<!--
  MilestoneTimeline — interactive horizontal timeline for milestones.

  Shows scheduled milestones as draggable markers on a timeline.
  Dragging a marker updates the target_date via API.
  Unscheduled milestones shown below.
  Emits 'edit' with a milestone to open the edit modal.
-->
<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { updateMilestone } from '@/api/milestones'
import { formatDate } from '@/composables/useDate'

const props = defineProps({
  milestones: { type: Array, required: true },
  slug: { type: String, required: true },
})

const emit = defineEmits(['edit'])

const PIXELS_PER_DAY = 10
const ROW_HEIGHT = 56 // px
const HEADER_HEIGHT = 32 // px

// ── Date range ────────────────────────────────────────────────────────────────

const todayDate = new Date()
todayDate.setHours(0, 0, 0, 0)

const scheduledMilestones = computed(() =>
  props.milestones.filter(m => m.target_date)
)

const unscheduledMilestones = computed(() =>
  props.milestones.filter(m => !m.target_date)
)

const timelineStart = computed(() => {
  const dates = scheduledMilestones.value.map(m => new Date(m.target_date))
  const earliest = dates.length ? new Date(Math.min(...dates, todayDate)) : new Date(todayDate)
  const start = new Date(earliest)
  start.setDate(1)
  start.setMonth(start.getMonth() - 1)
  start.setHours(0, 0, 0, 0)
  return start
})

const timelineEnd = computed(() => {
  const dates = scheduledMilestones.value.map(m => new Date(m.target_date))
  const latest = dates.length ? new Date(Math.max(...dates, todayDate)) : new Date(todayDate)
  const end = new Date(latest)
  end.setDate(1)
  end.setMonth(end.getMonth() + 2)
  end.setHours(0, 0, 0, 0)
  return end
})

const totalDays = computed(() =>
  Math.ceil((timelineEnd.value - timelineStart.value) / (1000 * 60 * 60 * 24))
)

const timelineWidth = computed(() => totalDays.value * PIXELS_PER_DAY)

// ── Month headers ─────────────────────────────────────────────────────────────

const monthHeaders = computed(() => {
  const headers = []
  const d = new Date(timelineStart.value)
  while (d < timelineEnd.value) {
    const dayOffset = (d - timelineStart.value) / (1000 * 60 * 60 * 24)
    headers.push({
      label: d.toLocaleDateString('en-US', { month: 'short', year: 'numeric' }),
      left: dayOffset * PIXELS_PER_DAY,
    })
    d.setMonth(d.getMonth() + 1)
  }
  return headers
})

// ── Positioning ───────────────────────────────────────────────────────────────

const todayLeft = computed(() => {
  const dayOffset = (todayDate - timelineStart.value) / (1000 * 60 * 60 * 24)
  return dayOffset * PIXELS_PER_DAY
})

function baseLeft(m) {
  if (!m.target_date) return 0
  const date = new Date(m.target_date)
  const dayOffset = (date - timelineStart.value) / (1000 * 60 * 60 * 24)
  return dayOffset * PIXELS_PER_DAY
}

function getMarkerLeft(m) {
  return dragOffsets.value[m.id] !== undefined ? dragOffsets.value[m.id] : baseLeft(m)
}

// ── Drag ──────────────────────────────────────────────────────────────────────

const scrollContainer = ref(null)
const dragging = ref(null) // { milestoneId, startClientX, startLeft }
const dragOffsets = ref({})
const isDragging = ref(false)

function onMarkerMousedown(event, milestone) {
  event.preventDefault()
  isDragging.value = false
  dragging.value = {
    milestoneId: milestone.id,
    startClientX: event.clientX,
    startLeft: baseLeft(milestone),
  }
}

function onDocumentMousemove(event) {
  if (!dragging.value) return
  const dx = event.clientX - dragging.value.startClientX
  if (Math.abs(dx) > 3) isDragging.value = true
  const newLeft = Math.max(0, Math.min(timelineWidth.value, dragging.value.startLeft + dx))
  dragOffsets.value[dragging.value.milestoneId] = newLeft
}

function onDocumentMouseup(event) {
  if (!dragging.value) return
  const { milestoneId, startClientX, startLeft } = dragging.value
  const dx = event.clientX - startClientX

  if (isDragging.value) {
    const newLeft = Math.max(0, Math.min(timelineWidth.value, startLeft + dx))
    const days = Math.round(newLeft / PIXELS_PER_DAY)
    const newDate = new Date(timelineStart.value)
    newDate.setDate(newDate.getDate() + days)
    doUpdate({ id: milestoneId, data: { target_date: newDate.toISOString() } })
  }

  dragging.value = null
  isDragging.value = false
  delete dragOffsets.value[milestoneId]
}

function onMarkerClick(event, milestone) {
  if (isDragging.value) return
  emit('edit', milestone)
}

onMounted(() => {
  document.addEventListener('mousemove', onDocumentMousemove)
  document.addEventListener('mouseup', onDocumentMouseup)
})

onUnmounted(() => {
  document.removeEventListener('mousemove', onDocumentMousemove)
  document.removeEventListener('mouseup', onDocumentMouseup)
})

// ── Mutation ──────────────────────────────────────────────────────────────────

const queryClient = useQueryClient()

const { mutate: doUpdate } = useMutation({
  mutationFn: ({ id, data }) => updateMilestone(props.slug, id, data),
  onMutate: async ({ id, data }) => {
    await queryClient.cancelQueries({ queryKey: ['milestones', props.slug] })
    const previous = queryClient.getQueryData(['milestones', props.slug])
    queryClient.setQueryData(['milestones', props.slug], old =>
      old?.map(m => m.id === id ? { ...m, ...data } : m) ?? old
    )
    return { previous }
  },
  onError: (_err, _vars, context) => {
    queryClient.setQueryData(['milestones', props.slug], context.previous)
  },
  onSettled: () => {
    queryClient.invalidateQueries({ queryKey: ['milestones', props.slug] })
  },
})

// ── Helpers ───────────────────────────────────────────────────────────────────

function draggedDate(m) {
  const left = getMarkerLeft(m)
  const days = Math.round(left / PIXELS_PER_DAY)
  const d = new Date(timelineStart.value)
  d.setDate(d.getDate() + days)
  return formatDate(d.toISOString())
}

function isOverdue(m) {
  if (!m.target_date || m.closed_at) return false
  return new Date(m.target_date) < todayDate
}

function markerColorClass(m) {
  if (m.closed_at) return 'bg-emerald-100 dark:bg-emerald-900/40 border-emerald-400 text-emerald-700 dark:text-emerald-400'
  if (isOverdue(m)) return 'bg-red-100 dark:bg-red-900/40 border-red-400 text-red-700 dark:text-red-400'
  return 'bg-blue-100 dark:bg-blue-900/40 border-blue-400 text-blue-700 dark:text-blue-400'
}
</script>

<template>
  <!-- Nothing scheduled -->
  <div v-if="scheduledMilestones.length === 0" class="py-6 text-center text-sm text-slate-400">
    No milestones have a target date. Add a target date to see them on the timeline.
  </div>

  <template v-else>
    <!-- Timeline -->
    <div class="flex border border-slate-200 dark:border-slate-700 rounded-lg overflow-hidden">

      <!-- Left: milestone labels (fixed) -->
      <div class="flex-shrink-0 w-44 border-r border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900">
        <!-- Header spacer -->
        <div
          :style="{ height: HEADER_HEIGHT + 'px' }"
          class="border-b border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/50"
        />
        <!-- Rows -->
        <div
          v-for="m in scheduledMilestones"
          :key="m.id"
          :style="{ height: ROW_HEIGHT + 'px' }"
          class="flex items-center px-3 border-b border-slate-100 dark:border-slate-800 last:border-b-0"
        >
          <div class="min-w-0">
            <p
              class="text-xs font-medium truncate"
              :class="m.closed_at ? 'text-slate-400 dark:text-slate-500 line-through' : 'text-slate-700 dark:text-slate-300'"
            >{{ m.title }}</p>
            <p v-if="isOverdue(m)" class="text-xs text-red-500 mt-0.5">overdue</p>
            <p v-else-if="m.closed_at" class="text-xs text-emerald-500 mt-0.5">closed</p>
          </div>
        </div>
      </div>

      <!-- Right: scrollable timeline -->
      <div ref="scrollContainer" class="flex-1 overflow-x-auto">
        <div
          :style="{ minWidth: timelineWidth + 'px', position: 'relative' }"
          :class="dragging ? 'cursor-grabbing select-none' : ''"
        >

          <!-- Month headers -->
          <div
            :style="{ height: HEADER_HEIGHT + 'px' }"
            class="relative border-b border-slate-200 bg-slate-50"
          >
            <div
              v-for="header in monthHeaders"
              :key="header.label"
              :style="{ left: header.left + 'px' }"
              class="absolute inset-y-0 flex items-center pl-2 text-xs text-slate-500 dark:text-slate-400 border-l border-slate-200 dark:border-slate-700"
            >
              {{ header.label }}
            </div>
          </div>

          <!-- Today vertical line -->
          <div
            :style="{ left: todayLeft + 'px', top: 0, bottom: 0, width: '1px', background: '#f87171', position: 'absolute', zIndex: 10, pointerEvents: 'none' }"
          >
            <span class="absolute top-1 left-1 text-xs text-red-400 font-medium whitespace-nowrap">Today</span>
          </div>

          <!-- Milestone rows -->
          <div
            v-for="m in scheduledMilestones"
            :key="m.id"
            :style="{ height: ROW_HEIGHT + 'px' }"
            class="relative border-b border-slate-100 dark:border-slate-800 last:border-b-0"
          >
            <!-- Month separator lines -->
            <div
              v-for="header in monthHeaders"
              :key="header.label"
              :style="{ left: header.left + 'px' }"
              class="absolute inset-y-0 w-px bg-slate-100 dark:bg-slate-800"
            />

            <!-- Horizontal guide line -->
            <div class="absolute inset-x-0 top-1/2 h-px bg-slate-150" style="background: #e2e8f0;" />

            <!-- Marker -->
            <div
              :style="{ left: getMarkerLeft(m) + 'px' }"
              class="absolute top-1/2 -translate-x-1/2 -translate-y-1/2 group"
              :class="dragging?.milestoneId === m.id ? 'z-20' : 'z-10'"
            >
              <!-- Diamond marker -->
              <div
                class="w-4 h-4 rotate-45 border-2 cursor-grab transition-transform group-hover:scale-125"
                :class="[
                  markerColorClass(m),
                  dragging?.milestoneId === m.id ? 'scale-125 cursor-grabbing' : '',
                ]"
                @mousedown="onMarkerMousedown($event, m)"
                @click="onMarkerClick($event, m)"
              />

              <!-- Tooltip on hover / drag label -->
              <div
                class="absolute bottom-6 left-1/2 -translate-x-1/2 whitespace-nowrap bg-slate-800 text-white text-xs rounded px-2 py-1 pointer-events-none"
                :class="dragging?.milestoneId === m.id ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'"
                style="transition: opacity 0.1s;"
              >
                {{ dragging?.milestoneId === m.id ? draggedDate(m) : formatDate(m.target_date) }}
              </div>
            </div>
          </div>

        </div>
      </div>
    </div>

    <!-- Progress legend -->
    <div class="mt-3 flex items-center gap-4 text-xs text-slate-400">
      <div class="flex items-center gap-1.5">
        <div class="w-3 h-3 rotate-45 border-2 border-blue-400 bg-blue-100 dark:bg-blue-900/40" />
        open
      </div>
      <div class="flex items-center gap-1.5">
        <div class="w-3 h-3 rotate-45 border-2 border-red-400 bg-red-100 dark:bg-red-900/40" />
        overdue
      </div>
      <div class="flex items-center gap-1.5">
        <div class="w-3 h-3 rotate-45 border-2 border-emerald-400 bg-emerald-100 dark:bg-emerald-900/40" />
        closed
      </div>
    </div>
  </template>

  <!-- Unscheduled milestones -->
  <div v-if="unscheduledMilestones.length > 0" class="mt-6">
    <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-2">Unscheduled</p>
    <div class="flex flex-wrap gap-2">
      <button
        v-for="m in unscheduledMilestones"
        :key="m.id"
        class="flex items-center gap-1.5 rounded-md border border-dashed border-slate-300 dark:border-slate-600 px-3 py-1.5 text-sm text-slate-500 dark:text-slate-400 hover:border-blue-400 dark:hover:border-blue-500 hover:text-blue-600 dark:hover:text-blue-400 transition-colors cursor-pointer"
        @click="emit('edit', m)"
      >
        <span class="w-2.5 h-2.5 rotate-45 border border-slate-300 inline-block flex-shrink-0" />
        {{ m.title }}
      </button>
    </div>
    <p class="text-xs text-slate-400 mt-2">Click to add a target date</p>
  </div>
</template>
