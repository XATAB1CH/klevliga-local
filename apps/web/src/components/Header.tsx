import { useState } from 'react';
import { Link, NavLink, useNavigate } from 'react-router-dom';
import { Bell, ChevronDown, Fish, MapPin, Menu, Search, UserRound, X } from 'lucide-react';

interface Props {
  menuOpen: boolean;
  onMenuToggle: () => void;
}

export function Header({ menuOpen, onMenuToggle }: Props) {
  const [query, setQuery] = useState('');
  const navigate = useNavigate();

  const submit = () => {
    const q = query.trim();
    if (q.length < 2) return;
    navigate(`/search?q=${encodeURIComponent(q)}`);
  };

  return (
    <header className="header">
      <div className="header-inner">
        <button className="mobile-menu icon-btn" onClick={onMenuToggle} aria-label={menuOpen ? 'Закрыть меню' : 'Открыть меню'}>
          {menuOpen ? <X /> : <Menu />}
        </button>
        <Link to="/" className="logo"><Fish size={30} /><span>Клёв<b>Лига</b></span></Link>
        <nav className="header-nav">
          <NavLink to="/forum">Форум</NavLink>
          <NavLink to="/waters">Водоёмы</NavLink>
          <NavLink to="/forum?category=reports">Отчёты</NavLink>
          <NavLink to="/knowledge">База знаний</NavLink>
          <NavLink to="/profile">Сообщество</NavLink>
        </nav>
        <div className="search-field">
          <Search size={18} />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => { if (e.key === 'Enter') submit(); }}
            placeholder="Поиск тем, водоёмов, пользователей..."
            aria-label="Поиск"
          />
          <kbd>Enter</kbd>
        </div>
        <button className="location" onClick={() => navigate('/waters')}><MapPin size={18} /> Москва <ChevronDown size={15} /></button>
        <button className="icon-btn bell" onClick={() => navigate('/notifications')} aria-label="Уведомления"><Bell size={20} /></button>
        <Link to="/profile" className="account"><UserRound size={21} /></Link>
      </div>
    </header>
  );
}
