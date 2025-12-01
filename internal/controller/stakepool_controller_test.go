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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cardanov1alpha1 "github.com/AficaLabs/cardano-operator/api/v1alpha1"
)

var _ = Describe("StakePool Controller", func() {
	const (
		timeout  = time.Second * 30
		interval = time.Millisecond * 500
	)

	Context("When reconciling a new StakePool resource", func() {
		ctx := context.Background()

		It("should set initial phase to Pending and create child resources", func() {
			resourceName := "test-sp-" + time.Now().Format("150405")
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}

			By("creating the custom resource for the Kind StakePool")
			stakepool := newValidStakePool(resourceName, "default")
			Expect(k8sClient.Create(ctx, stakepool)).To(Succeed())

			// Ensure cleanup
			defer func() {
				sp := &cardanov1alpha1.StakePool{}
				if err := k8sClient.Get(ctx, typeNamespacedName, sp); err == nil {
					// Remove finalizer if present
					sp.Finalizers = nil
					_ = k8sClient.Update(ctx, sp)
					_ = k8sClient.Delete(ctx, sp)
				}
			}()

			controllerReconciler := &StakePoolReconciler{
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

			sp := &cardanov1alpha1.StakePool{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, sp)).To(Succeed())
			Expect(sp.Status.Phase).To(Equal(cardanov1alpha1.StakePoolPhasePending))
			Expect(sp.Finalizers).To(ContainElement(StakePoolFinalizer))

			By("Third reconcile - create nodes and transition to Provisioning")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(k8sClient.Get(ctx, typeNamespacedName, sp)).To(Succeed())
			Expect(sp.Status.Phase).To(Equal(cardanov1alpha1.StakePoolPhaseProvisioning))

			By("Verifying block producer node was created")
			bpNodeName := types.NamespacedName{
				Name:      resourceName + "-bp",
				Namespace: "default",
			}
			bpNode := &cardanov1alpha1.CardanoNode{}
			Eventually(func() error {
				return k8sClient.Get(ctx, bpNodeName, bpNode)
			}, timeout, interval).Should(Succeed())

			Expect(bpNode.Spec.Type).To(Equal(cardanov1alpha1.CardanoNodeTypeBlockProducer))
			Expect(bpNode.Spec.Network).To(Equal(cardanov1alpha1.NetworkPreprod))
			Expect(bpNode.Spec.StakePoolRef).To(Equal(resourceName))

			By("Verifying relay node was created")
			relayNodeName := types.NamespacedName{
				Name:      resourceName + "-relay-0",
				Namespace: "default",
			}
			relayNode := &cardanov1alpha1.CardanoNode{}
			Eventually(func() error {
				return k8sClient.Get(ctx, relayNodeName, relayNode)
			}, timeout, interval).Should(Succeed())

			Expect(relayNode.Spec.Type).To(Equal(cardanov1alpha1.CardanoNodeTypeRelay))

			By("Verifying keys secret was created for managed mode")
			secretName := types.NamespacedName{
				Name:      resourceName + "-keys",
				Namespace: "default",
			}
			secret := &corev1.Secret{}
			Eventually(func() error {
				return k8sClient.Get(ctx, secretName, secret)
			}, timeout, interval).Should(Succeed())

			Expect(secret.Data).To(HaveKey("cold.skey"))
			Expect(secret.Data).To(HaveKey("cold.vkey"))
			Expect(secret.Data).To(HaveKey("vrf.skey"))
			Expect(secret.Data).To(HaveKey("vrf.vkey"))
			Expect(secret.Data).To(HaveKey("kes.skey"))
			Expect(secret.Data).To(HaveKey("kes.vkey"))

			By("Verifying node status is updated")
			Expect(k8sClient.Get(ctx, typeNamespacedName, sp)).To(Succeed())
			Expect(len(sp.Status.Nodes)).To(BeNumerically(">=", 2))
		})

		It("should handle external key mode without creating keys secret", func() {
			resourceName := "test-sp-ext-" + time.Now().Format("150405")
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}

			By("creating StakePool with external key mode")
			stakepool := newValidStakePool(resourceName, "default")
			stakepool.Spec.KeyManagement.Mode = cardanov1alpha1.KeyManagementModeExternal
			Expect(k8sClient.Create(ctx, stakepool)).To(Succeed())

			defer func() {
				sp := &cardanov1alpha1.StakePool{}
				if err := k8sClient.Get(ctx, typeNamespacedName, sp); err == nil {
					sp.Finalizers = nil
					_ = k8sClient.Update(ctx, sp)
					_ = k8sClient.Delete(ctx, sp)
				}
			}()

			controllerReconciler := &StakePoolReconciler{
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

			By("Verifying no keys secret was created")
			secretName := types.NamespacedName{
				Name:      resourceName + "-keys",
				Namespace: "default",
			}
			secret := &corev1.Secret{}
			err := k8sClient.Get(ctx, secretName, secret)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})
	})

	Context("When StakePool is being deleted", func() {
		ctx := context.Background()

		It("should handle deletion with finalizer", func() {
			resourceName := "test-sp-del-" + time.Now().Format("150405")
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}

			By("Creating a StakePool")
			stakepool := newValidStakePool(resourceName, "default")
			Expect(k8sClient.Create(ctx, stakepool)).To(Succeed())

			controllerReconciler := &StakePoolReconciler{
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

			By("Verifying block producer was created")
			bpNodeName := types.NamespacedName{
				Name:      resourceName + "-bp",
				Namespace: "default",
			}
			bpNode := &cardanov1alpha1.CardanoNode{}
			Eventually(func() error {
				return k8sClient.Get(ctx, bpNodeName, bpNode)
			}, timeout, interval).Should(Succeed())

			By("Deleting the StakePool")
			sp := &cardanov1alpha1.StakePool{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, sp)).To(Succeed())
			Expect(k8sClient.Delete(ctx, sp)).To(Succeed())

			By("Reconciling to handle deletion")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying StakePool is deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, typeNamespacedName, &cardanov1alpha1.StakePool{})
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("Phase transitions", func() {
		ctx := context.Background()

		It("should transition through phases correctly", func() {
			resourceName := "test-sp-phase-" + time.Now().Format("150405")
			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}

			By("Creating a StakePool")
			stakepool := newValidStakePool(resourceName, "default")
			Expect(k8sClient.Create(ctx, stakepool)).To(Succeed())

			defer func() {
				sp := &cardanov1alpha1.StakePool{}
				if err := k8sClient.Get(ctx, typeNamespacedName, sp); err == nil {
					sp.Finalizers = nil
					_ = k8sClient.Update(ctx, sp)
					_ = k8sClient.Delete(ctx, sp)
				}
			}()

			controllerReconciler := &StakePoolReconciler{
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

			sp := &cardanov1alpha1.StakePool{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, sp)).To(Succeed())
			Expect(sp.Status.Phase).To(Equal(cardanov1alpha1.StakePoolPhasePending))

			By("Third reconcile - transition to Provisioning")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(k8sClient.Get(ctx, typeNamespacedName, sp)).To(Succeed())
			Expect(sp.Status.Phase).To(Equal(cardanov1alpha1.StakePoolPhaseProvisioning))
		})
	})
})

// newValidStakePool creates a valid StakePool for testing
func newValidStakePool(name, namespace string) *cardanov1alpha1.StakePool { //nolint:unparam
	return &cardanov1alpha1.StakePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: cardanov1alpha1.StakePoolSpec{
			Network: cardanov1alpha1.NetworkPreprod,
			PoolParams: cardanov1alpha1.PoolParams{
				Pledge: "500000000000",
				Margin: "0.03",
				Cost:   "340000000",
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
				},
			},
			NodeConfig: cardanov1alpha1.NodeConfig{
				BlockProducer: cardanov1alpha1.NodeSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("2"),
							corev1.ResourceMemory: resource.MustParse("8Gi"),
						},
					},
				},
				RelayCount: 1,
				RelaySpec: cardanov1alpha1.NodeSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("4Gi"),
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
}
