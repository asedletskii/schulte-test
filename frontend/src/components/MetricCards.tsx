import { Card, CardContent, Chip, Grid, Typography } from '@mui/material';
import type { Metrics } from '../api/types';
import { formatNumber, formatSeconds, stabilityText, warmUpText } from '../utils/format';

function Tile(props: { title: string; value: string; hint: string; level?: 'good' | 'bad' }) {
  return (
    <Card sx={{ height: '100%' }}>
      <CardContent>
        <Typography variant="overline" color="text.secondary">{props.title}</Typography>
        <Typography variant="h4" sx={{ my: 0.5 }}>{props.value}</Typography>
        {props.level && (
          <Chip
            size="small" color={props.level === 'good' ? 'success' : 'warning'} label={props.hint}
            sx={{ height: 'auto', '& .MuiChip-label': { whiteSpace: 'normal', py: 0.5 } }}
          />
        )}
        {!props.level && <Typography variant="body2" color="text.secondary">{props.hint}</Typography>}
      </CardContent>
    </Card>
  );
}

export function MetricCards({ m }: { m: Metrics }) {
  return (
    <Grid container spacing={2}>
      <Grid item xs={12} sm={6} lg={3}>
        <Tile title="Эффективность работы (ЭР)" value={formatSeconds(m.workEfficiency)} hint="Среднее время на одну таблицу" />
      </Grid>
      <Grid item xs={12} sm={6} lg={3}>
        <Tile title="Врабатываемость (ВР)" value={formatNumber(m.warmUp)} level={m.warmUpLevel} hint={warmUpText(m.warmUpLevel)} />
      </Grid>
      <Grid item xs={12} sm={6} lg={3}>
        <Tile title="Психическая устойчивость (ПУ)" value={formatNumber(m.stability)} level={m.stabilityLevel} hint={stabilityText(m.stabilityLevel)} />
      </Grid>
      <Grid item xs={12} sm={6} lg={3}>
        <Tile title="Всего ошибок" value={String(m.totalErrors)} hint={`Общее время: ${formatSeconds(m.totalSeconds, 1)}`} />
      </Grid>
    </Grid>
  );
}
