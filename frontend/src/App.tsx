import { useEffect } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';

import MainLayout from '@/components/layout/MainLayout';
import Login from '@/pages/auth/Login';
import ContainerDetail from '@/pages/containers/ContainerDetail';
import ContainerForm from '@/pages/containers/ContainerForm';
import ContainerList from '@/pages/containers/ContainerList';
import ContainerVersions from '@/pages/containers/ContainerVersions';
import Dashboard from '@/pages/dashboard/Dashboard';
import DatasetDetail from '@/pages/datasets/DatasetDetail';
import DatasetForm from '@/pages/datasets/DatasetForm';
import DatasetList from '@/pages/datasets/DatasetList';
import EvaluationForm from '@/pages/evaluations/EvaluationForm';
import EvaluationList from '@/pages/evaluations/EvaluationList';
import ExecutionDetail from '@/pages/executions/ExecutionDetail';
import ExecutionForm from '@/pages/executions/ExecutionForm';
import ExecutionList from '@/pages/executions/ExecutionList';
import InjectionCreate from '@/pages/injections/InjectionCreate';
import InjectionList from '@/pages/injections/InjectionList';
import ProjectList from '@/pages/projects/ProjectList';
import Settings from '@/pages/settings/Settings';
import UserProfile from '@/pages/settings/UserProfile';
import SystemSettings from '@/pages/system/SystemSettings';
import TaskDetail from '@/pages/tasks/TaskDetail';
import TaskList from '@/pages/tasks/TaskList';
import UtilityTest from '@/pages/UtilityTest';
import { useAuthStore } from '@/store/auth';

function App() {
  const { isAuthenticated, loadUser } = useAuthStore();

  useEffect(() => {
    if (isAuthenticated) {
      loadUser();
    }
  }, [isAuthenticated, loadUser]);

  // Debug logging
  useEffect(() => {
    // console.log('Auth state changed:', { isAuthenticated })
  }, [isAuthenticated]);

  return (
    <Routes>
      {/* Public routes */}
      <Route path='/login' element={<Login />} />

      {/* Protected routes - Authentication bypassed for development */}
      <Route path='/*' element={<MainLayout />}>
        <Route index element={<Navigate to='/dashboard' replace />} />
        <Route path='dashboard' element={<Dashboard />} />

        {/* Projects */}
        <Route path='projects' element={<ProjectList />} />

        {/* Containers */}
        <Route path='containers' element={<ContainerList />} />
        <Route path='containers/new' element={<ContainerForm />} />
        <Route path='containers/:id' element={<ContainerDetail />} />
        <Route path='containers/:id/edit' element={<ContainerForm />} />
        <Route path='containers/:id/versions' element={<ContainerVersions />} />

        {/* Datasets */}
        <Route path='datasets' element={<DatasetList />} />
        <Route path='datasets/new' element={<DatasetForm />} />
        <Route path='datasets/:id' element={<DatasetDetail />} />
        <Route path='datasets/:id/edit' element={<DatasetForm />} />

        {/* Injections */}
        <Route path='injections' element={<InjectionList />} />
        <Route path='injections/create' element={<InjectionCreate />} />

        {/* Executions */}
        <Route path='executions' element={<ExecutionList />} />
        <Route path='executions/new' element={<ExecutionForm />} />
        <Route path='executions/:id' element={<ExecutionDetail />} />

        {/* Evaluations */}
        <Route path='evaluations' element={<EvaluationList />} />
        <Route path='evaluations/new' element={<EvaluationForm />} />

        {/* Tasks */}
        <Route path='tasks' element={<TaskList />} />
        <Route path='tasks/:id' element={<TaskDetail />} />

        {/* System */}
        <Route path='system' element={<SystemSettings />} />

        {/* Settings */}
        <Route path='settings/profile' element={<UserProfile />} />
        <Route path='settings' element={<Settings />} />

        {/* Utility Test */}
        <Route path='utility-test' element={<UtilityTest />} />

        {/* Fallback */}
        <Route path='*' element={<Navigate to='/dashboard' replace />} />
      </Route>
    </Routes>
  );
}

export default App;
