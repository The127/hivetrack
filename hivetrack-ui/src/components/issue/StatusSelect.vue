<!--
  StatusSelect — inline status picker for issue rows.

  Displays the current status icon + label. Clicking opens a dropdown to
  pick a new status. Closes on selection or outside click.

  Props:
    status    — current issue status string
    archetype — 'software' | 'support' (determines available statuses)

  Emits:
    update:status — new status string
-->
<script setup>
import { ref, computed, nextTick, onMounted, onBeforeUnmount } from 'vue'
import {
  CircleIcon,
  CircleDotIcon,
  GitPullRequestIcon,
  CheckCircle2Icon,
  XCircleIcon,
} from 'lucide-vue-next'

const props = defineProps({
  status: { type: String, required: true },
  archetype: { type: String, required: true },
})

const emit = defineEmits(['update:status'])

const STATUS_META = {
  todo:        { label: 'To Do',       scheme: 'gray',   icon: CircleIcon },
  in_progress: { label: 'In Progress', scheme: 'blue',   icon: CircleDotIcon },
  in_review:   { label: 'In Review',   scheme: 'violet', icon: GitPullRequestIcon },
  done:        { label: 'Done',        scheme: 'green',  icon: CheckCircle2Icon },
  cancelled:   { label: 'Cancelled',   scheme: 'gray',   icon: XCircleIcon },
  open:        { label: 'Open',        scheme: 'sky',    icon: CircleIcon },
  resolved:    { label: 'Resolved',    scheme: 'teal',   icon: CheckCircle2Icon },
  closed:      { label: 'Closed',      scheme: 'gray',   icon: XCircleIcon },
}

const SOFTWARE_STATUSES = ['todo', 'in_progress', 'in_review', 'done', 'cancelled']
const SUPPORT_STATUSES = ['open', 'in_progress', 'resolved', 'closed']

const ICON_COLORS = {
  gray:   'text-slate-400',
  blue:   'text-blue-500',
  violet: 'text-violet-500',
  green:  'text-green-500',
  sky:    'text-sky-500',
  teal:   'text-teal-500',
}

const open = ref(false)
const root = ref(null)
const triggerBtn = ref(null)
const dropdownStyle = ref({})

const meta = computed(() => STATUS_META[props.status] ?? { label: props.status, scheme: 'gray', icon: CircleIcon })
const iconColor = computed(() => ICON_COLORS[meta.value.scheme] ?? 'text-slate-400')

const statuses = computed(() => {
  const keys = props.archetype === 'support' ? SUPPORT_STATUSES : SOFTWARE_STATUSES
  return keys.map(s => ({ key: s, ...STATUS_META[s] }))
})

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
  if (key !== props.status) {
    emit('update:status', key)
  }
  open.value = false
}

const dropdownEl = ref(null)

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
      class="flex items-center gap-1 cursor-pointer rounded px-2 py-1.5 -mx-2 -my-1.5 hover:bg-slate-100 transition-colors"
      @click="toggle"
    >
      <component :is="meta.icon" class="size-3.5" :class="iconColor" />
      <span class="text-xs text-slate-500 w-20">{{ meta.label }}</span>
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
          class="bg-white border border-slate-200 rounded-md shadow-md py-1 min-w-36"
        >
          <button
            v-for="s in statuses"
            :key="s.key"
            class="w-full flex items-center gap-2 px-3 py-1.5 text-xs text-left cursor-pointer transition-colors"
            :class="s.key === status ? 'bg-slate-50 font-medium text-slate-900' : 'text-slate-700 hover:bg-slate-50'"
            @click="select(s.key, $event)"
          >
            <component :is="s.icon" class="size-3.5" :class="ICON_COLORS[s.scheme]" />
            {{ s.label }}
          </button>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
