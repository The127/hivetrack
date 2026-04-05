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
import { ref, computed } from "vue";
import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { CopyIcon, CheckIcon } from "lucide-vue-next";
import Modal from "@/components/ui/Modal.vue";
import Button from "@/components/ui/Button.vue";
import { createDroneToken } from "@/api/drones";

const props = defineProps({
  open: { type: Boolean, required: true },
  slug: { type: String, required: true },
});

const emit = defineEmits(["close"]);
const queryClient = useQueryClient();

const AVAILABLE_CAPS = [
  {
    value: "refinement",
    label: "Refinement",
    description: "Refine issue descriptions into structured user stories",
  },
  {
    value: "implementation",
    label: "Implementation",
    description: "Write code to implement issues, create commits and PRs",
  },
  {
    value: "code-review",
    label: "Code Review",
    description: "Review pull requests and provide feedback",
  },
];

const selectedCaps = ref(new Set());
const maxConcurrency = ref(1);
const generatedToken = ref(null);
const copied = ref(false);

function toggleCap(cap) {
  if (selectedCaps.value.has(cap)) {
    selectedCaps.value.delete(cap);
  } else {
    selectedCaps.value.add(cap);
  }
}

const hasAnyCap = computed(() => selectedCaps.value.size > 0);

const { mutate: doGenerate, isPending: generating } = useMutation({
  mutationFn: () =>
    createDroneToken(props.slug, {
      capabilities: [...selectedCaps.value],
      max_concurrency: maxConcurrency.value,
    }),
  onSuccess: (data) => {
    generatedToken.value = data.token;
  },
});

const showingToken = computed(() => generatedToken.value !== null);

function copyToken() {
  if (!generatedToken.value) return;
  navigator.clipboard.writeText(generatedToken.value);
  copied.value = true;
  setTimeout(() => (copied.value = false), 2000);
}

function close() {
  if (generatedToken.value) {
    queryClient.invalidateQueries({ queryKey: ["drones", props.slug] });
  }
  generatedToken.value = null;
  selectedCaps.value = new Set();
  maxConcurrency.value = 1;
  copied.value = false;
  emit("close");
}
</script>

<template>
  <Modal
    :open="open"
    :title="showingToken ? 'Drone token generated' : 'Add drone'"
    @close="close"
  >
    <!-- Step 1: form -->
    <div v-if="!showingToken" class="flex flex-col gap-4">
      <div class="flex flex-col gap-1.5">
        <span class="text-sm font-medium text-slate-700 dark:text-slate-300">
          Capabilities
        </span>
        <div class="flex flex-col gap-2">
          <label
            v-for="cap in AVAILABLE_CAPS"
            :key="cap.value"
            class="flex items-start gap-2.5 rounded-md border px-3 py-2.5 cursor-pointer transition-colors"
            :class="
              selectedCaps.has(cap.value)
                ? 'border-blue-300 dark:border-blue-700 bg-blue-50 dark:bg-blue-900/20'
                : 'border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 hover:border-slate-300 dark:hover:border-slate-600'
            "
          >
            <input
              type="checkbox"
              :checked="selectedCaps.has(cap.value)"
              class="mt-0.5 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500 cursor-pointer"
              @change="toggleCap(cap.value)"
            />
            <div>
              <span
                class="text-sm font-medium text-slate-700 dark:text-slate-300"
                >{{ cap.label }}</span
              >
              <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">
                {{ cap.description }}
              </p>
            </div>
          </label>
        </div>
      </div>
      <div class="flex flex-col gap-1.5">
        <label
          class="text-sm font-medium text-slate-700 dark:text-slate-300"
          for="drone-concurrency"
        >
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
        <code
          class="flex-1 text-xs bg-slate-100 dark:bg-slate-800 rounded-md px-3 py-2.5 text-slate-700 dark:text-slate-300 font-mono break-all select-all"
        >
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
      <div
        class="rounded-md bg-slate-50 dark:bg-slate-800/60 p-3 flex flex-col gap-2"
      >
        <p class="text-xs text-slate-500 dark:text-slate-400">
          Run the drone with one of:
        </p>
        <code
          class="text-xs text-slate-700 dark:text-slate-300 font-mono break-all block"
          >hivemind-drone claude --hivemind-url &lt;HIVEMIND_GRPC_URL&gt;
          --token {{ generatedToken }}</code
        >
        <code
          class="text-xs text-slate-700 dark:text-slate-300 font-mono break-all block"
          >hivemind-drone ollama --hivemind-url &lt;HIVEMIND_GRPC_URL&gt;
          --token {{ generatedToken }}</code
        >
        <code
          class="text-xs text-slate-700 dark:text-slate-300 font-mono break-all block"
          >hivemind-drone local --hivemind-url &lt;HIVEMIND_GRPC_URL&gt; --token
          {{ generatedToken }}</code
        >
      </div>
    </div>

    <template #footer>
      <Button
        v-if="!showingToken"
        variant="secondary"
        :disabled="generating"
        @click="close"
        >Cancel</Button
      >
      <Button
        v-if="!showingToken"
        variant="primary"
        :loading="generating"
        :disabled="!hasAnyCap"
        @click="doGenerate"
        >Generate token</Button
      >
      <Button v-if="showingToken" variant="primary" @click="close">Done</Button>
    </template>
  </Modal>
</template>
