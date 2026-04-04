<!--
  IssueDetailView — full detail page for a single issue.

  Route: /projects/:slug/issues/:number

  For epics: shows EpicChildList with progress bar.
  For tasks: shows EpicSelector to assign/change/clear parent epic.
-->
<script setup>
import { computed, ref, nextTick, useId } from "vue";
import { useRoute, RouterLink } from "vue-router";
import { ESTIMATE_LABEL } from "@/composables/issueConstants";
import { useQuery, useMutation, useQueryClient } from "@tanstack/vue-query";
import { ArrowLeftIcon, LayersIcon, ZapIcon, ScissorsIcon, LinkIcon, SparklesIcon } from "lucide-vue-next";
import MainLayout from "@/layouts/MainLayout.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import Spinner from "@/components/ui/Spinner.vue";
import EpicSelector from "@/components/issue/EpicSelector.vue";
import EpicChildList from "@/components/issue/EpicChildList.vue";
import CommentSection from "@/components/issue/CommentSection.vue";
import StatusSelect from "@/components/issue/StatusSelect.vue";
import PrioritySelect from "@/components/issue/PrioritySelect.vue";
import AssigneeSelect from "@/components/issue/AssigneeSelect.vue";
import OwnerSelect from "@/components/issue/OwnerSelect.vue";
import LabelSelect from "@/components/issue/LabelSelect.vue";
import MilestoneSelect from "@/components/issue/MilestoneSelect.vue";
import { MdPreview } from "md-editor-v3";
import "md-editor-v3/lib/style.css";
import MarkdownEditor from "@/components/ui/MarkdownEditor.vue";
import RelativeTime from "@/components/ui/RelativeTime.vue";
import SplitIssueModal from "@/components/issue/SplitIssueModal.vue";
import RefinementPanel from "@/components/issue/RefinementPanel.vue";
import Modal from "@/components/ui/Modal.vue";
import { fetchIssue, updateIssue, createIssueLink } from "@/api/issues";
import { fetchProject } from "@/api/projects";
import { fetchSprints } from "@/api/sprints";
import { useTheme } from "@/composables/useTheme";
import { useRefinement } from "@/composables/useRefinement";

const { isDark } = useTheme();
const editorTheme = computed(() => (isDark.value ? "dark" : "light"));

const route = useRoute();
const queryClient = useQueryClient();
const previewId = useId();

const slug = computed(() => route.params.slug);
const number = computed(() => Number(route.params.number));

// ── Refinement ───────────────────────────────────────────────────────────
const refinement = useRefinement(slug, number);

// ── Data ──────────────────────────────────────────────────────────────────

const { data: project } = useQuery({
  queryKey: ["project", slug],
  queryFn: () => fetchProject(slug.value),
});

const { data: sprintsResult } = useQuery({
  queryKey: ["sprints", slug],
  queryFn: () => fetchSprints(slug.value),
  enabled: computed(() => !!slug.value),
});

const currentSprint = computed(() => {
  if (!issue.value?.sprint_id) return null;
  return (
    (sprintsResult.value?.sprints ?? []).find(
      (s) => s.id === issue.value.sprint_id,
    ) ?? null
  );
});

const { data: issue, isLoading } = useQuery({
  queryKey: ["issue", slug, number],
  queryFn: () => fetchIssue(slug.value, number.value),
  enabled: computed(() => !!slug.value && !!number.value),
});


// ── Status mutation ───────────────────────────────────────────────────────────

const showCancelConfirm = ref(false);
const cancelReasonDraft = ref("");

const { mutate: doUpdateStatus, isPending: cancelPending } = useMutation({
  mutationFn: ({ status, cancelReason }) =>
    updateIssue(slug.value, number.value, {
      status,
      ...(cancelReason ? { cancel_reason: cancelReason } : {}),
    }),
  onSuccess: () => {
    showCancelConfirm.value = false;
    cancelReasonDraft.value = "";
    queryClient.invalidateQueries({
      queryKey: ["issue", slug.value, number.value],
    });
    queryClient.invalidateQueries({ queryKey: ["issues", slug.value] });
  },
});

function updateStatus(status) {
  if (status === "cancelled") {
    cancelReasonDraft.value = "";
    showCancelConfirm.value = true;
  } else {
    doUpdateStatus({ status });
  }
}

function confirmCancel() {
  doUpdateStatus({ status: "cancelled", cancelReason: cancelReasonDraft.value.trim() || undefined });
}

// ── Link mutation ─────────────────────────────────────────────────────────────

const showLinkForm = ref(false);
const linkTypeDraft = ref("blocks");
const linkTargetDraft = ref("");

const { mutate: doAddLink, isPending: linkPending } = useMutation({
  mutationFn: (payload) => createIssueLink(slug.value, number.value, payload),
  onSuccess: () => {
    showLinkForm.value = false;
    linkTypeDraft.value = "blocks";
    linkTargetDraft.value = "";
    queryClient.invalidateQueries({
      queryKey: ["issue", slug.value, number.value],
    });
  },
});

function confirmLink() {
  const target = parseInt(linkTargetDraft.value, 10);
  if (!target) return;
  doAddLink({ link_type: linkTypeDraft.value, target_number: target });
}

// ── Hold mutation ─────────────────────────────────────────────────────────────

const showHoldModal = ref(false);
const holdReasonDraft = ref("waiting_on_external");
const holdNoteDraft = ref("");

const { mutate: doToggleHold, isPending: holdPending } = useMutation({
  mutationFn: (payload) => updateIssue(slug.value, number.value, payload),
  onSuccess: () => {
    showHoldModal.value = false;
    holdReasonDraft.value = "waiting_on_external";
    holdNoteDraft.value = "";
    queryClient.invalidateQueries({
      queryKey: ["issue", slug.value, number.value],
    });
    queryClient.invalidateQueries({ queryKey: ["issues", slug.value] });
  },
});

function confirmHold() {
  doToggleHold({
    on_hold: true,
    hold_reason: holdReasonDraft.value,
    hold_note: holdNoteDraft.value.trim() || undefined,
  });
}

function clearHold() {
  doToggleHold({ on_hold: false });
}

// ── Priority mutation ─────────────────────────────────────────────────────────

const { mutate: updatePriority } = useMutation({
  mutationFn: (priority) => updateIssue(slug.value, number.value, { priority }),
  onSuccess: () => {
    queryClient.invalidateQueries({
      queryKey: ["issue", slug.value, number.value],
    });
    queryClient.invalidateQueries({ queryKey: ["issues", slug.value] });
  },
});

// ── Epic assignment mutation (for tasks) ────────────────────────────────────

const { mutate: updateParent } = useMutation({
  mutationFn: (parentId) =>
    updateIssue(slug.value, number.value, { parent_id: parentId }),
  onSuccess: () => {
    queryClient.invalidateQueries({
      queryKey: ["issue", slug.value, number.value],
    });
    queryClient.invalidateQueries({ queryKey: ["issues", slug.value] });
  },
});

// ── Assignee mutation ─────────────────────────────────────────────────────────

const { mutate: updateAssignees } = useMutation({
  mutationFn: (assigneeIds) =>
    updateIssue(slug.value, number.value, { assignee_ids: assigneeIds }),
  onSuccess: () => {
    queryClient.invalidateQueries({
      queryKey: ["issue", slug.value, number.value],
    });
    queryClient.invalidateQueries({ queryKey: ["issues", slug.value] });
    queryClient.invalidateQueries({ queryKey: ["me", "issues"] });
  },
});

// ── Owner mutation ────────────────────────────────────────────────────────────

const { mutate: updateOwner } = useMutation({
  mutationFn: (ownerId) =>
    updateIssue(slug.value, number.value, { owner_id: ownerId ?? null }),
  onSuccess: () => {
    queryClient.invalidateQueries({
      queryKey: ["issue", slug.value, number.value],
    });
    queryClient.invalidateQueries({ queryKey: ["issues", slug.value] });
  },
});

// ── Title mutation ────────────────────────────────────────────────────────────

const editingTitle = ref(false);
const titleDraft = ref("");
const titleInputEl = ref(null);

function startEditingTitle() {
  titleDraft.value = issue.value.title;
  editingTitle.value = true;
  nextTick(() => {
    titleInputEl.value?.select();
  });
}

function cancelEditingTitle() {
  editingTitle.value = false;
}

const { mutate: updateTitle } = useMutation({
  mutationFn: (title) => updateIssue(slug.value, number.value, { title }),
  onSuccess: () => {
    editingTitle.value = false;
    queryClient.invalidateQueries({
      queryKey: ["issue", slug.value, number.value],
    });
    queryClient.invalidateQueries({ queryKey: ["issues", slug.value] });
  },
  onError: () => {
    editingTitle.value = false;
  },
});

function saveTitle() {
  const trimmed = titleDraft.value.trim();
  if (!trimmed || trimmed === issue.value.title) {
    editingTitle.value = false;
    return;
  }
  updateTitle(trimmed);
}

// ── Description mutation ──────────────────────────────────────────────────────

const editingDescription = ref(false);
const descriptionDraft = ref("");
const descriptionInputEl = ref(null);

function startEditingDescription() {
  descriptionDraft.value = issue.value.description ?? "";
  editingDescription.value = true;
  nextTick(() => {
    descriptionInputEl.value?.focus?.();
  });
}

function cancelEditingDescription() {
  editingDescription.value = false;
}

const { mutate: updateDescription } = useMutation({
  mutationFn: (description) =>
    updateIssue(slug.value, number.value, { description }),
  onSuccess: () => {
    editingDescription.value = false;
    queryClient.invalidateQueries({
      queryKey: ["issue", slug.value, number.value],
    });
  },
  onError: () => {
    editingDescription.value = false;
  },
});

function saveDescription() {
  const trimmed = descriptionDraft.value.trim() || null;
  if (trimmed === (issue.value.description ?? null)) {
    editingDescription.value = false;
    return;
  }
  updateDescription(trimmed);
}

// ── Split modal ───────────────────────────────────────────────────────────────

const showSplitModal = ref(false)

const isTerminal = computed(() => {
  const s = issue.value?.status
  return s === 'done' || s === 'cancelled' || s === 'closed'
})

// ── Label mutation ───────────────────────────────────────────────────────────

const { mutate: updateLabels } = useMutation({
  mutationFn: (labelIds) =>
    updateIssue(slug.value, number.value, { label_ids: labelIds }),
  onSuccess: () => {
    queryClient.invalidateQueries({
      queryKey: ["issue", slug.value, number.value],
    });
    queryClient.invalidateQueries({ queryKey: ["issues", slug.value] });
  },
});

// ── Milestone mutation ────────────────────────────────────────────────────────

const { mutate: updateMilestone } = useMutation({
  mutationFn: (milestoneId) =>
    updateIssue(slug.value, number.value, { milestone_id: milestoneId ?? "null" }),
  onSuccess: () => {
    queryClient.invalidateQueries({
      queryKey: ["issue", slug.value, number.value],
    });
    queryClient.invalidateQueries({ queryKey: ["milestones", slug.value] });
  },
});
</script>

<template>
  <MainLayout>
    <div class="flex flex-col h-full">
      <!-- Header -->
      <div
        class="flex-shrink-0 flex items-center gap-3 px-6 py-3 border-b border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900"
      >
        <RouterLink
          :to="`/projects/${slug}/backlog`"
          class="inline-flex items-center gap-1 text-sm text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 transition-colors"
        >
          <ArrowLeftIcon class="size-4" />
          Back
        </RouterLink>
        <div v-if="project" class="flex items-center gap-2 text-slate-400">
          <span
            class="size-6 rounded flex items-center justify-center text-[10px] font-semibold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300"
          >
            {{ project.slug.slice(0, 2).toUpperCase() }}
          </span>
          <span class="text-sm font-medium text-slate-600 dark:text-slate-400">{{
            project.name
          }}</span>
        </div>
        <div class="ml-auto flex items-center gap-2">
          <Button
            v-if="issue && issue.type === 'task' && !isTerminal"
            variant="secondary"
            size="sm"
            @click="refinement.open()"
          >
            <SparklesIcon class="size-3.5" />
            {{ refinement.session.value ? 'Continue refining' : 'Refine' }}
          </Button>
          <Button
            v-if="issue && issue.type === 'task' && !isTerminal"
            variant="secondary"
            size="sm"
            @click="showSplitModal = true"
          >
            <ScissorsIcon class="size-3.5" />
            Split
          </Button>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="isLoading" class="h-32 flex items-center justify-center">
        <Spinner class="size-6 text-slate-400" />
      </div>

      <!-- Content -->
      <div v-else-if="issue" class="flex-1 overflow-y-auto">
        <div class="max-w-3xl mx-auto px-6 py-8 space-y-8">
          <!-- Issue header -->
          <div class="space-y-3">
            <div class="flex items-center gap-2">
              <span class="text-xs font-mono text-slate-400 dark:text-slate-500"
                >{{ slug.toUpperCase() }}-{{ issue.number }}</span
              >
              <Badge v-if="issue.type === 'epic'" color-scheme="violet" compact>
                <LayersIcon class="size-3" />
                Epic
              </Badge>
              <Badge v-else color-scheme="blue" compact>Task</Badge>
            </div>
            <h1
              v-if="!editingTitle"
              class="text-2xl font-semibold text-slate-900 dark:text-slate-100 cursor-pointer hover:text-slate-700 dark:hover:text-slate-300 group"
              @click="startEditingTitle"
            >
              {{ issue.title }}
            </h1>
            <input
              v-else
              ref="titleInputEl"
              v-model="titleDraft"
              class="text-2xl font-semibold text-slate-900 dark:text-slate-100 w-full border-b-2 border-blue-500 focus:outline-none bg-transparent"
              @keydown.enter="saveTitle"
              @keydown.escape="cancelEditingTitle"
              @blur="saveTitle"
            />
          </div>

          <!-- Metadata grid -->
          <div class="grid grid-cols-2 gap-x-8 gap-y-4">
            <!-- Status -->
            <div class="space-y-1">
              <span class="text-xs font-medium text-slate-500 dark:text-slate-400">Status</span>
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
              <span class="text-xs font-medium text-slate-500 dark:text-slate-400">Priority</span>
              <div class="pt-1">
                <PrioritySelect
                  :priority="issue.priority ?? 'none'"
                  @update:priority="updatePriority"
                />
              </div>
            </div>

            <!-- Estimate -->
            <div class="space-y-1">
              <span class="text-xs font-medium text-slate-500 dark:text-slate-400">Estimate</span>
              <div>
                <span
                  v-if="ESTIMATE_LABEL[issue.estimate]"
                  class="text-sm font-medium text-slate-600 dark:text-slate-300 bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded"
                >
                  {{ ESTIMATE_LABEL[issue.estimate] }}
                </span>
                <span v-else class="text-sm text-slate-400 dark:text-slate-500">None</span>
              </div>
            </div>

            <!-- Assignees (tasks only) -->
            <div v-if="issue.type !== 'epic'" class="space-y-1">
              <div class="max-w-xs">
                <AssigneeSelect
                  :project-slug="slug"
                  :model-value="issue.assignees ?? []"
                  @update:model-value="updateAssignees"
                />
              </div>
            </div>

            <!-- Owner -->
            <div class="space-y-1">
              <div class="max-w-xs">
                <OwnerSelect
                  :project-slug="slug"
                  :model-value="issue.owner ?? null"
                  @update:model-value="updateOwner"
                />
              </div>
            </div>

            <!-- Labels -->
            <div class="space-y-1">
              <div class="max-w-xs">
                <LabelSelect
                  :project-slug="slug"
                  :model-value="issue.labels ?? []"
                  @update:model-value="updateLabels"
                />
              </div>
            </div>

            <!-- Sprint -->
            <div class="space-y-1">
              <span class="text-xs font-medium text-slate-500 dark:text-slate-400 flex items-center gap-1">
                <ZapIcon class="size-3" />
                Sprint
              </span>
              <div>
                <RouterLink
                  v-if="currentSprint"
                  :to="`/projects/${slug}/backlog`"
                  class="inline-flex items-center gap-1 text-sm text-blue-600 hover:text-blue-700 hover:underline"
                >
                  {{ currentSprint.name }}
                </RouterLink>
                <span v-else class="text-sm text-slate-400 dark:text-slate-500">Backlog</span>
              </div>
            </div>

            <!-- Milestone -->
            <div class="space-y-1">
              <div class="max-w-xs">
                <MilestoneSelect
                  :project-slug="slug"
                  :model-value="issue.milestone_id ?? null"
                  @update:model-value="updateMilestone"
                />
              </div>
            </div>

            <!-- Epic (tasks only) -->
            <div v-if="issue.type === 'task'" class="space-y-1">
              <div class="max-w-xs">
                <EpicSelector
                  :project-slug="slug"
                  :model-value="issue.parent_id"
                  @update:model-value="updateParent"
                />
              </div>
            </div>

            <!-- On hold -->
            <div class="space-y-1">
              <span class="text-xs font-medium text-slate-500 dark:text-slate-400">Hold</span>
              <div v-if="issue.on_hold" class="flex items-center gap-2">
                <Badge color-scheme="amber" compact>{{
                  issue.hold_reason?.replace(/_/g, ' ') ?? "on hold"
                }}</Badge>
                <span
                  v-if="issue.hold_note"
                  class="text-xs text-slate-500 dark:text-slate-400 italic truncate max-w-[160px]"
                  :title="issue.hold_note"
                  >{{ issue.hold_note }}</span
                >
                <button
                  class="text-xs text-amber-600 dark:text-amber-400 hover:underline"
                  @click="clearHold"
                >clear</button>
              </div>
              <div v-else class="pt-1">
                <button
                  class="text-xs text-slate-500 dark:text-slate-400 hover:text-amber-600 dark:hover:text-amber-400 hover:underline"
                  @click="showHoldModal = true"
                >Put on hold</button>
              </div>
            </div>

            <!-- Cancel reason -->
            <div v-if="issue.status === 'cancelled' && issue.cancel_reason" class="space-y-1 col-span-2">
              <span class="text-xs font-medium text-slate-500 dark:text-slate-400">Cancelled because</span>
              <p class="text-sm text-slate-500 dark:text-slate-400 italic">{{ issue.cancel_reason }}</p>
            </div>
          </div>

          <!-- Description -->
          <div class="space-y-2">
            <h2 class="text-sm font-medium text-slate-700 dark:text-slate-300">Description</h2>
            <MarkdownEditor
              v-if="editingDescription"
              ref="descriptionInputEl"
              v-model="descriptionDraft"
              placeholder="Add a description…"
              @save="saveDescription"
              @cancel="cancelEditingDescription"
            />
            <div v-else-if="issue.description" @click="startEditingDescription" class="cursor-pointer">
              <MdPreview :id="previewId" :model-value="issue.description" language="en-US" :theme="editorTheme" />
            </div>
            <button
              v-else
              class="text-sm text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-400 italic"
              @click="startEditingDescription"
            >
              Add a description…
            </button>
            <div v-if="editingDescription" class="flex items-center gap-2">
              <button
                class="text-xs font-medium text-white bg-blue-600 hover:bg-blue-700 px-3 py-1 rounded"
                @click="saveDescription"
              >
                Save
              </button>
              <button
                class="text-xs text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 px-2 py-1"
                @click="cancelEditingDescription"
              >
                Cancel
              </button>
              <span class="text-xs text-slate-400 dark:text-slate-500 ml-1">Ctrl+Enter to save · Esc to cancel</span>
            </div>
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
          <div
            v-if="issue.type === 'task' && issue.checklist?.length"
            class="space-y-2"
          >
            <h2 class="text-sm font-medium text-slate-700 dark:text-slate-300">Checklist</h2>
            <div class="space-y-1">
              <div
                v-for="item in issue.checklist"
                :key="item.id"
                class="flex items-center gap-2"
              >
                <input
                  type="checkbox"
                  :checked="item.done"
                  disabled
                  class="rounded border-slate-300"
                />
                <span
                  class="text-sm"
                  :class="
                    item.done ? 'text-slate-400 dark:text-slate-500 line-through' : 'text-slate-700 dark:text-slate-300'
                  "
                  >{{ item.text }}</span
                >
              </div>
            </div>
          </div>

          <!-- Links -->
          <div class="space-y-2">
            <h2 class="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
              <LinkIcon class="size-3.5" />
              Linked issues
            </h2>
            <div v-if="issue.links?.length" class="space-y-1">
              <div
                v-for="link in issue.links"
                :key="link.id"
                class="flex items-center gap-2 text-sm"
              >
                <span class="text-xs text-slate-400 dark:text-slate-500 capitalize min-w-[5rem]">
                  {{ link.link_type.replace(/_/g, ' ') }}
                </span>
                <RouterLink
                  :to="`/projects/${slug}/issues/${link.linked_issue_number}`"
                  class="text-blue-600 hover:text-blue-700 hover:underline font-mono text-xs"
                >
                  {{ slug.toUpperCase() }}-{{ link.linked_issue_number }}
                </RouterLink>
              </div>
            </div>
            <div v-if="showLinkForm" class="rounded-md border border-slate-200 dark:border-slate-700 p-3 space-y-3">
              <div class="flex items-center gap-2">
                <span class="text-xs text-slate-500 dark:text-slate-400">This issue</span>
                <select
                  v-model="linkTypeDraft"
                  class="rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-300 px-2 py-1.5 text-xs focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
                >
                  <option value="blocks">blocks</option>
                  <option value="is_blocked_by">is blocked by</option>
                  <option value="relates_to">relates to</option>
                  <option value="duplicates">duplicates</option>
                </select>
                <span class="text-xs text-slate-500 dark:text-slate-400">#</span>
                <input
                  v-model="linkTargetDraft"
                  type="number"
                  min="1"
                  placeholder="number"
                  class="w-20 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-300 px-2 py-1.5 text-xs focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
                  @keydown.enter="confirmLink"
                />
              </div>
              <div class="flex items-center gap-2">
                <Button size="xs" variant="primary" :loading="linkPending" @click="confirmLink">Add link</Button>
                <button
                  class="text-xs text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300"
                  @click="showLinkForm = false"
                >Cancel</button>
              </div>
            </div>
            <button
              v-if="!showLinkForm"
              class="text-xs text-slate-500 dark:text-slate-400 hover:text-blue-600 dark:hover:text-blue-400"
              @click="showLinkForm = true"
            >+ Add link</button>
          </div>

          <!-- Dates -->
          <div class="flex items-center gap-4 text-xs text-slate-400 dark:text-slate-500">
            <span>Created <RelativeTime :datetime="issue.created_at" /></span>
            <span v-if="issue.updated_at !== issue.created_at">
              · Updated <RelativeTime :datetime="issue.updated_at" />
            </span>
          </div>

          <!-- Comments -->
          <CommentSection :project-slug="slug" :issue-number="number" />
        </div>
      </div>

      <!-- Not found -->
      <div v-else class="flex-1 flex items-center justify-center">
        <p class="text-sm text-slate-400 dark:text-slate-500">Issue not found.</p>
      </div>
    </div>

    <!-- Split modal -->
    <SplitIssueModal
      v-if="issue"
      :open="showSplitModal"
      :issue="issue"
      :project-slug="slug"
      @close="showSplitModal = false"
      @split="showSplitModal = false"
    />

    <!-- Put on hold dialog -->
    <Modal
      :open="showHoldModal"
      title="Put issue on hold"
      description="This issue will be marked as on hold and shown with a visual indicator."
      @close="showHoldModal = false"
    >
      <div class="flex flex-col gap-3">
        <div class="flex flex-col gap-1.5">
          <label class="text-sm font-medium text-slate-700 dark:text-slate-300" for="hold-reason">Reason</label>
          <select
            id="hold-reason"
            v-model="holdReasonDraft"
            class="w-full rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-300 px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
          >
            <option value="waiting_on_customer">Waiting on customer</option>
            <option value="waiting_on_external">Waiting on external</option>
            <option value="blocked_by_issue">Blocked by issue</option>
          </select>
        </div>
        <div class="flex flex-col gap-1.5">
          <label class="text-sm font-medium text-slate-700 dark:text-slate-300" for="hold-note">
            Note <span class="font-normal text-slate-400 dark:text-slate-500">(optional)</span>
          </label>
          <textarea
            id="hold-note"
            v-model="holdNoteDraft"
            rows="2"
            placeholder="Additional context..."
            class="w-full rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-300 placeholder:text-slate-400 dark:placeholder:text-slate-500 px-3 py-2 text-sm resize-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
          />
        </div>
      </div>
      <template #footer>
        <Button variant="secondary" :disabled="holdPending" @click="showHoldModal = false">Cancel</Button>
        <Button variant="primary" :loading="holdPending" @click="confirmHold">Put on hold</Button>
      </template>
    </Modal>

    <!-- Cancel issue dialog (with optional reason) -->
    <Modal
      :open="showCancelConfirm"
      title="Cancel this issue?"
      description="This issue will be marked as cancelled and won't appear on the active board."
      @close="showCancelConfirm = false"
    >
      <div class="flex flex-col gap-1.5">
        <label class="text-sm font-medium text-slate-700 dark:text-slate-300" for="cancel-reason">
          Reason <span class="font-normal text-slate-400 dark:text-slate-500">(optional)</span>
        </label>
        <textarea
          id="cancel-reason"
          v-model="cancelReasonDraft"
          rows="2"
          placeholder="Why is this being cancelled?"
          class="w-full rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-300 placeholder:text-slate-400 dark:placeholder:text-slate-500 px-3 py-2 text-sm resize-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
        />
      </div>
      <template #footer>
        <Button variant="secondary" :disabled="cancelPending" @click="showCancelConfirm = false">Keep it</Button>
        <Button variant="destructive" :loading="cancelPending" @click="confirmCancel">Cancel issue</Button>
      </template>
    </Modal>

    <!-- Refinement panel -->
    <RefinementPanel
      :open="refinement.isOpen.value"
      :session="refinement.session.value"
      :loading="refinement.sessionLoading.value"
      :send-pending="refinement.sendPending.value"
      :accept-pending="refinement.acceptPending.value"
      @close="refinement.close()"
      @start="refinement.startSession()"
      @send="(content) => refinement.sendMessage(content)"
      @accept="refinement.acceptProposal()"
    />
  </MainLayout>
</template>
