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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cardanov1alpha1 "github.com/AficaLabs/cardano-operator/api/v1alpha1"
)

var _ = Describe("KESRotation Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-kesrotation"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		kesrotation := &cardanov1alpha1.KESRotation{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind KESRotation")
			err := k8sClient.Get(ctx, typeNamespacedName, kesrotation)
			if err != nil && errors.IsNotFound(err) {
				resource := newValidKESRotation(resourceName, "default")
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &cardanov1alpha1.KESRotation{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				By("Cleanup the specific resource instance KESRotation")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &KESRotationReconciler{
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

// newValidKESRotation creates a valid KESRotation for testing
func newValidKESRotation(name, namespace string) *cardanov1alpha1.KESRotation {
	return &cardanov1alpha1.KESRotation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: cardanov1alpha1.KESRotationSpec{
			StakePoolRef:       "test-stakepool",
			AutoRotate:         true,
			RotationLeadEpochs: 2,
		},
	}
}
