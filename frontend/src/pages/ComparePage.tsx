import { useState } from 'react';
import {
  Alert, Autocomplete, Card, CardContent, Grid, Paper, Stack, Table, TableBody, TableCell, TableContainer,
  TableHead, TableRow, TextField, Typography,
} from '@mui/material';
import { BarChart } from '@mui/x-charts/BarChart';
import { api } from '../api/client';
import type { Comparison, TestSession } from '../api/types';
import { PageHeader } from '../components/PageHeader';
import { SubjectPicker } from '../components/SubjectPicker';
import { useAsync } from '../hooks/useAsync';
import { formatDateTime, formatNumber, formatSeconds } from '../utils/format';

function SessionSelect(props: { title: string; value: string | null; onChange: (id: string | null) => void }) {
  const [subjectId, setSubjectId] = useState<string | null>(null);
  const list = useAsync<TestSession[]>(() => (subjectId ? api.listResults(subjectId) : Promise.resolve([])), [subjectId]);
  return (
    <Stack spacing={2}>
      <Typography variant="h6">{props.title}</Typography>
      <SubjectPicker value={subjectId} onChange={(id) => { setSubjectId(id); props.onChange(null); }} />
      <Autocomplete
        sx={{ maxWidth: 480 }} disabled={!subjectId} options={list.data ?? []}
        value={(list.data ?? []).find((t) => t.id === props.value) ?? null}
        onChange={(_, t) => props.onChange(t?.id ?? null)}
        getOptionLabel={(t) => `${formatDateTime(t.finishedAt ?? t.startedAt)} — ЭР ${formatNumber(t.metrics!.workEfficiency)} с`}
        noOptionsText="Нет завершённых тестов"
        renderInput={(p) => <TextField {...p} label="Тестирование" />}
      />
    </Stack>
  );
}

function verdict(c: Comparison): string {
  if (c.better === 'equal') return 'Результаты A и B совпадают по эффективности работы.';
  const who = c.better === 'A' ? 'A' : 'B';
  return `Эффективность работы выше у ${who}: среднее время на таблицу меньше на ${formatSeconds(Math.abs(c.diffEfficiency))}.`;
}

export function ComparePage() {
  const [a, setA] = useState<string | null>(null);
  const [b, setB] = useState<string | null>(null);
  const cmp = useAsync<Comparison | null>(() => (a && b ? api.compare(a, b) : Promise.resolve(null)), [a, b]);
  const c = cmp.data;

  return (
    <>
      <PageHeader title="Сравнение результатов" subtitle="Сравните два тестирования одного или разных испытуемых." />
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}><SessionSelect title="Тестирование A" value={a} onChange={setA} /></Grid>
        <Grid item xs={12} md={6}><SessionSelect title="Тестирование B" value={b} onChange={setB} /></Grid>
      </Grid>
      {cmp.error && <Alert severity="error" sx={{ mt: 3 }}>{cmp.error}</Alert>}
      {c && c.a.metrics && c.b.metrics && (
        <>
          <Alert severity="info" sx={{ mt: 3 }}>{verdict(c)}</Alert>
          <TableContainer component={Paper} variant="outlined" sx={{ mt: 3 }}>
            <Table size="small">
              <TableHead><TableRow><TableCell>Показатель</TableCell><TableCell align="right">A</TableCell><TableCell align="right">B</TableCell><TableCell align="right">B − A</TableCell></TableRow></TableHead>
              <TableBody>
                <TableRow><TableCell>ЭР, с</TableCell><TableCell align="right">{formatNumber(c.a.metrics.workEfficiency)}</TableCell><TableCell align="right">{formatNumber(c.b.metrics.workEfficiency)}</TableCell><TableCell align="right">{formatNumber(c.diffEfficiency)}</TableCell></TableRow>
                <TableRow><TableCell>ВР</TableCell><TableCell align="right">{formatNumber(c.a.metrics.warmUp)}</TableCell><TableCell align="right">{formatNumber(c.b.metrics.warmUp)}</TableCell><TableCell align="right">{formatNumber(c.diffWarmUp)}</TableCell></TableRow>
                <TableRow><TableCell>ПУ</TableCell><TableCell align="right">{formatNumber(c.a.metrics.stability)}</TableCell><TableCell align="right">{formatNumber(c.b.metrics.stability)}</TableCell><TableCell align="right">{formatNumber(c.diffStability)}</TableCell></TableRow>
                <TableRow><TableCell>Ошибок</TableCell><TableCell align="right">{c.a.metrics.totalErrors}</TableCell><TableCell align="right">{c.b.metrics.totalErrors}</TableCell><TableCell align="right">{c.diffErrors}</TableCell></TableRow>
              </TableBody>
            </Table>
          </TableContainer>
          <Card sx={{ mt: 3 }}><CardContent>
            <Typography variant="h6">Время по таблицам, с</Typography>
            <BarChart
              height={300}
              xAxis={[{ scaleType: 'band', data: ['Таблица 1', 'Таблица 2', 'Таблица 3', 'Таблица 4', 'Таблица 5'] }]}
              series={[
                { data: c.a.metrics.fatigueCurve, label: 'A', valueFormatter: (v) => (v == null ? '' : formatNumber(v)) },
                { data: c.b.metrics.fatigueCurve, label: 'B', valueFormatter: (v) => (v == null ? '' : formatNumber(v)) },
              ]}
              margin={{ left: 60, right: 20, top: 20, bottom: 40 }}
            />
          </CardContent></Card>
        </>
      )}
    </>
  );
}
