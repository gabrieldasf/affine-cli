# Resolved Blocker: Phase 4 Live Fixture Approval

## Status

Resolved on 2026-06-19. The user explicitly approved a robust Live Zap live smoke against the AFFiNE workspace and document.

## Completed Locally

- Added a live `canvas apply --live` / `--apply` path for `canvas_transform` plans.
- Required `--workspace`, `--doc`, `--backup-dir`, and `--yes` before live apply.
- Preserved dry-run behavior for layout and transform plans.
- Added backup-before and delta artifact writes through the existing restrictive backup helper.
- Added pre-push local integrity and post-push reload integrity in the live code path.
- Added semantic diff preview from transform operations before live push.
- Kept `canvas apply` outside `mcp:read-only` annotations.
- Added `workflow_verify.yaml` with non-mutating Canvas transform/apply and pre-network live-gate workflows.
- Updated `workflow-verify-report.json` so Canvas workflow proof is no longer skipped.
- Extended `canvas apply --live` to support Canvas layout plans, not only transform plans.

## Verification Run

```powershell
go test ./internal/canvaswrite ./internal/cli
go test ./...
go run ./cmd/affine-pp-cli canvas apply --help
go run ./cmd/affine-pp-cli which "canvas apply" --json
```

Result: passed.

Dry-run transform-plan smoke passed and reported the required live gates. Local workflow proof passed without external mutation.

Live gate smokes failed before network as expected:

- missing `--yes`
- missing `--backup-dir`

## Live Smoke

- Workspace: `727cc066-a25e-4560-b68d-414b67cbc5c8`
- Document: `B6pvUw-r5SSfWKam-wncU`
- Anchor: existing Livess App content found under note `CDwHI00A3O`.
- Layout apply: created 20 Live Zap architecture and infrastructure cards plus 28 connectors below the Livess App area.
- Transform apply: set `phase4_proof=live-zap-layout-smoke` on the Live Zap hub card.
- Backups:
  - `C:\Users\bielx\AppData\Local\Temp\affine-live-zap-20260619-010552`
  - `C:\Users\bielx\AppData\Local\Temp\affine-live-zap-transform-20260619-010651`
- Verification: post-apply integrity stayed OK; search in the target bounds found 20 new note cards; diff from the pre-smoke snapshot reported 68 additions.

## Resolution

The approval blocker is closed. Phase 4 can use the live smoke evidence in `VERIFICATION.md` for completion and ship review.
