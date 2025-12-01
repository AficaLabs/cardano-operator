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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RotationEvent records a KES rotation event
type RotationEvent struct {
	// KES period that was rotated
	KESPeriod int64 `json:"kesPeriod"`

	// Rotation timestamp
	RotatedAt metav1.Time `json:"rotatedAt"`

	// Expiry epoch of new KES
	ExpiryEpoch int64 `json:"expiryEpoch"`

	// Certificate update transaction hash
	// +optional
	CertificateTxHash string `json:"certificateTxHash,omitempty"`

	// Whether rotation succeeded
	Success bool `json:"success"`

	// Error if failed
	// +optional
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// KESRotationSpec defines the desired state of KESRotation
type KESRotationSpec struct {
	// Associated StakePool name
	StakePoolRef string `json:"stakePoolRef"`

	// Enable automatic rotation (default: true)
	// +kubebuilder:default=true
	// +optional
	AutoRotate bool `json:"autoRotate,omitempty"`

	// Epochs before expiry to trigger rotation (default: 2)
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	// +kubebuilder:default=2
	// +optional
	RotationLeadEpochs int32 `json:"rotationLeadEpochs,omitempty"`

	// Max KES evolutions (protocol param, usually 62)
	// +kubebuilder:default=62
	// +optional
	MaxKESEvolutions int32 `json:"maxKESEvolutions,omitempty"`
}

// KESRotationStatus defines the observed state of KESRotation
type KESRotationStatus struct {
	// Current KES period number
	// +optional
	CurrentKESPeriod int64 `json:"currentKESPeriod,omitempty"`

	// Slot when current KES became active
	// +optional
	KESStartSlot int64 `json:"kesStartSlot,omitempty"`

	// Epoch when current KES expires
	// +optional
	KESExpiryEpoch int64 `json:"kesExpiryEpoch,omitempty"`

	// Planned rotation epoch
	// +optional
	NextRotationEpoch int64 `json:"nextRotationEpoch,omitempty"`

	// Recent rotation events (last 10)
	// +optional
	RotationHistory []RotationEvent `json:"rotationHistory,omitempty"`

	// Standard conditions
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=kesr,categories=cardano
// +kubebuilder:printcolumn:name="StakePool",type=string,JSONPath=`.spec.stakePoolRef`
// +kubebuilder:printcolumn:name="KES Period",type=integer,JSONPath=`.status.currentKESPeriod`
// +kubebuilder:printcolumn:name="Expiry Epoch",type=integer,JSONPath=`.status.kesExpiryEpoch`
// +kubebuilder:printcolumn:name="Next Rotation",type=integer,JSONPath=`.status.nextRotationEpoch`
// +kubebuilder:printcolumn:name="Auto",type=boolean,JSONPath=`.spec.autoRotate`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// KESRotation is the Schema for the kesrotations API
type KESRotation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KESRotationSpec   `json:"spec,omitempty"`
	Status KESRotationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KESRotationList contains a list of KESRotation
type KESRotationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KESRotation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KESRotation{}, &KESRotationList{})
}

// Condition types for KESRotation
const (
	// KESRotationConditionReady indicates the KES key is valid and rotation is configured
	KESRotationConditionReady = "Ready"
	// KESRotationConditionRotationPending indicates a rotation is scheduled
	KESRotationConditionRotationPending = "RotationPending"
	// KESRotationConditionRotationInProgress indicates rotation is in progress
	KESRotationConditionRotationInProgress = "RotationInProgress"
)
