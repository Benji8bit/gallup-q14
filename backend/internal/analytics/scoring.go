package analytics

import (
	"strings"

	"github.com/sapiens-solutions/gallup-q14/internal/models"
)

// ScoreContext describes how to interpret a raw survey answer.
type ScoreContext struct {
	IncludeInEngagement bool
	IsSatisfaction      bool
	IsEnps              bool
	Score5pt            int
	Favorable           bool
}

func BuildScoreContext(questionID, questionRole string, rawValue int) ScoreContext {
	role := strings.ToLower(strings.TrimSpace(questionRole))
	if role == "" {
		role = inferRole(questionID)
	}

	switch role {
	case "satisfaction":
		if rawValue < 1 || rawValue > 5 {
			return ScoreContext{}
		}
		return ScoreContext{
			IsSatisfaction: true,
			Score5pt:       rawValue,
		}
	case "enps":
		if rawValue < 0 || rawValue > 10 {
			return ScoreContext{}
		}
		return ScoreContext{IsEnps: true}
	default:
		score5 := normalizeEngagementValue(questionID, rawValue)
		if score5 == 0 {
			return ScoreContext{}
		}
		return ScoreContext{
			IncludeInEngagement: true,
			Score5pt:            score5,
			Favorable:           score5 >= 4,
		}
	}
}

func inferRole(questionID string) string {
	switch questionID {
	case "Q00":
		return "satisfaction"
	case "E01":
		return "enps"
	default:
		return "engagement"
	}
}

func normalizeEngagementValue(questionID string, rawValue int) int {
	// Legacy submissions used 1–6; new TZ uses 1–5 directly.
	if rawValue >= 1 && rawValue <= 5 {
		return rawValue
	}
	if rawValue == 6 {
		return mapSixToFive(rawValue)
	}
	return 0
}

type EnpsBucket struct {
	Promoters   int
	Passives    int
	Detractors  int
	Total       int
	PromotersPct float64
	PassivesPct  float64
	DetractorsPct float64
	Score        float64
}

func ClassifyEnps(value int) string {
	switch {
	case value >= 9:
		return "promoter"
	case value >= 7:
		return "passive"
	default:
		return "detractor"
	}
}

func BuildEnpsScore(values []int) EnpsBucket {
	out := EnpsBucket{}
	for _, v := range values {
		if v < 0 || v > 10 {
			continue
		}
		out.Total++
		switch ClassifyEnps(v) {
		case "promoter":
			out.Promoters++
		case "passive":
			out.Passives++
		default:
			out.Detractors++
		}
	}
	if out.Total == 0 {
		return out
	}
	out.PromotersPct = pct(out.Promoters, out.Total)
	out.PassivesPct = pct(out.Passives, out.Total)
	out.DetractorsPct = pct(out.Detractors, out.Total)
	out.Score = out.PromotersPct - out.DetractorsPct
	return out
}

func EnpsScoreModel(bucket EnpsBucket) *models.EnpsScore {
	if bucket.Total == 0 {
		return nil
	}
	return &models.EnpsScore{
		Score:         bucket.Score,
		Promoters:     bucket.Promoters,
		Passives:      bucket.Passives,
		Detractors:    bucket.Detractors,
		Total:         bucket.Total,
		PromotersPct:  bucket.PromotersPct,
		PassivesPct:   bucket.PassivesPct,
		DetractorsPct: bucket.DetractorsPct,
	}
}

func EnpsTrendModel(roundCode string, bucket EnpsBucket) models.EnpsTrendPoint {
	return models.EnpsTrendPoint{
		RoundCode:     roundCode,
		Score:         bucket.Score,
		Total:         bucket.Total,
		PromotersPct:  bucket.PromotersPct,
		PassivesPct:   bucket.PassivesPct,
		DetractorsPct: bucket.DetractorsPct,
	}
}
