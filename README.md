# happy-wakey-lib-core

The shared Happy Wakey domain and persistence library. It imports the immutable public contracts from `happy-wakey-interfaces` and supplies equivalent SeaORM, Drizzle, Prisma, GORM, and gRPC integration surfaces.

| Surface | Location | Boundary |
| --- | --- | --- |
| SeaORM | `src/lib.rs` | Rust read/write capabilities; read-only by default |
| Drizzle | `drizzle/` | Typed PostgreSQL schema and subject predicate builder |
| Prisma | `prisma/schema.prisma` | PostgreSQL/CockroachDB declarative ORM schema |
| GORM | `gorm/` | Compiled Go model and subject-scoped repository |
| gRPC | `proto/happy_wakey/v1/core.proto` | Credential-free request messages; Shared Auth is interceptor metadata |

The declarative SQL in `happy-wakey-interfaces` remains authoritative. These
ORM views must change with it and the polyglot validation gate checks the
security-critical table, owner, and RPC invariants.

## Boundaries

- `happy-wakey-interfaces` remains the authority for public and private wire types. This crate re-exports those types; it does not copy them.
- `ReadContext` is the default surface for server-side reads. It owns its SeaORM connection and exposes named operations rather than a raw connection.
- `WriteContext` is available only behind the `read-write` feature and remains subject to product authorization in the API server.
- Shared Auth establishes identity, realm, audience, client, session, scope, and assurance. Happy Wakey membership and resource authorization remain product-local.
- PostgreSQL and CockroachDB use SeaORM's PostgreSQL driver. Database principals and network policy remain the final enforcement boundary.
- This library never runs schema migrations or DDL at application startup.

The web/API topology can use this crate for the direct read-only database avenue. Stateless HTTP, stateful TCP, and asynchronous NATS are transport boundaries owned by the web and API servers; all four avenues must return the same `happy-wakey-interfaces` contracts and enforce the same Shared Auth identity.

## Dependency management and validation

Use the released `zed-pkg` CLI as the dependency entry point:

```sh
zed validate
zed install --adapter rust
zed run cargo test --all-targets --all-features --locked
zed run cargo fmt --all -- --check
zed run cargo clippy --all-targets --all-features --locked -- -D warnings
python3 scripts/validate_polyglot_contracts.py
(cd drizzle && npm ci && npm test)
(cd gorm && go test ./...)
protoc --proto_path=proto --descriptor_set_out=/tmp/happy-wakey-core.pb \
  proto/happy_wakey/v1/core.proto
```

The Cargo dependency on `happy-wakey-interfaces` is pinned to reviewed commit
`d6278ec8f6b2263678728b147a32dff92d52d8c8`, which includes the shared
Bluetooth lifecycle lane. Do not replace it with a branch or an unpinned Git
head.
