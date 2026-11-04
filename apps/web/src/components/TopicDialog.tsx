import { useState } from 'react';
import { Eye, MessageSquare, Send, UserRound, X } from 'lucide-react';
import { usePosts, useReply } from '../api/hooks';
import type { Topic } from '../api/types';
import { fmtCount, kindLabel, timeAgo } from '../lib/format';
import { useAuth } from '../auth/AuthContext';

export function TopicDialog({ topic, onClose }: { topic: Topic; onClose: () => void }) {
  const { data: posts } = usePosts(topic.id);
  const { isAuthenticated } = useAuth();
  const reply = useReply(topic.id);
  const [body, setBody] = useState('');
  const [error, setError] = useState<string | null>(null);

  const submit = async () => {
    if (!body.trim()) return;
    setError(null);
    try {
      await reply.mutateAsync({ body: body.trim() });
      setBody('');
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка');
    }
  };

  return (
    <div className="overlay" onMouseDown={onClose} role="presentation">
      <div className="dialog" role="dialog" aria-modal="true" onMouseDown={(e) => e.stopPropagation()}>
        <div className="dialog-head">
          <span className="post-badge">{kindLabel(topic.type)}</span>
          <button className="icon-btn" onClick={onClose} aria-label="Закрыть"><X /></button>
        </div>
        <h2>{topic.title}</h2>
        <p className="muted">{topic.author_name} · {timeAgo(topic.created_at)} · {topic.category_name}</p>
        <p>{topic.body}</p>
        <div className="post-stats">
          <span><MessageSquare /> {topic.replies_count} ответов</span>
          <span><Eye /> {fmtCount(topic.views_count)} просмотров</span>
        </div>

        <div style={{ margin: '18px 0 0' }}>
          {posts?.items.map((p) => (
            <div key={p.id} style={{ borderTop: '1px solid #1a3946', padding: '12px 0' }}>
              <div className="post-meta"><UserRound size={13} />{p.author_name} <span>·</span>{timeAgo(p.created_at)}</div>
              <p style={{ margin: '8px 0 0', color: '#cfe0e7' }}>{p.body}</p>
            </div>
          ))}
        </div>

        {isAuthenticated ? (
          <div style={{ marginTop: '14px' }}>
            <textarea rows={3} value={body} onChange={(e) => setBody(e.target.value)} placeholder="Ваш ответ…" style={{ width: '100%', resize: 'vertical' }} />
            {error && <p className="muted">{error}</p>}
            <button className="primary-btn" disabled={!body.trim() || reply.isPending} onClick={submit}><Send size={17} /> Ответить</button>
          </div>
        ) : (
          <p className="muted" style={{ marginTop: '14px' }}>Войдите, чтобы ответить в теме.</p>
        )}
      </div>
    </div>
  );
}
