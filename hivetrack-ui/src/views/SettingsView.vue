<!--
  SettingsView — instance-level user preferences.

  Currently exposes the theme preference (System / Light / Dark).
  More settings can be added here as needed.
-->
<script setup>
import { MonitorIcon, SunIcon, MoonIcon } from 'lucide-vue-next'
import MainLayout from '@/layouts/MainLayout.vue'
import { useTheme } from '@/composables/useTheme'

const { theme, setTheme } = useTheme()

const THEME_OPTIONS = [
  { value: 'system', label: 'System', icon: MonitorIcon },
  { value: 'light',  label: 'Light',  icon: SunIcon },
  { value: 'dark',   label: 'Dark',   icon: MoonIcon },
]
</script>

<template>
  <MainLayout>
    <div class="max-w-2xl mx-auto px-6 py-8">
      <h1 class="text-xl font-semibold text-slate-900 dark:text-slate-100 mb-6">
        Instance Settings
      </h1>

      <!-- Appearance -->
      <section class="rounded-lg border border-slate-200 dark:border-slate-700 overflow-hidden">
        <div class="px-5 py-4 border-b border-slate-100 dark:border-slate-800">
          <h2 class="text-sm font-semibold text-slate-900 dark:text-slate-100">Appearance</h2>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Choose how Hivetrack looks to you. Defaults to your OS setting.
          </p>
        </div>

        <div class="px-5 py-4 bg-white dark:bg-slate-900">
          <div class="flex gap-2">
            <button
              v-for="opt in THEME_OPTIONS"
              :key="opt.value"
              class="flex-1 flex flex-col items-center gap-2 rounded-lg border px-4 py-3 text-sm font-medium transition-colors cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
              :class="theme === opt.value
                ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 ring-1 ring-blue-500'
                : 'border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 hover:border-slate-300 dark:hover:border-slate-600 hover:bg-slate-50 dark:hover:bg-slate-800'"
              @click="setTheme(opt.value)"
            >
              <component :is="opt.icon" class="size-5" />
              {{ opt.label }}
            </button>
          </div>
        </div>
      </section>
    </div>
  </MainLayout>
</template>
