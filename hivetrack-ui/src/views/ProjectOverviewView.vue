<!--
  ProjectOverviewView — at-a-glance summary for a project.

  Shows project metadata, active sprint status, issue counts by status,
  and triage inbox count. All data is composed from existing queries —
  no dedicated overview endpoint needed.
-->
<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import {
  KanbanIcon,
  ListIcon,
  InboxIcon,
  LayersIcon,
  FlagIcon,
  CodeIcon,
  HeadphonesIcon,
  CalendarIcon,
  UsersIcon,
  TagIcon,
  CircleIcon,
  CircleDotIcon,
  GitPullRequestIcon,
  CheckCircle2Icon,
  XCircleIcon,
  Trash2Icon,
  SlidersHorizontalIcon,
} from 'lucide-vue-next'
import MainLayout from '@/layouts/MainLayout.vue'
import Badge from '@/components/ui/Badge.vue'
import Spinner from '@/components/ui/Spinner.vue'
import ProgressBar from '@/components/ui/ProgressBar.vue'
import SprintBurndownChart from '@/components/sprint/SprintBurndownChart.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import { fetchProject, removeProjectMember, updateProject } from '@/api/projects'
import { fetchLabels, deleteLabel } from '@/api/labels'
import { fetchIssues } from '@/api/issues'
import { fetchSprints, fetchSprintBurndown } from '@/api/sprints'

const route = useRoute()
const slug = computed(() => route.params.slug)
const queryClient = useQueryClient()

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

const { data: labelsData } = useQuery({
  queryKey: ['labels', slug],
  queryFn: () => fetchLabels(slug.value),
  enabled: computed(() => !!slug.value),
})

const labels = computed(() => labelsData.value?.labels ?? [])

// ── Mutations ─────────────────────────────────────────────────────────────────

const memberToRemove = ref(null)
const labelToDelete = ref(null)

const { mutate: doRemoveMember, isPending: removeMemberPending } = useMutation({
  mutationFn: (userId) => removeProjectMember(slug.value, userId),
  onSuccess: () => {
    memberToRemove.value = null
    queryClient.invalidateQueries({ queryKey: ['project', slug.value] })
  },
})

const { mutate: doDeleteLabel, isPending: deleteLabelPending } = useMutation({
  mutationFn: (labelId) => deleteLabel(slug.value, labelId),
  onSuccess: () => {
    labelToDelete.value = null
    queryClient.invalidateQueries({ queryKey: ['labels', slug.value] })
  },
})

// ── WIP limit settings ────────────────────────────────────────────────────────

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
    queryClient.invalidateQueries({ queryKey: ['project', slug.value] })
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

const SOFTWARE_STATUSES = [
  { key: 'todo', label: 'To Do', icon: CircleIcon, scheme: 'gray' },
  { key: 'in_progress', label: 'In Progress', icon: CircleDotIcon, scheme: 'blue' },
  { key: 'in_review', label: 'In Review', icon: GitPullRequestIcon, scheme: 'violet' },
  { key: 'done', label: 'Done', icon: CheckCircle2Icon, scheme: 'green' },
  { key: 'cancelled', label: 'Cancelled', icon: XCircleIcon, scheme: 'gray' },
]

const SUPPORT_STATUSES = [
  { key: 'open', label: 'Open', icon: CircleIcon, scheme: 'sky' },
  { key: 'in_progress', label: 'In Progress', icon: CircleDotIcon, scheme: 'blue' },
  { key: 'resolved', label: 'Resolved', icon: CheckCircle2Icon, scheme: 'teal' },
  { key: 'closed', label: 'Closed', icon: XCircleIcon, scheme: 'gray' },
]

const statusDefs = computed(() =>
  project.value?.archetype === 'support' ? SUPPORT_STATUSES : SOFTWARE_STATUSES,
)

const statusCounts = computed(() => {
  const counts = {}
  for (const s of statusDefs.value) counts[s.key] = 0
  for (const issue of allIssues.value) {
    if (counts[issue.status] !== undefined) counts[issue.status]++
  }
  return counts
})

const TERMINAL = { software: new Set(['done', 'cancelled']), support: new Set(['resolved', 'closed']) }

const sprintIssues = computed(() => {
  if (!activeSprint.value) return []
  return allIssues.value.filter((i) => i.sprint_id === activeSprint.value.id)
})

const sprintDone = computed(() => {
  if (!activeSprint.value) return 0
  const terminal = TERMINAL[project.value?.archetype] ?? TERMINAL.software
  return sprintIssues.value.filter((i) => terminal.has(i.status)).length
})

const SCHEME_ICON_CLASS = {
  gray: 'text-slate-400',
  blue: 'text-blue-500',
  violet: 'text-violet-500',
  green: 'text-green-500',
  sky: 'text-sky-500',
  teal: 'text-teal-500',
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}

function formatDateRange(start, end) {
  return `${formatDate(start)} – ${formatDate(end)}`
}
</script>

<template>
  <MainLayout>
    <!-- Loading -->
    <div v-if="loadingProject" class="flex justify-center items-center h-40">
      <Spinner class="size-5 text-slate-400" />
    </div>

    <div v-else-if="project" class="max-w-3xl mx-auto px-6 py-8 space-y-8">

      <!-- ── Project header ──────────────────────────────────────────────────── -->
      <div>
        <div class="flex items-start gap-4">
          <span
            class="size-12 rounded-lg flex items-center justify-center text-base font-bold bg-slate-100 text-slate-600 flex-shrink-0"
          >
            {{ project.slug.slice(0, 2).toUpperCase() }}
          </span>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 flex-wrap">
              <h1 class="text-xl font-semibold text-slate-900">{{ project.name }}</h1>
              <Badge :colorScheme="project.archetype === 'software' ? 'blue' : 'teal'" compact>
                <component
                  :is="project.archetype === 'software' ? CodeIcon : HeadphonesIcon"
                  class="size-3 mr-0.5"
                />
                {{ project.archetype }}
              </Badge>
            </div>
            <p v-if="project.description" class="text-sm text-slate-500 mt-0.5">
              {{ project.description }}
            </p>
            <p class="text-xs text-slate-400 mt-1 font-mono">{{ project.slug }}</p>
          </div>
        </div>
      </div>

      <!-- ── Quick navigation ────────────────────────────────────────────────── -->
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <RouterLink
          :to="`/projects/${slug}/board`"
          class="flex items-center gap-2.5 rounded-lg border border-slate-200 px-4 py-3 hover:bg-slate-50 hover:border-slate-300 transition-colors group"
        >
          <KanbanIcon class="size-4 text-slate-500 group-hover:text-slate-700 flex-shrink-0" />
          <span class="text-sm font-medium text-slate-700 group-hover:text-slate-900">Board</span>
        </RouterLink>
        <RouterLink
          :to="`/projects/${slug}/backlog`"
          class="flex items-center gap-2.5 rounded-lg border border-slate-200 px-4 py-3 hover:bg-slate-50 hover:border-slate-300 transition-colors group"
        >
          <ListIcon class="size-4 text-slate-500 group-hover:text-slate-700 flex-shrink-0" />
          <span class="text-sm font-medium text-slate-700 group-hover:text-slate-900">Backlog</span>
        </RouterLink>
        <RouterLink
          :to="`/projects/${slug}/epics`"
          class="flex items-center gap-2.5 rounded-lg border border-slate-200 px-4 py-3 hover:bg-slate-50 hover:border-slate-300 transition-colors group"
        >
          <LayersIcon class="size-4 text-slate-500 group-hover:text-slate-700 flex-shrink-0" />
          <span class="text-sm font-medium text-slate-700 group-hover:text-slate-900">Epics</span>
        </RouterLink>
        <RouterLink
          :to="`/projects/${slug}/milestones`"
          class="flex items-center gap-2.5 rounded-lg border border-slate-200 px-4 py-3 hover:bg-slate-50 hover:border-slate-300 transition-colors group"
        >
          <FlagIcon class="size-4 text-slate-500 group-hover:text-slate-700 flex-shrink-0" />
          <span class="text-sm font-medium text-slate-700 group-hover:text-slate-900">Milestones</span>
        </RouterLink>
      </div>

      <!-- ── Active sprint ────────────────────────────────────────────────────── -->
      <section v-if="activeSprint">
        <h2 class="text-sm font-medium text-slate-700 mb-3">Active Sprint</h2>
        <div class="rounded-lg border border-blue-100 bg-blue-50 px-4 py-3">
          <div class="flex items-center justify-between gap-3">
            <div class="flex items-center gap-2 min-w-0">
              <span
                class="text-xs font-medium text-blue-700 bg-blue-100 px-1.5 py-0.5 rounded uppercase tracking-wide flex-shrink-0"
              >Sprint</span>
              <span class="text-sm font-semibold text-slate-900 truncate">{{ activeSprint.name }}</span>
              <span v-if="activeSprint.start_date" class="text-xs text-slate-500 flex-shrink-0">
                <CalendarIcon class="size-3 inline mr-0.5 -mt-px" />
                {{ formatDateRange(activeSprint.start_date, activeSprint.end_date) }}
              </span>
            </div>
            <div class="w-32 flex-shrink-0">
              <ProgressBar :done="sprintDone" :total="sprintIssues.length" />
            </div>
          </div>
          <p v-if="activeSprint.goal" class="text-xs text-slate-600 mt-2">
            <span class="text-slate-400 font-medium">Goal:</span> {{ activeSprint.goal }}
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
        <h2 class="text-sm font-medium text-slate-700 mb-3">Issues</h2>
        <div class="rounded-lg border border-slate-200 divide-y divide-slate-100 overflow-hidden">
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
            <span class="text-sm text-slate-700 flex-1">{{ s.label }}</span>
            <span class="text-sm font-medium text-slate-900 tabular-nums">
              {{ statusCounts[s.key] }}
            </span>
          </div>
          <!-- Inbox row -->
          <RouterLink
            :to="`/projects/${slug}/triage`"
            class="flex items-center gap-3 px-4 py-2.5 hover:bg-slate-50 transition-colors group"
          >
            <InboxIcon class="size-4 flex-shrink-0 text-amber-500" />
            <span class="text-sm text-slate-700 flex-1 group-hover:text-slate-900">Inbox (untriaged)</span>
            <span
              class="text-sm font-medium tabular-nums"
              :class="inboxCount > 0 ? 'text-amber-600' : 'text-slate-900'"
            >
              {{ inboxCount }}
            </span>
          </RouterLink>
        </div>
      </section>

      <!-- ── Members ─────────────────────────────────────────────────────────── -->
      <section v-if="project.members?.length">
        <h2 class="text-sm font-medium text-slate-700 mb-3 flex items-center gap-1.5">
          <UsersIcon class="size-4 text-slate-500" />
          Members
          <span class="text-xs font-normal text-slate-500">{{ project.members.length }}</span>
        </h2>
        <div class="flex flex-wrap gap-2">
          <div
            v-for="m in project.members"
            :key="m.user_id"
            class="group flex items-center gap-2 rounded-md border border-slate-200 px-2.5 py-1.5"
          >
            <span class="text-sm text-slate-700">{{ m.display_name }}</span>
            <Badge colorScheme="gray" compact>{{ m.role.replace('project_', '') }}</Badge>
            <button
              class="opacity-0 group-hover:opacity-100 ml-0.5 rounded p-0.5 text-slate-400 hover:text-red-500 hover:bg-red-50 transition-all cursor-pointer"
              title="Remove member"
              @click="memberToRemove = m"
            >
              <Trash2Icon class="size-3" />
            </button>
          </div>
        </div>
      </section>

      <!-- ── Labels ─────────────────────────────────────────────────────────── -->
      <section v-if="labels.length">
        <h2 class="text-sm font-medium text-slate-700 mb-3 flex items-center gap-1.5">
          <TagIcon class="size-4 text-slate-500" />
          Labels
          <span class="text-xs font-normal text-slate-500">{{ labels.length }}</span>
        </h2>
        <div class="flex flex-wrap gap-2">
          <div
            v-for="label in labels"
            :key="label.id"
            class="group flex items-center gap-1.5 rounded-full border px-2.5 py-0.5"
            :style="{ borderColor: label.color + '66', backgroundColor: label.color + '22' }"
          >
            <span class="text-xs font-medium" :style="{ color: label.color }">{{ label.name }}</span>
            <button
              class="opacity-0 group-hover:opacity-100 rounded-full p-0.5 text-slate-400 hover:text-red-500 hover:bg-red-100 transition-all cursor-pointer"
              title="Delete label"
              @click="labelToDelete = label"
            >
              <XCircleIcon class="size-3" />
            </button>
          </div>
        </div>
      </section>

      <!-- ── Board settings (WIP limits, software only) ──────────────────── -->
      <section v-if="project.archetype === 'software'">
        <h2 class="text-sm font-medium text-slate-700 mb-3 flex items-center gap-1.5">
          <SlidersHorizontalIcon class="size-4 text-slate-500" />
          Board
        </h2>
        <div class="rounded-lg border border-slate-200 divide-y divide-slate-100 overflow-hidden">
          <div class="flex items-center gap-3 px-4 py-2.5">
            <CircleDotIcon class="size-4 flex-shrink-0 text-blue-500" />
            <span class="text-sm text-slate-700 flex-1">In Progress limit</span>
            <input
              v-model="wipInProgressInput"
              type="number"
              min="1"
              placeholder="None"
              class="w-20 text-sm text-right border border-slate-200 rounded px-2 py-0.5 text-slate-700 focus:outline-none focus:ring-1 focus:ring-blue-400 focus:border-blue-400"
              :disabled="savingWip"
              @blur="saveWipInProgress"
              @keydown.enter="$event.target.blur()"
            />
          </div>
          <div class="flex items-center gap-3 px-4 py-2.5">
            <GitPullRequestIcon class="size-4 flex-shrink-0 text-violet-500" />
            <span class="text-sm text-slate-700 flex-1">In Review limit</span>
            <input
              v-model="wipInReviewInput"
              type="number"
              min="1"
              placeholder="None"
              class="w-20 text-sm text-right border border-slate-200 rounded px-2 py-0.5 text-slate-700 focus:outline-none focus:ring-1 focus:ring-blue-400 focus:border-blue-400"
              :disabled="savingWip"
              @blur="saveWipInReview"
              @keydown.enter="$event.target.blur()"
            />
          </div>
        </div>
        <p class="text-xs text-slate-400 mt-1.5">Informational only — the board highlights columns that exceed these limits.</p>
      </section>

    </div>

    <!-- Remove member confirmation -->
    <ConfirmDialog
      v-if="memberToRemove"
      :open="!!memberToRemove"
      title="Remove member?"
      :message="`Remove ${memberToRemove.display_name} from this project? They will lose access immediately.`"
      confirm-text="Remove member"
      :loading="removeMemberPending"
      @confirm="doRemoveMember(memberToRemove.user_id)"
      @cancel="memberToRemove = null"
    />

    <!-- Delete label confirmation -->
    <ConfirmDialog
      v-if="labelToDelete"
      :open="!!labelToDelete"
      title="Delete label?"
      :message="`Delete '${labelToDelete.name}'? It will be removed from all issues.`"
      confirm-text="Delete label"
      :loading="deleteLabelPending"
      @confirm="doDeleteLabel(labelToDelete.id)"
      @cancel="labelToDelete = null"
    />
  </MainLayout>
</template>
