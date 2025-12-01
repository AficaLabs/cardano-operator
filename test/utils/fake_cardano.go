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

package utils

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/AficaLabs/cardano-operator/internal/cardano"
)

// CallRecord records a method call for assertions
type CallRecord struct {
	Method string
	Args   []interface{}
}

// FakeCardanoClient is a fake implementation of cardano.Client for testing
type FakeCardanoClient struct {
	mu sync.Mutex

	// Configurable responses
	TipResponse                *cardano.Tip
	UTxOResponse               []cardano.UTxO
	ProtocolParametersResponse *cardano.ProtocolParameters
	PoolParamsResponse         *cardano.PoolParams
	PoolIDResponse             string
	ColdKeysResponse           *cardano.KeyPair
	VRFKeysResponse            *cardano.KeyPair
	KESKeysResponse            *cardano.KeyPair
	PaymentKeysResponse        *cardano.KeyPair
	StakeKeysResponse          *cardano.KeyPair
	OpCertResponse             *cardano.OperationalCertificate
	RegistrationCertResponse   []byte
	UpdateCertResponse         []byte
	RetirementCertResponse     []byte
	TxBodyResponse             []byte
	SignedTxResponse           []byte
	SubmitTxResponse           *cardano.TransactionSubmitResult
	AddressResponse            string
	StakeAddressResponse       string
	MinFeeResponse             int64
	KESPeriodResponse          int64

	// Error injection
	QueryTipError                    error
	QueryUTxOError                   error
	QueryProtocolParametersError     error
	QueryPoolParamsError             error
	QueryStakePoolIDError            error
	GenerateColdKeysError            error
	GenerateVRFKeysError             error
	GenerateKESKeysError             error
	GeneratePaymentKeysError         error
	GenerateStakeKeysError           error
	CreateOpCertError                error
	CreateRegistrationCertError      error
	CreateUpdateCertError            error
	CreateRetirementCertError        error
	BuildTxError                     error
	SignTxError                      error
	SubmitTxError                    error
	BuildAddressError                error
	BuildStakeAddressError           error
	CalculateMinFeeError             error
	GetCurrentKESPeriodError         error

	// Call recording
	Calls []CallRecord
}

// NewFakeCardanoClient creates a new FakeCardanoClient with default responses
func NewFakeCardanoClient() *FakeCardanoClient {
	return &FakeCardanoClient{
		TipResponse: &cardano.Tip{
			Slot:      1000000,
			Epoch:     100,
			Block:     50000,
			Hash:      "abc123",
			SyncState: "100.00",
		},
		UTxOResponse: []cardano.UTxO{
			{
				TxHash:  "tx123",
				TxIndex: 0,
				Address: "addr_test1qz...",
				Value:   cardano.Value{Lovelace: 1000000000},
			},
		},
		ProtocolParametersResponse: &cardano.ProtocolParameters{
			MinFeeA:            44,
			MinFeeB:            155381,
			MaxBlockSize:       90112,
			MaxTxSize:          16384,
			KeyDeposit:         2000000,
			PoolDeposit:        500000000,
			MaxEpoch:           18,
			NOpt:               500,
			PoolPledgeInfluence: "0.3",
			MinPoolCost:        340000000,
			SlotsPerKESPeriod:  129600,
			MaxKESEvolutions:   62,
		},
		PoolParamsResponse: &cardano.PoolParams{
			PoolID:     "pool1abc...",
			VRFKeyHash: "vrf123...",
			Pledge:     500000000000,
			Cost:       340000000,
			Margin:     0.03,
		},
		PoolIDResponse: "pool1abc123...",
		ColdKeysResponse: &cardano.KeyPair{
			SigningKey:      []byte("cold-skey"),
			VerificationKey: []byte("cold-vkey"),
		},
		VRFKeysResponse: &cardano.KeyPair{
			SigningKey:      []byte("vrf-skey"),
			VerificationKey: []byte("vrf-vkey"),
		},
		KESKeysResponse: &cardano.KeyPair{
			SigningKey:      []byte("kes-skey"),
			VerificationKey: []byte("kes-vkey"),
		},
		PaymentKeysResponse: &cardano.KeyPair{
			SigningKey:      []byte("payment-skey"),
			VerificationKey: []byte("payment-vkey"),
		},
		StakeKeysResponse: &cardano.KeyPair{
			SigningKey:      []byte("stake-skey"),
			VerificationKey: []byte("stake-vkey"),
		},
		OpCertResponse: &cardano.OperationalCertificate{
			Certificate: []byte("op-cert"),
			Counter:     1,
			KESPeriod:   100,
			ExpiryEpoch: 162,
		},
		RegistrationCertResponse: []byte("registration-cert"),
		UpdateCertResponse:       []byte("update-cert"),
		RetirementCertResponse:   []byte("retirement-cert"),
		TxBodyResponse:           []byte("tx-body"),
		SignedTxResponse:         []byte("signed-tx"),
		SubmitTxResponse: &cardano.TransactionSubmitResult{
			TxHash:  "submitted-tx-hash",
			Success: true,
		},
		AddressResponse:      "addr_test1qz...",
		StakeAddressResponse: "stake_test1up...",
		MinFeeResponse:       200000,
		KESPeriodResponse:    100,
		Calls:                make([]CallRecord, 0),
	}
}

// recordCall records a method call
func (f *FakeCardanoClient) recordCall(method string, args ...interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, CallRecord{Method: method, Args: args})
}

// GetCalls returns a copy of all recorded calls
func (f *FakeCardanoClient) GetCalls() []CallRecord {
	f.mu.Lock()
	defer f.mu.Unlock()
	calls := make([]CallRecord, len(f.Calls))
	copy(calls, f.Calls)
	return calls
}

// ResetCalls clears all recorded calls
func (f *FakeCardanoClient) ResetCalls() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = make([]CallRecord, 0)
}

// CallCount returns the number of calls to a specific method
func (f *FakeCardanoClient) CallCount(method string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	count := 0
	for _, call := range f.Calls {
		if call.Method == method {
			count++
		}
	}
	return count
}

// WasMethodCalled checks if a method was called
func (f *FakeCardanoClient) WasMethodCalled(method string) bool {
	return f.CallCount(method) > 0
}

// Query operations

func (f *FakeCardanoClient) QueryTip(ctx context.Context) (*cardano.Tip, error) {
	f.recordCall("QueryTip")
	if f.QueryTipError != nil {
		return nil, f.QueryTipError
	}
	return f.TipResponse, nil
}

func (f *FakeCardanoClient) QueryUTxO(ctx context.Context, address string) ([]cardano.UTxO, error) {
	f.recordCall("QueryUTxO", address)
	if f.QueryUTxOError != nil {
		return nil, f.QueryUTxOError
	}
	return f.UTxOResponse, nil
}

func (f *FakeCardanoClient) QueryProtocolParameters(ctx context.Context) (*cardano.ProtocolParameters, error) {
	f.recordCall("QueryProtocolParameters")
	if f.QueryProtocolParametersError != nil {
		return nil, f.QueryProtocolParametersError
	}
	return f.ProtocolParametersResponse, nil
}

func (f *FakeCardanoClient) QueryPoolParams(ctx context.Context, poolID string) (*cardano.PoolParams, error) {
	f.recordCall("QueryPoolParams", poolID)
	if f.QueryPoolParamsError != nil {
		return nil, f.QueryPoolParamsError
	}
	return f.PoolParamsResponse, nil
}

func (f *FakeCardanoClient) QueryStakePoolID(ctx context.Context, coldVKeyPath string) (string, error) {
	f.recordCall("QueryStakePoolID", coldVKeyPath)
	if f.QueryStakePoolIDError != nil {
		return "", f.QueryStakePoolIDError
	}
	return f.PoolIDResponse, nil
}

// Key generation

func (f *FakeCardanoClient) GenerateColdKeys(ctx context.Context) (*cardano.KeyPair, error) {
	f.recordCall("GenerateColdKeys")
	if f.GenerateColdKeysError != nil {
		return nil, f.GenerateColdKeysError
	}
	return f.ColdKeysResponse, nil
}

func (f *FakeCardanoClient) GenerateVRFKeys(ctx context.Context) (*cardano.KeyPair, error) {
	f.recordCall("GenerateVRFKeys")
	if f.GenerateVRFKeysError != nil {
		return nil, f.GenerateVRFKeysError
	}
	return f.VRFKeysResponse, nil
}

func (f *FakeCardanoClient) GenerateKESKeys(ctx context.Context) (*cardano.KeyPair, error) {
	f.recordCall("GenerateKESKeys")
	if f.GenerateKESKeysError != nil {
		return nil, f.GenerateKESKeysError
	}
	return f.KESKeysResponse, nil
}

func (f *FakeCardanoClient) GeneratePaymentKeys(ctx context.Context) (*cardano.KeyPair, error) {
	f.recordCall("GeneratePaymentKeys")
	if f.GeneratePaymentKeysError != nil {
		return nil, f.GeneratePaymentKeysError
	}
	return f.PaymentKeysResponse, nil
}

func (f *FakeCardanoClient) GenerateStakeKeys(ctx context.Context) (*cardano.KeyPair, error) {
	f.recordCall("GenerateStakeKeys")
	if f.GenerateStakeKeysError != nil {
		return nil, f.GenerateStakeKeysError
	}
	return f.StakeKeysResponse, nil
}

// Certificate operations

func (f *FakeCardanoClient) CreateOperationalCertificate(ctx context.Context, kesSKey []byte, coldSKey []byte, counter int64, kesPeriod int64) (*cardano.OperationalCertificate, error) {
	f.recordCall("CreateOperationalCertificate", kesSKey, coldSKey, counter, kesPeriod)
	if f.CreateOpCertError != nil {
		return nil, f.CreateOpCertError
	}
	return f.OpCertResponse, nil
}

func (f *FakeCardanoClient) CreatePoolRegistrationCertificate(ctx context.Context, params *cardano.PoolParams, coldVKey []byte) ([]byte, error) {
	f.recordCall("CreatePoolRegistrationCertificate", params, coldVKey)
	if f.CreateRegistrationCertError != nil {
		return nil, f.CreateRegistrationCertError
	}
	return f.RegistrationCertResponse, nil
}

func (f *FakeCardanoClient) CreatePoolUpdateCertificate(ctx context.Context, params *cardano.PoolParams, coldVKey []byte) ([]byte, error) {
	f.recordCall("CreatePoolUpdateCertificate", params, coldVKey)
	if f.CreateUpdateCertError != nil {
		return nil, f.CreateUpdateCertError
	}
	return f.UpdateCertResponse, nil
}

func (f *FakeCardanoClient) CreatePoolRetirementCertificate(ctx context.Context, poolID string, retirementEpoch int64, coldVKey []byte) ([]byte, error) {
	f.recordCall("CreatePoolRetirementCertificate", poolID, retirementEpoch, coldVKey)
	if f.CreateRetirementCertError != nil {
		return nil, f.CreateRetirementCertError
	}
	return f.RetirementCertResponse, nil
}

// Transaction operations

func (f *FakeCardanoClient) BuildTx(ctx context.Context, opts *cardano.TxBuildOptions) ([]byte, error) {
	f.recordCall("BuildTx", opts)
	if f.BuildTxError != nil {
		return nil, f.BuildTxError
	}
	return f.TxBodyResponse, nil
}

func (f *FakeCardanoClient) SignTx(ctx context.Context, txBody []byte, signingKeys ...[]byte) ([]byte, error) {
	f.recordCall("SignTx", txBody, signingKeys)
	if f.SignTxError != nil {
		return nil, f.SignTxError
	}
	return f.SignedTxResponse, nil
}

func (f *FakeCardanoClient) SubmitTx(ctx context.Context, signedTx []byte) (*cardano.TransactionSubmitResult, error) {
	f.recordCall("SubmitTx", signedTx)
	if f.SubmitTxError != nil {
		return nil, f.SubmitTxError
	}
	return f.SubmitTxResponse, nil
}

// Address operations

func (f *FakeCardanoClient) BuildAddress(ctx context.Context, paymentVKey []byte, stakeVKey []byte) (string, error) {
	f.recordCall("BuildAddress", paymentVKey, stakeVKey)
	if f.BuildAddressError != nil {
		return "", f.BuildAddressError
	}
	return f.AddressResponse, nil
}

func (f *FakeCardanoClient) BuildStakeAddress(ctx context.Context, stakeVKey []byte) (string, error) {
	f.recordCall("BuildStakeAddress", stakeVKey)
	if f.BuildStakeAddressError != nil {
		return "", f.BuildStakeAddressError
	}
	return f.StakeAddressResponse, nil
}

// Utility

func (f *FakeCardanoClient) CalculateMinFee(ctx context.Context, txBody []byte, witnessCount int) (int64, error) {
	f.recordCall("CalculateMinFee", txBody, witnessCount)
	if f.CalculateMinFeeError != nil {
		return 0, f.CalculateMinFeeError
	}
	return f.MinFeeResponse, nil
}

func (f *FakeCardanoClient) GetCurrentKESPeriod(ctx context.Context) (int64, error) {
	f.recordCall("GetCurrentKESPeriod")
	if f.GetCurrentKESPeriodError != nil {
		return 0, f.GetCurrentKESPeriodError
	}
	return f.KESPeriodResponse, nil
}

// Verify FakeCardanoClient implements cardano.Client
var _ cardano.Client = &FakeCardanoClient{}

// Helper functions for creating test data

// GenerateTestPoolID generates a valid-looking pool ID for testing
func GenerateTestPoolID() string {
	return "pool1" + hex.EncodeToString([]byte("testpool12345678901234567"))[:52]
}

// GenerateTestTxHash generates a valid-looking transaction hash for testing
func GenerateTestTxHash() string {
	return hex.EncodeToString([]byte("testtxhash123456789012345678901"))[:64]
}

// GenerateTestAddress generates a valid-looking address for testing
func GenerateTestAddress(network string) string {
	if network == "mainnet" {
		return "addr1qx2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3jcu5d8ps7zex2k2xt3uqxgjqnnj83ws8lhrn648jjxtwq2ytjqp"
	}
	return "addr_test1qz2fxv2umyhttkxyxp8x0dlpdt3k6cwng5pxj3jhsydzer3jcu5d8ps7zex2k2xt3uqxgjqnnj83ws8lhrn648jjxtwqhpj9f2"
}

// WithQueryTipResponse sets the QueryTip response
func (f *FakeCardanoClient) WithQueryTipResponse(tip *cardano.Tip) *FakeCardanoClient {
	f.TipResponse = tip
	return f
}

// WithQueryTipError sets an error for QueryTip
func (f *FakeCardanoClient) WithQueryTipError(err error) *FakeCardanoClient {
	f.QueryTipError = err
	return f
}

// WithUTxOResponse sets the UTxO response
func (f *FakeCardanoClient) WithUTxOResponse(utxos []cardano.UTxO) *FakeCardanoClient {
	f.UTxOResponse = utxos
	return f
}

// WithUTxOError sets an error for QueryUTxO
func (f *FakeCardanoClient) WithUTxOError(err error) *FakeCardanoClient {
	f.QueryUTxOError = err
	return f
}

// WithSubmitTxError sets an error for SubmitTx
func (f *FakeCardanoClient) WithSubmitTxError(err error) *FakeCardanoClient {
	f.SubmitTxError = err
	return f
}

// WithSubmitTxFailure configures SubmitTx to return a failure result
func (f *FakeCardanoClient) WithSubmitTxFailure(errorMsg string) *FakeCardanoClient {
	f.SubmitTxResponse = &cardano.TransactionSubmitResult{
		TxHash:  "",
		Success: false,
		Error:   errorMsg,
	}
	return f
}

// SimulateInsufficientFunds configures the fake client to simulate insufficient funds
func (f *FakeCardanoClient) SimulateInsufficientFunds() *FakeCardanoClient {
	f.UTxOResponse = []cardano.UTxO{
		{
			TxHash:  "tx123",
			TxIndex: 0,
			Address: "addr_test1qz...",
			Value:   cardano.Value{Lovelace: 1000000}, // Only 1 ADA
		},
	}
	return f
}

// SimulateNodeNotSynced configures the fake client to simulate a non-synced node
func (f *FakeCardanoClient) SimulateNodeNotSynced() *FakeCardanoClient {
	f.TipResponse = &cardano.Tip{
		Slot:      500000,
		Epoch:     50,
		Block:     25000,
		Hash:      "abc123",
		SyncState: "50.00", // 50% synced
	}
	return f
}

// SimulateKESExpiringSoon configures the fake client to simulate KES keys expiring soon
func (f *FakeCardanoClient) SimulateKESExpiringSoon(currentEpoch int64, expiryEpoch int64) *FakeCardanoClient {
	f.TipResponse = &cardano.Tip{
		Slot:      currentEpoch * 432000, // Approximate slot for epoch
		Epoch:     currentEpoch,
		Block:     int64(currentEpoch * 21600),
		Hash:      "abc123",
		SyncState: "100.00",
	}
	f.OpCertResponse = &cardano.OperationalCertificate{
		Certificate: []byte("op-cert"),
		Counter:     1,
		KESPeriod:   currentEpoch * 3, // Approximate KES period
		ExpiryEpoch: expiryEpoch,
	}
	return f
}

// SimulateNetworkError configures the fake client to return network errors
func (f *FakeCardanoClient) SimulateNetworkError() *FakeCardanoClient {
	networkErr := fmt.Errorf("network error: connection refused")
	f.QueryTipError = networkErr
	f.QueryUTxOError = networkErr
	f.QueryProtocolParametersError = networkErr
	f.SubmitTxError = networkErr
	return f
}
