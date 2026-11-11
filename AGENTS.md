# AGENTS.md

Instructions for coding agents (Claude Code, Codex and others) working on `gitquest`.
`CLAUDE.md` only imports this file, so edit here.

## What this is

`gitquest` is a Go CLI escape-room game that teaches git. Each of the 12 levels is a real,
deterministically built git repo inside a marked game root; `gitquest check` grades it with
git plumbing. Layout and data flow: [ARCHITECTURE.md](ARCHITECTURE.md).

## Setup

- Go version from `go.mod`, git 2.32 or newer (`internal/gitx` refuses older).
- Optional tools: `golangci-lint` v2, `shellcheck`, `goreleaser` v2.
- Pre-commit hooks: `lefthook install` (runs gitleaks, gofmt, golangci-lint and typos; see `lefthook.yml`).
- No env vars are needed to develop. Runtime overrides: `GITQUEST_ROOT`, `GITQUEST_HOME`, `GITQUEST_INSTALL_DIR` (see README).

## Commands

Run all of these before calling work done (CI runs the same checks):

```bash
test -z "$(gofmt -l .)"
go vet ./...
golangci-lint run ./...
go test ./...
shellcheck install.sh && go test ./scripts   # when install.sh changes
goreleaser check                             # when .goreleaser.yaml changes
```

## Rules

- Standard library only. Do not add Go module dependencies.
- Checks use the plumbing methods on `levels.Ctx`. Never parse porcelain output (`git status`, `git log`).
- Level setup must stay deterministic: run git through `internal/gitx` isolated envs, never the host clock or git config.
- Only create or delete inside the game root, and only in dirs carrying a valid `.gitquest-level` marker.
- Tests never touch the real `HOME`: use `t.Setenv("HOME", t.TempDir())` and temp roots.
- Every level has a test that applies the solution and passes, and one that the unsolved state fails.

## Git

- Default branch is `master`. Conventional Commits (`feat:`, `fix:`, `test:`, `docs:`, `refactor:`, `ci:`, `chore:`), small and atomic.
- Never commit `.env*`, `*.local.*`, `CLAUDE.local.md` or `STEPS-FOR-OWNER.md`.

## Agent workflow

Features go through the [superpowers](https://github.com/obra/superpowers) loop:
brainstorm, write a spec, turn it into a task-by-task plan, then implement each task with
subagents and review.

- Specs: [`docs/agents/specs/`](docs/agents/specs/). The v0.1 spec is the source of truth for behavior; if a plan disagrees, the spec wins.
- Plans: [`docs/agents/plans/`](docs/agents/plans/).
- Name new files `YYYY-MM-DD-<topic>.md` (specs end in `-design.md`).
- Shared plugins: `.claude/settings.json` (Claude Code) and `.codex/config.toml` (Codex).
  Personal overrides go in the gitignored `.claude/settings.local.json`.
