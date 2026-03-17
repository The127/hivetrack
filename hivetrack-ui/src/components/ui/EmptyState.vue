<!--
  EmptyState — shown when a list or section has no items.

  Every empty state should explain what's missing and offer a clear
  next step (per the "no dead ends" principle).

  @example
  <!-- Empty backlog -->
  <EmptyState
    title="Nothing in the backlog"
    description="Issues that haven't been added to a sprint live here."
    action-label="Create issue"
    @action="openCreateDialog"
  />

  <!-- Empty state without an action (read-only view) -->
  <EmptyState
    title="No milestones yet"
    description="Milestones mark target dates for groups of issues."
  />
-->
<script setup>
defineProps({
  /** Short, descriptive heading. */
  title: {
    type: String,
    required: true,
  },
  /** One or two sentences explaining the empty state and what the user can do. */
  description: {
    type: String,
    default: null,
  },
  /** Label for the primary action button. Omit to render without an action. */
  actionLabel: {
    type: String,
    default: null,
  },
})

defineEmits([
  /** Emitted when the action button is clicked. */
  'action',
])
</script>

<template>
  <div class="flex flex-col items-center justify-center gap-3 py-16 text-center">
    <!-- Icon slot — pass a Lucide icon or similar -->
    <div v-if="$slots.icon" class="text-slate-300">
      <slot name="icon" />
    </div>
    <div class="space-y-1">
      <p class="text-sm font-medium text-slate-700">{{ title }}</p>
      <p v-if="description" class="text-sm text-slate-500 max-w-xs">{{ description }}</p>
    </div>
    <button
      v-if="actionLabel"
      class="mt-1 inline-flex items-center gap-1.5 rounded-md bg-blue-600 px-3.5 h-8 text-sm font-medium text-white hover:bg-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-1 transition-colors"
      @click="$emit('action')"
    >
      {{ actionLabel }}
    </button>
  </div>
</template>
