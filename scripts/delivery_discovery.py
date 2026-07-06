"""Read-only discovery against Delivery Sapiens PostgreSQL."""
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

    cfg = {
        "host": host,
        "port": int(env("DELIVERY_SAPIENS_DB_PORT", "5432")),
        "dbname": env("DELIVERY_SAPIENS_DB_NAME", "postgres"),
        "user": user,
        "password": password,
        "connect_timeout": 15,
    }

    conn = psycopg2.connect(**cfg)
    cur = conn.cursor(cursor_factory=RealDictCursor)

    def query(sql: str, params=None):
        cur.execute(sql, params or ())
        return cur.fetchall()

    out: dict = {}
    out["connection"] = query("SELECT current_user, session_user, current_database(), version()")[0]

    out["schemas"] = query(
        """
        SELECT nspname AS schema_name
        FROM pg_namespace
        WHERE nspname NOT LIKE 'pg_%%' AND nspname <> 'information_schema'
        ORDER BY 1
        """
    )

    patterns = (
        "employee",
        "person",
        "staff",
        "position",
        "role",
        "structure",
        "unit",
        "department",
        "org",
        "project",
        "resource",
        "manday",
        "grade",
        "practice",
        "universe",
        "roster",
        "team",
        "division",
        "cost_center",
    )
    like_clause = " OR ".join(f"table_name ILIKE '%%{p}%%'" for p in patterns)
    out["hr_objects"] = query(
        f"""
        SELECT table_schema, table_name, table_type
        FROM information_schema.tables
        WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
          AND ({like_clause})
        ORDER BY 1, 2
        """
    )

    out["ods_vdm_objects"] = query(
        """
        SELECT table_schema, table_name, table_type
        FROM information_schema.tables
        WHERE table_schema IN ('ods', 'vdm_query')
        ORDER BY 1, 2, 3
        """
    )

    for schema, table in (
        ("ods", "resource_plan"),
        ("vdm_query", "v_data_mart_without_total"),
    ):
        out[f"columns_{schema}_{table}"] = query(
            """
            SELECT column_name, data_type, is_nullable
            FROM information_schema.columns
            WHERE table_schema = %s AND table_name = %s
            ORDER BY ordinal_position
            """,
            (schema, table),
        )

    access_checks = [
        ("ods.resource_plan", "SELECT COUNT(*) AS cnt FROM ods.resource_plan"),
        (
            "vdm_query.v_data_mart_without_total",
            "SELECT COUNT(*) AS cnt FROM vdm_query.v_data_mart_without_total",
        ),
    ]
    out["restricted_locations"] = query(
        """
        SELECT table_schema, table_name, table_type
        FROM information_schema.tables
        WHERE table_name IN ('projects', 'v_resource_plan', 'vu_universe_project')
        ORDER BY 1, 2
        """
    )
    for schema, table, _ in out["restricted_locations"]:
        access_checks.append(
            (f"{schema}.{table}", f'SELECT COUNT(*) AS cnt FROM {schema}."{table}"')
        )

    out["access"] = {}
    for label, sql in access_checks:
        try:
            cur.execute(sql)
            out["access"][label] = {"ok": True, "result": dict(cur.fetchone())}
        except Exception as exc:
            conn.rollback()
            out["access"][label] = {"ok": False, "error": str(exc).split("\n")[0]}

    out["stage_id"] = query(
        """
        SELECT stage_id, COUNT(*) AS cnt
        FROM ods.resource_plan
        GROUP BY 1
        ORDER BY 2 DESC
        LIMIT 30
        """
    )

    out["calmonth_activity"] = query(
        """
        SELECT calmonth,
               COUNT(DISTINCT email) AS people,
               COUNT(*) AS rows,
               COALESCE(SUM(mandays), 0) AS mandays_sum
        FROM vdm_query.v_data_mart_without_total
        WHERE calmonth >= TO_CHAR(CURRENT_DATE - INTERVAL '6 months', 'YYYYMM')
        GROUP BY 1
        ORDER BY 1
        """
    )

    out["load_bands"] = query(
        """
        WITH q AS (
          SELECT email, SUM(mandays) AS mandays_q
          FROM vdm_query.v_data_mart_without_total
          WHERE calmonth >= TO_CHAR(DATE_TRUNC('quarter', CURRENT_DATE), 'YYYYMM')
          GROUP BY 1
        )
        SELECT CASE
                 WHEN mandays_q = 0 THEN '0'
                 WHEN mandays_q <= 40 THEN '1-40'
                 WHEN mandays_q <= 80 THEN '41-80'
                 ELSE '81+'
               END AS load_band,
               COUNT(*) AS people
        FROM q
        GROUP BY 1
        ORDER BY 1
        """
    )

    out["active_people_qtd"] = query(
        """
        SELECT COUNT(DISTINCT email) AS active_people_qtd
        FROM vdm_query.v_data_mart_without_total
        WHERE calmonth >= TO_CHAR(DATE_TRUNC('quarter', CURRENT_DATE), 'YYYYMM')
          AND mandays > 0
        """
    )[0]

    out["v_data_mart_field_cardinality"] = query(
        """
        SELECT COUNT(*) AS total_rows,
               COUNT(DISTINCT email) AS distinct_email,
               COUNT(DISTINCT fio) AS distinct_fio,
               COUNT(DISTINCT project_id) AS distinct_project_id,
               COUNT(DISTINCT project_title) AS distinct_project_title,
               COUNT(DISTINCT stage_id) AS distinct_stage_id,
               COUNT(DISTINCT version) AS distinct_version,
               COUNT(DISTINCT calmonth) AS distinct_calmonth
        FROM vdm_query.v_data_mart_without_total
        """
    )[0]

    out["v_data_mart_sample"] = query(
        """
        SELECT project_id, project_title, version, calmonth, mandays, stage_id
        FROM vdm_query.v_data_mart_without_total
        LIMIT 5
        """
    )

    out["hr_table_columns"] = {}
    for row in out["hr_objects"]:
        schema = row["table_schema"]
        table = row["table_name"]
        key = f'{schema}."{table}"'
        try:
            cols = query(
                """
                SELECT column_name, data_type
                FROM information_schema.columns
                WHERE table_schema = %s AND table_name = %s
                ORDER BY ordinal_position
                """,
                (schema, table),
            )
            cur.execute(f'SELECT COUNT(*) AS cnt FROM {schema}."{table}"')
            cnt = dict(cur.fetchone())
            out["hr_table_columns"][key] = {
                "type": row["table_type"],
                "row_count": cnt["cnt"],
                "columns": cols,
            }
        except Exception as exc:
            conn.rollback()
            out["hr_table_columns"][key] = {"error": str(exc).split("\n")[0]}

    # Extra: all columns in vdm_query and ods for manual scan
    out["all_columns_ods_vdm"] = query(
        """
        SELECT table_schema, table_name, column_name, data_type
        FROM information_schema.columns
        WHERE table_schema IN ('ods', 'vdm_query')
        ORDER BY 1, 2, ordinal_position
        """
    )

    conn.close()
    print(json.dumps(out, indent=2, default=str))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
