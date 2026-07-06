package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/sapiens-solutions/gallup-q14/internal/analytics"
	"github.com/sapiens-solutions/gallup-q14/internal/models"
	"github.com/sapiens-solutions/gallup-q14/migrations"
)

var (
	ErrDuplicateSubmission = errors.New("survey already submitted this quarter")
	ErrInvalidPayload      = errors.New("invalid survey payload")
)

type Store struct {
	db    *sql.DB
	clock func() time.Time
}

func New(dbPath string, clock func() time.Time) (*Store, error) {
	if clock == nil {
		clock = time.Now
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	s := &Store{db: db, clock: clock}
	if err := s.applyMigrations(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) applyMigrations(ctx context.Context) error {
	patterns, err := fs.Glob(migrations.Files, "*.sql")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(patterns)

	if _, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	for _, filename := range patterns {
		var exists int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM schema_migrations WHERE name = ?`, filename).Scan(&exists); err != nil {
			return fmt.Errorf("check migration %s: %w", filename, err)
		}
		if exists > 0 {
			continue
		}

		content, err := fs.ReadFile(migrations.Files, filename)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filename, err)
		}

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration tx %s: %w", filename, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", filename, err)
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (name) VALUES (?)`, filename); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("save migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", filename, err)
		}
	}

	return nil
}

func (s *Store) GetCurrentSurvey(ctx context.Context) (models.SurveyRound, []models.Question, error) {
	round, err := s.EnsureCurrentRound(ctx)
	if err != nil {
		return models.SurveyRound{}, nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, code, text_ru, dimension, sort_order, scale_min, scale_max, question_role
		FROM questions
		WHERE is_active = 1
		ORDER BY sort_order ASC
	`)
	if err != nil {
		return models.SurveyRound{}, nil, fmt.Errorf("query active questions: %w", err)
	}
	defer rows.Close()

	questions := make([]models.Question, 0, 20)
	for rows.Next() {
		var q models.Question
		if err := rows.Scan(&q.ID, &q.Code, &q.TextRU, &q.Dimension, &q.SortOrder, &q.ScaleMin, &q.ScaleMax, &q.QuestionRole); err != nil {
			return models.SurveyRound{}, nil, fmt.Errorf("scan question: %w", err)
		}
		questions = append(questions, q)
	}

	return round, questions, rows.Err()
}

func (s *Store) EnsureCurrentRound(ctx context.Context) (models.SurveyRound, error) {
	now := s.clock().UTC()
	year, quarter := quarterOf(now)
	roundCode := fmt.Sprintf("%04dQ%d", year, quarter)
	start, end := quarterRange(year, quarter)

	var round models.SurveyRound
	var startText string
	var endText string

	err := s.db.QueryRowContext(ctx, `
		SELECT id, code, year, quarter, start_date, end_date, is_active
		FROM survey_rounds
		WHERE code = ?
	`, roundCode).Scan(&round.ID, &round.Code, &round.Year, &round.Quarter, &startText, &endText, &round.IsActive)

	if err == nil {
		round.StartDate = parseStoredTime(startText)
		round.EndDate = parseStoredTime(endText)
		if !round.IsActive {
			_, _ = s.db.ExecContext(ctx, `UPDATE survey_rounds SET is_active = 0`)
			_, _ = s.db.ExecContext(ctx, `UPDATE survey_rounds SET is_active = 1 WHERE id = ?`, round.ID)
			round.IsActive = true
		}
		return round, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return models.SurveyRound{}, fmt.Errorf("query current round: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.SurveyRound{}, fmt.Errorf("begin round tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `UPDATE survey_rounds SET is_active = 0`); err != nil {
		return models.SurveyRound{}, fmt.Errorf("deactivate rounds: %w", err)
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO survey_rounds (code, year, quarter, start_date, end_date, is_active)
		VALUES (?, ?, ?, ?, ?, 1)
	`, roundCode, year, quarter, start.Format(time.RFC3339), end.Format(time.RFC3339))
	if err != nil {
		return models.SurveyRound{}, fmt.Errorf("insert round: %w", err)
	}

	roundID, err := res.LastInsertId()
	if err != nil {
		return models.SurveyRound{}, fmt.Errorf("round id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return models.SurveyRound{}, fmt.Errorf("commit round: %w", err)
	}

	return models.SurveyRound{
		ID:        roundID,
		Code:      roundCode,
		Year:      year,
		Quarter:   quarter,
		StartDate: start,
		EndDate:   end,
		IsActive:  true,
	}, nil
}

func (s *Store) SubmitSurvey(ctx context.Context, payload models.SubmitSurveyRequest) error {
	token := strings.TrimSpace(payload.AnonymousToken)
	if token == "" || len(token) > 128 {
		return ErrInvalidPayload
	}

	round, err := s.EnsureCurrentRound(ctx)
	if err != nil {
		return err
	}

	requiredQuestions, err := s.activeQuestionSpecs(ctx)
	if err != nil {
		return err
	}

	if len(payload.Answers) != len(requiredQuestions) {
		return ErrInvalidPayload
	}

	for _, spec := range requiredQuestions {
		value, ok := payload.Answers[spec.ID]
		if !ok {
			return ErrInvalidPayload
		}
		if value < spec.ScaleMin || value > spec.ScaleMax {
			return ErrInvalidPayload
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin submission tx: %w", err)
	}
	defer tx.Rollback()

	direction, positionGroup, gradeBand, employeeType, tenure := normalizeSubmissionMetadata(payload)
	deptLegacy := direction
	if deptLegacy == "" {
		deptLegacy = strings.TrimSpace(payload.Department)
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO submissions (
			round_id, anon_token, submitted_at,
			department, tenure, direction, position_group, grade_band, employee_type
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, round.ID, token, s.clock().UTC().Format(time.RFC3339),
		nullIfEmpty(deptLegacy),
		nullIfEmpty(tenure),
		nullIfEmpty(direction),
		nullIfEmpty(positionGroup),
		nullIfEmpty(gradeBand),
		nullIfEmpty(employeeType),
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return ErrDuplicateSubmission
		}
		return fmt.Errorf("insert submission: %w", err)
	}

	submissionID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("get submission id: %w", err)
	}

	for _, spec := range requiredQuestions {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO answers (submission_id, question_id, value)
			VALUES (?, ?, ?)
		`, submissionID, spec.ID, payload.Answers[spec.ID]); err != nil {
			return fmt.Errorf("insert answer for %s: %w", spec.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit submission: %w", err)
	}
	return nil
}

type questionSpec struct {
	ID           string
	ScaleMin     int
	ScaleMax     int
	QuestionRole string
}

func (s *Store) activeQuestionSpecs(ctx context.Context) ([]questionSpec, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, scale_min, scale_max, question_role
		FROM questions
		WHERE is_active = 1
		ORDER BY sort_order ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query question specs: %w", err)
	}
	defer rows.Close()

	specs := make([]questionSpec, 0, 20)
	for rows.Next() {
		var spec questionSpec
		if err := rows.Scan(&spec.ID, &spec.ScaleMin, &spec.ScaleMax, &spec.QuestionRole); err != nil {
			return nil, fmt.Errorf("scan question spec: %w", err)
		}
		specs = append(specs, spec)
	}

	if len(specs) == 0 {
		return nil, fmt.Errorf("no active questions configured")
	}
	return specs, rows.Err()
}

func (s *Store) activeQuestionIDs(ctx context.Context) ([]string, error) {
	specs, err := s.activeQuestionSpecs(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(specs))
	for _, spec := range specs {
		ids = append(ids, spec.ID)
	}
	return ids, nil
}

func (s *Store) AnalyticsRecords(ctx context.Context) ([]models.AnswerRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.code, q.id, q.code, q.dimension, q.question_role, a.value
		FROM answers a
		JOIN submissions s ON s.id = a.submission_id
		JOIN survey_rounds r ON r.id = s.round_id
		JOIN questions q ON q.id = a.question_id
		ORDER BY r.year ASC, r.quarter ASC, q.sort_order ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query analytics records: %w", err)
	}
	defer rows.Close()

	var records []models.AnswerRecord
	for rows.Next() {
		var rec models.AnswerRecord
		if err := rows.Scan(&rec.RoundCode, &rec.Question, &rec.QuestionCode, &rec.Dimension, &rec.QuestionRole, &rec.Value); err != nil {
			return nil, fmt.Errorf("scan analytics record: %w", err)
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

func (s *Store) AnalyticsGroupedRecords(ctx context.Context) ([]models.GroupedAnswerRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.code,
			COALESCE(NULLIF(TRIM(s.department), ''), 'Не указан') AS department,
			COALESCE(NULLIF(TRIM(s.direction), ''), COALESCE(NULLIF(TRIM(s.department), ''), 'Не указан')) AS direction,
			COALESCE(NULLIF(TRIM(s.position_group), ''), 'Не указан') AS position_group,
			COALESCE(NULLIF(TRIM(s.grade_band), ''), 'Не указан') AS grade_band,
			COALESCE(NULLIF(TRIM(s.employee_type), ''), 'Не указан') AS employee_type,
			COALESCE(NULLIF(TRIM(s.tenure), ''), 'Не указан') AS tenure,
			q.id, q.code, q.dimension, q.question_role, a.value
		FROM answers a
		JOIN submissions s ON s.id = a.submission_id
		JOIN survey_rounds r ON r.id = s.round_id
		JOIN questions q ON q.id = a.question_id
		ORDER BY r.year ASC, r.quarter ASC, q.sort_order ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query grouped analytics records: %w", err)
	}
	defer rows.Close()

	var records []models.GroupedAnswerRecord
	for rows.Next() {
		var rec models.GroupedAnswerRecord
		if err := rows.Scan(
			&rec.RoundCode,
			&rec.Department,
			&rec.Direction,
			&rec.PositionGroup,
			&rec.GradeBand,
			&rec.EmployeeType,
			&rec.Tenure,
			&rec.Question,
			&rec.QuestionCode,
			&rec.Dimension,
			&rec.QuestionRole,
			&rec.Value,
		); err != nil {
			return nil, fmt.Errorf("scan grouped analytics record: %w", err)
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

func (s *Store) QuestionTexts(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT code, text_ru FROM questions WHERE is_active = 1 ORDER BY sort_order ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query question texts: %w", err)
	}
	defer rows.Close()

	texts := map[string]string{}
	for rows.Next() {
		var code, text string
		if err := rows.Scan(&code, &text); err != nil {
			return nil, fmt.Errorf("scan question text: %w", err)
		}
		texts[code] = text
	}
	return texts, rows.Err()
}

func (s *Store) SubmissionCountsByRound(ctx context.Context) (map[string]int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.code, COUNT(DISTINCT s.id)
		FROM submissions s
		JOIN survey_rounds r ON r.id = s.round_id
		GROUP BY r.code
	`)
	if err != nil {
		return nil, fmt.Errorf("query submission counts: %w", err)
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var code string
		var count int
		if err := rows.Scan(&code, &count); err != nil {
			return nil, fmt.Errorf("scan submission count: %w", err)
		}
		counts[code] = count
	}
	return counts, rows.Err()
}

func (s *Store) DepartmentSubmissionCounts(ctx context.Context, roundCode string) (map[string]int, error) {
	return s.SegmentSubmissionCounts(ctx, roundCode, "direction")
}

func (s *Store) CurrentRoundSubmissionCount(ctx context.Context) (string, int, error) {
	round, err := s.EnsureCurrentRound(ctx)
	if err != nil {
		return "", 0, err
	}

	var count int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(1) FROM submissions WHERE round_id = ?
	`, round.ID).Scan(&count); err != nil {
		return "", 0, fmt.Errorf("count current submissions: %w", err)
	}
	return round.Code, count, nil
}

func (s *Store) ExportRecords(ctx context.Context) ([]models.ExportRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.code, s.submitted_at, s.anon_token, q.id, q.text_ru, q.dimension, q.question_role, a.value
		FROM answers a
		JOIN submissions s ON s.id = a.submission_id
		JOIN survey_rounds r ON r.id = s.round_id
		JOIN questions q ON q.id = a.question_id
		ORDER BY r.year ASC, r.quarter ASC, s.submitted_at ASC, q.sort_order ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query export records: %w", err)
	}
	defer rows.Close()

	exportRows := make([]models.ExportRecord, 0, 256)
	for rows.Next() {
		var record models.ExportRecord
		var submittedText string
		var questionRole string
		if err := rows.Scan(
			&record.RoundCode,
			&submittedText,
			&record.AnonToken,
			&record.QuestionID,
			&record.QuestionRU,
			&record.Dimension,
			&questionRole,
			&record.Value6,
		); err != nil {
			return nil, fmt.Errorf("scan export record: %w", err)
		}
		record.SubmittedAt = parseStoredTime(submittedText)
		ctx := analytics.BuildScoreContext(record.QuestionID, questionRole, record.Value6)
		if ctx.IsEnps {
			record.Value5 = record.Value6
			record.Favorable = false
		} else if ctx.IsSatisfaction {
			record.Value5 = record.Value6
			record.Favorable = false
		} else if ctx.IncludeInEngagement {
			record.Value5 = ctx.Score5pt
			record.Favorable = ctx.Favorable
		}
		exportRows = append(exportRows, record)
	}

	return exportRows, rows.Err()
}

func quarterOf(now time.Time) (year int, quarter int) {
	month := int(now.Month())
	quarter = (month-1)/3 + 1
	return now.Year(), quarter
}

func quarterRange(year, quarter int) (time.Time, time.Time) {
	startMonth := time.Month((quarter-1)*3 + 1)
	start := time.Date(year, startMonth, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 3, 0).Add(-time.Second)
	return start, end
}

func parseStoredTime(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC()
		}
	}
	return time.Time{}
}

func nullIfEmpty(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func RequiredDimensions() []string {
	return []string{
		"Basic Needs",
		"Growth",
		"Individual Contribution",
		"Teamwork",
	}
}
