import { Database } from 'lucide-react';
import type { DeliveryContextStat } from '../lib/api';

type Props = {
  stats: DeliveryContextStat[];
  syncedAt?: string;
  expectedRespondents: number;
  responseRatePct: number;
};

export const DeliveryContextPanel = ({ stats, syncedAt, expectedRespondents, responseRatePct }: Props) => {
  const loadBands = stats.filter((s) => s.key === 'load_band');
  const directionCounts = stats.filter((s) => s.key === 'direction_headcount');
  const staffTotal = stats.find((s) => s.key === 'summary' && s.value === 'staff_total');
  const activeQtd = stats.find((s) => s.key === 'summary' && s.value === 'active_delivery_qtd');

  if (!stats.length) {
    return (
      <div className="card mb-8 border-dashed">
        <p className="text-sm text-ink-muted">
          Справочник Delivery не синхронизирован. Запустите синхронизацию из дашборда HR.
        </p>
      </div>
    );
  }

  return (
    <div className="card mb-8">
      <div className="flex items-center gap-2 mb-4">
        <Database className="w-5 h-5 text-brand" />
        <h3 className="text-lg font-bold uppercase tracking-wide">Контекст Delivery Sapiens</h3>
      </div>

      {syncedAt && (
        <p className="text-xs text-ink-muted mb-4">
          Синхронизировано: {new Date(syncedAt).toLocaleString('ru-RU')}
        </p>
      )}

      <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-3 mb-6">
        <div className="rounded-md border border-line bg-page p-3">
          <p className="text-xs text-ink-muted">Штат компании (актуальный)</p>
          <p className="text-xl font-bold text-navy">{staffTotal?.numericValue ?? expectedRespondents}</p>
        </div>
        <div className="rounded-md border border-line bg-page p-3">
          <p className="text-xs text-ink-muted">Delivery-активность (квартал)</p>
          <p className="text-xl font-bold text-navy">{activeQtd?.numericValue ?? 0}</p>
        </div>
        <div className="rounded-md border border-line bg-page p-3">
          <p className="text-xs text-ink-muted">Охват опроса</p>
          <p className="text-xl font-bold text-navy">{responseRatePct.toFixed(1)}%</p>
        </div>
        <div className="rounded-md border border-line bg-page p-3">
          <p className="text-xs text-ink-muted">Знаменатель охвата</p>
          <p className="text-xl font-bold text-navy">{expectedRespondents}</p>
        </div>
      </div>

      {loadBands.length > 0 && (
        <div className="mb-4">
          <p className="text-xs font-bold uppercase tracking-wide text-brand mb-2">Нагрузка (mandays/квартал)</p>
          <div className="flex flex-wrap gap-2">
            {loadBands.map((b) => (
              <span key={b.value} className="text-xs px-2 py-1 rounded-full bg-brand-light border border-brand/20">
                {b.value}: {b.numericValue} чел.
              </span>
            ))}
          </div>
        </div>
      )}

      {directionCounts.length > 0 && (
        <div>
          <p className="text-xs font-bold uppercase tracking-wide text-brand mb-2">Штат по направлениям</p>
          <div className="flex flex-wrap gap-2">
            {directionCounts.map((d) => (
              <span key={d.value} className="text-xs px-2 py-1 rounded-full bg-page border border-line">
                {d.value}: {d.numericValue}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
