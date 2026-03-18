<!--
  OwnerSelect — single-user selector for the issue owner (DRI).

  Props:
    projectSlug — project slug (to fetch members)
    modelValue  — { id, display_name, avatar_url } | null (current owner)

  Emits:
    update:modelValue — uuid string | null
-->
<script setup>
import { ref, computed, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { fetchProject } from '@/api/projects'
import { UserIcon, SearchIcon, CheckIcon, XIcon } from 'lucide-vue-next'
import Avatar from '@/components/ui/Avatar.vue'

const props = defineProps({
  projectSlug: { type: String, required: true },
  modelValue: { type: Object, default: null },
})

const emit = defineEmits(['update:modelValue'])

const { data: project } = useQuery({
  queryKey: ['project', computed(() => props.projectSlug)],
  queryFn: () => fetchProject(props.projectSlug),
  enabled: computed(() => !!props.projectSlug),
})

const members = computed(() => project.value?.members ?? [])

// ── Dropdown state ───────────────────────────────────────────────────────────

const open = ref(false)
const root = ref(null)
const triggerBtn = ref(null)
const dropdownEl = ref(null)
const searchInput = ref(null)
const dropdownStyle = ref({})
const search = ref('')

const filteredMembers = computed(() => {
  if (!search.value) return members.value
  const q = search.value.toLowerCase()
  return members.value.filter(m => m.display_name.toLowerCase().includes(q))
})

function positionDropdown() {
  if (!triggerBtn.value) return
  const rect = triggerBtn.value.getBoundingClientRect()
  const spaceBelow = window.innerHeight - rect.bottom
  const goUp = spaceBelow < 240 && rect.top > spaceBelow

  dropdownStyle.value = {
    position: 'fixed',
    left: `${rect.left}px`,
    width: `${Math.max(rect.width, 220)}px`,
    zIndex: 9999,
    ...(goUp
      ? { bottom: `${window.innerHeight - rect.top + 4}px` }
      : { top: `${rect.bottom + 4}px` }),
  }
}

function toggle() {
  open.value = !open.value
  if (open.value) {
    search.value = ''
    nextTick(() => {
      positionDropdown()
      searchInput.value?.focus()
    })
  }
}

function selectMember(userId) {
  emit('update:modelValue', userId)
  open.value = false
}

function clearOwner() {
  emit('update:modelValue', null)
  open.value = false
}

// ── Click outside ────────────────────────────────────────────────────────────

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
  <div ref="root" class="flex flex-col gap-1.5">
    <label class="text-xs font-medium text-slate-500 flex items-center gap-1">
      <UserIcon class="size-3" />
      Owner
    </label>

    <!-- Trigger -->
    <button
      ref="triggerBtn"
      class="w-full flex items-center gap-2 rounded-md border border-slate-200 px-2.5 py-1.5 text-sm text-left cursor-pointer bg-white hover:border-slate-300 hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors"
      @click="toggle"
    >
      <template v-if="modelValue">
        <Avatar :name="modelValue.display_name" :src="modelValue.avatar_url" size="xs" />
        <span class="flex-1 min-w-0 truncate text-slate-700">{{ modelValue.display_name }}</span>
      </template>
      <template v-else>
        <span class="flex-1 text-slate-400">No owner</span>
      </template>
    </button>

    <!-- Dropdown -->
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
          class="bg-white border border-slate-200 rounded-lg shadow-lg overflow-hidden"
        >
          <!-- Search -->
          <div v-if="members.length > 5" class="p-2 border-b border-slate-100">
            <div class="relative">
              <SearchIcon class="absolute left-2 top-1/2 -translate-y-1/2 size-3.5 text-slate-400" />
              <input
                ref="searchInput"
                v-model="search"
                type="text"
                placeholder="Search members..."
                class="w-full pl-7 pr-2 py-1 text-sm text-slate-800 placeholder:text-slate-400 bg-slate-50 rounded border-none focus:outline-none"
                @keydown.escape="open = false"
              />
            </div>
          </div>

          <!-- Options -->
          <div class="max-h-52 overflow-y-auto py-1">
            <!-- Clear option -->
            <button
              v-if="modelValue"
              class="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-left cursor-pointer text-slate-500 hover:bg-slate-50 transition-colors"
              @click="clearOwner"
            >
              <XIcon class="size-3.5 flex-shrink-0" />
              <span>No owner</span>
            </button>

            <button
              v-for="member in filteredMembers"
              :key="member.user_id"
              class="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-left cursor-pointer transition-colors"
              :class="modelValue?.id === member.user_id ? 'bg-slate-50 font-medium text-slate-900' : 'text-slate-700 hover:bg-slate-50'"
              @click="selectMember(member.user_id)"
            >
              <CheckIcon v-if="modelValue?.id === member.user_id" class="size-3.5 text-blue-500 flex-shrink-0" />
              <span v-else class="size-3.5 flex-shrink-0" />
              <Avatar :name="member.display_name" :src="member.avatar_url" size="xs" />
              <span class="flex-1 min-w-0 truncate">{{ member.display_name }}</span>
            </button>

            <p v-if="search && !filteredMembers.length" class="px-3 py-2 text-xs text-slate-400">
              No members match "{{ search }}"
            </p>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
