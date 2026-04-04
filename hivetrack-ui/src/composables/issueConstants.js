export const PRIORITY_BORDER = {
  none: 'border-l-slate-200',
  low: 'border-l-sky-400',
  medium: 'border-l-amber-400',
  high: 'border-l-orange-500',
  critical: 'border-l-red-500',
}

export const ESTIMATE_LABEL = {
  none: null,
  xs: 'XS',
  s: 'S',
  m: 'M',
  l: 'L',
  xl: 'XL',
}

export function priorityBorder(priority) {
  return PRIORITY_BORDER[priority] ?? 'border-l-slate-200'
}

export function estimateLabel(estimate) {
  return ESTIMATE_LABEL[estimate] ?? null
}

// ── Status config ────────────────────────────────────────────────────────────

import {
  CircleIcon,
  CircleDotIcon,
  GitPullRequestIcon,
  CheckCircle2Icon,
  XCircleIcon,
} from 'lucide-vue-next'

export const STATUS_META = {
  todo:        { label: 'To Do',       scheme: 'gray',   icon: CircleIcon },
  in_progress: { label: 'In Progress', scheme: 'blue',   icon: CircleDotIcon },
  in_review:   { label: 'In Review',   scheme: 'violet', icon: GitPullRequestIcon },
  done:        { label: 'Done',        scheme: 'green',  icon: CheckCircle2Icon },
  cancelled:   { label: 'Cancelled',   scheme: 'gray',   icon: XCircleIcon },
  open:        { label: 'Open',        scheme: 'sky',    icon: CircleIcon },
  resolved:    { label: 'Resolved',    scheme: 'teal',   icon: CheckCircle2Icon },
  closed:      { label: 'Closed',      scheme: 'gray',   icon: XCircleIcon },
}

export const SOFTWARE_STATUSES = ['todo', 'in_progress', 'in_review', 'done', 'cancelled']
export const SUPPORT_STATUSES = ['open', 'in_progress', 'resolved', 'closed']

export function statusLabel(status) {
  return STATUS_META[status]?.label ?? status
}

export function statusScheme(status) {
  return STATUS_META[status]?.scheme ?? 'gray'
}

export function statusColumns(archetype) {
  const keys = archetype === 'support' ? SUPPORT_STATUSES : SOFTWARE_STATUSES
  return keys.map(key => ({ key, ...STATUS_META[key] }))
}

// ── Terminal statuses ────────────────────────────────────────────────────────

export const TERMINAL_STATUSES = {
  software: new Set(['done', 'cancelled']),
  support: new Set(['resolved', 'closed']),
}

export function isTerminalStatus(status, archetype = 'software') {
  return (TERMINAL_STATUSES[archetype] ?? TERMINAL_STATUSES.software).has(status)
}
