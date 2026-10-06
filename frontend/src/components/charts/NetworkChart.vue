<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useECharts } from '@/composables/useECharts'
import { formatTime24 } from '@/utils/format'
import { resolveInterface, sortInterfacesForDisplay, isVirtualInterface } from '@/utils/network'
import { useMetricsStore } from '@/stores/metricsStore'
import type { ServerMetrics } from '@/types/metrics'
import type { EChartsOption } from '@/lib/echarts'

const props = defineProps<{
  history: ServerMetrics[]
  selectedInterface: string
}>()

const store = useMetricsStore()
const chartRef = ref<HTMLElement | null>(null)


const baseOption: EChartsOption = {
  backgroundColor: 'transparent',
  grid: {
    top: 25,
    right: 15,
    bottom: 20,
    left: 45,
    containLabel: false
  },
  tooltip: {
    trigger: 'axis',
    backgroundColor: '#0f172a',
    borderColor: '#334155',
    textStyle: { color: '#e2e8f0', fontFamily: 'monospace', fontSize: 11 },
    formatter: (params: any) => {
      if (!Array.isArray(params) || params.length === 0) return ''
      const rx = params.find(p => p.seriesName === 'RX (Download)')
      const tx = params.find(p => p.seriesName === 'TX (Upload)')
      return `<div class="font-mono text-xs">
        <span class="text-slate-400">${params[0].name}</span><br/>
        <span class="text-emerald-400 font-bold">RX:</span> ${rx ? rx.value : 0} MB/s<br/>
        <span class="text-blue-400 font-bold">TX:</span> ${tx ? tx.value : 0} MB/s
      </div>`
    }
  },
  legend: {
    show: true,
    top: 0,
    right: 10,
    textStyle: { color: '#94a3b8', fontSize: 10, fontFamily: 'monospace' },
    itemWidth: 10,
    itemHeight: 6
  },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: [],
    axisLine: { lineStyle: { color: '#1e293b' } },
    axisTick: { show: false },
    axisLabel: { show: false }
  },
  yAxis: {
    type: 'value',
    splitLine: { lineStyle: { color: '#1e293b', type: 'dashed' } },
    axisLabel: {
      color: '#64748b',
      fontFamily: 'monospace',
      fontSize: 10,
      formatter: '{value}M'
    }
  },
  series: [
    {
      name: 'RX (Download)',
      type: 'line',
      smooth: true,
      showSymbol: false,
      lineStyle: { width: 2, color: '#10b981' },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(16, 185, 129, 0.35)' },
            { offset: 1, color: 'rgba(16, 185, 129, 0.0)' }
          ]
        }
      },
      data: []
    },
    {
      name: 'TX (Upload)',
      type: 'line',
      smooth: true,
      showSymbol: false,
      lineStyle: { width: 2, color: '#3b82f6' },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(59, 130, 246, 0.35)' },
            { offset: 1, color: 'rgba(59, 130, 246, 0.0)' }
          ]
        }
      },
      data: []
    }
  ]
}

const { setOption } = useECharts(chartRef, baseOption)

const availableInterfaces = computed(() => {
  const latest = props.history[props.history.length - 1] || store.latestMetrics
  // Physical NICs first: a Docker host reports forty veth pairs and bridges.
  return sortInterfacesForDisplay(latest?.network)
})

const latestRates = computed(() => {
  if (props.history.length === 0) return { rx: '0.00', tx: '0.00' }
  const latest = props.history[props.history.length - 1]
  const iface = resolveInterface(latest.network, props.selectedInterface)
  if (!iface) return { rx: '0.00', tx: '0.00' }

  return {
    rx: (iface.rx_bytes_sec / (1024 * 1024)).toFixed(2),
    tx: (iface.tx_bytes_sec / (1024 * 1024)).toFixed(2),
  }
})

watch(
  () => [props.history, props.selectedInterface],
  () => {
    if (!props.history || props.history.length === 0) return
    const timestamps = props.history.map(m => formatTime24(m.timestamp))
    
    const rxValues = props.history.map(m => {
      const iface = resolveInterface(m.network, props.selectedInterface)
      return iface ? +(iface.rx_bytes_sec / (1024 * 1024)).toFixed(2) : 0
    })

    const txValues = props.history.map(m => {
      const iface = resolveInterface(m.network, props.selectedInterface)
      return iface ? +(iface.tx_bytes_sec / (1024 * 1024)).toFixed(2) : 0
    })

    setOption({
      xAxis: { data: timestamps },
      series: [
        { name: 'RX (Download)', data: rxValues },
        { name: 'TX (Upload)', data: txValues }
      ]
    })
  }
)
</script>

<template>
  <div class="h-full w-full flex flex-col font-mono">
    <div class="flex items-center justify-between px-3 pt-2">
      <!-- Title + Interface Selector Dropdown -->
      <div class="flex items-center gap-2">
        <span class="text-xs font-semibold uppercase tracking-wider text-slate-400">
          Network Traffic
        </span>
        <div v-if="availableInterfaces.length > 0" class="flex items-center gap-1">
          <select
            v-model="store.selectedIface"
            class="bg-dark-950 border border-slate-700/80 hover:border-cyan-500/50 text-cyan-400 rounded px-1.5 py-0.5 text-[10px] focus:outline-none transition-colors"
          >
            <option v-for="iface in availableInterfaces" :key="iface.name" :value="iface.name">
              {{ iface.name }}{{ isVirtualInterface(iface.name) ? ' (virtual)' : '' }}
            </option>
          </select>
        </div>
      </div>

      <!-- Realtime Rates -->
      <div class="flex items-center gap-2 text-xs font-bold">
        <span class="text-emerald-400">↓ {{ latestRates.rx }} MB/s</span>
        <span class="text-blue-400">↑ {{ latestRates.tx }} MB/s</span>
      </div>
    </div>
    <div ref="chartRef" class="flex-1 w-full min-h-0"></div>
  </div>
</template>
