<script setup lang="ts">
import { ref, watch } from 'vue'
import { useECharts } from '@/composables/useECharts'
import type { HistoryTrendPoint } from '@/services/prometheus'
import type { EChartsOption } from '@/lib/echarts'
import { Loader2 } from 'lucide-vue-next'

const props = defineProps<{
  points: HistoryTrendPoint[]
  loading?: boolean
  rangeLabel?: string
}>()

const chartRef = ref<HTMLElement | null>(null)

const baseOption: EChartsOption = {
  backgroundColor: 'transparent',
  grid: {
    top: 32,
    right: 15,
    bottom: 25,
    left: 40,
    containLabel: false,
  },
  tooltip: {
    trigger: 'axis',
    backgroundColor: '#0f172a',
    borderColor: '#334155',
    textStyle: { color: '#e2e8f0', fontFamily: 'monospace', fontSize: 11 },
    formatter: (params: any) => {
      if (!Array.isArray(params) || params.length === 0) return ''
      const cpu = params.find(p => p.seriesName === 'CPU')
      const mem = params.find(p => p.seriesName === 'RAM')
      return `<div class="font-mono text-xs">
        <span class="text-slate-400">${params[0].name}</span><br/>
        <span class="text-cyan-400 font-bold">CPU:</span> ${cpu ? cpu.value : 0}%<br/>
        <span class="text-purple-400 font-bold">RAM:</span> ${mem ? mem.value : 0}%
      </div>`
    },
  },
  legend: {
    data: ['CPU', 'RAM'],
    top: 4,
    right: 12,
    textStyle: { color: '#94a3b8', fontFamily: 'monospace', fontSize: 10 },
    itemWidth: 10,
    itemHeight: 6,
  },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: [],
    axisLine: { lineStyle: { color: '#1e293b' } },
    axisTick: { show: false },
    axisLabel: {
      color: '#64748b',
      fontFamily: 'monospace',
      fontSize: 9,
      interval: 'auto',
    },
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
      fontSize: 9,
      formatter: '{value}%',
    },
  },
  series: [
    {
      name: 'CPU',
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
            { offset: 0, color: 'rgba(6, 182, 212, 0.30)' },
            { offset: 1, color: 'rgba(6, 182, 212, 0.0)' },
          ],
        },
      },
      data: [],
    },
    {
      name: 'RAM',
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
            { offset: 0, color: 'rgba(168, 85, 247, 0.28)' },
            { offset: 1, color: 'rgba(168, 85, 247, 0.0)' },
          ],
        },
      },
      data: [],
    },
  ],
}

const { setOption } = useECharts(chartRef, baseOption)

watch(
  () => props.points,
  (points) => {
    if (!points || points.length === 0) return
    setOption({
      xAxis: { data: points.map(p => p.time) },
      series: [
        { name: 'CPU', data: points.map(p => p.cpu) },
        { name: 'RAM', data: points.map(p => p.memory) },
      ],
    })
  },
  { immediate: true },
)
</script>

<template>
  <div class="h-full w-full flex flex-col relative font-mono">
    <div class="flex items-center justify-between px-2.5 sm:px-3 pt-2 shrink-0 flex-wrap sm:flex-nowrap gap-1">
      <div class="flex items-center gap-1.5 sm:gap-2">
        <span class="text-xs font-semibold uppercase tracking-wider text-slate-300">
          <span class="hidden sm:inline">Grafik </span>Historis CPU &amp; RAM
        </span>
        <span v-if="rangeLabel" class="text-[9px] sm:text-[10px] px-1.5 py-0.5 rounded bg-dark-950 border border-slate-800 text-cyan-400 font-bold">
          {{ rangeLabel }}
        </span>
      </div>
      <div v-if="loading" class="flex items-center gap-1 text-[10px] text-cyan-400">
        <Loader2 class="w-3 h-3 animate-spin" />
        <span>Memuat<span class="hidden sm:inline"> data Prometheus</span>...</span>
      </div>
    </div>

    <div class="flex-1 w-full min-h-[260px] relative">
      <div ref="chartRef" class="h-full w-full min-h-[260px]"></div>
      <div
        v-if="!loading && points.length === 0"
        class="absolute inset-0 flex items-center justify-center text-xs text-slate-500"
      >
        Belum ada data metrik historis untuk rentang ini di Prometheus.
      </div>
    </div>
  </div>
</template>
