import { Link } from 'react-router-dom';
import { ShieldCheck, BarChart3, Users, Database } from 'lucide-react';

export const LandingPage = () => {
  return (
    <div className="flex-1 flex flex-col">
      {/* Hero — 1:1 со sapiens.solutions index-header */}
      <section className="hero-section">
        <img src="/hero/bg2.png" alt="" className="hero-bg-dots" aria-hidden="true" />
        <img src="/hero/man2.png" alt="" className="hero-character" aria-hidden="true" />

        <div className="hero-text-block">
          <h1 className="hero-title">
            Опрос
            <br />
            вовлечённости
            <br />
            по методологии <span className="highlight-tag">Gallup Q14</span>
          </h1>
          <p className="hero-subtitle">
            Ежеквартальный анонимный мониторинг для команды Sapiens Solutions.
            <br className="hidden sm:block" />
            Gallup Q12 для DE-консалтинга (шкала 1–5) + быстрый индекс eNPS.
          </p>
          <div className="mt-8">
            <Link to="/survey" className="btn-hero">
              Пройти опрос
            </Link>
          </div>
        </div>

        <div className="hero-carousel-dots" aria-hidden="true">
          <span className="hero-dot hero-dot--active" />
          <span className="hero-dot" />
          <span className="hero-dot" />
          <span className="hero-dot" />
        </div>
      </section>

      {/* Принципы */}
      <section className="py-16 sm:py-20 bg-white">
        <div className="max-w-[177vmin] mx-auto px-6 sm:px-10 lg:px-16">
          <h2 className="section-title">Принципы опроса</h2>
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
            {[
              {
                icon: ShieldCheck,
                title: '100% анонимно',
                desc: 'Персональные данные не собираются. Токен в браузере нужен только для защиты от повторной отправки.',
              },
              {
                icon: BarChart3,
                title: 'Методология Gallup',
                desc: '12 вопросов вовлечённости по ТЗ руководства (DE-консалтинг) + eNPS 0–10.',
              },
              {
                icon: Database,
                title: 'Накопление данных',
                desc: 'Результаты сохраняются по кварталам и передаются в HR для анализа трендов.',
              },
              {
                icon: Users,
                title: 'Рекомендации',
                desc: 'Дашборд формирует рекомендации для руководства на основе зон с низкими показателями.',
              },
            ].map(({ icon: Icon, title, desc }) => (
              <div key={title} className="flex flex-col">
                <div className="w-12 h-12 bg-brand-light text-brand rounded flex items-center justify-center mb-4">
                  <Icon className="w-6 h-6" />
                </div>
                <h3 className="text-lg font-bold mb-2 uppercase tracking-wide text-navy">{title}</h3>
                <p className="text-ink-muted text-sm leading-relaxed">{desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Структура */}
      <section className="py-16 sm:py-20 bg-surface">
        <div className="max-w-[177vmin] mx-auto px-6 sm:px-10 lg:px-16">
          <h2 className="section-title">Структура опроса</h2>
          <div className="grid md:grid-cols-3 gap-6">
            {[
              { title: 'Q00', desc: 'Общая удовлетворённость компанией (шкала 1–5)' },
              { title: 'Q01–Q12', desc: 'Адаптированный Gallup Q12: роль, признание, команда, развитие (1–5)' },
              { title: 'E01 eNPS', desc: 'Готовность рекомендовать компанию: промоутеры, нейтралы, критики (0–10)' },
            ].map((item) => (
              <div key={item.title} className="card hover:border-brand/40 transition-colors">
                <h3 className="text-brand font-bold text-lg mb-2 uppercase">{item.title}</h3>
                <p className="text-ink-muted text-sm leading-relaxed">{item.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>
    </div>
  );
};
