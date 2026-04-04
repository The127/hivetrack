<script setup>
import { ref, computed, nextTick, onMounted, onBeforeUnmount } from "vue";
import {
  LayersIcon,
  ChevronDownIcon,
  CheckIcon,
  SearchIcon,
} from "lucide-vue-next";

const props = defineProps({
  epics: { type: Array, required: true },
  modelValue: { type: String, default: null },
});

const emit = defineEmits(["update:modelValue"]);

const epicFilterOpen = ref(false);
const epicFilterRoot = ref(null);
const epicFilterDropdownEl = ref(null);
const epicFilterTrigger = ref(null);
const epicFilterStyle = ref({});
const epicFilterSearch = ref("");
const epicFilterSearchInput = ref(null);

const selectedEpic = computed(
  () => props.epics.find((e) => e.id === props.modelValue) ?? null,
);

const filteredEpicOptions = computed(() => {
  if (!epicFilterSearch.value) return props.epics;
  const q = epicFilterSearch.value.toLowerCase();
  return props.epics.filter(
    (e) => e.title.toLowerCase().includes(q) || String(e.number).includes(q),
  );
});

function positionEpicFilter() {
  if (!epicFilterTrigger.value) return;
  const rect = epicFilterTrigger.value.getBoundingClientRect();
  epicFilterStyle.value = {
    position: "fixed",
    top: `${rect.bottom + 4}px`,
    left: `${rect.right}px`,
    transform: "translateX(-100%)",
    zIndex: 9999,
  };
}

function toggleEpicFilter() {
  epicFilterOpen.value = !epicFilterOpen.value;
  if (epicFilterOpen.value) {
    epicFilterSearch.value = "";
    nextTick(() => {
      positionEpicFilter();
      epicFilterSearchInput.value?.focus();
    });
  }
}

function selectEpicFilter(epicId) {
  emit("update:modelValue", epicId);
  epicFilterOpen.value = false;
}

function onEpicFilterClickOutside(e) {
  if (!epicFilterOpen.value) return;
  if (epicFilterRoot.value?.contains(e.target)) return;
  if (epicFilterDropdownEl.value?.contains(e.target)) return;
  epicFilterOpen.value = false;
}

onMounted(() =>
  document.addEventListener("pointerdown", onEpicFilterClickOutside, true),
);
onBeforeUnmount(() =>
  document.removeEventListener("pointerdown", onEpicFilterClickOutside, true),
);
</script>

<template>
  <div ref="epicFilterRoot" class="relative">
    <button
      ref="epicFilterTrigger"
      class="flex items-center gap-1.5 cursor-pointer rounded-md border border-slate-200 dark:border-slate-700 px-2.5 h-8 hover:bg-slate-50 dark:hover:bg-slate-800 hover:border-slate-300 dark:hover:border-slate-600 transition-colors"
      @click="toggleEpicFilter"
    >
      <LayersIcon class="size-3.5 text-violet-400 flex-shrink-0" />
      <span
        class="text-sm"
        :class="
          selectedEpic ? 'text-slate-700 dark:text-slate-200 font-medium' : 'text-slate-500 dark:text-slate-400'
        "
      >
        {{ selectedEpic ? selectedEpic.title : "All issues" }}
      </span>
      <ChevronDownIcon class="size-3 text-slate-400 ml-0.5" />
    </button>

    <Teleport to="body">
      <Transition
        enter-active-class="transition-opacity duration-75"
        enter-from-class="opacity-0"
        leave-active-class="transition-opacity duration-75"
        leave-to-class="opacity-0"
      >
        <div
          v-if="epicFilterOpen"
          ref="epicFilterDropdownEl"
          :style="epicFilterStyle"
          class="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg shadow-lg overflow-hidden min-w-52"
        >
          <!-- Search (when many epics) -->
          <div
            v-if="epics.length > 5"
            class="p-2 border-b border-slate-100 dark:border-slate-700"
          >
            <div class="relative">
              <SearchIcon
                class="absolute left-2 top-1/2 -translate-y-1/2 size-3.5 text-slate-400"
              />
              <input
                ref="epicFilterSearchInput"
                v-model="epicFilterSearch"
                type="text"
                placeholder="Search epics..."
                class="w-full pl-7 pr-2 py-1 text-sm text-slate-800 dark:text-slate-200 placeholder:text-slate-400 dark:placeholder:text-slate-500 bg-slate-50 dark:bg-slate-700 rounded border-none focus:outline-none"
                @keydown.escape="epicFilterOpen = false"
              />
            </div>
          </div>

          <div class="max-h-52 overflow-y-auto py-1">
            <!-- All issues option -->
            <button
              class="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-left cursor-pointer transition-colors"
              :class="
                !modelValue
                  ? 'bg-slate-50 dark:bg-slate-700 font-medium text-slate-900 dark:text-slate-100'
                  : 'text-slate-500 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700'
              "
              @click="selectEpicFilter(null)"
            >
              <CheckIcon
                v-if="!modelValue"
                class="size-3.5 text-blue-500 flex-shrink-0"
              />
              <span v-else class="size-3.5 flex-shrink-0" />
              <span>All issues</span>
            </button>

            <button
              v-for="epic in filteredEpicOptions"
              :key="epic.id"
              class="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-left cursor-pointer transition-colors"
              :class="
                epic.id === modelValue
                  ? 'bg-slate-50 dark:bg-slate-700 font-medium text-slate-900 dark:text-slate-100'
                  : 'text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700'
              "
              @click="selectEpicFilter(epic.id)"
            >
              <CheckIcon
                v-if="epic.id === modelValue"
                class="size-3.5 text-blue-500 flex-shrink-0"
              />
              <LayersIcon
                v-else
                class="size-3.5 text-violet-400 flex-shrink-0"
              />
              <span class="flex-1 min-w-0 truncate">{{
                epic.title
              }}</span>
            </button>

            <p
              v-if="epicFilterSearch && !filteredEpicOptions.length"
              class="px-3 py-2 text-xs text-slate-400 dark:text-slate-500"
            >
              No epics match "{{ epicFilterSearch }}"
            </p>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
