CREATE TABLE IF NOT EXISTS delivery_sync_meta (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    synced_at DATETIME NOT NULL,
    quarter_code TEXT NOT NULL,
    staff_total INTEGER NOT NULL DEFAULT 0,
    active_delivery_qtd INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS delivery_org_options (
    option_type TEXT NOT NULL,
    option_value TEXT NOT NULL,
    label_ru TEXT NOT NULL,
    employee_count INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (option_type, option_value)
);

CREATE TABLE IF NOT EXISTS delivery_context_stats (
    stat_key TEXT NOT NULL,
    stat_value TEXT NOT NULL,
    metric TEXT NOT NULL,
    numeric_value REAL NOT NULL,
    PRIMARY KEY (stat_key, stat_value, metric)
);

ALTER TABLE submissions ADD COLUMN direction TEXT;
ALTER TABLE submissions ADD COLUMN position_group TEXT;
ALTER TABLE submissions ADD COLUMN grade_band TEXT;
ALTER TABLE submissions ADD COLUMN employee_type TEXT;

INSERT OR IGNORE INTO delivery_org_options (option_type, option_value, label_ru, employee_count, sort_order) VALUES
    ('direction', 'Data Engineering', 'Data Engineering', 0, 1),
    ('direction', 'Development', 'Development', 0, 2),
    ('direction', 'backoffice', 'Backoffice', 0, 3),
    ('direction', 'easy_report', 'Easy Report', 0, 4),
    ('tenure_band', '<1 год', 'Менее 1 года', 0, 1),
    ('tenure_band', '1-3 года', '1–3 года', 0, 2),
    ('tenure_band', '3+ лет', '3+ лет', 0, 3),
    ('grade_band', 'Junior', 'Junior', 0, 1),
    ('grade_band', 'Middle', 'Middle', 0, 2),
    ('grade_band', 'Senior', 'Senior', 0, 3),
    ('grade_band', 'Lead', 'Lead / Expert', 0, 4),
    ('grade_band', 'Другое', 'Другое', 0, 5),
    ('employee_type', 'staff', 'Штат', 0, 1),
    ('employee_type', 'outstaff', 'Аутстафф', 0, 2),
    ('employee_type', 'freelance(staff)', 'Фриланс (штат)', 0, 3),
    ('employee_type', 'freelance(outstaff)', 'Фриланс (аутстафф)', 0, 4),
    ('employee_type', 'backoffice', 'Backoffice', 0, 5);
