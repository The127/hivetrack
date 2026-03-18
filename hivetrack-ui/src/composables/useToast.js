import { reactive } from 'vue'

let nextId = 0
const toasts = reactive([])

const DURATION = 8000 // ms

function add(message, variant = 'success', to = null) {
  const id = ++nextId
  toasts.push({ id, message, variant, to })
  setTimeout(() => remove(id), DURATION)
}

function remove(id) {
  const idx = toasts.findIndex((t) => t.id === id)
  if (idx !== -1) toasts.splice(idx, 1)
}

export function useToast() {
  return {
    toasts,
    success: (message, to = null) => add(message, 'success', to),
    error: (message, to = null) => add(message, 'error', to),
    info: (message, to = null) => add(message, 'info', to),
    remove,
  }
}
