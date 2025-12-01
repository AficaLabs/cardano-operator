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

package utils

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cardanov1alpha1 "github.com/AficaLabs/cardano-operator/api/v1alpha1"
)

// StakePoolOption is a function that modifies a StakePool
type StakePoolOption func(*cardanov1alpha1.StakePool)

// NewTestStakePool creates a valid StakePool for testing
func NewTestStakePool(name, namespace string, opts ...StakePoolOption) *cardanov1alpha1.StakePool {
	sp := &cardanov1alpha1.StakePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: cardanov1alpha1.StakePoolSpec{
			Network: cardanov1alpha1.NetworkPreprod,
			PoolParams: cardanov1alpha1.PoolParams{
				Pledge: "500000000000", // 500,000 ADA
				Margin: "0.03",         // 3%
				Cost:   "340000000",    // 340 ADA
				Metadata: cardanov1alpha1.PoolMetadata{
					URL:  "https://example.com/pool.json",
					Hash: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
				},
				Relays: []cardanov1alpha1.RelayConfig{
					{
						Type:     cardanov1alpha1.RelayTypeDNS,
						Hostname: "relay1.example.com",
						Port:     6000,
					},
					{
						Type:     cardanov1alpha1.RelayTypeDNS,
						Hostname: "relay2.example.com",
						Port:     6000,
					},
				},
			},
			NodeConfig: cardanov1alpha1.NodeConfig{
				BlockProducer: cardanov1alpha1.NodeSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("2"),
							corev1.ResourceMemory: resource.MustParse("8Gi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("4"),
							corev1.ResourceMemory: resource.MustParse("16Gi"),
						},
					},
				},
				RelayCount: 2,
				RelaySpec: cardanov1alpha1.NodeSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("4Gi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("2"),
							corev1.ResourceMemory: resource.MustParse("8Gi"),
						},
					},
				},
				NodeVersion: "10.1.4",
			},
			KeyManagement: cardanov1alpha1.KeyManagement{
				Mode:                  cardanov1alpha1.KeyManagementModeManaged,
				KESRotationLeadEpochs: 2,
			},
			PaymentConfig: cardanov1alpha1.PaymentConfig{
				Mode: cardanov1alpha1.PaymentModeExternal,
			},
			Storage: cardanov1alpha1.StorageConfig{
				Size: "200Gi",
			},
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(sp)
	}

	return sp
}

// WithNetwork sets the network for a StakePool
func WithNetwork(network cardanov1alpha1.Network) StakePoolOption {
	return func(sp *cardanov1alpha1.StakePool) {
		sp.Spec.Network = network
	}
}

// WithPledge sets the pledge for a StakePool
func WithPledge(pledge string) StakePoolOption {
	return func(sp *cardanov1alpha1.StakePool) {
		sp.Spec.PoolParams.Pledge = pledge
	}
}

// WithMargin sets the margin for a StakePool
func WithMargin(margin string) StakePoolOption {
	return func(sp *cardanov1alpha1.StakePool) {
		sp.Spec.PoolParams.Margin = margin
	}
}

// WithCost sets the cost for a StakePool
func WithCost(cost string) StakePoolOption {
	return func(sp *cardanov1alpha1.StakePool) {
		sp.Spec.PoolParams.Cost = cost
	}
}

// WithRelayCount sets the relay count for a StakePool
func WithRelayCount(count int32) StakePoolOption {
	return func(sp *cardanov1alpha1.StakePool) {
		sp.Spec.NodeConfig.RelayCount = count
	}
}

// WithKeyManagementMode sets the key management mode for a StakePool
func WithKeyManagementMode(mode cardanov1alpha1.KeyManagementMode) StakePoolOption {
	return func(sp *cardanov1alpha1.StakePool) {
		sp.Spec.KeyManagement.Mode = mode
	}
}

// WithExternalSigning sets up external signing for a StakePool
func WithExternalSigning(offlineSigningNodeRef string) StakePoolOption {
	return func(sp *cardanov1alpha1.StakePool) {
		sp.Spec.KeyManagement.Mode = cardanov1alpha1.KeyManagementModeExternal
		sp.Spec.KeyManagement.OfflineSigningNodeRef = offlineSigningNodeRef
	}
}

// WithAutomatedPayment sets up automated payment for a StakePool
func WithAutomatedPayment(secretName string) StakePoolOption {
	return func(sp *cardanov1alpha1.StakePool) {
		sp.Spec.PaymentConfig.Mode = cardanov1alpha1.PaymentModeAutomated
		sp.Spec.PaymentConfig.PaymentKeySecretRef = &cardanov1alpha1.SecretReference{
			Name: secretName,
			Key:  "payment.skey",
		}
	}
}

// WithStorageSize sets the storage size for a StakePool
func WithStorageSize(size string) StakePoolOption {
	return func(sp *cardanov1alpha1.StakePool) {
		sp.Spec.Storage.Size = size
	}
}

// WithStorageClass sets the storage class for a StakePool
func WithStorageClass(className string) StakePoolOption {
	return func(sp *cardanov1alpha1.StakePool) {
		sp.Spec.Storage.StorageClassName = className
	}
}

// WithPhase sets the status phase for a StakePool
func WithPhase(phase cardanov1alpha1.StakePoolPhase) StakePoolOption {
	return func(sp *cardanov1alpha1.StakePool) {
		sp.Status.Phase = phase
	}
}

// WithPoolID sets the pool ID in status
func WithPoolID(poolID string) StakePoolOption {
	return func(sp *cardanov1alpha1.StakePool) {
		sp.Status.PoolID = poolID
	}
}

// CardanoNodeOption is a function that modifies a CardanoNode
type CardanoNodeOption func(*cardanov1alpha1.CardanoNode)

// NewTestCardanoNode creates a valid CardanoNode for testing
func NewTestCardanoNode(name, namespace string, opts ...CardanoNodeOption) *cardanov1alpha1.CardanoNode {
	cn := &cardanov1alpha1.CardanoNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: cardanov1alpha1.CardanoNodeSpec{
			Type:        cardanov1alpha1.CardanoNodeTypeRelay,
			Network:     cardanov1alpha1.NetworkPreprod,
			NodeVersion: "10.1.4",
			Topology: &cardanov1alpha1.TopologyConfig{
				Mode: cardanov1alpha1.TopologyModeP2P,
				P2PConfig: &cardanov1alpha1.P2PConfig{
					TargetNumberOfActivePeers:      20,
					TargetNumberOfEstablishedPeers: 40,
				},
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("4Gi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("2"),
					corev1.ResourceMemory: resource.MustParse("8Gi"),
				},
			},
			Storage: cardanov1alpha1.StorageConfig{
				Size: "200Gi",
			},
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(cn)
	}

	return cn
}

// WithNodeType sets the node type
func WithNodeType(nodeType cardanov1alpha1.CardanoNodeType) CardanoNodeOption {
	return func(cn *cardanov1alpha1.CardanoNode) {
		cn.Spec.Type = nodeType
	}
}

// WithNodeNetwork sets the network for a CardanoNode
func WithNodeNetwork(network cardanov1alpha1.Network) CardanoNodeOption {
	return func(cn *cardanov1alpha1.CardanoNode) {
		cn.Spec.Network = network
	}
}

// WithStakePoolRef sets the stake pool reference
func WithStakePoolRef(ref string) CardanoNodeOption {
	return func(cn *cardanov1alpha1.CardanoNode) {
		cn.Spec.StakePoolRef = ref
	}
}

// WithStaticTopology sets up static topology with peers
func WithStaticTopology(peers []cardanov1alpha1.Peer) CardanoNodeOption {
	return func(cn *cardanov1alpha1.CardanoNode) {
		cn.Spec.Topology = &cardanov1alpha1.TopologyConfig{
			Mode:        cardanov1alpha1.TopologyModeStatic,
			StaticPeers: peers,
		}
	}
}

// WithNodePhase sets the status phase for a CardanoNode
func WithNodePhase(phase cardanov1alpha1.CardanoNodePhase) CardanoNodeOption {
	return func(cn *cardanov1alpha1.CardanoNode) {
		cn.Status.Phase = phase
	}
}

// WithSyncProgress sets the sync progress
func WithSyncProgress(progress string) CardanoNodeOption {
	return func(cn *cardanov1alpha1.CardanoNode) {
		cn.Status.SyncProgress = progress
	}
}

// WithPeerCount sets the peer count
func WithPeerCount(count int32) CardanoNodeOption {
	return func(cn *cardanov1alpha1.CardanoNode) {
		cn.Status.PeerCount = count
	}
}

// KESRotationOption is a function that modifies a KESRotation
type KESRotationOption func(*cardanov1alpha1.KESRotation)

// NewTestKESRotation creates a valid KESRotation for testing
func NewTestKESRotation(name, namespace string, opts ...KESRotationOption) *cardanov1alpha1.KESRotation {
	kr := &cardanov1alpha1.KESRotation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: cardanov1alpha1.KESRotationSpec{
			StakePoolRef:       name + "-pool",
			AutoRotate:         true,
			RotationLeadEpochs: 2,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(kr)
	}

	return kr
}

// WithKESStakePoolRef sets the stake pool reference
func WithKESStakePoolRef(ref string) KESRotationOption {
	return func(kr *cardanov1alpha1.KESRotation) {
		kr.Spec.StakePoolRef = ref
	}
}

// WithAutoRotate sets auto-rotate
func WithAutoRotate(enabled bool) KESRotationOption {
	return func(kr *cardanov1alpha1.KESRotation) {
		kr.Spec.AutoRotate = enabled
	}
}

// WithKESRotationLeadEpochs sets the rotation lead epochs
func WithKESRotationLeadEpochs(epochs int32) KESRotationOption {
	return func(kr *cardanov1alpha1.KESRotation) {
		kr.Spec.RotationLeadEpochs = epochs
	}
}

// WithCurrentKESPeriod sets the current KES period
func WithCurrentKESPeriod(period int64) KESRotationOption {
	return func(kr *cardanov1alpha1.KESRotation) {
		kr.Status.CurrentKESPeriod = period
	}
}

// WithKESExpiryEpoch sets the KES expiry epoch
func WithKESExpiryEpoch(epoch int64) KESRotationOption {
	return func(kr *cardanov1alpha1.KESRotation) {
		kr.Status.KESExpiryEpoch = epoch
	}
}

// OfflineSigningNodeOption is a function that modifies an OfflineSigningNode
type OfflineSigningNodeOption func(*cardanov1alpha1.OfflineSigningNode)

// NewTestOfflineSigningNode creates a valid OfflineSigningNode for testing
func NewTestOfflineSigningNode(
	name, namespace string, opts ...OfflineSigningNodeOption,
) *cardanov1alpha1.OfflineSigningNode {
	osn := &cardanov1alpha1.OfflineSigningNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: cardanov1alpha1.OfflineSigningNodeSpec{
			StakePoolRef: name + "-pool",
			SigningMode:  cardanov1alpha1.SigningModeManual,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(osn)
	}

	return osn
}

// WithOSNStakePoolRef sets the stake pool reference
func WithOSNStakePoolRef(ref string) OfflineSigningNodeOption {
	return func(osn *cardanov1alpha1.OfflineSigningNode) {
		osn.Spec.StakePoolRef = ref
	}
}

// WithSigningMode sets the signing mode
func WithSigningMode(mode cardanov1alpha1.SigningMode) OfflineSigningNodeOption {
	return func(osn *cardanov1alpha1.OfflineSigningNode) {
		osn.Spec.SigningMode = mode
	}
}

// Helper functions for creating common test scenarios

// NewTestStakePoolMainnet creates a mainnet StakePool for testing
func NewTestStakePoolMainnet(name, namespace string) *cardanov1alpha1.StakePool {
	return NewTestStakePool(name, namespace, WithNetwork(cardanov1alpha1.NetworkMainnet))
}

// NewTestBlockProducerNode creates a block producer CardanoNode for testing
func NewTestBlockProducerNode(name, namespace, stakePoolRef string) *cardanov1alpha1.CardanoNode {
	return NewTestCardanoNode(name, namespace,
		WithNodeType(cardanov1alpha1.CardanoNodeTypeBlockProducer),
		WithStakePoolRef(stakePoolRef),
		WithStaticTopology([]cardanov1alpha1.Peer{
			{Address: "relay1.example.com", Port: 6000},
			{Address: "relay2.example.com", Port: 6000},
		}),
	)
}

// NewTestRelayNode creates a relay CardanoNode for testing
func NewTestRelayNode(name, namespace string) *cardanov1alpha1.CardanoNode {
	return NewTestCardanoNode(name, namespace, WithNodeType(cardanov1alpha1.CardanoNodeTypeRelay))
}

// NewTestPaymentSecret creates a Secret containing payment keys for testing
func NewTestPaymentSecret(name, namespace string) *corev1.Secret {
	// nolint:lll // JSON key format requires specific structure
	paymentSKey := `{"type": "PaymentSigningKeyShelley_ed25519", "cborHex": "5820..."}`
	paymentVKey := `{"type": "PaymentVerificationKeyShelley_ed25519", "cborHex": "5820..."}`
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"payment.skey": []byte(paymentSKey),
			"payment.vkey": []byte(paymentVKey),
		},
	}
}

// NewTestColdKeySecret creates a Secret containing cold keys for testing
func NewTestColdKeySecret(name, namespace string) *corev1.Secret {
	// nolint:lll // JSON key format requires specific structure
	coldSKey := `{"type": "StakePoolSigningKey_ed25519", "cborHex": "5820..."}`
	coldVKey := `{"type": "StakePoolVerificationKey_ed25519", "cborHex": "5820..."}`
	coldCounter := `{"type": "NodeOperationalCertificateIssueCounter", "cborHex": "8200..."}`
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"cold.skey":    []byte(coldSKey),
			"cold.vkey":    []byte(coldVKey),
			"cold.counter": []byte(coldCounter),
		},
	}
}

// GenerateUniqueName generates a unique name for test resources
func GenerateUniqueName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, metav1.Now().UnixNano())
}
