package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/sapiens-solutions/gallup-q14/internal/models"
)

var orgOptionGroupLabels = map[string]string{
	"direction":     "Направление",
	"position":      "Должность",
	"grade_band":    "Грейд",
	"employee_type": "Тип занятости",
	"tenure_band":   "Стаж в компании",
}

var orgOptionGroupOrder = []string{
	"direction",
	"position",
	"grade_band",
	"employee_type",
	"tenure_band",
}

func (s *Store) GetOrgOptions(ctx context.Context) ([]models.OrgOptionGroup, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT option_type, option_value, label_ru, employee_count
		FROM delivery_org_options
		ORDER BY option_type, sort_order ASC, employee_count DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query org options: %w", err)
	}
	defer rows.Close()

	grouped := map[string][]models.OrgOption{}
	for rows.Next() {
		var optionType, value, label string
		var count int
		if err := rows.Scan(&optionType, &value, &label, &count); err != nil {
			return nil, fmt.Errorf("scan org option: %w", err)
		}
		grouped[optionType] = append(grouped[optionType], models.OrgOption{
			Value:         value,
			LabelRU:       label,
			EmployeeCount: count,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]models.OrgOptionGroup, 0, len(orgOptionGroupOrder))
	for _, optionType := range orgOptionGroupOrder {
		options := grouped[optionType]
		if len(options) == 0 {
			continue
		}
		out = append(out, models.OrgOptionGroup{
			Type:    optionType,
			LabelRU: orgOptionGroupLabels[optionType],
			Options: options,
		})
	}
	return out, nil
}

func (s *Store) GetOrgOptionScopes(ctx context.Context) ([]models.OrgOptionScope, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT option_type, option_value, scope_type, scope_value
		FROM delivery_org_option_scopes
		WHERE employee_count > 0
		ORDER BY scope_type, scope_value, option_type, option_value
	`)
	if err != nil {
		return nil, fmt.Errorf("query org option scopes: %w", err)
	}
	defer rows.Close()

	out := make([]models.OrgOptionScope, 0, 256)
	for rows.Next() {
		var scope models.OrgOptionScope
		if err := rows.Scan(&scope.OptionType, &scope.OptionValue, &scope.ScopeType, &scope.ScopeValue); err != nil {
			return nil, fmt.Errorf("scan org option scope: %w", err)
		}
		out = append(out, scope)
	}
	return out, rows.Err()
}

func (s *Store) GetDeliverySyncMeta(ctx context.Context) (*models.DeliverySyncMeta, error) {
	var meta models.DeliverySyncMeta
	var syncedAt string
	err := s.db.QueryRowContext(ctx, `
		SELECT synced_at, quarter_code, staff_total, active_delivery_qtd
		FROM delivery_sync_meta WHERE id = 1
	`).Scan(&syncedAt, &meta.QuarterCode, &meta.StaffTotal, &meta.ActiveDeliveryQTD)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query delivery sync meta: %w", err)
	}
	meta.SyncedAt = parseStoredTime(syncedAt)
	return &meta, nil
}

func (s *Store) GetDeliveryContextStats(ctx context.Context) ([]models.DeliveryContextStat, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT stat_key, stat_value, metric, numeric_value
		FROM delivery_context_stats
		ORDER BY stat_key, stat_value
	`)
	if err != nil {
		return nil, fmt.Errorf("query delivery context stats: %w", err)
	}
	defer rows.Close()

	stats := make([]models.DeliveryContextStat, 0, 32)
	for rows.Next() {
		var stat models.DeliveryContextStat
		if err := rows.Scan(&stat.Key, &stat.Value, &stat.Metric, &stat.NumericValue); err != nil {
			return nil, fmt.Errorf("scan delivery context stat: %w", err)
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

func (s *Store) ExpectedCountsBySegment(ctx context.Context, segmentType string) (map[string]int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT option_value, employee_count
		FROM delivery_org_options
		WHERE option_type = ?
	`, segmentType)
	if err != nil {
		return nil, fmt.Errorf("query expected counts for %s: %w", segmentType, err)
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var value string
		var count int
		if err := rows.Scan(&value, &count); err != nil {
			return nil, fmt.Errorf("scan expected count: %w", err)
		}
		counts[value] = count
	}
	return counts, rows.Err()
}

func (s *Store) SegmentSubmissionCounts(ctx context.Context, roundCode, segmentType string) (map[string]int, error) {
	column, err := segmentColumn(segmentType)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(NULLIF(TRIM(s.%s), ''), 'Не указан'), COUNT(DISTINCT s.id)
		FROM submissions s
		JOIN survey_rounds r ON r.id = s.round_id
		WHERE r.code = ?
		GROUP BY 1
	`, column)

	rows, err := s.db.QueryContext(ctx, query, roundCode)
	if err != nil {
		return nil, fmt.Errorf("query segment submission counts: %w", err)
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var value string
		var count int
		if err := rows.Scan(&value, &count); err != nil {
			return nil, fmt.Errorf("scan segment submission count: %w", err)
		}
		counts[value] = count
	}
	return counts, rows.Err()
}

func segmentColumn(segmentType string) (string, error) {
	switch segmentType {
	case "direction":
		return "direction", nil
	case "position":
		return "position_group", nil
	case "grade_band":
		return "grade_band", nil
	case "employee_type":
		return "employee_type", nil
	case "tenure_band":
		return "tenure", nil
	default:
		return "", fmt.Errorf("unknown segment type: %s", segmentType)
	}
}

func normalizeSubmissionMetadata(payload models.SubmitSurveyRequest) (direction, position, grade, empType, tenure string) {
	direction = strings.TrimSpace(payload.Direction)
	if direction == "" {
		direction = strings.TrimSpace(payload.Department)
	}
	return direction,
		strings.TrimSpace(payload.PositionGroup),
		strings.TrimSpace(payload.GradeBand),
		strings.TrimSpace(payload.EmployeeType),
		strings.TrimSpace(payload.Tenure)
}

func deliverySyncedAtPtr(meta *models.DeliverySyncMeta) *time.Time {
	if meta == nil {
		return nil
	}
	t := meta.SyncedAt
	return &t
}
