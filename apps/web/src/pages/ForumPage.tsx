import { Link, useLocation } from 'react-router-dom';
import { ChevronRight, MessageSquare, Plus } from 'lucide-react';
import { useCategories, useTopics } from '../api/hooks';
import type { Topic } from '../api/types';
import { PostCard } from '../components/PostCard';

interface Props {
  onOpen: (topic: Topic) => void;
  onFollow: (topic: Topic) => void;
  onNew: () => void;
}

export function ForumPage({ onOpen, onFollow, onNew }: Props) {
  const params = new URLSearchParams(useLocation().search);
  const catSlug = params.get('category');
  const { data: categories = [] } = useCategories();
  const current = categories.find((c) => c.slug === catSlug);
  const { data: topics } = useTopics({ category_id: current?.id, limit: 50 });

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <span className="eyebrow">СООБЩЕСТВО / ФОРУМ</span>
          <h1>{current?.name ?? 'Разделы форума'}</h1>
          <p>Темы, обсуждения и отчёты рыболовов.</p>
        </div>
        <button className="primary-btn" onClick={onNew}><Plus size={17} /> Создать тему</button>
      </div>
      <div className="grid-categories">
        {categories.map((c) => (
          <Link key={c.id} to={`/forum?category=${encodeURIComponent(c.slug)}`} className={catSlug === c.slug ? 'current' : ''}>
            <MessageSquare size={20} />{c.name}<ChevronRight size={16} />
          </Link>
        ))}
      </div>
      <div className="panel">
        <div className="panel-head">Последние обсуждения</div>
        {(topics?.items ?? []).map((t) => <PostCard key={t.id} topic={t} onOpen={onOpen} onFollow={onFollow} />)}
        {topics && topics.items.length === 0 && <div className="empty">В этом разделе пока нет тем.</div>}
      </div>
    </div>
  );
}
