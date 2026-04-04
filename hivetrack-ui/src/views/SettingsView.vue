<script setup>
import { MonitorIcon, SunIcon, MoonIcon } from 'lucide-vue-next'
import MainLayout from '@/layouts/MainLayout.vue'
import { useTheme } from '@/composables/useTheme'
import { useAccessibility } from '@/composables/useAccessibility'

const { theme, setTheme } = useTheme()
const { colorVision, highContrast, font, setColorVision, setHighContrast, setFont } = useAccessibility()

const THEME_OPTIONS = [
  { value: 'system', label: 'System', icon: MonitorIcon },
  { value: 'light',  label: 'Light',  icon: SunIcon },
  { value: 'dark',   label: 'Dark',   icon: MoonIcon },
]

const COLOR_VISION_OPTIONS = [
  { value: 'normal',       label: 'Normal vision' },
  { value: 'protanopia',   label: 'Protanopia (red-blind)' },
  { value: 'deuteranopia', label: 'Deuteranopia (green-blind)' },
  { value: 'tritanopia',   label: 'Tritanopia (blue-blind)' },
]

const FONT_OPTIONS = [
  { value: 'default',  label: 'App default' },
  { value: 'system',   label: 'System font' },
  { value: 'dyslexia', label: 'Dyslexia-friendly (OpenDyslexic)' },
]
</script>

<template>
  <MainLayout>
    <div class="max-w-2xl mx-auto px-6 py-8 space-y-6">
      <h1 class="text-xl font-semibold text-slate-900 dark:text-slate-100">
        User Settings
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

      <!-- Accessibility -->
      <section class="rounded-lg border border-slate-200 dark:border-slate-700 overflow-hidden">
        <div class="px-5 py-4 border-b border-slate-100 dark:border-slate-800">
          <h2 class="text-sm font-semibold text-slate-900 dark:text-slate-100">Accessibility</h2>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Adjust color and contrast settings for better readability.
          </p>
        </div>

        <div class="px-5 py-4 bg-white dark:bg-slate-900 space-y-5">
          <!-- Color vision -->
          <div>
            <label
              for="color-vision"
              class="block text-sm font-medium text-slate-900 dark:text-slate-100 mb-1.5"
            >
              Color vision
            </label>
            <select
              id="color-vision"
              :value="colorVision"
              class="w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 px-3 py-2 text-sm text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              @change="setColorVision($event.target.value)"
            >
              <option
                v-for="opt in COLOR_VISION_OPTIONS"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.label }}
              </option>
            </select>
          </div>

          <!-- High contrast -->
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-slate-900 dark:text-slate-100">High contrast</p>
              <p class="text-xs text-slate-500 dark:text-slate-400">
                Increase border and text contrast
              </p>
            </div>
            <button
              type="button"
              role="switch"
              :aria-checked="highContrast"
              class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 dark:focus-visible:ring-offset-slate-900"
              :class="highContrast ? 'bg-blue-600' : 'bg-slate-200 dark:bg-slate-700'"
              @click="setHighContrast(!highContrast)"
            >
              <span
                aria-hidden="true"
                class="pointer-events-none inline-block size-5 transform rounded-full bg-white shadow ring-0 transition duration-200"
                :class="highContrast ? 'translate-x-5' : 'translate-x-0'"
              />
            </button>
          </div>
        </div>
      </section>

      <!-- Typography -->
      <section class="rounded-lg border border-slate-200 dark:border-slate-700 overflow-hidden">
        <div class="px-5 py-4 border-b border-slate-100 dark:border-slate-800">
          <h2 class="text-sm font-semibold text-slate-900 dark:text-slate-100">Typography</h2>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Choose the font used throughout the application.
          </p>
        </div>

        <div class="px-5 py-4 bg-white dark:bg-slate-900">
          <label class="block text-sm font-medium text-slate-900 dark:text-slate-100 mb-2">
            Font
          </label>
          <div class="space-y-1.5">
            <button
              v-for="opt in FONT_OPTIONS"
              :key="opt.value"
              class="w-full text-left rounded-lg border px-4 py-2.5 text-sm transition-colors cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500"
              :class="font === opt.value
                ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 ring-1 ring-blue-500'
                : 'border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 hover:border-slate-300 dark:hover:border-slate-600 hover:bg-slate-50 dark:hover:bg-slate-800'"
              @click="setFont(opt.value)"
            >
              {{ opt.label }}
            </button>
          </div>
        </div>
      </section>
    </div>
  </MainLayout>
</template>
