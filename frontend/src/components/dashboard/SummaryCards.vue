<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue'
import { useMetricsStore, TIME_RANGES } from '@/stores/metricsStore'
import MetricCard from '@/components/common/MetricCard.vue'
import { getStatusFromValue, worstFilesystem, filesystemTotals, THRESHOLDS } from '@/config/thresholds'
import { toGB, toMB } from '@/utils/format'
import { resolveInterface } from '@/utils/network'
import { prometheusServerApi } from '@/services/prometheus'

const store = useMetricsStore()
const metrics = computed(() => store.latestMetrics)
const loading = computed(() => !store.hasData)

const slaPercent = ref<number | null>(null)
const slaLoading = ref(false)

const loadSla = async () => {
  if (!store.activeServerId) return
  slaLoading.value = true
  try {
    const val = await prometheusServerApi.getServerSla(
      store.activeServerId,
      store.timeRange,
      store.activeServer?.description
    )
    slaPercent.value = val
  } catch {
    slaPercent.value = null
  } finally {
    slaLoading.value = false
  }
}

onMounted(() => {
  loadSla()
})

watch(() => [store.activeServerId, store.timeRange], () => {
  loadSla()
})

const slaStatus = computed<'healthy' | 'warning' | 'critical'>(() => {
  if (slaPercent.value === null) return 'healthy'
  if (slaPercent.value >= 99.0) return 'healthy'
  if (slaPercent.value >= 95.0) return 'warning'
  return 'critical'
})

const cpuCard = computed(() => {
  const usage = metrics.value?.cpu.usage || 0
  const cores = metrics.value?.cpu.cores || store.hostInfo?.cpu_cores || 0
  const freq = metrics.value?.cpu.frequency_mhz ? `${metrics.value.cpu.frequency_mhz.toFixed(0)} MHz` : ''
  const status = getStatusFromValue(usage, THRESHOLDS.cpu.warning, THRESHOLDS.cpu.critical)

  return {
    value: usage.toFixed(1),
    unit: '%',
    percentage: usage,
    subValue: `${cores} Cores ${freq ? '• ' + freq : ''}`,
    status
  }
})

const memCard = computed(() => {
  const mem = metrics.value?.memory
  if (!mem) return { value: '0.0', unit: '%', percentage: 0, subValue: '0 / 0 GB', status: 'healthy' as const }

  const status = getStatusFromValue(mem.usage, THRESHOLDS.memory.warning, THRESHOLDS.memory.critical)

  return {
    value: mem.usage.toFixed(1),
    unit: '%',
    percentage: mem.usage,
    subValue: `${toGB(mem.used)} / ${toGB(mem.total)} GB • Swap ${toGB(mem.swap_used)} GB`,
    status
  }
})

/**
 * Disk pressure is reported per mount, never as a fleet-wide ratio.
 */
const diskCard = computed(() => {
  const filesystems = metrics.value?.filesystems || []
  const worst = worstFilesystem(filesystems)
  const totals = filesystemTotals(filesystems)

  if (!worst) {
    return { value: '0.0', unit: '%', percentage: 0, subValue: 'Tidak ada filesystem', status: 'healthy' as const }
  }

  const status = getStatusFromValue(worst.usage, THRESHOLDS.disk.warning, THRESHOLDS.disk.critical)
  const io = metrics.value?.disk_io
  const ioStr = io ? ` • R ${toMB(io.read_bytes_sec, 1)}M W ${toMB(io.write_bytes_sec, 1)}M` : ''
  const others = filesystems.length > 1 ? ` (+${filesystems.length - 1} mount)` : ''

  return {
    value: worst.usage.toFixed(1),
    unit: '%',
    percentage: worst.usage,
    subValue: `${worst.mountpoint}${others} • total ${toGB(totals.used, 0)}/${toGB(totals.total, 0)} GB${ioStr}`,
    status
  }
})

const netCard = computed(() => {
  const activeIface = resolveInterface(metrics.value?.network, store.selectedIface)
  if (!activeIface) return { value: '0.00', unit: 'MB/s', subValue: 'Tidak ada interface', status: 'healthy' as const }

  const rxMb = toMB(activeIface.rx_bytes_sec)
  const txMb = toMB(activeIface.tx_bytes_sec)

  return {
    value: rxMb,
    unit: 'MB/s',
    subValue: `↓ ${rxMb} • ↑ ${txMb} MB/s (${activeIface.name})`,
    status: 'healthy' as const
  }
})
</script>

<template>
  <div class="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-5 gap-2">
    <MetricCard
      title="CPU USAGE"
      :value="cpuCard.value"
      :unit="cpuCard.unit"
      :percentage="cpuCard.percentage"
      :sub-value="cpuCard.subValue"
      :status="cpuCard.status"
      :loading="loading"
    />

    <MetricCard
      title="MEMORY (RAM)"
      :value="memCard.value"
      :unit="memCard.unit"
      :percentage="memCard.percentage"
      :sub-value="memCard.subValue"
      :status="memCard.status"
      :loading="loading"
    />

    <MetricCard
      title="DISK TERPENUH"
      :value="diskCard.value"
      :unit="diskCard.unit"
      :percentage="diskCard.percentage"
      :sub-value="diskCard.subValue"
      :status="diskCard.status"
      :loading="loading"
    />

    <MetricCard
      title="NETWORK (PRIMARY)"
      :value="netCard.value"
      :unit="netCard.unit"
      :sub-value="netCard.subValue"
      :status="netCard.status"
      :loading="loading"
    />

    <!-- SLA Availability Card with PromQL Range Selector -->
    <div
      class="col-span-2 md:col-span-1 bg-dark-900/90 border rounded-lg px-3 py-2 flex flex-col justify-between shadow-lg backdrop-blur transition-all duration-200"
      :class="slaLoading ? 'border-slate-800/80' : slaStatus === 'critical' ? 'border-rose-500/30 shadow-rose-500/5' : slaStatus === 'warning' ? 'border-amber-500/30 shadow-amber-500/5' : 'border-emerald-500/30 shadow-emerald-500/5'"
    >
      <div class="flex items-center justify-between gap-1">
        <span class="text-[11px] font-semibold uppercase tracking-wider text-slate-400 font-mono truncate">
          SLA UPTIME
        </span>
        <!-- Time Range Selector Pills -->
        <div class="flex items-center gap-0.5 bg-dark-950 p-0.5 rounded border border-slate-800 shrink-0">
          <button
            v-for="r in TIME_RANGES"
            :key="r.key"
            type="button"
            class="px-1.5 py-0.5 text-[9px] font-mono font-bold rounded transition-colors"
            :class="store.timeRange === r.key
              ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40'
              : 'text-slate-500 hover:text-slate-300'"
            :title="`Pilih rentang SLA ${r.label}`"
            @click="store.setTimeRange(r.key)"
          >
            {{ r.key }}
          </button>
        </div>
      </div>

      <div class="my-1.5 flex items-baseline gap-1">
        <span
          class="text-2xl font-bold font-mono tracking-tight tabular-nums"
          :class="slaLoading ? 'text-slate-600' : slaStatus === 'critical' ? 'text-rose-400' : slaStatus === 'warning' ? 'text-amber-400' : 'text-emerald-400'"
        >
          {{ slaLoading ? '--' : slaPercent !== null ? `${slaPercent.toFixed(2)}` : '--' }}
        </span>
        <span class="text-[11px] text-slate-400 font-mono">%</span>
      </div>

      <!-- Track -->
      <div class="w-full bg-dark-950 rounded-full h-1.5 overflow-hidden">
        <div
          v-if="slaPercent !== null && !slaLoading"
          class="h-full rounded-full transition-all duration-300"
          :class="slaStatus === 'critical' ? 'bg-rose-500' : slaStatus === 'warning' ? 'bg-amber-500' : 'bg-emerald-500'"
          :style="{ width: `${Math.min(100, Math.max(0, slaPercent))}%` }"
        ></div>
      </div>

      <div class="mt-1.5 text-[10px] font-mono truncate text-slate-400">
        {{ slaLoading ? 'menghitung SLA...' : `Target 99.9% • TSDB (${TIME_RANGES.find(r => r.key === store.timeRange)?.label})` }}
      </div>
    </div>
  </div>
</template>
