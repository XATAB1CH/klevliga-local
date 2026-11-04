import { useState } from 'react';
import { Send, X } from 'lucide-react';
import { useCategories, useCreateTopic } from '../api/hooks';

interface Props {
  onClose: () => void;
  onCreated?: () => void;
}

export function CreateTopicDialog({ onClose, onCreated }: Props) {
  const { data: categories = [] } = useCategories();
  const create = useCreateTopic();
  const [categoryId, setCategoryId] = useState('');
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [error, setError] = useState<string | null>(null);

  const effectiveCategory = categoryId || categories[0]?.id || '';

  const submit = async () => {
    setError(null);
    try {
      await create.mutateAsync({ categoryId: effectiveCategory, title: title.trim(), body: body.trim() });
      onClose();
      onCreated?.();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка');
    }
  };

  return (
    <div className="overlay" onMouseDown={onClose} role="presentation">
      <div className="dialog" role="dialog" aria-modal="true" onMouseDown={(e) => e.stopPropagation()}>
        <div className="dialog-head">
          <h2>Новая тема</h2>
          <button className="icon-btn" onClick={onClose} aria-label="Закрыть"><X /></button>
        </div>
        <label>Раздел
          <select value={effectiveCategory} onChange={(e) => setCategoryId(e.target.value)} style={{ width: '100%', marginTop: 8, padding: 12, background: '#0b202c', color: '#eaf5f6', border: '1px solid #31515f', borderRadius: 6 }}>
            {categories.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
          </select>
        </label>
        <label>Заголовок<input placeholder="О чём хотите рассказать?" value={title} onChange={(e) => setTitle(e.target.value)} /></label>
        <label>Текст темы<textarea rows={6} placeholder="Условия, снасти, наблюдения…" value={body} onChange={(e) => setBody(e.target.value)} /></label>
        {error && <p className="muted">{error}</p>}
        <button className="primary-btn" disabled={!title.trim() || !body.trim() || create.isPending} onClick={submit}><Send size={17} /> Опубликовать</button>
      </div>
    </div>
  );
}
