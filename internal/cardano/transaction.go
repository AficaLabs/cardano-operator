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
	"fmt"
)

// TransactionBuilder provides high-level transaction building operations
type TransactionBuilder struct {
	client Client
}

// NewTransactionBuilder creates a new TransactionBuilder
func NewTransactionBuilder(client Client) *TransactionBuilder {
	return &TransactionBuilder{client: client}
}

// PoolRegistrationTxParams contains parameters for pool registration
type PoolRegistrationTxParams struct {
	// Pool parameters
	PoolParams *PoolParams

	// Cold verification key
	ColdVKey []byte

	// Payment address to fund the transaction and receive change
	PaymentAddress string

	// Signing keys for the transaction
	ColdSKey    []byte
	PaymentSKey []byte
	StakeSKey   []byte
}

// BuildPoolRegistrationTx builds a pool registration transaction
func (tb *TransactionBuilder) BuildPoolRegistrationTx(ctx context.Context, params *PoolRegistrationTxParams) (*BuiltTransaction, error) {
	// Create the pool registration certificate
	cert, err := tb.client.CreatePoolRegistrationCertificate(ctx, params.PoolParams, params.ColdVKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool registration certificate: %w", err)
	}

	// Query UTxOs for funding
	utxos, err := tb.client.QueryUTxO(ctx, params.PaymentAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to query UTxOs: %w", err)
	}

	if len(utxos) == 0 {
		return nil, fmt.Errorf("no UTxOs found at payment address")
	}

	// Build the transaction
	txBody, err := tb.client.BuildTx(ctx, &TxBuildOptions{
		Inputs:        utxos,
		Certificates:  [][]byte{cert},
		ChangeAddress: params.PaymentAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	return &BuiltTransaction{
		TxBody:       txBody,
		Certificates: [][]byte{cert},
		RequiredSigners: [][]byte{
			params.ColdSKey,
			params.PaymentSKey,
			params.StakeSKey,
		},
	}, nil
}

// PoolUpdateTxParams contains parameters for pool update
type PoolUpdateTxParams struct {
	// Updated pool parameters
	PoolParams *PoolParams

	// Cold verification key
	ColdVKey []byte

	// Payment address
	PaymentAddress string

	// Signing keys
	ColdSKey    []byte
	PaymentSKey []byte
}

// BuildPoolUpdateTx builds a pool update transaction
func (tb *TransactionBuilder) BuildPoolUpdateTx(ctx context.Context, params *PoolUpdateTxParams) (*BuiltTransaction, error) {
	// Create the pool update certificate
	cert, err := tb.client.CreatePoolUpdateCertificate(ctx, params.PoolParams, params.ColdVKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool update certificate: %w", err)
	}

	// Query UTxOs for funding
	utxos, err := tb.client.QueryUTxO(ctx, params.PaymentAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to query UTxOs: %w", err)
	}

	if len(utxos) == 0 {
		return nil, fmt.Errorf("no UTxOs found at payment address")
	}

	// Build the transaction
	txBody, err := tb.client.BuildTx(ctx, &TxBuildOptions{
		Inputs:        utxos,
		Certificates:  [][]byte{cert},
		ChangeAddress: params.PaymentAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	return &BuiltTransaction{
		TxBody:       txBody,
		Certificates: [][]byte{cert},
		RequiredSigners: [][]byte{
			params.ColdSKey,
			params.PaymentSKey,
		},
	}, nil
}

// PoolRetirementTxParams contains parameters for pool retirement
type PoolRetirementTxParams struct {
	// Pool ID to retire
	PoolID string

	// Retirement epoch
	RetirementEpoch int64

	// Cold verification key
	ColdVKey []byte

	// Payment address
	PaymentAddress string

	// Signing keys
	ColdSKey    []byte
	PaymentSKey []byte
}

// BuildPoolRetirementTx builds a pool retirement transaction
func (tb *TransactionBuilder) BuildPoolRetirementTx(ctx context.Context, params *PoolRetirementTxParams) (*BuiltTransaction, error) {
	// Create the retirement certificate
	cert, err := tb.client.CreatePoolRetirementCertificate(ctx, params.PoolID, params.RetirementEpoch, params.ColdVKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create retirement certificate: %w", err)
	}

	// Query UTxOs for funding
	utxos, err := tb.client.QueryUTxO(ctx, params.PaymentAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to query UTxOs: %w", err)
	}

	if len(utxos) == 0 {
		return nil, fmt.Errorf("no UTxOs found at payment address")
	}

	// Build the transaction
	txBody, err := tb.client.BuildTx(ctx, &TxBuildOptions{
		Inputs:        utxos,
		Certificates:  [][]byte{cert},
		ChangeAddress: params.PaymentAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	return &BuiltTransaction{
		TxBody:       txBody,
		Certificates: [][]byte{cert},
		RequiredSigners: [][]byte{
			params.ColdSKey,
			params.PaymentSKey,
		},
	}, nil
}

// KESCertificateTxParams contains parameters for KES certificate update
type KESCertificateTxParams struct {
	// Operational certificate data
	OpCert *OperationalCertificate

	// Payment address
	PaymentAddress string

	// Signing keys
	PaymentSKey []byte
}

// BuildKESCertificateTx builds a transaction to submit a new KES certificate
// Note: In Cardano, KES certificates are not submitted as transactions.
// They are included in block headers. This function is a placeholder
// for any on-chain operations that might be needed.
func (tb *TransactionBuilder) BuildKESCertificateTx(ctx context.Context, params *KESCertificateTxParams) (*BuiltTransaction, error) {
	// KES certificates are not submitted as transactions in Cardano
	// They are used directly by the block producer
	// This is a placeholder for future use
	return nil, fmt.Errorf("KES certificates are not submitted as transactions; they are used directly by the block producer")
}

// BuiltTransaction contains a built transaction ready for signing
type BuiltTransaction struct {
	// Raw transaction body
	TxBody []byte

	// Certificates included in the transaction
	Certificates [][]byte

	// Signing keys required for this transaction
	RequiredSigners [][]byte
}

// Sign signs the transaction with all required signers
func (bt *BuiltTransaction) Sign(ctx context.Context, client Client) ([]byte, error) {
	return client.SignTx(ctx, bt.TxBody, bt.RequiredSigners...)
}

// SignAndSubmit signs and submits the transaction
func (bt *BuiltTransaction) SignAndSubmit(ctx context.Context, client Client) (*TransactionSubmitResult, error) {
	signedTx, err := bt.Sign(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return client.SubmitTx(ctx, signedTx)
}

// CalculateRequiredDeposit calculates the deposit required for a pool registration
func (tb *TransactionBuilder) CalculateRequiredDeposit(ctx context.Context) (int64, error) {
	params, err := tb.client.QueryProtocolParameters(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to query protocol parameters: %w", err)
	}

	return params.PoolDeposit, nil
}

// CalculateMaxRetirementEpoch calculates the maximum epoch for pool retirement
func (tb *TransactionBuilder) CalculateMaxRetirementEpoch(ctx context.Context) (int64, error) {
	params, err := tb.client.QueryProtocolParameters(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to query protocol parameters: %w", err)
	}

	tip, err := tb.client.QueryTip(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to query tip: %w", err)
	}

	return tip.Epoch + params.MaxEpoch, nil
}

// ValidateRetirementEpoch validates that a retirement epoch is valid
func (tb *TransactionBuilder) ValidateRetirementEpoch(ctx context.Context, retirementEpoch int64) error {
	tip, err := tb.client.QueryTip(ctx)
	if err != nil {
		return fmt.Errorf("failed to query tip: %w", err)
	}

	if retirementEpoch <= tip.Epoch {
		return fmt.Errorf("retirement epoch %d must be greater than current epoch %d", retirementEpoch, tip.Epoch)
	}

	maxEpoch, err := tb.CalculateMaxRetirementEpoch(ctx)
	if err != nil {
		return fmt.Errorf("failed to calculate max retirement epoch: %w", err)
	}

	if retirementEpoch > maxEpoch {
		return fmt.Errorf("retirement epoch %d exceeds maximum allowed epoch %d", retirementEpoch, maxEpoch)
	}

	return nil
}
