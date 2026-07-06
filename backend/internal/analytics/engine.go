package analytics

import (
	"sort"
	"time"

	"github.com/sapiens-solutions/gallup-q14/internal/models"
)

type Engine struct{}

func New() *Engine {
	return &Engine{}
}

type BuildInput struct {
	Records                 []models.AnswerRecord
	GroupedRecords          []models.GroupedAnswerRecord
	QuestionTexts           map[string]string
	DeptSubmissionCounts    map[string]int
	SubmissionCountsByRound map[string]int
	CurrentRoundCode        string
	CurrentRoundSubmissions int
	DeliveryMeta            *models.DeliverySyncMeta
	DeliveryContext         []models.DeliveryContextStat
	ExpectedBySegment       map[string]map[string]int
	SubmissionBySegment     map[string]map[string]int
}

func (e *Engine) BuildDashboard(input BuildInput) models.DashboardResponse {
	now := time.Now().UTC()
	if len(input.Records) == 0 {
		return models.DashboardResponse{
			GeneratedAt:             now,
			CurrentRoundCode:        input.CurrentRoundCode,
			EngagementScore:         0,
			SatisfactionScore:       0,
			CurrentRoundSubmissions: input.CurrentRoundSubmissions,
			DimensionBreakdown:      defaultDimensions(),
			QuestionScores:          []models.QuestionScore{},
			Trends:                  []models.TrendPoint{},
			RecommendationsRU:       []string{"Недостаточно данных для рекомендаций. Запустите сбор ответов в текущем квартале."},
			MethodologyGuide:        BuildMethodologyGuide(input.QuestionTexts, nil),
			Recommendations:         []models.RecommendationItem{},
			DeliveryContext:         input.DeliveryContext,
		}
	}

	var favorableAll int
	var totalAll int
	var satisfactionSum int
	var satisfactionCount int
	var enpsValues []int
	byRoundEnps := map[string][]int{}

	type agg struct {
		fav int
		sum int
		cnt int
	}
	byDimension := map[string]*agg{}
	byRound := map[string]*agg{}
	byQuestion := map[string]*agg{}
	questionMeta := map[string]struct {
		code      string
		dimension string
	}{}

	for _, rec := range input.Records {
		ctx := BuildScoreContext(rec.Question, rec.QuestionRole, rec.Value)

		if ctx.IsSatisfaction {
			satisfactionSum += rec.Value
			satisfactionCount++
			continue
		}

		if ctx.IsEnps {
			byRoundEnps[rec.RoundCode] = append(byRoundEnps[rec.RoundCode], rec.Value)
			if rec.RoundCode == input.CurrentRoundCode {
				enpsValues = append(enpsValues, rec.Value)
			}
			continue
		}

		if !ctx.IncludeInEngagement {
			continue
		}

		score5 := ctx.Score5pt
		totalAll++
		if ctx.Favorable {
			favorableAll++
		}

		dim := byDimension[rec.Dimension]
		if dim == nil {
			dim = &agg{}
			byDimension[rec.Dimension] = dim
		}
		dim.cnt++
		dim.sum += score5
		if ctx.Favorable {
			dim.fav++
		}

		round := byRound[rec.RoundCode]
		if round == nil {
			round = &agg{}
			byRound[rec.RoundCode] = round
		}
		round.cnt++
		if ctx.Favorable {
			round.fav++
		}

		q := byQuestion[rec.Question]
		if q == nil {
			q = &agg{}
			byQuestion[rec.Question] = q
		}
		q.cnt++
		q.sum += score5
		if ctx.Favorable {
			q.fav++
		}
		questionMeta[rec.Question] = struct {
			code      string
			dimension string
		}{code: rec.QuestionCode, dimension: rec.Dimension}
	}

	dimensions := make([]models.DimensionScore, 0, len(byDimension))
	for dimName, dimAgg := range byDimension {
		if dimName == "eNPS" {
			continue
		}
		dimensions = append(dimensions, models.DimensionScore{
			Dimension:       dimName,
			FavorablePct:    pct(dimAgg.fav, dimAgg.cnt),
			ResponseCount:   dimAgg.cnt,
			AverageScore5pt: avg(dimAgg.sum, dimAgg.cnt),
		})
	}
	sort.Slice(dimensions, func(i, j int) bool { return dimensions[i].FavorablePct < dimensions[j].FavorablePct })

	questionScores := make([]models.QuestionScore, 0, len(byQuestion))
	for questionID, qAgg := range byQuestion {
		meta := questionMeta[questionID]
		code := meta.code
		if code == "" {
			code = questionID
		}
		questionScores = append(questionScores, models.QuestionScore{
			QuestionID:      questionID,
			Code:            code,
			Dimension:       meta.dimension,
			FavorablePct:    pct(qAgg.fav, qAgg.cnt),
			AverageScore5pt: avg(qAgg.sum, qAgg.cnt),
			ResponseCount:   qAgg.cnt,
		})
	}
	sort.Slice(questionScores, func(i, j int) bool { return questionScores[i].Code < questionScores[j].Code })

	roundCodes := make([]string, 0, len(byRound))
	for roundCode := range byRound {
		roundCodes = append(roundCodes, roundCode)
	}
	sort.Strings(roundCodes)

	trends := make([]models.TrendPoint, 0, len(roundCodes))
	for _, roundCode := range roundCodes {
		roundAgg := byRound[roundCode]
		submissionCount := input.SubmissionCountsByRound[roundCode]
		if submissionCount == 0 {
			submissionCount = roundAgg.cnt
		}
		trends = append(trends, models.TrendPoint{
			RoundCode:       roundCode,
			EngagementPct:   pct(roundAgg.fav, roundAgg.cnt),
			SubmissionCount: submissionCount,
		})
	}

	satisfaction := 0.0
	if satisfactionCount > 0 {
		satisfaction = float64(satisfactionSum) / float64(satisfactionCount)
	}

	enps := EnpsScoreModel(BuildEnpsScore(enpsValues))

	enpsRoundCodes := make([]string, 0, len(byRoundEnps))
	for roundCode := range byRoundEnps {
		enpsRoundCodes = append(enpsRoundCodes, roundCode)
	}
	sort.Strings(enpsRoundCodes)
	enpsTrends := make([]models.EnpsTrendPoint, 0, len(enpsRoundCodes))
	for _, roundCode := range enpsRoundCodes {
		bucket := BuildEnpsScore(byRoundEnps[roundCode])
		if bucket.Total == 0 {
			continue
		}
		enpsTrends = append(enpsTrends, EnpsTrendModel(roundCode, bucket))
	}

	deptBreakdown := buildDepartmentBreakdown(input.GroupedRecords, input.CurrentRoundCode, input.DeptSubmissionCounts)
	segments := buildSegmentBreakdown(segmentInput{
		GroupedRecords:         input.GroupedRecords,
		CurrentRoundCode:       input.CurrentRoundCode,
		ExpectedByType:         input.ExpectedBySegment,
		SubmissionCountsByType: input.SubmissionBySegment,
	})
	enpsSegments := buildEnpsSegmentBreakdown(segmentInput{
		GroupedRecords:         input.GroupedRecords,
		CurrentRoundCode:       input.CurrentRoundCode,
		ExpectedByType:         input.ExpectedBySegment,
		SubmissionCountsByType: input.SubmissionBySegment,
	})

	expected := 0
	if input.DeliveryMeta != nil {
		expected = input.DeliveryMeta.StaffTotal
	}
	respRate := responseRate(input.CurrentRoundSubmissions, expected)

	var syncedAt *time.Time
	if input.DeliveryMeta != nil && !input.DeliveryMeta.SyncedAt.IsZero() {
		t := input.DeliveryMeta.SyncedAt
		syncedAt = &t
	}

	engagementPct := pct(favorableAll, totalAll)
	recs := buildRecommendations(recommendationInput{
		EngagementScore:   engagementPct,
		SatisfactionScore: satisfaction,
		EnpsScore:         enps,
		Dimensions:        dimensions,
		QuestionScores:    questionScores,
		DepartmentScores:  deptBreakdown,
		CurrentRoundCode:  input.CurrentRoundCode,
	})

	return models.DashboardResponse{
		GeneratedAt:             now,
		CurrentRoundCode:        input.CurrentRoundCode,
		EngagementScore:         engagementPct,
		SatisfactionScore:       satisfaction,
		EnpsScore:               enps,
		EnpsTrends:              enpsTrends,
		EnpsSegmentBreakdown:    enpsSegments,
		CurrentRoundSubmissions: input.CurrentRoundSubmissions,
		DimensionBreakdown:      dimensions,
		QuestionScores:          questionScores,
		Trends:                  trends,
		RecommendationsRU:       recommendationsSummary(recs),
		MethodologyGuide:        BuildMethodologyGuide(input.QuestionTexts, questionScores),
		Recommendations:         recs,
		DepartmentBreakdown:     deptBreakdown,
		SegmentBreakdown:        segments,
		DeliveryContext:         input.DeliveryContext,
		ExpectedRespondents:     expected,
		ResponseRatePct:         respRate,
		DeliverySyncedAt:        syncedAt,
	}
}

func mapSixToFive(value int) int {
	switch value {
	case 1, 2:
		return 1
	case 3:
		return 2
	case 4:
		return 3
	case 5:
		return 4
	case 6:
		return 5
	default:
		return 0
	}
}

func pct(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) * 100 / float64(total)
}

func avg(sum, count int) float64 {
	if count == 0 {
		return 0
	}
	return float64(sum) / float64(count)
}

func defaultDimensions() []models.DimensionScore {
	dims := []string{
		"Basic Needs",
		"Individual Contribution",
		"Teamwork",
		"Growth",
	}
	out := make([]models.DimensionScore, 0, len(dims))
	for _, dim := range dims {
		out = append(out, models.DimensionScore{
			Dimension:       dim,
			FavorablePct:    0,
			ResponseCount:   0,
			AverageScore5pt: 0,
		})
	}
	return out
}
