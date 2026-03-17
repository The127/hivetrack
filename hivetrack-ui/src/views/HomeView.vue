<script setup>
import { useQuery } from "@tanstack/vue-query";

const { data: health, isLoading } = useQuery({
  queryKey: ["health"],
  queryFn: () => fetch("/api/v1/health").then((r) => r.json()),
});
</script>

<template>
  <div
    class="flex min-h-screen items-center justify-center bg-gray-950 text-white"
  >
    <div class="text-center">
      <h1 class="text-4xl font-bold tracking-tight">Hivetrack</h1>
      <p class="mt-2 text-gray-400">
        Lean task planning for high-performing teams.
      </p>
      <p v-if="isLoading" class="mt-6 text-sm text-gray-500">connecting...</p>
      <p v-else-if="health" class="mt-6 text-sm text-emerald-400">
        ● backend {{ health.status }}
      </p>
      <p v-else class="mt-6 text-sm text-red-400">✕ backend unreachable</p>
    </div>
  </div>
</template>
