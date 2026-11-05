import { Link } from 'react-router-dom';
import { Bell } from 'lucide-react';
import { useInfiniteNotifications, useMe } from '../api/hooks';
import { useAuth } from '../auth/AuthContext';
import { timeAgo } from '../lib/format';
import { ErrorNotice, Loading, LoadMore } from '../components/QueryState';

export function NotificationsPage({ onLogin }: { onLogin: () => void }) {
  const { isAuthenticated, isLoading, error } = useAuth();
  const me = useMe();
  const notifications = useInfiniteNotifications();
  const items = notifications.data?.pages.flatMap((page) => page.items) ?? [];
  return <div className="page">
    <div className="page-header"><div><span className="eyebrow">АКТИВНОСТЬ</span><h1>Уведомления</h1></div></div>
    {isLoading ? <Loading /> : error ? <ErrorNotice error={error} retry={me.refetch} /> : !isAuthenticated ?
      <section className="panel profile"><div className="profile-circle"><Bell size={40} /></div>
        <h2>Требуется вход</h2><p>Войдите, чтобы видеть уведомления.</p>
        <button className="primary-btn" onClick={onLogin}>Войти / Регистрация</button></section> :
      <div className="panel">
        {notifications.isPending && <Loading />}
        {notifications.isError && <ErrorNotice error={notifications.error} retry={notifications.refetch} />}
        {!notifications.isPending && !notifications.isError && !items.length && <div className="empty">Уведомлений пока нет.</div>}
        {items.map((notification) => <article key={notification.id} className="notification-item">
          {notification.target_url.startsWith('/') && !notification.target_url.startsWith('//') ?
            <Link to={notification.target_url} className={notification.read ? 'muted' : ''}><b>{notification.title}</b></Link> :
            <b className={notification.read ? 'muted' : ''}>{notification.title}</b>}
          <div className="muted tiny">{timeAgo(notification.created_at)}</div>
        </article>)}
        <LoadMore {...notifications} />
      </div>}
  </div>;
}
