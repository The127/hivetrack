<!--
  MainLayout — the primary application shell.

  Renders a fixed left sidebar with navigation and a scrollable main
  content area. Used for all authenticated views.

  The sidebar adapts based on the current route:
  - Always shows: Dashboard, Projects, Search hint
  - When route has :slug param: also shows the project-level navigation
    (Board, Backlog, Triage, Sprints, Milestones, Settings)

  Keyboard shortcuts (global, active while this layout is mounted):
    /         Focus the search input (not yet implemented — placeholder)
    Cmd+K     Open command palette (not yet implemented — placeholder)
    C         Create a new issue (emits 'create-issue')
-->
<script setup>
import { computed, onMounted, onBeforeUnmount, ref, watch } from "vue";
import { RouterLink, useRoute } from "vue-router";
import {
  LayoutDashboardIcon,
  FolderKanbanIcon,
  LayoutIcon,
  KanbanIcon,
  ListIcon,
  LayersIcon,
  InboxIcon,
  FlagIcon,
  TimerIcon,
  SettingsIcon,
  SearchIcon,
  ChevronRightIcon,
  ChevronLeftIcon,
  LogOutIcon,
} from "lucide-vue-next";
import Avatar from "@/components/ui/Avatar.vue";
import { useAuth } from "@/composables/useAuth";

const route = useRoute();
const { user, signOut } = useAuth();

// True when navigated inside a project (route has a :slug param).
const projectSlug = computed(() => route.params.slug ?? null);

// Display name for the current user. Falls back to email, then "You".
const userName = computed(
  () => user.value?.profile?.name ?? user.value?.profile?.email ?? "You",
);

// ── Sidebar collapsed state ────────────────────────────────────────────────────

const collapsed = ref(localStorage.getItem('hivetrack:sidebar:collapsed') === 'true');
watch(collapsed, (v) => localStorage.setItem('hivetrack:sidebar:collapsed', String(v)));

// ── Keyboard shortcuts ────────────────────────────────────────────────────────

const emit = defineEmits(["create-issue"]);

function handleKeydown(e) {
  // Ignore shortcuts when focus is inside an input, textarea, or contenteditable.
  const tag = document.activeElement?.tagName;
  if (
    tag === "INPUT" ||
    tag === "TEXTAREA" ||
    document.activeElement?.isContentEditable
  )
    return;

  if ((e.metaKey || e.ctrlKey) && e.key === "k") {
    e.preventDefault();
    // TODO: open command palette
    return;
  }

  if (e.key === "c" && !e.metaKey && !e.ctrlKey) {
    emit("create-issue");
  }
}

onMounted(() => window.addEventListener("keydown", handleKeydown));
onBeforeUnmount(() => window.removeEventListener("keydown", handleKeydown));
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-white">
    <!-- ── Sidebar ─────────────────────────────────────────────────────── -->
    <aside
      :class="collapsed ? 'w-14' : 'w-56'"
      class="hidden lg:flex flex-shrink-0 flex-col bg-slate-900 text-slate-100 overflow-y-auto transition-[width] duration-200 ease-in-out"
    >
      <!-- App identity + collapse toggle -->
      <!-- Expanded: single row with logo, title, chevron -->
      <div
        v-if="!collapsed"
        class="flex items-center gap-2 px-3 py-3 border-b border-slate-800"
      >
        <img src="@/assets/logo.svg" alt="Hivetrack" class="size-6 flex-shrink-0" />
        <span class="font-semibold tracking-tight text-white text-sm flex-1">Hivetrack</span>
        <button
          class="flex items-center justify-center size-6 rounded text-slate-500 hover:text-slate-200 hover:bg-slate-800 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100 cursor-pointer flex-shrink-0"
          title="Collapse sidebar"
          @click="collapsed = !collapsed"
        >
          <ChevronLeftIcon class="size-4" />
        </button>
      </div>
      <!-- Collapsed: logo above chevron, centered -->
      <div
        v-else
        class="flex flex-col items-center gap-1 px-2 py-3 border-b border-slate-800"
      >
        <img src="@/assets/logo.svg" alt="Hivetrack" class="size-6" />
        <button
          class="flex items-center justify-center size-6 rounded text-slate-500 hover:text-slate-200 hover:bg-slate-800 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100 cursor-pointer"
          title="Expand sidebar"
          @click="collapsed = !collapsed"
        >
          <ChevronRightIcon class="size-4" />
        </button>
      </div>

      <!-- Global navigation -->
      <nav class="flex-1 px-2 py-2 space-y-0.5">
        <RouterLink
          to="/"
          :class="[
            collapsed ? 'justify-center px-0' : 'gap-2.5 px-2',
            route.path === '/' ? 'bg-slate-800 text-slate-100' : '',
          ]"
          class="flex items-center w-full rounded-md py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
          exact-active-class="bg-slate-800 text-slate-100"
          title="My Work"
        >
          <LayoutDashboardIcon class="size-4 flex-shrink-0" />
          <span v-if="!collapsed">My Work</span>
        </RouterLink>

        <RouterLink
          to="/projects"
          :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
          class="flex items-center w-full rounded-md py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
          active-class="bg-slate-800 text-slate-100"
          title="Projects"
        >
          <FolderKanbanIcon class="size-4 flex-shrink-0" />
          <span v-if="!collapsed">Projects</span>
        </RouterLink>

        <!-- Project-level navigation — only shown when inside a project -->
        <template v-if="projectSlug">
          <div v-if="!collapsed" class="pt-3 pb-1 px-2 flex items-center gap-1">
            <RouterLink
              to="/projects"
              class="text-[11px] font-medium uppercase tracking-wider text-slate-500 hover:text-slate-300 transition-colors"
            >
              Projects
            </RouterLink>
            <ChevronRightIcon class="size-3 text-slate-600" />
            <span
              class="text-[11px] font-medium uppercase tracking-wider text-slate-400 truncate max-w-24"
            >
              {{ projectSlug }}
            </span>
          </div>
          <div v-else class="pt-2 pb-1 flex justify-center">
            <div class="w-5 border-t border-slate-700" />
          </div>

          <RouterLink
            :to="`/projects/${projectSlug}/overview`"
            :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
            class="flex items-center w-full rounded-md py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-slate-800 text-slate-100"
            title="Overview"
          >
            <LayoutIcon class="size-4 flex-shrink-0" />
            <span v-if="!collapsed">Overview</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/board`"
            :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
            class="flex items-center w-full rounded-md py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-slate-800 text-slate-100"
            title="Board"
          >
            <KanbanIcon class="size-4 flex-shrink-0" />
            <span v-if="!collapsed">Board</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/backlog`"
            :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
            class="flex items-center w-full rounded-md py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-slate-800 text-slate-100"
            title="Backlog"
          >
            <ListIcon class="size-4 flex-shrink-0" />
            <span v-if="!collapsed">Backlog</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/sprints`"
            :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
            class="flex items-center w-full rounded-md py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-slate-800 text-slate-100"
            title="Sprints"
          >
            <TimerIcon class="size-4 flex-shrink-0" />
            <span v-if="!collapsed">Sprints</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/epics`"
            :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
            class="flex items-center w-full rounded-md py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-slate-800 text-slate-100"
            title="Epics"
          >
            <LayersIcon class="size-4 flex-shrink-0" />
            <span v-if="!collapsed">Epics</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/triage`"
            :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
            class="flex items-center w-full rounded-md py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-slate-800 text-slate-100"
            title="Triage"
          >
            <InboxIcon class="size-4 flex-shrink-0" />
            <span v-if="!collapsed">Triage</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/milestones`"
            :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
            class="flex items-center w-full rounded-md py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-slate-800 text-slate-100"
            title="Milestones"
          >
            <FlagIcon class="size-4 flex-shrink-0" />
            <span v-if="!collapsed">Milestones</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/settings`"
            :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
            class="flex items-center w-full rounded-md py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-slate-800 text-slate-100"
            title="Settings"
          >
            <SettingsIcon class="size-4 flex-shrink-0" />
            <span v-if="!collapsed">Settings</span>
          </RouterLink>
        </template>
      </nav>

      <!-- Search hint + bottom section -->
      <div class="px-2 py-2 border-t border-slate-800 space-y-0.5">
        <button
          :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
          class="flex items-center w-full rounded-md py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100 text-left cursor-pointer"
          title="Press / to search"
          @click="
            () => {
              /* TODO: open search */
            }
          "
        >
          <SearchIcon class="size-4 flex-shrink-0" />
          <span v-if="!collapsed" class="flex-1">Search</span>
          <kbd
            v-if="!collapsed"
            class="text-[10px] font-mono text-slate-600 bg-slate-800 px-1 rounded"
            >/</kbd
          >
        </button>

        <RouterLink
          to="/settings"
          :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
          class="flex items-center w-full rounded-md py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
          active-class="bg-slate-800 text-slate-100"
          title="Instance settings"
        >
          <SettingsIcon class="size-4 flex-shrink-0" />
          <span v-if="!collapsed">Instance settings</span>
        </RouterLink>
      </div>

      <!-- User profile -->
      <div
        :class="collapsed ? 'justify-center px-2' : 'px-3 gap-2.5'"
        class="py-2.5 border-t border-slate-800 flex items-center min-w-0"
      >
        <Avatar :name="userName" size="sm" :src="user?.profile?.picture" :title="collapsed ? userName : undefined" />
        <template v-if="!collapsed">
          <span class="text-sm text-slate-300 truncate flex-1 min-w-0">{{ userName }}</span>
          <button
            class="flex-shrink-0 text-slate-500 hover:text-slate-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 rounded transition-colors duration-100 cursor-pointer"
            title="Sign out"
            @click="signOut"
          >
            <LogOutIcon class="size-4" />
          </button>
        </template>
      </div>
    </aside>

    <!-- ── Main content ────────────────────────────────────────────────── -->
    <main class="flex-1 overflow-y-auto min-w-0">
      <slot />
    </main>
  </div>
</template>
