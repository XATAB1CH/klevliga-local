import { useState } from 'react';
import { X } from 'lucide-react';
import { ApiError } from '../api/client';
import { useAuth } from '../auth/AuthContext';

export function AuthDialog({ onClose }: { onClose: () => void }) {
  const { login, register } = useAuth();
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const submit = async () => {
    setError(null);
    setBusy(true);
    try {
      if (mode === 'login') await login(email, password);
      else await register(email, username, password);
      onClose();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Не удалось выполнить действие');
    } finally {
      setBusy(false);
    }
  };

  const canSubmit = !!email && !!password && (mode === 'login' || username.length >= 3);

  return (
    <div className="overlay" onMouseDown={onClose} role="presentation">
      <div className="dialog" role="dialog" aria-modal="true" onMouseDown={(e) => e.stopPropagation()}>
        <div className="dialog-head">
          <h2>{mode === 'login' ? 'Вход' : 'Регистрация'}</h2>
          <button className="icon-btn" onClick={onClose} aria-label="Закрыть"><X /></button>
        </div>
        <div className="choices">
          <button className={mode === 'login' ? 'choice active' : 'choice'} onClick={() => { setMode('login'); setError(null); }}>Вход</button>
          <button className={mode === 'register' ? 'choice active' : 'choice'} onClick={() => { setMode('register'); setError(null); }}>Регистрация</button>
        </div>
        <label>Email<input type="email" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="you@example.com" /></label>
        {mode === 'register' && <label>Ник (от 3 символов)<input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="Рыболов" /></label>}
        <label>Пароль (от 12 символов)<input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="••••••••" /></label>
        {error && <p className="muted">{error}</p>}
        <button className="primary-btn" disabled={busy || !canSubmit} onClick={submit}>
          {mode === 'login' ? 'Войти' : 'Создать аккаунт'}
        </button>
      </div>
    </div>
  );
}
