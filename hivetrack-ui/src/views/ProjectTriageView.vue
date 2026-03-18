<!--
  ProjectTriageView — the triage inbox.

  Issues with triaged=false land here from quick-capture (title-only creates)
  or external integrations (CI, monitoring, webhooks). Triaging means placing
  an issue into the workflow by assigning it a status and optionally a sprint
  and milestone.

  Each row expands inline to reveal a triage form. On confirm, the issue is
  removed from the inbox and placed on the board.

  Special triage actions:
  - Cancel: triage with terminal status (cancelled / closed)
  - Duplicate: triage as cancelled + create a "duplicates" link to the target
-->
<script setup>
import { ref, computed, reactive, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import {
  InboxIcon,
  PlusIcon,
  CheckIcon,
  XIcon,
  CopyIcon,
  SearchIcon,
} from 'lucide-vue-next'
import MainLayout from '@/layouts/MainLayout.vue'
import Badge from '@/components/ui/Badge.vue'
import Spinner from '@/components/ui/Spinner.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import MilestoneSelect from '@/components/issue/MilestoneSelect.vue'
import CreateIssueModal from '@/components/issue/CreateIssueModal.vue'
import Button from '@/components/ui/Button.vue'
import RelativeTime from '@/components/ui/RelativeTime.vue'
import { fetchIssues, triageIssue, createIssueLink } from '@/api/issues'
import { fetchProject } from '@/api/projects'
import { fetchSprints } from '@/api/sprints'

const route = useRoute()
const slug = computed(() => route.params.slug)
const queryClient = useQueryClient()

// ── Data ──────────────────────────────────────────────────────────────────────

const { data: project, isLoading: loadingProject } = useQuery({
  queryKey: ['project', slug],
  queryFn: () => fetchProject(slug.value),
  enabled: computed(() => !!slug.value),
})

const INBOX_KEY = computed(() => ['issues', slug.value, { triaged: false, limit: 500 }])

const { data: inboxResult, isLoading: loadingIssues } = useQuery({
  queryKey: INBOX_KEY,
  queryFn: () => fetchIssues(slug.value, { triaged: false, limit: 500 }),
  enabled: computed(() => !!slug.value),
})

const { data: sprintsResult } = useQuery({
  queryKey: ['sprints', slug],
  queryFn: () => fetchSprints(slug.value),
  enabled: computed(() => !!slug.value),
})

const isLoading = computed(() => loadingProject.value || loadingIssues.value)
const inbox = computed(() => inboxResult.value?.items ?? [])

const activeSprints = computed(() =>
  (sprintsResult.value?.sprints ?? []).filter(s =>
    s.status === 'active' || s.status === 'planning'
  )
)

// ── Status options per archetype ───────────────────────────────────────────────

const SOFTWARE_STATUSES = [
  { key: 'todo',        label: 'To Do',      terminal: false },
  { key: 'in_progress', label: 'In Progress', terminal: false },
  { key: 'in_review',   label: 'In Review',   terminal: false },
  { key: 'done',        label: 'Done',        terminal: true  },
  { key: 'cancelled',   label: 'Cancelled',   terminal: true  },
]
const SUPPORT_STATUSES = [
  { key: 'open',        label: 'Open',       terminal: false },
  { key: 'in_progress', label: 'In Progress', terminal: false },
  { key: 'resolved',    label: 'Resolved',   terminal: true  },
  { key: 'closed',      label: 'Closed',     terminal: true  },
]

const triageStatuses = computed(() =>
  project.value?.archetype === 'support' ? SUPPORT_STATUSES : SOFTWARE_STATUSES
)

const terminalStatus = computed(() =>
  project.value?.archetype === 'support' ? 'closed' : 'cancelled'
)

const defaultStatus = computed(() =>
  project.value?.archetype === 'support' ? 'open' : 'todo'
)

// ── Priority border ───────────────────────────────────────────────────────────

const PRIORITY_BORDER = {
  none:     'border-l-slate-200',
  low:      'border-l-sky-400',
  medium:   'border-l-amber-400',
  high:     'border-l-orange-500',
  critical: 'border-l-red-500',
}

function priorityBorder(priority) {
  return PRIORITY_BORDER[priority] ?? 'border-l-slate-200'
}

// ── Triage form state ─────────────────────────────────────────────────────────

// mode: null | 'triage' | 'duplicate'
const triagingId = ref(null)
const triageMode = ref(null)
const form = reactive({
  status: 'todo',
  sprint_id: null,
  milestone_id: null,
})

// Duplicate-specific state
const dupSearch = ref('')
const dupSelected = ref(null) // { id, number, title }

const { data: dupSearchResult } = useQuery({
  queryKey: computed(() => ['issues', slug.value, { text: dupSearch.value, limit: 10 }]),
  queryFn: () => fetchIssues(slug.value, { text: dupSearch.value, limit: 10 }),
  enabled: computed(() => !!slug.value && dupSearch.value.length >= 2),
})

// Filter out the issue being triaged from search results
const dupResults = computed(() => {
  const results = dupSearchResult.value?.items ?? []
  return results.filter(i => i.id !== triagingId.value)
})

function startTriage(issue) {
  triagingId.value = issue.id
  triageMode.value = 'triage'
  form.status = defaultStatus.value
  form.sprint_id = null
  form.milestone_id = null
  dupSearch.value = ''
  dupSelected.value = null
}

function startDuplicate(issue) {
  triagingId.value = issue.id
  triageMode.value = 'duplicate'
  form.status = terminalStatus.value
  form.sprint_id = null
  form.milestone_id = null
  dupSearch.value = ''
  dupSelected.value = null
}

function cancelTriage() {
  triagingId.value = null
  triageMode.value = null
}

// Auto-set terminal status when duplicate mode is active
watch(terminalStatus, (v) => {
  if (triageMode.value === 'duplicate') {
    form.status = v
  }
})

// ── Triage mutation ───────────────────────────────────────────────────────────

const { mutate: doTriage, isPending: triaging } = useMutation({
  mutationFn: ({ number, duplicateOf }) =>
    triageIssue(slug.value, number, {
      status: form.status,
      ...(form.sprint_id ? { sprint_id: form.sprint_id } : {}),
      ...(form.milestone_id ? { milestone_id: form.milestone_id } : {}),
    }).then(async () => {
      if (duplicateOf != null) {
        await createIssueLink(slug.value, number, {
          link_type: 'duplicates',
          target_number: duplicateOf,
        })
      }
    }),
  onMutate: async ({ issueId }) => {
    await queryClient.cancelQueries({ queryKey: INBOX_KEY.value })
    const previous = queryClient.getQueryData(INBOX_KEY.value)
    queryClient.setQueryData(INBOX_KEY.value, old =>
      old
        ? { ...old, items: old.items.filter(i => i.id !== issueId), total: old.total - 1 }
        : old
    )
    return { previous }
  },
  onError: (_err, _vars, ctx) => {
    if (ctx?.previous) {
      queryClient.setQueryData(INBOX_KEY.value, ctx.previous)
    }
  },
  onSettled: () => {
    triagingId.value = null
    triageMode.value = null
    queryClient.invalidateQueries({ queryKey: ['issues', slug.value] })
  },
})

function confirmTriage(issue) {
  doTriage({ number: issue.number, issueId: issue.id, duplicateOf: null })
}

function confirmDuplicate(issue) {
  if (!dupSelected.value) return
  doTriage({ number: issue.number, issueId: issue.id, duplicateOf: dupSelected.value.number })
}

// ── Quick capture ─────────────────────────────────────────────────────────────

const showCreate = ref(false)
</script>

<template>
  <MainLayout @create-issue="showCreate = true">
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

          <div class="flex items-center gap-1.5">
            <InboxIcon class="size-4 text-slate-400" />
            <span class="text-sm font-medium text-slate-600">Triage</span>
            <span
              v-if="inboxResult && inboxResult.total > 0"
              class="text-xs text-slate-400 tabular-nums"
            >
              ({{ inboxResult.total }})
            </span>
          </div>
        </div>

        <Button size="sm" @click="showCreate = true">
          <PlusIcon class="size-3.5" />
          Capture
        </Button>
      </div>

      <!-- ── Loading ────────────────────────────────────────────────────── -->
      <div v-if="isLoading" class="flex-1 flex items-center justify-center">
        <Spinner />
      </div>

      <!-- ── Empty inbox ────────────────────────────────────────────────── -->
      <div v-else-if="inbox.length === 0" class="flex-1 flex items-center justify-center">
        <EmptyState
          title="Inbox is clear"
          description="New quick-captures and external integrations land here for review."
        >
          <Button size="sm" @click="showCreate = true">
            <PlusIcon class="size-3.5" />
            Capture issue
          </Button>
        </EmptyState>
      </div>

      <!-- ── Issue list ──────────────────────────────────────────────────── -->
      <div v-else class="flex-1 overflow-y-auto">
        <div class="max-w-4xl mx-auto px-6 py-4">
          <div class="border border-slate-200 rounded-lg overflow-hidden">
            <div v-for="issue in inbox" :key="issue.id">

              <!-- Issue row -->
              <div
                class="group flex items-center gap-3 px-4 py-3 bg-white border-l-4 border-b border-slate-100 last:border-b-0"
                :class="priorityBorder(issue.priority)"
              >
                <!-- Issue number -->
                <router-link
                  :to="`/projects/${slug}/issues/${issue.number}`"
                  class="text-[11px] font-mono text-slate-400 hover:text-blue-600 flex-shrink-0 w-20"
                >
                  {{ slug.toUpperCase() }}-{{ issue.number }}
                </router-link>

                <!-- Type badge -->
                <span
                  class="flex-shrink-0 text-[10px] font-medium px-1.5 py-0.5 rounded"
                  :class="issue.type === 'epic' ? 'bg-violet-100 text-violet-700' : 'bg-slate-100 text-slate-600'"
                >
                  {{ issue.type }}
                </span>

                <!-- Title -->
                <router-link
                  :to="`/projects/${slug}/issues/${issue.number}`"
                  class="flex-1 min-w-0 text-sm text-slate-800 truncate hover:underline"
                >
                  {{ issue.title }}
                </router-link>

                <!-- Age -->
                <RelativeTime :datetime="issue.created_at" class="flex-shrink-0 text-xs text-slate-400" />

                <!-- Action buttons (when row is not expanded) -->
                <template v-if="triagingId !== issue.id">
                  <Button size="sm" variant="secondary" class="flex-shrink-0" @click="startTriage(issue)">
                    Triage
                  </Button>
                  <Button size="sm" variant="ghost" class="flex-shrink-0" title="Mark as duplicate" @click="startDuplicate(issue)">
                    <CopyIcon class="size-3.5" />
                  </Button>
                </template>
                <Button v-else size="sm" variant="ghost" class="flex-shrink-0" @click="cancelTriage">
                  <XIcon class="size-3.5" />
                  Cancel
                </Button>
              </div>

              <!-- ── Triage form ──────────────────────────────────────── -->
              <div
                v-if="triagingId === issue.id && triageMode === 'triage'"
                class="px-4 py-4 bg-slate-50 border-b border-slate-200 border-l-4"
                :class="priorityBorder(issue.priority)"
              >
                <div class="flex flex-col gap-4 max-w-xl">

                  <!-- Status (required) -->
                  <div class="flex flex-col gap-1.5">
                    <label class="text-xs font-medium text-slate-500">
                      Status <span class="text-red-400">*</span>
                    </label>
                    <div class="flex flex-wrap gap-1.5">
                      <button
                        v-for="s in triageStatuses"
                        :key="s.key"
                        class="px-3 py-1.5 text-xs font-medium rounded-md border transition-colors cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
                        :class="form.status === s.key
                          ? s.terminal
                            ? 'bg-slate-600 text-white border-slate-600'
                            : 'bg-blue-600 text-white border-blue-600'
                          : s.terminal
                            ? 'bg-white text-slate-500 border-slate-200 hover:border-slate-400 hover:text-slate-700'
                            : 'bg-white text-slate-600 border-slate-200 hover:border-slate-300 hover:bg-slate-50'"
                        @click="form.status = s.key"
                      >
                        {{ s.label }}
                      </button>
                    </div>
                  </div>

                  <!-- Sprint (optional, only if active/planning sprints exist) -->
                  <div v-if="activeSprints.length" class="flex flex-col gap-1.5">
                    <label class="text-xs font-medium text-slate-500">Sprint</label>
                    <select
                      v-model="form.sprint_id"
                      class="w-full rounded-md border border-slate-200 px-2.5 py-1.5 text-sm text-slate-700 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 cursor-pointer"
                    >
                      <option :value="null">Backlog (no sprint)</option>
                      <option v-for="s in activeSprints" :key="s.id" :value="s.id">
                        {{ s.name }}{{ s.status === 'active' ? ' (active)' : ' (planning)' }}
                      </option>
                    </select>
                  </div>

                  <!-- Milestone (optional) -->
                  <MilestoneSelect
                    :projectSlug="slug"
                    :modelValue="form.milestone_id"
                    @update:modelValue="form.milestone_id = $event"
                  />

                  <!-- Confirm / Cancel -->
                  <div class="flex items-center gap-2 pt-1">
                    <Button size="sm" :loading="triaging" @click="confirmTriage(issue)">
                      <CheckIcon class="size-3.5" />
                      Triage
                    </Button>
                    <Button size="sm" variant="ghost" @click="cancelTriage">
                      Cancel
                    </Button>
                  </div>

                </div>
              </div>

              <!-- ── Duplicate form ───────────────────────────────────── -->
              <div
                v-if="triagingId === issue.id && triageMode === 'duplicate'"
                class="px-4 py-4 bg-slate-50 border-b border-slate-200 border-l-4"
                :class="priorityBorder(issue.priority)"
              >
                <div class="flex flex-col gap-4 max-w-xl">

                  <p class="text-xs text-slate-500">
                    This issue will be marked as <span class="font-medium text-slate-700">{{ terminalStatus }}</span> and linked as a duplicate of the issue you select.
                  </p>

                  <!-- Issue search -->
                  <div class="flex flex-col gap-1.5">
                    <label class="text-xs font-medium text-slate-500">
                      Duplicate of <span class="text-red-400">*</span>
                    </label>

                    <!-- Selected target -->
                    <div v-if="dupSelected" class="flex items-center gap-2 px-3 py-2 bg-blue-50 border border-blue-200 rounded-md">
                      <span class="text-[11px] font-mono text-slate-500 flex-shrink-0">
                        {{ slug.toUpperCase() }}-{{ dupSelected.number }}
                      </span>
                      <span class="flex-1 text-sm text-slate-800 truncate">{{ dupSelected.title }}</span>
                      <button
                        class="text-slate-400 hover:text-slate-600 flex-shrink-0 cursor-pointer"
                        @click="dupSelected = null; dupSearch = ''"
                      >
                        <XIcon class="size-3.5" />
                      </button>
                    </div>

                    <!-- Search input + results -->
                    <div v-else class="relative">
                      <div class="relative">
                        <SearchIcon class="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-slate-400 pointer-events-none" />
                        <input
                          v-model="dupSearch"
                          type="text"
                          placeholder="Search issues…"
                          class="w-full pl-8 pr-3 py-1.5 text-sm rounded-md border border-slate-200 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                      </div>
                      <div
                        v-if="dupSearch.length >= 2 && dupResults.length"
                        class="absolute top-full left-0 right-0 mt-1 bg-white border border-slate-200 rounded-md shadow-md overflow-hidden z-10"
                      >
                        <button
                          v-for="r in dupResults"
                          :key="r.id"
                          class="w-full flex items-center gap-2 px-3 py-2 text-sm text-left hover:bg-slate-50 cursor-pointer transition-colors"
                          @click="dupSelected = r; dupSearch = ''"
                        >
                          <span class="text-[11px] font-mono text-slate-400 flex-shrink-0 w-20">
                            {{ slug.toUpperCase() }}-{{ r.number }}
                          </span>
                          <span class="flex-1 min-w-0 text-slate-800 truncate">{{ r.title }}</span>
                        </button>
                      </div>
                      <p v-else-if="dupSearch.length >= 2 && !dupResults.length" class="mt-1 text-xs text-slate-400">
                        No issues found.
                      </p>
                    </div>
                  </div>

                  <!-- Confirm / Cancel -->
                  <div class="flex items-center gap-2 pt-1">
                    <Button
                      size="sm"
                      :loading="triaging"
                      :disabled="!dupSelected"
                      @click="confirmDuplicate(issue)"
                    >
                      <CopyIcon class="size-3.5" />
                      Mark as duplicate
                    </Button>
                    <Button size="sm" variant="ghost" @click="cancelTriage">
                      Cancel
                    </Button>
                  </div>

                </div>
              </div>

            </div>
          </div>
        </div>
      </div>

    </div>

    <!-- ── Quick-capture modal (no defaultStatus → lands in inbox) ──────── -->
    <CreateIssueModal
      :open="showCreate"
      :projectSlug="slug"
      @close="showCreate = false"
      @created="() => { showCreate = false; queryClient.invalidateQueries({ queryKey: INBOX_KEY }) }"
    />
  </MainLayout>
</template>
