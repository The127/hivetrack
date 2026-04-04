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

export const TERMINAL_STATUSES = {
  software: new Set(['done', 'cancelled']),
  support: new Set(['resolved', 'closed']),
}

export function isTerminalStatus(status, archetype = 'software') {
  return (TERMINAL_STATUSES[archetype] ?? TERMINAL_STATUSES.software).has(status)
}
