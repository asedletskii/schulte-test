import { Link as RouterLink, useParams } from 'react-router-dom';
import {
  Alert, Box, Button, Card, CardContent, CircularProgress, Paper, Table, TableBody, TableCell, TableContainer,
  TableHead, TableRow, Typography,
} from '@mui/material';
import { api } from '../api/client';
import type { Subject, TestSession } from '../api/types';
import { useAuth } from '../auth/AuthContext';
import { FatigueChart } from '../components/FatigueChart';
import { MetricCards } from '../components/MetricCards';
import { PageHeader } from '../components/PageHeader';
import { useAsync } from '../hooks/useAsync';
import { formatDateTime, formatSeconds } from '../utils/format';

export function ResultDetailsPage() {
  const { id = '' } = useParams();
  const { role } = useAuth();
  const result = useAsync<TestSession>(() => api.getResult(id), [id]);
  const subject = useAsync<Subject | null>(
    () => (result.data ? api.getSubject(result.data.subjectId) : Promise.resolve(null)),
    [result.data?.subjectId],
  );

  if (result.loading) return <CircularProgress />;
  if (result.error || !result.data?.metrics) return <Alert severity="error">{result.error ?? 'Результат не найден'}</Alert>;
  const t = result.data;
  const m = t.metrics!;

  return (
    <>
      <PageHeader
        title="Результат тестирования"
        subtitle={`${subject.data?.fullName ?? ''} · ${formatDateTime(t.finishedAt ?? t.startedAt)}`}
      />
      <MetricCards m={m} />
      <Card sx={{ mt: 3 }}><CardContent>
        <Typography variant="h6">Кривая утомляемости</Typography>
        <Typography variant="body2" color="text.secondary">
          Рост времени к концу теста говорит об утомляемости, резкие скачки — о неустойчивости внимания.
        </Typography>
        <FatigueChart curve={m.fatigueCurve} average={m.workEfficiency} />
      </CardContent></Card>

      <TableContainer component={Paper} variant="outlined" sx={{ mt: 3 }}>
        <Table size="small">
          <TableHead><TableRow><TableCell>Таблица</TableCell><TableCell align="right">Время</TableCell><TableCell align="right">Ошибок</TableCell></TableRow></TableHead>
          <TableBody>
            {(t.results ?? []).map((r) => (
              <TableRow key={r.number}><TableCell>№ {r.number}</TableCell><TableCell align="right">{formatSeconds(r.seconds)}</TableCell><TableCell align="right">{r.errors}</TableCell></TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      <Box sx={{ mt: 3, display: 'flex', gap: 2, flexWrap: 'wrap' }}>
        <Button variant="outlined" component={RouterLink} to="/results" state={{ subjectId: t.subjectId }}>К списку результатов</Button>
        {role === 'admin' && <Button variant="outlined" component={RouterLink} to="/dynamics" state={{ subjectId: t.subjectId }}>Динамика</Button>}
        <Button variant="outlined" onClick={() => window.print()}>Печать</Button>
      </Box>
    </>
  );
}
