import { apiFetch } from '@/composables/useApi'

export const fetchProjects = () => apiFetch('/api/v1/projects')

export const fetchProject = (slug) => apiFetch(`/api/v1/projects/${slug}`)

export const createProject = (data) =>
  apiFetch('/api/v1/projects', {
    method: 'POST',
    body: JSON.stringify(data),
  })

export const updateProject = (id, data) =>
  apiFetch(`/api/v1/projects/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })

export const deleteProject = (id) =>
  apiFetch(`/api/v1/projects/${id}`, { method: 'DELETE' })
