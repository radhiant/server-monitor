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
    left: 40,
    containLabel: false
  },
  tooltip: {
    trigger: 'axis',
    backgroundColor: '#0f172a',
    borderColor: '#334155',
    textStyle: { color: '#e2e8f0', fontFamily: 'monospace', fontSize: 11 },
    formatter: (params: any) => {
      if (!Array.isArray(params) || params.length === 0) return ''
      const item = params[0]
      return `<div class="font-mono text-xs">
        <span class="text-slate-400">${item.name}</span><br/>
        <span class="text-cyan-400 font-bold">CPU Usage:</span> ${item.value}%
      </div>`
    }
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
    max: 100,
    interval: 25,
    splitLine: { lineStyle: { color: '#1e293b', type: 'dashed' } },
    axisLabel: {
      color: '#64748b',
      fontFamily: 'monospace',
      fontSize: 10,
      formatter: '{value}%'
    }
  },
  series: [
    {
      name: 'CPU Usage',
      type: 'line',
      smooth: true,
      showSymbol: false,
      lineStyle: {
        width: 2,
        color: '#06b6d4'
      },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(6, 182, 212, 0.35)' },
            { offset: 1, color: 'rgba(6, 182, 212, 0.0)' }
          ]
        }
      },
      data: []
    }
  ]
}

const { setOption } = useECharts(chartRef, baseOption)

watch(
  () => props.history,
  (newHistory) => {
    if (!newHistory) return
    const timestamps = newHistory.map(m => formatTime24(m.timestamp))
    const cpuValues = newHistory.map(m => +(m.cpu.usage.toFixed(1)))

    setOption({
      xAxis: { data: timestamps },
      series: [{ data: cpuValues }]
    })
  }
)
</script>

<template>
  <div class="h-full w-full flex flex-col">
    <div class="flex items-center justify-between px-3 pt-2">
      <span class="text-xs font-mono font-semibold uppercase tracking-wider text-slate-400">
        CPU Realtime (60s)
      </span>
      <span v-if="history.length > 0" class="text-xs font-mono font-bold text-cyan-400">
        {{ history[history.length - 1].cpu.usage.toFixed(1) }}%
      </span>
    </div>
    <div ref="chartRef" class="flex-1 w-full min-h-0"></div>
  </div>
</template>
