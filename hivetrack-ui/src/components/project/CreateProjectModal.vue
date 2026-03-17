<!--
  CreateProjectModal — form to create a new project.

  Emits:
    close   — close without creating
    created — project was created; payload: { id, slug }
-->
<script setup>
import { ref, watch, computed } from 'vue'
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { CodeIcon, HeadphonesIcon } from 'lucide-vue-next'
import Modal from '@/components/ui/Modal.vue'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import { createProject } from '@/api/projects'

const props = defineProps({
  open: {
    type: Boolean,
    required: true,
  },
})

const emit = defineEmits(['close', 'created'])

const queryClient = useQueryClient()

// ── Form state ─────────────────────────────────────────────────────────────

const name = ref('')
const slug = ref('')
const archetype = ref('software')
const description = ref('')
const slugManuallyEdited = ref(false)
const errors = ref({})

// Auto-generate slug from name unless user has edited it manually
watch(name, (val) => {
  if (!slugManuallyEdited.value) {
    slug.value = val
      .toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-]/g, '')
      .replace(/-+/g, '-')
      .replace(/^-|-$/g, '')
      .slice(0, 20)
  }
})

function onSlugInput(val) {
  slugManuallyEdited.value = true
  slug.value = val
    .toLowerCase()
    .replace(/[^a-z0-9-]/g, '')
    .replace(/-+/g, '-')
}

// ── Reset when closed ───────────────────────────────────────────────────────

watch(
  () => props.open,
  (open) => {
    if (!open) {
      name.value = ''
      slug.value = ''
      archetype.value = 'software'
      description.value = ''
      slugManuallyEdited.value = false
      errors.value = {}
    }
  },
)

// ── Validation ──────────────────────────────────────────────────────────────

function validate() {
  const e = {}
  if (!name.value.trim()) e.name = 'Name is required.'
  if (!slug.value.trim()) e.slug = 'Slug is required.'
  else if (!/^[a-z0-9-]+$/.test(slug.value)) e.slug = 'Only lowercase letters, numbers, and hyphens.'
  errors.value = e
  return Object.keys(e).length === 0
}

// ── Mutation ────────────────────────────────────────────────────────────────

const { mutate, isPending, error: serverError } = useMutation({
  mutationFn: (data) => createProject(data),
  onSuccess: (result) => {
    queryClient.invalidateQueries({ queryKey: ['projects'] })
    emit('created', result)
  },
})

const submitError = computed(() => {
  if (!serverError.value) return null
  // Try to extract a human-readable message from the API error shape
  return serverError.value?.errors?.[0]?.message ?? 'Something went wrong. Please try again.'
})

function submit() {
  if (!validate()) return
  mutate({
    name: name.value.trim(),
    slug: slug.value.trim(),
    archetype: archetype.value,
    description: description.value.trim() || undefined,
  })
}
</script>

<template>
  <Modal
    :open="open"
    title="New project"
    description="Projects group issues, sprints, and milestones for your team."
    @close="emit('close')"
  >
    <form class="flex flex-col gap-5" @submit.prevent="submit">
      <!-- Name -->
      <Input
        label="Name"
        v-model="name"
        placeholder="e.g. Platform, Infra, Customer Support"
        :error="errors.name"
        autofocus
        required
      />

      <!-- Slug -->
      <Input
        label="Slug"
        :model-value="slug"
        @update:model-value="onSlugInput"
        placeholder="e.g. platform"
        :error="errors.slug"
        hint="Short identifier used in issue numbers and URLs. Lowercase, hyphens allowed."
        required
      />

      <!-- Archetype -->
      <div class="flex flex-col gap-1.5">
        <span class="text-sm font-medium text-slate-700">Type</span>
        <div class="grid grid-cols-2 gap-2">
          <button
            type="button"
            :class="[
              'flex flex-col items-start gap-1.5 rounded-lg border p-3 text-left transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500',
              archetype === 'software'
                ? 'border-blue-500 bg-blue-50 ring-1 ring-blue-500'
                : 'border-slate-200 hover:border-slate-300 hover:bg-slate-50',
            ]"
            @click="archetype = 'software'"
          >
            <CodeIcon
              :class="['size-4', archetype === 'software' ? 'text-blue-600' : 'text-slate-400']"
            />
            <div>
              <p
                :class="[
                  'text-sm font-medium',
                  archetype === 'software' ? 'text-blue-700' : 'text-slate-700',
                ]"
              >
                Software
              </p>
              <p class="text-xs text-slate-500">Sprints, backlog, board</p>
            </div>
          </button>

          <button
            type="button"
            :class="[
              'flex flex-col items-start gap-1.5 rounded-lg border p-3 text-left transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500',
              archetype === 'support'
                ? 'border-teal-500 bg-teal-50 ring-1 ring-teal-500'
                : 'border-slate-200 hover:border-slate-300 hover:bg-slate-50',
            ]"
            @click="archetype = 'support'"
          >
            <HeadphonesIcon
              :class="['size-4', archetype === 'support' ? 'text-teal-600' : 'text-slate-400']"
            />
            <div>
              <p
                :class="[
                  'text-sm font-medium',
                  archetype === 'support' ? 'text-teal-700' : 'text-slate-700',
                ]"
              >
                Support
              </p>
              <p class="text-xs text-slate-500">Email intake, token tracking</p>
            </div>
          </button>
        </div>
      </div>

      <!-- Description (optional) -->
      <div class="flex flex-col gap-1">
        <label class="text-sm font-medium text-slate-700">Description <span class="text-slate-400 font-normal">(optional)</span></label>
        <textarea
          v-model="description"
          rows="2"
          placeholder="What is this project for?"
          class="block w-full rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-offset-0 focus:border-blue-400 focus:ring-blue-200 resize-none transition-colors"
        />
      </div>

      <!-- Server error -->
      <p v-if="submitError" class="text-sm text-red-600">{{ submitError }}</p>
    </form>

    <template #footer>
      <Button variant="secondary" @click="emit('close')">Cancel</Button>
      <Button :loading="isPending" @click="submit">Create project</Button>
    </template>
  </Modal>
</template>
