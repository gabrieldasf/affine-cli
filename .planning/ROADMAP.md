# Roadmap: v1.1 Canvas Power Tools

## Milestone Goal

Turn AFFiNE Canvas support from diagnostic helpers into a safe operator workflow:

1. Search and select exact Canvas entities.
2. Diff expected, historical, snapshot, and live states.
3. Produce machine-applyable dry-run transforms.
4. Apply live changes only through explicit, verified, backed-up gates.

## Phase Sequence

### Phase 1: Canvas Search And Selection

Build the canonical Canvas snapshot model and `canvas search`. This phase is read-only.

Requirements covered: CANVAS-01, CANVAS-02, CANVAS-06 partial.

Exit gate: JSON selectors from search can be reused by tests and later phase fixtures.

Status: complete on 2026-06-18. Evidence: `.planning/phases/01-canvas-search-and-selection/VERIFICATION.md`.

### Phase 2: Canvas Diff And History Diagnosis

Build `canvas diff` on top of the canonical model. Support live-vs-snapshot, snapshot-vs-snapshot, and timestamp/history comparisons.

Requirements covered: CANVAS-01, CANVAS-03, CANVAS-07 partial.

Exit gate: diff categories are stable and fixture-backed.

Status: complete on 2026-06-19. Evidence: `.planning/phases/02-canvas-diff-and-history-diagnosis/VERIFICATION.md`.

### Phase 3: Canvas Transform Dry-Run Plans

Build `canvas transform` as a no-mutation planner that consumes selectors and emits machine-applyable operation plans.

Requirements covered: CANVAS-04, CANVAS-06 partial, CANVAS-07 partial.

Exit gate: transform output is accepted by `canvas apply --dry-run` and rejected when integrity or validation fails.

Status: complete on 2026-06-19. Evidence: `.planning/phases/03-canvas-transform-dry-run-plans/VERIFICATION.md`.

### Phase 4: Gated Live Apply And Publish Proof

Enable live apply only after Phase 3 plans are stable. Add backups, delta capture, reload verification, docs, patch records, MCP annotation review, and official publish proof.

Requirements covered: CANVAS-05, CANVAS-06, CANVAS-07.

Exit gate: live fixture smoke passes after explicit approval, publish validation passes, and workflow proof is non-skipped.

## Parallelism Rule

No more than two subagents may run at the same time.

Recommended per-phase waves:

- Wave A: implementation researcher plus verifier, or two disjoint implementation workers.
- Wave B: plan checker plus integration reviewer.
- Main agent integrates outputs, updates state, and owns final acceptance.

## Milestone Risks

- Y.js live writes lack a server-side compare-and-swap guard; use state vectors, backup artifacts, reload verification, and narrow operation contracts.
- Current `canvas doc audit` and `canvas doc integrity` snapshot examples imply local-only usage but still require workspace/doc flags; fix during Phase 1.
- Patch ledger drift exists; current patch records mention some helper paths that are absent. Audit patch records before relying on them.
- MCP read-only annotations must be reviewed before any mutating command is exposed.
