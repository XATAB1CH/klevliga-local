import { useState } from 'react';
import { NavLink, Route, Routes, useNavigate } from 'react-router-dom';
import { Home, MapPin, MessageSquare, Plus, Users } from 'lucide-react';
import { AuthDialog } from './components/AuthDialog';
import { CreateTopicDialog } from './components/CreateTopicDialog';
import { Header } from './components/Header';
import { Sidebar } from './components/Sidebar';
import { errorMessage } from './components/QueryState';
import { useFollow } from './api/hooks';
import { useAuth } from './auth/AuthContext';
import type { Topic } from './api/types';
import { ForumPage } from './pages/ForumPage';
import { HomePage } from './pages/HomePage';
import { NotificationsPage } from './pages/NotificationsPage';
import { Placeholder } from './pages/Placeholder';
import { ProfilePage } from './pages/ProfilePage';
import { SearchResultsPage } from './pages/SearchResultsPage';
import { SubscriptionsPage } from './pages/SubscriptionsPage';
import { WaterPage } from './pages/WaterPage';
import { WatersPage } from './pages/WatersPage';
import { TopicPage } from './pages/TopicPage';
import { NotFoundPage } from './pages/NotFoundPage';
import { AdminPage } from './pages/AdminPage';

export default function App() {
  const [menu, setMenu] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [authOpen, setAuthOpen] = useState(false);
  const [followNotice, setFollowNotice] = useState<{ text: string; error: boolean } | null>(null);
  const navigate = useNavigate();
  const openTopic = (topic: Topic) => navigate(`/forum/topics/${topic.id}`);
  const { isAuthenticated } = useAuth();
  const follow = useFollow();

  const handleNew = () => {
    if (!isAuthenticated) setAuthOpen(true);
    else setCreateOpen(true);
  };

  const handleFollow = (topic: Topic) => {
    if (!isAuthenticated) {
      setAuthOpen(true);
      return;
    }
    setFollowNotice(null);
    follow.mutate({ targetType: 'topic', targetId: topic.id }, {
      onSuccess: () => setFollowNotice({ text: 'Подписка сохранена. Новые ответы появятся в уведомлениях.', error: false }),
      onError: (error) => setFollowNotice({ text: errorMessage(error, 'Не удалось подписаться. Попробуйте ещё раз.'), error: true }),
    });
  };

  return (
    <div className="app">
      <Header menuOpen={menu} onMenuToggle={() => setMenu((v) => !v)} />
      <div className="app-grid">
        {menu && <button className="drawer-backdrop" onClick={() => setMenu(false)} aria-label="Закрыть меню" />}
        <aside id="site-sidebar" className={`sidebar ${menu ? 'opened' : ''}`}><Sidebar close={() => setMenu(false)} /></aside>
        <main className="main">
          {followNotice && <div className="action-notice" role={followNotice.error ? 'alert' : 'status'}>{followNotice.text}
            <button type="button" className="icon-btn" onClick={() => setFollowNotice(null)} aria-label="Скрыть сообщение">×</button>
          </div>}
          <Routes>
            <Route path="/" element={<HomePage onOpen={openTopic} onFollow={handleFollow} onNew={handleNew} />} />
            <Route path="/forum" element={<ForumPage onOpen={openTopic} onFollow={handleFollow} onNew={handleNew} />} />
            <Route path="/forum/topics/:id" element={<TopicPage />} />
            <Route path="/waters" element={<WatersPage />} />
            <Route path="/waters/:id" element={<WaterPage />} />
            <Route path="/search" element={<SearchResultsPage onOpen={openTopic} onFollow={handleFollow} />} />
            <Route path="/profile" element={<ProfilePage onLogin={() => setAuthOpen(true)} />} />
            <Route path="/admin" element={<AdminPage onLogin={() => setAuthOpen(true)} />} />
            <Route path="/notifications" element={<NotificationsPage onLogin={() => setAuthOpen(true)} />} />
            <Route path="/subscriptions" element={<SubscriptionsPage />} />
            <Route path="/knowledge" element={<Placeholder section="База знаний" />} />
            <Route path="/market" element={<Placeholder section="Барахолка" />} />
            <Route path="/assistant" element={<Placeholder section="КлёвЛига AI" />} />
            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </main>
      </div>
      <nav className="bottom-nav">
        <NavLink to="/"><Home />Главная</NavLink>
        <NavLink to="/forum"><MessageSquare />Форум</NavLink>
        <button className="new-mobile" onClick={handleNew} aria-label="Создать"><Plus /></button>
        <NavLink to="/waters"><MapPin />Водоёмы</NavLink>
        <NavLink to="/profile"><Users />Профиль</NavLink>
      </nav>
      {createOpen && <CreateTopicDialog onClose={() => setCreateOpen(false)} />}
      {authOpen && <AuthDialog onClose={() => setAuthOpen(false)} />}
    </div>
  );
}
