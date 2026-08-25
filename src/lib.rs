#![forbid(unsafe_code)]
//! Happy Wakey's role-aware domain and SeaORM boundary.
//!
//! Public contracts are re-exported from the canonical interfaces crate. Raw
//! database connections remain private so callers use named, subject-scoped
//! operations.

pub use happy_wakey_interfaces as interfaces;

use sea_orm::{
    ConnectOptions, ConnectionTrait, Database, DatabaseConnection, DbErr, Statement, TryGetable,
};
use std::time::Duration;

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum DatabaseFlavor {
    PostgreSql,
    CockroachDb,
}

/// A database capability suitable for trusted server-side, read-only work.
pub struct ReadContext {
    connection: DatabaseConnection,
    flavor: DatabaseFlavor,
}

/// A database capability for the API mutation boundary.
///
/// Cargo features are not an authorization mechanism. The API must still
/// authenticate with Shared Auth and enforce product-local ownership.
#[cfg(feature = "read-write")]
pub struct WriteContext {
    connection: DatabaseConnection,
    flavor: DatabaseFlavor,
}

impl ReadContext {
    pub async fn connect(
        database_url: &str,
        flavor: DatabaseFlavor,
        max_connections: u32,
    ) -> Result<Self, DbErr> {
        Ok(Self {
            connection: connect(database_url, max_connections).await?,
            flavor,
        })
    }

    pub const fn flavor(&self) -> DatabaseFlavor {
        self.flavor
    }

    pub async fn healthcheck(&self) -> Result<(), DbErr> {
        healthcheck(&self.connection).await
    }

    /// Count alarms owned by one already-verified Shared Auth subject.
    ///
    /// The caller supplies only the subject established by the server guard;
    /// no owner value from an untrusted request body may reach this method.
    pub async fn alarm_count_for_subject(&self, subject: &str) -> Result<u64, DbErr> {
        validate_subject(subject)?;
        let backend = self.connection.get_database_backend();
        let row = self
            .connection
            .query_one(Statement::from_sql_and_values(
                backend,
                "SELECT COUNT(*) AS alarm_count FROM happy_wakey_alarms WHERE owner_id = $1",
                [subject.into()],
            ))
            .await?
            .ok_or_else(|| DbErr::Custom("alarm count query returned no row".to_owned()))?;
        let count: i64 = i64::try_get(&row, "", "alarm_count")?;
        u64::try_from(count).map_err(|_| DbErr::Custom("alarm count was negative".to_owned()))
    }
}

#[cfg(feature = "read-write")]
impl WriteContext {
    pub async fn connect(
        database_url: &str,
        flavor: DatabaseFlavor,
        max_connections: u32,
    ) -> Result<Self, DbErr> {
        Ok(Self {
            connection: connect(database_url, max_connections).await?,
            flavor,
        })
    }

    pub const fn flavor(&self) -> DatabaseFlavor {
        self.flavor
    }

    pub async fn healthcheck(&self) -> Result<(), DbErr> {
        healthcheck(&self.connection).await
    }
}

fn validate_subject(subject: &str) -> Result<(), DbErr> {
    if subject.is_empty() || subject.len() > 512 || subject.chars().any(char::is_whitespace) {
        return Err(DbErr::Custom(
            "verified subject must be non-empty, bounded, and whitespace-free".to_owned(),
        ));
    }
    Ok(())
}

async fn connect(database_url: &str, max_connections: u32) -> Result<DatabaseConnection, DbErr> {
    if max_connections == 0 {
        return Err(DbErr::Custom(
            "max_connections must be greater than zero".to_owned(),
        ));
    }

    let mut options = ConnectOptions::new(database_url.to_owned());
    options
        .max_connections(max_connections)
        .min_connections(0)
        .connect_timeout(Duration::from_secs(10))
        .acquire_timeout(Duration::from_secs(10))
        .idle_timeout(Duration::from_secs(300))
        .sqlx_logging(false);
    Database::connect(options).await
}

async fn healthcheck(connection: &DatabaseConnection) -> Result<(), DbErr> {
    let backend = connection.get_database_backend();
    connection
        .query_one(Statement::from_string(backend, "SELECT 1"))
        .await
        .map(|_| ())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn database_flavors_remain_explicit() {
        assert_ne!(DatabaseFlavor::PostgreSql, DatabaseFlavor::CockroachDb);
    }

    #[test]
    fn subject_validation_is_fail_closed() {
        assert!(validate_subject("").is_err());
        assert!(validate_subject("subject with spaces").is_err());
        assert!(validate_subject(&"x".repeat(513)).is_err());
        assert!(validate_subject("customer:01J00000000000000000000000").is_ok());
    }
}
