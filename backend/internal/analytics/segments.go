package analytics

import (
	"sort"

	"github.com/sapiens-solutions/gallup-q14/internal/models"
)

type segmentInput struct {
	GroupedRecords         []models.GroupedAnswerRecord
	CurrentRoundCode       string
	ExpectedByType         map[string]map[string]int
	SubmissionCountsByType map[string]map[string]int
}

func buildSegmentBreakdown(input segmentInput) []models.SegmentScore {
	segmentTypes := []struct {
		Type  string
		Value func(models.GroupedAnswerRecord) string
	}{
		{"grade_band", func(r models.GroupedAnswerRecord) string { return r.GradeBand }},
		{"role", func(r models.GroupedAnswerRecord) string { return r.PositionGroup }},
	}

	out := make([]models.SegmentScore, 0, 24)
	for _, seg := range segmentTypes {
		type agg struct {
			fav, total int
		}
		byValue := map[string]*agg{}
		for _, rec := range input.GroupedRecords {
			if rec.RoundCode != input.CurrentRoundCode {
				continue
			}
			ctx := BuildScoreContext(rec.Question, rec.QuestionRole, rec.Value)
			if !ctx.IncludeInEngagement {
				continue
			}
			value := seg.Value(rec)
			if value == "" || value == "Не указан" {
				continue
			}
			a := byValue[value]
			if a == nil {
				a = &agg{}
				byValue[value] = a
			}
			a.total++
			if ctx.Favorable {
				a.fav++
			}
		}

		expected := input.ExpectedByType[seg.Type]
		submissions := input.SubmissionCountsByType[seg.Type]
		for value, a := range byValue {
			out = append(out, models.SegmentScore{
				SegmentType:     seg.Type,
				SegmentValue:    value,
				EngagementScore: pct(a.fav, a.total),
				SubmissionCount: submissions[value],
				ExpectedCount:   expected[value],
			})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].SegmentType == out[j].SegmentType {
			return out[i].EngagementScore < out[j].EngagementScore
		}
		return out[i].SegmentType < out[j].SegmentType
	})
	return out
}

func buildEnpsSegmentBreakdown(input segmentInput) []models.EnpsSegmentScore {
	segmentTypes := []struct {
		Type  string
		Value func(models.GroupedAnswerRecord) string
	}{
		{"grade_band", func(r models.GroupedAnswerRecord) string { return r.GradeBand }},
		{"role", func(r models.GroupedAnswerRecord) string { return r.PositionGroup }},
	}

	out := make([]models.EnpsSegmentScore, 0, 24)
	for _, seg := range segmentTypes {
		byValue := map[string][]int{}
		for _, rec := range input.GroupedRecords {
			if rec.RoundCode != input.CurrentRoundCode {
				continue
			}
			ctx := BuildScoreContext(rec.Question, rec.QuestionRole, rec.Value)
			if !ctx.IsEnps {
				continue
			}
			value := seg.Value(rec)
			if value == "" || value == "Не указан" {
				continue
			}
			byValue[value] = append(byValue[value], rec.Value)
		}

		expected := input.ExpectedByType[seg.Type]
		submissions := input.SubmissionCountsByType[seg.Type]
		for value, values := range byValue {
			bucket := BuildEnpsScore(values)
			if bucket.Total == 0 {
				continue
			}
			out = append(out, models.EnpsSegmentScore{
				SegmentType:     seg.Type,
				SegmentValue:    value,
				Score:           bucket.Score,
				PromotersPct:    bucket.PromotersPct,
				PassivesPct:     bucket.PassivesPct,
				DetractorsPct:   bucket.DetractorsPct,
				Total:           bucket.Total,
				SubmissionCount: submissions[value],
				ExpectedCount:   expected[value],
			})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].SegmentType == out[j].SegmentType {
			return out[i].Score < out[j].Score
		}
		return out[i].SegmentType < out[j].SegmentType
	})
	return out
}

func responseRate(submissions, expected int) float64 {
	if expected <= 0 {
		return 0
	}
	return float64(submissions) / float64(expected) * 100
}
