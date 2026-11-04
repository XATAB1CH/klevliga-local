import { Bookmark, Eye, MessageSquare, UserRound } from 'lucide-react';
import type { Topic } from '../api/types';
import { fmtCount, kindLabel, timeAgo } from '../lib/format';

interface Props {
  topic: Topic;
  onOpen: (topic: Topic) => void;
  onFollow?: (topic: Topic) => void;
}

export function PostCard({ topic, onOpen, onFollow }: Props) {
  const badge = kindLabel(topic.type);

  return (
    <article className="post-card">
      <button className="post-avatar" onClick={() => onOpen(topic)} aria-label={`Открыть ${topic.title}`}>
        {(topic.author_name || '?').slice(0, 1).toUpperCase()}
      </button>
      <div className="post-content">
        <div className="post-heading">
          <span className={`post-badge badge-${badge}`}>{badge}</span>
          <button className="post-name" onClick={() => onOpen(topic)}>{topic.title}</button>
        </div>
        <div className="post-meta">
          <UserRound size={13} />{topic.author_name} <span>·</span>{timeAgo(topic.last_activity_at)}<span>·</span>{topic.category_name}
        </div>
        <p className="post-desc">{topic.body}</p>
      </div>
      <div className="post-stats">
        <span><MessageSquare size={17} /> {topic.replies_count} <small>ответов</small></span>
        <span><Eye size={17} /> {fmtCount(topic.views_count)} <small>просмотров</small></span>
        {onFollow && <button className="post-saved" aria-label="Подписаться" onClick={() => onFollow(topic)}><Bookmark size={18} /></button>}
      </div>
    </article>
  );
}
