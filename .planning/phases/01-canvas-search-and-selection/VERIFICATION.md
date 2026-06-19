# Verification: Phase 1 Canvas Search And Selection

## Implementation

- Added `canvas search` as a read-only command.
- Added `internal/canvaswrite.SearchDoc` and `SearchBlocks` with stable selector JSON for blocks, cards, and connectors.
- Supported filters: text, exact ID, flavour, type, display mode, source mode, geometry bounds, connector source, connector target, limit, and text limit.
- Reused the existing AFFiNE Y.js snapshot/live loader path from audit/integrity.
- Fixed local snapshot access so `--snapshot-file` does not require live workspace/doc flags.
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
go run ./cmd/affine-pp-cli canvas search --help
```

Result: passed. The command exposes text, ID, flavour, type, display-mode, source-mode, bounds, connector endpoint, snapshot, history and live flags.

```powershell
go run ./cmd/affine-pp-cli which canvas --json
```

Result: passed. `canvas search` appears in the Canvas capability index.

## Integration Review

- External mutation: none. No AFFiNE live write, deploy, provider mutation or secret access was used.
- MCP safety: `canvas search` is annotated `mcp:read-only`.
- Snapshot safety: local snapshot mode works without requiring workspace/doc parameters.
- Durability: hand-authored generated-CLI changes are represented in `library/affine/.printing-press-patches/affine-canvas-search-selection.json`.

## Open Ends

- Phase 2 should decide whether the selector model should move into a named shared canonical snapshot package before adding diff semantics.
