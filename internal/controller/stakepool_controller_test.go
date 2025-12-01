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
	Context("When reconciling a resource", func() {
		const resourceName = "test-stakepool"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		stakepool := &cardanov1alpha1.StakePool{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind StakePool")
			err := k8sClient.Get(ctx, typeNamespacedName, stakepool)
			if err != nil && errors.IsNotFound(err) {
				resource := newValidStakePool(resourceName, "default")
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &cardanov1alpha1.StakePool{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				By("Cleanup the specific resource instance StakePool")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &StakePoolReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

// newValidStakePool creates a valid StakePool for testing
func newValidStakePool(name, namespace string) *cardanov1alpha1.StakePool {
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
