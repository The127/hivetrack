<!--
  SplitIssueModal — split a task into two or more smaller tasks.

  The original issue is cancelled and linked to each new issue via "relates_to".
  Pre-populates from checklist items when present; otherwise starts with two empty inputs.

  Props:
    open  — controls visibility
    issue — IssueDetail of the task to split

  Emits:
    close  — close without splitting
    split  — split completed; payload: { newIssues: [{ id, number }] }
-->
<script setup>
import { ref, watch, computed } from 'vue'
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'
import { PlusIcon, XIcon } from 'lucide-vue-next'
import Modal from '@/components/ui/Modal.vue'
import Button from '@/components/ui/Button.vue'
import { splitIssue } from '@/api/issues'

const props = defineProps({
  open: {
    type: Boolean,
    required: true,
  },
  issue: {
    type: Object,
    required: true,
  },
  projectSlug: {
    type: String,
    required: true,
  },
})

const emit = defineEmits(['close', 'split'])

const queryClient = useQueryClient()
const router = useRouter()

const titles = ref([])

// Pre-populate from checklist if available, otherwise two empty inputs
watch(
  () => props.open,
  (open) => {
    if (!open) return
    const checklist = props.issue?.checklist ?? []
    if (checklist.length >= 2) {
      titles.value = checklist.map((item) => item.text)
    } else {
      titles.value = ['', '']
    }
    error.value = null
  },
  { immediate: true },
)

const error = ref(null)

const canSubmit = computed(() => titles.value.filter((t) => t.trim()).length >= 2)

function addTitle() {
  titles.value = [...titles.value, '']
}

function removeTitle(index) {
  if (titles.value.length <= 2) return
  titles.value = titles.value.filter((_, i) => i !== index)
}

const { mutate, isPending } = useMutation({
  mutationFn: () =>
    splitIssue(
      props.projectSlug,
      props.issue.number,
      titles.value.map((t) => t.trim()).filter(Boolean),
    ),
  onSuccess: (result) => {
    queryClient.invalidateQueries({ queryKey: ['issue', props.projectSlug, props.issue.number] })
    queryClient.invalidateQueries({ queryKey: ['issues', props.projectSlug] })
    emit('split', result)
    if (result.new_issues?.length) {
      router.push(`/projects/${props.projectSlug}/issues/${result.new_issues[0].number}`)
    }
  },
  onError: (err) => {
    error.value = err?.errors?.[0]?.message ?? 'Something went wrong. Please try again.'
  },
})

function submit() {
  error.value = null
  const filled = titles.value.filter((t) => t.trim())
  if (filled.length < 2) {
    error.value = 'At least 2 titles are required.'
    return
  }
  mutate()
}
</script>

<template>
  <Modal
    :open="open"
    title="Split issue"
    description="The original issue will be cancelled and linked to the new issues."
    @close="emit('close')"
  >
    <div class="flex flex-col gap-4">
      <div class="space-y-2">
        <div
          v-for="(_, index) in titles"
          :key="index"
          class="flex items-center gap-2"
        >
          <input
            v-model="titles[index]"
            type="text"
            :placeholder="`New issue ${index + 1} title`"
            class="flex-1 rounded-md border border-slate-200 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100 px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
            @keydown.enter.prevent="submit"
          />
          <button
            v-if="titles.length > 2"
            type="button"
            class="flex-shrink-0 rounded-md p-1.5 text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors cursor-pointer"
            @click="removeTitle(index)"
          >
            <XIcon class="size-4" />
          </button>
        </div>
      </div>

      <button
        type="button"
        class="inline-flex items-center gap-1.5 text-sm text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 transition-colors cursor-pointer w-fit"
        @click="addTitle"
      >
        <PlusIcon class="size-4" />
        Add another issue
      </button>

      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
    </div>

    <template #footer>
      <Button variant="secondary" @click="emit('close')">Cancel</Button>
      <Button :loading="isPending" :disabled="!canSubmit" @click="submit">Split issue</Button>
    </template>
  </Modal>
</template>
