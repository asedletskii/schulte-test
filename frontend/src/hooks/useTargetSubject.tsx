import { useState, type ReactNode } from 'react';
import { useLocation } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';
import { SubjectPicker } from '../components/SubjectPicker';

export function useTargetSubject(): { subjectId: string | null; picker: ReactNode } {
  const { role, subjectId } = useAuth();
  const location = useLocation();
  const fromNav = (location.state as { subjectId?: string } | null)?.subjectId ?? null;
  const [selected, setSelected] = useState<string | null>(fromNav);
  if (role === 'subject') return { subjectId, picker: null };
  return { subjectId: selected, picker: <SubjectPicker value={selected} onChange={setSelected} /> };
}
