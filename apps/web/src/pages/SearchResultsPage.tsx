import { Link, useLocation } from 'react-router-dom';
import { useSearch } from '../api/hooks';
import type { Topic } from '../api/types';
import { PostCard } from '../components/PostCard';
import { waterKindLabel } from '../lib/format';

interface Props {
  onOpen: (topic: Topic) => void;
  onFollow: (topic: Topic) => void;
}

export function SearchResultsPage({ onOpen, onFollow }: Props) {
  const q = new URLSearchParams(useLocation().search).get('q') ?? '';
  const { data, isLoading } = useSearch({ q });

  return (
    <div className="page">
      <div className="page-header">
        <div><span className="eyebrow">ПОИСК</span><h1>Результаты по запросу «{q}»</h1></div>
      </div>
      {isLoading && <div className="empty">Поиск…</div>}
      {data && (
        <>
          {data.topics.length > 0 && (
            <>
              <h2>Темы</h2>
              {data.topics.map((t) => <PostCard key={t.id} topic={t} onOpen={onOpen} onFollow={onFollow} />)}
            </>
          )}
          {data.reports.length > 0 && (
            <>
              <h2>Отчёты</h2>
              {data.reports.map((r) => (
                <div key={r.id} className="panel" style={{ marginBottom: 10, padding: 14 }}>
                  <b>{r.title}</b><p className="muted">{r.body}</p>
                </div>
              ))}
            </>
          )}
          {data.waters.length > 0 && (
            <>
              <h2>Водоёмы</h2>
              {data.waters.map((w) => (
                <Link key={w.id} to={`/waters/${w.id}`} className="secondary-btn" style={{ margin: '4px 4px 4px 0' }}>{w.name} · {waterKindLabel(w.kind)}</Link>
              ))}
            </>
          )}
          {data.users.length > 0 && (
            <>
              <h2>Пользователи</h2>
              {data.users.map((u) => (
                <span key={u.id} className="secondary-btn" style={{ margin: '4px 4px 4px 0' }}>{u.display_name} (@{u.username})</span>
              ))}
            </>
          )}
        </>
      )}
    </div>
  );
}
