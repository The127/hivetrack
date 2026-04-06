<!--
  MarkdownEditor — thin wrapper around MdEditor from md-editor-v3.

  @prop {String} modelValue — markdown source
  @emits update:modelValue
  @emits save   — Ctrl+Enter / ⌘+Enter (via onSave)
  @emits cancel — Escape
-->
<script setup>
import { useId, computed } from "vue";
import { MdEditor } from "md-editor-v3";
import "md-editor-v3/lib/style.css";
import { useTheme } from "@/composables/useTheme";

defineProps({
  modelValue: { type: String, default: "" },
});

const emit = defineEmits(["update:modelValue", "save", "cancel"]);

const editorId = useId();
const { isDark } = useTheme();
const editorTheme = computed(() => isDark.value ? "dark" : "light");

function onKeydown(e) {
  if (e.key === "Escape") emit("cancel");
  if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) emit("save");
}
</script>

<template>
  <div @keydown="onKeydown">
    <MdEditor
      :id="editorId"
      :model-value="modelValue"
      :theme="editorTheme"
      language="en-US"
      @update:model-value="emit('update:modelValue', $event)"
    />
  </div>
</template>
