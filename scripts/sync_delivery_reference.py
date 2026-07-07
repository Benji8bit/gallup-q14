#!/usr/bin/env python3
"""Apply Delivery reference data from the local mirror into gallup-q14 SQLite.

Survey uses only Data Engineering grades from the mirror; roles are fixed in code.
Refresh mirror from PostgreSQL quarterly (or when grades change) — see README.
"""
from __future__ import annotations

import sqlite3
import sys
from datetime import date, datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from delivery_mirror import mirror_age_days, mirror_meta, resolve_app_db_path, resolve_mirror_path

SURVEY_DIRECTION = "Data Engineering"
SURVEY_ROLES = [
    "Менеджер проекта",
    "Тимлид",
    "Инженер данных/Разработчик",
]


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
    _ = value
    return value  # legacy import guard


def active_company_cte() -> str:
    """Active Data Engineering employees from local mirror."""
    return f"""
        WITH ranked_v AS (
            SELECT lower(email) AS email,
                   date_to,
                   dismissal,
                   date_from,
                   ROW_NUMBER() OVER (
                       PARTITION BY lower(email) ORDER BY date_from DESC
                   ) AS rn
            FROM mirror_v_employee
        ),
        latest_v AS (
            SELECT email, date_to, dismissal
            FROM ranked_v
            WHERE rn = 1
        ),
        active_emails AS (
            SELECT email
            FROM latest_v
            WHERE date(date_to) >= date('now')
              AND lower(COALESCE(NULLIF(TRIM(dismissal), ''), 'false')) NOT IN ('true', '1')
        ),
        ranked_e AS (
            SELECT lower(e.email) AS email,
                   e.employee_direction,
                   TRIM(e.grade) AS grade,
                   ROW_NUMBER() OVER (
                       PARTITION BY lower(e.email) ORDER BY e.date_from DESC
                   ) AS rn
            FROM mirror_employee e
            INNER JOIN active_emails a ON lower(e.email) = a.email
        ),
        latest_e AS (
            SELECT email, employee_direction, grade
            FROM ranked_e
            WHERE rn = 1
              AND employee_direction = '{SURVEY_DIRECTION}'
        )
    """


def fetch_mirror_rows(conn: sqlite3.Connection) -> dict:
    conn.row_factory = sqlite3.Row
    cur = conn.cursor()
    months = quarter_months()
    months_sql = ", ".join(f"'{m}'" for m in months)
    company_from = active_company_cte()

    def company_count(sql_body: str) -> list:
        cur.execute(company_from + sql_body)
        return [dict(row) for row in cur.fetchall()]

    grades = company_count(
        """
        SELECT grade AS value, COUNT(DISTINCT email) AS cnt
        FROM latest_e
        WHERE grade IS NOT NULL AND TRIM(grade) <> '' AND grade <> '-'
        GROUP BY 1 ORDER BY 1 ASC
        """
    )

    cur.execute(company_from + "SELECT COUNT(DISTINCT email) AS cnt FROM latest_e")
    staff_total = cur.fetchone()["cnt"]

    cur.execute(
        company_from
        + f"""
        SELECT COUNT(DISTINCT dm.email) AS cnt
        FROM mirror_data_mart dm
        INNER JOIN latest_e ON lower(dm.email) = latest_e.email
        WHERE dm.calmonth IN ({months_sql}) AND dm.mandays > 0
        """
    )
    active_delivery = cur.fetchone()["cnt"]

    cur.execute(
        company_from
        + f"""
        , mandays_q AS (
          SELECT lower(dm.email) AS email, SUM(dm.mandays) AS mandays_q
          FROM mirror_data_mart dm
          INNER JOIN latest_e ON lower(dm.email) = latest_e.email
          WHERE dm.calmonth IN ({months_sql})
          GROUP BY 1
        )
        SELECT CASE
                 WHEN mandays_q = 0 THEN '0'
                 WHEN mandays_q <= 40 THEN '1-40'
                 WHEN mandays_q <= 80 THEN '41-80'
                 ELSE '81+'
               END AS band,
               COUNT(*) AS people
        FROM mandays_q GROUP BY 1 ORDER BY 1
        """
    )
    load_bands = [dict(row) for row in cur.fetchall()]

    return {
        "grades": grades,
        "staff_total": staff_total,
        "active_delivery": active_delivery,
        "load_bands": load_bands,
    }


def write_sqlite(db_path: Path, data: dict) -> None:
    conn = sqlite3.connect(db_path)
    cur = conn.cursor()

    cur.execute("DELETE FROM delivery_org_options")
    cur.execute("DELETE FROM delivery_org_option_scopes")
    cur.execute("DELETE FROM delivery_context_stats")

    cur.execute(
        """
        INSERT INTO delivery_org_options (option_type, option_value, label_ru, employee_count, sort_order)
        VALUES ('direction', ?, ?, ?, 1)
        """,
        (SURVEY_DIRECTION, SURVEY_DIRECTION, data["staff_total"]),
    )

    for idx, row in enumerate(data["grades"], start=1):
        cur.execute(
            """
            INSERT INTO delivery_org_options (option_type, option_value, label_ru, employee_count, sort_order)
            VALUES ('grade_band', ?, ?, ?, ?)
            """,
            (row["value"], row["value"], row["cnt"], idx),
        )

    for idx, role in enumerate(SURVEY_ROLES, start=1):
        cur.execute(
            """
            INSERT INTO delivery_org_options (option_type, option_value, label_ru, employee_count, sort_order)
            VALUES ('role', ?, ?, 0, ?)
            """,
            (role, role, idx),
        )

    for row in data["load_bands"]:
        cur.execute(
            """
            INSERT INTO delivery_context_stats (stat_key, stat_value, metric, numeric_value)
            VALUES ('load_band', ?, 'people', ?)
            """,
            (row["band"], row["people"]),
        )

    cur.execute(
        """
        INSERT INTO delivery_context_stats (stat_key, stat_value, metric, numeric_value)
        VALUES ('direction_headcount', ?, 'people', ?)
        """,
        (SURVEY_DIRECTION, data["staff_total"]),
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
        f"Synced Delivery reference (DE only): staff={data['staff_total']}, "
        f"grades={len(data['grades'])}, active_qtd={data['active_delivery']}, quarter={qcode}, db={db_path}"
    )


def main() -> int:
    mirror_path = resolve_mirror_path()
    if not mirror_path.exists():
        print(
            f"Delivery mirror not found: {mirror_path}\n"
            "Run pull_delivery_mirror.py once (with VPN) to create the local copy.",
            file=sys.stderr,
        )
        return 1

    mirror_conn = sqlite3.connect(f"file:{mirror_path}?mode=ro", uri=True)
    age = mirror_age_days(mirror_conn)
    meta = mirror_meta(mirror_conn)
    if age is not None and age > 100:
        print(
            f"Warning: mirror is {age:.0f} days old (pulled {meta['pulled_at']}). "
            "Run pull_delivery_mirror.py when VPN is available (typically quarterly).",
            file=sys.stderr,
        )

    data = fetch_mirror_rows(mirror_conn)
    mirror_conn.close()

    db_path = resolve_app_db_path()
    db_path.parent.mkdir(parents=True, exist_ok=True)
    write_sqlite(db_path, data)
    if meta:
        print(f"Mirror source: {meta['source_host']}, pulled {meta['pulled_at']}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
