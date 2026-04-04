<script setup>
import ToastContainer from '@/components/ui/ToastContainer.vue'
</script>

<template>
  <RouterView />
  <ToastContainer />

  <!-- Color vision simulation filters (applied via CSS filter: url(#cv-xxx) on <html>) -->
  <svg class="absolute h-0 w-0" aria-hidden="true">
    <defs>
      <filter id="cv-protanopia">
        <feColorMatrix type="matrix" values="
          0.567, 0.433, 0,     0, 0
          0.558, 0.442, 0,     0, 0
          0,     0.242, 0.758, 0, 0
          0,     0,     0,     1, 0
        " />
      </filter>
      <filter id="cv-deuteranopia">
        <feColorMatrix type="matrix" values="
          0.625, 0.375, 0,   0, 0
          0.7,   0.3,   0,   0, 0
          0,     0.3,   0.7, 0, 0
          0,     0,     0,   1, 0
        " />
      </filter>
      <filter id="cv-tritanopia">
        <feColorMatrix type="matrix" values="
          0.95, 0.05,  0,     0, 0
          0,    0.433, 0.567, 0, 0
          0,    0.475, 0.525, 0, 0
          0,    0,     0,     1, 0
        " />
      </filter>
    </defs>
  </svg>
</template>
