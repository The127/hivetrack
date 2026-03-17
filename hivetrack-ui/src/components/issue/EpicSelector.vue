<!--
  EpicSelector — dropdown to pick an epic in a project.

  Props:
    projectSlug — project to fetch epics from
    modelValue  — current parent_id (uuid or null)

  Emits:
    update:modelValue — selected epic ID or null to clear
-->
<script setup>
import { computed } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { fetchIssues } from '@/api/issues'
import { LayersIcon } from 'lucide-vue-next'

const props = defineProps({
  projectSlug: { type: String, required: true },
  modelValue: { type: String, default: null },
})

const emit = defineEmits(['update:modelValue'])

const { data: epicsResult } = useQuery({
  queryKey: ['issues', props.projectSlug, { type: 'epic' }],
  queryFn: () => fetchIssues(props.projectSlug, { type: 'epic', limit: 200 }),
  enabled: computed(() => !!props.projectSlug),
})

const epics = computed(() => epicsResult.value?.items ?? [])

function onChange(e) {
  const val = e.target.value
  emit('update:modelValue', val || null)
}
</script>

<template>
  <div class="flex flex-col gap-1.5">
    <label class="text-xs font-medium text-slate-500 flex items-center gap-1">
      <LayersIcon class="size-3" />
      Epic
    </label>
    <select
      :value="modelValue ?? ''"
      class="w-full rounded-md border border-slate-200 px-2.5 py-1.5 text-sm text-slate-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 cursor-pointer bg-white"
      @change="onChange"
    >
      <option value="">No epic</option>
      <option v-for="epic in epics" :key="epic.id" :value="epic.id">
        {{ epic.title }}
      </option>
    </select>
  </div>
</template>
