<script setup>
import Badge from "@/components/ui/Badge.vue";

defineProps({
  project: { type: Object, required: true },
});
</script>

<template>
  <div class="flex items-center gap-2 min-w-0">
    <span
      class="size-7 rounded flex items-center justify-center text-xs font-semibold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 flex-shrink-0"
    >
      {{ project.slug.slice(0, 2).toUpperCase() }}
    </span>
    <span class="font-semibold text-slate-900 dark:text-slate-100 truncate">{{
      project.name
    }}</span>
    <Badge
      :color-scheme="project.archetype === 'software' ? 'blue' : 'teal'"
      compact
    >
      {{ project.archetype }}
    </Badge>
  </div>
</template>
