import { apiFetch } from '@/composables/useApi'

export const fetchSprints = (slug) =>
  apiFetch(`/api/v1/projects/${slug}/sprints`)

export const createSprint = (slug, data) =>
  apiFetch(`/api/v1/projects/${slug}/sprints`, {
    method: 'POST',
    body: JSON.stringify(data),
  })

export const updateSprint = (slug, id, data) =>
  apiFetch(`/api/v1/projects/${slug}/sprints/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })

export const deleteSprint = (slug, id) =>
  apiFetch(`/api/v1/projects/${slug}/sprints/${id}`, { method: 'DELETE' })

export const fetchSprintBurndown = (slug, sprintId) =>
  apiFetch(`/api/v1/projects/${slug}/sprints/${sprintId}/burndown`)
