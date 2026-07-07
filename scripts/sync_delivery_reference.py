#!/usr/bin/env python3
"""Apply Delivery reference data from the local mirror into gallup-q14 SQLite.

Reads backend/data/delivery_mirror.db (offline copy). To refresh the mirror from
PostgreSQL (VPN), run pull_delivery_mirror.py — typically once a month.
"""
from __future__ import annotations

import sqlite3
import sys
from datetime import date, datetime
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from delivery_mirror import mirror_age_days, mirror_meta, resolve_app_db_path, resolve_mirror_path


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


def active_company_cte() -> str:
    """Current employees from local mirror (SQLite window functions)."""
    return """
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
                   e.position,
                   TRIM(e.grade) AS grade,
                   e.employee_type,
                   e.date_from,
                   ROW_NUMBER() OVER (
                       PARTITION BY lower(e.email) ORDER BY e.date_from DESC
                   ) AS rn
            FROM mirror_employee e
            INNER JOIN active_emails a ON lower(e.email) = a.email
        ),
        latest_e AS (
            SELECT email, employee_direction, position, grade, employee_type, date_from
            FROM ranked_e
            WHERE rn = 1
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

    directions = company_count(
        """
        SELECT employee_direction AS value, COUNT(DISTINCT email) AS cnt
        FROM latest_e
        WHERE employee_direction IS NOT NULL AND TRIM(employee_direction) <> ''
        GROUP BY 1 ORDER BY 2 DESC
        """
    )

    positions = company_count(
        """
        SELECT position AS value, COUNT(DISTINCT email) AS cnt
        FROM latest_e
        WHERE position IS NOT NULL AND TRIM(position) <> ''
        GROUP BY 1 ORDER BY 2 DESC, 1 ASC
        """
    )

    grades = company_count(
        """
        SELECT grade AS value, COUNT(DISTINCT email) AS cnt
        FROM latest_e
        WHERE grade IS NOT NULL AND TRIM(grade) <> '' AND grade <> '-'
        GROUP BY 1 ORDER BY 2 DESC, 1 ASC
        """
    )

    employee_types = company_count(
        """
        SELECT employee_type AS value, COUNT(DISTINCT email) AS cnt
        FROM latest_e
        WHERE employee_type IS NOT NULL AND TRIM(employee_type) <> ''
        GROUP BY 1 ORDER BY 2 DESC
        """
    )

    position_by_direction = company_count(
        """
        SELECT employee_direction AS scope_value, position AS option_value,
               COUNT(DISTINCT email) AS cnt
        FROM latest_e
        WHERE employee_direction IS NOT NULL AND TRIM(employee_direction) <> ''
          AND position IS NOT NULL AND TRIM(position) <> ''
        GROUP BY 1, 2 ORDER BY 1, 3 DESC, 2 ASC
        """
    )

    grade_by_direction = company_count(
        """
        SELECT employee_direction AS scope_value, grade AS option_value,
               COUNT(DISTINCT email) AS cnt
        FROM latest_e
        WHERE employee_direction IS NOT NULL AND TRIM(employee_direction) <> ''
          AND grade IS NOT NULL AND TRIM(grade) <> '' AND grade <> '-'
        GROUP BY 1, 2 ORDER BY 1, 3 DESC, 2 ASC
        """
    )

    tenure_counts: dict[str, int] = {}
    cur.execute(
        company_from
        + """
        SELECT date_from FROM latest_e
        """
    )
    for row in cur.fetchall():
        tb = tenure_band(date.fromisoformat(row["date_from"]) if row["date_from"] else None)
        if tb:
            tenure_counts[tb] = tenure_counts.get(tb, 0) + 1

    cur.execute(company_from + "SELECT COUNT(DISTINCT email) AS cnt FROM latest_e")
    staff_total = cur.fetchone()["cnt"]

    cur.execute(
        f"""
        SELECT COUNT(DISTINCT email) AS cnt
        FROM mirror_data_mart
        WHERE calmonth IN ({months_sql}) AND mandays > 0
        """
    )
    active_delivery = cur.fetchone()["cnt"]

    cur.execute(
        f"""
        WITH q AS (
          SELECT email, SUM(mandays) AS mandays_q
          FROM mirror_data_mart
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
    load_bands = [dict(row) for row in cur.fetchall()]

    direction_headcount = company_count(
        """
        SELECT employee_direction AS direction, COUNT(DISTINCT email) AS people
        FROM latest_e
        WHERE employee_direction IS NOT NULL AND TRIM(employee_direction) <> ''
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
    if age is not None and age > 35:
        print(
            f"Warning: mirror is {age:.0f} days old (pulled {meta['pulled_at']}). "
            "Run pull_delivery_mirror.py when VPN is available.",
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
