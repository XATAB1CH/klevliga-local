import { BookOpen, Bookmark, Settings, UserRound } from 'lucide-react';
import { useAuth } from '../auth/AuthContext';

export function ProfilePage({ onLogin }: { onLogin: () => void }) {
  const { user, isAuthenticated, isLoading, logout } = useAuth();

  return (
    <div className="page">
      <div className="page-header">
        <div><span className="eyebrow">МОЙ ПРОФИЛЬ</span><h1>Профиль рыбака</h1><p>Личный кабинет, история публикаций и подписок.</p></div>
      </div>
      {isLoading ? (
        <div className="empty">Загрузка…</div>
      ) : !isAuthenticated ? (
        <section className="panel profile">
          <div className="profile-circle"><UserRound size={40} /></div>
          <h2>Вы не вошли</h2>
          <p>Войдите или зарегистрируйтесь, чтобы видеть профиль, подписки и уведомления.</p>
          <button className="primary-btn" onClick={onLogin}>Войти / Регистрация</button>
        </section>
      ) : (
        <section className="panel profile">
          <div className="profile-circle"><UserRound size={40} /></div>
          <h2>{user?.display_name}</h2>
          <p>@{user?.username}</p>
          {user?.bio && <p>{user.bio}</p>}
          {user?.region && <p className="muted tiny">Регион: {user.region}</p>}
          <p className="muted tiny">На КлёвЛиге с {user ? new Date(user.joined_at).toLocaleDateString('ru-RU') : ''}</p>
          <div className="profile-metrics">
            <span><BookOpen /> Публикации</span>
            <span><Bookmark /> Подписки</span>
            <span><Settings /> Настройки</span>
          </div>
          <button className="secondary-btn" style={{ marginTop: 22 }} onClick={() => void logout()}>Выйти</button>
        </section>
      )}
    </div>
  );
}
