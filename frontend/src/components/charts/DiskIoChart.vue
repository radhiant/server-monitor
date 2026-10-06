<script setup lang="ts">
import { ref, watch, computed } from 'vue'
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
      const read = params.find(p => p.seriesName === 'Read')
      const write = params.find(p => p.seriesName === 'Write')
      return `<div class="font-mono text-xs">
        <span class="text-slate-400">${params[0].name}</span><br/>
        <span class="text-cyan-400 font-bold">Read:</span> ${read ? read.value : 0} MB/s<br/>
        <span class="text-rose-400 font-bold">Write:</span> ${write ? write.value : 0} MB/s
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
      name: 'Read',
      type: 'line',
      smooth: true,
      showSymbol: false,
      lineStyle: { width: 2, color: '#06b6d4' },
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
    },
    {
      name: 'Write',
      type: 'line',
      smooth: true,
      showSymbol: false,
      lineStyle: { width: 2, color: '#f43f5e' },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(244, 63, 94, 0.35)' },
            { offset: 1, color: 'rgba(244, 63, 94, 0.0)' }
          ]
        }
      },
      data: []
    }
  ]
}

const { setOption } = useECharts(chartRef, baseOption)

const latestIo = computed(() => {
  if (props.history.length === 0) return { read: '0.00', write: '0.00' }
  const latest = props.history[props.history.length - 1]
  return {
    read: (latest.disk_io.read_bytes_sec / (1024 * 1024)).toFixed(2),
    write: (latest.disk_io.write_bytes_sec / (1024 * 1024)).toFixed(2),
  }
})

watch(
  () => props.history,
  (newHistory) => {
    if (!newHistory) return
    const timestamps = newHistory.map(m => formatTime24(m.timestamp))
    const readValues = newHistory.map(m => +(m.disk_io.read_bytes_sec / (1024 * 1024)).toFixed(2))
    const writeValues = newHistory.map(m => +(m.disk_io.write_bytes_sec / (1024 * 1024)).toFixed(2))

    setOption({
      xAxis: { data: timestamps },
      series: [
        { name: 'Read', data: readValues },
        { name: 'Write', data: writeValues }
      ]
    })
  }
)
</script>

<template>
  <div class="h-full w-full flex flex-col">
    <div class="flex items-center justify-between px-3 pt-2">
      <span class="text-xs font-mono font-semibold uppercase tracking-wider text-slate-400">
        Disk I/O Throughput (MB/s)
      </span>
      <div class="flex items-center gap-2 text-xs font-mono font-bold">
        <span class="text-cyan-400">R: {{ latestIo.read }} MB/s</span>
        <span class="text-rose-400">W: {{ latestIo.write }} MB/s</span>
      </div>
    </div>
    <div ref="chartRef" class="flex-1 w-full min-h-0"></div>
  </div>
</template>
