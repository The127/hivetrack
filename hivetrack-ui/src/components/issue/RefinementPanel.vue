<!--
  RefinementPanel — centered modal for phase-gated issue refinement via Hivemind.
  Two-column layout: left shows the story being built, right shows the active conversation.
-->
<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import {
  XIcon,
  SendIcon,
  SparklesIcon,
  CheckIcon,
  ChevronRightIcon,
  ChevronDownIcon,
  ArrowLeftIcon,
  UserIcon,
  TargetIcon,
  ListOrderedIcon,
  ShieldAlertIcon,
  ClipboardCheckIcon,
} from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import MarkdownContent from '@/components/ui/MarkdownContent.vue'
import Spinner from '@/components/ui/Spinner.vue'
import { REFINEMENT_PHASES } from '@/composables/useRefinement'

const props = defineProps({
  open: { type: Boolean, required: true },
  session: { type: Object, default: null },
  loading: { type: Boolean, default: false },
  sendPending: { type: Boolean, default: false },
  acceptPending: { type: Boolean, default: false },
  advancePending: { type: Boolean, default: false },
  currentPhase: { type: String, default: 'actor_goal' },
})

const emit = defineEmits(['close', 'send', 'accept', 'start', 'advance-phase'])

const messageInput = ref('')
const messagesEnd = ref(null)
const inputRef = ref(null)
const forceShowInput = ref(false)
const collapsedPhases = ref(new Set())

const currentPhaseIndex = computed(() =>
  REFINEMENT_PHASES.findIndex((p) => p.id === props.currentPhase),
)

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
  () => {
    scrollToBottom()
    forceShowInput.value = false
  },
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

// Messages grouped by phase
function messagesForPhase(phaseId) {
  return messages.value.filter((m) => m.phase === phaseId)
}

const currentPhaseMessages = computed(() => messagesForPhase(props.currentPhase))

// The latest assistant message in the current phase (the active question)
const latestAssistantMessage = computed(() => {
  for (let i = currentPhaseMessages.value.length - 1; i >= 0; i--) {
    if (currentPhaseMessages.value[i].role === 'assistant') {
      return currentPhaseMessages.value[i]
    }
  }
  return null
})

// Previous exchanges in the current phase (everything except the latest assistant message)
const previousExchanges = computed(() => {
  const msgs = currentPhaseMessages.value
  if (!latestAssistantMessage.value) return []
  const lastIdx = msgs.lastIndexOf(latestAssistantMessage.value)
  return msgs.slice(0, lastIdx)
})

const latestProposal = computed(() => {
  for (let i = messages.value.length - 1; i >= 0; i--) {
    if (messages.value[i].message_type === 'proposal' && messages.value[i].proposal) {
      return messages.value[i]
    }
  }
  return null
})

const waitingForResponse = computed(() => {
  if (!hasActiveSession.value) return false
  const msgs = messages.value
  if (msgs.length === 0) return true
  // If the current phase has no messages yet, we're waiting for the first response
  if (currentPhaseMessages.value.length === 0) return true
  return msgs[msgs.length - 1].role === 'user'
})

const isLastPhase = computed(
  () => currentPhaseIndex.value === REFINEMENT_PHASES.length - 1,
)

// Detect if the latest assistant message is asking for confirmation (but not a proposal)
const isConfirmationQuestion = computed(() => {
  if (!latestAssistantMessage.value) return false
  if (latestAssistantMessage.value.message_type === 'proposal') return false
  const content = latestAssistantMessage.value.content.toLowerCase()
  return (
    content.includes('does this look right') ||
    content.includes('look correct') ||
    content.includes('confirm') ||
    content.includes('anything to adjust') ||
    content.includes('anything to change') ||
    content.includes('look good') ||
    content.includes('shall i proceed') ||
    content.includes('ready to move on')
  )
})

const canAdvance = computed(() => {
  if (!hasActiveSession.value || isLastPhase.value || waitingForResponse.value) return false
  return currentPhaseMessages.value.some((m) => m.role === 'assistant')
})

// Only show "Next Phase" when the user isn't being asked a question that needs an answer
const showNextPhase = computed(() => {
  if (!canAdvance.value) return false
  // If Hivemind is asking for confirmation, show Confirm button instead
  if (isConfirmationQuestion.value) return false
  return true
})

const canGoBack = computed(() => hasActiveSession.value && currentPhaseIndex.value > 0)

// Hide footer when a proposal is showing — it has its own Accept/Request changes buttons
const showFooter = computed(() => {
  if (!hasActiveSession.value) return false
  if (latestAssistantMessage.value?.message_type === 'proposal') return false
  return true
})

// Check if a phase has any conversation content
function phaseHasContent(phaseId) {
  return messagesForPhase(phaseId).length > 0
}

// Get structured phase_data for a phase (from the latest message that has it)
function phaseData(phaseId) {
  const msgs = messagesForPhase(phaseId)
  for (let i = msgs.length - 1; i >= 0; i--) {
    if (msgs[i].phase_data) return msgs[i].phase_data
  }
  return null
}

// Extract structured-ish data from conversation text when phase_data is missing.
// Parses bold labels like **Goal:** X from assistant messages.
function extractedPhaseInfo(phaseId) {
  const msgs = messagesForPhase(phaseId)
  const assistantMsgs = msgs.filter((m) => m.role === 'assistant').map((m) => m.content)
  const all = assistantMsgs.join('\n')

  function extract(label) {
    const re = new RegExp(`\\*\\*${label}:?\\*\\*:?\\s*(.+)`, 'i')
    const match = all.match(re)
    return match ? match[1].trim() : null
  }

  if (phaseId === 'actor_goal') {
    const actor = extract('Actor')
    const goal = extract('Goal')
    if (actor || goal) return { actor, goal }
  }
  return null
}

function togglePhaseCollapse(phaseId) {
  if (collapsedPhases.value.has(phaseId)) {
    collapsedPhases.value.delete(phaseId)
  } else {
    collapsedPhases.value.add(phaseId)
  }
}

// Strip [Pass N/4 — Label] prefixes from Hivemind messages
function cleanContent(text) {
  return text.replace(/^\[Pass \d+\/\d+\s*[—–-]\s*[^\]]*\]\s*/i, '')
}

const PHASE_ICONS = {
  actor_goal: UserIcon,
  main_scenario: ListOrderedIcon,
  extensions: ShieldAlertIcon,
  acceptance_criteria: ClipboardCheckIcon,
}
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-150"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-100"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <!-- Backdrop -->
        <div
          class="absolute inset-0 bg-black/40 dark:bg-black/60"
          @click="emit('close')"
        />

        <!-- Modal -->
        <div class="relative z-10 flex flex-col w-full max-w-5xl max-h-[90vh] rounded-xl bg-white dark:bg-slate-900 shadow-2xl ring-1 ring-slate-900/10 dark:ring-slate-700">
          <!-- Header with phase stepper -->
          <div class="flex-shrink-0 border-b border-slate-100 dark:border-slate-800">
            <div class="flex items-center justify-between gap-4 px-6 pt-5 pb-3">
              <div class="flex items-center gap-2.5">
                <SparklesIcon class="size-5 text-violet-500" />
                <h2 class="text-base font-semibold text-slate-900 dark:text-slate-100">Refinement</h2>
              </div>
              <button
                class="rounded-md p-1.5 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
                @click="emit('close')"
              >
                <XIcon class="size-4" />
              </button>
            </div>

            <!-- Phase stepper -->
            <div v-if="hasActiveSession" class="px-6 pb-4">
              <div class="flex items-center">
                <template v-for="(phase, idx) in REFINEMENT_PHASES" :key="phase.id">
                  <button
                    :class="[
                      'relative flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition-all',
                      idx < currentPhaseIndex
                        ? 'bg-emerald-50 dark:bg-emerald-950/30 text-emerald-700 dark:text-emerald-400 hover:bg-emerald-100 dark:hover:bg-emerald-900/40 cursor-pointer'
                        : idx === currentPhaseIndex
                          ? 'bg-violet-100 dark:bg-violet-950/40 text-violet-700 dark:text-violet-300 ring-2 ring-violet-300 dark:ring-violet-700'
                          : 'bg-slate-50 dark:bg-slate-800/50 text-slate-400 dark:text-slate-500',
                    ]"
                    :disabled="idx >= currentPhaseIndex || advancePending"
                    @click="idx < currentPhaseIndex && emit('advance-phase', phase.id)"
                  >
                    <span
                      :class="[
                        'flex items-center justify-center size-5 rounded-full text-[10px] font-bold',
                        idx < currentPhaseIndex
                          ? 'bg-emerald-200 dark:bg-emerald-800 text-emerald-700 dark:text-emerald-300'
                          : idx === currentPhaseIndex
                            ? 'bg-violet-200 dark:bg-violet-800 text-violet-700 dark:text-violet-300'
                            : 'bg-slate-200 dark:bg-slate-700 text-slate-400 dark:text-slate-500',
                      ]"
                    >
                      <CheckIcon v-if="idx < currentPhaseIndex" class="size-3" />
                      <span v-else>{{ idx + 1 }}</span>
                    </span>
                    {{ phase.label }}
                  </button>
                  <ChevronRightIcon
                    v-if="idx < REFINEMENT_PHASES.length - 1"
                    class="mx-0.5 size-3.5 shrink-0 text-slate-300 dark:text-slate-600"
                  />
                </template>
              </div>
            </div>
          </div>

          <!-- Body -->
          <div class="flex-1 flex flex-col overflow-hidden min-h-0">
            <!-- Loading -->
            <div v-if="loading" class="flex items-center justify-center py-16">
              <Spinner class="size-6 text-slate-400" />
            </div>

            <!-- No session yet -->
            <div v-else-if="!session" class="flex flex-col items-center justify-center py-16 text-center px-8">
              <SparklesIcon class="size-10 text-slate-200 dark:text-slate-700 mb-4" />
              <h3 class="text-lg font-semibold text-slate-700 dark:text-slate-300 mb-2">Refine this issue</h3>
              <p class="text-sm text-slate-500 dark:text-slate-400 mb-6 max-w-md">
                Hivemind will guide you through four phases to build a structured user story
                with actor, goal, scenario, extensions, and acceptance criteria.
              </p>
              <Button variant="primary" @click="emit('start')">
                <SparklesIcon class="size-4" />
                Start refinement
              </Button>
            </div>

            <!-- Active session: two-column layout -->
            <div v-else class="flex flex-1 min-h-0">
              <!-- Left column: Story progress -->
              <div class="w-72 flex-shrink-0 border-r border-slate-100 dark:border-slate-800 overflow-y-auto bg-slate-50/50 dark:bg-slate-800/20">
                <div class="p-4 space-y-1">
                  <p class="text-[10px] font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-3 px-1">Story Progress</p>

                  <template v-for="(phase, idx) in REFINEMENT_PHASES" :key="phase.id">
                    <div
                      :class="[
                        'rounded-lg p-3 transition-colors',
                        phase.id === currentPhase
                          ? 'bg-violet-50 dark:bg-violet-950/30 ring-1 ring-violet-200 dark:ring-violet-800'
                          : phaseHasContent(phase.id)
                            ? 'bg-white dark:bg-slate-800/60'
                            : 'opacity-40',
                      ]"
                    >
                      <button
                        class="flex items-center gap-2 w-full text-left"
                        :disabled="!phaseHasContent(phase.id)"
                        @click="phaseHasContent(phase.id) && togglePhaseCollapse(phase.id)"
                      >
                        <component
                          :is="PHASE_ICONS[phase.id]"
                          :class="[
                            'size-3.5 shrink-0',
                            idx < currentPhaseIndex ? 'text-emerald-500' :
                            phase.id === currentPhase ? 'text-violet-500' : 'text-slate-400',
                          ]"
                        />
                        <span
                          :class="[
                            'text-xs font-semibold flex-1',
                            idx < currentPhaseIndex ? 'text-emerald-600 dark:text-emerald-400' :
                            phase.id === currentPhase ? 'text-violet-600 dark:text-violet-400' :
                            'text-slate-400 dark:text-slate-500',
                          ]"
                        >
                          {{ phase.label }}
                        </span>
                        <CheckIcon v-if="idx < currentPhaseIndex" class="size-3 text-emerald-500 shrink-0" />
                        <ChevronDownIcon
                          v-if="phaseHasContent(phase.id)"
                          :class="[
                            'size-3 shrink-0 transition-transform text-slate-400',
                            collapsedPhases.has(phase.id) ? '-rotate-90' : '',
                          ]"
                        />
                      </button>

                      <div v-if="!collapsedPhases.has(phase.id)" class="mt-1.5">
                        <!-- Structured phase data -->
                        <template v-if="phaseData(phase.id)">
                          <!-- Actor & Goal -->
                          <div v-if="phase.id === 'actor_goal'" class="text-[11px] leading-relaxed text-slate-500 dark:text-slate-400 space-y-1">
                            <p><span class="font-semibold text-slate-600 dark:text-slate-300">Actor:</span> {{ phaseData(phase.id).actor }}</p>
                            <p><span class="font-semibold text-slate-600 dark:text-slate-300">Goal:</span> {{ phaseData(phase.id).goal }}</p>
                          </div>
                          <!-- Main Scenario -->
                          <div v-else-if="phase.id === 'main_scenario'" class="text-[11px] leading-relaxed text-slate-500 dark:text-slate-400">
                            <ol class="list-decimal list-inside space-y-0.5">
                              <li v-for="(step, i) in phaseData(phase.id).main_success_scenario" :key="i">{{ step }}</li>
                            </ol>
                          </div>
                          <!-- Extensions -->
                          <div v-else-if="phase.id === 'extensions'" class="text-[11px] leading-relaxed text-slate-500 dark:text-slate-400 space-y-1.5">
                            <div v-if="phaseData(phase.id).preconditions?.length">
                              <p class="font-semibold text-slate-600 dark:text-slate-300">Preconditions</p>
                              <ul class="list-disc list-inside">
                                <li v-for="(p, i) in phaseData(phase.id).preconditions" :key="i">{{ p }}</li>
                              </ul>
                            </div>
                            <div v-if="phaseData(phase.id).extensions?.length">
                              <p class="font-semibold text-slate-600 dark:text-slate-300">Extensions</p>
                              <ul class="list-disc list-inside">
                                <li v-for="(e, i) in phaseData(phase.id).extensions" :key="i">{{ e }}</li>
                              </ul>
                            </div>
                          </div>
                          <!-- Acceptance Criteria -->
                          <div v-else-if="phase.id === 'acceptance_criteria'" class="text-[11px] leading-relaxed text-slate-500 dark:text-slate-400">
                            <ul class="list-disc list-inside space-y-0.5">
                              <li v-for="(c, i) in (phaseData(phase.id).acceptance_criteria || [])" :key="i">{{ c }}</li>
                            </ul>
                          </div>
                        </template>
                        <!-- In progress -->
                        <div v-else-if="phaseHasContent(phase.id)" class="flex items-center gap-2 py-1">
                          <Spinner v-if="phase.id === currentPhase" class="size-3 text-violet-400" />
                          <span class="text-[11px] text-slate-400 dark:text-slate-500">
                            {{ phase.id === currentPhase ? 'In progress...' : 'Completed' }}
                          </span>
                        </div>
                        <p v-else class="text-[11px] text-slate-300 dark:text-slate-600 italic">Not started</p>
                      </div>
                    </div>
                  </template>
                </div>
              </div>

              <!-- Right column: Current phase interaction -->
              <div class="flex-1 flex flex-col min-w-0 min-h-0">
                <!-- Current phase conversation -->
                <div class="flex-1 overflow-y-auto px-6 py-5 space-y-4">
                  <!-- Previous Q&A in this phase (compact) -->
                  <template v-for="(msg, idx) in previousExchanges" :key="msg.id">
                    <div v-if="msg.role === 'assistant'" class="text-sm text-slate-500 dark:text-slate-400 border-l-2 border-slate-200 dark:border-slate-700 pl-3 py-1">
                      <MarkdownContent :content="cleanContent(msg.content)" />
                    </div>
                    <div v-else class="text-sm text-slate-600 dark:text-slate-300 border-l-2 border-blue-200 dark:border-blue-800 pl-3 py-1">
                      {{ msg.content }}
                    </div>
                  </template>

                  <!-- Divider if there are previous exchanges -->
                  <div v-if="previousExchanges.length > 0 && latestAssistantMessage" class="border-t border-slate-100 dark:border-slate-800" />

                  <!-- Latest assistant message: the active question/prompt -->
                  <div v-if="latestAssistantMessage">
                    <!-- Proposal -->
                    <div v-if="latestAssistantMessage.message_type === 'proposal' && latestAssistantMessage.proposal" class="rounded-xl border border-violet-200 dark:border-violet-800 bg-violet-50/50 dark:bg-violet-950/20 p-6 space-y-4">
                      <div class="flex items-center gap-2">
                        <SparklesIcon class="size-4 text-violet-500" />
                        <span class="text-xs font-semibold text-violet-600 dark:text-violet-400 uppercase tracking-wide">Final Proposal</span>
                      </div>
                      <div class="space-y-3">
                        <p class="text-lg font-semibold text-slate-900 dark:text-slate-100">{{ latestAssistantMessage.proposal.title }}</p>
                        <div class="text-sm text-slate-600 dark:text-slate-300 prose prose-sm dark:prose-invert max-w-none">
                          <MarkdownContent :content="latestAssistantMessage.proposal.description" />
                        </div>
                      </div>
                      <div v-if="hasActiveSession && latestAssistantMessage === latestProposal" class="flex items-center gap-3 pt-3 border-t border-violet-200 dark:border-violet-800">
                        <Button variant="primary" :loading="acceptPending" @click="emit('accept')">
                          <CheckIcon class="size-4" />
                          Accept &amp; apply to issue
                        </Button>
                        <Button variant="secondary" :disabled="acceptPending" @click="inputRef?.focus()">
                          Request changes
                        </Button>
                      </div>
                    </div>

                    <!-- Regular question — rendered as a prompt card -->
                    <div v-else class="rounded-xl bg-slate-50 dark:bg-slate-800/60 border border-slate-100 dark:border-slate-700 p-5">
                      <div class="flex items-start gap-3">
                        <SparklesIcon class="size-4 text-violet-400 mt-0.5 shrink-0" />
                        <div class="text-sm text-slate-700 dark:text-slate-300 prose prose-sm dark:prose-invert max-w-none">
                          <MarkdownContent :content="cleanContent(latestAssistantMessage.content)" />
                        </div>
                      </div>
                    </div>
                  </div>

                  <!-- User's pending answer + waiting indicator -->
                  <template v-if="waitingForResponse">
                    <!-- Show the user's last message that's waiting for a response -->
                    <div
                      v-for="msg in currentPhaseMessages.filter(m => m.role === 'user' && currentPhaseMessages.indexOf(m) > currentPhaseMessages.lastIndexOf(latestAssistantMessage))"
                      :key="msg.id"
                      class="flex justify-end"
                    >
                      <div class="max-w-[75%] rounded-lg bg-blue-50 dark:bg-blue-950/40 px-4 py-3">
                        <p class="text-sm text-slate-800 dark:text-slate-200 whitespace-pre-wrap">{{ msg.content }}</p>
                      </div>
                    </div>
                    <div class="flex items-center gap-2.5 py-4">
                      <Spinner class="size-4 text-violet-400" />
                      <span class="text-sm text-slate-400 dark:text-slate-500">Hivemind is thinking...</span>
                    </div>
                  </template>

                  <div ref="messagesEnd" />
                </div>

                <!-- Footer: input + navigation (hidden when proposal is showing) -->
                <div v-if="showFooter" class="flex-shrink-0 border-t border-slate-100 dark:border-slate-800 px-6 py-4 space-y-3">
                  <!-- Confirmation buttons — shown when Hivemind asks "does this look right?" -->
                  <div v-if="isConfirmationQuestion && !waitingForResponse && !forceShowInput" class="flex items-center gap-2">
                    <Button
                      v-if="canGoBack"
                      variant="secondary"
                      :disabled="advancePending"
                      @click="emit('advance-phase', REFINEMENT_PHASES[currentPhaseIndex - 1].id)"
                    >
                      <ArrowLeftIcon class="size-4" />
                    </Button>
                    <div class="flex-1" />
                    <Button
                      variant="secondary"
                      @click="forceShowInput = true; nextTick(() => inputRef?.focus())"
                    >
                      Make changes
                    </Button>
                    <Button
                      v-if="!isLastPhase"
                      variant="primary"
                      :loading="advancePending"
                      @click="emit('advance-phase', null)"
                    >
                      <CheckIcon class="size-4" />
                      Confirm &amp; continue
                    </Button>
                    <Button
                      v-else
                      variant="primary"
                      :loading="sendPending"
                      @click="emit('send', 'Confirmed, looks good.')"
                    >
                      <CheckIcon class="size-4" />
                      Confirm
                    </Button>
                  </div>

                  <!-- Normal input row -->
                  <div v-else class="flex items-end gap-2">
                    <Button
                      v-if="canGoBack"
                      variant="secondary"
                      :disabled="advancePending"
                      @click="emit('advance-phase', REFINEMENT_PHASES[currentPhaseIndex - 1].id)"
                    >
                      <ArrowLeftIcon class="size-4" />
                    </Button>

                    <textarea
                      ref="inputRef"
                      v-model="messageInput"
                      rows="2"
                      placeholder="Your answer..."
                      class="flex-1 resize-none rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-sm text-slate-700 dark:text-slate-300 placeholder:text-slate-400 dark:placeholder:text-slate-500 px-4 py-3 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-violet-500"
                      :disabled="sendPending"
                      @keydown="onKeydown"
                    />

                    <Button
                      variant="secondary"
                      :disabled="!messageInput.trim() || sendPending"
                      :loading="sendPending"
                      @click="sendMessage"
                    >
                      <SendIcon class="size-4" />
                    </Button>

                    <Button
                      v-if="showNextPhase"
                      variant="primary"
                      :loading="advancePending"
                      @click="emit('advance-phase', null)"
                    >
                      {{ REFINEMENT_PHASES[currentPhaseIndex + 1]?.label }}
                      <ChevronRightIcon class="size-4" />
                    </Button>
                  </div>
                  <p v-if="!isConfirmationQuestion || waitingForResponse" class="text-[11px] text-slate-400 dark:text-slate-500">Ctrl+Enter to send</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
