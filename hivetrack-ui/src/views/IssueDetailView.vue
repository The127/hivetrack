<!--
  IssueDetailView — full detail page for a single issue.

  Route: /projects/:slug/issues/:number

  For epics: shows EpicChildList with progress bar.
  For tasks: shows EpicSelector to assign/change/clear parent epic.
-->
<script setup>
import { computed, ref, nextTick } from "vue";
import { useRoute, RouterLink } from "vue-router";
import { isTerminalStatus } from "@/composables/issueConstants";
import { useQuery, useMutation, useQueryClient } from "@tanstack/vue-query";
import { ArrowLeftIcon, LayersIcon, ScissorsIcon, SparklesIcon } from "lucide-vue-next";
import MainLayout from "@/layouts/MainLayout.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import Spinner from "@/components/ui/Spinner.vue";
import EpicChildList from "@/components/issue/EpicChildList.vue";
import CommentSection from "@/components/issue/CommentSection.vue";
import IssueDetailSidebar from "@/components/issue/IssueDetailSidebar.vue";
import IssueDescription from "@/components/issue/IssueDescription.vue";
import IssueLinks from "@/components/issue/IssueLinks.vue";
import "md-editor-v3/lib/style.css";
import RelativeTime from "@/components/ui/RelativeTime.vue";
import SplitIssueModal from "@/components/issue/SplitIssueModal.vue";
import RefinementPanel from "@/components/issue/RefinementPanel.vue";
import { fetchIssue, updateIssue } from "@/api/issues";
import { fetchProject } from "@/api/projects";
import { fetchSprints } from "@/api/sprints";
import { useRefinement } from "@/composables/useRefinement";

const route = useRoute();
const queryClient = useQueryClient();

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

// ── Title editing ────────────────────────────────────────────────────────

const editingTitle = ref(false);
const titleDraft = ref("");
const titleInputEl = ref(null);

function handleTitleClick() {
  if (window.getSelection()?.toString()) return;
  startEditingTitle();
}

function startEditingTitle() {
  titleDraft.value = issue.value.title;
  editingTitle.value = true;
  nextTick(() => titleInputEl.value?.select());
}

function cancelEditingTitle() {
  editingTitle.value = false;
}

const { mutate: updateTitle } = useMutation({
  mutationFn: (title) => updateIssue(slug.value, number.value, { title }),
  onSuccess: () => {
    editingTitle.value = false;
    queryClient.invalidateQueries({ queryKey: ["issue", slug.value, number.value] });
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

// ── Split modal ──────────────────────────────────────────────────────────

const showSplitModal = ref(false);

const isTerminal = computed(() =>
  isTerminalStatus(issue.value?.status, project.value?.archetype),
);
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
              @mouseup="handleTitleClick"
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

          <!-- Metadata grid + modals -->
          <IssueDetailSidebar
            :issue="issue"
            :slug="slug"
            :number="number"
            :archetype="project?.archetype ?? 'software'"
            :current-sprint="currentSprint"
          />

          <!-- Description -->
          <IssueDescription
            :slug="slug"
            :number="number"
            :description="issue.description"
          />

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
          <IssueLinks
            :slug="slug"
            :number="number"
            :links="issue.links"
          />

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

    <!-- Refinement panel -->
    <RefinementPanel
      :open="refinement.isOpen.value"
      :session="refinement.session.value"
      :loading="refinement.sessionLoading.value"
      :send-pending="refinement.sendPending.value"
      :accept-pending="refinement.acceptPending.value"
      :advance-pending="refinement.advancePending.value"
      :current-phase="refinement.currentPhase.value"
      @close="refinement.close()"
      @start="refinement.startSession()"
      @send="(content) => refinement.sendMessage(content)"
      @accept="refinement.acceptProposal()"
      @advance-phase="(targetPhase) => refinement.advancePhase(targetPhase)"
    />
  </MainLayout>
</template>
