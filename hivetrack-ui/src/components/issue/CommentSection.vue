<!--
  CommentSection — displays and manages comments on an issue.

  Shows a chronological list of comments with author info and timestamps,
  plus a form to add new comments.

  @prop {String} projectSlug — project slug for API calls
  @prop {Number} issueNumber — issue number for API calls
-->
<script setup>
import { ref } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { fetchComments, createComment, updateComment, deleteComment } from '@/api/comments'
import Avatar from '@/components/ui/Avatar.vue'
import Button from '@/components/ui/Button.vue'
import MarkdownContent from '@/components/ui/MarkdownContent.vue'

const props = defineProps({
  projectSlug: { type: String, required: true },
  issueNumber: { type: Number, required: true },
})

const queryClient = useQueryClient()
const queryKey = ['comments', props.projectSlug, props.issueNumber]

const { data: commentsData, isLoading } = useQuery({
  queryKey,
  queryFn: () => fetchComments(props.projectSlug, props.issueNumber),
})

// ── New comment ─────────────────────────────────────────────────────────────

const newBody = ref('')

const { mutate: submitComment, isPending: isSubmitting } = useMutation({
  mutationFn: (body) => createComment(props.projectSlug, props.issueNumber, { body }),
  onSuccess: () => {
    newBody.value = ''
    queryClient.invalidateQueries({ queryKey })
  },
})

function handleSubmit() {
  const body = newBody.value.trim()
  if (!body) return
  submitComment(body)
}

// ── Edit comment ────────────────────────────────────────────────────────────

const editingId = ref(null)
const editBody = ref('')

function startEdit(comment) {
  editingId.value = comment.id
  editBody.value = comment.body
}

function cancelEdit() {
  editingId.value = null
  editBody.value = ''
}

const { mutate: saveEdit } = useMutation({
  mutationFn: ({ id, body }) => updateComment(props.projectSlug, props.issueNumber, id, { body }),
  onSuccess: () => {
    editingId.value = null
    editBody.value = ''
    queryClient.invalidateQueries({ queryKey })
  },
})

function handleSaveEdit(id) {
  const body = editBody.value.trim()
  if (!body) return
  saveEdit({ id, body })
}

// ── Delete comment ──────────────────────────────────────────────────────────

const { mutate: removeComment } = useMutation({
  mutationFn: (id) => deleteComment(props.projectSlug, props.issueNumber, id),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey })
  },
})

// ── Formatting ──────────────────────────────────────────────────────────────

function relativeTime(dateStr) {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now - date
  const diffMin = Math.floor(diffMs / 60000)
  if (diffMin < 1) return 'just now'
  if (diffMin < 60) return `${diffMin}m ago`
  const diffHr = Math.floor(diffMin / 60)
  if (diffHr < 24) return `${diffHr}h ago`
  const diffDay = Math.floor(diffHr / 24)
  if (diffDay < 30) return `${diffDay}d ago`
  return date.toLocaleDateString()
}
</script>

<template>
  <div class="space-y-4">
    <h2 class="text-sm font-medium text-slate-700">Comments</h2>

    <!-- Loading -->
    <div v-if="isLoading" class="text-sm text-slate-400">Loading comments...</div>

    <!-- Comment list -->
    <div v-else-if="commentsData?.items?.length" class="space-y-4">
      <div
        v-for="comment in commentsData.items"
        :key="comment.id"
        class="flex gap-3"
      >
        <Avatar
          :name="comment.author_name || comment.author_email || 'User'"
          :src="comment.avatar_url"
          size="sm"
          class="mt-0.5"
        />
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-slate-700">
              {{ comment.author_name || comment.author_email || 'User' }}
            </span>
            <span class="text-xs text-slate-400">{{ relativeTime(comment.created_at) }}</span>
            <span v-if="comment.updated_at !== comment.created_at" class="text-xs text-slate-400 italic">(edited)</span>
          </div>

          <!-- Editing mode -->
          <div v-if="editingId === comment.id" class="mt-1 space-y-2">
            <textarea
              v-model="editBody"
              rows="3"
              class="w-full rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-700 placeholder-slate-400 focus:border-blue-400 focus:ring-1 focus:ring-blue-400 outline-none resize-none"
            />
            <div class="flex gap-2">
              <Button size="sm" @click="handleSaveEdit(comment.id)">Save</Button>
              <Button size="sm" variant="ghost" @click="cancelEdit">Cancel</Button>
            </div>
          </div>

          <!-- Display mode -->
          <div v-else>
            <MarkdownContent :content="comment.body" class="mt-0.5" />
            <div class="flex gap-2 mt-1">
              <button
                class="text-xs text-slate-400 hover:text-slate-600 transition-colors"
                @click="startEdit(comment)"
              >
                Edit
              </button>
              <button
                class="text-xs text-slate-400 hover:text-red-500 transition-colors"
                @click="removeComment(comment.id)"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <p v-else class="text-sm text-slate-400">No comments yet.</p>

    <!-- New comment form -->
    <div class="space-y-2 pt-2 border-t border-slate-100">
      <textarea
        v-model="newBody"
        rows="3"
        placeholder="Add a comment..."
        class="w-full rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-700 placeholder-slate-400 focus:border-blue-400 focus:ring-1 focus:ring-blue-400 outline-none resize-none"
        @keydown.meta.enter="handleSubmit"
        @keydown.ctrl.enter="handleSubmit"
      />
      <div class="flex items-center justify-between">
        <span class="text-xs text-slate-400">Ctrl+Enter to submit</span>
        <Button size="sm" :loading="isSubmitting" :disabled="!newBody.trim()" @click="handleSubmit">
          Comment
        </Button>
      </div>
    </div>
  </div>
</template>
