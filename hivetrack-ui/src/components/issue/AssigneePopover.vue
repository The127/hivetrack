<!--
  AssigneePopover — avatar stack with a hover popover listing all assignee names.

  Renders the first 2 avatars and a +N badge when there are more. On hover,
  shows a popover (teleported to body) listing all assignees by name.

  If no assignees, renders "Unassigned" text.

  Props:
    assignees — [{ id, display_name, avatar_url }]
-->
<script setup>
import { ref } from 'vue'
import Avatar from '@/components/ui/Avatar.vue'

const props = defineProps({
  assignees: {
    type: Array,
    default: () => [],
  },
})

const triggerRef = ref(null)
const showPopover = ref(false)
const popoverStyle = ref({})

function openPopover() {
  if (!props.assignees?.length) return
  const rect = triggerRef.value.getBoundingClientRect()
  popoverStyle.value = {
    position: 'fixed',
    top: `${rect.bottom + 6}px`,
    left: `${rect.left + rect.width / 2}px`,
    transform: 'translateX(-50%)',
    zIndex: 9999,
  }
  showPopover.value = true
}

function closePopover() {
  showPopover.value = false
}
</script>

<template>
  <div
    ref="triggerRef"
    class="flex items-center"
    @mouseenter="openPopover"
    @mouseleave="closePopover"
  >
    <!-- No assignees -->
    <span v-if="!assignees?.length" class="text-[11px] text-slate-400">
      Unassigned
    </span>

    <!-- Avatar stack -->
    <template v-else>
      <div class="flex -space-x-1">
        <Avatar
          v-for="a in assignees.slice(0, 2)"
          :key="a.id"
          :name="a.display_name"
          :src="a.avatar_url"
          size="xs"
          class="ring-1 ring-white"
        />
        <span
          v-if="assignees.length > 2"
          class="size-5 rounded-full bg-slate-100 dark:bg-slate-700 text-[10px] font-medium text-slate-500 dark:text-slate-400 flex items-center justify-center ring-1 ring-white dark:ring-slate-800"
        >
          +{{ assignees.length - 2 }}
        </span>
      </div>
    </template>
  </div>

  <!-- Popover -->
  <Teleport to="body">
    <Transition name="popover-fade">
      <div
        v-if="showPopover && assignees?.length"
        :style="popoverStyle"
        class="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg shadow-lg py-1.5 px-2 min-w-[130px]"
      >
        <div
          v-for="a in assignees"
          :key="a.id"
          class="flex items-center gap-2 py-1 px-1"
        >
          <Avatar :name="a.display_name" :src="a.avatar_url" size="xs" />
          <span class="text-xs text-slate-700 dark:text-slate-300 whitespace-nowrap">{{ a.display_name }}</span>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.popover-fade-enter-active,
.popover-fade-leave-active {
  transition: opacity 0.1s ease;
}
.popover-fade-enter-from,
.popover-fade-leave-to {
  opacity: 0;
}
</style>
