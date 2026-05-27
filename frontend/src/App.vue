<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import ActionButton from './ui/ActionButton.vue';
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

type UsageOverview = {
  range: {
    days: number;
    start: string;
    end: string;
  };
  totals: {
    tokens: {
      total: number;
      input: number;
      output: number;
      reasoning: number;
      cacheRead: number;
      cacheWrite: number;
    };
    actualCost: number;
    agentSessions: number;
    modelCalls: number;
  };
  daily: Array<{
    date: string;
    tokens: number;
    actualCost: number;
  }>;
};

const status = ref<Status | null>(null);
const overview = ref<UsageOverview | null>(null);
const error = ref(false);
const syncing = ref(false);
const overviewLoading = ref(false);
const overviewError = ref(false);
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

async function loadOverview() {
  overviewLoading.value = true;
  try {
    const response = await fetch('/api/usage-overview?days=30');
    if (!response.ok) {
      throw new Error('Usage Overview request failed');
    }
    const nextOverview = (await response.json()) as UsageOverview;
    if (nextOverview.totals && Array.isArray(nextOverview.daily)) {
      overview.value = nextOverview;
      overviewError.value = false;
    }
  } catch {
    overviewError.value = true;
  } finally {
    overviewLoading.value = false;
  }
}

onMounted(async () => {
  await loadStatus();
  if (status.value?.usageSource?.available) {
    await runManualSync();
    await loadOverview();
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
    await loadOverview();
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

function formatInteger(value?: number) {
  return value == null ? '0' : new Intl.NumberFormat('en-US').format(value);
}

function formatCost(value?: number) {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value ?? 0);
}

function tokenBarWidth(tokens: number) {
  const max = Math.max(...(overview.value?.daily.map((day) => day.tokens) ?? [0]), 1);
  return `${Math.max((tokens / max) * 100, 4)}%`;
}

function costBarWidth(cost: number) {
  const max = Math.max(...(overview.value?.daily.map((day) => day.actualCost) ?? [0]), 1);
  return `${Math.max((cost / max) * 100, 4)}%`;
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

    <section class="mt-12 max-w-6xl space-y-5" aria-label="Last 30 days Usage Overview">
      <div class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p class="text-app-accent text-xs font-bold tracking-[0.12em] uppercase">Last 30 days</p>
          <h2 class="text-app-fg-strong m-0 text-3xl font-bold">Usage Metadata</h2>
        </div>
        <p class="text-app-muted m-0 text-sm">
          {{
            overviewLoading
              ? 'Loading Usage Overview...'
              : overviewError
                ? 'Usage Overview unavailable'
                : 'Synced Model Calls'
          }}
        </p>
      </div>

      <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <PanelCard class="space-y-2" aria-label="Total tokens">
          <p class="text-app-accent m-0 text-xs font-bold tracking-[0.12em] uppercase">Tokens</p>
          <p class="text-app-fg-strong m-0 text-4xl font-bold">
            {{ formatInteger(overview?.totals.tokens.total) }}
          </p>
          <p class="text-app-muted m-0 text-sm">
            Input {{ formatInteger(overview?.totals.tokens.input) }} / Output
            {{ formatInteger(overview?.totals.tokens.output) }}
          </p>
        </PanelCard>
        <PanelCard class="space-y-2" aria-label="Actual Cost">
          <p class="text-app-accent m-0 text-xs font-bold tracking-[0.12em] uppercase">
            Actual Cost
          </p>
          <p class="text-app-fg-strong m-0 text-4xl font-bold">
            {{ formatCost(overview?.totals.actualCost) }}
          </p>
          <p class="text-app-muted m-0 text-sm">Source-reported spend</p>
        </PanelCard>
        <PanelCard class="space-y-2" aria-label="Agent Sessions">
          <p class="text-app-accent m-0 text-xs font-bold tracking-[0.12em] uppercase">
            Agent Sessions
          </p>
          <p class="text-app-fg-strong m-0 text-4xl font-bold">
            {{ formatInteger(overview?.totals.agentSessions) }}
          </p>
          <p class="text-app-muted m-0 text-sm">Sessions with Model Calls</p>
        </PanelCard>
        <PanelCard class="space-y-2" aria-label="Model Calls">
          <p class="text-app-accent m-0 text-xs font-bold tracking-[0.12em] uppercase">
            Model Calls
          </p>
          <p class="text-app-fg-strong m-0 text-4xl font-bold">
            {{ formatInteger(overview?.totals.modelCalls) }}
          </p>
          <p class="text-app-muted m-0 text-sm">Synced assistant calls</p>
        </PanelCard>
      </div>

      <div class="grid gap-5 lg:grid-cols-2">
        <PanelCard class="space-y-4" aria-label="Daily Tokens chart">
          <h3 class="text-app-fg-strong m-0 text-xl font-bold">Daily Tokens</h3>
          <p v-if="!overview?.daily.length" class="text-app-muted m-0">No daily token data yet.</p>
          <ol v-else class="m-0 space-y-3 p-0">
            <li
              v-for="day in overview.daily"
              :key="`tokens-${day.date}`"
              class="list-none space-y-1"
            >
              <div class="flex justify-between gap-4 text-sm">
                <span class="text-app-muted">{{ day.date }}</span>
                <span class="text-app-fg-strong"
                  >{{ day.date }}: {{ formatInteger(day.tokens) }} tokens</span
                >
              </div>
              <div class="bg-app-shell-deep h-3 overflow-hidden rounded-full">
                <div
                  class="bg-app-accent h-full rounded-full"
                  :style="{ width: tokenBarWidth(day.tokens) }"
                ></div>
              </div>
            </li>
          </ol>
        </PanelCard>
        <PanelCard class="space-y-4" aria-label="Daily Actual Cost chart">
          <h3 class="text-app-fg-strong m-0 text-xl font-bold">Daily Actual Cost</h3>
          <p v-if="!overview?.daily.length" class="text-app-muted m-0">
            No daily Actual Cost data yet.
          </p>
          <ol v-else class="m-0 space-y-3 p-0">
            <li v-for="day in overview.daily" :key="`cost-${day.date}`" class="list-none space-y-1">
              <div class="flex justify-between gap-4 text-sm">
                <span class="text-app-muted">{{ day.date }}</span>
                <span class="text-app-fg-strong"
                  >{{ day.date }}: {{ formatCost(day.actualCost) }}</span
                >
              </div>
              <div class="bg-app-shell-deep h-3 overflow-hidden rounded-full">
                <div
                  class="bg-status-success h-full rounded-full"
                  :style="{ width: costBarWidth(day.actualCost) }"
                ></div>
              </div>
            </li>
          </ol>
        </PanelCard>
      </div>
    </section>

    <PanelCard class="mt-14 space-y-5" aria-label="Usage Sync status">
      <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <StatusItem
          label="Usage Sync"
          :status="syncStatus()"
          :value="syncing ? 'Sync in progress...' : formatSyncCounts(lastRun)"
        />
        <ActionButton :disabled="syncing" @click="runManualSync">
          {{ syncing ? 'Syncing...' : 'Refresh Usage' }}
        </ActionButton>
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
