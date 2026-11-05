import { Link, useParams } from 'react-router-dom';
import { ArrowRight } from 'lucide-react';
import { ApiError } from '../api/client';
import { useInfiniteReports, useWater } from '../api/hooks';
import { waterKindLabel } from '../lib/format';
import { ErrorNotice, Loading, LoadMore } from '../components/QueryState';

function WaterReports({ waterId }: { waterId: string }) {
  const reports = useInfiniteReports({ water_id: waterId, limit: 20 });
  const items = reports.data?.pages.flatMap((page) => page.items) ?? [];
  return <section className="panel"><h2 className="panel-head">Отчёты сообщества</h2>
    {reports.isPending && <Loading text="Загрузка отчётов…" />}
    {reports.isError && <ErrorNotice error={reports.error} retry={reports.refetch} />}
    {!reports.isPending && !reports.isError && !items.length && <div className="empty">Отчётов об этом водоёме пока нет.</div>}
    {items.map((report) => <Link className="report-link" to={'/forum/topics/' + report.topic_id} key={report.id}>
      <b>{report.title}</b><p>{report.body}</p><span className="muted">{report.author_name} · {report.fished_at}</span>
    </Link>)}
    <LoadMore {...reports} />
  </section>;
}
export function WaterPage() {
  const { id = '' } = useParams();
  const water = useWater(id);
  return <div className="page">
    <Link to="/waters" className="back">← Все водоёмы</Link>
    {water.isPending ? <Loading text="Загрузка водоёма…" /> : water.isError ?
      water.error instanceof ApiError && water.error.status === 404 ?
        <section className="panel query-message"><h1>Водоём не найден</h1><p>Проверьте ссылку или выберите водоём из каталога.</p></section> :
        <ErrorNotice error={water.error} retry={water.refetch} /> : water.data && <>
      <div className="water-header" style={{ backgroundImage: 'linear-gradient(90deg,#061923ed,#06192377)' }}>
        <span className="eyebrow">{waterKindLabel(water.data.kind)}</span><h1>{water.data.name}</h1>
        <p>Отчёты сообщества, обсуждения и информация о водоёме.</p>
      </div>
      <section className="panel"><h2 className="panel-head">О водоёме</h2>
        <p className="water-note">{water.data.description ?? 'Описание водоёма появится позже.'}</p>
        <Link to="/forum" className="secondary-btn">Перейти к обсуждениям <ArrowRight size={16} /></Link>
      </section>
      <WaterReports waterId={id} />
    </>}
  </div>;
}
