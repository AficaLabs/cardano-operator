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
	"time"
)

// EpochInfo contains information about epochs and timing
type EpochInfo struct {
	// Current epoch
	CurrentEpoch int64

	// Current slot within the epoch
	SlotInEpoch int64

	// Slot at the start of the current epoch
	EpochStartSlot int64

	// Estimated time remaining in current epoch
	TimeRemainingInEpoch time.Duration

	// Epoch length in slots
	EpochLength int64

	// Slot length in seconds
	SlotLength float64
}

// KESInfo contains KES-related information
type KESInfo struct {
	// Current KES period
	CurrentKESPeriod int64

	// Slots per KES period
	SlotsPerKESPeriod int64

	// Maximum KES evolutions
	MaxKESEvolutions int64

	// KES period when a key with the given start period will expire
	ExpiryPeriod int64

	// Estimated epoch when KES will expire
	ExpiryEpoch int64
}

// ProtocolTracker provides protocol parameter tracking and epoch monitoring
type ProtocolTracker struct {
	client       Client
	cachedParams *ProtocolParameters
	cachedTip    *Tip
	cacheTime    time.Time
	cacheTTL     time.Duration
}

// NewProtocolTracker creates a new ProtocolTracker
func NewProtocolTracker(client Client) *ProtocolTracker {
	return &ProtocolTracker{
		client:   client,
		cacheTTL: 30 * time.Second, // Cache for 30 seconds
	}
}

// GetProtocolParameters returns cached or fresh protocol parameters
func (pt *ProtocolTracker) GetProtocolParameters(ctx context.Context) (*ProtocolParameters, error) {
	if pt.cachedParams != nil && time.Since(pt.cacheTime) < pt.cacheTTL {
		return pt.cachedParams, nil
	}

	params, err := pt.client.QueryProtocolParameters(ctx)
	if err != nil {
		return nil, err
	}

	pt.cachedParams = params
	pt.cacheTime = time.Now()

	return params, nil
}

// GetCurrentTip returns cached or fresh chain tip
func (pt *ProtocolTracker) GetCurrentTip(ctx context.Context) (*Tip, error) {
	if pt.cachedTip != nil && time.Since(pt.cacheTime) < pt.cacheTTL {
		return pt.cachedTip, nil
	}

	tip, err := pt.client.QueryTip(ctx)
	if err != nil {
		return nil, err
	}

	pt.cachedTip = tip
	pt.cacheTime = time.Now()

	return tip, nil
}

// GetEpochInfo returns detailed epoch information
func (pt *ProtocolTracker) GetEpochInfo(ctx context.Context) (*EpochInfo, error) {
	tip, err := pt.GetCurrentTip(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tip: %w", err)
	}

	// Cardano constants (these vary by network, simplified here)
	var epochLength int64 = 432000  // 5 days in slots (for mainnet)
	var slotLength float64 = 1.0    // 1 second per slot

	// Calculate epoch start slot and slot within epoch
	epochStartSlot := tip.Epoch * epochLength
	slotInEpoch := tip.Slot - epochStartSlot

	// Calculate time remaining
	slotsRemaining := epochLength - slotInEpoch
	timeRemaining := time.Duration(float64(slotsRemaining) * slotLength * float64(time.Second))

	return &EpochInfo{
		CurrentEpoch:         tip.Epoch,
		SlotInEpoch:          slotInEpoch,
		EpochStartSlot:       epochStartSlot,
		TimeRemainingInEpoch: timeRemaining,
		EpochLength:          epochLength,
		SlotLength:           slotLength,
	}, nil
}

// GetKESInfo returns KES-related information
func (pt *ProtocolTracker) GetKESInfo(ctx context.Context, kesStartPeriod int64) (*KESInfo, error) {
	params, err := pt.GetProtocolParameters(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get protocol parameters: %w", err)
	}

	tip, err := pt.GetCurrentTip(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tip: %w", err)
	}

	currentKESPeriod := tip.Slot / params.SlotsPerKESPeriod
	expiryPeriod := kesStartPeriod + params.MaxKESEvolutions

	// Estimate expiry epoch
	expirySlot := expiryPeriod * params.SlotsPerKESPeriod
	// Simplified epoch calculation
	var epochLength int64 = 432000 // slots per epoch (mainnet)
	expiryEpoch := expirySlot / epochLength

	return &KESInfo{
		CurrentKESPeriod:  currentKESPeriod,
		SlotsPerKESPeriod: params.SlotsPerKESPeriod,
		MaxKESEvolutions:  params.MaxKESEvolutions,
		ExpiryPeriod:      expiryPeriod,
		ExpiryEpoch:       expiryEpoch,
	}, nil
}

// ShouldRotateKES checks if KES rotation is needed
func (pt *ProtocolTracker) ShouldRotateKES(ctx context.Context, kesStartPeriod int64, leadEpochs int64) (bool, error) {
	kesInfo, err := pt.GetKESInfo(ctx, kesStartPeriod)
	if err != nil {
		return false, err
	}

	tip, err := pt.GetCurrentTip(ctx)
	if err != nil {
		return false, err
	}

	// Rotate if we're within leadEpochs of expiry
	return tip.Epoch >= (kesInfo.ExpiryEpoch - leadEpochs), nil
}

// GetMinPoolCost returns the minimum pool cost from protocol parameters
func (pt *ProtocolTracker) GetMinPoolCost(ctx context.Context) (int64, error) {
	params, err := pt.GetProtocolParameters(ctx)
	if err != nil {
		return 0, err
	}
	return params.MinPoolCost, nil
}

// GetPoolDeposit returns the pool deposit amount
func (pt *ProtocolTracker) GetPoolDeposit(ctx context.Context) (int64, error) {
	params, err := pt.GetProtocolParameters(ctx)
	if err != nil {
		return 0, err
	}
	return params.PoolDeposit, nil
}

// GetKeyDeposit returns the stake key deposit amount
func (pt *ProtocolTracker) GetKeyDeposit(ctx context.Context) (int64, error) {
	params, err := pt.GetProtocolParameters(ctx)
	if err != nil {
		return 0, err
	}
	return params.KeyDeposit, nil
}

// ValidatePoolCost validates that a pool cost meets minimum requirements
func (pt *ProtocolTracker) ValidatePoolCost(ctx context.Context, cost int64) error {
	minCost, err := pt.GetMinPoolCost(ctx)
	if err != nil {
		return fmt.Errorf("failed to get minimum pool cost: %w", err)
	}

	if cost < minCost {
		return fmt.Errorf("pool cost %d is below minimum %d", cost, minCost)
	}

	return nil
}

// ValidatePoolMargin validates that a pool margin is within valid range
func ValidatePoolMargin(margin float64) error {
	if margin < 0 || margin > 1 {
		return fmt.Errorf("pool margin %.4f must be between 0 and 1", margin)
	}
	return nil
}

// NetworkConstants contains network-specific constants
type NetworkConstants struct {
	NetworkMagic      int64
	EpochLength       int64 // slots
	SlotLength        float64 // seconds
	ActiveSlotCoeff   float64
	SecurityParam     int64
	SlotsPerKESPeriod int64
}

// GetNetworkConstants returns constants for a specific network
func GetNetworkConstants(network Network) *NetworkConstants {
	switch network {
	case NetworkMainnet:
		return &NetworkConstants{
			NetworkMagic:      764824073,
			EpochLength:       432000,
			SlotLength:        1.0,
			ActiveSlotCoeff:   0.05,
			SecurityParam:     2160,
			SlotsPerKESPeriod: 129600,
		}
	case NetworkPreprod:
		return &NetworkConstants{
			NetworkMagic:      1,
			EpochLength:       432000,
			SlotLength:        1.0,
			ActiveSlotCoeff:   0.05,
			SecurityParam:     2160,
			SlotsPerKESPeriod: 129600,
		}
	case NetworkPreview:
		return &NetworkConstants{
			NetworkMagic:      2,
			EpochLength:       86400,  // 1 day
			SlotLength:        1.0,
			ActiveSlotCoeff:   0.05,
			SecurityParam:     432,
			SlotsPerKESPeriod: 129600,
		}
	default:
		return nil
	}
}
