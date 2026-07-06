import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts';
import { formatRoundCode, type EnpsTrendPoint } from '../lib/api';
import { CHART, ENPS_BENCHMARK } from '../lib/chartTheme';

type Props = {
  trends: EnpsTrendPoint[];
};

export const EnpsTrendPanel = ({ trends }: Props) => {
  if (!trends.length) {
    return null;
  }

  const chartData = trends.map((t) => ({
    name: formatRoundCode(t.roundCode),
    score: Number(t.score.toFixed(1)),
    total: t.total,
    promotersPct: Number(t.promotersPct.toFixed(1)),
    passivesPct: Number(t.passivesPct.toFixed(1)),
    detractorsPct: Number(t.detractorsPct.toFixed(1)),
  }));

  return (
    <div className="card">
      <h3 className="text-lg font-bold mb-2 uppercase tracking-wide">Динамика eNPS по кварталам</h3>
      <p className="text-sm text-ink-muted mb-6">
        Индекс лояльности E01: % промоутеров (9–10) минус % критиков (0–6). Пунктир — порог +30.
      </p>
      <div className="h-72">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" stroke={CHART.grid} />
            <XAxis dataKey="name" stroke={CHART.axis} />
            <YAxis domain={[-100, 100]} stroke={CHART.axis} />
            <ReferenceLine y={0} stroke={CHART.axis} strokeDasharray="4 4" />
            <ReferenceLine y={30} stroke={ENPS_BENCHMARK} strokeDasharray="4 4" label={{ value: '+30', position: 'right', fill: CHART.axis, fontSize: 11 }} />
            <Tooltip
              contentStyle={{ backgroundColor: CHART.tooltipBg, borderColor: CHART.tooltipBorder }}
              formatter={(value, _name, props) => {
                const p = props.payload as {
                  score?: number;
                  total?: number;
                  promotersPct?: number;
                  passivesPct?: number;
                  detractorsPct?: number;
                };
                const score = Number(value ?? 0);
                return [
                  `${score >= 0 ? '+' : ''}${score} · n=${p.total ?? 0} · +${p.promotersPct ?? 0}% / ${p.passivesPct ?? 0}% / −${p.detractorsPct ?? 0}%`,
                  'eNPS',
                ];
              }}
            />
            <Line
              type="monotone"
              dataKey="score"
              stroke={CHART.primary}
              strokeWidth={3}
              dot={{ r: 6, fill: '#fff', stroke: CHART.primary, strokeWidth: 2 }}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
};
