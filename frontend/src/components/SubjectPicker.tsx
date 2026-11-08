import { Alert, Autocomplete, CircularProgress, TextField } from '@mui/material';
import { api } from '../api/client';
import type { Subject } from '../api/types';
import { useAsync } from '../hooks/useAsync';

interface Props {
  value: string | null;
  onChange: (id: string | null) => void;
  label?: string;
}

export function SubjectPicker({ value, onChange, label = 'Испытуемый' }: Props) {
  const { data, error, loading } = useAsync<Subject[]>(() => api.listSubjects(), []);
  if (error) return <Alert severity="error">{error}</Alert>;
  const options = data ?? [];
  return (
    <Autocomplete
      options={options}
      loading={loading}
      value={options.find((s) => s.id === value) ?? null}
      onChange={(_, s) => onChange(s?.id ?? null)}
      getOptionLabel={(s) => (s.group ? `${s.fullName} (${s.group})` : s.fullName)}
      isOptionEqualToValue={(a, b) => a.id === b.id}
      noOptionsText="Испытуемых нет — добавьте их в разделе «Испытуемые»"
      renderInput={(p) => (
        <TextField
          {...p} label={label}
          InputProps={{ ...p.InputProps, endAdornment: (<>{loading && <CircularProgress size={18} />}{p.InputProps.endAdornment}</>) }}
        />
      )}
      sx={{ maxWidth: 480 }}
    />
  );
}
