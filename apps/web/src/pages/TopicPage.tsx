import { Link, useParams } from 'react-router-dom';
import { ApiError } from '../api/client';
import { useTopic } from '../api/hooks';
import { TopicContent } from '../components/TopicDialog';
import { ErrorNotice, Loading } from '../components/QueryState';

export function TopicPage() {
  const { id = '' } = useParams();
  const topic = useTopic(id);
  return <div className="page">
    <Link to="/forum" className="back">← Все обсуждения</Link>
    {topic.isPending ? <Loading text="Загрузка темы…" /> : topic.isError ?
      topic.error instanceof ApiError && topic.error.status === 404 ?
        <section className="panel query-message"><h1>Тема не найдена</h1><p>Она удалена или ссылка устарела.</p></section> :
        <ErrorNotice error={topic.error} retry={topic.refetch} /> :
      topic.data && <article className="panel topic-page"><h1>{topic.data.title}</h1><TopicContent key={topic.data.id} topic={topic.data} /></article>}
  </div>;
}
