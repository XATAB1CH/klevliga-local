import { ApiError } from '../api/client';

export function errorMessage(error: unknown, fallback = 'Не удалось загрузить данные. Попробуйте ещё раз.') {
  if (error instanceof ApiError) {
    if (error.code === 'INVALID_CREDENTIALS') return 'Неверный email или пароль.';
    if (error.status === 401) return 'Сессия завершилась. Войдите ещё раз.';
    if (error.code === 'CONFLICT') return 'Этот email или ник уже занят.';
    if (error.code === 'INVALID_VERIFICATION_CODE') return 'Код неверный или истёк. Запросите новый код.';
    if (error.code === 'RATE_LIMITED' || error.status === 429) return 'Слишком много попыток. Подождите немного.';
    if (error.status === 413) return 'Фото слишком большое. Выберите файл до 5 МБ.';
    if (error.status === 415) return 'Поддерживаются только JPEG и PNG.';
    if (error.status === 403) return 'Это действие сейчас недоступно для вашего аккаунта.';
    if (error.status === 400) return 'Проверьте заполненные поля и попробуйте ещё раз.';
  }
  return fallback;
}
export function ErrorNotice({ error, retry, message }: { error?: unknown; retry?: () => unknown; message?: string }) {
  return <div className="query-message query-error" role="alert"><p>{message ?? errorMessage(error)}</p>
    {retry && <button type="button" className="secondary-btn" onClick={() => { void retry(); }}>Повторить</button>}
  </div>;
}
export function Loading({ text = 'Загрузка…' }: { text?: string }) {
  return <div className="empty" role="status">{text}</div>;
}
export function LoadMore({ hasNextPage, isFetchingNextPage, fetchNextPage }: {
  hasNextPage: boolean; isFetchingNextPage: boolean; fetchNextPage: () => unknown;
}) {
  return hasNextPage ? <div className="load-more"><button type="button" className="secondary-btn"
    disabled={isFetchingNextPage} onClick={() => { void fetchNextPage(); }}>
    {isFetchingNextPage ? 'Загрузка…' : 'Показать ещё'}
  </button></div> : null;
}
