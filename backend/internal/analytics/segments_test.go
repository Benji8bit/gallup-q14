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
			{RoundCode: "2026-Q2", GradeBand: "DE3", PositionGroup: "Тимлид", Question: "q1", QuestionRole: "engagement", Dimension: "Basic Needs", Value: 5},
			{RoundCode: "2026-Q2", GradeBand: "DE3", PositionGroup: "Тимлид", Question: "q2", QuestionRole: "engagement", Dimension: "Basic Needs", Value: 3},
			{RoundCode: "2026-Q2", GradeBand: "DE2", PositionGroup: "Инженер данных/Разработчик", Question: "q3", QuestionRole: "engagement", Dimension: "Teamwork", Value: 4},
			{RoundCode: "2026-Q2", GradeBand: "DE3", PositionGroup: "Тимлид", Question: "e1", QuestionRole: "enps", Dimension: "eNPS", Value: 10},
			{RoundCode: "2026-Q2", GradeBand: "DE2", PositionGroup: "Инженер данных/Разработчик", Question: "e2", QuestionRole: "enps", Dimension: "eNPS", Value: 0},
			{RoundCode: "2026-Q1", GradeBand: "DE3", PositionGroup: "Тимлид", Question: "old", QuestionRole: "engagement", Dimension: "Basic Needs", Value: 5},
			{RoundCode: "2026-Q2", GradeBand: "", PositionGroup: "Тимлид", Question: "skip", QuestionRole: "engagement", Dimension: "Basic Needs", Value: 5},
		},
		ExpectedByType: map[string]map[string]int{
			"grade_band": {"DE3": 40, "DE2": 20},
			"role":       {"Тимлид": 15},
		},
		SubmissionCountsByType: map[string]map[string]int{
			"grade_band": {"DE3": 5, "DE2": 2},
			"role":       {"Тимлид": 3},
		},
	}
}

func TestBuildSegmentBreakdown(t *testing.T) {
	out := buildSegmentBreakdown(segmentFixture())
	if len(out) == 0 {
		t.Fatal("expected segment scores")
	}

	var de3Score float64
	for _, s := range out {
		if s.SegmentType == "grade_band" && s.SegmentValue == "DE3" {
			de3Score = s.EngagementScore
			if s.ExpectedCount != 40 || s.SubmissionCount != 5 {
				t.Fatalf("DE3 metadata: expected=%d submissions=%d", s.ExpectedCount, s.SubmissionCount)
			}
		}
		if s.SegmentValue == "" || s.SegmentValue == "Не указан" {
			t.Fatalf("unexpected empty segment: %+v", s)
		}
	}
	assertFloat(t, de3Score, 50)
}

func TestBuildEnpsSegmentBreakdown(t *testing.T) {
	out := buildEnpsSegmentBreakdown(segmentFixture())
	if len(out) == 0 {
		t.Fatal("expected eNPS segment scores")
	}

	var de3Score, de2Score float64
	for _, s := range out {
		if s.SegmentType != "grade_band" {
			continue
		}
		switch s.SegmentValue {
		case "DE3":
			de3Score = s.Score
		case "DE2":
			de2Score = s.Score
		}
	}
	assertFloat(t, de3Score, 100)
	assertFloat(t, de2Score, -100)
}
