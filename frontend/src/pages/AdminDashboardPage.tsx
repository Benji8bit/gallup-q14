import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip as RechartsTooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
  Cell,
} from 'recharts';
import { Users, TrendingUp, Download, Loader2, RefreshCw, Percent } from 'lucide-react';
import { DeliveryContextPanel } from '../components/DeliveryContextPanel';
import { DimensionRadarPanel } from '../components/DimensionRadarPanel';
import { EnpsPanel } from '../components/EnpsPanel';
import { EnpsSegmentBreakdownPanel } from '../components/EnpsSegmentBreakdownPanel';
import { EnpsTrendPanel } from '../components/EnpsTrendPanel';
import { MethodologyGuidePanel } from '../components/MethodologyGuidePanel';
import { RecommendationsPanel } from '../components/RecommendationsPanel';
import { SegmentBreakdownPanel } from '../components/SegmentBreakdownPanel';
import {
  clearAdminToken,
  exportCsv,
  fetchDashboard,
  formatRoundCode,
  getAdminToken,
  syncDeliveryReference,
  type DashboardResponse,
} from '../lib/api';
import { CHART, scoreColor, BAR_FILL } from '../lib/chartTheme';

export const AdminDashboardPage = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [exporting, setExporting] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [syncMessage, setSyncMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [data, setData] = useState<DashboardResponse | null>(null);

  useEffect(() => {
    if (!getAdminToken()) {
      navigate('/admin');
      return;
    }

    fetchDashboard()
      .then(setData)
      .catch((err) => {
        if (err instanceof Error && err.message === 'Сессия истекла') {
          navigate('/admin');
          return;
        }
        setError(err instanceof Error ? err.message : 'Ошибка загрузки');
      })
      .finally(() => setLoading(false));
  }, [navigate]);

  const handleSync = async () => {
    setSyncing(true);
    setSyncMessage(null);
    try {
      const message = await syncDeliveryReference();
      setSyncMessage(message);
      const refreshed = await fetchDashboard();
      setData(refreshed);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка синхронизации');
    } finally {
      setSyncing(false);
    }
  };

  const handleExport = async () => {
    setExporting(true);
    try {
      await exportCsv();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка экспорта');
    } finally {
      setExporting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center min-h-screen bg-surface">
        <Loader2 className="w-8 h-8 animate-spin text-brand" />
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="flex-1 flex items-center justify-center p-4 min-h-screen bg-surface">
        <div className="card text-center max-w-md">
          <p className="text-red-500 mb-4">{error || 'Нет данных'}</p>
          <button onClick={() => window.location.reload()} className="btn-primary">
            Обновить
          </button>
        </div>
      </div>
    );
  }

  const trendData = data.trends.map((t) => ({
    name: formatRoundCode(t.roundCode),
    score: Number(t.engagementPct.toFixed(1)),
    submissions: t.submissionCount,
  }));

  const questionData = data.questionScores.map((q) => ({
    name: q.code,
    score: Number(q.averageScore5pt.toFixed(2)),
    favorable: Number(q.favorablePct.toFixed(1)),
  }));

  const latestTrend = trendData[trendData.length - 1];
  const previousTrend = trendData.length > 1 ? trendData[trendData.length - 2] : null;
  const trendDelta =
    latestTrend && previousTrend ? Number((latestTrend.score - previousTrend.score).toFixed(1)) : null;

  return (
    <div className="flex-1 min-h-screen bg-surface">
      {/* Navy header block — в тон hero sapiens.solutions */}
      <div className="dashboard-header">
        <div className="max-w-7xl mx-auto">
          <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-8">
            <div>
              <p className="text-brand-hover text-sm font-bold uppercase tracking-wide mb-1">HR Analytics</p>
              <h1 className="text-3xl font-bold uppercase tracking-wide text-white">Дашборд вовлечённости</h1>
              <p className="text-white/60 text-sm mt-1">
                {formatRoundCode(data.currentRoundCode)} · обновлено{' '}
                {new Date(data.generatedAt).toLocaleString('ru-RU')}
              </p>
            </div>
            <div className="flex gap-3 flex-wrap">
              <button
                onClick={handleSync}
                disabled={syncing}
                className="inline-flex items-center gap-2 px-4 py-2 rounded border border-white/30 text-white text-sm font-bold uppercase tracking-wide hover:bg-white/10 transition-colors"
              >
                {syncing ? <Loader2 className="w-4 h-4 animate-spin" /> : <RefreshCw className="w-4 h-4" />}
                Delivery
              </button>
              <button
                onClick={handleExport}
                disabled={exporting}
                className="inline-flex items-center gap-2 px-4 py-2 rounded border border-white/30 text-white text-sm font-bold uppercase tracking-wide hover:bg-white/10 transition-colors"
              >
                {exporting ? <Loader2 className="w-4 h-4 animate-spin" /> : <Download className="w-4 h-4" />}
                CSV
              </button>
              <button
                onClick={() => {
                  clearAdminToken();
                  navigate('/admin');
                }}
                className="btn-hero !min-w-0 !py-2 !px-5 !text-sm"
              >
                Выйти
              </button>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4">
            <div className="dashboard-kpi flex items-center gap-4">
              <div className="w-12 h-12 rounded-full bg-brand/30 flex items-center justify-center text-brand-hover">
                <TrendingUp className="w-6 h-6" />
              </div>
              <div>
                <p className="text-white/60 text-sm">Индекс вовлечённости</p>
                <p className="text-2xl font-bold text-white">
                  {data.engagementScore.toFixed(1)}%
                  {trendDelta !== null && (
                    <span
                      className={`text-sm font-normal ml-2 ${trendDelta >= 0 ? 'text-brand-hover' : 'text-red-400'}`}
                    >
                      {trendDelta >= 0 ? '+' : ''}
                      {trendDelta} п.п.
                    </span>
                  )}
                </p>
              </div>
            </div>
            <div className="dashboard-kpi flex items-center gap-4">
              <div className="w-12 h-12 rounded-full bg-brand/30 flex items-center justify-center text-brand-hover">
                <Users className="w-6 h-6" />
              </div>
              <div>
                <p className="text-white/60 text-sm">Участники ({formatRoundCode(data.currentRoundCode)})</p>
                <p className="text-2xl font-bold text-white">{data.currentRoundSubmissions}</p>
              </div>
            </div>
            <div className="dashboard-kpi flex items-center gap-4">
              <div className="w-12 h-12 rounded-full bg-brand/30 flex items-center justify-center text-brand-hover">
                <Percent className="w-6 h-6" />
              </div>
              <div>
                <p className="text-white/60 text-sm">Охват (Data Engineering)</p>
                <p className="text-2xl font-bold text-white">
                  {data.responseRatePct.toFixed(1)}%
                  <span className="text-sm text-white/50 font-normal ml-1">
                    / {data.expectedRespondents}
                  </span>
                </p>
              </div>
            </div>
            <div className="dashboard-kpi flex items-center gap-4">
              <div className="w-12 h-12 rounded-full bg-brand/30 flex items-center justify-center text-brand-hover">
                <TrendingUp className="w-6 h-6" />
              </div>
              <div>
                <p className="text-white/60 text-sm">eNPS (E01)</p>
                <p className="text-2xl font-bold text-white">
                  {data.enpsScore ? (
                    <>
                      {data.enpsScore.score >= 0 ? '+' : ''}
                      {data.enpsScore.score.toFixed(0)}
                    </>
                  ) : (
                    '—'
                  )}
                </p>
              </div>
            </div>
          </div>
          {syncMessage && (
            <p className="text-brand-hover text-xs mt-4">{syncMessage}</p>
          )}
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 xl:grid-cols-[1fr_22rem] gap-8 items-start">
          <div>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <div className="card">
          <h3 className="text-lg font-bold mb-6 uppercase tracking-wide">Динамика вовлечённости по кварталам</h3>
          <div className="h-72">
            {trendData.length > 0 ? (
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={trendData}>
                  <CartesianGrid strokeDasharray="3 3" stroke={CHART.grid} />
                  <XAxis dataKey="name" stroke={CHART.axis} />
                  <YAxis domain={[0, 100]} stroke={CHART.axis} unit="%" />
                  <RechartsTooltip
                    contentStyle={{ backgroundColor: CHART.tooltipBg, borderColor: CHART.tooltipBorder }}
                    itemStyle={{ color: CHART.primary }}
                    formatter={(value) => [`${Number(value ?? 0)}%`, 'Favorable']}
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
            ) : (
              <div className="h-full flex items-center justify-center text-ink-muted">Нет данных за прошлые кварталы</div>
            )}
          </div>
        </div>
        </div>

        <div className="mb-8">
        <DimensionRadarPanel breakdown={data.dimensionBreakdown} />
        </div>

      {(data.enpsTrends?.length ?? 0) > 0 && (
        <div className="mb-8">
          <EnpsTrendPanel trends={data.enpsTrends ?? []} />
        </div>
      )}

      {data.enpsScore && <EnpsPanel enps={data.enpsScore} />}

      <div className="card mb-8">
        <h3 className="text-lg font-bold mb-6 uppercase tracking-wide">Результаты по вопросам</h3>
        <div className="h-80">
          {questionData.length > 0 ? (
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={questionData} margin={{ top: 20, right: 30, left: 0, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" stroke={CHART.grid} vertical={false} />
                <XAxis dataKey="name" stroke={CHART.axis} />
                <YAxis domain={[0, 5]} stroke={CHART.axis} />
                <RechartsTooltip
                  contentStyle={{ backgroundColor: CHART.tooltipBg, borderColor: CHART.tooltipBorder }}
                  cursor={{ fill: '#f0f5ff' }}
                  formatter={(value, _name, props) => [
                    `${Number(value ?? 0)} (favorable: ${props.payload?.favorable ?? 0}%)`,
                    'Средний балл',
                  ]}
                />
                <Bar dataKey="score" radius={[4, 4, 0, 0]} fillOpacity={BAR_FILL.opacity}>
                  {questionData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={scoreColor(entry.score)} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <div className="h-full flex items-center justify-center text-ink-muted">Нет ответов</div>
          )}
        </div>
      </div>

      <DeliveryContextPanel
        stats={data.deliveryContext ?? []}
        syncedAt={data.deliverySyncedAt}
        expectedRespondents={data.expectedRespondents}
        responseRatePct={data.responseRatePct}
      />

      <SegmentBreakdownPanel segments={data.segmentBreakdown ?? []} />

      <EnpsSegmentBreakdownPanel segments={data.enpsSegmentBreakdown ?? []} />

      <RecommendationsPanel recommendations={data.recommendations ?? []} />
          </div>

          {data.methodologyGuide && <MethodologyGuidePanel guide={data.methodologyGuide} />}
        </div>
      </div>
    </div>
  );
};
