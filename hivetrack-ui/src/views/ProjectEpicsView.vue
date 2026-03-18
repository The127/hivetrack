<!--
  ProjectEpicsView — dedicated view for managing epics and their child tasks.

  Shows all epics in a project with progress bars, status, priority, and
  assignees. Each epic row expands to show child tasks via EpicChildList.
-->
<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import {
  LayersIcon,
  PlusIcon,
  ChevronRightIcon,
  ChevronDownIcon,
} from 'lucide-vue-next'
import MainLayout from '@/layouts/MainLayout.vue'
import Badge from '@/components/ui/Badge.vue'
import Spinner from '@/components/ui/Spinner.vue'
import Avatar from '@/components/ui/Avatar.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ProgressBar from '@/components/ui/ProgressBar.vue'
import StatusSelect from '@/components/issue/StatusSelect.vue'
import EpicChildList from '@/components/issue/EpicChildList.vue'
import PrioritySelect from '@/components/issue/PrioritySelect.vue'
import CreateIssueModal from '@/components/issue/CreateIssueModal.vue'
import Button from '@/components/ui/Button.vue'
import { fetchProject } from '@/api/projects'
import { fetchIssues, updateIssue } from '@/api/issues'

const route = useRoute()
const slug = computed(() => route.params.slug)
const queryClient = useQueryClient()

// ── Data ─────────────────────────────────────────────────────────────────────

const { data: project, isLoading: loadingProject } = useQuery({
  queryKey: ['project', slug],
  queryFn: () => fetchProject(slug.value),
})

const { data: epicsResult, isLoading: loadingEpics } = useQuery({
  queryKey: ['issues', slug, { type: 'epic', limit: 500 }],
  queryFn: () => fetchIssues(slug.value, { type: 'epic', limit: 500 }),
  enabled: computed(() => !!slug.value),
})

const epics = computed(() => epicsResult.value?.items ?? [])
const isLoading = computed(() => loadingProject.value || loadingEpics.value)

// ── Priority styling ─────────────────────────────────────────────────────────

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

// ── Expand / collapse ────────────────────────────────────────────────────────

const expandedEpics = ref(new Set())

function toggleEpic(epicId) {
  const next = new Set(expandedEpics.value)
  if (next.has(epicId)) {
    next.delete(epicId)
  } else {
    next.add(epicId)
  }
  expandedEpics.value = next
}

// ── Inline status update ─────────────────────────────────────────────────────

const { mutate: updateEpicStatus } = useMutation({
  mutationFn: ({ number, status }) => updateIssue(slug.value, number, { status }),
  onMutate: async ({ number, status }) => {
    const key = ['issues', slug.value, { type: 'epic', limit: 500 }]
    await queryClient.cancelQueries({ queryKey: key })
    const previous = queryClient.getQueryData(key)
    queryClient.setQueryData(key, old => {
      if (!old) return old
      return { ...old, items: old.items.map(i => i.number === number ? { ...i, status } : i) }
    })
    return { previous }
  },
  onError: (_err, _vars, context) => {
    if (context?.previous) {
      queryClient.setQueryData(['issues', slug.value, { type: 'epic', limit: 500 }], context.previous)
    }
  },
  onSettled: () => {
    queryClient.invalidateQueries({ queryKey: ['issues', slug.value] })
  },
})

const { mutate: updateEpicPriority } = useMutation({
  mutationFn: ({ number, priority }) => updateIssue(slug.value, number, { priority }),
  onMutate: async ({ number, priority }) => {
    const key = ['issues', slug.value, { type: 'epic', limit: 500 }]
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
      queryClient.setQueryData(['issues', slug.value, { type: 'epic', limit: 500 }], context.previous)
    }
  },
  onSettled: () => {
    queryClient.invalidateQueries({ queryKey: ['issues', slug.value] })
  },
})

// ── Move task to different epic ──────────────────────────────────────────────

const openMoveDropdown = ref(null)

function toggleMoveDropdown(childId) {
  openMoveDropdown.value = openMoveDropdown.value === childId ? null : childId
}

const { mutate: moveTaskToEpic } = useMutation({
  mutationFn: ({ number, parentId }) => updateIssue(slug.value, number, { parent_id: parentId }),
  onSuccess: () => {
    openMoveDropdown.value = null
    queryClient.invalidateQueries({ queryKey: ['issues', slug.value] })
    queryClient.invalidateQueries({ queryKey: ['issue', slug.value] })
  },
})

// ── Create epic modal ────────────────────────────────────────────────────────

const showCreateEpic = ref(false)

const defaultCreateStatus = computed(() => {
  if (!project.value) return null
  return project.value.archetype === 'support' ? 'open' : 'todo'
})
</script>

<template>
  <MainLayout @create-issue="showCreateEpic = true">
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
            <LayersIcon class="size-4" />
            <span class="text-sm font-medium text-slate-600">Epics</span>
          </div>
        </div>

        <Button size="sm" @click="showCreateEpic = true">
          <PlusIcon class="size-3.5" />
          New epic
        </Button>
      </div>

      <!-- ── Loading ────────────────────────────────────────────────────── -->
      <div v-if="isLoading" class="flex-1 flex items-center justify-center">
        <Spinner />
      </div>

      <!-- ── Empty state ────────────────────────────────────────────────── -->
      <div v-else-if="!epics.length" class="flex-1 flex items-center justify-center">
        <EmptyState
          title="No epics yet"
          description="Epics group related tasks together. Create one to start organizing work."
        >
          <Button size="sm" @click="showCreateEpic = true">
            <PlusIcon class="size-3.5" />
            New epic
          </Button>
        </EmptyState>
      </div>

      <!-- ── Epic list ──────────────────────────────────────────────────── -->
      <div v-else class="flex-1 overflow-y-auto">
        <div class="max-w-5xl mx-auto px-6 py-4 space-y-0">
          <div class="border border-slate-200 rounded-lg overflow-hidden">
            <div v-for="epic in epics" :key="epic.id">
              <!-- Epic row -->
              <div
                class="group flex items-center gap-3 px-4 py-3 hover:bg-slate-50 transition-colors border-l-4 cursor-pointer select-none"
                :class="[
                  priorityBorder(epic.priority),
                  expandedEpics.has(epic.id) ? 'bg-slate-50' : 'bg-white',
                  epic.id !== epics[epics.length - 1]?.id || expandedEpics.has(epic.id) ? 'border-b border-slate-100' : '',
                ]"
                @click="toggleEpic(epic.id)"
              >
                <!-- Expand chevron -->
                <component
                  :is="expandedEpics.has(epic.id) ? ChevronDownIcon : ChevronRightIcon"
                  class="size-4 text-slate-400 flex-shrink-0"
                />

                <!-- Issue number -->
                <router-link
                  :to="`/projects/${slug}/issues/${epic.number}`"
                  class="text-[11px] font-mono text-slate-400 hover:text-blue-600 flex-shrink-0"
                  @click.stop
                >
                  {{ slug.toUpperCase() }}-{{ epic.number }}
                </router-link>

                <!-- Title -->
                <span class="flex-1 min-w-0 text-sm font-medium text-slate-800 truncate">{{ epic.title }}</span>

                <!-- On hold -->
                <span v-if="epic.on_hold" class="flex-shrink-0 text-[10px] font-medium bg-amber-100 text-amber-700 px-1.5 py-0.5 rounded">on hold</span>

                <!-- Progress -->
                <div class="flex-shrink-0 w-28" @click.stop>
                  <ProgressBar :done="epic.child_done_count ?? 0" :total="epic.child_count ?? 0" />
                </div>

                <!-- Status -->
                <div class="flex-shrink-0" @click.stop>
                  <StatusSelect
                    v-if="project"
                    :status="epic.status"
                    :archetype="project.archetype"
                    @update:status="updateEpicStatus({ number: epic.number, status: $event })"
                  />
                </div>

                <!-- Priority -->
                <div class="flex-shrink-0" @click.stop>
                  <PrioritySelect
                    :priority="epic.priority ?? 'none'"
                    @update:priority="updateEpicPriority({ number: epic.number, priority: $event })"
                  />
                </div>

                <!-- Assignee avatars -->
                <div class="flex-shrink-0 flex -space-x-1 w-10 justify-end">
                  <Avatar v-for="a in (epic.assignees ?? []).slice(0, 2)" :key="a.id" :name="a.display_name" :src="a.avatar_url" size="xs" class="ring-1 ring-white" />
                </div>
              </div>

              <!-- Expanded: child tasks -->
              <div
                v-if="expandedEpics.has(epic.id) && project"
                class="px-8 py-4 bg-slate-50/50 border-b border-slate-100"
              >
                <EpicChildList
                  :projectSlug="slug"
                  :epicId="epic.id"
                  :archetype="project.archetype"
                  :childCount="epic.child_count ?? 0"
                  :childDoneCount="epic.child_done_count ?? 0"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

    </div>

    <!-- ── Create epic modal ──────────────────────────────────────────── -->
    <CreateIssueModal
      :open="showCreateEpic"
      :projectSlug="slug"
      :defaultStatus="defaultCreateStatus"
      defaultType="epic"
      @close="showCreateEpic = false"
      @created="showCreateEpic = false"
    />
  </MainLayout>
</template>
