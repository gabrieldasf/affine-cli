# Phase 3: Canvas Transform Dry-Run Plans

## Goal

Add `canvas transform` as a dry-run planner that consumes selectors and emits machine-applyable operation plans.

## Context

This phase bridges diagnosis and mutation. It must not write live documents. Its output becomes the input contract for `canvas apply --dry-run` and later `canvas apply --live`.

## Work Items

1. Define the operation plan schema with plan ID, source doc metadata, affected IDs, operations, before/after values, integrity status, backup target, rollback/proof fields, and warnings.
2. Implement transform operations: move, resize, align, distribute, set display mode, and selected metadata updates.
3. Accept selector JSON from `canvas search` and optionally direct IDs.
4. Add `canvas apply --dry-run` support for operation plans while preserving existing plan behavior or adding compatibility routing.
5. Validate transforms against geometry, connector endpoints, hierarchy, and integrity rules.
6. Add tests proving dry-run plans do not call live push APIs.
7. Update docs, MCP metadata, command index, and patch records.

## Subagent Waves

- Wave A, max 2: one implementation worker for transform schema/operations, one verifier for no-mutation and validation tests.
- Wave B, max 2: one plan checker for operation contract, one integration reviewer for apply dry-run compatibility.

## Verification

- `go test ./internal/canvaswrite ./internal/cli`
- Dry-run output is deterministic and accepted by `canvas apply --dry-run`.
- Invalid selectors and unsafe transforms fail before operation plan output.
- No Socket.IO push path is reachable from transform commands.

## Exit Criteria

Transform plans are specific enough for Phase 4 live apply without inventing a second mutation contract.
