import { ref } from 'vue'
import githubLightUrl from 'highlight.js/styles/github.css?url'
import githubDarkUrl from 'highlight.js/styles/github-dark.css?url'

const STORAGE_KEY = 'hivetrack:theme'

// 'system' | 'light' | 'dark'
const theme = ref(localStorage.getItem(STORAGE_KEY) ?? 'system')
const isDark = ref(false)

function getSystemDark() {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

function syncHighlightTheme(dark) {
  let link = document.getElementById('hljs-theme')
  if (!link) {
    link = document.createElement('link')
    link.id = 'hljs-theme'
    link.rel = 'stylesheet'
    document.head.appendChild(link)
  }
  link.href = dark ? githubDarkUrl : githubLightUrl
}

function apply() {
  const dark = theme.value === 'dark' || (theme.value === 'system' && getSystemDark())
  isDark.value = dark
  document.documentElement.classList.toggle('dark', dark)
  syncHighlightTheme(dark)
}

export function initTheme() {
  apply()
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', apply)
}

export function useTheme() {
  return {
    theme,
    isDark,
    setTheme(value) {
      theme.value = value
      localStorage.setItem(STORAGE_KEY, value)
      apply()
    },
  }
}
