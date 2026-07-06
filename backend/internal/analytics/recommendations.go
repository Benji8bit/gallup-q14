package analytics

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sapiens-solutions/gallup-q14/internal/models"
)

type recommendationInput struct {
	EngagementScore   float64
	SatisfactionScore float64
	EnpsScore         *models.EnpsScore
	Dimensions        []models.DimensionScore
	QuestionScores    []models.QuestionScore
	DepartmentScores  []models.DepartmentScore
	CurrentRoundCode  string
}

func buildRecommendations(input recommendationInput) []models.RecommendationItem {
	items := make([]models.RecommendationItem, 0, 12)

	if len(input.Dimensions) == 0 {
		return []models.RecommendationItem{{
			ID:           "empty",
			Scope:        "general",
			Title:        "Недостаточно данных",
			Action:       "Запустите сбор ответов в текущем квартале и дождитесь репрезентативной выборки.",
			TargetGroups: []string{"HR", "Генеральное руководство"},
			Basis:        "В базе нет ответов для расчёта измерений и групповых срезов.",
		}}
	}

	sortedDims := append([]models.DimensionScore(nil), input.Dimensions...)
	sort.Slice(sortedDims, func(i, j int) bool {
		if sortedDims[i].FavorablePct == sortedDims[j].FavorablePct {
			return sortedDims[i].AverageScore5pt < sortedDims[j].AverageScore5pt
		}
		return sortedDims[i].FavorablePct < sortedDims[j].FavorablePct
	})

	sortedQuestions := append([]models.QuestionScore(nil), input.QuestionScores...)
	sort.Slice(sortedQuestions, func(i, j int) bool {
		return sortedQuestions[i].FavorablePct < sortedQuestions[j].FavorablePct
	})

	eng := input.EngagementScore
	engCopy := eng
	items = append(items, models.RecommendationItem{
		ID:           "general-engagement",
		Scope:        "general",
		Title:        "Общий уровень вовлечённости",
		Action:       generalEngagementAction(eng),
		TargetGroups: []string{"Генеральное руководство", "HR"},
		Basis: fmt.Sprintf(
			"Индекс вовлечённости за %s составляет %.1f%% (%s). Порог внимания HR — ниже 50%%, зона риска — ниже 35%%.",
			input.CurrentRoundCode, eng, engagementBand(eng),
		),
		Evidence: models.RecommendationEvidence{EngagementScore: &engCopy},
	})

	if input.EnpsScore != nil && input.EnpsScore.Total > 0 {
		enps := input.EnpsScore.Score
		enpsCopy := enps
		items = append(items, models.RecommendationItem{
			ID:           "general-enps",
			Scope:        "general",
			Title:        "eNPS — готовность рекомендовать компанию",
			Action:       enpsAction(enps),
			TargetGroups: []string{"Генеральное руководство", "HR", "Руководители практик"},
			Basis: fmt.Sprintf(
				"eNPS за %s: %.0f (промоутеры %.0f%%, нейтралы %.0f%%, критики %.0f%%). "+
					"Норма — выше +30; ниже 0 — тревожный сигнал.",
				input.CurrentRoundCode,
				enps,
				input.EnpsScore.PromotersPct,
				input.EnpsScore.PassivesPct,
				input.EnpsScore.DetractorsPct,
			),
			Evidence: models.RecommendationEvidence{EngagementScore: &enpsCopy},
		})
	}

	if input.SatisfactionScore > 0 {
		sat := input.SatisfactionScore
		items = append(items, models.RecommendationItem{
			ID:           "general-satisfaction",
			Scope:        "general",
			Title:        "Удовлетворённость компанией (Q00)",
			Action:       "Проведите качественные интервью по мотивации удержания: что удерживает и что уводит из компании.",
			TargetGroups: []string{"HR", "Генеральное руководство"},
			Basis: fmt.Sprintf(
				"Средняя удовлетворённость Q00 — %.1f/5. Показатель ниже 3.5 сигнализирует о системных проблемах опыта сотрудника.",
				sat,
			),
			RelatedQuestions: []string{"Q00"},
			Evidence:         models.RecommendationEvidence{SatisfactionScore: &sat},
		})
	}

	topN := 3
	if len(sortedDims) < topN {
		topN = len(sortedDims)
	}
	for i := 0; i < topN; i++ {
		d := sortedDims[i]
		if d.FavorablePct >= 70 {
			continue
		}
		fav := d.FavorablePct
		avg := d.AverageScore5pt
		relatedQ := weakestQuestionsInDimension(sortedQuestions, d.Dimension, 2)
		items = append(items, models.RecommendationItem{
			ID:               fmt.Sprintf("general-dim-%d", i),
			Scope:            "general",
			Title:            fmt.Sprintf("Измерение «%s»", dimensionLabels()[d.Dimension]),
			Action:           dimensionAction(d.Dimension),
			TargetGroups:     strings.Split(dimensionAudiences()[d.Dimension], ", "),
			Basis: fmt.Sprintf(
				"Favorable по измерению «%s» — %.1f%%, средний балл %.2f/5 (n=%d). Слабые вопросы: %s.",
				dimensionLabels()[d.Dimension], fav, avg, d.ResponseCount, strings.Join(relatedQ, ", "),
			),
			RelatedDimension: d.Dimension,
			RelatedQuestions: relatedQ,
			Evidence: models.RecommendationEvidence{
				FavorablePct:    &fav,
				AverageScore5pt: &avg,
				ResponseCount:   d.ResponseCount,
			},
		})
	}

	deptByDim := groupDepartmentScores(input.DepartmentScores)
	type deptIssue struct {
		dept  string
		dim   string
		score models.DepartmentScore
	}
	var deptIssues []deptIssue
	for dept, dims := range deptByDim {
		if dept == "Не указан" {
			continue
		}
		for dim, sc := range dims {
			if sc.FavorablePct < 65 {
				deptIssues = append(deptIssues, deptIssue{dept: dept, dim: dim, score: sc})
			}
		}
	}
	sort.Slice(deptIssues, func(i, j int) bool {
		return deptIssues[i].score.FavorablePct < deptIssues[j].score.FavorablePct
	})
	limit := 4
	if len(deptIssues) < limit {
		limit = len(deptIssues)
	}
	for i := 0; i < limit; i++ {
		issue := deptIssues[i]
		fav := issue.score.FavorablePct
		avg := issue.score.AverageScore5pt
		items = append(items, models.RecommendationItem{
			ID:           fmt.Sprintf("group-dept-%d", i),
			Scope:        "group",
			Title:        fmt.Sprintf("Отдел «%s» — %s", issue.dept, dimensionLabels()[issue.dim]),
			Action:       departmentAction(issue.dept, issue.dim),
			TargetGroups: departmentTargets(issue.dept, issue.dim),
			Basis: fmt.Sprintf(
				"В отделе «%s» favorable по «%s» — %.1f%%, средний %.2f/5 (n=%d). "+
					"Основано на self-reported отделе из анонимного опроса.",
				issue.dept, dimensionLabels()[issue.dim], fav, avg, issue.score.ResponseCount,
			),
			RelatedDimension: issue.dim,
			Evidence: models.RecommendationEvidence{
				FavorablePct:    &fav,
				AverageScore5pt: &avg,
				Department:      issue.dept,
				ResponseCount:   issue.score.ResponseCount,
			},
		})
	}

	qLimit := 3
	if len(sortedQuestions) < qLimit {
		qLimit = len(sortedQuestions)
	}
	for i := 0; i < qLimit; i++ {
		q := sortedQuestions[i]
		if q.Code == "Q00" || q.FavorablePct >= 60 {
			continue
		}
		fav := q.FavorablePct
		avg := q.AverageScore5pt
		meta := questionMeta()[q.Code]
		items = append(items, models.RecommendationItem{
			ID:    fmt.Sprintf("group-q-%s", q.Code),
			Scope: "group",
			Title: fmt.Sprintf("Вопрос %s", q.Code),
			Action: fmt.Sprintf("%s Сфокусируйте action plan на: %s.", dimensionAction(q.Dimension), meta.leadershipFocus),
			TargetGroups: questionTargets(q.Code, q.Dimension),
			Basis: fmt.Sprintf(
				"%s — favorable %.1f%%, средний %.2f/5 (n=%d). Измеряет: %s.",
				q.Code, fav, avg, q.ResponseCount, meta.whatItMeasures,
			),
			RelatedDimension: q.Dimension,
			RelatedQuestions: []string{q.Code},
			Evidence: models.RecommendationEvidence{
				FavorablePct:    &fav,
				AverageScore5pt: &avg,
				QuestionCode:    q.Code,
				ResponseCount:   q.ResponseCount,
			},
		})
	}

	return items
}

func buildDepartmentBreakdown(grouped []models.GroupedAnswerRecord, currentRound string, deptSubmissions map[string]int) []models.DepartmentScore {
	type agg struct {
		fav, sum, cnt int
	}
	type subAgg map[string]*agg
	byDeptDim := map[string]subAgg{}

	for _, rec := range grouped {
		if rec.RoundCode != currentRound {
			continue
		}
		ctx := BuildScoreContext(rec.Question, rec.QuestionRole, rec.Value)
		if !ctx.IncludeInEngagement {
			continue
		}
		score5 := ctx.Score5pt
		dept := rec.Direction
		if dept == "" || dept == "Не указан" {
			dept = rec.Department
		}
		if byDeptDim[dept] == nil {
			byDeptDim[dept] = subAgg{}
		}
		a := byDeptDim[dept][rec.Dimension]
		if a == nil {
			a = &agg{}
			byDeptDim[dept][rec.Dimension] = a
		}
		a.cnt++
		a.sum += score5
		if score5 >= 4 {
			a.fav++
		}
	}

	out := make([]models.DepartmentScore, 0)
	for dept, dims := range byDeptDim {
		subCount := deptSubmissions[dept]
		for dim, a := range dims {
			out = append(out, models.DepartmentScore{
				Department:      dept,
				Dimension:       dim,
				FavorablePct:    pct(a.fav, a.cnt),
				AverageScore5pt: avg(a.sum, a.cnt),
				ResponseCount:   a.cnt,
				SubmissionCount: subCount,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Department == out[j].Department {
			return out[i].FavorablePct < out[j].FavorablePct
		}
		return out[i].Department < out[j].Department
	})
	return out
}

func recommendationsSummary(items []models.RecommendationItem) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, fmt.Sprintf("%s: %s", item.Title, item.Action))
	}
	return out
}

func groupDepartmentScores(scores []models.DepartmentScore) map[string]map[string]models.DepartmentScore {
	out := map[string]map[string]models.DepartmentScore{}
	for _, s := range scores {
		if out[s.Department] == nil {
			out[s.Department] = map[string]models.DepartmentScore{}
		}
		out[s.Department][s.Dimension] = s
	}
	return out
}

func weakestQuestionsInDimension(questions []models.QuestionScore, dimension string, n int) []string {
	filtered := make([]models.QuestionScore, 0)
	for _, q := range questions {
		if q.Dimension == dimension && q.Code != "Q00" {
			filtered = append(filtered, q)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].FavorablePct < filtered[j].FavorablePct
	})
	if len(filtered) < n {
		n = len(filtered)
	}
	codes := make([]string, 0, n)
	for i := 0; i < n; i++ {
		codes = append(codes, filtered[i].Code)
	}
	return codes
}

func engagementBand(pct float64) string {
	switch {
	case pct >= 70:
		return "сильная зона"
	case pct >= 50:
		return "зона внимания"
	default:
		return "критическая зона"
	}
}

func generalEngagementAction(pct float64) string {
	if pct >= 70 {
		return "Закрепите практики, давшие высокий результат: масштабируйте успешные форматы 1:1, признания и онбординга на другие команды."
	}
	if pct >= 50 {
		return "Сформируйте квартальный action plan по 2–3 слабым измерениям; назначьте владельцев из числа руководителей практик."
	}
	return "Эскалируйте тему вовлечённости на уровень топ-менеджмента: нужен системный план с измеримыми KPI и ежемесячным контролем."
}

func enpsAction(score float64) string {
	if score >= 30 {
		return "Поддерживайте сильные стороны EVP: публично делитесь историями промоутеров, усиливайте реферальную программу и менторство."
	}
	if score >= 0 {
		return "Проанализируйте причины нейтральности: exit-интервью, фокус-группы с критиками и пассивами; сформируйте план по удержанию."
	}
	return "Срочно: разберите драйверы недовольства (нагрузка, компенсация, руководство, проекты); назначьте ответственных и сроки улучшений."
}

func dimensionAction(dim string) string {
	actions := map[string]string{
		"Basic Needs":             "Проведите аудит ожиданий по ролям и доступности инструментов на DE-проектах; обновите onboarding checklist.",
		"Individual Contribution": "Введите регулярное признание (минимум раз в неделю) и структурированные 1:1 с фокусом на развитие.",
		"Teamwork":                "Усилите ритуалы командной работы: ретро, peer feedback, совместные стандарты качества delivery.",
		"Growth":                  "Запустите IDP на 6 месяцев для ключевых ролей DE; выделите бюджет на обучение и сертификации.",
		"Wellbeing":               "Пересмотрите загрузку и дедлайны; внедрите мониторинг переработок и политику «no meeting» блоков.",
		"Company Specific":        "Улучшите передачу клиентского контекста на проекты; назначьте технического куратора на сложные engagement.",
	}
	if a, ok := actions[dim]; ok {
		return a
	}
	return "Проведите фокус-группу по результатам измерения и сформируйте план улучшений."
}

func departmentAction(dept, dim string) string {
	base := dimensionAction(dim)
	switch dept {
	case "Data Engineering":
		return base + " Приоритет — стандарты DE, архитектурные ревью и доступ к средам разработки."
	case "Development":
		return base + " Приоритет — прозрачность требований, code review и связь разработки с бизнес-ценностью."
	case "backoffice":
		return base + " Приоритет — внутренние сервисы HR/админ, поддержка delivery-команд."
	case "easy_report":
		return base + " Приоритет — инструменты отчётности и связь с заказчиками аналитики."
	case "Дата-инжиниринг":
		return base + " Приоритет — стандарты DE, архитектурные ревью и доступ к средам разработки."
	case "Аналитика":
		return base + " Приоритет — прозрачность требований к отчётности и связь аналитики с бизнес-ценностью."
	case "Управление проектами":
		return base + " Приоритет — realistic planning, эскалации и баланс ожиданий клиента и команды."
	case "HR":
		return base + " Приоритет — программы удержания, обучение руководителей feedback-культуре."
	default:
		return base
	}
}

func departmentTargets(dept, dim string) []string {
	targets := []string{fmt.Sprintf("Руководитель направления «%s»", dept), "HR"}
	for _, t := range strings.Split(dimensionAudiences()[dim], ", ") {
		targets = append(targets, t)
	}
	return uniqueStrings(targets)
}

func questionTargets(code, dim string) []string {
	targets := strings.Split(dimensionAudiences()[dim], ", ")
	switch code {
	case "Q11", "Q12":
		targets = append(targets, "HR / L&D")
	case "Q04", "Q05", "Q06":
		targets = append(targets, "Линейные руководители")
	case "S01", "S02":
		targets = append(targets, "Руководитель DE-практики", "Аккаунт-менеджер")
	}
	return uniqueStrings(targets)
}

func uniqueStrings(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
