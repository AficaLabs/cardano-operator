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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cardanov1alpha1 "github.com/AficaLabs/cardano-operator/api/v1alpha1"
)

var _ = Describe("CardanoNode Controller", func() {
	const (
		timeout  = time.Second * 30
		interval = time.Millisecond * 500
	)

	Context("When reconciling a new CardanoNode", func() {
		ctx := context.Background()

		It("should set initial phase and create child resources", func() {
			resourceName := "test-cn-" + time.Now().Format("150405")
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}

			By("creating the custom resource for the Kind CardanoNode")
			cardanonode := newValidCardanoNode(resourceName, "default")
			Expect(k8sClient.Create(ctx, cardanonode)).To(Succeed())

			// Ensure cleanup
			defer func() {
				cn := &cardanov1alpha1.CardanoNode{}
				if err := k8sClient.Get(ctx, typeNamespacedName, cn); err == nil {
					cn.Finalizers = nil
					_ = k8sClient.Update(ctx, cn)
					_ = k8sClient.Delete(ctx, cn)
				}
			}()

			controllerReconciler := &CardanoNodeReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			By("First reconcile - add finalizer")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Second reconcile - set Pending phase")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			cn := &cardanov1alpha1.CardanoNode{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, cn)).To(Succeed())
			Expect(cn.Status.Phase).To(Equal(cardanov1alpha1.CardanoNodePhasePending))
			Expect(cn.Finalizers).To(ContainElement(CardanoNodeFinalizer))

			By("Third reconcile - create resources and transition to Starting")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(k8sClient.Get(ctx, typeNamespacedName, cn)).To(Succeed())
			Expect(cn.Status.Phase).To(Equal(cardanov1alpha1.CardanoNodePhaseStarting))

			By("Verifying PVC was created")
			pvc := &corev1.PersistentVolumeClaim{}
			pvcName := types.NamespacedName{
				Name:      resourceName + "-data",
				Namespace: "default",
			}
			Eventually(func() error {
				return k8sClient.Get(ctx, pvcName, pvc)
			}, timeout, interval).Should(Succeed())
			Expect(pvc.Spec.Resources.Requests[corev1.ResourceStorage]).To(Equal(resource.MustParse("200Gi")))

			By("Verifying Service was created")
			svc := &corev1.Service{}
			svcName := types.NamespacedName{
				Name:      resourceName + "-svc",
				Namespace: "default",
			}
			Eventually(func() error {
				return k8sClient.Get(ctx, svcName, svc)
			}, timeout, interval).Should(Succeed())
			Expect(svc.Spec.Ports).To(HaveLen(2)) // node port + prometheus port
			Expect(svc.Spec.Ports[0].Port).To(Equal(int32(6000)))

			By("Verifying ConfigMap was created")
			cm := &corev1.ConfigMap{}
			cmName := types.NamespacedName{
				Name:      resourceName + "-config",
				Namespace: "default",
			}
			Eventually(func() error {
				return k8sClient.Get(ctx, cmName, cm)
			}, timeout, interval).Should(Succeed())
			Expect(cm.Data).To(HaveKey("topology.json"))
		})

		It("should create block producer node with NetworkPolicy", func() {
			resourceName := "test-bp-" + time.Now().Format("150405")
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}

			By("creating a block producer CardanoNode")
			cardanonode := newValidCardanoNode(resourceName, "default")
			cardanonode.Spec.Type = cardanov1alpha1.CardanoNodeTypeBlockProducer
			cardanonode.Spec.StakePoolRef = "my-pool"
			Expect(k8sClient.Create(ctx, cardanonode)).To(Succeed())

			defer func() {
				cn := &cardanov1alpha1.CardanoNode{}
				if err := k8sClient.Get(ctx, typeNamespacedName, cn); err == nil {
					cn.Finalizers = nil
					_ = k8sClient.Update(ctx, cn)
					_ = k8sClient.Delete(ctx, cn)
				}
			}()

			controllerReconciler := &CardanoNodeReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// Run reconcile loops
			for i := 0; i < 3; i++ {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())
			}

			By("Verifying NetworkPolicy was created for block producer")
			np := &networkingv1.NetworkPolicy{}
			npName := types.NamespacedName{
				Name:      resourceName + "-network-policy",
				Namespace: "default",
			}
			Eventually(func() error {
				return k8sClient.Get(ctx, npName, np)
			}, timeout, interval).Should(Succeed())
		})
	})

	Context("When CardanoNode is being deleted", func() {
		ctx := context.Background()

		It("should handle deletion with finalizer", func() {
			resourceName := "test-cn-del-" + time.Now().Format("150405")
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}

			By("Creating a CardanoNode")
			cardanonode := newValidCardanoNode(resourceName, "default")
			Expect(k8sClient.Create(ctx, cardanonode)).To(Succeed())

			controllerReconciler := &CardanoNodeReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			By("Reconciling to add finalizer and create resources")
			for i := 0; i < 3; i++ {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())
			}

			By("Verifying PVC was created")
			pvc := &corev1.PersistentVolumeClaim{}
			pvcName := types.NamespacedName{
				Name:      resourceName + "-data",
				Namespace: "default",
			}
			Eventually(func() error {
				return k8sClient.Get(ctx, pvcName, pvc)
			}, timeout, interval).Should(Succeed())

			By("Deleting the CardanoNode")
			cn := &cardanov1alpha1.CardanoNode{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, cn)).To(Succeed())
			Expect(k8sClient.Delete(ctx, cn)).To(Succeed())

			By("Reconciling to handle deletion")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying CardanoNode is deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, typeNamespacedName, &cardanov1alpha1.CardanoNode{})
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("Static topology configuration", func() {
		ctx := context.Background()

		It("should generate static topology config", func() {
			resourceName := "test-static-" + time.Now().Format("150405")
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}

			By("creating CardanoNode with static topology")
			cardanonode := newValidCardanoNode(resourceName, "default")
			cardanonode.Spec.Topology = &cardanov1alpha1.TopologyConfig{
				Mode: cardanov1alpha1.TopologyModeStatic,
				StaticPeers: []cardanov1alpha1.Peer{
					{Address: "relay1.example.com", Port: 6000},
					{Address: "relay2.example.com", Port: 6001},
				},
			}
			Expect(k8sClient.Create(ctx, cardanonode)).To(Succeed())

			defer func() {
				cn := &cardanov1alpha1.CardanoNode{}
				if err := k8sClient.Get(ctx, typeNamespacedName, cn); err == nil {
					cn.Finalizers = nil
					_ = k8sClient.Update(ctx, cn)
					_ = k8sClient.Delete(ctx, cn)
				}
			}()

			controllerReconciler := &CardanoNodeReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// Run reconcile loops
			for i := 0; i < 3; i++ {
				_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())
			}

			By("Verifying ConfigMap contains static peers")
			cm := &corev1.ConfigMap{}
			cmName := types.NamespacedName{
				Name:      resourceName + "-config",
				Namespace: "default",
			}
			Eventually(func() error {
				return k8sClient.Get(ctx, cmName, cm)
			}, timeout, interval).Should(Succeed())

			Expect(cm.Data["topology.json"]).To(ContainSubstring("relay1.example.com"))
			Expect(cm.Data["topology.json"]).To(ContainSubstring("relay2.example.com"))
		})
	})

	Context("When reconciling non-existent resource", func() {
		It("should not error for missing resource", func() {
			ctx := context.Background()
			controllerReconciler := &CardanoNodeReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "non-existent-node",
					Namespace: "default",
				},
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

// newValidCardanoNode creates a valid CardanoNode for testing
func newValidCardanoNode(name, namespace string) *cardanov1alpha1.CardanoNode { //nolint:unparam
	return &cardanov1alpha1.CardanoNode{
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
			},
			Storage: cardanov1alpha1.StorageConfig{
				Size: "200Gi",
			},
		},
	}
}
