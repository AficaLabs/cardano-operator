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

func TestNewUTxOManager(t *testing.T) {
	mock := NewMockClient()
	um := NewUTxOManager(mock)

	if um == nil {
		t.Fatal("NewUTxOManager() returned nil")
	}
	if um.client == nil {
		t.Error("NewUTxOManager() client is nil")
	}
}

func TestUTxOManager_GetBalance(t *testing.T) {
	ctx := context.Background()

	t.Run("returns balance with single UTxO", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{
			{TxHash: "tx1", TxIndex: 0, Value: Value{Lovelace: 5000000}},
		}
		um := NewUTxOManager(mock)

		balance, err := um.GetBalance(ctx, "addr_test1")
		if err != nil {
			t.Errorf("GetBalance() error = %v", err)
		}
		if balance.TotalLovelace != 5000000 {
			t.Errorf("GetBalance().TotalLovelace = %v, want 5000000", balance.TotalLovelace)
		}
		if balance.TotalADA != 5.0 {
			t.Errorf("GetBalance().TotalADA = %v, want 5.0", balance.TotalADA)
		}
		if balance.UTxOCount != 1 {
			t.Errorf("GetBalance().UTxOCount = %v, want 1", balance.UTxOCount)
		}
	})

	t.Run("returns balance with multiple UTxOs", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{
			{TxHash: "tx1", TxIndex: 0, Value: Value{Lovelace: 5000000}},
			{TxHash: "tx2", TxIndex: 0, Value: Value{Lovelace: 3000000}},
			{TxHash: "tx3", TxIndex: 1, Value: Value{Lovelace: 2000000}},
		}
		um := NewUTxOManager(mock)

		balance, err := um.GetBalance(ctx, "addr_test1")
		if err != nil {
			t.Errorf("GetBalance() error = %v", err)
		}
		if balance.TotalLovelace != 10000000 {
			t.Errorf("GetBalance().TotalLovelace = %v, want 10000000", balance.TotalLovelace)
		}
		if balance.UTxOCount != 3 {
			t.Errorf("GetBalance().UTxOCount = %v, want 3", balance.UTxOCount)
		}
	})

	t.Run("returns balance with assets", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{
			{
				TxHash:  "tx1",
				TxIndex: 0,
				Value: Value{
					Lovelace: 5000000,
					Assets:   map[string]int64{"token1": 100, "token2": 50},
				},
			},
			{
				TxHash:  "tx2",
				TxIndex: 0,
				Value: Value{
					Lovelace: 3000000,
					Assets:   map[string]int64{"token1": 50},
				},
			},
		}
		um := NewUTxOManager(mock)

		balance, err := um.GetBalance(ctx, "addr_test1")
		if err != nil {
			t.Errorf("GetBalance() error = %v", err)
		}
		if balance.Assets["token1"] != 150 {
			t.Errorf("GetBalance().Assets[token1] = %v, want 150", balance.Assets["token1"])
		}
		if balance.Assets["token2"] != 50 {
			t.Errorf("GetBalance().Assets[token2] = %v, want 50", balance.Assets["token2"])
		}
	})

	t.Run("returns empty balance for no UTxOs", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{}
		um := NewUTxOManager(mock)

		balance, err := um.GetBalance(ctx, "addr_test1")
		if err != nil {
			t.Errorf("GetBalance() error = %v", err)
		}
		if balance.TotalLovelace != 0 {
			t.Errorf("GetBalance().TotalLovelace = %v, want 0", balance.TotalLovelace)
		}
		if balance.UTxOCount != 0 {
			t.Errorf("GetBalance().UTxOCount = %v, want 0", balance.UTxOCount)
		}
	})

	t.Run("propagates client error", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOError = errors.New("connection failed")
		um := NewUTxOManager(mock)

		_, err := um.GetBalance(ctx, "addr_test1")
		if err == nil {
			t.Error("GetBalance() expected error")
		}
	})
}

func TestUTxOManager_HasSufficientFunds(t *testing.T) {
	ctx := context.Background()

	t.Run("returns true when sufficient", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{
			{TxHash: "tx1", TxIndex: 0, Value: Value{Lovelace: 10000000}},
		}
		um := NewUTxOManager(mock)

		sufficient, err := um.HasSufficientFunds(ctx, "addr_test1", 5000000)
		if err != nil {
			t.Errorf("HasSufficientFunds() error = %v", err)
		}
		if !sufficient {
			t.Error("HasSufficientFunds() = false, want true")
		}
	})

	t.Run("returns false when insufficient", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{
			{TxHash: "tx1", TxIndex: 0, Value: Value{Lovelace: 5000000}},
		}
		um := NewUTxOManager(mock)

		sufficient, err := um.HasSufficientFunds(ctx, "addr_test1", 10000000)
		if err != nil {
			t.Errorf("HasSufficientFunds() error = %v", err)
		}
		if sufficient {
			t.Error("HasSufficientFunds() = true, want false")
		}
	})

	t.Run("returns true when exactly enough", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{
			{TxHash: "tx1", TxIndex: 0, Value: Value{Lovelace: 5000000}},
		}
		um := NewUTxOManager(mock)

		sufficient, err := um.HasSufficientFunds(ctx, "addr_test1", 5000000)
		if err != nil {
			t.Errorf("HasSufficientFunds() error = %v", err)
		}
		if !sufficient {
			t.Error("HasSufficientFunds() = false, want true")
		}
	})

	t.Run("propagates client error", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOError = errors.New("error")
		um := NewUTxOManager(mock)

		_, err := um.HasSufficientFunds(ctx, "addr_test1", 1000)
		if err == nil {
			t.Error("HasSufficientFunds() expected error")
		}
	})
}

func TestUTxOManager_SelectUTxOs(t *testing.T) {
	ctx := context.Background()

	t.Run("selects UTxOs to cover target", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{
			{TxHash: "tx1", TxIndex: 0, Value: Value{Lovelace: 3000000}},
			{TxHash: "tx2", TxIndex: 0, Value: Value{Lovelace: 5000000}},
			{TxHash: "tx3", TxIndex: 0, Value: Value{Lovelace: 2000000}},
		}
		um := NewUTxOManager(mock)

		selected, change, err := um.SelectUTxOs(ctx, "addr_test1", 4000000)
		if err != nil {
			t.Errorf("SelectUTxOs() error = %v", err)
		}
		if len(selected) != 1 {
			t.Errorf("SelectUTxOs() selected %d UTxOs, expected 1 (largest first)", len(selected))
		}
		if change != 1000000 {
			t.Errorf("SelectUTxOs() change = %v, want 1000000", change)
		}
	})

	t.Run("selects multiple UTxOs when needed", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{
			{TxHash: "tx1", TxIndex: 0, Value: Value{Lovelace: 3000000}},
			{TxHash: "tx2", TxIndex: 0, Value: Value{Lovelace: 5000000}},
			{TxHash: "tx3", TxIndex: 0, Value: Value{Lovelace: 2000000}},
		}
		um := NewUTxOManager(mock)

		selected, change, err := um.SelectUTxOs(ctx, "addr_test1", 7000000)
		if err != nil {
			t.Errorf("SelectUTxOs() error = %v", err)
		}
		if len(selected) < 2 {
			t.Errorf("SelectUTxOs() selected %d UTxOs, expected at least 2", len(selected))
		}
		// 5M + 3M = 8M, change = 1M
		if change != 1000000 {
			t.Errorf("SelectUTxOs() change = %v, want 1000000", change)
		}
	})

	t.Run("returns error when no UTxOs", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{}
		um := NewUTxOManager(mock)

		_, _, err := um.SelectUTxOs(ctx, "addr_test1", 1000000)
		if err == nil {
			t.Error("SelectUTxOs() expected error for empty UTxOs")
		}
	})

	t.Run("returns error when insufficient funds", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{
			{TxHash: "tx1", TxIndex: 0, Value: Value{Lovelace: 1000000}},
		}
		um := NewUTxOManager(mock)

		_, _, err := um.SelectUTxOs(ctx, "addr_test1", 5000000)
		if err == nil {
			t.Error("SelectUTxOs() expected error for insufficient funds")
		}
	})

	t.Run("propagates client error", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOError = errors.New("error")
		um := NewUTxOManager(mock)

		_, _, err := um.SelectUTxOs(ctx, "addr_test1", 1000000)
		if err == nil {
			t.Error("SelectUTxOs() expected error")
		}
	})
}

func TestUTxOManager_SelectUTxOsForPoolRegistration(t *testing.T) {
	ctx := context.Background()

	t.Run("selects UTxOs for pool registration", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{
			{TxHash: "tx1", TxIndex: 0, Value: Value{Lovelace: 600000000}}, // 600 ADA
		}
		um := NewUTxOManager(mock)

		poolDeposit := int64(500000000) // 500 ADA
		estimatedFee := int64(200000)   // 0.2 ADA

		selected, change, err := um.SelectUTxOsForPoolRegistration(ctx, "addr_test1", poolDeposit, estimatedFee)
		if err != nil {
			t.Errorf("SelectUTxOsForPoolRegistration() error = %v", err)
		}
		if len(selected) != 1 {
			t.Errorf("SelectUTxOsForPoolRegistration() selected %d UTxOs, want 1", len(selected))
		}
		// 600M - (500M + 0.2M + 2M minimum change) = 97.8M
		expectedChange := int64(600000000 - 500000000 - 200000 - 2000000)
		if change != expectedChange {
			t.Errorf("SelectUTxOsForPoolRegistration() change = %v, want %v", change, expectedChange)
		}
	})
}

func TestUTxOManager_FindLargestUTxO(t *testing.T) {
	ctx := context.Background()

	t.Run("finds largest UTxO", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{
			{TxHash: "tx1", TxIndex: 0, Value: Value{Lovelace: 3000000}},
			{TxHash: "tx2", TxIndex: 0, Value: Value{Lovelace: 8000000}},
			{TxHash: "tx3", TxIndex: 0, Value: Value{Lovelace: 5000000}},
		}
		um := NewUTxOManager(mock)

		largest, err := um.FindLargestUTxO(ctx, "addr_test1")
		if err != nil {
			t.Errorf("FindLargestUTxO() error = %v", err)
		}
		if largest.TxHash != "tx2" {
			t.Errorf("FindLargestUTxO().TxHash = %v, want tx2", largest.TxHash)
		}
		if largest.Value.Lovelace != 8000000 {
			t.Errorf("FindLargestUTxO().Value.Lovelace = %v, want 8000000", largest.Value.Lovelace)
		}
	})

	t.Run("returns error for no UTxOs", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{}
		um := NewUTxOManager(mock)

		_, err := um.FindLargestUTxO(ctx, "addr_test1")
		if err == nil {
			t.Error("FindLargestUTxO() expected error for empty UTxOs")
		}
	})

	t.Run("propagates client error", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOError = errors.New("error")
		um := NewUTxOManager(mock)

		_, err := um.FindLargestUTxO(ctx, "addr_test1")
		if err == nil {
			t.Error("FindLargestUTxO() expected error")
		}
	})
}

func TestUTxOManager_GetUTxOCount(t *testing.T) {
	ctx := context.Background()

	t.Run("returns correct count", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{
			{TxHash: "tx1", TxIndex: 0, Value: Value{Lovelace: 1000000}},
			{TxHash: "tx2", TxIndex: 0, Value: Value{Lovelace: 2000000}},
			{TxHash: "tx3", TxIndex: 0, Value: Value{Lovelace: 3000000}},
		}
		um := NewUTxOManager(mock)

		count, err := um.GetUTxOCount(ctx, "addr_test1")
		if err != nil {
			t.Errorf("GetUTxOCount() error = %v", err)
		}
		if count != 3 {
			t.Errorf("GetUTxOCount() = %v, want 3", count)
		}
	})

	t.Run("returns zero for no UTxOs", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{}
		um := NewUTxOManager(mock)

		count, err := um.GetUTxOCount(ctx, "addr_test1")
		if err != nil {
			t.Errorf("GetUTxOCount() error = %v", err)
		}
		if count != 0 {
			t.Errorf("GetUTxOCount() = %v, want 0", count)
		}
	})

	t.Run("propagates client error", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOError = errors.New("error")
		um := NewUTxOManager(mock)

		_, err := um.GetUTxOCount(ctx, "addr_test1")
		if err == nil {
			t.Error("GetUTxOCount() expected error")
		}
	})
}

func TestUTxOManager_SelectUTxOsWithStrategy(t *testing.T) {
	ctx := context.Background()

	t.Run("LargestFirst strategy", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{
			{TxHash: "tx1", TxIndex: 0, Value: Value{Lovelace: 2000000}},
			{TxHash: "tx2", TxIndex: 0, Value: Value{Lovelace: 5000000}},
			{TxHash: "tx3", TxIndex: 0, Value: Value{Lovelace: 1000000}},
		}
		um := NewUTxOManager(mock)

		selected, _, err := um.SelectUTxOsWithStrategy(ctx, "addr_test1", 4000000, LargestFirst)
		if err != nil {
			t.Errorf("SelectUTxOsWithStrategy() error = %v", err)
		}
		if len(selected) != 1 {
			t.Errorf("SelectUTxOsWithStrategy() selected %d UTxOs, want 1", len(selected))
		}
		if selected[0].TxHash != "tx2" {
			t.Errorf("SelectUTxOsWithStrategy() first selected tx = %v, want tx2", selected[0].TxHash)
		}
	})

	t.Run("SmallestFirst strategy", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{
			{TxHash: "tx1", TxIndex: 0, Value: Value{Lovelace: 2000000}},
			{TxHash: "tx2", TxIndex: 0, Value: Value{Lovelace: 5000000}},
			{TxHash: "tx3", TxIndex: 0, Value: Value{Lovelace: 1000000}},
		}
		um := NewUTxOManager(mock)

		selected, _, err := um.SelectUTxOsWithStrategy(ctx, "addr_test1", 2500000, SmallestFirst)
		if err != nil {
			t.Errorf("SelectUTxOsWithStrategy() error = %v", err)
		}
		// Should select 1M + 2M = 3M >= 2.5M
		if len(selected) != 2 {
			t.Errorf("SelectUTxOsWithStrategy() selected %d UTxOs, want 2", len(selected))
		}
		if selected[0].TxHash != "tx3" {
			t.Errorf("SelectUTxOsWithStrategy() first selected tx = %v, want tx3 (smallest)", selected[0].TxHash)
		}
	})

	t.Run("RandomImprove strategy (falls back to largest first)", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{
			{TxHash: "tx1", TxIndex: 0, Value: Value{Lovelace: 2000000}},
			{TxHash: "tx2", TxIndex: 0, Value: Value{Lovelace: 5000000}},
		}
		um := NewUTxOManager(mock)

		selected, _, err := um.SelectUTxOsWithStrategy(ctx, "addr_test1", 3000000, RandomImprove)
		if err != nil {
			t.Errorf("SelectUTxOsWithStrategy() error = %v", err)
		}
		if len(selected) != 1 {
			t.Errorf("SelectUTxOsWithStrategy() selected %d UTxOs, want 1", len(selected))
		}
	})

	t.Run("returns error for no UTxOs", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{}
		um := NewUTxOManager(mock)

		_, _, err := um.SelectUTxOsWithStrategy(ctx, "addr_test1", 1000000, LargestFirst)
		if err == nil {
			t.Error("SelectUTxOsWithStrategy() expected error for empty UTxOs")
		}
	})

	t.Run("returns error for insufficient funds", func(t *testing.T) {
		mock := NewMockClient()
		mock.UTxOResult = []UTxO{
			{TxHash: "tx1", TxIndex: 0, Value: Value{Lovelace: 1000000}},
		}
		um := NewUTxOManager(mock)

		_, _, err := um.SelectUTxOsWithStrategy(ctx, "addr_test1", 5000000, LargestFirst)
		if err == nil {
			t.Error("SelectUTxOsWithStrategy() expected error for insufficient funds")
		}
	})
}

func TestUTxORef(t *testing.T) {
	ref := UTxORef("abc123", 0)
	if ref != "abc123#0" {
		t.Errorf("UTxORef() = %v, want abc123#0", ref)
	}

	ref = UTxORef("def456", 5)
	if ref != "def456#5" {
		t.Errorf("UTxORef() = %v, want def456#5", ref)
	}
}

func TestLovelaceToADA(t *testing.T) {
	tests := []struct {
		lovelace int64
		want     float64
	}{
		{0, 0},
		{1000000, 1.0},
		{5000000, 5.0},
		{1500000, 1.5},
		{500000000, 500.0},
	}

	for _, tt := range tests {
		got := LovelaceToADA(tt.lovelace)
		if got != tt.want {
			t.Errorf("LovelaceToADA(%d) = %v, want %v", tt.lovelace, got, tt.want)
		}
	}
}

func TestADAToLovelace(t *testing.T) {
	tests := []struct {
		ada  float64
		want int64
	}{
		{0, 0},
		{1.0, 1000000},
		{5.0, 5000000},
		{1.5, 1500000},
		{500.0, 500000000},
	}

	for _, tt := range tests {
		got := ADAToLovelace(tt.ada)
		if got != tt.want {
			t.Errorf("ADAToLovelace(%v) = %d, want %d", tt.ada, got, tt.want)
		}
	}
}

func TestFormatADA(t *testing.T) {
	tests := []struct {
		lovelace int64
		want     string
	}{
		{0, "0.000000 ADA"},
		{1000000, "1.000000 ADA"},
		{5500000, "5.500000 ADA"},
		{500000000, "500.000000 ADA"},
	}

	for _, tt := range tests {
		got := FormatADA(tt.lovelace)
		if got != tt.want {
			t.Errorf("FormatADA(%d) = %v, want %v", tt.lovelace, got, tt.want)
		}
	}
}

func TestMinUTxOValue(t *testing.T) {
	// Without assets
	minNoAssets := MinUTxOValue(false)
	if minNoAssets != 1000000 {
		t.Errorf("MinUTxOValue(false) = %d, want 1000000", minNoAssets)
	}

	// With assets
	minWithAssets := MinUTxOValue(true)
	if minWithAssets != 1500000 {
		t.Errorf("MinUTxOValue(true) = %d, want 1500000", minWithAssets)
	}
}

func TestBalanceInfoStruct(t *testing.T) {
	info := BalanceInfo{
		TotalLovelace: 10000000,
		TotalADA:      10.0,
		UTxOCount:     5,
		Assets:        map[string]int64{"token1": 100},
	}

	if info.TotalLovelace != 10000000 {
		t.Errorf("BalanceInfo.TotalLovelace = %v, want 10000000", info.TotalLovelace)
	}
	if info.TotalADA != 10.0 {
		t.Errorf("BalanceInfo.TotalADA = %v, want 10.0", info.TotalADA)
	}
	if info.UTxOCount != 5 {
		t.Errorf("BalanceInfo.UTxOCount = %v, want 5", info.UTxOCount)
	}
	if info.Assets["token1"] != 100 {
		t.Errorf("BalanceInfo.Assets[token1] = %v, want 100", info.Assets["token1"])
	}
}

func TestCoinSelectionStrategyConstants(t *testing.T) {
	if LargestFirst != 0 {
		t.Errorf("LargestFirst = %v, want 0", LargestFirst)
	}
	if SmallestFirst != 1 {
		t.Errorf("SmallestFirst = %v, want 1", SmallestFirst)
	}
	if RandomImprove != 2 {
		t.Errorf("RandomImprove = %v, want 2", RandomImprove)
	}
}
