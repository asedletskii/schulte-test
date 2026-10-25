import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Accordion, AccordionDetails, AccordionSummary, Alert, Box, Button, FormControlLabel, Slider, Stack, Switch,
  TextField, Typography,
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import { api } from '../api/client';
import type { Settings, TableResult, TestSession } from '../api/types';
import { PageHeader } from '../components/PageHeader';
import { useTargetSubject } from '../hooks/useTargetSubject';
import { TestRunner } from '../features/test/TestRunner';

export function TestPage() {
  const navigate = useNavigate();
  const { subjectId, picker } = useTargetSubject();
  const [settings, setSettings] = useState<Settings | null>(null);
  const [session, setSession] = useState<TestSession | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    api.getSettings().then(setSettings).catch((e: Error) => setError(e.message));
  }, []);

  const start = async () => {
    if (!subjectId || !settings) return;
    setError(null);
    setBusy(true);
    try {
      setSession(await api.startTest(subjectId, settings));
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  };

  const finish = useCallback(async (results: TableResult[]) => {
    if (!session) return;
    try {
      const done = await api.finishTest(session.id, results);
      navigate(`/results/${done.id}`, { replace: true });
    } catch (e) {
      setError((e as Error).message);
    }
  }, [session, navigate]);

  if (session?.tables) {
    return (
      <>
        <PageHeader title="Тестирование по методике Шульте" />
        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
        <TestRunner tables={session.tables} settings={session.settings} onFinish={finish} />
      </>
    );
  }

  return (
    <>
      <PageHeader
        title="Пройти тест"
        subtitle="Пять таблиц с числами от 1 до 25. Нужно как можно быстрее находить числа по возрастанию."
      >
        {picker}
      </PageHeader>

      {settings && (
        <Accordion variant="outlined" sx={{ maxWidth: 640, mb: 3 }}>
          <AccordionSummary expandIcon={<ExpandMoreIcon />}>
            <Typography>Параметры отображения (только для этого теста)</Typography>
          </AccordionSummary>
          <AccordionDetails>
            <Stack spacing={2}>
              <Box>
                <Typography gutterBottom>Размер шрифта: {settings.fontSize} px</Typography>
                <Slider min={14} max={64} value={settings.fontSize} onChange={(_, v) => setSettings({ ...settings, fontSize: v as number })} />
              </Box>
              <TextField
                label="Цвет фона таблицы" type="color" value={settings.backgroundColor} sx={{ width: 200 }}
                onChange={(e) => setSettings({ ...settings, backgroundColor: e.target.value })}
              />
              <FormControlLabel control={<Switch checked={settings.showTimer} onChange={(e) => setSettings({ ...settings, showTimer: e.target.checked })} />} label="Показывать секундомер" />
              <FormControlLabel control={<Switch checked={settings.soundOnError} onChange={(e) => setSettings({ ...settings, soundOnError: e.target.checked })} />} label="Звуковой сигнал при ошибке" />
            </Stack>
          </AccordionDetails>
        </Accordion>
      )}

      {error && <Alert severity="error" sx={{ mb: 2, maxWidth: 640 }}>{error}</Alert>}
      <Button variant="contained" size="large" disabled={!subjectId || !settings || busy} onClick={start}>
        К тесту
      </Button>
      {!subjectId && <Typography color="text.secondary" sx={{ mt: 1 }}>Сначала выберите испытуемого.</Typography>}
    </>
  );
}
