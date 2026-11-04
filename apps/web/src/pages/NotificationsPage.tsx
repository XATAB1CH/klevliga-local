import { Bell } from 'lucide-react';
import { useNotifications } from '../api/hooks';
import { useAuth } from '../auth/AuthContext';
import { timeAgo } from '../lib/format';

export function NotificationsPage({ onLogin }: { onLogin: () => void }) {
  const { isAuthenticated } = useAuth();
  const { data, isLoading } = useNotifications();

  if (!isAuthenticated) {
    return (
      <div className="page">
        <div className="page-header"><div><span className="eyebrow">АКТИВНОСТЬ</span><h1>Уведомления</h1></div></div>
        <section className="panel profile">
          <div className="profile-circle"><Bell size={40} /></div>
          <h2>Требуется вход</h2>
          <p>Войдите, чтобы видеть уведомления.</p>
          <button className="primary-btn" onClick={onLogin}>Войти / Регистрация</button>
        </section>
      </div>
    );
  }

  return (
    <div className="page">
      <div className="page-header"><div><span className="eyebrow">АКТИВНОСТЬ</span><h1>Уведомления</h1></div></div>
      <div className="panel">
        {isLoading && <div className="empty">Загрузка…</div>}
        {!isLoading && (data?.items ?? []).length === 0 && <div className="empty">Уведомлений пока нет.</div>}
        {(data?.items ?? []).map((n) => (
          <div key={n.id} style={{ padding: '14px 18px', borderBottom: '1px solid #1a3946' }}>
            <b style={{ color: n.read ? '#8ba3b0' : '#eaf3f6' }}>{n.title}</b>
            <div className="muted tiny">{timeAgo(n.created_at)}</div>
          </div>
        ))}
      </div>
    </div>
  );
}
