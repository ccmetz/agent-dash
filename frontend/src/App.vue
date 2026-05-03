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
  <main
    class="min-h-screen bg-[radial-gradient(circle_at_top_left,#24446a,#101722_45%,#080b10)] px-5 py-10 font-sans text-[#e5edf7] sm:px-8 md:p-16"
  >
    <section class="max-w-2xl space-y-5">
      <p class="text-xs font-bold tracking-[0.12em] text-[#8fb5e3] uppercase">Agent Dash</p>
      <h1 class="m-0 pb-2 text-[clamp(2.5rem,9vw,6rem)] leading-[0.95]">Usage Overview</h1>
      <p class="max-w-xl text-xl leading-8 text-[#b7c8dc]">
        Local Usage Metadata will appear here after the first Usage Sync.
      </p>
    </section>

    <section
      class="mt-14 w-full max-w-lg space-y-4 rounded-[1.25rem] border border-white/15 bg-white/8 p-6 shadow-[0_24px_70px_rgb(0_0_0/0.35)] sm:p-8"
      aria-label="Backend status"
    >
      <span
        v-if="status?.ok"
        class="inline-flex rounded-full bg-[#78f0aa] px-3 py-1.5 font-bold text-[#062113]"
      >
        Connected
      </span>
      <span
        v-else-if="error"
        class="inline-flex rounded-full bg-[#ff8f8f] px-3 py-1.5 font-bold text-[#2c0707]"
      >
        Disconnected
      </span>
      <span
        v-else
        class="inline-flex rounded-full bg-[#bdd1e8] px-3 py-1.5 font-bold text-[#0d1724]"
      >
        Connecting
      </span>
      <p class="pt-2 text-xs font-bold tracking-[0.12em] text-[#8fb5e3] uppercase">
        Analytics Store
      </p>
      <p class="m-0 [overflow-wrap:anywhere] font-mono text-[#f8fbff]">
        {{ status?.analyticsStorePath ?? 'Waiting for backend status...' }}
      </p>
    </section>
  </main>
</template>
