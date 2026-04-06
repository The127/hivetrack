import { test, expect } from '@playwright/test'

/**
 * E2E: Refinement phase-gated UI
 *
 * Uses project e2e-test with issues #1 and #2 (created by global-setup).
 * Each test uses a separate issue to avoid state leaking.
 */

const PROJECT_SLUG = 'e2e-test'

test.describe('Refinement Phase UI', () => {
  test('refinement panel opens and shows correct state', async ({ page }) => {
    // Issue #1 — may or may not have an active session
    await page.goto(`/projects/${PROJECT_SLUG}/issues/1`)
    await page.waitForLoadState('networkidle')

    // Click the Refine button
    const refineButton = page.getByRole('button', { name: /refine/i })
    await expect(refineButton).toBeVisible({ timeout: 10_000 })
    await refineButton.click()

    // The panel should open with the Refinement header
    await expect(page.getByText('Refinement').first()).toBeVisible({ timeout: 5_000 })

    // Either "Start session" (no session) or phase stepper (active session)
    const startButton = page.getByRole('button', { name: /start session/i })
    const actorGoal = page.getByText('Actor & Goal')
    await expect(startButton.or(actorGoal)).toBeVisible({ timeout: 5_000 })
  })

  test('phase stepper shows all four phases when session is active', async ({ page }) => {
    // Issue #2 — start a session if one doesn't exist
    await page.goto(`/projects/${PROJECT_SLUG}/issues/2`)
    await page.waitForLoadState('networkidle')

    // Open refinement panel
    const refineButton = page.getByRole('button', { name: /refine/i })
    await expect(refineButton).toBeVisible({ timeout: 10_000 })
    await refineButton.click()

    // Start the session if not already active
    const startButton = page.getByRole('button', { name: /start session/i })
    const actorGoal = page.getByText('Actor & Goal')

    // Wait for either Start session or the phase stepper
    await expect(startButton.or(actorGoal)).toBeVisible({ timeout: 5_000 })

    if (await startButton.isVisible()) {
      await startButton.click()
      await page.waitForTimeout(3000)
    }

    // Phase stepper should appear with all 4 phases
    await expect(page.getByText('Actor & Goal')).toBeVisible({ timeout: 15_000 })
    await expect(page.getByText('Main Scenario')).toBeVisible()
    await expect(page.getByText('Extensions')).toBeVisible()
    await expect(page.getByText('Acceptance Criteria')).toBeVisible()
  })
})
