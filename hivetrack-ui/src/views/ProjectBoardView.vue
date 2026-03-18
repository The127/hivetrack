<!--
  ProjectBoardView — Kanban board for a project.

  Shows the active sprint's issues grouped by status. If no sprint is active,
  falls back to the backlog (triaged issues with no sprint).

  A context bar below the header shows either:
  - Active sprint: name, dates, goal, issue count + "Complete sprint" action
  - No sprint: a note that the backlog is being shown + link to create a sprint

  Drag-and-drop: Issues can be reordered within columns and dragged across
  columns to change status. Uses fractional indexing for O(1) rank updates.
-->
<script setup>
import { ref, computed, watch, nextTick } from "vue";
import { useRoute, RouterLink } from "vue-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/vue-query";
import { VueDraggable } from "vue-draggable-plus";
import { generateKeyBetween } from "fractional-indexing";
import {
  PlusIcon,
  CircleIcon,
  CircleDotIcon,
  GitPullRequestIcon,
  CheckCircle2Icon,
  XCircleIcon,
  ChevronDownIcon,
  InboxIcon,
  LayersIcon,
  CheckIcon,
  SearchIcon,
  XIcon,
} from "lucide-vue-next";
import MainLayout from "@/layouts/MainLayout.vue";
import Badge from "@/components/ui/Badge.vue";
import Spinner from "@/components/ui/Spinner.vue";
import AssigneePopover from "@/components/issue/AssigneePopover.vue";
import Alert from "@/components/ui/Alert.vue";
import ProgressBar from "@/components/ui/ProgressBar.vue";
import CreateIssueModal from "@/components/issue/CreateIssueModal.vue";
import CompleteSprintModal from "@/components/sprint/CompleteSprintModal.vue";
import PrioritySelect from "@/components/issue/PrioritySelect.vue";
import { fetchProject } from "@/api/projects";
import { fetchIssues, updateIssue } from "@/api/issues";
import { fetchSprints, updateSprint } from "@/api/sprints";

const route = useRoute();
const slug = computed(() => route.params.slug);
const queryClient = useQueryClient();

// ── Data ─────────────────────────────────────────────────────────────────────

const { data: project, isLoading: loadingProject } = useQuery({
  queryKey: ["project", slug],
  queryFn: () => fetchProject(slug.value),
});

const { data: sprintsResult, isLoading: loadingSprints } = useQuery({
  queryKey: ["sprints", slug],
  queryFn: () => fetchSprints(slug.value),
  enabled: computed(() => !!slug.value),
});

const { data: issuesResult, isLoading: loadingIssues } = useQuery({
  queryKey: ["issues", slug, { triaged: true, type: "task" }],
  queryFn: () =>
    fetchIssues(slug.value, { triaged: true, type: "task", limit: 500 }),
  enabled: computed(() => !!slug.value),
});

const { data: epicsResult } = useQuery({
  queryKey: ["issues", slug, { type: "epic" }],
  queryFn: () => fetchIssues(slug.value, { type: "epic", limit: 500 }),
  enabled: computed(() => !!slug.value),
});

const epicMap = computed(() => {
  const map = {};
  for (const epic of epicsResult.value?.items ?? []) {
    map[epic.id] = epic;
  }
  return map;
});

const isLoading = computed(
  () => loadingProject.value || loadingSprints.value || loadingIssues.value,
);

// ── Active sprint + issue source ──────────────────────────────────────────────

const activeSprint = computed(
  () =>
    (sprintsResult.value?.sprints ?? []).find((s) => s.status === "active") ??
    null,
);

const boardIssues = computed(() => {
  const all = issuesResult.value?.items ?? [];
  if (activeSprint.value) {
    return all.filter((i) => i.sprint_id === activeSprint.value.id);
  }
  return all.filter((i) => i.sprint_id == null);
});

// ── Complete sprint ──────────────────────────────────────────────────────────

const showCompleteSprintModal = ref(false);

const TERMINAL_STATUSES_SOFTWARE = new Set(["done", "cancelled"]);
const TERMINAL_STATUSES_SUPPORT = new Set(["resolved", "closed"]);

const openIssuesInSprint = computed(() => {
  const terminal =
    project.value?.archetype === "support"
      ? TERMINAL_STATUSES_SUPPORT
      : TERMINAL_STATUSES_SOFTWARE;
  return boardIssues.value.filter((i) => !terminal.has(i.status));
});

const doneInSprint = computed(
  () => boardIssues.value.length - openIssuesInSprint.value.length,
);

const otherSprints = computed(() =>
  (sprintsResult.value?.sprints ?? []).filter((s) => s.status === "planning"),
);

function requestCompleteSprint() {
  if (openIssuesInSprint.value.length > 0) {
    showCompleteSprintModal.value = true;
  } else {
    doCompleteSprint({ moveToSprintId: null });
  }
}

const { mutate: completeSprintMutation } = useMutation({
  mutationFn: ({ sprintId, moveToSprintId }) =>
    updateSprint(slug.value, sprintId, {
      status: "completed",
      move_open_issues_to_sprint_id: moveToSprintId ?? null,
    }),
  onSuccess: () => {
    showCompleteSprintModal.value = false;
    queryClient.invalidateQueries({ queryKey: ["sprints", slug.value] });
    queryClient.invalidateQueries({ queryKey: ["issues", slug.value] });
  },
});

function doCompleteSprint({ moveToSprintId }) {
  completeSprintMutation({ sprintId: activeSprint.value.id, moveToSprintId });
}

// ── Status column config ──────────────────────────────────────────────────────

const SOFTWARE_COLUMNS = [
  { key: "todo", label: "To Do", scheme: "gray", icon: CircleIcon },
  {
    key: "in_progress",
    label: "In Progress",
    scheme: "blue",
    icon: CircleDotIcon,
  },
  {
    key: "in_review",
    label: "In Review",
    scheme: "violet",
    icon: GitPullRequestIcon,
  },
  { key: "done", label: "Done", scheme: "green", icon: CheckCircle2Icon },
];

const SUPPORT_COLUMNS = [
  { key: "open", label: "Open", scheme: "sky", icon: CircleIcon },
  {
    key: "in_progress",
    label: "In Progress",
    scheme: "blue",
    icon: CircleDotIcon,
  },
  {
    key: "resolved",
    label: "Resolved",
    scheme: "teal",
    icon: CheckCircle2Icon,
  },
  { key: "closed", label: "Closed", scheme: "gray", icon: XCircleIcon },
];

const wipLimits = computed(() => ({
  in_progress: project.value?.wip_limit_in_progress ?? null,
  in_review: project.value?.wip_limit_in_review ?? null,
}));

const columns = computed(() => {
  if (!project.value) return [];
  return project.value.archetype === "support"
    ? SUPPORT_COLUMNS
    : SOFTWARE_COLUMNS;
});

// ── Board search / filter ────────────────────────────────────────────────────

const searchQuery = ref("");
const showSearch = ref(false);
const searchInputEl = ref(null);

function toggleSearch() {
  showSearch.value = !showSearch.value;
  if (showSearch.value) {
    nextTick(() => searchInputEl.value?.focus());
  } else {
    searchQuery.value = "";
  }
}

function matchesSearch(issue) {
  if (!searchQuery.value) return true;
  const q = searchQuery.value.toLowerCase();
  const id = `${slug.value}-${issue.number}`.toLowerCase();
  return issue.title.toLowerCase().includes(q) || id.includes(q);
}

// ── Drag-and-drop state ─────────────────────────────────────────────────────

const isDragging = ref(false);
const columnIssues = ref({});

function rebuildColumnIssues() {
  const newMap = {};
  for (const col of columns.value) {
    newMap[col.key] = (boardIssues.value ?? [])
      .filter((i) => i.status === col.key && matchesSearch(i))
      .slice();
  }
  columnIssues.value = newMap;
}

watch(
  [boardIssues, columns, searchQuery],
  () => {
    if (!isDragging.value) rebuildColumnIssues();
  },
  { immediate: true },
);

function computeRank(items, newIdx) {
  const prev = newIdx > 0 ? (items[newIdx - 1]?.rank ?? null) : null;
  const next =
    newIdx < items.length - 1 ? (items[newIdx + 1]?.rank ?? null) : null;
  try {
    return generateKeyBetween(prev, next);
  } catch {
    return Date.now().toString(36) + Math.random().toString(36).slice(2, 6);
  }
}

// ── Reorder mutation ────────────────────────────────────────────────────────

const { mutate: reorderIssue } = useMutation({
  mutationFn: ({ issueNumber, data }) =>
    updateIssue(slug.value, issueNumber, data),
  onMutate: async ({ issueNumber, data }) => {
    const queryKey = ["issues", slug.value, { triaged: true }];
    await queryClient.cancelQueries({ queryKey });
    const previous = queryClient.getQueryData(queryKey);
    queryClient.setQueryData(queryKey, (old) => {
      if (!old) return old;
      return {
        ...old,
        items: old.items.map((i) =>
          i.number === issueNumber ? { ...i, ...data } : i,
        ),
      };
    });
    return { previous };
  },
  onError: (_err, _vars, context) => {
    if (context?.previous) {
      queryClient.setQueryData(
        ["issues", slug.value, { triaged: true }],
        context.previous,
      );
    }
  },
  onSettled: () => {
    isDragging.value = false;
    queryClient.invalidateQueries({ queryKey: ["issues", slug.value] });
  },
});

// ── Drag handlers ───────────────────────────────────────────────────────────

function onDragStart() {
  isDragging.value = true;
}

function onDragEnd() {
  setTimeout(() => {
    isDragging.value = false;
  }, 0);
}

function onWithinColumnDrag(evt, colKey) {
  const items = columnIssues.value[colKey];
  const newIdx = evt.newDraggableIndex;
  const movedItem = items[newIdx];
  const newRank = computeRank(items, newIdx);
  movedItem.rank = newRank;
  reorderIssue({ issueNumber: movedItem.number, data: { rank: newRank } });
}

function onCrossColumnDrop(evt, toColKey) {
  const items = columnIssues.value[toColKey];
  const newIdx = evt.newDraggableIndex;
  const movedItem = items[newIdx];
  const newRank = computeRank(items, newIdx);
  movedItem.rank = newRank;
  movedItem.status = toColKey;
  reorderIssue({
    issueNumber: movedItem.number,
    data: { rank: newRank, status: toColKey },
  });
}

// ── Priority / estimate helpers ───────────────────────────────────────────────

const PRIORITY_BORDER = {
  none: "border-l-slate-200",
  low: "border-l-sky-400",
  medium: "border-l-amber-400",
  high: "border-l-orange-500",
  critical: "border-l-red-500",
};

const ESTIMATE_LABEL = {
  none: null,
  xs: "XS",
  s: "S",
  m: "M",
  l: "L",
  xl: "XL",
};

function priorityBorder(priority) {
  return PRIORITY_BORDER[priority] ?? "border-l-slate-200";
}

function estimateLabel(estimate) {
  return ESTIMATE_LABEL[estimate] ?? null;
}

function formatDateRange(startDate, endDate) {
  const fmt = (d) =>
    new Date(d).toLocaleDateString("en-US", { month: "short", day: "numeric" });
  return `${fmt(startDate)} – ${fmt(endDate)}`;
}

// ── Cancelled issues (collapsible in Done column) ────────────────────────────

const showCancelled = ref(false)

const cancelledBoardIssues = computed(() =>
  boardIssues.value.filter(i => i.status === 'cancelled' && matchesSearch(i))
)

// ── New issue modal ───────────────────────────────────────────────────────────

const showCreateIssue = ref(false);

const defaultCreateStatus = computed(() => {
  if (!project.value) return null;
  return project.value.archetype === "support" ? "open" : "todo";
});
</script>

<template>
  <MainLayout @create-issue="showCreateIssue = true">
    <div class="flex flex-col h-full">
      <!-- ── Board header ───────────────────────────────────────────────── -->
      <div
        class="flex-shrink-0 flex items-center justify-between px-6 py-3 border-b border-slate-200 bg-white"
      >
        <div class="flex items-center gap-3 min-w-0">
          <div v-if="project" class="flex items-center gap-2 min-w-0">
            <span
              class="size-7 rounded flex items-center justify-center text-xs font-semibold bg-slate-100 text-slate-600 flex-shrink-0"
            >
              {{ project.slug.slice(0, 2).toUpperCase() }}
            </span>
            <span class="font-semibold text-slate-900 truncate">{{
              project.name
            }}</span>
            <Badge
              :color-scheme="project.archetype === 'software' ? 'blue' : 'teal'"
              compact
            >
              {{ project.archetype }}
            </Badge>
          </div>
          <div
            v-else-if="loadingProject"
            class="h-5 w-40 rounded bg-slate-100 animate-pulse"
          />
        </div>

        <div class="flex items-center gap-2">
          <!-- Search -->
          <div v-if="showSearch" class="relative">
            <SearchIcon
              class="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-slate-400"
            />
            <input
              ref="searchInputEl"
              v-model="searchQuery"
              type="text"
              placeholder="Filter issues..."
              class="w-56 rounded-md border border-slate-200 pl-8 pr-8 h-8 text-sm text-slate-800 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 transition-colors"
              @keydown.escape="toggleSearch"
            />
            <button
              class="absolute right-2 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 cursor-pointer"
              @click="toggleSearch"
            >
              <XIcon class="size-3.5" />
            </button>
          </div>
          <button
            v-else
            class="inline-flex items-center justify-center rounded-md border border-slate-200 size-8 text-slate-500 hover:bg-slate-50 hover:text-slate-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors cursor-pointer"
            @click="toggleSearch"
          >
            <SearchIcon class="size-4" />
          </button>

          <button
            class="inline-flex items-center gap-1.5 rounded-md bg-blue-600 px-3 h-8 text-sm font-medium text-white hover:bg-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-1 transition-colors cursor-pointer"
            @click="showCreateIssue = true"
          >
            <PlusIcon class="size-4" />
            New issue
          </button>
        </div>
      </div>

      <!-- ── Context bar ────────────────────────────────────────────────── -->
      <template v-if="!isLoading">
        <!-- Active sprint -->
        <div
          v-if="activeSprint"
          class="flex-shrink-0 px-6 py-2 bg-blue-50 border-b border-blue-100"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2.5 min-w-0">
              <span
                class="text-xs font-medium text-blue-700 bg-blue-100 px-1.5 py-0.5 rounded uppercase tracking-wide flex-shrink-0"
                >Sprint</span
              >
              <span class="text-sm font-semibold text-slate-900 truncate">{{
                activeSprint.name
              }}</span>
              <span
                v-if="activeSprint.start_date"
                class="text-xs text-slate-500 flex-shrink-0"
              >
                {{
                  formatDateRange(
                    activeSprint.start_date,
                    activeSprint.end_date,
                  )
                }}
              </span>
              <div class="w-32 flex-shrink-0">
                <ProgressBar
                  :done="doneInSprint"
                  :total="boardIssues.length"
                />
              </div>
            </div>
            <button
              class="inline-flex items-center gap-1.5 rounded-md border border-slate-300 bg-white px-2.5 h-7 text-xs font-medium text-slate-600 hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors cursor-pointer flex-shrink-0"
              @click="requestCompleteSprint()"
            >
              <CheckIcon class="size-3.5" />
              Complete sprint
            </button>
          </div>
          <p v-if="activeSprint.goal" class="text-xs text-slate-600 mt-1.5">
            <span class="text-slate-400 font-medium">Goal:</span>
            {{ activeSprint.goal }}
          </p>
        </div>

        <!-- No active sprint — showing backlog -->
        <div v-else class="flex-shrink-0 px-6 py-2 border-b border-slate-100">
          <Alert>
            Showing backlog — no active sprint.
            <RouterLink
              :to="`/projects/${slug}/backlog`"
              class="ml-1 text-blue-600 hover:text-blue-700 hover:underline"
            >
              Go to Backlog to create and activate a sprint.
            </RouterLink>
          </Alert>
        </div>
      </template>

      <!-- ── Loading ────────────────────────────────────────────────────── -->
      <div v-if="isLoading" class="h-32 flex items-center justify-center">
        <Spinner class="size-6 text-slate-400" />
      </div>

      <!-- ── Board columns ──────────────────────────────────────────────── -->
      <div
        v-else-if="boardIssues.length > 0"
        class="flex-1 overflow-x-auto overflow-y-hidden"
      >
        <div class="flex h-full gap-3 px-6 py-4" style="min-width: max-content">
          <div
            v-for="col in columns"
            :key="col.key"
            class="flex flex-col w-72 flex-shrink-0"
          >
            <!-- Column header -->
            <div class="flex items-center gap-2 mb-3 px-1">
              <component
                :is="col.icon"
                class="size-3.5 flex-shrink-0"
                :class="{
                  'text-slate-400': col.scheme === 'gray',
                  'text-blue-500': col.scheme === 'blue',
                  'text-violet-500': col.scheme === 'violet',
                  'text-green-500': col.scheme === 'green',
                  'text-sky-500': col.scheme === 'sky',
                  'text-teal-500': col.scheme === 'teal',
                }"
              />
              <span class="text-sm font-medium text-slate-700">{{
                col.label
              }}</span>
              <span
                class="ml-auto text-xs tabular-nums"
                :class="
                  wipLimits[col.key] != null &&
                  (columnIssues[col.key]?.length ?? 0) > wipLimits[col.key]
                    ? 'text-amber-500 font-medium'
                    : 'text-slate-400'
                "
              >
                <template v-if="wipLimits[col.key] != null">
                  {{ columnIssues[col.key]?.length ?? 0 }} / {{ wipLimits[col.key] }}
                </template>
                <template v-else>
                  {{ columnIssues[col.key]?.length ?? 0 }}
                </template>
              </span>
            </div>

            <!-- Draggable issue cards -->
            <div class="flex-1 overflow-y-auto pb-4 pr-1 -mr-1 relative">
              <VueDraggable
                v-model="columnIssues[col.key]"
                :group="{ name: 'board' }"
                :animation="150"
                ghost-class="opacity-30"
                class="space-y-2 min-h-full"
                @start="onDragStart"
                @end="onDragEnd"
                @update="(evt) => onWithinColumnDrag(evt, col.key)"
                @add="(evt) => onCrossColumnDrop(evt, col.key)"
              >
                <div
                  v-for="issue in columnIssues[col.key]"
                  :key="issue.id"
                  class="group rounded-lg border border-slate-200 bg-white px-3 py-2.5 shadow-sm hover:shadow-md hover:border-slate-300 transition-all cursor-grab active:cursor-grabbing border-l-4"
                  :class="priorityBorder(issue.priority)"
                >
                  <!-- Issue number + type + assignees -->
                  <div class="flex items-center gap-1.5 mb-1.5">
                    <RouterLink
                      :to="`/projects/${slug}/issues/${issue.number}`"
                      class="text-[11px] font-mono text-slate-400 hover:text-blue-600 hover:underline"
                      @click.stop
                    >
                      {{ slug.toUpperCase() }}-{{ issue.number }}
                    </RouterLink>
                    <LayersIcon
                      v-if="issue.type === 'epic'"
                      class="size-3 text-violet-400"
                    />
                    <span
                      v-if="issue.on_hold"
                      class="text-[10px] font-medium bg-amber-100 text-amber-700 px-1.5 py-0.5 rounded"
                    >
                      on hold
                    </span>
                    <span class="flex-1" />
                    <AssigneePopover :assignees="issue.assignees ?? []" />
                  </div>

                  <!-- Title -->
                  <p
                    class="text-sm text-slate-800 leading-snug line-clamp-2 group-hover:text-slate-900 mb-2"
                  >
                    {{ issue.title }}
                  </p>

                  <!-- Epic + priority row -->
                  <div class="flex items-center gap-1.5 mb-2">
                    <RouterLink
                      v-if="issue.parent_id && epicMap[issue.parent_id]"
                      :to="`/projects/${slug}/issues/${epicMap[issue.parent_id].number}`"
                      class="inline-flex items-center gap-1 text-[11px] font-medium text-violet-600 bg-violet-50 hover:bg-violet-100 px-1.5 py-0.5 rounded min-w-0 max-w-[10rem] transition-colors"
                      @click.stop
                    >
                      <LayersIcon class="size-3 flex-shrink-0" />
                      <span class="truncate">{{
                        epicMap[issue.parent_id].title
                      }}</span>
                    </RouterLink>
                    <span class="flex-1" />
                    <PrioritySelect
                      :priority="issue.priority ?? 'none'"
                      @update:priority="
                        reorderIssue({
                          issueNumber: issue.number,
                          data: { priority: $event },
                        })
                      "
                    />
                  </div>

                  <!-- Footer: estimate + labels -->
                  <div class="flex items-center gap-1.5 flex-wrap">
                    <span
                      v-if="estimateLabel(issue.estimate)"
                      class="text-[11px] font-medium text-slate-500 bg-slate-100 px-1.5 py-0.5 rounded"
                    >
                      {{ estimateLabel(issue.estimate) }}
                    </span>
                    <Badge
                      v-for="l in (issue.labels ?? []).slice(0, 3)"
                      :key="l.id"
                      dot
                      :dot-color="l.color"
                      compact
                      >{{ l.name }}</Badge
                    >
                    <span
                      v-if="(issue.labels ?? []).length > 3"
                      class="text-[10px] text-slate-400"
                    >
                      +{{ issue.labels.length - 3 }}
                    </span>
                  </div>
                </div>
              </VueDraggable>

              <!-- Empty column placeholder (positioned over the draggable area) -->
              <div
                v-if="!columnIssues[col.key]?.length && !isDragging"
                class="absolute inset-0 rounded-lg border-2 border-dashed border-slate-200 flex items-center justify-center"
              >
                <p class="text-xs text-slate-400">No issues</p>
              </div>
            </div>

            <!-- Cancelled issues (Done column only) -->
            <div v-if="col.key === 'done' && cancelledBoardIssues.length > 0" class="mt-2">
              <button
                class="flex items-center gap-1.5 w-full px-1 py-1.5 text-xs text-slate-400 hover:text-slate-600 transition-colors cursor-pointer"
                @click="showCancelled = !showCancelled"
              >
                <ChevronDownIcon
                  class="size-3.5 transition-transform"
                  :class="showCancelled ? 'rotate-0' : '-rotate-90'"
                />
                <XCircleIcon class="size-3.5" />
                <span>Cancelled · {{ cancelledBoardIssues.length }}</span>
              </button>
              <div v-if="showCancelled" class="space-y-2 mt-1">
                <div
                  v-for="issue in cancelledBoardIssues"
                  :key="issue.id"
                  class="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2.5 opacity-60 border-l-4 border-l-slate-200"
                >
                  <div class="flex items-center gap-1.5 mb-1">
                    <RouterLink
                      :to="`/projects/${slug}/issues/${issue.number}`"
                      class="text-[11px] font-mono text-slate-400 hover:text-blue-600 hover:underline"
                    >
                      {{ slug.toUpperCase() }}-{{ issue.number }}
                    </RouterLink>
                  </div>
                  <p class="text-sm text-slate-500 leading-snug line-clamp-2 line-through">
                    {{ issue.title }}
                  </p>
                </div>
              </div>
            </div>

          </div>
        </div>
      </div>

      <!-- ── Empty board ────────────────────────────────────────────────── -->
      <div
        v-else-if="!isLoading && boardIssues.length === 0"
        class="flex-1 flex items-center justify-center"
      >
        <div class="text-center">
          <InboxIcon class="size-10 text-slate-300 mx-auto mb-3" />
          <p class="text-sm font-medium text-slate-600">
            {{ activeSprint ? "Sprint is empty" : "Backlog is empty" }}
          </p>
          <p class="text-sm text-slate-400 mt-1">
            {{
              activeSprint
                ? "Add issues to this sprint from the Backlog."
                : "Create an issue or triage items from the inbox."
            }}
          </p>
        </div>
      </div>
    </div>

    <!-- ── Create issue modal ──────────────────────────────────────────── -->
    <CreateIssueModal
      :open="showCreateIssue"
      :project-slug="slug"
      :default-status="defaultCreateStatus"
      @close="showCreateIssue = false"
      @created="showCreateIssue = false"
    />

    <!-- ── Complete sprint modal ─────────────────────────────────────── -->
    <CompleteSprintModal
      :open="showCompleteSprintModal"
      :open-issue-count="openIssuesInSprint.length"
      :sprints="otherSprints"
      @close="showCompleteSprintModal = false"
      @confirm="doCompleteSprint"
    />
  </MainLayout>
</template>
