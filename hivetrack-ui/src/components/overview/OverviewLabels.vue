<!--
  OverviewLabels — project labels list with delete action.
-->
<script setup>
import { computed, ref } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { TagIcon, XCircleIcon } from 'lucide-vue-next'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import { fetchLabels, deleteLabel } from '@/api/labels'

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

const { mutate: doDeleteLabel, isPending: deleteLabelPending } = useMutation({
  mutationFn: (labelId) => deleteLabel(props.slug, labelId),
  onSuccess: () => {
    labelToDelete.value = null
    queryClient.invalidateQueries({ queryKey: ['labels', props.slug] })
  },
})
</script>

<template>
  <section v-if="labels.length">
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
