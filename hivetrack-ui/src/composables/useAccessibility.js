import { ref } from 'vue'

const STORAGE_PREFIX = 'hivetrack:'

const colorVision = ref(localStorage.getItem(`${STORAGE_PREFIX}colorVision`) ?? 'normal')
const highContrast = ref(localStorage.getItem(`${STORAGE_PREFIX}highContrast`) === 'true')
const font = ref(localStorage.getItem(`${STORAGE_PREFIX}font`) ?? 'default')

function apply() {
  const root = document.documentElement

  // Build combined filter string
  const filters = []
  if (colorVision.value !== 'normal') filters.push(`url(#cv-${colorVision.value})`)
  if (highContrast.value) filters.push('contrast(1.25)')
  root.style.filter = filters.join(' ') || ''

  // Font
  root.classList.remove('font-system', 'font-dyslexia')
  if (font.value === 'system') root.classList.add('font-system')
  if (font.value === 'dyslexia') root.classList.add('font-dyslexia')
}

export function initAccessibility() {
  apply()
}

export function useAccessibility() {
  return {
    colorVision,
    highContrast,
    font,

    setColorVision(value) {
      colorVision.value = value
      localStorage.setItem(`${STORAGE_PREFIX}colorVision`, value)
      apply()
    },

    setHighContrast(value) {
      highContrast.value = value
      localStorage.setItem(`${STORAGE_PREFIX}highContrast`, String(value))
      apply()
    },

    setFont(value) {
      font.value = value
      localStorage.setItem(`${STORAGE_PREFIX}font`, value)
      apply()
    },
  }
}
