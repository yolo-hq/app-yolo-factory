// Root layout — wraps all pages with SDK provider + admin chrome.

import { Outlet, useRouter } from '@tanstack/react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { YoloProvider, type YoloConfig } from '@yolo-hq/sdk'
import { AdminLayout } from '@yolo-hq/components'
import { adminConfig } from '../config'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 30_000, retry: 1 },
  },
})

const yoloConfig: YoloConfig = {
  baseUrl: '/api/v1',
}

export function RootLayout() {
  const router = useRouter()
  const pathname = router.state.location.pathname

  // Derive active entity from path: /factory/tasks → Task
  const segments = pathname.split('/').filter(Boolean)
  const activeEntity = segments[1]
    ? segments[1].charAt(0).toUpperCase() + segments[1].slice(1, -1) // tasks → Task
    : undefined

  // Breadcrumb from path
  const breadcrumb = segments.map((seg, i) => ({
    label: seg.charAt(0).toUpperCase() + seg.slice(1),
    href: i < segments.length - 1 ? '/' + segments.slice(0, i + 1).join('/') : undefined,
  }))

  const handleNavigate = (entity: string) => {
    const path = `/factory/${entity.toLowerCase()}s`
    router.navigate({ to: path })
  }

  return (
    <QueryClientProvider client={queryClient}>
      <YoloProvider config={yoloConfig}>
        <AdminLayout
          config={adminConfig}
          breadcrumb={breadcrumb}
          activeEntity={activeEntity}
          onNavigate={handleNavigate}
        >
          <Outlet />
        </AdminLayout>
      </YoloProvider>
    </QueryClientProvider>
  )
}
