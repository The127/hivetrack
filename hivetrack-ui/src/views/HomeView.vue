<!--
  HomeView — Personal dashboard ("My Work").

  The default view after login. Answers: what should I work on right now?

  Sections (each backed by its own TanStack Query):
    1. My open issues    — /api/v1/me/issues
    2. Projects          — /api/v1/projects

  Keyboard shortcuts (from MainLayout):
    C  → create new issue
-->
<script setup>
import { ref, computed } from "vue";
import { useQuery, useMutation, useQueryClient } from "@tanstack/vue-query";
import { statusScheme, STATUS_META } from "@/composables/issueConstants";
import { useRouter } from "vue-router";
import {
  PlusIcon,
  InboxIcon,
  FolderKanbanIcon,
  CircleDotIcon,
  UserIcon,
  SearchIcon,
  XIcon,
} from "lucide-vue-next";
import MainLayout from "@/layouts/MainLayout.vue";
import AssigneePopover from "@/components/issue/AssigneePopover.vue";
import Badge from "@/components/ui/Badge.vue";
import EmptyState from "@/components/ui/EmptyState.vue";
import Spinner from "@/components/ui/Spinner.vue";
import CreateProjectModal from "@/components/project/CreateProjectModal.vue";
import CreateIssueModal from "@/components/issue/CreateIssueModal.vue";
import PrioritySelect from "@/components/issue/PrioritySelect.vue";
import { apiFetch } from "@/composables/useApi";
import { updateIssue } from "@/api/issues";
import { useAuth } from "@/composables/useAuth";

const { user } = useAuth();
const router = useRouter();
const queryClient = useQueryClient();

const showCreateProject = ref(false);
const showCreateIssue = ref(false);

function onProjectCreated(result) {
  showCreateProject.value = false;
  router.push(`/projects/${result.slug}/board`);
}

function onIssueCreated() {
  showCreateIssue.value = false;
}

const userName =
  user.value?.profile?.name ?? user.value?.profile?.email ?? "You";

// ── Queries ───────────────────────────────────────────────────────────────────

const { data: myIssues, isLoading: loadingIssues } = useQuery({
  queryKey: ["me", "issues"],
  queryFn: () => apiFetch("/api/v1/me/issues"),
});

const { data: myCreatedIssues, isLoading: loadingCreated } = useQuery({
  queryKey: ["me", "created-issues"],
  queryFn: () => apiFetch("/api/v1/me/created-issues"),
});

const { data: projects, isLoading: loadingProjects } = useQuery({
  queryKey: ["projects"],
  queryFn: () => apiFetch("/api/v1/projects"),
});

const totalUntriaged = computed(() =>
  (projects.value?.items ?? []).reduce(
    (sum, p) => sum + (p.untriaged_count ?? 0),
    0,
  ),
);

// ── Status display helpers ────────────────────────────────────────────────────

function formatStatus(s) {
  return (
    STATUS_META[s]?.label ??
    s.replace(/_/g, " ").replace(/^\w/, (c) => c.toUpperCase())
  );
}

const PRIORITIES = ["none", "low", "medium", "high", "critical"];

// ── Filters (my open issues) ────────────────────────────────────────────────
const assignedSearch = ref("");
const assignedStatus = ref("");
const assignedPriority = ref("");

const filteredAssigned = computed(() => {
  let items = myIssues.value?.items ?? [];
  if (assignedSearch.value) {
    const q = assignedSearch.value.toLowerCase();
    items = items.filter(
      (i) =>
        i.title.toLowerCase().includes(q) ||
        `${i.project_slug}-${i.number}`.toLowerCase().includes(q),
    );
  }
  if (assignedStatus.value)
    items = items.filter((i) => i.status === assignedStatus.value);
  if (assignedPriority.value)
    items = items.filter(
      (i) => (i.priority ?? "none") === assignedPriority.value,
    );
  return items;
});

const assignedStatuses = computed(() => {
  const set = new Set((myIssues.value?.items ?? []).map((i) => i.status));
  return [...set];
});

// ── Filters (created by me) ─────────────────────────────────────────────────
const createdSearch = ref("");
const createdStatus = ref("");
const createdPriority = ref("");

const filteredCreated = computed(() => {
  let items = myCreatedIssues.value?.items ?? [];
  if (createdSearch.value) {
    const q = createdSearch.value.toLowerCase();
    items = items.filter(
      (i) =>
        i.title.toLowerCase().includes(q) ||
        `${i.project_slug}-${i.number}`.toLowerCase().includes(q),
    );
  }
  if (createdStatus.value)
    items = items.filter((i) => i.status === createdStatus.value);
  if (createdPriority.value)
    items = items.filter(
      (i) => (i.priority ?? "none") === createdPriority.value,
    );
  return items;
});

const createdStatuses = computed(() => {
  const set = new Set(
    (myCreatedIssues.value?.items ?? []).map((i) => i.status),
  );
  return [...set];
});

// ── Priority mutation (list view) ─────────────────────────────────────────────

const { mutate: updateMyIssuePriority } = useMutation({
  mutationFn: ({ projectSlug, number, priority }) =>
    updateIssue(projectSlug, number, { priority }),
  onMutate: async ({ number, priority }) => {
    const key = ["me", "issues"];
    await queryClient.cancelQueries({ queryKey: key });
    const previous = queryClient.getQueryData(key);
    queryClient.setQueryData(key, (old) => {
      if (!old) return old;
      return {
        ...old,
        items: old.items.map((i) =>
          i.number === number ? { ...i, priority } : i,
        ),
      };
    });
    return { previous };
  },
  onError: (_err, _vars, context) => {
    if (context?.previous) {
      queryClient.setQueryData(["me", "issues"], context.previous);
    }
  },
  onSettled: () => {
    queryClient.invalidateQueries({ queryKey: ["me", "issues"] });
  },
});
</script>

<template>
  <MainLayout @create-issue="showCreateIssue = true">
    <div class="px-6 py-8">
      <!-- Page header -->
      <div class="max-w-3xl mx-auto mb-8 flex items-start justify-between">
        <div>
          <h1 class="text-xl font-semibold text-slate-900 dark:text-slate-100">
            My Work
          </h1>
          <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            Welcome back, {{ userName }}
          </p>
        </div>
        <button
          class="inline-flex items-center gap-1.5 rounded-md bg-blue-600 px-3 h-8 text-sm font-medium text-white hover:bg-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-1 transition-colors cursor-pointer"
          @click="showCreateIssue = true"
        >
          <PlusIcon class="size-4" />
          New issue
        </button>
      </div>

      <!-- ── My open issues ────────────────────────────────────────────── -->
      <section class="mb-8">
        <div class="max-w-3xl mx-auto flex items-center gap-3 mb-3">
          <h2
            class="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center gap-2"
          >
            <CircleDotIcon class="size-4 text-blue-500" />
            My open issues
            <span
              v-if="myIssues?.items?.length"
              class="text-xs font-normal text-slate-500"
            >
              {{ myIssues.items.length }}
            </span>
          </h2>
          <div class="ml-auto flex items-center gap-2">
            <div
              class="flex items-center gap-1.5 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 px-2 h-7"
            >
              <SearchIcon class="size-3 text-slate-400" />
              <input
                v-model="assignedSearch"
                type="text"
                placeholder="Filter..."
                class="w-24 text-xs bg-transparent text-slate-700 dark:text-slate-300 placeholder:text-slate-400 focus:outline-none"
              />
              <button
                v-if="assignedSearch"
                class="text-slate-400 hover:text-slate-600 cursor-pointer"
                @click="assignedSearch = ''"
              >
                <XIcon class="size-3" />
              </button>
            </div>
            <select
              v-model="assignedStatus"
              class="h-7 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs text-slate-600 dark:text-slate-300 px-1.5 focus:outline-none focus:ring-2 focus:ring-blue-500 cursor-pointer"
            >
              <option value="">All statuses</option>
              <option v-for="s in assignedStatuses" :key="s" :value="s">
                {{ formatStatus(s) }}
              </option>
            </select>
            <select
              v-model="assignedPriority"
              class="h-7 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs text-slate-600 dark:text-slate-300 px-1.5 focus:outline-none focus:ring-2 focus:ring-blue-500 cursor-pointer"
            >
              <option value="">All priorities</option>
              <option v-for="p in PRIORITIES" :key="p" :value="p">
                {{
                  p === "none" ? "None" : p.charAt(0).toUpperCase() + p.slice(1)
                }}
              </option>
            </select>
          </div>
        </div>

        <div v-if="loadingIssues" class="h-32 flex items-center justify-center">
          <Spinner class="size-5 text-slate-400" />
        </div>

        <template v-else-if="myIssues?.items?.length">
          <div
            v-if="filteredAssigned.length"
            class="max-w-3xl mx-auto rounded-lg border border-slate-200 dark:border-slate-700 divide-y divide-slate-100 dark:divide-slate-800 overflow-hidden"
          >
            <router-link
              v-for="issue in filteredAssigned"
              :key="issue.id"
              :to="`/projects/${issue.project_slug}/issues/${issue.number}`"
              class="flex items-center gap-3 px-4 py-2.5 bg-white dark:bg-slate-900 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors cursor-pointer group"
            >
              <span
                class="text-xs font-mono text-slate-400 dark:text-slate-500 flex-shrink-0 w-14 text-right"
              >
                {{ issue.project_slug?.toUpperCase() }}-{{ issue.number }}
              </span>
              <span
                class="flex-1 min-w-0 text-sm text-slate-800 dark:text-slate-200 truncate group-hover:text-slate-900 dark:group-hover:text-slate-100"
              >
                {{ issue.title }}
              </span>
              <PrioritySelect
                :priority="issue.priority ?? 'none'"
                @update:priority="
                  updateMyIssuePriority({
                    projectSlug: issue.project_slug,
                    number: issue.number,
                    priority: $event,
                  })
                "
              />
              <Badge :color-scheme="statusScheme(issue.status)" compact>
                {{ formatStatus(issue.status) }}
              </Badge>
              <AssigneePopover :assignees="issue.assignees ?? []" />
            </router-link>
          </div>
          <p
            v-else
            class="max-w-3xl mx-auto text-sm text-slate-400 dark:text-slate-500 text-center py-6"
          >
            No issues match the current filters.
          </p>
        </template>

        <EmptyState
          v-else
          class="max-w-3xl mx-auto"
          title="No open issues assigned to you"
          description="Issues assigned to you across all projects will appear here."
        >
          <template #icon>
            <CircleDotIcon class="size-8" />
          </template>
        </EmptyState>
      </section>

      <!-- ── Created by me ───────────────────────────────────────────── -->
      <section class="mb-8">
        <div class="max-w-3xl mx-auto flex items-center gap-3 mb-3">
          <h2
            class="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center gap-2"
          >
            <UserIcon class="size-4 text-violet-500" />
            Created by me
            <span
              v-if="myCreatedIssues?.items?.length"
              class="text-xs font-normal text-slate-500"
            >
              {{ myCreatedIssues.items.length }}
            </span>
          </h2>
          <div class="ml-auto flex items-center gap-2">
            <div
              class="flex items-center gap-1.5 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 px-2 h-7"
            >
              <SearchIcon class="size-3 text-slate-400" />
              <input
                v-model="createdSearch"
                type="text"
                placeholder="Filter..."
                class="w-24 text-xs bg-transparent text-slate-700 dark:text-slate-300 placeholder:text-slate-400 focus:outline-none"
              />
              <button
                v-if="createdSearch"
                class="text-slate-400 hover:text-slate-600 cursor-pointer"
                @click="createdSearch = ''"
              >
                <XIcon class="size-3" />
              </button>
            </div>
            <select
              v-model="createdStatus"
              class="h-7 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs text-slate-600 dark:text-slate-300 px-1.5 focus:outline-none focus:ring-2 focus:ring-blue-500 cursor-pointer"
            >
              <option value="">All statuses</option>
              <option v-for="s in createdStatuses" :key="s" :value="s">
                {{ formatStatus(s) }}
              </option>
            </select>
            <select
              v-model="createdPriority"
              class="h-7 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs text-slate-600 dark:text-slate-300 px-1.5 focus:outline-none focus:ring-2 focus:ring-blue-500 cursor-pointer"
            >
              <option value="">All priorities</option>
              <option v-for="p in PRIORITIES" :key="p" :value="p">
                {{
                  p === "none" ? "None" : p.charAt(0).toUpperCase() + p.slice(1)
                }}
              </option>
            </select>
          </div>
        </div>

        <div
          v-if="loadingCreated"
          class="h-32 flex items-center justify-center"
        >
          <Spinner class="size-5 text-slate-400" />
        </div>

        <template v-else-if="myCreatedIssues?.items?.length">
          <div
            v-if="filteredCreated.length"
            class="max-w-3xl mx-auto rounded-lg border border-slate-200 dark:border-slate-700 divide-y divide-slate-100 dark:divide-slate-800 overflow-hidden"
          >
            <router-link
              v-for="issue in filteredCreated"
              :key="issue.id"
              :to="`/projects/${issue.project_slug}/issues/${issue.number}`"
              class="flex items-center gap-3 px-4 py-2.5 bg-white dark:bg-slate-900 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors cursor-pointer group"
            >
              <span
                class="text-xs font-mono text-slate-400 dark:text-slate-500 flex-shrink-0 w-14 text-right"
              >
                {{ issue.project_slug?.toUpperCase() }}-{{ issue.number }}
              </span>
              <span
                class="flex-1 min-w-0 text-sm text-slate-800 dark:text-slate-200 truncate group-hover:text-slate-900 dark:group-hover:text-slate-100"
              >
                {{ issue.title }}
              </span>
              <Badge :color-scheme="statusScheme(issue.status)" compact>
                {{ formatStatus(issue.status) }}
              </Badge>
              <AssigneePopover :assignees="issue.assignees ?? []" />
            </router-link>
          </div>
          <p
            v-else
            class="max-w-3xl mx-auto text-sm text-slate-400 dark:text-slate-500 text-center py-6"
          >
            No issues match the current filters.
          </p>
        </template>

        <EmptyState
          v-else
          class="max-w-3xl mx-auto"
          title="No open issues created by you"
          description="Issues you create across all projects will appear here."
        >
          <template #icon>
            <UserIcon class="size-8" />
          </template>
        </EmptyState>
      </section>

      <!-- ── Projects ──────────────────────────────────────────────────── -->
      <section class="max-w-3xl mx-auto">
        <div class="flex items-center justify-between mb-3">
          <h2
            class="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center gap-2"
          >
            <FolderKanbanIcon
              class="size-4 text-slate-500 dark:text-slate-400"
            />
            Projects
            <span
              v-if="projects?.items?.length"
              class="text-xs font-normal text-slate-500"
            >
              {{ projects.items.length }}
            </span>
          </h2>
          <button
            class="inline-flex items-center gap-1.5 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 px-2.5 h-7 text-xs font-medium text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700 hover:text-slate-800 dark:hover:text-slate-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-1 transition-colors cursor-pointer"
            @click="showCreateProject = true"
          >
            <PlusIcon class="size-3.5" />
            New project
          </button>
        </div>

        <div
          v-if="loadingProjects"
          class="h-32 flex items-center justify-center"
        >
          <Spinner class="size-5 text-slate-400" />
        </div>

        <div
          v-else-if="projects?.items?.length"
          class="rounded-lg border border-slate-200 dark:border-slate-700 divide-y divide-slate-100 dark:divide-slate-800 overflow-hidden"
        >
          <RouterLink
            v-for="project in projects.items"
            :key="project.id"
            :to="`/projects/${project.slug}/board`"
            class="flex items-center gap-3 px-4 py-3 bg-white dark:bg-slate-900 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors group"
          >
            <!-- Project initial -->
            <span
              class="size-7 rounded flex items-center justify-center text-xs font-semibold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 flex-shrink-0"
            >
              {{ project.slug.slice(0, 2).toUpperCase() }}
            </span>

            <div class="flex-1 min-w-0">
              <p
                class="text-sm font-medium text-slate-800 dark:text-slate-200 group-hover:text-slate-900 dark:group-hover:text-slate-100 truncate"
              >
                {{ project.name }}
              </p>
              <p class="text-xs text-slate-500 dark:text-slate-400">
                {{ project.slug }}
              </p>
            </div>

            <Badge v-if="project.untriaged_count" color-scheme="amber" compact>
              <InboxIcon class="size-3 mr-0.5" />
              {{ project.untriaged_count }}
            </Badge>

            <Badge
              :color-scheme="project.archetype === 'software' ? 'blue' : 'teal'"
              compact
            >
              {{ project.archetype }}
            </Badge>

            <span v-if="project.archived" class="text-xs text-slate-400 italic"
              >archived</span
            >
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
        v-if="totalUntriaged > 0"
        class="mt-6 max-w-3xl mx-auto rounded-lg border border-amber-200 dark:border-amber-700/50 bg-amber-50 dark:bg-amber-900/20 px-4 py-3 flex items-center gap-3"
      >
        <InboxIcon class="size-4 text-amber-600 flex-shrink-0" />
        <p class="text-sm text-amber-800 dark:text-amber-300">
          <strong>{{ totalUntriaged }}</strong>
          {{ totalUntriaged === 1 ? "issue" : "issues" }} waiting in triage
          across your projects.
        </p>
      </div>
    </div>
  </MainLayout>

  <CreateProjectModal
    :open="showCreateProject"
    @close="showCreateProject = false"
    @created="onProjectCreated"
  />

  <CreateIssueModal
    :open="showCreateIssue"
    @close="showCreateIssue = false"
    @created="onIssueCreated"
  />
</template>
