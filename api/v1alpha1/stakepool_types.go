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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StakePoolPhase represents the current phase of a StakePool
// +kubebuilder:validation:Enum=Pending;Provisioning;Syncing;Registering;Active;Updating;Retiring;Retired;Error
type StakePoolPhase string

const (
	StakePoolPhasePending      StakePoolPhase = "Pending"
	StakePoolPhaseProvisioning StakePoolPhase = "Provisioning"
	StakePoolPhaseSyncing      StakePoolPhase = "Syncing"
	StakePoolPhaseRegistering  StakePoolPhase = "Registering"
	StakePoolPhaseActive       StakePoolPhase = "Active"
	StakePoolPhaseUpdating     StakePoolPhase = "Updating"
	StakePoolPhaseRetiring     StakePoolPhase = "Retiring"
	StakePoolPhaseRetired      StakePoolPhase = "Retired"
	StakePoolPhaseError        StakePoolPhase = "Error"
)

// Network represents the Cardano network
// +kubebuilder:validation:Enum=mainnet;preprod;preview
type Network string

const (
	NetworkMainnet Network = "mainnet"
	NetworkPreprod Network = "preprod"
	NetworkPreview Network = "preview"
)

// KeyManagementMode defines how keys are managed
// +kubebuilder:validation:Enum=managed;external
type KeyManagementMode string

const (
	KeyManagementModeManaged  KeyManagementMode = "managed"
	KeyManagementModeExternal KeyManagementMode = "external"
)

// PaymentMode defines how payments are handled
// +kubebuilder:validation:Enum=automated;external
type PaymentMode string

const (
	PaymentModeAutomated PaymentMode = "automated"
	PaymentModeExternal  PaymentMode = "external"
)

// RelayType defines the type of relay endpoint
// +kubebuilder:validation:Enum=dns;ip
type RelayType string

const (
	RelayTypeDNS RelayType = "dns"
	RelayTypeIP  RelayType = "ip"
)

// PoolParams defines the on-chain pool parameters
type PoolParams struct {
	// Pledge amount in lovelace (e.g., "500000000000")
	// +kubebuilder:validation:Pattern=`^[0-9]+$`
	Pledge string `json:"pledge"`

	// Pool margin as decimal (0.0 to 1.0, e.g., "0.03" for 3%)
	// +kubebuilder:validation:Pattern=`^0(\.[0-9]+)?$|^1(\.0+)?$`
	Margin string `json:"margin"`

	// Fixed cost per epoch in lovelace (min 340 ADA = 340000000)
	// +kubebuilder:validation:Pattern=`^[0-9]+$`
	Cost string `json:"cost"`

	// Pool metadata reference
	Metadata PoolMetadata `json:"metadata"`

	// Relay endpoint declarations (min 1 for testnet, min 2 for production)
	// +kubebuilder:validation:MinItems=1
	Relays []RelayConfig `json:"relays"`

	// Reward account address (defaults to pool's stake address)
	// +optional
	RewardAccount string `json:"rewardAccount,omitempty"`
}

// PoolMetadata defines the pool metadata reference
type PoolMetadata struct {
	// URL to metadata JSON (max 64 chars)
	// +kubebuilder:validation:MaxLength=64
	URL string `json:"url"`

	// Pre-computed metadata hash (auto-computed if omitted)
	// +kubebuilder:validation:Pattern=`^[a-f0-9]{64}$`
	// +optional
	Hash string `json:"hash,omitempty"`
}

// RelayConfig defines a relay endpoint
type RelayConfig struct {
	// Relay type: dns or ip
	Type RelayType `json:"type"`

	// DNS hostname (required if type=dns)
	// +optional
	Hostname string `json:"hostname,omitempty"`

	// IPv4 address (required if type=ip)
	// +optional
	IPv4 string `json:"ipv4,omitempty"`

	// IPv6 address
	// +optional
	IPv6 string `json:"ipv6,omitempty"`

	// Port number
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default=6000
	Port int32 `json:"port"`
}

// NodeConfig defines the node deployment configuration
type NodeConfig struct {
	// Block producer node configuration
	BlockProducer NodeSpec `json:"blockProducer"`

	// Number of relay nodes (min 2 for production)
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=2
	RelayCount int32 `json:"relayCount"`

	// Relay node configuration template
	RelaySpec NodeSpec `json:"relaySpec"`

	// Cardano node version (default: latest stable)
	// +optional
	NodeVersion string `json:"nodeVersion,omitempty"`
}

// NodeSpec defines the configuration for a Cardano node
type NodeSpec struct {
	// CPU/memory requests and limits
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Node selector for scheduling
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Pod tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Pod affinity rules
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
}

// KeyManagement defines key handling mode and references
type KeyManagement struct {
	// Key management mode: managed (operator generates) or external (user provides)
	Mode KeyManagementMode `json:"mode"`

	// Secret containing cold key (if mode=managed for storing, or importing existing)
	// +optional
	ColdKeySecretRef *SecretReference `json:"coldKeySecretRef,omitempty"`

	// OfflineSigningNode name for external signing mode
	// +optional
	OfflineSigningNodeRef string `json:"offlineSigningNodeRef,omitempty"`

	// Epochs before KES expiry to rotate (default: 2)
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	// +kubebuilder:default=2
	// +optional
	KESRotationLeadEpochs int32 `json:"kesRotationLeadEpochs,omitempty"`
}

// SecretReference references a key in a Secret
type SecretReference struct {
	// Name of the Secret
	Name string `json:"name"`

	// Key within the Secret
	// +kubebuilder:default="cold.skey"
	// +optional
	Key string `json:"key,omitempty"`
}

// PaymentConfig defines transaction funding configuration
type PaymentConfig struct {
	// Payment mode: automated or external
	Mode PaymentMode `json:"mode"`

	// Secret with payment signing key (if mode=automated)
	// +optional
	PaymentKeySecretRef *SecretReference `json:"paymentKeySecretRef,omitempty"`

	// Address for funding (if mode=external, for display purposes)
	// +optional
	PaymentAddress string `json:"paymentAddress,omitempty"`
}

// StorageConfig defines PVC configuration for blockchain data
type StorageConfig struct {
	// StorageClass name (default: cluster default)
	// +optional
	StorageClassName string `json:"storageClassName,omitempty"`

	// PVC size (e.g., "200Gi")
	// +kubebuilder:validation:Pattern=`^[0-9]+[KMGTPE]i?$`
	Size string `json:"size"`

	// Access modes (default: ReadWriteOnce)
	// +optional
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`
}

// StakePoolSpec defines the desired state of StakePool
type StakePoolSpec struct {
	// Network identifier: mainnet, preprod, preview
	Network Network `json:"network"`

	// On-chain pool parameters
	PoolParams PoolParams `json:"poolParams"`

	// Node deployment configuration
	NodeConfig NodeConfig `json:"nodeConfig"`

	// Key handling mode and references
	KeyManagement KeyManagement `json:"keyManagement"`

	// Transaction funding configuration
	PaymentConfig PaymentConfig `json:"paymentConfig"`

	// PVC configuration for blockchain data
	Storage StorageConfig `json:"storage"`
}

// NodeStatusInfo holds status information about a CardanoNode
type NodeStatusInfo struct {
	// Node name
	Name string `json:"name"`

	// Node type (block-producer, relay)
	Type string `json:"type"`

	// Current phase
	Phase string `json:"phase"`

	// Sync progress percentage
	SyncProgress string `json:"syncProgress,omitempty"`
}

// StakePoolStatus defines the observed state of StakePool
type StakePoolStatus struct {
	// Current phase of the stake pool
	// +optional
	Phase StakePoolPhase `json:"phase,omitempty"`

	// Bech32 pool ID (pool1...)
	// +optional
	PoolID string `json:"poolId,omitempty"`

	// Hex pool ID
	// +optional
	PoolIDHex string `json:"poolIdHex,omitempty"`

	// Current Cardano epoch
	// +optional
	CurrentEpoch int64 `json:"currentEpoch,omitempty"`

	// Epoch when current KES expires
	// +optional
	KESExpiryEpoch int64 `json:"kesExpiryEpoch,omitempty"`

	// Total blocks minted by this pool
	// +optional
	BlocksMinted int64 `json:"blocksMinted,omitempty"`

	// Slot of last minted block
	// +optional
	LastBlockSlot int64 `json:"lastBlockSlot,omitempty"`

	// Registration transaction hash
	// +optional
	RegistrationTxHash string `json:"registrationTxHash,omitempty"`

	// Status of each CardanoNode
	// +optional
	Nodes []NodeStatusInfo `json:"nodes,omitempty"`

	// Standard Kubernetes conditions
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=sp,categories=cardano
// +kubebuilder:printcolumn:name="Network",type=string,JSONPath=`.spec.network`
// +kubebuilder:printcolumn:name="Pool ID",type=string,JSONPath=`.status.poolId`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="KES Expiry",type=integer,JSONPath=`.status.kesExpiryEpoch`
// +kubebuilder:printcolumn:name="Blocks",type=integer,JSONPath=`.status.blocksMinted`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// StakePool is the Schema for the stakepools API
type StakePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StakePoolSpec   `json:"spec,omitempty"`
	Status StakePoolStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StakePoolList contains a list of StakePool
type StakePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StakePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StakePool{}, &StakePoolList{})
}

// Condition types for StakePool
const (
	// ConditionTypeReady indicates the pool is fully operational
	ConditionTypeReady = "Ready"
	// ConditionTypeNodesProvisioned indicates all nodes are running
	ConditionTypeNodesProvisioned = "NodesProvisioned"
	// ConditionTypeNodesSynced indicates all nodes are synced
	ConditionTypeNodesSynced = "NodesSynced"
	// ConditionTypeKeysReady indicates all keys are available
	ConditionTypeKeysReady = "KeysReady"
	// ConditionTypePoolRegistered indicates pool is registered on-chain
	ConditionTypePoolRegistered = "PoolRegistered"
	// ConditionTypeKESValid indicates current KES is valid
	ConditionTypeKESValid = "KESValid"
)
