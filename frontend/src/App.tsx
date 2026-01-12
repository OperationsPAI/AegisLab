import { Routes, Route, Navigate } from 'react-router-dom'
import { useEffect } from 'react'
import { useAuthStore } from '@/store/auth'
import MainLayout from '@/components/layout/MainLayout'
import Login from '@/pages/auth/Login'
import Dashboard from '@/pages/dashboard/Dashboard'
import ProjectList from '@/pages/projects/ProjectList'
import ContainerList from '@/pages/containers/ContainerList'
import InjectionList from '@/pages/injections/InjectionList'
import InjectionCreate from '@/pages/injections/InjectionCreate'
import ExecutionList from '@/pages/executions/ExecutionList'

function App() {
  const { isAuthenticated, loadUser } = useAuthStore()

  useEffect(() => {
    if (isAuthenticated) {
      loadUser()
    }
  }, [isAuthenticated, loadUser])

  return (
    <Routes>
      {/* Public routes */}
      <Route path="/login" element={<Login />} />

      {/* Protected routes */}
      <Route
        path="/*"
        element={
          isAuthenticated ? <MainLayout /> : <Navigate to="/login" replace />
        }
      >
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<Dashboard />} />

        {/* Projects */}
        <Route path="projects" element={<ProjectList />} />

        {/* Containers */}
        <Route path="containers" element={<ContainerList />} />

        {/* Injections */}
        <Route path="injections" element={<InjectionList />} />
        <Route path="injections/create" element={<InjectionCreate />} />

        {/* Executions */}
        <Route path="executions" element={<ExecutionList />} />

        {/* Fallback */}
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Route>
    </Routes>
  )
}

export default App
