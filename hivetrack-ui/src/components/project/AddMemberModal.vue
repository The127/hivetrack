<!--
  AddMemberModal — select a user and role, then add them to the project.

  Props:
    open            — controls visibility
    slug            — project slug
    existingMembers — array of current project members (to filter out)

  Events:
    close   — modal dismissed
    added   — member was added successfully
-->
<script setup>
import { ref, computed, watch } from 'vue'
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import Modal from '@/components/ui/Modal.vue'
import Button from '@/components/ui/Button.vue'
import { addProjectMember } from '@/api/projects'
import { apiFetch } from '@/composables/useApi'

const props = defineProps({
  open: { type: Boolean, required: true },
  slug: { type: String, required: true },
  existingMembers: { type: Array, default: () => [] },
})

const emit = defineEmits(['close', 'added'])
const queryClient = useQueryClient()

const ROLES = [
  { value: 'project_admin', label: 'Admin' },
  { value: 'project_member', label: 'Member' },
  { value: 'viewer', label: 'Viewer' },
]

const selectedUserId = ref('')
const selectedRole = ref('project_member')
const allUsers = ref([])
const loadingUsers = ref(false)
const serverError = ref(null)

const existingUserIds = computed(() => new Set(props.existingMembers.map((m) => m.user_id)))

const availableUsers = computed(() =>
  allUsers.value.filter((u) => !existingUserIds.value.has(u.id)),
)

const canSubmit = computed(() => !!selectedUserId.value && !!selectedRole.value)

async function loadUsers() {
  if (allUsers.value.length) return
  loadingUsers.value = true
  try {
    const data = await apiFetch('/api/v1/users')
    allUsers.value = data.users ?? data ?? []
  } catch {
    allUsers.value = []
  } finally {
    loadingUsers.value = false
  }
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      loadUsers()
      serverError.value = null
    }
  },
)

const { mutate: doAdd, isPending: adding } = useMutation({
  mutationFn: () =>
    addProjectMember(props.slug, {
      user_id: selectedUserId.value,
      role: selectedRole.value,
    }),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['project', props.slug] })
    emit('added')
    closeModal()
  },
  onError: (err) => {
    serverError.value = err?.errors?.[0]?.message ?? 'Something went wrong. Please try again.'
  },
})

function closeModal() {
  selectedUserId.value = ''
  selectedRole.value = 'project_member'
  serverError.value = null
  emit('close')
}
</script>

<template>
  <Modal
    :open="open"
    title="Add member"
    description="Add a user to this project and assign their role."
    @close="closeModal"
  >
    <div class="flex flex-col gap-4">
      <!-- User select -->
      <div class="flex flex-col gap-1.5">
        <label class="text-sm font-medium text-slate-700 dark:text-slate-300" for="add-member-user">
          User
        </label>
        <select
          id="add-member-user"
          v-model="selectedUserId"
          :disabled="loadingUsers"
          class="block w-full rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-offset-0 focus:border-blue-400 focus:ring-blue-200 transition-colors disabled:opacity-50 cursor-pointer"
        >
          <option value="" disabled>{{ loadingUsers ? 'Loading…' : 'Select a user' }}</option>
          <option
            v-for="user in availableUsers"
            :key="user.id"
            :value="user.id"
          >
            {{ user.display_name ?? user.username ?? user.email ?? user.id }}
          </option>
        </select>
        <p v-if="!loadingUsers && availableUsers.length === 0" class="text-xs text-slate-400 dark:text-slate-500">
          No users available to add.
        </p>
      </div>

      <!-- Role select -->
      <div class="flex flex-col gap-1.5">
        <span class="text-sm font-medium text-slate-700 dark:text-slate-300">Role</span>
        <div class="flex gap-2">
          <label
            v-for="r in ROLES"
            :key="r.value"
            class="flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-sm cursor-pointer transition-colors"
            :class="
              selectedRole === r.value
                ? 'border-blue-300 dark:border-blue-700 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 font-medium'
                : 'border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:border-slate-300 dark:hover:border-slate-600'
            "
          >
            <input
              v-model="selectedRole"
              type="radio"
              :value="r.value"
              class="sr-only"
            />
            {{ r.label }}
          </label>
        </div>
      </div>

      <!-- Server error -->
      <p v-if="serverError" class="text-sm text-red-600">{{ serverError }}</p>
    </div>

    <template #footer>
      <Button variant="secondary" :disabled="adding" @click="closeModal">Cancel</Button>
      <Button variant="primary" :loading="adding" :disabled="!canSubmit" @click="doAdd">
        Add member
      </Button>
    </template>
  </Modal>
</template>
