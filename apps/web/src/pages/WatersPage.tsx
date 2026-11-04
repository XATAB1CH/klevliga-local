import { Link } from 'react-router-dom';
import { Waves } from 'lucide-react';
import { useWaters } from '../api/hooks';
import { waterKindLabel } from '../lib/format';

export function WatersPage() {
  const { data: waters } = useWaters({ limit: 100 });

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <span className="eyebrow">РЫБОЛОВНЫЙ НАВИГАТОР</span>
          <h1>Водоёмы и места</h1>
          <p>Карточки водоёмов, отчёты и обсуждения. Карта с точками будет в 1.2.</p>
        </div>
      </div>
      <div className="water-cards">
        {(waters?.items ?? []).map((w) => (
          <Link className="water-card" key={w.id} to={`/waters/${w.id}`}>
            <div style={{ height: 130, display: 'grid', placeItems: 'center', background: '#0b202c', color: '#4de6d7' }}><Waves size={44} /></div>
            <div><span>{waterKindLabel(w.kind)}</span><h2>{w.name}</h2><p>{w.description ?? 'Описание появится позже.'}</p></div>
          </Link>
        ))}
      </div>
    </div>
  );
}
