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
import {
  ref,
  reactive,
  computed,
  watch,
  nextTick,
  onMounted,
  onBeforeUnmount,
} from "vue";
import { useRoute } from "vue-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/vue-query";
import { VueDraggable } from "vue-draggable-plus";
import { generateKeyBetween } from "fractional-indexing";
import {
  PlusIcon,
  ListIcon,
  LayersIcon,
  ChevronDownIcon,
  PlayIcon,
  CheckIcon,
  Trash2Icon,
  ArrowRightIcon,
  SearchIcon,
} from "lucide-vue-next";
import MainLayout from "@/layouts/MainLayout.vue";
import Badge from "@/components/ui/Badge.vue";
import Spinner from "@/components/ui/Spinner.vue";
import Avatar from "@/components/ui/Avatar.vue";
import EmptyState from "@/components/ui/EmptyState.vue";
import CreateIssueModal from "@/components/issue/CreateIssueModal.vue";
import CompleteSprintModal from "@/components/sprint/CompleteSprintModal.vue";
import StatusSelect from "@/components/issue/StatusSelect.vue";
import PrioritySelect from "@/components/issue/PrioritySelect.vue";
import ProgressBar from "@/components/ui/ProgressBar.vue";
import { fetchProject } from "@/api/projects";
import { fetchIssues, createIssue, updateIssue } from "@/api/issues";
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
const epicFilterOpen = ref(false);
const epicFilterRoot = ref(null);
const epicFilterDropdownEl = ref(null);
const epicFilterTrigger = ref(null);
const epicFilterStyle = ref({});
const epicFilterSearch = ref("");
const epicFilterSearchInput = ref(null);

function positionEpicFilter() {
  if (!epicFilterTrigger.value) return;
  const rect = epicFilterTrigger.value.getBoundingClientRect();
  epicFilterStyle.value = {
    position: "fixed",
    top: `${rect.bottom + 4}px`,
    left: `${rect.right}px`,
    transform: "translateX(-100%)",
    zIndex: 9999,
  };
}

function toggleEpicFilter() {
  epicFilterOpen.value = !epicFilterOpen.value;
  if (epicFilterOpen.value) {
    epicFilterSearch.value = "";
    nextTick(() => {
      positionEpicFilter();
      epicFilterSearchInput.value?.focus();
    });
  }
}

function selectEpicFilter(epicId) {
  selectedEpicId.value = epicId;
  epicFilterOpen.value = false;
}

function onEpicFilterClickOutside(e) {
  if (!epicFilterOpen.value) return;
  if (epicFilterRoot.value?.contains(e.target)) return;
  if (epicFilterDropdownEl.value?.contains(e.target)) return;
  epicFilterOpen.value = false;
}

onMounted(() =>
  document.addEventListener("pointerdown", onEpicFilterClickOutside, true),
);
onBeforeUnmount(() =>
  document.removeEventListener("pointerdown", onEpicFilterClickOutside, true),
);

const filteredEpicOptions = computed(() => {
  if (!epicFilterSearch.value) return epics.value;
  const q = epicFilterSearch.value.toLowerCase();
  return epics.value.filter(
    (e) => e.title.toLowerCase().includes(q) || String(e.number).includes(q),
  );
});

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

const TERMINAL_STATUSES_SOFTWARE = new Set(["done", "cancelled"]);
const TERMINAL_STATUSES_SUPPORT = new Set(["resolved", "closed"]);

const allIssues = computed(() => {
  const terminal =
    project.value?.archetype === "support"
      ? TERMINAL_STATUSES_SUPPORT
      : TERMINAL_STATUSES_SOFTWARE;
  return (issuesResult.value?.items ?? []).filter(
    (i) => !terminal.has(i.status),
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
const isDragging = ref(false);
const sectionIssues = ref({});
const overlayDropZones = reactive({});

function rebuildSectionIssues() {
  const sections = { [BACKLOG_KEY]: [] };
  for (const sprint of allSprints.value) {
    sections[sprint.id] = [];
  }
  for (const issue of allIssues.value) {
    const key = issue.sprint_id ?? BACKLOG_KEY;
    if (sections[key]) {
      sections[key].push(issue);
    }
  }
  sectionIssues.value = sections;
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

function onSectionDragStart() {
  isDragging.value = true;
}

function onSectionDragEnd() {
  setTimeout(() => {
    isDragging.value = false;
  }, 0);
}

function onWithinSectionDrag(evt, sectionId) {
  const items = sectionIssues.value[sectionId];
  if (!items) return;
  const newIdx = evt.newDraggableIndex;
  const movedItem = items[newIdx];
  const newRank = computeRank(items, newIdx);
  movedItem.rank = newRank;
  reorderIssue({ issueNumber: movedItem.number, data: { rank: newRank } });
}

function onCrossSectionDrop(evt, toSectionId) {
  const items = sectionIssues.value[toSectionId];
  if (!items) return;
  const newIdx = evt.newDraggableIndex;
  const movedItem = items[newIdx];
  const newRank = computeRank(items, newIdx);
  const newSprintId = toSectionId === BACKLOG_KEY ? null : toSectionId;
  movedItem.rank = newRank;
  movedItem.sprint_id = newSprintId;
  reorderIssue({
    issueNumber: movedItem.number,
    data: { rank: newRank, sprint_id: newSprintId },
  });
}

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

// ── Priority / estimate display ───────────────────────────────────────────────

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
const newSprint = ref({ name: "", start_date: "", end_date: "", goal: "" });
const newSprintError = ref("");

const { mutate: submitNewSprint, isPending: creatingSprintPending } =
  useMutation({
    mutationFn: (data) => createSprint(slug.value, data),
    onSuccess: () => {
      showNewSprintForm.value = false;
      newSprint.value = { name: "", start_date: "", end_date: "", goal: "" };
      newSprintError.value = "";
      queryClient.invalidateQueries({ queryKey: ["sprints", slug.value] });
    },
    onError: () => {
      newSprintError.value = "Failed to create sprint.";
    },
  });

function handleCreateSprint() {
  if (!newSprint.value.name.trim()) {
    newSprintError.value = "Name is required.";
    return;
  }
  const data = { name: newSprint.value.name.trim() };
  if (newSprint.value.start_date)
    data.start_date = newSprint.value.start_date + "T00:00:00Z";
  if (newSprint.value.end_date)
    data.end_date = newSprint.value.end_date + "T00:00:00Z";
  if (newSprint.value.goal.trim()) data.goal = newSprint.value.goal.trim();
  submitNewSprint(data);
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

const activeInlineCreate = ref(null); // section ID currently being edited
const inlineCreateTitle = ref("");
const inlineCreateError = ref("");
const inlineCreateInputs = ref({});

function setInlineCreateRef(sectionId) {
  return (el) => {
    inlineCreateInputs.value[sectionId] = el;
  };
}

function activateInlineCreate(sectionId) {
  activeInlineCreate.value = sectionId;
  inlineCreateTitle.value = "";
  inlineCreateError.value = "";
  nextTick(() => {
    const el = inlineCreateInputs.value[sectionId];
    el?.focus();
    el?.scrollIntoView({ behavior: "smooth", block: "nearest" });
  });
}

const { mutate: inlineCreate, isPending: inlineCreatePending } = useMutation({
  mutationFn: (data) => createIssue(slug.value, data),
  onSuccess: (_result, _variables) => {
    inlineCreateTitle.value = "";
    inlineCreateError.value = "";
    queryClient.invalidateQueries({ queryKey: ["issues", slug.value] });
    nextTick(() => {
      if (activeInlineCreate.value) {
        const el = inlineCreateInputs.value[activeInlineCreate.value];
        el?.focus();
        el?.scrollIntoView({ behavior: "smooth", block: "nearest" });
      }
    });
  },
  onError: () => {
    inlineCreateError.value = "Failed";
  },
});

function submitInlineCreate(sectionId) {
  const title = inlineCreateTitle.value.trim();
  if (!title) return;
  if (inlineCreatePending.value) return;
  const status = project.value?.archetype === "support" ? "open" : "todo";
  const sprintId = sectionId === BACKLOG_KEY ? undefined : sectionId;
  inlineCreate({ title, type: "task", status, sprint_id: sprintId });
}

function cancelInlineCreate() {
  if (inlineCreatePending.value) return;
  activeInlineCreate.value = null;
  inlineCreateTitle.value = "";
  inlineCreateError.value = "";
}

// ── Default status for issue creation (modal — header button) ───────────────

const defaultCreateStatus = computed(() => {
  if (!project.value) return null;
  return project.value.archetype === "support" ? "open" : "todo";
});

const showCreateIssue = ref(false);

// ── Helpers ───────────────────────────────────────────────────────────────────

function formatDateRange(startDate, endDate) {
  const fmt = (d) =>
    new Date(d).toLocaleDateString("en-US", { month: "short", day: "numeric" });
  return `${fmt(startDate)} – ${fmt(endDate)}`;
}
</script>

<template>
  <MainLayout @create-issue="showCreateIssue = true">
    <div class="flex flex-col h-full">
      <!-- ── Header ─────────────────────────────────────────────────────── -->
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

          <div class="flex items-center gap-1.5 text-slate-400">
            <ListIcon class="size-4" />
            <span class="text-sm font-medium text-slate-600">Backlog</span>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <!-- Epic filter -->
          <div v-if="epics.length" ref="epicFilterRoot" class="relative">
            <button
              ref="epicFilterTrigger"
              class="flex items-center gap-1.5 cursor-pointer rounded-md border border-slate-200 px-2.5 h-8 hover:bg-slate-50 hover:border-slate-300 transition-colors"
              @click="toggleEpicFilter"
            >
              <LayersIcon class="size-3.5 text-violet-400 flex-shrink-0" />
              <span
                class="text-sm"
                :class="
                  selectedEpic ? 'text-slate-700 font-medium' : 'text-slate-500'
                "
              >
                {{ selectedEpic ? selectedEpic.title : "All issues" }}
              </span>
              <ChevronDownIcon class="size-3 text-slate-400 ml-0.5" />
            </button>

            <Teleport to="body">
              <Transition
                enter-active-class="transition-opacity duration-75"
                enter-from-class="opacity-0"
                leave-active-class="transition-opacity duration-75"
                leave-to-class="opacity-0"
              >
                <div
                  v-if="epicFilterOpen"
                  ref="epicFilterDropdownEl"
                  :style="epicFilterStyle"
                  class="bg-white border border-slate-200 rounded-lg shadow-lg overflow-hidden min-w-52"
                >
                  <!-- Search (when many epics) -->
                  <div
                    v-if="epics.length > 5"
                    class="p-2 border-b border-slate-100"
                  >
                    <div class="relative">
                      <SearchIcon
                        class="absolute left-2 top-1/2 -translate-y-1/2 size-3.5 text-slate-400"
                      />
                      <input
                        ref="epicFilterSearchInput"
                        v-model="epicFilterSearch"
                        type="text"
                        placeholder="Search epics..."
                        class="w-full pl-7 pr-2 py-1 text-sm text-slate-800 placeholder:text-slate-400 bg-slate-50 rounded border-none focus:outline-none"
                        @keydown.escape="epicFilterOpen = false"
                      />
                    </div>
                  </div>

                  <div class="max-h-52 overflow-y-auto py-1">
                    <!-- All issues option -->
                    <button
                      class="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-left cursor-pointer transition-colors"
                      :class="
                        !selectedEpicId
                          ? 'bg-slate-50 font-medium text-slate-900'
                          : 'text-slate-500 hover:bg-slate-50'
                      "
                      @click="selectEpicFilter(null)"
                    >
                      <CheckIcon
                        v-if="!selectedEpicId"
                        class="size-3.5 text-blue-500 flex-shrink-0"
                      />
                      <span v-else class="size-3.5 flex-shrink-0" />
                      <span>All issues</span>
                    </button>

                    <button
                      v-for="epic in filteredEpicOptions"
                      :key="epic.id"
                      class="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-left cursor-pointer transition-colors"
                      :class="
                        epic.id === selectedEpicId
                          ? 'bg-slate-50 font-medium text-slate-900'
                          : 'text-slate-700 hover:bg-slate-50'
                      "
                      @click="selectEpicFilter(epic.id)"
                    >
                      <CheckIcon
                        v-if="epic.id === selectedEpicId"
                        class="size-3.5 text-blue-500 flex-shrink-0"
                      />
                      <LayersIcon
                        v-else
                        class="size-3.5 text-violet-400 flex-shrink-0"
                      />
                      <span class="flex-1 min-w-0 truncate">{{
                        epic.title
                      }}</span>
                    </button>

                    <p
                      v-if="epicFilterSearch && !filteredEpicOptions.length"
                      class="px-3 py-2 text-xs text-slate-400"
                    >
                      No epics match "{{ epicFilterSearch }}"
                    </p>
                  </div>
                </div>
              </Transition>
            </Teleport>
          </div>

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
        class="flex-shrink-0 flex items-center gap-3 px-6 py-2 border-b border-violet-200 bg-violet-50"
      >
        <LayersIcon class="size-4 text-violet-500 flex-shrink-0" />
        <span
          class="text-xs font-medium text-violet-600 uppercase tracking-wide"
          >Filtered by epic</span
        >
        <router-link
          :to="`/projects/${slug}/issues/${selectedEpic.number}`"
          class="font-medium text-sm text-violet-700 hover:underline"
        >
          {{ selectedEpic.title }}
        </router-link>
        <button
          class="ml-auto text-xs text-violet-500 hover:text-violet-700 cursor-pointer"
          @click="selectedEpicId = null"
        >
          Clear filter
        </button>
      </div>

      <!-- ── Loading ────────────────────────────────────────────────────── -->
      <div v-if="isLoading" class="flex-1 flex items-center justify-center">
        <Spinner class="size-6 text-slate-400" />
      </div>

      <!-- ── Content ────────────────────────────────────────────────────── -->
      <div v-else class="flex-1 overflow-y-auto">
        <!-- ── Active Sprint ─────────────────────────────────────────────── -->
        <template v-if="activeSprint">
          <div
            class="px-6 py-2.5 border-b border-slate-100 bg-blue-50 flex items-center justify-between"
          >
            <div class="flex items-center gap-2 min-w-0">
              <span
                class="text-xs font-medium text-blue-700 bg-blue-100 px-1.5 py-0.5 rounded uppercase tracking-wide flex-shrink-0"
                >Active</span
              >
              <span class="font-semibold text-slate-900 text-sm truncate">{{
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
              <span
                v-if="activeSprint.goal"
                class="text-xs text-slate-500 italic truncate max-w-48"
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
                class="inline-flex items-center gap-1.5 rounded-md border border-slate-300 bg-white px-2.5 h-7 text-xs font-medium text-slate-600 hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors cursor-pointer"
                @click="requestCompleteSprint()"
              >
                <CheckIcon class="size-3.5" />
                Complete sprint
              </button>
              <button
                class="inline-flex items-center rounded-md border border-slate-300 bg-white px-2 h-7 text-slate-400 hover:text-red-600 hover:border-red-300 focus-visible:outline-none transition-colors cursor-pointer"
                title="Delete sprint (moves issues to backlog)"
                @click="doDeleteSprint(activeSprint.id)"
              >
                <Trash2Icon class="size-3.5" />
              </button>
            </div>
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
              class="group flex items-center gap-3 px-6 py-2.5 hover:bg-slate-50 transition-colors cursor-grab active:cursor-grabbing border-l-4 border-b border-slate-100"
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
              <div class="flex-shrink-0 flex -space-x-1 w-10 justify-end">
                <Avatar
                  v-for="a in (issue.assignees ?? []).slice(0, 2)"
                  :key="a.id"
                  :name="a.display_name"
                  :src="a.avatar_url"
                  size="xs"
                  class="ring-1 ring-white"
                />
              </div>
              <button
                class="flex-shrink-0 opacity-0 group-hover:opacity-100 text-[11px] text-slate-400 hover:text-slate-600 transition-opacity cursor-pointer px-1.5 py-0.5 rounded hover:bg-slate-100"
                @click.stop="moveToBacklog(issue)"
              >
                ↓ Backlog
              </button>
            </div>
          </VueDraggable>
          <!-- Inline create row for active sprint -->
          <div
            v-if="activeInlineCreate === activeSprint.id"
            class="flex items-center gap-3 px-6 py-2.5 border-b border-slate-100 border-l-4 border-l-blue-400 bg-blue-50/30"
          >
            <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
              <PlusIcon class="size-3 text-blue-400" />
            </div>
            <input
              :ref="setInlineCreateRef(activeSprint.id)"
              v-model="inlineCreateTitle"
              type="text"
              placeholder="Issue title — Enter to create, Esc to close"
              class="flex-1 min-w-0 text-sm text-slate-800 bg-transparent placeholder:text-slate-400 focus:outline-none"
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
            class="flex items-center gap-3 px-6 py-2 border-b border-slate-100 border-l-4 border-l-transparent cursor-text group/create"
            @click="activateInlineCreate(activeSprint.id)"
          >
            <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
              <PlusIcon
                class="size-3 text-slate-300 group-hover/create:text-slate-400"
              />
            </div>
            <span
              class="text-sm text-slate-300 group-hover/create:text-slate-400"
              >Create issue</span
            >
          </div>
        </template>

        <!-- ── Planning Sprints ──────────────────────────────────────────── -->
        <template v-for="sprint in planningSprints" :key="sprint.id">
          <div
            class="px-6 py-2.5 border-b border-slate-100 bg-slate-50 flex items-center justify-between"
          >
            <div class="flex items-center gap-2 min-w-0">
              <span
                class="text-xs font-medium text-slate-500 bg-slate-200 px-1.5 py-0.5 rounded uppercase tracking-wide flex-shrink-0"
                >Planning</span
              >
              <span class="font-semibold text-slate-900 text-sm truncate">{{
                sprint.name
              }}</span>
              <span
                v-if="sprint.start_date"
                class="text-xs text-slate-500 flex-shrink-0"
              >
                {{ formatDateRange(sprint.start_date, sprint.end_date) }}
              </span>
              <span class="text-xs text-slate-400 tabular-nums flex-shrink-0">
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
                class="inline-flex items-center gap-1.5 rounded-md border border-slate-300 bg-white px-2.5 h-7 text-xs font-medium text-slate-600 hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors cursor-pointer"
                @click="activateSprint(sprint.id)"
              >
                <PlayIcon class="size-3.5" />
                Activate
              </button>
              <button
                class="inline-flex items-center rounded-md border border-slate-300 bg-white px-2 h-7 text-slate-400 hover:text-red-600 hover:border-red-300 focus-visible:outline-none transition-colors cursor-pointer"
                title="Delete sprint (moves issues to backlog)"
                @click="doDeleteSprint(sprint.id)"
              >
                <Trash2Icon class="size-3.5" />
              </button>
            </div>
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
              class="group flex items-center gap-3 px-6 py-2.5 hover:bg-slate-50 transition-colors cursor-grab active:cursor-grabbing border-l-4 border-b border-slate-100"
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
              <div class="flex-shrink-0 flex -space-x-1 w-10 justify-end">
                <Avatar
                  v-for="a in (issue.assignees ?? []).slice(0, 2)"
                  :key="a.id"
                  :name="a.display_name"
                  :src="a.avatar_url"
                  size="xs"
                  class="ring-1 ring-white"
                />
              </div>
              <button
                class="flex-shrink-0 opacity-0 group-hover:opacity-100 text-[11px] text-slate-400 hover:text-slate-600 transition-opacity cursor-pointer px-1.5 py-0.5 rounded hover:bg-slate-100"
                @click.stop="moveToBacklog(issue)"
              >
                ↓ Backlog
              </button>
            </div>
          </VueDraggable>
          <!-- Inline create row for planning sprint -->
          <div
            v-if="activeInlineCreate === sprint.id"
            class="flex items-center gap-3 px-6 py-2.5 border-b border-slate-100 border-l-4 border-l-blue-400 bg-blue-50/30"
          >
            <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
              <PlusIcon class="size-3 text-blue-400" />
            </div>
            <input
              :ref="setInlineCreateRef(sprint.id)"
              v-model="inlineCreateTitle"
              type="text"
              placeholder="Issue title — Enter to create, Esc to close"
              class="flex-1 min-w-0 text-sm text-slate-800 bg-transparent placeholder:text-slate-400 focus:outline-none"
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
            class="flex items-center gap-3 px-6 py-2 border-b border-slate-100 border-l-4 border-l-transparent cursor-text group/create"
            @click="activateInlineCreate(sprint.id)"
          >
            <div class="flex items-center gap-1.5 flex-shrink-0 w-24">
              <PlusIcon
                class="size-3 text-slate-300 group-hover/create:text-slate-400"
              />
            </div>
            <span
              class="text-sm text-slate-300 group-hover/create:text-slate-400"
              >Create issue</span
            >
          </div>
        </template>

        <!-- ── Backlog section header ─────────────────────────────────────── -->
        <div
          class="px-6 py-2.5 border-b border-slate-100 bg-white flex items-center justify-between"
        >
          <div class="flex items-center gap-2">
            <span class="font-semibold text-slate-900 text-sm">Backlog</span>
            <span class="text-xs text-slate-400 tabular-nums"
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
          <div class="max-w-lg space-y-3">
            <div class="text-sm font-medium text-slate-700">New sprint</div>

            <input
              v-model="newSprint.name"
              type="text"
              placeholder="Sprint name (required)"
              class="w-full rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />

            <div class="flex gap-3">
              <div class="flex-1">
                <label class="text-xs text-slate-500 mb-1 block"
                  >Start date</label
                >
                <input
                  v-model="newSprint.start_date"
                  type="date"
                  class="w-full rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
              <div class="flex-1">
                <label class="text-xs text-slate-500 mb-1 block"
                  >End date</label
                >
                <input
                  v-model="newSprint.end_date"
                  type="date"
                  class="w-full rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>

            <input
              v-model="newSprint.goal"
              type="text"
              placeholder="Sprint goal (optional)"
              class="w-full rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />

            <div v-if="newSprintError" class="text-xs text-red-600">
              {{ newSprintError }}
            </div>

            <div class="flex items-center gap-2">
              <button
                class="inline-flex items-center gap-1.5 rounded-md bg-blue-600 px-3 h-7 text-xs font-medium text-white hover:bg-blue-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors cursor-pointer disabled:opacity-50"
                :disabled="creatingSprintPending"
                @click="handleCreateSprint"
              >
                Create sprint
              </button>
              <button
                class="inline-flex items-center rounded-md border border-slate-300 bg-white px-3 h-7 text-xs font-medium text-slate-600 hover:bg-slate-50 focus-visible:outline-none transition-colors cursor-pointer"
                @click="
                  showNewSprintForm = false;
                  newSprintError = '';
                "
              >
                Cancel
              </button>
            </div>
          </div>
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
            class="group relative flex items-center gap-3 px-6 py-2.5 hover:bg-slate-50 transition-colors cursor-grab active:cursor-grabbing border-l-4 border-b border-slate-100"
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
            <div class="flex-shrink-0 flex -space-x-1 w-10 justify-end">
              <Avatar
                v-for="a in (issue.assignees ?? []).slice(0, 2)"
                :key="a.id"
                :name="a.display_name"
                :src="a.avatar_url"
                size="xs"
                class="ring-1 ring-white"
              />
            </div>

            <!-- Move to sprint dropdown -->
            <div v-if="targetSprints.length" class="relative flex-shrink-0">
              <button
                class="opacity-0 group-hover:opacity-100 text-[11px] text-slate-400 hover:text-slate-600 transition-opacity cursor-pointer px-1.5 py-0.5 rounded hover:bg-slate-100 inline-flex items-center gap-0.5"
                @click.stop="toggleDropdown(issue.id)"
              >
                → Sprint <ChevronDownIcon class="size-3" />
              </button>
              <div
                v-if="openDropdown === issue.id"
                class="absolute right-0 top-full mt-1 z-10 bg-white border border-slate-200 rounded-md shadow-md py-1 min-w-36"
              >
                <button
                  v-for="sprint in targetSprints"
                  :key="sprint.id"
                  class="w-full text-left px-3 py-1.5 text-xs text-slate-700 hover:bg-slate-50 cursor-pointer truncate"
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
              class="size-3 text-slate-300 group-hover/create:text-slate-400"
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
            class="text-[11px] font-medium text-slate-500 uppercase tracking-wide px-1 mb-0.5 flex items-center gap-1"
          >
            <ArrowRightIcon class="size-3" />
            Move to
          </div>
          <div v-for="sprint in targetSprints" :key="sprint.id">
            <VueDraggable
              v-model="overlayDropZones[sprint.id]"
              :group="{ name: 'backlog', put: true, pull: false }"
              :animation="0"
              class="rounded-lg border-2 border-dashed px-3 py-3 text-center min-h-12 transition-colors bg-white/90 backdrop-blur-sm shadow-lg"
              :class="
                overlayDropZones[sprint.id]?.length
                  ? 'border-blue-400 bg-blue-50/90'
                  : 'border-slate-300 hover:border-blue-300'
              "
              @add="(evt) => onDropToOverlayZone(evt, sprint.id)"
            >
              <div
                v-for="item in overlayDropZones[sprint.id]"
                :key="item.id"
                class="hidden"
              />
            </VueDraggable>
            <div class="text-xs text-slate-600 font-medium truncate mt-1 px-1">
              {{ sprint.name }}
              <span v-if="sprint.status === 'active'" class="text-blue-600"
                >(active)</span
              >
            </div>
          </div>
          <!-- Backlog drop zone -->
          <div class="mt-1 pt-2 border-t border-slate-200">
            <VueDraggable
              v-model="overlayDropZones[BACKLOG_KEY]"
              :group="{ name: 'backlog', put: true, pull: false }"
              :animation="0"
              class="rounded-lg border-2 border-dashed px-3 py-3 text-center min-h-12 transition-colors bg-white/90 backdrop-blur-sm shadow-lg"
              :class="
                overlayDropZones[BACKLOG_KEY]?.length
                  ? 'border-blue-400 bg-blue-50/90'
                  : 'border-slate-300 hover:border-blue-300'
              "
              @add="(evt) => onDropToOverlayZone(evt, BACKLOG_KEY)"
            >
              <div
                v-for="item in overlayDropZones[BACKLOG_KEY]"
                :key="item.id"
                class="hidden"
              />
            </VueDraggable>
            <div class="text-xs text-slate-600 font-medium mt-1 px-1">
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
