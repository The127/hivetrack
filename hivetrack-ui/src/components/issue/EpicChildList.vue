<!--
  EpicChildList — shows child tasks of an epic with progress bar.

  Uses the same row layout as the backlog for visual consistency.
  Supports drag-and-drop via VueDraggable with group "epic-tasks".

  Props:
    projectSlug    — project slug
    epicId         — the epic's UUID
    archetype      — project archetype
    childCount     — total children (from epic detail)
    childDoneCount — completed children (from epic detail)

  Provides "Add task" (creates new task with parent_id) and
  "Attach existing" (search unparented tasks, set parent_id).
-->
<script setup>
import { ref, computed, watch } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { VueDraggable } from 'vue-draggable-plus'
import {
  PlusIcon,
  LinkIcon,
  CircleIcon,
  CircleDotIcon,
  CheckCircle2Icon,
  XCircleIcon,
  GitPullRequestIcon,
  SearchIcon,
  XIcon,
} from 'lucide-vue-next'
import Avatar from '@/components/ui/Avatar.vue'
import ProgressBar from '@/components/ui/ProgressBar.vue'
import StatusSelect from '@/components/issue/StatusSelect.vue'
import PrioritySelect from '@/components/issue/PrioritySelect.vue'
import { fetchIssues, createIssue, updateIssue } from '@/api/issues'

const props = defineProps({
  projectSlug: { type: String, required: true },
  epicId: { type: String, required: true },
  archetype: { type: String, required: true },
  childCount: { type: Number, default: 0 },
  childDoneCount: { type: Number, default: 0 },
})

const emit = defineEmits(['task-added', 'task-removed'])

const queryClient = useQueryClient()

// ── Child tasks query ──────────────────────────────────────────────────────

const { data: childrenResult, isLoading } = useQuery({
  queryKey: ['issues', props.projectSlug, { parent_id: props.epicId }],
  queryFn: () => fetchIssues(props.projectSlug, { parent_id: props.epicId, limit: 500 }),
})

// Mutable list for DnD — synced from query data when not dragging
const childList = ref([])
const isDragging = ref(false)

watch(() => childrenResult.value?.items, (items) => {
  if (!isDragging.value) {
    childList.value = items ? [...items] : []
  }
}, { immediate: true })

// ── Priority styling ─────────────────────────────────────────────────────

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

// ── Drag-and-drop ──────────────────────────────────────────────────────────

const { mutate: assignToEpic } = useMutation({
  mutationFn: ({ number, parentId }) => updateIssue(props.projectSlug, number, { parent_id: parentId }),
  onSettled: () => {
    isDragging.value = false
    queryClient.invalidateQueries({ queryKey: ['issues', props.projectSlug] })
    queryClient.invalidateQueries({ queryKey: ['issue', props.projectSlug] })
  },
})

function onDragStart() {
  isDragging.value = true
}

function onDragEnd() {
  setTimeout(() => { isDragging.value = false }, 0)
}

function onTaskAdded(evt) {
  const task = childList.value[evt.newDraggableIndex]
  if (!task) return
  assignToEpic({ number: task.number, parentId: props.epicId })
  emit('task-added', task)
}

function onTaskRemoved() {
  emit('task-removed')
}

// ── Inline create ──────────────────────────────────────────────────────────

const showInlineCreate = ref(false)
const newTaskTitle = ref('')
const inlineInput = ref(null)

const { mutate: doCreateTask, isPending: creating } = useMutation({
  mutationFn: (data) => createIssue(props.projectSlug, data),
  onSuccess: () => {
    newTaskTitle.value = ''
    queryClient.invalidateQueries({ queryKey: ['issues', props.projectSlug] })
    queryClient.invalidateQueries({ queryKey: ['issue', props.projectSlug] })
    inlineInput.value?.focus()
  },
})

function submitCreate() {
  const title = newTaskTitle.value.trim()
  if (!title || creating.value) return
  doCreateTask({ title, type: 'task', parent_id: props.epicId, status: 'todo' })
}

// ── Attach existing ────────────────────────────────────────────────────────

const showAttach = ref(false)
const searchText = ref('')

const { data: searchResult } = useQuery({
  queryKey: computed(() => ['issues', props.projectSlug, { text: searchText.value, type: 'task', noParent: true }]),
  queryFn: () => fetchIssues(props.projectSlug, { type: 'task', text: searchText.value, limit: 20 }),
  enabled: computed(() => showAttach.value && searchText.value.length >= 2),
})

const searchResults = computed(() => {
  const items = searchResult.value?.items ?? []
  return items.filter(i => !i.parent_id)
})

const { mutate: attachTask } = useMutation({
  mutationFn: ({ number }) => updateIssue(props.projectSlug, number, { parent_id: props.epicId }),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['issues', props.projectSlug] })
    queryClient.invalidateQueries({ queryKey: ['issue', props.projectSlug] })
    searchText.value = ''
  },
})

// ── Inline status update ──────────────────────────────────────────────────

const { mutate: updateChildStatus } = useMutation({
  mutationFn: ({ number, status }) => updateIssue(props.projectSlug, number, { status }),
  onMutate: async ({ number, status }) => {
    const key = ['issues', props.projectSlug, { parent_id: props.epicId }]
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
      queryClient.setQueryData(['issues', props.projectSlug, { parent_id: props.epicId }], context.previous)
    }
  },
  onSettled: () => {
    queryClient.invalidateQueries({ queryKey: ['issues', props.projectSlug] })
    queryClient.invalidateQueries({ queryKey: ['issue', props.projectSlug] })
  },
})

const { mutate: updateChildPriority } = useMutation({
  mutationFn: ({ number, priority }) => updateIssue(props.projectSlug, number, { priority }),
  onMutate: async ({ number, priority }) => {
    const key = ['issues', props.projectSlug, { parent_id: props.epicId }]
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
      queryClient.setQueryData(['issues', props.projectSlug, { parent_id: props.epicId }], context.previous)
    }
  },
  onSettled: () => {
    queryClient.invalidateQueries({ queryKey: ['issues', props.projectSlug] })
    queryClient.invalidateQueries({ queryKey: ['issue', props.projectSlug] })
  },
})
</script>

<template>
  <div class="space-y-3">
    <!-- Progress header -->
    <div class="flex items-center justify-between">
      <h3 class="text-sm font-medium text-slate-700">Child tasks</h3>
      <div class="w-40">
        <ProgressBar :done="childDoneCount" :total="childCount" />
      </div>
    </div>

    <!-- Task list (backlog-style rows) -->
    <div v-if="isLoading" class="text-sm text-slate-400">Loading...</div>
    <div v-else class="border border-slate-200 rounded-lg overflow-hidden">
      <VueDraggable
        v-model="childList"
        :group="{ name: 'epic-tasks' }"
        :animation="150"
        ghost-class="opacity-30"
        class="min-h-8"
        @start="onDragStart"
        @end="onDragEnd"
        @add="onTaskAdded"
        @remove="onTaskRemoved"
      >
        <div
          v-for="child in childList"
          :key="child.id"
          class="group flex items-center gap-3 px-6 py-2.5 hover:bg-slate-50 transition-colors cursor-grab active:cursor-grabbing border-l-4 border-b border-slate-100"
          :class="priorityBorder(child.priority)"
        >
          <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
            <router-link
              :to="`/projects/${projectSlug}/issues/${child.number}`"
              class="text-[11px] font-mono text-slate-400 hover:text-blue-600"
              @click.stop
            >
              {{ projectSlug.toUpperCase() }}-{{ child.number }}
            </router-link>
          </div>
          <router-link
            :to="`/projects/${projectSlug}/issues/${child.number}`"
            class="flex-1 min-w-0 text-sm text-slate-800 truncate group-hover:text-slate-900 hover:underline"
            @click.stop
          >
            {{ child.title }}
          </router-link>
          <span v-if="child.on_hold" class="flex-shrink-0 text-[10px] font-medium bg-amber-100 text-amber-700 px-1.5 py-0.5 rounded">on hold</span>
          <div @click.stop>
            <StatusSelect :status="child.status" :archetype="archetype" @update:status="updateChildStatus({ number: child.number, status: $event })" />
          </div>
          <div @click.stop>
            <PrioritySelect :priority="child.priority ?? 'none'" @update:priority="updateChildPriority({ number: child.number, priority: $event })" />
          </div>
          <span v-if="estimateLabel(child.estimate)" class="flex-shrink-0 text-[11px] font-medium text-slate-500 bg-slate-100 px-1.5 py-0.5 rounded w-7 text-center">{{ estimateLabel(child.estimate) }}</span>
          <span v-else class="w-7 flex-shrink-0" />
          <div class="flex-shrink-0 flex -space-x-1 w-10 justify-end">
            <Avatar v-for="a in (child.assignees ?? []).slice(0, 2)" :key="a.id" :name="a.display_name" :src="a.avatar_url" size="xs" class="ring-1 ring-white" />
          </div>
        </div>
      </VueDraggable>

      <!-- Inline create row (matches backlog inline-create style) -->
      <div
        v-if="showInlineCreate"
        class="flex items-center gap-3 px-6 py-2.5 border-b border-slate-100 border-l-4 border-l-blue-400 bg-blue-50/30"
      >
        <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
          <PlusIcon class="size-3 text-blue-400" />
        </div>
        <input
          ref="inlineInput"
          v-model="newTaskTitle"
          type="text"
          placeholder="Task title — Enter to create, Esc to close"
          class="flex-1 min-w-0 text-sm text-slate-800 bg-transparent placeholder:text-slate-400 focus:outline-none"
          @keydown.enter.prevent="submitCreate"
          @keydown.escape="showInlineCreate = false; newTaskTitle = ''"
        />
      </div>
    </div>

    <!-- Attach search -->
    <div v-if="showAttach" class="space-y-2">
      <div class="relative">
        <SearchIcon class="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-slate-400" />
        <input
          v-model="searchText"
          type="text"
          placeholder="Search tasks to attach..."
          class="w-full rounded-md border border-slate-200 pl-8 pr-8 py-1.5 text-sm text-slate-800 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <button
          class="absolute right-2 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 cursor-pointer"
          @click="showAttach = false; searchText = ''"
        >
          <XIcon class="size-3.5" />
        </button>
      </div>
      <div v-if="searchResults.length" class="border border-slate-200 rounded-lg overflow-hidden max-h-48 overflow-y-auto">
        <button
          v-for="task in searchResults"
          :key="task.id"
          class="w-full group flex items-center gap-3 px-6 py-2.5 text-left hover:bg-slate-50 transition-colors cursor-pointer border-b border-slate-100 last:border-b-0"
          @click="attachTask({ number: task.number })"
        >
          <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
            <span class="text-[11px] font-mono text-slate-400">{{ projectSlug.toUpperCase() }}-{{ task.number }}</span>
          </div>
          <span class="flex-1 min-w-0 text-sm text-slate-700 truncate">{{ task.title }}</span>
        </button>
      </div>
      <div v-else-if="searchText.length >= 2" class="text-xs text-slate-400">No unparented tasks found.</div>
    </div>

    <!-- Action buttons -->
    <div class="flex items-center gap-2">
      <button
        v-if="!showInlineCreate"
        class="inline-flex items-center gap-1 text-xs text-slate-500 hover:text-slate-700 cursor-pointer"
        @click="showInlineCreate = true"
      >
        <PlusIcon class="size-3" />
        Add task
      </button>
      <button
        v-if="!showAttach"
        class="inline-flex items-center gap-1 text-xs text-slate-500 hover:text-slate-700 cursor-pointer"
        @click="showAttach = true"
      >
        <LinkIcon class="size-3" />
        Attach existing
      </button>
    </div>
  </div>
</template>
