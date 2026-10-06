<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  title: string
  value: string | number
  unit?: string
  subValue?: string
  percentage?: number
  status?: 'healthy' | 'warning' | 'critical'
  /** Before the first snapshot lands, show placeholders instead of a real-looking 0.0%. */
  loading?: boolean
}>()

const statusColor = computed(() => {
  switch (props.status) {
    case 'critical':
      return {
        bar: 'bg-rose-500',
        text: 'text-rose-400',
        border: 'border-rose-500/30',
        glow: 'shadow-rose-500/5'
      }
    case 'warning':
      return {
        bar: 'bg-amber-500',
        text: 'text-amber-400',
        border: 'border-amber-500/30',
        glow: 'shadow-amber-500/5'
      }
    case 'healthy':
    default:
      return {
        bar: 'bg-cyan-500',
        text: 'text-cyan-400',
        border: 'border-cyan-500/20',
        glow: 'shadow-cyan-500/5'
      }
  }
})
</script>

<template>
  <div
    class="bg-dark-900/90 border rounded-lg px-3 py-2 flex flex-col justify-between shadow-lg backdrop-blur transition-all duration-200"
    :class="loading ? 'border-slate-800/80' : [statusColor.border, statusColor.glow]"
  >
    <div class="flex items-center justify-between gap-2">
      <span class="text-[11px] font-semibold uppercase tracking-wider text-slate-400 font-mono truncate">
        {{ title }}
      </span>
      <div
        v-if="percentage !== undefined && !loading"
        class="text-[11px] font-mono font-bold shrink-0"
        :class="statusColor.text"
      >
        {{ percentage.toFixed(1) }}%
      </div>
    </div>

    <div class="my-1.5 flex items-baseline gap-1">
      <span class="text-2xl font-bold font-mono tracking-tight tabular-nums" :class="loading ? 'text-slate-600' : 'text-slate-100'">
        {{ loading ? '--' : value }}
      </span>
      <span v-if="unit" class="text-[11px] text-slate-400 font-mono">
        {{ unit }}
      </span>
    </div>

    <!-- Track is always rendered so cards without a percentage (network) keep
         the same height as the rest of the row. -->
    <div class="w-full bg-dark-950 rounded-full h-1.5 overflow-hidden">
      <div
        v-if="percentage !== undefined && !loading"
        class="h-full rounded-full transition-all duration-300"
        :class="statusColor.bar"
        :style="{ width: `${Math.min(100, Math.max(0, percentage))}%` }"
      ></div>
    </div>

    <div class="mt-1.5 text-[10px] font-mono truncate" :class="loading ? 'text-slate-600' : 'text-slate-400'">
      {{ loading ? 'menunggu telemetri...' : subValue }}
    </div>
  </div>
</template>
