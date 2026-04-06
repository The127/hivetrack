<!--
  ProjectBacklogView — fused backlog + sprint planning view.

  Shows the active sprint, planning sprints, and the unsprinted backlog in a
  single page (Jira-style). Controls to create sprints, move issues between
  sprints, activate, and complete sprints all live here.

  Drag-and-drop: Issues can be reordered within sections and dragged between
  sprints/backlog. When dragging from the backlog, a fixed sprint drop panel
  appears on the right edge so sprints don't need to be in view.
-->
<script setup>
import { priorityBorder, estimateLabel, isTerminalStatus } from "@/composables/issueConstants";
import { ref, reactive, computed, watch } from "vue";
import { useRoute } from "vue-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/vue-query";
import { VueDraggable } from "vue-draggable-plus";
import { useDragReorder } from "@/composables/useDragReorder";
import { useInlineCreate } from "@/composables/useInlineCreate";
import { formatDateRange } from "@/composables/useDate";
import {
  PlusIcon,
  ListIcon,
  LayersIcon,
  PlayIcon,
  CheckIcon,
  Trash2Icon,
  ArrowRightIcon,
  PencilIcon,
} from "lucide-vue-next";
import MainLayout from "@/layouts/MainLayout.vue";
import Badge from "@/components/ui/Badge.vue";
import ProjectHeader from "@/components/project/ProjectHeader.vue";
import Spinner from "@/components/ui/Spinner.vue";
import AssigneePopover from "@/components/issue/AssigneePopover.vue";
import EmptyState from "@/components/ui/EmptyState.vue";
import CreateIssueModal from "@/components/issue/CreateIssueModal.vue";
import CompleteSprintModal from "@/components/sprint/CompleteSprintModal.vue";
import StatusSelect from "@/components/issue/StatusSelect.vue";
import PrioritySelect from "@/components/issue/PrioritySelect.vue";
import ProgressBar from "@/components/ui/ProgressBar.vue";
import BacklogEpicFilter from "@/components/backlog/BacklogEpicFilter.vue";
import BacklogSprintForm from "@/components/backlog/BacklogSprintForm.vue";
import { fetchProject } from "@/api/projects";
import { fetchIssues, updateIssue } from "@/api/issues";
import {
  fetchSprints,
  createSprint,
  updateSprint,
  deleteSprint,
} from "@/api/sprints";

const route = useRoute();
const slug = computed(() => route.params.slug);
const queryClient = useQueryClient();

// ── Epic filter ──────────────────────────────────────────────────────────────

const selectedEpicId = ref(null);

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

const issueParams = computed(() => {
  const params = { triaged: true, type: "task", limit: 1000 };
  if (selectedEpicId.value) {
    params.parent_id = selectedEpicId.value;
  }
  return params;
});

const { data: issuesResult, isLoading: loadingIssues } = useQuery({
  queryKey: computed(() => ["issues", slug.value, issueParams.value]),
  queryFn: () => fetchIssues(slug.value, issueParams.value),
  enabled: computed(() => !!slug.value),
});

const { data: epicsResult } = useQuery({
  queryKey: ["issues", slug, { type: "epic" }],
  queryFn: () => fetchIssues(slug.value, { type: "epic", limit: 200 }),
  enabled: computed(() => !!slug.value),
});

const epics = computed(() => epicsResult.value?.items ?? []);
const selectedEpic = computed(
  () => epics.value.find((e) => e.id === selectedEpicId.value) ?? null,
);

const isLoading = computed(
  () => loadingProject.value || loadingSprints.value || loadingIssues.value,
);

// ── Grouping ──────────────────────────────────────────────────────────────────

const allIssues = computed(() => {
  const arch = project.value?.archetype ?? 'software';
  return (issuesResult.value?.items ?? []).filter(
    (i) => !isTerminalStatus(i.status, arch),
  );
});
const allSprints = computed(() => sprintsResult.value?.sprints ?? []);

const activeSprint = computed(
  () => allSprints.value.find((s) => s.status === "active") ?? null,
);
const planningSprints = computed(() =>
  allSprints.value.filter((s) => s.status === "planning"),
);

const targetSprints = computed(() => {
  const sprints = [];
  if (activeSprint.value) sprints.push(activeSprint.value);
  sprints.push(...planningSprints.value);
  return sprints;
});

// ── Drag-and-drop state ─────────────────────────────────────────────────────

const BACKLOG_KEY = "__backlog__";
const sectionIssues = ref({});
const overlayDropZones = reactive({});

function rebuildSectionIssues() {
  const sections = { [BACKLOG_KEY]: [] };
  for (const sprint of allSprints.value) {
    sections[sprint.id] = [];
  }
  for (const issue of allIssues.value) {
    const key = issue.sprint_id ?? BACKLOG_KEY;
    const dest = sections[key] ? key : BACKLOG_KEY;
    sections[dest].push(issue);
  }
  sectionIssues.value = sections;
}

// ── Drag-and-drop ──────────────────────────────────────────────────────────

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

const { isDragging, onDragStart: onSectionDragStart, onDragEnd: onSectionDragEnd, handleDrag } = useDragReorder(
  sectionIssues,
  (item, data) => reorderIssue({ issueNumber: item.number, data }),
);

function onWithinSectionDrag(evt, sectionId) {
  handleDrag(evt, sectionId);
}

function onCrossSectionDrop(evt, toSectionId) {
  const newSprintId = toSectionId === BACKLOG_KEY ? null : toSectionId;
  handleDrag(evt, toSectionId, { sprint_id: newSprintId });
}

watch(
  [allIssues, allSprints],
  () => {
    if (!isDragging.value) rebuildSectionIssues();
  },
  { immediate: true },
);

watch(
  targetSprints,
  (sprints) => {
    for (const sprint of sprints) {
      if (!overlayDropZones[sprint.id]) {
        overlayDropZones[sprint.id] = [];
      }
    }
    if (!overlayDropZones[BACKLOG_KEY]) {
      overlayDropZones[BACKLOG_KEY] = [];
    }
  },
  { immediate: true },
);

function onDropToOverlayZone(evt, targetId) {
  const arr = overlayDropZones[targetId];
  if (!arr?.length) return;
  const droppedItem = arr[0];
  overlayDropZones[targetId] = [];

  const sectionKey = targetId === BACKLOG_KEY ? BACKLOG_KEY : targetId;
  if (!sectionIssues.value[sectionKey]) {
    sectionIssues.value[sectionKey] = [];
  }
  sectionIssues.value[sectionKey].push(droppedItem);

  const newSprintId = targetId === BACKLOG_KEY ? null : targetId;
  droppedItem.sprint_id = newSprintId;

  isDragging.value = false;
  reorderIssue({
    issueNumber: droppedItem.number,
    data: { sprint_id: newSprintId },
  });
}

// ── Move issue mutation (for button-based moves) ────────────────────────────

const { mutate: moveIssue } = useMutation({
  mutationFn: ({ issueNumber, sprintId }) =>
    updateIssue(slug.value, issueNumber, { sprint_id: sprintId }),
  onMutate: async ({ issueNumber, sprintId }) => {
    const key = ["issues", slug.value, { triaged: true }];
    await queryClient.cancelQueries({ queryKey: key });
    const previous = queryClient.getQueryData(key);
    queryClient.setQueryData(key, (old) => {
      if (!old) return old;
      return {
        ...old,
        items: old.items.map((i) =>
          i.number === issueNumber ? { ...i, sprint_id: sprintId } : i,
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
    queryClient.invalidateQueries({ queryKey: ["issues", slug.value] });
  },
});


// ── Sprint status mutations ───────────────────────────────────────────────────

const sprintErrors = ref({});

const { mutate: activateSprint } = useMutation({
  mutationFn: (sprintId) =>
    updateSprint(slug.value, sprintId, { status: "active" }),
  onSuccess: (_data, sprintId) => {
    delete sprintErrors.value[sprintId];
    queryClient.invalidateQueries({ queryKey: ["sprints", slug.value] });
  },
  onError: (err, sprintId) => {
    sprintErrors.value[sprintId] =
      err?.status === 409 || String(err?.message).includes("409")
        ? "Another sprint is already active."
        : "Failed to activate sprint.";
    queryClient.invalidateQueries({ queryKey: ["sprints", slug.value] });
  },
});

const showCompleteSprintModal = ref(false);

// Since allIssues already excludes terminal statuses, sectionIssues only contains
// non-terminal issues. openIssuesInActiveSprint == sectionIssues for active sprint.
const openIssuesInActiveSprint = computed(() => {
  if (!activeSprint.value) return [];
  return sectionIssues.value[activeSprint.value.id] ?? [];
});

const doneInActiveSprint = computed(() => {
  if (!activeSprint.value) return 0;
  const total = (sectionIssues.value[activeSprint.value.id] ?? []).length;
  return total - openIssuesInActiveSprint.value.length;
});

const completionTargetSprints = computed(() => planningSprints.value);

function requestCompleteSprint() {
  if (openIssuesInActiveSprint.value.length > 0) {
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
  onError: () => {
    queryClient.invalidateQueries({ queryKey: ["sprints", slug.value] });
  },
});

function doCompleteSprint({ moveToSprintId }) {
  completeSprintMutation({ sprintId: activeSprint.value.id, moveToSprintId });
}

const { mutate: doDeleteSprint } = useMutation({
  mutationFn: (sprintId) => deleteSprint(slug.value, sprintId),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ["sprints", slug.value] });
    queryClient.invalidateQueries({ queryKey: ["issues", slug.value] });
  },
  onError: () => {
    queryClient.invalidateQueries({ queryKey: ["sprints", slug.value] });
  },
});

// ── New sprint form ───────────────────────────────────────────────────────────

const showNewSprintForm = ref(false);
const newSprintError = ref("");

const { mutate: submitNewSprint, isPending: creatingSprintPending } =
  useMutation({
    mutationFn: (data) => createSprint(slug.value, data),
    onSuccess: () => {
      showNewSprintForm.value = false;
      newSprintError.value = "";
      queryClient.invalidateQueries({ queryKey: ["sprints", slug.value] });
    },
    onError: () => {
      newSprintError.value = "Failed to create sprint.";
    },
  });

function handleCreateSprint(formData) {
  if (!formData.name.trim()) {
    newSprintError.value = "Name is required.";
    return;
  }
  const data = { name: formData.name.trim() };
  if (formData.start_date)
    data.start_date = formData.start_date + "T00:00:00Z";
  if (formData.end_date)
    data.end_date = formData.end_date + "T00:00:00Z";
  if (formData.goal.trim()) data.goal = formData.goal.trim();
  submitNewSprint(data);
}

// ── Edit sprint ───────────────────────────────────────────────────────────────

const editingSprintId = ref(null);
const editSprintError = ref("");

function isoToDateInput(iso) {
  if (!iso) return "";
  return iso.slice(0, 10);
}

function editSprintInitialData(sprint) {
  return {
    name: sprint.name,
    start_date: isoToDateInput(sprint.start_date),
    end_date: isoToDateInput(sprint.end_date),
    goal: sprint.goal ?? "",
  };
}

function startEditSprint(sprint) {
  editingSprintId.value = sprint.id;
  editSprintError.value = "";
}

function cancelEditSprint() {
  editingSprintId.value = null;
  editSprintError.value = "";
}

const { mutate: submitEditSprint, isPending: editingSprintPending } =
  useMutation({
    mutationFn: ({ id, data }) => updateSprint(slug.value, id, data),
    onSuccess: () => {
      editingSprintId.value = null;
      editSprintError.value = "";
      queryClient.invalidateQueries({ queryKey: ["sprints", slug.value] });
    },
    onError: () => {
      editSprintError.value = "Failed to update sprint.";
    },
  });

function handleEditSprint(formData) {
  if (!formData.name.trim()) {
    editSprintError.value = "Name is required.";
    return;
  }
  const data = { name: formData.name.trim() };
  if (formData.start_date)
    data.start_date = formData.start_date + "T00:00:00Z";
  if (formData.end_date)
    data.end_date = formData.end_date + "T00:00:00Z";
  if (formData.goal.trim())
    data.goal = formData.goal.trim();
  submitEditSprint({ id: editingSprintId.value, data });
}

// ── Move-to-sprint dropdown ───────────────────────────────────────────────────

const openDropdown = ref(null);

function toggleDropdown(issueId) {
  openDropdown.value = openDropdown.value === issueId ? null : issueId;
}

function moveToSprint(issue, sprintId) {
  openDropdown.value = null;
  moveIssue({ issueNumber: issue.number, sprintId });
}

function moveToBacklog(issue) {
  moveIssue({ issueNumber: issue.number, sprintId: null });
}

// ── Inline status update ────────────────────────────────────────────────────

function updateStatus(issue, newStatus) {
  reorderIssue({ issueNumber: issue.number, data: { status: newStatus } });
}

function updatePriority(issue, newPriority) {
  reorderIssue({ issueNumber: issue.number, data: { priority: newPriority } });
}

// ── Inline issue creation (per-section) ─────────────────────────────────────

const projectArchetype = computed(() => project.value?.archetype ?? "software");

const {
  activeInlineCreate,
  inlineCreateTitle,
  inlineCreateError,
  setInlineCreateRef,
  activateInlineCreate,
  submitInlineCreate,
  cancelInlineCreate,
} = useInlineCreate(slug, projectArchetype, queryClient);

// ── Default status for issue creation (modal — header button) ───────────────

const defaultCreateStatus = computed(() => {
  if (!project.value) return null;
  return projectArchetype.value === "support" ? "open" : "todo";
});

const showCreateIssue = ref(false);

// ── Helpers ───────────────────────────────────────────────────────────────────

</script>

<template>
  <MainLayout @create-issue="showCreateIssue = true">
    <div class="flex flex-col h-full">
      <!-- ── Header ─────────────────────────────────────────────────────── -->
      <div
        class="flex-shrink-0 flex items-center justify-between px-6 py-3 border-b border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900"
      >
        <div class="flex items-center gap-3 min-w-0">
          <ProjectHeader v-if="project" :project="project" />
          <div
            v-else-if="loadingProject"
            class="h-5 w-40 rounded bg-slate-100 dark:bg-slate-800 animate-pulse"
          />

          <div class="flex items-center gap-1.5 text-slate-400 dark:text-slate-500">
            <ListIcon class="size-4" />
            <span class="text-sm font-medium text-slate-600 dark:text-slate-300">Backlog</span>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <!-- Epic filter -->
          <BacklogEpicFilter
            v-if="epics.length"
            :epics="epics"
            :model-value="selectedEpicId"
            @update:model-value="selectedEpicId = $event"
          />

          <button
            class="inline-flex items-center gap-1.5 rounded-md bg-blue-600 px-3 h-8 text-sm font-medium text-white hover:bg-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-1 transition-colors cursor-pointer"
            @click="showCreateIssue = true"
          >
            <PlusIcon class="size-4" />
            New issue
          </button>
        </div>
      </div>

      <!-- ── Epic filter banner ────────────────────────────────────────── -->
      <div
        v-if="selectedEpic"
        class="flex-shrink-0 flex items-center gap-3 px-6 py-2 border-b border-violet-200 dark:border-violet-800/50 bg-violet-50 dark:bg-violet-900/20"
      >
        <LayersIcon class="size-4 text-violet-500 flex-shrink-0" />
        <span
          class="text-xs font-medium text-violet-600 dark:text-violet-400 uppercase tracking-wide"
          >Filtered by epic</span
        >
        <router-link
          :to="`/projects/${slug}/issues/${selectedEpic.number}`"
          class="font-medium text-sm text-violet-700 dark:text-violet-300 hover:underline"
        >
          {{ selectedEpic.title }}
        </router-link>
        <button
          class="ml-auto text-xs text-violet-500 dark:text-violet-400 hover:text-violet-700 dark:hover:text-violet-200 cursor-pointer"
          @click="selectedEpicId = null"
        >
          Clear filter
        </button>
      </div>

      <!-- ── Loading ────────────────────────────────────────────────────── -->
      <div v-if="isLoading" class="h-32 flex items-center justify-center">
        <Spinner class="size-6 text-slate-400" />
      </div>

      <!-- ── Content ────────────────────────────────────────────────────── -->
      <div v-else class="flex-1 overflow-y-auto">
        <!-- ── Active Sprint ─────────────────────────────────────────────── -->
        <template v-if="activeSprint">
          <div
            class="px-6 py-2.5 border-b border-slate-100 dark:border-slate-800 bg-blue-50 dark:bg-blue-900/20 flex items-center justify-between"
          >
            <div class="flex items-center gap-2 min-w-0">
              <span
                class="text-xs font-medium text-blue-700 dark:text-blue-300 bg-blue-100 dark:bg-blue-800/50 px-1.5 py-0.5 rounded uppercase tracking-wide flex-shrink-0"
                >Active</span
              >
              <span class="font-semibold text-slate-900 dark:text-slate-100 text-sm truncate">{{
                activeSprint.name
              }}</span>
              <span
                v-if="activeSprint.start_date"
                class="text-xs text-slate-500 dark:text-slate-400 flex-shrink-0"
              >
                {{
                  formatDateRange(
                    activeSprint.start_date,
                    activeSprint.end_date,
                  )
                }}
              </span>
              <span
                v-if="activeSprint.goal"
                class="text-xs text-slate-500 dark:text-slate-400 italic truncate max-w-48"
                >{{ activeSprint.goal }}</span
              >
              <div class="w-28 flex-shrink-0">
                <ProgressBar
                  :done="doneInActiveSprint"
                  :total="(sectionIssues[activeSprint.id] ?? []).length"
                />
              </div>
            </div>
            <div class="flex items-center gap-2 flex-shrink-0">
              <button
                class="inline-flex items-center gap-1.5 rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 px-2.5 h-7 text-xs font-medium text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors cursor-pointer"
                @click="requestCompleteSprint()"
              >
                <CheckIcon class="size-3.5" />
                Complete sprint
              </button>
              <button
                class="inline-flex items-center rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 px-2 h-7 text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300 hover:border-slate-400 dark:hover:border-slate-500 focus-visible:outline-none transition-colors cursor-pointer"
                title="Edit sprint"
                @click="startEditSprint(activeSprint)"
              >
                <PencilIcon class="size-3.5" />
              </button>
              <button
                class="inline-flex items-center rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 px-2 h-7 text-slate-400 dark:text-slate-500 hover:text-red-600 hover:border-red-300 focus-visible:outline-none transition-colors cursor-pointer"
                title="Delete sprint (moves issues to backlog)"
                @click="doDeleteSprint(activeSprint.id)"
              >
                <Trash2Icon class="size-3.5" />
              </button>
            </div>
          </div>

          <!-- Inline edit form for active sprint -->
          <div
            v-if="editingSprintId === activeSprint.id"
            class="px-6 py-4 border-b border-slate-100 dark:border-slate-800 bg-blue-50/40 dark:bg-blue-900/10"
          >
            <BacklogSprintForm
              mode="edit"
              :initial-data="editSprintInitialData(activeSprint)"
              :loading="editingSprintPending"
              :error="editSprintError"
              @submit="handleEditSprint"
              @cancel="cancelEditSprint"
            />
          </div>

          <VueDraggable
            v-model="sectionIssues[activeSprint.id]"
            :group="{ name: 'backlog' }"
            :animation="150"
            ghost-class="opacity-30"
            :class="isDragging ? 'min-h-10' : ''"
            @start="onSectionDragStart"
            @end="onSectionDragEnd"
            @update="(evt) => onWithinSectionDrag(evt, activeSprint.id)"
            @add="(evt) => onCrossSectionDrop(evt, activeSprint.id)"
          >
            <div
              v-for="issue in sectionIssues[activeSprint.id]"
              :key="issue.id"
              class="group flex items-center gap-3 px-6 py-2.5 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors cursor-grab active:cursor-grabbing border-l-4 border-b border-slate-100 dark:border-slate-800"
              :class="priorityBorder(issue.priority)"
            >
              <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
                <router-link
                  :to="`/projects/${slug}/issues/${issue.number}`"
                  class="text-[11px] font-mono text-slate-400 dark:text-slate-500 hover:text-blue-600 dark:hover:text-blue-400 hover:underline"
                  >{{ slug.toUpperCase() }}-{{ issue.number }}</router-link
                >
                <LayersIcon
                  v-if="issue.type === 'epic'"
                  class="size-3 text-violet-400 flex-shrink-0"
                />
              </div>
              <router-link
                :to="`/projects/${slug}/issues/${issue.number}`"
                class="flex-1 min-w-0 text-sm text-slate-800 dark:text-slate-200 truncate group-hover:text-slate-900 dark:group-hover:text-slate-100 hover:underline"
                >{{ issue.title }}</router-link
              >
              <span
                v-if="issue.on_hold"
                class="flex-shrink-0 text-[10px] font-medium bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-400 px-1.5 py-0.5 rounded"
                >on hold</span
              >
              <StatusSelect
                :status="issue.status"
                :archetype="project.archetype"
                @update:status="updateStatus(issue, $event)"
              />
              <PrioritySelect
                :priority="issue.priority ?? 'none'"
                @update:priority="updatePriority(issue, $event)"
              />
              <span
                v-if="estimateLabel(issue.estimate)"
                class="flex-shrink-0 text-[11px] font-medium text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-slate-700 px-1.5 py-0.5 rounded w-7 text-center"
                >{{ estimateLabel(issue.estimate) }}</span
              >
              <span v-else class="w-7 flex-shrink-0" />
              <div class="flex-shrink-0 flex gap-1 max-w-32">
                <Badge
                  v-for="l in (issue.labels ?? []).slice(0, 2)"
                  :key="l.id"
                  dot
                  :dot-color="l.color"
                  compact
                  >{{ l.name }}</Badge
                >
              </div>
              <div class="flex-shrink-0 flex justify-end w-10">
                <AssigneePopover :assignees="issue.assignees ?? []" />
              </div>
              <button
                class="flex-shrink-0 opacity-0 group-hover:opacity-100 text-[11px] text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300 transition-opacity cursor-pointer px-1.5 py-0.5 rounded hover:bg-slate-100 dark:hover:bg-slate-700"
                @click.stop="moveToBacklog(issue)"
              >
                ↓ Backlog
              </button>
            </div>
          </VueDraggable>
          <!-- Inline create row for active sprint -->
          <div
            v-if="activeInlineCreate === activeSprint.id"
            class="flex items-center gap-3 px-6 py-2.5 border-b border-slate-100 dark:border-slate-800 border-l-4 border-l-blue-400 bg-blue-50/30 dark:bg-blue-900/10"
          >
            <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
              <PlusIcon class="size-3 text-blue-400" />
            </div>
            <input
              :ref="setInlineCreateRef(activeSprint.id)"
              v-model="inlineCreateTitle"
              type="text"
              placeholder="Issue title — Enter to create, Esc to close"
              class="flex-1 min-w-0 text-sm text-slate-800 dark:text-slate-200 bg-transparent placeholder:text-slate-400 dark:placeholder:text-slate-500 focus:outline-none"
              @keydown.enter.prevent="submitInlineCreate(activeSprint.id)"
              @keydown.escape="cancelInlineCreate"
              @blur="cancelInlineCreate"
            />
            <span
              v-if="inlineCreateError"
              class="flex-shrink-0 text-xs text-red-500"
              >{{ inlineCreateError }}</span
            >
          </div>
          <div
            v-else
            class="flex items-center gap-3 px-6 py-2 border-b border-slate-100 dark:border-slate-800 border-l-4 border-l-transparent cursor-text group/create"
            @click="activateInlineCreate(activeSprint.id)"
          >
            <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
              <PlusIcon
                class="size-3 text-slate-300 dark:text-slate-600 group-hover/create:text-slate-400 dark:group-hover/create:text-slate-500"
              />
            </div>
            <span
              class="text-sm text-slate-300 dark:text-slate-600 group-hover/create:text-slate-400 dark:group-hover/create:text-slate-500"
              >Create issue</span
            >
          </div>
        </template>

        <!-- ── Planning Sprints ──────────────────────────────────────────── -->
        <template v-for="sprint in planningSprints" :key="sprint.id">
          <div
            class="px-6 py-2.5 border-b border-slate-100 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50 flex items-center justify-between"
          >
            <div class="flex items-center gap-2 min-w-0">
              <span
                class="text-xs font-medium text-slate-500 dark:text-slate-400 bg-slate-200 dark:bg-slate-700 px-1.5 py-0.5 rounded uppercase tracking-wide flex-shrink-0"
                >Planning</span
              >
              <span class="font-semibold text-slate-900 dark:text-slate-100 text-sm truncate">{{
                sprint.name
              }}</span>
              <span
                v-if="sprint.start_date"
                class="text-xs text-slate-500 dark:text-slate-400 flex-shrink-0"
              >
                {{ formatDateRange(sprint.start_date, sprint.end_date) }}
              </span>
              <span class="text-xs text-slate-400 dark:text-slate-500 tabular-nums flex-shrink-0">
                {{ (sectionIssues[sprint.id] ?? []).length }} issues
              </span>
            </div>
            <div class="flex items-center gap-2 flex-shrink-0">
              <span
                v-if="sprintErrors[sprint.id]"
                class="text-xs text-red-600"
                >{{ sprintErrors[sprint.id] }}</span
              >
              <button
                class="inline-flex items-center gap-1.5 rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 px-2.5 h-7 text-xs font-medium text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors cursor-pointer"
                @click="activateSprint(sprint.id)"
              >
                <PlayIcon class="size-3.5" />
                Activate
              </button>
              <button
                class="inline-flex items-center rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 px-2 h-7 text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300 hover:border-slate-400 dark:hover:border-slate-500 focus-visible:outline-none transition-colors cursor-pointer"
                title="Edit sprint"
                @click="startEditSprint(sprint)"
              >
                <PencilIcon class="size-3.5" />
              </button>
              <button
                class="inline-flex items-center rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 px-2 h-7 text-slate-400 dark:text-slate-500 hover:text-red-600 hover:border-red-300 focus-visible:outline-none transition-colors cursor-pointer"
                title="Delete sprint (moves issues to backlog)"
                @click="doDeleteSprint(sprint.id)"
              >
                <Trash2Icon class="size-3.5" />
              </button>
            </div>
          </div>

          <!-- Inline edit form for planning sprint -->
          <div
            v-if="editingSprintId === sprint.id"
            class="px-6 py-4 border-b border-slate-100 dark:border-slate-800 bg-slate-50 dark:bg-slate-800/50"
          >
            <BacklogSprintForm
              mode="edit"
              :initial-data="editSprintInitialData(sprint)"
              :loading="editingSprintPending"
              :error="editSprintError"
              @submit="handleEditSprint"
              @cancel="cancelEditSprint"
            />
          </div>

          <VueDraggable
            v-model="sectionIssues[sprint.id]"
            :group="{ name: 'backlog' }"
            :animation="150"
            ghost-class="opacity-30"
            :class="isDragging ? 'min-h-10' : ''"
            @start="onSectionDragStart"
            @end="onSectionDragEnd"
            @update="(evt) => onWithinSectionDrag(evt, sprint.id)"
            @add="(evt) => onCrossSectionDrop(evt, sprint.id)"
          >
            <div
              v-for="issue in sectionIssues[sprint.id]"
              :key="issue.id"
              class="group flex items-center gap-3 px-6 py-2.5 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors cursor-grab active:cursor-grabbing border-l-4 border-b border-slate-100 dark:border-slate-800"
              :class="priorityBorder(issue.priority)"
            >
              <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
                <router-link
                  :to="`/projects/${slug}/issues/${issue.number}`"
                  class="text-[11px] font-mono text-slate-400 dark:text-slate-500 hover:text-blue-600 dark:hover:text-blue-400 hover:underline"
                  >{{ slug.toUpperCase() }}-{{ issue.number }}</router-link
                >
                <LayersIcon
                  v-if="issue.type === 'epic'"
                  class="size-3 text-violet-400 flex-shrink-0"
                />
              </div>
              <router-link
                :to="`/projects/${slug}/issues/${issue.number}`"
                class="flex-1 min-w-0 text-sm text-slate-800 dark:text-slate-200 truncate group-hover:text-slate-900 dark:group-hover:text-slate-100 hover:underline"
                >{{ issue.title }}</router-link
              >
              <span
                v-if="issue.on_hold"
                class="flex-shrink-0 text-[10px] font-medium bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-400 px-1.5 py-0.5 rounded"
                >on hold</span
              >
              <StatusSelect
                :status="issue.status"
                :archetype="project.archetype"
                @update:status="updateStatus(issue, $event)"
              />
              <PrioritySelect
                :priority="issue.priority ?? 'none'"
                @update:priority="updatePriority(issue, $event)"
              />
              <span
                v-if="estimateLabel(issue.estimate)"
                class="flex-shrink-0 text-[11px] font-medium text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-slate-700 px-1.5 py-0.5 rounded w-7 text-center"
                >{{ estimateLabel(issue.estimate) }}</span
              >
              <span v-else class="w-7 flex-shrink-0" />
              <div class="flex-shrink-0 flex gap-1 max-w-32">
                <Badge
                  v-for="l in (issue.labels ?? []).slice(0, 2)"
                  :key="l.id"
                  dot
                  :dot-color="l.color"
                  compact
                  >{{ l.name }}</Badge
                >
              </div>
              <div class="flex-shrink-0 flex justify-end w-10">
                <AssigneePopover :assignees="issue.assignees ?? []" />
              </div>
              <button
                class="flex-shrink-0 opacity-0 group-hover:opacity-100 text-[11px] text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300 transition-opacity cursor-pointer px-1.5 py-0.5 rounded hover:bg-slate-100 dark:hover:bg-slate-700"
                @click.stop="moveToBacklog(issue)"
              >
                ↓ Backlog
              </button>
            </div>
          </VueDraggable>
          <!-- Inline create row for planning sprint -->
          <div
            v-if="activeInlineCreate === sprint.id"
            class="flex items-center gap-3 px-6 py-2.5 border-b border-slate-100 dark:border-slate-800 border-l-4 border-l-blue-400 bg-blue-50/30 dark:bg-blue-900/10"
          >
            <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
              <PlusIcon class="size-3 text-blue-400" />
            </div>
            <input
              :ref="setInlineCreateRef(sprint.id)"
              v-model="inlineCreateTitle"
              type="text"
              placeholder="Issue title — Enter to create, Esc to close"
              class="flex-1 min-w-0 text-sm text-slate-800 dark:text-slate-200 bg-transparent placeholder:text-slate-400 dark:placeholder:text-slate-500 focus:outline-none"
              @keydown.enter.prevent="submitInlineCreate(sprint.id)"
              @keydown.escape="cancelInlineCreate"
              @blur="cancelInlineCreate"
            />
            <span
              v-if="inlineCreateError"
              class="flex-shrink-0 text-xs text-red-500"
              >{{ inlineCreateError }}</span
            >
          </div>
          <div
            v-else
            class="flex items-center gap-3 px-6 py-2 border-b border-slate-100 dark:border-slate-800 border-l-4 border-l-transparent cursor-text group/create"
            @click="activateInlineCreate(sprint.id)"
          >
            <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
              <PlusIcon
                class="size-3 text-slate-300 dark:text-slate-600 group-hover/create:text-slate-400 dark:group-hover/create:text-slate-500"
              />
            </div>
            <span
              class="text-sm text-slate-300 dark:text-slate-600 group-hover/create:text-slate-400 dark:group-hover/create:text-slate-500"
              >Create issue</span
            >
          </div>
        </template>

        <!-- ── Backlog section header ─────────────────────────────────────── -->
        <div
          class="px-6 py-2.5 border-b border-slate-100 dark:border-slate-800 bg-white dark:bg-slate-900 flex items-center justify-between"
        >
          <div class="flex items-center gap-2">
            <span class="font-semibold text-slate-900 dark:text-slate-100 text-sm">Backlog</span>
            <span class="text-xs text-slate-400 dark:text-slate-500 tabular-nums"
              >{{ (sectionIssues[BACKLOG_KEY] ?? []).length }} issues</span
            >
          </div>
          <button
            class="inline-flex items-center gap-1.5 rounded-md border border-slate-300 bg-white px-2.5 h-7 text-xs font-medium text-slate-600 hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors cursor-pointer"
            @click="showNewSprintForm = !showNewSprintForm"
          >
            <PlusIcon class="size-3.5" />
            New sprint
          </button>
        </div>

        <!-- Inline new sprint form -->
        <div
          v-if="showNewSprintForm"
          class="px-6 py-4 border-b border-slate-100 bg-slate-50"
        >
          <BacklogSprintForm
            mode="create"
            :loading="creatingSprintPending"
            :error="newSprintError"
            @submit="handleCreateSprint"
            @cancel="showNewSprintForm = false; newSprintError = ''"
          />
        </div>

        <!-- Backlog issues -->
        <VueDraggable
          v-model="sectionIssues[BACKLOG_KEY]"
          :group="{ name: 'backlog' }"
          :animation="150"
          ghost-class="opacity-30"
          :class="isDragging ? 'min-h-16' : ''"
          @start="onSectionDragStart"
          @end="onSectionDragEnd"
          @update="(evt) => onWithinSectionDrag(evt, BACKLOG_KEY)"
          @add="(evt) => onCrossSectionDrop(evt, BACKLOG_KEY)"
        >
          <div
            v-for="issue in sectionIssues[BACKLOG_KEY]"
            :key="issue.id"
            class="group relative flex items-center gap-3 px-6 py-2.5 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors cursor-grab active:cursor-grabbing border-l-4 border-b border-slate-100 dark:border-slate-800"
            :class="priorityBorder(issue.priority)"
          >
            <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
              <router-link
                :to="`/projects/${slug}/issues/${issue.number}`"
                class="text-[11px] font-mono text-slate-400 hover:text-blue-600 hover:underline"
                >{{ slug.toUpperCase() }}-{{ issue.number }}</router-link
              >
              <LayersIcon
                v-if="issue.type === 'epic'"
                class="size-3 text-violet-400 flex-shrink-0"
              />
            </div>
            <router-link
              :to="`/projects/${slug}/issues/${issue.number}`"
              class="flex-1 min-w-0 text-sm text-slate-800 truncate group-hover:text-slate-900 hover:underline"
              >{{ issue.title }}</router-link
            >
            <span
              v-if="issue.on_hold"
              class="flex-shrink-0 text-[10px] font-medium bg-amber-100 text-amber-700 px-1.5 py-0.5 rounded"
              >on hold</span
            >
            <StatusSelect
              :status="issue.status"
              :archetype="project.archetype"
              @update:status="updateStatus(issue, $event)"
            />
            <PrioritySelect
              :priority="issue.priority ?? 'none'"
              @update:priority="updatePriority(issue, $event)"
            />
            <span
              v-if="estimateLabel(issue.estimate)"
              class="flex-shrink-0 text-[11px] font-medium text-slate-500 bg-slate-100 px-1.5 py-0.5 rounded w-7 text-center"
              >{{ estimateLabel(issue.estimate) }}</span
            >
            <span v-else class="w-7 flex-shrink-0" />
            <div class="flex-shrink-0 flex gap-1 max-w-32">
              <Badge
                v-for="l in (issue.labels ?? []).slice(0, 2)"
                :key="l.id"
                dot
                :dot-color="l.color"
                compact
                >{{ l.name }}</Badge
              >
            </div>
            <div class="flex-shrink-0 flex justify-end w-10">
              <AssigneePopover :assignees="issue.assignees ?? []" />
            </div>

            <!-- Move to sprint dropdown -->
            <div v-if="targetSprints.length" class="relative flex-shrink-0">
              <button
                class="opacity-0 group-hover:opacity-100 text-[11px] text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300 transition-opacity cursor-pointer px-1.5 py-0.5 rounded hover:bg-slate-100 dark:hover:bg-slate-700 inline-flex items-center gap-0.5"
                @click.stop="toggleDropdown(issue.id)"
              >
                → Sprint <ChevronDownIcon class="size-3" />
              </button>
              <div
                v-if="openDropdown === issue.id"
                class="absolute right-0 top-full mt-1 z-10 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-md shadow-md py-1 min-w-36"
              >
                <button
                  v-for="sprint in targetSprints"
                  :key="sprint.id"
                  class="w-full text-left px-3 py-1.5 text-xs text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 cursor-pointer truncate"
                  @click="moveToSprint(issue, sprint.id)"
                >
                  {{ sprint.name }}
                </button>
              </div>
            </div>
          </div>
        </VueDraggable>
        <!-- Inline create row for backlog -->
        <div
          v-if="activeInlineCreate === BACKLOG_KEY"
          class="flex items-center gap-3 px-6 py-2.5 border-b border-slate-100 border-l-4 border-l-blue-400 bg-blue-50/30"
        >
          <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
            <PlusIcon class="size-3 text-blue-400" />
          </div>
          <input
            :ref="setInlineCreateRef(BACKLOG_KEY)"
            v-model="inlineCreateTitle"
            type="text"
            placeholder="Issue title — Enter to create, Esc to close"
            class="flex-1 min-w-0 text-sm text-slate-800 bg-transparent placeholder:text-slate-400 focus:outline-none"
            @keydown.enter.prevent="submitInlineCreate(BACKLOG_KEY)"
            @keydown.escape="cancelInlineCreate"
            @blur="cancelInlineCreate"
          />
          <span
            v-if="inlineCreateError"
            class="flex-shrink-0 text-xs text-red-500"
            >{{ inlineCreateError }}</span
          >
        </div>
        <div
          v-else
          class="flex items-center gap-3 px-6 py-2 border-b border-slate-100 border-l-4 border-l-transparent cursor-text group/create"
          @click="activateInlineCreate(BACKLOG_KEY)"
        >
          <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
            <PlusIcon
              class="size-3 text-slate-300 dark:text-slate-600 group-hover/create:text-slate-400 dark:group-hover/create:text-slate-500"
            />
          </div>
          <span class="text-sm text-slate-300 group-hover/create:text-slate-400"
            >Create issue</span
          >
        </div>

        <!-- Empty state when nothing anywhere -->
        <div
          v-if="
            !(sectionIssues[BACKLOG_KEY] ?? []).length &&
            !activeSprint &&
            !planningSprints.length
          "
          class="flex items-center justify-center py-16"
        >
          <EmptyState
            title="Backlog is empty"
            description="Triaged issues not assigned to a sprint will appear here."
            action-label="New issue"
            @action="activateInlineCreate(BACKLOG_KEY)"
          >
            <template #icon>
              <ListIcon class="size-8" />
            </template>
          </EmptyState>
        </div>
      </div>
    </div>

    <!-- ── Quick-drop overlay (shown when dragging) ──────────────────── -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-opacity duration-150"
        enter-from-class="opacity-0"
        leave-active-class="transition-opacity duration-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="isDragging"
          class="fixed right-4 top-1/2 -translate-y-1/2 z-50 flex flex-col gap-2 w-52"
        >
          <div
            class="text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide px-1 mb-0.5 flex items-center gap-1"
          >
            <ArrowRightIcon class="size-3" />
            Move to
          </div>
          <div v-for="sprint in targetSprints" :key="sprint.id">
            <VueDraggable
              v-model="overlayDropZones[sprint.id]"
              :group="{ name: 'backlog', put: true, pull: false }"
              :animation="0"
              class="rounded-lg border-2 border-dashed px-3 py-3 text-center min-h-12 transition-colors bg-white/90 dark:bg-slate-800/90 backdrop-blur-sm shadow-lg"
              :class="
                overlayDropZones[sprint.id]?.length
                  ? 'border-blue-400 bg-blue-50/90 dark:bg-blue-900/40'
                  : 'border-slate-300 dark:border-slate-600 hover:border-blue-300'
              "
              @add="(evt) => onDropToOverlayZone(evt, sprint.id)"
            >
              <div
                v-for="item in overlayDropZones[sprint.id]"
                :key="item.id"
                class="hidden"
              />
            </VueDraggable>
            <div class="text-xs text-slate-600 dark:text-slate-400 font-medium truncate mt-1 px-1">
              {{ sprint.name }}
              <span v-if="sprint.status === 'active'" class="text-blue-600"
                >(active)</span
              >
            </div>
          </div>
          <!-- Backlog drop zone -->
          <div class="mt-1 pt-2 border-t border-slate-200 dark:border-slate-700">
            <VueDraggable
              v-model="overlayDropZones[BACKLOG_KEY]"
              :group="{ name: 'backlog', put: true, pull: false }"
              :animation="0"
              class="rounded-lg border-2 border-dashed px-3 py-3 text-center min-h-12 transition-colors bg-white/90 dark:bg-slate-800/90 backdrop-blur-sm shadow-lg"
              :class="
                overlayDropZones[BACKLOG_KEY]?.length
                  ? 'border-blue-400 bg-blue-50/90 dark:bg-blue-900/40'
                  : 'border-slate-300 dark:border-slate-600 hover:border-blue-300'
              "
              @add="(evt) => onDropToOverlayZone(evt, BACKLOG_KEY)"
            >
              <div
                v-for="item in overlayDropZones[BACKLOG_KEY]"
                :key="item.id"
                class="hidden"
              />
            </VueDraggable>
            <div class="text-xs text-slate-600 dark:text-slate-400 font-medium mt-1 px-1">
              Backlog
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

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
      :open-issue-count="openIssuesInActiveSprint.length"
      :sprints="completionTargetSprints"
      @close="showCompleteSprintModal = false"
      @confirm="doCompleteSprint"
    />
  </MainLayout>
</template>
