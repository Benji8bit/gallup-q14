import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Lock, Loader2 } from 'lucide-react';
import { setAdminToken, verifyAdminPassword } from '../lib/api';

export const AdminLoginPage = () => {
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    const ok = await verifyAdminPassword(password);
    if (ok) {
      setAdminToken(password);
      navigate('/admin/dashboard');
    } else {
      setError('Неверный пароль');
    }
    setLoading(false);
  };

  return (
    <div className="flex-1 flex items-center justify-center pt-24 pb-20 px-4 bg-surface min-h-[70vh]">
      <div className="card max-w-md w-full shadow-md">
        <div className="flex justify-center mb-6">
          <div className="w-16 h-16 bg-brand-light rounded-full flex items-center justify-center">
            <Lock className="w-8 h-8 text-brand" />
          </div>
        </div>
        <h1 className="text-2xl font-bold text-center mb-2 uppercase tracking-wide">Вход для HR</h1>
        <p className="text-ink-muted text-center text-sm mb-8">
          Доступ к дашборду вовлечённости Sapiens Solutions
        </p>

        <form onSubmit={handleLogin} className="space-y-6">
          <div>
            <label className="block text-sm font-bold mb-2 uppercase tracking-wide">Пароль доступа</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full bg-white border border-line rounded py-3 px-4 focus:outline-none focus:border-brand focus:ring-1 focus:ring-brand"
              placeholder="Введите пароль"
              required
            />
            {error && <p className="text-red-600 text-sm mt-2">{error}</p>}
          </div>

          <button type="submit" disabled={loading} className="btn-primary w-full flex items-center justify-center gap-2">
            {loading && <Loader2 className="w-4 h-4 animate-spin" />}
            Войти
          </button>
        </form>
      </div>
    </div>
  );
};
