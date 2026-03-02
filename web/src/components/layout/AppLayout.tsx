import { Outlet } from 'react-router-dom'
import AppSidebar from './AppSidebar'

export default function AppLayout() {
  return (
    <div className="flex h-screen overflow-hidden">
      <AppSidebar />
      <div className="flex-1 flex flex-col overflow-hidden">
        <main className="flex-1 overflow-y-auto p-6 bg-background" role="main" aria-label="Page content">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
