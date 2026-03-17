<!--
  Badge — compact inline label for status, priority, count, or category.

  Use the `colorScheme` prop for semantically coloured badges.
  Use the `dot` prop with `dotColor` for project-label style coloured dots.

  Status → colorScheme mapping (use StatusBadge for automatic mapping):
    todo         → gray
    in_progress  → blue
    in_review    → violet
    done         → green
    cancelled    → gray (muted)
    open         → sky
    in_progress  → blue
    resolved     → teal
    closed       → gray

  Priority → colorScheme mapping (use PriorityBadge for automatic mapping):
    none     → gray
    low      → sky
    medium   → amber
    high     → orange
    critical → red

  @example
  <Badge colorScheme="green">done</Badge>
  <Badge colorScheme="amber">medium</Badge>
  <Badge dot dotColor="#e2b340">bug</Badge>
  <Badge colorScheme="gray">42</Badge>
-->
<script setup>
import { computed } from 'vue'

const props = defineProps({
  /**
   * Colour palette for the badge.
   * Maps to a background/text/border combination.
   */
  colorScheme: {
    type: String,
    default: 'gray',
    validator: (v) =>
      ['gray', 'blue', 'green', 'red', 'amber', 'orange', 'violet', 'teal', 'sky', 'pink'].includes(v),
  },
  /** Show a coloured dot before the label (for issue labels). */
  dot: {
    type: Boolean,
    default: false,
  },
  /** Hex colour for the dot. Only relevant when `dot` is true. */
  dotColor: {
    type: String,
    default: null,
  },
  /** Compact size with less padding. */
  compact: {
    type: Boolean,
    default: false,
  },
})

const schemeClasses = {
  gray: 'bg-slate-100 text-slate-600 ring-slate-200',
  blue: 'bg-blue-50 text-blue-700 ring-blue-200',
  green: 'bg-emerald-50 text-emerald-700 ring-emerald-200',
  red: 'bg-red-50 text-red-700 ring-red-200',
  amber: 'bg-amber-50 text-amber-700 ring-amber-200',
  orange: 'bg-orange-50 text-orange-700 ring-orange-200',
  violet: 'bg-violet-50 text-violet-700 ring-violet-200',
  teal: 'bg-teal-50 text-teal-700 ring-teal-200',
  sky: 'bg-sky-50 text-sky-700 ring-sky-200',
  pink: 'bg-pink-50 text-pink-700 ring-pink-200',
}

const classes = computed(() => [
  'inline-flex items-center gap-1 rounded font-medium ring-1',
  props.compact ? 'px-1.5 py-0 text-xs leading-5' : 'px-2 py-0.5 text-xs',
  schemeClasses[props.colorScheme],
])
</script>

<template>
  <span :class="classes">
    <span
      v-if="dot"
      class="size-1.5 rounded-full flex-shrink-0"
      :style="dotColor ? { backgroundColor: dotColor } : {}"
    />
    <slot />
  </span>
</template>
