# Retrospective

## Milestone: v1.1 Canvas Power Tools

**Shipped:** 2026-06-19
**Phases:** 4
**Audit:** `.planning/v1.1-MILESTONE-AUDIT.md`

### What Was Built

- `canvas search` for reusable Canvas selectors.
- `canvas diff` for semantic Canvas state comparison.
- `canvas transform` for dry-run operation plans.
- `canvas apply --live` for explicitly gated transform and layout mutation.
- Workflow and publish validation for the generated AFFiNE CLI.

### What Worked

- Serial phase order matched the dependency chain: search enabled diff, diff informed transform, transform enabled live apply.
- Keeping mutation disabled until Phase 4 prevented early live-write risk.
- The live smoke gave concrete proof that the CLI can write useful Canvas structures, not just pass local tests.

### What Was Inefficient

- Missing phase `SUMMARY.md` files weakened the automated GSD closeout path.
- `workflow_verify.yaml` started too shell-like and had to be corrected into native CLI commands.
- Requirements were readable but not machine-traceable.

### Patterns Established

- Canvas live writes require explicit target, backup directory, confirmation, pre/post integrity, and reload proof.
- Patch records are required for generated CLI durability.
- Live smoke evidence belongs in phase verification before archive.

### Key Lessons

- Durable workflow manifests should exercise the real CLI command surface, not shell pipelines.
- GSD phase summaries are cheap compared with reconstructing milestone accomplishments later.
- Live integrity checks are necessary; semantic post-apply assertions are the next hardening step.

## Cross-Milestone Trends

| Theme | Observation |
|---|---|
| Safety gates | Mutation work needs proof artifacts before archive. |
| Traceability | Next milestones should use requirement tables from the start. |
