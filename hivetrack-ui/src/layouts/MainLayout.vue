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
  MenuIcon,
  XIcon,
} from "lucide-vue-next";
import Avatar from "@/components/ui/Avatar.vue";
import SearchDialog from "@/components/search/SearchDialog.vue";
import { useAuth } from "@/composables/useAuth";
import { useQuery } from "@tanstack/vue-query";
import { fetchIssues } from "@/api/issues";

const route = useRoute();
const { user, signOut } = useAuth();

// True when navigated inside a project (route has a :slug param).
const projectSlug = computed(() => route.params.slug ?? null);

// Untriaged count for the triage nav badge.
const { data: triageCountResult } = useQuery({
  queryKey: computed(() => [
    "issues",
    projectSlug.value,
    { triaged: false, limit: 1 },
  ]),
  queryFn: () => fetchIssues(projectSlug.value, { triaged: false, limit: 1 }),
  enabled: computed(() => !!projectSlug.value),
});
const triageCount = computed(() => triageCountResult.value?.total ?? 0);

// Display name for the current user. Falls back to email, then "You".
const userName = computed(
  () => user.value?.profile?.name ?? user.value?.profile?.email ?? "You",
);

// ── Platform detection ────────────────────────────────────────────────────────
const isMac = navigator.platform.toUpperCase().indexOf("MAC") >= 0;

// ── Sidebar collapsed state ────────────────────────────────────────────────────

const collapsed = ref(
  localStorage.getItem("hivetrack:sidebar:collapsed") === "true",
);
watch(collapsed, (v) =>
  localStorage.setItem("hivetrack:sidebar:collapsed", String(v)),
);

// ── Search dialog ────────────────────────────────────────────────────────────
const searchOpen = ref(false);

// ── Mobile sidebar drawer ─────────────────────────────────────────────────────
const mobileOpen = ref(false);

// Close mobile drawer on route change.
watch(
  () => route.fullPath,
  () => {
    mobileOpen.value = false;
  },
);

// ── Keyboard shortcuts ────────────────────────────────────────────────────────

const emit = defineEmits(["create-issue"]);

function handleKeydown(e) {
  // Ctrl+K / Cmd+K always opens search, even from inputs.
  if ((e.metaKey || e.ctrlKey) && e.key === "k") {
    e.preventDefault();
    searchOpen.value = true;
    return;
  }

  // Remaining shortcuts are ignored when focus is inside an input/textarea.
  const tag = document.activeElement?.tagName;
  if (
    tag === "INPUT" ||
    tag === "TEXTAREA" ||
    document.activeElement?.isContentEditable
  )
    return;

  if (e.key === "/" && !e.metaKey && !e.ctrlKey) {
    e.preventDefault();
    searchOpen.value = true;
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
  <div class="flex h-screen overflow-hidden bg-white dark:bg-slate-950">
    <!-- ── Mobile sidebar backdrop ─────────────────────────────────────── -->
    <Transition
      enter-active-class="transition-opacity duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="mobileOpen"
        class="fixed inset-0 z-40 bg-black/40 lg:hidden"
        @click="mobileOpen = false"
      />
    </Transition>

    <!-- ── Mobile sidebar drawer ───────────────────────────────────────── -->
    <Transition
      enter-active-class="transition-transform duration-200 ease-out"
      enter-from-class="-translate-x-full"
      enter-to-class="translate-x-0"
      leave-active-class="transition-transform duration-150 ease-in"
      leave-from-class="translate-x-0"
      leave-to-class="-translate-x-full"
    >
      <aside
        v-if="mobileOpen"
        class="fixed inset-y-0 left-0 z-50 w-64 flex flex-col bg-gray-50 dark:bg-slate-900 text-gray-700 dark:text-gray-300 border-r border-gray-200 dark:border-gray-800 overflow-y-auto lg:hidden"
      >
        <!-- Header with close button -->
        <div
          class="flex items-center gap-2 px-3 py-4 border-b border-gray-200 dark:border-gray-800"
        >
          <img
            src="@/assets/logo.svg"
            alt="Hivetrack"
            class="size-6 flex-shrink-0"
          />
          <span
            class="font-semibold tracking-tight text-gray-900 dark:text-gray-100 text-sm flex-1"
            >Hivetrack</span
          >
          <button
            class="flex items-center justify-center size-6 rounded text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100 cursor-pointer flex-shrink-0"
            title="Close menu"
            @click="mobileOpen = false"
          >
            <XIcon class="size-4" />
          </button>
        </div>

        <!-- Navigation (always expanded) -->
        <nav class="flex-1 px-2 py-2 space-y-0.5">
          <RouterLink
            to="/"
            :class="
              route.path === '/'
                ? 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100'
                : ''
            "
            class="flex items-center gap-2.5 px-2 w-full rounded-md py-1.5 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            exact-active-class="bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100"
          >
            <LayoutDashboardIcon class="size-4 flex-shrink-0" />
            <span>My Work</span>
          </RouterLink>

          <RouterLink
            to="/projects"
            class="flex items-center gap-2.5 px-2 w-full rounded-md py-1.5 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100"
          >
            <FolderKanbanIcon class="size-4 flex-shrink-0" />
            <span>Projects</span>
          </RouterLink>

          <template v-if="projectSlug">
            <div class="pt-3 pb-1 px-2 flex items-center gap-1">
              <RouterLink
                to="/projects"
                class="text-[11px] font-medium uppercase tracking-wider text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
              >
                Projects
              </RouterLink>
              <ChevronRightIcon
                class="size-3 text-gray-400 dark:text-gray-600"
              />
              <span
                class="text-[11px] font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400 truncate max-w-24"
              >
                {{ projectSlug }}
              </span>
            </div>

            <RouterLink
              v-for="item in [
                {
                  to: `/projects/${projectSlug}/overview`,
                  icon: LayoutIcon,
                  label: 'Overview',
                },
                {
                  to: `/projects/${projectSlug}/board`,
                  icon: KanbanIcon,
                  label: 'Board',
                },
                {
                  to: `/projects/${projectSlug}/backlog`,
                  icon: ListIcon,
                  label: 'Backlog',
                },
                {
                  to: `/projects/${projectSlug}/sprints`,
                  icon: TimerIcon,
                  label: 'Sprints',
                },
                {
                  to: `/projects/${projectSlug}/epics`,
                  icon: LayersIcon,
                  label: 'Epics',
                },
                {
                  to: `/projects/${projectSlug}/triage`,
                  icon: InboxIcon,
                  label: 'Triage',
                },
                {
                  to: `/projects/${projectSlug}/milestones`,
                  icon: FlagIcon,
                  label: 'Milestones',
                },
              ]"
              :key="item.to"
              :to="item.to"
              class="flex items-center gap-2.5 px-2 w-full rounded-md py-1.5 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
              active-class="bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            >
              <component :is="item.icon" class="size-4 flex-shrink-0" />
              <span class="flex-1">{{ item.label }}</span>
              <span
                v-if="item.label === 'Triage' && triageCount > 0"
                class="flex-shrink-0 min-w-[18px] h-[18px] flex items-center justify-center rounded-full text-[10px] font-semibold tabular-nums bg-amber-500 text-white px-1"
              >
                {{ triageCount > 99 ? "99+" : triageCount }}
              </span>
            </RouterLink>
          </template>
        </nav>

        <!-- Bottom section -->
        <div
          class="px-2 py-2 border-t border-gray-200 dark:border-gray-800 space-y-0.5"
        >
          <RouterLink
            to="/settings"
            class="flex items-center gap-2.5 px-2 w-full rounded-md py-1.5 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100"
          >
            <SettingsIcon class="size-4 flex-shrink-0" />
            <span>User settings</span>
          </RouterLink>
        </div>

        <!-- User profile -->
        <div
          class="px-3 py-2.5 border-t border-gray-200 dark:border-gray-800 flex items-center gap-2.5 min-w-0"
        >
          <Avatar :name="userName" size="sm" :src="user?.profile?.picture" />
          <span
            class="text-sm text-gray-700 dark:text-gray-300 truncate flex-1 min-w-0"
            >{{ userName }}</span
          >
          <button
            class="flex-shrink-0 text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 rounded transition-colors duration-100 cursor-pointer"
            title="Sign out"
            @click="signOut"
          >
            <LogOutIcon class="size-4" />
          </button>
        </div>
      </aside>
    </Transition>

    <!-- ── Desktop sidebar ─────────────────────────────────────────────── -->
    <aside
      :class="collapsed ? 'w-14' : 'w-56'"
      class="hidden lg:flex flex-shrink-0 flex-col bg-gray-50 dark:bg-slate-900 text-gray-700 dark:text-gray-300 border-r border-gray-200 dark:border-gray-800 overflow-y-auto transition-[width] duration-200 ease-in-out"
    >
      <!-- App identity + collapse toggle -->
      <!-- Expanded: single row with logo, title, chevron -->
      <div
        v-if="!collapsed"
        class="flex items-center gap-2 px-3 py-4 border-b border-gray-200 dark:border-gray-800"
      >
        <img
          src="@/assets/logo.svg"
          alt="Hivetrack"
          class="size-6 flex-shrink-0"
        />
        <span
          class="font-semibold tracking-tight text-gray-900 dark:text-gray-100 text-sm flex-1"
          >Hivetrack</span
        >
        <button
          class="flex items-center justify-center size-6 rounded text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100 cursor-pointer flex-shrink-0"
          title="Collapse sidebar"
          @click="collapsed = !collapsed"
        >
          <ChevronLeftIcon class="size-4" />
        </button>
      </div>
      <!-- Collapsed: just the expand chevron, same height as expanded header -->
      <div
        v-else
        class="flex items-center justify-center px-2 py-4 border-b border-gray-200 dark:border-gray-800"
      >
        <button
          class="flex items-center justify-center size-6 rounded text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100 cursor-pointer"
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
            route.path === '/'
              ? 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100'
              : '',
          ]"
          class="flex items-center w-full rounded-md py-1.5 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
          exact-active-class="bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100"
          title="My Work"
        >
          <LayoutDashboardIcon class="size-4 flex-shrink-0" />
          <span v-if="!collapsed">My Work</span>
        </RouterLink>

        <RouterLink
          to="/projects"
          :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
          class="flex items-center w-full rounded-md py-1.5 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
          active-class="bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100"
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
              class="text-[11px] font-medium uppercase tracking-wider text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
            >
              Projects
            </RouterLink>
            <ChevronRightIcon class="size-3 text-gray-400 dark:text-gray-600" />
            <span
              class="text-[11px] font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400 truncate max-w-24"
            >
              {{ projectSlug }}
            </span>
          </div>
          <div v-else class="pt-2 pb-1 flex justify-center">
            <div class="w-5 border-t border-gray-300 dark:border-gray-700" />
          </div>

          <RouterLink
            :to="`/projects/${projectSlug}/overview`"
            :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
            class="flex items-center w-full rounded-md py-1.5 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            title="Overview"
          >
            <LayoutIcon class="size-4 flex-shrink-0" />
            <span v-if="!collapsed">Overview</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/board`"
            :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
            class="flex items-center w-full rounded-md py-1.5 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            title="Board"
          >
            <KanbanIcon class="size-4 flex-shrink-0" />
            <span v-if="!collapsed">Board</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/backlog`"
            :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
            class="flex items-center w-full rounded-md py-1.5 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            title="Backlog"
          >
            <ListIcon class="size-4 flex-shrink-0" />
            <span v-if="!collapsed">Backlog</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/sprints`"
            :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
            class="flex items-center w-full rounded-md py-1.5 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            title="Sprints"
          >
            <TimerIcon class="size-4 flex-shrink-0" />
            <span v-if="!collapsed">Sprints</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/epics`"
            :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
            class="flex items-center w-full rounded-md py-1.5 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            title="Epics"
          >
            <LayersIcon class="size-4 flex-shrink-0" />
            <span v-if="!collapsed">Epics</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/triage`"
            :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
            class="flex items-center w-full rounded-md py-1.5 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            title="Triage"
          >
            <InboxIcon class="size-4 flex-shrink-0" />
            <span v-if="!collapsed" class="flex-1">Triage</span>
            <span
              v-if="triageCount > 0 && !collapsed"
              class="flex-shrink-0 min-w-[18px] h-[18px] flex items-center justify-center rounded-full text-[10px] font-semibold tabular-nums bg-amber-500 text-white px-1"
              :title="`${triageCount} untriaged`"
            >
              {{ triageCount > 99 ? "99+" : triageCount }}
            </span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/milestones`"
            :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
            class="flex items-center w-full rounded-md py-1.5 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            title="Milestones"
          >
            <FlagIcon class="size-4 flex-shrink-0" />
            <span v-if="!collapsed">Milestones</span>
          </RouterLink>
        </template>
      </nav>

      <!-- Search hint + bottom section -->
      <div
        class="px-2 py-2 border-t border-gray-200 dark:border-gray-800 space-y-0.5"
      >
        <button
          :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
          class="flex items-center w-full rounded-md py-1.5 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100 text-left cursor-pointer"
          :title="isMac ? '⌘K to search' : 'Ctrl+K to search'"
          @click="searchOpen = true"
        >
          <SearchIcon class="size-4 flex-shrink-0" />
          <span v-if="!collapsed" class="flex-1">Search</span>
          <kbd
            v-if="!collapsed"
            class="text-[10px] font-mono text-gray-400 dark:text-gray-500 bg-gray-200 dark:bg-gray-700 px-1 rounded"
            >{{ isMac ? "⌘K" : "Ctrl+K" }}</kbd
          >
        </button>

        <RouterLink
          to="/settings"
          :class="collapsed ? 'justify-center px-0' : 'gap-2.5 px-2'"
          class="flex items-center w-full rounded-md py-1.5 text-sm text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-gray-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
          active-class="bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-gray-100"
          title="User settings"
        >
          <SettingsIcon class="size-4 flex-shrink-0" />
          <span v-if="!collapsed">User settings</span>
        </RouterLink>
      </div>

      <!-- User profile -->
      <div
        :class="collapsed ? 'justify-center px-2' : 'px-3 gap-2.5'"
        class="py-2.5 border-t border-gray-200 dark:border-gray-800 flex items-center min-w-0"
      >
        <Avatar
          :name="userName"
          size="sm"
          :src="user?.profile?.picture"
          :title="collapsed ? userName : undefined"
        />
        <template v-if="!collapsed">
          <span
            class="text-sm text-gray-700 dark:text-gray-300 truncate flex-1 min-w-0"
            >{{ userName }}</span
          >
          <button
            class="flex-shrink-0 text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 rounded transition-colors duration-100 cursor-pointer"
            title="Sign out"
            @click="signOut"
          >
            <LogOutIcon class="size-4" />
          </button>
        </template>
      </div>
    </aside>

    <!-- ── Main content ────────────────────────────────────────────────── -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- Mobile top bar with hamburger -->
      <div
        class="flex items-center gap-2 px-3 py-2 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-slate-900 lg:hidden"
      >
        <button
          class="flex items-center justify-center size-8 rounded text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100 cursor-pointer"
          title="Open menu"
          @click="mobileOpen = true"
        >
          <MenuIcon class="size-5" />
        </button>
        <img src="@/assets/logo.svg" alt="Hivetrack" class="size-5" />
        <span
          class="font-semibold tracking-tight text-gray-900 dark:text-gray-100 text-sm"
          >Hivetrack</span
        >
      </div>

      <main class="flex-1 overflow-y-auto min-w-0 bg-white dark:bg-slate-950">
        <slot />
      </main>
    </div>

    <SearchDialog
      :open="searchOpen"
      :project-slug="projectSlug"
      @close="searchOpen = false"
    />
  </div>
</template>
