import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell, ReferenceLine } from 'recharts';
import { SEGMENT_TYPE_LABELS, type EnpsSegmentScore } from '../lib/api';
import { CHART, enpsColor, ENPS_BENCHMARK, BAR_FILL } from '../lib/chartTheme';

type Props = {
  segments: EnpsSegmentScore[];
};

function formatEnps(score: number): string {
  return `${score >= 0 ? '+' : ''}${score.toFixed(0)}`;
}

function segmentChartHeight(itemCount: number): number {
  return Math.min(720, Math.max(224, itemCount * 40));
}

function truncateLabel(value: string, max = 32): string {
  return value.length > max ? `${value.slice(0, max - 1)}…` : value;
}

export const EnpsSegmentBreakdownPanel = ({ segments }: Props) => {
  if (!segments.length) {
    return null;
  }

  const types = [...new Set(segments.map((s) => s.segmentType))];

  return (
    <div className="card mb-8">
      <h3 className="text-lg font-bold mb-2 uppercase tracking-wide">eNPS по группам (Delivery)</h3>
      <p className="text-sm text-ink-muted mb-6">
        Индекс eNPS в текущем квартале по self-reported группам из справочника Delivery. В подсказке — доли
        промоутеров/нейтралов/критиков и охват ответов / актуальный штат компании.
        промоутеров, нейтралов и критиков.
      </p>

      <div className="space-y-8">
        {types.map((type) => {
          const items = segments
            .filter((s) => s.segmentType === type)
            .sort((a, b) => a.score - b.score);
          const chartData = items.map((s) => ({
            name: truncateLabel(s.segmentValue),
            fullName: s.segmentValue,
            score: Number(s.score.toFixed(1)),
            promotersPct: Number(s.promotersPct.toFixed(1)),
            passivesPct: Number(s.passivesPct.toFixed(1)),
            detractorsPct: Number(s.detractorsPct.toFixed(1)),
            total: s.total,
            submissions: s.submissionCount,
            expected: s.expectedCount,
          }));

          return (
            <div key={type}>
              <h4 className="text-sm font-bold uppercase tracking-wide text-brand mb-3">
                {SEGMENT_TYPE_LABELS[type] || type}
              </h4>
              <div style={{ height: segmentChartHeight(chartData.length) }}>
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={chartData} layout="vertical" margin={{ left: 8, right: 16 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke={CHART.grid} horizontal={false} />
                    <XAxis type="number" domain={[-100, 100]} stroke={CHART.axis} />
                    <YAxis
                      type="category"
                      dataKey="name"
                      width={type === 'position' ? 200 : 160}
                      stroke={CHART.axis}
                      tick={{ fontSize: 11 }}
                    />
                    <ReferenceLine x={0} stroke={CHART.axis} strokeDasharray="4 4" />
                    <ReferenceLine x={30} stroke={ENPS_BENCHMARK} strokeDasharray="4 4" />
                    <Tooltip
                      contentStyle={{ backgroundColor: CHART.tooltipBg, borderColor: CHART.tooltipBorder }}
                      formatter={(value, _name, props) => {
                        const p = props.payload as {
                          score?: number;
                          total?: number;
                          promotersPct?: number;
                          passivesPct?: number;
                          detractorsPct?: number;
                          submissions?: number;
                          expected?: number;
                          fullName?: string;
                        };
                        const score = Number(value ?? 0);
                        return [
                          `${formatEnps(score)} · n=${p.total ?? 0} · +${p.promotersPct ?? 0}% / ${p.passivesPct ?? 0}% / −${p.detractorsPct ?? 0}% · ${p.submissions ?? 0}/${p.expected ?? 0} чел.`,
                          p.fullName,
                        ];
                      }}
                    />
                    <Bar dataKey="score" radius={[0, 4, 4, 0]} fillOpacity={BAR_FILL.opacity}>
                      {chartData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={enpsColor(entry.score)} />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
