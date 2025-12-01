# Quickstart: Cardano Stake Pool Operator

**Date**: 2025-11-30
**Branch**: `001-cardano-stakepool-operator`

This guide walks through deploying a Cardano stake pool on a Kubernetes cluster using the Cardano Operator.

---

## Prerequisites

| Requirement | Details |
|-------------|---------|
| Kubernetes | v1.28+ cluster with kubectl configured |
| Storage | StorageClass supporting ReadWriteOnce PVCs (200Gi+ available) |
| Networking | NetworkPolicy support (CNI: Calico, Cilium, etc.) |
| Resources | 16GB RAM, 4 CPU cores minimum per node |
| ADA | 500+ ADA for pool deposit + transaction fees |
| Metadata | Pool metadata JSON hosted at a public URL |

---

## Step 1: Install the Operator

### Option A: Helm (Recommended)

```bash
# Add the Helm repository
helm repo add cardano-operator https://ghcr.io/user/charts
helm repo update

# Install the operator
helm install cardano-operator cardano-operator/cardano-operator \
  --namespace cardano-system \
  --create-namespace \
  --set metrics.enabled=true
```

### Option B: kubectl

```bash
# Install CRDs and operator
kubectl apply -f https://github.com/user/cardano-operator/releases/latest/download/install.yaml
```

### Verify Installation

```bash
kubectl get pods -n cardano-system
# Expected: cardano-operator-controller-manager-xxx Running

kubectl get crds | grep cardano.org
# Expected: stakepools.cardano.org, cardanonodes.cardano.org, etc.
```

---

## Step 2: Prepare Secrets (Testnet - Managed Mode)

For testnet deployments with operator-managed keys:

```bash
# Create namespace for your stake pool
kubectl create namespace my-pool

# Create a secret with your payment key (for transaction fees)
kubectl create secret generic payment-keys \
  --namespace my-pool \
  --from-file=payment.skey=./payment.skey \
  --from-file=payment.vkey=./payment.vkey
```

> **Note**: For production/mainnet, use `external` key management mode with OfflineSigningNode. See [Production Deployment](#production-deployment-external-mode) below.

---

## Step 3: Create Pool Metadata

Create a JSON file with your pool metadata and host it at a public URL:

```json
{
  "name": "My Test Pool",
  "description": "A test stake pool managed by the Cardano Operator",
  "ticker": "TEST",
  "homepage": "https://example.com"
}
```

Host this file and note the URL (max 64 characters).

---

## Step 4: Deploy a Stake Pool (Testnet)

Create `stakepool.yaml`:

```yaml
apiVersion: cardano.org/v1alpha1
kind: StakePool
metadata:
  name: my-testnet-pool
  namespace: my-pool
spec:
  # Network configuration
  network: preview  # or preprod, mainnet

  # Pool parameters (on-chain)
  poolParams:
    pledge: "500000000000"      # 500,000 ADA in lovelace
    margin: "0.03"              # 3% margin
    cost: "340000000"           # 340 ADA fixed cost (minimum)
    metadata:
      url: "https://example.com/pool-metadata.json"
    relays:
      - type: dns
        hostname: relay1.example.com
        port: 6000
      - type: dns
        hostname: relay2.example.com
        port: 6000

  # Node configuration
  nodeConfig:
    relayCount: 2
    nodeVersion: "9.2.0"  # Optional: specific version
    blockProducer:
      resources:
        requests:
          cpu: "2"
          memory: "16Gi"
        limits:
          cpu: "4"
          memory: "24Gi"
    relaySpec:
      resources:
        requests:
          cpu: "1"
          memory: "8Gi"
        limits:
          cpu: "2"
          memory: "16Gi"

  # Key management (testnet: operator-managed)
  keyManagement:
    mode: managed
    kesRotationLeadEpochs: 2

  # Payment configuration
  paymentConfig:
    mode: automated
    paymentKeySecretRef:
      name: payment-keys
      key: payment.skey

  # Storage configuration
  storage:
    storageClassName: standard  # Your StorageClass
    size: "200Gi"
```

Apply the configuration:

```bash
kubectl apply -f stakepool.yaml
```

---

## Step 5: Monitor Deployment Progress

```bash
# Watch StakePool status
kubectl get stakepool my-testnet-pool -n my-pool -w

# Check detailed status
kubectl describe stakepool my-testnet-pool -n my-pool

# View operator logs
kubectl logs -n cardano-system -l app.kubernetes.io/name=cardano-operator -f

# Check node sync progress
kubectl get cardanonodes -n my-pool
```

### Expected Progression

| Phase | Duration | Description |
|-------|----------|-------------|
| Pending | ~1 min | Resources being created |
| Provisioning | ~5 min | Pods starting, PVCs bound |
| Syncing | 10-60 min | Nodes syncing blockchain (depends on snapshot) |
| Registering | ~5 min | Pool registration transaction |
| Active | Ongoing | Pool operational, producing blocks |

---

## Step 6: Verify Registration

Once the pool reaches `Active` status:

```bash
# Get pool ID
kubectl get stakepool my-testnet-pool -n my-pool -o jsonpath='{.status.poolId}'

# Check on-chain using cardano-cli (from relay pod)
kubectl exec -n my-pool deploy/my-testnet-pool-relay-0 -- \
  cardano-cli query stake-pools --testnet-magic 2
```

---

## Production Deployment (External Mode)

For mainnet with air-gap security:

### Step 1: Create OfflineSigningNode

```yaml
apiVersion: cardano.org/v1alpha1
kind: OfflineSigningNode
metadata:
  name: my-pool-signing
  namespace: my-pool
spec:
  stakePoolRef: my-mainnet-pool
  signingMode: manual
  requestRetentionDays: 90
```

### Step 2: Configure StakePool for External Signing

```yaml
apiVersion: cardano.org/v1alpha1
kind: StakePool
metadata:
  name: my-mainnet-pool
  namespace: my-pool
spec:
  network: mainnet

  # ... poolParams, nodeConfig, storage ...

  keyManagement:
    mode: external
    offlineSigningNodeRef: my-pool-signing
    kesRotationLeadEpochs: 3

  paymentConfig:
    mode: external
    paymentAddress: addr1q...  # Your funding address
```

### Step 3: Handle Signing Requests

```bash
# Check for pending signing requests
kubectl get offlinesigningnode my-pool-signing -n my-pool

# Get unsigned transaction
kubectl get offlinesigningnode my-pool-signing -n my-pool \
  -o jsonpath='{.status.pendingRequests[0].unsignedTxBody}' | base64 -d > unsigned.tx

# Sign offline (on air-gapped machine)
cardano-cli transaction sign \
  --tx-body-file unsigned.tx \
  --signing-key-file cold.skey \
  --mainnet \
  --out-file signed.tx

# Upload signed transaction
SIGNED_TX=$(base64 < signed.tx)
kubectl patch offlinesigningnode my-pool-signing -n my-pool --type=merge \
  -p "{\"status\":{\"pendingRequests\":[{\"id\":\"<request-id>\",\"signedTx\":\"${SIGNED_TX}\"}]}}"
```

---

## Monitoring

### Prometheus Metrics

The operator exposes metrics at `:8080/metrics`:

| Metric | Description |
|--------|-------------|
| `cardano_stakepool_blocks_minted_total` | Total blocks minted |
| `cardano_stakepool_kes_expiry_epoch` | KES expiry epoch |
| `cardano_node_sync_progress` | Node sync percentage |
| `cardano_node_peer_count` | Connected peers |

### Grafana Dashboard

Import the included dashboard:

```bash
kubectl apply -f https://github.com/user/cardano-operator/releases/latest/download/grafana-dashboard.yaml
```

---

## Common Operations

### Update Pool Parameters

```bash
# Edit the StakePool resource
kubectl edit stakepool my-testnet-pool -n my-pool

# Or patch specific field
kubectl patch stakepool my-testnet-pool -n my-pool --type=merge \
  -p '{"spec":{"poolParams":{"margin":"0.02"}}}'
```

### Manual KES Rotation

```bash
# Trigger early KES rotation
kubectl patch kesrotation my-testnet-pool-kes -n my-pool --type=merge \
  -p '{"spec":{"forceRotation":true}}'
```

### Retire Pool

```bash
# Delete the StakePool resource (triggers retirement)
kubectl delete stakepool my-testnet-pool -n my-pool

# Check retirement status
kubectl get stakepool my-testnet-pool -n my-pool
# Phase should transition: Active → Retiring → Retired
```

---

## Troubleshooting

### Pool Stuck in Syncing

```bash
# Check node logs
kubectl logs -n my-pool -l app.kubernetes.io/component=block-producer -f

# Check PVC status
kubectl get pvc -n my-pool

# Verify network connectivity
kubectl exec -n my-pool deploy/my-testnet-pool-relay-0 -- \
  curl -s localhost:12798/metrics | grep cardano_node_peer_count
```

### KES Rotation Failed

```bash
# Check KESRotation status
kubectl describe kesrotation my-testnet-pool-kes -n my-pool

# View rotation history
kubectl get kesrotation my-testnet-pool-kes -n my-pool \
  -o jsonpath='{.status.rotationHistory}' | jq .
```

### Transaction Submission Failed

```bash
# Check operator logs for transaction errors
kubectl logs -n cardano-system -l app.kubernetes.io/name=cardano-operator | grep -i "tx\|transaction"

# Verify payment address has funds
kubectl exec -n my-pool deploy/my-testnet-pool-relay-0 -- \
  cardano-cli query utxo --address <payment-address> --testnet-magic 2
```

---

## Cleanup

```bash
# Delete stake pool (will trigger retirement if registered)
kubectl delete stakepool my-testnet-pool -n my-pool

# Wait for retirement to complete, then delete namespace
kubectl delete namespace my-pool

# Uninstall operator
helm uninstall cardano-operator -n cardano-system
kubectl delete namespace cardano-system
```

---

## Next Steps

- [Full API Reference](./data-model.md)
- [Security Best Practices](./research.md#5-security-best-practices)
- [CI/CD Integration](./research.md#6-cicd-pipeline-design)
