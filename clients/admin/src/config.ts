// Admin configuration — loaded from admin.ui.yml at build time.
// For now, inline config. In production, loaded via yaml import.

import type { AdminConfig } from '@yolo-hq/components'

export const adminConfig: AdminConfig = {
  layout: {
    mode: 'sidebar',
    sidebar: {
      collapsible: true,
      collapse_mode: 'icon',
      width: 240,
      persist_state: true,
      keyboard_toggle: true,
    },
    header: {
      fixed: true,
      show_breadcrumb: true,
      show_search: true,
      show_theme_toggle: true,
    },
  },
  branding: {
    app_name: 'YOLO Factory',
  },
  theme: {
    default: 'dark',
    colors: {
      primary: '#6366f1',
      sidebar_bg: '#1e1e2e',
    },
  },
  navigation: {
    items: [
      {
        group: 'Factory',
        order: 1,
        items: [
          { entity: 'Task', icon: 'check-square', label: 'Tasks' },
          { entity: 'Run', icon: 'play', label: 'Runs' },
          { entity: 'Repo', icon: 'git-branch', label: 'Repos' },
          { entity: 'Question', icon: 'help-circle', label: 'Questions' },
        ],
      },
      {
        group: 'System',
        order: 99,
        position: 'bottom',
        items: [
          { label: 'Settings', icon: 'settings', href: '/settings' },
        ],
      },
    ],
  },
}
