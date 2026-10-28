<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import PortalApp from './App.vue'
import LandingHome from './LandingHome.vue'

function resolveView(): 'home' | 'account' {
  const path = window.location.pathname.replace(/\/+$/, '') || '/'
  return path === '/app' || path.startsWith('/app/') || ['/login', '/register', '/forgot'].includes(path) ? 'account' : 'home'
}
const view = ref<'home' | 'account'>(resolveView())
function syncRoute() { view.value = resolveView() }
onMounted(() => window.addEventListener('popstate', syncRoute))
onUnmounted(() => window.removeEventListener('popstate', syncRoute))
</script>

<template>
  <PortalApp v-if="view === 'account'" />
  <LandingHome v-else />
</template>