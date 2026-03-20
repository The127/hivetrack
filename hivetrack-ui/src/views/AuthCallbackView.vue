<!--
  AuthCallbackView — handles the OIDC redirect callback.

  The OIDC provider redirects here after a successful login with the
  authorisation code in the URL. This view completes the exchange,
  stores the token, and navigates to the intended destination.
-->
<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import Spinner from '@/components/ui/Spinner.vue'

const router = useRouter()
const { handleCallback } = useAuth()

onMounted(async () => {
  try {
    await handleCallback()
    // Return to the page the user originally wanted, or dashboard.
    const returnUrl = sessionStorage.getItem('oidc_return_url') ?? '/'
    sessionStorage.removeItem('oidc_return_url')
    router.replace(returnUrl)
  } catch (e) {
    console.error('[Hivetrack] Auth callback failed', e)
    router.replace('/')
  }
})
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-slate-950">
    <div class="flex flex-col items-center gap-3 text-slate-500 dark:text-slate-400">
      <Spinner class="size-6" />
      <p class="text-sm">Signing you in…</p>
    </div>
  </div>
</template>
