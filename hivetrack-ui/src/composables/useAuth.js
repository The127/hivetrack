import { ref, computed } from 'vue'
import { UserManager } from 'oidc-client-ts'

// Module-level shared state — one instance for the lifetime of the app.
const user = ref(null)
const isLoading = ref(true)
const initError = ref(null)

/** @type {UserManager | null} */
let _manager = null

/**
 * Fetches OIDC config from the backend and initialises oidc-client-ts.
 * Call once at app startup (main.js or App.vue).
 */
export async function initAuth() {
  try {
    const config = await fetch('/api/v1/auth/oidc-config').then((r) => r.json())

    _manager = new UserManager({
      authority: config.authority,
      client_id: config.client_id,
      redirect_uri: `${window.location.origin}/auth/callback`,
      post_logout_redirect_uri: window.location.origin,
      scope: 'openid profile email',
      response_type: 'code',
      automaticSilentRenew: true,
    })

    user.value = await _manager.getUser()

    _manager.events.addUserLoaded((u) => {
      user.value = u
    })
    _manager.events.addUserUnloaded(() => {
      user.value = null
    })
    _manager.events.addSilentRenewError(() => {
      // Token renewal failed — user needs to sign in again
      user.value = null
    })
  } catch (e) {
    initError.value = e
  } finally {
    isLoading.value = false
  }
}

/**
 * Auth composable. Returns reactive auth state and actions.
 *
 * @example
 * const { user, isAuthenticated, signIn, signOut } = useAuth()
 */
export function useAuth() {
  return {
    /** The current OIDC user object, or null if not signed in. */
    user: computed(() => user.value),

    /** True when a valid, non-expired user session exists. */
    isAuthenticated: computed(() => !!user.value && !user.value.expired),

    /** True while the initial auth check is in progress. */
    isLoading: computed(() => isLoading.value),

    /** Non-null if auth initialisation failed (e.g. backend unreachable). */
    initError: computed(() => initError.value),

    /** Redirects the user to the OIDC provider login page. */
    signIn: () => _manager?.signinRedirect() ?? Promise.resolve(),

    /** Redirects the user to the OIDC provider logout page. */
    signOut: () => _manager?.signoutRedirect(),

    /** Returns the current Bearer token string, or null. */
    getAccessToken: () => user.value?.access_token ?? null,

    /** Completes the redirect callback after returning from the OIDC provider. */
    handleCallback: () => _manager?.signinRedirectCallback(),
  }
}
