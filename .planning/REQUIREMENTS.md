# Requirements: v1.1 Canvas Power Tools

## Must Haves

### CANVAS-01: Canonical Canvas Snapshot Model

The CLI must expose an internal canonical model for Canvas blocks, notes/cards, connectors, text content, geometry, display mode, image/source metadata, hierarchy, and integrity signals. The model must work from live doc load, local snapshot files, and AFFiNE history timestamps.

### CANVAS-02: Block-Level Search And Selection

`canvas search` must find Canvas entities by text, exact ID, flavour, type, display mode, geometry bounds, connector endpoints, snapshot source, and history timestamp. JSON output must be stable and reusable as selector input for later commands.

### CANVAS-03: Semantic Canvas Diff

`canvas diff` must compare two Canvas states and report added, removed, text-changed, geometry-changed, display-mode-changed, relinked, image/source-changed, orphaned, unreachable, and integrity-changed entities.

### CANVAS-04: Machine-Applyable Transform Plans

`canvas transform` must produce dry-run operation plans for move, resize, align, distribute, set display mode, and selected metadata updates. Output must include affected IDs, before/after values, preflight integrity, validation results, and rollback/proof fields.

### CANVAS-05: Gated Live Apply

`canvas apply --live` must remain unavailable until the operation plan contract is stable. When enabled, live apply must require explicit apply flags, workspace/doc targeting, backup-before, delta artifact, pre/post integrity, reload verification, and clear MCP mutation annotations.

### CANVAS-06: Printing Press Durability

All hand-authored Canvas changes must be represented in patch records so they survive Printing Press regeneration. README, SKILL, `which`, command help, MCP metadata, and publish proof must describe the same command surface.

### CANVAS-07: Workflow Verification

The milestone must add a real Canvas workflow proof instead of relying only on publish validation. Verification must cover no-mutation dry-run behavior, Y.js integrity, fixture smoke, publish validation, and live dogfood only after explicit approval.

## Non-Goals

- Rewriting AFFiNE's renderer or Blocksuite internals.
- Broad live mutation support before search, diff, and transform contracts are stable.
- Hardcoding Quartzo workspace, document, or block IDs into reusable code or docs.
- Treating generated GraphQL search endpoints as a substitute for Canvas block-level search.

## Definition Of Done

- All four phases are complete and verified.
- `go test ./...` passes from `library/affine`.
- Canvas-specific tests cover search, diff, transform dry-run, apply safety gates, and integrity failure paths.
- `cli-printing-press publish validate --cli library/affine` or the current equivalent passes.
- A Canvas workflow proof exists and is referenced from patch records.
- The public Printing Press publish path remains compatible with the official protocol.
