<script setup lang="ts">
defineProps<{
  status: 'connected' | 'connecting' | 'disconnected';
}>();
</script>

<template>
  <span
    v-if="status === 'connected'"
    class="bg-status-success text-status-success-fg inline-flex rounded-full px-3 py-1.5 font-bold"
  >
    Connected
  </span>
  <span
    v-else-if="status === 'disconnected'"
    class="bg-status-danger text-status-danger-fg inline-flex rounded-full px-3 py-1.5 font-bold"
  >
    Disconnected
  </span>
  <span
    v-else
    class="bg-status-neutral text-status-neutral-fg inline-flex rounded-full px-3 py-1.5 font-bold"
  >
    Connecting
  </span>
</template>
