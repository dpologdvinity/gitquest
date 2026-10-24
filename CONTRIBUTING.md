# Contributing to gitquest

Thank you for contributing to `gitquest`! This document details the local development setup, coding standards, and pull request workflow.

For architectural decisions and internal layout, refer to [ARCHITECTURE.md](ARCHITECTURE.md). For detailed product specifications, refer to [docs/SPEC.md](docs/SPEC.md).

---

## 1. Prerequisites

- **Go**: Version `1.26` or higher (`go 1.26` in `go.mod`).
- **Git**: Version `2.32` or higher (required for `GIT_CONFIG_GLOBAL` support).
- **golangci-lint**: Version `v2.x` (recommended for local lint checks).
- **GoReleaser**: Version `v2.x` (optional, for release validation).

---

## 2. Local Setup & Build

Clone the repository and verify the build:

```bash
git clone https://github.com/dpologdvinity/gitquest.git
cd gitquest

# Build the CLI executable
go build -o gitquest .

# Verify the CLI responds
./gitquest --version
./gitquest --help
```

---

## 3. Development Commands

| Task | Command |
|---|---|
| Run all tests | `go test -v ./...` |
| Run tests with race detection | `go test -race ./...` |
| Generate test coverage | `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out` |
| Format codebase | `test -z "$(gofmt -l .)"` (or fix with `gofmt -w .`) |
| Static analysis | `go vet ./...` |
| Linter suite | `golangci-lint run ./...` |
| Validate release config | `goreleaser check` |
| Test snapshot build | `goreleaser build --snapshot --clean` |

---

## 4. Engineering Standards

1. **Standard Library Only:** Zero third-party Go module dependencies. All operations rely strictly on the Go standard library and the host `git` executable.
2. **Deterministic Behavior:** Never introduce wall-clock time or host user configuration into level setups or checks. Always use `internal/gitx` isolated environments.
3. **Plumbing Only:** In checks, use `Ctx` plumbing methods (`rev-parse`, `cat-file`, `merge-base`, etc.). Never parse porcelain output like `git status` or `git log`.
4. **Defensive Safety:** Respect game root boundaries. Never write or delete files outside `$GITQUEST_ROOT`. Check markers before any write or remove operations.
5. **Clean Module Separation:** Maintain loose coupling between `cli`, `game`, `levels`, and `gitx`.

---

## 5. Commit & Pull Request Guidelines

### Conventional Commits
All commits must follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:
- `feat:` New features, commands, or levels
- `fix:` Bug fixes or corrections to game logic
- `test:` Adding or refining unit/integration tests
- `docs:` Documentation improvements
- `refactor:` Code restructuring without functional changes
- `ci:` Updates to GitHub Actions workflows or release configurations
- `chore:` Housekeeping and repository hygiene

### Pull Request Process
1. Create a feature branch from `master`.
2. Ensure all tests, linters, and format checks pass locally:
   ```bash
   test -z "$(gofmt -l .)" && go vet ./... && golangci-lint run ./... && go test ./...
   ```
3. Keep changes atomic and accompanied by relevant tests.
4. Open a pull request targeting `master`.
5. Automated CI checks will run across Linux and macOS runners to validate dependencies, formatting, static analysis, unit tests, coverage, and snapshot builds.
