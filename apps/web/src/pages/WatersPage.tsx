import { Link } from 'react-router-dom';
import { Waves } from 'lucide-react';
import { useInfiniteWaters } from '../api/hooks';
import { waterKindLabel } from '../lib/format';
import { ErrorNotice, Loading, LoadMore } from '../components/QueryState';

export function WatersPage() {
  const waters = useInfiniteWaters({ limit: 20 });
  const items = waters.data?.pages.flatMap((page) => page.items) ?? [];
  return <div className="page">
    <div className="page-header"><div>
      <span className="eyebrow">РЫБОЛОВНЫЙ НАВИГАТОР</span><h1>Водоёмы и места</h1>
      <p>Карточки водоёмов, отчёты и обсуждения. Точные рыболовные точки не публикуются.</p>
    </div></div>
    {waters.isPending && <Loading text="Загрузка водоёмов…" />}
    {waters.isError && <ErrorNotice error={waters.error} retry={waters.refetch} />}
    {!waters.isPending && !waters.isError && !items.length && <div className="empty">Водоёмы пока не добавлены.</div>}
    <div className="water-cards">{items.map((water) =>
      <Link className="water-card" key={water.id} to={'/waters/' + water.id}>
        <div className="water-placeholder"><Waves size={44} /></div>
        <div><span>{waterKindLabel(water.kind)}</span><h2>{water.name}</h2><p>{water.description ?? 'Описание появится позже.'}</p></div>
      </Link>)}</div>
    <LoadMore {...waters} />
  </div>;
}
