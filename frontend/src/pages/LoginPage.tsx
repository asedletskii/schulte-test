import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Alert, Autocomplete, Box, Button, Card, CardContent, Container, Tab, Tabs, TextField, Typography,
} from '@mui/material';
import { useAuth } from '../auth/AuthContext';
import { api } from '../api/client';
import type { PublicSubject } from '../api/types';
import { useAsync } from '../hooks/useAsync';

export function LoginPage() {
  const { loginAdmin, loginSubject } = useAuth();
  const navigate = useNavigate();
  const [tab, setTab] = useState<'subject' | 'admin'>('subject');
  const [password, setPassword] = useState('');
  const [subject, setSubject] = useState<PublicSubject | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const { data: subjects, error: listError } = useAsync<PublicSubject[]>(() => api.publicSubjects(), []);

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setBusy(true);
    try {
      if (tab === 'admin') await loginAdmin(password);
      else if (subject) await loginSubject(subject.id);
      navigate(tab === 'admin' ? '/subjects' : '/test', { replace: true });
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <Container maxWidth="xs" sx={{ minHeight: '100vh', display: 'flex', alignItems: 'center' }}>
      <Card sx={{ width: '100%' }}>
        <CardContent component="form" onSubmit={submit} sx={{ p: 3 }}>
          <Typography variant="h5" align="center" gutterBottom>Таблицы Шульте</Typography>
          <Typography align="center" color="text.secondary" sx={{ mb: 2 }}>Тестирование устойчивости внимания</Typography>
          <Tabs value={tab} onChange={(_, v) => { setTab(v); setError(null); }} variant="fullWidth" sx={{ mb: 2 }}>
            <Tab value="subject" label="Испытуемый" />
            <Tab value="admin" label="Психолог" />
          </Tabs>
          {tab === 'subject' ? (
            <Autocomplete
              options={subjects ?? []} value={subject} onChange={(_, v) => setSubject(v)}
              getOptionLabel={(s) => (s.group ? `${s.fullName} (${s.group})` : s.fullName)}
              noOptionsText="Список пуст — попросите психолога добавить вас"
              renderInput={(p) => <TextField {...p} label="Выберите себя в списке" autoFocus />}
            />
          ) : (
            <TextField
              fullWidth type="password" label="Пароль психолога" value={password} autoFocus
              onChange={(e) => setPassword(e.target.value)} autoComplete="current-password"
            />
          )}
          {(error || listError) && <Alert severity="error" sx={{ mt: 2 }}>{error ?? listError}</Alert>}
          <Box sx={{ mt: 3 }}>
            <Button
              type="submit" variant="contained" fullWidth size="large"
              disabled={busy || (tab === 'admin' ? !password : !subject)}
            >
              Войти
            </Button>
          </Box>
        </CardContent>
      </Card>
    </Container>
  );
}
