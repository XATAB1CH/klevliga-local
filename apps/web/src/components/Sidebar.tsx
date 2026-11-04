import { useState } from 'react';
import { Link, NavLink } from 'react-router-dom';
import { Bell, BookOpen, ChevronDown, Fish, Home, MapPin, MessageSquare, Star, UserRound, Waves } from 'lucide-react';
import { useCategories, useWaters } from '../api/hooks';
import { waterKindLabel } from '../lib/format';

export function Sidebar({ close }: { close: () => void }) {
  const [forumOpen, setForumOpen] = useState(false);
  const [watersOpen, setWatersOpen] = useState(false);
  const { data: categories = [] } = useCategories();
  const { data: watersPage } = useWaters({ limit: 6 });

  return (
    <>
      <nav className="sidebar-primary">
        <NavLink end to="/" onClick={close}><Home size={19} /> Главная</NavLink>
        <NavLink to="/subscriptions" onClick={close}><Star size={19} /> Мои подписки</NavLink>
        <NavLink to="/notifications" onClick={close}><Bell size={19} /> Уведомления</NavLink>
        <NavLink to="/profile" onClick={close}><UserRound size={19} /> Профиль</NavLink>
      </nav>
      <div className="side-divider" />
      <div className="desktop-side-label">РАЗДЕЛЫ ФОРУМА</div>
      <button className="section-toggle" onClick={() => setForumOpen((x) => !x)} aria-expanded={forumOpen}><BookOpen size={19} /> Разделы форума <ChevronDown size={17} /></button>
      <div className={`categories ${forumOpen ? 'expanded' : ''}`}>
        {categories.map((c, i) => (
          <Link key={c.id} onClick={close} to={`/forum?category=${encodeURIComponent(c.slug)}`}>
            <span className="category-icon">
              {i === 2 ? <MapPin size={18} /> : i === 0 ? <MessageSquare size={18} /> : i === 1 ? <BookOpen size={18} /> : <Fish size={18} />}
            </span>
            {c.name}
          </Link>
        ))}
      </div>
      <div className="side-divider" />
      <div className="desktop-side-label">ПОПУЛЯРНЫЕ ВОДОЁМЫ</div>
      <button className="section-toggle" onClick={() => setWatersOpen((x) => !x)} aria-expanded={watersOpen}><Waves size={19} /> Популярные водоёмы <ChevronDown size={17} /></button>
      <div className={`side-waters ${watersOpen ? 'expanded' : ''}`}>
        {watersPage?.items.map((w) => (
          <Link key={w.id} to={`/waters/${w.id}`} onClick={close}>
            <Waves size={18} />
            <span><b>{w.name}</b><small>{waterKindLabel(w.kind)}</small></span>
          </Link>
        ))}
      </div>
    </>
  );
}
