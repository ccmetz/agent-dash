<script setup lang="ts">
import { onMounted, ref } from 'vue';
import PanelCard from './ui/PanelCard.vue';
import StatusItem from './ui/StatusItem.vue';

type Status = {
  ok: boolean;
  analyticsStorePath: string;
  usageSource?: {
    name: string;
    path: string;
    available: boolean;
    state: string;
  };
};

const status = ref<Status | null>(null);
const error = ref(false);

const waitingMessage = 'Waiting for backend status...';

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

function backendStatus() {
  if (status.value?.ok) return 'connected';
  if (error.value) return 'disconnected';
  return 'connecting';
}

function usageSourceStatus() {
  if (status.value?.usageSource?.available) return 'connected';
  if (status.value?.usageSource?.state === 'missing' || error.value) return 'disconnected';
  return 'connecting';
}
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

    <PanelCard
      v-if="status?.usageSource?.state === 'missing'"
      class="mt-10 space-y-4"
      aria-label="Usage Source setup"
    >
      <p class="text-app-accent text-xs font-bold tracking-[0.12em] uppercase">Setup needed</p>
      <h2 class="text-app-fg-strong m-0 text-2xl font-bold">OpenCode Usage Source missing</h2>
      <p class="text-app-muted m-0 max-w-xl text-lg leading-7">
        Agent Dash could not find OpenCode's local database, so there is no Usage Metadata to show
        yet.
      </p>
      <p class="text-app-fg-strong m-0 font-bold">
        Open OpenCode once, then refresh this dashboard.
      </p>
      <p class="text-app-muted m-0">Checked path:</p>
      <p class="text-app-fg-strong m-0 wrap-anywhere font-mono">{{ status.usageSource.path }}</p>
    </PanelCard>

    <PanelCard class="mt-14 divide-y divide-panel-border" aria-label="Backend status">
      <StatusItem
        class="pb-6"
        label="Analytics Store"
        :status="backendStatus()"
        :value="status?.analyticsStorePath ?? waitingMessage"
      />
      <StatusItem
        class="pt-6"
        label="OpenCode Usage Source"
        :status="usageSourceStatus()"
        :value="status?.usageSource?.path ?? waitingMessage"
      />
    </PanelCard>
  </main>
</template>
