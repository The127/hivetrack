import { apiFetch } from '@/composables/useApi'

export const fetchLabels = (slug) =>
  apiFetch(`/api/v1/projects/${slug}/labels`)

export const createLabel = (slug, data) =>
  apiFetch(`/api/v1/projects/${slug}/labels`, {
    method: 'POST',
    body: JSON.stringify(data),
  })

export const updateLabel = (slug, labelId, data) =>
  apiFetch(`/api/v1/projects/${slug}/labels/${labelId}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })

export const deleteLabel = (slug, labelId) =>
  apiFetch(`/api/v1/projects/${slug}/labels/${labelId}`, { method: 'DELETE' })
