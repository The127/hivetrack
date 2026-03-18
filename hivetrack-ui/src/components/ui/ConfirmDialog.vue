<!--
  ConfirmDialog — reusable destructive action confirmation modal.

  Props:
    open        — controls visibility
    title       — dialog heading
    message     — body copy specific to the action
    confirmText — label for the confirm button (default: "Confirm")
    loading     — shows spinner on confirm button while mutation runs

  Events:
    confirm — user clicked the destructive confirm button
    cancel  — user dismissed the dialog (Escape / Cancel / backdrop)

  Usage:
    <ConfirmDialog
      :open="showDialog"
      title="Cancel this issue?"
      message="This issue will be moved to cancelled and won't appear on the board."
      confirm-text="Cancel issue"
      :loading="isPending"
      @confirm="doCancel"
      @cancel="showDialog = false"
    />
-->
<script setup>
import Modal from '@/components/ui/Modal.vue'
import Button from '@/components/ui/Button.vue'

defineProps({
  open: { type: Boolean, required: true },
  title: { type: String, required: true },
  message: { type: String, required: true },
  confirmText: { type: String, default: 'Confirm' },
  loading: { type: Boolean, default: false },
})

const emit = defineEmits(['confirm', 'cancel'])
</script>

<template>
  <Modal :open="open" :title="title" @close="emit('cancel')">
    <p class="text-sm text-slate-600">{{ message }}</p>

    <template #footer>
      <Button variant="secondary" :disabled="loading" @click="emit('cancel')">
        Keep it
      </Button>
      <Button variant="destructive" :loading="loading" @click="emit('confirm')">
        {{ confirmText }}
      </Button>
    </template>
  </Modal>
</template>
