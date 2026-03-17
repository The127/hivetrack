<!--
  ProjectBacklogView — fused backlog + sprint planning view.

  Shows the active sprint, planning sprints, and the unsprinted backlog in a
  single page (Jira-style). Controls to create sprints, move issues between
  sprints, activate, and complete sprints all live here.
-->
<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import {
  PlusIcon,
  ListIcon,
  LayersIcon,
  CircleIcon,
  CircleDotIcon,
  GitPullRequestIcon,
  CheckCircle2Icon,
  XCircleIcon,
  ChevronDownIcon,
  PlayIcon,
  CheckIcon,
  Trash2Icon,
} from 'lucide-vue-next'
import MainLayout from '@/layouts/MainLayout.vue'
import Badge from '@/components/ui/Badge.vue'
import Spinner from '@/components/ui/Spinner.vue'
import Avatar from '@/components/ui/Avatar.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import CreateIssueModal from '@/components/issue/CreateIssueModal.vue'
import { fetchProject } from '@/api/projects'
import { fetchIssues, updateIssue } from '@/api/issues'
import { fetchSprints, createSprint, updateSprint, deleteSprint } from '@/api/sprints'

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
  queryFn: () => fetchIssues(slug.value, { triaged: true, limit: 1000 }),
  enabled: computed(() => !!slug.value),
})

const isLoading = computed(() => loadingProject.value || loadingSprints.value || loadingIssues.value)

// ── Grouping ──────────────────────────────────────────────────────────────────

const allIssues = computed(() => issuesResult.value?.items ?? [])
const allSprints = computed(() => sprintsResult.value?.sprints ?? [])

const activeSprint = computed(() => allSprints.value.find(s => s.status === 'active') ?? null)
const planningSprints = computed(() => allSprints.value.filter(s => s.status === 'planning'))

const sprintIssuesMap = computed(() => {
  const map = {}
  for (const issue of allIssues.value) {
    const key = issue.sprint_id ?? '__backlog__'
    if (!map[key]) map[key] = []
    map[key].push(issue)
  }
  return map
})

function issuesForSprint(sprintId) {
  return sprintIssuesMap.value[sprintId] ?? []
}

const backlogIssues = computed(() => {
  const issues = sprintIssuesMap.value['__backlog__'] ?? []
  return [...issues].sort((a, b) => {
    const pa = PRIORITY_ORDER[a.priority] ?? 4
    const pb = PRIORITY_ORDER[b.priority] ?? 4
    if (pa !== pb) return pa - pb
    return new Date(a.created_at) - new Date(b.created_at)
  })
})

// ── Status / priority display ─────────────────────────────────────────────────

const PRIORITY_ORDER = { critical: 0, high: 1, medium: 2, low: 3, none: 4 }

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

function statusIconClass(scheme) {
  return {
    'text-slate-400':  scheme === 'gray',
    'text-blue-500':   scheme === 'blue',
    'text-violet-500': scheme === 'violet',
    'text-green-500':  scheme === 'green',
    'text-sky-500':    scheme === 'sky',
    'text-teal-500':   scheme === 'teal',
  }
}

// ── Move issue mutation ───────────────────────────────────────────────────────

const { mutate: moveIssue } = useMutation({
  mutationFn: ({ issueNumber, sprintId }) =>
    updateIssue(slug.value, issueNumber, { sprint_id: sprintId }),
  onMutate: async ({ issueNumber, sprintId }) => {
    const key = ['issues', slug.value, { triaged: true }]
    await queryClient.cancelQueries({ queryKey: key })
    const previous = queryClient.getQueryData(key)
    queryClient.setQueryData(key, old => {
      if (!old) return old
      return {
        ...old,
        items: old.items.map(i =>
          i.number === issueNumber ? { ...i, sprint_id: sprintId } : i
        ),
      }
    })
    return { previous }
  },
  onError: (_err, _vars, context) => {
    if (context?.previous) {
      queryClient.setQueryData(['issues', slug.value, { triaged: true }], context.previous)
    }
  },
  onSettled: () => {
    queryClient.invalidateQueries({ queryKey: ['issues', slug.value] })
  },
})

// ── Sprint status mutations ───────────────────────────────────────────────────

const sprintErrors = ref({})

const { mutate: activateSprint } = useMutation({
  mutationFn: (sprintId) => updateSprint(slug.value, sprintId, { status: 'active' }),
  onSuccess: (_data, sprintId) => {
    delete sprintErrors.value[sprintId]
    queryClient.invalidateQueries({ queryKey: ['sprints', slug.value] })
  },
  onError: (err, sprintId) => {
    sprintErrors.value[sprintId] = (err?.status === 409 || String(err?.message).includes('409'))
      ? 'Another sprint is already active.'
      : 'Failed to activate sprint.'
    queryClient.invalidateQueries({ queryKey: ['sprints', slug.value] })
  },
})

const { mutate: completeSprint } = useMutation({
  mutationFn: (sprintId) => updateSprint(slug.value, sprintId, { status: 'completed' }),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['sprints', slug.value] })
    queryClient.invalidateQueries({ queryKey: ['issues', slug.value] })
  },
  onError: () => {
    queryClient.invalidateQueries({ queryKey: ['sprints', slug.value] })
  },
})

const { mutate: doDeleteSprint } = useMutation({
  mutationFn: (sprintId) => deleteSprint(slug.value, sprintId),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['sprints', slug.value] })
    queryClient.invalidateQueries({ queryKey: ['issues', slug.value] })
  },
  onError: () => {
    queryClient.invalidateQueries({ queryKey: ['sprints', slug.value] })
  },
})

// ── New sprint form ───────────────────────────────────────────────────────────

const showNewSprintForm = ref(false)
const newSprint = ref({ name: '', start_date: '', end_date: '', goal: '' })
const newSprintError = ref('')

const { mutate: submitNewSprint, isPending: creatingSprintPending } = useMutation({
  mutationFn: (data) => createSprint(slug.value, data),
  onSuccess: () => {
    showNewSprintForm.value = false
    newSprint.value = { name: '', start_date: '', end_date: '', goal: '' }
    newSprintError.value = ''
    queryClient.invalidateQueries({ queryKey: ['sprints', slug.value] })
  },
  onError: () => {
    newSprintError.value = 'Failed to create sprint.'
  },
})

function handleCreateSprint() {
  if (!newSprint.value.name.trim()) {
    newSprintError.value = 'Name is required.'
    return
  }
  const data = { name: newSprint.value.name.trim() }
  if (newSprint.value.start_date) data.start_date = newSprint.value.start_date
  if (newSprint.value.end_date) data.end_date = newSprint.value.end_date
  if (newSprint.value.goal.trim()) data.goal = newSprint.value.goal.trim()
  submitNewSprint(data)
}

// ── Move-to-sprint dropdown ───────────────────────────────────────────────────

const openDropdown = ref(null)

function toggleDropdown(issueId) {
  openDropdown.value = openDropdown.value === issueId ? null : issueId
}

function moveToSprint(issue, sprintId) {
  openDropdown.value = null
  moveIssue({ issueNumber: issue.number, sprintId })
}

function moveToBacklog(issue) {
  moveIssue({ issueNumber: issue.number, sprintId: null })
}

const targetSprints = computed(() => {
  const sprints = []
  if (activeSprint.value) sprints.push(activeSprint.value)
  sprints.push(...planningSprints.value)
  return sprints
})

// ── Default status for issue creation ────────────────────────────────────────

const defaultCreateStatus = computed(() => {
  if (!project.value) return null
  return project.value.archetype === 'support' ? 'open' : 'todo'
})

const showCreateIssue = ref(false)

// ── Helpers ───────────────────────────────────────────────────────────────────

function formatDateRange(startDate, endDate) {
  const fmt = (d) => new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  return `${fmt(startDate)} – ${fmt(endDate)}`
}
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

      <!-- ── Content ────────────────────────────────────────────────────── -->
      <div v-else class="flex-1 overflow-y-auto">

        <!-- ── Active Sprint ─────────────────────────────────────────────── -->
        <template v-if="activeSprint">
          <div class="px-6 py-2.5 border-b border-slate-100 bg-blue-50 flex items-center justify-between">
            <div class="flex items-center gap-2 min-w-0">
              <span class="text-xs font-medium text-blue-700 bg-blue-100 px-1.5 py-0.5 rounded uppercase tracking-wide flex-shrink-0">Active</span>
              <span class="font-semibold text-slate-900 text-sm truncate">{{ activeSprint.name }}</span>
              <span v-if="activeSprint.start_date" class="text-xs text-slate-500 flex-shrink-0">
                {{ formatDateRange(activeSprint.start_date, activeSprint.end_date) }}
              </span>
              <span v-if="activeSprint.goal" class="text-xs text-slate-500 italic truncate max-w-48">{{ activeSprint.goal }}</span>
              <span class="text-xs text-slate-400 tabular-nums flex-shrink-0">
                {{ issuesForSprint(activeSprint.id).length }} issues
              </span>
            </div>
            <div class="flex items-center gap-2 flex-shrink-0">
              <button
                class="inline-flex items-center gap-1.5 rounded-md border border-slate-300 bg-white px-2.5 h-7 text-xs font-medium text-slate-600 hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors cursor-pointer"
                @click="completeSprint(activeSprint.id)"
              >
                <CheckIcon class="size-3.5" />
                Complete sprint
              </button>
              <button
                class="inline-flex items-center rounded-md border border-slate-300 bg-white px-2 h-7 text-slate-400 hover:text-red-600 hover:border-red-300 focus-visible:outline-none transition-colors cursor-pointer"
                title="Delete sprint (moves issues to backlog)"
                @click="doDeleteSprint(activeSprint.id)"
              >
                <Trash2Icon class="size-3.5" />
              </button>
            </div>
          </div>

          <div v-if="issuesForSprint(activeSprint.id).length" class="divide-y divide-slate-100">
            <div
              v-for="issue in issuesForSprint(activeSprint.id)"
              :key="issue.id"
              class="group flex items-center gap-3 px-6 py-2.5 hover:bg-slate-50 transition-colors cursor-pointer border-l-4"
              :class="priorityBorder(issue.priority)"
            >
              <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
                <span class="text-[11px] font-mono text-slate-400">{{ slug.toUpperCase() }}-{{ issue.number }}</span>
                <LayersIcon v-if="issue.type === 'epic'" class="size-3 text-violet-400 flex-shrink-0" />
              </div>
              <span class="flex-1 min-w-0 text-sm text-slate-800 truncate group-hover:text-slate-900">{{ issue.title }}</span>
              <span v-if="issue.on_hold" class="flex-shrink-0 text-[10px] font-medium bg-amber-100 text-amber-700 px-1.5 py-0.5 rounded">on hold</span>
              <component :is="statusMeta(issue.status).icon" class="size-3.5 flex-shrink-0" :class="statusIconClass(statusMeta(issue.status).scheme)" />
              <span class="flex-shrink-0 text-xs text-slate-500 w-20">{{ statusMeta(issue.status).label }}</span>
              <Badge v-if="issue.priority && issue.priority !== 'none'" :colorScheme="priorityScheme(issue.priority)" compact class="flex-shrink-0">{{ issue.priority }}</Badge>
              <span v-else class="w-14 flex-shrink-0" />
              <span v-if="estimateLabel(issue.estimate)" class="flex-shrink-0 text-[11px] font-medium text-slate-500 bg-slate-100 px-1.5 py-0.5 rounded w-7 text-center">{{ estimateLabel(issue.estimate) }}</span>
              <span v-else class="w-7 flex-shrink-0" />
              <div class="flex-shrink-0 flex -space-x-1 w-10 justify-end">
                <Avatar v-for="a in (issue.assignees ?? []).slice(0, 2)" :key="a" :name="`${a}`" size="xs" class="ring-1 ring-white" />
              </div>
              <button
                class="flex-shrink-0 opacity-0 group-hover:opacity-100 text-[11px] text-slate-400 hover:text-slate-600 transition-opacity cursor-pointer px-1.5 py-0.5 rounded hover:bg-slate-100"
                @click.stop="moveToBacklog(issue)"
              >
                ↓ Backlog
              </button>
            </div>
          </div>
          <div v-else class="px-6 py-3 text-sm text-slate-400 italic">No issues in this sprint.</div>
        </template>

        <!-- ── Planning Sprints ──────────────────────────────────────────── -->
        <template v-for="sprint in planningSprints" :key="sprint.id">
          <div class="px-6 py-2.5 border-b border-slate-100 bg-slate-50 flex items-center justify-between">
            <div class="flex items-center gap-2 min-w-0">
              <span class="text-xs font-medium text-slate-500 bg-slate-200 px-1.5 py-0.5 rounded uppercase tracking-wide flex-shrink-0">Planning</span>
              <span class="font-semibold text-slate-900 text-sm truncate">{{ sprint.name }}</span>
              <span v-if="sprint.start_date" class="text-xs text-slate-500 flex-shrink-0">
                {{ formatDateRange(sprint.start_date, sprint.end_date) }}
              </span>
              <span class="text-xs text-slate-400 tabular-nums flex-shrink-0">
                {{ issuesForSprint(sprint.id).length }} issues
              </span>
            </div>
            <div class="flex items-center gap-2 flex-shrink-0">
              <span v-if="sprintErrors[sprint.id]" class="text-xs text-red-600">{{ sprintErrors[sprint.id] }}</span>
              <button
                class="inline-flex items-center gap-1.5 rounded-md border border-slate-300 bg-white px-2.5 h-7 text-xs font-medium text-slate-600 hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors cursor-pointer"
                @click="activateSprint(sprint.id)"
              >
                <PlayIcon class="size-3.5" />
                Activate
              </button>
              <button
                class="inline-flex items-center rounded-md border border-slate-300 bg-white px-2 h-7 text-slate-400 hover:text-red-600 hover:border-red-300 focus-visible:outline-none transition-colors cursor-pointer"
                title="Delete sprint (moves issues to backlog)"
                @click="doDeleteSprint(sprint.id)"
              >
                <Trash2Icon class="size-3.5" />
              </button>
            </div>
          </div>

          <div v-if="issuesForSprint(sprint.id).length" class="divide-y divide-slate-100">
            <div
              v-for="issue in issuesForSprint(sprint.id)"
              :key="issue.id"
              class="group flex items-center gap-3 px-6 py-2.5 hover:bg-slate-50 transition-colors cursor-pointer border-l-4"
              :class="priorityBorder(issue.priority)"
            >
              <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
                <span class="text-[11px] font-mono text-slate-400">{{ slug.toUpperCase() }}-{{ issue.number }}</span>
                <LayersIcon v-if="issue.type === 'epic'" class="size-3 text-violet-400 flex-shrink-0" />
              </div>
              <span class="flex-1 min-w-0 text-sm text-slate-800 truncate group-hover:text-slate-900">{{ issue.title }}</span>
              <span v-if="issue.on_hold" class="flex-shrink-0 text-[10px] font-medium bg-amber-100 text-amber-700 px-1.5 py-0.5 rounded">on hold</span>
              <component :is="statusMeta(issue.status).icon" class="size-3.5 flex-shrink-0" :class="statusIconClass(statusMeta(issue.status).scheme)" />
              <span class="flex-shrink-0 text-xs text-slate-500 w-20">{{ statusMeta(issue.status).label }}</span>
              <Badge v-if="issue.priority && issue.priority !== 'none'" :colorScheme="priorityScheme(issue.priority)" compact class="flex-shrink-0">{{ issue.priority }}</Badge>
              <span v-else class="w-14 flex-shrink-0" />
              <span v-if="estimateLabel(issue.estimate)" class="flex-shrink-0 text-[11px] font-medium text-slate-500 bg-slate-100 px-1.5 py-0.5 rounded w-7 text-center">{{ estimateLabel(issue.estimate) }}</span>
              <span v-else class="w-7 flex-shrink-0" />
              <div class="flex-shrink-0 flex -space-x-1 w-10 justify-end">
                <Avatar v-for="a in (issue.assignees ?? []).slice(0, 2)" :key="a" :name="`${a}`" size="xs" class="ring-1 ring-white" />
              </div>
              <button
                class="flex-shrink-0 opacity-0 group-hover:opacity-100 text-[11px] text-slate-400 hover:text-slate-600 transition-opacity cursor-pointer px-1.5 py-0.5 rounded hover:bg-slate-100"
                @click.stop="moveToBacklog(issue)"
              >
                ↓ Backlog
              </button>
            </div>
          </div>
          <div v-else class="px-6 py-3 text-sm text-slate-400 italic">No issues in this sprint.</div>
        </template>

        <!-- ── Backlog section header ─────────────────────────────────────── -->
        <div class="px-6 py-2.5 border-b border-slate-100 bg-white flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="font-semibold text-slate-900 text-sm">Backlog</span>
            <span class="text-xs text-slate-400 tabular-nums">{{ backlogIssues.length }} issues</span>
          </div>
          <button
            class="inline-flex items-center gap-1.5 rounded-md border border-slate-300 bg-white px-2.5 h-7 text-xs font-medium text-slate-600 hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors cursor-pointer"
            @click="showNewSprintForm = !showNewSprintForm"
          >
            <PlusIcon class="size-3.5" />
            New sprint
          </button>
        </div>

        <!-- Backlog issues -->
        <div v-if="backlogIssues.length" class="divide-y divide-slate-100">
          <div
            v-for="issue in backlogIssues"
            :key="issue.id"
            class="group relative flex items-center gap-3 px-6 py-2.5 hover:bg-slate-50 transition-colors cursor-pointer border-l-4"
            :class="priorityBorder(issue.priority)"
          >
            <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
              <span class="text-[11px] font-mono text-slate-400">{{ slug.toUpperCase() }}-{{ issue.number }}</span>
              <LayersIcon v-if="issue.type === 'epic'" class="size-3 text-violet-400 flex-shrink-0" />
            </div>
            <span class="flex-1 min-w-0 text-sm text-slate-800 truncate group-hover:text-slate-900">{{ issue.title }}</span>
            <span v-if="issue.on_hold" class="flex-shrink-0 text-[10px] font-medium bg-amber-100 text-amber-700 px-1.5 py-0.5 rounded">on hold</span>
            <component :is="statusMeta(issue.status).icon" class="size-3.5 flex-shrink-0" :class="statusIconClass(statusMeta(issue.status).scheme)" />
            <span class="flex-shrink-0 text-xs text-slate-500 w-20">{{ statusMeta(issue.status).label }}</span>
            <Badge v-if="issue.priority && issue.priority !== 'none'" :colorScheme="priorityScheme(issue.priority)" compact class="flex-shrink-0">{{ issue.priority }}</Badge>
            <span v-else class="w-14 flex-shrink-0" />
            <span v-if="estimateLabel(issue.estimate)" class="flex-shrink-0 text-[11px] font-medium text-slate-500 bg-slate-100 px-1.5 py-0.5 rounded w-7 text-center">{{ estimateLabel(issue.estimate) }}</span>
            <span v-else class="w-7 flex-shrink-0" />
            <div class="flex-shrink-0 flex -space-x-1 w-10 justify-end">
              <Avatar v-for="a in (issue.assignees ?? []).slice(0, 2)" :key="a" :name="`${a}`" size="xs" class="ring-1 ring-white" />
            </div>

            <!-- Move to sprint dropdown -->
            <div v-if="targetSprints.length" class="relative flex-shrink-0">
              <button
                class="opacity-0 group-hover:opacity-100 text-[11px] text-slate-400 hover:text-slate-600 transition-opacity cursor-pointer px-1.5 py-0.5 rounded hover:bg-slate-100 inline-flex items-center gap-0.5"
                @click.stop="toggleDropdown(issue.id)"
              >
                → Sprint <ChevronDownIcon class="size-3" />
              </button>
              <div
                v-if="openDropdown === issue.id"
                class="absolute right-0 top-full mt-1 z-10 bg-white border border-slate-200 rounded-md shadow-md py-1 min-w-36"
              >
                <button
                  v-for="sprint in targetSprints"
                  :key="sprint.id"
                  class="w-full text-left px-3 py-1.5 text-xs text-slate-700 hover:bg-slate-50 cursor-pointer truncate"
                  @click="moveToSprint(issue, sprint.id)"
                >
                  {{ sprint.name }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Empty state when nothing anywhere -->
        <div
          v-else-if="!activeSprint && !planningSprints.length"
          class="flex items-center justify-center py-16"
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

        <!-- Inline new sprint form -->
        <div v-if="showNewSprintForm" class="px-6 py-4 border-b border-slate-100 bg-slate-50">
          <div class="max-w-lg space-y-3">
            <div class="text-sm font-medium text-slate-700">New sprint</div>

            <input
              v-model="newSprint.name"
              type="text"
              placeholder="Sprint name (required)"
              class="w-full rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />

            <div class="flex gap-3">
              <div class="flex-1">
                <label class="text-xs text-slate-500 mb-1 block">Start date</label>
                <input
                  v-model="newSprint.start_date"
                  type="date"
                  class="w-full rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
              <div class="flex-1">
                <label class="text-xs text-slate-500 mb-1 block">End date</label>
                <input
                  v-model="newSprint.end_date"
                  type="date"
                  class="w-full rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>

            <input
              v-model="newSprint.goal"
              type="text"
              placeholder="Sprint goal (optional)"
              class="w-full rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />

            <div v-if="newSprintError" class="text-xs text-red-600">{{ newSprintError }}</div>

            <div class="flex items-center gap-2">
              <button
                class="inline-flex items-center gap-1.5 rounded-md bg-blue-600 px-3 h-7 text-xs font-medium text-white hover:bg-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors cursor-pointer disabled:opacity-50"
                :disabled="creatingSprintPending"
                @click="handleCreateSprint"
              >
                Create sprint
              </button>
              <button
                class="inline-flex items-center rounded-md border border-slate-300 bg-white px-3 h-7 text-xs font-medium text-slate-600 hover:bg-slate-50 focus-visible:outline-none transition-colors cursor-pointer"
                @click="showNewSprintForm = false; newSprintError = ''"
              >
                Cancel
              </button>
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
