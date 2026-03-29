// Run list page.
import { useState, useCallback } from 'react'
import { useQuery } from '@yolo-hq/sdk'
import { DataTable, DetailSheet } from '@yolo-hq/components'
import { runSchema } from '../../../schemas'

const uiColumns = [
  { field: 'task_id', title: 'Task', width: '200px' },
  { field: 'status', title: 'Status', sortable: true },
  { field: 'agent', title: 'Agent', width: '100px' },
  { field: 'model', title: 'Model', width: '100px' },
  { field: 'cost', title: 'Cost', width: '80px', sortable: true },
  { field: 'duration', title: 'Duration', width: '80px' },
  { field: 'started_at', title: 'Started', width: '150px', sortable: true },
]

export function RunListPage() {
  const [sortKey, setSortKey] = useState('started_at')
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc')
  const [detailItem, setDetailItem] = useState<any>(null)

  const { data, isLoading } = useQuery({
    domain: 'factory',
    entity: 'run' as any,
    query: 'list' as any,
    sort: `${sortDir === 'desc' ? '-' : ''}${sortKey}`,
    fields: ['id', 'task_id', 'status', 'agent', 'model', 'cost', 'duration', 'started_at'] as const,
  })

  const handleSort = useCallback((key: string, dir: 'asc' | 'desc') => {
    setSortKey(key)
    setSortDir(dir)
  }, [])

  const runs = (data as any)?.runs ?? (data as any)?.data ?? []

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold">Runs</h2>
      <DataTable
        data={runs}
        entity={runSchema}
        uiColumns={uiColumns}
        loading={isLoading}
        sortKey={sortKey}
        sortDirection={sortDir}
        onSort={handleSort}
        onRowClick={(item) => setDetailItem(item)}
      />
      <DetailSheet
        entity={runSchema}
        data={detailItem ?? {}}
        open={!!detailItem}
        onClose={() => setDetailItem(null)}
      />
    </div>
  )
}
