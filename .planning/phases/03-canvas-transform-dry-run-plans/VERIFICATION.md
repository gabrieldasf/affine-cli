# Verification: Phase 3 Canvas Transform Dry-Run Plans

## Implementation

- Added `canvas transform` as a dry-run planner for selected AFFiNE canvas entities.
- Added `internal/canvaswrite.TransformPlan` and `BuildTransformPlan` on top of the Phase 1 selector model.
- Supported direct IDs and selector JSON from `canvas search`.
- Supported operations: move, resize, align, distribute, set display mode, and selected metadata updates.
- Added deterministic plan IDs, affected IDs, before/after values, integrity status, backup target, rollback/proof fields, warnings, and dry-run metadata.
- Added transform-plan compatibility to `canvas apply --dry-run` while preserving existing layout-plan behavior.
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
go run ./cmd/affine-pp-cli canvas transform --help
```

Result: passed. The command exposes selector input, direct IDs, move, resize, align, distribute, display-mode, metadata, and backup-target flags.

```powershell
go run ./cmd/affine-pp-cli which "canvas transform" --json
```

Result: passed. `canvas transform` appears as the top Canvas capability match.

```powershell
$selector = '{"doc_id":"doc-1","source_mode":"snapshot","count":1,"entities":[{"id":"card","kind":"card","display_mode":"edgeless","xywh":[10,20,100,80]}]}'
$plan = $selector | go run ./cmd/affine-pp-cli canvas transform --selectors - --move 5,0 --json
$plan | go run ./cmd/affine-pp-cli canvas apply --dry-run --json
```

Result: passed. `canvas apply --dry-run` accepted the transform plan and returned `live_write_supported: false`.

## Integration Review

- External mutation: none. No AFFiNE live write, deploy, provider mutation or secret access was used.
- Live safety: `canvas transform` only consumes selectors or direct IDs and does not open Socket.IO or push APIs.
- Apply safety: transform plans are accepted only by `canvas apply --dry-run`; live transform apply remains a Phase 4 gate.
- Validation: geometry operations require XYWH values, invalid resize values fail, unsupported align/distribute values fail, and selectors with integrity issues fail before plan output.
- Durability: hand-authored generated-CLI changes are represented in `library/affine/.printing-press-patches/affine-canvas-transform-plans.json`.

## Open Ends

- Phase 4 should add explicit backups, live apply gates, reload verification, and publish proof before enabling mutation.
