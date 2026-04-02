<!--
  ProjectsView — list of all projects the current user has access to.

  Shows projects as a list with name, slug, archetype, description, and
  archived status. Clicking a project navigates to its board. A "New project"
  button opens the CreateProjectModal.
-->
<script setup>
import { ref, computed } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { useRouter } from 'vue-router'
import {
  PlusIcon,
  FolderKanbanIcon,
  ArchiveIcon,
  CodeIcon,
  HeadphonesIcon,
} from 'lucide-vue-next'
import MainLayout from '@/layouts/MainLayout.vue'
import Badge from '@/components/ui/Badge.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import Spinner from '@/components/ui/Spinner.vue'
import CreateProjectModal from '@/components/project/CreateProjectModal.vue'
import { fetchProjects } from '@/api/projects'

const router = useRouter()

const showCreateProject = ref(false)
const showArchived = ref(false)

function onProjectCreated(result) {
  showCreateProject.value = false
  router.push(`/projects/${result.slug}/board`)
}

// ── Query ──────────────────────────────────────────────────────────────────

const { data: projects, isLoading } = useQuery({
  queryKey: ['projects'],
  queryFn: fetchProjects,
})

const activeProjects = computed(() =>
  projects.value?.items?.filter((p) => !p.archived) ?? [],
)

const archivedProjects = computed(() =>
  projects.value?.items?.filter((p) => p.archived) ?? [],
)
</script>

<template>
  <MainLayout>
    <div class="max-w-3xl mx-auto px-6 py-8">
      <!-- Page header -->
      <div class="mb-6 flex items-start justify-between">
        <div>
          <h1 class="text-xl font-semibold text-slate-900 dark:text-slate-100">Projects</h1>
          <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            All projects you have access to.
          </p>
        </div>
        <button
          class="inline-flex items-center gap-1.5 rounded-md bg-blue-600 px-3 h-8 text-sm font-medium text-white hover:bg-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-1 transition-colors cursor-pointer"
          @click="showCreateProject = true"
        >
          <PlusIcon class="size-4" />
          New project
        </button>
      </div>

      <!-- Loading -->
      <div v-if="isLoading" class="h-32 flex items-center justify-center">
        <Spinner class="size-5 text-slate-400" />
      </div>

      <!-- Has projects -->
      <template v-else-if="projects?.items?.length">
        <!-- Active projects -->
        <section>
          <h2 class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-3 flex items-center gap-2">
            <FolderKanbanIcon class="size-4 text-slate-500 dark:text-slate-400" />
            Active
            <span v-if="activeProjects.length" class="text-xs font-normal text-slate-500">
              {{ activeProjects.length }}
            </span>
          </h2>

          <div
            v-if="activeProjects.length"
            class="rounded-lg border border-slate-200 dark:border-slate-700 divide-y divide-slate-100 dark:divide-slate-800 overflow-hidden"
          >
            <RouterLink
              v-for="project in activeProjects"
              :key="project.id"
              :to="`/projects/${project.slug}/board`"
              class="flex items-center gap-3 px-4 py-3 bg-white dark:bg-slate-900 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors group"
            >
              <!-- Project initial -->
              <span
                class="size-8 rounded-md flex items-center justify-center text-xs font-semibold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 flex-shrink-0"
              >
                {{ project.slug.slice(0, 2).toUpperCase() }}
              </span>

              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-slate-800 dark:text-slate-200 group-hover:text-slate-900 dark:group-hover:text-slate-100 truncate">
                  {{ project.name }}
                </p>
                <p class="text-xs text-slate-500 truncate">
                  {{ project.slug }}
                  <span v-if="project.description" class="text-slate-400">
                    &middot; {{ project.description }}
                  </span>
                </p>
              </div>

              <Badge :colorScheme="project.archetype === 'software' ? 'blue' : 'teal'" compact>
                <component
                  :is="project.archetype === 'software' ? CodeIcon : HeadphonesIcon"
                  class="size-3 mr-0.5"
                />
                {{ project.archetype }}
              </Badge>
            </RouterLink>
          </div>

          <p v-else class="text-sm text-slate-500 dark:text-slate-400 py-4">
            No active projects. Create one to get started.
          </p>
        </section>

        <!-- Archived projects -->
        <section v-if="archivedProjects.length" class="mt-6">
          <button
            class="text-sm font-medium text-slate-500 mb-3 flex items-center gap-2 hover:text-slate-700 transition-colors cursor-pointer"
            @click="showArchived = !showArchived"
          >
            <ArchiveIcon class="size-4" />
            Archived
            <span class="text-xs font-normal">{{ archivedProjects.length }}</span>
            <span class="text-xs">{{ showArchived ? '(hide)' : '(show)' }}</span>
          </button>

          <div
            v-if="showArchived"
            class="rounded-lg border border-slate-200 dark:border-slate-700 divide-y divide-slate-100 dark:divide-slate-800 overflow-hidden opacity-75"
          >
            <RouterLink
              v-for="project in archivedProjects"
              :key="project.id"
              :to="`/projects/${project.slug}/board`"
              class="flex items-center gap-3 px-4 py-3 bg-white dark:bg-slate-900 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors group"
            >
              <span
                class="size-8 rounded-md flex items-center justify-center text-xs font-semibold bg-slate-100 text-slate-500 flex-shrink-0"
              >
                {{ project.slug.slice(0, 2).toUpperCase() }}
              </span>

              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-slate-600 dark:text-slate-400 group-hover:text-slate-700 dark:group-hover:text-slate-300 truncate">
                  {{ project.name }}
                </p>
                <p class="text-xs text-slate-400 truncate">
                  {{ project.slug }}
                  <span v-if="project.description"> &middot; {{ project.description }}</span>
                </p>
              </div>

              <Badge colorScheme="gray" compact>archived</Badge>
            </RouterLink>
          </div>
        </section>
      </template>

      <!-- Empty state -->
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
    </div>
  </MainLayout>

  <CreateProjectModal
    :open="showCreateProject"
    @close="showCreateProject = false"
    @created="onProjectCreated"
  />
</template>
