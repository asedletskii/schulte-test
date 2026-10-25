import type { TableResult } from '../../api/types';

/**
 * Состояние прохождения теста из пяти таблиц.
 * Редьюсер чистый (время передаётся в действиях), поэтому легко тестируется.
 */
export type Phase = 'intro' | 'running' | 'between' | 'finished';

export interface RunState {
  phase: Phase;
  tableIndex: number; // 0…4
  next: number; // следующее искомое число 1…26
  errors: number;
  startedAt: number | null; // момент показа текущей таблицы, мс
  results: TableResult[];
  lastWrong: number | null; // последнее ошибочно выбранное число (для подсветки)
}

export const TABLES = 5;
export const CELLS = 25;

export const initialRunState: RunState = {
  phase: 'intro',
  tableIndex: 0,
  next: 1,
  errors: 0,
  startedAt: null,
  results: [],
  lastWrong: null,
};

export type RunAction =
  | { type: 'start'; now: number } // начать/продолжить: показать текущую таблицу
  | { type: 'pick'; value: number; now: number } // выбор числа в таблице
  | { type: 'reset' };

export function runReducer(state: RunState, action: RunAction): RunState {
  switch (action.type) {
    case 'start':
      if (state.phase !== 'intro' && state.phase !== 'between') return state;
      return { ...state, phase: 'running', next: 1, errors: 0, startedAt: action.now, lastWrong: null };

    case 'pick': {
      if (state.phase !== 'running' || state.startedAt === null) return state;
      if (action.value !== state.next) {
        return { ...state, errors: state.errors + 1, lastWrong: action.value };
      }
      if (state.next < CELLS) {
        return { ...state, next: state.next + 1, lastWrong: null };
      }
      // последнее число найдено — таблица пройдена
      const result: TableResult = {
        number: state.tableIndex + 1,
        seconds: Math.max(0.001, (action.now - state.startedAt) / 1000),
        errors: state.errors,
      };
      const results = [...state.results, result];
      const done = results.length === TABLES;
      return {
        ...state,
        results,
        phase: done ? 'finished' : 'between',
        tableIndex: done ? state.tableIndex : state.tableIndex + 1,
        next: CELLS + 1,
        startedAt: null,
        lastWrong: null,
      };
    }

    case 'reset':
      return initialRunState;
  }
}
