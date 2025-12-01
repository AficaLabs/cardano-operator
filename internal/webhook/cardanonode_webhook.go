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
var cardanonodelog = logf.Log.WithName("cardanonode-webhook")

// MinStorageSizeGi is the recommended minimum storage size in Gi
const MinStorageSizeGi = 100

// CardanoNodeValidator validates CardanoNode resources
type CardanoNodeValidator struct{}

// SetupCardanoNodeWebhookWithManager registers the webhook with the manager
func SetupCardanoNodeWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&cardanov1alpha1.CardanoNode{}).
		WithValidator(&CardanoNodeValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/validate-cardano-org-v1alpha1-cardanonode,mutating=false,failurePolicy=fail,sideEffects=None,groups=cardano.org,resources=cardanonodes,verbs=create;update,versions=v1alpha1,name=vcardanonode.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &CardanoNodeValidator{}

// ValidateCreate implements webhook.CustomValidator
func (v *CardanoNodeValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	cn, ok := obj.(*cardanov1alpha1.CardanoNode)
	if !ok {
		return nil, fmt.Errorf("expected CardanoNode, got %T", obj)
	}

	cardanonodelog.Info("validating CardanoNode creation", "name", cn.Name, "namespace", cn.Namespace)

	var warnings admission.Warnings

	// Validate node type
	if err := validateNodeType(cn.Spec.Type); err != nil {
		return warnings, err
	}

	// Validate network
	if err := validateNodeNetwork(cn.Spec.Network); err != nil {
		return warnings, err
	}

	// Validate storage
	storageSizeGi, err := parseStorageSizeGi(cn.Spec.Storage.Size)
	if err != nil {
		return warnings, fmt.Errorf("invalid storage size: %w", err)
	}

	// Warn about small storage
	if storageSizeGi < MinStorageSizeGi {
		warnings = append(warnings, fmt.Sprintf(
			"storage size %s is below recommended minimum of %dGi for %s",
			cn.Spec.Storage.Size, MinStorageSizeGi, cn.Spec.Network))
	}

	// Validate topology configuration
	if cn.Spec.Topology != nil {
		if err := validateTopology(cn.Spec.Topology, cn.Spec.Type); err != nil {
			return warnings, err
		}
	}

	// Block producer specific validations
	if cn.Spec.Type == cardanov1alpha1.CardanoNodeTypeBlockProducer {
		if cn.Spec.StakePoolRef == "" {
			warnings = append(warnings, "block-producer node has no stakePoolRef; it won't be able to mint blocks")
		}
	}

	return warnings, nil
}

// ValidateUpdate implements webhook.CustomValidator
func (v *CardanoNodeValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	cn, ok := newObj.(*cardanov1alpha1.CardanoNode)
	if !ok {
		return nil, fmt.Errorf("expected CardanoNode, got %T", newObj)
	}

	oldCN, ok := oldObj.(*cardanov1alpha1.CardanoNode)
	if !ok {
		return nil, fmt.Errorf("expected old CardanoNode, got %T", oldObj)
	}

	cardanonodelog.Info("validating CardanoNode update", "name", cn.Name, "namespace", cn.Namespace)

	var warnings admission.Warnings

	// Node type cannot be changed after creation
	if cn.Spec.Type != oldCN.Spec.Type {
		return warnings, fmt.Errorf("node type cannot be changed after creation (was %s, attempted %s)",
			oldCN.Spec.Type, cn.Spec.Type)
	}

	// Network cannot be changed after creation
	if cn.Spec.Network != oldCN.Spec.Network {
		return warnings, fmt.Errorf("network cannot be changed after creation (was %s, attempted %s)",
			oldCN.Spec.Network, cn.Spec.Network)
	}

	// Validate storage
	storageSizeGi, err := parseStorageSizeGi(cn.Spec.Storage.Size)
	if err != nil {
		return warnings, fmt.Errorf("invalid storage size: %w", err)
	}

	// Warn about small storage
	if storageSizeGi < MinStorageSizeGi {
		warnings = append(warnings, fmt.Sprintf(
			"storage size %s is below recommended minimum of %dGi for %s",
			cn.Spec.Storage.Size, MinStorageSizeGi, cn.Spec.Network))
	}

	// Warn about storage size reduction (PVC resize down may not work)
	oldStorageSizeGi, err := parseStorageSizeGi(oldCN.Spec.Storage.Size)
	if err == nil && storageSizeGi < oldStorageSizeGi {
		warnings = append(warnings, fmt.Sprintf(
			"reducing storage size from %s to %s may not be supported by your storage class",
			oldCN.Spec.Storage.Size, cn.Spec.Storage.Size))
	}

	// Validate topology configuration
	if cn.Spec.Topology != nil {
		if err := validateTopology(cn.Spec.Topology, cn.Spec.Type); err != nil {
			return warnings, err
		}
	}

	return warnings, nil
}

// ValidateDelete implements webhook.CustomValidator
func (v *CardanoNodeValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	cn, ok := obj.(*cardanov1alpha1.CardanoNode)
	if !ok {
		return nil, fmt.Errorf("expected CardanoNode, got %T", obj)
	}

	cardanonodelog.Info("validating CardanoNode deletion", "name", cn.Name, "namespace", cn.Namespace)

	var warnings admission.Warnings

	// Warn about deleting block producer
	if cn.Spec.Type == cardanov1alpha1.CardanoNodeTypeBlockProducer {
		warnings = append(warnings, "deleting a block-producer node will stop block production for the stake pool")
	}

	return warnings, nil
}

// validateNodeType validates the node type
func validateNodeType(nodeType cardanov1alpha1.CardanoNodeType) error {
	switch nodeType {
	case cardanov1alpha1.CardanoNodeTypeBlockProducer,
		cardanov1alpha1.CardanoNodeTypeRelay,
		cardanov1alpha1.CardanoNodeTypeOfflineSigning:
		return nil
	default:
		return fmt.Errorf("invalid node type: %s (must be block-producer, relay, or offline-signing)", nodeType)
	}
}

// validateNodeNetwork validates the network for a CardanoNode
func validateNodeNetwork(network cardanov1alpha1.Network) error {
	switch network {
	case cardanov1alpha1.NetworkMainnet, cardanov1alpha1.NetworkPreprod, cardanov1alpha1.NetworkPreview:
		return nil
	default:
		return fmt.Errorf("invalid network: %s (must be mainnet, preprod, or preview)", network)
	}
}

// validateTopology validates the topology configuration
func validateTopology(topology *cardanov1alpha1.TopologyConfig, nodeType cardanov1alpha1.CardanoNodeType) error {
	switch topology.Mode {
	case cardanov1alpha1.TopologyModeStatic:
		// Static topology requires peers to be configured
		if len(topology.StaticPeers) == 0 {
			return fmt.Errorf("static topology mode requires at least one peer to be configured")
		}
		// Validate each peer
		for i, peer := range topology.StaticPeers {
			if peer.Address == "" {
				return fmt.Errorf("peer %d: address is required", i)
			}
			if peer.Port < 1 || peer.Port > 65535 {
				return fmt.Errorf("peer %d: port must be between 1 and 65535", i)
			}
		}
	case cardanov1alpha1.TopologyModeP2P:
		// P2P mode is valid for relay nodes
		// Block producers typically use static topology to their relays
		// This is a warning case, not an error - warnings are handled in validate functions
		_ = nodeType // Intentionally unused here; warning logic is in validateCardanoNode
	default:
		return fmt.Errorf("invalid topology mode: %s (must be static or p2p)", topology.Mode)
	}

	return nil
}

// parseStorageSizeGi parses a storage size string and returns the size in Gi
func parseStorageSizeGi(size string) (int64, error) {
	size = strings.TrimSpace(size)
	if size == "" {
		return 0, fmt.Errorf("storage size is empty")
	}

	// Handle different suffixes and convert to Gi
	var multiplier int64 = 1

	if strings.HasSuffix(size, "Ti") {
		size = strings.TrimSuffix(size, "Ti")
		multiplier = 1024
	} else if strings.HasSuffix(size, "T") {
		size = strings.TrimSuffix(size, "T")
		multiplier = 1000 // TB to GB approximation
	} else if strings.HasSuffix(size, "Gi") {
		size = strings.TrimSuffix(size, "Gi")
		multiplier = 1
	} else if strings.HasSuffix(size, "G") {
		size = strings.TrimSuffix(size, "G")
		multiplier = 1 // GB approximately equals Gi for validation purposes
	} else if strings.HasSuffix(size, "Mi") {
		size = strings.TrimSuffix(size, "Mi")
		multiplier = 0 // Will be < 1 Gi
	} else if strings.HasSuffix(size, "M") {
		size = strings.TrimSuffix(size, "M")
		multiplier = 0 // Will be < 1 Gi
	} else if strings.HasSuffix(size, "Pi") {
		size = strings.TrimSuffix(size, "Pi")
		multiplier = 1024 * 1024
	} else if strings.HasSuffix(size, "P") {
		size = strings.TrimSuffix(size, "P")
		multiplier = 1000 * 1000
	} else if strings.HasSuffix(size, "Ei") {
		size = strings.TrimSuffix(size, "Ei")
		multiplier = 1024 * 1024 * 1024
	} else if strings.HasSuffix(size, "E") {
		size = strings.TrimSuffix(size, "E")
		multiplier = 1000 * 1000 * 1000
	} else if strings.HasSuffix(size, "Ki") {
		size = strings.TrimSuffix(size, "Ki")
		multiplier = 0 // Will be < 1 Gi
	} else if strings.HasSuffix(size, "K") {
		size = strings.TrimSuffix(size, "K")
		multiplier = 0 // Will be < 1 Gi
	}

	sizeNum, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %s", size)
	}

	if sizeNum < 0 {
		return 0, fmt.Errorf("storage size cannot be negative")
	}

	return sizeNum * multiplier, nil
}
