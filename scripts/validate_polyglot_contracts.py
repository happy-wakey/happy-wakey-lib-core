#!/usr/bin/env python3
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

surfaces = {
    "SeaORM": ROOT / "src/lib.rs",
    "Drizzle": ROOT / "drizzle/src/schema.ts",
    "Prisma": ROOT / "prisma/schema.prisma",
    "GORM": ROOT / "gorm/alarm.go",
    "GORM morning": ROOT / "gorm/morning_intelligence.go",
    "gRPC": ROOT / "proto/happy_wakey/v1/core.proto",
}
for name, path in surfaces.items():
    assert path.is_file(), f"missing {name} surface: {path}"

for name in ("SeaORM", "Drizzle", "Prisma", "GORM"):
    text = surfaces[name].read_text(encoding="utf-8")
    assert "happy_wakey_alarms" in text, f"{name} lost the canonical alarm table"
    assert "owner_id" in text, f"{name} lost the internal owner column"

morning_tables = (
    "happy_wakey_tenants",
    "happy_wakey_tenant_memberships",
    "happy_wakey_connector_consents",
    "happy_wakey_source_item_candidates",
    "happy_wakey_usefulness_decisions",
    "happy_wakey_safe_deep_links",
    "happy_wakey_morning_briefings",
    "happy_wakey_embeddings",
    "happy_wakey_correlation_findings",
    "happy_wakey_chat_sessions",
)
for name in ("Drizzle", "Prisma", "GORM morning"):
    text = surfaces[name].read_text(encoding="utf-8")
    for table in morning_tables:
        assert table in text, f"{name} lost the morning table {table}"
    assert "tenant_id" in text and "subject_id" in text
    assert "created_at" in text, f"{name} lost audited creation timestamps"

authority = json.loads((ROOT / "schema-authority.json").read_text(encoding="utf-8"))
expected_revision = "3b9161f5314417cfcccd2e76c32f66d84c2eea0a"
assert authority["revision"] == expected_revision
assert authority["runtimeBoundary"] == "happy-wakey/happy-wakey-orm-core"
cargo = (ROOT / "Cargo.toml").read_text(encoding="utf-8")
assert f'rev = "{expected_revision}"' in cargo
lock = (ROOT / "Cargo.lock").read_text(encoding="utf-8")
assert expected_revision in lock

proto = surfaces["gRPC"].read_text(encoding="utf-8")
for forbidden in ("owner_id", "bearer_token", "access_token", "refresh_token", "session_id"):
    assert forbidden not in proto, f"gRPC request surface exposes {forbidden}"
for operation in ("ListAlarms", "CreateAlarm", "TransitionOccurrence"):
    assert f"rpc {operation}" in proto, f"gRPC surface lost {operation}"
for operation in (
    "GetMorningBriefing",
    "ListUsefulDeepLinks",
    "RecordUsefulnessDecision",
    "RecordEmbedding",
):
    assert f"rpc {operation}" in proto, f"gRPC surface lost {operation}"
assert not re.search(r"\b(?:tenant|subject|owner)_id\s*=", proto)
assert "repeated float vector" in proto

print("validated exact schema authority plus SeaORM, Drizzle, Prisma, GORM, and credential-free gRPC surfaces")
