---
name: "scaffold-rename"
description: "Rename a copied Go scaffold project: replace module path, project name and all references. Invoke right after copying the scaffold to a new project, when user says 'rename project to X' / '重命名工程' / '改 module 路径'."
---

# Scaffold Rename

When a copy of `go-ai-scaffold` is used as the starting point for a new project, this skill renames the placeholder module path and project name everywhere so the new project is self-consistent and `go build` still passes.

## When to invoke

- Right after the scaffold is copied to a new directory.
- User says: "重命名工程为 X" / "rename project to X" / "改 module 为 github.com/foo/bar".
- User wants to change the placeholder `github.com/example/go-ai-scaffold`.

## Inputs

Ask the user (only if not provided) for:
1. `new_module_path` — e.g. `github.com/myorg/myapi`
2. `new_project_name` — short name used in Makefile / config / README, e.g. `myapi`

## Steps

1. **Confirm** the old placeholder values:
   - Old module path: `github.com/example/go-ai-scaffold`
   - Old project name: `go-ai-scaffold`
2. **Full-replace** the old module path with the new one across the WHOLE project (use `Grep` to list files first, then `Edit` with `replace_all` per file). Touch at least:
   - `go.mod` (`module ...` line)
   - All `*.go` import paths
   - `configs/config.yaml` (service name / app name fields)
3. **Rename project name** (the short name `go-ai-scaffold`) in:
   - `configs/config.yaml`
4. **Do NOT** rename the `.trae/skills/` directory or skill names — they are scaffold tooling, not project code.
5. **Tidy & verify**:
   - Run `go mod tidy`
   - Run `go build ./...`
   - Run `go vet ./...`
6. Report what changed and confirm the build is green.

## Rules

- Always replace the FULL module path string, never partial substrings.
- Keep import ordering valid; run `gofmt -w .` if available.
- Never silently delete files; if a rename requires directory moves, ask the user first.
- If `go build` fails after rename, fix import paths until green — do not leave the project broken.
