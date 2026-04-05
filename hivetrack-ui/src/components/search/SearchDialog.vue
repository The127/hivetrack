<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { SearchIcon, XIcon, LoaderIcon } from "lucide-vue-next";
import Badge from "@/components/ui/Badge.vue";
import { fetchIssues } from "@/api/issues";
import { fetchProjects } from "@/api/projects";
import { statusLabel, statusScheme } from "@/composables/issueConstants";

const props = defineProps({
  open: { type: Boolean, required: true },
  projectSlug: { type: String, default: null },
});

const emit = defineEmits(["close"]);

const router = useRouter();
const inputRef = ref(null);
const query = ref("");
const loading = ref(false);
const results = ref([]);
const activeIndex = ref(0);
const scopeSlug = ref(null);

// Sync scope when dialog opens
watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      query.value = "";
      results.value = [];
      activeIndex.value = 0;
      scopeSlug.value = props.projectSlug;
      nextTick(() => inputRef.value?.focus());
    }
  },
);

// Debounced search
let debounceTimer = null;
watch(query, (text) => {
  clearTimeout(debounceTimer);
  if (!text.trim()) {
    results.value = [];
    loading.value = false;
    return;
  }
  loading.value = true;
  debounceTimer = setTimeout(() => doSearch(text.trim()), 300);
});

// Re-search when scope changes
watch(scopeSlug, () => {
  if (query.value.trim()) {
    loading.value = true;
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => doSearch(query.value.trim()), 100);
  }
});

async function doSearch(text) {
  try {
    if (scopeSlug.value) {
      const res = await fetchIssues(scopeSlug.value, { text, limit: 20 });
      results.value = (res.items ?? []).map((i) => ({
        ...i,
        _slug: scopeSlug.value,
      }));
    } else {
      const projectsRes = await fetchProjects();
      const projects = projectsRes.items ?? projectsRes ?? [];
      const all = await Promise.all(
        projects.map(async (p) => {
          const res = await fetchIssues(p.slug, { text, limit: 10 });
          return (res.items ?? []).map((i) => ({ ...i, _slug: p.slug }));
        }),
      );
      results.value = all.flat();
    }
  } catch {
    results.value = [];
  } finally {
    loading.value = false;
    activeIndex.value = 0;
  }
}

// Flat list for keyboard nav
const flatResults = computed(() => results.value);

// Group results by project when searching all
const grouped = computed(() => {
  if (scopeSlug.value) return null;
  const map = new Map();
  for (const r of results.value) {
    if (!map.has(r._slug)) map.set(r._slug, []);
    map.get(r._slug).push(r);
  }
  return map;
});

function clearScope() {
  scopeSlug.value = null;
}

function navigateTo(issue) {
  router.push(`/projects/${issue._slug}/issues/${issue.number}`);
  emit("close");
}

function onKeydown(e) {
  if (e.key === "Escape") {
    emit("close");
    return;
  }
  if (e.key === "ArrowDown") {
    e.preventDefault();
    activeIndex.value = Math.min(
      activeIndex.value + 1,
      flatResults.value.length - 1,
    );
    scrollActiveIntoView();
    return;
  }
  if (e.key === "ArrowUp") {
    e.preventDefault();
    activeIndex.value = Math.max(activeIndex.value - 1, 0);
    scrollActiveIntoView();
    return;
  }
  if (e.key === "Enter") {
    e.preventDefault();
    const item = flatResults.value[activeIndex.value];
    if (item) navigateTo(item);
    return;
  }
}

function scrollActiveIntoView() {
  nextTick(() => {
    const el = document.querySelector("[data-search-active]");
    el?.scrollIntoView({ block: "nearest" });
  });
}

function onBackdropClick() {
  emit("close");
}

// Global keydown for Escape even when input not focused
onMounted(() => document.addEventListener("keydown", onKeydownGlobal));
onUnmounted(() => document.removeEventListener("keydown", onKeydownGlobal));

function onKeydownGlobal(e) {
  if (e.key === "Escape" && props.open) {
    emit("close");
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-150"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-100"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="open"
        class="fixed inset-0 z-50 flex items-start justify-center pt-[15vh] p-4"
      >
        <!-- Backdrop -->
        <div
          class="absolute inset-0 bg-black/40 dark:bg-black/60"
          aria-hidden="true"
          @click="onBackdropClick"
        />

        <!-- Panel -->
        <div
          role="dialog"
          aria-label="Search issues"
          class="relative z-10 w-full max-w-xl rounded-xl bg-white dark:bg-slate-900 shadow-2xl ring-1 ring-slate-900/10 dark:ring-slate-700 flex flex-col max-h-[60vh]"
        >
          <!-- Search input -->
          <div
            class="flex items-center gap-3 px-4 border-b border-slate-100 dark:border-slate-800"
          >
            <SearchIcon
              class="size-4 text-slate-400 dark:text-slate-500 flex-shrink-0"
            />
            <input
              ref="inputRef"
              v-model="query"
              type="text"
              placeholder="Search issues..."
              class="flex-1 h-12 bg-transparent text-sm text-slate-900 dark:text-slate-100 placeholder:text-slate-400 dark:placeholder:text-slate-500 focus:outline-none"
              @keydown="onKeydown"
            />
            <kbd
              class="hidden sm:inline-block text-[10px] font-mono text-slate-400 dark:text-slate-500 bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded"
              >ESC</kbd
            >
          </div>

          <!-- Scope chip -->
          <div
            v-if="scopeSlug"
            class="flex items-center gap-2 px-4 py-2 border-b border-slate-100 dark:border-slate-800"
          >
            <span class="text-xs text-slate-500 dark:text-slate-400"
              >Searching in</span
            >
            <span
              class="inline-flex items-center gap-1 text-xs font-medium text-slate-700 dark:text-slate-300 bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded-full"
            >
              {{ scopeSlug }}
              <button
                class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 cursor-pointer"
                title="Search all projects"
                @click="clearScope"
              >
                <XIcon class="size-3" />
              </button>
            </span>
          </div>

          <!-- Results -->
          <div class="overflow-y-auto flex-1">
            <!-- Loading -->
            <div
              v-if="loading"
              class="flex items-center justify-center py-8 text-slate-400 dark:text-slate-500"
            >
              <LoaderIcon class="size-5 animate-spin" />
            </div>

            <!-- No query -->
            <div
              v-else-if="!query.trim()"
              class="py-8 text-center text-sm text-slate-400 dark:text-slate-500"
            >
              Type to search issues
            </div>

            <!-- No results -->
            <div
              v-else-if="!results.length"
              class="py-8 text-center text-sm text-slate-400 dark:text-slate-500"
            >
              No issues found
            </div>

            <!-- Scoped results (flat list) -->
            <template v-else-if="scopeSlug">
              <button
                v-for="(issue, i) in flatResults"
                :key="issue.id"
                :data-search-active="i === activeIndex ? '' : undefined"
                :class="
                  i === activeIndex ? 'bg-blue-50 dark:bg-blue-900/30' : ''
                "
                class="w-full flex items-center gap-3 px-4 py-2.5 text-left hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors cursor-pointer"
                @click="navigateTo(issue)"
                @mouseenter="activeIndex = i"
              >
                <span
                  class="text-[11px] font-mono text-slate-400 dark:text-slate-500 flex-shrink-0 w-24"
                  >{{ issue._slug.toUpperCase() }}-{{ issue.number }}</span
                >
                <span
                  class="flex-1 min-w-0 text-sm text-slate-800 dark:text-slate-200 truncate"
                  >{{ issue.title }}</span
                >
                <Badge :color-scheme="statusScheme(issue.status)" compact>
                  {{ statusLabel(issue.status) }}
                </Badge>
              </button>
            </template>

            <!-- Grouped results (all projects) -->
            <template v-else>
              <template v-for="[slug, issues] in grouped" :key="slug">
                <div
                  class="px-4 pt-3 pb-1 text-[11px] font-medium uppercase tracking-wider text-slate-400 dark:text-slate-500"
                >
                  {{ slug }}
                </div>
                <button
                  v-for="issue in issues"
                  :key="issue.id"
                  :data-search-active="
                    flatResults.indexOf(issue) === activeIndex ? '' : undefined
                  "
                  :class="
                    flatResults.indexOf(issue) === activeIndex
                      ? 'bg-blue-50 dark:bg-blue-900/30'
                      : ''
                  "
                  class="w-full flex items-center gap-3 px-4 py-2.5 text-left hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors cursor-pointer"
                  @click="navigateTo(issue)"
                  @mouseenter="activeIndex = flatResults.indexOf(issue)"
                >
                  <span
                    class="text-[11px] font-mono text-slate-400 dark:text-slate-500 flex-shrink-0 w-24"
                    >{{ slug.toUpperCase() }}-{{ issue.number }}</span
                  >
                  <span
                    class="flex-1 min-w-0 text-sm text-slate-800 dark:text-slate-200 truncate"
                    >{{ issue.title }}</span
                  >
                  <Badge :color-scheme="statusScheme(issue.status)" compact>
                    {{ statusLabel(issue.status) }}
                  </Badge>
                </button>
              </template>
            </template>
          </div>

          <!-- Footer hint -->
          <div
            v-if="results.length"
            class="flex items-center gap-3 px-4 py-2 border-t border-slate-100 dark:border-slate-800 text-[11px] text-slate-400 dark:text-slate-500"
          >
            <span><kbd class="font-mono">↑↓</kbd> navigate</span>
            <span><kbd class="font-mono">↵</kbd> open</span>
            <span><kbd class="font-mono">esc</kbd> close</span>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
