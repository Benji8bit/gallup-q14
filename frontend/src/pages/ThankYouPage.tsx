import { Link } from 'react-router-dom';
import { CheckCircle } from 'lucide-react';

export const ThankYouPage = () => {
  return (
    <div className="flex-1 flex items-center justify-center pt-24 pb-20 px-4 bg-surface">
      <div className="card max-w-md w-full text-center py-12">
        <div className="flex justify-center mb-6">
          <CheckCircle className="w-20 h-20 text-brand" />
        </div>
        <h1 className="text-3xl font-bold mb-4 uppercase tracking-wide">Спасибо за участие!</h1>
        <p className="text-ink-muted mb-8 leading-relaxed">
          Ваши ответы успешно сохранены. Ваше мнение очень важно для нас и поможет сделать нашу компанию лучше.
        </p>
        <Link to="/" className="btn-primary inline-block">
          Вернуться на главную
        </Link>
      </div>
    </div>
  );
};
