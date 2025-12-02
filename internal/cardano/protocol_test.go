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
	"time"
)

func TestNewProtocolTracker(t *testing.T) {
	mock := NewMockClient()
	pt := NewProtocolTracker(mock)

	if pt == nil {
		t.Fatal("NewProtocolTracker() returned nil")
	}
	if pt.client == nil {
		t.Error("NewProtocolTracker() client is nil")
	}
	if pt.cacheTTL != 30*time.Second {
		t.Errorf("NewProtocolTracker() cacheTTL = %v, want 30s", pt.cacheTTL)
	}
}

func TestProtocolTracker_GetProtocolParameters(t *testing.T) {
	ctx := context.Background()

	t.Run("first call fetches from client", func(t *testing.T) {
		mock := NewMockClient()
		pt := NewProtocolTracker(mock)

		params, err := pt.GetProtocolParameters(ctx)
		if err != nil {
			t.Errorf("GetProtocolParameters() error = %v", err)
		}
		if params.PoolDeposit != 500000000 {
			t.Errorf("GetProtocolParameters().PoolDeposit = %v, want 500000000", params.PoolDeposit)
		}
		if mock.QueryProtocolParametersCalled != 1 {
			t.Errorf("QueryProtocolParametersCalled = %v, want 1", mock.QueryProtocolParametersCalled)
		}
	})

	t.Run("second call uses cache", func(t *testing.T) {
		mock := NewMockClient()
		pt := NewProtocolTracker(mock)

		_, _ = pt.GetProtocolParameters(ctx)
		_, _ = pt.GetProtocolParameters(ctx)
		_, _ = pt.GetProtocolParameters(ctx)

		// Should only have called once due to caching
		if mock.QueryProtocolParametersCalled != 1 {
			t.Errorf("QueryProtocolParametersCalled = %v, want 1 (cache should prevent additional calls)", mock.QueryProtocolParametersCalled)
		}
	})

	t.Run("client error propagates", func(t *testing.T) {
		mock := NewMockClient()
		mock.ProtocolParamsError = errors.New("connection failed")
		pt := NewProtocolTracker(mock)

		_, err := pt.GetProtocolParameters(ctx)
		if err == nil {
			t.Error("GetProtocolParameters() expected error")
		}
	})
}

func TestProtocolTracker_GetCurrentTip(t *testing.T) {
	ctx := context.Background()

	t.Run("first call fetches from client", func(t *testing.T) {
		mock := NewMockClient()
		pt := NewProtocolTracker(mock)

		tip, err := pt.GetCurrentTip(ctx)
		if err != nil {
			t.Errorf("GetCurrentTip() error = %v", err)
		}
		if tip.Slot != 1000000 {
			t.Errorf("GetCurrentTip().Slot = %v, want 1000000", tip.Slot)
		}
		if mock.QueryTipCalled != 1 {
			t.Errorf("QueryTipCalled = %v, want 1", mock.QueryTipCalled)
		}
	})

	t.Run("second call uses cache", func(t *testing.T) {
		mock := NewMockClient()
		pt := NewProtocolTracker(mock)

		_, _ = pt.GetCurrentTip(ctx)
		_, _ = pt.GetCurrentTip(ctx)

		if mock.QueryTipCalled != 1 {
			t.Errorf("QueryTipCalled = %v, want 1 (cache should prevent additional calls)", mock.QueryTipCalled)
		}
	})

	t.Run("client error propagates", func(t *testing.T) {
		mock := NewMockClient()
		mock.TipError = errors.New("node not synced")
		pt := NewProtocolTracker(mock)

		_, err := pt.GetCurrentTip(ctx)
		if err == nil {
			t.Error("GetCurrentTip() expected error")
		}
	})
}

func TestProtocolTracker_GetEpochInfo(t *testing.T) {
	ctx := context.Background()

	t.Run("returns valid epoch info", func(t *testing.T) {
		mock := NewMockClient()
		mock.TipResult = &Tip{
			Slot:  43200100, // epoch 100, slot 100 within epoch
			Epoch: 100,
		}
		pt := NewProtocolTracker(mock)

		info, err := pt.GetEpochInfo(ctx)
		if err != nil {
			t.Errorf("GetEpochInfo() error = %v", err)
		}
		if info.CurrentEpoch != 100 {
			t.Errorf("GetEpochInfo().CurrentEpoch = %v, want 100", info.CurrentEpoch)
		}
		if info.EpochLength != 432000 {
			t.Errorf("GetEpochInfo().EpochLength = %v, want 432000", info.EpochLength)
		}
		if info.SlotLength != 1.0 {
			t.Errorf("GetEpochInfo().SlotLength = %v, want 1.0", info.SlotLength)
		}
	})

	t.Run("client error propagates", func(t *testing.T) {
		mock := NewMockClient()
		mock.TipError = errors.New("connection failed")
		pt := NewProtocolTracker(mock)

		_, err := pt.GetEpochInfo(ctx)
		if err == nil {
			t.Error("GetEpochInfo() expected error")
		}
	})
}

func TestProtocolTracker_GetKESInfo(t *testing.T) {
	ctx := context.Background()

	t.Run("returns valid KES info", func(t *testing.T) {
		mock := NewMockClient()
		mock.TipResult = &Tip{Slot: 12960000, Epoch: 100}
		mock.ProtocolParamsResult = &ProtocolParameters{
			SlotsPerKESPeriod: 129600,
			MaxKESEvolutions:  62,
		}
		pt := NewProtocolTracker(mock)

		info, err := pt.GetKESInfo(ctx, 90)
		if err != nil {
			t.Errorf("GetKESInfo() error = %v", err)
		}
		if info.SlotsPerKESPeriod != 129600 {
			t.Errorf("GetKESInfo().SlotsPerKESPeriod = %v, want 129600", info.SlotsPerKESPeriod)
		}
		if info.MaxKESEvolutions != 62 {
			t.Errorf("GetKESInfo().MaxKESEvolutions = %v, want 62", info.MaxKESEvolutions)
		}
		// Expiry period = start (90) + max evolutions (62) = 152
		if info.ExpiryPeriod != 152 {
			t.Errorf("GetKESInfo().ExpiryPeriod = %v, want 152", info.ExpiryPeriod)
		}
	})

	t.Run("protocol params error propagates", func(t *testing.T) {
		mock := NewMockClient()
		mock.ProtocolParamsError = errors.New("params error")
		pt := NewProtocolTracker(mock)

		_, err := pt.GetKESInfo(ctx, 90)
		if err == nil {
			t.Error("GetKESInfo() expected error")
		}
	})

	t.Run("tip error propagates", func(t *testing.T) {
		mock := NewMockClient()
		mock.TipError = errors.New("tip error")
		pt := NewProtocolTracker(mock)

		_, err := pt.GetKESInfo(ctx, 90)
		if err == nil {
			t.Error("GetKESInfo() expected error")
		}
	})
}

func TestProtocolTracker_ShouldRotateKES(t *testing.T) {
	ctx := context.Background()

	t.Run("should not rotate when far from expiry", func(t *testing.T) {
		mock := NewMockClient()
		mock.TipResult = &Tip{Slot: 12960000, Epoch: 30}
		mock.ProtocolParamsResult = &ProtocolParameters{
			SlotsPerKESPeriod: 129600,
			MaxKESEvolutions:  62,
		}
		pt := NewProtocolTracker(mock)

		// KES start period 90, expiry will be well in the future
		shouldRotate, err := pt.ShouldRotateKES(ctx, 90, 5)
		if err != nil {
			t.Errorf("ShouldRotateKES() error = %v", err)
		}
		if shouldRotate {
			t.Error("ShouldRotateKES() = true, expected false (far from expiry)")
		}
	})

	t.Run("should rotate when near expiry", func(t *testing.T) {
		mock := NewMockClient()
		// Current epoch very close to calculated expiry epoch
		mock.TipResult = &Tip{Slot: 12960000, Epoch: 9999}
		mock.ProtocolParamsResult = &ProtocolParameters{
			SlotsPerKESPeriod: 129600,
			MaxKESEvolutions:  62,
		}
		pt := NewProtocolTracker(mock)

		shouldRotate, err := pt.ShouldRotateKES(ctx, 90, 5)
		if err != nil {
			t.Errorf("ShouldRotateKES() error = %v", err)
		}
		if !shouldRotate {
			t.Error("ShouldRotateKES() = false, expected true (near expiry)")
		}
	})

	t.Run("error propagates", func(t *testing.T) {
		mock := NewMockClient()
		mock.ProtocolParamsError = errors.New("error")
		pt := NewProtocolTracker(mock)

		_, err := pt.ShouldRotateKES(ctx, 90, 5)
		if err == nil {
			t.Error("ShouldRotateKES() expected error")
		}
	})
}

func TestProtocolTracker_GetMinPoolCost(t *testing.T) {
	ctx := context.Background()

	t.Run("returns min pool cost", func(t *testing.T) {
		mock := NewMockClient()
		pt := NewProtocolTracker(mock)

		cost, err := pt.GetMinPoolCost(ctx)
		if err != nil {
			t.Errorf("GetMinPoolCost() error = %v", err)
		}
		if cost != 340000000 {
			t.Errorf("GetMinPoolCost() = %v, want 340000000", cost)
		}
	})

	t.Run("error propagates", func(t *testing.T) {
		mock := NewMockClient()
		mock.ProtocolParamsError = errors.New("error")
		pt := NewProtocolTracker(mock)

		_, err := pt.GetMinPoolCost(ctx)
		if err == nil {
			t.Error("GetMinPoolCost() expected error")
		}
	})
}

func TestProtocolTracker_GetPoolDeposit(t *testing.T) {
	ctx := context.Background()

	t.Run("returns pool deposit", func(t *testing.T) {
		mock := NewMockClient()
		pt := NewProtocolTracker(mock)

		deposit, err := pt.GetPoolDeposit(ctx)
		if err != nil {
			t.Errorf("GetPoolDeposit() error = %v", err)
		}
		if deposit != 500000000 {
			t.Errorf("GetPoolDeposit() = %v, want 500000000", deposit)
		}
	})

	t.Run("error propagates", func(t *testing.T) {
		mock := NewMockClient()
		mock.ProtocolParamsError = errors.New("error")
		pt := NewProtocolTracker(mock)

		_, err := pt.GetPoolDeposit(ctx)
		if err == nil {
			t.Error("GetPoolDeposit() expected error")
		}
	})
}

func TestProtocolTracker_GetKeyDeposit(t *testing.T) {
	ctx := context.Background()

	t.Run("returns key deposit", func(t *testing.T) {
		mock := NewMockClient()
		pt := NewProtocolTracker(mock)

		deposit, err := pt.GetKeyDeposit(ctx)
		if err != nil {
			t.Errorf("GetKeyDeposit() error = %v", err)
		}
		if deposit != 2000000 {
			t.Errorf("GetKeyDeposit() = %v, want 2000000", deposit)
		}
	})

	t.Run("error propagates", func(t *testing.T) {
		mock := NewMockClient()
		mock.ProtocolParamsError = errors.New("error")
		pt := NewProtocolTracker(mock)

		_, err := pt.GetKeyDeposit(ctx)
		if err == nil {
			t.Error("GetKeyDeposit() expected error")
		}
	})
}

func TestProtocolTracker_ValidatePoolCost(t *testing.T) {
	ctx := context.Background()

	t.Run("valid cost above minimum", func(t *testing.T) {
		mock := NewMockClient()
		pt := NewProtocolTracker(mock)

		err := pt.ValidatePoolCost(ctx, 500000000) // 500 ADA, above 340 ADA minimum
		if err != nil {
			t.Errorf("ValidatePoolCost() error = %v, want nil", err)
		}
	})

	t.Run("valid cost at minimum", func(t *testing.T) {
		mock := NewMockClient()
		pt := NewProtocolTracker(mock)

		err := pt.ValidatePoolCost(ctx, 340000000) // exactly minimum
		if err != nil {
			t.Errorf("ValidatePoolCost() error = %v, want nil", err)
		}
	})

	t.Run("invalid cost below minimum", func(t *testing.T) {
		mock := NewMockClient()
		pt := NewProtocolTracker(mock)

		err := pt.ValidatePoolCost(ctx, 100000000) // 100 ADA, below minimum
		if err == nil {
			t.Error("ValidatePoolCost() expected error for cost below minimum")
		}
	})

	t.Run("error when fetching min cost", func(t *testing.T) {
		mock := NewMockClient()
		mock.ProtocolParamsError = errors.New("error")
		pt := NewProtocolTracker(mock)

		err := pt.ValidatePoolCost(ctx, 500000000)
		if err == nil {
			t.Error("ValidatePoolCost() expected error")
		}
	})
}

func TestValidatePoolMargin(t *testing.T) {
	tests := []struct {
		name    string
		margin  float64
		wantErr bool
	}{
		{"valid zero margin", 0, false},
		{"valid small margin", 0.01, false},
		{"valid margin", 0.5, false},
		{"valid max margin", 1.0, false},
		{"invalid negative margin", -0.01, true},
		{"invalid margin above 1", 1.01, true},
		{"invalid large margin", 2.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePoolMargin(tt.margin)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePoolMargin(%v) error = %v, wantErr %v", tt.margin, err, tt.wantErr)
			}
		})
	}
}

func TestGetNetworkConstants(t *testing.T) {
	tests := []struct {
		network   Network
		wantMagic int64
		wantEpoch int64
		wantNil   bool
	}{
		{NetworkMainnet, 764824073, 432000, false},
		{NetworkPreprod, 1, 432000, false},
		{NetworkPreview, 2, 86400, false},
		{Network("invalid"), 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.network), func(t *testing.T) {
			constants := GetNetworkConstants(tt.network)

			if tt.wantNil {
				if constants != nil {
					t.Errorf("GetNetworkConstants(%v) = %v, want nil", tt.network, constants)
				}
				return
			}

			if constants == nil {
				t.Fatalf("GetNetworkConstants(%v) = nil, want non-nil", tt.network)
			}
			if constants.NetworkMagic != tt.wantMagic {
				t.Errorf("GetNetworkConstants(%v).NetworkMagic = %v, want %v", tt.network, constants.NetworkMagic, tt.wantMagic)
			}
			if constants.EpochLength != tt.wantEpoch {
				t.Errorf("GetNetworkConstants(%v).EpochLength = %v, want %v", tt.network, constants.EpochLength, tt.wantEpoch)
			}
		})
	}
}

func TestNetworkConstantsValues(t *testing.T) {
	t.Run("mainnet constants", func(t *testing.T) {
		c := GetNetworkConstants(NetworkMainnet)
		if c.SlotLength != 1.0 {
			t.Errorf("Mainnet SlotLength = %v, want 1.0", c.SlotLength)
		}
		if c.ActiveSlotCoeff != 0.05 {
			t.Errorf("Mainnet ActiveSlotCoeff = %v, want 0.05", c.ActiveSlotCoeff)
		}
		if c.SecurityParam != 2160 {
			t.Errorf("Mainnet SecurityParam = %v, want 2160", c.SecurityParam)
		}
		if c.SlotsPerKESPeriod != 129600 {
			t.Errorf("Mainnet SlotsPerKESPeriod = %v, want 129600", c.SlotsPerKESPeriod)
		}
	})

	t.Run("preprod constants", func(t *testing.T) {
		c := GetNetworkConstants(NetworkPreprod)
		if c.SecurityParam != 2160 {
			t.Errorf("Preprod SecurityParam = %v, want 2160", c.SecurityParam)
		}
	})

	t.Run("preview constants", func(t *testing.T) {
		c := GetNetworkConstants(NetworkPreview)
		if c.EpochLength != 86400 {
			t.Errorf("Preview EpochLength = %v, want 86400", c.EpochLength)
		}
		if c.SecurityParam != 432 {
			t.Errorf("Preview SecurityParam = %v, want 432", c.SecurityParam)
		}
	})
}

func TestEpochInfoStruct(t *testing.T) {
	info := EpochInfo{
		CurrentEpoch:         100,
		SlotInEpoch:          50000,
		EpochStartSlot:       43200000,
		TimeRemainingInEpoch: time.Hour * 24,
		EpochLength:          432000,
		SlotLength:           1.0,
	}

	if info.CurrentEpoch != 100 {
		t.Errorf("EpochInfo.CurrentEpoch = %v, want 100", info.CurrentEpoch)
	}
	if info.SlotInEpoch != 50000 {
		t.Errorf("EpochInfo.SlotInEpoch = %v, want 50000", info.SlotInEpoch)
	}
}

func TestKESInfoStruct(t *testing.T) {
	info := KESInfo{
		CurrentKESPeriod:  100,
		SlotsPerKESPeriod: 129600,
		MaxKESEvolutions:  62,
		ExpiryPeriod:      162,
		ExpiryEpoch:       200,
	}

	if info.CurrentKESPeriod != 100 {
		t.Errorf("KESInfo.CurrentKESPeriod = %v, want 100", info.CurrentKESPeriod)
	}
	if info.ExpiryPeriod != 162 {
		t.Errorf("KESInfo.ExpiryPeriod = %v, want 162", info.ExpiryPeriod)
	}
}
