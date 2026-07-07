package analytics

import (
	"testing"

	"github.com/sapiens-solutions/gallup-q14/internal/models"
)

func TestPct(t *testing.T) {
	assertFloat(t, pct(3, 10), 30)
	assertFloat(t, pct(0, 0), 0)
}

func TestAvg(t *testing.T) {
	assertFloat(t, avg(10, 4), 2.5)
	assertFloat(t, avg(0, 0), 0)
}

func TestMapSixToFive(t *testing.T) {
	cases := map[int]int{
		1: 1, 2: 1, 3: 2, 4: 3, 5: 4, 6: 5,
	}
	for in, want := range cases {
		if got := mapSixToFive(in); got != want {
			t.Fatalf("mapSixToFive(%d) = %d want %d", in, got, want)
		}
	}
	if got := mapSixToFive(0); got != 0 {
		t.Fatalf("mapSixToFive(0) = %d", got)
	}
}

func TestBuildDashboard_empty(t *testing.T) {
	out := New().BuildDashboard(BuildInput{
		CurrentRoundCode:        "2026-Q2",
		CurrentRoundSubmissions: 0,
	})
	if out.EngagementScore != 0 || out.SatisfactionScore != 0 {
		t.Fatalf("scores: engagement=%v satisfaction=%v", out.EngagementScore, out.SatisfactionScore)
	}
	if len(out.DimensionBreakdown) != 4 {
		t.Fatalf("default dimensions = %d", len(out.DimensionBreakdown))
	}
	if len(out.RecommendationsRU) == 0 {
		t.Fatal("expected placeholder recommendation")
	}
}

func TestBuildDashboard_engagementAndEnps(t *testing.T) {
	records := []models.AnswerRecord{
		{RoundCode: "2026-Q2", Question: "q1", QuestionCode: "Q01", Dimension: "Basic Needs", QuestionRole: "engagement", Value: 5},
		{RoundCode: "2026-Q2", Question: "q2", QuestionCode: "Q02", Dimension: "Basic Needs", QuestionRole: "engagement", Value: 4},
		{RoundCode: "2026-Q2", Question: "q3", QuestionCode: "Q03", Dimension: "Basic Needs", QuestionRole: "engagement", Value: 2},
		{RoundCode: "2026-Q2", Question: "q0", QuestionCode: "Q00", Dimension: "Satisfaction", QuestionRole: "satisfaction", Value: 4},
		{RoundCode: "2026-Q2", Question: "e1", QuestionCode: "E01", Dimension: "eNPS", QuestionRole: "enps", Value: 10},
		{RoundCode: "2026-Q2", Question: "e1b", QuestionCode: "E01", Dimension: "eNPS", QuestionRole: "enps", Value: 9},
		{RoundCode: "2026-Q1", Question: "q1old", QuestionCode: "Q01", Dimension: "Basic Needs", QuestionRole: "engagement", Value: 3},
		{RoundCode: "2026-Q1", Question: "e1old", QuestionCode: "E01", Dimension: "eNPS", QuestionRole: "enps", Value: 8},
	}

	out := New().BuildDashboard(BuildInput{
		Records:                 records,
		CurrentRoundCode:        "2026-Q2",
		CurrentRoundSubmissions: 2,
		SubmissionCountsByRound: map[string]int{"2026-Q2": 2, "2026-Q1": 1},
		DeliveryMeta:            &models.DeliverySyncMeta{StaffTotal: 100, ActiveDeliveryQTD: 50},
	})

	assertFloat(t, out.EngagementScore, 50) // 2 favorable of 4 engagement answers (Q2 + Q1)
	assertFloat(t, out.SatisfactionScore, 4)
	if out.EnpsScore == nil {
		t.Fatal("expected eNPS score")
	}
	assertFloat(t, out.EnpsScore.Score, 100) // two promoters in current quarter
	if len(out.Trends) != 2 {
		t.Fatalf("trends = %d", len(out.Trends))
	}
	assertFloat(t, out.ResponseRatePct, 2) // 2/100 * 100
	if len(out.QuestionScores) != 4 {
		t.Fatalf("question scores = %d", len(out.QuestionScores))
	}
}
