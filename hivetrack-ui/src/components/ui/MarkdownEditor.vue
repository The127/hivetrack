<!--
  MarkdownEditor — thin wrapper around MdEditor from md-editor-v3.

  @prop {String} modelValue — markdown source
  @emits update:modelValue
  @emits save   — Ctrl+Enter (via onSave)
  @emits cancel — Escape
-->
<script setup>
import { useId } from "vue";
import { MdEditor } from "md-editor-v3";
import "md-editor-v3/lib/style.css";

defineProps({
  modelValue: { type: String, default: "" },
});

const emit = defineEmits(["update:modelValue", "save", "cancel"]);

const editorId = useId();

function onKeydown(e) {
  if (e.key === "Escape") emit("cancel");
  if (e.key === "Enter" && e.ctrlKey) emit("save");
}
</script>

<template>
  <div @keydown="onKeydown">
    <MdEditor
      :id="editorId"
      :model-value="modelValue"
      language="en-US"
      @update:model-value="emit('update:modelValue', $event)"
    />
  </div>
</template>
