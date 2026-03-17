<!--
  PublicLayout — unauthenticated layout for the customer support portal.

  No auth required. Used for:
  - /public/support/:slug — issue submission form
  - /support/track/:token — issue status tracking

  Shows a minimal branded header without any internal navigation.
-->
<script setup>
defineProps({
  /** Project name to show in the header alongside the Hivetrack brand. */
  projectName: {
    type: String,
    default: null,
  },
})
</script>

<template>
  <div class="min-h-screen bg-slate-50 flex flex-col">
    <!-- Minimal public header -->
    <header class="h-12 border-b border-slate-200 bg-white flex items-center px-6 gap-3">
      <span class="text-sm font-semibold text-slate-800 tracking-tight">Hivetrack</span>
      <template v-if="projectName">
        <span class="text-slate-300">/</span>
        <span class="text-sm text-slate-600">{{ projectName }}</span>
      </template>
    </header>

    <!-- Content -->
    <main class="flex-1 flex items-start justify-center px-4 py-10">
      <div class="w-full max-w-xl">
        <slot />
      </div>
    </main>

    <footer class="py-4 text-center text-xs text-slate-400">
      Powered by Hivetrack
    </footer>
  </div>
</template>
