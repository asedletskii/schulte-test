import { Alert, Card, CardContent, Typography } from '@mui/material';
import { LineChart } from '@mui/x-charts/LineChart';
import { api } from '../api/client';
import type { Dynamics } from '../api/types';
import { PageHeader } from '../components/PageHeader';
import { useAsync } from '../hooks/useAsync';
import { useTargetSubject } from '../hooks/useTargetSubject';
import { formatNumber, trendText } from '../utils/format';

export function DynamicsPage() {
  const { subjectId, picker } = useTargetSubject();
  const dyn = useAsync<Dynamics | null>(() => (subjectId ? api.dynamics(subjectId) : Promise.resolve(null)), [subjectId]);
  const d = dyn.data;

  const labels = (d?.points ?? []).map(
    (p, i) => `№${i + 1} · ${new Date(p.date).toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' })}`,
  );
  const fmt = (v: number | null) => (v == null ? '' : formatNumber(v));

  return (
    <>
      <PageHeader title="Динамика результатов" subtitle="Изменение характеристик внимания испытуемого с течением времени.">
        {picker}
      </PageHeader>
      {dyn.error && <Alert severity="error">{dyn.error}</Alert>}
      {!subjectId && <Typography color="text.secondary">Выберите испытуемого.</Typography>}
      {d && (
        <>
          <Alert severity={d.trend === 'improving' ? 'success' : d.trend === 'worsening' ? 'warning' : 'info'} sx={{ mb: 3 }}>
            {trendText(d.trend)}
            {d.points.length >= 2 && ` Изменение ЭР: ${formatNumber(d.slope)} с за тест.`}
          </Alert>
          {d.points.length > 0 && (
            <>
              <Card sx={{ mb: 3 }}><CardContent>
                <Typography variant="h6">Эффективность работы (ЭР), с</Typography>
                <LineChart
                  height={280} xAxis={[{ data: labels, scaleType: 'point' }]} yAxis={[{ min: 0 }]}
                  series={[{ data: d.points.map((p) => p.workEfficiency), label: 'ЭР, с', valueFormatter: fmt, curve: 'linear' }]}
                  margin={{ left: 60, right: 40, top: 20, bottom: 40 }}
                />
              </CardContent></Card>
              <Card><CardContent>
                <Typography variant="h6">Врабатываемость и психическая устойчивость</Typography>
                <Typography variant="body2" color="text.secondary">Значения ниже 1,0 — хороший результат.</Typography>
                <LineChart
                  height={280} xAxis={[{ data: labels, scaleType: 'point' }]} yAxis={[{ min: 0 }]}
                  series={[
                    { data: d.points.map((p) => p.warmUp), label: 'ВР', valueFormatter: fmt, curve: 'linear' },
                    { data: d.points.map((p) => p.stability), label: 'ПУ', valueFormatter: fmt, curve: 'linear' },
                  ]}
                  margin={{ left: 60, right: 40, top: 20, bottom: 40 }}
                />
              </CardContent></Card>
            </>
          )}
        </>
      )}
    </>
  );
}
