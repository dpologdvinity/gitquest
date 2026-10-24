# Architecture of gitquest

`gitquest` is an interactive escape-room puzzle game designed to teach Git through hands-on troubleshooting in real Git repositories.

---

## 1. Technical Stack

| Layer | Technology | Details |
|---|---|---|
| **Language** | Go (`go 1.26+`) | Standard library only; zero third-party dependencies |
| **VCS Engine** | Git (>= 2.32) | Invoked via child process with sanitized environments |
| **Packaging & Release** | GoReleaser v2 | Cross-compiled static binaries for Linux and macOS (`amd64`/`arm64`) with `CGO_ENABLED=0` |
| **CI/CD** | GitHub Actions | Multi-stage pipeline: Dependency validation, linting, matrix testing with coverage, and artifact builds |

---

## 2. Component Boundaries & Package Layout

```
gitquest/
├── main.go                       # Entrypoint: constructs RealDeps and exits with cli.Run
├── internal/
│   ├── cli/                      # CLI argument parsing, command dispatch, and output formatting
│   │   ├── cli.go                # Run, Deps, readyText, status, level, check, hint, reset
│   │   ├── cli_test.go           # Basic CLI unit tests
│   │   └── play_test.go          # Integrated CLI playthrough tests
│   ├── game/                     # Game lifecycle, root boundaries, progress, and building
│   │   ├── root.go               # EnsureRoot, SafeRemove, Root, LevelDir, markers (.gitquest-root)
│   │   ├── progress.go           # Progress persistence (progress.json), quarantine, CurrentLevel
│   │   ├── build.go              # Build(root, Level), Check(root, Level)
│   │   ├── build_test.go         # Build isolation and determinism tests
│   │   └── play_test.go          # Matrix test runner across all 12 levels
│   ├── levels/                   # Escape room level specifications and check logic
│   │   ├── level.go              # Level struct, Step table, Ctx (Git plumbing helpers), Registry
│   │   ├── l01_first_commit.go   # Level 1: The Empty Logbook
│   │   ├── ...                   # Levels 2-11
│   │   └── l12_gitignore.go      # Level 12: Secrets in the Attic
│   └── gitx/                     # Git execution engine and environment sandboxing
│       ├── gitx.go               # Git runner, SetupEnv, CheckEnv, ParseVersion, VersionOK
│       └── gitx_test.go          # Environment sanitization and git runner tests
├── docs/
│   ├── SPEC.md                   # Source of truth specification (v0.1)
│   ├── PLAN.md                   # Original step-by-step implementation plan
│   └── agents/                   # Agent and owner ledgers (git-ignored)
├── .github/workflows/            # GitHub Actions CI & Release automation
├── .goreleaser.yaml              # Release configuration
├── .golangci.yml                 # Linter configuration (v2 format)
└── README.md                     # Project overview and player guide
```

---

## 3. Data Flow & Execution Lifecycles

### 3.1 Command Routing (`internal/cli`)
1. User executes `gitquest <command> [args]`.
2. `cli.Run` checks for flags (`-V`, `--version`, `-h`, `--help`).
3. CLI dependencies (`cli.Deps`) provide abstract access to standard streams, environment variables, system clock, and current working directory for deterministic testing.
4. Commands (`start`, `level`, `check`, `hint`, `status`, `reset`) validate system `git` compatibility (>= 2.32) via `gitx.VersionOK()` before modifying state.

### 3.2 Root Sandboxing & Safe Deletion (`internal/game`)
To protect against accidental directory modification or deletion:
- Game repositories are constrained strictly to `$GITQUEST_ROOT` (defaults to `~/gitquest`).
- `EnsureRoot` writes a `.gitquest-root` sentinel (`gitquest game root v1\n`). If the path already exists as an unmarked non-empty directory, Gitquest refuses with a `RootError`.
- `SafeRemove` validates:
  1. The game root has a valid `.gitquest-root` marker.
  2. The target level folder is a real directory (not a symlink).
  3. Symlinks resolve strictly inside the game root.
  4. The target directory contains `.gitquest-level` with matching level metadata.
  If any check fails, `SafeRemove` refuses deletion.

### 3.3 Deterministic Level Generation (`internal/levels` & `internal/gitx`)
Levels are constructed from deterministic sequences of steps (`Write`, `Commit`, `Branch`, `Switch`, `ResetHard`, `Save`):
- Git commits use a fixed baseline timestamp (`1767258000`, 2026-01-01 09:00:00 UTC), advancing by 60 seconds per commit.
- `SetupEnv` clears all host `GIT_*` environment variables, sets `GIT_CONFIG_NOSYSTEM=1`, `GIT_CONFIG_GLOBAL=os.DevNull`, and assigns narrator identity.
- Local repository configuration forces predictable Git behavior: disables GPG signing, sets `core.autocrlf=false`, `merge.ff=true`, and sets `core.hooksPath` to an empty non-existent directory.
- As a result, building any level across different machines yields identical commit hashes and trees.

### 3.4 Plumbing-Only Verification (`Ctx`)
Verification functions (`Level.Check`) inspect player work through `Ctx` plumbing methods:
- `symbolic-ref` for current branch (`Head`).
- `rev-parse` for commit and tree hashes (`Rev`, `Tree`, `Has`).
- `cat-file` for raw blobs and commit messages (`Blob`, `Message`).
- `merge-base` for ancestry relationships (`IsAncestor`).
- `check-ignore` for ignore verification (`Ignored`).
Porcelain commands (`git status`, `git log`) are never parsed, ensuring checks are immune to formatting differences, aliases, and terminal width.

### 3.5 Progress State Ledger
Player progress is persisted in `progress.json`:
- Location: `$GITQUEST_HOME/progress.json` (or `$XDG_STATE_HOME/gitquest`, falling back to `~/.local/state/gitquest`).
- Written atomically with file mode `0600` inside directory `0700`.
- Corrupt files are preserved as `progress.json.corrupt-<timestamp>` rather than deleted.
