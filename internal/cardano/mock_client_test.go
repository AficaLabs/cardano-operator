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

func TestMockClient_QueryOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("QueryTip success", func(t *testing.T) {
		mock := NewMockClient()
		tip, err := mock.QueryTip(ctx)
		if err != nil {
			t.Errorf("QueryTip() error = %v", err)
		}
		if tip.Slot != 1000000 {
			t.Errorf("QueryTip().Slot = %v, want 1000000", tip.Slot)
		}
		if mock.QueryTipCalled != 1 {
			t.Errorf("QueryTipCalled = %v, want 1", mock.QueryTipCalled)
		}
	})

	t.Run("QueryTip error", func(t *testing.T) {
		mock := NewMockClient()
		mock.TipError = errors.New("connection failed")
		_, err := mock.QueryTip(ctx)
		if err == nil {
			t.Error("QueryTip() expected error")
		}
	})

	t.Run("QueryUTxO success", func(t *testing.T) {
		mock := NewMockClient()
		utxos, err := mock.QueryUTxO(ctx, "addr_test1qz")
		if err != nil {
			t.Errorf("QueryUTxO() error = %v", err)
		}
		if len(utxos) != 1 {
			t.Errorf("QueryUTxO() returned %d UTxOs, want 1", len(utxos))
		}
		if mock.QueryUTxOCalled != 1 {
			t.Errorf("QueryUTxOCalled = %v, want 1", mock.QueryUTxOCalled)
		}
	})

	t.Run("QueryUTxO error", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOError = errors.New("address not found")
		_, err := mock.QueryUTxO(ctx, "invalid")
		if err == nil {
			t.Error("QueryUTxO() expected error")
		}
	})

	t.Run("QueryProtocolParameters success", func(t *testing.T) {
		mock := NewMockClient()
		params, err := mock.QueryProtocolParameters(ctx)
		if err != nil {
			t.Errorf("QueryProtocolParameters() error = %v", err)
		}
		if params.PoolDeposit != 500000000 {
			t.Errorf("QueryProtocolParameters().PoolDeposit = %v, want 500000000", params.PoolDeposit)
		}
		if mock.QueryProtocolParametersCalled != 1 {
			t.Errorf("QueryProtocolParametersCalled = %v, want 1", mock.QueryProtocolParametersCalled)
		}
	})

	t.Run("QueryProtocolParameters error", func(t *testing.T) {
		mock := NewMockClient()
		mock.ProtocolParamsError = errors.New("node not synced")
		_, err := mock.QueryProtocolParameters(ctx)
		if err == nil {
			t.Error("QueryProtocolParameters() expected error")
		}
	})

	t.Run("QueryPoolParams success", func(t *testing.T) {
		mock := NewMockClient()
		params, err := mock.QueryPoolParams(ctx, "pool1abc")
		if err != nil {
			t.Errorf("QueryPoolParams() error = %v", err)
		}
		if params.PoolID != "pool1abc123" {
			t.Errorf("QueryPoolParams().PoolID = %v, want pool1abc123", params.PoolID)
		}
		if mock.QueryPoolParamsCalled != 1 {
			t.Errorf("QueryPoolParamsCalled = %v, want 1", mock.QueryPoolParamsCalled)
		}
	})

	t.Run("QueryPoolParams error", func(t *testing.T) {
		mock := NewMockClient()
		mock.PoolParamsError = errors.New("pool not found")
		_, err := mock.QueryPoolParams(ctx, "invalid")
		if err == nil {
			t.Error("QueryPoolParams() expected error")
		}
	})

	t.Run("QueryStakePoolID success", func(t *testing.T) {
		mock := NewMockClient()
		poolID, err := mock.QueryStakePoolID(ctx, "/path/to/cold.vkey")
		if err != nil {
			t.Errorf("QueryStakePoolID() error = %v", err)
		}
		if poolID != "pool1abc123def456" {
			t.Errorf("QueryStakePoolID() = %v, want pool1abc123def456", poolID)
		}
		if mock.QueryStakePoolIDCalled != 1 {
			t.Errorf("QueryStakePoolIDCalled = %v, want 1", mock.QueryStakePoolIDCalled)
		}
	})

	t.Run("QueryStakePoolID error", func(t *testing.T) {
		mock := NewMockClient()
		mock.StakePoolIDError = errors.New("invalid key")
		_, err := mock.QueryStakePoolID(ctx, "invalid")
		if err == nil {
			t.Error("QueryStakePoolID() expected error")
		}
	})
}

func TestMockClient_KeyGeneration(t *testing.T) {
	ctx := context.Background()

	t.Run("GenerateColdKeys success", func(t *testing.T) {
		mock := NewMockClient()
		keys, err := mock.GenerateColdKeys(ctx)
		if err != nil {
			t.Errorf("GenerateColdKeys() error = %v", err)
		}
		if keys == nil {
			t.Error("GenerateColdKeys() returned nil")
		}
		if mock.GenerateColdKeysCalled != 1 {
			t.Errorf("GenerateColdKeysCalled = %v, want 1", mock.GenerateColdKeysCalled)
		}
	})

	t.Run("GenerateColdKeys error", func(t *testing.T) {
		mock := NewMockClient()
		mock.ColdKeysError = errors.New("key generation failed")
		_, err := mock.GenerateColdKeys(ctx)
		if err == nil {
			t.Error("GenerateColdKeys() expected error")
		}
	})

	t.Run("GenerateVRFKeys success", func(t *testing.T) {
		mock := NewMockClient()
		keys, err := mock.GenerateVRFKeys(ctx)
		if err != nil {
			t.Errorf("GenerateVRFKeys() error = %v", err)
		}
		if keys == nil {
			t.Error("GenerateVRFKeys() returned nil")
		}
		if mock.GenerateVRFKeysCalled != 1 {
			t.Errorf("GenerateVRFKeysCalled = %v, want 1", mock.GenerateVRFKeysCalled)
		}
	})

	t.Run("GenerateVRFKeys error", func(t *testing.T) {
		mock := NewMockClient()
		mock.VRFKeysError = errors.New("VRF generation failed")
		_, err := mock.GenerateVRFKeys(ctx)
		if err == nil {
			t.Error("GenerateVRFKeys() expected error")
		}
	})

	t.Run("GenerateKESKeys success", func(t *testing.T) {
		mock := NewMockClient()
		keys, err := mock.GenerateKESKeys(ctx)
		if err != nil {
			t.Errorf("GenerateKESKeys() error = %v", err)
		}
		if keys == nil {
			t.Error("GenerateKESKeys() returned nil")
		}
		if mock.GenerateKESKeysCalled != 1 {
			t.Errorf("GenerateKESKeysCalled = %v, want 1", mock.GenerateKESKeysCalled)
		}
	})

	t.Run("GenerateKESKeys error", func(t *testing.T) {
		mock := NewMockClient()
		mock.KESKeysError = errors.New("KES generation failed")
		_, err := mock.GenerateKESKeys(ctx)
		if err == nil {
			t.Error("GenerateKESKeys() expected error")
		}
	})

	t.Run("GeneratePaymentKeys success", func(t *testing.T) {
		mock := NewMockClient()
		keys, err := mock.GeneratePaymentKeys(ctx)
		if err != nil {
			t.Errorf("GeneratePaymentKeys() error = %v", err)
		}
		if keys == nil {
			t.Error("GeneratePaymentKeys() returned nil")
		}
		if mock.GeneratePaymentKeysCalled != 1 {
			t.Errorf("GeneratePaymentKeysCalled = %v, want 1", mock.GeneratePaymentKeysCalled)
		}
	})

	t.Run("GeneratePaymentKeys error", func(t *testing.T) {
		mock := NewMockClient()
		mock.PaymentKeysError = errors.New("payment key generation failed")
		_, err := mock.GeneratePaymentKeys(ctx)
		if err == nil {
			t.Error("GeneratePaymentKeys() expected error")
		}
	})

	t.Run("GenerateStakeKeys success", func(t *testing.T) {
		mock := NewMockClient()
		keys, err := mock.GenerateStakeKeys(ctx)
		if err != nil {
			t.Errorf("GenerateStakeKeys() error = %v", err)
		}
		if keys == nil {
			t.Error("GenerateStakeKeys() returned nil")
		}
		if mock.GenerateStakeKeysCalled != 1 {
			t.Errorf("GenerateStakeKeysCalled = %v, want 1", mock.GenerateStakeKeysCalled)
		}
	})

	t.Run("GenerateStakeKeys error", func(t *testing.T) {
		mock := NewMockClient()
		mock.StakeKeysError = errors.New("stake key generation failed")
		_, err := mock.GenerateStakeKeys(ctx)
		if err == nil {
			t.Error("GenerateStakeKeys() expected error")
		}
	})
}

func TestMockClient_CertificateOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateOperationalCertificate success", func(t *testing.T) {
		mock := NewMockClient()
		cert, err := mock.CreateOperationalCertificate(ctx, []byte("kes"), []byte("cold"), 5, 100)
		if err != nil {
			t.Errorf("CreateOperationalCertificate() error = %v", err)
		}
		if cert.Counter != 5 {
			t.Errorf("CreateOperationalCertificate().Counter = %v, want 5", cert.Counter)
		}
		if mock.CreateOpCertCalled != 1 {
			t.Errorf("CreateOpCertCalled = %v, want 1", mock.CreateOpCertCalled)
		}
	})

	t.Run("CreateOperationalCertificate error", func(t *testing.T) {
		mock := NewMockClient()
		mock.OpCertError = errors.New("cert creation failed")
		_, err := mock.CreateOperationalCertificate(ctx, nil, nil, 0, 0)
		if err == nil {
			t.Error("CreateOperationalCertificate() expected error")
		}
	})

	t.Run("CreatePoolRegistrationCertificate success", func(t *testing.T) {
		mock := NewMockClient()
		cert, err := mock.CreatePoolRegistrationCertificate(ctx, &PoolParams{}, []byte("cold"))
		if err != nil {
			t.Errorf("CreatePoolRegistrationCertificate() error = %v", err)
		}
		if cert == nil {
			t.Error("CreatePoolRegistrationCertificate() returned nil")
		}
		if mock.CreatePoolRegCertCalled != 1 {
			t.Errorf("CreatePoolRegCertCalled = %v, want 1", mock.CreatePoolRegCertCalled)
		}
	})

	t.Run("CreatePoolRegistrationCertificate error", func(t *testing.T) {
		mock := NewMockClient()
		mock.PoolRegCertError = errors.New("invalid params")
		_, err := mock.CreatePoolRegistrationCertificate(ctx, nil, nil)
		if err == nil {
			t.Error("CreatePoolRegistrationCertificate() expected error")
		}
	})

	t.Run("CreatePoolUpdateCertificate success", func(t *testing.T) {
		mock := NewMockClient()
		cert, err := mock.CreatePoolUpdateCertificate(ctx, &PoolParams{}, []byte("cold"))
		if err != nil {
			t.Errorf("CreatePoolUpdateCertificate() error = %v", err)
		}
		if cert == nil {
			t.Error("CreatePoolUpdateCertificate() returned nil")
		}
		if mock.CreatePoolUpdateCertCalled != 1 {
			t.Errorf("CreatePoolUpdateCertCalled = %v, want 1", mock.CreatePoolUpdateCertCalled)
		}
	})

	t.Run("CreatePoolUpdateCertificate error", func(t *testing.T) {
		mock := NewMockClient()
		mock.PoolUpdateCertError = errors.New("update failed")
		_, err := mock.CreatePoolUpdateCertificate(ctx, nil, nil)
		if err == nil {
			t.Error("CreatePoolUpdateCertificate() expected error")
		}
	})

	t.Run("CreatePoolRetirementCertificate success", func(t *testing.T) {
		mock := NewMockClient()
		cert, err := mock.CreatePoolRetirementCertificate(ctx, "pool1abc", 200, []byte("cold"))
		if err != nil {
			t.Errorf("CreatePoolRetirementCertificate() error = %v", err)
		}
		if cert == nil {
			t.Error("CreatePoolRetirementCertificate() returned nil")
		}
		if mock.CreatePoolRetireCertCalled != 1 {
			t.Errorf("CreatePoolRetireCertCalled = %v, want 1", mock.CreatePoolRetireCertCalled)
		}
	})

	t.Run("CreatePoolRetirementCertificate error", func(t *testing.T) {
		mock := NewMockClient()
		mock.PoolRetireCertError = errors.New("retirement failed")
		_, err := mock.CreatePoolRetirementCertificate(ctx, "", 0, nil)
		if err == nil {
			t.Error("CreatePoolRetirementCertificate() expected error")
		}
	})
}

func TestMockClient_TransactionOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("BuildTx success", func(t *testing.T) {
		mock := NewMockClient()
		tx, err := mock.BuildTx(ctx, &TxBuildOptions{})
		if err != nil {
			t.Errorf("BuildTx() error = %v", err)
		}
		if tx == nil {
			t.Error("BuildTx() returned nil")
		}
		if mock.BuildTxCalled != 1 {
			t.Errorf("BuildTxCalled = %v, want 1", mock.BuildTxCalled)
		}
	})

	t.Run("BuildTx error", func(t *testing.T) {
		mock := NewMockClient()
		mock.BuildTxError = errors.New("insufficient funds")
		_, err := mock.BuildTx(ctx, nil)
		if err == nil {
			t.Error("BuildTx() expected error")
		}
	})

	t.Run("SignTx success", func(t *testing.T) {
		mock := NewMockClient()
		signedTx, err := mock.SignTx(ctx, []byte("tx"), []byte("key1"), []byte("key2"))
		if err != nil {
			t.Errorf("SignTx() error = %v", err)
		}
		if signedTx == nil {
			t.Error("SignTx() returned nil")
		}
		if mock.SignTxCalled != 1 {
			t.Errorf("SignTxCalled = %v, want 1", mock.SignTxCalled)
		}
	})

	t.Run("SignTx error", func(t *testing.T) {
		mock := NewMockClient()
		mock.SignTxError = errors.New("invalid key")
		_, err := mock.SignTx(ctx, nil)
		if err == nil {
			t.Error("SignTx() expected error")
		}
	})

	t.Run("SubmitTx success", func(t *testing.T) {
		mock := NewMockClient()
		result, err := mock.SubmitTx(ctx, []byte("signed-tx"))
		if err != nil {
			t.Errorf("SubmitTx() error = %v", err)
		}
		if !result.Success {
			t.Error("SubmitTx() Success = false, want true")
		}
		if result.TxHash != "tx-hash-abc" {
			t.Errorf("SubmitTx().TxHash = %v, want tx-hash-abc", result.TxHash)
		}
		if mock.SubmitTxCalled != 1 {
			t.Errorf("SubmitTxCalled = %v, want 1", mock.SubmitTxCalled)
		}
	})

	t.Run("SubmitTx error", func(t *testing.T) {
		mock := NewMockClient()
		mock.SubmitTxError = errors.New("submission failed")
		_, err := mock.SubmitTx(ctx, nil)
		if err == nil {
			t.Error("SubmitTx() expected error")
		}
	})
}

func TestMockClient_AddressOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("BuildAddress success", func(t *testing.T) {
		mock := NewMockClient()
		addr, err := mock.BuildAddress(ctx, []byte("pay"), []byte("stake"))
		if err != nil {
			t.Errorf("BuildAddress() error = %v", err)
		}
		if addr != "addr_test1qz123" {
			t.Errorf("BuildAddress() = %v, want addr_test1qz123", addr)
		}
		if mock.BuildAddressCalled != 1 {
			t.Errorf("BuildAddressCalled = %v, want 1", mock.BuildAddressCalled)
		}
	})

	t.Run("BuildAddress error", func(t *testing.T) {
		mock := NewMockClient()
		mock.BuildAddressError = errors.New("invalid key")
		_, err := mock.BuildAddress(ctx, nil, nil)
		if err == nil {
			t.Error("BuildAddress() expected error")
		}
	})

	t.Run("BuildStakeAddress success", func(t *testing.T) {
		mock := NewMockClient()
		addr, err := mock.BuildStakeAddress(ctx, []byte("stake"))
		if err != nil {
			t.Errorf("BuildStakeAddress() error = %v", err)
		}
		if addr != "stake_test1uz123" {
			t.Errorf("BuildStakeAddress() = %v, want stake_test1uz123", addr)
		}
		if mock.BuildStakeAddressCalled != 1 {
			t.Errorf("BuildStakeAddressCalled = %v, want 1", mock.BuildStakeAddressCalled)
		}
	})

	t.Run("BuildStakeAddress error", func(t *testing.T) {
		mock := NewMockClient()
		mock.BuildStakeAddressError = errors.New("invalid stake key")
		_, err := mock.BuildStakeAddress(ctx, nil)
		if err == nil {
			t.Error("BuildStakeAddress() expected error")
		}
	})
}

func TestMockClient_UtilityOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("CalculateMinFee success", func(t *testing.T) {
		mock := NewMockClient()
		fee, err := mock.CalculateMinFee(ctx, []byte("tx"), 2)
		if err != nil {
			t.Errorf("CalculateMinFee() error = %v", err)
		}
		if fee != 200000 {
			t.Errorf("CalculateMinFee() = %v, want 200000", fee)
		}
		if mock.CalculateMinFeeCalled != 1 {
			t.Errorf("CalculateMinFeeCalled = %v, want 1", mock.CalculateMinFeeCalled)
		}
	})

	t.Run("CalculateMinFee error", func(t *testing.T) {
		mock := NewMockClient()
		mock.MinFeeError = errors.New("calculation failed")
		_, err := mock.CalculateMinFee(ctx, nil, 0)
		if err == nil {
			t.Error("CalculateMinFee() expected error")
		}
	})

	t.Run("GetCurrentKESPeriod success", func(t *testing.T) {
		mock := NewMockClient()
		period, err := mock.GetCurrentKESPeriod(ctx)
		if err != nil {
			t.Errorf("GetCurrentKESPeriod() error = %v", err)
		}
		if period != 100 {
			t.Errorf("GetCurrentKESPeriod() = %v, want 100", period)
		}
		if mock.GetCurrentKESPeriodCalled != 1 {
			t.Errorf("GetCurrentKESPeriodCalled = %v, want 1", mock.GetCurrentKESPeriodCalled)
		}
	})

	t.Run("GetCurrentKESPeriod error", func(t *testing.T) {
		mock := NewMockClient()
		mock.KESPeriodError = errors.New("node not synced")
		_, err := mock.GetCurrentKESPeriod(ctx)
		if err == nil {
			t.Error("GetCurrentKESPeriod() expected error")
		}
	})
}

func TestMockClient_CallTracking(t *testing.T) {
	ctx := context.Background()
	mock := NewMockClient()

	// Call multiple operations
	_, _ = mock.QueryTip(ctx)
	_, _ = mock.QueryTip(ctx)
	_, _ = mock.QueryUTxO(ctx, "addr")
	_, _ = mock.GenerateColdKeys(ctx)
	_, _ = mock.GenerateKESKeys(ctx)
	_, _ = mock.GenerateKESKeys(ctx)
	_, _ = mock.GenerateKESKeys(ctx)

	// Verify call counts
	if mock.QueryTipCalled != 2 {
		t.Errorf("QueryTipCalled = %v, want 2", mock.QueryTipCalled)
	}
	if mock.QueryUTxOCalled != 1 {
		t.Errorf("QueryUTxOCalled = %v, want 1", mock.QueryUTxOCalled)
	}
	if mock.GenerateColdKeysCalled != 1 {
		t.Errorf("GenerateColdKeysCalled = %v, want 1", mock.GenerateColdKeysCalled)
	}
	if mock.GenerateKESKeysCalled != 3 {
		t.Errorf("GenerateKESKeysCalled = %v, want 3", mock.GenerateKESKeysCalled)
	}
}

func TestNewMockClient(t *testing.T) {
	mock := NewMockClient()

	// Verify defaults are set
	if mock.TipResult == nil {
		t.Error("TipResult should not be nil")
	}
	if mock.UTxOResult == nil {
		t.Error("UTxOResult should not be nil")
	}
	if mock.ProtocolParamsResult == nil {
		t.Error("ProtocolParamsResult should not be nil")
	}
	if mock.PoolParamsResult == nil {
		t.Error("PoolParamsResult should not be nil")
	}
	if mock.ColdKeysResult == nil {
		t.Error("ColdKeysResult should not be nil")
	}
	if mock.OpCertResult == nil {
		t.Error("OpCertResult should not be nil")
	}
	if mock.SubmitTxResult == nil {
		t.Error("SubmitTxResult should not be nil")
	}

	// Verify all errors are nil
	if mock.TipError != nil {
		t.Error("TipError should be nil")
	}
	if mock.UTxOError != nil {
		t.Error("UTxOError should be nil")
	}

	// Verify call counts start at 0
	if mock.QueryTipCalled != 0 {
		t.Errorf("QueryTipCalled = %v, want 0", mock.QueryTipCalled)
	}
}
