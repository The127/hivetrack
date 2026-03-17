<!--
  HomeView — Personal dashboard ("My Work").

  The default view after login. Answers: what should I work on right now?

  Sections (each backed by its own TanStack Query):
    1. My open issues    — /api/v1/me/issues
    2. Triage inbox      — /api/v1/projects/:slug/triage (cross-project, aggregated)
    3. Projects          — /api/v1/projects

  Keyboard shortcuts (from MainLayout):
    C  → create new issue
-->
<script setup>
import { ref } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'
import { PlusIcon, InboxIcon, FolderKanbanIcon, CircleDotIcon } from 'lucide-vue-next'
import MainLayout from '@/layouts/MainLayout.vue'
import Avatar from '@/components/ui/Avatar.vue'
import Badge from '@/components/ui/Badge.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import Spinner from '@/components/ui/Spinner.vue'
import CreateProjectModal from '@/components/project/CreateProjectModal.vue'
import { apiFetch } from '@/composables/useApi'
import { useAuth } from '@/composables/useAuth'

const { user } = useAuth()
const router = useRouter()

const showCreateProject = ref(false)

function onProjectCreated(result) {
  showCreateProject.value = false
  router.push(`/projects/${result.slug}/board`)
}

const userName = user.value?.profile?.name ?? user.value?.profile?.email ?? 'You'

// ── Queries ───────────────────────────────────────────────────────────────────

const { data: myIssues, isLoading: loadingIssues } = useQuery({
  queryKey: ['me', 'issues'],
  queryFn: () => apiFetch('/api/v1/me/issues'),
})

const { data: projects, isLoading: loadingProjects } = useQuery({
  queryKey: ['projects'],
  queryFn: () => apiFetch('/api/v1/projects'),
})

// ── Status display helpers ────────────────────────────────────────────────────

const STATUS_SCHEME = {
  todo: 'gray',
  in_progress: 'blue',
  in_review: 'violet',
  done: 'green',
  cancelled: 'gray',
  open: 'sky',
  resolved: 'teal',
  closed: 'gray',
}

const PRIORITY_SCHEME = {
  none: 'gray',
  low: 'sky',
  medium: 'amber',
  high: 'orange',
  critical: 'red',
}

function statusScheme(status) {
  return STATUS_SCHEME[status] ?? 'gray'
}

function priorityScheme(priority) {
  return PRIORITY_SCHEME[priority] ?? 'gray'
}

// Format status string for display: "in_progress" → "In progress"
function formatStatus(s) {
  return s.replace(/_/g, ' ').replace(/^\w/, (c) => c.toUpperCase())
}
</script>

<template>
  <MainLayout>
    <div class="max-w-3xl mx-auto px-6 py-8">
      <!-- Page header -->
      <div class="flex items-center justify-between mb-8">
        <div>
          <h1 class="text-xl font-semibold text-slate-900">My Work</h1>
          <p class="text-sm text-slate-500 mt-0.5">
            Welcome back, {{ userName }}
          </p>
        </div>
        <button
          class="inline-flex items-center gap-1.5 rounded-md bg-blue-600 px-3 h-8 text-sm font-medium text-white hover:bg-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-1 transition-colors cursor-pointer"
          title="Create issue (C)"
        >
          <PlusIcon class="size-4" />
          New issue
        </button>
        <button
          class="inline-flex items-center gap-1.5 rounded-md border border-slate-300 bg-white px-3 h-8 text-sm font-medium text-slate-700 hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-1 transition-colors cursor-pointer"
          @click="showCreateProject = true"
        >
          <PlusIcon class="size-4" />
          New project
        </button>
      </div>

      <!-- ── My open issues ────────────────────────────────────────────── -->
      <section class="mb-8">
        <h2 class="text-sm font-medium text-slate-700 mb-3 flex items-center gap-2">
          <CircleDotIcon class="size-4 text-blue-500" />
          My open issues
          <span
            v-if="myIssues?.items?.length"
            class="text-xs font-normal text-slate-500"
          >
            {{ myIssues.items.length }}
          </span>
        </h2>

        <div v-if="loadingIssues" class="flex justify-center py-8">
          <Spinner class="size-5 text-slate-400" />
        </div>

        <div
          v-else-if="myIssues?.items?.length"
          class="rounded-lg border border-slate-200 divide-y divide-slate-100 overflow-hidden"
        >
          <div
            v-for="issue in myIssues.items"
            :key="issue.id"
            class="flex items-center gap-3 px-4 py-2.5 hover:bg-slate-50 transition-colors cursor-pointer group"
          >
            <!-- Issue number -->
            <span class="text-xs font-mono text-slate-400 flex-shrink-0 w-14 text-right">
              {{ issue.project_slug?.toUpperCase() }}-{{ issue.number }}
            </span>

            <!-- Title -->
            <span class="flex-1 min-w-0 text-sm text-slate-800 truncate group-hover:text-slate-900">
              {{ issue.title }}
            </span>

            <!-- Priority -->
            <Badge
              v-if="issue.priority && issue.priority !== 'none'"
              :colorScheme="priorityScheme(issue.priority)"
              compact
            >
              {{ issue.priority }}
            </Badge>

            <!-- Status -->
            <Badge :colorScheme="statusScheme(issue.status)" compact>
              {{ formatStatus(issue.status) }}
            </Badge>

            <!-- Assignees -->
            <div v-if="issue.assignees?.length" class="flex -space-x-1">
              <Avatar
                v-for="a in issue.assignees.slice(0, 3)"
                :key="a.id"
                :name="a.display_name"
                :src="a.avatar_url"
                size="xs"
                class="ring-1 ring-white"
              />
            </div>
          </div>
        </div>

        <EmptyState
          v-else
          title="No open issues assigned to you"
          description="Issues assigned to you across all projects will appear here."
        >
          <template #icon>
            <CircleDotIcon class="size-8" />
          </template>
        </EmptyState>
      </section>

      <!-- ── Projects ──────────────────────────────────────────────────── -->
      <section>
        <h2 class="text-sm font-medium text-slate-700 mb-3 flex items-center gap-2">
          <FolderKanbanIcon class="size-4 text-slate-500" />
          Projects
          <span
            v-if="projects?.items?.length"
            class="text-xs font-normal text-slate-500"
          >
            {{ projects.items.length }}
          </span>
        </h2>

        <div v-if="loadingProjects" class="flex justify-center py-8">
          <Spinner class="size-5 text-slate-400" />
        </div>

        <div
          v-else-if="projects?.items?.length"
          class="rounded-lg border border-slate-200 divide-y divide-slate-100 overflow-hidden"
        >
          <RouterLink
            v-for="project in projects.items"
            :key="project.id"
            :to="`/projects/${project.slug}/board`"
            class="flex items-center gap-3 px-4 py-3 hover:bg-slate-50 transition-colors group"
          >
            <!-- Project initial -->
            <span
              class="size-7 rounded flex items-center justify-center text-xs font-semibold bg-slate-100 text-slate-600 flex-shrink-0"
            >
              {{ project.slug.slice(0, 2).toUpperCase() }}
            </span>

            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-slate-800 group-hover:text-slate-900 truncate">
                {{ project.name }}
              </p>
              <p class="text-xs text-slate-500">{{ project.slug }}</p>
            </div>

            <Badge :colorScheme="project.archetype === 'software' ? 'blue' : 'teal'" compact>
              {{ project.archetype }}
            </Badge>

            <span v-if="project.archived" class="text-xs text-slate-400 italic">archived</span>
          </RouterLink>
        </div>

        <EmptyState
          v-else
          title="No projects yet"
          description="Create your first project to start tracking work."
          action-label="New project"
          @action="showCreateProject = true"
        >
          <template #icon>
            <FolderKanbanIcon class="size-8" />
          </template>
        </EmptyState>
      </section>

      <!-- ── Triage inbox hint ──────────────────────────────────────────── -->
      <div
        v-if="projects?.items?.length"
        class="mt-6 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 flex items-center gap-3"
      >
        <InboxIcon class="size-4 text-amber-600 flex-shrink-0" />
        <p class="text-sm text-amber-800">
          Open a project and go to
          <strong>Triage</strong> to review incoming issues.
        </p>
      </div>
    </div>
  </MainLayout>

  <CreateProjectModal
    :open="showCreateProject"
    @close="showCreateProject = false"
    @created="onProjectCreated"
  />
</template>
