<!--
  ProjectBoardView — Kanban board for a project.

  Shows triaged issues grouped by status as swimlane columns.
  Columns are determined by the project archetype:
    software: todo | in_progress | in_review | done | cancelled
    support:  open | in_progress | resolved  | closed

  Issues in the triage inbox (triaged=false) are not shown here — they
  live in the Triage view.
-->
<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useQuery } from '@tanstack/vue-query'
import { PlusIcon, CircleIcon, CircleDotIcon, GitPullRequestIcon, CheckCircle2Icon, XCircleIcon, InboxIcon, LayersIcon } from 'lucide-vue-next'
import MainLayout from '@/layouts/MainLayout.vue'
import Badge from '@/components/ui/Badge.vue'
import Spinner from '@/components/ui/Spinner.vue'
import Avatar from '@/components/ui/Avatar.vue'
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
  queryKey: ['issues', slug, { triaged: true }],
  queryFn: () => fetchIssues(slug.value, { triaged: true, limit: 500 }),
  enabled: computed(() => !!slug.value),
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
  const issues = issuesResult.value?.items ?? []
  const map = {}
  for (const col of columns.value) map[col.key] = []
  for (const issue of issues) {
    if (map[issue.status]) map[issue.status].push(issue)
  }
  return map
})

// ── Priority config ───────────────────────────────────────────────────────────

const PRIORITY_BORDER = {
  none:     'border-l-slate-200',
  low:      'border-l-sky-400',
  medium:   'border-l-amber-400',
  high:     'border-l-orange-500',
  critical: 'border-l-red-500',
}

const ESTIMATE_LABEL = {
  none: null,
  xs:   'XS',
  s:    'S',
  m:    'M',
  l:    'L',
  xl:   'XL',
}

function priorityBorder(priority) {
  return PRIORITY_BORDER[priority] ?? 'border-l-slate-200'
}

function estimateLabel(estimate) {
  return ESTIMATE_LABEL[estimate] ?? null
}

const isLoading = computed(() => loadingProject.value || loadingIssues.value)
</script>

<template>
  <MainLayout>
    <div class="flex flex-col h-full">

      <!-- ── Board header ───────────────────────────────────────────────── -->
      <div class="flex-shrink-0 flex items-center justify-between px-6 py-3 border-b border-slate-200 bg-white">
        <div class="flex items-center gap-3 min-w-0">
          <div v-if="project" class="flex items-center gap-2 min-w-0">
            <span
              class="size-7 rounded flex items-center justify-center text-xs font-semibold bg-slate-100 text-slate-600 flex-shrink-0"
            >
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
        >
          <PlusIcon class="size-4" />
          New issue
        </button>
      </div>

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
                  <!-- Assignee count (UUIDs only — no names available in summary) -->
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
        v-if="!isLoading && issuesResult?.total === 0"
        class="absolute inset-0 flex items-center justify-center pointer-events-none"
      >
        <div class="text-center">
          <InboxIcon class="size-10 text-slate-300 mx-auto mb-3" />
          <p class="text-sm font-medium text-slate-600">Board is empty</p>
          <p class="text-sm text-slate-400 mt-1">
            Create an issue or triage items from the inbox.
          </p>
        </div>
      </div>

    </div>
  </MainLayout>
</template>
