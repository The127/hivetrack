export function formatDate(dateStr, { year = false } = {}) {
  if (!dateStr) return "";
  const opts = { month: "short", day: "numeric" };
  if (year) opts.year = "numeric";
  return new Date(dateStr).toLocaleDateString("en-US", opts);
}

export function formatDateRange(startDate, endDate, opts) {
  return `${formatDate(startDate, opts)} – ${formatDate(endDate, opts)}`;
}
