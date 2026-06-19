# Phase 4: Gated Live Apply And Publish Proof

## Goal

Enable live Canvas apply for validated operation plans and prove the final CLI is safe, documented, and publishable through the Printing Press path.

## Context

Existing `canvas apply` currently requires `--dry-run`. Existing `canvas doc repair` demonstrates the safer mutation model: explicit apply, backup-before, delta artifact, local integrity, push, reload, and post-push integrity. Live apply should follow that stricter model.

## Work Items

1. Add explicit live flags for apply, requiring workspace/doc, operation plan input, backup directory, and confirmation.
2. Run preflight integrity and semantic diff preview before any push.
3. Write backup-before and delta artifacts with restrictive file permissions.
4. Apply only validated operations using Y.js state vector and delta encoding.
5. Reload the document and verify post-apply integrity plus expected semantic diff.
6. Ensure mutating commands are not marked read-only in MCP annotations.
7. Add fixture-gated live smoke tests requiring explicit approval/environment.
8. Add a real Canvas workflow manifest/proof and ensure workflow verification is not skipped.
9. Update README, SKILL, `which`, command help, patch records, changelog/release notes as appropriate.
10. Run publish validation through the current Printing Press workflow.

## Subagent Waves

- Wave A, max 2: one implementation worker for apply gates/live write path, one verifier for safety tests and fixture smoke protocol.
- Wave B, max 2: one security/reliability reviewer for mutation gates, one project standards reviewer for Printing Press publish proof.

## Verification

- `go test ./...` from `library/affine`
- Canvas fixture smoke after explicit approval only
- Dry-run no-mutation regression tests
- Y.js integrity pre/post tests
- Printing Press publish validation
- Workflow proof verification with non-skipped Canvas manifest

## Exit Criteria

Live apply is available only through explicit, backed-up, verified operation plans, and the CLI remains publish-ready under the official Printing Press protocol.
