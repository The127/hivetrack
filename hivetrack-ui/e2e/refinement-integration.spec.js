/**
 * E2E Integration: Full refinement flow with a real Claude drone
 *
 * Requires hivemind running on :8080/:50051 and a Claude drone registered
 * for the e2e-test project with the 'refinement' capability.
 * Started automatically by `just ui-e2e` / `just ui-e2e-ui` via _e2e-services.
 *
 * Creates a fresh issue before each test so there is never leftover session
 * state from a previous run.
 *
 * These tests are intentionally slow — they drive real LLM conversations
 * across all four refinement phases.
 */
import { test, expect } from "@playwright/test";

const HIVEMIND_MGMT = "http://localhost:8080";
const PROJECT_SLUG = "e2e-test";

async function hivemindAvailable() {
  try {
    const res = await fetch(`${HIVEMIND_MGMT}/api/v1/drones`);
    return res.ok;
  } catch {
    return false;
  }
}

async function droneAvailableForProject(slug) {
  try {
    const res = await fetch(
      `${HIVEMIND_MGMT}/api/v1/drones?project_slug=${slug}`,
    );
    if (!res.ok) return false;
    const body = await res.json();
    return (
      Array.isArray(body.drones) &&
      body.drones.some((d) => d.status === "available")
    );
  } catch {
    return false;
  }
}

/**
 * Create a fresh issue using the browser's OIDC token from sessionStorage.
 * Mirrors the pattern in global-setup.js since page.request doesn't carry
 * Bearer tokens stored in sessionStorage.
 */
async function createFreshIssue(page, projectSlug) {
  return page.evaluate(async (slug) => {
    const entry = Object.entries(sessionStorage).find(([k]) =>
      k.startsWith("oidc."),
    )?.[1];
    const accessToken = entry ? JSON.parse(entry).access_token : null;
    if (!accessToken) throw new Error("no OIDC access_token in sessionStorage");

    const resp = await fetch(`/api/v1/projects/${slug}/issues`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify({
        title: "PR Review Automation Tool",
        type: "task",
        description:
          "As a software developer I want to automatically review pull requests " +
          "so that code quality checks happen without manual effort.",
      }),
    });
    if (!resp.ok)
      throw new Error(`create issue: ${resp.status} ${await resp.text()}`);
    const body = await resp.json();
    return body.Number;
  }, projectSlug);
}

/**
 * Wait for the drone to finish responding.
 *
 * The "Hivemind is thinking..." spinner appears when is_generating=true and
 * partial_response=''. After the drone finishes, it disappears.
 *
 * If the drone is so fast that the thinking spinner never appears during a
 * 2s polling window, we fall back to a short stabilisation wait before
 * checking the next actionable element.
 */
async function waitForDroneResponse(page) {
  const thinkingText = page.getByText("Hivemind is thinking...");

  // Wait up to 15s for the thinking state to appear.
  const sawThinking = await expect(thinkingText)
    .toBeVisible({ timeout: 15_000 })
    .then(() => true)
    .catch(() => false);

  if (sawThinking) {
    // Drone is processing — wait until it's done (generous timeout for LLM).
    await expect(thinkingText).not.toBeVisible({ timeout: 90_000 });
  } else {
    // Drone responded before the 2s poll could capture the thinking state.
    // The next poll has already updated the UI — nothing more to wait for.
  }
}

test.describe("Refinement integration (real drone)", () => {
  // Full LLM conversation across 4 phases: allow generous time.
  test.setTimeout(15 * 60_000);

  let issueNumber;

  test.beforeEach(async ({ page }) => {
    if (!(await hivemindAvailable())) {
      test.skip("Hivemind not reachable on :8080");
    }
    if (!(await droneAvailableForProject(PROJECT_SLUG))) {
      test.skip(`No drone registered for project ${PROJECT_SLUG}`);
    }

    // Navigate first so sessionStorage (OIDC token) is populated.
    await page.goto(`/projects/${PROJECT_SLUG}/overview`);
    await page.waitForLoadState("networkidle");

    issueNumber = await createFreshIssue(page, PROJECT_SLUG);
  });

  test("full refinement flow from start to accepted proposal", async ({
    page,
  }) => {
    await page.goto(`/projects/${PROJECT_SLUG}/issues/${issueNumber}`);
    await page.waitForLoadState("networkidle");

    // Open refinement panel.
    const refineButton = page.getByRole("button", { name: /refine/i });
    await expect(refineButton).toBeVisible({ timeout: 10_000 });
    await refineButton.click();

    // Fresh issue always shows "Start refinement" button.
    await expect(
      page.getByRole("button", { name: /start refinement/i }),
    ).toBeVisible({ timeout: 10_000 });
    await page.getByRole("button", { name: /start refinement/i }).click();

    // Wait for the drone to deliver the first question before entering the loop.
    await waitForDroneResponse(page);

    // Drive the full conversation:
    //   - When an answer input is visible → send a concise answer, then wait
    //     for the drone to respond before the next iteration.
    //   - When "Confirm & continue" is visible → advance to next phase, then
    //     wait for the drone to open the new phase.
    //   - When "Accept & apply" is visible → accept the proposal and finish.
    //
    // Up to 30 exchanges (well above any realistic 4-phase flow).
    const ANSWER = "Software developer who wants to automate code reviews";

    for (let step = 0; step < 30; step++) {
      const acceptBtn = page.getByRole("button", { name: /accept.*apply/i });
      const confirmBtn = page.getByRole("button", {
        name: /confirm.*continue/i,
      });
      const answerInput = page.getByPlaceholder(/answer/i).first();

      // The next actionable state is already set up by waitForDroneResponse.
      if (await acceptBtn.isVisible()) {
        await acceptBtn.click();
        break;
      }

      if (await confirmBtn.isVisible()) {
        await confirmBtn.click();
        await waitForDroneResponse(page);
        continue;
      }

      // Question is ready — answer it, then wait for drone to respond.
      await expect(answerInput).toBeVisible({ timeout: 5_000 });
      await answerInput.fill(ANSWER);
      await page.keyboard.press("Meta+Enter");
      await waitForDroneResponse(page);
    }

    // After accepting the proposal:
    // - session.status → "completed"
    // - issue.refined = true
    // - panel closes automatically
    // - "Refine" button disappears (only shown when !issue.refined)
    await expect(page.getByRole("button", { name: /refine/i })).not.toBeVisible(
      { timeout: 30_000 },
    );
  });
});
