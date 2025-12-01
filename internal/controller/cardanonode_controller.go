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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	cardanov1alpha1 "github.com/AficaLabs/cardano-operator/api/v1alpha1"
)

// CardanoNodeFinalizer is the finalizer name for CardanoNode resources
const CardanoNodeFinalizer = "cardanonode.cardano.org/finalizer"

// Default Cardano node image
const defaultCardanoNodeImage = "ghcr.io/intersectmbo/cardano-node:10.1.4"

// CardanoNodeReconciler reconciles a CardanoNode object
type CardanoNodeReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=cardano.cardano.org,resources=cardanonodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cardano.cardano.org,resources=cardanonodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cardano.cardano.org,resources=cardanonodes/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CardanoNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the CardanoNode instance
	node := &cardanov1alpha1.CardanoNode{}
	if err := r.Get(ctx, req.NamespacedName, node); err != nil {
		if errors.IsNotFound(err) {
			log.Info("CardanoNode resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get CardanoNode")
		return ctrl.Result{}, err
	}

	// Check if the CardanoNode instance is marked to be deleted
	if node.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(node, CardanoNodeFinalizer) {
			// Run finalization logic
			if err := r.finalizeCardanoNode(ctx, node); err != nil {
				return ctrl.Result{}, err
			}

			// Remove finalizer
			controllerutil.RemoveFinalizer(node, CardanoNodeFinalizer)
			if err := r.Update(ctx, node); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(node, CardanoNodeFinalizer) {
		controllerutil.AddFinalizer(node, CardanoNodeFinalizer)
		if err := r.Update(ctx, node); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Handle based on current phase
	switch node.Status.Phase {
	case "":
		return r.handlePendingPhase(ctx, node)
	case cardanov1alpha1.CardanoNodePhasePending:
		return r.handleStartingTransition(ctx, node)
	case cardanov1alpha1.CardanoNodePhaseStarting:
		return r.handleStartingPhase(ctx, node)
	case cardanov1alpha1.CardanoNodePhaseSyncing:
		return r.handleSyncingPhase(ctx, node)
	case cardanov1alpha1.CardanoNodePhaseRunning:
		return r.handleRunningPhase(ctx, node)
	case cardanov1alpha1.CardanoNodePhaseError:
		return r.handleErrorPhase(ctx, node)
	default:
		log.Info("Unknown phase", "phase", node.Status.Phase)
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}
}

// handlePendingPhase sets the initial Pending phase
func (r *CardanoNodeReconciler) handlePendingPhase(ctx context.Context, node *cardanov1alpha1.CardanoNode) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Setting initial phase to Pending")

	node.Status.Phase = cardanov1alpha1.CardanoNodePhasePending
	if err := r.Status().Update(ctx, node); err != nil {
		return ctrl.Result{}, err
	}

	r.recordEvent(node, corev1.EventTypeNormal, "PhaseChange", "CardanoNode phase set to Pending")
	return ctrl.Result{Requeue: true}, nil
}

// handleStartingTransition creates resources and transitions to Starting
func (r *CardanoNodeReconciler) handleStartingTransition(ctx context.Context, node *cardanov1alpha1.CardanoNode) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Creating node resources and transitioning to Starting")

	// Create PVC
	if err := r.ensurePVC(ctx, node); err != nil {
		log.Error(err, "Failed to ensure PVC")
		return r.setErrorPhase(ctx, node, "FailedToCreatePVC", err.Error())
	}

	// Create Service
	if err := r.ensureService(ctx, node); err != nil {
		log.Error(err, "Failed to ensure Service")
		return r.setErrorPhase(ctx, node, "FailedToCreateService", err.Error())
	}

	// Create ConfigMap for topology
	if err := r.ensureConfigMap(ctx, node); err != nil {
		log.Error(err, "Failed to ensure ConfigMap")
		return r.setErrorPhase(ctx, node, "FailedToCreateConfigMap", err.Error())
	}

	// Create Deployment
	if err := r.ensureDeployment(ctx, node); err != nil {
		log.Error(err, "Failed to ensure Deployment")
		return r.setErrorPhase(ctx, node, "FailedToCreateDeployment", err.Error())
	}

	// Create NetworkPolicy for block producer isolation
	if node.Spec.Type == cardanov1alpha1.CardanoNodeTypeBlockProducer {
		if err := r.ensureNetworkPolicy(ctx, node); err != nil {
			log.Error(err, "Failed to ensure NetworkPolicy")
			// Don't fail on NetworkPolicy - it's optional
		}
	}

	// Update status
	node.Status.Phase = cardanov1alpha1.CardanoNodePhaseStarting
	node.Status.ServiceName = node.Name + "-svc"
	if err := r.Status().Update(ctx, node); err != nil {
		return ctrl.Result{}, err
	}

	r.recordEvent(node, corev1.EventTypeNormal, "PhaseChange", "CardanoNode phase set to Starting")
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

// handleStartingPhase waits for the pod to be running
func (r *CardanoNodeReconciler) handleStartingPhase(ctx context.Context, node *cardanov1alpha1.CardanoNode) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Check if deployment is ready
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: node.Name, Namespace: node.Namespace}, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Deployment not found, recreating")
			return r.handleStartingTransition(ctx, node)
		}
		return ctrl.Result{}, err
	}

	// Check if pod is running
	if deployment.Status.ReadyReplicas > 0 {
		node.Status.Phase = cardanov1alpha1.CardanoNodePhaseSyncing
		node.Status.SyncProgress = "0%"
		if err := r.Status().Update(ctx, node); err != nil {
			return ctrl.Result{}, err
		}
		r.recordEvent(node, corev1.EventTypeNormal, "PhaseChange", "CardanoNode phase set to Syncing")
	}

	// Get pod name for status
	if err := r.updatePodStatus(ctx, node); err != nil {
		log.Error(err, "Failed to update pod status")
	}

	if err := r.Status().Update(ctx, node); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
}

// handleSyncingPhase monitors sync progress
func (r *CardanoNodeReconciler) handleSyncingPhase(ctx context.Context, node *cardanov1alpha1.CardanoNode) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Update pod status
	if err := r.updatePodStatus(ctx, node); err != nil {
		log.Error(err, "Failed to update pod status")
	}

	// In a real implementation, we would query the node's sync status via cardano-cli
	// For now, we'll simulate by checking if the pod has been running for a while
	// TODO: Implement actual sync progress monitoring

	// Check deployment health
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: node.Name, Namespace: node.Namespace}, deployment)
	if err != nil {
		return ctrl.Result{}, err
	}

	if deployment.Status.ReadyReplicas == 0 {
		// Pod not ready, go back to Starting
		node.Status.Phase = cardanov1alpha1.CardanoNodePhaseStarting
	}

	if err := r.Status().Update(ctx, node); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// handleRunningPhase monitors the running node
func (r *CardanoNodeReconciler) handleRunningPhase(ctx context.Context, node *cardanov1alpha1.CardanoNode) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.V(1).Info("Monitoring running node")

	// Update pod status
	if err := r.updatePodStatus(ctx, node); err != nil {
		log.Error(err, "Failed to update pod status")
	}

	// Check deployment health
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: node.Name, Namespace: node.Namespace}, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			return r.handleStartingTransition(ctx, node)
		}
		return ctrl.Result{}, err
	}

	if deployment.Status.ReadyReplicas == 0 {
		node.Status.Phase = cardanov1alpha1.CardanoNodePhaseStarting
		r.recordEvent(node, corev1.EventTypeWarning, "PodNotReady", "Node pod is not ready")
	}

	// TODO: Monitor actual node health via cardano-cli

	if err := r.Status().Update(ctx, node); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

// handleErrorPhase attempts recovery from error state
func (r *CardanoNodeReconciler) handleErrorPhase(ctx context.Context, node *cardanov1alpha1.CardanoNode) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Attempting recovery from error state")

	// Try to recover by going back to Pending
	node.Status.Phase = cardanov1alpha1.CardanoNodePhasePending
	if err := r.Status().Update(ctx, node); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

// setErrorPhase transitions to error phase with reason
func (r *CardanoNodeReconciler) setErrorPhase(ctx context.Context, node *cardanov1alpha1.CardanoNode, reason, message string) (ctrl.Result, error) {
	node.Status.Phase = cardanov1alpha1.CardanoNodePhaseError
	setCondition(&node.Status.Conditions, metav1.Condition{
		Type:    "Error",
		Status:  metav1.ConditionTrue,
		Reason:  reason,
		Message: message,
	})
	if err := r.Status().Update(ctx, node); err != nil {
		return ctrl.Result{}, err
	}
	r.recordEvent(node, corev1.EventTypeWarning, reason, message)
	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

// ensurePVC creates or updates the PVC for blockchain data
func (r *CardanoNodeReconciler) ensurePVC(ctx context.Context, node *cardanov1alpha1.CardanoNode) error {
	log := logf.FromContext(ctx)
	pvcName := node.Name + "-data"

	// Check if PVC already exists
	existingPVC := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, types.NamespacedName{Name: pvcName, Namespace: node.Namespace}, existingPVC)
	if err == nil {
		log.V(1).Info("PVC already exists", "pvc", pvcName)
		return nil
	}
	if !errors.IsNotFound(err) {
		return err
	}

	// Parse storage size
	storageSize, err := resource.ParseQuantity(node.Spec.Storage.Size)
	if err != nil {
		return fmt.Errorf("invalid storage size: %w", err)
	}

	// Set access modes
	accessModes := node.Spec.Storage.AccessModes
	if len(accessModes) == 0 {
		accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	}

	// Create PVC
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: node.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "cardano-node",
				"app.kubernetes.io/managed-by": "cardano-operator",
				"cardano.org/node":             node.Name,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: accessModes,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storageSize,
				},
			},
		},
	}

	if node.Spec.Storage.StorageClassName != "" {
		pvc.Spec.StorageClassName = &node.Spec.Storage.StorageClassName
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(node, pvc, r.Scheme); err != nil {
		return err
	}

	log.Info("Creating PVC", "pvc", pvcName)
	return r.Create(ctx, pvc)
}

// ensureService creates or updates the Service
func (r *CardanoNodeReconciler) ensureService(ctx context.Context, node *cardanov1alpha1.CardanoNode) error {
	log := logf.FromContext(ctx)
	svcName := node.Name + "-svc"

	// Check if Service already exists
	existingSvc := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: svcName, Namespace: node.Namespace}, existingSvc)
	if err == nil {
		log.V(1).Info("Service already exists", "service", svcName)
		return nil
	}
	if !errors.IsNotFound(err) {
		return err
	}

	// Create Service
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: node.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "cardano-node",
				"app.kubernetes.io/managed-by": "cardano-operator",
				"cardano.org/node":             node.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app.kubernetes.io/name": "cardano-node",
				"cardano.org/node":       node.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "node",
					Port:       6000,
					TargetPort: intstr.FromInt(6000),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "prometheus",
					Port:       12798,
					TargetPort: intstr.FromInt(12798),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(node, svc, r.Scheme); err != nil {
		return err
	}

	log.Info("Creating Service", "service", svcName)
	return r.Create(ctx, svc)
}

// ensureConfigMap creates or updates the ConfigMap for topology
func (r *CardanoNodeReconciler) ensureConfigMap(ctx context.Context, node *cardanov1alpha1.CardanoNode) error {
	log := logf.FromContext(ctx)
	cmName := node.Name + "-config"

	// Check if ConfigMap already exists
	existingCM := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: cmName, Namespace: node.Namespace}, existingCM)
	if err == nil {
		log.V(1).Info("ConfigMap already exists", "configmap", cmName)
		return nil
	}
	if !errors.IsNotFound(err) {
		return err
	}

	// Generate topology config
	topologyJSON := r.generateTopologyConfig(node)

	// Create ConfigMap
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: node.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "cardano-node",
				"app.kubernetes.io/managed-by": "cardano-operator",
				"cardano.org/node":             node.Name,
			},
		},
		Data: map[string]string{
			"topology.json": topologyJSON,
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(node, cm, r.Scheme); err != nil {
		return err
	}

	log.Info("Creating ConfigMap", "configmap", cmName)
	return r.Create(ctx, cm)
}

// generateTopologyConfig generates the topology JSON for the node
func (r *CardanoNodeReconciler) generateTopologyConfig(node *cardanov1alpha1.CardanoNode) string {
	// Generate P2P topology configuration
	if node.Spec.Topology != nil && node.Spec.Topology.Mode == cardanov1alpha1.TopologyModeP2P {
		targetActive := int32(20)
		targetEstablished := int32(40)
		if node.Spec.Topology.P2PConfig != nil {
			if node.Spec.Topology.P2PConfig.TargetNumberOfActivePeers > 0 {
				targetActive = node.Spec.Topology.P2PConfig.TargetNumberOfActivePeers
			}
			if node.Spec.Topology.P2PConfig.TargetNumberOfEstablishedPeers > 0 {
				targetEstablished = node.Spec.Topology.P2PConfig.TargetNumberOfEstablishedPeers
			}
		}

		// Get public roots based on network
		publicRoots := getPublicRoots(node.Spec.Network)

		return fmt.Sprintf(`{
  "localRoots": [],
  "publicRoots": [%s],
  "useLedgerAfterSlot": -1,
  "bootstrapPeers": null,
  "localHotValency": 0,
  "localWarmValency": 0,
  "targetNumberOfActivePeers": %d,
  "targetNumberOfEstablishedPeers": %d,
  "targetNumberOfKnownPeers": 100,
  "targetNumberOfRootPeers": 100
}`, publicRoots, targetActive, targetEstablished)
	}

	// Static topology
	if node.Spec.Topology != nil && len(node.Spec.Topology.StaticPeers) > 0 {
		peers := ""
		for i, peer := range node.Spec.Topology.StaticPeers {
			if i > 0 {
				peers += ","
			}
			peers += fmt.Sprintf(`{"addr": "%s", "port": %d, "valency": 1}`, peer.Address, peer.Port)
		}
		return fmt.Sprintf(`{"Producers": [%s]}`, peers)
	}

	// Default empty topology
	return `{"Producers": []}`
}

// getPublicRoots returns the public roots for the given network
func getPublicRoots(network cardanov1alpha1.Network) string {
	switch network {
	case cardanov1alpha1.NetworkMainnet:
		return `{"accessPoints": [{"address": "backbone.cardano.iog.io", "port": 3001}], "advertise": false}`
	case cardanov1alpha1.NetworkPreprod:
		return `{"accessPoints": [{"address": "preprod-node.world.dev.cardano.org", "port": 30000}], "advertise": false}`
	case cardanov1alpha1.NetworkPreview:
		return `{"accessPoints": [{"address": "preview-node.world.dev.cardano.org", "port": 30002}], "advertise": false}`
	default:
		return ""
	}
}

// ensureDeployment creates or updates the Deployment
func (r *CardanoNodeReconciler) ensureDeployment(ctx context.Context, node *cardanov1alpha1.CardanoNode) error {
	log := logf.FromContext(ctx)

	// Check if Deployment already exists
	existingDep := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: node.Name, Namespace: node.Namespace}, existingDep)
	if err == nil {
		log.V(1).Info("Deployment already exists", "deployment", node.Name)
		return nil
	}
	if !errors.IsNotFound(err) {
		return err
	}

	// Get image
	image := defaultCardanoNodeImage
	if node.Spec.NodeVersion != "" {
		image = "ghcr.io/intersectmbo/cardano-node:" + node.Spec.NodeVersion
	}

	// Get network flag
	networkFlag := fmt.Sprintf("--%s", node.Spec.Network)

	// Build command
	cmd := []string{
		"run",
		"--config", fmt.Sprintf("/opt/cardano/config/%s/config.json", node.Spec.Network),
		"--topology", "/opt/cardano/topology/topology.json",
		"--database-path", "/opt/cardano/data/db",
		"--socket-path", "/opt/cardano/ipc/node.socket",
		"--port", "6000",
		networkFlag,
	}

	replicas := int32(1)

	// Create Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      node.Name,
			Namespace: node.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "cardano-node",
				"app.kubernetes.io/managed-by": "cardano-operator",
				"cardano.org/node":             node.Name,
				"cardano.org/node-type":        string(node.Spec.Type),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": "cardano-node",
					"cardano.org/node":       node.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name": "cardano-node",
						"cardano.org/node":       node.Name,
						"cardano.org/node-type":  string(node.Spec.Type),
					},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: ptrBool(true),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Containers: []corev1.Container{
						{
							Name:      "cardano-node",
							Image:     image,
							Command:   []string{"cardano-node"},
							Args:      cmd,
							Resources: node.Spec.Resources,
							Ports: []corev1.ContainerPort{
								{
									Name:          "node",
									ContainerPort: 6000,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "prometheus",
									ContainerPort: 12798,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: ptrBool(false),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
								RunAsNonRoot: ptrBool(true),
								RunAsUser:    ptrInt64(1000),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/opt/cardano/data",
								},
								{
									Name:      "ipc",
									MountPath: "/opt/cardano/ipc",
								},
								{
									Name:      "topology",
									MountPath: "/opt/cardano/topology",
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"test", "-S", "/opt/cardano/ipc/node.socket",
										},
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"test", "-S", "/opt/cardano/ipc/node.socket",
										},
									},
								},
								InitialDelaySeconds: 60,
								PeriodSeconds:       30,
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: node.Name + "-data",
								},
							},
						},
						{
							Name: "ipc",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "topology",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: node.Name + "-config",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Add node selector if specified
	if len(node.Spec.Resources.Limits) > 0 || len(node.Spec.Resources.Requests) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Resources = node.Spec.Resources
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(node, deployment, r.Scheme); err != nil {
		return err
	}

	log.Info("Creating Deployment", "deployment", node.Name)
	return r.Create(ctx, deployment)
}

// ensureNetworkPolicy creates a NetworkPolicy for block producer isolation
func (r *CardanoNodeReconciler) ensureNetworkPolicy(ctx context.Context, node *cardanov1alpha1.CardanoNode) error {
	log := logf.FromContext(ctx)
	npName := node.Name + "-network-policy"

	// Check if NetworkPolicy already exists
	existingNP := &networkingv1.NetworkPolicy{}
	err := r.Get(ctx, types.NamespacedName{Name: npName, Namespace: node.Namespace}, existingNP)
	if err == nil {
		log.V(1).Info("NetworkPolicy already exists", "networkpolicy", npName)
		return nil
	}
	if !errors.IsNotFound(err) {
		return err
	}

	// Create NetworkPolicy to restrict ingress to only relay nodes
	np := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      npName,
			Namespace: node.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "cardano-node",
				"app.kubernetes.io/managed-by": "cardano-operator",
				"cardano.org/node":             node.Name,
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"cardano.org/node": node.Name,
				},
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"cardano.org/node-type": "relay",
								},
							},
						},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Port: ptrIntOrString(intstr.FromInt(6000)),
						},
					},
				},
			},
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(node, np, r.Scheme); err != nil {
		return err
	}

	log.Info("Creating NetworkPolicy", "networkpolicy", npName)
	return r.Create(ctx, np)
}

// updatePodStatus updates the status with pod information
func (r *CardanoNodeReconciler) updatePodStatus(ctx context.Context, node *cardanov1alpha1.CardanoNode) error {
	// List pods for this node
	podList := &corev1.PodList{}
	if err := r.List(ctx, podList, client.InNamespace(node.Namespace), client.MatchingLabels{
		"cardano.org/node": node.Name,
	}); err != nil {
		return err
	}

	if len(podList.Items) > 0 {
		node.Status.PodName = podList.Items[0].Name
	}

	return nil
}

// finalizeCardanoNode handles cleanup when the CardanoNode is deleted
//
//nolint:unparam // Error return reserved for future cleanup logic
func (r *CardanoNodeReconciler) finalizeCardanoNode(ctx context.Context, node *cardanov1alpha1.CardanoNode) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing CardanoNode", "node", node.Name)

	// Child resources will be garbage collected due to owner references
	r.recordEvent(node, corev1.EventTypeNormal, "Finalized", "CardanoNode cleanup completed")
	return nil
}

// recordEvent records an event if the recorder is set
func (r *CardanoNodeReconciler) recordEvent(node *cardanov1alpha1.CardanoNode, eventType, reason, message string) {
	if r.Recorder != nil {
		r.Recorder.Event(node, eventType, reason, message)
	}
}

// Helper functions
func ptrBool(b bool) *bool {
	return &b
}

func ptrInt64(i int64) *int64 {
	return &i
}

func ptrIntOrString(i intstr.IntOrString) *intstr.IntOrString {
	return &i
}

// SetupWithManager sets up the controller with the Manager.
func (r *CardanoNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cardanov1alpha1.CardanoNode{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&networkingv1.NetworkPolicy{}).
		Named("cardanonode").
		Complete(r)
}
