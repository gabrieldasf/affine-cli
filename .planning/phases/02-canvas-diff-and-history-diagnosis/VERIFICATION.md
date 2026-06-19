# Verification: Phase 2 Canvas Diff And History Diagnosis

## Implementation

- Added `canvas diff` as a read-only command.
- Added `internal/canvaswrite.DiffDoc` and `DiffSearchResults` on top of the Phase 1 search selector model.
- Supported input pairs:
  - `--before-snapshot` plus `--after-snapshot`;
  - `--before-snapshot` plus live after-state through `--workspace` and `--doc`;
  - `--before-timestamp` plus live after-state;
  - `--before-timestamp` plus `--after-timestamp`.
- Added stable diff JSON with source metadata, issue count, category summary, severity, affected ID, before/after values, and suggested next action.
- Covered semantic categories: added, removed, text changed, geometry changed, display mode changed, relinked, image/source changed, unreachable and integrity changed.
- Updated README, SKILL, root help highlights, `which`, and Printing Press patch records.

## Verification

```powershell
go test ./internal/canvaswrite ./internal/cli
```

Result: passed.

```powershell
go test ./...
```

Result: passed.

```powershell
go run ./cmd/affine-pp-cli canvas diff --help
```

Result: passed. The command exposes before/after snapshot and timestamp inputs plus workspace/doc live comparison flags.

```powershell
go run ./cmd/affine-pp-cli which "canvas diff" --json
```

Result: passed. `canvas diff` appears as the top Canvas capability match.

## Integration Review

- External mutation: none. No AFFiNE live write, deploy, provider mutation or secret access was used.
- MCP safety: `canvas diff` is annotated `mcp:read-only`.
- Selector compatibility: diff consumes the Phase 1 `SearchEntity` output shape instead of re-reading raw Y.js structures directly.
- Durability: hand-authored generated-CLI changes are represented in `library/affine/.printing-press-patches/affine-canvas-diff-diagnosis.json`.

## Open Ends

- Phase 3 should convert diff/search selections into dry-run transform plans without enabling live apply.
