import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { VueQueryPlugin } from '@tanstack/vue-query'
import App from './App.vue'
import './style.css'
import { initAuth, useAuth } from '@/composables/useAuth'

// ── Router ────────────────────────────────────────────────────────────────────
// Routes are added as views are built.
// meta.requiresAuth = true  → redirect to OIDC login if not authenticated
// meta.layout       = name  → hint for which layout the view uses (documentation)

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: () => import('./views/HomeView.vue'),
      meta: { requiresAuth: true, layout: 'main' },
    },
    {
      path: '/auth/callback',
      component: () => import('./views/AuthCallbackView.vue'),
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

// ── Bootstrap ─────────────────────────────────────────────────────────────────

// Initialise OIDC before installing the router — app.use(router) triggers the
// initial navigation synchronously, so the guard must have auth state ready.
await initAuth()

const app = createApp(App)
app.use(router)
app.use(VueQueryPlugin)
app.mount('#app')
