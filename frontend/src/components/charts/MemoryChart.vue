<script setup lang="ts">
import { ref, watch } from 'vue'
import { useECharts } from '@/composables/useECharts'
import { formatTime24 } from '@/utils/format'
import type { ServerMetrics } from '@/types/metrics'
import type { EChartsOption } from '@/lib/echarts'

const props = defineProps<{
  history: ServerMetrics[]
}>()

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
      const ram = params.find(p => p.seriesName === 'RAM Used')
      const swap = params.find(p => p.seriesName === 'Swap Used')
      return `<div class="font-mono text-xs">
        <span class="text-slate-400">${params[0].name}</span><br/>
        <span class="text-purple-400 font-bold">RAM:</span> ${ram ? ram.value : 0} GB<br/>
        <span class="text-amber-400 font-bold">Swap:</span> ${swap ? swap.value : 0} GB
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
    min: 0,
    // `max` is set per update from the host's installed RAM, see the watcher.
    splitNumber: 4,
    splitLine: { lineStyle: { color: '#1e293b', type: 'dashed' } },
    axisLabel: {
      color: '#64748b',
      fontFamily: 'monospace',
      fontSize: 10,
      formatter: '{value}G'
    }
  },
  series: [
    {
      name: 'RAM Used',
      type: 'line',
      smooth: true,
      showSymbol: false,
      lineStyle: { width: 2, color: '#a855f7' },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(168, 85, 247, 0.35)' },
            { offset: 1, color: 'rgba(168, 85, 247, 0.0)' }
          ]
        }
      },
      data: []
    },
    {
      name: 'Swap Used',
      type: 'line',
      smooth: true,
      showSymbol: false,
      lineStyle: { width: 1.5, color: '#f59e0b', type: 'dashed' },
      data: []
    }
  ]
}

/** Stick sizes a physical machine is actually built from. */
const STANDARD_GB = [1, 2, 4, 8, 12, 16, 24, 32, 48, 64, 96, 128, 192, 256, 384, 512]

/**
 * Pick the axis ceiling for a host's RAM.
 *
 * Linux reports a little under the marketed size (15.21 GiB on a 16 GB box),
 * so snapping up to the nearest standard size gives both a familiar number and
 * clean gridlines. The 12% guard keeps that from lying about a VM: 15.21 snaps
 * to 16, but an 8.1 GiB VM stays at 8.1 rather than claiming a 16 GB ceiling
 * (or, with plain even-rounding, an invented 10).
 */
function axisCeilGb(gb: number): number {
  if (!Number.isFinite(gb) || gb <= 0) return 0
  const standard = STANDARD_GB.find(size => size >= gb && size <= gb * 1.12)
  return standard ?? Math.ceil(gb * 10) / 10
}

/** Roughly four gridlines, always on a whole number of gigabytes. */
function axisStepGb(max: number): number {
  return Math.max(1, Math.round(max / 4))
}

const { setOption } = useECharts(chartRef, baseOption)

watch(
  () => props.history,
  (newHistory) => {
    if (!newHistory) return
    const timestamps = newHistory.map(m => formatTime24(m.timestamp))
    const ramValues = newHistory.map(m => +(m.memory.used / (1024 * 1024 * 1024)).toFixed(2))
    const swapValues = newHistory.map(m => +(m.memory.swap_used / (1024 * 1024 * 1024)).toFixed(2))

    // Fix the ceiling at installed RAM rather than letting ECharts fit the
    // peak. Auto-scaling made 3.9 GB used out of 15.2 GB draw as a full chart,
    // which reads as a machine about to run out when it has ample headroom.
    const capacityGb = (newHistory[newHistory.length - 1]?.memory.total || 0) / (1024 * 1024 * 1024)
    const peakGb = Math.max(0, ...ramValues, ...swapValues)
    // Guard against a host whose swap outgrows RAM: never clip a plotted line.
    const axisMax = Math.max(axisCeilGb(capacityGb), axisCeilGb(peakGb)) || undefined

    setOption({
      xAxis: { data: timestamps },
      yAxis: axisMax
        ? { max: axisMax, interval: axisStepGb(axisMax) }
        : {},
      series: [
        { name: 'RAM Used', data: ramValues },
        { name: 'Swap Used', data: swapValues }
      ]
    })
  }
)
</script>

<template>
  <div class="h-full w-full flex flex-col">
    <div class="flex items-center justify-between px-3 pt-2">
      <span class="text-xs font-mono font-semibold uppercase tracking-wider text-slate-400">
        Memory & Swap Realtime
      </span>
      <span v-if="history.length > 0" class="text-xs font-mono font-bold text-purple-400">
        {{ (history[history.length - 1].memory.used / (1024 * 1024 * 1024)).toFixed(2) }} / {{ (history[history.length - 1].memory.total / (1024 * 1024 * 1024)).toFixed(1) }} GB
      </span>
    </div>
    <div ref="chartRef" class="flex-1 w-full min-h-0"></div>
  </div>
</template>
