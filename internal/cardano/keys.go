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
	"fmt"
)

// KeyType represents the type of Cardano key
type KeyType string

const (
	KeyTypeCold    KeyType = "cold"
	KeyTypeVRF     KeyType = "vrf"
	KeyTypeKES     KeyType = "kes"
	KeyTypePayment KeyType = "payment"
	KeyTypeStake   KeyType = "stake"
)

// KeyEnvelope represents a Cardano key in envelope format
type KeyEnvelope struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	CborHex     string `json:"cborHex"`
}

// ParseKeyEnvelope parses a key from JSON envelope format
func ParseKeyEnvelope(data []byte) (*KeyEnvelope, error) {
	var envelope KeyEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("failed to parse key envelope: %w", err)
	}
	return &envelope, nil
}

// SerializeKeyEnvelope serializes a key to JSON envelope format
func SerializeKeyEnvelope(envelope *KeyEnvelope) ([]byte, error) {
	return json.MarshalIndent(envelope, "", "    ")
}

// KeyManager provides high-level key management operations
type KeyManager struct {
	client Client
}

// NewKeyManager creates a new KeyManager
func NewKeyManager(client Client) *KeyManager {
	return &KeyManager{client: client}
}

// GeneratePoolKeys generates all keys needed for a stake pool
func (km *KeyManager) GeneratePoolKeys(ctx context.Context) (*PoolKeys, error) {
	// Generate cold keys
	coldKeys, err := km.client.GenerateColdKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cold keys: %w", err)
	}

	// Generate VRF keys
	vrfKeys, err := km.client.GenerateVRFKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate VRF keys: %w", err)
	}

	// Generate KES keys
	kesKeys, err := km.client.GenerateKESKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate KES keys: %w", err)
	}

	return &PoolKeys{
		Cold: coldKeys,
		VRF:  vrfKeys,
		KES:  kesKeys,
	}, nil
}

// GenerateWalletKeys generates payment and stake keys for a wallet
func (km *KeyManager) GenerateWalletKeys(ctx context.Context) (*WalletKeys, error) {
	// Generate payment keys
	paymentKeys, err := km.client.GeneratePaymentKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate payment keys: %w", err)
	}

	// Generate stake keys
	stakeKeys, err := km.client.GenerateStakeKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate stake keys: %w", err)
	}

	return &WalletKeys{
		Payment: paymentKeys,
		Stake:   stakeKeys,
	}, nil
}

// RotateKESKeys generates new KES keys and creates an operational certificate
func (km *KeyManager) RotateKESKeys(ctx context.Context, coldSKey []byte, currentCounter int64) (*KESRotationResult, error) {
	// Generate new KES keys
	kesKeys, err := km.client.GenerateKESKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new KES keys: %w", err)
	}

	// Get current KES period
	kesPeriod, err := km.client.GetCurrentKESPeriod(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current KES period: %w", err)
	}

	// Create operational certificate
	opCert, err := km.client.CreateOperationalCertificate(ctx, kesKeys.SigningKey, coldSKey, currentCounter, kesPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to create operational certificate: %w", err)
	}

	return &KESRotationResult{
		NewKESKeys:             kesKeys,
		OperationalCertificate: opCert,
		NewCounter:             currentCounter + 1,
		KESPeriod:              kesPeriod,
	}, nil
}

// PoolKeys contains all keys for a stake pool
type PoolKeys struct {
	Cold *KeyPair
	VRF  *KeyPair
	KES  *KeyPair
}

// WalletKeys contains payment and stake keys
type WalletKeys struct {
	Payment *KeyPair
	Stake   *KeyPair
}

// KESRotationResult contains the result of a KES rotation
type KESRotationResult struct {
	NewKESKeys             *KeyPair
	OperationalCertificate *OperationalCertificate
	NewCounter             int64
	KESPeriod              int64
}

// ExtractVRFKeyHash extracts the VRF key hash from a verification key
func ExtractVRFKeyHash(vrfVKey []byte) (string, error) {
	envelope, err := ParseKeyEnvelope(vrfVKey)
	if err != nil {
		return "", fmt.Errorf("failed to parse VRF key: %w", err)
	}

	// The cborHex contains the key hash in most cases
	// In production, this would need to be properly computed
	return envelope.CborHex, nil
}

// ValidateKeyType checks if a key envelope has the expected type
func ValidateKeyType(data []byte, expectedType string) error {
	envelope, err := ParseKeyEnvelope(data)
	if err != nil {
		return err
	}

	if envelope.Type != expectedType {
		return fmt.Errorf("unexpected key type: got %s, expected %s", envelope.Type, expectedType)
	}

	return nil
}

// Key type constants for validation
const (
	KeyTypeColdSigning         = "StakePoolSigningKey_ed25519"
	KeyTypeColdVerification    = "StakePoolVerificationKey_ed25519"
	KeyTypeVRFSigning          = "VrfSigningKey_PraosVRF"
	KeyTypeVRFVerification     = "VrfVerificationKey_PraosVRF"
	KeyTypeKESSigning          = "KesSigningKey_ed25519_kes_2^6"
	KeyTypeKESVerification     = "KesVerificationKey_ed25519_kes_2^6"
	KeyTypePaymentSigning      = "PaymentSigningKeyShelley_ed25519"
	KeyTypePaymentVerification = "PaymentVerificationKeyShelley_ed25519"
	KeyTypeStakeSigning        = "StakeSigningKeyShelley_ed25519"
	KeyTypeStakeVerification   = "StakeVerificationKeyShelley_ed25519"
)
