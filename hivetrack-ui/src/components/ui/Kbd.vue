<!--
  Kbd — keyboard shortcut display.

  Renders one or more keys in a monospaced, button-like style.
  Use the `keys` prop for modifier combos, or the default slot for
  a single key label.

  @example
  <Kbd>C</Kbd>
  <Kbd :keys="['⌘', 'K']" />
  <Kbd :keys="['Ctrl', 'K']" />
-->
<script setup>
defineProps({
  /**
   * Ordered list of key labels to render as a chord (joined with +).
   * If omitted, use the default slot for a single key.
   */
  keys: {
    type: Array,
    default: null,
  },
})
</script>

<template>
  <span class="inline-flex items-center gap-0.5">
    <template v-if="keys">
      <kbd
        v-for="key in keys"
        :key="key"
        class="inline-flex items-center rounded border border-slate-300 bg-slate-50 px-1 py-0.5 font-mono text-[11px] text-slate-600 shadow-[0_1px_0_rgba(0,0,0,0.15)] leading-none"
      >
        {{ key }}
      </kbd>
      <!-- The "+" separators between keys -->
      <!-- intentionally nothing — keys are spaced by gap -->
    </template>
    <kbd
      v-else
      class="inline-flex items-center rounded border border-slate-300 bg-slate-50 px-1 py-0.5 font-mono text-[11px] text-slate-600 shadow-[0_1px_0_rgba(0,0,0,0.15)] leading-none"
    >
      <slot />
    </kbd>
  </span>
</template>
