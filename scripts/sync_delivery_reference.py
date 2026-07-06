#!/usr/bin/env python3
"""Sync Delivery Sapiens reference data into gallup-q14 SQLite (no PII stored).

Org options mirror ods.employee structure with active-in-quarter counts (distinct email,
mandays > 0 in current quarter). Position and grade lists cascade by employee_direction
via delivery_org_option_scopes.
"""
from __future__ import annotations

import os
import sqlite3
import sys
from datetime import date, datetime
from pathlib import Path

try:
    import psycopg2
    from psycopg2.extras import RealDictCursor
except ImportError:
    print("Install psycopg2-binary: pip install psycopg2-binary", file=sys.stderr)
    raise SystemExit(1)


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


def direction_label(value: str) -> str:
    labels = {
        "Data Engineering": "Data Engineering",
        "Development": "Development",
        "backoffice": "Backoffice",
        "easy_report": "Easy Report",
    }
    return labels.get(value, value)


def tenure_band(date_from: date | None) -> str:
    if not date_from:
        return ""
    years = (date.today() - date_from).days / 365.25
    if years < 1:
        return "<1 год"
    if years < 3:
        return "1-3 года"
    return "3+ лет"


def quarter_code() -> str:
    today = date.today()
    q = (today.month - 1) // 3 + 1
    return f"{today.year}Q{q}"


def quarter_months() -> list[str]:
    today = date.today()
    q = (today.month - 1) // 3 + 1
    start_month = (q - 1) * 3 + 1
    return [f"{m:02d}.{today.year}" for m in range(start_month, start_month + 3)]


def employee_type_label(value: str) -> str:
    labels = {
        "staff": "Штат",
        "outstaff": "Аутстафф",
        "freelance(staff)": "Фриланс (штат)",
        "freelance(outstaff)": "Фриланс (аутстафф)",
        "backoffice": "Backoffice",
        "easy_report": "Easy Report",
    }
    return labels.get(value, value)


def fetch_pg_rows(conn) -> dict:
    cur = conn.cursor(cursor_factory=RealDictCursor)
    months = quarter_months()
    months_sql = ", ".join(f"'{m}'" for m in months)

    active_from = f"""
        WITH active_emails AS (
            SELECT DISTINCT lower(email) AS email
            FROM vdm_query.v_data_mart_without_total
            WHERE calmonth IN ({months_sql}) AND mandays::numeric > 0
        )
    """

    def active_count(sql_body: str) -> list:
        cur.execute(active_from + sql_body)
        return cur.fetchall()

    directions = active_count(
        """
        SELECT e.employee_direction AS value, COUNT(DISTINCT e.email) AS cnt
        FROM ods.employee e
        INNER JOIN active_emails a ON lower(e.email) = a.email
        WHERE e.employee_direction IS NOT NULL AND TRIM(e.employee_direction) <> ''
        GROUP BY 1 ORDER BY 2 DESC
        """
    )

    positions = active_count(
        """
        SELECT e.position AS value, COUNT(DISTINCT e.email) AS cnt
        FROM ods.employee e
        INNER JOIN active_emails a ON lower(e.email) = a.email
        WHERE e.position IS NOT NULL AND TRIM(e.position) <> ''
        GROUP BY 1 ORDER BY 2 DESC, 1 ASC
        """
    )

    grades = active_count(
        """
        SELECT TRIM(e.grade) AS value, COUNT(DISTINCT e.email) AS cnt
        FROM ods.employee e
        INNER JOIN active_emails a ON lower(e.email) = a.email
        WHERE e.grade IS NOT NULL AND TRIM(e.grade) <> '' AND TRIM(e.grade) <> '-'
        GROUP BY 1 ORDER BY 2 DESC, 1 ASC
        """
    )

    employee_types = active_count(
        """
        SELECT e.employee_type AS value, COUNT(DISTINCT e.email) AS cnt
        FROM ods.employee e
        INNER JOIN active_emails a ON lower(e.email) = a.email
        WHERE e.employee_type IS NOT NULL AND TRIM(e.employee_type) <> ''
        GROUP BY 1 ORDER BY 2 DESC
        """
    )

    position_by_direction = active_count(
        """
        SELECT e.employee_direction AS scope_value, e.position AS option_value,
               COUNT(DISTINCT e.email) AS cnt
        FROM ods.employee e
        INNER JOIN active_emails a ON lower(e.email) = a.email
        WHERE e.employee_direction IS NOT NULL AND TRIM(e.employee_direction) <> ''
          AND e.position IS NOT NULL AND TRIM(e.position) <> ''
        GROUP BY 1, 2 ORDER BY 1, 3 DESC, 2 ASC
        """
    )

    grade_by_direction = active_count(
        """
        SELECT e.employee_direction AS scope_value, TRIM(e.grade) AS option_value,
               COUNT(DISTINCT e.email) AS cnt
        FROM ods.employee e
        INNER JOIN active_emails a ON lower(e.email) = a.email
        WHERE e.employee_direction IS NOT NULL AND TRIM(e.employee_direction) <> ''
          AND e.grade IS NOT NULL AND TRIM(e.grade) <> '' AND TRIM(e.grade) <> '-'
        GROUP BY 1, 2 ORDER BY 1, 3 DESC, 2 ASC
        """
    )

    tenure_counts: dict[str, int] = {}
    cur.execute(
        f"""
        {active_from}
        SELECT e.date_from
        FROM ods.employee e
        INNER JOIN active_emails a ON lower(e.email) = a.email
        """
    )
    for row in cur.fetchall():
        tb = tenure_band(row["date_from"])
        if tb:
            tenure_counts[tb] = tenure_counts.get(tb, 0) + 1

    cur.execute("SELECT COUNT(DISTINCT email) AS cnt FROM ods.employee")
    staff_total = cur.fetchone()["cnt"]

    cur.execute(
        f"""
        SELECT COUNT(DISTINCT email) AS cnt
        FROM vdm_query.v_data_mart_without_total
        WHERE calmonth IN ({months_sql}) AND mandays::numeric > 0
        """
    )
    active_delivery = cur.fetchone()["cnt"]

    cur.execute(
        f"""
        WITH q AS (
          SELECT email, SUM(mandays::numeric) AS mandays_q
          FROM vdm_query.v_data_mart_without_total
          WHERE calmonth IN ({months_sql})
          GROUP BY 1
        )
        SELECT CASE
                 WHEN mandays_q = 0 THEN '0'
                 WHEN mandays_q <= 40 THEN '1-40'
                 WHEN mandays_q <= 80 THEN '41-80'
                 ELSE '81+'
               END AS band,
               COUNT(*) AS people
        FROM q GROUP BY 1 ORDER BY 1
        """
    )
    load_bands = cur.fetchall()

    direction_headcount = active_count(
        """
        SELECT e.employee_direction AS direction, COUNT(DISTINCT e.email) AS people
        FROM ods.employee e
        INNER JOIN active_emails a ON lower(e.email) = a.email
        WHERE e.employee_direction IS NOT NULL AND TRIM(e.employee_direction) <> ''
        GROUP BY 1 ORDER BY 2 DESC
        """
    )

    return {
        "directions": directions,
        "positions": positions,
        "grades": grades,
        "employee_types": employee_types,
        "position_by_direction": position_by_direction,
        "grade_by_direction": grade_by_direction,
        "tenure_counts": tenure_counts,
        "staff_total": staff_total,
        "active_delivery": active_delivery,
        "load_bands": load_bands,
        "direction_headcount": direction_headcount,
    }


def write_sqlite(db_path: Path, data: dict) -> None:
    conn = sqlite3.connect(db_path)
    cur = conn.cursor()

    cur.execute("DELETE FROM delivery_org_options")
    cur.execute("DELETE FROM delivery_org_option_scopes")
    cur.execute("DELETE FROM delivery_context_stats")

    sort = 0
    for row in data["directions"]:
        sort += 1
        value = row["value"]
        cur.execute(
            """
            INSERT INTO delivery_org_options (option_type, option_value, label_ru, employee_count, sort_order)
            VALUES ('direction', ?, ?, ?, ?)
            """,
            (value, direction_label(value), row["cnt"], sort),
        )

    sort = 0
    for row in data["positions"]:
        sort += 1
        cur.execute(
            """
            INSERT INTO delivery_org_options (option_type, option_value, label_ru, employee_count, sort_order)
            VALUES ('position', ?, ?, ?, ?)
            """,
            (row["value"], row["value"], row["cnt"], sort),
        )
    cur.execute(
        """
        INSERT INTO delivery_org_options (option_type, option_value, label_ru, employee_count, sort_order)
        VALUES ('position', 'Другое', 'Другая должность', 0, 999)
        """
    )

    sort = 0
    for row in data["grades"]:
        sort += 1
        value = row["value"]
        cur.execute(
            """
            INSERT INTO delivery_org_options (option_type, option_value, label_ru, employee_count, sort_order)
            VALUES ('grade_band', ?, ?, ?, ?)
            """,
            (value, value, row["cnt"], sort),
        )

    tenure_order = ["<1 год", "1-3 года", "3+ лет"]
    labels = {"<1 год": "Менее 1 года", "1-3 года": "1–3 года", "3+ лет": "3+ лет"}
    for idx, band in enumerate(tenure_order, start=1):
        cnt = data["tenure_counts"].get(band, 0)
        cur.execute(
            """
            INSERT INTO delivery_org_options (option_type, option_value, label_ru, employee_count, sort_order)
            VALUES ('tenure_band', ?, ?, ?, ?)
            """,
            (band, labels[band], cnt, idx),
        )

    sort = 0
    for row in data["employee_types"]:
        sort += 1
        value = row["value"]
        cur.execute(
            """
            INSERT INTO delivery_org_options (option_type, option_value, label_ru, employee_count, sort_order)
            VALUES ('employee_type', ?, ?, ?, ?)
            """,
            (value, employee_type_label(value), row["cnt"], sort),
        )

    for row in data["position_by_direction"]:
        cur.execute(
            """
            INSERT INTO delivery_org_option_scopes
                (option_type, option_value, scope_type, scope_value, employee_count)
            VALUES ('position', ?, 'direction', ?, ?)
            """,
            (row["option_value"], row["scope_value"], row["cnt"]),
        )

    for row in data["grade_by_direction"]:
        cur.execute(
            """
            INSERT INTO delivery_org_option_scopes
                (option_type, option_value, scope_type, scope_value, employee_count)
            VALUES ('grade_band', ?, 'direction', ?, ?)
            """,
            (row["option_value"], row["scope_value"], row["cnt"]),
        )

    for row in data["load_bands"]:
        cur.execute(
            """
            INSERT INTO delivery_context_stats (stat_key, stat_value, metric, numeric_value)
            VALUES ('load_band', ?, 'people', ?)
            """,
            (row["band"], row["people"]),
        )

    for row in data["direction_headcount"]:
        cur.execute(
            """
            INSERT INTO delivery_context_stats (stat_key, stat_value, metric, numeric_value)
            VALUES ('direction_headcount', ?, 'people', ?)
            """,
            (row["direction"], row["people"]),
        )

    cur.execute(
        """
        INSERT INTO delivery_context_stats (stat_key, stat_value, metric, numeric_value)
        VALUES ('summary', 'staff_total', 'people', ?)
        """,
        (data["staff_total"],),
    )
    cur.execute(
        """
        INSERT INTO delivery_context_stats (stat_key, stat_value, metric, numeric_value)
        VALUES ('summary', 'active_delivery_qtd', 'people', ?)
        """,
        (data["active_delivery"],),
    )

    now = datetime.utcnow().isoformat() + "Z"
    qcode = quarter_code()
    cur.execute("DELETE FROM delivery_sync_meta")
    cur.execute(
        """
        INSERT INTO delivery_sync_meta (id, synced_at, quarter_code, staff_total, active_delivery_qtd)
        VALUES (1, ?, ?, ?, ?)
        """,
        (now, qcode, data["staff_total"], data["active_delivery"]),
    )

    conn.commit()
    conn.close()
    print(
        f"Synced Delivery reference: staff={data['staff_total']}, "
        f"active_qtd={data['active_delivery']}, quarter={qcode}, db={db_path}"
    )


def main() -> int:
    password = env("DELIVERY_SAPIENS_DB_PASSWORD")
    if not password:
        print("DELIVERY_SAPIENS_DB_PASSWORD is not set", file=sys.stderr)
        return 1

    db_path = Path(env("DB_PATH", "./data/gallup-q14.db"))
    if not db_path.is_absolute():
        repo_root = Path(__file__).resolve().parents[1]
        candidates = [
            repo_root / db_path,
            repo_root / "backend" / db_path,
            repo_root / "backend" / "data" / "gallup-q14.db",
        ]
        db_path = next((p for p in candidates if p.exists()), candidates[0])

    host = env("DELIVERY_SAPIENS_DB_HOST")
    user = env("DELIVERY_SAPIENS_DB_USER")
    if not host or not user:
        print("DELIVERY_SAPIENS_DB_HOST and DELIVERY_SAPIENS_DB_USER are required", file=sys.stderr)
        return 1

    pg = psycopg2.connect(
        host=host,
        port=int(env("DELIVERY_SAPIENS_DB_PORT", "5432")),
        dbname=env("DELIVERY_SAPIENS_DB_NAME", "postgres"),
        user=user,
        password=password,
        connect_timeout=20,
    )
    data = fetch_pg_rows(pg)
    pg.close()
    write_sqlite(db_path, data)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
