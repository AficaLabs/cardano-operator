# Research: Cardano Stake Pool Kubernetes Operator

**Date**: 2025-11-30
**Branch**: `001-cardano-stakepool-operator`

## Overview

This document captures research findings for implementing the Cardano Stake Pool Kubernetes Operator using Go and operator-sdk, following latest best practices for security, CI/CD, and release management.

---

## 1. Operator SDK & Project Structure

### Decision: operator-sdk v1.37+ with kubebuilder v4 scaffolding

**Rationale**: operator-sdk v1.37+ uses kubebuilder v4 under the hood, providing:
- Go 1.22+ support with improved generics
- controller-runtime v0.18+ with enhanced caching
- Simplified project layout with `internal/controller/` convention
- Built-in support for multi-group APIs
- Improved webhook scaffolding

**Alternatives Considered**:
- Pure kubebuilder: Viable but operator-sdk adds OLM integration, scorecard testing
- Metacontroller: Too limited for complex reconciliation logic
- KUDO: Abandoned/less active development

**Implementation**:
```bash
operator-sdk init --domain cardano.org --repo github.com/user/cardano-operator
operator-sdk create api --group cardano --version v1alpha1 --kind StakePool --resource --controller
operator-sdk create api --group cardano --version v1alpha1 --kind CardanoNode --resource --controller
operator-sdk create api --group cardano --version v1alpha1 --kind OfflineSigningNode --resource --controller
operator-sdk create api --group cardano --version v1alpha1 --kind KESRotation --resource --controller
```

---

## 2. Controller-Runtime Reconciliation Patterns

### Decision: Event-driven reconciliation with rate limiting and exponential backoff

**Rationale**:
- Watches on owned resources (Deployments, Services, PVCs) trigger reconciliation
- Rate limiting prevents thundering herd during cluster events
- Exponential backoff for transient failures (network, Cardano node sync)

**Key Patterns**:

| Pattern | Usage |
|---------|-------|
| `ctrl.Result{RequeueAfter: time.Minute}` | Periodic re-check for KES expiration |
| `ctrl.Result{Requeue: true}` | Immediate retry after transient error |
| Finalizers | Cleanup pool retirement, remove on-chain registration |
| Status conditions | Track phase: Pending → Syncing → Registered → Active |
| Owner references | Auto-cleanup of Deployments, Services, PVCs, Secrets |

**Alternatives Considered**:
- Polling-based reconciliation: Less efficient, higher API server load
- External queue (Redis/NATS): Over-engineering for this use case

---

## 3. CRD Design Best Practices

### Decision: Status subresource + validation webhooks + printer columns

**Rationale**:
- Status subresource enables RBAC separation (spec vs status updates)
- Validation webhooks catch errors before persistence
- Printer columns improve `kubectl get` UX

**CRD Features to Implement**:

| Feature | Implementation |
|---------|----------------|
| Status subresource | `+kubebuilder:subresource:status` |
| Validation | `+kubebuilder:validation:*` markers + webhook |
| Default values | `+kubebuilder:default:=` markers |
| Printer columns | `+kubebuilder:printcolumn:*` for pool ID, status, epoch |
| Short names | `+kubebuilder:resource:shortName=sp` for StakePool |
| Categories | `+kubebuilder:resource:categories=cardano` |

**Status Conditions Pattern**:
```go
// Standard condition types for StakePool
const (
    ConditionReady          = "Ready"
    ConditionNodesProvisioned = "NodesProvisioned"
    ConditionKeysGenerated  = "KeysGenerated"
    ConditionPoolRegistered = "PoolRegistered"
    ConditionKESValid       = "KESValid"
)
```

---

## 4. Testing Strategy

### Decision: Three-tier testing with envtest, kind, and testnet

**Rationale**:
- Unit tests with envtest: Fast, no cluster needed, test controller logic
- Integration tests with kind: Real K8s API, NetworkPolicy validation
- E2E tests with Cardano testnet: Full protocol validation

**Test Framework Stack**:

| Layer | Tool | Purpose |
|-------|------|---------|
| Unit | `go test` + envtest | Controller reconciliation logic |
| Integration | kind + Ginkgo | Multi-resource interactions |
| E2E | kind + cardano-node (preview) | Full stake pool lifecycle |
| Contract | OpenAPI validation | CRD schema compliance |

**Coverage Requirements** (per constitution):
- Unit: >70% for `internal/controller/` and `internal/cardano/`
- Integration: All happy paths + key error scenarios
- E2E: Complete lifecycle (create → register → rotate KES → retire)

**Alternatives Considered**:
- Only unit tests: Misses K8s API interactions
- Only e2e: Too slow for CI, expensive testnet resources
- Mock Cardano: Doesn't validate protocol compliance

---

## 5. Security Best Practices

### Decision: Minimal RBAC + secret encryption + network isolation

**Key Security Measures**:

| Area | Implementation |
|------|----------------|
| RBAC | Namespace-scoped by default; cluster-wide optional |
| Secrets | External Secrets Operator integration; never log key material |
| Network | NetworkPolicy: block-producer isolated from public |
| Images | Distroless base; non-root user; read-only filesystem |
| Supply chain | Signed images (cosign); SBOM generation |
| Scanning | Trivy (vulnerabilities), gosec (SAST), govulncheck |

**RBAC Minimal Privileges**:
```yaml
# Only required permissions
- apiGroups: [""]
  resources: ["pods", "services", "persistentvolumeclaims", "secrets", "configmaps"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "statefulsets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["networking.k8s.io"]
  resources: ["networkpolicies"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

**Secret Handling**:
- Keys stored in Kubernetes Secrets with `kubernetes.io/tls` or custom type
- Support for External Secrets Operator (AWS SM, Vault, GCP SM)
- Keys never appear in: logs, events, status fields, metrics labels
- Memory-only handling during signing operations

---

## 6. CI/CD Pipeline Design

### Decision: GitHub Actions with semantic-release and GoReleaser

**Rationale**:
- GitHub Actions: Native integration, free for public repos, good operator ecosystem
- semantic-release: Automated versioning from conventional commits
- GoReleaser: Multi-arch builds, checksums, changelog generation

**Pipeline Structure**:

| Workflow | Trigger | Actions |
|----------|---------|---------|
| `ci.yaml` | PR, push to main | Lint, test, build, security scan |
| `release.yaml` | Tag push (v*) | GoReleaser, Docker build, Helm push |
| `security.yaml` | Schedule (daily) | Dependency audit, CVE scan |
| `e2e.yaml` | PR (label), manual | Full e2e with testnet |

**CI Pipeline Steps**:
```yaml
jobs:
  lint:
    - golangci-lint (includes go vet, staticcheck, gofmt)
    - hadolint (Dockerfile)
    - helm lint

  test:
    - go test -race -coverprofile
    - envtest (controller tests)
    - Upload coverage to Codecov

  security:
    - gosec (SAST)
    - govulncheck (dependency CVEs)
    - trivy (container scan)

  build:
    - Docker build (linux/amd64, linux/arm64)
    - Push to ghcr.io (on main)
```

**Alternatives Considered**:
- GitLab CI: Good but less K8s operator ecosystem tooling
- CircleCI: Cost concerns for open source
- Jenkins: Too much maintenance overhead

---

## 7. Multi-Architecture Container Builds

### Decision: Docker buildx with manifest lists

**Rationale**:
- Native buildx support in GitHub Actions
- Single tag serves both amd64 and arm64
- Supports Apple Silicon development and ARM-based K8s nodes

**Dockerfile Strategy**:
```dockerfile
# Multi-stage build
FROM --platform=$BUILDPLATFORM golang:1.22 AS builder
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o manager cmd/main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /workspace/manager /manager
USER 65532:65532
ENTRYPOINT ["/manager"]
```

**Build Matrix**:
- `linux/amd64`: Primary production target
- `linux/arm64`: ARM servers, Apple Silicon dev

---

## 8. Helm Chart Distribution

### Decision: OCI registry (ghcr.io) + Artifact Hub listing

**Rationale**:
- OCI registry is the modern standard (Helm 3.8+)
- ghcr.io provides unified container + chart hosting
- Artifact Hub increases discoverability

**Chart Structure**:
```text
charts/cardano-operator/
├── Chart.yaml
├── values.yaml
├── values.schema.json      # JSON Schema for values validation
├── templates/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── rbac.yaml
│   ├── crds.yaml           # Or separate CRD chart
│   └── _helpers.tpl
├── crds/                   # CRDs as separate files (Helm 3 pattern)
└── README.md
```

**Distribution**:
```bash
# Push to OCI registry
helm package charts/cardano-operator
helm push cardano-operator-*.tgz oci://ghcr.io/user/charts

# Install
helm install cardano-operator oci://ghcr.io/user/charts/cardano-operator --version 0.1.0
```

---

## 9. Cardano-Specific Integration

### Decision: cardano-cli via sidecar container + cardano-node for queries

**Rationale**:
- cardano-cli is the canonical tool for transaction building
- Avoids reimplementing complex ledger logic
- Sidecar pattern allows version pinning per pool

**Integration Architecture**:

| Component | Purpose | Container |
|-----------|---------|-----------|
| cardano-node | Blockchain sync, queries | StatefulSet per node type |
| cardano-cli | Transaction building, signing | Init/sidecar in operator |
| ogmios (optional) | WebSocket API for queries | Optional deployment |

**CLI Wrapper Design**:
```go
// internal/cardano/client.go
type Client interface {
    QueryTip(ctx context.Context) (*Tip, error)
    QueryUTxO(ctx context.Context, address string) ([]UTxO, error)
    BuildTx(ctx context.Context, tx *TxBuilder) ([]byte, error)
    SignTx(ctx context.Context, txBody []byte, signingKey []byte) ([]byte, error)
    SubmitTx(ctx context.Context, signedTx []byte) (string, error)
    GenerateKeys(ctx context.Context, keyType KeyType) (*KeyPair, error)
}
```

---

## 10. Air-Gap Signing Workflow

### Decision: Signing request CRD with status-based workflow

**Rationale**:
- Declarative model fits Kubernetes patterns
- Status transitions track signing ceremony progress
- Supports both automated (in-cluster) and manual (air-gap) modes

**Workflow States**:
```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐     ┌───────────┐
│ Pending     │ ──▶ │ Unsigned     │ ──▶ │ Awaiting    │ ──▶ │ Signed    │
│ (created)   │     │ (tx built)   │     │ Signature   │     │ (complete)│
└─────────────┘     └──────────────┘     └─────────────┘     └───────────┘
                                               │
                                               ▼
                                         [User signs offline,
                                          uploads signature]
```

**OfflineSigningNode Status Fields**:
```go
type OfflineSigningNodeStatus struct {
    Phase              string    `json:"phase"`
    UnsignedTxHash     string    `json:"unsignedTxHash,omitempty"`
    UnsignedTxBody     string    `json:"unsignedTxBody,omitempty"`  // Base64
    SignedTx           string    `json:"signedTx,omitempty"`        // User provides
    TxHash             string    `json:"txHash,omitempty"`          // After submission
    LastTransitionTime metav1.Time `json:"lastTransitionTime"`
}
```

---

## Summary of Key Decisions

| Area | Decision |
|------|----------|
| Framework | operator-sdk v1.37+ with kubebuilder v4 |
| Testing | envtest + kind + Cardano testnet (3-tier) |
| CI/CD | GitHub Actions + semantic-release + GoReleaser |
| Container | Multi-arch (amd64/arm64) distroless images |
| Distribution | Helm via OCI registry (ghcr.io) |
| Security | Minimal RBAC, secret encryption, NetworkPolicy |
| Cardano CLI | Sidecar container with version pinning |
| Air-gap | OfflineSigningNode CRD with status workflow |

---

## References

- [operator-sdk documentation](https://sdk.operatorframework.io/)
- [kubebuilder book](https://book.kubebuilder.io/)
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)
- [Cardano Developer Portal](https://developers.cardano.org/)
- [cardano-cli documentation](https://github.com/IntersectMBO/cardano-cli)
