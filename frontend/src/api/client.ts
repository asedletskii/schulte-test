import type {
  AuthSession, Comparison, Dynamics, GroupStats, PublicSubject, Settings, Subject, SubjectInput,
  TableResult, TestSession,
} from './types';

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
  }
}

let token: string | null = null;
let onUnauthorized: (() => void) | null = null;

export function setToken(t: string | null) {
  token = t;
}

export function setUnauthorizedHandler(fn: (() => void) | null) {
  onUnauthorized = fn;
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = {};
  if (body !== undefined) headers['Content-Type'] = 'application/json';
  if (token) headers['Authorization'] = `Bearer ${token}`;

  let resp: Response;
  try {
    resp = await fetch(`/api${path}`, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch {
    throw new ApiError(0, 'Сервер недоступен. Проверьте подключение.');
  }

  if (resp.status === 204) return undefined as T;
  const data = await resp.json().catch(() => ({}));
  if (!resp.ok) {
    if (resp.status === 401 && token) onUnauthorized?.();
    throw new ApiError(resp.status, data.error ?? `Ошибка ${resp.status}`);
  }
  return data as T;
}

const q = encodeURIComponent;

export const api = {
  loginAdmin: (password: string) => request<AuthSession>('POST', '/auth/admin', { password }),
  loginSubject: (subjectId: string) => request<AuthSession>('POST', '/auth/subject', { subjectId }),
  publicSubjects: () => request<PublicSubject[]>('GET', '/public/subjects'),

  listSubjects: (group = '') => request<Subject[]>('GET', `/subjects${group ? `?group=${q(group)}` : ''}`),
  getSubject: (id: string) => request<Subject>('GET', `/subjects/${q(id)}`),
  createSubject: (s: SubjectInput) => request<Subject>('POST', '/subjects', s),
  updateSubject: (id: string, s: SubjectInput) => request<Subject>('PUT', `/subjects/${q(id)}`, s),
  deleteSubject: (id: string) => request<void>('DELETE', `/subjects/${q(id)}`),
  listGroups: () => request<string[]>('GET', '/groups'),
  groupStats: (name: string) => request<GroupStats>('GET', `/groups/${q(name)}/stats`),

  getSettings: () => request<Settings>('GET', '/settings'),
  updateSettings: (s: Settings) => request<Settings>('PUT', '/settings', s),

  startTest: (subjectId: string, settings?: Settings) =>
    request<TestSession>('POST', '/tests', { subjectId, settings }),
  finishTest: (id: string, results: TableResult[]) =>
    request<TestSession>('POST', `/tests/${q(id)}/finish`, { results }),
  listResults: (subjectId: string) => request<TestSession[]>('GET', `/subjects/${q(subjectId)}/results`),
  getResult: (id: string) => request<TestSession>('GET', `/results/${q(id)}`),
  deleteResult: (id: string) => request<void>('DELETE', `/results/${q(id)}`),
  dynamics: (subjectId: string) => request<Dynamics>('GET', `/subjects/${q(subjectId)}/dynamics`),
  compare: (a: string, b: string) => request<Comparison>('GET', `/compare?a=${q(a)}&b=${q(b)}`),
};
