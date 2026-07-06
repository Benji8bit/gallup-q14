import { AlertTriangle, Target, Users } from 'lucide-react';
import { DIMENSION_LABELS, type RecommendationItem } from '../lib/api';

type Props = {
  recommendations: RecommendationItem[];
};

const RecommendationCard = ({ item }: { item: RecommendationItem }) => (
  <div className="p-4 rounded-md border flex flex-col gap-3 bg-brand-light border-brand/20 text-navy">
    <div className="flex gap-3">
      <AlertTriangle className="w-5 h-5 shrink-0 text-brand mt-0.5" />
      <div className="min-w-0">
        <p className="font-bold">{item.title}</p>
        <p className="mt-2 text-sm leading-relaxed">{item.action}</p>
      </div>
    </div>

    <div className="ml-8 space-y-2 text-sm">
      <div className="rounded-md bg-white/70 border border-brand/10 p-3">
        <p className="text-xs font-bold uppercase tracking-wide text-brand mb-1">На чём основано</p>
        <p className="text-sm leading-relaxed">{item.basis}</p>
      </div>

      {item.targetGroups.length > 0 && (
        <div className="flex flex-wrap items-start gap-2">
          <Users className="w-4 h-4 shrink-0 text-brand mt-0.5" />
          <div>
            <p className="text-xs font-bold uppercase tracking-wide text-ink-muted mb-1">Для кого</p>
            <div className="flex flex-wrap gap-1.5">
              {item.targetGroups.map((group) => (
                <span
                  key={group}
                  className="inline-block text-xs px-2 py-0.5 rounded-full bg-white border border-brand/20"
                >
                  {group}
                </span>
              ))}
            </div>
          </div>
        </div>
      )}

      {(item.relatedDimension || (item.relatedQuestions && item.relatedQuestions.length > 0)) && (
        <div className="flex flex-wrap items-center gap-2 text-xs text-ink-muted">
          <Target className="w-4 h-4 shrink-0" />
          {item.relatedDimension && (
            <span>Измерение: {DIMENSION_LABELS[item.relatedDimension] || item.relatedDimension}</span>
          )}
          {item.relatedQuestions && item.relatedQuestions.length > 0 && (
            <span>Вопросы: {item.relatedQuestions.join(', ')}</span>
          )}
        </div>
      )}
    </div>
  </div>
);

export const RecommendationsPanel = ({ recommendations }: Props) => {
  const general = recommendations.filter((r) => r.scope === 'general');
  const grouped = recommendations.filter((r) => r.scope === 'group');

  if (recommendations.length === 0) {
    return (
      <div className="card">
        <h3 className="text-lg font-bold mb-4 uppercase tracking-wide">Рекомендации для руководства</h3>
        <p className="text-ink-muted text-sm">Недостаточно данных для формирования рекомендаций.</p>
      </div>
    );
  }

  return (
    <div className="card">
      <h3 className="text-lg font-bold mb-2 uppercase tracking-wide">Рекомендации для руководства</h3>
      <p className="text-sm text-ink-muted mb-6">
        Каждая рекомендация привязана к метрикам опроса и указывает, кому адресовано действие. Используйте
        справочник справа, чтобы понять, что означает реакция сотрудников на конкретные вопросы.
      </p>

      {general.length > 0 && (
        <section className="mb-8">
          <h4 className="text-sm font-bold uppercase tracking-wide text-brand mb-3">Общие</h4>
          <div className="space-y-3">
            {general.map((item) => (
              <RecommendationCard key={item.id} item={item} />
            ))}
          </div>
        </section>
      )}

      {grouped.length > 0 && (
        <section>
          <h4 className="text-sm font-bold uppercase tracking-wide text-brand mb-3">По группам и зонам риска</h4>
          <div className="space-y-3">
            {grouped.map((item) => (
              <RecommendationCard key={item.id} item={item} />
            ))}
          </div>
        </section>
      )}
    </div>
  );
};
