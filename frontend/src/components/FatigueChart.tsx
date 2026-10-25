import { LineChart } from '@mui/x-charts/LineChart';

/** Кривая утомляемости: время прохождения каждой из пяти таблиц. */
export function FatigueChart({ curve, average }: { curve: number[]; average?: number }) {
  const x = curve.map((_, i) => i + 1);
  const series = [{ curve: 'linear' as const, data: curve, label: 'Время на таблицу, с', color: '#1565c0', valueFormatter: (v: number | null) => (v == null ? '' : `${v.toFixed(2)} с`) }];
  if (average !== undefined) {
    series.push({ curve: 'linear' as const, data: curve.map(() => average), label: 'ЭР (среднее)', color: '#00897b', valueFormatter: (v: number | null) => (v == null ? '' : `${v.toFixed(2)} с`) });
  }
  return (
    <LineChart
      height={300}
      xAxis={[{ data: x, scaleType: 'point', label: 'Номер таблицы' }]}
      yAxis={[{ label: 'Время, с', min: 0 }]}
      series={series}
      margin={{ left: 60, right: 20, top: 20, bottom: 50 }}
    />
  );
}
