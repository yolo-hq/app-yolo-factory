import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  YoloView, createYoloRouter, RouterProvider,
  type YoloViewConfig, type PageRouteConfig,
} from '@yolo-hq/view'

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 30_000, retry: 1 } },
})

const config: YoloViewConfig = {
  version: 1,
  app: { name: 'YOLO Factory', auth: 'required' },
  theme: {
    tokens: {
      colors: { primary: '#0ea5e9', sidebar: '#0c1222' },
      density: 'compact',
      radius: 'md',
      motion: 'subtle',
    },
    dark: 'auto',
  },
  layout: {
    mode: 'sidebar',
    sidebar: { style: 'single', collapsible: true },
    header: { sticky: true, search: true, breadcrumbs: true },
    mobile: { nav: 'hamburger' },
  },
  navigation: [
    {
      group: 'Factory',
      icon: 'cpu',
      items: [
        { path: '/dashboard', label: 'Dashboard', icon: 'layout-dashboard' },
        { path: '/tasks', label: 'Tasks', icon: 'check-square' },
        { path: '/runs', label: 'Runs', icon: 'play' },
        { path: '/repos', label: 'Repos', icon: 'git-branch' },
        { path: '/questions', label: 'Questions', icon: 'message-square' },
      ],
    },
  ],
}

const pages: Record<string, PageRouteConfig> = {
  '/dashboard': {
    page: {
      page: { type: 'dashboard', title: 'Dashboard', domain: 'factory' },
      sections: [
        {
          type: 'stats',
          title: 'Overview',
          items: [
            { field: 'total', label: 'Total Tasks', variant: 'default' },
            { field: 'total', label: 'Active Runs', variant: 'accent' },
            { field: 'total', label: 'Repos', variant: 'success' },
            { field: 'total', label: 'Open Questions', variant: 'warning' },
          ],
        },
        {
          type: 'entity_list',
          title: 'Recent Tasks',
          entity: 'Task',
          domain: 'factory',
          columns: ['title', 'status', 'type', 'created_at'],
          limit: 5,
        },
        {
          type: 'entity_list',
          title: 'Latest Runs',
          entity: 'Run',
          domain: 'factory',
          columns: ['agent', 'status', 'cost', 'started_at'],
          limit: 5,
        },
      ],
    },
  },
  '/tasks': {
    page: {
      page: { type: 'entity_list', title: 'Tasks', entity: 'Task', domain: 'factory' },
      toolbar: { search: true, views: ['table', 'board'], default_view: 'table', create_button: true },
    },
    entity: {
      entity: 'Task', domain: 'factory',
      meta: { icon: 'check-square', label: 'Task', plural: 'Tasks' },
      views: { default: 'table', available: ['table', 'board'] },
      table: {
        columns: ['title', 'type', 'status', 'priority', 'model', 'run_count', 'created_at'],
        default_sort: '-created_at',
      },
      board: { group_by: 'status', card_title: 'title', card_subtitle: 'type' },
      fields: {
        status: { color_map: { queued: 'gray', running: 'blue', completed: 'green', failed: 'red', cancelled: 'yellow' } },
        type: { color_map: { auto: 'blue', manual: 'gray', review: 'purple' } },
        created_at: { format: 'relative_date' },
      },
    },
  },
  '/runs': {
    page: {
      page: { type: 'entity_list', title: 'Runs', entity: 'Run', domain: 'factory' },
      toolbar: { search: true, views: ['table'], default_view: 'table' },
    },
    entity: {
      entity: 'Run', domain: 'factory',
      meta: { icon: 'play', label: 'Run', plural: 'Runs' },
      table: {
        columns: ['task_id', 'agent', 'model', 'status', 'cost', 'duration', 'started_at'],
        default_sort: '-started_at',
      },
      fields: {
        status: { color_map: { running: 'blue', completed: 'green', failed: 'red', cancelled: 'yellow' } },
        cost: { format: 'currency' },
        started_at: { format: 'relative_date' },
      },
    },
  },
  '/repos': {
    page: {
      page: { type: 'entity_list', title: 'Repos', entity: 'Repo', domain: 'factory' },
      toolbar: { search: true, views: ['table'], default_view: 'table' },
    },
    entity: {
      entity: 'Repo', domain: 'factory',
      meta: { icon: 'git-branch', label: 'Repo', plural: 'Repos' },
      table: { columns: ['name', 'url', 'target_branch', 'default_model', 'active', 'created_at'] },
      fields: { created_at: { format: 'relative_date' } },
    },
  },
  '/questions': {
    page: {
      page: { type: 'entity_list', title: 'Questions', entity: 'Question', domain: 'factory' },
      toolbar: { search: true, views: ['table'], default_view: 'table' },
    },
    entity: {
      entity: 'Question', domain: 'factory',
      meta: { icon: 'message-square', label: 'Question', plural: 'Questions' },
      table: {
        columns: ['body', 'status', 'task_id', 'created_at'],
        default_sort: '-created_at',
      },
      fields: {
        status: { color_map: { open: 'blue', resolved: 'green', dismissed: 'gray' } },
        created_at: { format: 'relative_date' },
      },
    },
  },
}

const router = createYoloRouter({ pages, defaultPath: '/dashboard' })

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <YoloView config={config} user={{ id: '1', role: 'admin', name: 'Admin' }}>
        <RouterProvider router={router} />
      </YoloView>
    </QueryClientProvider>
  )
}
