<!--
  PrioritySelect — inline priority picker for issue rows and detail views.

  Displays the current priority as a colored badge. Clicking opens a dropdown
  to pick a new priority. Closes on selection or outside click.

  Props:
    priority — current priority string (none|low|medium|high|critical)

  Emits:
    update:priority — new priority string
-->
<script setup>
import { ref, computed, nextTick, onMounted, onBeforeUnmount } from 'vue'
import {
  MinusIcon,
  ArrowDownIcon,
  ArrowRightIcon,
  ArrowUpIcon,
  AlertTriangleIcon,
} from 'lucide-vue-next'

const props = defineProps({
  priority: { type: String, required: true },
})

const emit = defineEmits(['update:priority'])

const PRIORITY_META = {
  none:     { label: 'No priority', scheme: 'gray',   icon: MinusIcon },
  low:      { label: 'Low',         scheme: 'sky',    icon: ArrowDownIcon },
  medium:   { label: 'Medium',      scheme: 'amber',  icon: ArrowRightIcon },
  high:     { label: 'High',        scheme: 'orange', icon: ArrowUpIcon },
  critical: { label: 'Critical',    scheme: 'red',    icon: AlertTriangleIcon },
}

const PRIORITIES = ['none', 'low', 'medium', 'high', 'critical']

const ICON_COLORS = {
  gray:   'text-slate-400',
  sky:    'text-sky-500',
  amber:  'text-amber-500',
  orange: 'text-orange-500',
  red:    'text-red-500',
}

const open = ref(false)
const root = ref(null)
const triggerBtn = ref(null)
const dropdownEl = ref(null)
const dropdownStyle = ref({})

const meta = computed(() => PRIORITY_META[props.priority] ?? PRIORITY_META.none)
const iconColor = computed(() => ICON_COLORS[meta.value.scheme] ?? 'text-slate-400')

const options = computed(() =>
  PRIORITIES.map(p => ({ key: p, ...PRIORITY_META[p] }))
)

function positionDropdown() {
  if (!triggerBtn.value) return
  const rect = triggerBtn.value.getBoundingClientRect()
  dropdownStyle.value = {
    position: 'fixed',
    top: `${rect.bottom + 4}px`,
    left: `${rect.right}px`,
    transform: 'translateX(-100%)',
    zIndex: 9999,
  }
}

function toggle(e) {
  e.preventDefault()
  e.stopPropagation()
  open.value = !open.value
  if (open.value) {
    nextTick(positionDropdown)
  }
}

function select(key, e) {
  e.preventDefault()
  e.stopPropagation()
  if (key !== props.priority) {
    emit('update:priority', key)
  }
  open.value = false
}

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
  <div ref="root" class="flex items-center flex-shrink-0">
    <button
      ref="triggerBtn"
      class="flex items-center gap-1 cursor-pointer rounded px-2 py-1.5 -mx-2 -my-1.5 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
      @click="toggle"
    >
      <component :is="meta.icon" class="size-3.5" :class="iconColor" />
      <span class="text-xs text-slate-500 dark:text-slate-400 w-20">{{ meta.label }}</span>
    </button>

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
          class="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-md shadow-md py-1 min-w-36"
        >
          <button
            v-for="p in options"
            :key="p.key"
            class="w-full flex items-center gap-2 px-3 py-1.5 text-xs text-left cursor-pointer transition-colors"
            :class="p.key === priority ? 'bg-slate-50 dark:bg-slate-700 font-medium text-slate-900 dark:text-slate-100' : 'text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700'"
            @click="select(p.key, $event)"
          >
            <component :is="p.icon" class="size-3.5" :class="ICON_COLORS[p.scheme]" />
            {{ p.label }}
          </button>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
