# AFFiNE CLI

## What This Is

AFFiNE CLI is a Printing Press published command-line and MCP interface for operating AFFiNE workspaces. It now includes Canvas Power Tools for search, semantic diff, dry-run transform planning, and gated live apply.

## Current State

**Shipped version:** v1.1 Canvas Power Tools on 2026-06-19.

The Canvas surface can:

- Locate Canvas blocks, cards, connectors, text, geometry, display mode, source mode, and history/snapshot states.
- Diff Canvas states with stable semantic categories.
- Produce machine-applyable dry-run transform plans.
- Apply transform and layout plans live only behind explicit workspace/doc/backup/confirmation gates.
- Capture backup and delta artifacts, perform integrity checks, reload the document, and preserve publish proof.

## Core Value

Safe operator workflows for AFFiNE Canvas: inspect first, plan second, mutate only with explicit proof.

## Validated Requirements

- CANVAS-01 Canonical Canvas Snapshot Model — v1.1
- CANVAS-02 Block-Level Search And Selection — v1.1
- CANVAS-03 Semantic Canvas Diff — v1.1
- CANVAS-04 Machine-Applyable Transform Plans — v1.1
- CANVAS-05 Gated Live Apply — v1.1
- CANVAS-06 Printing Press Durability — v1.1
- CANVAS-07 Workflow Verification — v1.1

## Known Tech Debt

- Add durable workflow smoke steps for search, diff, and doc integrity.
- Add exact post-reload field assertions for live transform operations.
- Use formal traceability tables and phase summaries in the next milestone.
- Run Nyquist validation when strict validation artifacts are required.

## Operating Constraints

- The generated CLI lives under `library/affine`; durable hand-authored changes require Printing Press patch records under `library/affine/.printing-press-patches/`.
- Canvas live writes use AFFiNE Socket.IO document load/update flow and Y.js deltas. Treat every live mutation as persistent data work.
- `canvas apply` remains dry-run by default. Live apply requires explicit flags and proof artifacts.

## Next Milestone Goals

Define the next milestone with `$gsd-new-milestone`. Good candidates are:

- harden live apply verification,
- broaden durable workflow coverage,
- clean up generated CLI patch durability,
- expand Canvas editing operations only where operator workflows need them.

<details>
<summary>Archived v1.1 project brief</summary>

The v1.1 brief focused on turning AFFiNE Canvas support from diagnostic helpers into safe operator workflows: find the right blocks, compare actual state against intended state, plan safe batch transformations, and only then apply validated live changes.

</details>

---

*Last updated: 2026-06-19 after v1.1 milestone*
