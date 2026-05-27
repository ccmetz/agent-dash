<template>
  <button
    class="border-app-accent/70 bg-app-shell/40 text-app-fg-strong hover:border-app-fg-strong hover:bg-app-accent/15 focus-visible:ring-app-accent cursor-pointer rounded-lg border px-4 py-2 text-sm font-bold shadow-sm shadow-black/20 transition hover:-translate-y-0.5 hover:shadow-md focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-app-shell focus-visible:outline-none active:translate-y-0 disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0 disabled:hover:border-app-accent/70 disabled:hover:bg-app-shell/40 disabled:hover:shadow-sm"
    type="button"
  >
    <slot />
  </button>
</template>
