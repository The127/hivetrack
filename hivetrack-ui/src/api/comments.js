import { apiFetch } from '@/composables/useApi'

export const fetchComments = (slug, number) =>
  apiFetch(`/api/v1/projects/${slug}/issues/${number}/comments`)

export const createComment = (slug, number, data) =>
  apiFetch(`/api/v1/projects/${slug}/issues/${number}/comments`, {
    method: 'POST',
    body: JSON.stringify(data),
  })

export const updateComment = (slug, number, commentId, data) =>
  apiFetch(`/api/v1/projects/${slug}/issues/${number}/comments/${commentId}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })

export const deleteComment = (slug, number, commentId) =>
  apiFetch(`/api/v1/projects/${slug}/issues/${number}/comments/${commentId}`, {
    method: 'DELETE',
  })
