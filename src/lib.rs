#![forbid(unsafe_code)]
//! Happy Wakey's role-aware domain and SeaORM boundary.
//!
//! Public contracts are re-exported from the canonical interfaces crate. Raw
//! database connections remain private so callers use named, subject-scoped
//! operations.

pub use happy_wakey_interfaces as interfaces;

use chrono::DateTime;
use sea_orm::{
    ConnectOptions, ConnectionTrait, Database, DatabaseConnection, DbErr, QueryResult, Statement,
    TryGetable,
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

    /// Return the canonical alarm contracts owned by one verified subject.
    ///
    /// This is a deliberately named read operation rather than a raw query or
    /// connection escape hatch. The owner remains a bound SQL value and every
    /// selected column is mapped into the interfaces authority before leaving
    /// the read-only capability.
    pub async fn alarms_for_subject(&self, subject: &str) -> Result<Vec<interfaces::Alarm>, DbErr> {
        validate_subject(subject)?;
        let backend = self.connection.get_database_backend();
        self.connection
            .query_all(Statement::from_sql_and_values(
                backend,
                r#"SELECT CAST(id AS TEXT) AS id,
                          label,
                          CAST(local_time AS TEXT) AS local_time,
                          time_zone,
                          CAST(weekdays AS TEXT) AS weekdays_json,
                          enabled,
                          sound,
                          CAST(volume AS DOUBLE PRECISION) AS volume,
                          gradual_seconds,
                          CAST(tags AS TEXT) AS tags_json,
                          generation,
                          CAST(created_at AS TEXT) AS created_at,
                          CAST(updated_at AS TEXT) AS updated_at
                   FROM happy_wakey_alarms
                   WHERE owner_id = $1
                   ORDER BY created_at ASC"#,
                [subject.into()],
            ))
            .await?
            .into_iter()
            .map(alarm_from_row)
            .collect()
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

fn alarm_from_row(row: QueryResult) -> Result<interfaces::Alarm, DbErr> {
    let weekdays: String = row.try_get("", "weekdays_json")?;
    let volume: f64 = row.try_get("", "volume")?;
    let gradual_seconds: i32 = row.try_get("", "gradual_seconds")?;
    let tags: String = row.try_get("", "tags_json")?;
    let generation: i64 = row.try_get("", "generation")?;
    let created_at: String = row.try_get("", "created_at")?;
    let updated_at: String = row.try_get("", "updated_at")?;
    let volume = volume as f32;
    if !volume.is_finite() {
        return Err(DbErr::Custom(
            "volume is not representable as f32".to_owned(),
        ));
    }

    Ok(interfaces::Alarm {
        id: row.try_get("", "id")?,
        label: row.try_get("", "label")?,
        local_time: row.try_get("", "local_time")?,
        time_zone: row.try_get("", "time_zone")?,
        weekdays: serde_json::from_str(&weekdays)
            .map_err(|error| DbErr::Custom(format!("invalid weekdays contract: {error}")))?,
        enabled: row.try_get("", "enabled")?,
        sound: row.try_get("", "sound")?,
        volume,
        gradual_seconds: u32::try_from(gradual_seconds)
            .map_err(|_| DbErr::Custom("gradual_seconds was negative".to_owned()))?,
        tags: serde_json::from_str(&tags)
            .map_err(|error| DbErr::Custom(format!("invalid tags contract: {error}")))?,
        generation: u64::try_from(generation)
            .map_err(|_| DbErr::Custom("alarm generation was negative".to_owned()))?,
        created_at: normalize_timestamp(&created_at)?,
        updated_at: normalize_timestamp(&updated_at)?,
    })
}

fn normalize_timestamp(value: &str) -> Result<String, DbErr> {
    DateTime::parse_from_rfc3339(value)
        .or_else(|_| DateTime::parse_from_str(value, "%Y-%m-%d %H:%M:%S%.f%#z"))
        .map(|value| value.to_rfc3339())
        .map_err(|_| DbErr::Custom("database timestamp was not RFC 3339 compatible".to_owned()))
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

    #[test]
    fn database_timestamps_are_normalized_to_rfc3339() {
        assert_eq!(
            normalize_timestamp("2026-08-25 12:34:56.123+00").unwrap(),
            "2026-08-25T12:34:56.123+00:00"
        );
        assert!(normalize_timestamp("not-a-timestamp").is_err());
    }
}
