// Форматирование чисел и относительного времени.

export function fmtCount(n: number): string {
  return n >= 1000 ? `${(n / 1000).toFixed(1)}k` : String(n);
}

// timeAgo превращает ISO-дату в короткую относительную строку на русском.
export function timeAgo(iso: string): string {
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return '';

  const sec = Math.round((Date.now() - then) / 1000);
  const abs = Math.abs(sec);

  if (abs < 60) return 'только что';
  if (abs < 3600) return `${Math.floor(abs / 60)} мин. назад`;
  if (abs < 86400) return `${Math.floor(abs / 3600)} ч назад`;
  if (abs < 86400 * 7) return `${Math.floor(abs / 86400)} дн. назад`;

  return new Date(then).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' });
}

// kindLabel превращает API-тип темы в человекочитаемую метку.
export function kindLabel(type: string): string {
  switch (type) {
    case 'report':
      return 'Отчёт';
    case 'question':
      return 'Вопрос';
    default:
      return 'Обсуждение';
  }
}

// waterKindLabel превращает тип водоёма в человекочитаемую метку.
export function waterKindLabel(kind: string): string {
  switch (kind) {
    case 'river':
      return 'Река';
    case 'lake':
      return 'Озеро';
    case 'reservoir':
      return 'Водохранилище';
    case 'pond':
      return 'Пруд';
    default:
      return 'Водоём';
  }
}
