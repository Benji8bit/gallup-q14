package analytics

import (
	"testing"

	"github.com/sapiens-solutions/gallup-q14/internal/models"
)

func TestResponseRate(t *testing.T) {
	assertFloat(t, responseRate(25, 100), 25)
	assertFloat(t, responseRate(10, 0), 0)
	assertFloat(t, responseRate(0, 50), 0)
}

func segmentFixture() segmentInput {
	return segmentInput{
		CurrentRoundCode: "2026-Q2",
		GroupedRecords: []models.GroupedAnswerRecord{
			{RoundCode: "2026-Q2", Direction: "DE", GradeBand: "K2", Question: "q1", QuestionRole: "engagement", Dimension: "Basic Needs", Value: 5},
			{RoundCode: "2026-Q2", Direction: "DE", GradeBand: "K2", Question: "q2", QuestionRole: "engagement", Dimension: "Basic Needs", Value: 3},
			{RoundCode: "2026-Q2", Direction: "Analytics", GradeBand: "K1", Question: "q3", QuestionRole: "engagement", Dimension: "Teamwork", Value: 4},
			{RoundCode: "2026-Q2", Direction: "DE", GradeBand: "K2", Question: "e1", QuestionRole: "enps", Dimension: "eNPS", Value: 10},
			{RoundCode: "2026-Q2", Direction: "Analytics", GradeBand: "K1", Question: "e2", QuestionRole: "enps", Dimension: "eNPS", Value: 0},
			{RoundCode: "2026-Q1", Direction: "DE", GradeBand: "K2", Question: "old", QuestionRole: "engagement", Dimension: "Basic Needs", Value: 5},
			{RoundCode: "2026-Q2", Direction: "", GradeBand: "K2", Question: "skip", QuestionRole: "engagement", Dimension: "Basic Needs", Value: 5},
		},
		ExpectedByType: map[string]map[string]int{
			"direction": {"DE": 40, "Analytics": 20},
		},
		SubmissionCountsByType: map[string]map[string]int{
			"direction": {"DE": 5, "Analytics": 2},
		},
	}
}

func TestBuildSegmentBreakdown(t *testing.T) {
	out := buildSegmentBreakdown(segmentFixture())
	if len(out) == 0 {
		t.Fatal("expected segment scores")
	}

	var deScore float64
	for _, s := range out {
		if s.SegmentType == "direction" && s.SegmentValue == "DE" {
			deScore = s.EngagementScore
			if s.ExpectedCount != 40 || s.SubmissionCount != 5 {
				t.Fatalf("DE metadata: expected=%d submissions=%d", s.ExpectedCount, s.SubmissionCount)
			}
		}
		if s.SegmentValue == "" || s.SegmentValue == "Не указан" {
			t.Fatalf("unexpected empty segment: %+v", s)
		}
	}
	assertFloat(t, deScore, 50) // one favorable of two DE answers in Q2
}

func TestBuildEnpsSegmentBreakdown(t *testing.T) {
	out := buildEnpsSegmentBreakdown(segmentFixture())
	if len(out) == 0 {
		t.Fatal("expected eNPS segment scores")
	}

	var deScore, analyticsScore float64
	for _, s := range out {
		if s.SegmentType != "direction" {
			continue
		}
		switch s.SegmentValue {
		case "DE":
			deScore = s.Score
		case "Analytics":
			analyticsScore = s.Score
		}
	}
	assertFloat(t, deScore, 100)
	assertFloat(t, analyticsScore, -100)
}
