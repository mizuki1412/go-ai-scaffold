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
2. `new_project_name` — short name used in config (key `project.name`) and logs, e.g. `myapi`

## Steps

1. **Confirm** the old placeholder values:
   - Old module path: `github.com/example/go-ai-scaffold`
   - Old project name: `go-ai-scaffold`
2. **Full-replace** the old module path with the new one across the WHOLE project (use `Grep` to list files first, then `Edit` with `replace_all` per file). Touch at least:
   - `go.mod` (`module ...` line)
   - All `*.go` import paths under `pkg/`, `mod/`, `cmd/`, `main.go`
   - The placeholder string `github.com/example/go-ai-scaffold` appears in every import; replace_all per file is the safest approach.
3. **Rename project name** (the short name `go-ai-scaffold`):
   - The scaffold has NO `configs/config.yaml` file by default — config is loaded by `configkit` from file/env at runtime. The project name is referenced via config key `project.name` (see `pkg/cli/configkey/project.go`).
   - If the user has added a config file (e.g. `config.yaml` / `config.json`), update the `project.name` field in it.
   - If no config file exists, the rename only needs the module path change (step 2) — the project name is a runtime config value, not a compile-time constant.
   - Update `.gitignore` if it references the old binary name (e.g. `go-ai-scaffold.exe` — per project_memory, the build artifact must be gitignored).
4. **Do NOT** rename the `.trae/skills/` directory or skill names — they are scaffold tooling, not project code.
5. **Do NOT** rename `pkg/` package directories — they are generic reusable libraries; only the import path prefix (module path) changes.
6. **Tidy & verify**:
   - Run `go mod tidy`
   - Run `go build ./...`
   - Run `go vet ./...`
   - Run `gofmt -l .` (must return no files)
7. Report what changed and confirm the build is green.

## What changes vs what stays

| Changes | Stays the same |
|---------|----------------|
| `go.mod` module line | `.trae/skills/` directory + skill names |
| All `*.go` import paths `github.com/example/go-ai-scaffold/...` | `pkg/` / `mod/` / `cmd/` directory structure |
| `project.name` in config file (if exists) | Package names within `pkg/` (class, library, cli, service) |
| `.gitignore` binary name (if listed) | File names within `pkg/class/`, `pkg/library/` |

## Rules

- Always replace the FULL module path string, never partial substrings.
- Keep import ordering valid; run `gofmt -w .` after edits.
- Never silently delete files; if a rename requires directory moves, ask the user first.
- If `go build` fails after rename, fix import paths until green — do not leave the project broken.
- The `.ai/conventions.md` file references the old placeholder module path in its examples — update those too so future AI reads see the new path.
