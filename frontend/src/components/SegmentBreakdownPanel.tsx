import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell } from 'recharts';
import { SEGMENT_TYPE_LABELS, type SegmentScore } from '../lib/api';
import { CHART, scoreColor, BAR_FILL } from '../lib/chartTheme';

function segmentChartHeight(itemCount: number): number {
  return Math.min(720, Math.max(224, itemCount * 40));
}

function truncateLabel(value: string, max = 32): string {
  return value.length > max ? `${value.slice(0, max - 1)}…` : value;
}

type Props = {
  segments: SegmentScore[];
};

export const SegmentBreakdownPanel = ({ segments }: Props) => {
  if (!segments.length) {
    return null;
  }

  const types = [...new Set(segments.map((s) => s.segmentType))];

  return (
    <div className="card mb-8">
      <h3 className="text-lg font-bold mb-2 uppercase tracking-wide">Срезы по группам (Delivery)</h3>
      <p className="text-sm text-ink-muted mb-6">
        Вовлечённость по self-reported группам из справочника Delivery Sapiens. Показан охват: ответы /
        актуальный штат компании.
      </p>

      <div className="space-y-8">
        {types.map((type) => {
          const items = segments
            .filter((s) => s.segmentType === type)
            .sort((a, b) => a.engagementScore - b.engagementScore);
          const chartData = items.map((s) => ({
            name: truncateLabel(s.segmentValue),
            fullName: s.segmentValue,
            score: Number(s.engagementScore.toFixed(1)),
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
                    <XAxis type="number" domain={[0, 100]} unit="%" stroke={CHART.axis} />
                    <YAxis
                      type="category"
                      dataKey="name"
                      width={type === 'position' ? 200 : 160}
                      stroke={CHART.axis}
                      tick={{ fontSize: 11 }}
                    />
                    <Tooltip
                      contentStyle={{ backgroundColor: CHART.tooltipBg, borderColor: CHART.tooltipBorder }}
                      formatter={(value, _name, props) => {
                        const p = props.payload as { submissions?: number; expected?: number; fullName?: string };
                        return [
                          `${Number(value ?? 0)}% · ${p.submissions ?? 0}/${p.expected ?? 0} чел.`,
                          p.fullName,
                        ];
                      }}
                    />
                    <Bar dataKey="score" radius={[0, 4, 4, 0]} fillOpacity={BAR_FILL.opacity}>
                      {chartData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={scoreColor(entry.score / 20)} />
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
