import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { v4 as uuidv4 } from 'uuid';
import { Loader2 } from 'lucide-react';
import { fetchCurrentSurvey, submitSurvey, formatRoundCode, type OrgOptionGroup, type Question } from '../lib/api';
import { SURVEY_DIRECTION, SURVEY_ROLES } from '../lib/surveyConstants';

type SurveyMetadata = {
  gradeBand: string;
  role: string;
};

const emptyMetadata = (): SurveyMetadata => ({
  gradeBand: '',
  role: '',
});

function scaleValues(question: Question): number[] {
  const values: number[] = [];
  for (let v = question.scaleMin; v <= question.scaleMax; v++) {
    values.push(v);
  }
  return values;
}

function QuestionScale({
  question,
  value,
  onChange,
}: {
  question: Question;
  value: number | null | undefined;
  onChange: (val: number) => void;
}) {
  const values = scaleValues(question);

  if (question.questionRole === 'enps') {
    return (
      <div className="space-y-2">
        <div className="grid grid-cols-6 sm:grid-cols-11 gap-2">
          {values.map((val) => (
            <button
              key={val}
              type="button"
              onClick={() => onChange(val)}
              className={`py-3 px-1 rounded-md border transition-colors ${
                value === val
                  ? 'bg-brand text-white border-brand font-bold shadow-btn'
                  : 'bg-white border-line hover:border-brand/50'
              }`}
            >
              {val}
            </button>
          ))}
        </div>
        <div className="w-full flex justify-between text-xs text-ink-muted px-1">
          <span>0 — точно не порекомендую</span>
          <span>10 — точно порекомендую</span>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      <div className="grid grid-cols-5 gap-2">
        {values.map((val) => (
          <button
            key={val}
            type="button"
            onClick={() => onChange(val)}
            title={
              val === 1
                ? 'Совершенно не согласен'
                : val === 5
                  ? 'Полностью согласен'
                  : undefined
            }
            className={`py-3 px-2 rounded-md border transition-colors flex flex-col items-center justify-center ${
              value === val
                ? 'bg-brand text-white border-brand font-bold shadow-btn'
                : 'bg-white border-line hover:border-brand/50'
            }`}
          >
            <span className="text-lg">{val}</span>
          </button>
        ))}
      </div>
      <div className="w-full flex justify-between text-xs text-ink-muted px-1">
        <span>Совершенно не согласен</span>
        <span>Полностью согласен</span>
      </div>
    </div>
  );
}

export const SurveyPage = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [roundCode, setRoundCode] = useState('');
  const [questions, setQuestions] = useState<Question[]>([]);
  const [gradeOptions, setGradeOptions] = useState<OrgOptionGroup['options']>([]);
  const [answers, setAnswers] = useState<Record<string, number | null>>({});
  const [metadata, setMetadata] = useState<SurveyMetadata>(emptyMetadata());

  useEffect(() => {
    fetchCurrentSurvey()
      .then((data) => {
        setRoundCode(data.round.code);
        setQuestions(data.questions);
        const grades = data.orgOptions?.find((g) => g.type === 'grade_band')?.options ?? [];
        setGradeOptions([...grades].sort((a, b) => a.labelRu.localeCompare(b.labelRu, 'ru')));
        setLoading(false);
      })
      .catch(() => {
        setError('Не удалось загрузить опрос. Пожалуйста, попробуйте позже.');
        setLoading(false);
      });
  }, []);

  const handleAnswer = (questionId: string, value: number | null) => {
    setAnswers((prev) => ({ ...prev, [questionId]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!metadata.gradeBand || !metadata.role) {
      setError('Заполните все поля в разделе «О вас».');
      document.getElementById('about-you')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
      return;
    }

    const unanswered = questions.filter(
      (q) => answers[q.id] === undefined || answers[q.id] === null
    );
    if (unanswered.length > 0) {
      setError('Пожалуйста, ответьте на все вопросы.');
      const firstUnansweredEl = document.getElementById(`question-${unanswered[0].id}`);
      firstUnansweredEl?.scrollIntoView({ behavior: 'smooth', block: 'center' });
      return;
    }

    setSubmitting(true);
    setError(null);

    let token = localStorage.getItem('respondent_token');
    if (!token) {
      token = uuidv4();
      localStorage.setItem('respondent_token', token);
    }

    const payloadAnswers: Record<string, number> = {};
    for (const q of questions) {
      const value = answers[q.id];
      if (value !== null && value !== undefined) {
        payloadAnswers[q.id] = value;
      }
    }

    try {
      await submitSurvey({
        anonymousToken: token,
        answers: payloadAnswers,
        direction: SURVEY_DIRECTION,
        gradeBand: metadata.gradeBand,
        role: metadata.role,
      });
      navigate('/thank-you');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка при отправке результатов.');
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <Loader2 className="w-8 h-8 animate-spin text-brand" />
      </div>
    );
  }

  if (error && questions.length === 0) {
    return (
      <div className="flex-1 flex items-center justify-center p-4">
        <div className="card text-center max-w-md w-full border-red-900/50">
          <p className="text-red-500 mb-4">{error}</p>
          <button onClick={() => window.location.reload()} className="btn-primary w-full">
            Обновить страницу
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col bg-surface">
      <div className="hero-section !min-h-[220px] !items-center">
        <img src="/hero/bg2.png" alt="" className="hero-bg-dots opacity-60" aria-hidden="true" />
        <div className="hero-text-block !py-0 !pb-8 !pt-28">
          <p className="text-brand-hover text-xs font-bold uppercase tracking-widest mb-2">Sapiens Solutions</p>
          <h1 className="hero-title !text-2xl sm:!text-3xl !max-w-none">Опрос вовлечённости</h1>
          {roundCode && (
            <p className="hero-period">{formatRoundCode(roundCode).replace(' ', ' · ')}</p>
          )}
          <p className="hero-subtitle !mt-3 !text-sm">
            Gallup Q12 для Data Engineering + eNPS. 13 вопросов. Ответы анонимны.
          </p>
        </div>
      </div>

      <div className="pb-10 px-4 sm:px-6 lg:px-8 max-w-3xl mx-auto w-full -mt-2">
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 p-4 rounded-md mb-8">{error}</div>
        )}

        <form onSubmit={handleSubmit} className="space-y-8">
          <div id="about-you" className="card space-y-4">
            <h2 className="text-xl font-semibold mb-2">О вас</h2>
            <p className="text-sm text-ink-muted mb-4">
              Обязательные поля для агрегированных срезов по направлению Data Engineering. Данные не
              идентифицируют вас лично.
            </p>

            <div className="grid sm:grid-cols-2 gap-4">
              <div className="sm:col-span-2">
                <label className="block text-sm font-medium mb-1">Направление</label>
                <div className="w-full bg-page border border-line rounded py-2 px-3 text-ink">
                  {SURVEY_DIRECTION}
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">
                  Грейд <span className="text-red-500">*</span>
                </label>
                <select
                  required
                  className="w-full bg-white border border-line rounded py-2 px-3 focus:outline-none focus:border-brand focus:ring-1 focus:ring-brand"
                  value={metadata.gradeBand}
                  onChange={(e) => setMetadata((prev) => ({ ...prev, gradeBand: e.target.value }))}
                >
                  <option value="">Выберите грейд</option>
                  {gradeOptions.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.labelRu}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">
                  Роль <span className="text-red-500">*</span>
                </label>
                <select
                  required
                  className="w-full bg-white border border-line rounded py-2 px-3 focus:outline-none focus:border-brand focus:ring-1 focus:ring-brand"
                  value={metadata.role}
                  onChange={(e) => setMetadata((prev) => ({ ...prev, role: e.target.value }))}
                >
                  <option value="">Выберите роль</option>
                  {SURVEY_ROLES.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.labelRu}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>

          <div className="space-y-6">
            {questions.map((q) => (
              <div key={q.id} id={`question-${q.id}`} className="card">
                <h3 className="text-lg font-medium mb-4">
                  <span className="text-brand mr-2 font-bold">{q.code}.</span>
                  {q.textRu}
                </h3>
                <QuestionScale question={q} value={answers[q.id]} onChange={(val) => handleAnswer(q.id, val)} />
              </div>
            ))}
          </div>

          <div className="pt-4 flex justify-end">
            <button
              type="submit"
              disabled={submitting}
              className="btn-primary w-full sm:w-auto flex items-center justify-center gap-2"
            >
              {submitting && <Loader2 className="w-4 h-4 animate-spin" />}
              Отправить ответы
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
