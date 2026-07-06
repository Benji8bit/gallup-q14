#!/usr/bin/env python3
"""Export delivery_org_options, delivery_context_stats, delivery_sync_meta to SQL."""
from __future__ import annotations

import sqlite3
import sys
from pathlib import Path


def sql_literal(value) -> str:
    if value is None:
        return "NULL"
    if isinstance(value, (int, float)) and not isinstance(value, bool):
        return str(value)
    return "'" + str(value).replace("'", "''") + "'"


def main() -> int:
    src = Path(sys.argv[1] if len(sys.argv) > 1 else "backend/data/gallup-q14.db")
    if not src.is_absolute():
        src = Path(__file__).resolve().parents[1] / src
    out = Path(sys.argv[2] if len(sys.argv) > 2 else "scripts/delivery_reference_seed.sql")
    if not out.is_absolute():
        out = Path(__file__).resolve().parents[1] / out

    conn = sqlite3.connect(src)
    conn.row_factory = sqlite3.Row
    cur = conn.cursor()

    lines = [
        "BEGIN TRANSACTION;",
        "DELETE FROM delivery_org_options;",
        "DELETE FROM delivery_context_stats;",
        "DELETE FROM delivery_sync_meta;",
    ]

    for row in cur.execute(
        """
        SELECT option_type, option_value, label_ru, employee_count, sort_order
        FROM delivery_org_options
        ORDER BY option_type, sort_order
        """
    ):
        vals = ", ".join(sql_literal(row[k]) for k in row.keys())
        lines.append(
            f"INSERT INTO delivery_org_options (option_type, option_value, label_ru, employee_count, sort_order) VALUES ({vals});"
        )

    for row in cur.execute(
        """
        SELECT stat_key, stat_value, metric, numeric_value
        FROM delivery_context_stats
        ORDER BY stat_key, stat_value
        """
    ):
        vals = ", ".join(sql_literal(row[k]) for k in row.keys())
        lines.append(
            f"INSERT INTO delivery_context_stats (stat_key, stat_value, metric, numeric_value) VALUES ({vals});"
        )

    for row in cur.execute(
        """
        SELECT id, synced_at, quarter_code, staff_total, active_delivery_qtd
        FROM delivery_sync_meta
        """
    ):
        vals = ", ".join(sql_literal(row[k]) for k in row.keys())
        lines.append(
            f"INSERT INTO delivery_sync_meta (id, synced_at, quarter_code, staff_total, active_delivery_qtd) VALUES ({vals});"
        )

    lines.append("COMMIT;")
    out.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"Exported {len(lines) - 5} rows to {out}")
    conn.close()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
