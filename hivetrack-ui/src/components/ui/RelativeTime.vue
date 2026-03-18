<!--
  RelativeTime — displays a timestamp as relative text with a tooltip showing the full date.

  Shows relative time for recent dates, falling back to an absolute date for older ones.
  The full datetime is always shown in a tooltip on hover.

  Relative format:
    < 1 min  → "just now"
    < 1 h    → "5m ago"
    < 24 h   → "3h ago"
    < 30 d   → "12d ago"
    ≥ 30 d   → "Mar 5, 2025"

  @example
  <RelativeTime :datetime="comment.created_at" />
  <RelativeTime :datetime="issue.updated_at" />
-->
<script setup>
import { computed } from 'vue'

const props = defineProps({
  /** ISO 8601 datetime string. */
  datetime: {
    type: String,
    required: true,
  },
})

function relativeText(dateStr) {
  const date = new Date(dateStr)
  const diffMs = Date.now() - date
  const diffMin = Math.floor(diffMs / 60_000)
  if (diffMin < 1) return 'just now'
  if (diffMin < 60) return `${diffMin}m ago`
  const diffHr = Math.floor(diffMin / 60)
  if (diffHr < 24) return `${diffHr}h ago`
  const diffDay = Math.floor(diffHr / 24)
  if (diffDay < 30) return `${diffDay}d ago`
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

function fullDateText(dateStr) {
  return new Date(dateStr).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

const relative = computed(() => relativeText(props.datetime))
const full = computed(() => fullDateText(props.datetime))
</script>

<template>
  <time :datetime="datetime" :title="full">{{ relative }}</time>
</template>
