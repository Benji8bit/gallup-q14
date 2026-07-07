-- Gallup Q14: примеры SQL для локального зеркала Delivery (SQLite)
-- Файл: backend/data/delivery_mirror.db
-- DBeaver: подключение "Delivery Mirror (Gallup Q14 local)"

-- Метаданные последнего pull
SELECT pulled_at, source_host, quarter_code,
       v_employee_rows, employee_rows, data_mart_rows
FROM mirror_meta;

-- Актуальный штат по направлениям (логика sync_delivery_reference.py)
WITH ranked_v AS (
    SELECT lower(email) AS email, date_to, dismissal, date_from,
           ROW_NUMBER() OVER (PARTITION BY lower(email) ORDER BY date_from DESC) AS rn
    FROM mirror_v_employee
),
latest_v AS (
    SELECT email, date_to, dismissal FROM ranked_v WHERE rn = 1
),
active_emails AS (
    SELECT email FROM latest_v
    WHERE date(date_to) >= date('now')
      AND lower(COALESCE(NULLIF(TRIM(dismissal), ''), 'false')) NOT IN ('true', '1')
),
ranked_e AS (
    SELECT lower(e.email) AS email, e.employee_direction, e.position,
           ROW_NUMBER() OVER (PARTITION BY lower(e.email) ORDER BY e.date_from DESC) AS rn
    FROM mirror_employee e
    INNER JOIN active_emails a ON lower(e.email) = a.email
),
latest_e AS (
    SELECT email, employee_direction, position FROM ranked_e WHERE rn = 1
)
SELECT employee_direction, COUNT(*) AS people
FROM latest_e
WHERE employee_direction IS NOT NULL AND TRIM(employee_direction) <> ''
GROUP BY 1
ORDER BY 2 DESC;
