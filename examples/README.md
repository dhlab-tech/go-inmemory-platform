# Examples

Run from the **repository root** (`go-inmemory-platform`):

| Directory | Command | Needs |
|-----------|---------|--------|
| [mongo_crud](mongo_crud/) | `go run ./examples/mongo_crud` | MongoDB replica set (`MONGODB_URI`), Change Streams |
| [mongo_indexes](mongo_indexes/) | `go run ./examples/mongo_indexes` | Same |
| [mongo_listeners](mongo_listeners/) | `go run ./examples/mongo_listeners` | Same |
| [postgres_widgets](postgres_widgets/) | `go run ./examples/postgres_widgets` | Postgres (`POSTGRES_DSN`) |

Mongo examples expect a replica set (e.g. `replicaSet=rs0`) because they use change streams.
