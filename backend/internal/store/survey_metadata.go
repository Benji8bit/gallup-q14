package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/sapiens-solutions/gallup-q14/internal/models"
)

const SurveyDirection = "Data Engineering"

var surveyRoles = []string{
	"Менеджер проекта",
	"Тимлид",
	"Инженер данных/Разработчик",
}

func (s *Store) GetSurveyOrgOptions(ctx context.Context) ([]models.OrgOptionGroup, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT option_value, label_ru
		FROM delivery_org_options
		WHERE option_type = 'grade_band'
		ORDER BY option_value ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query survey grade options: %w", err)
	}
	defer rows.Close()

	options := make([]models.OrgOption, 0, 32)
	for rows.Next() {
		var value, label string
		if err := rows.Scan(&value, &label); err != nil {
			return nil, fmt.Errorf("scan grade option: %w", err)
		}
		options = append(options, models.OrgOption{
			Value:   value,
			LabelRU: label,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return []models.OrgOptionGroup{{
		Type:    "grade_band",
		LabelRU: "Грейд",
		Options: options,
	}}, nil
}

func (s *Store) allowedGradeBands(ctx context.Context) (map[string]struct{}, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT option_value
		FROM delivery_org_options
		WHERE option_type = 'grade_band'
	`)
	if err != nil {
		return nil, fmt.Errorf("query allowed grades: %w", err)
	}
	defer rows.Close()

	allowed := map[string]struct{}{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		allowed[value] = struct{}{}
	}
	return allowed, rows.Err()
}

func isValidSurveyRole(role string) bool {
	for _, candidate := range surveyRoles {
		if role == candidate {
			return true
		}
	}
	return false
}

func normalizeSubmissionMetadata(payload models.SubmitSurveyRequest) (direction, role, grade string) {
	role = strings.TrimSpace(payload.Role)
	if role == "" {
		role = strings.TrimSpace(payload.PositionGroup)
	}
	grade = strings.TrimSpace(payload.GradeBand)
	direction = SurveyDirection
	return direction, role, grade
}

func (s *Store) validateSurveyMetadata(ctx context.Context, role, grade string) error {
	if grade == "" {
		return ErrInvalidPayload
	}
	if role == "" || !isValidSurveyRole(role) {
		return ErrInvalidPayload
	}

	allowed, err := s.allowedGradeBands(ctx)
	if err != nil {
		return err
	}
	if len(allowed) == 0 {
		return ErrInvalidPayload
	}
	if _, ok := allowed[grade]; !ok {
		return ErrInvalidPayload
	}
	return nil
}
