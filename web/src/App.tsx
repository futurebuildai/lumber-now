import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from '@/context/AuthContext'
import { TenantProvider } from '@/context/TenantContext'
import ErrorBoundary from '@/components/shared/ErrorBoundary'
import ProtectedRoute from '@/components/shared/ProtectedRoute'
import AppLayout from '@/components/layout/AppLayout'
import LoginPage from '@/features/auth/LoginPage'
import RegisterPage from '@/features/auth/RegisterPage'
import RequestFeed from '@/features/dealer-admin/RequestFeed'
import InventoryManager from '@/features/dealer-admin/InventoryManager'
import UserManagement from '@/features/dealer-admin/UserManagement'
import DealerSettings from '@/features/dealer-admin/DealerSettings'
import TenantList from '@/features/platform-admin/TenantList'
import { lazy, Suspense } from 'react'

const CreateDealerWizard = lazy(() => import('@/features/platform-admin/create-dealer/CreateDealerWizard'))
const EditDealerPage = lazy(() => import('@/features/platform-admin/EditDealerPage'))

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 30_000, retry: 1 },
  },
})

function WizardFallback() {
  return (
    <div className="flex items-center justify-center py-24">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
    </div>
  )
}

export default function App() {
  return (
    <ErrorBoundary>
    <QueryClientProvider client={queryClient}>
      <TenantProvider>
        <AuthProvider>
          <BrowserRouter>
            <Routes>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />
              <Route path="/" element={<ProtectedRoute><AppLayout /></ProtectedRoute>}>
                <Route index element={<Navigate to="/requests" replace />} />
                <Route path="requests" element={<RequestFeed />} />
                <Route path="inventory" element={
                  <ProtectedRoute roles={['dealer_admin', 'sales_rep', 'platform_admin']}>
                    <InventoryManager />
                  </ProtectedRoute>
                } />
                <Route path="admin/users" element={
                  <ProtectedRoute roles={['dealer_admin', 'platform_admin']}>
                    <UserManagement />
                  </ProtectedRoute>
                } />
                <Route path="admin/settings" element={
                  <ProtectedRoute roles={['dealer_admin', 'platform_admin']}>
                    <DealerSettings />
                  </ProtectedRoute>
                } />
                <Route path="platform" element={
                  <ProtectedRoute roles={['platform_admin']}>
                    <TenantList />
                  </ProtectedRoute>
                } />
                <Route path="platform/new-dealer" element={
                  <ProtectedRoute roles={['platform_admin']}>
                    <Suspense fallback={<WizardFallback />}>
                      <CreateDealerWizard />
                    </Suspense>
                  </ProtectedRoute>
                } />
                <Route path="platform/dealers/:id/edit" element={
                  <ProtectedRoute roles={['platform_admin']}>
                    <Suspense fallback={<WizardFallback />}>
                      <EditDealerPage />
                    </Suspense>
                  </ProtectedRoute>
                } />
              </Route>
            </Routes>
          </BrowserRouter>
        </AuthProvider>
      </TenantProvider>
    </QueryClientProvider>
    </ErrorBoundary>
  )
}
