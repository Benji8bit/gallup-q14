import { BookOpen, ChevronDown, ChevronUp } from 'lucide-react';
import { useState } from 'react';
import { DIMENSION_LABELS, type MethodologyGuide } from '../lib/api';

type Props = {
  guide: MethodologyGuide;
};

export const MethodologyGuidePanel = ({ guide }: Props) => {
  const [expandedQuestion, setExpandedQuestion] = useState<string | null>(null);
  const [showAllQuestions, setShowAllQuestions] = useState(false);

  const visibleQuestions = showAllQuestions ? guide.questions : guide.questions.slice(0, 6);

  return (
    <aside className="card lg:sticky lg:top-6 h-fit">
      <div className="flex items-center gap-2 mb-4">
        <BookOpen className="w-5 h-5 text-brand shrink-0" />
        <h3 className="text-lg font-bold uppercase tracking-wide">Справочник по методике</h3>
      </div>

      <p className="text-sm text-ink-muted leading-relaxed mb-6">{guide.overview}</p>

      <section className="mb-6">
        <h4 className="text-xs font-bold uppercase tracking-wide text-brand mb-3">Метрики дашборда</h4>
        <div className="space-y-3">
          {guide.metrics.map((metric) => (
            <div key={metric.name} className="rounded-md border border-line bg-page p-3">
              <p className="font-bold text-sm text-navy">{metric.name}</p>
              <p className="text-xs text-ink-muted mt-1">{metric.description}</p>
              <p className="text-xs mt-2">
                <span className="text-ink-muted">Формула: </span>
                <span className="font-mono text-navy">{metric.formula}</span>
              </p>
              <p className="text-xs mt-1 text-ink-muted">Шкала: {metric.scale}</p>
            </div>
          ))}
        </div>
      </section>

      <section className="mb-6">
        <h4 className="text-xs font-bold uppercase tracking-wide text-brand mb-3">Шкалы ответов</h4>
        <div className="space-y-3">
          {guide.scales.map((scale) => (
            <div key={scale.title} className="rounded-md border border-line bg-page p-3">
              <p className="font-bold text-sm text-navy">{scale.title}</p>
              <p className="text-xs text-ink-muted mt-1">{scale.description}</p>
              {scale.mapping && scale.mapping.length > 0 && (
                <table className="w-full text-xs mt-3 border-collapse">
                  <thead>
                    <tr className="text-left text-ink-muted border-b border-line">
                      <th className="py-1 pr-2">Форма</th>
                      <th className="py-1 pr-2">Дашборд</th>
                      <th className="py-1">Смысл</th>
                    </tr>
                  </thead>
                  <tbody>
                    {scale.mapping.map((row) => (
                      <tr key={row.surveyValue} className="border-b border-line/60 last:border-0">
                        <td className="py-1 pr-2 font-mono">{row.surveyValue}</td>
                        <td className="py-1 pr-2 font-mono">{row.dashboardValue}</td>
                        <td className="py-1 text-ink-muted">{row.meaning}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          ))}
        </div>
      </section>

      <section className="mb-6">
        <h4 className="text-xs font-bold uppercase tracking-wide text-brand mb-3">Измерения</h4>
        <div className="space-y-2">
          {guide.dimensions.map((dim) => (
            <div key={dim.key} className="rounded-md border border-line p-3">
              <p className="font-bold text-sm text-navy">{dim.labelRu}</p>
              <p className="text-xs text-ink-muted mt-1">{dim.description}</p>
              <p className="text-xs mt-2">
                <span className="text-ink-muted">Аудитория действий: </span>
                {dim.primaryAudience}
              </p>
            </div>
          ))}
        </div>
      </section>

      <section>
        <h4 className="text-xs font-bold uppercase tracking-wide text-brand mb-3">Как читать вопросы</h4>
        <p className="text-xs text-ink-muted mb-3">
          Низкий favorable по вопросу означает, что сотрудники не согласны с формулировкой — это сигнал для
          руководства, а не готовый action plan. Ниже — что измеряет каждый вопрос и на что смотреть лидеру.
        </p>
        <div className="space-y-2">
          {visibleQuestions.map((q) => {
            const open = expandedQuestion === q.code;
            return (
              <div key={q.code} className="rounded-md border border-line overflow-hidden">
                <button
                  type="button"
                  onClick={() => setExpandedQuestion(open ? null : q.code)}
                  className="w-full flex items-start justify-between gap-2 p-3 text-left hover:bg-page transition-colors"
                >
                  <div>
                    <span className="font-mono text-xs font-bold text-brand">{q.code}</span>
                    {q.dimension && (
                      <span className="text-xs text-ink-muted ml-2">
                        {DIMENSION_LABELS[q.dimension] || q.dimension}
                      </span>
                    )}
                    <p className="text-xs text-navy mt-1 line-clamp-2">{q.textRu}</p>
                    {q.currentFavPct !== undefined && q.currentFavPct > 0 && (
                      <p className="text-xs text-ink-muted mt-1">
                        Сейчас: {q.currentFavPct.toFixed(1)}% favorable · {q.currentAvg5pt?.toFixed(2)}/5
                      </p>
                    )}
                  </div>
                  {open ? (
                    <ChevronUp className="w-4 h-4 shrink-0 text-ink-muted mt-0.5" />
                  ) : (
                    <ChevronDown className="w-4 h-4 shrink-0 text-ink-muted mt-0.5" />
                  )}
                </button>
                {open && (
                  <div className="px-3 pb-3 pt-0 border-t border-line bg-page">
                    <p className="text-xs mt-3">
                      <span className="font-bold text-navy">Что измеряет: </span>
                      {q.whatItMeasures}
                    </p>
                    <p className="text-xs mt-2">
                      <span className="font-bold text-navy">Фокус руководства: </span>
                      {q.leadershipFocus}
                    </p>
                  </div>
                )}
              </div>
            );
          })}
        </div>
        {guide.questions.length > 6 && (
          <button
            type="button"
            onClick={() => setShowAllQuestions((v) => !v)}
            className="mt-3 text-xs font-bold text-brand hover:text-brand-hover"
          >
            {showAllQuestions ? 'Свернуть список' : `Показать все ${guide.questions.length} вопросов`}
          </button>
        )}
      </section>
    </aside>
  );
};
