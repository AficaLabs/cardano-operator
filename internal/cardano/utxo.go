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
	"sort"
)

// BalanceInfo contains balance information for an address
type BalanceInfo struct {
	// Total lovelace balance
	TotalLovelace int64

	// Total ADA balance (lovelace / 1_000_000)
	TotalADA float64

	// Number of UTxOs
	UTxOCount int

	// Native assets
	Assets map[string]int64
}

// UTxOManager provides UTxO query and selection operations
type UTxOManager struct {
	client Client
}

// NewUTxOManager creates a new UTxOManager
func NewUTxOManager(client Client) *UTxOManager {
	return &UTxOManager{client: client}
}

// GetBalance returns the balance for an address
func (um *UTxOManager) GetBalance(ctx context.Context, address string) (*BalanceInfo, error) {
	utxos, err := um.client.QueryUTxO(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to query UTxOs: %w", err)
	}

	balance := &BalanceInfo{
		Assets: make(map[string]int64),
	}

	for _, utxo := range utxos {
		balance.TotalLovelace += utxo.Value.Lovelace
		balance.UTxOCount++

		for asset, amount := range utxo.Value.Assets {
			balance.Assets[asset] += amount
		}
	}

	balance.TotalADA = float64(balance.TotalLovelace) / 1_000_000

	return balance, nil
}

// HasSufficientFunds checks if an address has sufficient funds for an amount
func (um *UTxOManager) HasSufficientFunds(ctx context.Context, address string, requiredLovelace int64) (bool, error) {
	balance, err := um.GetBalance(ctx, address)
	if err != nil {
		return false, err
	}

	return balance.TotalLovelace >= requiredLovelace, nil
}

// SelectUTxOs selects UTxOs to cover a target amount
func (um *UTxOManager) SelectUTxOs(ctx context.Context, address string, targetLovelace int64) ([]UTxO, int64, error) {
	utxos, err := um.client.QueryUTxO(ctx, address)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query UTxOs: %w", err)
	}

	if len(utxos) == 0 {
		return nil, 0, fmt.Errorf("no UTxOs found at address")
	}

	// Sort by value descending for efficient selection
	sort.Slice(utxos, func(i, j int) bool {
		return utxos[i].Value.Lovelace > utxos[j].Value.Lovelace
	})

	var selected []UTxO
	var total int64

	// Simple greedy selection
	for _, utxo := range utxos {
		selected = append(selected, utxo)
		total += utxo.Value.Lovelace

		if total >= targetLovelace {
			break
		}
	}

	if total < targetLovelace {
		return nil, 0, fmt.Errorf("insufficient funds: need %d lovelace, have %d", targetLovelace, total)
	}

	change := total - targetLovelace
	return selected, change, nil
}

// SelectUTxOsForPoolRegistration selects UTxOs for pool registration
func (um *UTxOManager) SelectUTxOsForPoolRegistration(ctx context.Context, address string, poolDeposit int64, estimatedFee int64) ([]UTxO, int64, error) {
	// Pool registration requires deposit + fee + minimum ADA for change
	minChangeADA := int64(2_000_000) // 2 ADA minimum for change output
	requiredAmount := poolDeposit + estimatedFee + minChangeADA

	return um.SelectUTxOs(ctx, address, requiredAmount)
}

// FindLargestUTxO finds the largest UTxO at an address
func (um *UTxOManager) FindLargestUTxO(ctx context.Context, address string) (*UTxO, error) {
	utxos, err := um.client.QueryUTxO(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to query UTxOs: %w", err)
	}

	if len(utxos) == 0 {
		return nil, fmt.Errorf("no UTxOs found at address")
	}

	var largest *UTxO
	for i := range utxos {
		if largest == nil || utxos[i].Value.Lovelace > largest.Value.Lovelace {
			largest = &utxos[i]
		}
	}

	return largest, nil
}

// GetUTxOCount returns the number of UTxOs at an address
func (um *UTxOManager) GetUTxOCount(ctx context.Context, address string) (int, error) {
	utxos, err := um.client.QueryUTxO(ctx, address)
	if err != nil {
		return 0, fmt.Errorf("failed to query UTxOs: %w", err)
	}

	return len(utxos), nil
}

// UTxORef creates a reference string for a UTxO
func UTxORef(txHash string, txIndex int) string {
	return fmt.Sprintf("%s#%d", txHash, txIndex)
}

// LovelaceToADA converts lovelace to ADA
func LovelaceToADA(lovelace int64) float64 {
	return float64(lovelace) / 1_000_000
}

// ADAToLovelace converts ADA to lovelace
func ADAToLovelace(ada float64) int64 {
	return int64(ada * 1_000_000)
}

// FormatADA formats an ADA amount for display
func FormatADA(lovelace int64) string {
	ada := LovelaceToADA(lovelace)
	return fmt.Sprintf("%.6f ADA", ada)
}

// MinUTxOValue calculates the minimum UTxO value
// This is a simplified calculation; the actual calculation depends on
// the UTxO's datum and assets
func MinUTxOValue(hasAssets bool) int64 {
	if hasAssets {
		return 1_500_000 // 1.5 ADA with assets
	}
	return 1_000_000 // 1 ADA without assets
}

// CoinSelectionStrategy defines different strategies for coin selection
type CoinSelectionStrategy int

const (
	// LargestFirst selects largest UTxOs first
	LargestFirst CoinSelectionStrategy = iota
	// SmallestFirst selects smallest UTxOs first
	SmallestFirst
	// RandomImprove uses random selection with improvement
	RandomImprove
)

// SelectUTxOsWithStrategy selects UTxOs using a specific strategy
func (um *UTxOManager) SelectUTxOsWithStrategy(ctx context.Context, address string, targetLovelace int64, strategy CoinSelectionStrategy) ([]UTxO, int64, error) {
	utxos, err := um.client.QueryUTxO(ctx, address)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query UTxOs: %w", err)
	}

	if len(utxos) == 0 {
		return nil, 0, fmt.Errorf("no UTxOs found at address")
	}

	// Sort based on strategy
	switch strategy {
	case LargestFirst:
		sort.Slice(utxos, func(i, j int) bool {
			return utxos[i].Value.Lovelace > utxos[j].Value.Lovelace
		})
	case SmallestFirst:
		sort.Slice(utxos, func(i, j int) bool {
			return utxos[i].Value.Lovelace < utxos[j].Value.Lovelace
		})
	case RandomImprove:
		// For now, use largest first as default
		sort.Slice(utxos, func(i, j int) bool {
			return utxos[i].Value.Lovelace > utxos[j].Value.Lovelace
		})
	}

	var selected []UTxO
	var total int64

	for _, utxo := range utxos {
		selected = append(selected, utxo)
		total += utxo.Value.Lovelace

		if total >= targetLovelace {
			break
		}
	}

	if total < targetLovelace {
		return nil, 0, fmt.Errorf("insufficient funds: need %d lovelace, have %d", targetLovelace, total)
	}

	change := total - targetLovelace
	return selected, change, nil
}
