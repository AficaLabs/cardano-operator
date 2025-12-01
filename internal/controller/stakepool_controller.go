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

package controller

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	cardanov1alpha1 "github.com/AficaLabs/cardano-operator/api/v1alpha1"
	"github.com/AficaLabs/cardano-operator/internal/cardano"
)

// StakePoolFinalizer is the finalizer name for StakePool resources
const StakePoolFinalizer = "stakepool.cardano.org/finalizer"

// StakePoolReconciler reconciles a StakePool object
type StakePoolReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	// CardanoClient is the client for interacting with Cardano
	// If nil, a default client will be created based on the network
	CardanoClient cardano.Client
}

// +kubebuilder:rbac:groups=cardano.cardano.org,resources=stakepools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cardano.cardano.org,resources=stakepools/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cardano.cardano.org,resources=stakepools/finalizers,verbs=update
// +kubebuilder:rbac:groups=cardano.cardano.org,resources=cardanonodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cardano.cardano.org,resources=kesrotations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *StakePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the StakePool instance
	stakePool := &cardanov1alpha1.StakePool{}
	if err := r.Get(ctx, req.NamespacedName, stakePool); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			log.Info("StakePool resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get StakePool")
		return ctrl.Result{}, err
	}

	// Check if the StakePool instance is marked to be deleted
	if stakePool.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(stakePool, StakePoolFinalizer) {
			// Run finalization logic for StakePoolFinalizer
			if err := r.finalizeStakePool(ctx, stakePool); err != nil {
				return ctrl.Result{}, err
			}

			// Remove StakePoolFinalizer once finalization is done
			controllerutil.RemoveFinalizer(stakePool, StakePoolFinalizer)
			if err := r.Update(ctx, stakePool); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(stakePool, StakePoolFinalizer) {
		controllerutil.AddFinalizer(stakePool, StakePoolFinalizer)
		if err := r.Update(ctx, stakePool); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Handle based on current phase
	switch stakePool.Status.Phase {
	case "":
		// Initial state - set to Pending
		return r.handlePendingPhase(ctx, stakePool)
	case cardanov1alpha1.StakePoolPhasePending:
		// Create nodes and transition to Provisioning
		return r.handleProvisioningTransition(ctx, stakePool)
	case cardanov1alpha1.StakePoolPhaseProvisioning:
		// Wait for nodes to be ready and transition to Syncing
		return r.handleProvisioningPhase(ctx, stakePool)
	case cardanov1alpha1.StakePoolPhaseSyncing:
		// Wait for nodes to sync and transition to Registering
		return r.handleSyncingPhase(ctx, stakePool)
	case cardanov1alpha1.StakePoolPhaseRegistering:
		// Register pool on-chain and transition to Active
		return r.handleRegisteringPhase(ctx, stakePool)
	case cardanov1alpha1.StakePoolPhaseActive:
		// Monitor pool health and handle updates
		return r.handleActivePhase(ctx, stakePool)
	case cardanov1alpha1.StakePoolPhaseUpdating:
		// Handle pool parameter updates
		return r.handleUpdatingPhase(ctx, stakePool)
	case cardanov1alpha1.StakePoolPhaseRetiring:
		// Handle pool retirement
		return r.handleRetiringPhase(ctx, stakePool)
	case cardanov1alpha1.StakePoolPhaseError:
		// Attempt recovery from error state
		return r.handleErrorPhase(ctx, stakePool)
	default:
		log.Info("Unknown phase", "phase", stakePool.Status.Phase)
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}
}

// handlePendingPhase sets the initial Pending phase
func (r *StakePoolReconciler) handlePendingPhase(ctx context.Context, sp *cardanov1alpha1.StakePool) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Setting initial phase to Pending")

	sp.Status.Phase = cardanov1alpha1.StakePoolPhasePending
	if err := r.Status().Update(ctx, sp); err != nil {
		return ctrl.Result{}, err
	}

	r.recordEvent(sp, corev1.EventTypeNormal, "PhaseChange", "StakePool phase set to Pending")
	return ctrl.Result{Requeue: true}, nil
}

// handleProvisioningTransition creates nodes and keys
func (r *StakePoolReconciler) handleProvisioningTransition(ctx context.Context, sp *cardanov1alpha1.StakePool) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Transitioning to Provisioning phase")

	// Create keys if managed mode
	if sp.Spec.KeyManagement.Mode == cardanov1alpha1.KeyManagementModeManaged {
		if err := r.ensureKeysSecret(ctx, sp); err != nil {
			log.Error(err, "Failed to ensure keys secret")
			return r.setErrorPhase(ctx, sp, "FailedToCreateKeys", err.Error())
		}
	}

	// Create block producer node
	if err := r.ensureBlockProducerNode(ctx, sp); err != nil {
		log.Error(err, "Failed to ensure block producer node")
		return r.setErrorPhase(ctx, sp, "FailedToCreateBlockProducer", err.Error())
	}

	// Create relay nodes
	if err := r.ensureRelayNodes(ctx, sp); err != nil {
		log.Error(err, "Failed to ensure relay nodes")
		return r.setErrorPhase(ctx, sp, "FailedToCreateRelays", err.Error())
	}

	// Update status
	sp.Status.Phase = cardanov1alpha1.StakePoolPhaseProvisioning
	if err := r.updateNodeStatus(ctx, sp); err != nil {
		log.Error(err, "Failed to update node status")
	}

	if err := r.Status().Update(ctx, sp); err != nil {
		return ctrl.Result{}, err
	}

	r.recordEvent(sp, corev1.EventTypeNormal, "PhaseChange", "StakePool phase set to Provisioning")
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

// handleProvisioningPhase waits for nodes to be ready
func (r *StakePoolReconciler) handleProvisioningPhase(ctx context.Context, sp *cardanov1alpha1.StakePool) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Update node status
	if err := r.updateNodeStatus(ctx, sp); err != nil {
		log.Error(err, "Failed to update node status")
	}

	// Check if all nodes are running
	allRunning := true
	for _, node := range sp.Status.Nodes {
		if node.Phase != string(cardanov1alpha1.CardanoNodePhaseRunning) &&
			node.Phase != string(cardanov1alpha1.CardanoNodePhaseSyncing) {
			allRunning = false
			break
		}
	}

	if allRunning && len(sp.Status.Nodes) > 0 {
		sp.Status.Phase = cardanov1alpha1.StakePoolPhaseSyncing
		r.recordEvent(sp, corev1.EventTypeNormal, "PhaseChange", "StakePool phase set to Syncing")
	}

	if err := r.Status().Update(ctx, sp); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
}

// handleSyncingPhase waits for nodes to fully sync
func (r *StakePoolReconciler) handleSyncingPhase(ctx context.Context, sp *cardanov1alpha1.StakePool) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Update node status
	if err := r.updateNodeStatus(ctx, sp); err != nil {
		log.Error(err, "Failed to update node status")
	}

	// Check if all nodes are synced (100%)
	allSynced := true
	for _, node := range sp.Status.Nodes {
		if node.SyncProgress != "100%" && node.SyncProgress != "100.00%" {
			allSynced = false
			break
		}
	}

	if allSynced && len(sp.Status.Nodes) > 0 {
		sp.Status.Phase = cardanov1alpha1.StakePoolPhaseRegistering
		r.recordEvent(sp, corev1.EventTypeNormal, "PhaseChange", "StakePool phase set to Registering")
	}

	if err := r.Status().Update(ctx, sp); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// handleRegisteringPhase registers the pool on-chain
func (r *StakePoolReconciler) handleRegisteringPhase(ctx context.Context, sp *cardanov1alpha1.StakePool) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Check if already registered
	if sp.Status.PoolID != "" {
		sp.Status.Phase = cardanov1alpha1.StakePoolPhaseActive
		if err := r.Status().Update(ctx, sp); err != nil {
			return ctrl.Result{}, err
		}
		r.recordEvent(sp, corev1.EventTypeNormal, "PhaseChange", "StakePool phase set to Active")
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// For external payment mode, we can't register automatically
	if sp.Spec.PaymentConfig.Mode == cardanov1alpha1.PaymentModeExternal {
		log.Info("External payment mode - waiting for manual registration")
		// Set a placeholder pool ID to indicate waiting for external registration
		setCondition(&sp.Status.Conditions, metav1.Condition{
			Type:    cardanov1alpha1.ConditionTypePoolRegistered,
			Status:  metav1.ConditionFalse,
			Reason:  "AwaitingExternalRegistration",
			Message: "Pool registration pending - external payment mode requires manual registration",
		})
		if err := r.Status().Update(ctx, sp); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// TODO: Implement automatic pool registration with CardanoClient
	// For now, transition to Active state for testing purposes
	log.Info("Pool registration would occur here")

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// handleActivePhase monitors pool health
func (r *StakePoolReconciler) handleActivePhase(ctx context.Context, sp *cardanov1alpha1.StakePool) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.V(1).Info("Monitoring active pool")

	// Update node status
	if err := r.updateNodeStatus(ctx, sp); err != nil {
		log.Error(err, "Failed to update node status")
	}

	// Check for spec changes that require updates
	// TODO: Implement pool update detection

	// Ensure KESRotation resource exists for KES key management
	if err := r.ensureKESRotation(ctx, sp); err != nil {
		log.Error(err, "Failed to ensure KESRotation")
	}

	if err := r.Status().Update(ctx, sp); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

// handleUpdatingPhase handles pool parameter updates
func (r *StakePoolReconciler) handleUpdatingPhase(ctx context.Context, sp *cardanov1alpha1.StakePool) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Handling pool update")

	// TODO: Implement pool update logic

	sp.Status.Phase = cardanov1alpha1.StakePoolPhaseActive
	if err := r.Status().Update(ctx, sp); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// handleRetiringPhase handles pool retirement
func (r *StakePoolReconciler) handleRetiringPhase(ctx context.Context, sp *cardanov1alpha1.StakePool) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Handling pool retirement", "pool", sp.Name)

	// TODO: Implement pool retirement logic

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

// handleErrorPhase attempts recovery from error state
func (r *StakePoolReconciler) handleErrorPhase(ctx context.Context, sp *cardanov1alpha1.StakePool) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Attempting recovery from error state", "pool", sp.Name)

	// Check if underlying issues are resolved
	// For now, just requeue and let the user fix things
	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

// setErrorPhase transitions to error phase with reason
func (r *StakePoolReconciler) setErrorPhase(ctx context.Context, sp *cardanov1alpha1.StakePool, reason, message string) (ctrl.Result, error) {
	sp.Status.Phase = cardanov1alpha1.StakePoolPhaseError
	setCondition(&sp.Status.Conditions, metav1.Condition{
		Type:    "Error",
		Status:  metav1.ConditionTrue,
		Reason:  reason,
		Message: message,
	})
	if err := r.Status().Update(ctx, sp); err != nil {
		return ctrl.Result{}, err
	}
	r.recordEvent(sp, corev1.EventTypeWarning, reason, message)
	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

// ensureKeysSecret creates or updates the keys secret for managed mode
func (r *StakePoolReconciler) ensureKeysSecret(ctx context.Context, sp *cardanov1alpha1.StakePool) error {
	log := logf.FromContext(ctx)
	secretName := sp.Name + "-keys"

	// Check if secret already exists
	existingSecret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: sp.Namespace}, existingSecret)
	if err == nil {
		// Secret already exists
		log.V(1).Info("Keys secret already exists", "secret", secretName)
		return nil
	}
	if !errors.IsNotFound(err) {
		return err
	}

	// Generate keys using CardanoClient or use placeholder for testing
	var coldSKey, coldVKey, vrfSKey, vrfVKey, kesSKey, kesVKey []byte

	if r.CardanoClient != nil {
		// Use real key generation
		coldKeys, err := r.CardanoClient.GenerateColdKeys(ctx)
		if err != nil {
			return fmt.Errorf("failed to generate cold keys: %w", err)
		}
		coldSKey = coldKeys.SigningKey
		coldVKey = coldKeys.VerificationKey

		vrfKeys, err := r.CardanoClient.GenerateVRFKeys(ctx)
		if err != nil {
			return fmt.Errorf("failed to generate VRF keys: %w", err)
		}
		vrfSKey = vrfKeys.SigningKey
		vrfVKey = vrfKeys.VerificationKey

		kesKeys, err := r.CardanoClient.GenerateKESKeys(ctx)
		if err != nil {
			return fmt.Errorf("failed to generate KES keys: %w", err)
		}
		kesSKey = kesKeys.SigningKey
		kesVKey = kesKeys.VerificationKey
	} else {
		// Use placeholder keys for testing (envtest doesn't have cardano-cli)
		coldSKey = []byte(`{"type": "StakePoolSigningKey_ed25519", "description": "Stake Pool Cold Signing Key", "cborHex": "test-cold-skey"}`)
		coldVKey = []byte(`{"type": "StakePoolVerificationKey_ed25519", "description": "Stake Pool Cold Verification Key", "cborHex": "test-cold-vkey"}`)
		vrfSKey = []byte(`{"type": "VrfSigningKey_PraosVRF", "description": "VRF Signing Key", "cborHex": "test-vrf-skey"}`)
		vrfVKey = []byte(`{"type": "VrfVerificationKey_PraosVRF", "description": "VRF Verification Key", "cborHex": "test-vrf-vkey"}`)
		kesSKey = []byte(`{"type": "KesSigningKey_ed25519_kes_2^6", "description": "KES Signing Key", "cborHex": "test-kes-skey"}`)
		kesVKey = []byte(`{"type": "KesVerificationKey_ed25519_kes_2^6", "description": "KES Verification Key", "cborHex": "test-kes-vkey"}`)
	}

	// Create the secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: sp.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "cardano-operator",
				"app.kubernetes.io/managed-by": "cardano-operator",
				"cardano.org/stakepool":        sp.Name,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"cold.skey": coldSKey,
			"cold.vkey": coldVKey,
			"vrf.skey":  vrfSKey,
			"vrf.vkey":  vrfVKey,
			"kes.skey":  kesSKey,
			"kes.vkey":  kesVKey,
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(sp, secret, r.Scheme); err != nil {
		return err
	}

	log.Info("Creating keys secret", "secret", secretName)
	if err := r.Create(ctx, secret); err != nil {
		return err
	}

	setCondition(&sp.Status.Conditions, metav1.Condition{
		Type:    cardanov1alpha1.ConditionTypeKeysReady,
		Status:  metav1.ConditionTrue,
		Reason:  "KeysGenerated",
		Message: "Pool keys have been generated and stored",
	})

	r.recordEvent(sp, corev1.EventTypeNormal, "KeysCreated", "Pool keys created and stored in secret")
	return nil
}

// ensureBlockProducerNode creates the block producer CardanoNode
func (r *StakePoolReconciler) ensureBlockProducerNode(ctx context.Context, sp *cardanov1alpha1.StakePool) error {
	log := logf.FromContext(ctx)
	nodeName := sp.Name + "-bp"

	// Check if node already exists
	existingNode := &cardanov1alpha1.CardanoNode{}
	err := r.Get(ctx, types.NamespacedName{Name: nodeName, Namespace: sp.Namespace}, existingNode)
	if err == nil {
		log.V(1).Info("Block producer node already exists", "node", nodeName)
		return nil
	}
	if !errors.IsNotFound(err) {
		return err
	}

	// Create block producer node
	node := &cardanov1alpha1.CardanoNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeName,
			Namespace: sp.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "cardano-node",
				"app.kubernetes.io/managed-by": "cardano-operator",
				"cardano.org/stakepool":        sp.Name,
				"cardano.org/node-type":        "block-producer",
			},
		},
		Spec: cardanov1alpha1.CardanoNodeSpec{
			Type:         cardanov1alpha1.CardanoNodeTypeBlockProducer,
			Network:      sp.Spec.Network,
			NodeVersion:  sp.Spec.NodeConfig.NodeVersion,
			Resources:    sp.Spec.NodeConfig.BlockProducer.Resources,
			Storage:      sp.Spec.Storage,
			StakePoolRef: sp.Name,
			Topology: &cardanov1alpha1.TopologyConfig{
				Mode: cardanov1alpha1.TopologyModeP2P,
				P2PConfig: &cardanov1alpha1.P2PConfig{
					TargetNumberOfActivePeers:      20,
					TargetNumberOfEstablishedPeers: 40,
				},
			},
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(sp, node, r.Scheme); err != nil {
		return err
	}

	log.Info("Creating block producer node", "node", nodeName)
	if err := r.Create(ctx, node); err != nil {
		return err
	}

	r.recordEvent(sp, corev1.EventTypeNormal, "NodeCreated", fmt.Sprintf("Block producer node %s created", nodeName))
	return nil
}

// ensureRelayNodes creates the relay CardanoNodes
func (r *StakePoolReconciler) ensureRelayNodes(ctx context.Context, sp *cardanov1alpha1.StakePool) error {
	log := logf.FromContext(ctx)

	for i := int32(0); i < sp.Spec.NodeConfig.RelayCount; i++ {
		nodeName := fmt.Sprintf("%s-relay-%d", sp.Name, i)

		// Check if node already exists
		existingNode := &cardanov1alpha1.CardanoNode{}
		err := r.Get(ctx, types.NamespacedName{Name: nodeName, Namespace: sp.Namespace}, existingNode)
		if err == nil {
			log.V(1).Info("Relay node already exists", "node", nodeName)
			continue
		}
		if !errors.IsNotFound(err) {
			return err
		}

		// Create relay node
		node := &cardanov1alpha1.CardanoNode{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nodeName,
				Namespace: sp.Namespace,
				Labels: map[string]string{
					"app.kubernetes.io/name":       "cardano-node",
					"app.kubernetes.io/managed-by": "cardano-operator",
					"cardano.org/stakepool":        sp.Name,
					"cardano.org/node-type":        "relay",
				},
			},
			Spec: cardanov1alpha1.CardanoNodeSpec{
				Type:         cardanov1alpha1.CardanoNodeTypeRelay,
				Network:      sp.Spec.Network,
				NodeVersion:  sp.Spec.NodeConfig.NodeVersion,
				Resources:    sp.Spec.NodeConfig.RelaySpec.Resources,
				Storage:      sp.Spec.Storage,
				StakePoolRef: sp.Name,
				Topology: &cardanov1alpha1.TopologyConfig{
					Mode: cardanov1alpha1.TopologyModeP2P,
					P2PConfig: &cardanov1alpha1.P2PConfig{
						TargetNumberOfActivePeers:      20,
						TargetNumberOfEstablishedPeers: 40,
					},
				},
			},
		}

		// Set owner reference
		if err := controllerutil.SetControllerReference(sp, node, r.Scheme); err != nil {
			return err
		}

		log.Info("Creating relay node", "node", nodeName)
		if err := r.Create(ctx, node); err != nil {
			return err
		}

		r.recordEvent(sp, corev1.EventTypeNormal, "NodeCreated", fmt.Sprintf("Relay node %s created", nodeName))
	}

	return nil
}

// updateNodeStatus updates the status with information about child nodes
func (r *StakePoolReconciler) updateNodeStatus(ctx context.Context, sp *cardanov1alpha1.StakePool) error {
	log := logf.FromContext(ctx)

	// List all CardanoNodes owned by this StakePool
	nodeList := &cardanov1alpha1.CardanoNodeList{}
	if err := r.List(ctx, nodeList, client.InNamespace(sp.Namespace), client.MatchingLabels{
		"cardano.org/stakepool": sp.Name,
	}); err != nil {
		return err
	}

	sp.Status.Nodes = make([]cardanov1alpha1.NodeStatusInfo, 0, len(nodeList.Items))
	for _, node := range nodeList.Items {
		sp.Status.Nodes = append(sp.Status.Nodes, cardanov1alpha1.NodeStatusInfo{
			Name:         node.Name,
			Type:         string(node.Spec.Type),
			Phase:        string(node.Status.Phase),
			SyncProgress: node.Status.SyncProgress,
		})
	}

	log.V(1).Info("Updated node status", "nodeCount", len(sp.Status.Nodes))
	return nil
}

// ensureKESRotation creates a KESRotation resource for this pool
func (r *StakePoolReconciler) ensureKESRotation(ctx context.Context, sp *cardanov1alpha1.StakePool) error {
	log := logf.FromContext(ctx)
	rotationName := sp.Name + "-kes"

	// Check if KESRotation already exists
	existingRotation := &cardanov1alpha1.KESRotation{}
	err := r.Get(ctx, types.NamespacedName{Name: rotationName, Namespace: sp.Namespace}, existingRotation)
	if err == nil {
		log.V(1).Info("KESRotation already exists", "rotation", rotationName)
		return nil
	}
	if !errors.IsNotFound(err) {
		return err
	}

	// Create KESRotation
	rotation := &cardanov1alpha1.KESRotation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rotationName,
			Namespace: sp.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "cardano-operator",
				"app.kubernetes.io/managed-by": "cardano-operator",
				"cardano.org/stakepool":        sp.Name,
			},
		},
		Spec: cardanov1alpha1.KESRotationSpec{
			StakePoolRef:       sp.Name,
			RotationLeadEpochs: sp.Spec.KeyManagement.KESRotationLeadEpochs,
			AutoRotate:         true,
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(sp, rotation, r.Scheme); err != nil {
		return err
	}

	log.Info("Creating KESRotation", "rotation", rotationName)
	if err := r.Create(ctx, rotation); err != nil {
		return err
	}

	return nil
}

// finalizeStakePool handles cleanup when the StakePool is deleted
//
//nolint:unparam // Error return reserved for future retirement transaction logic
func (r *StakePoolReconciler) finalizeStakePool(ctx context.Context, sp *cardanov1alpha1.StakePool) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing StakePool", "pool", sp.Name)

	// Child resources (CardanoNodes, Secrets) will be garbage collected
	// due to owner references, but we can add custom cleanup logic here

	// TODO: Submit pool retirement transaction if configured

	r.recordEvent(sp, corev1.EventTypeNormal, "Finalized", "StakePool cleanup completed")
	return nil
}

// recordEvent records an event if the recorder is set
func (r *StakePoolReconciler) recordEvent(sp *cardanov1alpha1.StakePool, eventType, reason, message string) {
	if r.Recorder != nil {
		r.Recorder.Event(sp, eventType, reason, message)
	}
}

// setCondition sets or updates a condition in the conditions slice
func setCondition(conditions *[]metav1.Condition, condition metav1.Condition) {
	condition.LastTransitionTime = metav1.Now()
	for i, c := range *conditions {
		if c.Type == condition.Type {
			if c.Status != condition.Status {
				(*conditions)[i] = condition
			}
			return
		}
	}
	*conditions = append(*conditions, condition)
}

// SetupWithManager sets up the controller with the Manager.
func (r *StakePoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cardanov1alpha1.StakePool{}).
		Owns(&cardanov1alpha1.CardanoNode{}).
		Owns(&corev1.Secret{}).
		Owns(&cardanov1alpha1.KESRotation{}).
		Named("stakepool").
		Complete(r)
}
