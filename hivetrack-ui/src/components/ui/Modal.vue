<!--
  Modal — accessible dialog overlay.

  Props:
    open        Boolean — controls visibility (controlled component)
    title       String  — dialog heading
    description String  — optional subtitle below title

  Events:
    close — emitted when the overlay or Escape key is pressed

  Slots:
    default — dialog body
    footer  — action buttons (right-aligned)

  @example
  <Modal :open="showModal" title="Create project" @close="showModal = false">
    <p>Body content here.</p>
    <template #footer>
      <Button variant="secondary" @click="showModal = false">Cancel</Button>
      <Button @click="submit">Create</Button>
    </template>
  </Modal>
-->
<script setup>
import { onMounted, onUnmounted } from 'vue'
import { XIcon } from 'lucide-vue-next'

const props = defineProps({
  open: {
    type: Boolean,
    required: true,
  },
  title: {
    type: String,
    required: true,
  },
  description: {
    type: String,
    default: null,
  },
})

const emit = defineEmits(['close'])

function onKeydown(e) {
  if (e.key === 'Escape' && props.open) emit('close')
}

onMounted(() => document.addEventListener('keydown', onKeydown))
onUnmounted(() => document.removeEventListener('keydown', onKeydown))
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
      <div
        v-if="open"
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
      >
        <!-- Backdrop -->
        <div
          class="absolute inset-0 bg-black/40 dark:bg-black/60"
          aria-hidden="true"
          @click="emit('close')"
        />

        <!-- Panel -->
        <div
          role="dialog"
          :aria-label="title"
          class="relative z-10 w-full max-w-lg rounded-xl bg-white dark:bg-slate-900 shadow-xl ring-1 ring-slate-900/10 dark:ring-slate-700"
        >
          <!-- Header -->
          <div class="flex items-start justify-between gap-4 px-6 pt-5 pb-4 border-b border-slate-100 dark:border-slate-800">
            <div>
              <h2 class="text-base font-semibold text-slate-900 dark:text-slate-100">{{ title }}</h2>
              <p v-if="description" class="mt-0.5 text-sm text-slate-500 dark:text-slate-400">{{ description }}</p>
            </div>
            <button
              class="flex-shrink-0 rounded-md p-1 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors cursor-pointer"
              @click="emit('close')"
            >
              <XIcon class="size-4" />
            </button>
          </div>

          <!-- Body -->
          <div class="px-6 py-5">
            <slot />
          </div>

          <!-- Footer -->
          <div
            v-if="$slots.footer"
            class="flex items-center justify-end gap-2 px-6 py-4 border-t border-slate-100 dark:border-slate-800"
          >
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
