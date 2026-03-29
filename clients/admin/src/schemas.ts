// Entity schemas for Factory — used by components until schema API is wired.
// These match the Go entity definitions in server/factory/entities/.

import type { EntitySchema } from '@yolo-hq/sdk'

export const taskSchema: EntitySchema = {
  name: 'Task',
  domain: 'factory',
  fields: [
    { name: 'id', type: 'string' },
    { name: 'title', type: 'string', label: 'Title' },
    { name: 'body', type: 'string', label: 'Description' },
    {
      name: 'status',
      type: 'enum',
      label: 'Status',
      enum_values: ['queued', 'running', 'done', 'failed', 'blocked', 'cancelled'],
      color_map: {
        queued: 'bg-yellow-100 text-yellow-800',
        running: 'bg-blue-100 text-blue-800',
        done: 'bg-green-100 text-green-800',
        failed: 'bg-red-100 text-red-800',
        blocked: 'bg-gray-100 text-gray-800',
        cancelled: 'bg-gray-100 text-gray-500',
      },
    },
    {
      name: 'priority',
      type: 'integer',
      label: 'Priority',
    },
    {
      name: 'type',
      type: 'enum',
      label: 'Type',
      enum_values: ['feature', 'bugfix', 'refactor', 'test', 'docs'],
    },
    { name: 'model', type: 'string', label: 'Model' },
    { name: 'max_retries', type: 'integer', label: 'Max Retries' },
    { name: 'timeout_secs', type: 'integer', label: 'Timeout (s)' },
    {
      name: 'repo_id',
      type: 'relation',
      label: 'Repo',
      relation: { entity: 'Repo', display_field: 'name' },
    },
    { name: 'cost', type: 'number', label: 'Cost' },
    { name: 'run_count', type: 'integer', label: 'Runs' },
    { name: 'created_at', type: 'datetime', label: 'Created' },
  ],
  actions: {
    create: { label: 'Create Task' },
    update: { label: 'Update' },
    execute: { label: 'Execute', placement: ['row'] },
    cancel: { label: 'Cancel', placement: ['row', 'bulk'], confirm: true },
    delete: { label: 'Delete', placement: ['row', 'bulk'], confirm: true },
  },
}

export const runSchema: EntitySchema = {
  name: 'Run',
  domain: 'factory',
  fields: [
    { name: 'id', type: 'string' },
    {
      name: 'task_id',
      type: 'relation',
      label: 'Task',
      relation: { entity: 'Task', display_field: 'title' },
    },
    {
      name: 'status',
      type: 'enum',
      label: 'Status',
      enum_values: ['running', 'success', 'failed'],
      color_map: {
        running: 'bg-blue-100 text-blue-800',
        success: 'bg-green-100 text-green-800',
        failed: 'bg-red-100 text-red-800',
      },
    },
    { name: 'agent', type: 'string', label: 'Agent' },
    { name: 'model', type: 'string', label: 'Model' },
    { name: 'cost', type: 'number', label: 'Cost' },
    { name: 'duration', type: 'number', label: 'Duration' },
    { name: 'started_at', type: 'datetime', label: 'Started' },
  ],
}

export const repoSchema: EntitySchema = {
  name: 'Repo',
  domain: 'factory',
  fields: [
    { name: 'id', type: 'string' },
    { name: 'name', type: 'string', label: 'Name' },
    { name: 'url', type: 'url', label: 'URL' },
    { name: 'local_path', type: 'string', label: 'Local Path' },
    { name: 'target_branch', type: 'string', label: 'Branch' },
    { name: 'default_model', type: 'string', label: 'Default Model' },
    { name: 'feedback_loops', type: 'integer', label: 'Feedback Loops' },
    { name: 'active', type: 'boolean', label: 'Active' },
  ],
  actions: {
    create: { label: 'Add Repo' },
    update: { label: 'Update' },
  },
}

export const questionSchema: EntitySchema = {
  name: 'Question',
  domain: 'factory',
  fields: [
    { name: 'id', type: 'string' },
    { name: 'body', type: 'string', label: 'Question' },
    {
      name: 'status',
      type: 'enum',
      label: 'Status',
      enum_values: ['open', 'resolved'],
      color_map: {
        open: 'bg-yellow-100 text-yellow-800',
        resolved: 'bg-green-100 text-green-800',
      },
    },
    {
      name: 'task_id',
      type: 'relation',
      label: 'Task',
      relation: { entity: 'Task', display_field: 'title' },
    },
    {
      name: 'run_id',
      type: 'relation',
      label: 'Run',
      relation: { entity: 'Run' },
    },
  ],
  actions: {
    resolve: { label: 'Resolve', placement: ['row'] },
  },
}
