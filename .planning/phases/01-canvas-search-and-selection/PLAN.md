# Phase 1: Canvas Search And Selection

## Goal

Create a read-only canonical Canvas snapshot model and expose `canvas search` for block/card/connector-level selection.

## Context

Current closest implementation is `canvas doc audit`, which counts keywords and samples text blocks. Search needs a reusable selector contract with IDs, source metadata, geometry, hierarchy, flavour/type, display mode, text matches, connector endpoints, and integrity signals.

## Work Items

1. Extract shared Canvas loading/modeling helpers from current audit, integrity, card inspect, block inspect, and model code.
2. Define stable JSON structs for canonical entities and selectors.
3. Implement `canvas search` for live doc, snapshot file, and history timestamp inputs.
4. Support filters for text, exact ID, flavour, type, display mode, geometry bounds, connector endpoint, and source mode.
5. Fix misleading snapshot examples where local snapshot mode still requires workspace/doc flags unnecessarily.
6. Update README, SKILL, `which`, command help, MCP read-only annotations, and patch records.
7. Add unit and fixture tests for selector output stability.

## Subagent Waves

- Wave A, max 2: one implementation worker for canonical model/search, one verifier for fixture and no-mutation tests.
- Wave B, max 2: one plan checker for selector contract, one project standards reviewer for docs and patch records.

## Verification

- `go test ./internal/canvaswrite ./internal/cli`
- Search fixture tests prove deterministic JSON output.
- Snapshot-file search works without unnecessary live AFFiNE access.
- MCP metadata remains read-only for search.

## Exit Criteria

`canvas search --json` emits selectors that Phase 2 and Phase 3 can consume without reinterpretation.
