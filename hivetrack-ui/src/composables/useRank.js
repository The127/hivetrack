import { generateKeyBetween } from "fractional-indexing";

export function computeRank(items, newIdx) {
  const prev = newIdx > 0 ? (items[newIdx - 1]?.rank ?? null) : null;
  const next =
    newIdx < items.length - 1 ? (items[newIdx + 1]?.rank ?? null) : null;
  try {
    return generateKeyBetween(prev, next);
  } catch {
    return Date.now().toString(36) + Math.random().toString(36).slice(2, 6);
  }
}
