<script setup>
import { ref, computed } from "vue";
import { RouterLink } from "vue-router";
import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { ZapIcon } from "lucide-vue-next";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import Modal from "@/components/ui/Modal.vue";
import StatusSelect from "@/components/issue/StatusSelect.vue";
import PrioritySelect from "@/components/issue/PrioritySelect.vue";
import AssigneeSelect from "@/components/issue/AssigneeSelect.vue";
import OwnerSelect from "@/components/issue/OwnerSelect.vue";
import LabelSelect from "@/components/issue/LabelSelect.vue";
import MilestoneSelect from "@/components/issue/MilestoneSelect.vue";
import EpicSelector from "@/components/issue/EpicSelector.vue";
import { ESTIMATE_LABEL } from "@/composables/issueConstants";
import { useIssueField } from "@/composables/useIssueField";
import { updateIssue } from "@/api/issues";

const props = defineProps({
  issue: { type: Object, required: true },
  slug: { type: String, required: true },
  number: { type: Number, required: true },
  archetype: { type: String, default: "software" },
  currentSprint: { type: Object, default: null },
});

const queryClient = useQueryClient();
const slugRef = computed(() => props.slug);
const numberRef = computed(() => props.number);

// ── Simple field mutations ──────────────────────────────────────────────────

const { mutate: updatePriority } = useIssueField(slugRef, numberRef, {
  toPayload: (v) => ({ priority: v }),
});
const { mutate: updateParent } = useIssueField(slugRef, numberRef, {
  toPayload: (v) => ({ parent_id: v }),
});
const { mutate: updateAssignees } = useIssueField(slugRef, numberRef, {
  toPayload: (v) => ({ assignee_ids: v }),
  extraInvalidate: [["me", "issues"]],
});
const { mutate: updateOwner } = useIssueField(slugRef, numberRef, {
  toPayload: (v) => ({ owner_id: v ?? null }),
});
const { mutate: updateLabels } = useIssueField(slugRef, numberRef, {
  toPayload: (v) => ({ label_ids: v }),
});
const { mutate: updateMilestone } = useIssueField(slugRef, numberRef, {
  toPayload: (v) => ({ milestone_id: v ?? "null" }),
  extraInvalidate: [["milestones", props.slug]],
});

// ── Status mutation (with cancel confirm) ───────────────────────────────────

const showCancelConfirm = ref(false);
const cancelReasonDraft = ref("");

const { mutate: doUpdateStatus, isPending: cancelPending } = useMutation({
  mutationFn: ({ status, cancelReason }) =>
    updateIssue(props.slug, props.number, {
      status,
      ...(cancelReason ? { cancel_reason: cancelReason } : {}),
    }),
  onSuccess: () => {
    showCancelConfirm.value = false;
    cancelReasonDraft.value = "";
    queryClient.invalidateQueries({
      queryKey: ["issue", props.slug, props.number],
    });
    queryClient.invalidateQueries({ queryKey: ["issues", props.slug] });
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
  doUpdateStatus({
    status: "cancelled",
    cancelReason: cancelReasonDraft.value.trim() || undefined,
  });
}

// ── Hold mutation ───────────────────────────────────────────────────────────

const showHoldModal = ref(false);
const holdReasonDraft = ref("waiting_on_external");
const holdNoteDraft = ref("");

const { mutate: doToggleHold, isPending: holdPending } = useMutation({
  mutationFn: (payload) => updateIssue(props.slug, props.number, payload),
  onSuccess: () => {
    showHoldModal.value = false;
    holdReasonDraft.value = "waiting_on_external";
    holdNoteDraft.value = "";
    queryClient.invalidateQueries({
      queryKey: ["issue", props.slug, props.number],
    });
    queryClient.invalidateQueries({ queryKey: ["issues", props.slug] });
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
</script>

<template>
  <!-- Metadata grid -->
  <div class="grid grid-cols-2 gap-x-8 gap-y-4">
    <!-- Status -->
    <div class="space-y-1">
      <span class="text-xs font-medium text-slate-500 dark:text-slate-400">Status</span>
      <div class="pt-1">
        <StatusSelect
          :status="issue.status"
          :archetype="archetype"
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
          issue.hold_reason?.replace(/_/g, " ") ?? "on hold"
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

  <!-- Cancel issue dialog -->
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
</template>
