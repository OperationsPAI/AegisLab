import { lazy, Suspense, useEffect } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';

import MainLayout from '@/components/layout/MainLayout';
import { useAuthStore } from '@/store/auth';

// Lazy load all page components
const Login = lazy(() => import('@/pages/auth/Login'));
const Dashboard = lazy(() => import('@/pages/dashboard/Dashboard'));
const ProjectList = lazy(() => import('@/pages/projects/ProjectList'));
const ProjectEdit = lazy(() => import('@/pages/projects/ProjectEdit'));
const ContainerList = lazy(() => import('@/pages/containers/ContainerList'));
const ContainerForm = lazy(() => import('@/pages/containers/ContainerForm'));
const ContainerDetail = lazy(() => import('@/pages/containers/ContainerDetail'));
const ContainerVersions = lazy(() => import('@/pages/containers/ContainerVersions'));
const DatasetList = lazy(() => import('@/pages/datasets/DatasetList'));
const DatasetForm = lazy(() => import('@/pages/datasets/DatasetForm'));
const DatasetDetail = lazy(() => import('@/pages/datasets/DatasetDetail'));
const DatapackList = lazy(() => import('@/pages/datapacks/DatapackList'));
const DatapackDetail = lazy(() => import('@/pages/datapacks/DatapackDetail'));
const InjectionList = lazy(() => import('@/pages/injections/InjectionList'));
const InjectionCreate = lazy(() => import('@/pages/injections/InjectionCreate'));
const InjectionDetail = lazy(() => import('@/pages/injections/InjectionDetail'));
const ExecutionList = lazy(() => import('@/pages/executions/ExecutionList'));
const ExecutionForm = lazy(() => import('@/pages/executions/ExecutionForm'));
const ExecutionDetail = lazy(() => import('@/pages/executions/ExecutionDetail'));
const EvaluationList = lazy(() => import('@/pages/evaluations/EvaluationList'));
const EvaluationForm = lazy(() => import('@/pages/evaluations/EvaluationForm'));
const EvaluationDetail = lazy(() => import('@/pages/evaluations/EvaluationDetail'));
const TaskList = lazy(() => import('@/pages/tasks/TaskList'));
const TaskDetail = lazy(() => import('@/pages/tasks/TaskDetail'));
const SystemSettings = lazy(() => import('@/pages/system/SystemSettings'));
const UserProfile = lazy(() => import('@/pages/settings/UserProfile'));
const Settings = lazy(() => import('@/pages/settings/Settings'));
const UtilityTest = lazy(() => import('@/pages/UtilityTest'));

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
      <Route path='/login' element={
        <Suspense fallback={<div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>Loading...</div>}>
          <Login />
        </Suspense>
      } />

      {/* Protected routes */}
      <Route path='/*' element={isAuthenticated ? <MainLayout /> : <Navigate to='/login' replace />}>
        <Route index element={<Navigate to='/dashboard' replace />} />
        <Route path='dashboard' element={
          <Suspense fallback={<div>Loading...</div>}>
            <Dashboard />
          </Suspense>
        } />

        {/* Projects */}
        <Route path='projects' element={
          <Suspense fallback={<div>Loading...</div>}>
            <ProjectList />
          </Suspense>
        } />
        <Route path='projects/:id/edit' element={
          <Suspense fallback={<div>Loading...</div>}>
            <ProjectEdit />
          </Suspense>
        } />

        {/* Containers */}
        <Route path='containers' element={
          <Suspense fallback={<div>Loading...</div>}>
            <ContainerList />
          </Suspense>
        } />
        <Route path='containers/new' element={
          <Suspense fallback={<div>Loading...</div>}>
            <ContainerForm />
          </Suspense>
        } />
        <Route path='containers/:id' element={
          <Suspense fallback={<div>Loading...</div>}>
            <ContainerDetail />
          </Suspense>
        } />
        <Route path='containers/:id/edit' element={
          <Suspense fallback={<div>Loading...</div>}>
            <ContainerForm />
          </Suspense>
        } />
        <Route path='containers/:id/versions' element={
          <Suspense fallback={<div>Loading...</div>}>
            <ContainerVersions />
          </Suspense>
        } />

        {/* Datasets */}
        <Route path='datasets' element={
          <Suspense fallback={<div>Loading...</div>}>
            <DatasetList />
          </Suspense>
        } />
        <Route path='datasets/new' element={
          <Suspense fallback={<div>Loading...</div>}>
            <DatasetForm />
          </Suspense>
        } />
        <Route path='datasets/:id' element={
          <Suspense fallback={<div>Loading...</div>}>
            <DatasetDetail />
          </Suspense>
        } />
        <Route path='datasets/:id/edit' element={
          <Suspense fallback={<div>Loading...</div>}>
            <DatasetForm />
          </Suspense>
        } />

        {/* Datapacks */}
        <Route path='datapacks' element={
          <Suspense fallback={<div>Loading...</div>}>
            <DatapackList />
          </Suspense>
        } />
        <Route path='datapacks/:id' element={
          <Suspense fallback={<div>Loading...</div>}>
            <DatapackDetail />
          </Suspense>
        } />

        {/* Injections */}
        <Route path='injections' element={
          <Suspense fallback={<div>Loading...</div>}>
            <InjectionList />
          </Suspense>
        } />
        <Route path='injections/create' element={
          <Suspense fallback={<div>Loading...</div>}>
            <InjectionCreate />
          </Suspense>
        } />
        <Route path='injections/:id' element={
          <Suspense fallback={<div>Loading...</div>}>
            <InjectionDetail />
          </Suspense>
        } />

        {/* Executions */}
        <Route path='executions' element={
          <Suspense fallback={<div>Loading...</div>}>
            <ExecutionList />
          </Suspense>
        } />
        <Route path='executions/new' element={
          <Suspense fallback={<div>Loading...</div>}>
            <ExecutionForm />
          </Suspense>
        } />
        <Route path='executions/:id' element={
          <Suspense fallback={<div>Loading...</div>}>
            <ExecutionDetail />
          </Suspense>
        } />

        {/* Evaluations */}
        <Route path='evaluations' element={
          <Suspense fallback={<div>Loading...</div>}>
            <EvaluationList />
          </Suspense>
        } />
        <Route path='evaluations/new' element={
          <Suspense fallback={<div>Loading...</div>}>
            <EvaluationForm />
          </Suspense>
        } />
        <Route path='evaluations/:id' element={
          <Suspense fallback={<div>Loading...</div>}>
            <EvaluationDetail />
          </Suspense>
        } />

        {/* Tasks */}
        <Route path='tasks' element={
          <Suspense fallback={<div>Loading...</div>}>
            <TaskList />
          </Suspense>
        } />
        <Route path='tasks/:id' element={
          <Suspense fallback={<div>Loading...</div>}>
            <TaskDetail />
          </Suspense>
        } />

        {/* System */}
        <Route path='system' element={
          <Suspense fallback={<div>Loading...</div>}>
            <SystemSettings />
          </Suspense>
        } />

        {/* Settings */}
        <Route path='settings/profile' element={
          <Suspense fallback={<div>Loading...</div>}>
            <UserProfile />
          </Suspense>
        } />
        <Route path='settings' element={
          <Suspense fallback={<div>Loading...</div>}>
            <Settings />
          </Suspense>
        } />

        {/* Utility Test */}
        <Route path='utility-test' element={
          <Suspense fallback={<div>Loading...</div>}>
            <UtilityTest />
          </Suspense>
        } />

        {/* Fallback */}
        <Route path='*' element={<Navigate to='/dashboard' replace />} />
      </Route>
    </Routes>
  );
}

export default App;
