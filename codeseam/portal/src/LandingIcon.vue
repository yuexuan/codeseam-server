<script setup lang="ts">
defineProps<{ name: string }>()
</script>

<template>
  <svg
    class="landing-line-icon"
    viewBox="0 0 24 24"
    xmlns="http://www.w3.org/2000/svg"
    fill="none"
    stroke="currentColor"
    stroke-width="1.65"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <template v-if="name === 'switch'">
      <path d="M3 7h18m-5-4 5 4-5 4M21 17H3m5-4-5 4 5 4" />
    </template>
    <template v-else-if="name === 'network'">
      <circle cx="12" cy="12" r="9" />
      <path d="M3 12h18M12 3a16 16 0 0 1 0 18M12 3a16 16 0 0 0 0 18" />
    </template>
    <template v-else-if="name === 'activity'">
      <path d="M3 12h4l3-7 4 14 3-7h4" />
    </template>
    <template v-else-if="name === 'sliders'">
      <path d="M4 7h16M4 17h16" />
      <circle cx="9" cy="7" r="2" fill="var(--surface)" />
      <circle cx="16" cy="17" r="2" fill="var(--surface)" />
    </template>
    <template v-else-if="name === 'link'">
      <path d="M10 13a5 5 0 0 0 7.5.5l2-2a5 5 0 0 0-7-7l-1.2 1.2M14 11a5 5 0 0 0-7.5-.5l-2 2a5 5 0 0 0 7 7l1.2-1.2" />
    </template>
    <template v-else-if="name === 'radar'">
      <circle cx="12" cy="12" r="9" />
      <circle cx="12" cy="12" r="4" />
      <path d="m12 12 5.5-5.5" />
      <circle cx="12" cy="12" r="1" fill="currentColor" stroke="none" />
    </template>
    <template v-else-if="name === 'chart'">
      <path d="M4 20V4M4 20h16M8 15l4-5 3 2 5-6" />
    </template>
    <template v-else-if="name === 'user'">
      <circle cx="12" cy="8" r="3.5" />
      <path d="M5 20v-2a7 7 0 0 1 14 0v2" />
    </template>
    <template v-else-if="name === 'download'">
      <path d="M12 3v12m-4-4 4 4 4-4M4 17v3h16v-3" />
    </template>
    <template v-else-if="name === 'terminal'">
      <rect x="3" y="4" width="18" height="16" rx="2" />
      <path d="m7 9 4 3-4 3m6 0h4" />
    </template>
    <template v-else-if="name === 'monitor'">
      <rect x="3" y="4" width="18" height="14" rx="2" />
      <path d="M8 22h8m-4-4v4" />
    </template>
    <template v-else-if="name === 'shield'">
      <path d="m12 2 8 4v6c0 5-3.2 8.3-8 10-4.8-1.7-8-5-8-10V6z" />
      <path d="m8.5 12 2.5 2.5 4.5-5" />
    </template>
    <template v-else-if="name === 'refresh'">
      <path d="M20 7v5h-5M4 17v-5h5" />
      <path d="M5.5 9A7.5 7.5 0 0 1 18 6l2 6M4 12l2 6a7.5 7.5 0 0 0 12.5-3" />
    </template>
  </svg>
</template>
