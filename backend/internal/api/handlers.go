package api

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/sapiens-solutions/gallup-q14/internal/analytics"
	"github.com/sapiens-solutions/gallup-q14/internal/delivery"
	"github.com/sapiens-solutions/gallup-q14/internal/models"
	"github.com/sapiens-solutions/gallup-q14/internal/store"
)

type Handler struct {
	store         *store.Store
	analytics     *analytics.Engine
	adminPassword string
	dbPath        string
}

func NewHandler(db *store.Store, analyticsEngine *analytics.Engine, adminPassword, dbPath string) *Handler {
	return &Handler{
		store:         db,
		analytics:     analyticsEngine,
		adminPassword: adminPassword,
		dbPath:        dbPath,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/api/health", h.health)
	r.Get("/api/survey/current", h.currentSurvey)
	r.Post("/api/survey/submit", h.submitSurvey)

	r.Route("/api/admin", func(admin chi.Router) {
		admin.Use(h.requireAdminAuth)
		admin.Get("/dashboard", h.dashboard)
		admin.Post("/sync-delivery", h.syncDelivery)
		admin.Get("/export", h.exportCSV)
	})
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"service":   "gallup-q14-backend",
		"timestamp": time.Now().UTC(),
	})
}

func (h *Handler) currentSurvey(w http.ResponseWriter, r *http.Request) {
	round, questions, err := h.store.GetCurrentSurvey(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось получить текущий опрос"})
		return
	}

	orgOptions, err := h.store.GetSurveyOrgOptions(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось получить справочники"})
		return
	}

	writeJSON(w, http.StatusOK, models.SurveyCurrentResponse{
		Round:      round,
		Questions:  questions,
		OrgOptions: orgOptions,
	})
}

func (h *Handler) submitSurvey(w http.ResponseWriter, r *http.Request) {
	var req models.SubmitSurveyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Некорректный JSON"})
		return
	}

	err := h.store.SubmitSurvey(r.Context(), req)
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, map[string]string{"status": "accepted"})
	case errors.Is(err, store.ErrDuplicateSubmission):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "Вы уже отправляли опрос в этом квартале"})
	case errors.Is(err, store.ErrInvalidPayload):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Заполните обязательные поля «О вас» и ответьте на все 13 вопросов"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось сохранить ответы"})
	}
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	records, err := h.store.AnalyticsRecords(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось получить аналитику"})
		return
	}

	groupedRecords, err := h.store.AnalyticsGroupedRecords(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось получить групповую аналитику"})
		return
	}

	questionTexts, err := h.store.QuestionTexts(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось получить тексты вопросов"})
		return
	}

	submissionCounts, err := h.store.SubmissionCountsByRound(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось получить статистику участия"})
		return
	}

	currentRoundCode, currentSubmissions, err := h.store.CurrentRoundSubmissionCount(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось получить текущий раунд"})
		return
	}

	deptSubmissionCounts, err := h.store.DepartmentSubmissionCounts(r.Context(), currentRoundCode)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось получить статистику по отделам"})
		return
	}

	deliveryMeta, err := h.store.GetDeliverySyncMeta(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось получить метаданные Delivery"})
		return
	}
	deliveryContext, err := h.store.GetDeliveryContextStats(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось получить контекст Delivery"})
		return
	}

	expectedBySegment := map[string]map[string]int{}
	submissionBySegment := map[string]map[string]int{}
	for _, segType := range []string{"grade_band", "role"} {
		expected, err := h.store.ExpectedCountsBySegment(r.Context(), segType)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось получить ожидаемые срезы"})
			return
		}
		expectedBySegment[segType] = expected

		submissions, err := h.store.SegmentSubmissionCounts(r.Context(), currentRoundCode, segType)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось получить срезы участия"})
			return
		}
		submissionBySegment[segType] = submissions
	}

	result := h.analytics.BuildDashboard(analytics.BuildInput{
		Records:                 records,
		GroupedRecords:          groupedRecords,
		QuestionTexts:           questionTexts,
		DeptSubmissionCounts:    deptSubmissionCounts,
		SubmissionCountsByRound: submissionCounts,
		CurrentRoundCode:        currentRoundCode,
		CurrentRoundSubmissions: currentSubmissions,
		DeliveryMeta:            deliveryMeta,
		DeliveryContext:         deliveryContext,
		ExpectedBySegment:       expectedBySegment,
		SubmissionBySegment:     submissionBySegment,
	})
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) syncDelivery(w http.ResponseWriter, r *http.Request) {
	message, err := delivery.RunSync(h.dbPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":  "Не удалось синхронизировать Delivery Sapiens",
			"detail": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"message": message,
	})
}

func (h *Handler) exportCSV(w http.ResponseWriter, r *http.Request) {
	rows, err := h.store.ExportRecords(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Не удалось выгрузить CSV"})
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="gallup-q14-export.csv"`)

	writer := csv.NewWriter(w)
	defer writer.Flush()

	_ = writer.Write([]string{
		"round_code", "submitted_at", "anonymous_token", "question_id", "question_ru",
		"dimension", "value_6pt", "value_5pt", "is_favorable",
	})

	for _, row := range rows {
		_ = writer.Write([]string{
			row.RoundCode,
			row.SubmittedAt.Format(time.RFC3339),
			row.AnonToken,
			row.QuestionID,
			row.QuestionRU,
			row.Dimension,
			intToStr(row.Value6),
			intToStr(row.Value5),
			boolToStr(row.Favorable),
		})
	}
}

func (h *Handler) requireAdminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.authorized(r.Header.Get("Authorization")) {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="admin", charset="UTF-8"`)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Требуется админ-аутентификация"})
	})
}

func (h *Handler) authorized(authHeader string) bool {
	if authHeader == "" || h.adminPassword == "" {
		return false
	}

	if token, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
		return secureCompare(strings.TrimSpace(token), h.adminPassword)
	}

	if encoded, ok := strings.CutPrefix(authHeader, "Basic "); ok {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
		if err != nil {
			return false
		}
		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 {
			return false
		}
		username := parts[0]
		password := parts[1]
		return secureCompare(username, "admin") && secureCompare(password, h.adminPassword)
	}

	return false
}

func secureCompare(left, right string) bool {
	if len(left) != len(right) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func intToStr(v int) string {
	return strconv.Itoa(v)
}

func boolToStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
