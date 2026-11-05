import { useState, type FormEvent } from 'react';
import { useAuth } from '../auth/AuthContext';
import { Dialog } from './Dialog';
import { errorMessage } from './QueryState';

export function AuthDialog({ onClose }: { onClose: () => void }) {
  const { login, register } = useAuth();
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (busy) return;
    setError(null); setBusy(true);
    try {
      if (mode === 'login') { await login(email.trim(), password); onClose(); }
      else {
        await register(email.trim(), username.trim(), password);
        setMode('login'); setPassword('');
        setNotice('Аккаунт создан. Войдите с вашим email и паролем.');
      }
    } catch (e) { setError(errorMessage(e, 'Не удалось выполнить действие. Проверьте соединение и попробуйте ещё раз.')); }
    finally { setBusy(false); }
  };
  return <Dialog title={mode === 'login' ? 'Вход' : 'Регистрация'} onClose={onClose}>
    <div className="choices" aria-label="Вход или регистрация">
      <button type="button" className={mode === 'login' ? 'choice active' : 'choice'} disabled={busy}
        onClick={() => { setMode('login'); setError(null); setNotice(null); }}>Вход</button>
      <button type="button" className={mode === 'register' ? 'choice active' : 'choice'} disabled={busy}
        onClick={() => { setMode('register'); setError(null); setNotice(null); }}>Регистрация</button>
    </div>
    {notice && <p className="success-message" role="status">{notice}</p>}
    <form onSubmit={(event) => { void submit(event); }}><fieldset disabled={busy}>
      <label>Email<input data-autofocus type="email" required maxLength={254} autoComplete="email" value={email}
        onChange={(e) => setEmail(e.target.value)} placeholder="you@example.com" /></label>
      {mode === 'register' && <label>Ник (3–32 символа)<input required minLength={3} maxLength={32} autoComplete="username"
        value={username} onChange={(e) => setUsername(e.target.value)} placeholder="Рыболов" /></label>}
      <label>{mode === 'register' ? 'Пароль (12–128 символов)' : 'Пароль'}<input type="password" required
        minLength={mode === 'register' ? 12 : undefined} maxLength={128}
        autoComplete={mode === 'register' ? 'new-password' : 'current-password'} value={password}
        onChange={(e) => setPassword(e.target.value)} placeholder="••••••••" /></label>
      {error && <p className="form-error" role="alert">{error}</p>}
      <button type="submit" className="primary-btn" disabled={busy}>
        {busy ? 'Подождите…' : mode === 'login' ? 'Войти' : 'Создать аккаунт'}
      </button>
    </fieldset></form>
  </Dialog>;
}
