import type { Level, Trend } from '../api/types';

/** Секунды → «12,34 с». */
export function formatSeconds(v: number, digits = 2): string {
  return `${v.toFixed(digits).replace('.', ',')} с`;
}

/** Число с запятой в качестве десятичного разделителя. */
export function formatNumber(v: number, digits = 2): string {
  return v.toFixed(digits).replace('.', ',');
}

/** Секундомер: 83.4 → «01:23.4». */
export function formatStopwatch(ms: number): string {
  const total = Math.max(0, ms) / 1000;
  const m = Math.floor(total / 60);
  const s = total - m * 60;
  return `${String(m).padStart(2, '0')}:${s.toFixed(1).padStart(4, '0')}`;
}

export function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString('ru-RU', {
    day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit',
  });
}

export function warmUpText(level: Level): string {
  return level === 'good'
    ? 'Хорошая врабатываемость'
    : 'Испытуемый долго сосредотачивается на основной работе';
}

export function stabilityText(level: Level): string {
  return level === 'good'
    ? 'Хорошая психическая устойчивость'
    : 'Низкая психическая устойчивость к выполнению работы';
}

export function trendText(t: Trend): string {
  switch (t) {
    case 'improving': return 'Результаты улучшаются: время прохождения таблицы сокращается.';
    case 'worsening': return 'Результаты ухудшаются: время прохождения таблицы растёт.';
    case 'stable': return 'Результаты стабильны.';
    default: return 'Недостаточно данных: нужно минимум два завершённых тестирования.';
  }
}
