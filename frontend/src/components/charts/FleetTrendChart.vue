<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useECharts } from '@/composables/useECharts'
import { formatTime24 } from '@/utils/format'
import type { FleetTrendSample } from '@/stores/fleetStore'
import type { EChartsOption } from '@/lib/echarts'

/**
 * Fleet-wide CPU and memory averages over the recent poll history.
 *
 * The cards and the health donut are both instantaneous — they show what the
 * fleet looks like this second and nothing about where it is heading. A steady
 * 60% and a 60% climbing out of 30% read identically there; here they do not.
 */
const props = defineProps<{
  samples: FleetTrendSample[]
}>()

const chartRef = ref<HTMLElement | null>(null)

const latest = computed(() => props.samples[props.samples.length - 1] || null)

const baseOption: EChartsOption = {
  backgroundColor: 'transparent',
  grid: {
    top: 8,
    right: 8,
    bottom: 6,
    left: 30,
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
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: [],
    axisLine: { lineStyle: { color: '#1e293b' } },
    axisTick: { show: false },
    axisLabel: { show: false },
  },
  yAxis: {
    type: 'value',
    min: 0,
    max: 100,
    interval: 50,
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
  () => props.samples,
  (samples) => {
    if (!samples || samples.length === 0) return
    setOption({
      xAxis: { data: samples.map(s => formatTime24(s.timestamp)) },
      series: [
        { name: 'CPU', data: samples.map(s => +s.cpu.toFixed(1)) },
        { name: 'RAM', data: samples.map(s => +s.memory.toFixed(1)) },
      ],
    })
  },
  { immediate: true },
)
</script>

<template>
  <section class="bg-dark-900/90 border border-slate-800/80 rounded-lg p-3 shadow-lg flex flex-col min-h-0 font-mono">
    <div class="flex items-center justify-between shrink-0 mb-1">
      <span class="text-[11px] font-semibold uppercase tracking-wider text-slate-400">
        Tren Rata-rata Fleet
      </span>
      <div class="flex items-center gap-2 text-[10px] font-bold">
        <span class="text-cyan-400">CPU {{ latest ? latest.cpu.toFixed(0) : '--' }}%</span>
        <span class="text-purple-400">RAM {{ latest ? latest.memory.toFixed(0) : '--' }}%</span>
      </div>
    </div>

    <div class="flex-1 min-h-0 relative">
      <div ref="chartRef" class="h-full w-full"></div>
      <div
        v-if="samples.length < 2"
        class="absolute inset-0 flex items-center justify-center text-[10px] text-slate-500"
      >
        Mengumpulkan sampel...
      </div>
    </div>

    <div class="shrink-0 text-[9px] text-slate-600 mt-1">
      Rata-rata {{ latest ? latest.online : 0 }} node terhubung &middot; 5 menit terakhir
    </div>
  </section>
</template>
