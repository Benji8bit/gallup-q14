#!/usr/bin/env python3
"""Analyze company headcount definitions against local mirror."""
from __future__ import annotations

import sqlite3
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from delivery_mirror import resolve_mirror_path  # noqa: E402


def active_company_sql(extra_latest_e: str = "") -> str:
    return f"""
        WITH ranked_v AS (
            SELECT lower(email) AS email, date_to, dismissal, date_from, employee_type,
                   ROW_NUMBER() OVER (PARTITION BY lower(email) ORDER BY date_from DESC) AS rn
            FROM mirror_v_employee
        ),
        latest_v AS (
            SELECT email, date_to, dismissal, employee_type FROM ranked_v WHERE rn = 1
        ),
        active_emails AS (
            SELECT email FROM latest_v
            WHERE date(date_to) >= date('now')
              AND lower(COALESCE(NULLIF(TRIM(dismissal), ''), 'false')) NOT IN ('true', '1')
        ),
        ranked_e AS (
            SELECT lower(e.email) AS email, e.employee_direction, e.position,
                   e.employee_type, e.grade,
                   ROW_NUMBER() OVER (PARTITION BY lower(e.email) ORDER BY e.date_from DESC) AS rn
            FROM mirror_employee e
            INNER JOIN active_emails a ON lower(e.email) = a.email
        ),
        latest_e AS (
            SELECT email, employee_direction, position, employee_type, grade
            FROM ranked_e WHERE rn = 1 {extra_latest_e}
        )
    """


def count(cur: sqlite3.Cursor, label: str, extra: str = "") -> int:
    cur.execute(active_company_sql(extra) + "SELECT COUNT(DISTINCT email) FROM latest_e")
    n = cur.fetchone()[0]
    print(f"{label}: {n}")
    return n


def main() -> int:
    path = resolve_mirror_path()
    if not path.exists():
        print(f"Mirror missing: {path}", file=sys.stderr)
        return 1

    conn = sqlite3.connect(path)
    cur = conn.cursor()

    meta = cur.execute(
        "SELECT pulled_at, source_host, v_employee_rows, employee_rows FROM mirror_meta"
    ).fetchone()
    print("mirror:", meta)

    print("\n=== Headcount variants ===")
    count(cur, "current sync (298 baseline)")

    count(
        cur,
        "staff + backoffice only",
        "AND employee_type IN ('staff', 'backoffice')",
    )
    count(
        cur,
        "exclude freelance/outstaff",
        "AND employee_type NOT IN ('freelance(staff)', 'freelance(outstaff)', 'outstaff')",
    )
    count(cur, "staff only", "AND employee_type = 'staff'")
    count(cur, "exclude easy_report direction", "AND employee_direction <> 'easy_report'")

    # v_employee only filters
    cur.execute(
        """
        WITH ranked_v AS (
            SELECT lower(email) AS email, date_to, dismissal, employee_type,
                   ROW_NUMBER() OVER (PARTITION BY lower(email) ORDER BY date_from DESC) AS rn
            FROM mirror_v_employee
        )
        SELECT employee_type, COUNT(*) FROM ranked_v
        WHERE rn = 1 AND date(date_to) >= date('now')
          AND lower(COALESCE(NULLIF(TRIM(dismissal), ''), 'false')) NOT IN ('true', '1')
        GROUP BY 1 ORDER BY 2 DESC
        """
    )
    print("\n=== By employee_type (current active filter) ===")
    for row in cur.fetchall():
        print(f"  {row[0] or '(null)'}: {row[1]}")

    cur.execute(
        active_company_sql()
        + """
        SELECT employee_direction, employee_type, COUNT(DISTINCT email) c
        FROM latest_e
        GROUP BY 1, 2 ORDER BY 3 DESC
        """
    )
    print("\n=== direction x employee_type (top 20) ===")
    for row in cur.fetchall()[:20]:
        print(f"  {row[0]} / {row[1]}: {row[2]}")

    # Check dismissal='True' on latest v row but still in active? shouldn't happen
    cur.execute(
        """
        WITH ranked_v AS (
            SELECT lower(email) AS email, date_to, dismissal,
                   ROW_NUMBER() OVER (PARTITION BY lower(email) ORDER BY date_from DESC) AS rn
            FROM mirror_v_employee
        )
        SELECT dismissal, COUNT(*) FROM ranked_v WHERE rn = 1 GROUP BY 1 ORDER BY 2 DESC
        """
    )
    print("\n=== dismissal on latest v_employee row ===")
    for row in cur.fetchall():
        print(f"  '{row[0]}': {row[1]}")

    # Multiple latest_e per person? shouldn't - distinct email
    # Duplicate emails in mirror_employee with different directions?
    cur.execute(
        active_company_sql()
        + """
        SELECT email, COUNT(*) c FROM latest_e GROUP BY 1 HAVING c > 1 LIMIT 5
        """
    )
    dups = cur.fetchall()
    print("\n=== duplicate emails in latest_e ===", len(dups))

    # Compare v_employee type vs employee type mismatch
    cur.execute(
        """
        WITH ranked_v AS (
            SELECT lower(email) AS email, employee_type AS v_type, date_to, dismissal, date_from,
                   ROW_NUMBER() OVER (PARTITION BY lower(email) ORDER BY date_from DESC) AS rn
            FROM mirror_v_employee
        ),
        active AS (
            SELECT email, v_type FROM ranked_v WHERE rn = 1
              AND date(date_to) >= date('now')
              AND lower(COALESCE(NULLIF(TRIM(dismissal), ''), 'false')) NOT IN ('true', '1')
        ),
        ranked_e AS (
            SELECT lower(e.email) AS email, e.employee_type AS e_type,
                   ROW_NUMBER() OVER (PARTITION BY lower(e.email) ORDER BY e.date_from DESC) AS rn
            FROM mirror_employee e
            INNER JOIN active a ON lower(e.email) = a.email
        ),
        latest_e AS (SELECT email, e_type FROM ranked_e WHERE rn = 1)
        SELECT a.v_type, e.e_type, COUNT(*) FROM active a
        LEFT JOIN latest_e e ON a.email = e.email
        GROUP BY 1, 2 ORDER BY 3 DESC LIMIT 15
        """
    )
    print("\n=== v_employee_type vs employee_type (active) ===")
    for row in cur.fetchall():
        print(f"  v={row[0]} / e={row[1]}: {row[2]}")

    # Strict: v dismissal empty AND employee_type staff/backoffice
    cur.execute(
        """
        WITH ranked_v AS (
            SELECT lower(email) AS email, date_to, dismissal, employee_type,
                   ROW_NUMBER() OVER (PARTITION BY lower(email) ORDER BY date_from DESC) AS rn
            FROM mirror_v_employee
        ),
        active AS (
            SELECT email FROM ranked_v WHERE rn = 1
              AND date(date_to) >= date('now')
              AND COALESCE(TRIM(dismissal), '') = ''
              AND employee_type IN ('staff', 'backoffice')
        ),
        ranked_e AS (
            SELECT lower(e.email) AS email,
                   ROW_NUMBER() OVER (PARTITION BY lower(e.email) ORDER BY e.date_from DESC) AS rn
            FROM mirror_employee e INNER JOIN active a ON lower(e.email) = a.email
        )
        SELECT COUNT(*) FROM ranked_e WHERE rn = 1
        """
    )
    print("\nstrict v: dismissal empty + v_type staff/backoffice:", cur.fetchone()[0])

    months = __import__("sync_delivery_reference", fromlist=["quarter_months"]).quarter_months()
    months_sql = ", ".join(f"'{m}'" for m in months)

    cur.execute(
        f"""
        WITH ranked_v AS (
            SELECT lower(email) AS email, date_to, dismissal,
                   ROW_NUMBER() OVER (PARTITION BY lower(email) ORDER BY date_from DESC) AS rn
            FROM mirror_v_employee
        ),
        active AS (
            SELECT email FROM ranked_v WHERE rn = 1
              AND date(date_to) >= date('now')
              AND lower(COALESCE(NULLIF(TRIM(dismissal), ''), 'false')) NOT IN ('true', '1')
        ),
        ranked_e AS (
            SELECT lower(e.email) AS email, e.employee_type,
                   ROW_NUMBER() OVER (PARTITION BY lower(e.email) ORDER BY e.date_from DESC) AS rn
            FROM mirror_employee e INNER JOIN active a ON lower(e.email) = a.email
        ),
        latest_e AS (SELECT email, employee_type FROM ranked_e WHERE rn = 1),
        dm AS (
            SELECT lower(email) AS email, SUM(mandays) AS md
            FROM mirror_data_mart WHERE calmonth IN ({months_sql}) GROUP BY 1
        )
        SELECT
            (SELECT COUNT(*) FROM latest_e) AS all_active,
            (SELECT COUNT(*) FROM latest_e e JOIN dm d ON e.email = d.email WHERE d.md > 0) AS with_mandays,
            (SELECT COUNT(*) FROM latest_e WHERE employee_type IN ('staff', 'backoffice', 'outstaff')) AS staff_bo_out,
            (SELECT COUNT(*) FROM latest_e e JOIN dm d ON e.email = d.email WHERE d.md > 0
                AND e.employee_type IN ('staff', 'backoffice', 'outstaff')) AS staff_bo_out_mandays
        """
    )
    row = cur.fetchone()
    print(f"\n=== vs data mart ({months}) ===")
    print(f"  all active (298 logic): {row[0]}")
    print(f"  with mandays>0 in quarter: {row[1]}")
    print(f"  staff+backoffice+outstaff: {row[2]}")
    print(f"  staff+backoffice+outstaff WITH mandays>0: {row[3]}")

    cur.execute(
        f"""
        WITH ranked_v AS (
            SELECT lower(email) AS email, date_to, dismissal,
                   ROW_NUMBER() OVER (PARTITION BY lower(email) ORDER BY date_from DESC) AS rn
            FROM mirror_v_employee
        ),
        active AS (
            SELECT email FROM ranked_v WHERE rn = 1
              AND date(date_to) >= date('now')
              AND lower(COALESCE(NULLIF(TRIM(dismissal), ''), 'false')) NOT IN ('true', '1')
        ),
        ranked_e AS (
            SELECT lower(e.email) AS email, e.employee_type,
                   ROW_NUMBER() OVER (PARTITION BY lower(e.email) ORDER BY e.date_from DESC) AS rn
            FROM mirror_employee e INNER JOIN active a ON lower(e.email) = a.email
        ),
        latest_e AS (SELECT email, employee_type FROM ranked_e WHERE rn = 1),
        dm AS (
            SELECT lower(email) AS email, SUM(mandays) AS md
            FROM mirror_data_mart WHERE calmonth IN ({months_sql}) GROUP BY 1
        )
        SELECT e.employee_type, COUNT(*) AS total,
               SUM(CASE WHEN COALESCE(d.md, 0) > 0 THEN 1 ELSE 0 END) AS with_mandays
        FROM latest_e e
        LEFT JOIN dm d ON e.email = d.email
        GROUP BY 1 ORDER BY 2 DESC
        """
    )
    print("\n=== employee_type | active | with mandays>0 ===")
    for et, total, wm in cur.fetchall():
        print(f"  {et or '(null)'}: {total} total, {wm} with mandays")

    # active without any mandays row in mirror at all
    cur.execute(
        f"""
        WITH ranked_v AS (
            SELECT lower(email) AS email, date_to, dismissal,
                   ROW_NUMBER() OVER (PARTITION BY lower(email) ORDER BY date_from DESC) AS rn
            FROM mirror_v_employee
        ),
        active AS (
            SELECT email FROM ranked_v WHERE rn = 1
              AND date(date_to) >= date('now')
              AND lower(COALESCE(NULLIF(TRIM(dismissal), ''), 'false')) NOT IN ('true', '1')
        ),
        ranked_e AS (
            SELECT lower(e.email) AS email, e.employee_type, e.employee_direction, e.position,
                   ROW_NUMBER() OVER (PARTITION BY lower(e.email) ORDER BY e.date_from DESC) AS rn
            FROM mirror_employee e INNER JOIN active a ON lower(e.email) = a.email
        ),
        latest_e AS (SELECT * FROM ranked_e WHERE rn = 1),
        dm AS (
            SELECT lower(email) AS email, SUM(mandays) AS md
            FROM mirror_data_mart WHERE calmonth IN ({months_sql}) GROUP BY 1
        )
        SELECT employee_type, COUNT(*) FROM latest_e e
        LEFT JOIN dm d ON e.email = d.email
        WHERE COALESCE(d.md, 0) = 0
        GROUP BY 1 ORDER BY 2 DESC
        """
    )
    print("\n=== active but mandays=0 in quarter (by type) ===")
    for row in cur.fetchall():
        print(f"  {row[0]}: {row[1]}")

    conn.close()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
