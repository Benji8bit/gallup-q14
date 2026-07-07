"""Paths and helpers for the local Delivery mirror (offline reference copy)."""
from __future__ import annotations

import os
import sqlite3
from datetime import datetime, timezone
from pathlib import Path


def repo_root() -> Path:
    return Path(__file__).resolve().parents[1]


def env(name: str, default: str = "") -> str:
    val = os.environ.get(name)
    if val:
        return val
    try:
        import winreg

        with winreg.OpenKey(winreg.HKEY_CURRENT_USER, "Environment") as key:
            val, _ = winreg.QueryValueEx(key, name)
            return val
    except OSError:
        return default


def resolve_mirror_path() -> Path:
    raw = env("DELIVERY_MIRROR_PATH")
    if raw:
        path = Path(raw)
        return path if path.is_absolute() else repo_root() / path
    return repo_root() / "backend" / "data" / "delivery_mirror.db"


def resolve_app_db_path() -> Path:
    raw = env("DB_PATH", "./data/gallup-q14.db")
    path = Path(raw)
    if path.is_absolute():
        return path
    root = repo_root()
    candidates = [
        root / path,
        root / "backend" / path,
        root / "backend" / "data" / "gallup-q14.db",
    ]
    return next((p for p in candidates if p.exists()), candidates[0])


MIRROR_SCHEMA = """
CREATE TABLE IF NOT EXISTS mirror_meta (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    pulled_at TEXT NOT NULL,
    source_host TEXT NOT NULL,
    quarter_code TEXT NOT NULL,
    v_employee_rows INTEGER NOT NULL DEFAULT 0,
    employee_rows INTEGER NOT NULL DEFAULT 0,
    data_mart_rows INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS mirror_v_employee (
    email TEXT NOT NULL,
    position TEXT,
    grade TEXT,
    employee_type TEXT,
    dismissal TEXT,
    date_from TEXT,
    date_to TEXT,
    direction TEXT
);

CREATE TABLE IF NOT EXISTS mirror_employee (
    email TEXT NOT NULL,
    employee_direction TEXT,
    position TEXT,
    grade TEXT,
    employee_type TEXT,
    date_from TEXT
);

CREATE TABLE IF NOT EXISTS mirror_data_mart (
    email TEXT NOT NULL,
    calmonth TEXT NOT NULL,
    mandays REAL NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_mirror_v_employee_email ON mirror_v_employee (lower(email));
CREATE INDEX IF NOT EXISTS idx_mirror_employee_email ON mirror_employee (lower(email));
CREATE INDEX IF NOT EXISTS idx_mirror_data_mart_month ON mirror_data_mart (calmonth);
"""


def init_mirror(conn: sqlite3.Connection) -> None:
    conn.executescript(MIRROR_SCHEMA)


def mirror_meta(conn: sqlite3.Connection) -> dict | None:
    conn.row_factory = sqlite3.Row
    row = conn.execute(
        "SELECT pulled_at, source_host, quarter_code FROM mirror_meta WHERE id = 1"
    ).fetchone()
    return dict(row) if row else None


def mirror_age_days(conn: sqlite3.Connection) -> float | None:
    meta = mirror_meta(conn)
    if not meta:
        return None
    pulled = datetime.fromisoformat(meta["pulled_at"].replace("Z", "+00:00"))
    if pulled.tzinfo is None:
        pulled = pulled.replace(tzinfo=timezone.utc)
    now = datetime.now(timezone.utc)
    return (now - pulled).total_seconds() / 86400
