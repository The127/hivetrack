import { ref, nextTick } from "vue";
import { useMutation } from "@tanstack/vue-query";
import { createIssue } from "@/api/issues";

/**
 * Composable for inline issue creation in backlog sections.
 *
 * @param {import('vue').Ref<string>} slug - project slug ref
 * @param {import('vue').ComputedRef<string>} archetype - project archetype computed
 * @param {import('@tanstack/vue-query').QueryClient} queryClient - TanStack query client
 */
export function useInlineCreate(slug, archetype, queryClient) {
  const activeInlineCreate = ref(null);
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
      if (!el) return;
      el.scrollIntoView({ behavior: "instant", block: "nearest" });
      el.focus();
    });
  }

  const { mutate: inlineCreate, isPending: inlineCreatePending } = useMutation({
    mutationFn: (data) => createIssue(slug.value, data),
    onSuccess: () => {
      inlineCreateTitle.value = "";
      inlineCreateError.value = "";
      queryClient.invalidateQueries({ queryKey: ["issues", slug.value] });
      nextTick(() => {
        if (activeInlineCreate.value) {
          const el = inlineCreateInputs.value[activeInlineCreate.value];
          if (!el) return;
          el.scrollIntoView({ behavior: "instant", block: "nearest" });
          el.focus();
        }
      });
    },
    onError: () => {
      inlineCreateError.value = "Failed";
    },
  });

  function submitInlineCreate(sectionId) {
    const BACKLOG_KEY = "__backlog__";
    const title = inlineCreateTitle.value.trim();
    if (!title) return;
    if (inlineCreatePending.value) return;
    const status = archetype.value === "support" ? "open" : "todo";
    const sprintId = sectionId === BACKLOG_KEY ? undefined : sectionId;
    inlineCreate({ title, type: "task", status, sprint_id: sprintId });
  }

  function cancelInlineCreate() {
    if (inlineCreatePending.value) return;
    activeInlineCreate.value = null;
    inlineCreateTitle.value = "";
    inlineCreateError.value = "";
  }

  return {
    activeInlineCreate,
    inlineCreateTitle,
    inlineCreateError,
    inlineCreatePending,
    setInlineCreateRef,
    activateInlineCreate,
    submitInlineCreate,
    cancelInlineCreate,
  };
}
