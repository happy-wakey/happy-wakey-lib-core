# happy-wakey-lib-core

The shared Happy Wakey domain and persistence library. It imports the immutable public contracts from `happy-wakey-interfaces`, adds implementation code and SeaORM connection policy, and is the only library that API and trusted web processes should use for direct database access.

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
```

The Cargo dependency on `happy-wakey-interfaces` is pinned to the reviewed immutable commit recorded in `Cargo.toml`. Do not replace it with a branch or an unpinned Git head.
