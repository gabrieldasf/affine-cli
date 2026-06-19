# Plan Index: v1.1 Canvas Power Tools

## Scope

This milestone adds Canvas search, diff, transform planning, and gated live apply to `library/affine`.

## Requirement Coverage

| Requirement | Phase |
| --- | --- |
| CANVAS-01 Canonical Canvas Snapshot Model | Phase 1, Phase 2 |
| CANVAS-02 Block-Level Search And Selection | Phase 1 |
| CANVAS-03 Semantic Canvas Diff | Phase 2 |
| CANVAS-04 Machine-Applyable Transform Plans | Phase 3 |
| CANVAS-05 Gated Live Apply | Phase 4 |
| CANVAS-06 Printing Press Durability | All phases, final in Phase 4 |
| CANVAS-07 Workflow Verification | Phase 2, Phase 3, Phase 4 |

## Subagent Operating Plan

Use at most two active subagents at any time.

For each phase:

1. Main agent opens the phase.
2. Up to two subagents perform read-only research, implementation on disjoint files, or verification.
3. Main agent integrates, runs checks, updates `.planning/STATE.md`, and decides whether to advance.

## Phase Plans

- [Phase 1: Canvas Search And Selection](../../phases/01-canvas-search-and-selection/PLAN.md)
- [Phase 2: Canvas Diff And History Diagnosis](../../phases/02-canvas-diff-and-history-diagnosis/PLAN.md)
- [Phase 3: Canvas Transform Dry-Run Plans](../../phases/03-canvas-transform-dry-run-plans/PLAN.md)
- [Phase 4: Gated Live Apply And Publish Proof](../../phases/04-gated-live-apply-and-publish-proof/PLAN.md)
