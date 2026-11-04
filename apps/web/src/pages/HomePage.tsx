import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { ArrowRight, Camera, ChevronDown, Flame, Heart, Map as MapIcon, MapPin, MessageCircle, Plus, Search, Waves } from 'lucide-react';
import { useFeed, useWaters } from '../api/hooks';
import type { Topic } from '../api/types';
import { PostCard } from '../components/PostCard';
import { waterKindLabel } from '../lib/format';

const pills = ['Все', 'Отчёты', 'Популярные'] as const;

interface Props {
  onOpen: (topic: Topic) => void;
  onFollow: (topic: Topic) => void;
  onNew: () => void;
}

export function HomePage({ onOpen, onFollow, onNew }: Props) {
  const [tab, setTab] = useState<(typeof pills)[number]>('Все');
  const [sort, setSort] = useState<'новые' | 'популярные'>('новые');
  const { data: watersPage } = useWaters({ limit: 4 });

  const kind = tab === 'Отчёты' ? 'report' : 'all';
  const { data: feed, isLoading } = useFeed({ kind, limit: 50 });

  const items = useMemo(() => {
    const src = feed?.items ?? [];
    const popular = sort === 'популярные' || tab === 'Популярные';
    return popular
      ? [...src].sort((a, b) => b.topic.replies_count - a.topic.replies_count)
      : src;
  }, [feed, sort, tab]);

  return (
    <div className="three-col">
      <div className="content-col">
        <section className="hero">
          <div className="hero-overlay" />
          <div className="hero-inner">
            <div className="eyebrow"><span /> ЖИВОЕ РЫБОЛОВНОЕ СООБЩЕСТВО</div>
            <h1>Рыбалка начинается здесь.</h1>
            <p>Свежие отчёты, советы и разговоры о главном —<br /> на одной волне.</p>
            <button className="primary-btn" onClick={() => document.getElementById('feed')?.scrollIntoView({ behavior: 'smooth' })}>Перейти к обсуждениям <ArrowRight size={18} /></button>
          </div>
        </section>

        <section className="feed" id="feed">
          <div className="tabs-row">
            <div className="tabs-scroll">
              {pills.map((x) => <button key={x} className={tab === x ? 'selected' : ''} onClick={() => setTab(x)}>{x}</button>)}
            </div>
            <button className="sort-control" onClick={() => setSort((x) => (x === 'новые' ? 'популярные' : 'новые'))}>Сначала {sort}<ChevronDown size={16} /></button>
          </div>
          {isLoading && <div className="empty">Загрузка…</div>}
          {!isLoading && items.length === 0 && <div className="empty"><Search />Пока нет публикаций. Создайте первую тему.</div>}
          {items.map((it) => <PostCard key={it.topic.id} topic={it.topic} onOpen={onOpen} onFollow={onFollow} />)}
          <div className="feed-bottom"><span>Лента подключена к Go API · /api/v1</span><button onClick={onNew}>Новая тема <Plus size={16} /></button></div>
        </section>
      </div>

      <aside className="right-col">
        <section className="widget">
          <div className="widget-title"><h3><Camera size={19} /> Фото дня</h3><Link to="/profile">Все фото <ArrowRight size={14} /></Link></div>
          <img src="/images/sunset-photo.jpg" className="photo-day" alt="Закат на озере" />
          <b className="photo-name">Вечер на озере</b>
          <div className="muted tiny">Фото сообщества · Авторский снимок</div>
          <div className="mini-social"><Heart size={16} /> 256 <MessageCircle size={16} /> 18 <MapPin size={15} /> Московская область</div>
        </section>

        <section className="widget">
          <div className="widget-title"><h3><Flame size={19} className="hot" /> Популярные темы</h3><Link to="/forum">Все темы <ArrowRight size={14} /></Link></div>
          <div className="top-list">
            {[...(feed?.items ?? [])].sort((a, b) => b.topic.replies_count - a.topic.replies_count).slice(0, 5).map((it, i) => (
              <button key={it.topic.id} onClick={() => onOpen(it.topic)}>
                <span>{i + 1}</span>
                <div>{it.topic.title}<small>{it.topic.replies_count} ответов</small></div>
              </button>
            ))}
          </div>
        </section>

        <section className="widget">
          <div className="widget-title"><h3><MapIcon size={19} /> Популярные водоёмы</h3><Link to="/waters">Все водоёмы <ArrowRight size={14} /></Link></div>
          <div className="water-grid">
            {watersPage?.items.map((w) => (
              <Link to={`/waters/${w.id}`} key={w.id}><Waves size={16} /><b>{w.name}</b><small>{waterKindLabel(w.kind)}</small></Link>
            ))}
          </div>
        </section>
      </aside>
    </div>
  );
}
