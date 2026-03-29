// Repo list page.
import { useState, useCallback } from 'react'
import { useQuery } from '@yolo-hq/sdk'
import { DataTable, DetailSheet, FormBuilder } from '@yolo-hq/components'
import { repoSchema } from '../../../schemas'

const uiColumns = [
  { field: 'name', title: 'Name', width: '2fr', sortable: true },
  { field: 'url', title: 'URL', width: '2fr' },
  { field: 'target_branch', title: 'Branch', width: '120px' },
  { field: 'active', title: 'Active', width: '80px' },
]

export function RepoListPage() {
  const [sortKey, setSortKey] = useState('name')
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('asc')
  const [detailItem, setDetailItem] = useState<any>(null)
  const [showCreate, setShowCreate] = useState(false)
  const [createValues, setCreateValues] = useState<Record<string, any>>({})

  const { data, isLoading } = useQuery({
    domain: 'factory',
    entity: 'repo' as any,
    query: 'list' as any,
    sort: `${sortDir === 'desc' ? '-' : ''}${sortKey}`,
    fields: ['id', 'name', 'url', 'target_branch', 'active'] as const,
  })

  const handleSort = useCallback((key: string, dir: 'asc' | 'desc') => {
    setSortKey(key)
    setSortDir(dir)
  }, [])

  const repos = (data as any)?.repos ?? (data as any)?.data ?? []

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Repos</h2>
        <button
          onClick={() => setShowCreate(true)}
          className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          Add Repo
        </button>
      </div>

      <DataTable
        data={repos}
        entity={repoSchema}
        uiColumns={uiColumns}
        loading={isLoading}
        sortKey={sortKey}
        sortDirection={sortDir}
        onSort={handleSort}
        onRowClick={(item) => setDetailItem(item)}
      />

      <DetailSheet
        entity={repoSchema}
        data={detailItem ?? {}}
        open={!!detailItem}
        onClose={() => setDetailItem(null)}
      />

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30">
          <div className="w-full max-w-lg rounded-lg border bg-background p-6 shadow-xl">
            <h3 className="mb-4 text-lg font-semibold">Add Repository</h3>
            <FormBuilder
              entity={repoSchema}
              config={{
                rows: [
                  { fields: ['name', 'url'] },
                  { fields: ['local_path'] },
                  { fields: ['target_branch', 'default_model'] },
                  { fields: ['feedback_loops', 'active'] },
                ],
              }}
              values={createValues}
              onChange={(field, value) => setCreateValues((v) => ({ ...v, [field]: value }))}
              onSubmit={() => {
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
