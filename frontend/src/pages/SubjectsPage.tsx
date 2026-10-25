import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Alert, Autocomplete, Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, IconButton, Paper,
  Stack, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, TextField, Tooltip, Typography,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import AssessmentIcon from '@mui/icons-material/Assessment';
import { api } from '../api/client';
import type { Subject, SubjectInput } from '../api/types';
import { PageHeader } from '../components/PageHeader';
import { useAsync } from '../hooks/useAsync';

const EMPTY: SubjectInput = { fullName: '', birthYear: 0, group: '', notes: '' };

function SubjectDialog(props: { open: boolean; initial: Subject | null; groups: string[]; onClose: () => void; onSaved: () => void }) {
  const { open, initial, groups, onClose, onSaved } = props;
  const [form, setForm] = useState<SubjectInput>(EMPTY);
  const [error, setError] = useState<string | null>(null);
  const [prevKey, setPrevKey] = useState('');

  // сброс формы при каждом открытии диалога
  const key = `${open}-${initial?.id ?? 'new'}`;
  if (key !== prevKey) {
    setPrevKey(key);
    setError(null);
    setForm(initial ? { fullName: initial.fullName, birthYear: initial.birthYear, group: initial.group, notes: initial.notes } : EMPTY);
  }

  const save = async () => {
    try {
      if (initial) await api.updateSubject(initial.id, form);
      else await api.createSubject(form);
      onSaved();
    } catch (e) {
      setError((e as Error).message);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>{initial ? 'Изменить данные испытуемого' : 'Новый испытуемый'}</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ mt: 1 }}>
          <TextField label="ФИО" required autoFocus value={form.fullName} onChange={(e) => setForm({ ...form, fullName: e.target.value })} />
          <TextField
            label="Год рождения" type="number" value={form.birthYear || ''}
            onChange={(e) => setForm({ ...form, birthYear: Number(e.target.value) || 0 })}
          />
          <Autocomplete
            freeSolo options={groups} inputValue={form.group}
            onInputChange={(_, v) => setForm({ ...form, group: v })}
            renderInput={(p) => <TextField {...p} label="Группа" />}
          />
          <TextField label="Примечание" multiline minRows={2} value={form.notes} onChange={(e) => setForm({ ...form, notes: e.target.value })} />
          {error && <Alert severity="error">{error}</Alert>}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Отмена</Button>
        <Button variant="contained" onClick={save} disabled={!form.fullName.trim()}>Сохранить</Button>
      </DialogActions>
    </Dialog>
  );
}

export function SubjectsPage() {
  const navigate = useNavigate();
  const [group, setGroup] = useState('');
  const subjects = useAsync<Subject[]>(() => api.listSubjects(group), [group]);
  const groups = useAsync<string[]>(() => api.listGroups(), []);
  const [editing, setEditing] = useState<Subject | null>(null);
  const [creating, setCreating] = useState(false);
  const [toDelete, setToDelete] = useState<Subject | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const refresh = () => { subjects.reload(); groups.reload(); };
  const confirmDelete = async () => {
    if (!toDelete) return;
    try {
      await api.deleteSubject(toDelete.id);
      setToDelete(null);
      refresh();
    } catch (e) {
      setActionError((e as Error).message);
    }
  };

  return (
    <>
      <PageHeader title="Испытуемые" subtitle="Ведение данных о лицах, проходящих тестирование.">
        <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} alignItems={{ sm: 'center' }}>
          <Button variant="contained" startIcon={<AddIcon />} onClick={() => setCreating(true)}>Добавить</Button>
          <Autocomplete
            sx={{ width: 260 }} size="small" options={groups.data ?? []} value={group || null}
            onChange={(_, v) => setGroup(v ?? '')} renderInput={(p) => <TextField {...p} label="Фильтр по группе" />}
          />
        </Stack>
      </PageHeader>

      {(subjects.error || actionError) && <Alert severity="error" sx={{ mb: 2 }}>{subjects.error ?? actionError}</Alert>}
      <TableContainer component={Paper} variant="outlined">
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>ФИО</TableCell><TableCell>Год рожд.</TableCell><TableCell>Группа</TableCell>
              <TableCell>Примечание</TableCell><TableCell align="right">Действия</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {(subjects.data ?? []).map((s) => (
              <TableRow key={s.id} hover>
                <TableCell>{s.fullName}</TableCell>
                <TableCell>{s.birthYear || '—'}</TableCell>
                <TableCell>{s.group || '—'}</TableCell>
                <TableCell sx={{ maxWidth: 260, overflow: 'hidden', textOverflow: 'ellipsis' }}>{s.notes || '—'}</TableCell>
                <TableCell align="right" sx={{ whiteSpace: 'nowrap' }}>
                  <Tooltip title="Провести тест"><IconButton onClick={() => navigate('/test', { state: { subjectId: s.id } })}><PlayArrowIcon /></IconButton></Tooltip>
                  <Tooltip title="Результаты"><IconButton onClick={() => navigate('/results', { state: { subjectId: s.id } })}><AssessmentIcon /></IconButton></Tooltip>
                  <Tooltip title="Изменить"><IconButton onClick={() => setEditing(s)}><EditIcon /></IconButton></Tooltip>
                  <Tooltip title="Удалить"><IconButton color="error" onClick={() => setToDelete(s)}><DeleteIcon /></IconButton></Tooltip>
                </TableCell>
              </TableRow>
            ))}
            {!subjects.loading && (subjects.data ?? []).length === 0 && (
              <TableRow><TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                <Typography color="text.secondary">Испытуемых пока нет. Нажмите «Добавить».</Typography>
              </TableCell></TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <SubjectDialog
        open={creating || editing !== null} initial={editing} groups={groups.data ?? []}
        onClose={() => { setCreating(false); setEditing(null); }}
        onSaved={() => { setCreating(false); setEditing(null); refresh(); }}
      />
      <Dialog open={toDelete !== null} onClose={() => setToDelete(null)}>
        <DialogTitle>Удалить испытуемого?</DialogTitle>
        <DialogContent>
          <Typography>
            «{toDelete?.fullName}» и все его результаты тестирования будут удалены без возможности восстановления.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Box sx={{ flexGrow: 1 }} />
          <Button onClick={() => setToDelete(null)}>Отмена</Button>
          <Button color="error" variant="contained" onClick={confirmDelete}>Удалить</Button>
        </DialogActions>
      </Dialog>
    </>
  );
}
