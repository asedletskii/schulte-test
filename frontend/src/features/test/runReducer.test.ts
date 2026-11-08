import { describe, expect, it } from 'vitest';
import { CELLS, initialRunState, runReducer, TABLES, type RunState } from './runReducer';

const pick = (s: RunState, value: number, now: number) => runReducer(s, { type: 'pick', value, now });

function playTable(s: RunState, t0: number, ms: number): RunState {
  s = runReducer(s, { type: 'start', now: t0 });
  for (let v = 1; v <= CELLS; v++) s = pick(s, v, t0 + (ms * v) / CELLS);
  return s;
}

describe('runReducer', () => {
  it('начинается с экрана инструкции', () => {
    expect(initialRunState.phase).toBe('intro');
  });

  it('start переводит в running и запоминает момент старта', () => {
    const s = runReducer(initialRunState, { type: 'start', now: 1000 });
    expect(s).toMatchObject({ phase: 'running', next: 1, errors: 0, startedAt: 1000 });
  });

  it('верный выбор увеличивает искомое число, неверный — счётчик ошибок', () => {
    let s = runReducer(initialRunState, { type: 'start', now: 0 });
    s = pick(s, 1, 100);
    expect(s.next).toBe(2);
    s = pick(s, 5, 200);
    expect(s).toMatchObject({ next: 2, errors: 1, lastWrong: 5 });
    s = pick(s, 2, 300);
    expect(s).toMatchObject({ next: 3, errors: 1, lastWrong: null });
  });

  it('выбор вне фазы running игнорируется', () => {
    expect(pick(initialRunState, 1, 0)).toBe(initialRunState);
  });

  it('после числа 25 таблица завершается и фиксируется время и ошибки', () => {
    let s = runReducer(initialRunState, { type: 'start', now: 0 });
    s = pick(s, 9, 10);
    for (let v = 1; v <= CELLS; v++) s = pick(s, v, 1000 * v);
    expect(s.phase).toBe('between');
    expect(s.results).toHaveLength(1);
    expect(s.results[0]).toEqual({ number: 1, seconds: 25, errors: 1 });
    expect(s.tableIndex).toBe(1);
  });

  it('после пятой таблицы наступает фаза finished с пятью результатами', () => {
    let s = initialRunState;
    for (let i = 0; i < TABLES; i++) s = playTable(s, i * 100_000, 30_000 + i * 1000);
    expect(s.phase).toBe('finished');
    expect(s.results.map((r) => r.seconds)).toEqual([30, 31, 32, 33, 34]);
    expect(s.results.map((r) => r.number)).toEqual([1, 2, 3, 4, 5]);
  });

  it('start в фазе running или finished ничего не меняет', () => {
    const running = runReducer(initialRunState, { type: 'start', now: 5 });
    expect(runReducer(running, { type: 'start', now: 99 })).toBe(running);
  });

  it('reset возвращает исходное состояние', () => {
    const s = playTable(initialRunState, 0, 1000);
    expect(runReducer(s, { type: 'reset' })).toEqual(initialRunState);
  });
});
