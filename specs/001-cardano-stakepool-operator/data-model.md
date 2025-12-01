# Data Model: Cardano Stake Pool Kubernetes Operator

**Date**: 2025-11-30
**Branch**: `001-cardano-stakepool-operator`

## Overview

This document defines the Custom Resource Definitions (CRDs) and their relationships for the Cardano Stake Pool Operator. All CRDs belong to the `cardano.org` API group with initial version `v1alpha1`.

---

## Entity Relationship Diagram

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                              StakePool (Primary)                             │
│  - Owns CardanoNode (block-producer, relays)                                │
│  - Owns KESRotation                                                         │
│  - References OfflineSigningNode (optional, for air-gap mode)               │
│  - References Secrets (payment key, cold key if managed)                    │
└─────────────────────────────────────────────────────────────────────────────┘
         │ owns                    │ owns                    │ references
         ▼                         ▼                         ▼
┌─────────────────┐    ┌─────────────────────┐    ┌─────────────────────────┐
│  CardanoNode    │    │    KESRotation      │    │   OfflineSigningNode    │
│  - Deployment   │    │    - Scheduled      │    │   - Signing requests    │
│  - Service      │    │    - Status         │    │   - Workflow status     │
│  - PVC          │    │                     │    │                         │
│  - NetworkPolicy│    │                     │    │                         │
└─────────────────┘    └─────────────────────┘    └─────────────────────────┘
```

---

## 1. StakePool

The primary CRD representing a Cardano stake pool with all configuration.

### Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `network` | `string` | Yes | Network identifier: `mainnet`, `preprod`, `preview` |
| `poolParams` | `PoolParams` | Yes | On-chain pool parameters |
| `nodeConfig` | `NodeConfig` | Yes | Node deployment configuration |
| `keyManagement` | `KeyManagement` | Yes | Key handling mode and references |
| `paymentConfig` | `PaymentConfig` | Yes | Transaction funding configuration |
| `storage` | `StorageConfig` | Yes | PVC configuration for blockchain data |

### PoolParams

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `pledge` | `string` | Yes | Pledge amount in lovelace (e.g., "500000000000") |
| `margin` | `string` | Yes | Pool margin as decimal (e.g., "0.03" for 3%) |
| `cost` | `string` | Yes | Fixed cost per epoch in lovelace (min 340 ADA) |
| `metadata` | `PoolMetadata` | Yes | Pool metadata reference |
| `relays` | `[]RelayConfig` | Yes | Relay endpoint declarations (min 2 for production) |
| `rewardAccount` | `string` | No | Reward account address (defaults to pool's stake address) |

### PoolMetadata

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `url` | `string` | Yes | URL to metadata JSON (max 64 chars) |
| `hash` | `string` | No | Pre-computed metadata hash (auto-computed if omitted) |

### RelayConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | `string` | Yes | `dns` or `ip` |
| `hostname` | `string` | Conditional | DNS hostname (if type=dns) |
| `ipv4` | `string` | Conditional | IPv4 address (if type=ip) |
| `ipv6` | `string` | No | IPv6 address |
| `port` | `int32` | Yes | Port number (default: 6000) |

### NodeConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `blockProducer` | `NodeSpec` | Yes | Block producer node configuration |
| `relayCount` | `int32` | Yes | Number of relay nodes (min 2) |
| `relaySpec` | `NodeSpec` | Yes | Relay node configuration template |
| `nodeVersion` | `string` | No | Cardano node version (default: latest stable) |

### NodeSpec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `resources` | `ResourceRequirements` | No | CPU/memory requests and limits |
| `nodeSelector` | `map[string]string` | No | Node selector for scheduling |
| `tolerations` | `[]Toleration` | No | Pod tolerations |
| `affinity` | `Affinity` | No | Pod affinity rules |

### KeyManagement

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `mode` | `string` | Yes | `managed` (operator generates) or `external` (user provides) |
| `coldKeySecretRef` | `SecretReference` | Conditional | Secret containing cold key (if mode=managed or importing) |
| `offlineSigningNodeRef` | `string` | Conditional | OfflineSigningNode name (if mode=external) |
| `kesRotationLeadEpochs` | `int32` | No | Epochs before KES expiry to rotate (default: 2) |

### PaymentConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `mode` | `string` | Yes | `automated` or `external` |
| `paymentKeySecretRef` | `SecretReference` | Conditional | Secret with payment signing key (if mode=automated) |
| `paymentAddress` | `string` | Conditional | Address for funding (if mode=external, for display) |

### StorageConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `storageClassName` | `string` | No | StorageClass name (default: cluster default) |
| `size` | `string` | Yes | PVC size (e.g., "200Gi") |
| `accessModes` | `[]string` | No | Access modes (default: ReadWriteOnce) |

### Status Fields

| Field | Type | Description |
|-------|------|-------------|
| `phase` | `string` | Current phase: `Pending`, `Provisioning`, `Syncing`, `Registering`, `Active`, `Updating`, `Retiring`, `Retired`, `Error` |
| `poolId` | `string` | Bech32 pool ID (pool1...) once registered |
| `poolIdHex` | `string` | Hex pool ID |
| `currentEpoch` | `int64` | Current Cardano epoch |
| `kesExpiryEpoch` | `int64` | Epoch when current KES expires |
| `blocksMinted` | `int64` | Total blocks minted by this pool |
| `lastBlockSlot` | `int64` | Slot of last minted block |
| `registrationTxHash` | `string` | Registration transaction hash |
| `conditions` | `[]Condition` | Standard Kubernetes conditions |
| `nodes` | `[]NodeStatus` | Status of each CardanoNode |

### Conditions

| Type | Description |
|------|-------------|
| `Ready` | Pool is fully operational and producing blocks |
| `NodesProvisioned` | All nodes (BP + relays) are running |
| `NodesSynced` | All nodes are synced to chain tip |
| `KeysReady` | All required keys are available |
| `PoolRegistered` | Pool registration confirmed on-chain |
| `KESValid` | Current KES key is valid (not expired) |

---

## 2. CardanoNode

Represents a single Cardano node instance (block producer or relay).

### Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | `string` | Yes | `block-producer`, `relay`, or `offline-signing` |
| `network` | `string` | Yes | Network: `mainnet`, `preprod`, `preview` |
| `nodeVersion` | `string` | No | Cardano node version |
| `topology` | `TopologyConfig` | Yes | Peer topology configuration |
| `resources` | `ResourceRequirements` | No | CPU/memory configuration |
| `storage` | `StorageConfig` | Yes | PVC configuration |
| `stakePoolRef` | `string` | No | Owning StakePool name (for BP nodes) |

### TopologyConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `mode` | `string` | Yes | `static` or `p2p` |
| `staticPeers` | `[]Peer` | Conditional | Static peer list (if mode=static) |
| `p2pConfig` | `P2PConfig` | Conditional | P2P configuration (if mode=p2p) |

### Peer

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `address` | `string` | Yes | IP or hostname |
| `port` | `int32` | Yes | Port number |

### Status Fields

| Field | Type | Description |
|-------|------|-------------|
| `phase` | `string` | `Pending`, `Starting`, `Syncing`, `Running`, `Error` |
| `syncProgress` | `string` | Sync percentage (e.g., "99.5%") |
| `tipSlot` | `int64` | Current chain tip slot |
| `tipEpoch` | `int64` | Current epoch |
| `peerCount` | `int32` | Connected peer count |
| `nodeVersion` | `string` | Running node version |
| `podName` | `string` | Associated Pod name |
| `serviceName` | `string` | Associated Service name |
| `conditions` | `[]Condition` | Standard conditions |

---

## 3. OfflineSigningNode

Manages air-gap signing workflows for cold key operations.

### Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `stakePoolRef` | `string` | Yes | Associated StakePool name |
| `signingMode` | `string` | Yes | `manual` (user exports/imports) or `isolated` (dedicated namespace) |
| `requestRetentionDays` | `int32` | No | Days to retain completed requests (default: 30) |

### Status Fields

| Field | Type | Description |
|-------|------|-------------|
| `phase` | `string` | `Idle`, `RequestPending`, `AwaitingSignature`, `Processing`, `Error` |
| `pendingRequests` | `[]SigningRequest` | Queue of pending signing requests |
| `completedRequests` | `[]SigningRequest` | Recently completed requests |
| `lastActivity` | `Time` | Last signing activity timestamp |

### SigningRequest (embedded)

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Unique request ID |
| `type` | `string` | `registration`, `update`, `kes-rotation`, `retirement` |
| `createdAt` | `Time` | Request creation time |
| `unsignedTxBody` | `string` | Base64-encoded unsigned transaction body |
| `unsignedTxHash` | `string` | Hash for verification |
| `signedTx` | `string` | Base64-encoded signed transaction (user provides) |
| `txHash` | `string` | Submitted transaction hash |
| `status` | `string` | `pending`, `awaiting-signature`, `signed`, `submitted`, `confirmed`, `failed` |
| `errorMessage` | `string` | Error details if failed |

---

## 4. KESRotation

Tracks KES key rotation schedule and history.

### Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `stakePoolRef` | `string` | Yes | Associated StakePool name |
| `autoRotate` | `bool` | No | Enable automatic rotation (default: true) |
| `rotationLeadEpochs` | `int32` | No | Epochs before expiry to trigger rotation (default: 2) |
| `maxKESEvolutions` | `int32` | No | Max KES evolutions (protocol param, usually 62) |

### Status Fields

| Field | Type | Description |
|-------|------|-------------|
| `currentKESPeriod` | `int64` | Current KES period number |
| `kesStartSlot` | `int64` | Slot when current KES became active |
| `kesExpiryEpoch` | `int64` | Epoch when current KES expires |
| `nextRotationEpoch` | `int64` | Planned rotation epoch |
| `rotationHistory` | `[]RotationEvent` | Recent rotation events |
| `conditions` | `[]Condition` | Standard conditions |

### RotationEvent (embedded)

| Field | Type | Description |
|-------|------|-------------|
| `kesPeriod` | `int64` | KES period that was rotated |
| `rotatedAt` | `Time` | Rotation timestamp |
| `expiryEpoch` | `int64` | Expiry epoch of new KES |
| `certificateTxHash` | `string` | Certificate update transaction |
| `success` | `bool` | Whether rotation succeeded |
| `errorMessage` | `string` | Error if failed |

---

## State Transitions

### StakePool Lifecycle

```text
                                    ┌──────────────┐
                                    │   Pending    │ (resource created)
                                    └──────┬───────┘
                                           │
                                           ▼
                                    ┌──────────────┐
                                    │ Provisioning │ (creating nodes, PVCs)
                                    └──────┬───────┘
                                           │
                                           ▼
                                    ┌──────────────┐
                                    │   Syncing    │ (nodes syncing blockchain)
                                    └──────┬───────┘
                                           │
                                           ▼
                                    ┌──────────────┐
                                    │ Registering  │ (submitting pool cert)
                                    └──────┬───────┘
                                           │
                   ┌───────────────────────┼───────────────────────┐
                   │                       │                       │
                   ▼                       ▼                       ▼
            ┌──────────────┐        ┌──────────────┐        ┌──────────────┐
            │    Active    │◀──────▶│   Updating   │        │    Error     │
            │ (producing)  │        │ (param change)│        │              │
            └──────┬───────┘        └──────────────┘        └──────────────┘
                   │
                   │ (delete resource)
                   ▼
            ┌──────────────┐
            │   Retiring   │ (retirement cert submitted)
            └──────┬───────┘
                   │
                   │ (retirement epoch reached)
                   ▼
            ┌──────────────┐
            │   Retired    │ (cleanup complete)
            └──────────────┘
```

### CardanoNode Lifecycle

```text
Pending → Starting → Syncing → Running
                         │
                         └──▶ Error (recoverable, retry)
```

### OfflineSigningNode Request Workflow

```text
pending → awaiting-signature → signed → submitted → confirmed
                │                           │
                └───────────────────────────┴──▶ failed
```

---

## Validation Rules

### StakePool

| Rule | Validation |
|------|------------|
| Pledge format | Must be valid lovelace amount (numeric string) |
| Margin range | 0.0 to 1.0 (0% to 100%) |
| Cost minimum | >= 340000000 lovelace (340 ADA, protocol minimum) |
| Relay count | >= 2 for production networks |
| Metadata URL | Valid URL, max 64 characters |
| Network | One of: mainnet, preprod, preview |

### CardanoNode

| Rule | Validation |
|------|------------|
| Type | One of: block-producer, relay, offline-signing |
| Storage size | >= 100Gi recommended |

### Cross-Resource

| Rule | Validation |
|------|------------|
| Cold key uniqueness | Same cold key cannot be used by multiple StakePools |
| Payment key isolation | Payment keys should not be reused across pools |

---

## Indexes and Queries

Recommended indexes for efficient reconciliation:

| Resource | Index | Purpose |
|----------|-------|---------|
| StakePool | `.spec.network` | Filter by network |
| StakePool | `.status.phase` | Find pools needing action |
| CardanoNode | `.spec.stakePoolRef` | Find nodes for a pool |
| KESRotation | `.status.nextRotationEpoch` | Find upcoming rotations |
| OfflineSigningNode | `.status.phase` | Find nodes with pending requests |
