import { test, expect } from "@playwright/test";

/**
 * E2E: Refinement phase-gated UI
 *
 * Mocks the session API so no real drone or Hivemind is needed.
 * Tests only that the panel and phase stepper render correctly.
 *
 * Uses project e2e-test with issues #1 and #2 (created by global-setup).
 * Each test uses a separate issue to avoid state leaking.
 */

const PROJECT_SLUG = "e2e-test";

const SESSION_1 = "**/api/v1/projects/e2e-test/issues/1/refinement/session";

const ACTIVE_SESSION = {
  id: "aaaaaaaa-0000-0000-0000-000000000002",
  issue_id: "bbbbbbbb-0000-0000-0000-000000000002",
  status: "active",
  current_phase: "actor_goal",
  messages: [],
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
  partial_response: "",
  is_generating: false,
};

test.describe("Refinement Phase UI", () => {
  test("refinement panel opens and shows start button when no session", async ({
    page,
  }) => {
    // Return 204 (no content) — apiFetch converts this to null, so session = null,
    // and TanStack Query does not retry (no error thrown).
    await page.route(SESSION_1, (route) => route.fulfill({ status: 204 }));

    await page.goto(`/projects/${PROJECT_SLUG}/issues/1`);
    await page.waitForLoadState("networkidle");

    const refineButton = page.getByRole("button", { name: /refine/i });
    await expect(refineButton).toBeVisible({ timeout: 10_000 });
    await refineButton.click();

    await expect(page.getByText("Refinement").first()).toBeVisible({
      timeout: 5_000,
    });
    await expect(
      page.getByRole("button", { name: /start refinement/i }),
    ).toBeVisible({ timeout: 5_000 });
  });

  test("phase stepper shows all five phases when refinement panel is opened", async ({
    page,
  }) => {
    // First fetch returns 204 (so "Refine" button shows), subsequent fetches return ACTIVE_SESSION
    let callCount = 0;
    await page.route(SESSION_1, (route) => {
      callCount++;
      if (callCount === 1) {
        route.fulfill({ status: 204 });
      } else {
        route.fulfill({ json: ACTIVE_SESSION });
      }
    });

    await page.goto(`/projects/${PROJECT_SLUG}/issues/1`);
    await page.waitForLoadState("networkidle");

    const refineButton = page.getByRole("button", { name: /refine/i });
    await expect(refineButton).toBeVisible({ timeout: 10_000 });
    await refineButton.click();

    // All five phases must appear in the phase stepper (wait for session to load after panel opens)
    await expect(page.getByText("Actor & Goal").first()).toBeVisible({
      timeout: 10_000,
    });
    await expect(page.getByText("Main Scenario").first()).toBeVisible();
    await expect(page.getByText("Extensions").first()).toBeVisible();
    await expect(page.getByText("Acceptance Criteria").first()).toBeVisible();
    await expect(page.getByText("BDD Scenarios").first()).toBeVisible();
  });

  test("story progress shows BDD scenarios when bdd_scenarios phase_result is present", async ({
    page,
  }) => {
    const sessionWithBdd = {
      id: "aaaaaaaa-0000-0000-0000-000000000001",
      issue_id: "bbbbbbbb-0000-0000-0000-000000000001",
      status: "active",
      current_phase: "acceptance_criteria",
      messages: [
        {
          id: "msg-proposal",
          role: "assistant",
          content:
            '{"type":"proposal","title":"Password Reset","description":"## User Story"}',
          message_type: "proposal",
          proposal: { title: "Password Reset", description: "## User Story" },
          phase: "acceptance_criteria",
          phase_data: null,
          created_at: "2026-01-01T00:01:00Z",
        },
        {
          id: "msg-bdd",
          role: "assistant",
          content: "",
          message_type: "phase_result",
          phase: "bdd_scenarios",
          phase_data: {
            scenarios: [
              {
                name: "Successful password reset",
                given: ["a registered user with email user@example.com"],
                when: ["they request a password reset"],
                then: ["a reset email is sent"],
              },
            ],
          },
          created_at: "2026-01-01T00:02:00Z",
        },
      ],
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
      partial_response: "",
      is_generating: false,
    };

    // First fetch returns 204 (so "Refine" button shows), subsequent fetches return sessionWithBdd
    let callCount = 0;
    await page.route(SESSION_1, (route) => {
      callCount++;
      if (callCount === 1) {
        route.fulfill({ status: 204 });
      } else {
        route.fulfill({ json: sessionWithBdd });
      }
    });

    await page.goto(`/projects/${PROJECT_SLUG}/issues/1`);
    await page.waitForLoadState("networkidle");

    const refineButton = page.getByRole("button", { name: /refine/i });
    await expect(refineButton).toBeVisible({ timeout: 10_000 });
    await refineButton.click();

    // BDD Scenarios phase should appear in Story Progress
    await expect(page.getByText("BDD Scenarios").first()).toBeVisible({
      timeout: 10_000,
    });

    // The scenario name and steps should be visible in Story Progress
    await expect(page.getByText("Successful password reset")).toBeVisible({
      timeout: 5_000,
    });
    await expect(
      page.getByText("a registered user with email user@example.com"),
    ).toBeVisible();
    await expect(page.getByText("they request a password reset")).toBeVisible();
    await expect(page.getByText("a reset email is sent")).toBeVisible();
  });
});
