import { ThumbsDown, Minus, ThumbsUp } from 'lucide-react';
import type { EnpsScore } from '../lib/api';

function enpsBand(score: number): { label: string; className: string } {
  if (score >= 30) return { label: 'Сильная зона', className: 'text-brand' };
  if (score >= 0) return { label: 'Зона внимания', className: 'text-amber-600' };
  return { label: 'Тревожный сигнал', className: 'text-red-600' };
}

export const EnpsPanel = ({ enps }: { enps: EnpsScore }) => {
  const band = enpsBand(enps.score);

  return (
    <div className="card mb-8">
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4 mb-6">
        <div>
          <h3 className="text-lg font-bold uppercase tracking-wide">eNPS — готовность рекомендовать</h3>
          <p className="text-sm text-ink-muted mt-1">
            E01: «Насколько вероятно, что вы порекомендуете компанию другу или коллеге?» (0–10)
          </p>
        </div>
        <div className="text-right">
          <p className={`text-4xl font-bold ${band.className}`}>
            {enps.score >= 0 ? '+' : ''}
            {enps.score.toFixed(0)}
          </p>
          <p className={`text-sm font-medium ${band.className}`}>{band.label}</p>
          <p className="text-xs text-ink-muted">норма &gt; +30; &lt; 0 — тревожно</p>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="rounded-lg border border-line p-4 bg-brand-light/40">
          <div className="flex items-center gap-2 text-brand mb-2">
            <ThumbsUp className="w-4 h-4" />
            <span className="text-sm font-bold uppercase">Промоутеры (9–10)</span>
          </div>
          <p className="text-2xl font-bold text-navy">{enps.promotersPct.toFixed(1)}%</p>
          <p className="text-xs text-ink-muted mt-1">{enps.promoters} из {enps.total} ответов</p>
        </div>
        <div className="rounded-lg border border-line p-4 bg-amber-50/50">
          <div className="flex items-center gap-2 text-amber-700 mb-2">
            <Minus className="w-4 h-4" />
            <span className="text-sm font-bold uppercase">Нейтралы (7–8)</span>
          </div>
          <p className="text-2xl font-bold text-amber-800">{enps.passivesPct.toFixed(1)}%</p>
          <p className="text-xs text-ink-muted mt-1">{enps.passives} из {enps.total} ответов</p>
        </div>
        <div className="rounded-lg border border-line p-4 bg-red-50/50">
          <div className="flex items-center gap-2 text-red-700 mb-2">
            <ThumbsDown className="w-4 h-4" />
            <span className="text-sm font-bold uppercase">Критики (0–6)</span>
          </div>
          <p className="text-2xl font-bold text-red-800">{enps.detractorsPct.toFixed(1)}%</p>
          <p className="text-xs text-ink-muted mt-1">{enps.detractors} из {enps.total} ответов</p>
        </div>
      </div>

      <p className="text-xs text-ink-muted mt-4">
        Формула: eNPS = % промоутеров − % критиков. Нейтралы не входят в расчёт индекса, но показаны для полной картины.
      </p>
    </div>
  );
};
