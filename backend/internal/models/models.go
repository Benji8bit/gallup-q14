package models

import "time"

type SurveyRound struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Year      int       `json:"year"`
	Quarter   int       `json:"quarter"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	IsActive  bool      `json:"isActive"`
}

type Question struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	TextRU       string `json:"textRu"`
	Dimension    string `json:"dimension"`
	SortOrder    int    `json:"sortOrder"`
	ScaleMin     int    `json:"scaleMin"`
	ScaleMax     int    `json:"scaleMax"`
	QuestionRole string `json:"questionRole"`
}

type SurveyCurrentResponse struct {
	Round           SurveyRound      `json:"round"`
	Questions       []Question       `json:"questions"`
	OrgOptions      []OrgOptionGroup `json:"orgOptions,omitempty"`
	OrgOptionScopes []OrgOptionScope `json:"orgOptionScopes,omitempty"`
}

type OrgOptionScope struct {
	OptionType  string `json:"optionType"`
	OptionValue string `json:"optionValue"`
	ScopeType   string `json:"scopeType"`
	ScopeValue  string `json:"scopeValue"`
}

type OrgOptionGroup struct {
	Type    string     `json:"type"`
	LabelRU string     `json:"labelRu"`
	Options []OrgOption `json:"options"`
}

type OrgOption struct {
	Value         string `json:"value"`
	LabelRU       string `json:"labelRu"`
	EmployeeCount int    `json:"employeeCount,omitempty"`
}

type SubmitSurveyRequest struct {
	AnonymousToken string         `json:"anonymousToken"`
	Answers        map[string]int `json:"answers"`
	Role           string         `json:"role,omitempty"`
	Department     string         `json:"department,omitempty"`
	Tenure         string         `json:"tenure,omitempty"`
	Direction      string         `json:"direction,omitempty"`
	PositionGroup  string         `json:"positionGroup,omitempty"`
	GradeBand      string         `json:"gradeBand,omitempty"`
	EmployeeType   string         `json:"employeeType,omitempty"`
}

type SurveySubmission struct {
	ID           int64
	RoundID      int64
	AnonymousRef string
	SubmittedAt  time.Time
}

type AnswerRecord struct {
	RoundCode    string
	Question     string
	QuestionCode string
	Dimension    string
	QuestionRole string
	Value        int
}

type ExportRecord struct {
	RoundCode   string
	SubmittedAt time.Time
	AnonToken   string
	QuestionID  string
	QuestionRU  string
	Dimension   string
	Value6      int
	Value5      int
	Favorable   bool
}

type DimensionScore struct {
	Dimension       string  `json:"dimension"`
	FavorablePct    float64 `json:"favorablePct"`
	ResponseCount   int     `json:"responseCount"`
	AverageScore5pt float64 `json:"averageScore5pt"`
}

type TrendPoint struct {
	RoundCode       string  `json:"roundCode"`
	EngagementPct   float64 `json:"engagementPct"`
	SubmissionCount int     `json:"submissionCount"`
}

type QuestionScore struct {
	QuestionID      string  `json:"questionId"`
	Code            string  `json:"code"`
	Dimension       string  `json:"dimension"`
	FavorablePct    float64 `json:"favorablePct"`
	AverageScore5pt float64 `json:"averageScore5pt"`
	ResponseCount   int     `json:"responseCount"`
}

type EnpsScore struct {
	Score         float64 `json:"score"`
	Promoters     int     `json:"promoters"`
	Passives      int     `json:"passives"`
	Detractors    int     `json:"detractors"`
	Total         int     `json:"total"`
	PromotersPct  float64 `json:"promotersPct"`
	PassivesPct   float64 `json:"passivesPct"`
	DetractorsPct float64 `json:"detractorsPct"`
}

type EnpsTrendPoint struct {
	RoundCode     string  `json:"roundCode"`
	Score         float64 `json:"score"`
	Total         int     `json:"total"`
	PromotersPct  float64 `json:"promotersPct"`
	PassivesPct   float64 `json:"passivesPct"`
	DetractorsPct float64 `json:"detractorsPct"`
}

type EnpsSegmentScore struct {
	SegmentType     string  `json:"segmentType"`
	SegmentValue    string  `json:"segmentValue"`
	Score           float64 `json:"score"`
	PromotersPct    float64 `json:"promotersPct"`
	PassivesPct     float64 `json:"passivesPct"`
	DetractorsPct   float64 `json:"detractorsPct"`
	Total           int     `json:"total"`
	SubmissionCount int     `json:"submissionCount"`
	ExpectedCount   int     `json:"expectedCount"`
}

type DashboardResponse struct {
	GeneratedAt             time.Time              `json:"generatedAt"`
	CurrentRoundCode        string                 `json:"currentRoundCode"`
	EngagementScore         float64                `json:"engagementScore"`
	SatisfactionScore       float64                `json:"satisfactionScore"`
	EnpsScore               *EnpsScore             `json:"enpsScore,omitempty"`
	EnpsTrends              []EnpsTrendPoint       `json:"enpsTrends,omitempty"`
	EnpsSegmentBreakdown    []EnpsSegmentScore     `json:"enpsSegmentBreakdown,omitempty"`
	CurrentRoundSubmissions int                    `json:"currentRoundSubmissions"`
	DimensionBreakdown      []DimensionScore       `json:"dimensionBreakdown"`
	QuestionScores          []QuestionScore        `json:"questionScores"`
	Trends                  []TrendPoint           `json:"trends"`
	RecommendationsRU       []string               `json:"recommendationsRu"`
	MethodologyGuide        MethodologyGuide       `json:"methodologyGuide"`
	Recommendations         []RecommendationItem   `json:"recommendations"`
	DepartmentBreakdown     []DepartmentScore      `json:"departmentBreakdown,omitempty"`
	SegmentBreakdown        []SegmentScore         `json:"segmentBreakdown,omitempty"`
	DeliveryContext         []DeliveryContextStat  `json:"deliveryContext,omitempty"`
	ExpectedRespondents     int                    `json:"expectedRespondents"`
	ResponseRatePct         float64                `json:"responseRatePct"`
	DeliverySyncedAt        *time.Time             `json:"deliverySyncedAt,omitempty"`
}

type SegmentScore struct {
	SegmentType     string  `json:"segmentType"`
	SegmentValue    string  `json:"segmentValue"`
	EngagementScore float64 `json:"engagementScore"`
	SubmissionCount int     `json:"submissionCount"`
	ExpectedCount   int     `json:"expectedCount"`
}

type DeliveryContextStat struct {
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	Metric      string  `json:"metric"`
	NumericValue float64 `json:"numericValue"`
}

type DeliverySyncMeta struct {
	SyncedAt          time.Time
	QuarterCode       string
	StaffTotal        int
	ActiveDeliveryQTD int
}

type MethodologyGuide struct {
	Overview string              `json:"overview"`
	Metrics  []MetricGuideItem   `json:"metrics"`
	Scales   []ScaleGuideSection `json:"scales"`
	Questions []QuestionGuideItem `json:"questions"`
	Dimensions []DimensionGuideItem `json:"dimensions"`
}

type MetricGuideItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Formula     string `json:"formula"`
	Scale       string `json:"scale"`
}

type ScaleGuideSection struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Mapping     []ScaleMappingRow `json:"mapping,omitempty"`
}

type ScaleMappingRow struct {
	SurveyValue string `json:"surveyValue"`
	DashboardValue string `json:"dashboardValue"`
	Meaning     string `json:"meaning"`
}

type QuestionGuideItem struct {
	Code            string  `json:"code"`
	TextRU          string  `json:"textRu"`
	Dimension       string  `json:"dimension"`
	WhatItMeasures  string  `json:"whatItMeasures"`
	LeadershipFocus string  `json:"leadershipFocus"`
	CurrentAvg5pt   float64 `json:"currentAvg5pt,omitempty"`
	CurrentFavPct   float64 `json:"currentFavPct,omitempty"`
}

type DimensionGuideItem struct {
	Key             string `json:"key"`
	LabelRU         string `json:"labelRu"`
	Description     string `json:"description"`
	PrimaryAudience string `json:"primaryAudience"`
}

type RecommendationItem struct {
	ID               string                 `json:"id"`
	Scope            string                 `json:"scope"`
	Title            string                 `json:"title"`
	Action           string                 `json:"action"`
	TargetGroups     []string               `json:"targetGroups"`
	Basis            string                 `json:"basis"`
	RelatedDimension string                 `json:"relatedDimension,omitempty"`
	RelatedQuestions []string               `json:"relatedQuestions,omitempty"`
	Evidence         RecommendationEvidence `json:"evidence"`
}

type RecommendationEvidence struct {
	FavorablePct      *float64 `json:"favorablePct,omitempty"`
	AverageScore5pt   *float64 `json:"averageScore5pt,omitempty"`
	EngagementScore   *float64 `json:"engagementScore,omitempty"`
	SatisfactionScore *float64 `json:"satisfactionScore,omitempty"`
	Department        string   `json:"department,omitempty"`
	Tenure            string   `json:"tenure,omitempty"`
	QuestionCode      string   `json:"questionCode,omitempty"`
	ResponseCount     int      `json:"responseCount,omitempty"`
}

type DepartmentScore struct {
	Department      string  `json:"department"`
	Dimension       string  `json:"dimension"`
	FavorablePct    float64 `json:"favorablePct"`
	AverageScore5pt float64 `json:"averageScore5pt"`
	ResponseCount   int     `json:"responseCount"`
	SubmissionCount int     `json:"submissionCount"`
}

type GroupedAnswerRecord struct {
	RoundCode     string
	Department    string
	Direction     string
	PositionGroup string
	GradeBand     string
	EmployeeType  string
	Tenure        string
	Question      string
	QuestionCode  string
	Dimension     string
	QuestionRole  string
	Value         int
}
