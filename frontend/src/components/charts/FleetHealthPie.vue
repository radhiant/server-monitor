<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useECharts } from '@/composables/useECharts'
import type { EChartsOption } from '@/lib/echarts'

/**
 * Fleet health donut: how the registry splits across healthy / warning /
 * critical / offline right now. Readable from across a room, which is the
 * whole point of the overview screen.
 */
const props = defineProps<{
  counts: {
    healthy: number
    warning: number
    critical: number
    offline: number
    pending: number
    total: number
  }
}>()

const chartRef = ref<HTMLElement | null>(null)

const SEGMENTS = [
  { key: 'healthy', name: 'Sehat', color: '#10b981' },
  { key: 'warning', name: 'Peringatan', color: '#f59e0b' },
  { key: 'critical', name: 'Kritis', color: '#ef4444' },
  { key: 'offline', name: 'Offline', color: '#f43f5e' },
  { key: 'pending', name: 'Memeriksa', color: '#334155' },
] as const

const segments = computed(() =>
  SEGMENTS.map(s => ({ ...s, value: props.counts[s.key] })).filter(s => s.value > 0)
)

const reachable = computed(() => props.counts.healthy + props.counts.warning + props.counts.critical)

const baseOption: EChartsOption = {
  backgroundColor: 'transparent',
  tooltip: {
    trigger: 'item',
    backgroundColor: '#0f172a',
    borderColor: '#334155',
    textStyle: { color: '#e2e8f0', fontFamily: 'monospace', fontSize: 11 },
    formatter: (params: any) => `${params.marker} <b>${params.name}</b><br/>${params.value} server &middot; ${params.percent}%`,
  },
  series: [
    {
      name: 'Status Fleet',
      type: 'pie',
      radius: ['60%', '86%'],
      center: ['50%', '50%'],
      avoidLabelOverlap: false,
      label: { show: false },
      labelLine: { show: false },
      itemStyle: { borderColor: '#0f172a', borderWidth: 3 },
      emphasis: { scaleSize: 5 },
      data: [],
    },
  ],
}

const { setOption } = useECharts(chartRef, baseOption)

watch(
  segments,
  (next) => {
    setOption({
      series: [
        {
          data: next.map(s => ({
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
  <div class="flex flex-col min-h-0 font-mono">
    <!-- Square on wide screens: a donut only gains from height when it gains
         width too, so tying the box to the rail width keeps the ring big
         without leaving slack above and below it. Fixed height on narrow
         layouts, where the rail spans the full page. -->
    <div class="relative w-full shrink-0 h-[200px] xl:h-auto xl:aspect-square">
      <div ref="chartRef" class="h-full w-full"></div>
      <div class="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
        <span class="text-3xl xl:text-4xl font-bold leading-none text-slate-100 tabular-nums">
          {{ reachable }}<span class="text-slate-500">/{{ counts.total }}</span>
        </span>
        <span class="text-[10px] uppercase tracking-wider text-slate-500 mt-1.5">terhubung</span>
      </div>
    </div>

    <div class="shrink-0 grid grid-cols-2 gap-x-2 gap-y-1 mt-2">
      <div
        v-for="segment in SEGMENTS.filter(s => s.key !== 'pending' || counts.pending > 0)"
        :key="segment.key"
        class="flex items-center gap-1.5 text-[11px]"
      >
        <span class="w-2 h-2 rounded-sm shrink-0" :style="{ backgroundColor: segment.color }"></span>
        <span class="text-slate-400 flex-1 truncate">{{ segment.name }}</span>
        <span class="text-slate-100 font-bold tabular-nums">{{ counts[segment.key] }}</span>
      </div>
    </div>
  </div>
</template>
