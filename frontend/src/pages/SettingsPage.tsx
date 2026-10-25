import { useEffect, useState } from 'react';
import {
  Alert, Box, Button, Card, CardContent, FormControlLabel, Slider, Snackbar, Stack, Switch, TextField, Typography,
} from '@mui/material';
import { api } from '../api/client';
import type { Settings } from '../api/types';
import { PageHeader } from '../components/PageHeader';

export function SettingsPage() {
  const [s, setS] = useState<Settings | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    api.getSettings().then(setS).catch((e: Error) => setError(e.message));
  }, []);

  const save = async () => {
    if (!s) return;
    try {
      setS(await api.updateSettings(s));
      setError(null);
      setSaved(true);
    } catch (e) {
      setError((e as Error).message);
    }
  };

  return (
    <>
      <PageHeader title="Настройки тестирования" subtitle="Параметры по умолчанию для всех испытуемых. Перед началом теста их можно изменить." />
      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
      {s && (
        <Card sx={{ maxWidth: 640 }}><CardContent>
          <Stack spacing={3}>
            <Box>
              <Typography gutterBottom>Размер шрифта: {s.fontSize} px</Typography>
              <Slider min={14} max={64} value={s.fontSize} onChange={(_, v) => setS({ ...s, fontSize: v as number })} />
            </Box>
            <TextField label="Цвет фона таблицы" type="color" value={s.backgroundColor} sx={{ width: 200 }} onChange={(e) => setS({ ...s, backgroundColor: e.target.value })} />
            <FormControlLabel control={<Switch checked={s.showTimer} onChange={(e) => setS({ ...s, showTimer: e.target.checked })} />} label="Показывать секундомер во время теста" />
            <FormControlLabel control={<Switch checked={s.soundOnError} onChange={(e) => setS({ ...s, soundOnError: e.target.checked })} />} label="Звуковой сигнал при ошибочном выборе" />
            <Box><Button variant="contained" onClick={save}>Сохранить</Button></Box>
          </Stack>
        </CardContent></Card>
      )}
      <Snackbar open={saved} autoHideDuration={2500} onClose={() => setSaved(false)} message="Настройки сохранены" />
    </>
  );
}
