import { useState, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { Search, ShieldCheck } from 'lucide-react';
import { useAuth } from '../auth/AuthContext';
import { useAuditEvents, useChangeAccess, useManagedUser, useManagedUsers, useMe } from '../api/hooks';
import type { AuditFilters, ManagedUser, Role } from '../api/types';
import { ErrorNotice, Loading, LoadMore, errorMessage } from '../components/QueryState';

const roleNames: Record<Role, string> = { user: 'Пользователь', moderator: 'Модератор', support: 'Поддержка', admin: 'Администратор' };
const actionNames: Record<string, string> = {
  'admin.bootstrap': 'Назначен первый администратор',
  'user.roles_changed': 'Изменены роли',
  'user.status_changed': 'Изменён статус аккаунта',
};

export function AdminPage({ onLogin }: { onLogin: () => void }) {
  const { user, isLoading, error } = useAuth();
  const me = useMe();
  if (isLoading) return <Loading />;
  if (error) return <ErrorNotice error={error} retry={me.refetch} />;
  if (!user) return <div className="page"><h1>Управление сообществом</h1><p>Войдите в аккаунт администратора.</p><button className="primary-btn" onClick={onLogin}>Войти</button></div>;
  if (!user.permissions?.includes('users.manage')) return <div className="page"><h1>Доступ ограничен</h1><p>Управление правами и журнал доступны администраторам.</p><Link className="secondary-btn" to="/profile">В профиль</Link></div>;
  return <AdminWorkspace key={user.id} />;
}

function AdminWorkspace() {
  const [query, setQuery] = useState('');
  const [search, setSearch] = useState('');
  const [selected, setSelected] = useState('');
  const [changeNotice, setChangeNotice] = useState('');
  const results = useManagedUsers(search);
  const account = useManagedUser(selected);
  const [filters, setFilters] = useState<AuditFilters>({});
  const events = useAuditEvents(filters);
  const [filterError, setFilterError] = useState('');

  const filterAudit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const from = String(form.get('from') ?? '');
    const until = String(form.get('until') ?? '');
    if (from && until && new Date(from) >= new Date(until)) {
      setFilterError('Конец периода должен быть позже начала.'); return;
    }
    setFilterError('');
    setFilters({ action: String(form.get('action') ?? '') || undefined,
      target_id: form.get('selected') ? selected : undefined,
      from: from ? new Date(from).toISOString() : undefined,
      until: until ? new Date(until).toISOString() : undefined });
  };

  return <div className="page access-page">
    <div className="page-header"><div><span className="eyebrow">КЛЁВЛИГА · СООБЩЕСТВО</span><h1><ShieldCheck aria-hidden="true" /> Управление сообществом</h1><p>Права участников и история служебных изменений.</p></div></div>
    <section className="panel access-section" aria-labelledby="access-users-title">
      <h2 id="access-users-title">Участники</h2>
      <form className="access-search" onSubmit={(event) => { event.preventDefault(); setSearch(query.trim()); setSelected(''); }}>
        <label>Ник или имя<input required minLength={2} maxLength={200} value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Найти участника" /></label>
        <button className="secondary-btn" type="submit"><Search size={18} aria-hidden="true" /> Найти</button>
      </form>
      {results.isFetching && <Loading text="Ищем участников…" />}
      {results.error && <ErrorNotice error={results.error} retry={results.refetch} />}
      {results.data && <div className="access-users" aria-label="Результаты поиска">
        {results.data.items.length === 0 && <p>Участники не найдены.</p>}
        {results.data.items.map((user) => <button key={user.id} type="button" className={selected === user.id ? 'access-user selected' : 'access-user'} aria-pressed={selected === user.id} onClick={() => setSelected(user.id)}>
          <strong>{user.display_name}</strong><span>@{user.username} · {user.status === 'blocked' ? 'Заблокирован' : 'Активен'}</span>
        </button>)}
        {results.data.items.length === 20 && <p className="tiny">Показаны первые 20 участников. Уточните имя для поиска.</p>}
      </div>}
      {selected && (account.isPending ? <Loading /> : account.error ? <ErrorNotice error={account.error} retry={account.refetch} /> : account.data &&
        <AccessEditor key={`${account.data.id}:${account.data.roles.join(',')}:${account.data.status}`} account={account.data} onNotice={setChangeNotice} />)}
      {changeNotice && <p className="success-message" role="status">{changeNotice}</p>}
    </section>
    <section className="panel access-section" aria-labelledby="access-audit-title">
      <h2 id="access-audit-title">Журнал действий</h2><p className="access-hint">Успешные изменения сохраняются с причиной. Записи журнала нельзя редактировать.</p>
      <form className="access-filters" onSubmit={filterAudit}>
        <label>Действие<select name="action"><option value="">Все действия</option>{Object.entries(actionNames).map(([value, label]) => <option key={value} value={value}>{label}</option>)}</select></label>
        <label>С начала периода<input type="datetime-local" name="from" /></label>
        <label>До конца периода<input type="datetime-local" name="until" /></label>
        <label className="access-check"><input type="checkbox" name="selected" disabled={!selected} /> Только выбранный участник</label>
        <button type="submit" className="secondary-btn">Применить</button>
      </form>
      {filterError && <p role="alert" className="form-error">{filterError}</p>}
      {events.isPending ? <Loading /> : events.error ? <ErrorNotice error={events.error} retry={events.refetch} /> : <>
        {events.data?.pages.every((page) => page.items.length === 0) && <p>В этом периоде записей нет.</p>}
        <ol className="audit-list">{events.data?.pages.flatMap((page) => page.items).map((entry) => <li key={entry.id}>
          <div className="audit-heading"><strong>{actionNames[entry.action] ?? entry.action}</strong><time dateTime={entry.created_at}>{new Date(entry.created_at).toLocaleString('ru-RU')}</time></div>
          <p className="user-text">{entry.reason}</p>
          {Object.entries(entry.changes).map(([field, change]) => {
            const describe = (value: string | Role[]) => Array.isArray(value) ? value.map((role) => roleNames[role]).join(', ') : value === 'blocked' ? 'Заблокирован' : value === 'active' ? 'Активен' : value;
            return <p key={field}>{field === 'roles' ? 'Роли' : 'Статус'}: {describe(change.before)} → {describe(change.after)}</p>;
          })}
          <details><summary>Участники действия</summary><p>Исполнитель: {entry.actor_type === 'bootstrap' ? 'Первоначальная настройка на сервере' : entry.actor_id}</p><p>Аккаунт: {entry.target_id}</p><p>Номер запроса: {entry.request_id}</p></details>
        </li>)}</ol>
        <LoadMore hasNextPage={events.hasNextPage} isFetchingNextPage={events.isFetchingNextPage} fetchNextPage={events.fetchNextPage} />
      </>}
    </section>
  </div>;
}

function AccessEditor({ account, onNotice }: { account: ManagedUser; onNotice: (notice: string) => void }) {
  const [roles, setRoles] = useState<Role[]>(account.roles);
  const [reason, setReason] = useState('');
  const [actionError, setActionError] = useState('');
  const mutation = useChangeAccess();
  const apply = async (kind: 'roles' | 'status') => {
    onNotice(''); setActionError('');
    if (reason.trim().length < 3) { setActionError('Укажите причину изменения: от 3 символов.'); return; }
    try {
      await mutation.mutateAsync(kind === 'roles' ? { kind, id: account.id, roles, reason } :
        { kind, id: account.id, status: account.status === 'blocked' ? 'active' : 'blocked', reason });
      onNotice('Изменение сохранено в журнале.');
    } catch (error) { setActionError(errorMessage(error, 'Не удалось сохранить изменение. Попробуйте ещё раз.')); }
  };
  const changed = [...roles].sort().join() !== [...account.roles].sort().join();
  return <form className="access-editor" onSubmit={(event) => { event.preventDefault(); void apply('roles'); }}>
    <h3>{account.display_name} <small>@{account.username}</small></h3>
    <p>Аккаунт {account.status === 'blocked' ? 'заблокирован' : 'активен'}. Служебные роли можно совмещать.</p>
    <fieldset disabled={mutation.isPending}><legend>Роли участника</legend><div className="access-roles">
      {(Object.keys(roleNames) as Role[]).map((role) => <label key={role} className="access-check"><input type="checkbox" checked={roles.includes(role)} disabled={role === 'user'} onChange={(event) => setRoles(event.target.checked ? [...roles, role] : roles.filter((item) => item !== role))} />{roleNames[role]}</label>)}
    </div>
      <label>Причина изменения<textarea required minLength={3} maxLength={1000} value={reason} onChange={(event) => setReason(event.target.value)} rows={2} /></label>
      <div className="access-actions"><button className="primary-btn" type="submit" disabled={!changed}>{mutation.isPending ? 'Сохраняем…' : 'Сохранить роли'}</button>
        <button className="secondary-btn" type="button" onClick={() => { void apply('status'); }}>{account.status === 'blocked' ? 'Разблокировать аккаунт' : 'Заблокировать аккаунт'}</button></div>
    </fieldset>
    {actionError && <p className="form-error" role="alert">{actionError}</p>}
  </form>;
}
