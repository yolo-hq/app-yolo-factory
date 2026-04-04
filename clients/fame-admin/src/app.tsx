import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  YoloView, createYoloRouter, RouterProvider,
  type YoloViewConfig, type PageRouteConfig,
} from '@yolo-hq/view'

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 30_000, retry: 1 } },
})

// ── FameCreators Admin Theme ──
const config: YoloViewConfig = {
  version: 1,
  app: { name: 'FameCreators', auth: 'required' },
  theme: {
    tokens: {
      colors: {
        primary: '#7c3aed',
        accent: '#a855f7',
        sidebar: '#1e1033',
      },
      density: 'comfortable',
      radius: 'lg',
      motion: 'subtle',
    },
    dark: 'auto',
  },
  layout: {
    mode: 'sidebar',
    sidebar: { style: 'single', collapsible: true, width: 250 },
    header: { sticky: true, search: true, breadcrumbs: true },
    mobile: {
      nav: 'bottom',
      items: [
        { label: 'Campaigns', icon: 'megaphone', path: '/campaigns' },
        { label: 'Creators', icon: 'users', path: '/creators' },
        { label: 'Finance', icon: 'wallet', path: '/finance' },
        { label: 'Support', icon: 'message-square', path: '/support' },
      ],
    },
  },
  navigation: [
    {
      group: 'Campaigns',
      icon: 'megaphone',
      items: [
        { path: '/campaigns', label: 'All Campaigns', icon: 'list' },
        { path: '/campaigns/board', label: 'Campaign Board', icon: 'columns' },
      ],
    },
    {
      group: 'People',
      icon: 'users',
      items: [
        { path: '/creators', label: 'Creators', icon: 'sparkles' },
        { path: '/brands', label: 'Brands', icon: 'building' },
      ],
    },
    {
      group: 'Finance',
      icon: 'wallet',
      items: [
        { path: '/finance', label: 'Transactions', icon: 'receipt' },
      ],
    },
    {
      group: 'Support',
      icon: 'message-square',
      items: [
        { path: '/support', label: 'Tickets', icon: 'ticket' },
      ],
    },
  ],
  components: {
    button: {
      classes: {
        base: 'rounded-xl font-medium transition-all shadow-sm',
        variants: {
          variant: {
            default: 'bg-gradient-to-r from-violet-500 to-purple-600 text-white hover:shadow-md',
            secondary: 'bg-white/10 backdrop-blur text-white border border-white/20',
            ghost: 'text-purple-300 hover:bg-white/5',
          },
        },
      },
    },
    card: {
      classes: {
        base: 'rounded-xl border border-white/10 bg-[var(--bg-surface)] shadow-sm',
      },
    },
  },
}

// ── Entity Configs ──
const pages: Record<string, PageRouteConfig> = {
  '/campaigns': {
    page: {
      page: { type: 'entity_list', title: 'All Campaigns', entity: 'Task', domain: 'factory' },
      toolbar: { search: true, views: ['table', 'card'], default_view: 'table', create_button: true },
    },
    entity: {
      entity: 'Task', domain: 'factory',
      meta: { icon: 'megaphone', label: 'Campaign', plural: 'Campaigns', color: '#7c3aed' },
      views: { default: 'table', available: ['table', 'card'] },
      table: {
        columns: ['title', 'status', 'type', 'priority', 'model', 'created_at'],
        default_sort: '-created_at',
        row_actions: ['edit', 'pause'],
      },
      card: { columns: 3, title_field: 'title', subtitle_field: 'type', badge_field: 'status' },
      fields: {
        title: { label: 'Campaign Name' },
        status: {
          label: 'Status',
          color_map: { queued: 'gray', running: 'blue', completed: 'green', failed: 'red', cancelled: 'yellow' },
        },
        type: { label: 'Category' },
        priority: { label: 'Priority' },
        model: { label: 'AI Model' },
        created_at: { label: 'Created', format: 'relative_date' },
      },
      actions: [
        { name: 'execute', label: 'Launch', icon: 'rocket', variant: 'default', pinned: true },
        { name: 'cancel', label: 'Cancel', icon: 'x', variant: 'destructive', confirm: true },
      ],
    },
  },
  '/campaigns/board': {
    page: {
      page: { type: 'entity_list', title: 'Campaign Board', entity: 'Task', domain: 'factory' },
      toolbar: { search: true, views: ['board'], default_view: 'board' },
    },
    entity: {
      entity: 'Task', domain: 'factory',
      meta: { icon: 'megaphone', label: 'Campaign', plural: 'Campaigns' },
      board: { group_by: 'status', card_title: 'title', card_subtitle: 'type' },
      fields: {
        status: { color_map: { queued: 'gray', running: 'blue', completed: 'green', failed: 'red' } },
      },
    },
  },
  '/creators': {
    page: {
      page: { type: 'entity_list', title: 'Creators', entity: 'Repo', domain: 'factory' },
      toolbar: { search: true, views: ['table', 'card'], default_view: 'card', create_button: true },
    },
    entity: {
      entity: 'Repo', domain: 'factory',
      meta: { icon: 'sparkles', label: 'Creator', plural: 'Creators', color: '#ec4899' },
      views: { default: 'card', available: ['table', 'card'] },
      table: { columns: ['name', 'url', 'default_model', 'active', 'created_at'] },
      card: { columns: 3, title_field: 'name', subtitle_field: 'url', badge_field: 'active' },
      fields: {
        name: { label: 'Creator Name' },
        url: { label: 'Profile' },
        default_model: { label: 'Tier' },
        active: { label: 'Verified' },
        created_at: { label: 'Joined', format: 'relative_date' },
      },
    },
  },
  '/brands': {
    page: {
      page: { type: 'entity_list', title: 'Brands', entity: 'Repo', domain: 'factory' },
      toolbar: { search: true, views: ['table'], default_view: 'table' },
    },
    entity: {
      entity: 'Repo', domain: 'factory',
      meta: { icon: 'building', label: 'Brand', plural: 'Brands' },
      table: { columns: ['name', 'url', 'target_branch', 'active', 'created_at'] },
      fields: {
        name: { label: 'Company' },
        url: { label: 'Website' },
        target_branch: { label: 'Industry' },
        active: { label: 'Active' },
        created_at: { label: 'Registered', format: 'relative_date' },
      },
    },
  },
  '/finance': {
    page: {
      page: { type: 'entity_list', title: 'Transactions', entity: 'Run', domain: 'factory' },
      toolbar: { search: true, views: ['table'], default_view: 'table' },
    },
    entity: {
      entity: 'Run', domain: 'factory',
      meta: { icon: 'receipt', label: 'Transaction', plural: 'Transactions' },
      table: { columns: ['task_id', 'agent', 'status', 'cost', 'duration', 'started_at'] },
      fields: {
        task_id: { label: 'Reference' },
        agent: { label: 'Payment Method' },
        status: {
          label: 'Status',
          color_map: { running: 'blue', completed: 'green', failed: 'red', cancelled: 'yellow' },
        },
        cost: { label: 'Amount', format: 'currency' },
        duration: { label: 'Processing Time' },
        started_at: { label: 'Date', format: 'relative_date' },
      },
    },
  },
  '/support': {
    page: {
      page: { type: 'entity_list', title: 'Support Tickets', entity: 'Question', domain: 'factory' },
      toolbar: { search: true, views: ['table', 'board'], default_view: 'table', create_button: true },
    },
    entity: {
      entity: 'Question', domain: 'factory',
      meta: { icon: 'ticket', label: 'Ticket', plural: 'Tickets' },
      views: { default: 'table', available: ['table', 'board'] },
      table: { columns: ['body', 'status', 'task_id', 'created_at'] },
      board: { group_by: 'status', card_title: 'body', card_subtitle: 'task_id' },
      fields: {
        body: { label: 'Description' },
        status: {
          label: 'Status',
          color_map: { open: 'blue', resolved: 'green', dismissed: 'gray' },
        },
        task_id: { label: 'Campaign' },
        created_at: { label: 'Opened', format: 'relative_date' },
      },
    },
  },
}

const router = createYoloRouter({ pages, defaultPath: '/campaigns' })

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <YoloView config={config} user={{ id: '1', role: 'admin', name: 'Admin' }}>
        <RouterProvider router={router} />
      </YoloView>
    </QueryClientProvider>
  )
}
