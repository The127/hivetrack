<!--
  OverviewLabels — project labels list with delete action.
-->
<script setup>
import { computed, ref, nextTick } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { TagIcon, XCircleIcon, PlusIcon } from 'lucide-vue-next'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import { fetchLabels, createLabel, deleteLabel } from '@/api/labels'

const props = defineProps({
  slug: { type: String, required: true },
})

const queryClient = useQueryClient()

const { data: labelsData } = useQuery({
  queryKey: computed(() => ['labels', props.slug]),
  queryFn: () => fetchLabels(props.slug),
  enabled: computed(() => !!props.slug),
})

const labels = computed(() => labelsData.value?.labels ?? [])

const labelToDelete = ref(null)
const showCreateForm = ref(false)
const newLabelName = ref('')
const newLabelColor = ref('#6366f1')
const nameInputRef = ref(null)

const defaultColors = [
  '#ef4444', '#f97316', '#eab308', '#22c55e', '#06b6d4',
  '#3b82f6', '#6366f1', '#a855f7', '#ec4899', '#64748b',
]

function openCreateForm() {
  showCreateForm.value = true
  newLabelName.value = ''
  newLabelColor.value = '#6366f1'
  nextTick(() => nameInputRef.value?.focus())
}

function cancelCreate() {
  showCreateForm.value = false
}

const { mutate: doCreateLabel, isPending: createLabelPending } = useMutation({
  mutationFn: () => createLabel(props.slug, { name: newLabelName.value.trim(), color: newLabelColor.value }),
  onSuccess: () => {
    showCreateForm.value = false
    queryClient.invalidateQueries({ queryKey: ['labels', props.slug] })
  },
})

function submitCreate() {
  if (!newLabelName.value.trim()) return
  doCreateLabel()
}

const { mutate: doDeleteLabel, isPending: deleteLabelPending } = useMutation({
  mutationFn: (labelId) => deleteLabel(props.slug, labelId),
  onSuccess: () => {
    labelToDelete.value = null
    queryClient.invalidateQueries({ queryKey: ['labels', props.slug] })
  },
})
</script>

<template>
  <section>
    <h2 class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-3 flex items-center gap-1.5">
      <TagIcon class="size-4 text-slate-500 dark:text-slate-400" />
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
          class="opacity-0 group-hover:opacity-100 rounded-full p-0.5 text-slate-400 dark:text-slate-500 hover:text-red-500 hover:bg-red-100 dark:hover:bg-red-900/30 transition-all cursor-pointer"
          title="Delete label"
          @click="labelToDelete = label"
        >
          <XCircleIcon class="size-3" />
        </button>
      </div>
      <button
        v-if="!showCreateForm"
        class="flex items-center gap-1 rounded-full border border-dashed border-slate-300 dark:border-slate-600 px-2.5 py-0.5 text-xs text-slate-500 dark:text-slate-400 hover:border-slate-400 dark:hover:border-slate-500 hover:text-slate-700 dark:hover:text-slate-300 transition-colors cursor-pointer"
        @click="openCreateForm"
      >
        <PlusIcon class="size-3" />
        Add label
      </button>
    </div>

    <!-- Create label form -->
    <div v-if="showCreateForm" class="mt-3 flex items-center gap-2">
      <div class="flex gap-1">
        <button
          v-for="c in defaultColors"
          :key="c"
          class="size-5 rounded-full border-2 transition-all cursor-pointer"
          :class="newLabelColor === c ? 'border-slate-900 dark:border-white scale-110' : 'border-transparent hover:scale-110'"
          :style="{ backgroundColor: c }"
          @click="newLabelColor = c"
        />
      </div>
      <input
        ref="nameInputRef"
        v-model="newLabelName"
        type="text"
        placeholder="Label name"
        class="h-7 rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 px-2 text-xs text-slate-900 dark:text-slate-100 placeholder:text-slate-400 focus:outline-none focus:ring-1 focus:ring-blue-500"
        @keydown.enter="submitCreate"
        @keydown.escape="cancelCreate"
      />
      <button
        class="h-7 rounded-md bg-blue-600 hover:bg-blue-700 px-3 text-xs font-medium text-white disabled:opacity-50 cursor-pointer"
        :disabled="!newLabelName.trim() || createLabelPending"
        @click="submitCreate"
      >
        Add
      </button>
      <button
        class="h-7 px-2 text-xs text-slate-500 hover:text-slate-700 dark:hover:text-slate-300 cursor-pointer"
        @click="cancelCreate"
      >
        Cancel
      </button>
    </div>
  </section>

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
</template>
