/**
 * Playwright global setup — authenticates via Keyline OIDC, saves browser
 * storage state, then creates test data (project + issue) via the API.
 *
 * Assumes the full stack is running:
 *   - Keyline API on :8081, Keyline UI on :5177
 *   - Hivetrack backend on :8086, frontend dev server on :5399
 */
import { test as setup, expect } from '@playwright/test'

const AUTH_FILE = './e2e/.auth/user.json'

setup('authenticate via Keyline OIDC', async ({ page }) => {
  // Navigate to the app — the auth guard will redirect to Keyline login
  await page.goto('/')

  // Wait for the Keyline login form to appear (we've been redirected)
  await page.waitForURL(/localhost:8081|localhost:5177/, { timeout: 15_000 })

  // Fill in the seeded admin credentials
  await page.getByLabel(/username/i).fill('admin')
  await page.getByLabel(/password/i).fill('admin')
  await page.getByRole('button', { name: 'Sign In', exact: true }).click()

  // Wait to be redirected back to the app
  await page.waitForURL('http://localhost:5173/**', { timeout: 15_000 })

  // Verify we're authenticated — the page should show something meaningful
  await expect(page.locator('body')).not.toContainText('Sign in', { timeout: 10_000 })

  // Save signed-in state
  await page.context().storageState({ path: AUTH_FILE })

  // Wait for initial API calls to complete (creates user in backend)
  await page.waitForTimeout(2000)

  // Create test project via API (using the browser's auth context)
  const createProjectResp = await page.evaluate(async () => {
    const token = Object.entries(sessionStorage)
      .find(([k]) => k.startsWith('oidc.'))?.[1]
    const parsed = token ? JSON.parse(token) : null
    const accessToken = parsed?.access_token

    if (!accessToken) return { error: 'no token found' }

    // Create project
    const projResp = await fetch('/api/v1/projects', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify({
        slug: 'e2e-test',
        name: 'E2E Test',
        archetype: 'software',
      }),
    })
    if (!projResp.ok) {
      const body = await projResp.text()
      if (body.includes('duplicate') || projResp.status === 409) {
        // Project already exists — that's fine, just create the issue
      } else {
        return { error: `project: ${projResp.status} ${body}` }
      }
    }

    // Create issues (one per test to avoid state leaking)
    // #1, #2: mocked tests; #3: integration test (real drone)
    for (let i = 1; i <= 3; i++) {
      const issueResp = await fetch('/api/v1/projects/e2e-test/issues', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${accessToken}`,
        },
        body: JSON.stringify({
          title: `Refinement Phase Test ${i}`,
          type: 'task',
          description: 'Test issue for E2E refinement phase tests',
        }),
      })
      // Ignore errors for issues that may already exist
      if (!issueResp.ok) {
        const body = await issueResp.text()
        if (!body.includes('duplicate')) {
          return { error: `issue ${i}: ${issueResp.status} ${body}` }
        }
      }
    }


    return { ok: true }
  })

  console.log('Test data setup:', JSON.stringify(createProjectResp))
})
