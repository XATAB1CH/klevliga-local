import { Link, useParams } from 'react-router-dom';
import { ArrowRight } from 'lucide-react';
import { useWater } from '../api/hooks';
import { waterKindLabel } from '../lib/format';

export function WaterPage() {
  const { id } = useParams();
  const { data: water } = useWater(id ?? '');

  if (!water) {
    return <div className="page">Водоём не найден. <Link to="/waters">Вернуться</Link></div>;
  }

  return (
    <div className="page">
      <Link to="/waters" className="back">← Все водоёмы</Link>
      <div className="water-header" style={{ backgroundImage: 'linear-gradient(90deg,#061923ed,#06192377)' }}>
        <span className="eyebrow">{waterKindLabel(water.kind)} · МОСКОВСКАЯ ОБЛАСТЬ</span>
        <h1>{water.name}</h1>
        <p>Отчёты сообщества, обсуждения и информация о водоёме.</p>
      </div>
      <div className="panel">
        <div className="panel-head">Отчёты и активность</div>
        <p className="water-note">{water.description ?? 'Описание водоёма появится позже.'}</p>
        <Link to="/forum" className="secondary-btn">Перейти к обсуждениям <ArrowRight size={16} /></Link>
      </div>
    </div>
  );
}
