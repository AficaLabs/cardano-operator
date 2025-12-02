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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cardanov1alpha1 "github.com/AficaLabs/cardano-operator/api/v1alpha1"
)

func newValidStakePool() *cardanov1alpha1.StakePool {
	return &cardanov1alpha1.StakePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pool",
			Namespace: "default",
		},
		Spec: cardanov1alpha1.StakePoolSpec{
			Network: cardanov1alpha1.NetworkPreprod,
			PoolParams: cardanov1alpha1.PoolParams{
				Pledge: "500000000000",
				Margin: "0.03",
				Cost:   "340000000",
				Metadata: cardanov1alpha1.PoolMetadata{
					URL:  "https://example.com/pool.json",
					Hash: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
				},
				Relays: []cardanov1alpha1.RelayConfig{
					{
						Type:     cardanov1alpha1.RelayTypeDNS,
						Hostname: "relay1.example.com",
						Port:     6000,
					},
				},
			},
			KeyManagement: cardanov1alpha1.KeyManagement{
				Mode: cardanov1alpha1.KeyManagementModeManaged,
			},
			PaymentConfig: cardanov1alpha1.PaymentConfig{
				Mode: cardanov1alpha1.PaymentModeExternal,
			},
			Storage: cardanov1alpha1.StorageConfig{
				Size: "200Gi",
			},
		},
	}
}

const invalidNumberStr = "not-a-number"

func TestStakePoolValidatorValidateCreate(t *testing.T) {
	ctx := context.Background()
	validator := &StakePoolValidator{}

	t.Run("valid stakepool", func(t *testing.T) {
		sp := newValidStakePool()
		warnings, err := validator.ValidateCreate(ctx, sp)
		if err != nil {
			t.Errorf("ValidateCreate() error = %v", err)
		}
		if len(warnings) != 0 {
			t.Errorf("ValidateCreate() warnings = %v, want none", warnings)
		}
	})

	t.Run("invalid object type", func(t *testing.T) {
		_, err := validator.ValidateCreate(ctx, &cardanov1alpha1.CardanoNode{})
		if err == nil {
			t.Error("ValidateCreate() expected error for wrong object type")
		}
	})

	t.Run("mainnet with single relay warning", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.Network = cardanov1alpha1.NetworkMainnet
		sp.Spec.PoolParams.Pledge = "1000000" // 1 ADA minimum for mainnet
		warnings, err := validator.ValidateCreate(ctx, sp)
		if err != nil {
			t.Errorf("ValidateCreate() error = %v", err)
		}
		if len(warnings) == 0 {
			t.Error("ValidateCreate() expected warning for single relay on mainnet")
		}
	})

	t.Run("invalid pledge format", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.PoolParams.Pledge = invalidNumberStr
		_, err := validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for invalid pledge")
		}
	})

	t.Run("negative pledge", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.PoolParams.Pledge = "-100"
		_, err := validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for negative pledge")
		}
	})

	t.Run("mainnet with insufficient pledge", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.Network = cardanov1alpha1.NetworkMainnet
		sp.Spec.PoolParams.Pledge = "100" // Less than 1 ADA
		_, err := validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for insufficient mainnet pledge")
		}
	})

	t.Run("invalid margin format", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.PoolParams.Margin = invalidNumberStr
		_, err := validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for invalid margin")
		}
	})

	t.Run("margin out of range", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.PoolParams.Margin = "1.5"
		_, err := validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for margin > 1")
		}

		sp.Spec.PoolParams.Margin = "-0.1"
		_, err = validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for margin < 0")
		}
	})

	t.Run("invalid cost format", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.PoolParams.Cost = invalidNumberStr
		_, err := validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for invalid cost")
		}
	})

	t.Run("cost below minimum", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.PoolParams.Cost = "100000000" // 100 ADA, less than 340 ADA
		_, err := validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for cost below minimum")
		}
	})

	t.Run("invalid network", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.Network = "invalid-network"
		_, err := validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for invalid network")
		}
	})

	t.Run("external key management without ref", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.KeyManagement.Mode = cardanov1alpha1.KeyManagementModeExternal
		sp.Spec.KeyManagement.OfflineSigningNodeRef = ""
		_, err := validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for external mode without ref")
		}
	})

	t.Run("external key management with ref", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.KeyManagement.Mode = cardanov1alpha1.KeyManagementModeExternal
		sp.Spec.KeyManagement.OfflineSigningNodeRef = "my-signing-node"
		_, err := validator.ValidateCreate(ctx, sp)
		if err != nil {
			t.Errorf("ValidateCreate() error = %v", err)
		}
	})

	t.Run("invalid key management mode", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.KeyManagement.Mode = "invalid-mode"
		_, err := validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for invalid key management mode")
		}
	})

	t.Run("automated payment without secret ref", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.PaymentConfig.Mode = cardanov1alpha1.PaymentModeAutomated
		sp.Spec.PaymentConfig.PaymentKeySecretRef = nil
		_, err := validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for automated payment without secret")
		}
	})

	t.Run("automated payment with empty secret name", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.PaymentConfig.Mode = cardanov1alpha1.PaymentModeAutomated
		sp.Spec.PaymentConfig.PaymentKeySecretRef = &cardanov1alpha1.SecretReference{Name: ""}
		_, err := validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for empty secret name")
		}
	})

	t.Run("automated payment with valid secret", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.PaymentConfig.Mode = cardanov1alpha1.PaymentModeAutomated
		sp.Spec.PaymentConfig.PaymentKeySecretRef = &cardanov1alpha1.SecretReference{
			Name: "payment-key",
			Key:  "payment.skey",
		}
		_, err := validator.ValidateCreate(ctx, sp)
		if err != nil {
			t.Errorf("ValidateCreate() error = %v", err)
		}
	})

	t.Run("invalid payment mode", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.PaymentConfig.Mode = "invalid-mode"
		_, err := validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for invalid payment mode")
		}
	})

	t.Run("empty storage size", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.Storage.Size = ""
		_, err := validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for empty storage size")
		}
	})

	t.Run("invalid storage size format", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.Storage.Size = "abc"
		_, err := validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for invalid storage size format")
		}
	})

	t.Run("negative storage size", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Spec.Storage.Size = "-100Gi"
		_, err := validator.ValidateCreate(ctx, sp)
		if err == nil {
			t.Error("ValidateCreate() expected error for negative storage size")
		}
	})
}

func TestStakePoolValidatorValidateUpdate(t *testing.T) {
	ctx := context.Background()
	validator := &StakePoolValidator{}

	t.Run("valid update", func(t *testing.T) {
		oldSP := newValidStakePool()
		newSP := newValidStakePool()
		newSP.Spec.PoolParams.Margin = "0.05"
		warnings, err := validator.ValidateUpdate(ctx, oldSP, newSP)
		if err != nil {
			t.Errorf("ValidateUpdate() error = %v", err)
		}
		if len(warnings) != 0 {
			t.Errorf("ValidateUpdate() warnings = %v, want none", warnings)
		}
	})

	t.Run("invalid new object type", func(t *testing.T) {
		oldSP := newValidStakePool()
		_, err := validator.ValidateUpdate(ctx, oldSP, &cardanov1alpha1.CardanoNode{})
		if err == nil {
			t.Error("ValidateUpdate() expected error for wrong new object type")
		}
	})

	t.Run("invalid old object type", func(t *testing.T) {
		newSP := newValidStakePool()
		_, err := validator.ValidateUpdate(ctx, &cardanov1alpha1.CardanoNode{}, newSP)
		if err == nil {
			t.Error("ValidateUpdate() expected error for wrong old object type")
		}
	})

	t.Run("network change rejected", func(t *testing.T) {
		oldSP := newValidStakePool()
		oldSP.Spec.Network = cardanov1alpha1.NetworkPreprod
		newSP := newValidStakePool()
		newSP.Spec.Network = cardanov1alpha1.NetworkMainnet
		_, err := validator.ValidateUpdate(ctx, oldSP, newSP)
		if err == nil {
			t.Error("ValidateUpdate() expected error for network change")
		}
	})

	t.Run("key management mode change warning", func(t *testing.T) {
		oldSP := newValidStakePool()
		oldSP.Spec.KeyManagement.Mode = cardanov1alpha1.KeyManagementModeManaged
		newSP := newValidStakePool()
		newSP.Spec.KeyManagement.Mode = cardanov1alpha1.KeyManagementModeExternal
		newSP.Spec.KeyManagement.OfflineSigningNodeRef = "my-signing-node"
		warnings, err := validator.ValidateUpdate(ctx, oldSP, newSP)
		if err != nil {
			t.Errorf("ValidateUpdate() error = %v", err)
		}
		if len(warnings) == 0 {
			t.Error("ValidateUpdate() expected warning for key management mode change")
		}
	})

	t.Run("mainnet with single relay warning", func(t *testing.T) {
		oldSP := newValidStakePool()
		oldSP.Spec.Network = cardanov1alpha1.NetworkMainnet
		oldSP.Spec.PoolParams.Pledge = "1000000"
		newSP := newValidStakePool()
		newSP.Spec.Network = cardanov1alpha1.NetworkMainnet
		newSP.Spec.PoolParams.Pledge = "1000000"
		warnings, err := validator.ValidateUpdate(ctx, oldSP, newSP)
		if err != nil {
			t.Errorf("ValidateUpdate() error = %v", err)
		}
		if len(warnings) == 0 {
			t.Error("ValidateUpdate() expected warning for single relay on mainnet")
		}
	})
}

func TestStakePoolValidatorValidateDelete(t *testing.T) {
	ctx := context.Background()
	validator := &StakePoolValidator{}

	t.Run("delete inactive pool", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Status.Phase = cardanov1alpha1.StakePoolPhasePending
		warnings, err := validator.ValidateDelete(ctx, sp)
		if err != nil {
			t.Errorf("ValidateDelete() error = %v", err)
		}
		if len(warnings) != 0 {
			t.Errorf("ValidateDelete() warnings = %v, want none", warnings)
		}
	})

	t.Run("delete active pool warning", func(t *testing.T) {
		sp := newValidStakePool()
		sp.Status.Phase = cardanov1alpha1.StakePoolPhaseActive
		warnings, err := validator.ValidateDelete(ctx, sp)
		if err != nil {
			t.Errorf("ValidateDelete() error = %v", err)
		}
		if len(warnings) == 0 {
			t.Error("ValidateDelete() expected warning for deleting active pool")
		}
	})

	t.Run("invalid object type", func(t *testing.T) {
		_, err := validator.ValidateDelete(ctx, &cardanov1alpha1.CardanoNode{})
		if err == nil {
			t.Error("ValidateDelete() expected error for wrong object type")
		}
	})
}

func TestValidateMetadataURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid https url", "https://example.com/pool.json", false},
		{"valid http url", "http://example.com/pool.json", false},
		{"empty url", "", true},
		{"url too long", "https://example.com/" + string(make([]byte, 100)), true},
		{"invalid scheme", "ftp://example.com/pool.json", true},
		{"no host", "https:///pool.json", true},
		{"invalid url", "not-a-url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMetadataURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMetadataURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRelay(t *testing.T) {
	tests := []struct {
		name    string
		relay   *cardanov1alpha1.RelayConfig
		index   int
		wantErr bool
	}{
		{
			name: "valid dns relay",
			relay: &cardanov1alpha1.RelayConfig{
				Type:     cardanov1alpha1.RelayTypeDNS,
				Hostname: "relay.example.com",
				Port:     6000,
			},
			index:   0,
			wantErr: false,
		},
		{
			name: "dns relay without hostname",
			relay: &cardanov1alpha1.RelayConfig{
				Type:     cardanov1alpha1.RelayTypeDNS,
				Hostname: "",
				Port:     6000,
			},
			index:   0,
			wantErr: true,
		},
		{
			name: "dns relay with invalid hostname",
			relay: &cardanov1alpha1.RelayConfig{
				Type:     cardanov1alpha1.RelayTypeDNS,
				Hostname: "relay example.com",
				Port:     6000,
			},
			index:   0,
			wantErr: true,
		},
		{
			name: "valid ip relay with ipv4",
			relay: &cardanov1alpha1.RelayConfig{
				Type: cardanov1alpha1.RelayTypeIP,
				IPv4: "192.168.1.1",
				Port: 6000,
			},
			index:   0,
			wantErr: false,
		},
		{
			name: "valid ip relay with ipv6",
			relay: &cardanov1alpha1.RelayConfig{
				Type: cardanov1alpha1.RelayTypeIP,
				IPv6: "2001:db8::1",
				Port: 6000,
			},
			index:   0,
			wantErr: false,
		},
		{
			name: "ip relay without address",
			relay: &cardanov1alpha1.RelayConfig{
				Type: cardanov1alpha1.RelayTypeIP,
				Port: 6000,
			},
			index:   0,
			wantErr: true,
		},
		{
			name: "ip relay with invalid ipv4",
			relay: &cardanov1alpha1.RelayConfig{
				Type: cardanov1alpha1.RelayTypeIP,
				IPv4: "not-an-ip",
				Port: 6000,
			},
			index:   0,
			wantErr: true,
		},
		{
			name: "ip relay with ipv6 in ipv4 field",
			relay: &cardanov1alpha1.RelayConfig{
				Type: cardanov1alpha1.RelayTypeIP,
				IPv4: "2001:db8::1",
				Port: 6000,
			},
			index:   0,
			wantErr: true,
		},
		{
			name: "ip relay with ipv4 in ipv6 field",
			relay: &cardanov1alpha1.RelayConfig{
				Type: cardanov1alpha1.RelayTypeIP,
				IPv6: "192.168.1.1",
				Port: 6000,
			},
			index:   0,
			wantErr: true,
		},
		{
			name: "invalid relay type",
			relay: &cardanov1alpha1.RelayConfig{
				Type:     "invalid",
				Hostname: "relay.example.com",
				Port:     6000,
			},
			index:   0,
			wantErr: true,
		},
		{
			name: "invalid port (0)",
			relay: &cardanov1alpha1.RelayConfig{
				Type:     cardanov1alpha1.RelayTypeDNS,
				Hostname: "relay.example.com",
				Port:     0,
			},
			index:   0,
			wantErr: true,
		},
		{
			name: "invalid port (too high)",
			relay: &cardanov1alpha1.RelayConfig{
				Type:     cardanov1alpha1.RelayTypeDNS,
				Hostname: "relay.example.com",
				Port:     70000,
			},
			index:   0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRelay(tt.relay, tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRelay() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNetwork(t *testing.T) {
	tests := []struct {
		name    string
		network cardanov1alpha1.Network
		wantErr bool
	}{
		{"mainnet", cardanov1alpha1.NetworkMainnet, false},
		{"preprod", cardanov1alpha1.NetworkPreprod, false},
		{"preview", cardanov1alpha1.NetworkPreview, false},
		{"invalid", "invalid", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNetwork(tt.network)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateNetwork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePoolParams(t *testing.T) {
	t.Run("no relays", func(t *testing.T) {
		params := &cardanov1alpha1.PoolParams{
			Pledge: "500000000000",
			Margin: "0.03",
			Cost:   "340000000",
			Metadata: cardanov1alpha1.PoolMetadata{
				URL:  "https://example.com/pool.json",
				Hash: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			},
			Relays: []cardanov1alpha1.RelayConfig{},
		}
		err := validatePoolParams(params, cardanov1alpha1.NetworkPreprod)
		if err == nil {
			t.Error("validatePoolParams() expected error for no relays")
		}
	})

	t.Run("invalid metadata hash length", func(t *testing.T) {
		params := &cardanov1alpha1.PoolParams{
			Pledge: "500000000000",
			Margin: "0.03",
			Cost:   "340000000",
			Metadata: cardanov1alpha1.PoolMetadata{
				URL:  "https://example.com/pool.json",
				Hash: "tooshort",
			},
			Relays: []cardanov1alpha1.RelayConfig{
				{Type: cardanov1alpha1.RelayTypeDNS, Hostname: "relay.example.com", Port: 6000},
			},
		}
		err := validatePoolParams(params, cardanov1alpha1.NetworkPreprod)
		if err == nil {
			t.Error("validatePoolParams() expected error for invalid hash length")
		}
	})

	t.Run("empty metadata hash allowed", func(t *testing.T) {
		params := &cardanov1alpha1.PoolParams{
			Pledge: "500000000000",
			Margin: "0.03",
			Cost:   "340000000",
			Metadata: cardanov1alpha1.PoolMetadata{
				URL:  "https://example.com/pool.json",
				Hash: "",
			},
			Relays: []cardanov1alpha1.RelayConfig{
				{Type: cardanov1alpha1.RelayTypeDNS, Hostname: "relay.example.com", Port: 6000},
			},
		}
		err := validatePoolParams(params, cardanov1alpha1.NetworkPreprod)
		if err != nil {
			t.Errorf("validatePoolParams() error = %v, want nil", err)
		}
	})
}

func TestValidateStorageConfig(t *testing.T) {
	tests := []struct {
		name    string
		size    string
		wantErr bool
	}{
		{"valid Gi", "200Gi", false},
		{"valid G", "200G", false},
		{"valid Ti", "1Ti", false},
		{"valid T", "1T", false},
		{"valid Mi", "500Mi", false},
		{"valid M", "500M", false},
		{"valid Ki", "1000Ki", false},
		{"valid K", "1000K", false},
		{"valid Pi", "1Pi", false},
		{"valid P", "1P", false},
		{"valid Ei", "1Ei", false},
		{"valid E", "1E", false},
		{"empty", "", true},
		{"invalid format", "abc", true},
		{"negative", "-100Gi", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &cardanov1alpha1.StorageConfig{Size: tt.size}
			err := validateStorageConfig(sc)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStorageConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
