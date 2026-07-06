"""Follow-up aggregates for Delivery Sapiens (no PII in output)."""
import json
import os
import sys

import psycopg2
from psycopg2.extras import RealDictCursor


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


def main() -> int:
    password = env("DELIVERY_SAPIENS_DB_PASSWORD")
    if not password:
        print("DELIVERY_SAPIENS_DB_PASSWORD is not set", file=sys.stderr)
        return 1

    host = env("DELIVERY_SAPIENS_DB_HOST")
    user = env("DELIVERY_SAPIENS_DB_USER")
    if not host or not user:
        print("DELIVERY_SAPIENS_DB_HOST and DELIVERY_SAPIENS_DB_USER are required", file=sys.stderr)
        return 1

    conn = psycopg2.connect(
        host=host,
        port=int(env("DELIVERY_SAPIENS_DB_PORT", "5432")),
        dbname=env("DELIVERY_SAPIENS_DB_NAME", "postgres"),
        user=user,
        password=password,
        connect_timeout=15,
    )
    cur = conn.cursor(cursor_factory=RealDictCursor)

    def query(sql: str):
        cur.execute(sql)
        return cur.fetchall()

    out: dict = {}
    for field in (
        "employee_direction",
        "position",
        "grade",
        "employee_type",
        "main_technology",
    ):
        out[f"employee_{field}"] = query(
            f"""
            SELECT {field} AS value, COUNT(*) AS cnt
            FROM ods.employee
            WHERE {field} IS NOT NULL AND TRIM({field}) <> ''
            GROUP BY 1
            ORDER BY 2 DESC
            LIMIT 25
            """
        )

    out["v_employee_columns"] = query(
        """
        SELECT column_name, data_type
        FROM information_schema.columns
        WHERE table_schema = 'ods' AND table_name = 'v_employee'
        ORDER BY ordinal_position
        """
    )
    out["v_employee_count"] = query("SELECT COUNT(*) AS cnt FROM ods.v_employee")[0]

    out["calmonth_recent"] = query(
        """
        SELECT calmonth,
               COUNT(DISTINCT email) AS people,
               COUNT(*) AS rows,
               SUM(mandays::numeric) AS mandays
        FROM vdm_query.v_data_mart_without_total
        GROUP BY 1
        ORDER BY 1 DESC
        LIMIT 12
        """
    )

    out["load_bands_q1_2026"] = query(
        """
        WITH q AS (
          SELECT email, SUM(mandays::numeric) AS mandays_q
          FROM vdm_query.v_data_mart_without_total
          WHERE calmonth IN ('01.2026', '02.2026', '03.2026')
          GROUP BY 1
        )
        SELECT CASE
                 WHEN mandays_q = 0 THEN '0'
                 WHEN mandays_q <= 40 THEN '1-40'
                 WHEN mandays_q <= 80 THEN '41-80'
                 ELSE '81+'
               END AS band,
               COUNT(*) AS people
        FROM q
        GROUP BY 1
        ORDER BY 1
        """
    )

    out["active_q1_2026"] = query(
        """
        SELECT COUNT(DISTINCT email) AS people
        FROM vdm_query.v_data_mart_without_total
        WHERE calmonth IN ('01.2026', '02.2026', '03.2026')
          AND mandays::numeric > 0
        """
    )[0]

    out["join_coverage"] = query(
        """
        SELECT COUNT(DISTINCT e.email) AS employees,
               COUNT(DISTINCT dm.email) AS in_datamart,
               COUNT(DISTINCT CASE WHEN dm.email IS NOT NULL THEN e.email END) AS matched
        FROM ods.employee e
        LEFT JOIN (
          SELECT DISTINCT email FROM vdm_query.v_data_mart_without_total
        ) dm ON lower(e.email) = lower(dm.email)
        """
    )[0]

    out["projects"] = query(
        """
        SELECT project_id, title, project_direction, type_project, status, technology
        FROM ods.projects
        ORDER BY project_id
        """
    )

    conn.close()
    print(json.dumps(out, indent=2, default=str))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
