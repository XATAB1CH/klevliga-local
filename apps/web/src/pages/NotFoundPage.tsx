import { Link } from 'react-router-dom';
export function NotFoundPage() {
  return <div className="page"><section className="panel query-message"><h1>Страница не найдена</h1>
    <p>Проверьте адрес или вернитесь к обсуждениям сообщества.</p>
    <Link className="secondary-btn" to="/">На главную</Link>
  </section></div>;
}
