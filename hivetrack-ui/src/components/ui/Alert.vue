<!--
  Alert — inline contextual message with an optional icon.

  Use for non-blocking notices, hints, and feedback that lives inline
  within a page rather than as a toast or modal.

  Props:
    variant — controls colour palette and default icon
      'info'    (default) — slate, InfoIcon
      'warning'           — amber, AlertTriangleIcon
      'success'           — green, CheckCircleIcon
      'error'             — red, XCircleIcon

  Slots:
    default — message content (text, links, anything inline)
    icon    — override the default icon

  @example
  <Alert>Showing backlog — no active sprint.</Alert>
  <Alert variant="warning">This action cannot be undone.</Alert>
  <Alert variant="success">Changes saved.</Alert>
-->
<script setup>
import { computed } from 'vue'
import { InfoIcon, AlertTriangleIcon, CheckCircleIcon, XCircleIcon } from 'lucide-vue-next'

const props = defineProps({
  variant: {
    type: String,
    default: 'info',
    validator: (v) => ['info', 'warning', 'success', 'error'].includes(v),
  },
})

const SCHEME = {
  info:    { wrapper: 'bg-slate-50 dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400',  icon: InfoIcon,             iconClass: 'text-slate-400 dark:text-slate-500'  },
  warning: { wrapper: 'bg-amber-50 dark:bg-amber-900/30 border-amber-200 dark:border-amber-700/50 text-amber-800 dark:text-amber-300',  icon: AlertTriangleIcon,    iconClass: 'text-amber-500 dark:text-amber-400'  },
  success: { wrapper: 'bg-green-50 dark:bg-green-900/30 border-green-200 dark:border-green-700/50 text-green-800 dark:text-green-300',  icon: CheckCircleIcon,      iconClass: 'text-green-500 dark:text-green-400'  },
  error:   { wrapper: 'bg-red-50 dark:bg-red-900/30 border-red-200 dark:border-red-700/50 text-red-800 dark:text-red-300',    icon: XCircleIcon,          iconClass: 'text-red-500 dark:text-red-400'    },
}

const scheme = computed(() => SCHEME[props.variant])
</script>

<template>
  <div
    class="flex items-start gap-2.5 rounded-md border px-3.5 py-2.5 text-sm"
    :class="scheme.wrapper"
  >
    <slot name="icon">
      <component :is="scheme.icon" class="size-4 flex-shrink-0 mt-px" :class="scheme.iconClass" />
    </slot>
    <div class="min-w-0">
      <slot />
    </div>
  </div>
</template>
