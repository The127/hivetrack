<!--
  ProjectMilestonesView — manage milestones for a project.

  Shows all milestones with progress bars, target dates, and closed status.
  Allows creating, editing, closing/reopening, and deleting milestones.
-->
<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { PlusIcon, PencilIcon, Trash2Icon, CheckCircle2Icon, RotateCcwIcon } from 'lucide-vue-next'
import MainLayout from '@/layouts/MainLayout.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Spinner from '@/components/ui/Spinner.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import Modal from '@/components/ui/Modal.vue'
import ProgressBar from '@/components/ui/ProgressBar.vue'
import { fetchMilestones, createMilestone, updateMilestone, deleteMilestone } from '@/api/milestones'

const route = useRoute()
const slug = computed(() => route.params.slug)
const queryClient = useQueryClient()

// ── Data ─────────────────────────────────────────────────────────────────────

const { data: milestonesResult, isLoading } = useQuery({
  queryKey: ['milestones', slug],
  queryFn: () => fetchMilestones(slug.value),
  enabled: computed(() => !!slug.value),
})

const milestones = computed(() => milestonesResult.value?.milestones ?? [])

// ── Modals ────────────────────────────────────────────────────────────────────

const showCreateModal = ref(false)
const showEditModal = ref(false)
const editingMilestone = ref(null)
const confirmDeleteId = ref(null)

// Form fields
const formTitle = ref('')
const formDescription = ref('')
const formTargetDate = ref('')

function openCreate() {
  formTitle.value = ''
  formDescription.value = ''
  formTargetDate.value = ''
  showCreateModal.value = true
}

function openEdit(milestone) {
  editingMilestone.value = milestone
  formTitle.value = milestone.title
  formDescription.value = milestone.description ?? ''
  formTargetDate.value = milestone.target_date ? milestone.target_date.slice(0, 10) : ''
  showEditModal.value = true
}

function closeModals() {
  showCreateModal.value = false
  showEditModal.value = false
  editingMilestone.value = null
}

// ── Mutations ─────────────────────────────────────────────────────────────────

const invalidate = () => queryClient.invalidateQueries({ queryKey: ['milestones', slug.value] })

const { mutate: doCreate, isPending: creating } = useMutation({
  mutationFn: (data) => createMilestone(slug.value, data),
  onSuccess: () => { closeModals(); invalidate() },
})

const { mutate: doUpdate, isPending: updating } = useMutation({
  mutationFn: ({ id, data }) => updateMilestone(slug.value, id, data),
  onSuccess: () => { closeModals(); invalidate() },
})

const { mutate: doDelete } = useMutation({
  mutationFn: (id) => deleteMilestone(slug.value, id),
  onSuccess: () => { confirmDeleteId.value = null; invalidate() },
})

// ── Handlers ──────────────────────────────────────────────────────────────────

function submitCreate() {
  if (!formTitle.value.trim()) return
  const data = { title: formTitle.value.trim() }
  if (formDescription.value.trim()) data.description = formDescription.value.trim()
  if (formTargetDate.value) data.target_date = new Date(formTargetDate.value).toISOString()
  doCreate(data)
}

function submitEdit() {
  if (!editingMilestone.value || !formTitle.value.trim()) return
  const data = {
    title: formTitle.value.trim(),
    description: formDescription.value.trim() || null,
    target_date: formTargetDate.value ? new Date(formTargetDate.value).toISOString() : null,
  }
  doUpdate({ id: editingMilestone.value.id, data })
}

function toggleClose(milestone) {
  const close = !milestone.closed_at
  doUpdate({ id: milestone.id, data: { close } })
}

// ── Formatting ────────────────────────────────────────────────────────────────

function formatDate(dateStr) {
  if (!dateStr) return null
  return new Date(dateStr).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}
</script>

<template>
  <MainLayout>
    <div class="max-w-3xl mx-auto px-6 py-8">

      <!-- Header -->
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-lg font-semibold text-slate-900">Milestones</h1>
        <Button size="sm" @click="openCreate">
          <PlusIcon class="size-3.5 mr-1" />
          New milestone
        </Button>
      </div>

      <!-- Loading -->
      <div v-if="isLoading" class="flex justify-center items-center h-32">
        <Spinner class="size-5 text-slate-400" />
      </div>

      <!-- Empty state -->
      <EmptyState
        v-else-if="milestones.length === 0"
        title="No milestones"
        description="Create a milestone to group issues around a target goal or release."
      />

      <!-- Milestone list -->
      <div v-else class="space-y-3">
        <div
          v-for="m in milestones"
          :key="m.id"
          class="rounded-lg border px-4 py-4 space-y-3"
          :class="m.closed_at ? 'border-slate-100 bg-slate-50' : 'border-slate-200 bg-white'"
        >
          <!-- Row: title + actions -->
          <div class="flex items-start justify-between gap-3">
            <div class="flex items-center gap-2 flex-wrap min-w-0">
              <span
                class="text-sm font-medium"
                :class="m.closed_at ? 'text-slate-500 line-through' : 'text-slate-900'"
              >{{ m.title }}</span>
              <Badge v-if="m.closed_at" colorScheme="gray" compact>closed</Badge>
              <span v-if="m.target_date" class="text-xs text-slate-400">
                → {{ formatDate(m.target_date) }}
              </span>
            </div>
            <div class="flex items-center gap-1 flex-shrink-0">
              <button
                class="p-1 rounded text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-colors cursor-pointer"
                :title="m.closed_at ? 'Reopen milestone' : 'Close milestone'"
                @click="toggleClose(m)"
              >
                <RotateCcwIcon v-if="m.closed_at" class="size-3.5" />
                <CheckCircle2Icon v-else class="size-3.5" />
              </button>
              <button
                class="p-1 rounded text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-colors cursor-pointer"
                title="Edit milestone"
                @click="openEdit(m)"
              >
                <PencilIcon class="size-3.5" />
              </button>
              <button
                class="p-1 rounded text-slate-400 hover:text-red-500 hover:bg-red-50 transition-colors cursor-pointer"
                title="Delete milestone"
                @click="confirmDeleteId = m.id"
              >
                <Trash2Icon class="size-3.5" />
              </button>
            </div>
          </div>

          <!-- Description -->
          <p v-if="m.description" class="text-xs text-slate-500">{{ m.description }}</p>

          <!-- Progress -->
          <ProgressBar :done="m.closed_issue_count" :total="m.issue_count" />
        </div>
      </div>
    </div>

    <!-- Create modal -->
    <Modal :open="showCreateModal" title="New milestone" @close="closeModals">
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-slate-700 mb-1">Title</label>
          <input
            v-model="formTitle"
            type="text"
            placeholder="e.g. v1.0 Release"
            class="w-full rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            @keydown.enter="submitCreate"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-700 mb-1">Description <span class="font-normal text-slate-400">(optional)</span></label>
          <textarea
            v-model="formDescription"
            rows="3"
            placeholder="What does this milestone represent?"
            class="w-full rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-700 mb-1">Target date <span class="font-normal text-slate-400">(optional)</span></label>
          <input
            v-model="formTargetDate"
            type="date"
            class="w-full rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>
      </div>
      <template #footer>
        <Button variant="secondary" @click="closeModals">Cancel</Button>
        <Button :loading="creating" :disabled="!formTitle.trim()" @click="submitCreate">Create</Button>
      </template>
    </Modal>

    <!-- Edit modal -->
    <Modal :open="showEditModal" title="Edit milestone" @close="closeModals">
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-slate-700 mb-1">Title</label>
          <input
            v-model="formTitle"
            type="text"
            class="w-full rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            @keydown.enter="submitEdit"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-700 mb-1">Description <span class="font-normal text-slate-400">(optional)</span></label>
          <textarea
            v-model="formDescription"
            rows="3"
            class="w-full rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-700 mb-1">Target date <span class="font-normal text-slate-400">(optional)</span></label>
          <input
            v-model="formTargetDate"
            type="date"
            class="w-full rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>
      </div>
      <template #footer>
        <Button variant="secondary" @click="closeModals">Cancel</Button>
        <Button :loading="updating" :disabled="!formTitle.trim()" @click="submitEdit">Save</Button>
      </template>
    </Modal>

    <!-- Delete confirm modal -->
    <Modal :open="!!confirmDeleteId" title="Delete milestone" @close="confirmDeleteId = null">
      <p class="text-sm text-slate-600">
        Are you sure you want to delete this milestone? Issues assigned to it will not be deleted.
      </p>
      <template #footer>
        <Button variant="secondary" @click="confirmDeleteId = null">Cancel</Button>
        <Button variant="destructive" @click="doDelete(confirmDeleteId)">Delete</Button>
      </template>
    </Modal>
  </MainLayout>
</template>
