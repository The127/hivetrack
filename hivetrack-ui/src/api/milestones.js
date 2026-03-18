import { apiFetch } from '@/composables/useApi'

export const fetchMilestones = (slug) =>
  apiFetch(`/api/v1/projects/${slug}/milestones`)

export const createMilestone = (slug, data) =>
  apiFetch(`/api/v1/projects/${slug}/milestones`, {
    method: 'POST',
    body: JSON.stringify(data),
  })

export const updateMilestone = (slug, milestoneId, data) =>
  apiFetch(`/api/v1/projects/${slug}/milestones/${milestoneId}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })

export const deleteMilestone = (slug, milestoneId) =>
  apiFetch(`/api/v1/projects/${slug}/milestones/${milestoneId}`, { method: 'DELETE' })
