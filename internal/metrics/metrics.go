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
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	// MetricsNamespace is the namespace for all Cardano operator metrics
	MetricsNamespace = "cardano"

	// MetricsSubsystemStakePool is the subsystem for stake pool metrics
	MetricsSubsystemStakePool = "stakepool"

	// MetricsSubsystemNode is the subsystem for node metrics
	MetricsSubsystemNode = "node"

	// MetricsSubsystemKES is the subsystem for KES metrics
	MetricsSubsystemKES = "kes"
)

var (
	// StakePoolBlocksMinted tracks the total number of blocks minted by a stake pool
	StakePoolBlocksMinted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: MetricsNamespace,
			Subsystem: MetricsSubsystemStakePool,
			Name:      "blocks_minted_total",
			Help:      "Total number of blocks minted by the stake pool",
		},
		[]string{"pool_id", "namespace", "name"},
	)

	// StakePoolKESExpiryEpoch tracks the epoch when KES keys will expire
	StakePoolKESExpiryEpoch = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: MetricsNamespace,
			Subsystem: MetricsSubsystemStakePool,
			Name:      "kes_expiry_epoch",
			Help:      "Epoch when the KES keys will expire",
		},
		[]string{"pool_id", "namespace", "name"},
	)

	// StakePoolKESPeriodsRemaining tracks the number of KES periods remaining
	StakePoolKESPeriodsRemaining = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: MetricsNamespace,
			Subsystem: MetricsSubsystemStakePool,
			Name:      "kes_periods_remaining",
			Help:      "Number of KES periods remaining before key expiration",
		},
		[]string{"pool_id", "namespace", "name"},
	)

	// StakePoolLiveStake tracks the live stake delegated to a pool
	StakePoolLiveStake = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: MetricsNamespace,
			Subsystem: MetricsSubsystemStakePool,
			Name:      "live_stake_lovelace",
			Help:      "Live stake delegated to the pool in lovelace",
		},
		[]string{"pool_id", "namespace", "name"},
	)

	// StakePoolDelegators tracks the number of delegators
	StakePoolDelegators = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: MetricsNamespace,
			Subsystem: MetricsSubsystemStakePool,
			Name:      "delegators_total",
			Help:      "Total number of delegators to the pool",
		},
		[]string{"pool_id", "namespace", "name"},
	)

	// StakePoolPhase tracks the current phase of the stake pool
	StakePoolPhase = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: MetricsNamespace,
			Subsystem: MetricsSubsystemStakePool,
			Name:      "phase",
			Help:      "Current phase of the stake pool (0=Pending, 1=Provisioning, 2=Syncing, 3=Registering, 4=Active, 5=Retiring, 6=Retired, 7=Failed)",
		},
		[]string{"namespace", "name"},
	)

	// NodeSyncProgress tracks the sync progress of a Cardano node (0-100%)
	NodeSyncProgress = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: MetricsNamespace,
			Subsystem: MetricsSubsystemNode,
			Name:      "sync_progress_percent",
			Help:      "Sync progress of the Cardano node as a percentage (0-100)",
		},
		[]string{"node_type", "namespace", "name"},
	)

	// NodePeerCount tracks the number of connected peers
	NodePeerCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: MetricsNamespace,
			Subsystem: MetricsSubsystemNode,
			Name:      "peer_count",
			Help:      "Number of connected peers",
		},
		[]string{"node_type", "namespace", "name"},
	)

	// NodeTipSlot tracks the current tip slot
	NodeTipSlot = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: MetricsNamespace,
			Subsystem: MetricsSubsystemNode,
			Name:      "tip_slot",
			Help:      "Current tip slot of the node",
		},
		[]string{"node_type", "namespace", "name"},
	)

	// NodeTipEpoch tracks the current tip epoch
	NodeTipEpoch = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: MetricsNamespace,
			Subsystem: MetricsSubsystemNode,
			Name:      "tip_epoch",
			Help:      "Current tip epoch of the node",
		},
		[]string{"node_type", "namespace", "name"},
	)

	// NodeTipBlockHeight tracks the current block height
	NodeTipBlockHeight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: MetricsNamespace,
			Subsystem: MetricsSubsystemNode,
			Name:      "tip_block_height",
			Help:      "Current block height of the node",
		},
		[]string{"node_type", "namespace", "name"},
	)

	// NodeReady tracks whether a node is ready (1) or not (0)
	NodeReady = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: MetricsNamespace,
			Subsystem: MetricsSubsystemNode,
			Name:      "ready",
			Help:      "Whether the node is ready (1) or not (0)",
		},
		[]string{"node_type", "namespace", "name"},
	)

	// KESRotationsTotal tracks the total number of KES rotations performed
	KESRotationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: MetricsNamespace,
			Subsystem: MetricsSubsystemKES,
			Name:      "rotations_total",
			Help:      "Total number of KES key rotations performed",
		},
		[]string{"pool_id", "namespace", "name"},
	)

	// KESRotationFailuresTotal tracks the total number of failed KES rotations
	KESRotationFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: MetricsNamespace,
			Subsystem: MetricsSubsystemKES,
			Name:      "rotation_failures_total",
			Help:      "Total number of failed KES key rotations",
		},
		[]string{"pool_id", "namespace", "name"},
	)

	// KESCurrentCounter tracks the current KES counter value
	KESCurrentCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: MetricsNamespace,
			Subsystem: MetricsSubsystemKES,
			Name:      "current_counter",
			Help:      "Current KES counter value",
		},
		[]string{"pool_id", "namespace", "name"},
	)

	// ReconcileTotal tracks the total number of reconciliations
	ReconcileTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: MetricsNamespace,
			Name:      "reconcile_total",
			Help:      "Total number of reconciliations by controller",
		},
		[]string{"controller", "result"},
	)

	// ReconcileDuration tracks the duration of reconciliations
	ReconcileDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: MetricsNamespace,
			Name:      "reconcile_duration_seconds",
			Help:      "Duration of reconciliations in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms to ~10s
		},
		[]string{"controller"},
	)

	// TransactionsSubmitted tracks the total number of transactions submitted
	TransactionsSubmitted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: MetricsNamespace,
			Name:      "transactions_submitted_total",
			Help:      "Total number of transactions submitted to the blockchain",
		},
		[]string{"type", "namespace", "name"},
	)

	// TransactionFailures tracks the total number of failed transactions
	TransactionFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: MetricsNamespace,
			Name:      "transaction_failures_total",
			Help:      "Total number of failed transaction submissions",
		},
		[]string{"type", "namespace", "name", "reason"},
	)
)

func init() {
	// Register all metrics with the controller-runtime metrics registry
	metrics.Registry.MustRegister(
		// Stake pool metrics
		StakePoolBlocksMinted,
		StakePoolKESExpiryEpoch,
		StakePoolKESPeriodsRemaining,
		StakePoolLiveStake,
		StakePoolDelegators,
		StakePoolPhase,

		// Node metrics
		NodeSyncProgress,
		NodePeerCount,
		NodeTipSlot,
		NodeTipEpoch,
		NodeTipBlockHeight,
		NodeReady,

		// KES metrics
		KESRotationsTotal,
		KESRotationFailuresTotal,
		KESCurrentCounter,

		// Operator metrics
		ReconcileTotal,
		ReconcileDuration,
		TransactionsSubmitted,
		TransactionFailures,
	)
}

// PhaseToMetricValue converts a StakePool phase string to a numeric value for metrics
func PhaseToMetricValue(phase string) float64 {
	switch phase {
	case "Pending":
		return 0
	case "Provisioning":
		return 1
	case "Syncing":
		return 2
	case "Registering":
		return 3
	case "Active":
		return 4
	case "Retiring":
		return 5
	case "Retired":
		return 6
	case "Failed":
		return 7
	default:
		return -1
	}
}

// RecordReconcileMetrics records metrics for a reconciliation
func RecordReconcileMetrics(controller string, result string, duration float64) {
	ReconcileTotal.WithLabelValues(controller, result).Inc()
	ReconcileDuration.WithLabelValues(controller).Observe(duration)
}

// RecordStakePoolMetrics records all stake pool metrics
func RecordStakePoolMetrics(poolID, namespace, name string, blocksMinted int64, kesExpiryEpoch int64, kesPeriodsRemaining int64) {
	labels := prometheus.Labels{
		"pool_id":   poolID,
		"namespace": namespace,
		"name":      name,
	}

	// Note: Counter values should only increase, so we use Set with the current total
	// The controller should track the cumulative count
	StakePoolKESExpiryEpoch.With(labels).Set(float64(kesExpiryEpoch))
	StakePoolKESPeriodsRemaining.With(labels).Set(float64(kesPeriodsRemaining))
}

// RecordNodeMetrics records all node metrics
func RecordNodeMetrics(nodeType, namespace, name string, syncProgress float64, peerCount int, tipSlot, tipEpoch, blockHeight int64, ready bool) {
	labels := prometheus.Labels{
		"node_type": nodeType,
		"namespace": namespace,
		"name":      name,
	}

	NodeSyncProgress.With(labels).Set(syncProgress)
	NodePeerCount.With(labels).Set(float64(peerCount))
	NodeTipSlot.With(labels).Set(float64(tipSlot))
	NodeTipEpoch.With(labels).Set(float64(tipEpoch))
	NodeTipBlockHeight.With(labels).Set(float64(blockHeight))

	readyValue := 0.0
	if ready {
		readyValue = 1.0
	}
	NodeReady.With(labels).Set(readyValue)
}

// DeleteStakePoolMetrics removes metrics for a deleted stake pool
func DeleteStakePoolMetrics(poolID, namespace, name string) {
	labels := prometheus.Labels{
		"pool_id":   poolID,
		"namespace": namespace,
		"name":      name,
	}

	StakePoolBlocksMinted.Delete(labels)
	StakePoolKESExpiryEpoch.Delete(labels)
	StakePoolKESPeriodsRemaining.Delete(labels)
	StakePoolLiveStake.Delete(labels)
	StakePoolDelegators.Delete(labels)
	KESRotationsTotal.Delete(labels)
	KESRotationFailuresTotal.Delete(labels)
	KESCurrentCounter.Delete(labels)

	// Delete phase metric (different labels)
	StakePoolPhase.Delete(prometheus.Labels{
		"namespace": namespace,
		"name":      name,
	})
}

// DeleteNodeMetrics removes metrics for a deleted node
func DeleteNodeMetrics(nodeType, namespace, name string) {
	labels := prometheus.Labels{
		"node_type": nodeType,
		"namespace": namespace,
		"name":      name,
	}

	NodeSyncProgress.Delete(labels)
	NodePeerCount.Delete(labels)
	NodeTipSlot.Delete(labels)
	NodeTipEpoch.Delete(labels)
	NodeTipBlockHeight.Delete(labels)
	NodeReady.Delete(labels)
}
