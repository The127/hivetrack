<!--
  LabelSelect — multi-select dropdown for picking issue labels.

  Props:
    projectSlug — project slug (to fetch labels)
    modelValue  — []{ id, name, color } (current labels)

  Emits:
    update:modelValue — []uuid (selected label IDs)
-->
<script setup>
import { ref, computed, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { fetchLabels } from '@/api/labels'
import { TagIcon, SearchIcon, CheckIcon } from 'lucide-vue-next'
import Badge from '@/components/ui/Badge.vue'

const props = defineProps({
  projectSlug: { type: String, required: true },
  modelValue: { type: Array, default: () => [] },
})

const emit = defineEmits(['update:modelValue'])

const { data: labelsData } = useQuery({
  queryKey: ['labels', computed(() => props.projectSlug)],
  queryFn: () => fetchLabels(props.projectSlug),
  enabled: computed(() => !!props.projectSlug),
})

const labels = computed(() => labelsData.value?.labels ?? [])

// ── Dropdown state ───────────────────────────────────────────────────────────

const open = ref(false)
const root = ref(null)
const triggerBtn = ref(null)
const dropdownEl = ref(null)
const searchInput = ref(null)
const dropdownStyle = ref({})
const search = ref('')

const selectedIds = computed(() => new Set(props.modelValue.map(l => l.id)))

const filteredLabels = computed(() => {
  if (!search.value) return labels.value
  const q = search.value.toLowerCase()
  return labels.value.filter(l => l.name.toLowerCase().includes(q))
})

function positionDropdown() {
  if (!triggerBtn.value) return
  const rect = triggerBtn.value.getBoundingClientRect()
  const spaceBelow = window.innerHeight - rect.bottom
  const goUp = spaceBelow < 240 && rect.top > spaceBelow

  dropdownStyle.value = {
    position: 'fixed',
    left: `${rect.left}px`,
    width: `${Math.max(rect.width, 240)}px`,
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

function toggleLabel(labelId) {
  const current = new Set(selectedIds.value)
  if (current.has(labelId)) {
    current.delete(labelId)
  } else {
    current.add(labelId)
  }
  emit('update:modelValue', [...current])
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
    <label class="text-xs font-medium text-slate-500 flex items-center gap-1">
      <TagIcon class="size-3" />
      Labels
    </label>

    <!-- Trigger -->
    <button
      ref="triggerBtn"
      class="w-full flex items-center gap-2 rounded-md border border-slate-200 px-2.5 py-1.5 text-sm text-left cursor-pointer bg-white hover:border-slate-300 hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors"
      @click="toggle"
    >
      <template v-if="modelValue.length">
        <div class="flex flex-wrap gap-1 flex-1 min-w-0">
          <Badge
            v-for="l in modelValue.slice(0, 3)"
            :key="l.id"
            dot
            :dot-color="l.color"
            compact
          >{{ l.name }}</Badge>
          <span v-if="modelValue.length > 3" class="text-xs text-slate-400">
            +{{ modelValue.length - 3 }}
          </span>
        </div>
      </template>
      <template v-else>
        <span class="flex-1 text-slate-400">No labels</span>
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
          class="bg-white border border-slate-200 rounded-lg shadow-lg overflow-hidden"
        >
          <!-- Search -->
          <div v-if="labels.length > 5" class="p-2 border-b border-slate-100">
            <div class="relative">
              <SearchIcon class="absolute left-2 top-1/2 -translate-y-1/2 size-3.5 text-slate-400" />
              <input
                ref="searchInput"
                v-model="search"
                type="text"
                placeholder="Search labels..."
                class="w-full pl-7 pr-2 py-1 text-sm text-slate-800 placeholder:text-slate-400 bg-slate-50 rounded border-none focus:outline-none"
                @keydown.escape="open = false"
              />
            </div>
          </div>

          <!-- Options -->
          <div class="max-h-52 overflow-y-auto py-1">
            <button
              v-for="label in filteredLabels"
              :key="label.id"
              class="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-left cursor-pointer transition-colors"
              :class="selectedIds.has(label.id) ? 'bg-slate-50 font-medium text-slate-900' : 'text-slate-700 hover:bg-slate-50'"
              @click="toggleLabel(label.id)"
            >
              <CheckIcon v-if="selectedIds.has(label.id)" class="size-3.5 text-blue-500 flex-shrink-0" />
              <span v-else class="size-3.5 flex-shrink-0" />
              <span
                class="size-2.5 rounded-full flex-shrink-0"
                :style="{ backgroundColor: label.color }"
              />
              <span class="flex-1 min-w-0 truncate">{{ label.name }}</span>
            </button>

            <!-- No results -->
            <p v-if="labels.length === 0" class="px-3 py-2 text-xs text-slate-400">
              No labels in this project
            </p>
            <p v-else-if="search && !filteredLabels.length" class="px-3 py-2 text-xs text-slate-400">
              No labels match "{{ search }}"
            </p>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
