import { ShieldCheck } from 'lucide-react';

export function SubscriptionsPage() {
  return (
    <div className="page">
      <div className="page-header">
        <div><span className="eyebrow">ПОДПИСКИ</span><h1>Мои подписки</h1><p>Темы и водоёмы, на которые вы подписаны.</p></div>
      </div>
      <div className="panel placeholder">
        <ShieldCheck size={35} />
        <b>Список подписок появится позже</b>
        <p className="muted">В API пока нет эндпоинта списка подписок (реализованы только подписаться/отписаться).</p>
      </div>
    </div>
  );
}
