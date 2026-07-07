#!/usr/bin/env python3
"""Pull Delivery Sapiens reference tables into a local SQLite mirror (VPN required).

Run monthly (or when HR data changed). Day-to-day sync uses the mirror only —
see sync_delivery_reference.py.
"""
from __future__ import annotations

import sys
from datetime import datetime, timezone
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

try:
    import psycopg2
except ImportError:
    print("Install psycopg2-binary: pip install psycopg2-binary", file=sys.stderr)
    raise SystemExit(1)

from delivery_mirror import env, init_mirror, repo_root, resolve_mirror_path  # noqa: E402
from sync_delivery_reference import quarter_code, quarter_months  # noqa: E402


def pg_connect():
    password = env("DELIVERY_SAPIENS_DB_PASSWORD")
    if not password:
        print("DELIVERY_SAPIENS_DB_PASSWORD is not set", file=sys.stderr)
        raise SystemExit(1)
    host = env("DELIVERY_SAPIENS_DB_HOST")
    user = env("DELIVERY_SAPIENS_DB_USER")
    if not host or not user:
        print("DELIVERY_SAPIENS_DB_HOST and DELIVERY_SAPIENS_DB_USER are required", file=sys.stderr)
        raise SystemExit(1)
    return psycopg2.connect(
        host=host,
        port=int(env("DELIVERY_SAPIENS_DB_PORT", "5432")),
        dbname=env("DELIVERY_SAPIENS_DB_NAME", "postgres"),
        user=user,
        password=password,
        connect_timeout=30,
    )


def pull() -> Path:
    mirror_path = resolve_mirror_path()
    mirror_path.parent.mkdir(parents=True, exist_ok=True)

    months = quarter_months()
    months_sql = ", ".join(f"'{m}'" for m in months)
    host = env("DELIVERY_SAPIENS_DB_HOST", "<delivery-host>")

    pg = pg_connect()
    cur = pg.cursor()

    cur.execute(
        """
        SELECT email, position, grade, employee_type, dismissal, date_from, date_to, direction
        FROM ods.v_employee
        """
    )
    v_rows = cur.fetchall()

    cur.execute(
        """
        SELECT email, employee_direction, position, grade, employee_type, date_from
        FROM ods.employee
        """
    )
    e_rows = cur.fetchall()

    cur.execute(
        f"""
        SELECT email, calmonth, mandays::numeric
        FROM vdm_query.v_data_mart_without_total
        WHERE calmonth IN ({months_sql})
        """
    )
    dm_rows = cur.fetchall()
    pg.close()

    import sqlite3

    conn = sqlite3.connect(mirror_path)
    init_mirror(conn)
    conn.execute("DELETE FROM mirror_v_employee")
    conn.execute("DELETE FROM mirror_employee")
    conn.execute("DELETE FROM mirror_data_mart")
    conn.execute("DELETE FROM mirror_meta")

    conn.executemany(
        """
        INSERT INTO mirror_v_employee
            (email, position, grade, employee_type, dismissal, date_from, date_to, direction)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        """,
        [
            (
                r[0],
                r[1],
                r[2],
                r[3],
                r[4],
                r[5].isoformat() if r[5] else None,
                r[6].isoformat() if r[6] else None,
                r[7],
            )
            for r in v_rows
        ],
    )
    conn.executemany(
        """
        INSERT INTO mirror_employee
            (email, employee_direction, position, grade, employee_type, date_from)
        VALUES (?, ?, ?, ?, ?, ?)
        """,
        [
            (
                r[0],
                r[1],
                r[2],
                r[3],
                r[4],
                r[5].isoformat() if r[5] else None,
            )
            for r in e_rows
        ],
    )
    conn.executemany(
        """
        INSERT INTO mirror_data_mart (email, calmonth, mandays)
        VALUES (?, ?, ?)
        """,
        [(r[0], r[1], float(r[2] or 0)) for r in dm_rows],
    )

    pulled_at = datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")
    conn.execute(
        """
        INSERT INTO mirror_meta
            (id, pulled_at, source_host, quarter_code, v_employee_rows, employee_rows, data_mart_rows)
        VALUES (1, ?, ?, ?, ?, ?, ?)
        """,
        (pulled_at, host, quarter_code(), len(v_rows), len(e_rows), len(dm_rows)),
    )
    conn.commit()
    conn.close()

    print(
        f"Pulled Delivery mirror: v_employee={len(v_rows)}, employee={len(e_rows)}, "
        f"data_mart={len(dm_rows)}, quarter={quarter_code()}, path={mirror_path}"
    )
    return mirror_path


def main() -> int:
    pull()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
