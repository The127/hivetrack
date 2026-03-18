<!--
  SprintsView — list of completed sprints for a project.

  Shows all completed sprints sorted newest first, with issue counts.
  Clicking a sprint navigates to the sprint detail page.
-->
<script setup>
import { computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'
import { ChevronRightIcon } from 'lucide-vue-next'
import MainLayout from '@/layouts/MainLayout.vue'
import Spinner from '@/components/ui/Spinner.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ProgressBar from '@/components/ui/ProgressBar.vue'
import { fetchSprints } from '@/api/sprints'

const route = useRoute()
const slug = computed(() => route.params.slug)

const { data: sprintsResult, isLoading } = useQuery({
  queryKey: ['sprints', slug],
  queryFn: () => fetchSprints(slug.value),
  enabled: computed(() => !!slug.value),
})

const completedSprints = computed(() => {
  const sprints = sprintsResult.value?.sprints ?? []
  return sprints
    .filter((s) => s.status === 'completed')
    .sort((a, b) => {
      const aEnd = a.end_date ? new Date(a.end_date) : new Date(0)
      const bEnd = b.end_date ? new Date(b.end_date) : new Date(0)
      return bEnd - aEnd
    })
})

function formatDate(dateStr) {
  if (!dateStr) return null
  return new Date(dateStr).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

function dateRange(sprint) {
  const start = formatDate(sprint.start_date)
  const end = formatDate(sprint.end_date)
  if (start && end) return `${start} – ${end}`
  if (start) return `Started ${start}`
  if (end) return `Ended ${end}`
  return null
}
</script>

<template>
  <MainLayout>
    <div class="max-w-3xl mx-auto px-6 py-8">

      <!-- Header -->
      <div class="mb-6">
        <h1 class="text-lg font-semibold text-slate-900">Completed Sprints</h1>
        <p class="text-sm text-slate-500 mt-1">Readonly history of past sprints.</p>
      </div>

      <!-- Loading -->
      <div v-if="isLoading" class="flex justify-center items-center h-32">
        <Spinner class="size-5 text-slate-400" />
      </div>

      <!-- Empty state -->
      <EmptyState
        v-else-if="completedSprints.length === 0"
        title="No completed sprints"
        description="Completed sprints will appear here once a sprint is marked done."
      />

      <!-- Sprint list -->
      <div v-else class="space-y-3">
        <RouterLink
          v-for="sprint in completedSprints"
          :key="sprint.id"
          :to="`/projects/${slug}/sprints/${sprint.id}`"
          class="block rounded-lg border border-slate-200 bg-white px-4 py-4 hover:border-slate-300 hover:bg-slate-50 transition-colors"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0 flex-1 space-y-1">
              <div class="flex items-center gap-2">
                <span class="text-xs font-mono text-slate-400">#{{ sprint.number }}</span>
                <span class="text-sm font-medium text-slate-900">{{ sprint.name }}</span>
                <span v-if="dateRange(sprint)" class="text-xs text-slate-400">{{ dateRange(sprint) }}</span>
              </div>
              <p v-if="sprint.goal" class="text-xs text-slate-500 truncate">{{ sprint.goal }}</p>
              <div class="pt-1">
                <ProgressBar :done="sprint.done_count" :total="sprint.issue_count" />
              </div>
              <p class="text-xs text-slate-400">{{ sprint.done_count }} / {{ sprint.issue_count }} issues done</p>
            </div>
            <ChevronRightIcon class="size-4 text-slate-400 flex-shrink-0 mt-0.5" />
          </div>
        </RouterLink>
      </div>

    </div>
  </MainLayout>
</template>
