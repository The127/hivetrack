<!--
  OverviewWipSettings — WIP limit configuration for software projects.
-->
<script setup>
import { computed, ref, watch } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { SlidersHorizontalIcon, CircleDotIcon, GitPullRequestIcon } from 'lucide-vue-next'
import { fetchProject, updateProject } from '@/api/projects'

const props = defineProps({
  slug: { type: String, required: true },
})

const queryClient = useQueryClient()

const { data: project } = useQuery({
  queryKey: computed(() => ['project', props.slug]),
  queryFn: () => fetchProject(props.slug),
})

const wipInProgressInput = ref(null)
const wipInReviewInput = ref(null)

watch(project, (p) => {
  if (p) {
    wipInProgressInput.value = p.wip_limit_in_progress ?? ''
    wipInReviewInput.value = p.wip_limit_in_review ?? ''
  }
}, { immediate: true })

const { mutate: saveWipLimits, isPending: savingWip } = useMutation({
  mutationFn: ({ field, value }) => {
    const body = {}
    body[field] = value
    return updateProject(project.value.id, body)
  },
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['project', props.slug] })
  },
})

function saveWipInProgress() {
  const raw = wipInProgressInput.value
  const parsed = raw === '' || raw === null ? null : parseInt(raw, 10)
  if (parsed === null || !isNaN(parsed)) {
    saveWipLimits({ field: 'wip_limit_in_progress', value: parsed })
  }
}

function saveWipInReview() {
  const raw = wipInReviewInput.value
  const parsed = raw === '' || raw === null ? null : parseInt(raw, 10)
  if (parsed === null || !isNaN(parsed)) {
    saveWipLimits({ field: 'wip_limit_in_review', value: parsed })
  }
}
</script>

<template>
  <section v-if="project?.archetype === 'software'">
    <h2 class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-3 flex items-center gap-1.5">
      <SlidersHorizontalIcon class="size-4 text-slate-500 dark:text-slate-400" />
      Board
    </h2>
    <div class="rounded-lg border border-slate-200 dark:border-slate-700 divide-y divide-slate-100 dark:divide-slate-800 overflow-hidden">
      <div class="flex items-center gap-3 px-4 py-2.5">
        <CircleDotIcon class="size-4 flex-shrink-0 text-blue-500" />
        <span class="text-sm text-slate-700 dark:text-slate-300 flex-1">In Progress limit</span>
        <input
          v-model="wipInProgressInput"
          type="number"
          min="1"
          placeholder="None"
          class="w-20 text-sm text-right border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 rounded px-2 py-0.5 text-slate-700 dark:text-slate-300 focus:outline-none focus:ring-1 focus:ring-blue-400 focus:border-blue-400"
          :disabled="savingWip"
          @blur="saveWipInProgress"
          @keydown.enter="$event.target.blur()"
        />
      </div>
      <div class="flex items-center gap-3 px-4 py-2.5">
        <GitPullRequestIcon class="size-4 flex-shrink-0 text-violet-500" />
        <span class="text-sm text-slate-700 dark:text-slate-300 flex-1">In Review limit</span>
        <input
          v-model="wipInReviewInput"
          type="number"
          min="1"
          placeholder="None"
          class="w-20 text-sm text-right border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 rounded px-2 py-0.5 text-slate-700 dark:text-slate-300 focus:outline-none focus:ring-1 focus:ring-blue-400 focus:border-blue-400"
          :disabled="savingWip"
          @blur="saveWipInReview"
          @keydown.enter="$event.target.blur()"
        />
      </div>
    </div>
    <p class="text-xs text-slate-400 dark:text-slate-500 mt-1.5">Informational only — the board highlights columns that exceed these limits.</p>
  </section>
</template>
