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
  gray:   'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 ring-slate-200 dark:ring-slate-700',
  blue:   'bg-blue-50 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300 ring-blue-200 dark:ring-blue-700/50',
  green:  'bg-emerald-50 dark:bg-emerald-900/40 text-emerald-700 dark:text-emerald-300 ring-emerald-200 dark:ring-emerald-700/50',
  red:    'bg-red-50 dark:bg-red-900/40 text-red-700 dark:text-red-300 ring-red-200 dark:ring-red-700/50',
  amber:  'bg-amber-50 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300 ring-amber-200 dark:ring-amber-700/50',
  orange: 'bg-orange-50 dark:bg-orange-900/40 text-orange-700 dark:text-orange-300 ring-orange-200 dark:ring-orange-700/50',
  violet: 'bg-violet-50 dark:bg-violet-900/40 text-violet-700 dark:text-violet-300 ring-violet-200 dark:ring-violet-700/50',
  teal:   'bg-teal-50 dark:bg-teal-900/40 text-teal-700 dark:text-teal-300 ring-teal-200 dark:ring-teal-700/50',
  sky:    'bg-sky-50 dark:bg-sky-900/40 text-sky-700 dark:text-sky-300 ring-sky-200 dark:ring-sky-700/50',
  pink:   'bg-pink-50 dark:bg-pink-900/40 text-pink-700 dark:text-pink-300 ring-pink-200 dark:ring-pink-700/50',
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
