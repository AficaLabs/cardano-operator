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
	"errors"
	"testing"
)

// mockClient is a mock implementation of the Client interface for testing
type mockClient struct {
	generateColdKeysFunc    func(ctx context.Context) (*KeyPair, error)
	generateVRFKeysFunc     func(ctx context.Context) (*KeyPair, error)
	generateKESKeysFunc     func(ctx context.Context) (*KeyPair, error)
	generatePaymentKeysFunc func(ctx context.Context) (*KeyPair, error)
	generateStakeKeysFunc   func(ctx context.Context) (*KeyPair, error)
	getCurrentKESPeriodFunc func(ctx context.Context) (int64, error)
	createOpCertFunc        func(ctx context.Context, kesSKey []byte, coldSKey []byte, counter int64, kesPeriod int64) (*OperationalCertificate, error)
}

func (m *mockClient) QueryTip(ctx context.Context) (*Tip, error) {
	return &Tip{Slot: 1000, Epoch: 100, Block: 500}, nil
}

func (m *mockClient) QueryUTxO(ctx context.Context, address string) ([]UTxO, error) {
	return []UTxO{{TxHash: "abc123", TxIndex: 0, Address: address, Value: Value{Lovelace: 10000000}}}, nil
}

func (m *mockClient) QueryProtocolParameters(ctx context.Context) (*ProtocolParameters, error) {
	return &ProtocolParameters{PoolDeposit: 500000000, MaxEpoch: 18}, nil
}

func (m *mockClient) QueryPoolParams(ctx context.Context, poolID string) (*PoolParams, error) {
	return nil, nil
}

func (m *mockClient) QueryStakePoolID(ctx context.Context, coldVKeyPath string) (string, error) {
	return "pool1abc123", nil
}

func (m *mockClient) GenerateColdKeys(ctx context.Context) (*KeyPair, error) {
	if m.generateColdKeysFunc != nil {
		return m.generateColdKeysFunc(ctx)
	}
	return &KeyPair{
		SigningKey:      []byte(`{"type":"StakePoolSigningKey_ed25519","description":"","cborHex":"coldSKey"}`),
		VerificationKey: []byte(`{"type":"StakePoolVerificationKey_ed25519","description":"","cborHex":"coldVKey"}`),
	}, nil
}

func (m *mockClient) GenerateVRFKeys(ctx context.Context) (*KeyPair, error) {
	if m.generateVRFKeysFunc != nil {
		return m.generateVRFKeysFunc(ctx)
	}
	return &KeyPair{
		SigningKey:      []byte(`{"type":"VrfSigningKey_PraosVRF","description":"","cborHex":"vrfSKey"}`),
		VerificationKey: []byte(`{"type":"VrfVerificationKey_PraosVRF","description":"","cborHex":"vrfVKey"}`),
	}, nil
}

func (m *mockClient) GenerateKESKeys(ctx context.Context) (*KeyPair, error) {
	if m.generateKESKeysFunc != nil {
		return m.generateKESKeysFunc(ctx)
	}
	return &KeyPair{
		SigningKey:      []byte(`{"type":"KesSigningKey_ed25519_kes_2^6","description":"","cborHex":"kesSKey"}`),
		VerificationKey: []byte(`{"type":"KesVerificationKey_ed25519_kes_2^6","description":"","cborHex":"kesVKey"}`),
	}, nil
}

func (m *mockClient) GeneratePaymentKeys(ctx context.Context) (*KeyPair, error) {
	if m.generatePaymentKeysFunc != nil {
		return m.generatePaymentKeysFunc(ctx)
	}
	return &KeyPair{
		SigningKey:      []byte(`{"type":"PaymentSigningKeyShelley_ed25519","description":"","cborHex":"paymentSKey"}`),
		VerificationKey: []byte(`{"type":"PaymentVerificationKeyShelley_ed25519","description":"","cborHex":"paymentVKey"}`),
	}, nil
}

func (m *mockClient) GenerateStakeKeys(ctx context.Context) (*KeyPair, error) {
	if m.generateStakeKeysFunc != nil {
		return m.generateStakeKeysFunc(ctx)
	}
	return &KeyPair{
		SigningKey:      []byte(`{"type":"StakeSigningKeyShelley_ed25519","description":"","cborHex":"stakeSKey"}`),
		VerificationKey: []byte(`{"type":"StakeVerificationKeyShelley_ed25519","description":"","cborHex":"stakeVKey"}`),
	}, nil
}

func (m *mockClient) CreateOperationalCertificate(ctx context.Context, kesSKey []byte, coldSKey []byte, counter int64, kesPeriod int64) (*OperationalCertificate, error) {
	if m.createOpCertFunc != nil {
		return m.createOpCertFunc(ctx, kesSKey, coldSKey, counter, kesPeriod)
	}
	return &OperationalCertificate{
		Certificate: []byte("mock-certificate"),
		Counter:     counter + 1,
		KESPeriod:   kesPeriod,
		ExpiryEpoch: 200,
	}, nil
}

func (m *mockClient) CreatePoolRegistrationCertificate(ctx context.Context, params *PoolParams, coldVKey []byte) ([]byte, error) {
	return []byte("mock-pool-reg-cert"), nil
}

func (m *mockClient) CreatePoolUpdateCertificate(ctx context.Context, params *PoolParams, coldVKey []byte) ([]byte, error) {
	return []byte("mock-pool-update-cert"), nil
}

func (m *mockClient) CreatePoolRetirementCertificate(ctx context.Context, poolID string, retirementEpoch int64, coldVKey []byte) ([]byte, error) {
	return []byte("mock-pool-retire-cert"), nil
}

func (m *mockClient) BuildTx(ctx context.Context, opts *TxBuildOptions) ([]byte, error) {
	return []byte("mock-tx-body"), nil
}

func (m *mockClient) SignTx(ctx context.Context, txBody []byte, signingKeys ...[]byte) ([]byte, error) {
	return []byte("mock-signed-tx"), nil
}

func (m *mockClient) SubmitTx(ctx context.Context, signedTx []byte) (*TransactionSubmitResult, error) {
	return &TransactionSubmitResult{TxHash: "tx-hash-123", Success: true}, nil
}

func (m *mockClient) BuildAddress(ctx context.Context, paymentVKey []byte, stakeVKey []byte) (string, error) {
	return "addr_test1abc123", nil
}

func (m *mockClient) BuildStakeAddress(ctx context.Context, stakeVKey []byte) (string, error) {
	return "stake_test1abc123", nil
}

func (m *mockClient) CalculateMinFee(ctx context.Context, txBody []byte, witnessCount int) (int64, error) {
	return 200000, nil
}

func (m *mockClient) GetCurrentKESPeriod(ctx context.Context) (int64, error) {
	if m.getCurrentKESPeriodFunc != nil {
		return m.getCurrentKESPeriodFunc(ctx)
	}
	return 100, nil
}

// TestParseKeyEnvelope tests parsing of key envelope JSON
func TestParseKeyEnvelope(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		wantType    string
		wantCborHex string
		wantErr     bool
	}{
		{
			name:        "valid cold signing key",
			input:       []byte(`{"type":"StakePoolSigningKey_ed25519","description":"Stake Pool Signing Key","cborHex":"5820abc123"}`),
			wantType:    "StakePoolSigningKey_ed25519",
			wantCborHex: "5820abc123",
			wantErr:     false,
		},
		{
			name:        "valid VRF key",
			input:       []byte(`{"type":"VrfVerificationKey_PraosVRF","description":"VRF Verification Key","cborHex":"5840def456"}`),
			wantType:    "VrfVerificationKey_PraosVRF",
			wantCborHex: "5840def456",
			wantErr:     false,
		},
		{
			name:    "invalid JSON",
			input:   []byte(`{invalid json`),
			wantErr: true,
		},
		{
			name:        "empty envelope",
			input:       []byte(`{}`),
			wantType:    "",
			wantCborHex: "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseKeyEnvelope(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseKeyEnvelope() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Type != tt.wantType {
					t.Errorf("ParseKeyEnvelope() Type = %v, want %v", got.Type, tt.wantType)
				}
				if got.CborHex != tt.wantCborHex {
					t.Errorf("ParseKeyEnvelope() CborHex = %v, want %v", got.CborHex, tt.wantCborHex)
				}
			}
		})
	}
}

// TestSerializeKeyEnvelope tests serializing key envelopes to JSON
func TestSerializeKeyEnvelope(t *testing.T) {
	envelope := &KeyEnvelope{
		Type:        "StakePoolSigningKey_ed25519",
		Description: "Test Key",
		CborHex:     "5820abc123",
	}

	data, err := SerializeKeyEnvelope(envelope)
	if err != nil {
		t.Fatalf("SerializeKeyEnvelope() error = %v", err)
	}

	// Parse it back to verify
	var parsed KeyEnvelope
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to parse serialized envelope: %v", err)
	}

	if parsed.Type != envelope.Type {
		t.Errorf("Type mismatch: got %v, want %v", parsed.Type, envelope.Type)
	}
	if parsed.Description != envelope.Description {
		t.Errorf("Description mismatch: got %v, want %v", parsed.Description, envelope.Description)
	}
	if parsed.CborHex != envelope.CborHex {
		t.Errorf("CborHex mismatch: got %v, want %v", parsed.CborHex, envelope.CborHex)
	}
}

// TestValidateKeyType tests key type validation
func TestValidateKeyType(t *testing.T) {
	tests := []struct {
		name         string
		data         []byte
		expectedType string
		wantErr      bool
	}{
		{
			name:         "matching cold signing key",
			data:         []byte(`{"type":"StakePoolSigningKey_ed25519","description":"","cborHex":"abc"}`),
			expectedType: KeyTypeColdSigning,
			wantErr:      false,
		},
		{
			name:         "matching VRF verification key",
			data:         []byte(`{"type":"VrfVerificationKey_PraosVRF","description":"","cborHex":"abc"}`),
			expectedType: KeyTypeVRFVerification,
			wantErr:      false,
		},
		{
			name:         "mismatched key type",
			data:         []byte(`{"type":"StakePoolSigningKey_ed25519","description":"","cborHex":"abc"}`),
			expectedType: KeyTypeVRFSigning,
			wantErr:      true,
		},
		{
			name:         "invalid JSON",
			data:         []byte(`{invalid`),
			expectedType: KeyTypeColdSigning,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKeyType(tt.data, tt.expectedType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateKeyType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestExtractVRFKeyHash tests VRF key hash extraction
func TestExtractVRFKeyHash(t *testing.T) {
	tests := []struct {
		name    string
		vrfVKey []byte
		want    string
		wantErr bool
	}{
		{
			name:    "valid VRF key",
			vrfVKey: []byte(`{"type":"VrfVerificationKey_PraosVRF","description":"","cborHex":"5840abc123def456"}`),
			want:    "5840abc123def456",
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			vrfVKey: []byte(`{invalid}`),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractVRFKeyHash(tt.vrfVKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractVRFKeyHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractVRFKeyHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestKeyManagerGeneratePoolKeys tests pool key generation
func TestKeyManagerGeneratePoolKeys(t *testing.T) {
	ctx := context.Background()

	t.Run("successful key generation", func(t *testing.T) {
		client := &mockClient{}
		km := NewKeyManager(client)

		keys, err := km.GeneratePoolKeys(ctx)
		if err != nil {
			t.Fatalf("GeneratePoolKeys() error = %v", err)
		}

		if keys.Cold == nil {
			t.Error("Cold keys should not be nil")
		}
		if keys.VRF == nil {
			t.Error("VRF keys should not be nil")
		}
		if keys.KES == nil {
			t.Error("KES keys should not be nil")
		}
	})

	t.Run("cold key generation failure", func(t *testing.T) {
		client := &mockClient{
			generateColdKeysFunc: func(ctx context.Context) (*KeyPair, error) {
				return nil, errors.New("cold key generation failed")
			},
		}
		km := NewKeyManager(client)

		_, err := km.GeneratePoolKeys(ctx)
		if err == nil {
			t.Error("Expected error when cold key generation fails")
		}
	})

	t.Run("VRF key generation failure", func(t *testing.T) {
		client := &mockClient{
			generateVRFKeysFunc: func(ctx context.Context) (*KeyPair, error) {
				return nil, errors.New("VRF key generation failed")
			},
		}
		km := NewKeyManager(client)

		_, err := km.GeneratePoolKeys(ctx)
		if err == nil {
			t.Error("Expected error when VRF key generation fails")
		}
	})

	t.Run("KES key generation failure", func(t *testing.T) {
		client := &mockClient{
			generateKESKeysFunc: func(ctx context.Context) (*KeyPair, error) {
				return nil, errors.New("KES key generation failed")
			},
		}
		km := NewKeyManager(client)

		_, err := km.GeneratePoolKeys(ctx)
		if err == nil {
			t.Error("Expected error when KES key generation fails")
		}
	})
}

// TestKeyManagerGenerateWalletKeys tests wallet key generation
func TestKeyManagerGenerateWalletKeys(t *testing.T) {
	ctx := context.Background()

	t.Run("successful key generation", func(t *testing.T) {
		client := &mockClient{}
		km := NewKeyManager(client)

		keys, err := km.GenerateWalletKeys(ctx)
		if err != nil {
			t.Fatalf("GenerateWalletKeys() error = %v", err)
		}

		if keys.Payment == nil {
			t.Error("Payment keys should not be nil")
		}
		if keys.Stake == nil {
			t.Error("Stake keys should not be nil")
		}
	})

	t.Run("payment key generation failure", func(t *testing.T) {
		client := &mockClient{
			generatePaymentKeysFunc: func(ctx context.Context) (*KeyPair, error) {
				return nil, errors.New("payment key generation failed")
			},
		}
		km := NewKeyManager(client)

		_, err := km.GenerateWalletKeys(ctx)
		if err == nil {
			t.Error("Expected error when payment key generation fails")
		}
	})

	t.Run("stake key generation failure", func(t *testing.T) {
		client := &mockClient{
			generateStakeKeysFunc: func(ctx context.Context) (*KeyPair, error) {
				return nil, errors.New("stake key generation failed")
			},
		}
		km := NewKeyManager(client)

		_, err := km.GenerateWalletKeys(ctx)
		if err == nil {
			t.Error("Expected error when stake key generation fails")
		}
	})
}

// TestKeyManagerRotateKESKeys tests KES key rotation
func TestKeyManagerRotateKESKeys(t *testing.T) {
	ctx := context.Background()

	t.Run("successful KES rotation", func(t *testing.T) {
		client := &mockClient{}
		km := NewKeyManager(client)

		result, err := km.RotateKESKeys(ctx, []byte("cold-skey"), 5)
		if err != nil {
			t.Fatalf("RotateKESKeys() error = %v", err)
		}

		if result.NewKESKeys == nil {
			t.Error("NewKESKeys should not be nil")
		}
		if result.OperationalCertificate == nil {
			t.Error("OperationalCertificate should not be nil")
		}
		if result.NewCounter != 6 {
			t.Errorf("NewCounter = %v, want 6", result.NewCounter)
		}
		if result.KESPeriod != 100 {
			t.Errorf("KESPeriod = %v, want 100", result.KESPeriod)
		}
	})

	t.Run("KES key generation failure", func(t *testing.T) {
		client := &mockClient{
			generateKESKeysFunc: func(ctx context.Context) (*KeyPair, error) {
				return nil, errors.New("KES key generation failed")
			},
		}
		km := NewKeyManager(client)

		_, err := km.RotateKESKeys(ctx, []byte("cold-skey"), 5)
		if err == nil {
			t.Error("Expected error when KES key generation fails")
		}
	})

	t.Run("get KES period failure", func(t *testing.T) {
		client := &mockClient{
			getCurrentKESPeriodFunc: func(ctx context.Context) (int64, error) {
				return 0, errors.New("failed to get KES period")
			},
		}
		km := NewKeyManager(client)

		_, err := km.RotateKESKeys(ctx, []byte("cold-skey"), 5)
		if err == nil {
			t.Error("Expected error when getting KES period fails")
		}
	})

	t.Run("operational certificate creation failure", func(t *testing.T) {
		client := &mockClient{
			createOpCertFunc: func(ctx context.Context, kesSKey []byte, coldSKey []byte, counter int64, kesPeriod int64) (*OperationalCertificate, error) {
				return nil, errors.New("op cert creation failed")
			},
		}
		km := NewKeyManager(client)

		_, err := km.RotateKESKeys(ctx, []byte("cold-skey"), 5)
		if err == nil {
			t.Error("Expected error when operational certificate creation fails")
		}
	})
}

// TestKeyTypeConstants verifies key type constants
func TestKeyTypeConstants(t *testing.T) {
	// Test that constants are defined and have expected values
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"KeyTypeColdSigning", KeyTypeColdSigning, "StakePoolSigningKey_ed25519"},
		{"KeyTypeColdVerification", KeyTypeColdVerification, "StakePoolVerificationKey_ed25519"},
		{"KeyTypeVRFSigning", KeyTypeVRFSigning, "VrfSigningKey_PraosVRF"},
		{"KeyTypeVRFVerification", KeyTypeVRFVerification, "VrfVerificationKey_PraosVRF"},
		{"KeyTypeKESSigning", KeyTypeKESSigning, "KesSigningKey_ed25519_kes_2^6"},
		{"KeyTypeKESVerification", KeyTypeKESVerification, "KesVerificationKey_ed25519_kes_2^6"},
		{"KeyTypePaymentSigning", KeyTypePaymentSigning, "PaymentSigningKeyShelley_ed25519"},
		{"KeyTypePaymentVerification", KeyTypePaymentVerification, "PaymentVerificationKeyShelley_ed25519"},
		{"KeyTypeStakeSigning", KeyTypeStakeSigning, "StakeSigningKeyShelley_ed25519"},
		{"KeyTypeStakeVerification", KeyTypeStakeVerification, "StakeVerificationKeyShelley_ed25519"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.constant, tt.expected)
			}
		})
	}
}
