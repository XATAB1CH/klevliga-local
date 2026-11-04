import { Link } from 'react-router-dom';
import { ArrowRight, ShieldCheck } from 'lucide-react';

export function Placeholder({ section }: { section: string }) {
  return (
    <div className="page">
      <div className="page-header">
        <div><span className="eyebrow">КЛЁВЛИГА / БУДУЩИЙ МОДУЛЬ</span><h1>{section}</h1><p>Модуль предусмотрен в дорожной карте. Сейчас формируем основу форума.</p></div>
      </div>
      <div className="panel placeholder">
        <ShieldCheck size={35} />
        <b>Экран будет реализован на следующем этапе</b>
        <Link to="/" className="secondary-btn">На главную <ArrowRight size={16} /></Link>
      </div>
    </div>
  );
}
