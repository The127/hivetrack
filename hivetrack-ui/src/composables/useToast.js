import { reactive } from 'vue'

let nextId = 0
const toasts = reactive([])

const DURATION = 4000 // ms

function add(message, variant = 'success') {
  const id = ++nextId
  toasts.push({ id, message, variant })
  setTimeout(() => remove(id), DURATION)
}

function remove(id) {
  const idx = toasts.findIndex((t) => t.id === id)
  if (idx !== -1) toasts.splice(idx, 1)
}

export function useToast() {
  return {
    toasts,
    success: (message) => add(message, 'success'),
    error: (message) => add(message, 'error'),
    info: (message) => add(message, 'info'),
    remove,
  }
}
