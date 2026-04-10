/**
 * E2E: Refinement conversation flow
 *
 * Mocks the refinement session API to verify the full start → thinking →
 * question → reply → next-question cycle without a real drone.
 *
 * Catches regressions like: "Hivemind is thinking..." never resolving.
 *
 * Uses issue #1 from the e2e-test project (created by global-setup).
 */
import { test, expect } from '@playwright/test'

const SESSION_URL = '**/api/v1/projects/e2e-test/issues/1/refinement/session'
const START_URL = '**/api/v1/projects/e2e-test/issues/1/refinement/start'
const MESSAGE_URL = '**/api/v1/projects/e2e-test/issues/1/refinement/message'

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

const ACTOR_QUESTION = 'Who is the primary actor that benefits from this feature?'
const ACTOR_QUESTION_MSG = {
  id: 'msg-agent-1',
  role: 'assistant',
  content: ACTOR_QUESTION,
  message_type: 'question',
  suggestions: ['Developer', 'End user', 'System administrator'],
  phase: 'actor_goal',
  created_at: '2026-01-01T00:01:00Z',
}

const GOAL_QUESTION = 'What goal does the end user want to achieve?'
const GOAL_QUESTION_MSG = {
  id: 'msg-agent-2',
  role: 'assistant',
  content: GOAL_QUESTION,
  message_type: 'question',
  suggestions: [],
  phase: 'actor_goal',
  created_at: '2026-01-01T00:03:00Z',
}

async function openRefinementPanel(page) {
  await page.goto('/projects/e2e-test/issues/1')
  await page.waitForLoadState('networkidle')
  const refineButton = page.getByRole('button', { name: /refine/i })
  await expect(refineButton).toBeVisible({ timeout: 10_000 })
  await refineButton.click()
  await expect(page.getByText('Refinement').first()).toBeVisible({ timeout: 5_000 })
}

test.describe('Refinement conversation flow', () => {
  test('thinking spinner resolves to first question', async ({ page }) => {
    // Simulate: session already generating (drone received request), no tokens yet,
    // then a question arrives on the next poll. Catches the "stuck thinking" regression.
    let callCount = 0
    await page.route(SESSION_URL, route => {
      callCount++
      if (callCount <= 2) {
        route.fulfill({ json: { ...BASE_SESSION, is_generating: true, partial_response: '' } })
      } else {
        route.fulfill({
          json: { ...BASE_SESSION, is_generating: false, messages: [ACTOR_QUESTION_MSG] },
        })
      }
    })

    await openRefinementPanel(page)

    // Spinner visible while generating but no tokens
    await expect(page.getByText('Hivemind is thinking...')).toBeVisible({ timeout: 10_000 })

    // Spinner resolves to first question (polling picks up the completed response)
    await expect(page.getByText(ACTOR_QUESTION, { exact: false })).toBeVisible({ timeout: 15_000 })
    await expect(page.getByText('Hivemind is thinking...')).not.toBeVisible()
  })

  test('sending a reply triggers message POST and shows next question', async ({ page }) => {
    const userReply = 'End user'
    const userReplyMsg = {
      id: 'msg-user-1',
      role: 'user',
      content: userReply,
      message_type: 'message',
      phase: 'actor_goal',
      created_at: '2026-01-01T00:02:00Z',
    }

    // Session already has the first question waiting (with suggestions → placeholder changes)
    let messagePosted = false
    await page.route(SESSION_URL, route => {
      if (!messagePosted) {
        route.fulfill({
          json: { ...BASE_SESSION, is_generating: false, messages: [ACTOR_QUESTION_MSG] },
        })
      } else {
        route.fulfill({
          json: {
            ...BASE_SESSION,
            is_generating: false,
            messages: [ACTOR_QUESTION_MSG, userReplyMsg, GOAL_QUESTION_MSG],
          },
        })
      }
    })
    await page.route(MESSAGE_URL, async route => {
      messagePosted = true
      await route.fulfill({ status: 200, json: {} })
    })

    await openRefinementPanel(page)

    // First question should be visible
    await expect(page.getByText(ACTOR_QUESTION, { exact: false })).toBeVisible({ timeout: 10_000 })

    // ACTOR_QUESTION_MSG has suggestions → placeholder is "Or type your own answer..."
    const input = page.getByPlaceholder(/type your own answer/i)
    await input.fill(userReply)
    await input.press('Meta+Enter')

    // User reply and next question should appear
    await expect(page.getByText(userReply, { exact: false })).toBeVisible({ timeout: 10_000 })
    await expect(page.getByText(GOAL_QUESTION, { exact: false })).toBeVisible({ timeout: 10_000 })
  })

  test('suggestion chips send pre-filled reply on click', async ({ page }) => {
    let messageBody = null
    await page.route(SESSION_URL, route =>
      route.fulfill({
        json: {
          ...BASE_SESSION,
          is_generating: false,
          messages: [ACTOR_QUESTION_MSG],
        },
      }),
    )
    await page.route(MESSAGE_URL, async route => {
      messageBody = route.request().postDataJSON()
      await route.fulfill({ status: 200, json: {} })
    })

    await openRefinementPanel(page)

    await expect(page.getByText(ACTOR_QUESTION, { exact: false })).toBeVisible({ timeout: 10_000 })

    // Click a suggestion chip
    await page.getByRole('button', { name: 'End user' }).click()

    // The message POST should contain the chip text
    await expect.poll(() => messageBody?.content, { timeout: 5_000 }).toBe('End user')
  })
})
