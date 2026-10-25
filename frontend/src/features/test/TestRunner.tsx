import { useEffect, useReducer, useRef, useState } from 'react';
import { Alert, Box, Button, Card, CardContent, LinearProgress, Paper, Stack, Typography } from '@mui/material';
import type { Settings, TableResult } from '../../api/types';
import { formatSeconds, formatStopwatch } from '../../utils/format';
import { beepError } from './sound';
import { initialRunState, runReducer, TABLES } from './runReducer';

interface Props {
  tables: number[][];
  settings: Settings;
  onFinish: (results: TableResult[]) => void;
}

/** Прохождение теста: пять таблиц 5×5, числа нужно находить по возрастанию от 1 до 25. */
export function TestRunner({ tables, settings, onFinish }: Props) {
  const [state, dispatch] = useReducer(runReducer, initialRunState);
  const [elapsed, setElapsed] = useState(0);
  const reported = useRef(false);

  // секундомер (только отображение; зачёт времени выполняется по меткам performance.now())
  useEffect(() => {
    if (state.phase !== 'running' || state.startedAt === null) return;
    const started = state.startedAt;
    const id = window.setInterval(() => setElapsed(performance.now() - started), 100);
    return () => window.clearInterval(id);
  }, [state.phase, state.startedAt]);

  useEffect(() => {
    if (state.phase === 'finished' && !reported.current) {
      reported.current = true;
      onFinish(state.results);
    }
  }, [state.phase, state.results, onFinish]);

  const pick = (value: number) => {
    if (state.phase !== 'running') return;
    if (value !== state.next && settings.soundOnError) beepError();
    dispatch({ type: 'pick', value, now: performance.now() });
  };
  const start = () => {
    setElapsed(0);
    dispatch({ type: 'start', now: performance.now() });
  };

  if (state.phase === 'intro') {
    return (
      <Card><CardContent>
        <Typography variant="h5" gutterBottom>Инструкция</Typography>
        <Typography paragraph>
          Вам будет показано пять таблиц, в каждой — числа от 1 до 25 в случайном порядке. Находите и нажимайте
          числа <b>по порядку: 1, 2, 3, … 25</b> как можно быстрее и без ошибок. Время засчитывается с момента показа
          таблицы до выбора числа 25.
        </Typography>
        <Typography paragraph color="text.secondary">
          Смотрите на всю таблицу целиком, не сосредотачиваясь на отдельных ячейках. После каждой таблицы можно
          немного передохнуть.
        </Typography>
        <Button variant="contained" size="large" onClick={start}>Начать тест</Button>
      </CardContent></Card>
    );
  }

  if (state.phase === 'between') {
    const last = state.results[state.results.length - 1];
    return (
      <Card><CardContent>
        <Typography variant="h5" gutterBottom>Таблица {last.number} из {TABLES} пройдена</Typography>
        <Typography paragraph>Время: <b>{formatSeconds(last.seconds)}</b>, ошибок: <b>{last.errors}</b>.</Typography>
        <LinearProgress variant="determinate" value={(state.results.length / TABLES) * 100} sx={{ mb: 3 }} />
        <Button variant="contained" size="large" onClick={start}>Таблица {state.tableIndex + 1} — начать</Button>
      </CardContent></Card>
    );
  }

  if (state.phase === 'finished') {
    return <Alert severity="info">Сохранение результатов…</Alert>;
  }

  const cells = tables[state.tableIndex];
  return (
    <Box>
      <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1, maxWidth: 560 }}>
        <Typography variant="h6">Таблица {state.tableIndex + 1} из {TABLES}</Typography>
        <Typography variant="h6" color="primary">Найдите: {state.next}</Typography>
        {settings.showTimer && (
          <Typography variant="h6" sx={{ fontVariantNumeric: 'tabular-nums' }}>{formatStopwatch(elapsed)}</Typography>
        )}
      </Stack>
      <Paper
        elevation={2}
        data-testid="schulte-table"
        sx={{
          display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: 0.5, p: 0.5, maxWidth: 560,
          backgroundColor: settings.backgroundColor, userSelect: 'none',
        }}
      >
        {cells.map((value) => {
          const found = value < state.next;
          const wrong = state.lastWrong === value;
          return (
            <Box
              key={value} role="button" tabIndex={0} aria-label={`число ${value}`}
              onClick={() => pick(value)}
              onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); pick(value); } }}
              sx={{
                aspectRatio: '1 / 1', display: 'flex', alignItems: 'center', justifyContent: 'center',
                fontSize: settings.fontSize, fontWeight: 600, cursor: 'pointer', borderRadius: 1,
                border: '1px solid', borderColor: 'divider',
                bgcolor: wrong ? 'error.light' : 'transparent',
                color: found ? 'text.disabled' : 'text.primary',
                '&:hover': { bgcolor: wrong ? 'error.light' : 'action.hover' },
              }}
            >
              {value}
            </Box>
          );
        })}
      </Paper>
      <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
        Ошибок в этой таблице: {state.errors}
      </Typography>
    </Box>
  );
}
