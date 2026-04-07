import { YoloApp, parseClientConfig, parseBlockPage } from '@yolo-hq/view'
import { PRDDetail } from './custom/PRDDetail'
import { QuestionSheet } from './custom/QuestionSheet'
import type { PageBlock } from '@yolo-hq/view'

// Import YAML configs as raw strings (Vite ?raw)
import clientYml from '../config/client.ui.yml?raw'
import dashboardYml from '../config/pages/dashboard.page.yml?raw'
import prdDetailYml from '../config/pages/prds/[id].page.yml?raw'

// ── Parse configs ──

const { data: config, errors } = parseClientConfig(clientYml)
if (!config) throw new Error(`Invalid client.ui.yml: ${errors.map(e => e.message).join(', ')}`)

function loadPage(yml: string, name: string): PageBlock {
  const { data, errors } = parseBlockPage(yml)
  if (!data) throw new Error(`Invalid ${name}: ${errors[0]?.message}`)
  return data
}

const pages: Record<string, PageBlock> = {
  '/dashboard': loadPage(dashboardYml, 'dashboard.page.yml'),
  '/prds/$id': loadPage(prdDetailYml, 'prds/[id].page.yml'),
}

// ── App ──

export default function App() {
  return (
    <YoloApp
      client="admin"
      config={config}
      sdk={{ baseUrl: 'http://localhost:9000', realtime: true }}
      components={{ PRDDetail, QuestionSheet }}
      pages={pages}
      user={{ id: '1', role: 'admin', name: 'Admin' }}
      defaultPath="/dashboard"
    />
  )
}
