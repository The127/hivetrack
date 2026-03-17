import { apiFetch } from '@/composables/useApi'

export const fetchIssues = (slug, params = {}) => {
  const qs = new URLSearchParams()
  if (params.status !== undefined) qs.set('status', params.status)
  if (params.priority !== undefined) qs.set('priority', params.priority)
  if (params.triaged !== undefined) qs.set('triaged', String(params.triaged))
  if (params.text !== undefined) qs.set('text', params.text)
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

export const fetchMyIssues = () => apiFetch('/api/v1/me/issues')
