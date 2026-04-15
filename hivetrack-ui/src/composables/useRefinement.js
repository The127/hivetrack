import { ref, computed, watch, onUnmounted } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { fetchEventSource } from '@microsoft/fetch-event-source'
import {
  startRefinementSession,
  sendRefinementMessage,
  getRefinementSession,
  acceptRefinementProposal,
  advanceRefinementPhase,
} from '@/api/refinement'
import { useAuth } from '@/composables/useAuth'

export const REFINEMENT_PHASES = [
  { id: 'actor_goal', label: 'Actor & Goal' },
  { id: 'main_scenario', label: 'Main Scenario' },
  { id: 'extensions', label: 'Extensions' },
  { id: 'acceptance_criteria', label: 'Acceptance Criteria' },
]

export function useRefinement(slug, number) {
  const queryClient = useQueryClient()
  const { getAccessToken } = useAuth()
  const isOpen = ref(false)
  let pollInterval = null
  let streamCtrl = null

  const sessionQueryKey = ['refinement-session', slug, number]

  const {
    data: session,
    isLoading: sessionLoading,
    refetch: refetchSession,
  } = useQuery({
    queryKey: sessionQueryKey,
    queryFn: () => getRefinementSession(slug.value, number.value),
    enabled: isOpen,
    staleTime: 0,
  })

  const { mutate: doStart, isPending: startPending } = useMutation({
    mutationFn: () => startRefinementSession(slug.value, number.value),
    onSuccess: () => {
      refetchSession()
      startPolling()
      startStream()
    },
  })

  const { mutate: doSend, isPending: sendPending } = useMutation({
    mutationFn: (content) => sendRefinementMessage(slug.value, number.value, content),
    onSuccess: () => {
      refetchSession()
    },
  })

  const { mutate: doAccept, isPending: acceptPending } = useMutation({
    mutationFn: () => acceptRefinementProposal(slug.value, number.value),
    onSuccess: () => {
      stopPolling()
      stopStream()
      isOpen.value = false
      refetchSession()
      queryClient.invalidateQueries({ queryKey: ['issue', slug.value, number.value] })
      queryClient.invalidateQueries({ queryKey: ['issues', slug.value] })
    },
  })

  const { mutate: doAdvance, isPending: advancePending } = useMutation({
    mutationFn: (targetPhase) => advanceRefinementPhase(slug.value, number.value, targetPhase),
    onSuccess: () => {
      refetchSession()
    },
  })

  const currentPhase = computed(() => session.value?.current_phase ?? 'actor_goal')

  function open() {
    isOpen.value = true
    startPolling()
    startStream()
  }

  function close() {
    isOpen.value = false
    stopPolling()
    stopStream()
  }

  function startPolling() {
    stopPolling()
    pollInterval = setInterval(() => {
      if (isOpen.value) {
        refetchSession()
      }
    }, 2000)
  }

  function stopPolling() {
    if (pollInterval) {
      clearInterval(pollInterval)
      pollInterval = null
    }
  }

  function startStream() {
    stopStream()
    const ctrl = new AbortController()
    streamCtrl = ctrl
    const token = getAccessToken()
    fetchEventSource(
      `/api/v1/projects/${slug.value}/issues/${number.value}/refinement/stream`,
      {
        method: 'GET',
        signal: ctrl.signal,
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        openWhenHidden: true,
        onmessage() {
          refetchSession()
        },
        onerror(err) {
          // Return a retry delay (ms) so the lib reconnects automatically.
          // Polling is the safety net if this stays broken.
          if (ctrl.signal.aborted) throw err
          return 3000
        },
      },
    ).catch(() => {
      // Swallow — stream is best-effort, polling keeps data flowing.
    })
  }

  function stopStream() {
    if (streamCtrl) {
      streamCtrl.abort()
      streamCtrl = null
    }
  }

  watch(
    () => session.value?.status,
    (status) => {
      if (status === 'completed' || status === 'abandoned' || status === 'failed') {
        stopPolling()
        stopStream()
      }
    },
  )

  onUnmounted(() => {
    stopPolling()
    stopStream()
  })

  return {
    session,
    sessionLoading,
    isOpen,
    startPending,
    sendPending,
    acceptPending,
    advancePending,
    currentPhase,
    open,
    close,
    startSession: doStart,
    sendMessage: doSend,
    acceptProposal: doAccept,
    advancePhase: doAdvance,
  }
}
