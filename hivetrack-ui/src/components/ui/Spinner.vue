<!--
  Spinner — animated loading indicator.

  Inherits colour from the parent via `currentColor`.
  Use `class` to control size (e.g. `class="size-4"`) and colour
  (e.g. `class="text-blue-600"`).

  @example
  <Spinner class="size-4 text-blue-600" />
  <Spinner class="size-5" />
-->
<template>
  <svg
    class="animate-spin"
    xmlns="http://www.w3.org/2000/svg"
    fill="none"
    viewBox="0 0 24 24"
    aria-hidden="true"
    v-bind="$attrs"
  >
    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
    <path
      class="opacity-75"
      fill="currentColor"
      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
    />
  </svg>
</template>
