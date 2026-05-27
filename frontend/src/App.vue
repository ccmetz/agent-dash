<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
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
  usageSync?: UsageSync;
};

type SyncRun = {
  status: string;
  startedAt: string;
  finishedAt: string;
  inserted: number;
  updated: number;
  skipped: number;
  errorMessage?: string;
};

type UsageSync = {
  status: string;
  lastRun?: SyncRun;
  recentRuns: SyncRun[];
  nextPollAt: string;
  pollSeconds: number;
};

const status = ref<Status | null>(null);
const error = ref(false);
const syncing = ref(false);
const nextPollAt = ref<string | null>(null);
let pollTimer: number | undefined;

const waitingMessage = 'Waiting for backend status...';

async function loadStatus() {
  try {
    const response = await fetch('/api/status');
    if (!response.ok) {
      throw new Error('Status request failed');
    }
    status.value = await response.json();
    nextPollAt.value = status.value?.usageSync?.nextPollAt ?? null;
    error.value = false;
  } catch {
    error.value = true;
  }
}

onMounted(async () => {
  await loadStatus();
  if (status.value?.usageSource?.available) {
    await runManualSync();
  }
  pollTimer = window.setInterval(loadStatus, 60_000);
});

onUnmounted(() => {
  if (pollTimer) window.clearInterval(pollTimer);
});

async function runManualSync() {
  syncing.value = true;
  try {
    const response = await fetch('/api/usage-sync', { method: 'POST' });
    const run = (await response.json()) as SyncRun;
    status.value = {
      ...(status.value ?? { ok: true, analyticsStorePath: '' }),
      usageSync: {
        status: run.status,
        lastRun: run,
        recentRuns: [run, ...(status.value?.usageSync?.recentRuns ?? [])],
        nextPollAt: new Date(Date.now() + 60_000).toISOString(),
        pollSeconds: 60,
      },
    };
    await loadStatus();
  } finally {
    syncing.value = false;
  }
}

const lastRun = computed(() => status.value?.usageSync?.lastRun);
const latestError = computed(() => lastRun.value?.errorMessage);

function syncStatus() {
  if (syncing.value) return 'connecting';
  if (error.value) return 'disconnected';
  if (status.value?.usageSync?.status === 'error') return 'disconnected';
  if (status.value?.usageSync?.lastRun) return 'connected';
  return 'connecting';
}

function formatSyncCounts(run?: SyncRun) {
  if (!run) return 'No sync runs yet';
  return `${run.inserted} inserted, ${run.updated} updated, ${run.skipped} skipped`;
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

    <PanelCard class="mt-14 space-y-5" aria-label="Usage Sync status">
      <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <StatusItem
          label="Usage Sync"
          :status="syncStatus()"
          :value="syncing ? 'Sync in progress...' : formatSyncCounts(lastRun)"
        />
        <button
          class="border-app-accent/70 bg-app-shell/40 text-app-fg-strong hover:border-app-fg-strong hover:bg-app-accent/15 focus-visible:ring-app-accent cursor-pointer rounded-lg border px-4 py-2 text-sm font-bold shadow-sm shadow-black/20 transition hover:-translate-y-0.5 hover:shadow-md focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-app-shell focus-visible:outline-none active:translate-y-0 disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0 disabled:hover:border-app-accent/70 disabled:hover:bg-app-shell/40 disabled:hover:shadow-sm"
          type="button"
          :disabled="syncing"
          @click="runManualSync"
        >
          {{ syncing ? 'Syncing...' : 'Refresh Usage' }}
        </button>
      </div>
      <dl class="text-app-muted grid gap-3 text-sm sm:grid-cols-2">
        <div>
          <dt class="font-bold">Last sync</dt>
          <dd class="m-0">{{ lastRun?.finishedAt ?? 'Never synced' }}</dd>
        </div>
        <div>
          <dt class="font-bold">Next poll</dt>
          <dd class="m-0">{{ nextPollAt ?? 'Waiting for backend status...' }}</dd>
        </div>
        <div>
          <dt class="font-bold">Source path</dt>
          <dd class="m-0 wrap-anywhere font-mono">
            {{ status?.usageSource?.path ?? waitingMessage }}
          </dd>
        </div>
        <div>
          <dt class="font-bold">Analytics Store</dt>
          <dd class="m-0 wrap-anywhere font-mono">
            {{ status?.analyticsStorePath ?? waitingMessage }}
          </dd>
        </div>
        <div v-if="latestError">
          <dt class="font-bold">Latest error</dt>
          <dd class="m-0">{{ latestError }}</dd>
        </div>
      </dl>
    </PanelCard>
  </main>
</template>
