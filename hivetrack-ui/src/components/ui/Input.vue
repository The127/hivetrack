<!--
  Input — single-line text field with optional label and error message.

  Supports all standard <input> attributes via $attrs (passed through
  to the underlying <input> element).

  @example
  <Input label="Title" v-model="title" placeholder="Untitled issue" />
  <Input
    label="Email"
    type="email"
    v-model="email"
    :error="errors.email"
    required
  />
  <Input label="Search" v-model="q" prefix-icon="search" />
-->
<script setup>
import { computed, useId } from 'vue'

defineOptions({ inheritAttrs: false })

const props = defineProps({
  /** Field label displayed above the input. If omitted, no label is rendered. */
  label: {
    type: String,
    default: null,
  },
  /** Error message shown below the input. Also applies error styling. */
  error: {
    type: String,
    default: null,
  },
  /** Hint text shown below the input (hidden when error is present). */
  hint: {
    type: String,
    default: null,
  },
  /** v-model binding. */
  modelValue: {
    type: [String, Number],
    default: '',
  },
})

const emit = defineEmits(['update:modelValue'])

const id = useId()

const inputClasses = computed(() => [
  'block w-full rounded-md border px-3 text-sm text-slate-900 placeholder:text-slate-400',
  'h-8',
  'focus:outline-none focus:ring-2 focus:ring-offset-0',
  'disabled:cursor-not-allowed disabled:bg-slate-50 disabled:text-slate-500',
  'transition-colors',
  props.error
    ? 'border-red-400 focus:border-red-400 focus:ring-red-300'
    : 'border-slate-300 focus:border-blue-400 focus:ring-blue-200',
])
</script>

<template>
  <div class="flex flex-col gap-1">
    <label v-if="label" :for="id" class="text-sm font-medium text-slate-700">
      {{ label }}
    </label>
    <input
      :id="id"
      v-bind="$attrs"
      :value="modelValue"
      :class="inputClasses"
      @input="emit('update:modelValue', $event.target.value)"
    />
    <p v-if="error" class="text-xs text-red-600">{{ error }}</p>
    <p v-else-if="hint" class="text-xs text-slate-500">{{ hint }}</p>
  </div>
</template>
