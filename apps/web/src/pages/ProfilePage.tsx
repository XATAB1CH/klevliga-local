import { useRef, useState, type ChangeEvent } from 'react';
import { Link } from 'react-router-dom';
import { Bell, Bookmark, CalendarDays, Camera, MapPin, UserRound } from 'lucide-react';
import { useAuth } from '../auth/AuthContext';
import { useMe, useUploadAvatar } from '../api/hooks';
import { ErrorNotice, Loading, errorMessage } from '../components/QueryState';

const maxAvatarBytes = 5 * 1024 * 1024;

export function ProfilePage({ onLogin }: { onLogin: () => void }) {
  const { user, isAuthenticated, isLoading, error, isLoggingOut, logout } = useAuth();
  const me = useMe();
  const upload = useUploadAvatar();
  const fileInput = useRef<HTMLInputElement>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  const signOut = async () => {
    setActionError(null);
    try { await logout(); } catch (e) { setActionError(errorMessage(e, 'Не удалось выйти. Попробуйте ещё раз.')); }
  };
  const selectAvatar = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file) return;
    setActionError(null);
    setNotice(null);
    if (!['image/jpeg', 'image/png'].includes(file.type)) {
      setActionError('Выберите изображение JPEG или PNG.'); return;
    }
    if (file.size > maxAvatarBytes) {
      setActionError('Фото слишком большое. Выберите файл до 5 МБ.'); return;
    }
    try {
      await upload.mutateAsync(file);
      setNotice('Аватар обновлён.');
    } catch (e) { setActionError(errorMessage(e, 'Не удалось загрузить фото. Попробуйте ещё раз.')); }
  };

  return <div className="page profile-page">
    <div className="page-header"><div><span className="eyebrow">МОЙ ПРОФИЛЬ</span><h1>Профиль рыбака</h1><p>Ваше место в сообществе КлёвЛиги.</p></div></div>
    {isLoading ? <Loading /> : error ? <ErrorNotice error={error} retry={me.refetch} /> :
      <section className="panel profile-shell">
        <div className="profile-cover"><span>КЛЁВЛИГА · ГЛУБОКАЯ ВОДА</span></div>
        <div className="profile-content">
          <div className="profile-identity">
            <div className="profile-avatar" aria-label={user ? `Аватар ${user.display_name}` : 'Аватар профиля'}>
              {user?.avatar_url ? <img src={user.avatar_url} alt="" /> : <UserRound size={48} aria-hidden="true" />}
            </div>
            <div className="profile-heading">
              <span className="profile-kicker">{isAuthenticated ? 'УЧАСТНИК СООБЩЕСТВА' : 'ПРИСОЕДИНЯЙТЕСЬ К СООБЩЕСТВУ'}</span>
              <h2>{isAuthenticated ? user?.display_name : 'Ваша история начинается здесь'}</h2>
              <p>{isAuthenticated ? `@${user?.username}` : 'Создайте профиль, делитесь уловом и находите единомышленников.'}</p>
            </div>
          </div>

          {isAuthenticated && user ? <>
            <div className="profile-details">
              <span><CalendarDays size={17} aria-hidden="true" /> На КлёвЛиге с {new Date(user.joined_at).toLocaleDateString('ru-RU')}</span>
              {user.region && <span><MapPin size={17} aria-hidden="true" /> {user.region}</span>}
            </div>
            {user.bio && <p className="profile-bio user-text">{user.bio}</p>}
            <div className="profile-actions">
              <input ref={fileInput} className="visually-hidden" type="file" accept="image/jpeg,image/png"
                aria-label="Выберите аватар" onChange={(event) => { void selectAvatar(event); }} />
              <button className="primary-btn" type="button" disabled={upload.isPending}
                onClick={() => fileInput.current?.click()}>
                <Camera size={18} aria-hidden="true" /> {upload.isPending ? 'Загружаем фото…' : user.avatar_url ? 'Сменить аватар' : 'Загрузить аватар'}
              </button>
              <span className="profile-upload-hint">JPEG или PNG, до 5 МБ. Фото будет обрезано по центру.</span>
            </div>
            {notice && <p className="success-message" role="status">{notice}</p>}
            {actionError && <p className="form-error" role="alert">{actionError}</p>}
            <div className="profile-links">
              <Link to="/subscriptions"><Bookmark size={21} aria-hidden="true" /><span><strong>Мои подписки</strong><small>Темы и водоёмы, за которыми вы следите</small></span></Link>
              <Link to="/notifications"><Bell size={21} aria-hidden="true" /><span><strong>Уведомления</strong><small>Ответы и события сообщества</small></span></Link>
            </div>
            <div className="profile-footer"><button type="button" className="profile-signout" disabled={isLoggingOut} onClick={() => { void signOut(); }}>
              {isLoggingOut ? 'Выходим…' : 'Выйти из аккаунта'}
            </button></div>
          </> : <div className="profile-actions"><button className="primary-btn" onClick={onLogin}>Войти / Регистрация</button></div>}
        </div>
      </section>}
  </div>;
}
