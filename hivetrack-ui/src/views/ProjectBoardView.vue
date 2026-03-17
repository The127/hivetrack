<!--
  ProjectBoardView — Kanban board for a project.

  Shows the active sprint's issues grouped by status. If no sprint is active,
  falls back to the backlog (triaged issues with no sprint).

  A context bar below the header shows either:
  - Active sprint: name, dates, goal, issue count + "Complete sprint" action
  - No sprint: a note that the backlog is being shown + link to create a sprint
-->
<script setup>
import { ref, computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import {
  PlusIcon,
  CircleIcon,
  CircleDotIcon,
  GitPullRequestIcon,
  CheckCircle2Icon,
  XCircleIcon,
  InboxIcon,
  LayersIcon,
  CheckIcon,
  InfoIcon,
} from 'lucide-vue-next'
import MainLayout from '@/layouts/MainLayout.vue'
import Badge from '@/components/ui/Badge.vue'
import Spinner from '@/components/ui/Spinner.vue'
import Avatar from '@/components/ui/Avatar.vue'
import CreateIssueModal from '@/components/issue/CreateIssueModal.vue'
import { fetchProject } from '@/api/projects'
import { fetchIssues } from '@/api/issues'
import { fetchSprints, updateSprint } from '@/api/sprints'

const route = useRoute()
const slug = computed(() => route.params.slug)
const queryClient = useQueryClient()

// ── Data ─────────────────────────────────────────────────────────────────────

const { data: project, isLoading: loadingProject } = useQuery({
  queryKey: ['project', slug],
  queryFn: () => fetchProject(slug.value),
})

const { data: sprintsResult, isLoading: loadingSprints } = useQuery({
  queryKey: ['sprints', slug],
  queryFn: () => fetchSprints(slug.value),
  enabled: computed(() => !!slug.value),
})

const { data: issuesResult, isLoading: loadingIssues } = useQuery({
  queryKey: ['issues', slug, { triaged: true }],
  queryFn: () => fetchIssues(slug.value, { triaged: true, limit: 500 }),
  enabled: computed(() => !!slug.value),
})

const isLoading = computed(() => loadingProject.value || loadingSprints.value || loadingIssues.value)

// ── Active sprint + issue source ──────────────────────────────────────────────

const activeSprint = computed(() =>
  (sprintsResult.value?.sprints ?? []).find(s => s.status === 'active') ?? null
)

// Issues shown on the board: sprint issues when a sprint is active, otherwise backlog.
const boardIssues = computed(() => {
  const all = issuesResult.value?.items ?? []
  if (activeSprint.value) {
    return all.filter(i => i.sprint_id === activeSprint.value.id)
  }
  return all.filter(i => i.sprint_id == null)
})

// ── Complete sprint mutation ──────────────────────────────────────────────────

const { mutate: completeSprint } = useMutation({
  mutationFn: (sprintId) => updateSprint(slug.value, sprintId, { status: 'completed' }),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['sprints', slug.value] })
    queryClient.invalidateQueries({ queryKey: ['issues', slug.value] })
  },
})

// ── Status column config ──────────────────────────────────────────────────────

const SOFTWARE_COLUMNS = [
  { key: 'todo',        label: 'To Do',       scheme: 'gray',   icon: CircleIcon },
  { key: 'in_progress', label: 'In Progress', scheme: 'blue',   icon: CircleDotIcon },
  { key: 'in_review',   label: 'In Review',   scheme: 'violet', icon: GitPullRequestIcon },
  { key: 'done',        label: 'Done',        scheme: 'green',  icon: CheckCircle2Icon },
  { key: 'cancelled',   label: 'Cancelled',   scheme: 'gray',   icon: XCircleIcon },
]

const SUPPORT_COLUMNS = [
  { key: 'open',        label: 'Open',        scheme: 'sky',    icon: CircleIcon },
  { key: 'in_progress', label: 'In Progress', scheme: 'blue',   icon: CircleDotIcon },
  { key: 'resolved',    label: 'Resolved',    scheme: 'teal',   icon: CheckCircle2Icon },
  { key: 'closed',      label: 'Closed',      scheme: 'gray',   icon: XCircleIcon },
]

const columns = computed(() => {
  if (!project.value) return []
  return project.value.archetype === 'support' ? SUPPORT_COLUMNS : SOFTWARE_COLUMNS
})

// ── Group issues by status ────────────────────────────────────────────────────

const issuesByStatus = computed(() => {
  const map = {}
  for (const col of columns.value) map[col.key] = []
  for (const issue of boardIssues.value) {
    if (map[issue.status]) map[issue.status].push(issue)
  }
  return map
})

// ── Priority / estimate helpers ───────────────────────────────────────────────

const PRIORITY_BORDER = {
  none:     'border-l-slate-200',
  low:      'border-l-sky-400',
  medium:   'border-l-amber-400',
  high:     'border-l-orange-500',
  critical: 'border-l-red-500',
}

const ESTIMATE_LABEL = { none: null, xs: 'XS', s: 'S', m: 'M', l: 'L', xl: 'XL' }

function priorityBorder(priority) {
  return PRIORITY_BORDER[priority] ?? 'border-l-slate-200'
}

function estimateLabel(estimate) {
  return ESTIMATE_LABEL[estimate] ?? null
}

function formatDateRange(startDate, endDate) {
  const fmt = (d) => new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  return `${fmt(startDate)} – ${fmt(endDate)}`
}

// ── New issue modal ───────────────────────────────────────────────────────────

const showCreateIssue = ref(false)

const defaultCreateStatus = computed(() => {
  if (!project.value) return null
  return project.value.archetype === 'support' ? 'open' : 'todo'
})
</script>

<template>
  <MainLayout @create-issue="showCreateIssue = true">
    <div class="flex flex-col h-full">

      <!-- ── Board header ───────────────────────────────────────────────── -->
      <div class="flex-shrink-0 flex items-center justify-between px-6 py-3 border-b border-slate-200 bg-white">
        <div class="flex items-center gap-3 min-w-0">
          <div v-if="project" class="flex items-center gap-2 min-w-0">
            <span class="size-7 rounded flex items-center justify-center text-xs font-semibold bg-slate-100 text-slate-600 flex-shrink-0">
              {{ project.slug.slice(0, 2).toUpperCase() }}
            </span>
            <span class="font-semibold text-slate-900 truncate">{{ project.name }}</span>
            <Badge :colorScheme="project.archetype === 'software' ? 'blue' : 'teal'" compact>
              {{ project.archetype }}
            </Badge>
          </div>
          <div v-else-if="loadingProject" class="h-5 w-40 rounded bg-slate-100 animate-pulse" />
        </div>

        <button
          class="inline-flex items-center gap-1.5 rounded-md bg-blue-600 px-3 h-8 text-sm font-medium text-white hover:bg-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-1 transition-colors cursor-pointer"
          @click="showCreateIssue = true"
        >
          <PlusIcon class="size-4" />
          New issue
        </button>
      </div>

      <!-- ── Context bar ────────────────────────────────────────────────── -->
      <template v-if="!isLoading">
        <!-- Active sprint -->
        <div v-if="activeSprint" class="flex-shrink-0 flex items-center justify-between px-6 py-2 bg-blue-50 border-b border-blue-100">
          <div class="flex items-center gap-2.5 min-w-0">
            <span class="text-xs font-medium text-blue-700 bg-blue-100 px-1.5 py-0.5 rounded uppercase tracking-wide flex-shrink-0">Sprint</span>
            <span class="text-sm font-semibold text-slate-900 truncate">{{ activeSprint.name }}</span>
            <span v-if="activeSprint.start_date" class="text-xs text-slate-500 flex-shrink-0">
              {{ formatDateRange(activeSprint.start_date, activeSprint.end_date) }}
            </span>
            <span v-if="activeSprint.goal" class="text-xs text-slate-500 italic truncate max-w-64">{{ activeSprint.goal }}</span>
            <span class="text-xs text-slate-400 tabular-nums flex-shrink-0">{{ boardIssues.length }} issues</span>
          </div>
          <button
            class="inline-flex items-center gap-1.5 rounded-md border border-slate-300 bg-white px-2.5 h-7 text-xs font-medium text-slate-600 hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors cursor-pointer flex-shrink-0"
            @click="completeSprint(activeSprint.id)"
          >
            <CheckIcon class="size-3.5" />
            Complete sprint
          </button>
        </div>

        <!-- No active sprint — showing backlog -->
        <div v-else class="flex-shrink-0 flex items-center gap-2 px-6 py-2 bg-slate-50 border-b border-slate-100">
          <InfoIcon class="size-3.5 text-slate-400 flex-shrink-0" />
          <span class="text-xs text-slate-500">Showing backlog — no active sprint.</span>
          <RouterLink
            :to="`/projects/${slug}/backlog`"
            class="text-xs text-blue-600 hover:text-blue-700 hover:underline"
          >
            Go to Backlog to create and activate a sprint.
          </RouterLink>
        </div>
      </template>

      <!-- ── Loading ────────────────────────────────────────────────────── -->
      <div v-if="isLoading" class="flex-1 flex items-center justify-center">
        <Spinner class="size-6 text-slate-400" />
      </div>

      <!-- ── Board columns ──────────────────────────────────────────────── -->
      <div v-else class="flex-1 overflow-x-auto overflow-y-hidden">
        <div class="flex h-full gap-3 px-6 py-4" style="min-width: max-content">

          <div
            v-for="col in columns"
            :key="col.key"
            class="flex flex-col w-72 flex-shrink-0"
          >
            <!-- Column header -->
            <div class="flex items-center gap-2 mb-3 px-1">
              <component
                :is="col.icon"
                class="size-3.5 flex-shrink-0"
                :class="{
                  'text-slate-400':  col.scheme === 'gray',
                  'text-blue-500':   col.scheme === 'blue',
                  'text-violet-500': col.scheme === 'violet',
                  'text-green-500':  col.scheme === 'green',
                  'text-sky-500':    col.scheme === 'sky',
                  'text-teal-500':   col.scheme === 'teal',
                }"
              />
              <span class="text-sm font-medium text-slate-700">{{ col.label }}</span>
              <span class="ml-auto text-xs text-slate-400 tabular-nums">
                {{ issuesByStatus[col.key]?.length ?? 0 }}
              </span>
            </div>

            <!-- Issue cards -->
            <div class="flex-1 overflow-y-auto space-y-2 pb-4 pr-1 -mr-1">

              <!-- Empty column -->
              <div
                v-if="!issuesByStatus[col.key]?.length"
                class="rounded-lg border-2 border-dashed border-slate-200 py-8 text-center"
              >
                <p class="text-xs text-slate-400">No issues</p>
              </div>

              <!-- Issue card -->
              <div
                v-for="issue in issuesByStatus[col.key]"
                :key="issue.id"
                class="group rounded-lg border border-slate-200 bg-white px-3 py-2.5 shadow-sm hover:shadow-md hover:border-slate-300 transition-all cursor-pointer border-l-4"
                :class="priorityBorder(issue.priority)"
              >
                <!-- Issue number + type -->
                <div class="flex items-center gap-1.5 mb-1.5">
                  <span class="text-[11px] font-mono text-slate-400">
                    {{ slug.toUpperCase() }}-{{ issue.number }}
                  </span>
                  <LayersIcon v-if="issue.type === 'epic'" class="size-3 text-violet-400 ml-auto" />
                  <span
                    v-if="issue.on_hold"
                    class="ml-auto text-[10px] font-medium bg-amber-100 text-amber-700 px-1.5 py-0.5 rounded"
                  >
                    on hold
                  </span>
                </div>

                <!-- Title -->
                <p class="text-sm text-slate-800 leading-snug line-clamp-2 group-hover:text-slate-900 mb-2">
                  {{ issue.title }}
                </p>

                <!-- Footer: estimate + assignees -->
                <div class="flex items-center gap-1.5">
                  <span
                    v-if="estimateLabel(issue.estimate)"
                    class="text-[11px] font-medium text-slate-500 bg-slate-100 px-1.5 py-0.5 rounded"
                  >
                    {{ estimateLabel(issue.estimate) }}
                  </span>
                  <span class="flex-1" />
                  <div
                    v-if="issue.assignees?.length"
                    class="flex items-center gap-1 text-xs text-slate-400"
                  >
                    <Avatar :name="`${issue.assignees.length}`" size="xs" />
                    <span v-if="issue.assignees.length > 1" class="text-[11px]">
                      +{{ issue.assignees.length - 1 }}
                    </span>
                  </div>
                </div>
              </div>

            </div>
          </div>

        </div>
      </div>

      <!-- ── Empty board ────────────────────────────────────────────────── -->
      <div
        v-if="!isLoading && boardIssues.length === 0"
        class="absolute inset-0 flex items-center justify-center pointer-events-none"
      >
        <div class="text-center">
          <InboxIcon class="size-10 text-slate-300 mx-auto mb-3" />
          <p class="text-sm font-medium text-slate-600">
            {{ activeSprint ? 'Sprint is empty' : 'Backlog is empty' }}
          </p>
          <p class="text-sm text-slate-400 mt-1">
            {{ activeSprint ? 'Add issues to this sprint from the Backlog.' : 'Create an issue or triage items from the inbox.' }}
          </p>
        </div>
      </div>

    </div>

    <!-- ── Create issue modal ──────────────────────────────────────────── -->
    <CreateIssueModal
      :open="showCreateIssue"
      :project-slug="slug"
      :default-status="defaultCreateStatus"
      @close="showCreateIssue = false"
      @created="showCreateIssue = false"
    />
  </MainLayout>
</template>
