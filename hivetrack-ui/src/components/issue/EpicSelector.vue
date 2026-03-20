<!--
  EpicSelector — custom dropdown to pick an epic in a project.

  Props:
    projectSlug — project to fetch epics from
    modelValue  — current parent_id (uuid or null)

  Emits:
    update:modelValue — selected epic ID or null to clear
-->
<script setup>
import { ref, computed, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { fetchIssues } from '@/api/issues'
import { LayersIcon, XIcon, SearchIcon, CheckIcon } from 'lucide-vue-next'

const props = defineProps({
  projectSlug: { type: String, required: true },
  modelValue: { type: String, default: null },
})

const emit = defineEmits(['update:modelValue'])

const { data: epicsResult } = useQuery({
  queryKey: ['issues', props.projectSlug, { type: 'epic' }],
  queryFn: () => fetchIssues(props.projectSlug, { type: 'epic', limit: 200 }),
  enabled: computed(() => !!props.projectSlug),
})

const epics = computed(() => epicsResult.value?.items ?? [])
const selectedEpic = computed(() => epics.value.find(e => e.id === props.modelValue) ?? null)

// ── Dropdown state ───────────────────────────────────────────────────────────

const open = ref(false)
const root = ref(null)
const triggerBtn = ref(null)
const dropdownEl = ref(null)
const searchInput = ref(null)
const dropdownStyle = ref({})
const search = ref('')

const filteredEpics = computed(() => {
  if (!search.value) return epics.value
  const q = search.value.toLowerCase()
  return epics.value.filter(e => e.title.toLowerCase().includes(q) || String(e.number).includes(q))
})

function positionDropdown() {
  if (!triggerBtn.value) return
  const rect = triggerBtn.value.getBoundingClientRect()
  const spaceBelow = window.innerHeight - rect.bottom
  const goUp = spaceBelow < 240 && rect.top > spaceBelow

  dropdownStyle.value = {
    position: 'fixed',
    left: `${rect.left}px`,
    width: `${Math.max(rect.width, 280)}px`,
    zIndex: 9999,
    ...(goUp
      ? { bottom: `${window.innerHeight - rect.top + 4}px` }
      : { top: `${rect.bottom + 4}px` }),
  }
}

function toggle() {
  open.value = !open.value
  if (open.value) {
    search.value = ''
    nextTick(() => {
      positionDropdown()
      searchInput.value?.focus()
    })
  }
}

function select(epicId) {
  if (epicId !== props.modelValue) {
    emit('update:modelValue', epicId)
  }
  open.value = false
}

function clear(e) {
  e.stopPropagation()
  emit('update:modelValue', null)
}

// ── Click outside ────────────────────────────────────────────────────────────

function onClickOutside(e) {
  if (!open.value) return
  if (root.value?.contains(e.target)) return
  if (dropdownEl.value?.contains(e.target)) return
  open.value = false
}

onMounted(() => document.addEventListener('pointerdown', onClickOutside, true))
onBeforeUnmount(() => document.removeEventListener('pointerdown', onClickOutside, true))
</script>

<template>
  <div ref="root" class="flex flex-col gap-1.5">
    <label class="text-xs font-medium text-slate-500 dark:text-slate-400 flex items-center gap-1">
      <LayersIcon class="size-3" />
      Epic
    </label>

    <!-- Trigger -->
    <button
      ref="triggerBtn"
      class="w-full flex items-center gap-2 rounded-md border border-slate-200 dark:border-slate-700 px-2.5 py-1.5 text-sm text-left cursor-pointer bg-white dark:bg-slate-800 hover:border-slate-300 dark:hover:border-slate-600 hover:bg-slate-50 dark:hover:bg-slate-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors"
      @click="toggle"
    >
      <template v-if="selectedEpic">
        <LayersIcon class="size-3.5 text-violet-400 flex-shrink-0" />
        <span class="flex-1 min-w-0 truncate text-slate-700 dark:text-slate-300">{{ selectedEpic.title }}</span>
        <button
          class="flex-shrink-0 p-0.5 rounded hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300 transition-colors cursor-pointer"
          @click="clear"
        >
          <XIcon class="size-3" />
        </button>
      </template>
      <template v-else>
        <span class="flex-1 text-slate-400 dark:text-slate-500">No epic</span>
      </template>
    </button>

    <!-- Dropdown -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-opacity duration-75"
        enter-from-class="opacity-0"
        leave-active-class="transition-opacity duration-75"
        leave-to-class="opacity-0"
      >
        <div
          v-if="open"
          ref="dropdownEl"
          :style="dropdownStyle"
          class="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg shadow-lg overflow-hidden"
        >
          <!-- Search -->
          <div v-if="epics.length > 5" class="p-2 border-b border-slate-100 dark:border-slate-700">
            <div class="relative">
              <SearchIcon class="absolute left-2 top-1/2 -translate-y-1/2 size-3.5 text-slate-400" />
              <input
                ref="searchInput"
                v-model="search"
                type="text"
                placeholder="Search epics..."
                class="w-full pl-7 pr-2 py-1 text-sm text-slate-800 dark:text-slate-200 placeholder:text-slate-400 dark:placeholder:text-slate-500 bg-slate-50 dark:bg-slate-700/50 rounded border-none focus:outline-none"
                @keydown.escape="open = false"
              />
            </div>
          </div>

          <!-- Options -->
          <div class="max-h-52 overflow-y-auto py-1">
            <!-- No epic option -->
            <button
              class="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-left cursor-pointer transition-colors"
              :class="!modelValue ? 'bg-slate-50 dark:bg-slate-700 font-medium text-slate-900 dark:text-slate-100' : 'text-slate-500 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700'"
              @click="select(null)"
            >
              <CheckIcon v-if="!modelValue" class="size-3.5 text-blue-500 flex-shrink-0" />
              <span v-else class="size-3.5 flex-shrink-0" />
              <span>No epic</span>
            </button>

            <!-- Epic options -->
            <button
              v-for="epic in filteredEpics"
              :key="epic.id"
              class="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-left cursor-pointer transition-colors"
              :class="epic.id === modelValue ? 'bg-slate-50 dark:bg-slate-700 font-medium text-slate-900 dark:text-slate-100' : 'text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700'"
              @click="select(epic.id)"
            >
              <CheckIcon v-if="epic.id === modelValue" class="size-3.5 text-blue-500 flex-shrink-0" />
              <LayersIcon v-else class="size-3.5 text-violet-400 flex-shrink-0" />
              <span class="flex-1 min-w-0 truncate">{{ epic.title }}</span>
              <span class="text-[11px] font-mono text-slate-400 dark:text-slate-500 flex-shrink-0">{{ epic.number }}</span>
            </button>

            <!-- No results -->
            <p v-if="search && !filteredEpics.length" class="px-3 py-2 text-xs text-slate-400">
              No epics match "{{ search }}"
            </p>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
