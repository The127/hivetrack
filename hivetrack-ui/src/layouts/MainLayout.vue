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
import { computed, onMounted, onBeforeUnmount } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import {
  LayoutDashboardIcon,
  FolderKanbanIcon,
  KanbanIcon,
  ListIcon,
  InboxIcon,
  CalendarIcon,
  FlagIcon,
  SettingsIcon,
  SearchIcon,
  ChevronRightIcon,
  LogOutIcon,
} from 'lucide-vue-next'
import Avatar from '@/components/ui/Avatar.vue'
import { useAuth } from '@/composables/useAuth'

const route = useRoute()
const { user, signOut } = useAuth()

// True when navigated inside a project (route has a :slug param).
const projectSlug = computed(() => route.params.slug ?? null)

// Display name for the current user. Falls back to email, then "You".
const userName = computed(
  () => user.value?.profile?.name ?? user.value?.profile?.email ?? 'You',
)

// ── Keyboard shortcuts ────────────────────────────────────────────────────────

const emit = defineEmits(['create-issue'])

function handleKeydown(e) {
  // Ignore shortcuts when focus is inside an input, textarea, or contenteditable.
  const tag = document.activeElement?.tagName
  if (tag === 'INPUT' || tag === 'TEXTAREA' || document.activeElement?.isContentEditable) return

  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    // TODO: open command palette
    console.debug('[Hivetrack] Command palette — Cmd+K (not yet implemented)')
    return
  }

  if (e.key === 'c' && !e.metaKey && !e.ctrlKey) {
    emit('create-issue')
  }
}

onMounted(() => window.addEventListener('keydown', handleKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', handleKeydown))
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-white">
    <!-- ── Sidebar ─────────────────────────────────────────────────────── -->
    <aside
      class="hidden lg:flex w-56 flex-shrink-0 flex-col bg-slate-900 text-slate-100 overflow-y-auto"
    >
      <!-- App identity -->
      <div class="flex items-center gap-2 px-4 py-3 border-b border-slate-800">
        <span class="font-semibold tracking-tight text-white text-sm">Hivetrack</span>
      </div>

      <!-- Global navigation -->
      <nav class="flex-1 px-2 py-3 space-y-0.5">
        <RouterLink
          to="/"
          class="flex items-center gap-2.5 w-full rounded-md px-2 py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
          :class="{ 'bg-slate-800 text-slate-100': route.path === '/' }"
          exact-active-class="bg-slate-800 text-slate-100"
        >
          <LayoutDashboardIcon class="size-4 flex-shrink-0" />
          <span>My Work</span>
        </RouterLink>

        <RouterLink
          to="/projects"
          class="flex items-center gap-2.5 w-full rounded-md px-2 py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
          active-class="bg-slate-800 text-slate-100"
        >
          <FolderKanbanIcon class="size-4 flex-shrink-0" />
          <span>Projects</span>
        </RouterLink>

        <!-- Project-level navigation — only shown when inside a project -->
        <template v-if="projectSlug">
          <div class="pt-3 pb-1 px-2 flex items-center gap-1">
            <RouterLink
              to="/projects"
              class="text-[11px] font-medium uppercase tracking-wider text-slate-500 hover:text-slate-300 transition-colors"
            >
              Projects
            </RouterLink>
            <ChevronRightIcon class="size-3 text-slate-600" />
            <span class="text-[11px] font-medium uppercase tracking-wider text-slate-400 truncate max-w-24">
              {{ projectSlug }}
            </span>
          </div>

          <RouterLink
            :to="`/projects/${projectSlug}/board`"
            class="flex items-center gap-2.5 w-full rounded-md px-2 py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-slate-800 text-slate-100"
          >
            <KanbanIcon class="size-4 flex-shrink-0" />
            <span>Board</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/backlog`"
            class="flex items-center gap-2.5 w-full rounded-md px-2 py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-slate-800 text-slate-100"
          >
            <ListIcon class="size-4 flex-shrink-0" />
            <span>Backlog</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/triage`"
            class="flex items-center gap-2.5 w-full rounded-md px-2 py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-slate-800 text-slate-100"
          >
            <InboxIcon class="size-4 flex-shrink-0" />
            <span>Triage</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/sprints`"
            class="flex items-center gap-2.5 w-full rounded-md px-2 py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-slate-800 text-slate-100"
          >
            <CalendarIcon class="size-4 flex-shrink-0" />
            <span>Sprints</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/milestones`"
            class="flex items-center gap-2.5 w-full rounded-md px-2 py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-slate-800 text-slate-100"
          >
            <FlagIcon class="size-4 flex-shrink-0" />
            <span>Milestones</span>
          </RouterLink>

          <RouterLink
            :to="`/projects/${projectSlug}/settings`"
            class="flex items-center gap-2.5 w-full rounded-md px-2 py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
            active-class="bg-slate-800 text-slate-100"
          >
            <SettingsIcon class="size-4 flex-shrink-0" />
            <span>Settings</span>
          </RouterLink>
        </template>
      </nav>

      <!-- Search hint + bottom section -->
      <div class="px-2 py-2 border-t border-slate-800 space-y-0.5">
        <button
          class="flex items-center gap-2.5 w-full rounded-md px-2 py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100 text-left"
          title="Press / to search"
          @click="() => {/* TODO: open search */}"
        >
          <SearchIcon class="size-4 flex-shrink-0" />
          <span class="flex-1">Search</span>
          <kbd class="text-[10px] font-mono text-slate-600 bg-slate-800 px-1 rounded">/</kbd>
        </button>

        <RouterLink
          to="/settings"
          class="flex items-center gap-2.5 w-full rounded-md px-2 py-1.5 text-sm text-slate-400 hover:bg-slate-800 hover:text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors duration-100"
          active-class="bg-slate-800 text-slate-100"
        >
          <SettingsIcon class="size-4 flex-shrink-0" />
          <span>Instance settings</span>
        </RouterLink>
      </div>

      <!-- User profile -->
      <div class="px-3 py-2.5 border-t border-slate-800 flex items-center gap-2.5 min-w-0">
        <Avatar :name="userName" size="sm" :src="user?.profile?.picture" />
        <span class="text-sm text-slate-300 truncate flex-1 min-w-0">{{ userName }}</span>
        <button
          class="flex-shrink-0 text-slate-500 hover:text-slate-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 rounded transition-colors duration-100"
          title="Sign out"
          @click="signOut"
        >
          <LogOutIcon class="size-4" />
        </button>
      </div>
    </aside>

    <!-- ── Main content ────────────────────────────────────────────────── -->
    <main class="flex-1 overflow-y-auto min-w-0">
      <slot />
    </main>
  </div>
</template>

