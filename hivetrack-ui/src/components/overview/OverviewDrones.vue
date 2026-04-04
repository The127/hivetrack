<!--
  OverviewDrones — Hivemind drone list with add/delete/revoke actions.
-->
<script setup>
import { computed, ref } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { CpuIcon, PlusIcon, Trash2Icon } from 'lucide-vue-next'
import Badge from '@/components/ui/Badge.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import AddDroneModal from '@/components/project/AddDroneModal.vue'
import RelativeTime from '@/components/ui/RelativeTime.vue'
import { fetchDrones, deleteDrone, revokeToken } from '@/api/drones'

const props = defineProps({
  slug: { type: String, required: true },
})

const queryClient = useQueryClient()

const { data: dronesData } = useQuery({
  queryKey: computed(() => ['drones', props.slug]),
  queryFn: () => fetchDrones(props.slug),
  enabled: computed(() => !!props.slug),
  refetchInterval: 10000,
  retry: false,
})

const drones = computed(() => dronesData.value?.drones ?? [])
const pendingTokens = computed(() => dronesData.value?.pending_tokens ?? [])
const hivemindAvailable = computed(() => dronesData.value !== undefined)
const showAddDroneModal = ref(false)
const droneToDelete = ref(null)
const tokenToRevoke = ref(null)

const { mutate: doDeleteDrone, isPending: deleteDronePending } = useMutation({
  mutationFn: (droneId) => deleteDrone(props.slug, droneId),
  onSuccess: () => {
    droneToDelete.value = null
    queryClient.invalidateQueries({ queryKey: ['drones', props.slug] })
  },
})

const { mutate: doRevokeToken, isPending: revokeTokenPending } = useMutation({
  mutationFn: (token) => revokeToken(props.slug, token),
  onSuccess: () => {
    tokenToRevoke.value = null
    queryClient.invalidateQueries({ queryKey: ['drones', props.slug] })
  },
})
</script>

<template>
  <section v-if="hivemindAvailable">
    <h2 class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-3 flex items-center gap-1.5">
      <CpuIcon class="size-4 text-slate-500 dark:text-slate-400" />
      Drones
      <span v-if="drones.length" class="text-xs font-normal text-slate-500 dark:text-slate-400">{{ drones.length }}</span>
      <button
        class="ml-auto rounded-md p-1 text-slate-400 hover:text-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/30 transition-colors cursor-pointer"
        title="Add drone"
        @click="showAddDroneModal = true"
      >
        <PlusIcon class="size-4" />
      </button>
    </h2>

    <!-- Empty state -->
    <div v-if="!drones.length && !pendingTokens.length" class="rounded-lg border border-dashed border-slate-200 dark:border-slate-700 px-4 py-6 text-center">
      <CpuIcon class="size-6 mx-auto text-slate-300 dark:text-slate-600 mb-2" />
      <p class="text-sm text-slate-500 dark:text-slate-400">No drones connected.</p>
      <p class="text-xs text-slate-400 dark:text-slate-500 mt-1">Add a drone to enable Hivemind-powered features.</p>
    </div>

    <!-- Drone list -->
    <div v-else class="rounded-lg border border-slate-200 dark:border-slate-700 divide-y divide-slate-100 dark:divide-slate-800 overflow-hidden">
      <!-- Connected drones -->
      <div
        v-for="drone in drones"
        :key="drone.drone_id"
        class="group flex items-center gap-3 px-4 py-2.5"
      >
        <span class="text-sm font-mono text-slate-700 dark:text-slate-300 flex-1 truncate">{{ drone.drone_id }}</span>
        <Badge
          :color-scheme="drone.status === 'available' ? 'green' : drone.status === 'busy' ? 'yellow' : 'red'"
          compact
        >
          {{ drone.status }}
        </Badge>
        <span v-for="cap in drone.capabilities" :key="cap" class="text-[10px] rounded-full bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 text-slate-500 dark:text-slate-400">
          {{ cap }}
        </span>
        <span v-if="drone.last_heartbeat" class="text-xs text-slate-400 dark:text-slate-500">
          <RelativeTime :date="drone.last_heartbeat" />
        </span>
        <button
          class="opacity-0 group-hover:opacity-100 rounded-md p-1 text-slate-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/30 transition-all cursor-pointer"
          title="Delete drone"
          @click="droneToDelete = drone"
        >
          <Trash2Icon class="size-3.5" />
        </button>
      </div>

      <!-- Pending tokens (not yet connected) -->
      <div
        v-for="(token, i) in pendingTokens"
        :key="'pending-' + i"
        class="group flex items-center gap-3 px-4 py-2.5 opacity-50 hover:opacity-75"
      >
        <span class="text-sm text-slate-400 dark:text-slate-500 flex-1">Waiting for drone...</span>
        <Badge color-scheme="gray" compact>Pending</Badge>
        <span v-for="cap in token.capabilities" :key="cap" class="text-[10px] rounded-full bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 text-slate-500 dark:text-slate-400">
          {{ cap }}
        </span>
        <span class="text-xs text-slate-400 dark:text-slate-500">
          <RelativeTime :date="token.created_at" />
        </span>
        <button
          class="opacity-0 group-hover:opacity-100 rounded-md p-1 text-slate-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/30 transition-all cursor-pointer"
          title="Revoke token"
          @click="tokenToRevoke = token"
        >
          <Trash2Icon class="size-3.5" />
        </button>
      </div>
    </div>
  </section>

  <!-- Add drone modal -->
  <AddDroneModal
    :open="showAddDroneModal"
    :slug="slug"
    @close="showAddDroneModal = false"
  />

  <!-- Delete drone confirmation -->
  <ConfirmDialog
    v-if="droneToDelete"
    :open="!!droneToDelete"
    title="Delete drone?"
    :message="`Delete '${droneToDelete.drone_id}'? It will be disconnected and its registration purged.`"
    confirm-text="Delete"
    :loading="deleteDronePending"
    @confirm="doDeleteDrone(droneToDelete.drone_id)"
    @cancel="droneToDelete = null"
  />

  <!-- Revoke token confirmation -->
  <ConfirmDialog
    v-if="tokenToRevoke"
    :open="!!tokenToRevoke"
    title="Revoke token?"
    message="Revoke this pending drone token? Any drone attempting to register with it will be rejected."
    confirm-text="Revoke token"
    :loading="revokeTokenPending"
    @confirm="doRevokeToken(tokenToRevoke.token)"
    @cancel="tokenToRevoke = null"
  />
</template>
