# API Stability Policy

**Status:** Public normative document  
**Version:** 1.0 (go-inmemory-platform)  
**Purpose:** Define public API boundaries and stability guarantees

---

## Public API vs internal

### Public API

The following import paths are treated as **public API**:

| Area | Packages |
|------|-----------|
| Domain | `pkg/domain` |
| Projection | `pkg/projection` |
| Mongo | `pkg/mongo`, `pkg/mongo/inmemory` |
| PostgreSQL | `pkg/postgres`, `pkg/postgres/migrate`, `pkg/postgres/inmemory` |
| Codegen contract | `pkg/entitymeta` |
| CLI | `cmd/entitygen` (command-line tool; stable flags and generated output shape are versioned with the module) |

**Public API guarantees (intent):**

- Breaking changes are recorded in [CHANGELOG.md](CHANGELOG.md).
- Deprecation is communicated via Go doc comments where practical.
- [Semantic Versioning](https://semver.org/) applies once the project leaves early `v0.0.x`.

### Internal API

The following are **not** covered by stability guarantees:

- `internal/**` (including `internal/entitygen` implementation details)
- Unexported symbols
- Test-only code and `examples/**` (samples may change without notice)

---

## Versioning policy

### Early `v0.0.x`

**Current baseline:** `v0.0.1` — early, consolidated module (Mongo + Postgres + projection + codegen).

**Policy:**

- Breaking changes are **allowed**; they should be **documented** in the changelog.
- Prefer deprecation comments before removal when feasible.
- Patch releases should avoid intentional breaking changes.

### Pre-1.0 (`v0.x` after stabilization)

- Minor bumps (`v0.1.0` → `v0.2.0`) may include breaking changes; describe them in the changelog.
- Patch bumps should remain compatible for typical consumer code.

### Post-1.0 (`v1+`)

- **MAJOR:** breaking changes  
- **MINOR:** backward-compatible features  
- **PATCH:** backward-compatible fixes  
- Deprecations should survive at least one minor release before removal.

---

## Deprecation

Deprecated APIs should be marked in Go doc:

```go
// Deprecated: use NewThing instead; OldThing will be removed in v0.x.
func OldThing() { ... }
```

Removal timing follows the versioning tier above.

---

## What counts as a breaking change

**Breaking:** removing or renaming exported symbols, changing signatures or observable behavior that existing callers rely on, narrowing generated output incompatibly without a major version bump (post-1.0).

**Non-breaking:** new exports, bug fixes, performance work, internal refactors, documentation updates.

---

## Reporting API issues

Unexpected breaks or unclear boundaries: **GitHub Issues** or **Discussions** on `github.com/dhlab-tech/go-inmemory-platform`.

---

## Related documentation

- [CHANGELOG.md](CHANGELOG.md)
- [ARCHITECTURE_CONTRACT.md](ARCHITECTURE_CONTRACT.md)
- [NON_GOALS.md](NON_GOALS.md)
- [CONTRIBUTING.md](CONTRIBUTING.md)
- [docs/positioning.md](docs/positioning.md)
- [docs/anti-patterns.md](docs/anti-patterns.md)
