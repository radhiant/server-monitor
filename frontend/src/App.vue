<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useMetricsStore } from '@/stores/metricsStore'
import { fetchServerRegistry } from '@/services/registry'
import { useNightlyReload } from '@/composables/useNightlyReload'
import HeaderSection from '@/components/dashboard/HeaderSection.vue'
import FleetView from '@/components/fleet/FleetView.vue'
import ServerView from '@/components/dashboard/ServerView.vue'

const store = useMetricsStore()

// Flush browser heap at 04:00 AM daily — wall displays run 24/7 and ECharts
// canvas instances plus reactive proxies slowly accumulate memory.
useNightlyReload(4)

/**
 * Two views, no router.
 *
 * `?view=fleet` (the default) shows every server at once; `?view=server&server=id`
 * drills into one. Keeping it in the query string means the existing NOC
 * copy-links still resolve, and each view can be pinned to its own display.
 */
const isTv = computed(() => store.fitMode === 'tv')

onMounted(async () => {
  const registry = await fetchServerRegistry()
  store.setServers(registry)
})
</script>

<template>
  <div
    class="bg-dark-950 text-slate-200 flex flex-col selection:bg-cyan-500/20 selection:text-cyan-200 min-h-screen"
    :class="isTv ? 'lg:h-screen lg:overflow-hidden' : ''"
  >
    <HeaderSection />

    <FleetView v-if="store.view === 'fleet'" />
    <ServerView v-else />

    <footer
      class="shrink-0 py-1 px-3 text-center border-t border-slate-800/60 text-[10px] font-mono text-slate-600"
    >
      Server Monitor V1 • Distributed Realtime Telemetry
    </footer>
  </div>
</template>
