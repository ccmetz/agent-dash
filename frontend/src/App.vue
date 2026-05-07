<script setup lang="ts">
import { onMounted, ref } from 'vue';

type Status = {
  ok: boolean;
  analyticsStorePath: string;
};

const status = ref<Status | null>(null);
const error = ref(false);

onMounted(async () => {
  try {
    const response = await fetch('/api/status');
    if (!response.ok) {
      throw new Error('Status request failed');
    }
    status.value = await response.json();
  } catch {
    error.value = true;
  }
});
</script>

<template>
  <main class="bg-app-shell text-app-fg min-h-screen px-5 py-10 font-sans sm:px-8 md:p-16">
    <section class="max-w-2xl space-y-5">
      <p class="text-app-accent text-xs font-bold tracking-[0.12em] uppercase">Agent Dash</p>
      <h1 class="m-0 pb-2 text-[clamp(2.5rem,9vw,6rem)] leading-[0.95]">Usage Overview</h1>
      <p class="text-app-muted max-w-xl text-xl leading-8">
        Local Usage Metadata will appear here after the first Usage Sync.
      </p>
    </section>

    <section
      class="border-panel-border bg-panel shadow-panel rounded-panel mt-14 w-full max-w-lg space-y-4 border p-6 sm:p-8"
      aria-label="Backend status"
    >
      <span
        v-if="status?.ok"
        class="bg-status-success text-status-success-fg inline-flex rounded-full px-3 py-1.5 font-bold"
      >
        Connected
      </span>
      <span
        v-else-if="error"
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
      <p class="text-app-accent pt-2 text-xs font-bold tracking-[0.12em] uppercase">
        Analytics Store
      </p>
      <p class="text-app-fg-strong m-0 wrap-anywhere font-mono">
        {{ status?.analyticsStorePath ?? 'Waiting for backend status...' }}
      </p>
    </section>
  </main>
</template>
