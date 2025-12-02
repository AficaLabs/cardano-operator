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

package cardano

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

// skipIfNoCLI skips the test if cardano-cli is not available
func skipIfNoCLI(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("cardano-cli"); err != nil {
		t.Skip("cardano-cli not found in PATH, skipping integration test")
	}
}

func TestExecutor_GenerateColdKeys_Integration(t *testing.T) {
	skipIfNoCLI(t)

	ctx := context.Background()
	executor, err := NewExecutor(&ClientConfig{
		Network: NetworkPreprod,
	})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	keys, err := executor.GenerateColdKeys(ctx)
	if err != nil {
		t.Fatalf("GenerateColdKeys() error = %v", err)
	}

	if keys == nil {
		t.Fatal("GenerateColdKeys() returned nil")
	}
	if len(keys.SigningKey) == 0 {
		t.Error("GenerateColdKeys() SigningKey is empty")
	}
	if len(keys.VerificationKey) == 0 {
		t.Error("GenerateColdKeys() VerificationKey is empty")
	}

	// Verify keys are valid JSON envelopes
	var skey map[string]interface{}
	if err := json.Unmarshal(keys.SigningKey, &skey); err != nil {
		t.Errorf("SigningKey is not valid JSON: %v", err)
	}
	if skey["type"] == nil {
		t.Error("SigningKey missing 'type' field")
	}

	var vkey map[string]interface{}
	if err := json.Unmarshal(keys.VerificationKey, &vkey); err != nil {
		t.Errorf("VerificationKey is not valid JSON: %v", err)
	}
	if vkey["type"] == nil {
		t.Error("VerificationKey missing 'type' field")
	}
}

func TestExecutor_GenerateVRFKeys_Integration(t *testing.T) {
	skipIfNoCLI(t)

	ctx := context.Background()
	executor, err := NewExecutor(&ClientConfig{
		Network: NetworkPreprod,
	})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	keys, err := executor.GenerateVRFKeys(ctx)
	if err != nil {
		t.Fatalf("GenerateVRFKeys() error = %v", err)
	}

	if keys == nil {
		t.Fatal("GenerateVRFKeys() returned nil")
	}
	if len(keys.SigningKey) == 0 {
		t.Error("GenerateVRFKeys() SigningKey is empty")
	}
	if len(keys.VerificationKey) == 0 {
		t.Error("GenerateVRFKeys() VerificationKey is empty")
	}

	// Verify VRF key type
	var vkey map[string]interface{}
	if err := json.Unmarshal(keys.VerificationKey, &vkey); err != nil {
		t.Errorf("VerificationKey is not valid JSON: %v", err)
	}
	keyType, ok := vkey["type"].(string)
	if !ok || !strings.Contains(keyType, "VRF") {
		t.Errorf("VerificationKey type = %v, expected VRF type", vkey["type"])
	}
}

func TestExecutor_GenerateKESKeys_Integration(t *testing.T) {
	skipIfNoCLI(t)

	ctx := context.Background()
	executor, err := NewExecutor(&ClientConfig{
		Network: NetworkPreprod,
	})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	keys, err := executor.GenerateKESKeys(ctx)
	if err != nil {
		t.Fatalf("GenerateKESKeys() error = %v", err)
	}

	if keys == nil {
		t.Fatal("GenerateKESKeys() returned nil")
	}
	if len(keys.SigningKey) == 0 {
		t.Error("GenerateKESKeys() SigningKey is empty")
	}
	if len(keys.VerificationKey) == 0 {
		t.Error("GenerateKESKeys() VerificationKey is empty")
	}

	// Verify KES key type (case-insensitive check for "kes")
	var vkey map[string]interface{}
	if err := json.Unmarshal(keys.VerificationKey, &vkey); err != nil {
		t.Errorf("VerificationKey is not valid JSON: %v", err)
	}
	keyType, ok := vkey["type"].(string)
	if !ok || !strings.Contains(strings.ToLower(keyType), "kes") {
		t.Errorf("VerificationKey type = %v, expected KES type", vkey["type"])
	}
}

func TestExecutor_GeneratePaymentKeys_Integration(t *testing.T) {
	skipIfNoCLI(t)

	ctx := context.Background()
	executor, err := NewExecutor(&ClientConfig{
		Network: NetworkPreprod,
	})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	keys, err := executor.GeneratePaymentKeys(ctx)
	if err != nil {
		t.Fatalf("GeneratePaymentKeys() error = %v", err)
	}

	if keys == nil {
		t.Fatal("GeneratePaymentKeys() returned nil")
	}
	if len(keys.SigningKey) == 0 {
		t.Error("GeneratePaymentKeys() SigningKey is empty")
	}
	if len(keys.VerificationKey) == 0 {
		t.Error("GeneratePaymentKeys() VerificationKey is empty")
	}

	// Verify payment key type
	var vkey map[string]interface{}
	if err := json.Unmarshal(keys.VerificationKey, &vkey); err != nil {
		t.Errorf("VerificationKey is not valid JSON: %v", err)
	}
	keyType, ok := vkey["type"].(string)
	if !ok || !strings.Contains(keyType, "Payment") {
		t.Errorf("VerificationKey type = %v, expected Payment type", vkey["type"])
	}
}

func TestExecutor_GenerateStakeKeys_Integration(t *testing.T) {
	skipIfNoCLI(t)

	ctx := context.Background()
	executor, err := NewExecutor(&ClientConfig{
		Network: NetworkPreprod,
	})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	keys, err := executor.GenerateStakeKeys(ctx)
	if err != nil {
		t.Fatalf("GenerateStakeKeys() error = %v", err)
	}

	if keys == nil {
		t.Fatal("GenerateStakeKeys() returned nil")
	}
	if len(keys.SigningKey) == 0 {
		t.Error("GenerateStakeKeys() SigningKey is empty")
	}
	if len(keys.VerificationKey) == 0 {
		t.Error("GenerateStakeKeys() VerificationKey is empty")
	}

	// Verify stake key type
	var vkey map[string]interface{}
	if err := json.Unmarshal(keys.VerificationKey, &vkey); err != nil {
		t.Errorf("VerificationKey is not valid JSON: %v", err)
	}
	keyType, ok := vkey["type"].(string)
	if !ok || !strings.Contains(keyType, "Stake") {
		t.Errorf("VerificationKey type = %v, expected Stake type", vkey["type"])
	}
}

func TestExecutor_BuildAddress_Integration(t *testing.T) {
	skipIfNoCLI(t)

	ctx := context.Background()
	executor, err := NewExecutor(&ClientConfig{
		Network: NetworkPreprod,
	})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	// First generate the keys we need
	paymentKeys, err := executor.GeneratePaymentKeys(ctx)
	if err != nil {
		t.Fatalf("GeneratePaymentKeys() error = %v", err)
	}

	stakeKeys, err := executor.GenerateStakeKeys(ctx)
	if err != nil {
		t.Fatalf("GenerateStakeKeys() error = %v", err)
	}

	// Build the address
	addr, err := executor.BuildAddress(ctx, paymentKeys.VerificationKey, stakeKeys.VerificationKey)
	if err != nil {
		t.Fatalf("BuildAddress() error = %v", err)
	}

	if addr == "" {
		t.Error("BuildAddress() returned empty address")
	}

	// Preprod addresses start with addr_test1
	if !strings.HasPrefix(addr, "addr_test1") {
		t.Errorf("BuildAddress() = %v, expected addr_test1 prefix for preprod", addr)
	}
}

func TestExecutor_BuildStakeAddress_Integration(t *testing.T) {
	skipIfNoCLI(t)

	ctx := context.Background()
	executor, err := NewExecutor(&ClientConfig{
		Network: NetworkPreprod,
	})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	// First generate stake keys
	stakeKeys, err := executor.GenerateStakeKeys(ctx)
	if err != nil {
		t.Fatalf("GenerateStakeKeys() error = %v", err)
	}

	// Build the stake address
	addr, err := executor.BuildStakeAddress(ctx, stakeKeys.VerificationKey)
	if err != nil {
		t.Fatalf("BuildStakeAddress() error = %v", err)
	}

	if addr == "" {
		t.Error("BuildStakeAddress() returned empty address")
	}

	// Preprod stake addresses start with stake_test1
	if !strings.HasPrefix(addr, "stake_test1") {
		t.Errorf("BuildStakeAddress() = %v, expected stake_test1 prefix for preprod", addr)
	}
}

func TestExecutor_BuildAddress_Mainnet_Integration(t *testing.T) {
	skipIfNoCLI(t)

	ctx := context.Background()
	executor, err := NewExecutor(&ClientConfig{
		Network: NetworkMainnet,
	})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	// Generate keys
	paymentKeys, err := executor.GeneratePaymentKeys(ctx)
	if err != nil {
		t.Fatalf("GeneratePaymentKeys() error = %v", err)
	}

	stakeKeys, err := executor.GenerateStakeKeys(ctx)
	if err != nil {
		t.Fatalf("GenerateStakeKeys() error = %v", err)
	}

	// Build mainnet address
	addr, err := executor.BuildAddress(ctx, paymentKeys.VerificationKey, stakeKeys.VerificationKey)
	if err != nil {
		t.Fatalf("BuildAddress() error = %v", err)
	}

	// Mainnet addresses start with addr1
	if !strings.HasPrefix(addr, "addr1") {
		t.Errorf("BuildAddress() = %v, expected addr1 prefix for mainnet", addr)
	}
}

func TestExecutor_runCLI_Integration(t *testing.T) {
	skipIfNoCLI(t)

	ctx := context.Background()
	executor, err := NewExecutor(&ClientConfig{
		Network: NetworkPreprod,
	})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	// Test a simple command that doesn't need a node
	output, err := executor.runCLI(ctx, "--version")
	if err != nil {
		t.Fatalf("runCLI(--version) error = %v", err)
	}

	if !strings.Contains(string(output), "cardano-cli") {
		t.Errorf("runCLI(--version) output doesn't contain 'cardano-cli': %s", output)
	}
}

func TestExecutor_networkArgs_Integration(t *testing.T) {
	tests := []struct {
		name    string
		network Network
		magic   int64
		want    string
	}{
		{"mainnet", NetworkMainnet, 764824073, "--mainnet"},
		{"preprod", NetworkPreprod, 1, "--testnet-magic"},
		{"preview", NetworkPreview, 2, "--testnet-magic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, _ := NewExecutor(&ClientConfig{
				Network: tt.network,
			})

			args := executor.networkArgs()
			if len(args) == 0 {
				t.Error("networkArgs() returned empty")
			}
			if args[0] != tt.want {
				t.Errorf("networkArgs()[0] = %v, want %v", args[0], tt.want)
			}
		})
	}
}

func TestExecutor_AllKeyTypes_Integration(t *testing.T) {
	skipIfNoCLI(t)

	ctx := context.Background()
	executor, err := NewExecutor(&ClientConfig{
		Network: NetworkPreprod,
	})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}

	// Generate all key types that a stake pool needs
	t.Run("cold keys", func(t *testing.T) {
		keys, err := executor.GenerateColdKeys(ctx)
		if err != nil {
			t.Errorf("GenerateColdKeys() error = %v", err)
		}
		if keys == nil || len(keys.SigningKey) == 0 || len(keys.VerificationKey) == 0 {
			t.Error("GenerateColdKeys() returned invalid keys")
		}
	})

	t.Run("VRF keys", func(t *testing.T) {
		keys, err := executor.GenerateVRFKeys(ctx)
		if err != nil {
			t.Errorf("GenerateVRFKeys() error = %v", err)
		}
		if keys == nil || len(keys.SigningKey) == 0 || len(keys.VerificationKey) == 0 {
			t.Error("GenerateVRFKeys() returned invalid keys")
		}
	})

	t.Run("KES keys", func(t *testing.T) {
		keys, err := executor.GenerateKESKeys(ctx)
		if err != nil {
			t.Errorf("GenerateKESKeys() error = %v", err)
		}
		if keys == nil || len(keys.SigningKey) == 0 || len(keys.VerificationKey) == 0 {
			t.Error("GenerateKESKeys() returned invalid keys")
		}
	})

	t.Run("payment keys", func(t *testing.T) {
		keys, err := executor.GeneratePaymentKeys(ctx)
		if err != nil {
			t.Errorf("GeneratePaymentKeys() error = %v", err)
		}
		if keys == nil || len(keys.SigningKey) == 0 || len(keys.VerificationKey) == 0 {
			t.Error("GeneratePaymentKeys() returned invalid keys")
		}
	})

	t.Run("stake keys", func(t *testing.T) {
		keys, err := executor.GenerateStakeKeys(ctx)
		if err != nil {
			t.Errorf("GenerateStakeKeys() error = %v", err)
		}
		if keys == nil || len(keys.SigningKey) == 0 || len(keys.VerificationKey) == 0 {
			t.Error("GenerateStakeKeys() returned invalid keys")
		}
	})
}
