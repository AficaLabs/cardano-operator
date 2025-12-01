# Tasks: Cardano Stake Pool Kubernetes Operator

**Input**: Design documents from `/specs/001-cardano-stakepool-operator/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included per Constitution Principle V (Test-Driven Reliability) - 70% coverage required for controller packages.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4, US5)
- Include exact file paths in descriptions

## Path Conventions

- **Operator project**: Standard operator-sdk/kubebuilder layout
- `api/v1alpha1/` - CRD type definitions
- `internal/controller/` - Controller implementations
- `internal/cardano/` - Cardano CLI wrapper and business logic
- `config/` - Kubernetes manifests
- `charts/` - Helm chart

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization with operator-sdk and basic structure

- [x] T001 Initialize operator-sdk project with `operator-sdk init --domain cardano.org --repo github.com/user/cardano-operator` in repository root
- [x] T002 Configure Go 1.22+ in go.mod with required dependencies (controller-runtime, client-go, zap)
- [x] T003 [P] Create Makefile with targets: generate, manifests, build, test, docker-build, helm-package
- [x] T004 [P] Configure golangci-lint with staticcheck, go vet, gofmt rules in .golangci.yml
- [x] T005 [P] Create Dockerfile with multi-stage build for linux/amd64 and linux/arm64 in Dockerfile
- [x] T006 [P] Create GitHub Actions CI workflow in .github/workflows/ci.yaml (lint, test, build)
- [x] T007 [P] Create GitHub Actions release workflow in .github/workflows/release.yaml (GoReleaser, multi-arch)
- [x] T008 [P] Create GitHub Actions security workflow in .github/workflows/security.yaml (gosec, trivy, govulncheck)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### CRD Type Definitions

- [ ] T009 Create groupversion_info.go with cardano.org/v1alpha1 group in api/v1alpha1/groupversion_info.go
- [ ] T010 [P] Define StakePool CRD types (Spec, Status, PoolParams, NodeConfig, KeyManagement, PaymentConfig, StorageConfig) in api/v1alpha1/stakepool_types.go
- [ ] T011 [P] Define CardanoNode CRD types (Spec, Status, TopologyConfig) in api/v1alpha1/cardanonode_types.go
- [ ] T012 [P] Define OfflineSigningNode CRD types (Spec, Status, SigningRequest) in api/v1alpha1/offlinesigningnode_types.go
- [ ] T013 [P] Define KESRotation CRD types (Spec, Status, RotationEvent) in api/v1alpha1/kesrotation_types.go
- [ ] T014 Run `make generate` to create zz_generated.deepcopy.go
- [ ] T015 Run `make manifests` to generate CRD YAMLs in config/crd/bases/

### Cardano CLI Integration Layer

- [ ] T016 Define Cardano client interface (QueryTip, QueryUTxO, BuildTx, SignTx, SubmitTx, GenerateKeys) in internal/cardano/client.go
- [ ] T017 Implement cardano-cli command executor with timeout and error handling in internal/cardano/executor.go
- [ ] T018 [P] Implement key generation functions (cold, KES, VRF, payment keys) in internal/cardano/keys.go
- [ ] T019 [P] Implement transaction building functions (registration, update, retirement, KES cert) in internal/cardano/transaction.go
- [ ] T020 [P] Implement protocol parameter queries and epoch tracking in internal/cardano/protocol.go
- [ ] T021 [P] Implement UTxO queries and balance checking in internal/cardano/utxo.go

### Observability Infrastructure

- [ ] T022 Define Prometheus metrics (stakepool_blocks_minted, node_sync_progress, kes_expiry_epoch, peer_count) in internal/metrics/metrics.go
- [ ] T023 [P] Configure structured logging with zap in cmd/main.go

### Validation Webhooks

- [ ] T024 Implement StakePool validation webhook (pledge format, margin range, cost minimum, relay count) in internal/webhook/stakepool_webhook.go
- [ ] T025 [P] Implement CardanoNode validation webhook (type enum, storage size) in internal/webhook/cardanonode_webhook.go

### Helm Chart Skeleton

- [ ] T026 Create Helm chart structure with Chart.yaml, values.yaml in charts/cardano-operator/
- [ ] T027 [P] Create Helm templates for Deployment, ServiceAccount, RBAC in charts/cardano-operator/templates/
- [ ] T028 [P] Create values.schema.json for Helm values validation in charts/cardano-operator/

### Test Infrastructure

- [ ] T029 Configure envtest suite with test environment setup in internal/controller/suite_test.go
- [ ] T030 [P] Create test utilities for fake Cardano client in test/utils/fake_cardano.go
- [ ] T031 [P] Create test fixtures for sample CRs in test/utils/fixtures.go

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Deploy a New Stake Pool (Priority: P1) 🎯 MVP

**Goal**: Enable declarative stake pool deployment via StakePool CRD

**Independent Test**: Apply a StakePool CR and verify block producer, relay nodes, keys generated, and pool registered on testnet

### Tests for User Story 1

- [ ] T032 [P] [US1] Unit test for StakePool controller reconciliation logic in internal/controller/stakepool_controller_test.go
- [ ] T033 [P] [US1] Unit test for CardanoNode controller reconciliation logic in internal/controller/cardanonode_controller_test.go
- [ ] T034 [P] [US1] Unit test for key generation in internal/cardano/keys_test.go
- [ ] T035 [P] [US1] Unit test for transaction building in internal/cardano/transaction_test.go
- [ ] T036 [US1] Integration test for StakePool creation flow in test/e2e/stakepool_create_test.go

### Implementation for User Story 1

- [ ] T037 [US1] Implement StakePool controller scaffold with Reconcile method in internal/controller/stakepool_controller.go
- [ ] T038 [US1] Implement phase transitions (Pending → Provisioning → Syncing → Registering → Active) in internal/controller/stakepool_controller.go
- [ ] T039 [US1] Implement CardanoNode creation from StakePool spec (block producer + relays) in internal/controller/stakepool_controller.go
- [ ] T040 [US1] Implement key generation workflow (cold, VRF, KES for managed mode) in internal/controller/stakepool_controller.go
- [ ] T041 [US1] Implement pool registration transaction building and submission in internal/controller/stakepool_controller.go
- [ ] T042 [US1] Implement CardanoNode controller scaffold with Reconcile method in internal/controller/cardanonode_controller.go
- [ ] T043 [US1] Implement Deployment creation for Cardano node pods in internal/controller/cardanonode_controller.go
- [ ] T044 [US1] Implement Service creation for node endpoints in internal/controller/cardanonode_controller.go
- [ ] T045 [US1] Implement PVC creation for blockchain storage in internal/controller/cardanonode_controller.go
- [ ] T046 [US1] Implement NetworkPolicy for block producer isolation in internal/controller/cardanonode_controller.go
- [ ] T047 [US1] Implement node sync status monitoring and status updates in internal/controller/cardanonode_controller.go
- [ ] T048 [US1] Implement OfflineSigningNode controller for external key mode in internal/controller/offlinesigningnode_controller.go
- [ ] T049 [US1] Implement signing request workflow (pending → awaiting-signature → signed → submitted) in internal/controller/offlinesigningnode_controller.go
- [ ] T050 [US1] Implement finalizers for StakePool cleanup in internal/controller/stakepool_controller.go
- [ ] T051 [US1] Implement owner references for child resources (CardanoNode, Secrets) in internal/controller/stakepool_controller.go
- [ ] T052 [US1] Add Kubernetes Events for state transitions in internal/controller/stakepool_controller.go
- [ ] T053 [US1] Create sample StakePool CR for testnet in config/samples/cardano_v1alpha1_stakepool.yaml
- [ ] T054 [US1] Create sample StakePool CR with external signing in config/samples/cardano_v1alpha1_stakepool_external.yaml

**Checkpoint**: User Story 1 complete - stake pool can be deployed via CR

---

## Phase 4: User Story 2 - Monitor Pool Health and Performance (Priority: P2)

**Goal**: Expose Prometheus metrics for pool health, sync status, and block production

**Independent Test**: Deploy a stake pool and verify metrics are exposed at /metrics endpoint

### Tests for User Story 2

- [ ] T055 [P] [US2] Unit test for metrics registration and updates in internal/metrics/metrics_test.go
- [ ] T056 [US2] Integration test for metrics endpoint in test/e2e/metrics_test.go

### Implementation for User Story 2

- [ ] T057 [US2] Implement node sync progress metric collection in internal/controller/cardanonode_controller.go
- [ ] T058 [US2] Implement peer count metric collection in internal/controller/cardanonode_controller.go
- [ ] T059 [US2] Implement block minted counter metric in internal/controller/stakepool_controller.go
- [ ] T060 [US2] Implement KES expiry epoch gauge metric in internal/controller/stakepool_controller.go
- [ ] T061 [US2] Implement tip slot and epoch metrics in internal/controller/cardanonode_controller.go
- [ ] T062 [US2] Add ServiceMonitor CR for Prometheus Operator integration in config/prometheus/monitor.yaml
- [ ] T063 [US2] Add Grafana dashboard JSON in config/grafana/dashboard.json
- [ ] T064 [US2] Add alert rules for degraded connectivity and KES expiration in config/prometheus/alerts.yaml

**Checkpoint**: User Story 2 complete - pool metrics exposed via Prometheus

---

## Phase 5: User Story 3 - Automated KES Key Rotation (Priority: P2)

**Goal**: Automatically rotate KES keys before expiration to prevent block production interruption

**Independent Test**: Deploy a pool and verify KES rotation triggers automatically as expiration approaches

### Tests for User Story 3

- [ ] T065 [P] [US3] Unit test for KESRotation controller reconciliation in internal/controller/kesrotation_controller_test.go
- [ ] T066 [P] [US3] Unit test for KES key generation and certificate creation in internal/cardano/keys_test.go
- [ ] T067 [US3] Integration test for KES rotation workflow in test/e2e/kesrotation_test.go

### Implementation for User Story 3

- [ ] T068 [US3] Implement KESRotation controller scaffold with Reconcile method in internal/controller/kesrotation_controller.go
- [ ] T069 [US3] Implement KES expiration monitoring (check epochs remaining) in internal/controller/kesrotation_controller.go
- [ ] T070 [US3] Implement automatic rotation trigger based on lead epochs threshold in internal/controller/kesrotation_controller.go
- [ ] T071 [US3] Implement new KES key generation in internal/controller/kesrotation_controller.go
- [ ] T072 [US3] Implement operational certificate creation with cardano-cli in internal/controller/kesrotation_controller.go
- [ ] T073 [US3] Implement block producer pod restart with new certificate in internal/controller/kesrotation_controller.go
- [ ] T074 [US3] Implement rotation history tracking in KESRotation status in internal/controller/kesrotation_controller.go
- [ ] T075 [US3] Implement rotation failure handling and alerting in internal/controller/kesrotation_controller.go
- [ ] T076 [US3] Add KESRotation creation from StakePool controller in internal/controller/stakepool_controller.go

**Checkpoint**: User Story 3 complete - KES keys rotate automatically

---

## Phase 6: User Story 4 - Update Pool Parameters (Priority: P3)

**Goal**: Allow pool parameter updates via StakePool spec modification

**Independent Test**: Modify StakePool spec and verify update transaction is submitted on-chain

### Tests for User Story 4

- [ ] T077 [P] [US4] Unit test for pool update detection in internal/controller/stakepool_controller_test.go
- [ ] T078 [P] [US4] Unit test for pool update certificate building in internal/cardano/transaction_test.go
- [ ] T079 [US4] Integration test for pool parameter update flow in test/e2e/stakepool_update_test.go

### Implementation for User Story 4

- [ ] T080 [US4] Implement spec change detection (compare current vs desired pool params) in internal/controller/stakepool_controller.go
- [ ] T081 [US4] Implement pool update certificate building in internal/cardano/transaction.go
- [ ] T082 [US4] Implement update transaction submission workflow in internal/controller/stakepool_controller.go
- [ ] T083 [US4] Implement metadata hash computation for URL changes in internal/cardano/metadata.go
- [ ] T084 [US4] Implement relay update handling in internal/controller/stakepool_controller.go
- [ ] T085 [US4] Implement pledge/margin/cost update validation in internal/webhook/stakepool_webhook.go
- [ ] T086 [US4] Add "Updating" phase handling in status transitions in internal/controller/stakepool_controller.go

**Checkpoint**: User Story 4 complete - pool parameters can be updated via CR

---

## Phase 7: User Story 5 - Retire a Stake Pool (Priority: P3)

**Goal**: Gracefully retire pool when StakePool resource is deleted

**Independent Test**: Delete StakePool CR and verify retirement certificate is submitted and resources cleaned up

### Tests for User Story 5

- [ ] T087 [P] [US5] Unit test for retirement certificate building in internal/cardano/transaction_test.go
- [ ] T088 [P] [US5] Unit test for finalizer cleanup logic in internal/controller/stakepool_controller_test.go
- [ ] T089 [US5] Integration test for pool retirement flow in test/e2e/stakepool_retire_test.go

### Implementation for User Story 5

- [ ] T090 [US5] Implement retirement certificate building with epoch selection in internal/cardano/transaction.go
- [ ] T091 [US5] Implement finalizer-based retirement workflow in internal/controller/stakepool_controller.go
- [ ] T092 [US5] Implement "Retiring" phase with retirement epoch tracking in internal/controller/stakepool_controller.go
- [ ] T093 [US5] Implement resource cleanup after retirement epoch in internal/controller/stakepool_controller.go
- [ ] T094 [US5] Implement retirement cancellation (re-registration if CR recreated) in internal/controller/stakepool_controller.go
- [ ] T095 [US5] Add retirement events and status updates in internal/controller/stakepool_controller.go

**Checkpoint**: User Story 5 complete - pools can be retired gracefully

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Production readiness, documentation, and hardening

- [ ] T096 [P] Add godoc comments to all public types in api/v1alpha1/*.go
- [ ] T097 [P] Add godoc comments to all exported functions in internal/cardano/*.go
- [ ] T098 [P] Create README.md with installation and usage instructions
- [ ] T099 [P] Create CONTRIBUTING.md with development setup guide
- [ ] T100 Update Helm chart with all configurable values in charts/cardano-operator/values.yaml
- [ ] T101 [P] Add Helm chart README with configuration reference in charts/cardano-operator/README.md
- [ ] T102 Add leader election configuration for HA deployment in cmd/main.go
- [ ] T103 Add health and readiness probes in cmd/main.go
- [ ] T104 Run full e2e test suite against preview testnet
- [ ] T105 Verify >70% test coverage for controller packages
- [ ] T106 Security audit: verify keys never logged, no secrets in status
- [ ] T107 Create GitHub release with changelog for v0.1.0

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - User stories can proceed in parallel OR sequentially (P1 → P2 → P3)
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories. **MVP**
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Uses infrastructure from US1 but independently testable
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - Adds to US1 functionality but independently testable
- **User Story 4 (P3)**: Can start after Foundational (Phase 2) - Extends US1 controller, independently testable
- **User Story 5 (P3)**: Can start after Foundational (Phase 2) - Extends US1 controller, independently testable

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- CRD types before controllers
- Client/library code before controllers that use it
- Core implementation before edge cases
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel (T003-T008)
- All CRD type definitions marked [P] can run in parallel (T010-T013)
- All Cardano client functions marked [P] can run in parallel (T018-T021)
- Tests within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel after Foundational phase

---

## Parallel Example: Phase 2 Foundational

```bash
# Launch all CRD type definitions together:
Task T010: "Define StakePool CRD types in api/v1alpha1/stakepool_types.go"
Task T011: "Define CardanoNode CRD types in api/v1alpha1/cardanonode_types.go"
Task T012: "Define OfflineSigningNode CRD types in api/v1alpha1/offlinesigningnode_types.go"
Task T013: "Define KESRotation CRD types in api/v1alpha1/kesrotation_types.go"

# Then generate code (depends on above):
Task T014: "Run make generate"
Task T015: "Run make manifests"

# Launch all Cardano client functions together:
Task T018: "Implement key generation in internal/cardano/keys.go"
Task T019: "Implement transaction building in internal/cardano/transaction.go"
Task T020: "Implement protocol queries in internal/cardano/protocol.go"
Task T021: "Implement UTxO queries in internal/cardano/utxo.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Deploy Stake Pool)
4. **STOP and VALIDATE**: Test stake pool deployment on preview testnet
5. Release v0.1.0-alpha

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Release v0.1.0-alpha (MVP!)
3. Add User Stories 2 + 3 → Test independently → Release v0.2.0 (Monitoring + KES)
4. Add User Stories 4 + 5 → Test independently → Release v0.3.0 (Full Lifecycle)
5. Complete Polish → Release v1.0.0

### Parallel Team Strategy

With multiple developers after Foundational:
- Developer A: User Story 1 (core deployment)
- Developer B: User Story 2 (metrics/monitoring)
- Developer C: User Story 3 (KES rotation)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Constitution requires: 70% test coverage, no keys in logs, idempotent reconciliation
