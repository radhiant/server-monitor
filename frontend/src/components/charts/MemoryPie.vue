<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useECharts } from '@/composables/useECharts'
import { toGB } from '@/utils/format'
import type { MemoryMetrics } from '@/types/metrics'
import type { EChartsOption } from '@/lib/echarts'

/**
 * Memory composition donut.
 *
 * A single "RAM 62%" figure is the number operators most often misread on
 * Linux: page cache counts toward it but is reclaimable on demand. Splitting
 * the same total into application / cache / buffers / free is genuinely
 * compositional data — parts of one whole — which is what a pie is for.
 */
const props = defineProps<{
  memory: MemoryMetrics | null | undefined
}>()

const chartRef = ref<HTMLElement | null>(null)

const SLICE_COLORS = {
  used: '#a855f7',
  cached: '#06b6d4',
  buffers: '#3b82f6',
  free: '#1e293b',
} as const

const breakdown = computed(() => {
  const m = props.memory
  if (!m || !m.total) return null

  const used = Math.max(0, m.used)
  const cached = Math.max(0, m.cached)
  const buffers = Math.max(0, m.buffers)
  // Derive free from the other three rather than trusting the reported field,
  // so the slices always sum to exactly `total` on every platform.
  const free = Math.max(0, m.total - used - cached - buffers)

  return {
    total: m.total,
    usage: m.usage,
    slices: [
      { key: 'used', name: 'Aplikasi', value: used, color: SLICE_COLORS.used },
      { key: 'cached', name: 'Cache', value: cached, color: SLICE_COLORS.cached },
      { key: 'buffers', name: 'Buffers', value: buffers, color: SLICE_COLORS.buffers },
      { key: 'free', name: 'Bebas', value: free, color: SLICE_COLORS.free },
    ],
  }
})

const baseOption: EChartsOption = {
  backgroundColor: 'transparent',
  tooltip: {
    trigger: 'item',
    backgroundColor: '#0f172a',
    borderColor: '#334155',
    textStyle: { color: '#e2e8f0', fontFamily: 'monospace', fontSize: 11 },
    formatter: (params: any) => {
      const gb = (params.value / (1024 * 1024 * 1024)).toFixed(2)
      return `${params.marker} <b>${params.name}</b><br/>${gb} GB &middot; ${params.percent}%`
    },
  },
  series: [
    {
      name: 'Memory',
      type: 'pie',
      radius: ['62%', '88%'],
      center: ['50%', '50%'],
      avoidLabelOverlap: false,
      label: { show: false },
      labelLine: { show: false },
      itemStyle: {
        borderColor: '#0f172a',
        borderWidth: 2,
      },
      emphasis: {
        scaleSize: 4,
        itemStyle: { borderWidth: 2 },
      },
      data: [],
    },
  ],
}

const { setOption } = useECharts(chartRef, baseOption)

watch(
  breakdown,
  (next) => {
    if (!next) return
    setOption({
      series: [
        {
          data: next.slices.map(s => ({
            name: s.name,
            value: s.value,
            itemStyle: { color: s.color },
          })),
        },
      ],
    })
  },
  { immediate: true },
)
</script>

<template>
  <div class="bg-dark-900/90 border border-slate-800/80 rounded-lg p-2.5 shadow-lg flex flex-col min-h-0 font-mono">
    <div class="flex items-center justify-between shrink-0 mb-1">
      <span class="text-[11px] font-semibold uppercase tracking-wider text-slate-400">
        Komposisi Memory
      </span>
      <span class="text-[11px] font-bold text-purple-400">
        {{ breakdown ? toGB(breakdown.total, 1) + ' GB' : '--' }}
      </span>
    </div>

    <div class="h-[160px] sm:h-[170px] xl:h-auto xl:flex-1 flex items-center gap-2 min-h-[150px] xl:min-h-0">
      <!-- Donut with the headline figure held in the ring -->
      <div class="relative h-full flex-1 min-w-0 min-h-[140px]">
        <div ref="chartRef" class="h-full w-full min-h-[140px]"></div>
        <div class="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
          <span class="text-lg font-bold leading-none text-slate-100">
            {{ breakdown ? breakdown.usage.toFixed(0) : '--' }}<span class="text-[11px] text-slate-400">%</span>
          </span>
          <span class="text-[9px] uppercase tracking-wider text-slate-500 mt-0.5">terpakai</span>
        </div>
      </div>

      <!-- Value legend: the chart answers "what shape", this answers "how much" -->
      <div class="shrink-0 space-y-1 pr-1">
        <div
          v-for="slice in breakdown?.slices || []"
          :key="slice.key"
          class="flex items-center gap-1.5 text-[10px] leading-tight"
        >
          <span class="w-2 h-2 rounded-sm shrink-0" :style="{ backgroundColor: slice.color }"></span>
          <span class="text-slate-400 w-14">{{ slice.name }}</span>
          <span class="text-slate-200 font-semibold tabular-nums">{{ toGB(slice.value, 1) }}G</span>
        </div>
        <div v-if="!breakdown" class="text-[10px] text-slate-500">Menunggu data...</div>
      </div>
    </div>
  </div>
</template>
