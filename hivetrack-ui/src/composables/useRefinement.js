import { ref, watch, onUnmounted } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import {
  startRefinementSession,
  sendRefinementMessage,
  getRefinementSession,
  acceptRefinementProposal,
} from '@/api/refinement'

export function useRefinement(slug, number) {
  const queryClient = useQueryClient()
  const isOpen = ref(false)
  let pollInterval = null

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
      refetchSession()
      queryClient.invalidateQueries({ queryKey: ['issue', slug.value, number.value] })
      queryClient.invalidateQueries({ queryKey: ['issues', slug.value] })
    },
  })

  function open() {
    isOpen.value = true
    startPolling()
  }

  function close() {
    isOpen.value = false
    stopPolling()
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

  watch(
    () => session.value?.status,
    (status) => {
      if (status === 'completed' || status === 'abandoned') {
        stopPolling()
      }
    },
  )

  onUnmounted(() => {
    stopPolling()
  })

  return {
    session,
    sessionLoading,
    isOpen,
    startPending,
    sendPending,
    acceptPending,
    open,
    close,
    startSession: doStart,
    sendMessage: doSend,
    acceptProposal: doAccept,
  }
}
