# Implementation Plan: Cardano Stake Pool Kubernetes Operator

**Branch**: `001-cardano-stakepool-operator` | **Date**: 2025-11-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-cardano-stakepool-operator/spec.md`

## Summary

Build a Kubernetes operator using Go and operator-sdk that enables declarative deployment and management of Cardano stake pools. The operator manages the complete stake pool lifecycle including node provisioning (block producer, relay, offline-signing), cryptographic key management with air-gap support, automated KES key rotation, pool registration/updates/retirement, and comprehensive observability. The design prioritizes security for cryptographic operations while providing flexibility through hybrid modes for both key management and transaction signing.

## Technical Context

**Language/Version**: Go 1.22+ (latest stable)
**Primary Dependencies**: operator-sdk v1.37+, controller-runtime, kubebuilder, client-go, cardano-cli (via container)
**Storage**: PersistentVolumeClaims for blockchain data (~150GB+), Kubernetes Secrets for keys
**Testing**: go test, envtest, kind clusters, Ginkgo/Gomega for BDD-style tests
**Target Platform**: Kubernetes 1.28+ (Linux amd64/arm64)
**Project Type**: Kubernetes Operator (single project with CRDs)
**Performance Goals**: Reconciliation latency <5s, support 10+ stake pools per cluster
**Constraints**: Block producer network isolation via NetworkPolicy, air-gap support for cold keys
**Scale/Scope**: Single operator managing multiple stake pools, HA deployment with leader election

**CI/CD Requirements** (per user input):
- GitHub Actions for CI/CD pipelines
- Automated testing (unit, integration, e2e)
- Container image builds with multi-arch support
- Semantic versioning with automated releases
- Helm chart packaging and distribution
- Security scanning (trivy, gosec, dependency audit)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Kubernetes-Native Design | ✅ PASS | CRDs for StakePool, CardanoNode, OfflineSigningNode, KESRotation; controller-runtime reconciliation; owner references for GC; Helm/GitOps compatible |
| II. Security-First Cryptographic Handling | ✅ PASS | Hybrid cold key modes (operator-managed/external); OfflineSigningNode CRD for air-gap ceremonies; keys in Secrets (Vault integration planned); keys never logged/exposed in status |
| III. Cardano Protocol Compliance | ✅ PASS | Uses cardano-cli for transaction building; supports mainnet/testnet; handles epoch boundaries; HFC support via node version management |
| IV. Operational Observability | ✅ PASS | Prometheus metrics for all operations; structured logging (zap); Kubernetes Events for state transitions; alert-ready metrics defined |
| V. Test-Driven Reliability | ✅ PASS | Unit tests >70% coverage; envtest for integration; e2e with kind + local testnet; CI gates before merge |

**Controller Requirements Compliance**:
- ✅ controller-runtime Reconciler interface
- ✅ Idempotent reconciliation
- ✅ Finalizers for external resource cleanup
- ✅ Status subresource for observed state
- ✅ Leader election for HA

**Development Workflow Compliance**:
- ✅ go vet, staticcheck, gofmt enforced in CI
- ✅ godoc comments for public APIs
- ✅ Dependency security scanning
- ✅ Semantic versioning for releases
- ✅ API version progression (v1alpha1 → v1beta1 → v1)

## Project Structure

### Documentation (this feature)

```text
specs/001-cardano-stakepool-operator/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (CRD schemas)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
.
├── api/
│   └── v1alpha1/
│       ├── stakepool_types.go
│       ├── cardanonode_types.go
│       ├── offlinesigningnode_types.go
│       ├── kesrotation_types.go
│       ├── groupversion_info.go
│       └── zz_generated.deepcopy.go
├── cmd/
│   └── main.go
├── internal/
│   ├── controller/
│   │   ├── stakepool_controller.go
│   │   ├── cardanonode_controller.go
│   │   ├── offlinesigningnode_controller.go
│   │   ├── kesrotation_controller.go
│   │   └── suite_test.go
│   ├── cardano/
│   │   ├── client.go           # cardano-cli wrapper
│   │   ├── transaction.go      # transaction building
│   │   ├── keys.go             # key generation/management
│   │   └── protocol.go         # protocol parameters
│   ├── metrics/
│   │   └── metrics.go          # Prometheus metrics
│   └── webhook/
│       └── validation.go       # Admission webhooks
├── config/
│   ├── crd/
│   │   └── bases/              # Generated CRD YAMLs
│   ├── manager/
│   │   └── manager.yaml
│   ├── rbac/
│   │   └── role.yaml
│   ├── samples/
│   │   └── cardano_v1alpha1_stakepool.yaml
│   └── default/
│       └── kustomization.yaml
├── charts/
│   └── cardano-operator/
│       ├── Chart.yaml
│       ├── values.yaml
│       └── templates/
├── test/
│   ├── e2e/
│   │   └── stakepool_test.go
│   └── utils/
│       └── testenv.go
├── hack/
│   └── scripts/
├── .github/
│   └── workflows/
│       ├── ci.yaml
│       ├── release.yaml
│       └── security.yaml
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
└── PROJECT                     # kubebuilder project file
```

**Structure Decision**: Standard operator-sdk/kubebuilder layout with:
- `api/v1alpha1/` for CRD type definitions
- `internal/controller/` for reconciliation logic
- `internal/cardano/` for Cardano-specific business logic (CLI wrapper, transactions, keys)
- `config/` for Kubernetes manifests (CRDs, RBAC, samples)
- `charts/` for Helm packaging
- `.github/workflows/` for CI/CD pipelines

## Complexity Tracking

> No violations requiring justification. All complexity is justified by constitution requirements.

| Design Decision | Justification | Constitution Alignment |
|-----------------|---------------|------------------------|
| Multiple CRDs (4) | Each represents a distinct lifecycle and concern | I. Kubernetes-Native (declarative resources) |
| OfflineSigningNode CRD | Air-gap security requirement for cold keys | II. Security-First |
| Hybrid key modes | Balance security (production) vs convenience (testnet) | II. Security-First + usability |
| NetworkPolicy isolation | Block producer must not be public-facing | II. Security-First |
| cardano-cli via container | Protocol compliance without reimplementing ledger | III. Protocol Compliance |
