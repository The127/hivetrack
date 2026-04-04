<script setup>
import { ref, watch } from "vue";

const props = defineProps({
  mode: { type: String, required: true, validator: (v) => ["create", "edit"].includes(v) },
  initialData: { type: Object, default: () => ({ name: "", start_date: "", end_date: "", goal: "" }) },
  loading: { type: Boolean, default: false },
  error: { type: String, default: "" },
});

const emit = defineEmits(["submit", "cancel"]);

const form = ref({ ...props.initialData });

watch(
  () => props.initialData,
  (val) => {
    form.value = { ...val };
  },
);

function handleSubmit() {
  emit("submit", { ...form.value });
}

function handleCancel() {
  emit("cancel");
}
</script>

<template>
  <div class="max-w-lg space-y-3">
    <div class="text-sm font-medium text-slate-700 dark:text-slate-300">
      {{ mode === "create" ? "New sprint" : "Edit sprint" }}
    </div>

    <input
      v-model="form.name"
      type="text"
      placeholder="Sprint name (required)"
      class="w-full rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 px-3 py-1.5 text-sm text-slate-900 dark:text-slate-100 placeholder:text-slate-400 dark:placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
    />

    <div class="flex gap-3">
      <div class="flex-1">
        <label class="text-xs text-slate-500 dark:text-slate-400 mb-1 block">Start date</label>
        <input
          v-model="form.start_date"
          type="date"
          class="w-full rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 px-3 py-1.5 text-sm text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
      </div>
      <div class="flex-1">
        <label class="text-xs text-slate-500 dark:text-slate-400 mb-1 block">End date</label>
        <input
          v-model="form.end_date"
          type="date"
          class="w-full rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 px-3 py-1.5 text-sm text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
      </div>
    </div>

    <input
      v-model="form.goal"
      type="text"
      placeholder="Sprint goal (optional)"
      class="w-full rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 px-3 py-1.5 text-sm text-slate-900 dark:text-slate-100 placeholder:text-slate-400 dark:placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
    />

    <div v-if="error" class="text-xs text-red-600">
      {{ error }}
    </div>

    <div class="flex items-center gap-2">
      <button
        class="inline-flex items-center gap-1.5 rounded-md bg-blue-600 px-3 h-7 text-xs font-medium text-white hover:bg-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors cursor-pointer disabled:opacity-50"
        :disabled="loading"
        @click="handleSubmit"
      >
        {{ mode === "create" ? "Create sprint" : "Save" }}
      </button>
      <button
        class="inline-flex items-center rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 px-3 h-7 text-xs font-medium text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 focus-visible:outline-none transition-colors cursor-pointer"
        @click="handleCancel"
      >
        Cancel
      </button>
    </div>
  </div>
</template>
