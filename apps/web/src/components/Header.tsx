import { useEffect, useRef, useState, type FormEvent } from 'react';
import { Link, NavLink, useNavigate } from 'react-router-dom';
import { Bell, ChevronDown, Fish, MapPin, Menu, Search, UserRound, X } from 'lucide-react';
import { useAuth } from '../auth/AuthContext';

interface Props { menuOpen: boolean; onMenuToggle: () => void }

export function Header({ menuOpen, onMenuToggle }: Props) {
  const { user } = useAuth();
  const [query, setQuery] = useState('');
  const [searchOpen, setSearchOpen] = useState(false);
  const input = useRef<HTMLInputElement>(null);
  const searchToggle = useRef<HTMLButtonElement>(null);
  const navigate = useNavigate();
  useEffect(() => { if (searchOpen) input.current?.focus(); }, [searchOpen]);
  const closeSearch = () => { setSearchOpen(false); searchToggle.current?.focus(); };
  const submit = (event: FormEvent) => {
    event.preventDefault();
    const q = query.trim();
    if (q.length < 2) return;
    navigate('/search?q=' + encodeURIComponent(q));
    setSearchOpen(false);
  };
  return <header className="header"><div className="header-inner">
    <button className="mobile-menu icon-btn" onClick={onMenuToggle}
      aria-expanded={menuOpen} aria-controls="site-sidebar" aria-label={menuOpen ? 'Закрыть меню' : 'Открыть меню'}>
      {menuOpen ? <X /> : <Menu />}
    </button>
    <Link to="/" className="logo"><Fish size={30} /><span>Клёв<b>Лига</b></span></Link>
    <nav className="header-nav" aria-label="Основная навигация">
      <NavLink to="/forum">Форум</NavLink><NavLink to="/waters">Водоёмы</NavLink>
      <NavLink to="/forum?category=reports">Отчёты</NavLink><NavLink to="/knowledge">База знаний</NavLink>
      <NavLink to="/profile">Сообщество</NavLink>
    </nav>
    <button ref={searchToggle} type="button" className="mobile-search icon-btn" onClick={() => setSearchOpen(true)}
      aria-label="Открыть поиск" aria-expanded={searchOpen} aria-controls="site-search"><Search size={20} /></button>
    <form id="site-search" role="search" className={searchOpen ? 'search-field search-expanded' : 'search-field'} onSubmit={submit}>
      <input ref={input} value={query} onChange={(e) => setQuery(e.target.value)}
        onKeyDown={(e) => { if (e.key === 'Escape' && searchOpen) closeSearch(); }}
        required minLength={2} maxLength={200} placeholder="Поиск тем, водоёмов, пользователей…" aria-label="Поиск" />
      <button className="search-submit icon-btn" type="submit" aria-label="Найти"><Search size={18} /></button>
      {searchOpen && <button type="button" className="icon-btn" onClick={closeSearch} aria-label="Закрыть поиск"><X size={18} /></button>}
    </form>
    <button className="location" onClick={() => navigate('/waters')}><MapPin size={18} /> Водоёмы <ChevronDown size={15} /></button>
    <button className="icon-btn bell" onClick={() => navigate('/notifications')} aria-label="Уведомления"><Bell size={20} /></button>
    <Link to="/profile" className="account" aria-label="Мой профиль">
      {user?.avatar_url ? <img src={user.avatar_url} alt="" /> : <UserRound size={21} />}
    </Link>
  </div></header>;
}
