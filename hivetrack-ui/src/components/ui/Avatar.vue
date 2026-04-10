<!--
  Avatar — user avatar with image fallback to coloured initials.

  When `src` is provided and loads successfully, the image is shown.
  Otherwise, up to two initials are derived from `name` and displayed
  on a deterministically coloured background.

  Sizes:
    xs   size-5  (20px) — inline in dense lists
    sm   size-6  (24px) — issue cards, compact rows
    md   size-8  (32px) — default
    lg   size-10 (40px) — profile sections

  @example
  <Avatar name="Jane Smith" />
  <Avatar name="Jane Smith" src="https://..." size="sm" />
  <Avatar name="Jane Smith" size="lg" />
-->
<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  /** Full name of the user. Used to derive initials and background colour. */
  name: {
    type: String,
    required: true,
  },
  /** URL of the avatar image. Falls back to initials if not provided or on error. */
  src: {
    type: String,
    default: null,
  },
  /** Size of the avatar. */
  size: {
    type: String,
    default: 'md',
    validator: (v) => ['xs', 'sm', 'md', 'lg'].includes(v),
  },
})

const imgError = ref(false)

const initials = computed(() => {
  const trimmed = props.name?.trim() ?? ''
  if (!trimmed) return '?'
  const parts = trimmed.split(/\s+/)
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase() || '?'
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
})

// Deterministic colour based on name — same name always gets same colour.
const PALETTE = [
  'bg-blue-500',
  'bg-violet-500',
  'bg-emerald-500',
  'bg-amber-500',
  'bg-rose-500',
  'bg-sky-500',
  'bg-teal-500',
  'bg-orange-500',
]

const bgColor = computed(() => {
  const name = props.name ?? ''
  if (!name) return PALETTE[0]
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return PALETTE[Math.abs(hash) % PALETTE.length]
})

const sizeClasses = {
  xs: 'size-5 text-[10px]',
  sm: 'size-6 text-[11px]',
  md: 'size-8 text-xs',
  lg: 'size-10 text-sm',
}

const showImage = computed(() => props.src && !imgError.value)
</script>

<template>
  <span
    :class="[
      'inline-flex items-center justify-center rounded-full font-semibold text-white flex-shrink-0 select-none',
      sizeClasses[size],
      !showImage ? bgColor : '',
    ]"
    :title="name"
  >
    <img
      v-if="showImage"
      :src="src"
      :alt="name"
      class="size-full rounded-full object-cover"
      @error="imgError = true"
    />
    <template v-else>{{ initials }}</template>
  </span>
</template>
