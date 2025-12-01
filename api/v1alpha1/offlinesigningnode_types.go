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

// OfflineSigningNodePhase represents the current phase
// +kubebuilder:validation:Enum=Idle;RequestPending;AwaitingSignature;Processing;Error
type OfflineSigningNodePhase string

const (
	OfflineSigningNodePhaseIdle              OfflineSigningNodePhase = "Idle"
	OfflineSigningNodePhaseRequestPending    OfflineSigningNodePhase = "RequestPending"
	OfflineSigningNodePhaseAwaitingSignature OfflineSigningNodePhase = "AwaitingSignature"
	OfflineSigningNodePhaseProcessing        OfflineSigningNodePhase = "Processing"
	OfflineSigningNodePhaseError             OfflineSigningNodePhase = "Error"
)

// SigningMode defines how signing operations are performed
// +kubebuilder:validation:Enum=manual;isolated
type SigningMode string

const (
	// SigningModeManual means user exports/imports transaction files
	SigningModeManual SigningMode = "manual"
	// SigningModeIsolated means a dedicated namespace with restricted access
	SigningModeIsolated SigningMode = "isolated"
)

// SigningRequestType defines the type of signing request
// +kubebuilder:validation:Enum=registration;update;kes-rotation;retirement
type SigningRequestType string

const (
	SigningRequestTypeRegistration SigningRequestType = "registration"
	SigningRequestTypeUpdate       SigningRequestType = "update"
	SigningRequestTypeKESRotation  SigningRequestType = "kes-rotation"
	SigningRequestTypeRetirement   SigningRequestType = "retirement"
)

// SigningRequestStatus defines the status of a signing request
// +kubebuilder:validation:Enum=pending;awaiting-signature;signed;submitted;confirmed;failed
type SigningRequestStatus string

const (
	SigningRequestStatusPending           SigningRequestStatus = "pending"
	SigningRequestStatusAwaitingSignature SigningRequestStatus = "awaiting-signature"
	SigningRequestStatusSigned            SigningRequestStatus = "signed"
	SigningRequestStatusSubmitted         SigningRequestStatus = "submitted"
	SigningRequestStatusConfirmed         SigningRequestStatus = "confirmed"
	SigningRequestStatusFailed            SigningRequestStatus = "failed"
)

// SigningRequest represents a transaction signing request
type SigningRequest struct {
	// Unique request ID
	ID string `json:"id"`

	// Type of signing request
	Type SigningRequestType `json:"type"`

	// Request creation time
	CreatedAt metav1.Time `json:"createdAt"`

	// Base64-encoded unsigned transaction body
	// +optional
	UnsignedTxBody string `json:"unsignedTxBody,omitempty"`

	// Hash for verification
	// +optional
	UnsignedTxHash string `json:"unsignedTxHash,omitempty"`

	// Base64-encoded signed transaction (user provides)
	// +optional
	SignedTx string `json:"signedTx,omitempty"`

	// Submitted transaction hash
	// +optional
	TxHash string `json:"txHash,omitempty"`

	// Current status
	Status SigningRequestStatus `json:"status"`

	// Error details if failed
	// +optional
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// OfflineSigningNodeSpec defines the desired state of OfflineSigningNode
type OfflineSigningNodeSpec struct {
	// Associated StakePool name
	StakePoolRef string `json:"stakePoolRef"`

	// Signing mode: manual (user exports/imports) or isolated (dedicated namespace)
	SigningMode SigningMode `json:"signingMode"`

	// Days to retain completed requests (default: 30)
	// +kubebuilder:default=30
	// +optional
	RequestRetentionDays int32 `json:"requestRetentionDays,omitempty"`
}

// OfflineSigningNodeStatus defines the observed state of OfflineSigningNode
type OfflineSigningNodeStatus struct {
	// Current phase
	// +optional
	Phase OfflineSigningNodePhase `json:"phase,omitempty"`

	// Queue of pending signing requests
	// +optional
	PendingRequests []SigningRequest `json:"pendingRequests,omitempty"`

	// Recently completed requests
	// +optional
	CompletedRequests []SigningRequest `json:"completedRequests,omitempty"`

	// Last signing activity timestamp
	// +optional
	LastActivity *metav1.Time `json:"lastActivity,omitempty"`

	// Standard conditions
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=osn,categories=cardano
// +kubebuilder:printcolumn:name="StakePool",type=string,JSONPath=`.spec.stakePoolRef`
// +kubebuilder:printcolumn:name="Mode",type=string,JSONPath=`.spec.signingMode`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Pending",type=integer,JSONPath=`.status.pendingRequests`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// OfflineSigningNode is the Schema for the offlinesigningnodes API
type OfflineSigningNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OfflineSigningNodeSpec   `json:"spec,omitempty"`
	Status OfflineSigningNodeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OfflineSigningNodeList contains a list of OfflineSigningNode
type OfflineSigningNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OfflineSigningNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OfflineSigningNode{}, &OfflineSigningNodeList{})
}
