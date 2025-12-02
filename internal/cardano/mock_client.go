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

// MockClient is a mock implementation of the Client interface for testing
type MockClient struct {
	// Query operation results
	TipResult            *Tip
	TipError             error
	UTxOResult           []UTxO
	UTxOError            error
	ProtocolParamsResult *ProtocolParameters
	ProtocolParamsError  error
	PoolParamsResult     *PoolParams
	PoolParamsError      error
	StakePoolIDResult    string
	StakePoolIDError     error

	// Key generation results
	ColdKeysResult    *KeyPair
	ColdKeysError     error
	VRFKeysResult     *KeyPair
	VRFKeysError      error
	KESKeysResult     *KeyPair
	KESKeysError      error
	PaymentKeysResult *KeyPair
	PaymentKeysError  error
	StakeKeysResult   *KeyPair
	StakeKeysError    error

	// Certificate operation results
	OpCertResult         *OperationalCertificate
	OpCertError          error
	PoolRegCertResult    []byte
	PoolRegCertError     error
	PoolUpdateCertResult []byte
	PoolUpdateCertError  error
	PoolRetireCertResult []byte
	PoolRetireCertError  error

	// Transaction operation results
	BuildTxResult  []byte
	BuildTxError   error
	SignTxResult   []byte
	SignTxError    error
	SubmitTxResult *TransactionSubmitResult
	SubmitTxError  error

	// Address operation results
	BuildAddressResult      string
	BuildAddressError       error
	BuildStakeAddressResult string
	BuildStakeAddressError  error

	// Utility results
	MinFeeResult    int64
	MinFeeError     error
	KESPeriodResult int64
	KESPeriodError  error

	// Call tracking
	QueryTipCalled                int
	QueryUTxOCalled               int
	QueryProtocolParametersCalled int
	QueryPoolParamsCalled         int
	QueryStakePoolIDCalled        int
	GenerateColdKeysCalled        int
	GenerateVRFKeysCalled         int
	GenerateKESKeysCalled         int
	GeneratePaymentKeysCalled     int
	GenerateStakeKeysCalled       int
	CreateOpCertCalled            int
	CreatePoolRegCertCalled       int
	CreatePoolUpdateCertCalled    int
	CreatePoolRetireCertCalled    int
	BuildTxCalled                 int
	SignTxCalled                  int
	SubmitTxCalled                int
	BuildAddressCalled            int
	BuildStakeAddressCalled       int
	CalculateMinFeeCalled         int
	GetCurrentKESPeriodCalled     int
}

// NewMockClient creates a new mock client with default successful responses
func NewMockClient() *MockClient {
	return &MockClient{
		TipResult: &Tip{
			Slot:      1000000,
			Epoch:     100,
			Block:     500000,
			Hash:      "abc123def456",
			SyncState: "100.00",
		},
		UTxOResult: []UTxO{
			{
				TxHash:  "tx123",
				TxIndex: 0,
				Address: "addr_test1qz",
				Value:   Value{Lovelace: 10000000000},
			},
		},
		ProtocolParamsResult: &ProtocolParameters{
			MinFeeA:             44,
			MinFeeB:             155381,
			MaxBlockSize:        90112,
			MaxTxSize:           16384,
			KeyDeposit:          2000000,
			PoolDeposit:         500000000,
			MaxEpoch:            18,
			NOpt:                500,
			PoolPledgeInfluence: "0.3",
			MinPoolCost:         340000000,
			SlotsPerKESPeriod:   129600,
			MaxKESEvolutions:    62,
		},
		PoolParamsResult: &PoolParams{
			PoolID:        "pool1abc123",
			VRFKeyHash:    "vrf123",
			Pledge:        500000000000,
			Cost:          340000000,
			Margin:        0.03,
			RewardAccount: "stake1abc",
			Owners:        []string{"stake1abc"},
			Relays:        []Relay{{Type: "dns", Hostname: "relay.example.com", Port: 6000}},
		},
		StakePoolIDResult: "pool1abc123def456",
		ColdKeysResult:    &KeyPair{SigningKey: []byte("cold-skey"), VerificationKey: []byte("cold-vkey")},
		VRFKeysResult:     &KeyPair{SigningKey: []byte("vrf-skey"), VerificationKey: []byte("vrf-vkey")},
		KESKeysResult:     &KeyPair{SigningKey: []byte("kes-skey"), VerificationKey: []byte("kes-vkey")},
		PaymentKeysResult: &KeyPair{SigningKey: []byte("pay-skey"), VerificationKey: []byte("pay-vkey")},
		StakeKeysResult:   &KeyPair{SigningKey: []byte("stake-skey"), VerificationKey: []byte("stake-vkey")},
		OpCertResult: &OperationalCertificate{
			Certificate: []byte("op-cert"),
			Counter:     5,
			KESPeriod:   100,
			ExpiryEpoch: 200,
		},
		PoolRegCertResult:       []byte("pool-reg-cert"),
		PoolUpdateCertResult:    []byte("pool-update-cert"),
		PoolRetireCertResult:    []byte("pool-retire-cert"),
		BuildTxResult:           []byte("tx-body"),
		SignTxResult:            []byte("signed-tx"),
		SubmitTxResult:          &TransactionSubmitResult{TxHash: "tx-hash-abc", Success: true},
		BuildAddressResult:      "addr_test1qz123",
		BuildStakeAddressResult: "stake_test1uz123",
		MinFeeResult:            200000,
		KESPeriodResult:         100,
	}
}

// Query operations
func (m *MockClient) QueryTip(ctx context.Context) (*Tip, error) {
	m.QueryTipCalled++
	return m.TipResult, m.TipError
}

func (m *MockClient) QueryUTxO(ctx context.Context, address string) ([]UTxO, error) {
	m.QueryUTxOCalled++
	return m.UTxOResult, m.UTxOError
}

func (m *MockClient) QueryProtocolParameters(ctx context.Context) (*ProtocolParameters, error) {
	m.QueryProtocolParametersCalled++
	return m.ProtocolParamsResult, m.ProtocolParamsError
}

func (m *MockClient) QueryPoolParams(ctx context.Context, poolID string) (*PoolParams, error) {
	m.QueryPoolParamsCalled++
	return m.PoolParamsResult, m.PoolParamsError
}

func (m *MockClient) QueryStakePoolID(ctx context.Context, coldVKeyPath string) (string, error) {
	m.QueryStakePoolIDCalled++
	return m.StakePoolIDResult, m.StakePoolIDError
}

// Key generation
func (m *MockClient) GenerateColdKeys(ctx context.Context) (*KeyPair, error) {
	m.GenerateColdKeysCalled++
	return m.ColdKeysResult, m.ColdKeysError
}

func (m *MockClient) GenerateVRFKeys(ctx context.Context) (*KeyPair, error) {
	m.GenerateVRFKeysCalled++
	return m.VRFKeysResult, m.VRFKeysError
}

func (m *MockClient) GenerateKESKeys(ctx context.Context) (*KeyPair, error) {
	m.GenerateKESKeysCalled++
	return m.KESKeysResult, m.KESKeysError
}

func (m *MockClient) GeneratePaymentKeys(ctx context.Context) (*KeyPair, error) {
	m.GeneratePaymentKeysCalled++
	return m.PaymentKeysResult, m.PaymentKeysError
}

func (m *MockClient) GenerateStakeKeys(ctx context.Context) (*KeyPair, error) {
	m.GenerateStakeKeysCalled++
	return m.StakeKeysResult, m.StakeKeysError
}

// Certificate operations
func (m *MockClient) CreateOperationalCertificate(ctx context.Context, kesSKey []byte, coldSKey []byte, counter int64, kesPeriod int64) (*OperationalCertificate, error) {
	m.CreateOpCertCalled++
	return m.OpCertResult, m.OpCertError
}

func (m *MockClient) CreatePoolRegistrationCertificate(ctx context.Context, params *PoolParams, coldVKey []byte) ([]byte, error) {
	m.CreatePoolRegCertCalled++
	return m.PoolRegCertResult, m.PoolRegCertError
}

func (m *MockClient) CreatePoolUpdateCertificate(ctx context.Context, params *PoolParams, coldVKey []byte) ([]byte, error) {
	m.CreatePoolUpdateCertCalled++
	return m.PoolUpdateCertResult, m.PoolUpdateCertError
}

func (m *MockClient) CreatePoolRetirementCertificate(ctx context.Context, poolID string, retirementEpoch int64, coldVKey []byte) ([]byte, error) {
	m.CreatePoolRetireCertCalled++
	return m.PoolRetireCertResult, m.PoolRetireCertError
}

// Transaction operations
func (m *MockClient) BuildTx(ctx context.Context, opts *TxBuildOptions) ([]byte, error) {
	m.BuildTxCalled++
	return m.BuildTxResult, m.BuildTxError
}

func (m *MockClient) SignTx(ctx context.Context, txBody []byte, signingKeys ...[]byte) ([]byte, error) {
	m.SignTxCalled++
	return m.SignTxResult, m.SignTxError
}

func (m *MockClient) SubmitTx(ctx context.Context, signedTx []byte) (*TransactionSubmitResult, error) {
	m.SubmitTxCalled++
	return m.SubmitTxResult, m.SubmitTxError
}

// Address operations
func (m *MockClient) BuildAddress(ctx context.Context, paymentVKey []byte, stakeVKey []byte) (string, error) {
	m.BuildAddressCalled++
	return m.BuildAddressResult, m.BuildAddressError
}

func (m *MockClient) BuildStakeAddress(ctx context.Context, stakeVKey []byte) (string, error) {
	m.BuildStakeAddressCalled++
	return m.BuildStakeAddressResult, m.BuildStakeAddressError
}

// Utility
func (m *MockClient) CalculateMinFee(ctx context.Context, txBody []byte, witnessCount int) (int64, error) {
	m.CalculateMinFeeCalled++
	return m.MinFeeResult, m.MinFeeError
}

func (m *MockClient) GetCurrentKESPeriod(ctx context.Context) (int64, error) {
	m.GetCurrentKESPeriodCalled++
	return m.KESPeriodResult, m.KESPeriodError
}

// Verify MockClient implements Client interface
var _ Client = (*MockClient)(nil)
