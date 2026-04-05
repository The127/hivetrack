<script setup>
import { ref, useId, computed, nextTick } from "vue";
import { useMutation, useQueryClient } from "@tanstack/vue-query";
import MarkdownEditor from "@/components/ui/MarkdownEditor.vue";
import { MdPreview } from "md-editor-v3";
import { useTheme } from "@/composables/useTheme";
import { updateIssue } from "@/api/issues";

const props = defineProps({
  slug: { type: String, required: true },
  number: { type: Number, required: true },
  description: { type: String, default: null },
});

const { isDark } = useTheme();
const editorTheme = computed(() => (isDark.value ? "dark" : "light"));
const previewId = useId();
const queryClient = useQueryClient();

const editing = ref(false);
const draft = ref("");
const editorEl = ref(null);

function handleDescriptionClick() {
  if (window.getSelection().toString()) return;
  startEditing();
}

function startEditing() {
  draft.value = props.description ?? "";
  editing.value = true;
  nextTick(() => editorEl.value?.focus?.());
}

function cancelEditing() {
  editing.value = false;
}

const { mutate: save } = useMutation({
  mutationFn: (description) =>
    updateIssue(props.slug, props.number, { description }),
  onSuccess: () => {
    editing.value = false;
    queryClient.invalidateQueries({
      queryKey: ["issue", props.slug, props.number],
    });
  },
  onError: () => {
    editing.value = false;
  },
});

function saveDescription() {
  const trimmed = draft.value.trim() || null;
  if (trimmed === (props.description ?? null)) {
    editing.value = false;
    return;
  }
  save(trimmed);
}
</script>

<template>
  <div class="space-y-2">
    <h2 class="text-sm font-medium text-slate-700 dark:text-slate-300">Description</h2>
    <MarkdownEditor
      v-if="editing"
      ref="editorEl"
      v-model="draft"
      placeholder="Add a description…"
      @save="saveDescription"
      @cancel="cancelEditing"
    />
    <div v-else-if="description" @click="handleDescriptionClick" class="cursor-pointer">
      <MdPreview :id="previewId" :model-value="description" language="en-US" :theme="editorTheme" />
    </div>
    <button
      v-else
      class="text-sm text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-400 italic"
      @click="startEditing"
    >
      Add a description…
    </button>
    <div v-if="editing" class="flex items-center gap-2">
      <button
        class="text-xs font-medium text-white bg-blue-600 hover:bg-blue-700 px-3 py-1 rounded"
        @click="saveDescription"
      >
        Save
      </button>
      <button
        class="text-xs text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 px-2 py-1"
        @click="cancelEditing"
      >
        Cancel
      </button>
      <span class="text-xs text-slate-400 dark:text-slate-500 ml-1">Ctrl+Enter to save · Esc to cancel</span>
    </div>
  </div>
</template>
