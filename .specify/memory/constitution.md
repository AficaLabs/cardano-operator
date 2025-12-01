<!--
Sync Impact Report
==================
Version change: N/A → 1.0.0 (Initial)
Modified principles: N/A (new constitution)
Added sections:
  - Core Principles (5 principles)
  - Kubernetes Operator Standards
  - Development Workflow
  - Governance
Removed sections: N/A
Templates requiring updates:
  - .specify/templates/plan-template.md: ✅ Compatible (no changes needed)
  - .specify/templates/spec-template.md: ✅ Compatible (no changes needed)
  - .specify/templates/tasks-template.md: ✅ Compatible (no changes needed)
Follow-up TODOs: None
-->

# Cardano Operator Constitution

## Core Principles

### I. Kubernetes-Native Design

All functionality MUST be implemented using Kubernetes-native patterns and APIs. The operator
MUST use Custom Resource Definitions (CRDs) to represent Cardano stake pool configurations.
Reconciliation loops MUST be idempotent and converge to the desired state. The operator MUST
integrate with standard Kubernetes tooling (kubectl, Helm, GitOps workflows).

**Rationale**: Kubernetes operators provide declarative infrastructure management, automatic
healing, and integration with the broader cloud-native ecosystem. Users expect consistent
Kubernetes semantics.

### II. Security-First Cryptographic Handling

All cryptographic keys (cold keys, KES keys, VRF keys, payment keys) MUST be handled with
defense-in-depth security. Key generation MUST occur in isolated, auditable contexts. Private
keys MUST never be logged, exposed in status fields, or transmitted unencrypted. The operator
MUST support integration with external secret management (Kubernetes Secrets, HashiCorp Vault,
AWS Secrets Manager). KES key rotation MUST be automated with configurable lead time before
expiration.

**Rationale**: Stake pool keys control significant financial assets. Compromise leads to
irreversible loss. Security is non-negotiable.

### III. Cardano Protocol Compliance

The operator MUST maintain compatibility with the current Cardano mainnet and testnet protocol
parameters. All stake pool registration, updates, and retirement operations MUST produce valid
transactions per the Cardano ledger specification. The operator MUST handle protocol parameter
changes (epoch boundaries, k parameter, pledge influence) gracefully. Hard fork combinator
(HFC) events MUST be supported without manual intervention.

**Rationale**: Invalid transactions waste fees and fail silently. Protocol non-compliance can
result in stake pool deregistration or missed rewards.

### IV. Operational Observability

The operator MUST expose Prometheus metrics for stake pool health, block production, rewards,
and operational status. All reconciliation events MUST be logged with structured logging at
appropriate levels. The operator MUST emit Kubernetes Events for significant state transitions
(pool registered, KES rotated, blocks minted). Alert-ready metrics MUST include: missed slot
leader checks, KES key expiration countdown, node sync status, and peer connectivity.

**Rationale**: Stake pool operators need visibility into pool performance to maximize rewards
and quickly diagnose issues. Silent failures are unacceptable in financial infrastructure.

### V. Test-Driven Reliability

All CRD controllers MUST have unit tests covering reconciliation logic. Integration tests MUST
validate against a local Cardano testnet or mock. End-to-end tests MUST verify complete stake
pool lifecycle (creation, operation, retirement). Tests MUST run in CI before any merge to
main. Mutation testing or coverage thresholds SHOULD be enforced for critical paths (key
handling, transaction building).

**Rationale**: Financial infrastructure demands high reliability. Untested code is a liability.
Operators run continuously; bugs compound over time.

## Kubernetes Operator Standards

### Custom Resource Definitions

- **StakePool**: Primary CRD representing a Cardano stake pool with desired configuration
- **CardanoNode**: CRD for managing Cardano node deployments (relay and block-producer)
- **KESRotation**: CRD for managing KES key operational certificates

### Controller Requirements

- Controllers MUST implement the controller-runtime Reconciler interface
- Reconciliation MUST be idempotent (same input produces same output, re-runnable safely)
- Controllers MUST use finalizers for cleanup of external resources
- Status subresources MUST reflect actual observed state, not desired state
- Controllers MUST implement leader election for high availability deployments

### Resource Management

- All created resources MUST have owner references for garbage collection
- Resource requests and limits MUST be configurable via CRD spec
- The operator MUST support running in a dedicated namespace or cluster-wide

## Development Workflow

### Code Quality Gates

- All code MUST pass `go vet` and `staticcheck` with zero warnings
- All code MUST be formatted with `gofmt`
- All public APIs MUST have godoc comments
- Dependency updates MUST be reviewed for security advisories

### Testing Requirements

- Unit test coverage MUST exceed 70% for controller packages
- Integration tests MUST use envtest or kind clusters
- E2E tests MUST complete stake pool lifecycle in under 30 minutes

### Release Process

- Semantic versioning MUST be used for operator releases
- Breaking CRD changes MUST increment the API version (v1alpha1 -> v1beta1 -> v1)
- Release notes MUST document upgrade paths and breaking changes

## Governance

This constitution defines the non-negotiable standards for the Cardano Operator project. All
pull requests and code reviews MUST verify compliance with these principles.

### Amendment Process

1. Proposed amendments MUST be documented with rationale
2. Breaking changes to principles require MAJOR version bump to constitution
3. New principles or material expansions require MINOR version bump
4. Clarifications and wording fixes require PATCH version bump

### Compliance Review

- PR reviewers MUST check constitution compliance before approval
- Complexity that violates principles MUST be justified in writing
- Unjustified violations MUST be rejected

### Versioning Policy

- MAJOR: Removal or redefinition of core principles
- MINOR: Addition of new principles or sections
- PATCH: Clarifications, typo fixes, non-semantic changes

**Version**: 1.0.0 | **Ratified**: 2025-11-30 | **Last Amended**: 2025-11-30
