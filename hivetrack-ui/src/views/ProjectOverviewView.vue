<!--
  ProjectOverviewView — at-a-glance summary for a project.

  Shows project metadata, active sprint status, issue counts by status,
  and triage inbox count. All data is composed from existing queries —
  no dedicated overview endpoint needed.
-->
<script setup>
import { computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { isTerminalStatus, statusColumns } from '@/composables/issueConstants'
import { formatDateRange } from '@/composables/useDate'
import { useQuery } from '@tanstack/vue-query'
import {
  KanbanIcon,
  ListIcon,
  InboxIcon,
  LayersIcon,
  FlagIcon,
  CodeIcon,
  HeadphonesIcon,
  CalendarIcon,
} from 'lucide-vue-next'
import MainLayout from '@/layouts/MainLayout.vue'
import Badge from '@/components/ui/Badge.vue'
import Spinner from '@/components/ui/Spinner.vue'
import ProgressBar from '@/components/ui/ProgressBar.vue'
import SprintBurndownChart from '@/components/sprint/SprintBurndownChart.vue'
import OverviewMembers from '@/components/overview/OverviewMembers.vue'
import OverviewLabels from '@/components/overview/OverviewLabels.vue'
import OverviewWipSettings from '@/components/overview/OverviewWipSettings.vue'
import OverviewDrones from '@/components/overview/OverviewDrones.vue'
import { fetchProject } from '@/api/projects'
import { fetchIssues } from '@/api/issues'
import { fetchSprints, fetchSprintBurndown } from '@/api/sprints'

const route = useRoute()
const slug = computed(() => route.params.slug)

// ── Queries ───────────────────────────────────────────────────────────────────

const { data: project, isLoading: loadingProject } = useQuery({
  queryKey: ['project', slug],
  queryFn: () => fetchProject(slug.value),
})

const { data: sprintsResult } = useQuery({
  queryKey: ['sprints', slug],
  queryFn: () => fetchSprints(slug.value),
  enabled: computed(() => !!slug.value),
})

const { data: issuesResult } = useQuery({
  queryKey: ['issues', slug, { triaged: true, type: 'task' }],
  queryFn: () => fetchIssues(slug.value, { triaged: true, type: 'task', limit: 500 }),
  enabled: computed(() => !!slug.value),
})

const { data: inboxResult } = useQuery({
  queryKey: ['issues', slug, { triaged: false }],
  queryFn: () => fetchIssues(slug.value, { triaged: false, limit: 1 }),
  enabled: computed(() => !!slug.value),
})

// ── Derived state ─────────────────────────────────────────────────────────────

const activeSprint = computed(
  () => (sprintsResult.value?.sprints ?? []).find((s) => s.status === 'active') ?? null,
)

const { data: burndownResult } = useQuery({
  queryKey: computed(() => ['sprint-burndown', slug.value, activeSprint.value?.id]),
  queryFn: () => fetchSprintBurndown(slug.value, activeSprint.value.id),
  enabled: computed(() => !!activeSprint.value),
})

const allIssues = computed(() => issuesResult.value?.items ?? [])
const inboxCount = computed(() => inboxResult.value?.total ?? 0)

const statusDefs = computed(() =>
  statusColumns(project.value?.archetype),
)

const statusCounts = computed(() => {
  const counts = {}
  for (const s of statusDefs.value) counts[s.key] = 0
  for (const issue of allIssues.value) {
    if (counts[issue.status] !== undefined) counts[issue.status]++
  }
  return counts
})

const sprintIssues = computed(() => {
  if (!activeSprint.value) return []
  return allIssues.value.filter((i) => i.sprint_id === activeSprint.value.id)
})

const sprintDone = computed(() => {
  if (!activeSprint.value) return 0
  const arch = project.value?.archetype ?? 'software'
  return sprintIssues.value.filter((i) => isTerminalStatus(i.status, arch)).length
})

const SCHEME_ICON_CLASS = {
  gray: 'text-slate-400',
  blue: 'text-blue-500',
  violet: 'text-violet-500',
  green: 'text-green-500',
  sky: 'text-sky-500',
  teal: 'text-teal-500',
}

</script>

<template>
  <MainLayout>
    <!-- Loading -->
    <div v-if="loadingProject" class="h-32 flex items-center justify-center">
      <Spinner class="size-5 text-slate-400" />
    </div>

    <div v-else-if="project" class="max-w-3xl mx-auto px-6 py-8 space-y-8">

      <!-- ── Project header ──────────────────────────────────────────────────── -->
      <div>
        <div class="flex items-start gap-4">
          <span
            class="size-12 rounded-lg flex items-center justify-center text-base font-bold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 flex-shrink-0"
          >
            {{ project.slug.slice(0, 2).toUpperCase() }}
          </span>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 flex-wrap">
              <h1 class="text-xl font-semibold text-slate-900 dark:text-slate-100">{{ project.name }}</h1>
              <Badge :colorScheme="project.archetype === 'software' ? 'blue' : 'teal'" compact>
                <component
                  :is="project.archetype === 'software' ? CodeIcon : HeadphonesIcon"
                  class="size-3 mr-0.5"
                />
                {{ project.archetype }}
              </Badge>
            </div>
            <p v-if="project.description" class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
              {{ project.description }}
            </p>
            <p class="text-xs text-slate-400 dark:text-slate-500 mt-1 font-mono">{{ project.slug }}</p>
          </div>
        </div>
      </div>

      <!-- ── Quick navigation ────────────────────────────────────────────────── -->
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <RouterLink
          :to="`/projects/${slug}/board`"
          class="flex items-center gap-2.5 rounded-lg border border-slate-200 dark:border-slate-700 px-4 py-3 hover:bg-slate-50 dark:hover:bg-slate-800 hover:border-slate-300 dark:hover:border-slate-600 transition-colors group"
        >
          <KanbanIcon class="size-4 text-slate-500 dark:text-slate-400 group-hover:text-slate-700 dark:group-hover:text-slate-200 flex-shrink-0" />
          <span class="text-sm font-medium text-slate-700 dark:text-slate-300 group-hover:text-slate-900 dark:group-hover:text-slate-100">Board</span>
        </RouterLink>
        <RouterLink
          :to="`/projects/${slug}/backlog`"
          class="flex items-center gap-2.5 rounded-lg border border-slate-200 dark:border-slate-700 px-4 py-3 hover:bg-slate-50 dark:hover:bg-slate-800 hover:border-slate-300 dark:hover:border-slate-600 transition-colors group"
        >
          <ListIcon class="size-4 text-slate-500 dark:text-slate-400 group-hover:text-slate-700 dark:group-hover:text-slate-200 flex-shrink-0" />
          <span class="text-sm font-medium text-slate-700 dark:text-slate-300 group-hover:text-slate-900 dark:group-hover:text-slate-100">Backlog</span>
        </RouterLink>
        <RouterLink
          :to="`/projects/${slug}/epics`"
          class="flex items-center gap-2.5 rounded-lg border border-slate-200 dark:border-slate-700 px-4 py-3 hover:bg-slate-50 dark:hover:bg-slate-800 hover:border-slate-300 dark:hover:border-slate-600 transition-colors group"
        >
          <LayersIcon class="size-4 text-slate-500 dark:text-slate-400 group-hover:text-slate-700 dark:group-hover:text-slate-200 flex-shrink-0" />
          <span class="text-sm font-medium text-slate-700 dark:text-slate-300 group-hover:text-slate-900 dark:group-hover:text-slate-100">Epics</span>
        </RouterLink>
        <RouterLink
          :to="`/projects/${slug}/milestones`"
          class="flex items-center gap-2.5 rounded-lg border border-slate-200 dark:border-slate-700 px-4 py-3 hover:bg-slate-50 dark:hover:bg-slate-800 hover:border-slate-300 dark:hover:border-slate-600 transition-colors group"
        >
          <FlagIcon class="size-4 text-slate-500 dark:text-slate-400 group-hover:text-slate-700 dark:group-hover:text-slate-200 flex-shrink-0" />
          <span class="text-sm font-medium text-slate-700 dark:text-slate-300 group-hover:text-slate-900 dark:group-hover:text-slate-100">Milestones</span>
        </RouterLink>
      </div>

      <!-- ── Active sprint ────────────────────────────────────────────────────── -->
      <section v-if="activeSprint">
        <h2 class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-3">Active Sprint</h2>
        <div class="rounded-lg border border-blue-100 dark:border-blue-800/50 bg-blue-50 dark:bg-blue-900/20 px-4 py-3">
          <div class="flex items-center justify-between gap-3">
            <div class="flex items-center gap-2 min-w-0">
              <span
                class="text-xs font-medium text-blue-700 dark:text-blue-300 bg-blue-100 dark:bg-blue-800/50 px-1.5 py-0.5 rounded uppercase tracking-wide flex-shrink-0"
              >Sprint</span>
              <span class="text-sm font-semibold text-slate-900 dark:text-slate-100 truncate">{{ activeSprint.name }}</span>
              <span v-if="activeSprint.start_date" class="text-xs text-slate-500 dark:text-slate-400 flex-shrink-0">
                <CalendarIcon class="size-3 inline mr-0.5 -mt-px" />
                {{ formatDateRange(activeSprint.start_date, activeSprint.end_date) }}
              </span>
            </div>
            <div class="w-32 flex-shrink-0">
              <ProgressBar :done="sprintDone" :total="sprintIssues.length" />
            </div>
          </div>
          <p v-if="activeSprint.goal" class="text-xs text-slate-600 dark:text-slate-400 mt-2">
            <span class="text-slate-400 dark:text-slate-500 font-medium">Goal:</span> {{ activeSprint.goal }}
          </p>
          <div v-if="burndownResult && activeSprint.start_date && activeSprint.end_date" class="mt-3">
            <SprintBurndownChart
              :points="burndownResult.points"
              :total="burndownResult.total"
              :start-date="activeSprint.start_date"
              :end-date="activeSprint.end_date"
            />
          </div>
        </div>
      </section>

      <!-- ── Issue stats ─────────────────────────────────────────────────────── -->
      <section>
        <h2 class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-3">Issues</h2>
        <div class="rounded-lg border border-slate-200 dark:border-slate-700 divide-y divide-slate-100 dark:divide-slate-800 overflow-hidden">
          <div
            v-for="s in statusDefs"
            :key="s.key"
            class="flex items-center gap-3 px-4 py-2.5"
          >
            <component
              :is="s.icon"
              class="size-4 flex-shrink-0"
              :class="SCHEME_ICON_CLASS[s.scheme]"
            />
            <span class="text-sm text-slate-700 dark:text-slate-300 flex-1">{{ s.label }}</span>
            <span class="text-sm font-medium text-slate-900 dark:text-slate-100 tabular-nums">
              {{ statusCounts[s.key] }}
            </span>
          </div>
          <!-- Inbox row -->
          <RouterLink
            :to="`/projects/${slug}/triage`"
            class="flex items-center gap-3 px-4 py-2.5 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors group"
          >
            <InboxIcon class="size-4 flex-shrink-0 text-amber-500" />
            <span class="text-sm text-slate-700 dark:text-slate-300 flex-1 group-hover:text-slate-900 dark:group-hover:text-slate-100">Inbox (untriaged)</span>
            <span
              class="text-sm font-medium tabular-nums"
              :class="inboxCount > 0 ? 'text-amber-600 dark:text-amber-400' : 'text-slate-900 dark:text-slate-100'"
            >
              {{ inboxCount }}
            </span>
          </RouterLink>
        </div>
      </section>

      <!-- ── Extracted sections ──────────────────────────────────────────────── -->
      <OverviewMembers :slug="slug" />
      <OverviewLabels :slug="slug" />
      <OverviewWipSettings :slug="slug" />
      <OverviewDrones :slug="slug" />

    </div>
  </MainLayout>
</template>
