<!--
  CreateIssueModal — quick-capture form for a new issue.

  Only title is required. Type and priority are optional and default to
  sensible values. Everything else can be set after creation.

  Per design decision #13: "No mandatory fields except title."

  Props:
    open        — controls visibility
    projectSlug — project to create the issue in (optional — shows project picker when omitted)

  Emits:
    close   — close without creating
    created — issue was created; payload: { id, number }
-->
<script setup>
import { ref, watch, computed } from 'vue'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import Modal from '@/components/ui/Modal.vue'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import { createIssue } from '@/api/issues'
import { apiFetch } from '@/composables/useApi'
import { useToast } from '@/composables/useToast'

const props = defineProps({
  open: {
    type: Boolean,
    required: true,
  },
  projectSlug: {
    type: String,
    default: null,
  },
  // When set, the issue is created with this status (landing triaged in backlog,
  // not in the inbox). Used by the backlog view.
  defaultStatus: {
    type: String,
    default: null,
  },
  // When set, the type toggle defaults to this value (e.g. 'epic').
  defaultType: {
    type: String,
    default: 'task',
  },
})

const emit = defineEmits(['close', 'created'])

const queryClient = useQueryClient()
const toast = useToast()

// ── Project picker (only when projectSlug is not provided) ──────────────────

const needsProjectPicker = computed(() => !props.projectSlug)

const { data: projectList } = useQuery({
  queryKey: ['projects'],
  queryFn: () => apiFetch('/api/v1/projects'),
  enabled: needsProjectPicker,
})

const selectedProject = ref('')
const resolvedSlug = computed(() => props.projectSlug ?? selectedProject.value)

// ── Form state ──────────────────────────────────────────────────────────────

const title = ref('')
const type = ref('task')
const priority = ref('none')
const errors = ref({})

// ── Reset when closed ───────────────────────────────────────────────────────

watch(
  () => props.open,
  (open) => {
    if (!open) {
      title.value = ''
      type.value = props.defaultType
      priority.value = 'none'
      selectedProject.value = ''
      errors.value = {}
    }
  },
)

// ── Priority styling ────────────────────────────────────────────────────────

const PRIORITY_ACTIVE = {
  none:     'border-slate-400 bg-slate-100 text-slate-700 ring-1 ring-slate-400',
  low:      'border-sky-500 bg-sky-50 text-sky-700 ring-1 ring-sky-500',
  medium:   'border-amber-500 bg-amber-50 text-amber-700 ring-1 ring-amber-500',
  high:     'border-orange-500 bg-orange-50 text-orange-700 ring-1 ring-orange-500',
  critical: 'border-red-500 bg-red-50 text-red-700 ring-1 ring-red-500',
}

function priorityActiveClass(p) {
  return PRIORITY_ACTIVE[p] ?? PRIORITY_ACTIVE.none
}

// ── Validation ──────────────────────────────────────────────────────────────

function validate() {
  const e = {}
  if (needsProjectPicker.value && !selectedProject.value) e.project = 'Select a project.'
  if (!title.value.trim()) e.title = 'Title is required.'
  errors.value = e
  return Object.keys(e).length === 0
}

// ── Mutation ────────────────────────────────────────────────────────────────

const { mutate, isPending, error: serverError } = useMutation({
  mutationFn: (data) => createIssue(resolvedSlug.value, data),
  onSuccess: (result) => {
    queryClient.invalidateQueries({ queryKey: ['issues', resolvedSlug.value] })
    queryClient.invalidateQueries({ queryKey: ['me', 'issues'] })
    toast.success(`Issue #${result.Number} created`, `/projects/${resolvedSlug.value}/issues/${result.Number}`)
    emit('created', result)
  },
})

const submitError = computed(() => {
  if (!serverError.value) return null
  return serverError.value?.errors?.[0]?.message ?? 'Something went wrong. Please try again.'
})

function submit() {
  if (!validate()) return
  mutate({
    title: title.value.trim(),
    type: type.value,
    priority: priority.value !== 'none' ? priority.value : undefined,
    status: props.defaultStatus ?? undefined,
  })
}
</script>

<template>
  <Modal
    :open="open"
    title="New issue"
    description="Only a title is required — everything else can be set later."
    @close="emit('close')"
  >
    <form class="flex flex-col gap-5" @submit.prevent="submit">
      <!-- Project (only when no projectSlug prop) -->
      <div v-if="needsProjectPicker" class="flex flex-col gap-1.5">
        <label class="text-sm font-medium text-slate-700 dark:text-slate-300" for="create-issue-project">Project</label>
        <select
          id="create-issue-project"
          v-model="selectedProject"
          :class="[
            'w-full rounded-md border px-3 py-2 text-sm bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 cursor-pointer',
            errors.project ? 'border-red-300 dark:border-red-700' : 'border-slate-200 dark:border-slate-700',
          ]"
        >
          <option value="" disabled>Select a project…</option>
          <option
            v-for="p in projectList?.items"
            :key="p.id"
            :value="p.slug"
          >
            {{ p.name }} ({{ p.slug }})
          </option>
        </select>
        <p v-if="errors.project" class="text-sm text-red-600">{{ errors.project }}</p>
      </div>

      <!-- Title -->
      <Input
        label="Title"
        v-model="title"
        placeholder="Short description of the work"
        :error="errors.title"
        autofocus
        required
      />

      <!-- Type (hidden when locked by caller via defaultType) -->
      <div v-if="defaultType === 'task'" class="flex flex-col gap-1.5">
        <span class="text-sm font-medium text-slate-700 dark:text-slate-300">Type</span>
        <div class="flex gap-2">
          <button
            type="button"
            :class="[
              'flex-1 rounded-md border px-3 py-2 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 cursor-pointer',
              type === 'task'
                ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 ring-1 ring-blue-500'
                : 'border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 hover:border-slate-300 dark:hover:border-slate-600 hover:bg-slate-50 dark:hover:bg-slate-800',
            ]"
            @click="type = 'task'"
          >
            Task
          </button>
          <button
            type="button"
            :class="[
              'flex-1 rounded-md border px-3 py-2 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 cursor-pointer',
              type === 'epic'
                ? 'border-violet-500 bg-violet-50 dark:bg-violet-900/30 text-violet-700 dark:text-violet-300 ring-1 ring-violet-500'
                : 'border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 hover:border-slate-300 dark:hover:border-slate-600 hover:bg-slate-50 dark:hover:bg-slate-800',
            ]"
            @click="type = 'epic'"
          >
            Epic
          </button>
        </div>
      </div>

      <!-- Priority -->
      <div class="flex flex-col gap-1.5">
        <span class="text-sm font-medium text-slate-700 dark:text-slate-300">
          Priority <span class="text-slate-400 dark:text-slate-500 font-normal">(optional)</span>
        </span>
        <div class="flex gap-2 flex-wrap">
          <button
            v-for="p in ['none', 'low', 'medium', 'high', 'critical']"
            :key="p"
            type="button"
            :class="[
              'rounded-md border px-3 py-1.5 text-xs font-medium capitalize transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 cursor-pointer',
              priority === p
                ? priorityActiveClass(p)
                : 'border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 hover:border-slate-300 dark:hover:border-slate-600 hover:bg-slate-50 dark:hover:bg-slate-800',
            ]"
            @click="priority = p"
          >
            {{ p === 'none' ? 'No priority' : p }}
          </button>
        </div>
      </div>

      <!-- Server error -->
      <p v-if="submitError" class="text-sm text-red-600">{{ submitError }}</p>
    </form>

    <template #footer>
      <Button variant="secondary" @click="emit('close')">Cancel</Button>
      <Button :loading="isPending" @click="submit">Create issue</Button>
    </template>
  </Modal>
</template>
