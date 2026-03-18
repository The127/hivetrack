import { apiFetch } from '@/composables/useApi'

export const fetchIssues = (slug, params = {}) => {
  const qs = new URLSearchParams()
  if (params.status !== undefined) qs.set('status', params.status)
  if (params.priority !== undefined) qs.set('priority', params.priority)
  if (params.triaged !== undefined) qs.set('triaged', String(params.triaged))
  if (params.refined !== undefined) qs.set('refined', String(params.refined))
  if (params.backlog !== undefined) qs.set('backlog', String(params.backlog))
  if (params.text !== undefined) qs.set('text', params.text)
  if (params.type !== undefined) qs.set('type', params.type)
  if (params.parent_id !== undefined) qs.set('parent_id', params.parent_id)
  if (params.sprint_id !== undefined) qs.set('sprint_id', params.sprint_id)
  if (params.no_parent !== undefined) qs.set('no_parent', String(params.no_parent))
  if (params.limit !== undefined) qs.set('limit', String(params.limit))
  if (params.offset !== undefined) qs.set('offset', String(params.offset))
  const query = qs.toString()
  return apiFetch(`/api/v1/projects/${slug}/issues${query ? `?${query}` : ''}`)
}

export const fetchIssue = (slug, number) =>
  apiFetch(`/api/v1/projects/${slug}/issues/${number}`)

export const createIssue = (slug, data) =>
  apiFetch(`/api/v1/projects/${slug}/issues`, {
    method: 'POST',
    body: JSON.stringify(data),
  })

export const updateIssue = (slug, number, data) =>
  apiFetch(`/api/v1/projects/${slug}/issues/${number}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })

export const deleteIssue = (slug, number) =>
  apiFetch(`/api/v1/projects/${slug}/issues/${number}`, { method: 'DELETE' })

export const triageIssue = (slug, number, data) =>
  apiFetch(`/api/v1/projects/${slug}/issues/${number}/triage`, {
    method: 'POST',
    body: JSON.stringify(data),
  })

export const refineIssue = (slug, number) =>
  apiFetch(`/api/v1/projects/${slug}/issues/${number}/refine`, {
    method: 'POST',
  })

export const fetchMyIssues = () => apiFetch('/api/v1/me/issues')

export const splitIssue = (slug, number, titles) =>
  apiFetch(`/api/v1/projects/${slug}/issues/${number}/split`, {
    method: 'POST',
    body: JSON.stringify({ titles }),
  })
