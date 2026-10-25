import { Navigate, Route, Routes } from 'react-router-dom';
import type { ReactElement } from 'react';
import { useAuth } from './auth/AuthContext';
import { Layout } from './components/Layout';
import { LoginPage } from './pages/LoginPage';
import { SubjectsPage } from './pages/SubjectsPage';
import { TestPage } from './pages/TestPage';
import { ResultsPage } from './pages/ResultsPage';
import { ResultDetailsPage } from './pages/ResultDetailsPage';
import { DynamicsPage } from './pages/DynamicsPage';
import { ComparePage } from './pages/ComparePage';
import { GroupPage } from './pages/GroupPage';
import { SettingsPage } from './pages/SettingsPage';
import { HelpPage } from './pages/HelpPage';

function AdminOnly({ children }: { children: ReactElement }) {
  const { role } = useAuth();
  return role === 'admin' ? children : <Navigate to="/test" replace />;
}

export default function App() {
  const { role } = useAuth();
  if (!role) {
    return (
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    );
  }
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/test" element={<TestPage />} />
        <Route path="/results" element={<ResultsPage />} />
        <Route path="/results/:id" element={<ResultDetailsPage />} />
        <Route path="/dynamics" element={<DynamicsPage />} />
        <Route path="/help" element={<HelpPage />} />
        <Route path="/subjects" element={<AdminOnly><SubjectsPage /></AdminOnly>} />
        <Route path="/compare" element={<AdminOnly><ComparePage /></AdminOnly>} />
        <Route path="/groups" element={<AdminOnly><GroupPage /></AdminOnly>} />
        <Route path="/settings" element={<AdminOnly><SettingsPage /></AdminOnly>} />
        <Route path="*" element={<Navigate to={role === 'admin' ? '/subjects' : '/test'} replace />} />
      </Route>
    </Routes>
  );
}
