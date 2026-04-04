import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { YoloProvider } from '@yolo-hq/sdk'
import {
  YoloView,
  createYoloRouter,
  RouterProvider,
  type YoloViewConfig,
  type PageRouteConfig,
} from '@yolo-hq/view'

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 10_000, retry: 1 } },
})

// ── Theme: Dark mission control ──
const config: YoloViewConfig = {
  version: 1,
  app: { name: 'YOLO Factory', auth: 'required' },
  theme: {
    tokens: {
      colors: {
        primary: '#6366f1',      // indigo
        accent: '#8b5cf6',
        sidebar: '#0f0a1e',      // deep dark purple
      },
      density: 'compact',
      radius: 'md',
      motion: 'subtle',
    },
    dark: 'auto',
  },
  layout: {
    mode: 'sidebar',
    sidebar: { style: 'single', collapsible: true, width: 220 },
    header: { sticky: true, search: true, breadcrumbs: true },
    mobile: { nav: 'hamburger' },
    footer: { text: 'YOLO Factory — Autonomous Dev Engine' },
  },
  navigation: [
    {
      group: 'Mission Control',
      icon: 'radar',
      items: [
        { path: '/dashboard', label: 'Dashboard', icon: 'layout-dashboard' },
        { path: '/questions', label: 'Questions', icon: 'message-circle' },
      ],
    },
    {
      group: 'Work',
      icon: 'file-text',
      items: [
        { path: '/prds', label: 'PRDs', icon: 'file-text' },
        { path: '/tasks', label: 'Tasks', icon: 'square-check' },
      ],
    },
    {
      group: 'Execution',
      icon: 'play',
      items: [
        { path: '/runs', label: 'Runs', icon: 'play-circle' },
      ],
    },
    {
      group: 'Intelligence',
      icon: 'brain',
      items: [
        { path: '/suggestions', label: 'Suggestions', icon: 'lightbulb' },
        { path: '/insights', label: 'Insights', icon: 'brain' },
      ],
    },
    {
      group: 'Platform',
      icon: 'settings',
      items: [
        { path: '/projects', label: 'Projects', icon: 'folder-git-2' },
      ],
    },
  ],
}

// ── Pages ──
const pages: Record<string, PageRouteConfig> = {
  // Dashboard
  '/dashboard': {
    page: {
      page: { type: 'dashboard', title: 'Mission Control', domain: 'factory' },
      sections: [
        {
          type: 'stats',
          title: 'Overview',
          items: [
            { field: 'total', label: 'Running Tasks', variant: 'accent' },
            { field: 'total', label: 'Open Questions', variant: 'warning' },
            { field: 'total', label: 'Spend Today', format: 'currency', variant: 'default' },
            { field: 'total', label: 'Active PRDs', variant: 'success' },
          ],
        },
        {
          type: 'entity_list',
          title: 'Active PRDs',
          entity: 'PRD',
          domain: 'factory',
          columns: ['title', 'status', 'completed_tasks', 'total_tasks', 'total_cost_usd'],
          limit: 5,
        },
        {
          type: 'entity_list',
          title: 'Recent Runs',
          entity: 'Run',
          domain: 'factory',
          columns: ['agent_type', 'model', 'status', 'cost_usd', 'duration_ms', 'started_at'],
          limit: 5,
        },
      ],
    },
  },

  // Questions
  '/questions': {
    page: {
      page: { type: 'entity_list', title: 'Questions', entity: 'Question', domain: 'factory' },
      toolbar: { search: true, views: ['table'], default_view: 'table' },
    },
    entity: {
      entity: 'Question', domain: 'factory',
      meta: { icon: 'message-circle', label: 'Question', plural: 'Questions', color: '#f59e0b' },
      table: {
        columns: ['body', 'confidence', 'status', 'task_id', 'created_at'],
        default_sort: '-created_at',
      },
      fields: {
        status: { color_map: { open: 'amber', answered: 'green', dismissed: 'gray', auto_resolved: 'blue' } },
        confidence: { color_map: { low: 'red', medium: 'yellow', high: 'green' } },
        created_at: { format: 'relative_date' },
      },
      actions: [
        { name: 'answer', label: 'Answer', icon: 'reply', variant: 'default', pinned: true },
      ],
      empty: { icon: 'message-circle', title: 'No questions', description: 'Agents haven\'t asked anything yet.' },
    },
  },

  // PRDs
  '/prds': {
    page: {
      page: { type: 'entity_list', title: 'PRDs', entity: 'PRD', domain: 'factory' },
      toolbar: { search: true, views: ['table', 'board'], default_view: 'board', create_button: true },
    },
    entity: {
      entity: 'PRD', domain: 'factory',
      meta: { icon: 'file-text', label: 'PRD', plural: 'PRDs', color: '#8b5cf6' },
      views: { default: 'board', available: ['table', 'board'] },
      table: {
        columns: ['title', 'status', 'source', 'total_tasks', 'completed_tasks', 'total_cost_usd', 'created_at'],
        default_sort: '-created_at',
      },
      board: { group_by: 'status', card_title: 'title', card_subtitle: 'source' },
      fields: {
        status: { color_map: { draft: 'gray', approved: 'blue', planning: 'indigo', in_progress: 'amber', completed: 'green', failed: 'red' } },
        total_cost_usd: { format: 'currency', label: 'Cost' },
        created_at: { format: 'relative_date' },
      },
      actions: [
        { name: 'approve', label: 'Approve', icon: 'check', variant: 'default', pinned: true },
        { name: 'execute', label: 'Execute', icon: 'play', variant: 'default', pinned: true },
      ],
      empty: { icon: 'file-text', title: 'No PRDs yet', description: 'Submit your first PRD to start building.', cta_label: 'Submit PRD' },
    },
  },

  // Tasks
  '/tasks': {
    page: {
      page: { type: 'entity_list', title: 'Tasks', entity: 'Task', domain: 'factory' },
      toolbar: { search: true, views: ['table', 'board'], default_view: 'table' },
    },
    entity: {
      entity: 'Task', domain: 'factory',
      meta: { icon: 'square-check', label: 'Task', plural: 'Tasks', color: '#6366f1' },
      views: { default: 'table', available: ['table', 'board'] },
      table: {
        columns: ['title', 'status', 'branch', 'model', 'run_count', 'cost_usd', 'created_at'],
        default_sort: '-created_at',
      },
      board: { group_by: 'status', card_title: 'title', card_subtitle: 'branch' },
      fields: {
        status: { color_map: { queued: 'gray', blocked: 'orange', running: 'blue', reviewing: 'purple', done: 'green', failed: 'red', cancelled: 'slate' } },
        cost_usd: { format: 'currency', label: 'Cost' },
        created_at: { format: 'relative_date' },
      },
      actions: [
        { name: 'retry', label: 'Retry', icon: 'refresh-cw', variant: 'default' },
        { name: 'cancel', label: 'Cancel', icon: 'x', variant: 'destructive', confirm: true },
      ],
    },
  },

  // Runs
  '/runs': {
    page: {
      page: { type: 'entity_list', title: 'Runs', entity: 'Run', domain: 'factory' },
      toolbar: { search: true, views: ['table'], default_view: 'table' },
    },
    entity: {
      entity: 'Run', domain: 'factory',
      meta: { icon: 'play-circle', label: 'Run', plural: 'Runs', color: '#3b82f6' },
      table: {
        columns: ['agent_type', 'model', 'status', 'cost_usd', 'tokens_in', 'tokens_out', 'duration_ms', 'started_at'],
        default_sort: '-started_at',
      },
      fields: {
        status: { color_map: { running: 'blue', completed: 'green', failed: 'red', cancelled: 'slate' } },
        cost_usd: { format: 'currency', label: 'Cost' },
        started_at: { format: 'relative_date' },
      },
    },
  },

  // Suggestions
  '/suggestions': {
    page: {
      page: { type: 'entity_list', title: 'Suggestions', entity: 'Suggestion', domain: 'factory' },
      toolbar: { search: true, views: ['table'], default_view: 'table' },
    },
    entity: {
      entity: 'Suggestion', domain: 'factory',
      meta: { icon: 'lightbulb', label: 'Suggestion', plural: 'Suggestions', color: '#eab308' },
      table: {
        columns: ['title', 'category', 'priority', 'source', 'status', 'created_at'],
        default_sort: '-created_at',
      },
      fields: {
        status: { color_map: { pending: 'amber', accepted: 'green', rejected: 'gray', converted: 'blue' } },
        priority: { color_map: { low: 'slate', medium: 'yellow', high: 'orange', critical: 'red' } },
        category: { color_map: { bug: 'red', feature: 'blue', refactor: 'purple', test: 'green', docs: 'slate' } },
        created_at: { format: 'relative_date' },
      },
      actions: [
        { name: 'approve', label: 'Approve', icon: 'check', variant: 'default', pinned: true },
        { name: 'reject', label: 'Reject', icon: 'x', variant: 'destructive' },
      ],
      empty: { icon: 'lightbulb', title: 'No suggestions', description: 'AI hasn\'t found improvements yet.' },
    },
  },

  // Insights
  '/insights': {
    page: {
      page: { type: 'entity_list', title: 'Insights', entity: 'Insight', domain: 'factory' },
      toolbar: { search: true, views: ['table', 'card'], default_view: 'card' },
    },
    entity: {
      entity: 'Insight', domain: 'factory',
      meta: { icon: 'brain', label: 'Insight', plural: 'Insights', color: '#8b5cf6' },
      views: { default: 'card', available: ['table', 'card'] },
      table: {
        columns: ['title', 'category', 'priority', 'status', 'created_at'],
      },
      card: { title_field: 'title', subtitle_field: 'category', badge_field: 'priority', columns: 2 },
      fields: {
        status: { color_map: { pending: 'amber', acknowledged: 'blue', applied: 'green', dismissed: 'gray' } },
        priority: { color_map: { low: 'slate', medium: 'yellow', high: 'orange', critical: 'red' } },
        created_at: { format: 'relative_date' },
      },
      actions: [
        { name: 'acknowledge', label: 'Acknowledge', icon: 'eye', variant: 'default' },
        { name: 'apply', label: 'Apply', icon: 'check', variant: 'default', pinned: true },
        { name: 'dismiss', label: 'Dismiss', icon: 'x', variant: 'ghost' },
      ],
      empty: { icon: 'brain', title: 'No insights yet', description: 'Factory will generate optimization insights as it runs.' },
    },
  },

  // Projects
  '/projects': {
    page: {
      page: { type: 'entity_list', title: 'Projects', entity: 'Project', domain: 'factory' },
      toolbar: { search: true, views: ['table', 'card'], default_view: 'card', create_button: true },
    },
    entity: {
      entity: 'Project', domain: 'factory',
      meta: { icon: 'folder-git-2', label: 'Project', plural: 'Projects', color: '#10b981' },
      views: { default: 'card', available: ['table', 'card'] },
      table: {
        columns: ['name', 'status', 'default_model', 'spent_this_month_usd', 'budget_monthly_usd', 'created_at'],
      },
      card: { title_field: 'name', subtitle_field: 'repo_url', badge_field: 'status', columns: 2 },
      fields: {
        status: { color_map: { active: 'green', paused: 'yellow', archived: 'gray' } },
        spent_this_month_usd: { format: 'currency', label: 'Spent' },
        budget_monthly_usd: { format: 'currency', label: 'Budget' },
        created_at: { format: 'relative_date' },
      },
      actions: [
        { name: 'pause', label: 'Pause', icon: 'pause', variant: 'default' },
        { name: 'resume', label: 'Resume', icon: 'play', variant: 'default' },
      ],
    },
  },
}

const router = createYoloRouter({ pages, defaultPath: '/dashboard' })

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <YoloProvider config={{ baseUrl: '/api/v1', realtime: true }}>
        <YoloView config={config} user={{ id: '1', role: 'admin', name: 'Admin' }}>
          <RouterProvider router={router} />
        </YoloView>
      </YoloProvider>
    </QueryClientProvider>
  )
}
