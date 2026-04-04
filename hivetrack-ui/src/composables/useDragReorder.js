import { ref } from "vue";
import { computeRank } from "@/composables/useRank";

/**
 * Shared drag-and-drop reorder logic for grouped issue lists.
 *
 * @param {import('vue').Ref<Record<string, Array>>} groupedItems — reactive map of group key → items array
 * @param {(movedItem: object, updateData: object) => void} onReorder — called with the moved item and the data to persist
 * @returns {{ isDragging: import('vue').Ref<boolean>, onDragStart: () => void, onDragEnd: () => void, handleDrag: (evt, groupKey, extraUpdates?) => void }}
 */
export function useDragReorder(groupedItems, onReorder) {
  const isDragging = ref(false);

  function onDragStart() {
    isDragging.value = true;
  }

  function onDragEnd() {
    setTimeout(() => {
      isDragging.value = false;
    }, 0);
  }

  function handleDrag(evt, groupKey, extraUpdates = {}) {
    const items = groupedItems.value[groupKey];
    if (!items) return;
    const newIdx = evt.newDraggableIndex;
    const movedItem = items[newIdx];
    const newRank = computeRank(items, newIdx);
    movedItem.rank = newRank;
    Object.assign(movedItem, extraUpdates);
    onReorder(movedItem, { rank: newRank, ...extraUpdates });
  }

  return { isDragging, onDragStart, onDragEnd, handleDrag };
}
