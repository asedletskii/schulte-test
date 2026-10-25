// Типы данных REST API (соответствуют DTO backend).

export type Role = 'admin' | 'subject';

export interface AuthSession {
  token: string;
  expiresAt: string;
  role: Role;
  subjectId?: string;
}

export interface Subject {
  id: string;
  fullName: string;
  birthYear: number;
  group: string;
  notes: string;
  createdAt: string;
}

export interface PublicSubject {
  id: string;
  fullName: string;
  group: string;
}

export type SubjectInput = Pick<Subject, 'fullName' | 'birthYear' | 'group' | 'notes'>;

export interface Settings {
  fontSize: number;
  backgroundColor: string;
  showTimer: boolean;
  soundOnError: boolean;
}

export interface TableResult {
  number: number;
  seconds: number;
  errors: number;
}

export type Level = 'good' | 'bad';

export interface Metrics {
  totalSeconds: number;
  workEfficiency: number;
  warmUp: number;
  stability: number;
  warmUpLevel: Level;
  stabilityLevel: Level;
  fatigueCurve: number[];
  totalErrors: number;
}

export interface TestSession {
  id: string;
  subjectId: string;
  status: 'in_progress' | 'completed';
  tables?: number[][];
  settings: Settings;
  startedAt: string;
  finishedAt?: string;
  results?: TableResult[];
  metrics?: Metrics;
}

export interface DynamicsPoint {
  sessionId: string;
  date: string;
  workEfficiency: number;
  warmUp: number;
  stability: number;
  errors: number;
}

export type Trend = 'improving' | 'worsening' | 'stable' | 'unknown';

export interface Dynamics {
  points: DynamicsPoint[];
  slope: number;
  trend: Trend;
}

export interface Comparison {
  a: TestSession;
  b: TestSession;
  diffEfficiency: number;
  diffWarmUp: number;
  diffStability: number;
  diffErrors: number;
  better: 'A' | 'B' | 'equal';
}

export interface GroupMember {
  subject: Subject;
  tests: number;
  latest?: Metrics;
  rank: number;
}

export interface GroupStats {
  group: string;
  members: GroupMember[];
  avgEfficiency: number;
  avgWarmUp: number;
  avgStability: number;
  participants: number;
}
