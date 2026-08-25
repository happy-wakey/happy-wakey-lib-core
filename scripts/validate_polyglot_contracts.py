#!/usr/bin/env python3
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

surfaces = {
    "SeaORM": ROOT / "src/lib.rs",
    "Drizzle": ROOT / "drizzle/src/schema.ts",
    "Prisma": ROOT / "prisma/schema.prisma",
    "GORM": ROOT / "gorm/alarm.go",
    "gRPC": ROOT / "proto/happy_wakey/v1/core.proto",
}
for name, path in surfaces.items():
    assert path.is_file(), f"missing {name} surface: {path}"

for name in ("SeaORM", "Drizzle", "Prisma", "GORM"):
    text = surfaces[name].read_text(encoding="utf-8")
    assert "happy_wakey_alarms" in text, f"{name} lost the canonical alarm table"
    assert "owner_id" in text, f"{name} lost the internal owner column"

proto = surfaces["gRPC"].read_text(encoding="utf-8")
for forbidden in ("owner_id", "bearer_token", "access_token", "refresh_token", "session_id"):
    assert forbidden not in proto, f"gRPC request surface exposes {forbidden}"
for operation in ("ListAlarms", "CreateAlarm", "TransitionOccurrence"):
    assert f"rpc {operation}" in proto, f"gRPC surface lost {operation}"

print("validated SeaORM, Drizzle, Prisma, GORM, and credential-free gRPC core surfaces")
