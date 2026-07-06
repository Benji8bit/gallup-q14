package analytics

import (
	"math"
	"testing"
)

func assertFloat(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.01 {
		t.Fatalf("got %.4f want %.4f", got, want)
	}
}

func TestBuildScoreContext_satisfaction(t *testing.T) {
	ctx := BuildScoreContext("Q00", "satisfaction", 4)
	if !ctx.IsSatisfaction || ctx.IsEnps || ctx.IncludeInEngagement {
		t.Fatalf("unexpected flags: %+v", ctx)
	}
	if ctx.Score5pt != 4 {
		t.Fatalf("Score5pt = %d", ctx.Score5pt)
	}

	invalid := BuildScoreContext("Q00", "satisfaction", 6)
	if invalid.IsSatisfaction || invalid.IsEnps || invalid.IncludeInEngagement {
		t.Fatalf("expected zero context for out-of-range satisfaction, got %+v", invalid)
	}
}

func TestBuildScoreContext_enps(t *testing.T) {
	ctx := BuildScoreContext("E01", "enps", 9)
	if !ctx.IsEnps || ctx.IsSatisfaction || ctx.IncludeInEngagement {
		t.Fatalf("unexpected flags: %+v", ctx)
	}

	invalid := BuildScoreContext("E01", "enps", 11)
	if invalid.IsEnps || invalid.IsSatisfaction || invalid.IncludeInEngagement {
		t.Fatalf("expected zero context for invalid eNPS, got %+v", invalid)
	}
}

func TestBuildScoreContext_engagement(t *testing.T) {
	fav := BuildScoreContext("Q01", "engagement", 5)
	if !fav.IncludeInEngagement || !fav.Favorable || fav.Score5pt != 5 {
		t.Fatalf("favorable: %+v", fav)
	}

	neutral := BuildScoreContext("Q02", "engagement", 3)
	if !neutral.IncludeInEngagement || neutral.Favorable || neutral.Score5pt != 3 {
		t.Fatalf("neutral: %+v", neutral)
	}
}

func TestBuildScoreContext_legacyScale6(t *testing.T) {
	ctx := BuildScoreContext("Q03", "", 6)
	if !ctx.IncludeInEngagement || ctx.Score5pt != 5 || !ctx.Favorable {
		t.Fatalf("legacy 6 should map to 5: %+v", ctx)
	}
}

func TestBuildScoreContext_inferRole(t *testing.T) {
	if got := inferRole("Q00"); got != "satisfaction" {
		t.Fatalf("Q00 role = %q", got)
	}
	if got := inferRole("E01"); got != "enps" {
		t.Fatalf("E01 role = %q", got)
	}
	if got := inferRole("Q07"); got != "engagement" {
		t.Fatalf("Q07 role = %q", got)
	}
}

func TestClassifyEnps(t *testing.T) {
	cases := []struct {
		value int
		want  string
	}{
		{10, "promoter"},
		{9, "promoter"},
		{8, "passive"},
		{7, "passive"},
		{6, "detractor"},
		{0, "detractor"},
	}
	for _, tc := range cases {
		if got := ClassifyEnps(tc.value); got != tc.want {
			t.Fatalf("ClassifyEnps(%d) = %q want %q", tc.value, got, tc.want)
		}
	}
}

func TestBuildEnpsScore(t *testing.T) {
	// 3 promoters, 2 passives, 2 detractors → score = 3/7*100 - 2/7*100
	bucket := BuildEnpsScore([]int{10, 9, 10, 8, 7, 6, 0})
	if bucket.Total != 7 {
		t.Fatalf("Total = %d", bucket.Total)
	}
	if bucket.Promoters != 3 || bucket.Passives != 2 || bucket.Detractors != 2 {
		t.Fatalf("counts: promoters=%d passives=%d detractors=%d", bucket.Promoters, bucket.Passives, bucket.Detractors)
	}
	assertFloat(t, bucket.PromotersPct, 42.86)
	assertFloat(t, bucket.DetractorsPct, 28.57)
	assertFloat(t, bucket.Score, 14.29)

	empty := BuildEnpsScore(nil)
	if empty.Total != 0 || empty.Score != 0 {
		t.Fatalf("empty bucket: %+v", empty)
	}

	skipped := BuildEnpsScore([]int{-1, 11, 5})
	if skipped.Total != 1 || skipped.Detractors != 1 {
		t.Fatalf("invalid values skipped: %+v", skipped)
	}
}

func TestEnpsScoreModel(t *testing.T) {
	if EnpsScoreModel(EnpsBucket{}) != nil {
		t.Fatal("expected nil for empty bucket")
	}
	model := EnpsScoreModel(BuildEnpsScore([]int{10, 9, 10, 0}))
	if model == nil || model.Total != 4 {
		t.Fatalf("model: %+v", model)
	}
	assertFloat(t, model.Score, 50)
}

func TestEnpsTrendModel(t *testing.T) {
	point := EnpsTrendModel("2026-Q2", BuildEnpsScore([]int{9, 8}))
	if point.RoundCode != "2026-Q2" || point.Total != 2 {
		t.Fatalf("trend point: %+v", point)
	}
	assertFloat(t, point.Score, 50)
}
