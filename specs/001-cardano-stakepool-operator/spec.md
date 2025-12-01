# Feature Specification: Cardano Stake Pool Kubernetes Operator

**Feature Branch**: `001-cardano-stakepool-operator`
**Created**: 2025-11-30
**Status**: Draft
**Input**: User description: "Build a kubernetes operator for deploying, operating, and managing a cardano stake pool"

## Clarifications

### Session 2025-11-30

- Q: How should the operator handle cold key operations that require air-gap security? → A: Hybrid mode with optional OfflineSigningNode CRD - users choose per-pool whether cold keys are operator-managed (convenience/testnet) or external (production security), plus an optional OfflineSigningNode CRD for managing signing ceremonies in isolated namespaces.
- Q: How should relay node count and topology be configured? → A: Configurable count with minimum of 2 relays required for production redundancy.
- Q: How should blockchain data storage be handled for Cardano nodes? → A: Operator manages PersistentVolumeClaims with configurable StorageClass.
- Q: How should the operator access ADA funds for transaction fees and deposits? → A: Support both modes configurable per-pool: (1) user provides payment key reference via Secret for automated transactions, or (2) operator generates unsigned transactions for external signing and submission.
- Q: How should network isolation between block producer and relay nodes be enforced? → A: Block producer only accessible to relay nodes via NetworkPolicy; relays are public-facing.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Deploy a New Stake Pool (Priority: P1)

As a stake pool operator, I want to deploy a fully functional Cardano stake pool by defining a single declarative configuration, so that I can participate in Cardano block production without manual infrastructure setup.

**Why this priority**: This is the core value proposition of the operator. Without the ability to deploy a stake pool, no other functionality is useful. This enables the primary use case of transforming complex multi-step stake pool setup into a single declarative resource.

**Independent Test**: Can be fully tested by applying a StakePool custom resource and verifying that all required components (block producer node, relay nodes, keys) are provisioned and the pool is registered on the Cardano network.

**Acceptance Scenarios**:

1. **Given** a Kubernetes cluster with the operator installed and no existing stake pool, **When** I apply a StakePool custom resource with valid configuration (pool name, pledge amount, margin, relay information), **Then** the operator provisions a block producer node, relay nodes, generates required keys, and registers the pool on the Cardano network within 30 minutes.

2. **Given** a StakePool resource is applied, **When** the pool registration transaction is submitted, **Then** the operator updates the StakePool status to reflect the registration state (pending, registered) and the pool ID.

3. **Given** an invalid StakePool configuration (e.g., pledge exceeds available funds), **When** I apply the resource, **Then** the operator reports a clear error in the status without crashing or leaving partial resources.

---

### User Story 2 - Monitor Pool Health and Performance (Priority: P2)

As a stake pool operator, I want to monitor my pool's health, synchronization status, and block production metrics, so that I can ensure my pool is operating correctly and maximize my rewards.

**Why this priority**: Once a pool is deployed, operators need visibility into its health to maintain uptime and optimize performance. This is essential for ongoing operations but not required for initial deployment.

**Independent Test**: Can be fully tested by deploying a stake pool and verifying that health metrics, sync status, and block production data are exposed through standard monitoring interfaces.

**Acceptance Scenarios**:

1. **Given** a running stake pool, **When** I query the monitoring endpoint, **Then** I can see node sync percentage, tip slot, peer count, and memory/CPU utilization.

2. **Given** the pool is elected as slot leader, **When** a block is successfully minted, **Then** the metrics reflect the minted block count and the event is logged.

3. **Given** the block producer node loses connectivity to peers, **When** the peer count drops below a healthy threshold, **Then** an alert-ready metric is exposed indicating degraded connectivity.

---

### User Story 3 - Automated KES Key Rotation (Priority: P2)

As a stake pool operator, I want the operator to automatically rotate KES (Key Evolving Signature) keys before they expire, so that my pool continues producing blocks without manual intervention.

**Why this priority**: KES keys expire after a fixed number of epochs. Manual rotation is error-prone and missing rotation results in lost block production. This automation is critical for reliable pool operation.

**Independent Test**: Can be fully tested by deploying a pool and verifying that as KES key expiration approaches, the operator generates new operational certificates and updates the block producer without downtime.

**Acceptance Scenarios**:

1. **Given** a stake pool with KES keys approaching expiration (configurable threshold, default 2 epochs before expiry), **When** the rotation threshold is reached, **Then** the operator generates new KES keys and operational certificate, updates the block producer, and logs the rotation event.

2. **Given** KES rotation is in progress, **When** the new certificate is applied, **Then** block production continues without interruption.

3. **Given** KES rotation fails (e.g., cold key unavailable), **When** the rotation is attempted, **Then** the operator reports the failure in status and emits an alert metric.

---

### User Story 4 - Update Pool Parameters (Priority: P3)

As a stake pool operator, I want to update my pool's metadata, margin, pledge, or relay information by modifying the StakePool resource, so that I can adjust my pool's configuration without manual transaction building.

**Why this priority**: Pool operators periodically need to update their pool parameters. This is a less frequent operation than deployment or monitoring but essential for ongoing pool management.

**Independent Test**: Can be fully tested by modifying a deployed StakePool resource and verifying the update transaction is submitted and reflected on-chain.

**Acceptance Scenarios**:

1. **Given** a registered stake pool, **When** I update the StakePool resource spec (e.g., change margin from 3% to 2%), **Then** the operator submits a pool update certificate and the change is reflected on-chain within the next epoch.

2. **Given** a pool metadata URL change, **When** I update the metadata reference in the StakePool spec, **Then** the operator updates the on-chain metadata hash and the new metadata is retrievable.

3. **Given** an invalid update (e.g., pledge increase beyond available funds), **When** the update is attempted, **Then** the operator rejects the update with a clear error message before submitting any transaction.

---

### User Story 5 - Retire a Stake Pool (Priority: P3)

As a stake pool operator, I want to retire my stake pool gracefully by deleting the StakePool resource, so that delegators are notified and my deposit is returned after the retirement epoch.

**Why this priority**: Pool retirement is an important lifecycle operation but occurs infrequently. It must be handled correctly to ensure delegators are properly notified and deposits are recovered.

**Independent Test**: Can be fully tested by deleting a StakePool resource and verifying the retirement certificate is submitted and the pool transitions to retired status.

**Acceptance Scenarios**:

1. **Given** a registered stake pool, **When** I delete the StakePool resource, **Then** the operator submits a retirement certificate specifying retirement at the current epoch + 2 (configurable) and updates status to "retiring".

2. **Given** a retiring stake pool, **When** the retirement epoch is reached, **Then** the operator cleans up Kubernetes resources (nodes, secrets) and the pool deposit is returned to the specified address.

3. **Given** a retirement in progress, **When** I recreate the StakePool resource before retirement epoch, **Then** the operator allows cancellation of retirement (re-registration).

---

### Edge Cases

- What happens when the Cardano network experiences a hard fork during pool operation?
  - The operator handles protocol parameter changes gracefully and updates nodes as required.
- How does the system handle insufficient funds for pool registration or updates?
  - The operator validates available funds before attempting transactions and reports clear errors.
- What happens when the block producer node crashes during block production?
  - The operator automatically restarts the node and restores connectivity.
- How does the system handle network partitions where relay nodes lose connectivity?
  - The operator monitors peer counts and attempts reconnection, exposing degraded status metrics.
- What happens when multiple StakePool resources attempt to use the same cold key?
  - The operator rejects duplicate cold key usage with a clear error.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow users to deploy a Cardano stake pool by applying a declarative custom resource.
- **FR-002**: System MUST provision and manage Cardano block producer nodes with network isolation enforced via NetworkPolicy (block producer accessible only to relay nodes, not public internet).
- **FR-003**: System MUST provision and manage Cardano relay nodes for network connectivity, with configurable count and a minimum of 2 relays required for production pools.
- **FR-004**: System MUST support hybrid cold key management: operator-managed keys (convenience mode for testnets) OR externally-provided keys with signing request/response workflow (production security mode), configurable per-pool.
- **FR-005**: System MUST register stake pools on the Cardano network by building and submitting registration certificates.
- **FR-006**: System MUST expose stake pool health metrics (sync status, peer count, block production) through a standard monitoring interface.
- **FR-007**: System MUST automatically rotate KES keys before expiration with configurable lead time.
- **FR-008**: System MUST allow pool parameter updates (margin, pledge, metadata, relays) through resource modification.
- **FR-009**: System MUST handle stake pool retirement by submitting retirement certificates when the resource is deleted.
- **FR-010**: System MUST store cryptographic keys securely using external secret management integration.
- **FR-011**: System MUST support both Cardano mainnet and testnet networks.
- **FR-012**: System MUST emit events for significant state changes (pool registered, KES rotated, block minted, retirement initiated).
- **FR-013**: System MUST recover from node failures by automatically restarting failed components.
- **FR-014**: System MUST validate pool configurations before attempting on-chain transactions.
- **FR-015**: System MUST support configurable resource limits (CPU, memory) for provisioned nodes.
- **FR-016**: System MUST provide an optional OfflineSigningNode CRD for managing cold key signing ceremonies in isolated namespaces, generating unsigned transaction payloads and accepting signed responses.
- **FR-017**: System MUST support three node types: block-producer (mints blocks, requires KES/VRF keys), relay (public-facing, forwards transactions), and offline-signing (isolated, handles cold key operations).
- **FR-018**: System MUST manage PersistentVolumeClaims for blockchain data storage with configurable StorageClass, size, and access modes.
- **FR-019**: System MUST support hybrid payment modes per-pool: (1) automated transactions using user-provided payment key reference, or (2) unsigned transaction generation for external signing and submission.

### Key Entities

- **StakePool**: Represents the desired state of a Cardano stake pool including pool parameters (pledge, margin, cost), metadata reference, relay configuration, and lifecycle state.
- **CardanoNode**: Represents a Cardano node instance with type (block-producer, relay, or offline-signing), configuration for network, topology, resource allocation, and security isolation level.
- **OfflineSigningNode**: Represents an isolated signing environment for cold key operations, managing signing request queue, approval workflow, and signed response ingestion.
- **PoolKeys**: Represents the cryptographic key set for a stake pool including cold key reference, KES key, VRF key, and operational certificate.
- **KESRotation**: Represents the KES key rotation schedule and status including current KES period, expiration epoch, and rotation history.

## Assumptions

- Users have a Kubernetes cluster with sufficient resources to run Cardano nodes (minimum 16GB RAM, 4 CPU cores per node recommended).
- Users have access to ADA funds for pool registration deposit (500 ADA) and transaction fees.
- The operator will support the current Cardano mainnet protocol version and the most recent testnet.
- External secret management (Kubernetes Secrets by default) is available for key storage.
- Users are responsible for providing reliable network connectivity and storage for blockchain data.
- Pool metadata (name, description, ticker) will be hosted externally by the user at a URL they control.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can deploy a fully operational stake pool within 30 minutes of applying a StakePool resource (assuming synced blockchain data is available).
- **SC-002**: Stake pools deployed by the operator achieve 99.9% uptime for block production (excluding network-wide outages).
- **SC-003**: KES key rotation completes automatically with zero missed blocks due to expired keys.
- **SC-004**: Pool parameter updates are reflected on-chain within 2 epochs of resource modification.
- **SC-005**: Users can monitor pool health through standard tooling without additional configuration.
- **SC-006**: Pool retirement completes successfully with deposit returned within the protocol-specified timeframe.
- **SC-007**: 90% of first-time users can deploy a testnet stake pool without external documentation beyond the resource specification.
