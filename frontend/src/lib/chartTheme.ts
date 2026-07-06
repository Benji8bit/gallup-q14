/** Brand colors for charts — Sapiens blue palette */
export const CHART = {
  primary: '#2965ff',
  secondary: '#5289ff',
  tertiary: '#94b2ff',
  light: '#cce1ff',
  grid: '#e8e8e8',
  axis: '#595959',
  tooltipBg: '#ffffff',
  tooltipBorder: '#e8e8e8',
  low: '#ef4444',
  mid: '#5289ff',
  high: '#2965ff',
} as const;

/** Softer bar fills — less saturated than line/radar accents */
export const BAR_FILL = {
  high: '#94b2ff',
  mid: '#b8cbff',
  low: '#f87171',
  opacity: 0.88,
} as const;

export function scoreColor(score: number): string {
  if (score >= 4) return BAR_FILL.high;
  if (score >= 3) return BAR_FILL.mid;
  return BAR_FILL.low;
}

export function enpsColor(score: number): string {
  if (score >= 30) return BAR_FILL.high;
  if (score >= 0) return BAR_FILL.mid;
  return BAR_FILL.low;
}

/** Benchmark line for eNPS (+30 threshold). */
export const ENPS_BENCHMARK = CHART.tertiary;
