import { Box, Typography } from '@mui/material';
import type { ReactNode } from 'react';

export function PageHeader({ title, subtitle, children }: { title: string; subtitle?: string; children?: ReactNode }) {
  return (
    <Box sx={{ mb: 3 }}>
      <Typography variant="h4" component="h1" gutterBottom>{title}</Typography>
      {subtitle && <Typography color="text.secondary" sx={{ mb: children ? 2 : 0 }}>{subtitle}</Typography>}
      {children}
    </Box>
  );
}
