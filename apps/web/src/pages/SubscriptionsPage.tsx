import { Link } from 'react-router-dom';
import { Bookmark } from 'lucide-react';

export function SubscriptionsPage() {
  return <div className="page">
    <div className="page-header"><div><span className="eyebrow">ПОДПИСКИ</span><h1>Мои подписки</h1>
      <p>Следите за интересными обсуждениями.</p></div></div>
    <section className="panel placeholder"><Bookmark size={35} /><b>Список подписок готовится</b>
      <p>Подписаться можно кнопкой закладки у темы. Новые ответы появятся в уведомлениях.</p>
      <Link to="/notifications" className="secondary-btn">Открыть уведомления</Link>
      <Link to="/forum" className="back">Найти обсуждение</Link>
    </section>
  </div>;
}
