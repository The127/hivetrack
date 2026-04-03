import { apiFetch } from '@/composables/useApi'

export const startRefinementSession = (slug, number) =>
  apiFetch(`/api/v1/projects/${slug}/issues/${number}/refinement/start`, {
    method: 'POST',
  })

export const sendRefinementMessage = (slug, number, content) =>
  apiFetch(`/api/v1/projects/${slug}/issues/${number}/refinement/message`, {
    method: 'POST',
    body: JSON.stringify({ content }),
  })

export const getRefinementSession = (slug, number) =>
  apiFetch(`/api/v1/projects/${slug}/issues/${number}/refinement/session`)

export const acceptRefinementProposal = (slug, number) =>
  apiFetch(`/api/v1/projects/${slug}/issues/${number}/refinement/accept`, {
    method: 'POST',
  })
