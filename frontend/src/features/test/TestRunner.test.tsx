import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TestRunner } from './TestRunner';
import type { Settings } from '../../api/types';

const settings: Settings = { fontSize: 24, backgroundColor: '#ffffff', showTimer: true, soundOnError: false };
const table = Array.from({ length: 25 }, (_, i) => i + 1);
const tables = [table, table, table, table, table];

describe('TestRunner', () => {
  it('проводит пять таблиц и возвращает результаты', () => {
    const onFinish = vi.fn();
    render(<TestRunner tables={tables} settings={settings} onFinish={onFinish} />);

    fireEvent.click(screen.getByRole('button', { name: 'Начать тест' }));
    for (let t = 1; t <= 5; t++) {
      expect(screen.getByText(`Таблица ${t} из 5`)).toBeInTheDocument();
      if (t === 1) fireEvent.click(screen.getByLabelText('число 7'));
      for (let v = 1; v <= 25; v++) fireEvent.click(screen.getByLabelText(`число ${v}`));
      if (t < 5) fireEvent.click(screen.getByRole('button', { name: new RegExp(`Таблица ${t + 1} — начать`) }));
    }
    expect(onFinish).toHaveBeenCalledTimes(1);
    const results = onFinish.mock.calls[0][0];
    expect(results).toHaveLength(5);
    expect(results[0].errors).toBe(1);
    expect(results[1].errors).toBe(0);
    expect(results.every((r: { seconds: number }) => r.seconds > 0)).toBe(true);
  });

  it('скрывает секундомер, если он отключён в настройках', () => {
    render(<TestRunner tables={tables} settings={{ ...settings, showTimer: false }} onFinish={() => {}} />);
    fireEvent.click(screen.getByRole('button', { name: 'Начать тест' }));
    expect(screen.queryByText(/^\d\d:\d\d\.\d$/)).not.toBeInTheDocument();
    expect(screen.getByText('Найдите: 1')).toBeInTheDocument();
  });

  it('показывает подсказку со следующим числом', () => {
    render(<TestRunner tables={tables} settings={settings} onFinish={() => {}} />);
    fireEvent.click(screen.getByRole('button', { name: 'Начать тест' }));
    fireEvent.click(screen.getByLabelText('число 1'));
    expect(screen.getByText('Найдите: 2')).toBeInTheDocument();
  });
});
