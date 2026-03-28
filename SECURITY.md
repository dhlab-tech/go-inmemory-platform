# Security policy

## Supported versions

| Version | Supported            |
| ------- | -------------------- |
| 0.0.x   | :white_check_mark:   |

Security fixes are applied to the latest patch of the supported minor line when practical.

## Reporting a vulnerability

**Do not** open a public issue for undisclosed security problems.

### Process

1. Email **security@digital-heroes.tech** with:
   - Description and impact
   - Steps to reproduce
   - Affected component (Mongo adapter, Postgres/replication, projection, codegen, etc.)
   - Suggested fix (optional)

### Timeline (intent)

- **First response:** within a few business days
- **Updates:** as the investigation progresses
- **Resolution:** depends on severity

### Disclosure

- Significant issues may receive an advisory after a fix is available.
- Reporter credit if you want it.

## Hardening in production

When running **`go-inmemory-platform`**:

- **MongoDB:** authentication, TLS, least-privilege users, current server versions; secured Change Stream access.
- **PostgreSQL:** authentication, TLS, replication roles with minimal rights; secure publication/slot management if using logical replication.
- **Application:** keep the module and drivers updated; monitor DB and app logs; avoid exposing admin interfaces.

## Contact

**security@digital-heroes.tech**
