import { YoloApp, parseClientConfig, parseBlockPage } from '@yolo-hq/view'
import { YoloProvider } from '@yolo-hq/sdk'
import { PRDDetail } from './custom/PRDDetail'
import { QuestionSheet } from './custom/QuestionSheet'
import type { PageBlock } from '@yolo-hq/view'

// Import YAML configs as raw strings (Vite ?raw)
import clientYml from '../config/client.ui.yml?raw'
import dashboardYml from '../config/pages/dashboard.page.yml?raw'
import prdDetailYml from '../config/pages/prds/[id].page.yml?raw'

// ── Parse configs ──

const { data: config } = parseClientConfig(clientYml)
if (!config) throw new Error('Invalid client.ui.yml')

function loadPage(yml: string): PageBlock {
  const { data, errors } = parseBlockPage(yml)
  if (!data) throw new Error(`Invalid page: ${errors[0]?.message}`)
  return data
}

const pages: Record<string, PageBlock> = {
  '/dashboard': loadPage(dashboardYml),
  '/prds/$id': loadPage(prdDetailYml),
}

// ── App ──

export default function App() {
  return (
    <YoloProvider config={{ baseUrl: '/api/v1', realtime: true }}>
      <YoloApp
        client="admin"
        config={config}
        components={{ PRDDetail, QuestionSheet }}
        pages={pages}
        user={{ id: '1', role: 'admin', name: 'Admin' }}
        defaultPath="/dashboard"
      />
    </YoloProvider>
  )
}
