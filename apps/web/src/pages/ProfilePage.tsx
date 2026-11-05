import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Bookmark, UserRound } from 'lucide-react';
import { useAuth } from '../auth/AuthContext';
import { useMe } from '../api/hooks';
import { ErrorNotice, Loading, errorMessage } from '../components/QueryState';

export function ProfilePage({ onLogin }: { onLogin: () => void }) {
  const { user, isAuthenticated, isLoading, error, isLoggingOut, logout } = useAuth();
  const me = useMe();
  const [logoutError, setLogoutError] = useState<string | null>(null);
  const signOut = async () => {
    setLogoutError(null);
    try { await logout(); } catch (e) { setLogoutError(errorMessage(e, 'Не удалось выйти. Попробуйте ещё раз.')); }
  };
  return <div className="page">
    <div className="page-header"><div><span className="eyebrow">МОЙ ПРОФИЛЬ</span><h1>Профиль рыбака</h1><p>Ваш профиль и подписки.</p></div></div>
    {isLoading ? <Loading /> : error ? <ErrorNotice error={error} retry={me.refetch} /> : !isAuthenticated ?
      <section className="panel profile">
        <div className="profile-circle"><UserRound size={40} /></div><h2>Вы не вошли</h2>
        <p>Войдите или зарегистрируйтесь, чтобы видеть профиль, подписки и уведомления.</p>
        <button className="primary-btn" onClick={onLogin}>Войти / Регистрация</button>
      </section> :
      <section className="panel profile">
        <div className="profile-circle"><UserRound size={40} /></div>
        <h2>{user?.display_name}</h2><p>@{user?.username}</p>
        {user?.bio && <p className="user-text">{user.bio}</p>}
        {user?.region && <p className="muted tiny">Регион: {user.region}</p>}
        <p className="muted tiny">На КлёвЛиге с {user ? new Date(user.joined_at).toLocaleDateString('ru-RU') : ''}</p>
        <Link to="/subscriptions" className="secondary-btn"><Bookmark size={16} /> Мои подписки</Link>
        {logoutError && <p className="form-error" role="alert">{logoutError}</p>}
        <div><button className="secondary-btn" disabled={isLoggingOut} onClick={() => { void signOut(); }}>
          {isLoggingOut ? 'Выход…' : 'Выйти'}
        </button></div>
      </section>}
  </div>;
}
