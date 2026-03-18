import { createRouter, createWebHistory } from 'vue-router'
import { useAuth } from '@/composables/useAuth'

// meta.requiresAuth = true  → redirect to OIDC login if not authenticated
// meta.layout       = name  → hint for which layout the view uses (documentation)

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: () => import('@/views/HomeView.vue'),
      meta: { requiresAuth: true, layout: 'main' },
    },
    {
      path: '/projects',
      component: () => import('@/views/ProjectsView.vue'),
      meta: { requiresAuth: true, layout: 'main' },
    },
    {
      path: '/projects/:slug/overview',
      component: () => import('@/views/ProjectOverviewView.vue'),
      meta: { requiresAuth: true, layout: 'main' },
    },
    {
      path: '/projects/:slug/board',
      component: () => import('@/views/ProjectBoardView.vue'),
      meta: { requiresAuth: true, layout: 'main' },
    },
    {
      path: '/projects/:slug/backlog',
      component: () => import('@/views/ProjectBacklogView.vue'),
      meta: { requiresAuth: true, layout: 'main' },
    },
    {
      path: '/projects/:slug/epics',
      component: () => import('@/views/ProjectEpicsView.vue'),
      meta: { requiresAuth: true, layout: 'main' },
    },
    {
      path: '/projects/:slug/triage',
      component: () => import('@/views/ProjectTriageView.vue'),
      meta: { requiresAuth: true, layout: 'main' },
    },
    {
      path: '/projects/:slug/milestones',
      component: () => import('@/views/ProjectMilestonesView.vue'),
      meta: { requiresAuth: true, layout: 'main' },
    },
    {
      path: '/projects/:slug/sprints',
      component: () => import('@/views/SprintsView.vue'),
      meta: { requiresAuth: true, layout: 'main' },
    },
    {
      path: '/projects/:slug/sprints/:sprintId',
      component: () => import('@/views/SprintDetailView.vue'),
      meta: { requiresAuth: true, layout: 'main' },
    },
    {
      path: '/projects/:slug/issues/:number',
      component: () => import('@/views/IssueDetailView.vue'),
      meta: { requiresAuth: true, layout: 'main' },
    },
    {
      path: '/auth/callback',
      component: () => import('@/views/AuthCallbackView.vue'),
      meta: { requiresAuth: false },
    },
  ],
})

// Auth guard: redirect unauthenticated users to OIDC login.
router.beforeEach(async (to) => {
  if (!to.meta.requiresAuth) return true

  const { isAuthenticated, initError, signIn } = useAuth()

  // If the backend was unreachable during init (e.g. local UI-only dev),
  // allow navigation so the UI is visible. API queries will fail gracefully.
  if (initError.value) return true

  if (!isAuthenticated.value) {
    // Store the intended destination so the callback can redirect back.
    sessionStorage.setItem('oidc_return_url', to.fullPath)
    await signIn()
    return false
  }

  return true
})

export default router
