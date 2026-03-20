<!--
  MinimalLayout — centred, distraction-free layout.

  Use for: onboarding, project creation, settings pages, empty states
  where the sidebar chrome would be distracting.

  Renders a simple top bar with the app name and a centred content slot.
-->
<script setup>
import { RouterLink } from 'vue-router'
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-slate-950 flex flex-col">
    <!-- Top bar -->
    <header class="h-12 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 flex items-center px-6">
      <RouterLink to="/" class="text-sm font-semibold text-slate-800 dark:text-slate-200 tracking-tight">
        Hivetrack
      </RouterLink>
    </header>

    <!-- Centred content -->
    <main class="flex-1 flex items-start justify-center px-4 py-12">
      <div class="w-full max-w-lg">
        <slot />
      </div>
    </main>
  </div>
</template>
