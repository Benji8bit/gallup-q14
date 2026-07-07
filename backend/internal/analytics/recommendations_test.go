package analytics

import (
	"strings"
	"testing"

	"github.com/sapiens-solutions/gallup-q14/internal/models"
)

func TestEngagementBand(t *testing.T) {
	if got := engagementBand(75); got != "сильная зона" {
		t.Fatalf("75%% = %q", got)
	}
	if got := engagementBand(55); got != "зона внимания" {
		t.Fatalf("55%% = %q", got)
	}
	if got := engagementBand(30); got != "критическая зона" {
		t.Fatalf("30%% = %q", got)
	}
}

func TestBuildRecommendations_empty(t *testing.T) {
	items := buildRecommendations(recommendationInput{})
	if len(items) != 1 || items[0].ID != "empty" {
		t.Fatalf("items: %+v", items)
	}
}

func TestBuildRecommendations_withData(t *testing.T) {
	items := buildRecommendations(recommendationInput{
		EngagementScore:  45,
		CurrentRoundCode: "2026-Q2",
		EnpsScore: &models.EnpsScore{
			Score: 10, Total: 5,
			PromotersPct: 40, PassivesPct: 30, DetractorsPct: 30,
		},
		Dimensions: []models.DimensionScore{
			{Dimension: "Basic Needs", FavorablePct: 40, AverageScore5pt: 3.1, ResponseCount: 10},
			{Dimension: "Teamwork", FavorablePct: 80, AverageScore5pt: 4.2, ResponseCount: 10},
		},
		QuestionScores: []models.QuestionScore{
			{Code: "Q01", Dimension: "Basic Needs", FavorablePct: 35, AverageScore5pt: 2.9, ResponseCount: 10},
			{Code: "Q07", Dimension: "Teamwork", FavorablePct: 75, AverageScore5pt: 3.8, ResponseCount: 10},
		},
	})

	if len(items) < 4 {
		t.Fatalf("expected general + dimension recs, got %d", len(items))
	}
	if items[0].ID != "general-engagement" {
		t.Fatalf("first item = %s", items[0].ID)
	}

	var hasEnps, hasWeakDim bool
	for _, item := range items {
		switch item.ID {
		case "general-enps":
			hasEnps = true
		}
		if strings.HasPrefix(item.ID, "general-dim-") && strings.Contains(item.Title, "Роль и ресурсы") {
			hasWeakDim = true
		}
	}
	if !hasEnps {
		t.Fatalf("missing general eNPS recommendation")
	}
	if !hasWeakDim {
		t.Fatal("expected recommendation for weak Basic Needs dimension")
	}
}

func TestRecommendationsSummary(t *testing.T) {
	out := recommendationsSummary([]models.RecommendationItem{
		{Title: "A", Action: "do A"},
		{Title: "B", Action: "do B"},
	})
	if len(out) != 2 || !strings.Contains(out[0], "A:") {
		t.Fatalf("summary: %v", out)
	}
}
