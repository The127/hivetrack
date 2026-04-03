<!--
  AddDroneModal — generate a drone join token.

  Two-step flow:
  1. Form: select capabilities + max concurrency → "Generate token"
  2. Display: show token once with copy button + command instructions

  Props:
    open — controls visibility
    slug — project slug

  Events:
    close — modal dismissed (triggers drone list refresh in parent)
-->
<script setup>
import { ref, computed } from 'vue'
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { CopyIcon, CheckIcon } from 'lucide-vue-next'
import Modal from '@/components/ui/Modal.vue'
import Button from '@/components/ui/Button.vue'
import { createDroneToken } from '@/api/drones'

const props = defineProps({
  open: { type: Boolean, required: true },
  slug: { type: String, required: true },
})

const emit = defineEmits(['close'])
const queryClient = useQueryClient()

const capabilities = ref('llm-inference')
const maxConcurrency = ref(1)
const generatedToken = ref(null)
const copied = ref(false)

const { mutate: doGenerate, isPending: generating } = useMutation({
  mutationFn: () =>
    createDroneToken(props.slug, {
      capabilities: capabilities.value.split(',').map((s) => s.trim()).filter(Boolean),
      max_concurrency: maxConcurrency.value,
    }),
  onSuccess: (data) => {
    generatedToken.value = data.token
  },
})

const showingToken = computed(() => generatedToken.value !== null)

function copyToken() {
  if (!generatedToken.value) return
  navigator.clipboard.writeText(generatedToken.value)
  copied.value = true
  setTimeout(() => (copied.value = false), 2000)
}

function close() {
  if (generatedToken.value) {
    queryClient.invalidateQueries({ queryKey: ['drones', props.slug] })
  }
  generatedToken.value = null
  capabilities.value = 'llm-inference'
  maxConcurrency.value = 1
  copied.value = false
  emit('close')
}
</script>

<template>
  <Modal :open="open" :title="showingToken ? 'Drone token generated' : 'Add drone'" @close="close">
    <!-- Step 1: form -->
    <div v-if="!showingToken" class="flex flex-col gap-4">
      <div class="flex flex-col gap-1.5">
        <label class="text-sm font-medium text-slate-700 dark:text-slate-300" for="drone-caps">
          Capabilities
        </label>
        <input
          id="drone-caps"
          v-model="capabilities"
          type="text"
          placeholder="llm-inference"
          class="w-full rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-300 placeholder:text-slate-400 dark:placeholder:text-slate-500 px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
        />
        <p class="text-xs text-slate-400 dark:text-slate-500">Comma-separated list of capabilities this drone will provide.</p>
      </div>
      <div class="flex flex-col gap-1.5">
        <label class="text-sm font-medium text-slate-700 dark:text-slate-300" for="drone-concurrency">
          Max concurrency
        </label>
        <input
          id="drone-concurrency"
          v-model.number="maxConcurrency"
          type="number"
          min="1"
          class="w-20 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-300 px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
        />
      </div>
    </div>

    <!-- Step 2: token display -->
    <div v-else class="flex flex-col gap-4">
      <p class="text-sm text-amber-600 dark:text-amber-400 font-medium">
        This token is shown once. Copy it now.
      </p>
      <div class="flex items-center gap-2">
        <code class="flex-1 text-xs bg-slate-100 dark:bg-slate-800 rounded-md px-3 py-2.5 text-slate-700 dark:text-slate-300 font-mono break-all select-all">
          {{ generatedToken }}
        </code>
        <button
          class="flex-shrink-0 rounded-md p-2 text-slate-500 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
          title="Copy token"
          @click="copyToken"
        >
          <CheckIcon v-if="copied" class="size-4 text-green-500" />
          <CopyIcon v-else class="size-4" />
        </button>
      </div>
      <div class="rounded-md bg-slate-50 dark:bg-slate-800/60 p-3">
        <p class="text-xs text-slate-500 dark:text-slate-400 mb-1.5">Run the drone with:</p>
        <code class="text-xs text-slate-700 dark:text-slate-300 font-mono break-all">
          hivemind-drone --token {{ generatedToken }} --url &lt;HIVEMIND_GRPC_URL&gt;
        </code>
      </div>
    </div>

    <template #footer>
      <Button v-if="!showingToken" variant="secondary" :disabled="generating" @click="close">Cancel</Button>
      <Button v-if="!showingToken" variant="primary" :loading="generating" @click="doGenerate">Generate token</Button>
      <Button v-if="showingToken" variant="primary" @click="close">Done</Button>
    </template>
  </Modal>
</template>
