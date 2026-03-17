<!--
  Button — the primary interactive element.

  Variants:
    primary     Blue filled. Default action.
    secondary   Subtle filled. Secondary action.
    ghost       No background. Tertiary / toolbar action.
    destructive Red filled. Irreversible actions only.
    link        Inline text link appearance.

  Sizes:
    sm   h-7  / text-xs  — compact, toolbar use
    md   h-8  / text-sm  — default
    lg   h-9  / text-sm  — prominent calls to action

  Renders as <button> by default. Pass `as="a"` or `as="RouterLink"`
  to render as another element while keeping button styles.

  @example
  <Button @click="save">Save changes</Button>
  <Button variant="ghost" size="sm" :loading="isSaving">
    <Trash2Icon class="size-4" /> Delete
  </Button>
  <Button variant="destructive" @click="confirmDelete">Delete project</Button>
-->
<script setup>
import { computed } from 'vue'
import Spinner from '@/components/ui/Spinner.vue'

const props = defineProps({
  /** Visual style of the button. */
  variant: {
    type: String,
    default: 'primary',
    validator: (v) => ['primary', 'secondary', 'ghost', 'destructive', 'link'].includes(v),
  },
  /** Size of the button. */
  size: {
    type: String,
    default: 'md',
    validator: (v) => ['sm', 'md', 'lg'].includes(v),
  },
  /** Shows a loading spinner and disables interaction. */
  loading: {
    type: Boolean,
    default: false,
  },
  /** Disables the button. */
  disabled: {
    type: Boolean,
    default: false,
  },
  /** HTML type attribute. */
  type: {
    type: String,
    default: 'button',
  },
  /** Renders as a different element (e.g. 'a', RouterLink). */
  as: {
    type: [String, Object],
    default: 'button',
  },
})

const variantClasses = {
  primary:
    'bg-blue-600 text-white hover:bg-blue-700 active:bg-blue-800 focus-visible:ring-blue-500',
  secondary:
    'bg-slate-100 text-slate-700 hover:bg-slate-200 active:bg-slate-300 focus-visible:ring-slate-400',
  ghost:
    'text-slate-600 hover:bg-slate-100 hover:text-slate-900 active:bg-slate-200 focus-visible:ring-slate-400',
  destructive:
    'bg-red-600 text-white hover:bg-red-700 active:bg-red-800 focus-visible:ring-red-500',
  link: 'text-blue-600 hover:underline focus-visible:ring-blue-500 h-auto px-0',
}

const sizeClasses = {
  sm: 'h-7 px-2.5 text-xs gap-1',
  md: 'h-8 px-3.5 text-sm gap-1.5',
  lg: 'h-9 px-4 text-sm gap-2',
}

const classes = computed(() => [
  'inline-flex items-center justify-center rounded-md font-medium',
  'transition-colors duration-150',
  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-1',
  'disabled:opacity-50 disabled:cursor-not-allowed disabled:pointer-events-none',
  variantClasses[props.variant],
  props.variant !== 'link' ? sizeClasses[props.size] : '',
])
</script>

<template>
  <component
    :is="as"
    :type="as === 'button' ? type : undefined"
    :class="classes"
    :disabled="disabled || loading"
    v-bind="$attrs"
  >
    <Spinner v-if="loading" class="size-3.5 opacity-70" />
    <slot />
  </component>
</template>
