import { useState } from 'react';
import {
  Alert, Autocomplete, Card, CardContent, Chip, Grid, Paper, Table, TableBody, TableCell, TableContainer,
  TableHead, TableRow, TextField, Typography,
} from '@mui/material';
import { BarChart } from '@mui/x-charts/BarChart';
import { api } from '../api/client';
import type { GroupStats } from '../api/types';
import { PageHeader } from '../components/PageHeader';
import { useAsync } from '../hooks/useAsync';
import { formatNumber } from '../utils/format';

export function GroupPage() {
  const groups = useAsync<string[]>(() => api.listGroups(), []);
  const [group, setGroup] = useState<string | null>(null);
  const stats = useAsync<GroupStats | null>(() => (group ? api.groupStats(group) : Promise.resolve(null)), [group]);
  const s = stats.data;
  const ranked = (s?.members ?? []).filter((m) => m.latest);

  return (
    <>
      <PageHeader title="Сравнение в группе" subtitle="Сравнение последних результатов участников группы по эффективности работы.">
        <Autocomplete
          sx={{ maxWidth: 360 }} options={groups.data ?? []} value={group} onChange={(_, v) => setGroup(v)}
          noOptionsText="Групп пока нет" renderInput={(p) => <TextField {...p} label="Группа" />}
        />
      </PageHeader>
      {(groups.error || stats.error) && <Alert severity="error">{groups.error ?? stats.error}</Alert>}
      {s && (
        <>
          <Grid container spacing={2} sx={{ mb: 3 }}>
            {[
              ['Участников с результатами', `${s.participants} из ${s.members.length}`],
              ['Средний ЭР, с', formatNumber(s.avgEfficiency)],
              ['Средняя ВР', formatNumber(s.avgWarmUp)],
              ['Средняя ПУ', formatNumber(s.avgStability)],
            ].map(([t, v]) => (
              <Grid item xs={6} md={3} key={t}><Card><CardContent>
                <Typography variant="overline" color="text.secondary">{t}</Typography>
                <Typography variant="h5">{v}</Typography>
              </CardContent></Card></Grid>
            ))}
          </Grid>
          <TableContainer component={Paper} variant="outlined">
            <Table size="small">
              <TableHead><TableRow>
                <TableCell>Место</TableCell><TableCell>ФИО</TableCell><TableCell align="right">Тестов</TableCell>
                <TableCell align="right">ЭР, с</TableCell><TableCell align="right">ВР</TableCell><TableCell align="right">ПУ</TableCell>
              </TableRow></TableHead>
              <TableBody>
                {[...s.members].sort((a, b) => (a.rank || 999) - (b.rank || 999)).map((m) => (
                  <TableRow key={m.subject.id}>
                    <TableCell>{m.rank ? <Chip size="small" label={m.rank} color={m.rank === 1 ? 'success' : 'default'} /> : '—'}</TableCell>
                    <TableCell>{m.subject.fullName}</TableCell>
                    <TableCell align="right">{m.tests}</TableCell>
                    <TableCell align="right">{m.latest ? formatNumber(m.latest.workEfficiency) : '—'}</TableCell>
                    <TableCell align="right">{m.latest ? formatNumber(m.latest.warmUp) : '—'}</TableCell>
                    <TableCell align="right">{m.latest ? formatNumber(m.latest.stability) : '—'}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
          {ranked.length > 0 && (
            <Card sx={{ mt: 3 }}><CardContent>
              <Typography variant="h6">ЭР участников, с</Typography>
              <BarChart
                height={300}
                xAxis={[{ scaleType: 'band', data: ranked.map((m) => m.subject.fullName) }]}
                series={[{ data: ranked.map((m) => m.latest!.workEfficiency), label: 'ЭР', valueFormatter: (v) => (v == null ? '' : formatNumber(v)) }]}
                margin={{ left: 60, right: 20, top: 20, bottom: 50 }}
              />
            </CardContent></Card>
          )}
        </>
      )}
    </>
  );
}
