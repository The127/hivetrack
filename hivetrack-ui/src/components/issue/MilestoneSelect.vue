<!--
  MilestoneSelect — single-select dropdown for picking a milestone.

  Props:
    projectSlug — project slug (to fetch milestones)
    modelValue  — current milestone id (string | null)

  Emits:
    update:modelValue — uuid string or null
-->
<script setup>
import { ref, computed, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { fetchMilestones } from '@/api/milestones'
import { FlagIcon, CheckIcon, XIcon } from 'lucide-vue-next'

const props = defineProps({
  projectSlug: { type: String, required: true },
  modelValue: { type: String, default: null },
})

const emit = defineEmits(['update:modelValue'])

const { data: milestonesData } = useQuery({
  queryKey: ['milestones', computed(() => props.projectSlug)],
  queryFn: () => fetchMilestones(props.projectSlug),
  enabled: computed(() => !!props.projectSlug),
})

const milestones = computed(() => (milestonesData.value?.milestones ?? []).filter(m => !m.closed_at))

const currentMilestone = computed(() =>
  milestonesData.value?.milestones?.find(m => m.id === props.modelValue) ?? null
)

// ── Dropdown state ───────────────────────────────────────────────────────────

const open = ref(false)
const root = ref(null)
const triggerBtn = ref(null)
const dropdownEl = ref(null)
const dropdownStyle = ref({})

function positionDropdown() {
  if (!triggerBtn.value) return
  const rect = triggerBtn.value.getBoundingClientRect()
  const spaceBelow = window.innerHeight - rect.bottom
  const goUp = spaceBelow < 200 && rect.top > spaceBelow

  dropdownStyle.value = {
    position: 'fixed',
    left: `${rect.left}px`,
    width: `${Math.max(rect.width, 220)}px`,
    zIndex: 9999,
    ...(goUp
      ? { bottom: `${window.innerHeight - rect.top + 4}px` }
      : { top: `${rect.bottom + 4}px` }),
  }
}

function toggle() {
  open.value = !open.value
  if (open.value) {
    nextTick(() => positionDropdown())
  }
}

function select(id) {
  emit('update:modelValue', id)
  open.value = false
}

function clear() {
  emit('update:modelValue', null)
  open.value = false
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
      <FlagIcon class="size-3" />
      Milestone
    </label>

    <!-- Trigger -->
    <button
      ref="triggerBtn"
      class="w-full flex items-center gap-2 rounded-md border border-slate-200 px-2.5 py-1.5 text-sm text-left cursor-pointer bg-white hover:border-slate-300 hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors"
      @click="toggle"
    >
      <span v-if="currentMilestone" class="flex-1 text-slate-700 truncate">{{ currentMilestone.title }}</span>
      <span v-else class="flex-1 text-slate-400">No milestone</span>
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
          <div class="max-h-52 overflow-y-auto py-1">
            <!-- Clear option -->
            <button
              v-if="modelValue"
              class="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-slate-500 hover:bg-slate-50 cursor-pointer transition-colors"
              @click="clear"
            >
              <XIcon class="size-3.5 text-slate-400 flex-shrink-0" />
              <span>Clear milestone</span>
            </button>

            <!-- Milestone options -->
            <button
              v-for="m in milestones"
              :key="m.id"
              class="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-left cursor-pointer transition-colors"
              :class="modelValue === m.id ? 'bg-slate-50 font-medium text-slate-900' : 'text-slate-700 hover:bg-slate-50'"
              @click="select(m.id)"
            >
              <CheckIcon v-if="modelValue === m.id" class="size-3.5 text-blue-500 flex-shrink-0" />
              <span v-else class="size-3.5 flex-shrink-0" />
              <span class="flex-1 min-w-0 truncate">{{ m.title }}</span>
            </button>

            <p v-if="milestones.length === 0 && !modelValue" class="px-3 py-2 text-xs text-slate-400">
              No open milestones
            </p>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
