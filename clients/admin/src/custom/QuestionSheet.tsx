import { useState } from 'react'
import { useQuery, useAction } from '@yolo-hq/sdk'
import { Drawer, DrawerContent } from '@yolo-hq/view'
import { Badge, Button, Textarea } from '@yolo-hq/view'
import { formatRelativeTime } from '@yolo-hq/view'

export interface QuestionSheetProps {
  questionId: string
  open: boolean
  onClose: () => void
}

const confidenceColors: Record<string, string> = {
  low: 'red',
  medium: 'yellow',
  high: 'green',
}

const statusColors: Record<string, { bg: string; text: string }> = {
  open: { bg: '#fef3c7', text: '#92400e' },
  answered: { bg: '#dcfce7', text: '#166534' },
  dismissed: { bg: '#f3f4f6', text: '#374151' },
  auto_resolved: { bg: '#dbeafe', text: '#1e40af' },
}

export function QuestionSheet({ questionId, open, onClose }: QuestionSheetProps) {
  const [answer, setAnswer] = useState('')

  const { data: question, isLoading } = useQuery({
    domain: 'factory',
    entity: 'Question' as any,
    query: 'get' as any,
    id: questionId,
    options: { enabled: !!questionId },
  })

  const answerMutation = useAction({
    domain: 'factory',
    entity: 'Question' as any,
    action: 'answer' as any,
    options: {
      onSuccess: () => {
        setAnswer('')
        onClose()
      },
    },
  })

  const q = question as Record<string, unknown> | null

  return (
    <Drawer open={open} onClose={onClose} title="Question" size="lg">
      <DrawerContent>
        {isLoading ? (
          <div className="text-sm text-[var(--text-tertiary)]">Loading...</div>
        ) : q ? (
          <div className="flex flex-col gap-6">
            {/* Status + Confidence */}
            <div className="flex items-center gap-2">
              <Badge
                variant="default"
                style={statusColors[q.status as string]
                  ? { backgroundColor: statusColors[q.status as string].bg, color: statusColors[q.status as string].text }
                  : undefined}
              >
                {String(q.status)}
              </Badge>
              <Badge variant="outline">
                {String(q.confidence)} confidence
              </Badge>
            </div>

            {/* Question body */}
            <div>
              <h3 className="text-xs font-semibold uppercase tracking-wider text-[var(--text-tertiary)] mb-2">Question</h3>
              <p className="text-sm text-[var(--text-default)] leading-relaxed whitespace-pre-wrap">
                {String(q.body ?? '')}
              </p>
            </div>

            {/* Context */}
            {q.context && (
              <div>
                <h3 className="text-xs font-semibold uppercase tracking-wider text-[var(--text-tertiary)] mb-2">Context</h3>
                <p className="text-sm text-[var(--text-secondary)] leading-relaxed whitespace-pre-wrap">
                  {String(q.context)}
                </p>
              </div>
            )}

            {/* Task link */}
            {q.task_id && (
              <div className="text-xs text-[var(--text-tertiary)]">
                Task: <span className="text-[var(--color-primary)]">{String(q.task_id).slice(0, 8)}...</span>
                {' · '}
                {q.created_at && formatRelativeTime(new Date(q.created_at as string))}
              </div>
            )}

            {/* Answer section */}
            {q.status === 'open' ? (
              <div className="border-t border-[var(--border-default)] pt-4">
                <h3 className="text-xs font-semibold uppercase tracking-wider text-[var(--text-tertiary)] mb-2">Your Answer</h3>
                <Textarea
                  value={answer}
                  onChange={(e) => setAnswer(e.target.value)}
                  rows={4}
                  placeholder="Type your answer..."
                />
                <div className="flex justify-end gap-2 mt-3">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={onClose}
                  >
                    Skip
                  </Button>
                  <Button
                    size="sm"
                    onClick={() => answerMutation.mutate({ id: questionId, answer } as any)}
                    isLoading={answerMutation.isPending}
                    disabled={!answer.trim()}
                  >
                    Submit Answer
                  </Button>
                </div>
              </div>
            ) : q.answer ? (
              <div className="border-t border-[var(--border-default)] pt-4">
                <h3 className="text-xs font-semibold uppercase tracking-wider text-[var(--text-tertiary)] mb-2">Answer</h3>
                <p className="text-sm text-[var(--text-default)] leading-relaxed whitespace-pre-wrap bg-[var(--bg-subtle)] rounded-[var(--radius-default)] p-3">
                  {String(q.answer)}
                </p>
                <div className="text-xs text-[var(--text-tertiary)] mt-2">
                  Answered by {String(q.answered_by ?? 'unknown')}
                  {q.answered_at && ` · ${formatRelativeTime(new Date(q.answered_at as string))}`}
                </div>
              </div>
            ) : null}
          </div>
        ) : (
          <div className="text-sm text-[var(--text-tertiary)]">Question not found</div>
        )}
      </DrawerContent>
    </Drawer>
  )
}
