/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhook

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	cardanov1alpha1 "github.com/AficaLabs/cardano-operator/api/v1alpha1"
)

// log is for logging in this package.
var stakepoollog = logf.Log.WithName("stakepool-webhook")

// MinCostLovelace is the minimum pool cost as per Cardano protocol (340 ADA)
const MinCostLovelace int64 = 340_000_000

// MinRelaysProduction is the minimum number of relays recommended for production
const MinRelaysProduction = 2

// MaxMetadataURLLength is the maximum length for metadata URL
const MaxMetadataURLLength = 64

// StakePoolValidator validates StakePool resources
type StakePoolValidator struct{}

// SetupStakePoolWebhookWithManager registers the webhook with the manager
func SetupStakePoolWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&cardanov1alpha1.StakePool{}).
		WithValidator(&StakePoolValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/validate-cardano-org-v1alpha1-stakepool,mutating=false,failurePolicy=fail,sideEffects=None,groups=cardano.org,resources=stakepools,verbs=create;update,versions=v1alpha1,name=vstakepool.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &StakePoolValidator{}

// ValidateCreate implements webhook.CustomValidator
func (v *StakePoolValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	sp, ok := obj.(*cardanov1alpha1.StakePool)
	if !ok {
		return nil, fmt.Errorf("expected StakePool, got %T", obj)
	}

	stakepoollog.Info("validating StakePool creation", "name", sp.Name, "namespace", sp.Namespace)

	var warnings admission.Warnings

	// Validate pool parameters
	if err := validatePoolParams(&sp.Spec.PoolParams, sp.Spec.Network); err != nil {
		return warnings, err
	}

	// Validate network
	if err := validateNetwork(sp.Spec.Network); err != nil {
		return warnings, err
	}

	// Validate key management configuration
	if err := validateKeyManagement(&sp.Spec.KeyManagement); err != nil {
		return warnings, err
	}

	// Validate payment configuration
	if err := validatePaymentConfig(&sp.Spec.PaymentConfig); err != nil {
		return warnings, err
	}

	// Validate storage configuration
	if err := validateStorageConfig(&sp.Spec.Storage); err != nil {
		return warnings, err
	}

	// Add warning for non-production relay count
	if len(sp.Spec.PoolParams.Relays) < MinRelaysProduction && sp.Spec.Network == cardanov1alpha1.NetworkMainnet {
		warnings = append(warnings, fmt.Sprintf(
			"only %d relay(s) configured; at least %d relays are recommended for mainnet production",
			len(sp.Spec.PoolParams.Relays), MinRelaysProduction))
	}

	return warnings, nil
}

// ValidateUpdate implements webhook.CustomValidator
func (v *StakePoolValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	sp, ok := newObj.(*cardanov1alpha1.StakePool)
	if !ok {
		return nil, fmt.Errorf("expected StakePool, got %T", newObj)
	}

	oldSP, ok := oldObj.(*cardanov1alpha1.StakePool)
	if !ok {
		return nil, fmt.Errorf("expected old StakePool, got %T", oldObj)
	}

	stakepoollog.Info("validating StakePool update", "name", sp.Name, "namespace", sp.Namespace)

	var warnings admission.Warnings

	// Network cannot be changed after creation
	if sp.Spec.Network != oldSP.Spec.Network {
		return warnings, fmt.Errorf("network cannot be changed after creation (was %s, attempted %s)",
			oldSP.Spec.Network, sp.Spec.Network)
	}

	// Validate pool parameters
	if err := validatePoolParams(&sp.Spec.PoolParams, sp.Spec.Network); err != nil {
		return warnings, err
	}

	// Validate key management configuration
	if err := validateKeyManagement(&sp.Spec.KeyManagement); err != nil {
		return warnings, err
	}

	// Validate payment configuration
	if err := validatePaymentConfig(&sp.Spec.PaymentConfig); err != nil {
		return warnings, err
	}

	// Validate storage configuration
	if err := validateStorageConfig(&sp.Spec.Storage); err != nil {
		return warnings, err
	}

	// Warn about key management mode changes
	if sp.Spec.KeyManagement.Mode != oldSP.Spec.KeyManagement.Mode {
		warnings = append(warnings, "changing key management mode may require manual key migration")
	}

	// Add warning for non-production relay count
	if len(sp.Spec.PoolParams.Relays) < MinRelaysProduction && sp.Spec.Network == cardanov1alpha1.NetworkMainnet {
		warnings = append(warnings, fmt.Sprintf(
			"only %d relay(s) configured; at least %d relays are recommended for mainnet production",
			len(sp.Spec.PoolParams.Relays), MinRelaysProduction))
	}

	return warnings, nil
}

// ValidateDelete implements webhook.CustomValidator
func (v *StakePoolValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	sp, ok := obj.(*cardanov1alpha1.StakePool)
	if !ok {
		return nil, fmt.Errorf("expected StakePool, got %T", obj)
	}

	stakepoollog.Info("validating StakePool deletion", "name", sp.Name, "namespace", sp.Namespace)

	var warnings admission.Warnings

	// Warn about pool retirement if active
	if sp.Status.Phase == cardanov1alpha1.StakePoolPhaseActive {
		warnings = append(warnings, "deleting an active stake pool; ensure the pool is properly retired on-chain first")
	}

	return warnings, nil
}

// validatePoolParams validates the pool parameters
func validatePoolParams(params *cardanov1alpha1.PoolParams, network cardanov1alpha1.Network) error {
	// Validate pledge is a valid lovelace amount
	pledge, err := strconv.ParseInt(params.Pledge, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid pledge format: must be a numeric value in lovelace")
	}
	if pledge < 0 {
		return fmt.Errorf("pledge cannot be negative")
	}
	// Mainnet pools should have a reasonable pledge (at least 1 ADA)
	if network == cardanov1alpha1.NetworkMainnet && pledge < 1000000 {
		return fmt.Errorf("mainnet pools should have at least 1 ADA pledge (1000000 lovelace)")
	}

	// Validate margin is between 0.0 and 1.0
	margin, err := strconv.ParseFloat(params.Margin, 64)
	if err != nil {
		return fmt.Errorf("invalid margin format: must be a decimal value between 0.0 and 1.0")
	}
	if margin < 0 || margin > 1 {
		return fmt.Errorf("margin must be between 0.0 and 1.0 (got %v)", margin)
	}

	// Validate cost meets minimum (340 ADA)
	cost, err := strconv.ParseInt(params.Cost, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid cost format: must be a numeric value in lovelace")
	}
	if cost < MinCostLovelace {
		return fmt.Errorf("pool cost must be at least 340 ADA (340000000 lovelace), got %d", cost)
	}

	// Validate metadata URL
	if err := validateMetadataURL(params.Metadata.URL); err != nil {
		return fmt.Errorf("invalid metadata URL: %w", err)
	}

	// Validate metadata hash if provided
	if params.Metadata.Hash != "" {
		if len(params.Metadata.Hash) != 64 {
			return fmt.Errorf("metadata hash must be 64 hex characters")
		}
	}

	// Validate relays
	if len(params.Relays) == 0 {
		return fmt.Errorf("at least one relay must be configured")
	}

	for i, relay := range params.Relays {
		if err := validateRelay(&relay, i); err != nil {
			return err
		}
	}

	return nil
}

// validateMetadataURL validates the pool metadata URL
func validateMetadataURL(metadataURL string) error {
	if len(metadataURL) > MaxMetadataURLLength {
		return fmt.Errorf("URL exceeds maximum length of %d characters (got %d)", MaxMetadataURLLength, len(metadataURL))
	}

	if metadataURL == "" {
		return fmt.Errorf("metadata URL is required")
	}

	parsed, err := url.Parse(metadataURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	if parsed.Host == "" {
		return fmt.Errorf("URL must have a host")
	}

	return nil
}

// validateRelay validates a relay configuration
func validateRelay(relay *cardanov1alpha1.RelayConfig, index int) error {
	switch relay.Type {
	case cardanov1alpha1.RelayTypeDNS:
		if relay.Hostname == "" {
			return fmt.Errorf("relay %d: hostname is required for dns type relay", index)
		}
		// Basic hostname validation
		if strings.ContainsAny(relay.Hostname, " \t\n") {
			return fmt.Errorf("relay %d: hostname contains invalid characters", index)
		}
	case cardanov1alpha1.RelayTypeIP:
		if relay.IPv4 == "" && relay.IPv6 == "" {
			return fmt.Errorf("relay %d: IPv4 or IPv6 address is required for ip type relay", index)
		}
		if relay.IPv4 != "" {
			if ip := net.ParseIP(relay.IPv4); ip == nil || ip.To4() == nil {
				return fmt.Errorf("relay %d: invalid IPv4 address: %s", index, relay.IPv4)
			}
		}
		if relay.IPv6 != "" {
			if ip := net.ParseIP(relay.IPv6); ip == nil || ip.To4() != nil {
				return fmt.Errorf("relay %d: invalid IPv6 address: %s", index, relay.IPv6)
			}
		}
	default:
		return fmt.Errorf("relay %d: unknown relay type: %s", index, relay.Type)
	}

	// Port validation is already handled by kubebuilder markers
	if relay.Port < 1 || relay.Port > 65535 {
		return fmt.Errorf("relay %d: port must be between 1 and 65535", index)
	}

	return nil
}

// validateNetwork validates the network
func validateNetwork(network cardanov1alpha1.Network) error {
	switch network {
	case cardanov1alpha1.NetworkMainnet, cardanov1alpha1.NetworkPreprod, cardanov1alpha1.NetworkPreview:
		return nil
	default:
		return fmt.Errorf("invalid network: %s (must be mainnet, preprod, or preview)", network)
	}
}

// validateKeyManagement validates the key management configuration
func validateKeyManagement(km *cardanov1alpha1.KeyManagement) error {
	switch km.Mode {
	case cardanov1alpha1.KeyManagementModeManaged:
		// Managed mode: operator generates keys
		// ColdKeySecretRef is optional (operator will create it)
	case cardanov1alpha1.KeyManagementModeExternal:
		// External mode: user provides keys via OfflineSigningNode
		if km.OfflineSigningNodeRef == "" {
			return fmt.Errorf("offlineSigningNodeRef is required when key management mode is 'external'")
		}
	default:
		return fmt.Errorf("invalid key management mode: %s (must be managed or external)", km.Mode)
	}

	return nil
}

// validatePaymentConfig validates the payment configuration
func validatePaymentConfig(pc *cardanov1alpha1.PaymentConfig) error {
	switch pc.Mode {
	case cardanov1alpha1.PaymentModeAutomated:
		// Automated mode: operator submits transactions
		if pc.PaymentKeySecretRef == nil {
			return fmt.Errorf("paymentKeySecretRef is required when payment mode is 'automated'")
		}
		if pc.PaymentKeySecretRef.Name == "" {
			return fmt.Errorf("paymentKeySecretRef.name is required")
		}
	case cardanov1alpha1.PaymentModeExternal:
		// External mode: user submits transactions manually
		// PaymentAddress is optional but recommended for display
	default:
		return fmt.Errorf("invalid payment mode: %s (must be automated or external)", pc.Mode)
	}

	return nil
}

// validateStorageConfig validates the storage configuration
func validateStorageConfig(sc *cardanov1alpha1.StorageConfig) error {
	if sc.Size == "" {
		return fmt.Errorf("storage size is required")
	}

	// Parse the size to validate format
	// The pattern validation is already in the CRD, but let's do a basic check
	size := strings.TrimSuffix(sc.Size, "i")
	size = strings.TrimSuffix(size, "G")
	size = strings.TrimSuffix(size, "T")
	size = strings.TrimSuffix(size, "M")
	size = strings.TrimSuffix(size, "K")
	size = strings.TrimSuffix(size, "P")
	size = strings.TrimSuffix(size, "E")

	sizeNum, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid storage size format: %s", sc.Size)
	}
	if sizeNum <= 0 {
		return fmt.Errorf("storage size must be positive")
	}

	return nil
}
