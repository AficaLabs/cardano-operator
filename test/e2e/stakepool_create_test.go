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

package e2e

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/AficaLabs/cardano-operator/test/utils"
)

var _ = Describe("StakePool Creation", Ordered, func() {
	const (
		stakePoolName      = "e2e-test-pool"
		stakePoolNamespace = "default"
	)

	BeforeAll(func() {
		By("ensuring the CRDs are installed")
		cmd := exec.Command("kubectl", "get", "crd", "stakepools.cardano.cardano.org")
		_, err := utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "StakePool CRD should be installed")

		cmd = exec.Command("kubectl", "get", "crd", "cardanonodes.cardano.cardano.org")
		_, err = utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "CardanoNode CRD should be installed")
	})

	AfterAll(func() {
		By("cleaning up the test StakePool")
		cmd := exec.Command("kubectl", "delete", "stakepool", stakePoolName,
			"-n", stakePoolNamespace, "--ignore-not-found=true")
		_, _ = utils.Run(cmd)

		By("waiting for cleanup to complete")
		Eventually(func() bool {
			cmd := exec.Command("kubectl", "get", "stakepool", stakePoolName, "-n", stakePoolNamespace)
			_, err := utils.Run(cmd)
			return err != nil // Should error when not found
		}, 2*time.Minute, 5*time.Second).Should(BeTrue(), "StakePool should be deleted")
	})

	Context("When creating a StakePool with managed keys", func() {
		It("should create the StakePool resource", func() {
			By("applying the StakePool CR")
			stakePoolYAML := fmt.Sprintf(`
apiVersion: cardano.cardano.org/v1alpha1
kind: StakePool
metadata:
  name: %s
  namespace: %s
spec:
  network: preprod
  poolParams:
    pledge: "500000000000"
    margin: "0.03"
    cost: "340000000"
    metadata:
      url: "https://example.com/pool.json"
      hash: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
    relays:
      - type: dns
        hostname: relay1.example.com
        port: 6000
  nodeConfig:
    blockProducer:
      resources:
        requests:
          cpu: "1"
          memory: "2Gi"
    relayCount: 1
    relaySpec:
      resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
    nodeVersion: "10.1.4"
  keyManagement:
    mode: managed
    kesRotationLeadEpochs: 2
  paymentConfig:
    mode: external
  storage:
    size: "50Gi"
`, stakePoolName, stakePoolNamespace)

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = strings.NewReader(stakePoolYAML)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred(), "Failed to create StakePool")
		})

		It("should transition to Pending phase", func() {
			By("verifying the StakePool enters Pending phase")
			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "stakepool", stakePoolName,
					"-n", stakePoolNamespace, "-o", "jsonpath={.status.phase}")
				output, err := utils.Run(cmd)
				if err != nil {
					return ""
				}
				return output
			}, 30*time.Second, 2*time.Second).Should(Equal("Pending"))
		})

		It("should have a finalizer", func() {
			By("verifying the finalizer is added")
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "stakepool", stakePoolName,
					"-n", stakePoolNamespace, "-o", "jsonpath={.metadata.finalizers}")
				output, err := utils.Run(cmd)
				if err != nil {
					return false
				}
				return strings.Contains(output, "stakepool.cardano.cardano.org/finalizer")
			}, 30*time.Second, 2*time.Second).Should(BeTrue(), "StakePool should have finalizer")
		})

		It("should transition to Provisioning phase", func() {
			By("verifying the StakePool enters Provisioning phase")
			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "stakepool", stakePoolName,
					"-n", stakePoolNamespace, "-o", "jsonpath={.status.phase}")
				output, err := utils.Run(cmd)
				if err != nil {
					return ""
				}
				return output
			}, 60*time.Second, 5*time.Second).Should(Equal("Provisioning"))
		})

		It("should create block producer CardanoNode", func() {
			By("verifying the block producer CardanoNode is created")
			bpNodeName := stakePoolName + "-bp"
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "cardanonode", bpNodeName, "-n", stakePoolNamespace)
				_, err := utils.Run(cmd)
				return err == nil
			}, 60*time.Second, 5*time.Second).Should(BeTrue(), "Block producer CardanoNode should be created")

			By("verifying block producer is of correct type")
			cmd := exec.Command("kubectl", "get", "cardanonode", bpNodeName,
				"-n", stakePoolNamespace, "-o", "jsonpath={.spec.type}")
			output, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal("BlockProducer"))
		})

		It("should create relay CardanoNode", func() {
			By("verifying the relay CardanoNode is created")
			relayNodeName := stakePoolName + "-relay-0"
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "cardanonode", relayNodeName, "-n", stakePoolNamespace)
				_, err := utils.Run(cmd)
				return err == nil
			}, 60*time.Second, 5*time.Second).Should(BeTrue(), "Relay CardanoNode should be created")

			By("verifying relay is of correct type")
			cmd := exec.Command("kubectl", "get", "cardanonode", relayNodeName,
				"-n", stakePoolNamespace, "-o", "jsonpath={.spec.type}")
			output, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal("Relay"))
		})

		It("should create keys secret for managed mode", func() {
			By("verifying the keys secret is created")
			secretName := stakePoolName + "-keys"
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "secret", secretName, "-n", stakePoolNamespace)
				_, err := utils.Run(cmd)
				return err == nil
			}, 60*time.Second, 5*time.Second).Should(BeTrue(), "Keys secret should be created")

			By("verifying secret contains expected keys")
			cmd := exec.Command("kubectl", "get", "secret", secretName,
				"-n", stakePoolNamespace, "-o", "jsonpath={.data}")
			output, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("cold.skey"))
			Expect(output).To(ContainSubstring("cold.vkey"))
			Expect(output).To(ContainSubstring("vrf.skey"))
			Expect(output).To(ContainSubstring("vrf.vkey"))
			Expect(output).To(ContainSubstring("kes.skey"))
			Expect(output).To(ContainSubstring("kes.vkey"))
		})

		It("should update status with node references", func() {
			By("verifying status contains node references")
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "stakepool", stakePoolName,
					"-n", stakePoolNamespace, "-o", "jsonpath={.status.nodes}")
				output, err := utils.Run(cmd)
				if err != nil {
					return false
				}
				return len(output) > 2 // At least "[]"
			}, 60*time.Second, 5*time.Second).Should(BeTrue())
		})

		It("should create PVCs for nodes", func() {
			By("verifying PVC for block producer")
			bpPVCName := stakePoolName + "-bp"
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "pvc", bpPVCName, "-n", stakePoolNamespace)
				_, err := utils.Run(cmd)
				return err == nil
			}, 60*time.Second, 5*time.Second).Should(BeTrue(), "Block producer PVC should be created")

			By("verifying PVC for relay")
			relayPVCName := stakePoolName + "-relay-0"
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "pvc", relayPVCName, "-n", stakePoolNamespace)
				_, err := utils.Run(cmd)
				return err == nil
			}, 60*time.Second, 5*time.Second).Should(BeTrue(), "Relay PVC should be created")
		})

		It("should create Services for nodes", func() {
			By("verifying Service for block producer")
			bpSvcName := stakePoolName + "-bp"
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "service", bpSvcName, "-n", stakePoolNamespace)
				_, err := utils.Run(cmd)
				return err == nil
			}, 60*time.Second, 5*time.Second).Should(BeTrue(), "Block producer Service should be created")

			By("verifying Service for relay")
			relaySvcName := stakePoolName + "-relay-0"
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "service", relaySvcName, "-n", stakePoolNamespace)
				_, err := utils.Run(cmd)
				return err == nil
			}, 60*time.Second, 5*time.Second).Should(BeTrue(), "Relay Service should be created")
		})

		It("should create NetworkPolicy for block producer", func() {
			By("verifying NetworkPolicy for block producer")
			npName := stakePoolName + "-bp"
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "networkpolicy", npName, "-n", stakePoolNamespace)
				_, err := utils.Run(cmd)
				return err == nil
			}, 60*time.Second, 5*time.Second).Should(BeTrue(), "Block producer NetworkPolicy should be created")
		})

		It("should create Deployments for nodes", func() {
			By("verifying Deployment for block producer")
			bpDeployName := stakePoolName + "-bp"
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "deployment", bpDeployName, "-n", stakePoolNamespace)
				_, err := utils.Run(cmd)
				return err == nil
			}, 60*time.Second, 5*time.Second).Should(BeTrue(), "Block producer Deployment should be created")

			By("verifying Deployment for relay")
			relayDeployName := stakePoolName + "-relay-0"
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "deployment", relayDeployName, "-n", stakePoolNamespace)
				_, err := utils.Run(cmd)
				return err == nil
			}, 60*time.Second, 5*time.Second).Should(BeTrue(), "Relay Deployment should be created")
		})
	})

	Context("When deleting the StakePool", func() {
		It("should delete child resources", func() {
			By("deleting the StakePool")
			cmd := exec.Command("kubectl", "delete", "stakepool", stakePoolName, "-n", stakePoolNamespace)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred(), "Failed to delete StakePool")

			By("verifying child CardanoNodes are deleted")
			bpNodeName := stakePoolName + "-bp"
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "cardanonode", bpNodeName, "-n", stakePoolNamespace)
				_, err := utils.Run(cmd)
				return err != nil // Should error when not found
			}, 2*time.Minute, 5*time.Second).Should(BeTrue(), "Block producer CardanoNode should be deleted")

			relayNodeName := stakePoolName + "-relay-0"
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "cardanonode", relayNodeName, "-n", stakePoolNamespace)
				_, err := utils.Run(cmd)
				return err != nil
			}, 2*time.Minute, 5*time.Second).Should(BeTrue(), "Relay CardanoNode should be deleted")

			By("verifying StakePool is fully deleted")
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "stakepool", stakePoolName, "-n", stakePoolNamespace)
				_, err := utils.Run(cmd)
				return err != nil
			}, 2*time.Minute, 5*time.Second).Should(BeTrue(), "StakePool should be deleted")
		})
	})
})

var _ = Describe("StakePool with External Keys", Ordered, func() {
	const (
		stakePoolName      = "e2e-ext-pool"
		stakePoolNamespace = "default"
	)

	AfterAll(func() {
		By("cleaning up the external test StakePool")
		cmd := exec.Command("kubectl", "delete", "stakepool", stakePoolName,
			"-n", stakePoolNamespace, "--ignore-not-found=true")
		_, _ = utils.Run(cmd)

		Eventually(func() bool {
			cmd := exec.Command("kubectl", "get", "stakepool", stakePoolName,
				"-n", stakePoolNamespace)
			_, err := utils.Run(cmd)
			return err != nil
		}, 2*time.Minute, 5*time.Second).Should(BeTrue())
	})

	Context("When creating a StakePool with external key management", func() {
		It("should create the StakePool without generating keys", func() {
			By("applying the StakePool CR with external key mode")
			stakePoolYAML := fmt.Sprintf(`
apiVersion: cardano.cardano.org/v1alpha1
kind: StakePool
metadata:
  name: %s
  namespace: %s
spec:
  network: preprod
  poolParams:
    pledge: "500000000000"
    margin: "0.03"
    cost: "340000000"
    metadata:
      url: "https://example.com/pool.json"
      hash: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
    relays:
      - type: dns
        hostname: relay1.example.com
        port: 6000
  nodeConfig:
    blockProducer:
      resources:
        requests:
          cpu: "1"
          memory: "2Gi"
    relayCount: 1
    relaySpec:
      resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
    nodeVersion: "10.1.4"
  keyManagement:
    mode: external
    kesRotationLeadEpochs: 2
  paymentConfig:
    mode: external
  storage:
    size: "50Gi"
`, stakePoolName, stakePoolNamespace)

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = strings.NewReader(stakePoolYAML)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred(), "Failed to create StakePool")
		})

		It("should not create keys secret for external mode", func() {
			By("waiting for provisioning to start")
			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "stakepool", stakePoolName,
					"-n", stakePoolNamespace, "-o", "jsonpath={.status.phase}")
				output, _ := utils.Run(cmd)
				return output
			}, 60*time.Second, 5*time.Second).Should(Equal("Provisioning"))

			By("verifying no keys secret is created")
			secretName := stakePoolName + "-keys"
			// Give it some time to ensure it's not created
			Consistently(func() bool {
				cmd := exec.Command("kubectl", "get", "secret", secretName, "-n", stakePoolNamespace)
				_, err := utils.Run(cmd)
				return err != nil // Should consistently error (not found)
			}, 10*time.Second, 2*time.Second).Should(BeTrue(), "Keys secret should NOT be created for external mode")
		})

		It("should still create CardanoNodes", func() {
			By("verifying CardanoNodes are created")
			bpNodeName := stakePoolName + "-bp"
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "cardanonode", bpNodeName, "-n", stakePoolNamespace)
				_, err := utils.Run(cmd)
				return err == nil
			}, 60*time.Second, 5*time.Second).Should(BeTrue(), "Block producer CardanoNode should be created")

			relayNodeName := stakePoolName + "-relay-0"
			Eventually(func() bool {
				cmd := exec.Command("kubectl", "get", "cardanonode", relayNodeName, "-n", stakePoolNamespace)
				_, err := utils.Run(cmd)
				return err == nil
			}, 60*time.Second, 5*time.Second).Should(BeTrue(), "Relay CardanoNode should be created")
		})
	})
})
