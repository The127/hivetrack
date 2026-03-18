<!--
  IssueDetailView — full detail page for a single issue.

  Route: /projects/:slug/issues/:number

  For epics: shows EpicChildList with progress bar.
  For tasks: shows EpicSelector to assign/change/clear parent epic.
-->
<script setup>
import { computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import {
  ArrowLeftIcon,
  LayersIcon,
} from 'lucide-vue-next'
import MainLayout from '@/layouts/MainLayout.vue'
import Badge from '@/components/ui/Badge.vue'
import Spinner from '@/components/ui/Spinner.vue'
import EpicSelector from '@/components/issue/EpicSelector.vue'
import EpicChildList from '@/components/issue/EpicChildList.vue'
import CommentSection from '@/components/issue/CommentSection.vue'
import StatusSelect from '@/components/issue/StatusSelect.vue'
import PrioritySelect from '@/components/issue/PrioritySelect.vue'
import AssigneeSelect from '@/components/issue/AssigneeSelect.vue'
import { fetchIssue, updateIssue } from '@/api/issues'
import { fetchProject } from '@/api/projects'

const route = useRoute()
const queryClient = useQueryClient()

const slug = computed(() => route.params.slug)
const number = computed(() => Number(route.params.number))

// ── Data ──────────────────────────────────────────────────────────────────

const { data: project } = useQuery({
  queryKey: ['project', slug],
  queryFn: () => fetchProject(slug.value),
})

const { data: issue, isLoading } = useQuery({
  queryKey: ['issue', slug, number],
  queryFn: () => fetchIssue(slug.value, number.value),
  enabled: computed(() => !!slug.value && !!number.value),
})

const ESTIMATE_LABEL = { none: null, xs: 'XS', s: 'S', m: 'M', l: 'L', xl: 'XL' }

// ── Status mutation ───────────────────────────────────────────────────────────

const { mutate: updateStatus } = useMutation({
  mutationFn: (status) => updateIssue(slug.value, number.value, { status }),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['issue', slug.value, number.value] })
    queryClient.invalidateQueries({ queryKey: ['issues', slug.value] })
  },
})

// ── Priority mutation ─────────────────────────────────────────────────────────

const { mutate: updatePriority } = useMutation({
  mutationFn: (priority) => updateIssue(slug.value, number.value, { priority }),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['issue', slug.value, number.value] })
    queryClient.invalidateQueries({ queryKey: ['issues', slug.value] })
  },
})

// ── Epic assignment mutation (for tasks) ────────────────────────────────────

const { mutate: updateParent } = useMutation({
  mutationFn: (parentId) => updateIssue(slug.value, number.value, { parent_id: parentId }),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['issue', slug.value, number.value] })
    queryClient.invalidateQueries({ queryKey: ['issues', slug.value] })
  },
})

// ── Assignee mutation ─────────────────────────────────────────────────────────

const { mutate: updateAssignees } = useMutation({
  mutationFn: (assigneeIds) => updateIssue(slug.value, number.value, { assignee_ids: assigneeIds }),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['issue', slug.value, number.value] })
    queryClient.invalidateQueries({ queryKey: ['issues', slug.value] })
    queryClient.invalidateQueries({ queryKey: ['me', 'issues'] })
  },
})
</script>

<template>
  <MainLayout>
    <div class="flex flex-col h-full">
      <!-- Header -->
      <div class="flex-shrink-0 flex items-center gap-3 px-6 py-3 border-b border-slate-200 bg-white">
        <RouterLink
          :to="`/projects/${slug}/backlog`"
          class="inline-flex items-center gap-1 text-sm text-slate-500 hover:text-slate-700 transition-colors"
        >
          <ArrowLeftIcon class="size-4" />
          Back
        </RouterLink>
        <div v-if="project" class="flex items-center gap-2 text-slate-400">
          <span class="size-6 rounded flex items-center justify-center text-[10px] font-semibold bg-slate-100 text-slate-600">
            {{ project.slug.slice(0, 2).toUpperCase() }}
          </span>
          <span class="text-sm font-medium text-slate-600">{{ project.name }}</span>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="isLoading" class="flex-1 flex items-center justify-center">
        <Spinner class="size-6 text-slate-400" />
      </div>

      <!-- Content -->
      <div v-else-if="issue" class="flex-1 overflow-y-auto">
        <div class="max-w-3xl mx-auto px-6 py-8 space-y-8">

          <!-- Issue header -->
          <div class="space-y-3">
            <div class="flex items-center gap-2">
              <span class="text-xs font-mono text-slate-400">{{ slug.toUpperCase() }}-{{ issue.number }}</span>
              <Badge v-if="issue.type === 'epic'" colorScheme="violet" compact>
                <LayersIcon class="size-3" />
                Epic
              </Badge>
              <Badge v-else colorScheme="blue" compact>Task</Badge>
            </div>
            <h1 class="text-2xl font-semibold text-slate-900">{{ issue.title }}</h1>
          </div>

          <!-- Metadata grid -->
          <div class="grid grid-cols-2 gap-x-8 gap-y-4">
            <!-- Status -->
            <div class="space-y-1">
              <span class="text-xs font-medium text-slate-500">Status</span>
              <div class="pt-1">
                <StatusSelect
                  :status="issue.status"
                  :archetype="project?.archetype ?? 'software'"
                  @update:status="updateStatus"
                />
              </div>
            </div>

            <!-- Priority -->
            <div class="space-y-1">
              <span class="text-xs font-medium text-slate-500">Priority</span>
              <div class="pt-1">
                <PrioritySelect
                  :priority="issue.priority ?? 'none'"
                  @update:priority="updatePriority"
                />
              </div>
            </div>

            <!-- Estimate -->
            <div class="space-y-1">
              <span class="text-xs font-medium text-slate-500">Estimate</span>
              <div>
                <span v-if="ESTIMATE_LABEL[issue.estimate]" class="text-sm font-medium text-slate-600 bg-slate-100 px-2 py-0.5 rounded">
                  {{ ESTIMATE_LABEL[issue.estimate] }}
                </span>
                <span v-else class="text-sm text-slate-400">None</span>
              </div>
            </div>

            <!-- Assignees -->
            <div class="space-y-1">
              <div class="max-w-xs">
                <AssigneeSelect
                  :project-slug="slug"
                  :model-value="issue.assignees ?? []"
                  @update:model-value="updateAssignees"
                />
              </div>
            </div>

            <!-- On hold -->
            <div v-if="issue.on_hold" class="space-y-1">
              <span class="text-xs font-medium text-slate-500">On Hold</span>
              <div class="flex items-center gap-2">
                <Badge colorScheme="amber" compact>{{ issue.hold_reason ?? 'on hold' }}</Badge>
                <span v-if="issue.hold_note" class="text-xs text-slate-500 italic">{{ issue.hold_note }}</span>
              </div>
            </div>
          </div>

          <!-- Description -->
          <div v-if="issue.description" class="space-y-2">
            <h2 class="text-sm font-medium text-slate-700">Description</h2>
            <div class="prose prose-sm prose-slate max-w-none text-slate-600">
              {{ issue.description }}
            </div>
          </div>

          <!-- Epic selector (for tasks) -->
          <div v-if="issue.type === 'task'" class="max-w-xs">
            <EpicSelector
              :project-slug="slug"
              :model-value="issue.parent_id"
              @update:model-value="updateParent"
            />
          </div>

          <!-- Child tasks (for epics) -->
          <div v-if="issue.type === 'epic'">
            <EpicChildList
              :project-slug="slug"
              :epic-id="issue.id"
              :archetype="project?.archetype ?? 'software'"
              :child-count="issue.child_count"
              :child-done-count="issue.child_done_count"
            />
          </div>

          <!-- Checklist (for tasks) -->
          <div v-if="issue.type === 'task' && issue.checklist?.length" class="space-y-2">
            <h2 class="text-sm font-medium text-slate-700">Checklist</h2>
            <div class="space-y-1">
              <div v-for="item in issue.checklist" :key="item.id" class="flex items-center gap-2">
                <input type="checkbox" :checked="item.done" disabled class="rounded border-slate-300" />
                <span class="text-sm" :class="item.done ? 'text-slate-400 line-through' : 'text-slate-700'">{{ item.text }}</span>
              </div>
            </div>
          </div>

          <!-- Comments -->
          <CommentSection :project-slug="slug" :issue-number="number" />

        </div>
      </div>

      <!-- Not found -->
      <div v-else class="flex-1 flex items-center justify-center">
        <p class="text-sm text-slate-400">Issue not found.</p>
      </div>
    </div>
  </MainLayout>
</template>
