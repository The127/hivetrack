<!--
  CompleteSprintModal — shown when completing a sprint that has open issues.

  Lets the user choose: move open issues to backlog, or to another sprint.

  Props:
    open            Boolean — controls visibility
    openIssueCount  Number  — count of non-terminal issues in the sprint
    sprints         Array   — available sprints to move issues to (planning sprints)

  Events:
    close     — cancel / dismiss
    confirm   — { moveToSprintId: string | null } — null means backlog
-->
<script setup>
import { ref, watch } from 'vue'
import Modal from '@/components/ui/Modal.vue'
import Button from '@/components/ui/Button.vue'
import { InboxIcon, ArrowRightIcon } from 'lucide-vue-next'

const props = defineProps({
  open: { type: Boolean, required: true },
  openIssueCount: { type: Number, default: 0 },
  sprints: { type: Array, default: () => [] },
})

const emit = defineEmits(['close', 'confirm'])

const choice = ref('backlog') // 'backlog' | 'sprint'
const selectedSprintId = ref(null)

watch(() => props.open, (open) => {
  if (open) {
    choice.value = 'backlog'
    selectedSprintId.value = props.sprints[0]?.id ?? null
  }
})

function confirm() {
  const moveToSprintId = choice.value === 'sprint' ? selectedSprintId.value : null
  emit('confirm', { moveToSprintId })
}
</script>

<template>
  <Modal :open="open" title="Complete sprint" @close="emit('close')">
    <div class="space-y-4">
      <p class="text-sm text-slate-600 dark:text-slate-400">
        <span class="font-semibold text-slate-900 dark:text-slate-100">{{ openIssueCount }}</span>
        {{ openIssueCount === 1 ? 'issue is' : 'issues are' }} not yet complete.
        Where should {{ openIssueCount === 1 ? 'it' : 'they' }} go?
      </p>

      <!-- Option: Move to backlog -->
      <label
        class="flex items-start gap-3 rounded-lg border p-3 cursor-pointer transition-colors"
        :class="choice === 'backlog' ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20' : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600'"
      >
        <input
          v-model="choice"
          type="radio"
          value="backlog"
          class="mt-0.5 accent-blue-600"
        />
        <div>
          <div class="flex items-center gap-1.5">
            <InboxIcon class="size-4 text-slate-500 dark:text-slate-400" />
            <span class="text-sm font-medium text-slate-900 dark:text-slate-100">Move to backlog</span>
          </div>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">Issues will be unassigned from any sprint.</p>
        </div>
      </label>

      <!-- Option: Move to another sprint -->
      <label
        class="flex items-start gap-3 rounded-lg border p-3 cursor-pointer transition-colors"
        :class="[
          choice === 'sprint' ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20' : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600',
          !sprints.length ? 'opacity-50 pointer-events-none' : '',
        ]"
      >
        <input
          v-model="choice"
          type="radio"
          value="sprint"
          :disabled="!sprints.length"
          class="mt-0.5 accent-blue-600"
        />
        <div class="flex-1">
          <div class="flex items-center gap-1.5">
            <ArrowRightIcon class="size-4 text-slate-500 dark:text-slate-400" />
            <span class="text-sm font-medium text-slate-900 dark:text-slate-100">Move to another sprint</span>
          </div>
          <p v-if="!sprints.length" class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">No other sprints available.</p>
          <select
            v-else-if="choice === 'sprint'"
            v-model="selectedSprintId"
            class="mt-2 w-full rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 px-2.5 py-1.5 text-sm text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            <option v-for="s in sprints" :key="s.id" :value="s.id">{{ s.name }}</option>
          </select>
        </div>
      </label>
    </div>

    <template #footer>
      <Button variant="secondary" @click="emit('close')">Cancel</Button>
      <Button @click="confirm">Complete sprint</Button>
    </template>
  </Modal>
</template>
