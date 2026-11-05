import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { ArrowRight, Camera, ChevronDown, Flame, Map as MapIcon, Plus, Search, Waves } from 'lucide-react';
import { useInfiniteFeed, useWaters } from '../api/hooks';
import type { Topic } from '../api/types';
import { PostCard } from '../components/PostCard';
import { waterKindLabel } from '../lib/format';
import { ErrorNotice, Loading, LoadMore } from '../components/QueryState';

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
  const feed = useInfiniteFeed({ kind, limit: 20 });
  const loadedItems = useMemo(() => feed.data?.pages.flatMap((page) => page.items) ?? [], [feed.data]);

  const items = useMemo(() => {
    const src = loadedItems;
    const popular = sort === 'популярные' || tab === 'Популярные';
    return popular
      ? [...src].sort((a, b) => b.topic.replies_count - a.topic.replies_count)
      : src;
  }, [loadedItems, sort, tab]);

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
          {(sort === 'популярные' || tab === 'Популярные') && <p className="search-note">По числу ответов среди загруженных тем.</p>}
          {feed.isPending && <Loading />}
          {feed.isError && <ErrorNotice error={feed.error} retry={feed.refetch} />}
          {!feed.isPending && !feed.isError && items.length === 0 && <div className="empty"><Search />Пока нет публикаций. Создайте первую тему.</div>}
          {items.map((it) => <PostCard key={it.topic.id} topic={it.topic} onOpen={onOpen} onFollow={onFollow} />)}
          <LoadMore {...feed} />
          <div className="feed-bottom"><span>Обсуждения сообщества</span><button onClick={onNew}>Новая тема <Plus size={16} /></button></div>
        </section>
      </div>

      <aside className="right-col">
        <section className="widget">
          <div className="widget-title"><h3><Camera size={19} /> Фото дня</h3><Link to="/profile">Все фото <ArrowRight size={14} /></Link></div>
          <img src="/images/sunset-photo.jpg" className="photo-day" alt="Закат на озере" />
          <b className="photo-name">Вечер на озере</b>
          <div className="muted tiny">Фото сообщества · Авторский снимок</div>
          <div className="mini-social">Иллюстративное фото · место съёмки не указано</div>
        </section>

        <section className="widget">
          <div className="widget-title"><h3><Flame size={19} className="hot" /> Популярные темы</h3><Link to="/forum">Все темы <ArrowRight size={14} /></Link></div>
          <div className="top-list">
            {[...loadedItems].sort((a, b) => b.topic.replies_count - a.topic.replies_count).slice(0, 5).map((it, i) => (
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
