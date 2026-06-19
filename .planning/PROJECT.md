# AFFiNE CLI

## Project Purpose

AFFiNE CLI is a Printing Press published command-line and MCP interface for operating AFFiNE workspaces. The current Canvas surface can plan, model, validate, inspect, audit, repair narrow integrity issues, and patch individual card images. The next product step is to make Canvas work powerful enough for repeatable operator workflows: find the right blocks, compare actual state against intended state, plan safe batch transformations, and only then apply validated live changes.

## Active Milestone

**v1.1 Canvas Power Tools**

Build block-level Canvas search, semantic diff, dry-run transform planning, and gated live apply. This milestone preserves the current safety stance: read-only and dry-run by default, explicit mutation only after backup, integrity checks, and reload verification.

## Operating Constraints

- The generated CLI lives under `library/affine`; durable hand-authored changes require Printing Press patch records under `library/affine/.printing-press-patches/`.
- Canvas live writes use AFFiNE Socket.IO document load/update flow and Y.js deltas. Treat every live mutation as high-risk persistent data work.
- `canvas apply` remains dry-run by default. Live apply requires explicit flags and proof artifacts.
- Planning and execution use at most two subagents per active wave.
- Phase execution is serial across the milestone: Phase 1 unlocks Phase 2, Phase 2 unlocks Phase 3, and Phase 3 unlocks Phase 4.

## Success Criteria

- Operators can locate Canvas blocks/cards/connectors by text, ID, flavour, type, geometry, timestamp, and snapshot source.
- Operators can diff live, snapshot, and history states using stable semantic categories.
- Operators can produce machine-applyable dry-run transform plans with affected IDs, before/after values, and validation results.
- Live apply uses backup-before, delta capture, pre/post integrity checks, reload verification, and publish proof.
- CLI, MCP metadata, README/SKILL/which docs, patch records, tests, and workflow proof stay aligned.
