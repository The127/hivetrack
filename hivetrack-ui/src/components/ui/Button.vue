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
import buttonVariants from './buttonVariants.js'

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

const classes = computed(() => buttonVariants({ variant: props.variant, size: props.size }))
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
