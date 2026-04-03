<!--
  RefinementPanel — right slide-out chat panel for interactive issue refinement via Hivemind.

  Props:
    open      Boolean — controls visibility
    session   Object  — refinement session data (from useRefinement)
    loading   Boolean — session loading state
    sendPending   Boolean — message send in progress
    acceptPending Boolean — proposal accept in progress

  Events:
    close    — user closed the panel
    send     — user sent a message (payload: string)
    accept   — user accepted a proposal
    start    — user wants to start a new session
-->
<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { XIcon, SendIcon, SparklesIcon, CheckIcon } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import MarkdownContent from '@/components/ui/MarkdownContent.vue'
import Spinner from '@/components/ui/Spinner.vue'

const props = defineProps({
  open: { type: Boolean, required: true },
  session: { type: Object, default: null },
  loading: { type: Boolean, default: false },
  sendPending: { type: Boolean, default: false },
  acceptPending: { type: Boolean, default: false },
})

const emit = defineEmits(['close', 'send', 'accept', 'start'])

const messageInput = ref('')
const messagesEnd = ref(null)
const inputRef = ref(null)

function sendMessage() {
  const content = messageInput.value.trim()
  if (!content) return
  emit('send', content)
  messageInput.value = ''
}

function onKeydown(e) {
  if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
    e.preventDefault()
    sendMessage()
  }
}

function onEscape(e) {
  if (e.key === 'Escape' && props.open) emit('close')
}

function scrollToBottom() {
  nextTick(() => {
    messagesEnd.value?.scrollIntoView({ behavior: 'smooth' })
  })
}

watch(
  () => props.session?.messages?.length,
  () => scrollToBottom(),
)

watch(
  () => props.open,
  (val) => {
    if (val) {
      nextTick(() => {
        inputRef.value?.focus()
        scrollToBottom()
      })
    }
  },
)

onMounted(() => document.addEventListener('keydown', onEscape))
onUnmounted(() => document.removeEventListener('keydown', onEscape))

const hasActiveSession = computed(() => props.session && props.session.status === 'active')
const messages = computed(() => props.session?.messages ?? [])
const latestProposal = computed(() => {
  for (let i = messages.value.length - 1; i >= 0; i--) {
    if (messages.value[i].message_type === 'proposal' && messages.value[i].proposal) {
      return messages.value[i]
    }
  }
  return null
})

// Detect if we're waiting for an assistant response
const waitingForResponse = computed(() => {
  if (!hasActiveSession.value) return false
  const msgs = messages.value
  if (msgs.length === 0) return true // just started, waiting for first response
  return msgs[msgs.length - 1].role === 'user'
})
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-transform duration-200 ease-out"
      enter-from-class="translate-x-full"
      enter-to-class="translate-x-0"
      leave-active-class="transition-transform duration-150 ease-in"
      leave-from-class="translate-x-0"
      leave-to-class="translate-x-full"
    >
      <div v-if="open" class="fixed inset-y-0 right-0 z-50 flex">
        <!-- Backdrop -->
        <div
          class="fixed inset-0 bg-black/20 dark:bg-black/40"
          @click="emit('close')"
        />

        <!-- Panel -->
        <div class="relative ml-auto flex h-full w-full max-w-lg flex-col bg-white dark:bg-slate-900 shadow-2xl ring-1 ring-slate-900/10 dark:ring-slate-700">
          <!-- Header -->
          <div class="flex items-center justify-between gap-4 px-5 py-4 border-b border-slate-100 dark:border-slate-800">
            <div class="flex items-center gap-2">
              <SparklesIcon class="size-4 text-violet-500" />
              <h2 class="text-sm font-semibold text-slate-900 dark:text-slate-100">Refinement</h2>
            </div>
            <button
              class="rounded-md p-1 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
              @click="emit('close')"
            >
              <XIcon class="size-4" />
            </button>
          </div>

          <!-- Messages area -->
          <div class="flex-1 overflow-y-auto px-5 py-4 space-y-4">
            <!-- Loading -->
            <div v-if="loading" class="flex items-center justify-center py-12">
              <Spinner class="size-5 text-slate-400" />
            </div>

            <!-- No session yet -->
            <div v-else-if="!session" class="flex flex-col items-center justify-center py-12 text-center">
              <SparklesIcon class="size-8 text-slate-300 dark:text-slate-600 mb-3" />
              <p class="text-sm text-slate-500 dark:text-slate-400 mb-4">
                Start a refinement session to interactively refine this issue with Hivemind.
              </p>
              <Button variant="primary" size="sm" @click="emit('start')">
                <SparklesIcon class="size-3.5" />
                Start session
              </Button>
            </div>

            <!-- Messages -->
            <template v-else>
              <div
                v-for="msg in messages"
                :key="msg.id"
                :class="[
                  'max-w-[85%] rounded-lg px-4 py-3',
                  msg.role === 'user'
                    ? 'ml-auto bg-blue-50 dark:bg-blue-950/40 text-slate-800 dark:text-slate-200'
                    : 'mr-auto bg-slate-50 dark:bg-slate-800/60 text-slate-800 dark:text-slate-200',
                ]"
              >
                <!-- Proposal card -->
                <div v-if="msg.message_type === 'proposal' && msg.proposal" class="space-y-3">
                  <MarkdownContent :content="msg.content" />
                  <div class="rounded-md border border-violet-200 dark:border-violet-800 bg-violet-50 dark:bg-violet-950/30 p-3 space-y-1.5">
                    <p class="text-xs font-semibold text-violet-600 dark:text-violet-400 uppercase tracking-wide">Proposal</p>
                    <p class="text-sm font-medium text-slate-900 dark:text-slate-100">{{ msg.proposal.title }}</p>
                    <div class="text-sm text-slate-600 dark:text-slate-300">
                      <MarkdownContent :content="msg.proposal.description" />
                    </div>
                  </div>
                  <div v-if="hasActiveSession && msg === latestProposal" class="flex items-center gap-2 pt-1">
                    <Button
                      variant="primary"
                      size="sm"
                      :loading="acceptPending"
                      @click="emit('accept')"
                    >
                      <CheckIcon class="size-3.5" />
                      Accept
                    </Button>
                    <Button
                      variant="secondary"
                      size="sm"
                      :disabled="acceptPending"
                      @click="inputRef?.focus()"
                    >
                      Continue refining
                    </Button>
                  </div>
                </div>

                <!-- Regular message -->
                <div v-else>
                  <MarkdownContent v-if="msg.role === 'assistant'" :content="msg.content" />
                  <p v-else class="text-sm whitespace-pre-wrap">{{ msg.content }}</p>
                </div>
              </div>

              <!-- Waiting indicator -->
              <div v-if="waitingForResponse" class="mr-auto flex items-center gap-2 text-slate-400 dark:text-slate-500 py-2">
                <Spinner class="size-3.5" />
                <span class="text-xs">Hivemind is thinking...</span>
              </div>

              <div ref="messagesEnd" />
            </template>
          </div>

          <!-- Input area -->
          <div v-if="hasActiveSession" class="border-t border-slate-100 dark:border-slate-800 px-5 py-3">
            <div class="flex items-end gap-2">
              <textarea
                ref="inputRef"
                v-model="messageInput"
                rows="2"
                placeholder="Ask a question or provide more context..."
                class="flex-1 resize-none rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-sm text-slate-700 dark:text-slate-300 placeholder:text-slate-400 dark:placeholder:text-slate-500 px-3 py-2 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
                :disabled="sendPending"
                @keydown="onKeydown"
              />
              <Button
                variant="primary"
                size="sm"
                :disabled="!messageInput.trim() || sendPending"
                :loading="sendPending"
                @click="sendMessage"
              >
                <SendIcon class="size-3.5" />
              </Button>
            </div>
            <p class="mt-1 text-[10px] text-slate-400 dark:text-slate-500">Ctrl+Enter to send</p>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
