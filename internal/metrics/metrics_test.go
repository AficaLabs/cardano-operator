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

package metrics

import (
	"testing"
)

func TestPhaseToMetricValue(t *testing.T) {
	tests := []struct {
		phase    string
		expected float64
	}{
		{"Pending", 0},
		{"Provisioning", 1},
		{"Syncing", 2},
		{"Registering", 3},
		{"Active", 4},
		{"Retiring", 5},
		{"Retired", 6},
		{"Failed", 7},
		{"Unknown", -1},
		{"", -1},
	}

	for _, tt := range tests {
		t.Run(tt.phase, func(t *testing.T) {
			result := PhaseToMetricValue(tt.phase)
			if result != tt.expected {
				t.Errorf("PhaseToMetricValue(%q) = %v, want %v", tt.phase, result, tt.expected)
			}
		})
	}
}

func TestRecordReconcileMetrics(t *testing.T) {
	// Test that RecordReconcileMetrics doesn't panic
	t.Run("record success", func(t *testing.T) {
		RecordReconcileMetrics("stakepool", "success", 0.5)
	})

	t.Run("record error", func(t *testing.T) {
		RecordReconcileMetrics("stakepool", "error", 1.0)
	})

	t.Run("record requeue", func(t *testing.T) {
		RecordReconcileMetrics("cardanonode", "requeue", 0.1)
	})

	t.Run("different controllers", func(t *testing.T) {
		RecordReconcileMetrics("kesrotation", "success", 0.2)
		RecordReconcileMetrics("offlinesigningnode", "success", 0.3)
	})
}

func TestRecordStakePoolMetrics(t *testing.T) {
	// Test that RecordStakePoolMetrics doesn't panic
	t.Run("record basic metrics", func(t *testing.T) {
		RecordStakePoolMetrics(
			"pool1abc123",
			"default",
			"my-pool",
			100, // blocksMinted
			500, // kesExpiryEpoch
			10,  // kesPeriodsRemaining
		)
	})

	t.Run("record zero values", func(t *testing.T) {
		RecordStakePoolMetrics(
			"pool1xyz789",
			"default",
			"new-pool",
			0, 0, 0,
		)
	})

	t.Run("record different namespaces", func(t *testing.T) {
		RecordStakePoolMetrics("pool1ns1", "namespace1", "pool1", 10, 100, 5)
		RecordStakePoolMetrics("pool1ns2", "namespace2", "pool2", 20, 200, 8)
	})
}

func TestRecordNodeMetrics(t *testing.T) {
	// Test that RecordNodeMetrics doesn't panic
	t.Run("record relay node metrics", func(t *testing.T) {
		RecordNodeMetrics(
			"relay",
			"default",
			"relay-0",
			99.5,    // syncProgress
			10,      // peerCount
			1000000, // tipSlot
			100,     // tipEpoch
			500000,  // blockHeight
			true,    // ready
		)
	})

	t.Run("record block producer metrics", func(t *testing.T) {
		RecordNodeMetrics(
			"block-producer",
			"default",
			"bp-0",
			100.0,
			5,
			1000001,
			100,
			500001,
			true,
		)
	})

	t.Run("record not ready node", func(t *testing.T) {
		RecordNodeMetrics(
			"relay",
			"default",
			"relay-1",
			50.0,
			2,
			500000,
			50,
			250000,
			false,
		)
	})

	t.Run("record zero values", func(t *testing.T) {
		RecordNodeMetrics(
			"relay",
			"test",
			"new-node",
			0.0, 0, 0, 0, 0, false,
		)
	})
}

func TestDeleteStakePoolMetrics(t *testing.T) {
	// First record some metrics, then delete them
	t.Run("delete after recording", func(t *testing.T) {
		// Record metrics first
		RecordStakePoolMetrics("pool-to-delete", "default", "delete-me", 50, 300, 5)

		// Delete should not panic
		DeleteStakePoolMetrics("pool-to-delete", "default", "delete-me")
	})

	t.Run("delete non-existent (should not panic)", func(t *testing.T) {
		// This should not panic even if metrics don't exist
		DeleteStakePoolMetrics("non-existent", "default", "not-there")
	})
}

func TestDeleteNodeMetrics(t *testing.T) {
	// First record some metrics, then delete them
	t.Run("delete after recording", func(t *testing.T) {
		// Record metrics first
		RecordNodeMetrics("relay", "default", "delete-me", 50.0, 5, 100, 10, 50, true)

		// Delete should not panic
		DeleteNodeMetrics("relay", "default", "delete-me")
	})

	t.Run("delete non-existent (should not panic)", func(t *testing.T) {
		// This should not panic even if metrics don't exist
		DeleteNodeMetrics("relay", "default", "not-there")
	})
}

func TestMetricConstants(t *testing.T) {
	// Verify constants are defined correctly
	if MetricsNamespace != "cardano" {
		t.Errorf("MetricsNamespace = %q, want %q", MetricsNamespace, "cardano")
	}
	if MetricsSubsystemStakePool != "stakepool" {
		t.Errorf("MetricsSubsystemStakePool = %q, want %q", MetricsSubsystemStakePool, "stakepool")
	}
	if MetricsSubsystemNode != "node" {
		t.Errorf("MetricsSubsystemNode = %q, want %q", MetricsSubsystemNode, "node")
	}
	if MetricsSubsystemKES != "kes" {
		t.Errorf("MetricsSubsystemKES = %q, want %q", MetricsSubsystemKES, "kes")
	}
}

func TestMetricsRegistration(t *testing.T) {
	// Test that all metrics are properly initialized (not nil)
	t.Run("stake pool metrics initialized", func(t *testing.T) {
		if StakePoolBlocksMinted == nil {
			t.Error("StakePoolBlocksMinted is nil")
		}
		if StakePoolKESExpiryEpoch == nil {
			t.Error("StakePoolKESExpiryEpoch is nil")
		}
		if StakePoolKESPeriodsRemaining == nil {
			t.Error("StakePoolKESPeriodsRemaining is nil")
		}
		if StakePoolLiveStake == nil {
			t.Error("StakePoolLiveStake is nil")
		}
		if StakePoolDelegators == nil {
			t.Error("StakePoolDelegators is nil")
		}
		if StakePoolPhase == nil {
			t.Error("StakePoolPhase is nil")
		}
	})

	t.Run("node metrics initialized", func(t *testing.T) {
		if NodeSyncProgress == nil {
			t.Error("NodeSyncProgress is nil")
		}
		if NodePeerCount == nil {
			t.Error("NodePeerCount is nil")
		}
		if NodeTipSlot == nil {
			t.Error("NodeTipSlot is nil")
		}
		if NodeTipEpoch == nil {
			t.Error("NodeTipEpoch is nil")
		}
		if NodeTipBlockHeight == nil {
			t.Error("NodeTipBlockHeight is nil")
		}
		if NodeReady == nil {
			t.Error("NodeReady is nil")
		}
	})

	t.Run("KES metrics initialized", func(t *testing.T) {
		if KESRotationsTotal == nil {
			t.Error("KESRotationsTotal is nil")
		}
		if KESRotationFailuresTotal == nil {
			t.Error("KESRotationFailuresTotal is nil")
		}
		if KESCurrentCounter == nil {
			t.Error("KESCurrentCounter is nil")
		}
	})

	t.Run("operator metrics initialized", func(t *testing.T) {
		if ReconcileTotal == nil {
			t.Error("ReconcileTotal is nil")
		}
		if ReconcileDuration == nil {
			t.Error("ReconcileDuration is nil")
		}
		if TransactionsSubmitted == nil {
			t.Error("TransactionsSubmitted is nil")
		}
		if TransactionFailures == nil {
			t.Error("TransactionFailures is nil")
		}
	})
}
