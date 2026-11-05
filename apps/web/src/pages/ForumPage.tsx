import { Link, useLocation } from 'react-router-dom';
import { ChevronRight, MessageSquare, Plus } from 'lucide-react';
import { useCategories, useInfiniteTopics } from '../api/hooks';
import type { Topic } from '../api/types';
import { PostCard } from '../components/PostCard';
import { ErrorNotice, Loading, LoadMore } from '../components/QueryState';

interface Props { onOpen: (topic: Topic) => void; onFollow: (topic: Topic) => void; onNew: () => void }
function TopicList({ categoryId, onOpen, onFollow }: Pick<Props, 'onOpen' | 'onFollow'> & { categoryId?: string }) {
  const topics = useInfiniteTopics({ category_id: categoryId, limit: 20 });
  const items = topics.data?.pages.flatMap((page) => page.items) ?? [];
  return <div className="panel">
    <div className="panel-head">Последние обсуждения</div>
    {topics.isPending && <Loading />}
    {topics.isError && <ErrorNotice error={topics.error} retry={topics.refetch} />}
    {items.map((topic) => <PostCard key={topic.id} topic={topic} onOpen={onOpen} onFollow={onFollow} />)}
    {!topics.isPending && !topics.isError && !items.length && <div className="empty">В этом разделе пока нет тем.</div>}
    <LoadMore {...topics} />
  </div>;
}
export function ForumPage({ onOpen, onFollow, onNew }: Props) {
  const catSlug = new URLSearchParams(useLocation().search).get('category');
  const categories = useCategories();
  const current = categories.data?.find((category) => category.slug === catSlug);
  return <div className="page">
    <div className="page-header"><div>
      <span className="eyebrow">СООБЩЕСТВО / ФОРУМ</span>
      <h1>{current?.name ?? 'Разделы форума'}</h1><p>Темы, обсуждения и отчёты рыболовов.</p>
    </div><button className="primary-btn" onClick={onNew}><Plus size={17} /> Создать тему</button></div>
    {categories.isPending ? <Loading text="Загрузка разделов…" /> : categories.isError ?
      <ErrorNotice error={categories.error} retry={categories.refetch} /> : <>
      <div className="grid-categories">{categories.data?.map((category) =>
        <Link key={category.id} to={'/forum?category=' + encodeURIComponent(category.slug)} className={catSlug === category.slug ? 'current' : ''}>
          <MessageSquare size={20} />{category.name}<ChevronRight size={16} />
        </Link>)}</div>
      {catSlug && !current ? <section className="panel query-message"><h2>Раздел не найден</h2>
        <Link className="back" to="/forum">Все разделы форума</Link></section> :
        <TopicList key={current?.id ?? 'all'} categoryId={current?.id} onOpen={onOpen} onFollow={onFollow} />}
    </>}
  </div>;
}
