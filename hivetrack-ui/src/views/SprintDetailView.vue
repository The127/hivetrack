<!--
  SprintDetailView — readonly view of a completed sprint's issues.

  Shows sprint header (name, dates, goal, done/total) and issues
  grouped by status. No drag-and-drop, no inline edits.
-->
<script setup>
import { computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { priorityBorder, estimateLabel, statusLabel, statusScheme } from '@/composables/issueConstants'
import { useQuery } from '@tanstack/vue-query'
import { ArrowLeftIcon } from 'lucide-vue-next'
import MainLayout from '@/layouts/MainLayout.vue'
import Spinner from '@/components/ui/Spinner.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ProgressBar from '@/components/ui/ProgressBar.vue'
import Badge from '@/components/ui/Badge.vue'
import Avatar from '@/components/ui/Avatar.vue'
import { fetchSprints } from '@/api/sprints'
import { fetchIssues } from '@/api/issues'
import { formatDate } from '@/composables/useDate'

const route = useRoute()
const slug = computed(() => route.params.slug)
const sprintId = computed(() => route.params.sprintId)

// ── Sprint data ───────────────────────────────────────────────────────────────

const { data: sprintsResult, isLoading: sprintsLoading } = useQuery({
  queryKey: ['sprints', slug],
  queryFn: () => fetchSprints(slug.value),
  enabled: computed(() => !!slug.value),
})

const sprint = computed(() =>
  (sprintsResult.value?.sprints ?? []).find((s) => s.id === sprintId.value) ?? null
)

// ── Issues ────────────────────────────────────────────────────────────────────

const { data: issuesResult, isLoading: issuesLoading } = useQuery({
  queryKey: ['issues', slug, { sprint_id: sprintId }],
  queryFn: () => fetchIssues(slug.value, { sprint_id: sprintId.value, limit: 500 }),
  enabled: computed(() => !!slug.value && !!sprintId.value),
})

const issues = computed(() => issuesResult.value?.items ?? [])

// Group issues by status, preserving a sensible display order
const STATUS_ORDER = ['todo', 'in_progress', 'in_review', 'done', 'cancelled', 'open', 'resolved', 'closed']

const groupedIssues = computed(() => {
  const groups = {}
  for (const issue of issues.value) {
    if (!groups[issue.status]) groups[issue.status] = []
    groups[issue.status].push(issue)
  }
  return STATUS_ORDER
    .filter((s) => groups[s]?.length)
    .map((s) => ({ status: s, issues: groups[s] }))
})

// ── Formatting ────────────────────────────────────────────────────────────────


const YEAR = { year: true }

function dateRange(s) {
  if (!s) return null
  const start = formatDate(s.start_date, YEAR)
  const end = formatDate(s.end_date, YEAR)
  if (start && end) return `${start} – ${end}`
  if (start) return `Started ${start}`
  if (end) return `Ended ${end}`
  return null
}

const isLoading = computed(() => sprintsLoading.value || issuesLoading.value)
</script>

<template>
  <MainLayout>
    <div class="max-w-3xl mx-auto px-6 py-8">

      <!-- Back link -->
      <RouterLink
        :to="`/projects/${slug}/sprints`"
        class="inline-flex items-center gap-1.5 text-sm text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300 mb-5"
      >
        <ArrowLeftIcon class="size-3.5" />
        All sprints
      </RouterLink>

      <!-- Loading -->
      <div v-if="isLoading" class="flex justify-center items-center h-32">
        <Spinner class="size-5 text-slate-400" />
      </div>

      <template v-else-if="sprint">
        <!-- Sprint header -->
        <div class="mb-6 space-y-1">
          <div class="flex items-center gap-3">
            <span class="text-sm font-mono text-slate-400">#{{ sprint.number }}</span>
            <h1 class="text-lg font-semibold text-slate-900 dark:text-slate-100">{{ sprint.name }}</h1>
            <span v-if="dateRange(sprint)" class="text-sm text-slate-400">{{ dateRange(sprint) }}</span>
          </div>
          <p v-if="sprint.goal" class="text-sm text-slate-500 dark:text-slate-400">{{ sprint.goal }}</p>
          <div class="flex items-center gap-3 pt-1">
            <div class="w-48">
              <ProgressBar :done="sprint.done_count" :total="sprint.issue_count" />
            </div>
            <span class="text-xs text-slate-400">{{ sprint.done_count }} / {{ sprint.issue_count }} issues done</span>
          </div>
        </div>

        <!-- Issue list by status -->
        <div v-if="groupedIssues.length === 0">
          <EmptyState title="No issues" description="This sprint had no issues." />
        </div>

        <div v-else class="space-y-5">
          <div v-for="group in groupedIssues" :key="group.status">
            <div class="flex items-center gap-2 mb-2">
              <Badge :colorScheme="statusScheme(group.status)" compact>{{ statusLabel(group.status) }}</Badge>
              <span class="text-xs text-slate-400">{{ group.issues.length }}</span>
            </div>
            <div class="border border-slate-200 dark:border-slate-700 rounded-lg overflow-hidden">
              <RouterLink
                v-for="issue in group.issues"
                :key="issue.id"
                :to="`/projects/${slug}/issues/${issue.number}`"
                class="flex items-center gap-3 px-4 py-2.5 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors border-b border-slate-100 dark:border-slate-800 last:border-b-0 border-l-4"
                :class="priorityBorder(issue.priority)"
              >
                <span class="text-[11px] font-mono text-slate-400 flex-shrink-0 w-20">
                  {{ slug.toUpperCase() }}-{{ issue.number }}
                </span>
                <span class="flex-1 min-w-0 text-sm text-slate-800 dark:text-slate-200 truncate">{{ issue.title }}</span>
                <span
                  v-if="estimateLabel(issue.estimate)"
                  class="flex-shrink-0 text-[11px] font-medium text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-slate-700 px-1.5 py-0.5 rounded w-7 text-center"
                >{{ estimateLabel(issue.estimate) }}</span>
                <div class="flex-shrink-0 flex -space-x-1 w-10 justify-end">
                  <Avatar
                    v-for="a in (issue.assignees ?? []).slice(0, 2)"
                    :key="a.id"
                    :name="a.display_name"
                    :src="a.avatar_url"
                    size="xs"
                    class="ring-1 ring-white"
                  />
                </div>
              </RouterLink>
            </div>
          </div>
        </div>
      </template>

      <!-- Sprint not found -->
      <EmptyState v-else title="Sprint not found" description="This sprint does not exist or has been deleted." />

    </div>
  </MainLayout>
</template>
