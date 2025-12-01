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

// CardanoNodePhase represents the current phase of a CardanoNode
// +kubebuilder:validation:Enum=Pending;Starting;Syncing;Running;Error
type CardanoNodePhase string

const (
	CardanoNodePhasePending  CardanoNodePhase = "Pending"
	CardanoNodePhaseStarting CardanoNodePhase = "Starting"
	CardanoNodePhaseSyncing  CardanoNodePhase = "Syncing"
	CardanoNodePhaseRunning  CardanoNodePhase = "Running"
	CardanoNodePhaseError    CardanoNodePhase = "Error"
)

// CardanoNodeType represents the type of Cardano node
// +kubebuilder:validation:Enum=block-producer;relay;offline-signing
type CardanoNodeType string

const (
	CardanoNodeTypeBlockProducer   CardanoNodeType = "block-producer"
	CardanoNodeTypeRelay           CardanoNodeType = "relay"
	CardanoNodeTypeOfflineSigning  CardanoNodeType = "offline-signing"
)

// TopologyMode defines the peer topology mode
// +kubebuilder:validation:Enum=static;p2p
type TopologyMode string

const (
	TopologyModeStatic TopologyMode = "static"
	TopologyModeP2P    TopologyMode = "p2p"
)

// Peer defines a static peer configuration
type Peer struct {
	// IP or hostname
	Address string `json:"address"`

	// Port number
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`
}

// P2PConfig defines P2P topology configuration
type P2PConfig struct {
	// Target number of active peers
	// +kubebuilder:default=20
	// +optional
	TargetNumberOfActivePeers int32 `json:"targetNumberOfActivePeers,omitempty"`

	// Target number of established peers
	// +kubebuilder:default=40
	// +optional
	TargetNumberOfEstablishedPeers int32 `json:"targetNumberOfEstablishedPeers,omitempty"`
}

// TopologyConfig defines peer topology configuration
type TopologyConfig struct {
	// Topology mode: static or p2p
	Mode TopologyMode `json:"mode"`

	// Static peer list (if mode=static)
	// +optional
	StaticPeers []Peer `json:"staticPeers,omitempty"`

	// P2P configuration (if mode=p2p)
	// +optional
	P2PConfig *P2PConfig `json:"p2pConfig,omitempty"`
}

// CardanoNodeSpec defines the desired state of CardanoNode
type CardanoNodeSpec struct {
	// Node type: block-producer, relay, or offline-signing
	Type CardanoNodeType `json:"type"`

	// Network: mainnet, preprod, preview
	Network Network `json:"network"`

	// Cardano node version
	// +optional
	NodeVersion string `json:"nodeVersion,omitempty"`

	// Peer topology configuration
	// +optional
	Topology *TopologyConfig `json:"topology,omitempty"`

	// CPU/memory configuration
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// PVC configuration
	Storage StorageConfig `json:"storage"`

	// Owning StakePool name (for BP nodes)
	// +optional
	StakePoolRef string `json:"stakePoolRef,omitempty"`
}

// CardanoNodeStatus defines the observed state of CardanoNode
type CardanoNodeStatus struct {
	// Current phase
	// +optional
	Phase CardanoNodePhase `json:"phase,omitempty"`

	// Sync percentage (e.g., "99.5%")
	// +optional
	SyncProgress string `json:"syncProgress,omitempty"`

	// Current chain tip slot
	// +optional
	TipSlot int64 `json:"tipSlot,omitempty"`

	// Current epoch
	// +optional
	TipEpoch int64 `json:"tipEpoch,omitempty"`

	// Connected peer count
	// +optional
	PeerCount int32 `json:"peerCount,omitempty"`

	// Running node version
	// +optional
	NodeVersion string `json:"nodeVersion,omitempty"`

	// Associated Pod name
	// +optional
	PodName string `json:"podName,omitempty"`

	// Associated Service name
	// +optional
	ServiceName string `json:"serviceName,omitempty"`

	// Standard conditions
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=cn,categories=cardano
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Network",type=string,JSONPath=`.spec.network`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Sync",type=string,JSONPath=`.status.syncProgress`
// +kubebuilder:printcolumn:name="Peers",type=integer,JSONPath=`.status.peerCount`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// CardanoNode is the Schema for the cardanonodes API
type CardanoNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CardanoNodeSpec   `json:"spec,omitempty"`
	Status CardanoNodeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CardanoNodeList contains a list of CardanoNode
type CardanoNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CardanoNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CardanoNode{}, &CardanoNodeList{})
}
