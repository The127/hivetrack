import { createApp } from 'vue'
import { VueQueryPlugin } from '@tanstack/vue-query'
import App from './App.vue'
import './style.css'
import { initAuth } from '@/composables/useAuth'
import { initTheme } from '@/composables/useTheme'
import { initAccessibility } from '@/composables/useAccessibility'
import router from '@/router/index.js'

// Apply theme + accessibility before first render to avoid flash.
initTheme()
initAccessibility()

// Initialise OIDC before installing the router — app.use(router) triggers the
// initial navigation synchronously, so the guard must have auth state ready.
await initAuth()

const app = createApp(App)
app.use(router)
app.use(VueQueryPlugin)
app.mount('#app')
