<script setup lang="ts">
import { computed } from 'vue'
import { useMetricsStore } from '@/stores/metricsStore'
import CoreGrid from '@/components/common/CoreGrid.vue'
import { toGB } from '@/utils/format'
import { THRESHOLDS } from '@/config/thresholds'
import { HardDrive, Cpu, Activity } from 'lucide-vue-next'

const store = useMetricsStore()
const metrics = computed(() => store.latestMetrics)

const cores = computed(() => metrics.value?.cpu.per_core || [])
const loadAvg = computed(() => metrics.value?.load)
const numCores = computed(() => metrics.value?.cpu.cores || store.hostInfo?.cpu_cores || 1)

/** Load is only meaningful relative to core count: 8.0 on 8 cores is 100%, not a crisis. */
const loadRatio = (value: number | undefined) =>
  value === undefined ? 0 : (value / numCores.value) * 100

const loads = computed(() => [
  { key: '1m', label: '1 MIN', value: loadAvg.value?.['1m'] },
  { key: '5m', label: '5 MIN', value: loadAvg.value?.['5m'] },
  { key: '15m', label: '15 MIN', value: loadAvg.value?.['15m'] },
])

function getLoadColor(ratio: number): { text: string; bar: string } {
  if (ratio >= THRESHOLDS.loadRatio.critical * 100) return { text: 'text-rose-400', bar: 'bg-rose-500' }
  if (ratio >= THRESHOLDS.loadRatio.warning * 100) return { text: 'text-amber-400', bar: 'bg-amber-500' }
  return { text: 'text-emerald-400', bar: 'bg-emerald-500' }
}

function getDiskColor(usage: number): { text: string; bar: string } {
  if (usage >= THRESHOLDS.disk.critical) return { text: 'text-rose-400', bar: 'bg-rose-500' }
  if (usage >= THRESHOLDS.disk.warning) return { text: 'text-amber-400', bar: 'bg-amber-500' }
  return { text: 'text-cyan-400', bar: 'bg-cyan-500' }
}

/** Fullest mount first: the one that will fill up is the one worth reading. */
const filesystems = computed(() =>
  [...(metrics.value?.filesystems || [])].sort((a, b) => b.usage - a.usage)
)
</script>

<template>
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-2 min-h-0">
    <!-- 1. CPU per-core breakdown -->
    <div class="bg-dark-900/90 border border-slate-800/80 rounded-lg p-2 shadow-lg flex flex-col min-h-[105px] lg:min-h-0">
      <div class="flex items-center justify-between mb-1.5 shrink-0">
        <div class="flex items-center gap-1.5 text-[11px] font-mono font-semibold text-slate-300">
          <Cpu class="w-3.5 h-3.5 text-cyan-400" />
          <span>CPU CORES ({{ cores.length }})</span>
        </div>
        <span class="text-[10px] font-mono text-slate-500 truncate max-w-[55%]" :title="store.hostInfo?.cpu_model">
          {{ store.hostInfo?.cpu_model || 'Multi-Core Processor' }}
        </span>
      </div>

      <div class="flex-1 min-h-0 overflow-y-auto">
        <CoreGrid v-if="cores.length > 0" :cores="cores" />
        <div v-else class="h-full flex items-center justify-center text-[11px] font-mono text-slate-500">
          Menunggu telemetri core...
        </div>
      </div>
    </div>

    <!-- 2. Load average, normalised to core count -->
    <div class="bg-dark-900/90 border border-slate-800/80 rounded-lg p-2 shadow-lg flex flex-col min-h-[105px] lg:min-h-0 font-mono">
      <div class="flex items-center justify-between mb-1.5 shrink-0">
        <div class="flex items-center gap-1.5 text-[11px] font-semibold text-slate-300">
          <Activity class="w-3.5 h-3.5 text-cyan-400" />
          <span>LOAD AVERAGE</span>
        </div>
        <span class="text-[10px] text-slate-500">Baseline {{ numCores }} core</span>
      </div>

      <div class="flex-1 min-h-0 grid grid-cols-3 gap-1.5 items-center">
        <div
          v-for="load in loads"
          :key="load.key"
          class="bg-dark-950/80 border border-slate-800 rounded p-1.5 text-center"
        >
          <div class="text-[9px] text-slate-500 mb-0.5">{{ load.label }}</div>
          <div class="text-base font-bold tabular-nums" :class="getLoadColor(loadRatio(load.value)).text">
            {{ load.value !== undefined ? load.value.toFixed(2) : '--' }}
          </div>
          <div class="w-full bg-slate-900 rounded-full h-1 mt-1 overflow-hidden">
            <div
              class="h-full rounded-full transition-all duration-300"
              :class="getLoadColor(loadRatio(load.value)).bar"
              :style="{ width: `${Math.min(100, loadRatio(load.value))}%` }"
            ></div>
          </div>
        </div>
      </div>
    </div>

    <!-- 3. Filesystems, fullest first -->
    <div class="bg-dark-900/90 border border-slate-800/80 rounded-lg p-2 shadow-lg flex flex-col min-h-[105px] lg:min-h-0 font-mono">
      <div class="flex items-center justify-between mb-1.5 shrink-0">
        <div class="flex items-center gap-1.5 text-[11px] font-semibold text-slate-300">
          <HardDrive class="w-3.5 h-3.5 text-cyan-400" />
          <span>FILESYSTEMS</span>
        </div>
        <span class="text-[10px] text-slate-500">{{ filesystems.length }} mount</span>
      </div>

      <div class="flex-1 min-h-0 overflow-y-auto space-y-1 pr-1">
        <div
          v-for="fs in filesystems"
          :key="fs.mountpoint"
          class="bg-dark-950/80 border border-slate-800/80 rounded px-1.5 py-1"
        >
          <div class="flex items-center justify-between text-[10px] mb-0.5 gap-2">
            <span class="font-bold text-slate-200 truncate" :title="fs.mountpoint">{{ fs.mountpoint }}</span>
            <span class="text-slate-500 shrink-0 tabular-nums">
              {{ toGB(fs.used, 0) }}/{{ toGB(fs.total, 0) }}G
              <span class="font-bold" :class="getDiskColor(fs.usage).text">{{ fs.usage.toFixed(0) }}%</span>
            </span>
          </div>
          <div class="w-full bg-slate-900 rounded-full h-1 overflow-hidden">
            <div
              class="h-full rounded-full transition-all duration-300"
              :class="getDiskColor(fs.usage).bar"
              :style="{ width: `${Math.min(100, fs.usage)}%` }"
            ></div>
          </div>
        </div>

        <div v-if="filesystems.length === 0" class="h-full flex items-center justify-center text-[11px] text-slate-500">
          Belum ada filesystem dilaporkan
        </div>
      </div>
    </div>
  </div>
</template>
