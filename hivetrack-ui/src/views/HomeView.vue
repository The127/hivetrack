<!--
  HomeView — Personal dashboard ("My Work").

  The default view after login. Answers: what should I work on right now?

  Sections (each backed by its own TanStack Query):
    1. My open issues    — /api/v1/me/issues (list or kanban board toggle)
    2. Projects          — /api/v1/projects

  Keyboard shortcuts (from MainLayout):
    C  → create new issue
-->
<script setup>
import { ref, computed, watch } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'
import {
  PlusIcon,
  InboxIcon,
  FolderKanbanIcon,
  CircleDotIcon,
  CircleIcon,
  GitPullRequestIcon,
  CheckCircle2Icon,
  XCircleIcon,
  ListIcon,
  KanbanIcon,
} from 'lucide-vue-next'
import { VueDraggable } from 'vue-draggable-plus'
import { generateKeyBetween } from 'fractional-indexing'
import MainLayout from '@/layouts/MainLayout.vue'
import AssigneePopover from '@/components/issue/AssigneePopover.vue'
import Badge from '@/components/ui/Badge.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import Spinner from '@/components/ui/Spinner.vue'
import CreateProjectModal from '@/components/project/CreateProjectModal.vue'
import CreateIssueModal from '@/components/issue/CreateIssueModal.vue'
import PrioritySelect from '@/components/issue/PrioritySelect.vue'
import { apiFetch } from '@/composables/useApi'
import { updateIssue } from '@/api/issues'
import { useAuth } from '@/composables/useAuth'

const { user } = useAuth()
const router = useRouter()
const queryClient = useQueryClient()

const showCreateProject = ref(false)
const showCreateIssue = ref(false)

function onProjectCreated(result) {
  showCreateProject.value = false
  router.push(`/projects/${result.slug}/board`)
}

function onIssueCreated() {
  showCreateIssue.value = false
}

const userName = user.value?.profile?.name ?? user.value?.profile?.email ?? 'You'

// ── View mode ─────────────────────────────────────────────────────────────────

const viewMode = ref('list') // 'list' | 'board'

// ── Queries ───────────────────────────────────────────────────────────────────

const { data: myIssues, isLoading: loadingIssues } = useQuery({
  queryKey: ['me', 'issues'],
  queryFn: () => apiFetch('/api/v1/me/issues'),
})

const { data: projects, isLoading: loadingProjects } = useQuery({
  queryKey: ['projects'],
  queryFn: () => apiFetch('/api/v1/projects'),
})

// ── Status display helpers ────────────────────────────────────────────────────

const STATUS_SCHEME = {
  todo: 'gray',
  in_progress: 'blue',
  in_review: 'violet',
  done: 'green',
  cancelled: 'gray',
  open: 'sky',
  resolved: 'teal',
  closed: 'gray',
}

function statusScheme(status) {
  return STATUS_SCHEME[status] ?? 'gray'
}

// Format status string for display: "in_progress" → "In progress"
function formatStatus(s) {
  return s.replace(/_/g, ' ').replace(/^\w/, (c) => c.toUpperCase())
}

// ── Priority mutation (list view) ─────────────────────────────────────────────

const { mutate: updateMyIssuePriority } = useMutation({
  mutationFn: ({ projectSlug, number, priority }) => updateIssue(projectSlug, number, { priority }),
  onMutate: async ({ number, priority }) => {
    const key = ['me', 'issues']
    await queryClient.cancelQueries({ queryKey: key })
    const previous = queryClient.getQueryData(key)
    queryClient.setQueryData(key, old => {
      if (!old) return old
      return { ...old, items: old.items.map(i => i.number === number ? { ...i, priority } : i) }
    })
    return { previous }
  },
  onError: (_err, _vars, context) => {
    if (context?.previous) {
      queryClient.setQueryData(['me', 'issues'], context.previous)
    }
  },
  onSettled: () => {
    queryClient.invalidateQueries({ queryKey: ['me', 'issues'] })
  },
})

// ── Board columns config ──────────────────────────────────────────────────────

const ALL_COLUMNS = [
  { key: 'todo',        label: 'To Do',      scheme: 'gray',   icon: CircleIcon },
  { key: 'open',        label: 'Open',        scheme: 'sky',    icon: CircleIcon },
  { key: 'in_progress', label: 'In Progress', scheme: 'blue',   icon: CircleDotIcon },
  { key: 'in_review',   label: 'In Review',   scheme: 'violet', icon: GitPullRequestIcon },
  { key: 'done',        label: 'Done',        scheme: 'green',  icon: CheckCircle2Icon },
  { key: 'resolved',    label: 'Resolved',    scheme: 'teal',   icon: CheckCircle2Icon },
  { key: 'closed',      label: 'Closed',      scheme: 'gray',   icon: XCircleIcon },
  { key: 'cancelled',   label: 'Cancelled',   scheme: 'gray',   icon: XCircleIcon },
]

const activeColumns = computed(() => {
  const usedStatuses = new Set((myIssues.value?.items ?? []).map(i => i.status))
  return ALL_COLUMNS.filter(c => usedStatuses.has(c.key))
})

// ── Drag-and-drop (board view) ────────────────────────────────────────────────

const isDragging = ref(false)
const columnIssues = ref({})

function rebuildColumnIssues() {
  const newMap = {}
  for (const col of activeColumns.value) {
    newMap[col.key] = (myIssues.value?.items ?? [])
      .filter(i => i.status === col.key)
      .slice()
  }
  columnIssues.value = newMap
}

watch(
  [myIssues, activeColumns],
  () => {
    if (!isDragging.value) rebuildColumnIssues()
  },
  { immediate: true },
)

function computeRank(items, newIdx) {
  const prev = newIdx > 0 ? (items[newIdx - 1]?.rank ?? null) : null
  const next = newIdx < items.length - 1 ? (items[newIdx + 1]?.rank ?? null) : null
  try {
    return generateKeyBetween(prev, next)
  } catch {
    return Date.now().toString(36) + Math.random().toString(36).slice(2, 6)
  }
}

const { mutate: reorderIssue } = useMutation({
  mutationFn: ({ projectSlug, issueNumber, data }) =>
    updateIssue(projectSlug, issueNumber, data),
  onMutate: async ({ issueNumber, data }) => {
    const key = ['me', 'issues']
    await queryClient.cancelQueries({ queryKey: key })
    const previous = queryClient.getQueryData(key)
    queryClient.setQueryData(key, old => {
      if (!old) return old
      return {
        ...old,
        items: old.items.map(i => i.number === issueNumber ? { ...i, ...data } : i),
      }
    })
    return { previous }
  },
  onError: (_err, _vars, context) => {
    if (context?.previous) {
      queryClient.setQueryData(['me', 'issues'], context.previous)
    }
  },
  onSettled: () => {
    isDragging.value = false
    queryClient.invalidateQueries({ queryKey: ['me', 'issues'] })
  },
})

function onDragStart() {
  isDragging.value = true
}

function onDragEnd() {
  setTimeout(() => {
    isDragging.value = false
  }, 0)
}

function onWithinColumnDrag(evt, colKey) {
  const items = columnIssues.value[colKey]
  const newIdx = evt.newDraggableIndex
  const movedItem = items[newIdx]
  const newRank = computeRank(items, newIdx)
  movedItem.rank = newRank
  reorderIssue({ projectSlug: movedItem.project_slug, issueNumber: movedItem.number, data: { rank: newRank } })
}

function onCrossColumnDrop(evt, toColKey) {
  const items = columnIssues.value[toColKey]
  const newIdx = evt.newDraggableIndex
  const movedItem = items[newIdx]
  const newRank = computeRank(items, newIdx)
  movedItem.rank = newRank
  movedItem.status = toColKey
  reorderIssue({
    projectSlug: movedItem.project_slug,
    issueNumber: movedItem.number,
    data: { rank: newRank, status: toColKey },
  })
}

const PRIORITY_BORDER = {
  none: 'border-l-slate-200',
  low: 'border-l-sky-400',
  medium: 'border-l-amber-400',
  high: 'border-l-orange-500',
  critical: 'border-l-red-500',
}

function priorityBorder(priority) {
  return PRIORITY_BORDER[priority] ?? 'border-l-slate-200'
}
</script>

<template>
  <MainLayout @create-issue="showCreateIssue = true">
    <div class="px-6 py-8">
      <!-- Page header -->
      <div class="max-w-3xl mx-auto mb-8 flex items-start justify-between">
        <div>
          <h1 class="text-xl font-semibold text-slate-900 dark:text-slate-100">My Work</h1>
          <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            Welcome back, {{ userName }}
          </p>
        </div>
        <button
          class="inline-flex items-center gap-1.5 rounded-md bg-blue-600 px-3 h-8 text-sm font-medium text-white hover:bg-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-1 transition-colors cursor-pointer"
          @click="showCreateIssue = true"
        >
          <PlusIcon class="size-4" />
          New issue
        </button>
      </div>

      <!-- ── My open issues ────────────────────────────────────────────── -->
      <section class="mb-8">
        <!-- Section header with view toggle -->
        <div class="max-w-3xl mx-auto flex items-center gap-3 mb-3">
          <h2 class="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center gap-2">
            <CircleDotIcon class="size-4 text-blue-500" />
            My open issues
            <span
              v-if="myIssues?.items?.length"
              class="text-xs font-normal text-slate-500"
            >
              {{ myIssues.items.length }}
            </span>
          </h2>

          <!-- List / Board toggle -->
          <div class="ml-auto flex items-center rounded-md border border-slate-200 dark:border-slate-700 overflow-hidden">
            <button
              class="inline-flex items-center gap-1.5 px-2.5 h-7 text-xs font-medium transition-colors cursor-pointer"
              :class="viewMode === 'list'
                ? 'bg-slate-100 dark:bg-slate-800 text-slate-800 dark:text-slate-200'
                : 'bg-white dark:bg-slate-900 text-slate-500 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 hover:text-slate-700 dark:hover:text-slate-300'"
              @click="viewMode = 'list'"
            >
              <ListIcon class="size-3.5" />
              List
            </button>
            <button
              class="inline-flex items-center gap-1.5 px-2.5 h-7 text-xs font-medium border-l border-slate-200 dark:border-slate-700 transition-colors cursor-pointer"
              :class="viewMode === 'board'
                ? 'bg-slate-100 dark:bg-slate-800 text-slate-800 dark:text-slate-200'
                : 'bg-white dark:bg-slate-900 text-slate-500 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 hover:text-slate-700 dark:hover:text-slate-300'"
              @click="viewMode = 'board'"
            >
              <KanbanIcon class="size-3.5" />
              Board
            </button>
          </div>
        </div>

        <div v-if="loadingIssues" class="flex justify-center py-8">
          <Spinner class="size-5 text-slate-400" />
        </div>

        <template v-else-if="myIssues?.items?.length">
          <!-- ── List view ─────────────────────────────────────────────── -->
          <div
            v-if="viewMode === 'list'"
            class="max-w-3xl mx-auto rounded-lg border border-slate-200 dark:border-slate-700 divide-y divide-slate-100 dark:divide-slate-800 overflow-hidden"
          >
            <div
              v-for="issue in myIssues.items"
              :key="issue.id"
              class="flex items-center gap-3 px-4 py-2.5 bg-white dark:bg-slate-900 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors cursor-pointer group"
            >
              <!-- Issue number -->
              <span class="text-xs font-mono text-slate-400 dark:text-slate-500 flex-shrink-0 w-14 text-right">
                {{ issue.project_slug?.toUpperCase() }}-{{ issue.number }}
              </span>

              <!-- Title -->
              <span class="flex-1 min-w-0 text-sm text-slate-800 dark:text-slate-200 truncate group-hover:text-slate-900 dark:group-hover:text-slate-100">
                {{ issue.title }}
              </span>

              <!-- Priority -->
              <PrioritySelect
                :priority="issue.priority ?? 'none'"
                @update:priority="updateMyIssuePriority({ projectSlug: issue.project_slug, number: issue.number, priority: $event })"
              />

              <!-- Status -->
              <Badge :colorScheme="statusScheme(issue.status)" compact>
                {{ formatStatus(issue.status) }}
              </Badge>

              <!-- Assignees -->
              <AssigneePopover :assignees="issue.assignees ?? []" />
            </div>
          </div>

          <!-- ── Board view ─────────────────────────────────────────────── -->
          <div
            v-else
            class="overflow-x-auto"
          >
            <div class="flex gap-3 pb-4" style="min-width: max-content">
              <div
                v-for="col in activeColumns"
                :key="col.key"
                class="flex flex-col w-72 flex-shrink-0"
              >
                <!-- Column header -->
                <div class="flex items-center gap-2 mb-3 px-1">
                  <component
                    :is="col.icon"
                    class="size-3.5 flex-shrink-0"
                    :class="{
                      'text-slate-400': col.scheme === 'gray',
                      'text-blue-500': col.scheme === 'blue',
                      'text-violet-500': col.scheme === 'violet',
                      'text-green-500': col.scheme === 'green',
                      'text-sky-500': col.scheme === 'sky',
                      'text-teal-500': col.scheme === 'teal',
                    }"
                  />
                  <span class="text-sm font-medium text-slate-700 dark:text-slate-300">{{ col.label }}</span>
                  <span class="ml-auto text-xs text-slate-400 dark:text-slate-500 tabular-nums">
                    {{ columnIssues[col.key]?.length ?? 0 }}
                  </span>
                </div>

                <!-- Draggable issue cards -->
                <div class="relative">
                  <VueDraggable
                    v-model="columnIssues[col.key]"
                    :group="{ name: 'my-work-board' }"
                    :animation="150"
                    ghost-class="opacity-30"
                    class="space-y-2 min-h-16"
                    @start="onDragStart"
                    @end="onDragEnd"
                    @update="(evt) => onWithinColumnDrag(evt, col.key)"
                    @add="(evt) => onCrossColumnDrop(evt, col.key)"
                  >
                    <div
                      v-for="issue in columnIssues[col.key]"
                      :key="issue.id"
                      class="group rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 px-3 py-2.5 shadow-sm hover:shadow-md hover:border-slate-300 dark:hover:border-slate-600 transition-all cursor-grab active:cursor-grabbing border-l-4"
                      :class="priorityBorder(issue.priority)"
                    >
                      <!-- Project badge + assignees -->
                      <div class="flex items-center gap-1.5 mb-1.5">
                        <span class="text-[11px] font-mono text-slate-400 dark:text-slate-500">
                          {{ issue.project_slug?.toUpperCase() }}-{{ issue.number }}
                        </span>
                        <span
                          v-if="issue.on_hold"
                          class="text-[10px] font-medium bg-amber-100 text-amber-700 px-1.5 py-0.5 rounded"
                        >
                          on hold
                        </span>
                        <span class="flex-1" />
                        <AssigneePopover :assignees="issue.assignees ?? []" />
                      </div>

                      <!-- Title -->
                      <p class="text-sm text-slate-800 dark:text-slate-200 leading-snug line-clamp-2 group-hover:text-slate-900 dark:group-hover:text-slate-100 mb-2">
                        {{ issue.title }}
                      </p>

                      <!-- Priority row -->
                      <div class="flex items-center gap-1.5">
                        <span class="flex-1" />
                        <PrioritySelect
                          :priority="issue.priority ?? 'none'"
                          @update:priority="reorderIssue({ projectSlug: issue.project_slug, issueNumber: issue.number, data: { priority: $event } })"
                        />
                      </div>
                    </div>
                  </VueDraggable>

                  <!-- Empty column placeholder -->
                  <div
                    v-if="!columnIssues[col.key]?.length && !isDragging"
                    class="absolute inset-0 rounded-lg border-2 border-dashed border-slate-200 dark:border-slate-700 flex items-center justify-center min-h-16"
                  >
                    <p class="text-xs text-slate-400">No issues</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </template>

        <EmptyState
          v-else
          class="max-w-3xl mx-auto"
          title="No open issues assigned to you"
          description="Issues assigned to you across all projects will appear here."
        >
          <template #icon>
            <CircleDotIcon class="size-8" />
          </template>
        </EmptyState>
      </section>

      <!-- ── Projects ──────────────────────────────────────────────────── -->
      <section class="max-w-3xl mx-auto">
        <div class="flex items-center justify-between mb-3">
          <h2 class="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center gap-2">
            <FolderKanbanIcon class="size-4 text-slate-500 dark:text-slate-400" />
            Projects
            <span
              v-if="projects?.items?.length"
              class="text-xs font-normal text-slate-500"
            >
              {{ projects.items.length }}
            </span>
          </h2>
          <button
            class="inline-flex items-center gap-1.5 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 px-2.5 h-7 text-xs font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700 hover:text-slate-800 dark:hover:text-slate-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-1 transition-colors cursor-pointer"
            @click="showCreateProject = true"
          >
            <PlusIcon class="size-3.5" />
            New project
          </button>
        </div>

        <div v-if="loadingProjects" class="flex justify-center py-8">
          <Spinner class="size-5 text-slate-400" />
        </div>

        <div
          v-else-if="projects?.items?.length"
          class="rounded-lg border border-slate-200 dark:border-slate-700 divide-y divide-slate-100 dark:divide-slate-800 overflow-hidden"
        >
          <RouterLink
            v-for="project in projects.items"
            :key="project.id"
            :to="`/projects/${project.slug}/board`"
            class="flex items-center gap-3 px-4 py-3 bg-white dark:bg-slate-900 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors group"
          >
            <!-- Project initial -->
            <span
              class="size-7 rounded flex items-center justify-center text-xs font-semibold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 flex-shrink-0"
            >
              {{ project.slug.slice(0, 2).toUpperCase() }}
            </span>

            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-slate-800 dark:text-slate-200 group-hover:text-slate-900 dark:group-hover:text-slate-100 truncate">
                {{ project.name }}
              </p>
              <p class="text-xs text-slate-500 dark:text-slate-400">{{ project.slug }}</p>
            </div>

            <Badge :colorScheme="project.archetype === 'software' ? 'blue' : 'teal'" compact>
              {{ project.archetype }}
            </Badge>

            <span v-if="project.archived" class="text-xs text-slate-400 italic">archived</span>
          </RouterLink>
        </div>

        <EmptyState
          v-else
          title="No projects yet"
          description="Create your first project to start tracking work."
          action-label="New project"
          @action="showCreateProject = true"
        >
          <template #icon>
            <FolderKanbanIcon class="size-8" />
          </template>
        </EmptyState>
      </section>

      <!-- ── Triage inbox hint ──────────────────────────────────────────── -->
      <div
        v-if="projects?.items?.length"
        class="mt-6 max-w-3xl mx-auto rounded-lg border border-amber-200 dark:border-amber-700/50 bg-amber-50 dark:bg-amber-900/20 px-4 py-3 flex items-center gap-3"
      >
        <InboxIcon class="size-4 text-amber-600 flex-shrink-0" />
        <p class="text-sm text-amber-800 dark:text-amber-300">
          Open a project and go to
          <strong>Triage</strong> to review incoming issues.
        </p>
      </div>
    </div>
  </MainLayout>

  <CreateProjectModal
    :open="showCreateProject"
    @close="showCreateProject = false"
    @created="onProjectCreated"
  />

  <CreateIssueModal
    :open="showCreateIssue"
    @close="showCreateIssue = false"
    @created="onIssueCreated"
  />
</template>
