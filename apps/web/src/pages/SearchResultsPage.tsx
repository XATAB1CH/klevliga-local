import { Link, useLocation } from 'react-router-dom';
import { useSearch } from '../api/hooks';
import type { Topic } from '../api/types';
import { PostCard } from '../components/PostCard';
import { ErrorNotice, Loading } from '../components/QueryState';
import { waterKindLabel } from '../lib/format';

interface Props { onOpen: (topic: Topic) => void; onFollow: (topic: Topic) => void }
export function SearchResultsPage({ onOpen, onFollow }: Props) {
  const q = new URLSearchParams(useLocation().search).get('q')?.trim() ?? '';
  const search = useSearch({ q });
  const data = search.data;
  const empty = data && !data.topics.length && !data.reports.length && !data.waters.length && !data.users.length;
  return <div className="page">
    <div className="page-header"><div><span className="eyebrow">ПОИСК</span><h1>Результаты по запросу «{q}»</h1></div></div>
    {q.length < 2 ? <p className="query-message">Введите в поиске не меньше двух символов.</p> : <>
      {search.isPending && <Loading text="Поиск…" />}
      {search.isError && <ErrorNotice error={search.error} retry={search.refetch} />}
      {!search.isError && empty && <div className="empty">Ничего не найдено. Попробуйте другой запрос.</div>}
      {data && <>
        {data.topics.length > 0 && <><h2>Темы</h2>{data.topics.map((topic) =>
          <PostCard key={topic.id} topic={topic} onOpen={onOpen} onFollow={onFollow} />)}</>}
        {data.reports.length > 0 && <><h2>Отчёты</h2>{data.reports.map((report) =>
          <Link key={report.id} to={'/forum/topics/' + report.topic_id} className="panel report-link">
            <b>{report.title}</b><p>{report.body}</p>
          </Link>)}</>}
        {data.waters.length > 0 && <><h2>Водоёмы</h2><div className="search-links">{data.waters.map((water) =>
          <Link key={water.id} to={'/waters/' + water.id} className="secondary-btn">{water.name} · {waterKindLabel(water.kind)}</Link>)}</div></>}
        {data.users.length > 0 && <><h2>Пользователи</h2><div className="search-links">{data.users.map((user) =>
          <span key={user.id} className="search-user">{user.display_name} (@{user.username})</span>)}</div></>}
      </>}
    </>}
  </div>;
}
