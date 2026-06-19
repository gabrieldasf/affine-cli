# Phase 2: Canvas Diff And History Diagnosis

## Goal

Add `canvas diff` to compare two canonical Canvas states and diagnose missing, changed, moved, relinked, orphaned, or integrity-damaged entities.

## Context

Diff must compare normalized model state, not raw Y.js bytes. It depends on Phase 1 selector and snapshot semantics.

## Work Items

1. Define diff input modes: live-vs-snapshot, snapshot-vs-snapshot, timestamp-vs-live, timestamp-vs-timestamp.
2. Define diff categories: added, removed, text changed, geometry changed, display mode changed, relinked, image/source changed, orphaned, unreachable, integrity changed.
3. Implement stable diff JSON with severity, affected IDs, before/after values, and suggested next action.
4. Reuse integrity checks to report structural risks alongside semantic diff.
5. Add tests for history/snapshot comparisons and missing Canvas region scenarios.
6. Add docs, examples, MCP read-only annotations, and patch records.
7. Review existing patch ledger drift and remove or replace stale helper references.

## Subagent Waves

- Wave A, max 2: one implementation worker for diff engine, one verifier for fixtures and expected diff categories.
- Wave B, max 2: one integration checker for Phase 1 selector compatibility, one standards reviewer for docs/patch ledger drift.

## Verification

- `go test ./internal/canvaswrite ./internal/cli`
- Diff fixtures cover all core categories.
- Diff refuses ambiguous input pairs with actionable errors.
- No live mutation path is introduced.

## Exit Criteria

`canvas diff --json` produces stable, test-backed diagnosis output that can inform transform planning.
