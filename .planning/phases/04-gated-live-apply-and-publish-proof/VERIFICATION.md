# Phase 4 Verification

Date: 2026-06-19

## Scope

Phase 4 adds explicit, backed-up live Canvas apply for validated plans. The completed scope covers:

- `canvas apply --live` / `--apply` for `canvas_transform` plans.
- `canvas apply --live` / `--apply` for Canvas layout plans.
- Required `--workspace`, `--doc`, `--backup-dir`, and `--yes` gates before mutation.
- Backup-before and delta artifacts.
- Pre-push local integrity and post-push reload integrity.
- Semantic diff preview for transform plans.
- Workflow proof that is no longer skipped.

## Live Smoke

Target:

- Workspace: `727cc066-a25e-4560-b68d-414b67cbc5c8`
- Document: `B6pvUw-r5SSfWKam-wncU`
- Anchor: existing Livess App content in note `CDwHI00A3O`

Smoke result:

- Created 20 detailed Live Zap architecture and infrastructure cards below the Livess App area.
- Created 28 connectors between the new cards.
- Verified the created area with `canvas search --bounds "-25350,4980,3300,2400" --flavour affine:note --json`; result count was 20.
- Verified pre/post document integrity with `canvas doc integrity --workspace ... --doc ... --json`; result stayed OK.
- Compared the live document against the pre-smoke snapshot; result showed 68 additions, matching 20 note blocks, 20 paragraph child blocks, and 28 connectors.
- Applied a second live transform to the hub card, setting `phase4_proof=live-zap-layout-smoke`.
- Verified the transformed hub block with `canvas block --block live-zap-20260619-hub --json`.

Backup artifacts:

- `C:\Users\bielx\AppData\Local\Temp\affine-live-zap-20260619-010552`
- `C:\Users\bielx\AppData\Local\Temp\affine-live-zap-transform-20260619-010651`

## Commands Run

```powershell
go test ./internal/cli ./internal/canvaswrite
go test ./...
go build ./cmd/affine-pp-cli ./cmd/affine-pp-mcp
go run ./cmd/affine-pp-cli canvas apply --help
go run ./cmd/affine-pp-cli which "canvas apply" --json
go run ./cmd/cli-printing-press workflow-verify --dir library/affine --json
go run ./cmd/cli-printing-press publish validate --dir library/affine --json
go run ./cmd/affine-pp-cli canvas doc integrity --workspace 727cc066-a25e-4560-b68d-414b67cbc5c8 --doc B6pvUw-r5SSfWKam-wncU --json
go run ./cmd/affine-pp-cli canvas search --workspace 727cc066-a25e-4560-b68d-414b67cbc5c8 --doc B6pvUw-r5SSfWKam-wncU --bounds "-25350,4980,3300,2400" --flavour affine:note --json
```

Additional proof:

- `workflow-verify` verdict: `workflow-pass`.
- `publish validate` result: `passed=true`.

## Verdict

Pass. Phase 4 meets the exit criteria: live apply is explicit, backed up, integrity checked, reload verified, documented, and proven against an approved live fixture.
