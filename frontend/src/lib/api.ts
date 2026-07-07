export interface SurveyRound {
  id: number;
  code: string;
  year: number;
  quarter: number;
  startDate: string;
  endDate: string;
  isActive: boolean;
}

export interface Question {
  id: string;
  code: string;
  textRu: string;
  dimension: string;
  sortOrder: number;
  scaleMin: number;
  scaleMax: number;
  questionRole: 'satisfaction' | 'engagement' | 'enps' | string;
}

export interface OrgOption {
  value: string;
  labelRu: string;
  employeeCount?: number;
}

export interface OrgOptionGroup {
  type: string;
  labelRu: string;
  options: OrgOption[];
}

export interface OrgOptionScope {
  optionType: string;
  optionValue: string;
  scopeType: string;
  scopeValue: string;
}

export interface SurveyCurrentResponse {
  round: SurveyRound;
  questions: Question[];
  orgOptions?: OrgOptionGroup[];
  orgOptionScopes?: OrgOptionScope[];
}

export interface DimensionScore {
  dimension: string;
  favorablePct: number;
  responseCount: number;
  averageScore5pt: number;
}

export interface QuestionScore {
  questionId: string;
  code: string;
  dimension: string;
  favorablePct: number;
  averageScore5pt: number;
  responseCount: number;
}

export interface TrendPoint {
  roundCode: string;
  engagementPct: number;
  submissionCount: number;
}

export interface MetricGuideItem {
  name: string;
  description: string;
  formula: string;
  scale: string;
}

export interface ScaleMappingRow {
  surveyValue: string;
  dashboardValue: string;
  meaning: string;
}

export interface ScaleGuideSection {
  title: string;
  description: string;
  mapping?: ScaleMappingRow[];
}

export interface QuestionGuideItem {
  code: string;
  textRu: string;
  dimension?: string;
  whatItMeasures: string;
  leadershipFocus: string;
  currentAvg5pt?: number;
  currentFavPct?: number;
}

export interface DimensionGuideItem {
  key: string;
  labelRu: string;
  description: string;
  primaryAudience: string;
}

export interface MethodologyGuide {
  overview: string;
  metrics: MetricGuideItem[];
  scales: ScaleGuideSection[];
  questions: QuestionGuideItem[];
  dimensions: DimensionGuideItem[];
}

export interface RecommendationEvidence {
  favorablePct?: number;
  averageScore5pt?: number;
  engagementScore?: number;
  satisfactionScore?: number;
  department?: string;
  tenure?: string;
  questionCode?: string;
  responseCount?: number;
}

export interface RecommendationItem {
  id: string;
  scope: 'general' | 'group' | string;
  title: string;
  action: string;
  targetGroups: string[];
  basis: string;
  relatedDimension?: string;
  relatedQuestions?: string[];
  evidence: RecommendationEvidence;
}

export interface DepartmentScore {
  department: string;
  dimension: string;
  favorablePct: number;
  averageScore5pt: number;
  responseCount: number;
  submissionCount: number;
}

export interface SegmentScore {
  segmentType: string;
  segmentValue: string;
  engagementScore: number;
  submissionCount: number;
  expectedCount: number;
}

export interface DeliveryContextStat {
  key: string;
  value: string;
  metric: string;
  numericValue: number;
}

export interface EnpsScore {
  score: number;
  promoters: number;
  passives: number;
  detractors: number;
  total: number;
  promotersPct: number;
  passivesPct: number;
  detractorsPct: number;
}

export interface EnpsTrendPoint {
  roundCode: string;
  score: number;
  total: number;
  promotersPct: number;
  passivesPct: number;
  detractorsPct: number;
}

export interface EnpsSegmentScore {
  segmentType: string;
  segmentValue: string;
  score: number;
  promotersPct: number;
  passivesPct: number;
  detractorsPct: number;
  total: number;
  submissionCount: number;
  expectedCount: number;
}

export interface DashboardResponse {
  generatedAt: string;
  currentRoundCode: string;
  engagementScore: number;
  satisfactionScore: number;
  enpsScore?: EnpsScore;
  enpsTrends?: EnpsTrendPoint[];
  enpsSegmentBreakdown?: EnpsSegmentScore[];
  currentRoundSubmissions: number;
  dimensionBreakdown: DimensionScore[];
  questionScores: QuestionScore[];
  trends: TrendPoint[];
  recommendationsRu: string[];
  methodologyGuide: MethodologyGuide;
  recommendations: RecommendationItem[];
  departmentBreakdown?: DepartmentScore[];
  segmentBreakdown?: SegmentScore[];
  deliveryContext?: DeliveryContextStat[];
  expectedRespondents: number;
  responseRatePct: number;
  deliverySyncedAt?: string;
}

const ADMIN_TOKEN_KEY = 'admin_token';

export function getAdminToken(): string | null {
  return localStorage.getItem(ADMIN_TOKEN_KEY);
}

export function setAdminToken(token: string): void {
  localStorage.setItem(ADMIN_TOKEN_KEY, token);
}

export function clearAdminToken(): void {
  localStorage.removeItem(ADMIN_TOKEN_KEY);
}

export async function fetchCurrentSurvey(): Promise<SurveyCurrentResponse> {
  const response = await fetch('/api/survey/current');
  if (!response.ok) {
    throw new Error('Не удалось загрузить опрос');
  }
  return response.json();
}

export async function submitSurvey(payload: {
  anonymousToken: string;
  answers: Record<string, number>;
  direction?: string;
  gradeBand?: string;
  role?: string;
}): Promise<void> {
  const response = await fetch('/api/survey/submit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });

  if (response.status === 409) {
    throw new Error('Вы уже отправляли опрос в этом квартале');
  }
  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data.error || 'Ошибка при отправке ответов');
  }
}

export async function verifyAdminPassword(password: string): Promise<boolean> {
  const response = await fetch('/api/admin/dashboard', {
    headers: { Authorization: `Bearer ${password}` },
  });
  return response.ok;
}

export async function fetchDashboard(): Promise<DashboardResponse> {
  const token = getAdminToken();
  if (!token) {
    throw new Error('Требуется авторизация');
  }

  const response = await fetch('/api/admin/dashboard', {
    headers: { Authorization: `Bearer ${token}` },
  });

  if (response.status === 401) {
    clearAdminToken();
    throw new Error('Сессия истекла');
  }
  if (!response.ok) {
    throw new Error('Не удалось загрузить дашборд');
  }
  return response.json();
}

export async function exportCsv(): Promise<void> {
  const token = getAdminToken();
  if (!token) {
    throw new Error('Требуется авторизация');
  }

  const response = await fetch('/api/admin/export', {
    headers: { Authorization: `Bearer ${token}` },
  });

  if (!response.ok) {
    throw new Error('Не удалось выгрузить CSV');
  }

  const blob = await response.blob();
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = 'gallup-q14-export.csv';
  link.click();
  URL.revokeObjectURL(url);
}

export async function syncDeliveryReference(): Promise<string> {
  const token = getAdminToken();
  if (!token) {
    throw new Error('Требуется авторизация');
  }

  const response = await fetch('/api/admin/sync-delivery', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  });

  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data.error || data.detail || 'Не удалось синхронизировать Delivery');
  }
  return data.message || 'Синхронизация завершена';
}

export const SEGMENT_TYPE_LABELS: Record<string, string> = {
  grade_band: 'Грейд',
  role: 'Роль',
};

export const DIMENSION_LABELS: Record<string, string> = {
  'Basic Needs': 'Роль и ресурсы',
  'Individual Contribution': 'Признание и поддержка',
  Teamwork: 'Голос, смысл и команда',
  Growth: 'Обратная связь и развитие',
};

/** Stable radar chart order (4 blocks per leadership TZ). */
export const DIMENSION_ORDER = [
  'Basic Needs',
  'Individual Contribution',
  'Teamwork',
  'Growth',
] as const;

export function formatRoundCode(code: string): string {
  const match = code.match(/^(\d{4})Q(\d)$/);
  if (!match) return code;
  return `Q${match[2]} ${match[1]}`;
}
