<script setup lang="ts">
import StatusBadge from './StatusBadge.vue';

defineProps<{
  label: string;
  value: string;
  status: 'connected' | 'connecting' | 'disconnected';
}>();
</script>

<template>
  <div class="space-y-4">
    <StatusBadge :status="status" />
    <p class="text-app-accent text-xs font-bold tracking-[0.12em] uppercase">{{ label }}</p>
    <p class="text-app-fg-strong m-0 wrap-anywhere font-mono">{{ value }}</p>
  </div>
</template>
