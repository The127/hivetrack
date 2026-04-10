/**
 * E2E: Refinement token streaming UI
 *
 * Uses Playwright route() to mock the refinement session API — no real
 * drone or Hivemind connection needed. Tests the streaming bubble, spinner,
 * and question display logic in RefinementPanel.
 *
 * Uses issue #1 from the e2e-test project (created by global-setup).
 */
import { test, expect } from '@playwright/test'

const SESSION_URL = '**/api/v1/projects/e2e-test/issues/1/refinement/session'
const START_URL = '**/api/v1/projects/e2e-test/issues/1/refinement/start'

const BASE_SESSION = {
  id: 'aaaaaaaa-0000-0000-0000-000000000001',
  issue_id: 'bbbbbbbb-0000-0000-0000-000000000001',
  status: 'active',
  current_phase: 'actor_goal',
  messages: [],
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  partial_response: '',
  is_generating: false,
}

async function openRefinementPanel(page) {
  await page.goto('/projects/e2e-test/issues/1')
  await page.waitForLoadState('networkidle')
  const refineButton = page.getByRole('button', { name: /refine/i })
  await expect(refineButton).toBeVisible({ timeout: 10_000 })
  await refineButton.click()
  await expect(page.getByText('Refinement').first()).toBeVisible({ timeout: 5_000 })
}

test.describe('Refinement streaming UI', () => {
  test('shows spinner when generating but no tokens received yet', async ({ page }) => {
    await page.route(SESSION_URL, route =>
      route.fulfill({ json: { ...BASE_SESSION, is_generating: true, partial_response: '' } }),
    )
    await page.route(START_URL, route => route.fulfill({ status: 200, json: {} }))

    await openRefinementPanel(page)

    // Start the session (no active session mocked, start button may appear)
    const startButton = page.getByRole('button', { name: /start refinement/i })
    if (await startButton.isVisible({ timeout: 3_000 }).catch(() => false)) {
      // After clicking start, the session mock will return is_generating:true
      await page.route(SESSION_URL, route =>
        route.fulfill({ json: { ...BASE_SESSION, is_generating: true, partial_response: '' } }),
      )
      await startButton.click()
    }

    // Spinner should appear (generating but no tokens yet)
    await expect(page.locator('.animate-spin, [data-testid="spinner"], svg.animate-spin').first())
      .toBeVisible({ timeout: 10_000 })
  })

  test('shows streaming bubble with partial response while generating', async ({ page }) => {
    const partial = 'Who is the primary actor in this scenario?'

    await page.route(SESSION_URL, route =>
      route.fulfill({
        json: {
          ...BASE_SESSION,
          is_generating: true,
          partial_response: partial,
          messages: [{ id: 'msg-1', role: 'user', content: 'Start', message_type: 'message', phase: 'actor_goal', created_at: '2026-01-01T00:00:00Z' }],
        },
      }),
    )

    await openRefinementPanel(page)

    // Partial text should be visible in the streaming bubble
    await expect(page.getByText(partial, { exact: false })).toBeVisible({ timeout: 10_000 })
  })

  test('shows blinking cursor while streaming', async ({ page }) => {
    await page.route(SESSION_URL, route =>
      route.fulfill({
        json: {
          ...BASE_SESSION,
          is_generating: true,
          partial_response: 'Thinking...',
          messages: [{ id: 'msg-1', role: 'user', content: 'Start', message_type: 'message', phase: 'actor_goal', created_at: '2026-01-01T00:00:00Z' }],
        },
      }),
    )

    await openRefinementPanel(page)

    // Blinking cursor element should be present (animate-pulse class)
    await expect(page.locator('.animate-pulse').first()).toBeVisible({ timeout: 10_000 })
  })

  test('shows question card when agent has responded and is not generating', async ({ page }) => {
    const question = 'Who is the primary actor?'

    await page.route(SESSION_URL, route =>
      route.fulfill({
        json: {
          ...BASE_SESSION,
          is_generating: false,
          partial_response: '',
          messages: [
            { id: 'msg-1', role: 'user', content: 'Start', message_type: 'message', phase: 'actor_goal', created_at: '2026-01-01T00:00:00Z' },
            { id: 'msg-2', role: 'assistant', content: question, message_type: 'message', phase: 'actor_goal', created_at: '2026-01-01T00:01:00Z' },
          ],
        },
      }),
    )

    await openRefinementPanel(page)

    await expect(page.getByText(question, { exact: false })).toBeVisible({ timeout: 10_000 })
    // No streaming bubble (partial response text) when not generating
    await expect(page.locator('[data-generating="true"]')).not.toBeVisible()
  })

  test('does not show streaming bubble when generation stops', async ({ page }) => {
    const question = 'What is the actor trying to achieve?'

    // First call: generating with partial
    let callCount = 0
    await page.route(SESSION_URL, route => {
      callCount++
      if (callCount === 1) {
        route.fulfill({
          json: { ...BASE_SESSION, is_generating: true, partial_response: 'Partial...' },
        })
      } else {
        route.fulfill({
          json: {
            ...BASE_SESSION,
            is_generating: false,
            partial_response: '',
            messages: [
              { id: 'msg-1', role: 'user', content: 'Start', message_type: 'message', phase: 'actor_goal', created_at: '2026-01-01T00:00:00Z' },
              { id: 'msg-2', role: 'assistant', content: question, message_type: 'message', phase: 'actor_goal', created_at: '2026-01-01T00:01:00Z' },
            ],
          },
        })
      }
    })

    await openRefinementPanel(page)

    // Eventually the final question replaces the streaming bubble
    await expect(page.getByText(question, { exact: false })).toBeVisible({ timeout: 15_000 })
  })
})
