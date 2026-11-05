import { useState, type FormEvent } from 'react';
import { Eye, MessageSquare, Send, UserRound } from 'lucide-react';
import { useInfinitePosts, useReply } from '../api/hooks';
import type { Post, Topic } from '../api/types';
import { fmtCount, kindLabel, timeAgo } from '../lib/format';
import { useAuth } from '../auth/AuthContext';
import { Dialog } from './Dialog';
import { ErrorNotice, Loading, LoadMore, errorMessage } from './QueryState';

export function TopicContent({ topic }: { topic: Topic }) {
  const posts = useInfinitePosts(topic.id);
  const { isAuthenticated } = useAuth();
  const reply = useReply(topic.id);
  const [body, setBody] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [sent, setSent] = useState(false);
  const [sentPost, setSentPost] = useState<Post | null>(null);
  const items = posts.data?.pages.flatMap((page) => page.items) ?? [];
  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (!body.trim() || reply.isPending) return;
    setError(null); setSent(false);
    try { const post = await reply.mutateAsync({ body: body.trim() }); setSentPost(post); setBody(''); setSent(true); }
    catch (e) { setError(errorMessage(e, 'Не удалось отправить ответ. Текст сохранён в форме.')); }
  };
  return <>
    <span className="post-badge">{kindLabel(topic.type)}</span>
    <p className="muted">{topic.author_name} · {timeAgo(topic.created_at)} · {topic.category_name}</p>
    <p className="user-text">{topic.body}</p>
    <div className="post-stats">
      <span><MessageSquare /> {topic.replies_count} ответов</span><span><Eye /> {fmtCount(topic.views_count)} просмотров</span>
    </div>
    <section className="topic-replies" aria-label="Ответы в теме">
      <h2>Ответы</h2>
      {posts.isPending && <Loading text="Загрузка ответов…" />}
      {posts.isError && <ErrorNotice error={posts.error} retry={posts.refetch} />}
      {!posts.isPending && !posts.isError && items.length === 0 && <p className="muted">Ответов пока нет. Начните обсуждение.</p>}
      {items.map((post) => <article key={post.id} className="topic-reply">
        <div className="post-meta"><UserRound size={13} />{post.author_name} <span>·</span>{timeAgo(post.created_at)}</div>
        <p className="user-text">{post.body}</p>
      </article>)}
      <LoadMore {...posts} />
      {sentPost && !items.some((post) => post.id === sentPost.id) && <article className="topic-reply sent-reply">
        <h3>Ваш опубликованный ответ</h3><p className="user-text">{sentPost.body}</p>
      </article>}
    </section>
    {topic.locked ? <p className="query-message">Обсуждение закрыто для новых ответов.</p> :
      isAuthenticated ? <form className="reply-form" onSubmit={(event) => { void submit(event); }}>
        <label>Ваш ответ<textarea required maxLength={20000} rows={3} value={body}
          onChange={(e) => { setBody(e.target.value); setSent(false); }} placeholder="Поделитесь опытом…" /></label>
        {error && <p className="form-error" role="alert">{error}</p>}
        {sent && <p className="success-message" role="status">Ответ опубликован.</p>}
        <button type="submit" className="primary-btn" disabled={!body.trim() || reply.isPending}>
          <Send size={17} /> {reply.isPending ? 'Отправка…' : 'Ответить'}
        </button>
      </form> : <p className="muted">Войдите через профиль, чтобы ответить в теме.</p>}
  </>;
}
export function TopicDialog({ topic, onClose }: { topic: Topic; onClose: () => void }) {
  return <Dialog title={topic.title} onClose={onClose}><TopicContent key={topic.id} topic={topic} /></Dialog>;
}
