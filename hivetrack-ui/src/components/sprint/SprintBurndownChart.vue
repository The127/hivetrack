<script setup>
import { computed } from 'vue'

const props = defineProps({
  points: {
    type: Array,
    required: true,
  },
  total: {
    type: Number,
    required: true,
  },
  startDate: {
    type: String,
    required: true,
  },
  endDate: {
    type: String,
    required: true,
  },
})

const W = 480
const H = 160
const PAD = { top: 12, right: 12, bottom: 28, left: 32 }
const chartW = W - PAD.left - PAD.right
const chartH = H - PAD.top - PAD.bottom

// Build date axis from startDate to endDate (inclusive)
const allDates = computed(() => {
  const dates = []
  const start = new Date(props.startDate)
  const end = new Date(props.endDate)
  for (let d = new Date(start); d <= end; d.setDate(d.getDate() + 1)) {
    dates.push(new Date(d).toISOString().slice(0, 10))
  }
  return dates
})

const totalDays = computed(() => Math.max(allDates.value.length - 1, 1))

// Map points by date string
const pointsByDate = computed(() => {
  const m = {}
  for (const p of props.points) {
    m[new Date(p.date).toISOString().slice(0, 10)] = p.remaining
  }
  return m
})

// X coordinate for a date index
function xFor(idx) {
  return PAD.left + (idx / totalDays.value) * chartW
}

// Y coordinate for a remaining count (0 = bottom, total = top)
function yFor(remaining) {
  if (props.total === 0) return PAD.top + chartH
  return PAD.top + chartH - (remaining / props.total) * chartH
}

// Ideal burndown line: (startDate, total) → (endDate, 0)
const idealPath = computed(() => {
  const x0 = xFor(0)
  const y0 = yFor(props.total)
  const x1 = xFor(totalDays.value)
  const y1 = yFor(0)
  return `M${x0},${y0} L${x1},${y1}`
})

// Actual burndown line from available data points
const actualPath = computed(() => {
  const pts = allDates.value
    .map((date, idx) => {
      const rem = pointsByDate.value[date]
      if (rem === undefined) return null
      return { x: xFor(idx), y: yFor(rem) }
    })
    .filter(Boolean)

  if (pts.length === 0) return ''
  return pts.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x},${p.y}`).join(' ')
})

// Today marker (vertical line), if within range
const todayX = computed(() => {
  const today = new Date().toISOString().slice(0, 10)
  const idx = allDates.value.indexOf(today)
  if (idx === -1) return null
  return xFor(idx)
})

// Y-axis ticks (4 ticks: 0, total/3, 2*total/3, total)
const yTicks = computed(() => {
  if (props.total === 0) return [0]
  const step = Math.ceil(props.total / 3)
  const ticks = []
  for (let v = 0; v <= props.total; v += step) ticks.push(v)
  if (ticks[ticks.length - 1] !== props.total) ticks.push(props.total)
  return ticks
})

// X-axis labels: show start and end only to keep it clean
const xLabels = computed(() => {
  const dates = allDates.value
  if (dates.length === 0) return []
  const labels = [{ idx: 0, label: formatShort(dates[0]) }]
  if (dates.length > 1) {
    labels.push({ idx: dates.length - 1, label: formatShort(dates[dates.length - 1]) })
  }
  return labels
})

function formatShort(dateStr) {
  return new Date(dateStr).toLocaleDateString('en-US', { month: 'short', day: 'numeric', timeZone: 'UTC' })
}
</script>

<template>
  <svg
    :viewBox="`0 0 ${W} ${H}`"
    :width="W"
    :height="H"
    class="w-full h-auto"
    aria-label="Sprint burndown chart"
  >
    <!-- Y-axis ticks and grid lines -->
    <g v-for="v in yTicks" :key="v">
      <line
        :x1="PAD.left"
        :y1="yFor(v)"
        :x2="PAD.left + chartW"
        :y2="yFor(v)"
        stroke="#e2e8f0"
        stroke-width="1"
      />
      <text
        :x="PAD.left - 4"
        :y="yFor(v)"
        text-anchor="end"
        dominant-baseline="middle"
        font-size="9"
        fill="#94a3b8"
      >{{ v }}</text>
    </g>

    <!-- X-axis labels -->
    <g v-for="lbl in xLabels" :key="lbl.idx">
      <text
        :x="xFor(lbl.idx)"
        :y="PAD.top + chartH + 14"
        text-anchor="middle"
        font-size="9"
        fill="#94a3b8"
      >{{ lbl.label }}</text>
    </g>

    <!-- Today marker -->
    <line
      v-if="todayX !== null"
      :x1="todayX"
      :y1="PAD.top"
      :x2="todayX"
      :y2="PAD.top + chartH"
      stroke="#94a3b8"
      stroke-width="1"
      stroke-dasharray="3,3"
    />

    <!-- Ideal burndown line (dashed gray) -->
    <path
      v-if="idealPath"
      :d="idealPath"
      fill="none"
      stroke="#cbd5e1"
      stroke-width="1.5"
      stroke-dasharray="4,3"
    />

    <!-- Actual burndown line (solid blue) -->
    <path
      v-if="actualPath"
      :d="actualPath"
      fill="none"
      stroke="#3b82f6"
      stroke-width="2"
      stroke-linejoin="round"
      stroke-linecap="round"
    />
  </svg>
</template>
