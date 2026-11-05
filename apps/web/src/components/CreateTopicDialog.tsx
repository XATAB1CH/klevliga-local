import { useState, type FormEvent } from 'react';
import { Send } from 'lucide-react';
import { useCategories, useCreateTopic } from '../api/hooks';
import { Dialog } from './Dialog';
import { ErrorNotice, Loading, errorMessage } from './QueryState';

interface Props { onClose: () => void; onCreated?: () => void }

export function CreateTopicDialog({ onClose, onCreated }: Props) {
  const categories = useCategories();
  const create = useCreateTopic();
  const [categoryId, setCategoryId] = useState('');
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [error, setError] = useState<string | null>(null);
  const effectiveCategory = categoryId || categories.data?.[0]?.id || '';
  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (create.isPending || !effectiveCategory) return;
    setError(null);
    try {
      await create.mutateAsync({ categoryId: effectiveCategory, title: title.trim(), body: body.trim() });
      onClose(); onCreated?.();
    } catch (e) { setError(errorMessage(e, 'Не удалось опубликовать тему. Ваш текст сохранён в форме.')); }
  };
  return <Dialog title="Новая тема" onClose={onClose}>
    {categories.isPending ? <Loading /> : categories.isError ? <ErrorNotice error={categories.error} retry={categories.refetch} /> :
      !categories.data?.length ? <p className="query-message">Сейчас нет доступных разделов для публикации.</p> :
      <form onSubmit={(event) => { void submit(event); }}><fieldset disabled={create.isPending}>
        <label>Раздел<select value={effectiveCategory} onChange={(e) => setCategoryId(e.target.value)} required>
          {categories.data.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
        </select></label>
        <label>Заголовок<input data-autofocus autoFocus required minLength={5} maxLength={250} placeholder="О чём хотите рассказать?"
          value={title} onChange={(e) => setTitle(e.target.value)} /></label>
        <label>Текст темы<textarea required maxLength={20000} rows={6} placeholder="Условия, снасти, наблюдения…"
          value={body} onChange={(e) => setBody(e.target.value)} /></label>
        {error && <p className="form-error" role="alert">{error}</p>}
        <button type="submit" className="primary-btn" disabled={title.trim().length < 5 || !body.trim() || create.isPending}>
          <Send size={17} />{create.isPending ? 'Публикация…' : 'Опубликовать'}
        </button>
      </fieldset></form>}
  </Dialog>;
}
