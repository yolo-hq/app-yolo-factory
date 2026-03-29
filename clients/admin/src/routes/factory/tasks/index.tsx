// Task list page — DataTable with columns from entity.ui.yml, filters, bulk actions.

import { useState, useCallback } from 'react'
import { useQuery, useAction } from '@yolo-hq/sdk'
import { DataTable, FilterBar, DetailSheet, FormBuilder } from '@yolo-hq/components'
import { taskSchema } from '../../../schemas'

const uiColumns = [
  { field: 'title', title: 'Title', width: '2fr', sortable: true },
  { field: 'status', title: 'Status', sortable: true },
  { field: 'priority', title: 'Priority', width: '80px', sortable: true },
  { field: 'type', title: 'Type' },
  { field: 'repo_id', title: 'Repo' },
  { field: 'created_at', title: 'Created', width: '150px', sortable: true },
]

const bulkActions = [
  { name: 'cancel', label: 'Cancel', placement: ['bulk'] },
  { name: 'delete', label: 'Delete', placement: ['bulk'] },
]

export function TaskListPage() {
  const [filters, setFilters] = useState<Record<string, any>>({})
  const [search, setSearch] = useState('')
  const [sortKey, setSortKey] = useState<string>('created_at')
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc')
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [detailItem, setDetailItem] = useState<any>(null)
  const [showCreate, setShowCreate] = useState(false)
  const [createValues, setCreateValues] = useState<Record<string, any>>({})

  const { data, isLoading } = useQuery({
    domain: 'factory',
    entity: 'task' as any,
    query: 'list' as any,
    filter: filters as any,
    sort: `${sortDir === 'desc' ? '-' : ''}${sortKey}`,
    search: search || undefined,
    fields: ['id', 'title', 'status', 'priority', 'type', 'repo_id', 'created_at'] as const,
  })

  const { mutate: cancelTask } = useAction({
    domain: 'factory',
    entity: 'task' as any,
    action: 'cancel' as any,
  })

  const { mutate: createTask } = useAction({
    domain: 'factory',
    entity: 'task' as any,
    action: 'create' as any,
  })

  const handleSort = useCallback((key: string, dir: 'asc' | 'desc') => {
    setSortKey(key)
    setSortDir(dir)
  }, [])

  const handleBulkAction = useCallback((action: string, ids: string[]) => {
    if (action === 'cancel') {
      ids.forEach((id) => cancelTask({ id } as any))
      setSelectedIds(new Set())
    }
  }, [cancelTask])

  const tasks = (data as any)?.tasks ?? (data as any)?.data ?? []

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Tasks</h2>
        <button
          onClick={() => setShowCreate(true)}
          className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          Create Task
        </button>
      </div>

      <FilterBar
        entity={taskSchema}
        fields={['status', 'priority', 'type']}
        values={filters}
        onChange={setFilters}
        onSearch={setSearch}
        searchValue={search}
      />

      <DataTable
        data={tasks}
        entity={taskSchema}
        uiColumns={uiColumns}
        loading={isLoading}
        sortKey={sortKey}
        sortDirection={sortDir}
        onSort={handleSort}
        selectable
        selectedIds={selectedIds}
        onSelectionChange={setSelectedIds}
        actions={bulkActions}
        onAction={handleBulkAction}
        onRowClick={(item) => setDetailItem(item)}
      />

      <DetailSheet
        entity={taskSchema}
        data={detailItem ?? {}}
        open={!!detailItem}
        onClose={() => setDetailItem(null)}
        actions={[
          { name: 'execute', label: 'Execute' },
          { name: 'cancel', label: 'Cancel' },
        ]}
      />

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30">
          <div className="w-full max-w-lg rounded-lg border bg-background p-6 shadow-xl">
            <h3 className="mb-4 text-lg font-semibold">Create Task</h3>
            <FormBuilder
              entity={taskSchema}
              config={{
                rows: [
                  { fields: ['title'] },
                  { fields: ['repo_id', 'type'] },
                  { fields: ['body'] },
                  { fields: ['priority', 'model'] },
                  { fields: ['max_retries', 'timeout_secs'] },
                ],
              }}
              values={createValues}
              onChange={(field, value) => setCreateValues((v) => ({ ...v, [field]: value }))}
              onSubmit={(values) => {
                createTask(values as any)
                setShowCreate(false)
                setCreateValues({})
              }}
            />
            <button
              onClick={() => setShowCreate(false)}
              className="mt-2 text-sm text-muted-foreground hover:text-foreground"
            >
              Cancel
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
