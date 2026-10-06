<script setup lang="ts">
import { computed } from 'vue'
import { useMetricsStore } from '@/stores/metricsStore'
import { useMetricsStream } from '@/composables/useMetricsStream'
import SummaryCards from '@/components/dashboard/SummaryCards.vue'
import ChartsSection from '@/components/dashboard/ChartsSection.vue'
import SystemSection from '@/components/dashboard/SystemSection.vue'
import ProcessTable from '@/components/dashboard/ProcessTable.vue'
import MemoryPie from '@/components/charts/MemoryPie.vue'
import PhysicalView from '@/components/common/PhysicalView.vue'
import { AlertTriangle, RefreshCw, Loader2 } from 'lucide-vue-next'

/**
 * Single-server detail dashboard.
 *
 * In TV mode every panel is sized as a fraction of the viewport so the whole
 * dashboard lands on one screen with nothing to scroll — the wall display can
 * be left unattended. Scroll mode keeps the classic long page for laptops,
 * where dividing a 768px viewport this many ways would be unreadable.
 */
const store = useMetricsStore()
const { initStream } = useMetricsStream()

const isTv = computed(() => store.fitMode === 'tv')

const showReconnecting = computed(() =>
  !store.isMockMode && store.connectionState === 'RECONNECTING'
)

const showOffline = computed(() =>
  !store.isMockMode &&
  (store.connectionState === 'DISCONNECTED' || store.connectionState === 'ERROR')
)

const serverLabel = computed(() =>
  store.activeServer?.name || store.activeServer?.baseUrl || 'server'
)
</script>

<template>
  <div
    class="flex-1 min-h-0 p-2 sm:p-2.5 flex flex-col gap-2"
    :class="isTv ? 'lg:overflow-hidden overflow-y-auto' : 'overflow-y-auto'"
  >
    <!-- Connection banners -->
    <div
      v-if="showReconnecting"
      class="shrink-0 bg-amber-950/40 border border-amber-500/40 rounded-lg px-3 py-1.5 flex flex-wrap sm:flex-nowrap items-center justify-between gap-2 sm:gap-3 font-mono text-[11px] text-amber-300 backdrop-blur"
    >
      <div class="flex items-center gap-2 min-w-0">
        <Loader2 class="w-4 h-4 text-amber-400 animate-spin shrink-0" />
        <span class="truncate">
          <strong>Menyambung ulang:</strong> koneksi ke
          <code class="bg-amber-900/50 px-1 py-0.5 rounded text-amber-200">{{ serverLabel }}</code>
          terputus.
        </span>
      </div>
      <button
        type="button"
        class="flex items-center gap-1.5 px-2.5 py-1 rounded bg-amber-900/60 hover:bg-amber-800/80 text-amber-200 transition-colors border border-amber-700/50 shrink-0"
        @click="initStream"
      >
        <RefreshCw class="w-3.5 h-3.5" />
        <span>Sambung Ulang</span>
      </button>
    </div>

    <div
      v-else-if="showOffline"
      class="shrink-0 bg-rose-950/40 border border-rose-500/40 rounded-lg px-3 py-1.5 flex flex-wrap sm:flex-nowrap items-center justify-between gap-2 sm:gap-3 font-mono text-[11px] text-rose-300 backdrop-blur"
    >
      <div class="flex items-center gap-2 min-w-0">
        <AlertTriangle class="w-4 h-4 text-rose-400 shrink-0" />
        <span class="truncate">
          <strong>Agent tidak dapat dihubungi:</strong>
          <code class="bg-rose-900/50 px-1 py-0.5 rounded text-rose-200">{{ serverLabel }}</code>
        </span>
      </div>
      <button
        type="button"
        class="flex items-center gap-1.5 px-2.5 py-1 rounded bg-rose-900/60 hover:bg-rose-800/80 text-rose-200 transition-colors border border-rose-700/50 shrink-0"
        @click="initStream"
      >
        <RefreshCw class="w-3.5 h-3.5" />
        <span>Coba Lagi</span>
      </button>
    </div>

    <!-- 1. Headline metrics -->
    <SummaryCards class="shrink-0" />

    <!-- 2. Charts and hardware panels, with the physical rail spanning both rows. -->
    <div
      class="flex flex-col xl:grid xl:grid-cols-12 gap-2 min-h-0"
      :class="isTv ? 'lg:flex-[1.6_1_0%]' : 'xl:shrink-0 xl:h-[590px]'"
    >
      <div class="col-span-12 xl:col-span-9 2xl:col-span-10 flex flex-col gap-2 min-h-0">
        <div class="flex-1 min-h-0">
          <ChartsSection />
        </div>
        <SystemSection class="shrink-0" :class="isTv ? 'lg:h-[126px]' : 'lg:h-[160px]'" />
      </div>

      <!-- Narrow rail on wide screens, responsive grid/stack on mobile -->
      <div class="col-span-12 xl:col-span-3 2xl:col-span-2 flex flex-col sm:grid sm:grid-cols-2 xl:flex xl:flex-col gap-2 min-h-0">
        <PhysicalView :server="store.activeServer" class="shrink-0" />
        <MemoryPie :memory="store.latestMetrics?.memory" class="shrink-0 xl:flex-1 min-h-[190px] xl:min-h-0" />
      </div>
    </div>

    <!-- 3. Process table -->
    <ProcessTable :class="isTv ? 'lg:flex-1 lg:min-h-0 h-[380px] lg:h-auto' : 'shrink-0 h-[360px] sm:h-[320px]'" />
  </div>
</template>
