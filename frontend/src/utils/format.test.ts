import { describe, expect, it } from 'vitest';
import { formatNumber, formatSeconds, formatStopwatch, stabilityText, trendText, warmUpText } from './format';

describe('format', () => {
  it('форматирует секунды с запятой', () => {
    expect(formatSeconds(12.345)).toBe('12,35 с');
    expect(formatSeconds(3, 1)).toBe('3,0 с');
  });
  it('форматирует числа', () => {
    expect(formatNumber(1.2)).toBe('1,20');
  });
  it('форматирует секундомер', () => {
    expect(formatStopwatch(0)).toBe('00:00.0');
    expect(formatStopwatch(83_400)).toBe('01:23.4');
    expect(formatStopwatch(-5)).toBe('00:00.0');
  });
  it('даёт интерпретации', () => {
    expect(warmUpText('good')).toMatch(/Хорошая/);
    expect(warmUpText('bad')).toMatch(/долго/);
    expect(stabilityText('good')).toMatch(/Хорошая/);
    expect(stabilityText('bad')).toMatch(/Низкая/);
    expect(trendText('improving')).toMatch(/улучшаются/);
    expect(trendText('unknown')).toMatch(/минимум два/);
  });
});
