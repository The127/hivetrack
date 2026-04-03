import { apiFetch } from '@/composables/useApi'

export const fetchDrones = (slug) =>
  apiFetch(`/api/v1/projects/${slug}/drones`)

export const createDroneToken = (slug, data) =>
  apiFetch(`/api/v1/projects/${slug}/drones/tokens`, {
    method: 'POST',
    body: JSON.stringify(data),
  })

export const fetchDrone = (slug, droneId) =>
  apiFetch(`/api/v1/projects/${slug}/drones/${droneId}`)

export const deregisterDrone = (slug, droneId) =>
  apiFetch(`/api/v1/projects/${slug}/drones/${droneId}/deregister`, {
    method: 'POST',
  })

export const revokeToken = (slug, token) =>
  apiFetch(`/api/v1/projects/${slug}/drones/tokens/${token}`, {
    method: 'DELETE',
  })
