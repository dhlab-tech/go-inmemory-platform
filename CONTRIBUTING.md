# Contributing to go-inmemory-platform

Thank you for your interest in contributing to **`go-inmemory-platform`** (`github.com/dhlab-tech/go-inmemory-platform`).

## Code style

- Run `go fmt` / `go vet` on changed packages.
- Prefer clear names and minimal, focused changes; match surrounding patterns.
- If the repo adds `golangci-lint` or similar, follow the checked-in config.

## Pull requests

1. Fork or branch from the default branch.
2. Implement the change with tests where it makes sense (`*_test.go`).
3. Update docs (README, `docs/`, or normative `*.md` in the root) if behavior or guarantees change.
4. Run `go test ./...` (and `-race` when touching concurrency-sensitive code).
5. Open a PR with a short **what / why / how / testing** summary.

## Tests

- New behavior should have unit tests; integration-style checks live under `internal/testwidget` (`-tags=integration`, `POSTGRES_TEST_URL`).
- Examples under `examples/` are smoke checks, not a substitute for tests.

## Commits

Use readable summaries; optional body with context (what & why).

## Scope

See **[NON_GOALS.md](NON_GOALS.md)** and **[ARCHITECTURE_CONTRACT.md](ARCHITECTURE_CONTRACT.md)**:

- **Welcome:** bug fixes, docs, tests, small API improvements aligned with the contract, codegen fixes.
- **Usually out of scope without prior discussion:** distributed cache features, cross-instance sync, built-in eviction, or large API expansions that blur the “projection of one database” model.

## Questions

- **GitHub Issues** — bugs and focused feature proposals.
- **GitHub Discussions** — usage questions.
- **[SUPPORT.md](SUPPORT.md)** — commercial support options.
