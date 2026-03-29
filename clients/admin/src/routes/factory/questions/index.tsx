// Question list page.
import { useState } from 'react'
import { useQuery, useAction } from '@yolo-hq/sdk'
import { DataTable, DetailSheet } from '@yolo-hq/components'
import { questionSchema } from '../../../schemas'

const uiColumns = [
  { field: 'body', title: 'Question', width: '3fr' },
  { field: 'status', title: 'Status' },
  { field: 'task_id', title: 'Task' },
  { field: 'run_id', title: 'Run' },
]

export function QuestionListPage() {
  const [detailItem, setDetailItem] = useState<any>(null)

  const { data, isLoading } = useQuery({
    domain: 'factory',
    entity: 'question' as any,
    query: 'list' as any,
    fields: ['id', 'body', 'status', 'task_id', 'run_id'] as const,
  })

  const { mutate: resolveQuestion } = useAction({
    domain: 'factory',
    entity: 'question' as any,
    action: 'resolve' as any,
  })

  const questions = (data as any)?.questions ?? (data as any)?.data ?? []

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold">Questions</h2>
      <DataTable
        data={questions}
        entity={questionSchema}
        uiColumns={uiColumns}
        loading={isLoading}
        onRowClick={(item) => setDetailItem(item)}
      />
      <DetailSheet
        entity={questionSchema}
        data={detailItem ?? {}}
        open={!!detailItem}
        onClose={() => setDetailItem(null)}
        actions={[{ name: 'resolve', label: 'Resolve' }]}
        onAction={(action) => {
          if (action === 'resolve' && detailItem) {
            resolveQuestion({ id: detailItem.id } as any)
            setDetailItem(null)
          }
        }}
      />
    </div>
  )
}
