import { useNavigate } from 'react-router-dom';
import {
  Alert, Button, IconButton, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow,
  Tooltip, Typography,
} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import { api } from '../api/client';
import type { TestSession } from '../api/types';
import { useAuth } from '../auth/AuthContext';
import { PageHeader } from '../components/PageHeader';
import { useAsync } from '../hooks/useAsync';
import { useTargetSubject } from '../hooks/useTargetSubject';
import { formatDateTime, formatNumber, formatSeconds } from '../utils/format';

export function ResultsPage() {
  const navigate = useNavigate();
  const { role } = useAuth();
  const { subjectId, picker } = useTargetSubject();
  const list = useAsync<TestSession[]>(() => (subjectId ? api.listResults(subjectId) : Promise.resolve([])), [subjectId]);

  const remove = async (id: string) => {
    if (!window.confirm('Удалить результат тестирования?')) return;
    await api.deleteResult(id);
    list.reload();
  };

  return (
    <>
      <PageHeader title="Результаты тестирования" subtitle="Сохранённые результаты. Выберите тест, чтобы увидеть подробности.">
        {picker}
      </PageHeader>
      {list.error && <Alert severity="error" sx={{ mb: 2 }}>{list.error}</Alert>}
      {!subjectId ? (
        <Typography color="text.secondary">Выберите испытуемого.</Typography>
      ) : (
        <TableContainer component={Paper} variant="outlined">
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Дата</TableCell><TableCell align="right">ЭР</TableCell><TableCell align="right">ВР</TableCell>
                <TableCell align="right">ПУ</TableCell><TableCell align="right">Ошибок</TableCell><TableCell />
              </TableRow>
            </TableHead>
            <TableBody>
              {(list.data ?? []).map((t) => (
                <TableRow key={t.id} hover>
                  <TableCell>{formatDateTime(t.finishedAt ?? t.startedAt)}</TableCell>
                  <TableCell align="right">{formatSeconds(t.metrics!.workEfficiency)}</TableCell>
                  <TableCell align="right">{formatNumber(t.metrics!.warmUp)}</TableCell>
                  <TableCell align="right">{formatNumber(t.metrics!.stability)}</TableCell>
                  <TableCell align="right">{t.metrics!.totalErrors}</TableCell>
                  <TableCell align="right" sx={{ whiteSpace: 'nowrap' }}>
                    <Button size="small" onClick={() => navigate(`/results/${t.id}`)}>Открыть</Button>
                    {role === 'admin' && (
                      <Tooltip title="Удалить"><IconButton size="small" color="error" onClick={() => remove(t.id)}><DeleteIcon fontSize="small" /></IconButton></Tooltip>
                    )}
                  </TableCell>
                </TableRow>
              ))}
              {!list.loading && (list.data ?? []).length === 0 && (
                <TableRow><TableCell colSpan={6} align="center" sx={{ py: 4 }}>
                  <Typography color="text.secondary">Тестирований пока нет.</Typography>
                </TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </>
  );
}
