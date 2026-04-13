<!--
  OverviewMembers — project members list with add/remove actions.
-->
<script setup>
import { computed, ref } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { UsersIcon, PlusIcon, Trash2Icon } from 'lucide-vue-next'
import Badge from '@/components/ui/Badge.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import AddMemberModal from '@/components/project/AddMemberModal.vue'
import { fetchProject, removeProjectMember } from '@/api/projects'

const props = defineProps({
  slug: { type: String, required: true },
})

const queryClient = useQueryClient()

const { data: project } = useQuery({
  queryKey: computed(() => ['project', props.slug]),
  queryFn: () => fetchProject(props.slug),
})

const members = computed(() => project.value?.members ?? [])

const showAddMemberModal = ref(false)
const memberToRemove = ref(null)

const { mutate: doRemoveMember, isPending: removeMemberPending } = useMutation({
  mutationFn: (userId) => removeProjectMember(props.slug, userId),
  onSuccess: () => {
    memberToRemove.value = null
    queryClient.invalidateQueries({ queryKey: ['project', props.slug] })
  },
})
</script>

<template>
  <section>
    <h2 class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-3 flex items-center gap-1.5">
      <UsersIcon class="size-4 text-slate-500 dark:text-slate-400" />
      Members
      <span v-if="members.length" class="text-xs font-normal text-slate-500 dark:text-slate-400">{{ members.length }}</span>
      <button
        class="ml-auto rounded-md p-1 text-slate-400 hover:text-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/30 transition-colors cursor-pointer"
        title="Add member"
        @click="showAddMemberModal = true"
      >
        <PlusIcon class="size-4" />
      </button>
    </h2>
    <div class="flex flex-wrap gap-2">
      <div
        v-for="m in members"
        :key="m.user_id"
        class="group flex items-center gap-2 rounded-md border border-slate-200 dark:border-slate-700 px-2.5 py-1.5"
      >
        <span class="text-sm text-slate-700 dark:text-slate-300">{{ m.display_name }}</span>
        <Badge colorScheme="gray" compact>{{ m.role.replace('project_', '') }}</Badge>
        <button
          class="opacity-0 group-hover:opacity-100 ml-0.5 rounded p-0.5 text-slate-400 dark:text-slate-500 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/30 transition-all cursor-pointer"
          title="Remove member"
          @click="memberToRemove = m"
        >
          <Trash2Icon class="size-3" />
        </button>
      </div>
    </div>
  </section>

  <!-- Add member modal -->
  <AddMemberModal
    :open="showAddMemberModal"
    :slug="slug"
    :existing-members="members"
    @close="showAddMemberModal = false"
    @added="showAddMemberModal = false"
  />

  <ConfirmDialog
    v-if="memberToRemove"
    :open="!!memberToRemove"
    title="Remove member?"
    :message="`Remove ${memberToRemove.display_name} from this project? They will lose access immediately.`"
    confirm-text="Remove member"
    :loading="removeMemberPending"
    @confirm="doRemoveMember(memberToRemove.user_id)"
    @cancel="memberToRemove = null"
  />
</template>
