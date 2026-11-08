import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import { api, setToken, setUnauthorizedHandler } from '../api/client';
import type { AuthSession, Role } from '../api/types';

interface AuthState {
  role: Role | null;
  subjectId: string | null;
  loginAdmin: (password: string) => Promise<void>;
  loginSubject: (subjectId: string) => Promise<void>;
  logout: () => void;
}

const STORAGE_KEY = 'schulte.session';
const AuthContext = createContext<AuthState | null>(null);

function loadSession(): AuthSession | null {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    const s = JSON.parse(raw) as AuthSession;
    return new Date(s.expiresAt).getTime() > Date.now() ? s : null;
  } catch {
    return null;
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<AuthSession | null>(() => {
    const s = loadSession();
    setToken(s?.token ?? null);
    return s;
  });

  const apply = useCallback((s: AuthSession | null) => {
    setToken(s?.token ?? null);
    if (s) sessionStorage.setItem(STORAGE_KEY, JSON.stringify(s));
    else sessionStorage.removeItem(STORAGE_KEY);
    setSession(s);
  }, []);

  useEffect(() => {
    setUnauthorizedHandler(() => apply(null));
    return () => setUnauthorizedHandler(null);
  }, [apply]);

  const value = useMemo<AuthState>(
    () => ({
      role: session?.role ?? null,
      subjectId: session?.subjectId ?? null,
      loginAdmin: async (password) => apply(await api.loginAdmin(password)),
      loginSubject: async (id) => apply(await api.loginSubject(id)),
      logout: () => apply(null),
    }),
    [session, apply],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth должен вызываться внутри AuthProvider');
  return ctx;
}
