<!--
  ProjectBacklogView — flat list of triaged issues not assigned to any sprint.

  The backlog contains all issues with triaged=true and sprint_id IS NULL.
  Issues in a sprint live on the board/sprint view. Issues with triaged=false
  live in the Triage inbox.

  Creating an issue from here sets a default status so it lands directly in
  the backlog (triaged=true) rather than the inbox.
-->
<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'
import {
  PlusIcon,
  ListIcon,
  LayersIcon,
  CircleIcon,
  CircleDotIcon,
  GitPullRequestIcon,
  CheckCircle2Icon,
  XCircleIcon,
} from 'lucide-vue-next'
import MainLayout from '@/layouts/MainLayout.vue'
import Badge from '@/components/ui/Badge.vue'
import Spinner from '@/components/ui/Spinner.vue'
import Avatar from '@/components/ui/Avatar.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import CreateIssueModal from '@/components/issue/CreateIssueModal.vue'
import { fetchProject } from '@/api/projects'
import { fetchIssues } from '@/api/issues'

const route = useRoute()
const slug = computed(() => route.params.slug)

// ── Data ─────────────────────────────────────────────────────────────────────

const { data: project, isLoading: loadingProject } = useQuery({
  queryKey: ['project', slug],
  queryFn: () => fetchProject(slug.value),
})

const { data: issuesResult, isLoading: loadingIssues } = useQuery({
  queryKey: ['issues', slug, { backlog: true, triaged: true }],
  queryFn: () => fetchIssues(slug.value, { backlog: true, triaged: true, limit: 500 }),
  enabled: computed(() => !!slug.value),
})

const isLoading = computed(() => loadingProject.value || loadingIssues.value)

const issues = computed(() => issuesResult.value?.items ?? [])

// ── Sorting ───────────────────────────────────────────────────────────────────

const PRIORITY_ORDER = { critical: 0, high: 1, medium: 2, low: 3, none: 4 }

const sortedIssues = computed(() =>
  [...issues.value].sort((a, b) => {
    const pa = PRIORITY_ORDER[a.priority] ?? 4
    const pb = PRIORITY_ORDER[b.priority] ?? 4
    if (pa !== pb) return pa - pb
    return new Date(a.created_at) - new Date(b.created_at)
  }),
)

// ── Status display ────────────────────────────────────────────────────────────

const STATUS_META = {
  todo:        { label: 'To Do',       scheme: 'gray',   icon: CircleIcon },
  in_progress: { label: 'In Progress', scheme: 'blue',   icon: CircleDotIcon },
  in_review:   { label: 'In Review',   scheme: 'violet', icon: GitPullRequestIcon },
  done:        { label: 'Done',        scheme: 'green',  icon: CheckCircle2Icon },
  cancelled:   { label: 'Cancelled',   scheme: 'gray',   icon: XCircleIcon },
  open:        { label: 'Open',        scheme: 'sky',    icon: CircleIcon },
  resolved:    { label: 'Resolved',    scheme: 'teal',   icon: CheckCircle2Icon },
  closed:      { label: 'Closed',      scheme: 'gray',   icon: XCircleIcon },
}

function statusMeta(status) {
  return STATUS_META[status] ?? { label: status, scheme: 'gray', icon: CircleIcon }
}

const PRIORITY_SCHEME = {
  none: 'gray', low: 'sky', medium: 'amber', high: 'orange', critical: 'red',
}

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

function priorityScheme(priority) {
  return PRIORITY_SCHEME[priority] ?? 'gray'
}

function estimateLabel(estimate) {
  return ESTIMATE_LABEL[estimate] ?? null
}

// ── Default status for issue creation ────────────────────────────────────────
// Passing a status causes the backend to mark the issue as triaged, landing it
// directly in the backlog instead of the inbox.

const defaultCreateStatus = computed(() => {
  if (!project.value) return null
  return project.value.archetype === 'support' ? 'open' : 'todo'
})

// ── New issue modal ───────────────────────────────────────────────────────────

const showCreateIssue = ref(false)
</script>

<template>
  <MainLayout @create-issue="showCreateIssue = true">
    <div class="flex flex-col h-full">

      <!-- ── Header ─────────────────────────────────────────────────────── -->
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

          <div class="flex items-center gap-1.5 text-slate-400">
            <ListIcon class="size-4" />
            <span class="text-sm font-medium text-slate-600">Backlog</span>
            <span
              v-if="!isLoading"
              class="text-xs text-slate-400 tabular-nums"
            >
              {{ issuesResult?.total ?? 0 }}
            </span>
          </div>
        </div>

        <button
          class="inline-flex items-center gap-1.5 rounded-md bg-blue-600 px-3 h-8 text-sm font-medium text-white hover:bg-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-1 transition-colors cursor-pointer"
          @click="showCreateIssue = true"
        >
          <PlusIcon class="size-4" />
          New issue
        </button>
      </div>

      <!-- ── Loading ────────────────────────────────────────────────────── -->
      <div v-if="isLoading" class="flex-1 flex items-center justify-center">
        <Spinner class="size-6 text-slate-400" />
      </div>

      <!-- ── Empty state ────────────────────────────────────────────────── -->
      <div
        v-else-if="!sortedIssues.length"
        class="flex-1 flex items-center justify-center"
      >
        <EmptyState
          title="Backlog is empty"
          description="Triaged issues not assigned to a sprint will appear here."
          action-label="New issue"
          @action="showCreateIssue = true"
        >
          <template #icon>
            <ListIcon class="size-8" />
          </template>
        </EmptyState>
      </div>

      <!-- ── Issue list ─────────────────────────────────────────────────── -->
      <div v-else class="flex-1 overflow-y-auto">
        <div class="divide-y divide-slate-100">
          <div
            v-for="issue in sortedIssues"
            :key="issue.id"
            class="group flex items-center gap-3 px-6 py-2.5 hover:bg-slate-50 transition-colors cursor-pointer border-l-4"
            :class="priorityBorder(issue.priority)"
          >
            <!-- Issue number + type indicator -->
            <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
              <span class="text-[11px] font-mono text-slate-400">
                {{ slug.toUpperCase() }}-{{ issue.number }}
              </span>
              <LayersIcon
                v-if="issue.type === 'epic'"
                class="size-3 text-violet-400 flex-shrink-0"
              />
            </div>

            <!-- Title -->
            <span class="flex-1 min-w-0 text-sm text-slate-800 truncate group-hover:text-slate-900">
              {{ issue.title }}
            </span>

            <!-- On hold indicator -->
            <span
              v-if="issue.on_hold"
              class="flex-shrink-0 text-[10px] font-medium bg-amber-100 text-amber-700 px-1.5 py-0.5 rounded"
            >
              on hold
            </span>

            <!-- Status -->
            <component
              :is="statusMeta(issue.status).icon"
              class="size-3.5 flex-shrink-0"
              :class="{
                'text-slate-400':  statusMeta(issue.status).scheme === 'gray',
                'text-blue-500':   statusMeta(issue.status).scheme === 'blue',
                'text-violet-500': statusMeta(issue.status).scheme === 'violet',
                'text-green-500':  statusMeta(issue.status).scheme === 'green',
                'text-sky-500':    statusMeta(issue.status).scheme === 'sky',
                'text-teal-500':   statusMeta(issue.status).scheme === 'teal',
              }"
            />
            <span class="flex-shrink-0 text-xs text-slate-500 w-20">
              {{ statusMeta(issue.status).label }}
            </span>

            <!-- Priority -->
            <Badge
              v-if="issue.priority && issue.priority !== 'none'"
              :colorScheme="priorityScheme(issue.priority)"
              compact
              class="flex-shrink-0"
            >
              {{ issue.priority }}
            </Badge>
            <span v-else class="w-14 flex-shrink-0" />

            <!-- Estimate -->
            <span
              v-if="estimateLabel(issue.estimate)"
              class="flex-shrink-0 text-[11px] font-medium text-slate-500 bg-slate-100 px-1.5 py-0.5 rounded w-7 text-center"
            >
              {{ estimateLabel(issue.estimate) }}
            </span>
            <span v-else class="w-7 flex-shrink-0" />

            <!-- Assignees -->
            <div class="flex-shrink-0 flex -space-x-1 w-10 justify-end">
              <Avatar
                v-for="a in (issue.assignees ?? []).slice(0, 2)"
                :key="a"
                :name="`${a}`"
                size="xs"
                class="ring-1 ring-white"
              />
            </div>
          </div>
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
