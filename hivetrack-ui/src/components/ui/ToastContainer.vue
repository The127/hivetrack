<!--
  ToastContainer — renders global toast notifications.

  Mount once in App.vue. Toasts are triggered via useToast().

  Variants: 'success' | 'error' | 'info'

  Optional `to` prop on a toast makes the message a RouterLink.
-->
<script setup>
import { CheckCircleIcon, XCircleIcon, InfoIcon, XIcon } from 'lucide-vue-next'
import { RouterLink } from 'vue-router'
import { useToast } from '@/composables/useToast'

const { toasts, remove } = useToast()

const SCHEME = {
  success: { wrapper: 'bg-green-50 border-green-200 text-green-900', icon: CheckCircleIcon, iconClass: 'text-green-500' },
  error:   { wrapper: 'bg-red-50 border-red-200 text-red-900',       icon: XCircleIcon,     iconClass: 'text-red-500'   },
  info:    { wrapper: 'bg-slate-50 border-slate-200 text-slate-900', icon: InfoIcon,        iconClass: 'text-slate-400' },
}

function scheme(variant) {
  return SCHEME[variant] ?? SCHEME.info
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2 pointer-events-none">
      <TransitionGroup
        enter-active-class="transition-all duration-200 ease-out"
        enter-from-class="opacity-0 translate-y-2"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition-all duration-150 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 translate-y-2"
      >
        <div
          v-for="toast in toasts"
          :key="toast.id"
          class="pointer-events-auto flex items-center gap-2.5 rounded-lg border px-4 py-3 text-sm shadow-md w-80 max-w-full"
          :class="scheme(toast.variant).wrapper"
        >
          <component
            :is="scheme(toast.variant).icon"
            class="size-4 flex-shrink-0"
            :class="scheme(toast.variant).iconClass"
          />
          <RouterLink
            v-if="toast.to"
            :to="toast.to"
            class="flex-1 min-w-0 underline underline-offset-2 hover:opacity-80"
            @click="remove(toast.id)"
          >
            {{ toast.message }}
          </RouterLink>
          <span v-else class="flex-1 min-w-0">{{ toast.message }}</span>
          <button
            class="flex-shrink-0 rounded p-0.5 opacity-60 hover:opacity-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-current cursor-pointer"
            @click="remove(toast.id)"
          >
            <XIcon class="size-3.5" />
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
