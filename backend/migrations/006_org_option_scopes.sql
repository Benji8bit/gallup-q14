-- Scoped org options: valid position/grade values per direction (from Delivery).
CREATE TABLE IF NOT EXISTS delivery_org_option_scopes (
    option_type TEXT NOT NULL,
    option_value TEXT NOT NULL,
    scope_type TEXT NOT NULL,
    scope_value TEXT NOT NULL,
    employee_count INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (option_type, option_value, scope_type, scope_value)
);

CREATE INDEX IF NOT EXISTS idx_delivery_org_option_scopes_lookup
    ON delivery_org_option_scopes (scope_type, scope_value, option_type);
