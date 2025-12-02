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
	"errors"
	"testing"
)

// txMockClient is a mock client for transaction tests with configurable behavior
type txMockClient struct {
	mockClient
	queryUTxOFunc         func(ctx context.Context, address string) ([]UTxO, error)
	createPoolRegCertFunc func(ctx context.Context, params *PoolParams, coldVKey []byte) ([]byte, error)
	createPoolUpdCertFunc func(ctx context.Context, params *PoolParams, coldVKey []byte) ([]byte, error)
	createPoolRetCertFunc func(ctx context.Context, poolID string, retirementEpoch int64, coldVKey []byte) ([]byte, error)
	buildTxFunc           func(ctx context.Context, opts *TxBuildOptions) ([]byte, error)
	signTxFunc            func(ctx context.Context, txBody []byte, signingKeys ...[]byte) ([]byte, error)
	submitTxFunc          func(ctx context.Context, signedTx []byte) (*TransactionSubmitResult, error)
	queryTipFunc          func(ctx context.Context) (*Tip, error)
	queryProtocolFunc     func(ctx context.Context) (*ProtocolParameters, error)
}

func (m *txMockClient) QueryUTxO(ctx context.Context, address string) ([]UTxO, error) {
	if m.queryUTxOFunc != nil {
		return m.queryUTxOFunc(ctx, address)
	}
	return []UTxO{{TxHash: "abc123", TxIndex: 0, Address: address, Value: Value{Lovelace: 10000000000}}}, nil
}

func (m *txMockClient) CreatePoolRegistrationCertificate(ctx context.Context, params *PoolParams, coldVKey []byte) ([]byte, error) {
	if m.createPoolRegCertFunc != nil {
		return m.createPoolRegCertFunc(ctx, params, coldVKey)
	}
	return []byte("mock-pool-reg-cert"), nil
}

func (m *txMockClient) CreatePoolUpdateCertificate(ctx context.Context, params *PoolParams, coldVKey []byte) ([]byte, error) {
	if m.createPoolUpdCertFunc != nil {
		return m.createPoolUpdCertFunc(ctx, params, coldVKey)
	}
	return []byte("mock-pool-update-cert"), nil
}

func (m *txMockClient) CreatePoolRetirementCertificate(ctx context.Context, poolID string, retirementEpoch int64, coldVKey []byte) ([]byte, error) {
	if m.createPoolRetCertFunc != nil {
		return m.createPoolRetCertFunc(ctx, poolID, retirementEpoch, coldVKey)
	}
	return []byte("mock-pool-retire-cert"), nil
}

func (m *txMockClient) BuildTx(ctx context.Context, opts *TxBuildOptions) ([]byte, error) {
	if m.buildTxFunc != nil {
		return m.buildTxFunc(ctx, opts)
	}
	return []byte("mock-tx-body"), nil
}

func (m *txMockClient) SignTx(ctx context.Context, txBody []byte, signingKeys ...[]byte) ([]byte, error) {
	if m.signTxFunc != nil {
		return m.signTxFunc(ctx, txBody, signingKeys...)
	}
	return []byte("mock-signed-tx"), nil
}

func (m *txMockClient) SubmitTx(ctx context.Context, signedTx []byte) (*TransactionSubmitResult, error) {
	if m.submitTxFunc != nil {
		return m.submitTxFunc(ctx, signedTx)
	}
	return &TransactionSubmitResult{TxHash: "tx-hash-123", Success: true}, nil
}

func (m *txMockClient) QueryTip(ctx context.Context) (*Tip, error) {
	if m.queryTipFunc != nil {
		return m.queryTipFunc(ctx)
	}
	return &Tip{Slot: 1000, Epoch: 100, Block: 500}, nil
}

func (m *txMockClient) QueryProtocolParameters(ctx context.Context) (*ProtocolParameters, error) {
	if m.queryProtocolFunc != nil {
		return m.queryProtocolFunc(ctx)
	}
	return &ProtocolParameters{PoolDeposit: 500000000, MaxEpoch: 18}, nil
}

// TestNewTransactionBuilder tests TransactionBuilder creation
func TestNewTransactionBuilder(t *testing.T) {
	client := &txMockClient{}
	tb := NewTransactionBuilder(client)

	if tb == nil {
		t.Fatal("NewTransactionBuilder returned nil")
	}
	if tb.client != client {
		t.Error("TransactionBuilder client not set correctly")
	}
}

// TestBuildPoolRegistrationTx tests pool registration transaction building
func TestBuildPoolRegistrationTx(t *testing.T) {
	ctx := context.Background()

	t.Run("successful pool registration", func(t *testing.T) {
		client := &txMockClient{}
		tb := NewTransactionBuilder(client)

		params := &PoolRegistrationTxParams{
			PoolParams: &PoolParams{
				VRFKeyHash:    "vrf-hash",
				Pledge:        500000000000,
				Cost:          340000000,
				Margin:        0.03,
				RewardAccount: "stake_test1abc",
				Owners:        []string{"stake_test1abc"},
				Relays:        []Relay{{Type: "dns", Hostname: "relay.example.com", Port: 6000}},
			},
			ColdVKey:       []byte("cold-vkey"),
			PaymentAddress: "addr_test1payment",
			ColdSKey:       []byte("cold-skey"),
			PaymentSKey:    []byte("payment-skey"),
			StakeSKey:      []byte("stake-skey"),
		}

		tx, err := tb.BuildPoolRegistrationTx(ctx, params)
		if err != nil {
			t.Fatalf("BuildPoolRegistrationTx() error = %v", err)
		}

		if tx == nil {
			t.Fatal("Transaction should not be nil")
		}
		if tx.TxBody == nil {
			t.Error("TxBody should not be nil")
		}
		if len(tx.Certificates) != 1 {
			t.Errorf("Expected 1 certificate, got %d", len(tx.Certificates))
		}
		if len(tx.RequiredSigners) != 3 {
			t.Errorf("Expected 3 required signers, got %d", len(tx.RequiredSigners))
		}
	})

	t.Run("certificate creation failure", func(t *testing.T) {
		client := &txMockClient{
			createPoolRegCertFunc: func(ctx context.Context, params *PoolParams, coldVKey []byte) ([]byte, error) {
				return nil, errors.New("certificate creation failed")
			},
		}
		tb := NewTransactionBuilder(client)

		params := &PoolRegistrationTxParams{
			PoolParams:     &PoolParams{},
			ColdVKey:       []byte("cold-vkey"),
			PaymentAddress: "addr_test1payment",
		}

		_, err := tb.BuildPoolRegistrationTx(ctx, params)
		if err == nil {
			t.Error("Expected error when certificate creation fails")
		}
	})

	t.Run("no UTxOs available", func(t *testing.T) {
		client := &txMockClient{
			queryUTxOFunc: func(ctx context.Context, address string) ([]UTxO, error) {
				return []UTxO{}, nil
			},
		}
		tb := NewTransactionBuilder(client)

		params := &PoolRegistrationTxParams{
			PoolParams:     &PoolParams{},
			ColdVKey:       []byte("cold-vkey"),
			PaymentAddress: "addr_test1payment",
		}

		_, err := tb.BuildPoolRegistrationTx(ctx, params)
		if err == nil {
			t.Error("Expected error when no UTxOs available")
		}
	})

	t.Run("UTxO query failure", func(t *testing.T) {
		client := &txMockClient{
			queryUTxOFunc: func(ctx context.Context, address string) ([]UTxO, error) {
				return nil, errors.New("UTxO query failed")
			},
		}
		tb := NewTransactionBuilder(client)

		params := &PoolRegistrationTxParams{
			PoolParams:     &PoolParams{},
			ColdVKey:       []byte("cold-vkey"),
			PaymentAddress: "addr_test1payment",
		}

		_, err := tb.BuildPoolRegistrationTx(ctx, params)
		if err == nil {
			t.Error("Expected error when UTxO query fails")
		}
	})

	t.Run("transaction build failure", func(t *testing.T) {
		client := &txMockClient{
			buildTxFunc: func(ctx context.Context, opts *TxBuildOptions) ([]byte, error) {
				return nil, errors.New("tx build failed")
			},
		}
		tb := NewTransactionBuilder(client)

		params := &PoolRegistrationTxParams{
			PoolParams:     &PoolParams{},
			ColdVKey:       []byte("cold-vkey"),
			PaymentAddress: "addr_test1payment",
		}

		_, err := tb.BuildPoolRegistrationTx(ctx, params)
		if err == nil {
			t.Error("Expected error when transaction build fails")
		}
	})
}

// TestBuildPoolUpdateTx tests pool update transaction building
func TestBuildPoolUpdateTx(t *testing.T) {
	ctx := context.Background()

	t.Run("successful pool update", func(t *testing.T) {
		client := &txMockClient{}
		tb := NewTransactionBuilder(client)

		params := &PoolUpdateTxParams{
			PoolParams: &PoolParams{
				PoolID:        "pool1abc123",
				VRFKeyHash:    "vrf-hash",
				Pledge:        600000000000,
				Cost:          340000000,
				Margin:        0.025,
				RewardAccount: "stake_test1abc",
				Owners:        []string{"stake_test1abc"},
				Relays:        []Relay{{Type: "dns", Hostname: "relay.example.com", Port: 6000}},
			},
			ColdVKey:       []byte("cold-vkey"),
			PaymentAddress: "addr_test1payment",
			ColdSKey:       []byte("cold-skey"),
			PaymentSKey:    []byte("payment-skey"),
		}

		tx, err := tb.BuildPoolUpdateTx(ctx, params)
		if err != nil {
			t.Fatalf("BuildPoolUpdateTx() error = %v", err)
		}

		if tx == nil {
			t.Fatal("Transaction should not be nil")
		}
		if len(tx.RequiredSigners) != 2 {
			t.Errorf("Expected 2 required signers, got %d", len(tx.RequiredSigners))
		}
	})

	t.Run("certificate creation failure", func(t *testing.T) {
		client := &txMockClient{
			createPoolUpdCertFunc: func(ctx context.Context, params *PoolParams, coldVKey []byte) ([]byte, error) {
				return nil, errors.New("certificate creation failed")
			},
		}
		tb := NewTransactionBuilder(client)

		params := &PoolUpdateTxParams{
			PoolParams:     &PoolParams{},
			ColdVKey:       []byte("cold-vkey"),
			PaymentAddress: "addr_test1payment",
		}

		_, err := tb.BuildPoolUpdateTx(ctx, params)
		if err == nil {
			t.Error("Expected error when certificate creation fails")
		}
	})
}

// TestBuildPoolRetirementTx tests pool retirement transaction building
func TestBuildPoolRetirementTx(t *testing.T) {
	ctx := context.Background()

	t.Run("successful pool retirement", func(t *testing.T) {
		client := &txMockClient{}
		tb := NewTransactionBuilder(client)

		params := &PoolRetirementTxParams{
			PoolID:          "pool1abc123",
			RetirementEpoch: 200,
			ColdVKey:        []byte("cold-vkey"),
			PaymentAddress:  "addr_test1payment",
			ColdSKey:        []byte("cold-skey"),
			PaymentSKey:     []byte("payment-skey"),
		}

		tx, err := tb.BuildPoolRetirementTx(ctx, params)
		if err != nil {
			t.Fatalf("BuildPoolRetirementTx() error = %v", err)
		}

		if tx == nil {
			t.Fatal("Transaction should not be nil")
		}
		if len(tx.RequiredSigners) != 2 {
			t.Errorf("Expected 2 required signers, got %d", len(tx.RequiredSigners))
		}
	})

	t.Run("certificate creation failure", func(t *testing.T) {
		client := &txMockClient{
			createPoolRetCertFunc: func(ctx context.Context, poolID string, retirementEpoch int64, coldVKey []byte) ([]byte, error) {
				return nil, errors.New("retirement certificate creation failed")
			},
		}
		tb := NewTransactionBuilder(client)

		params := &PoolRetirementTxParams{
			PoolID:          "pool1abc123",
			RetirementEpoch: 200,
			ColdVKey:        []byte("cold-vkey"),
			PaymentAddress:  "addr_test1payment",
		}

		_, err := tb.BuildPoolRetirementTx(ctx, params)
		if err == nil {
			t.Error("Expected error when certificate creation fails")
		}
	})
}

// TestBuildKESCertificateTx tests that KES certificate tx returns an error
func TestBuildKESCertificateTx(t *testing.T) {
	ctx := context.Background()
	client := &txMockClient{}
	tb := NewTransactionBuilder(client)

	params := &KESCertificateTxParams{
		OpCert: &OperationalCertificate{
			Certificate: []byte("cert"),
			Counter:     1,
			KESPeriod:   100,
		},
		PaymentAddress: "addr_test1",
		PaymentSKey:    []byte("payment-skey"),
	}

	_, err := tb.BuildKESCertificateTx(ctx, params)
	if err == nil {
		t.Error("Expected error for KES certificate transaction (not supported)")
	}
}

// TestBuiltTransactionSign tests signing a built transaction
func TestBuiltTransactionSign(t *testing.T) {
	ctx := context.Background()

	t.Run("successful signing", func(t *testing.T) {
		client := &txMockClient{}
		bt := &BuiltTransaction{
			TxBody:          []byte("tx-body"),
			Certificates:    [][]byte{[]byte("cert")},
			RequiredSigners: [][]byte{[]byte("skey1"), []byte("skey2")},
		}

		signedTx, err := bt.Sign(ctx, client)
		if err != nil {
			t.Fatalf("Sign() error = %v", err)
		}
		if signedTx == nil {
			t.Error("Signed transaction should not be nil")
		}
	})

	t.Run("signing failure", func(t *testing.T) {
		client := &txMockClient{
			signTxFunc: func(ctx context.Context, txBody []byte, signingKeys ...[]byte) ([]byte, error) {
				return nil, errors.New("signing failed")
			},
		}
		bt := &BuiltTransaction{
			TxBody:          []byte("tx-body"),
			RequiredSigners: [][]byte{[]byte("skey")},
		}

		_, err := bt.Sign(ctx, client)
		if err == nil {
			t.Error("Expected error when signing fails")
		}
	})
}

// TestBuiltTransactionSignAndSubmit tests signing and submitting a transaction
func TestBuiltTransactionSignAndSubmit(t *testing.T) {
	ctx := context.Background()

	t.Run("successful sign and submit", func(t *testing.T) {
		client := &txMockClient{}
		bt := &BuiltTransaction{
			TxBody:          []byte("tx-body"),
			RequiredSigners: [][]byte{[]byte("skey")},
		}

		result, err := bt.SignAndSubmit(ctx, client)
		if err != nil {
			t.Fatalf("SignAndSubmit() error = %v", err)
		}
		if result == nil {
			t.Fatal("Result should not be nil")
		}
		if !result.Success {
			t.Error("Transaction should be successful")
		}
		if result.TxHash == "" {
			t.Error("TxHash should not be empty")
		}
	})

	t.Run("signing failure", func(t *testing.T) {
		client := &txMockClient{
			signTxFunc: func(ctx context.Context, txBody []byte, signingKeys ...[]byte) ([]byte, error) {
				return nil, errors.New("signing failed")
			},
		}
		bt := &BuiltTransaction{
			TxBody:          []byte("tx-body"),
			RequiredSigners: [][]byte{[]byte("skey")},
		}

		_, err := bt.SignAndSubmit(ctx, client)
		if err == nil {
			t.Error("Expected error when signing fails")
		}
	})

	t.Run("submission failure", func(t *testing.T) {
		client := &txMockClient{
			submitTxFunc: func(ctx context.Context, signedTx []byte) (*TransactionSubmitResult, error) {
				return nil, errors.New("submission failed")
			},
		}
		bt := &BuiltTransaction{
			TxBody:          []byte("tx-body"),
			RequiredSigners: [][]byte{[]byte("skey")},
		}

		_, err := bt.SignAndSubmit(ctx, client)
		if err == nil {
			t.Error("Expected error when submission fails")
		}
	})
}

// TestCalculateRequiredDeposit tests deposit calculation
func TestCalculateRequiredDeposit(t *testing.T) {
	ctx := context.Background()

	t.Run("successful calculation", func(t *testing.T) {
		client := &txMockClient{}
		tb := NewTransactionBuilder(client)

		deposit, err := tb.CalculateRequiredDeposit(ctx)
		if err != nil {
			t.Fatalf("CalculateRequiredDeposit() error = %v", err)
		}
		if deposit != 500000000 {
			t.Errorf("Expected deposit 500000000, got %d", deposit)
		}
	})

	t.Run("protocol parameters query failure", func(t *testing.T) {
		client := &txMockClient{
			queryProtocolFunc: func(ctx context.Context) (*ProtocolParameters, error) {
				return nil, errors.New("query failed")
			},
		}
		tb := NewTransactionBuilder(client)

		_, err := tb.CalculateRequiredDeposit(ctx)
		if err == nil {
			t.Error("Expected error when protocol parameters query fails")
		}
	})
}

// TestCalculateMaxRetirementEpoch tests max retirement epoch calculation
func TestCalculateMaxRetirementEpoch(t *testing.T) {
	ctx := context.Background()

	t.Run("successful calculation", func(t *testing.T) {
		client := &txMockClient{}
		tb := NewTransactionBuilder(client)

		maxEpoch, err := tb.CalculateMaxRetirementEpoch(ctx)
		if err != nil {
			t.Fatalf("CalculateMaxRetirementEpoch() error = %v", err)
		}
		// Current epoch (100) + MaxEpoch (18) = 118
		if maxEpoch != 118 {
			t.Errorf("Expected max epoch 118, got %d", maxEpoch)
		}
	})

	t.Run("protocol parameters query failure", func(t *testing.T) {
		client := &txMockClient{
			queryProtocolFunc: func(ctx context.Context) (*ProtocolParameters, error) {
				return nil, errors.New("query failed")
			},
		}
		tb := NewTransactionBuilder(client)

		_, err := tb.CalculateMaxRetirementEpoch(ctx)
		if err == nil {
			t.Error("Expected error when protocol parameters query fails")
		}
	})

	t.Run("tip query failure", func(t *testing.T) {
		client := &txMockClient{
			queryTipFunc: func(ctx context.Context) (*Tip, error) {
				return nil, errors.New("tip query failed")
			},
		}
		tb := NewTransactionBuilder(client)

		_, err := tb.CalculateMaxRetirementEpoch(ctx)
		if err == nil {
			t.Error("Expected error when tip query fails")
		}
	})
}

// TestValidateRetirementEpoch tests retirement epoch validation
func TestValidateRetirementEpoch(t *testing.T) {
	ctx := context.Background()

	t.Run("valid retirement epoch", func(t *testing.T) {
		client := &txMockClient{}
		tb := NewTransactionBuilder(client)

		// Current epoch is 100, max is 118, so 110 should be valid
		err := tb.ValidateRetirementEpoch(ctx, 110)
		if err != nil {
			t.Errorf("ValidateRetirementEpoch() unexpected error = %v", err)
		}
	})

	t.Run("retirement epoch in the past", func(t *testing.T) {
		client := &txMockClient{}
		tb := NewTransactionBuilder(client)

		// Current epoch is 100, so 90 should be invalid
		err := tb.ValidateRetirementEpoch(ctx, 90)
		if err == nil {
			t.Error("Expected error for past retirement epoch")
		}
	})

	t.Run("retirement epoch equals current", func(t *testing.T) {
		client := &txMockClient{}
		tb := NewTransactionBuilder(client)

		// Current epoch is 100, so 100 should be invalid
		err := tb.ValidateRetirementEpoch(ctx, 100)
		if err == nil {
			t.Error("Expected error for retirement epoch equal to current")
		}
	})

	t.Run("retirement epoch too far in future", func(t *testing.T) {
		client := &txMockClient{}
		tb := NewTransactionBuilder(client)

		// Max epoch is 118, so 200 should be invalid
		err := tb.ValidateRetirementEpoch(ctx, 200)
		if err == nil {
			t.Error("Expected error for retirement epoch beyond max")
		}
	})

	t.Run("tip query failure", func(t *testing.T) {
		client := &txMockClient{
			queryTipFunc: func(ctx context.Context) (*Tip, error) {
				return nil, errors.New("tip query failed")
			},
		}
		tb := NewTransactionBuilder(client)

		err := tb.ValidateRetirementEpoch(ctx, 110)
		if err == nil {
			t.Error("Expected error when tip query fails")
		}
	})
}
