<script setup>
import { ref } from "vue";
import { RouterLink } from "vue-router";
import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { LinkIcon } from "lucide-vue-next";
import Button from "@/components/ui/Button.vue";
import { createIssueLink } from "@/api/issues";

const props = defineProps({
  slug: { type: String, required: true },
  number: { type: Number, required: true },
  links: { type: Array, default: () => [] },
});

const queryClient = useQueryClient();

const showForm = ref(false);
const linkTypeDraft = ref("blocks");
const linkTargetDraft = ref("");

const { mutate: doAddLink, isPending: linkPending } = useMutation({
  mutationFn: (payload) => createIssueLink(props.slug, props.number, payload),
  onSuccess: () => {
    showForm.value = false;
    linkTypeDraft.value = "blocks";
    linkTargetDraft.value = "";
    queryClient.invalidateQueries({
      queryKey: ["issue", props.slug, props.number],
    });
  },
});

function confirmLink() {
  const target = parseInt(linkTargetDraft.value, 10);
  if (!target) return;
  doAddLink({ link_type: linkTypeDraft.value, target_number: target });
}
</script>

<template>
  <div class="space-y-2">
    <h2 class="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
      <LinkIcon class="size-3.5" />
      Linked issues
    </h2>
    <div v-if="links?.length" class="space-y-1">
      <div
        v-for="link in links"
        :key="link.id"
        class="flex items-center gap-2 text-sm"
      >
        <span class="text-xs text-slate-400 dark:text-slate-500 capitalize min-w-[5rem]">
          {{ link.link_type.replace(/_/g, " ") }}
        </span>
        <RouterLink
          :to="`/projects/${slug}/issues/${link.linked_issue_number}`"
          class="text-blue-600 hover:text-blue-700 hover:underline font-mono text-xs"
        >
          {{ slug.toUpperCase() }}-{{ link.linked_issue_number }}
        </RouterLink>
      </div>
    </div>
    <div v-if="showForm" class="rounded-md border border-slate-200 dark:border-slate-700 p-3 space-y-3">
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
          @click="showForm = false"
        >Cancel</button>
      </div>
    </div>
    <button
      v-if="!showForm"
      class="text-xs text-slate-500 dark:text-slate-400 hover:text-blue-600 dark:hover:text-blue-400"
      @click="showForm = true"
    >+ Add link</button>
  </div>
</template>
