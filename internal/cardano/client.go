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
)

// Network represents the Cardano network
type Network string

const (
	NetworkMainnet Network = "mainnet"
	NetworkPreprod Network = "preprod"
	NetworkPreview Network = "preview"
)

// Tip represents the current chain tip
type Tip struct {
	Slot      int64  `json:"slot"`
	Epoch     int64  `json:"epoch"`
	Block     int64  `json:"block"`
	Hash      string `json:"hash"`
	SyncState string `json:"syncState"`
}

// UTxO represents an unspent transaction output
type UTxO struct {
	TxHash  string `json:"txHash"`
	TxIndex int    `json:"txIndex"`
	Address string `json:"address"`
	Value   Value  `json:"value"`
}

// Value represents the value in a UTxO
type Value struct {
	Lovelace int64             `json:"lovelace"`
	Assets   map[string]int64  `json:"assets,omitempty"`
}

// ProtocolParameters represents the current protocol parameters
type ProtocolParameters struct {
	MinFeeA            int64  `json:"txFeePerByte"`
	MinFeeB            int64  `json:"txFeeFixed"`
	MaxBlockSize       int64  `json:"maxBlockBodySize"`
	MaxTxSize          int64  `json:"maxTxSize"`
	KeyDeposit         int64  `json:"stakeAddressDeposit"`
	PoolDeposit        int64  `json:"stakePoolDeposit"`
	MaxEpoch           int64  `json:"poolRetireMaxEpoch"`
	NOpt               int64  `json:"stakePoolTargetNum"`
	PoolPledgeInfluence string `json:"poolPledgeInfluence"`
	MinPoolCost        int64  `json:"minPoolCost"`
	SlotsPerKESPeriod  int64  `json:"slotsPerKESPeriod"`
	MaxKESEvolutions   int64  `json:"maxKESEvolutions"`
}

// PoolParams represents stake pool parameters for registration
type PoolParams struct {
	PoolID        string   `json:"poolId,omitempty"`
	VRFKeyHash    string   `json:"vrfKeyHash"`
	Pledge        int64    `json:"pledge"`
	Cost          int64    `json:"cost"`
	Margin        float64  `json:"margin"`
	RewardAccount string   `json:"rewardAccount"`
	Owners        []string `json:"owners"`
	Relays        []Relay  `json:"relays"`
	MetadataURL   string   `json:"metadataUrl,omitempty"`
	MetadataHash  string   `json:"metadataHash,omitempty"`
}

// Relay represents a pool relay configuration
type Relay struct {
	Type     string `json:"type"` // "dns" or "ip"
	Hostname string `json:"hostname,omitempty"`
	IPv4     string `json:"ipv4,omitempty"`
	IPv6     string `json:"ipv6,omitempty"`
	Port     int32  `json:"port"`
}

// KeyPair represents a cryptographic key pair
type KeyPair struct {
	SigningKey      []byte `json:"signingKey"`
	VerificationKey []byte `json:"verificationKey"`
}

// OperationalCertificate represents a KES operational certificate
type OperationalCertificate struct {
	Certificate    []byte `json:"certificate"`
	Counter        int64  `json:"counter"`
	KESPeriod      int64  `json:"kesPeriod"`
	ExpiryEpoch    int64  `json:"expiryEpoch"`
}

// TransactionSubmitResult represents the result of submitting a transaction
type TransactionSubmitResult struct {
	TxHash  string `json:"txHash"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// Client defines the interface for interacting with Cardano
type Client interface {
	// Query operations
	QueryTip(ctx context.Context) (*Tip, error)
	QueryUTxO(ctx context.Context, address string) ([]UTxO, error)
	QueryProtocolParameters(ctx context.Context) (*ProtocolParameters, error)
	QueryPoolParams(ctx context.Context, poolID string) (*PoolParams, error)
	QueryStakePoolID(ctx context.Context, coldVKeyPath string) (string, error)

	// Key generation
	GenerateColdKeys(ctx context.Context) (*KeyPair, error)
	GenerateVRFKeys(ctx context.Context) (*KeyPair, error)
	GenerateKESKeys(ctx context.Context) (*KeyPair, error)
	GeneratePaymentKeys(ctx context.Context) (*KeyPair, error)
	GenerateStakeKeys(ctx context.Context) (*KeyPair, error)

	// Certificate operations
	CreateOperationalCertificate(ctx context.Context, kesSKey []byte, coldSKey []byte, counter int64, kesPeriod int64) (*OperationalCertificate, error)
	CreatePoolRegistrationCertificate(ctx context.Context, params *PoolParams, coldVKey []byte) ([]byte, error)
	CreatePoolUpdateCertificate(ctx context.Context, params *PoolParams, coldVKey []byte) ([]byte, error)
	CreatePoolRetirementCertificate(ctx context.Context, poolID string, retirementEpoch int64, coldVKey []byte) ([]byte, error)

	// Transaction operations
	BuildTx(ctx context.Context, opts *TxBuildOptions) ([]byte, error)
	SignTx(ctx context.Context, txBody []byte, signingKeys ...[]byte) ([]byte, error)
	SubmitTx(ctx context.Context, signedTx []byte) (*TransactionSubmitResult, error)

	// Address operations
	BuildAddress(ctx context.Context, paymentVKey []byte, stakeVKey []byte) (string, error)
	BuildStakeAddress(ctx context.Context, stakeVKey []byte) (string, error)

	// Utility
	CalculateMinFee(ctx context.Context, txBody []byte, witnessCount int) (int64, error)
	GetCurrentKESPeriod(ctx context.Context) (int64, error)
}

// TxBuildOptions contains options for building a transaction
type TxBuildOptions struct {
	// Inputs to spend
	Inputs []UTxO

	// Outputs to create
	Outputs []TxOutput

	// Certificates to include
	Certificates [][]byte

	// Change address
	ChangeAddress string

	// TTL (time to live) in slots
	TTLSlot int64

	// Fee (if 0, will be calculated)
	Fee int64

	// Metadata
	Metadata map[string]interface{}
}

// TxOutput represents a transaction output
type TxOutput struct {
	Address string
	Value   Value
}

// ClientConfig holds configuration for the Cardano client
type ClientConfig struct {
	// Network to connect to
	Network Network

	// Socket path for cardano-node
	SocketPath string

	// Path to cardano-cli binary (if not using container)
	CLIPath string

	// Timeout for operations
	TimeoutSeconds int

	// Magic number for the network (auto-detected if not set)
	NetworkMagic int64
}

// NewClient creates a new Cardano client
func NewClient(config *ClientConfig) (Client, error) {
	return NewExecutor(config)
}
