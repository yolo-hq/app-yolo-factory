import { useMemo } from 'react'
import { useQuery, useAction } from '@yolo-hq/sdk'
import {
  Badge, Button, Card, CardContent, CardHeader, CardTitle,
  ProgressBar, Tabs, TabsList, TabsTrigger, TabsContent,
  Skeleton, formatCurrency, formatRelativeTime,
} from '@yolo-hq/view'

const statusColors: Record<string, { bg: string; text: string }> = {
  draft:       { bg: '#f3f4f6', text: '#374151' },
  approved:    { bg: '#dbeafe', text: '#1e40af' },
  planning:    { bg: '#e0e7ff', text: '#3730a3' },
  in_progress: { bg: '#fef3c7', text: '#92400e' },
  completed:   { bg: '#dcfce7', text: '#166534' },
  failed:      { bg: '#fee2e2', text: '#991b1b' },
}

const taskStatusColors: Record<string, string> = {
  queued:    '#9ca3af',
  blocked:   '#f97316',
  running:   '#3b82f6',
  reviewing: '#8b5cf6',
  done:      '#22c55e',
  failed:    '#ef4444',
  cancelled: '#64748b',
}

export interface PRDDetailProps {
  prdId: string
}

export function PRDDetail({ prdId }: PRDDetailProps) {
  // Fetch PRD
  const { data: prd, isLoading } = useQuery({
    domain: 'factory',
    entity: 'PRD' as any,
    query: 'get' as any,
    id: prdId,
  })

  // Fetch tasks for this PRD
  const { data: tasksResult } = useQuery({
    domain: 'factory',
    entity: 'Task' as any,
    query: 'list' as any,
    filter: { prd_id: prdId } as any,
    options: { enabled: !!prdId },
  })

  // Actions
  const approveMutation = useAction({ domain: 'factory', entity: 'PRD' as any, action: 'approve' as any })
  const executeMutation = useAction({ domain: 'factory', entity: 'PRD' as any, action: 'execute' as any })

  const p = prd as Record<string, unknown> | null

  // Extract tasks array
  const tasks = useMemo(() => {
    if (!tasksResult) return []
    const data = (tasksResult as any)?.data ?? tasksResult
    if (Array.isArray(data)) return data
    if (typeof data === 'object') {
      for (const val of Object.values(data)) {
        if (Array.isArray(val)) return val as Record<string, unknown>[]
      }
    }
    return []
  }, [tasksResult])

  // Group tasks by status for board
  const taskGroups = useMemo(() => {
    const statuses = ['queued', 'blocked', 'running', 'reviewing', 'done', 'failed']
    return statuses.map(status => ({
      status,
      tasks: tasks.filter(t => t.status === status),
      color: taskStatusColors[status] ?? '#9ca3af',
    }))
  }, [tasks])

  if (isLoading || !p) {
    return (
      <div className="p-6 flex flex-col gap-4">
        <Skeleton className="h-16 w-full" />
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-64 w-full" />
      </div>
    )
  }

  const totalTasks = Number(p.total_tasks ?? 0)
  const completedTasks = Number(p.completed_tasks ?? 0)
  const failedTasks = Number(p.failed_tasks ?? 0)
  const status = String(p.status ?? 'draft')
  const sc = statusColors[status]

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="border-b border-[var(--border-default)] bg-[var(--bg-surface)] px-6 py-4">
        <div className="flex items-start justify-between gap-4">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-2">
              <Badge variant="default" style={sc ? { backgroundColor: sc.bg, color: sc.text } : undefined}>
                {status}
              </Badge>
              <span className="text-xs text-[var(--text-tertiary)]">
                {p.created_at && formatRelativeTime(new Date(p.created_at as string))}
              </span>
            </div>
            <h1 className="text-lg font-semibold text-[var(--text-default)] truncate">
              {String(p.title ?? 'Untitled PRD')}
            </h1>
          </div>

          <div className="flex items-center gap-2 shrink-0">
            {status === 'draft' && (
              <Button
                size="sm"
                onClick={() => approveMutation.mutate({ id: prdId } as any)}
                isLoading={approveMutation.isPending}
              >
                Approve
              </Button>
            )}
            {(status === 'draft' || status === 'approved') && (
              <Button
                size="sm"
                variant="secondary"
                onClick={() => executeMutation.mutate({ id: prdId } as any)}
                isLoading={executeMutation.isPending}
              >
                Execute
              </Button>
            )}
          </div>
        </div>

        {/* Progress bar */}
        {totalTasks > 0 && (
          <div className="mt-3">
            <ProgressBar
              value={completedTasks}
              max={totalTasks}
              showValue
              size="sm"
              variant={failedTasks > 0 ? 'danger' : completedTasks === totalTasks ? 'success' : 'default'}
            />
            <div className="flex gap-4 mt-1 text-xs text-[var(--text-tertiary)]">
              <span>{completedTasks} done</span>
              {failedTasks > 0 && <span className="text-[var(--status-danger)]">{failedTasks} failed</span>}
              <span className="ml-auto">{formatCurrency(Number(p.total_cost_usd ?? 0))}</span>
            </div>
          </div>
        )}
      </div>

      {/* Tabs */}
      <div className="flex-1 overflow-hidden">
        <Tabs defaultValue="tasks" className="flex flex-col h-full">
          <TabsList className="px-6 pt-2">
            <TabsTrigger value="tasks">Tasks ({tasks.length})</TabsTrigger>
            <TabsTrigger value="spec">Specification</TabsTrigger>
            <TabsTrigger value="cost">Cost</TabsTrigger>
          </TabsList>

          {/* Tasks tab — board view */}
          <TabsContent value="tasks" className="flex-1 overflow-auto p-4">
            <div className="flex gap-3 overflow-x-auto pb-4 min-h-[300px]">
              {taskGroups.map(group => (
                <div key={group.status} className="flex w-56 shrink-0 flex-col gap-2">
                  <div className="flex items-center gap-2 px-1 py-1">
                    <div className="h-2 w-2 rounded-full" style={{ backgroundColor: group.color }} />
                    <span className="text-xs font-medium text-[var(--text-secondary)]">
                      {group.status}
                    </span>
                    <span className="text-xs text-[var(--text-tertiary)]">{group.tasks.length}</span>
                  </div>

                  <div className="flex flex-col gap-1.5">
                    {group.tasks.map((task, i) => (
                      <Card key={String(task.id ?? i)} className="cursor-pointer hover:shadow-md transition-shadow">
                        <CardContent className="p-3">
                          <p className="text-xs font-medium text-[var(--text-default)] truncate">
                            {String(task.title ?? '')}
                          </p>
                          <div className="flex items-center gap-2 mt-1.5 text-[10px] text-[var(--text-tertiary)]">
                            <span>{String(task.branch ?? '')}</span>
                            {task.cost_usd && <span>{formatCurrency(Number(task.cost_usd))}</span>}
                          </div>
                        </CardContent>
                      </Card>
                    ))}

                    {group.tasks.length === 0 && (
                      <div className="rounded-[var(--radius-default)] border border-dashed border-[var(--border-default)] p-3 text-center text-[10px] text-[var(--text-tertiary)]">
                        Empty
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </TabsContent>

          {/* Spec tab */}
          <TabsContent value="spec" className="flex-1 overflow-auto p-6">
            <div className="max-w-3xl">
              <h3 className="text-sm font-medium text-[var(--text-default)] mb-3">Body</h3>
              <div className="text-sm text-[var(--text-secondary)] whitespace-pre-wrap leading-relaxed bg-[var(--bg-subtle)] rounded-[var(--radius-default)] p-4">
                {String(p.body ?? 'No specification provided.')}
              </div>

              {p.acceptance_criteria && (
                <div className="mt-6">
                  <h3 className="text-sm font-medium text-[var(--text-default)] mb-3">Acceptance Criteria</h3>
                  <ul className="flex flex-col gap-2">
                    {(p.acceptance_criteria as any[]).map((ac: any, i: number) => (
                      <li key={ac.id ?? i} className="flex items-start gap-2 text-sm text-[var(--text-secondary)]">
                        <span className="mt-0.5 h-4 w-4 rounded-full border border-[var(--border-default)] shrink-0 flex items-center justify-center text-[10px]">
                          {i + 1}
                        </span>
                        {ac.description ?? String(ac)}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          </TabsContent>

          {/* Cost tab */}
          <TabsContent value="cost" className="flex-1 overflow-auto p-6">
            <div className="max-w-3xl">
              <div className="grid grid-cols-3 gap-4 mb-6">
                <Card>
                  <CardContent className="p-4">
                    <p className="text-xs text-[var(--text-secondary)]">Total Cost</p>
                    <p className="text-xl font-semibold text-[var(--text-default)] tabular-nums">
                      {formatCurrency(Number(p.total_cost_usd ?? 0))}
                    </p>
                  </CardContent>
                </Card>
                <Card>
                  <CardContent className="p-4">
                    <p className="text-xs text-[var(--text-secondary)]">Tasks</p>
                    <p className="text-xl font-semibold text-[var(--text-default)] tabular-nums">{totalTasks}</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardContent className="p-4">
                    <p className="text-xs text-[var(--text-secondary)]">Avg per Task</p>
                    <p className="text-xl font-semibold text-[var(--text-default)] tabular-nums">
                      {totalTasks > 0 ? formatCurrency(Number(p.total_cost_usd ?? 0) / totalTasks) : '$0'}
                    </p>
                  </CardContent>
                </Card>
              </div>

              {/* Cost per task table */}
              <h3 className="text-sm font-medium text-[var(--text-default)] mb-3">Cost per Task</h3>
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-[var(--border-default)] bg-[var(--bg-subtle)]">
                    <th className="px-3 py-2 text-left text-xs font-medium text-[var(--text-secondary)]">Task</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-[var(--text-secondary)]">Status</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-[var(--text-secondary)]">Model</th>
                    <th className="px-3 py-2 text-right text-xs font-medium text-[var(--text-secondary)]">Cost</th>
                    <th className="px-3 py-2 text-right text-xs font-medium text-[var(--text-secondary)]">Runs</th>
                  </tr>
                </thead>
                <tbody>
                  {tasks.map((task, i) => (
                    <tr key={String(task.id ?? i)} className="border-b border-[var(--border-default)]">
                      <td className="px-3 py-2 text-[var(--text-default)] truncate max-w-[200px]">{String(task.title ?? '')}</td>
                      <td className="px-3 py-2">
                        <span className="inline-flex items-center gap-1">
                          <span className="h-1.5 w-1.5 rounded-full" style={{ backgroundColor: taskStatusColors[task.status as string] ?? '#9ca3af' }} />
                          <span className="text-xs">{String(task.status ?? '')}</span>
                        </span>
                      </td>
                      <td className="px-3 py-2 text-xs text-[var(--text-secondary)]">{String(task.model ?? '—')}</td>
                      <td className="px-3 py-2 text-right tabular-nums">{formatCurrency(Number(task.cost_usd ?? 0))}</td>
                      <td className="px-3 py-2 text-right tabular-nums text-[var(--text-secondary)]">{String(task.run_count ?? 0)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
