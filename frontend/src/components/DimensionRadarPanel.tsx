import {
  RadarChart,
  Radar,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  ResponsiveContainer,
  Tooltip,
} from 'recharts';
import { CHART } from '../lib/chartTheme';
import {
  DIMENSION_LABELS,
  DIMENSION_ORDER,
  type DimensionScore,
} from '../lib/api';

const DIMENSION_SHORT_LABELS: Record<string, string> = {
  'Basic Needs': 'Роль',
  'Individual Contribution': 'Признание',
  Teamwork: 'Команда',
  Growth: 'Развитие',
};

export function buildDimensionRadarData(breakdown: DimensionScore[]) {
  const byKey = Object.fromEntries(breakdown.map((d) => [d.dimension, d]));
  return DIMENSION_ORDER.map((key) => ({
    key,
    shortLabel: DIMENSION_SHORT_LABELS[key] ?? key,
    fullLabel: DIMENSION_LABELS[key] ?? key,
    A: Number((byKey[key]?.averageScore5pt ?? 0).toFixed(2)),
    favorable: Number((byKey[key]?.favorablePct ?? 0).toFixed(1)),
    hasData: Boolean(byKey[key]),
  }));
}

type Props = {
  breakdown: DimensionScore[];
};

export const DimensionRadarPanel = ({ breakdown }: Props) => {
  const data = buildDimensionRadarData(breakdown);
  const hasAny = data.some((d) => d.hasData);

  if (!hasAny) {
    return (
      <div className="card">
        <h3 className="text-lg font-bold mb-6 uppercase tracking-wide">Индексы по направлениям</h3>
        <div className="h-80 flex items-center justify-center text-ink-muted">Нет данных</div>
      </div>
    );
  }

  return (
    <div className="card">
      <h3 className="text-lg font-bold mb-2 uppercase tracking-wide">Индексы по направлениям</h3>
      <p className="text-sm text-ink-muted mb-4">Средний балл 1–5 по четырём блокам ТЗ</p>

      <div className="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_18rem] gap-6 items-center">
        <div className="h-[min(480px,70vw)] min-h-[320px] w-full">
          <ResponsiveContainer width="100%" height="100%">
            <RadarChart
              cx="50%"
              cy="50%"
              outerRadius="72%"
              data={data}
              margin={{ top: 8, right: 28, bottom: 8, left: 28 }}
            >
              <PolarGrid stroke={CHART.grid} />
              <PolarAngleAxis
                dataKey="shortLabel"
                tick={{ fill: CHART.axis, fontSize: 15, fontWeight: 700 }}
                tickLine={false}
              />
              <PolarRadiusAxis
                angle={90}
                domain={[0, 5]}
                tickCount={6}
                stroke={CHART.axis}
                fontSize={11}
                axisLine={false}
              />
              <Radar
                name="Средний балл"
                dataKey="A"
                stroke={CHART.primary}
                fill={CHART.primary}
                fillOpacity={0.22}
                strokeWidth={2.5}
                dot={{ r: 5, fill: CHART.primary, stroke: '#fff', strokeWidth: 2 }}
              />
              <Tooltip
                contentStyle={{ backgroundColor: CHART.tooltipBg, borderColor: CHART.tooltipBorder }}
                formatter={(value, _name, item) => {
                  const p = item?.payload as { fullLabel?: string; favorable?: number } | undefined;
                  return [
                    `${Number(value ?? 0)} / 5 · favorable ${p?.favorable ?? 0}%`,
                    p?.fullLabel ?? '',
                  ];
                }}
              />
            </RadarChart>
          </ResponsiveContainer>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-1 gap-2 text-sm">
          {data.map((d) => (
            <div
              key={d.key}
              className="flex justify-between gap-3 rounded border border-line px-3 py-2.5 bg-white/60"
            >
              <span className="text-ink-muted leading-snug">{d.fullLabel}</span>
              <span className="font-semibold text-navy whitespace-nowrap">
                {d.A.toFixed(1)}
                <span className="text-ink-muted font-normal text-xs ml-1">/ 5</span>
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
