<!--
  ProgressBar — shows progress as a bar + "done/total" label.

  Props:
    done  — number of completed items
    total — total number of items

  @example
  <ProgressBar :done="3" :total="7" />
-->
<script setup>
import { computed } from 'vue'

const props = defineProps({
  done: { type: Number, required: true },
  total: { type: Number, required: true },
})

const percent = computed(() => (props.total > 0 ? Math.round((props.done / props.total) * 100) : 0))
</script>

<template>
  <div class="flex items-center gap-2">
    <div class="flex-1 h-1.5 rounded-full bg-slate-200 overflow-hidden">
      <div
        class="h-full rounded-full transition-all duration-300"
        :class="percent === 100 ? 'bg-emerald-500' : 'bg-blue-500'"
        :style="{ width: `${percent}%` }"
      />
    </div>
    <span class="text-xs text-slate-500 tabular-nums flex-shrink-0">{{ done }}/{{ total }}</span>
  </div>
</template>
