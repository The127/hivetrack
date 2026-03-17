import { useAuth } from '@/composables/useAuth'

/**
 * Typed error thrown by apiFetch on non-2xx responses.
 * Carries the HTTP status and the parsed error body from the API.
 */
export class ApiError extends Error {
  /**
   * @param {string} message
   * @param {number} status
   * @param {{ errors: Array<{ code: string, message: string, field?: string }> }} body
   */
  constructor(message, status, body) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.body = body
  }
}

/**
 * Authenticated fetch wrapper. Attaches the OIDC Bearer token and handles
 * error parsing. Use this for all calls to /api/v1/... endpoints.
 *
 * Throws ApiError on non-2xx responses — TanStack Query's `error` state
 * will capture it automatically.
 *
 * Returns null for 204 No Content responses.
 *
 * @param {string} path - API path, e.g. '/api/v1/projects'
 * @param {RequestInit} [options] - Standard fetch options (method, body, headers, …)
 * @returns {Promise<any>}
 *
 * @example
 * // GET
 * const projects = await apiFetch('/api/v1/projects')
 *
 * // POST
 * const issue = await apiFetch('/api/v1/projects/ht/issues', {
 *   method: 'POST',
 *   body: JSON.stringify({ title: 'Fix login bug', type: 'task' }),
 * })
 *
 * // DELETE (returns null)
 * await apiFetch('/api/v1/projects/ht/issues/42', { method: 'DELETE' })
 */
export async function apiFetch(path, options = {}) {
  const { getAccessToken } = useAuth()
  const token = getAccessToken()

  const response = await fetch(path, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  })

  if (response.status === 204) return null

  if (!response.ok) {
    let body
    try {
      body = await response.json()
    } catch {
      body = { errors: [{ code: 'unknown', message: 'Request failed' }] }
    }
    const message = body?.errors?.[0]?.message ?? `HTTP ${response.status}`
    throw new ApiError(message, response.status, body)
  }

  return response.json()
}
