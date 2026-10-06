<script setup lang="ts">
import { ref, watch } from 'vue'
import { useMetricsStore, TIME_RANGES } from '@/stores/metricsStore'
import CpuChart from '@/components/charts/CpuChart.vue'
import MemoryChart from '@/components/charts/MemoryChart.vue'
import NetworkChart from '@/components/charts/NetworkChart.vue'
import DiskIoChart from '@/components/charts/DiskIoChart.vue'
import ServerHistoricalChart from '@/components/charts/ServerHistoricalChart.vue'
import { prometheusServerApi, type HistoryTrendPoint } from '@/services/prometheus'
import { Activity, History } from 'lucide-vue-next'

const store = useMetricsStore()
const chartMode = ref<'realtime' | 'history'>('realtime')
const trendPoints = ref<HistoryTrendPoint[]>([])
const trendLoading = ref(false)

const loadHistoricalTrend = async () => {
  if (!store.activeServerId || chartMode.value !== 'history') return
  trendLoading.value = true
  try {
    const pts = await prometheusServerApi.getServerTrend(
      store.activeServerId,
      store.timeRange,
      store.activeServer?.description
    )
    trendPoints.value = pts || []
  } catch {
    trendPoints.value = []
  } finally {
    trendLoading.value = false
  }
}

watch([() => store.activeServerId, () => store.timeRange, chartMode], () => {
  if (chartMode.value === 'history') {
    loadHistoricalTrend()
  }
})
</script>

<template>
  <div class="flex flex-col h-full min-h-0 gap-1.5">
    <!-- Chart Mode Toggle & Range Bar -->
    <div class="flex flex-wrap items-center justify-between gap-1.5 shrink-0 px-1 font-mono text-[11px]">
      <div class="flex items-center gap-1 bg-dark-950 p-0.5 rounded-md border border-slate-800">
        <button
          type="button"
          class="flex items-center gap-1.5 px-2.5 py-0.5 rounded font-semibold transition-colors text-[10px] sm:text-[11px]"
          :class="chartMode === 'realtime'
            ? 'bg-cyan-950/60 border border-cyan-500/40 text-cyan-300'
            : 'text-slate-400 hover:text-slate-200'"
          @click="chartMode = 'realtime'"
        >
          <Activity class="w-3 h-3 text-cyan-400" />
          <span>Realtime (60s)</span>
        </button>
        <button
          type="button"
          class="flex items-center gap-1.5 px-2.5 py-0.5 rounded font-semibold transition-colors text-[10px] sm:text-[11px]"
          :class="chartMode === 'history'
            ? 'bg-cyan-950/60 border border-cyan-500/40 text-cyan-300'
            : 'text-slate-400 hover:text-slate-200'"
          @click="chartMode = 'history'"
        >
          <History class="w-3 h-3 text-cyan-400" />
          <span>Historis<span class="hidden sm:inline"> Prometheus</span></span>
        </button>
      </div>

      <!-- Time Range Pills in Historical Mode -->
      <div v-if="chartMode === 'history'" class="flex items-center gap-0.5 sm:gap-1 bg-dark-950 p-0.5 rounded-md border border-slate-800 overflow-x-auto max-w-full">
        <button
          v-for="r in TIME_RANGES"
          :key="r.key"
          type="button"
          class="px-1.5 sm:px-2 py-0.5 rounded text-[10px] font-semibold transition-colors shrink-0"
          :class="store.timeRange === r.key
            ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40'
            : 'text-slate-400 hover:text-slate-200'"
          @click="store.setTimeRange(r.key)"
        >
          <span class="sm:hidden">{{ r.key }}</span>
          <span class="hidden sm:inline">{{ r.label }}</span>
        </button>
      </div>
    </div>

    <!-- Chart Grid: Realtime Mode -->
    <div v-if="chartMode === 'realtime'" class="grid grid-cols-1 lg:grid-cols-2 lg:grid-rows-2 gap-2 flex-1 min-h-0">
      <div class="bg-dark-900/90 border border-slate-800/80 rounded-lg shadow-lg flex flex-col overflow-hidden min-h-0 h-[190px] sm:h-[210px] lg:h-auto">
        <CpuChart :history="store.history" />
      </div>

      <div class="bg-dark-900/90 border border-slate-800/80 rounded-lg shadow-lg flex flex-col overflow-hidden min-h-0 h-[190px] sm:h-[210px] lg:h-auto">
        <MemoryChart :history="store.history" />
      </div>

      <div class="bg-dark-900/90 border border-slate-800/80 rounded-lg shadow-lg flex flex-col overflow-hidden min-h-0 h-[190px] sm:h-[210px] lg:h-auto">
        <NetworkChart :history="store.history" :selected-interface="store.selectedIface" />
      </div>

      <div class="bg-dark-900/90 border border-slate-800/80 rounded-lg shadow-lg flex flex-col overflow-hidden min-h-0 h-[190px] sm:h-[210px] lg:h-auto">
        <DiskIoChart :history="store.history" />
      </div>
    </div>

    <!-- Chart: Prometheus Historical Mode -->
    <div v-else class="bg-dark-900/90 border border-slate-800/80 rounded-lg shadow-lg flex-1 min-h-[320px] h-[340px] sm:h-[380px] lg:h-auto overflow-hidden">
      <ServerHistoricalChart
        :points="trendPoints"
        :loading="trendLoading"
        :range-label="TIME_RANGES.find(r => r.key === store.timeRange)?.label"
      />
    </div>
  </div>
</template>
