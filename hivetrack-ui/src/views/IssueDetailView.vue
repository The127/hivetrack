<!--
  IssueDetailView — full detail page for a single issue.

  Route: /projects/:slug/issues/:number

  For epics: shows EpicChildList with progress bar.
  For tasks: shows EpicSelector to assign/change/clear parent epic.
-->
<script setup>
import { computed, ref, nextTick, useId } from "vue";
import { useRoute, RouterLink } from "vue-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/vue-query";
import { ArrowLeftIcon, LayersIcon, ZapIcon, ScissorsIcon, LinkIcon } from "lucide-vue-next";
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
import ConfirmDialog from "@/components/ui/ConfirmDialog.vue";
import { fetchIssue, updateIssue } from "@/api/issues";
import { fetchProject } from "@/api/projects";
import { fetchSprints } from "@/api/sprints";

const route = useRoute();
const queryClient = useQueryClient();
const previewId = useId();

const slug = computed(() => route.params.slug);
const number = computed(() => Number(route.params.number));

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

const ESTIMATE_LABEL = {
  none: null,
  xs: "XS",
  s: "S",
  m: "M",
  l: "L",
  xl: "XL",
};

// ── Status mutation ───────────────────────────────────────────────────────────

const showCancelConfirm = ref(false);

const { mutate: doUpdateStatus, isPending: cancelPending } = useMutation({
  mutationFn: (status) => updateIssue(slug.value, number.value, { status }),
  onSuccess: () => {
    showCancelConfirm.value = false;
    queryClient.invalidateQueries({
      queryKey: ["issue", slug.value, number.value],
    });
    queryClient.invalidateQueries({ queryKey: ["issues", slug.value] });
  },
});

function updateStatus(status) {
  if (status === "cancelled") {
    showCancelConfirm.value = true;
  } else {
    doUpdateStatus(status);
  }
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
        class="flex-shrink-0 flex items-center gap-3 px-6 py-3 border-b border-slate-200 bg-white"
      >
        <RouterLink
          :to="`/projects/${slug}/backlog`"
          class="inline-flex items-center gap-1 text-sm text-slate-500 hover:text-slate-700 transition-colors"
        >
          <ArrowLeftIcon class="size-4" />
          Back
        </RouterLink>
        <div v-if="project" class="flex items-center gap-2 text-slate-400">
          <span
            class="size-6 rounded flex items-center justify-center text-[10px] font-semibold bg-slate-100 text-slate-600"
          >
            {{ project.slug.slice(0, 2).toUpperCase() }}
          </span>
          <span class="text-sm font-medium text-slate-600">{{
            project.name
          }}</span>
        </div>
        <div class="ml-auto">
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
      <div v-if="isLoading" class="flex-1 flex items-center justify-center">
        <Spinner class="size-6 text-slate-400" />
      </div>

      <!-- Content -->
      <div v-else-if="issue" class="flex-1 overflow-y-auto">
        <div class="max-w-3xl mx-auto px-6 py-8 space-y-8">
          <!-- Issue header -->
          <div class="space-y-3">
            <div class="flex items-center gap-2">
              <span class="text-xs font-mono text-slate-400"
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
              class="text-2xl font-semibold text-slate-900 cursor-pointer hover:text-slate-700 group"
              @click="startEditingTitle"
            >
              {{ issue.title }}
            </h1>
            <input
              v-else
              ref="titleInputEl"
              v-model="titleDraft"
              class="text-2xl font-semibold text-slate-900 w-full border-b-2 border-blue-500 focus:outline-none bg-transparent"
              @keydown.enter="saveTitle"
              @keydown.escape="cancelEditingTitle"
              @blur="saveTitle"
            />
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
                <span
                  v-if="ESTIMATE_LABEL[issue.estimate]"
                  class="text-sm font-medium text-slate-600 bg-slate-100 px-2 py-0.5 rounded"
                >
                  {{ ESTIMATE_LABEL[issue.estimate] }}
                </span>
                <span v-else class="text-sm text-slate-400">None</span>
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
              <span class="text-xs font-medium text-slate-500 flex items-center gap-1">
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
                <span v-else class="text-sm text-slate-400">Backlog</span>
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
            <div v-if="issue.on_hold" class="space-y-1">
              <span class="text-xs font-medium text-slate-500">On Hold</span>
              <div class="flex items-center gap-2">
                <Badge color-scheme="amber" compact>{{
                  issue.hold_reason ?? "on hold"
                }}</Badge>
                <span
                  v-if="issue.hold_note"
                  class="text-xs text-slate-500 italic"
                  >{{ issue.hold_note }}</span
                >
              </div>
            </div>
          </div>

          <!-- Description -->
          <div class="space-y-2">
            <h2 class="text-sm font-medium text-slate-700">Description</h2>
            <MarkdownEditor
              v-if="editingDescription"
              ref="descriptionInputEl"
              v-model="descriptionDraft"
              placeholder="Add a description…"
              @save="saveDescription"
              @cancel="cancelEditingDescription"
            />
            <div v-else-if="issue.description" @click="startEditingDescription" class="cursor-pointer">
              <MdPreview :id="previewId" :model-value="issue.description" language="en-US" />
            </div>
            <button
              v-else
              class="text-sm text-slate-400 hover:text-slate-600 italic"
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
                class="text-xs text-slate-500 hover:text-slate-700 px-2 py-1"
                @click="cancelEditingDescription"
              >
                Cancel
              </button>
              <span class="text-xs text-slate-400 ml-1">Ctrl+Enter to save · Esc to cancel</span>
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
            <h2 class="text-sm font-medium text-slate-700">Checklist</h2>
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
                    item.done ? 'text-slate-400 line-through' : 'text-slate-700'
                  "
                  >{{ item.text }}</span
                >
              </div>
            </div>
          </div>

          <!-- Links -->
          <div v-if="issue.links?.length" class="space-y-2">
            <h2 class="text-sm font-medium text-slate-700 flex items-center gap-1.5">
              <LinkIcon class="size-3.5" />
              Linked issues
            </h2>
            <div class="space-y-1">
              <div
                v-for="link in issue.links"
                :key="link.id"
                class="flex items-center gap-2 text-sm"
              >
                <span class="text-xs text-slate-400 capitalize min-w-[5rem]">
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
          </div>

          <!-- Dates -->
          <div class="flex items-center gap-4 text-xs text-slate-400">
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
        <p class="text-sm text-slate-400">Issue not found.</p>
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

    <!-- Cancel issue confirmation -->
    <ConfirmDialog
      :open="showCancelConfirm"
      title="Cancel this issue?"
      message="This issue will be marked as cancelled and won't appear on the active board."
      confirm-text="Cancel issue"
      :loading="cancelPending"
      @confirm="doUpdateStatus('cancelled')"
      @cancel="showCancelConfirm = false"
    />
  </MainLayout>
</template>
