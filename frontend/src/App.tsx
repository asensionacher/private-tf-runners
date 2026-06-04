import { Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './hooks/useAuth';
import LoginPage from './pages/LoginPage';
import Dashboard from './pages/Dashboard';
import StacksList from './pages/StacksList';
import StackDetail from './pages/StackDetail';
import RunsList from './pages/RunsList';
import RunDetail from './pages/RunDetail';
import CreateStackPage from './pages/CreateStackPage';
import CreateRunPage from './pages/CreateRunPage';
import RunnersList from './pages/RunnersList';
import UsersList from './pages/UsersList';
import Layout from './components/Layout';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Dashboard />} />
        <Route path="stacks" element={<StacksList />} />
        <Route path="stacks/new" element={<CreateStackPage />} />
        <Route path="stacks/:id" element={<StackDetail />} />
        <Route path="runs/new" element={<CreateRunPage />} />
        <Route path="runs" element={<RunsList />} />
        <Route path="runs/:id" element={<RunDetail />} />
        <Route path="runners" element={<RunnersList />} />
        <Route path="users" element={<UsersList />} />
      </Route>
    </Routes>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <AppRoutes />
    </AuthProvider>
  );
}
