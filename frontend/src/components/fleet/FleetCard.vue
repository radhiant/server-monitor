<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { ServerRegistryItem } from '@/types/server'
import type { FleetEntry } from '@/stores/fleetStore'
import { serverHealth, worstFilesystem, THRESHOLDS, getStatusFromValue } from '@/config/thresholds'
import { resolveServerImage, useServerImages } from '@/services/serverImages'
import { formatUptimeShort, toGB } from '@/utils/format'
import { Server as ServerIcon, PlugZap, ChevronRight, MapPin } from 'lucide-vue-next'

const props = defineProps<{
  server: ServerRegistryItem
  entry: FleetEntry | null
  sla?: number | null
}>()

const emit = defineEmits<{ (e: 'open', id: string): void }>()

const overrides = useServerImages()

const isOffline = computed(() => props.entry?.state === 'OFFLINE')
const isPending = computed(() => !props.entry || props.entry.state === 'PENDING')
const metrics = computed(() => props.entry?.metrics || null)

const status = computed(() => {
  if (isPending.value) return 'pending'
  if (isOffline.value) return 'offline'
  return serverHealth(metrics.value)
})

/** One palette drives border, dot and label so the card reads as a single signal. */
const theme = computed(() => {
  switch (status.value) {
    case 'critical':
      return { border: 'border-rose-500/60', ring: 'hover:border-rose-400', dot: 'bg-rose-500', text: 'text-rose-400', label: 'KRITIS' }
    case 'warning':
      return { border: 'border-amber-500/50', ring: 'hover:border-amber-400', dot: 'bg-amber-500', text: 'text-amber-400', label: 'PERINGATAN' }
    case 'offline':
      return { border: 'border-rose-600/70', ring: 'hover:border-rose-500', dot: 'bg-rose-600', text: 'text-rose-400', label: 'OFFLINE' }
    case 'pending':
      return { border: 'border-slate-700/70', ring: 'hover:border-slate-600', dot: 'bg-slate-500', text: 'text-slate-400', label: 'MEMERIKSA' }
    default:
      return { border: 'border-emerald-500/30', ring: 'hover:border-emerald-400/70', dot: 'bg-emerald-500', text: 'text-emerald-400', label: 'SEHAT' }
  }
})

const imageSrc = computed(() => {
  void overrides.value
  return resolveServerImage(props.server.id, props.server.image)
})

const loadFailed = ref(false)
watch(imageSrc, () => { loadFailed.value = false })

const showPhoto = computed(() => Boolean(imageSrc.value) && !loadFailed.value)

const barColor = (value: number, warn: number, crit: number) => {
  const level = getStatusFromValue(value, warn, crit)
  if (level === 'critical') return 'bg-rose-500'
  if (level === 'warning') return 'bg-amber-500'
  return 'bg-cyan-500'
}

const cpu = computed(() => metrics.value?.cpu.usage ?? 0)
const mem = computed(() => metrics.value?.memory.usage ?? 0)

/** Disk pressure is per-mount, so the fullest partition is what the card shows. */
const disk = computed(() => {
  const worst = worstFilesystem(metrics.value?.filesystems)
  return worst ? { usage: worst.usage, mount: worst.mountpoint } : null
})

const load1 = computed(() => metrics.value?.load?.['1m'])
const cores = computed(() => metrics.value?.cpu.cores ?? 0)

const bars = computed(() => [
  { key: 'cpu', label: 'CPU', value: cpu.value, color: barColor(cpu.value, THRESHOLDS.cpu.warning, THRESHOLDS.cpu.critical) },
  { key: 'mem', label: 'RAM', value: mem.value, color: barColor(mem.value, THRESHOLDS.memory.warning, THRESHOLDS.memory.critical) },
  {
    key: 'disk',
    label: 'DSK',
    value: disk.value?.usage ?? 0,
    color: barColor(disk.value?.usage ?? 0, THRESHOLDS.disk.warning, THRESHOLDS.disk.critical),
  },
])
</script>

<template>
  <button
    type="button"
    class="text-left bg-dark-900/90 border rounded-lg shadow-lg flex flex-col min-h-[270px] sm:min-h-[260px] xl:min-h-[200px] h-full overflow-hidden transition-all duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-cyan-500/70"
    :class="[theme.border, theme.ring]"
    :aria-label="`Buka detail ${server.name}, status ${theme.label}`"
    @click="emit('open', server.id)"
  >
    <!-- Identity row -->
    <div class="flex items-center gap-2 px-2.5 py-2 border-b border-slate-800/70 shrink-0 bg-dark-900/80">
      <span class="relative flex h-2.5 w-2.5 shrink-0">
        <span
          v-if="status !== 'pending'"
          class="absolute inline-flex h-full w-full rounded-full opacity-60 animate-ping"
          :class="theme.dot"
        ></span>
        <span class="relative inline-flex rounded-full h-2.5 w-2.5" :class="theme.dot"></span>
      </span>

      <div class="min-w-0 flex-1">
        <div class="font-mono text-xs font-bold text-slate-100 truncate">{{ server.name }}</div>
        <div class="font-mono text-[10px] text-slate-500 truncate">
          {{ server.role || server.description }}
        </div>
      </div>

      <span
        v-if="sla !== undefined && sla !== null"
        class="font-mono text-[9px] font-bold px-1.5 py-0.5 rounded bg-dark-950 border shrink-0 tabular-nums"
        :class="sla >= 99.0 ? 'text-emerald-400 border-emerald-500/30' : sla >= 95.0 ? 'text-cyan-400 border-cyan-500/30' : 'text-amber-400 border-amber-500/30'"
        :title="`SLA Uptime Prometheus: ${sla.toFixed(2)}%`"
      >
        {{ sla.toFixed(1) }}%
      </span>

      <span class="font-mono text-[9px] font-bold tracking-wider shrink-0" :class="theme.text">
        {{ theme.label }}
      </span>
      <ChevronRight class="w-3.5 h-3.5 text-slate-600 shrink-0" />
    </div>

    <!-- Photo of the physical machine carries the card, with live figures laid over it -->
    <div class="flex-1 min-h-[190px] sm:min-h-[180px] xl:min-h-[120px] relative bg-dark-950">
      <img
        v-if="showPhoto"
        :src="imageSrc!"
        :alt="`Foto fisik ${server.name}`"
        class="absolute inset-0 w-full h-full object-cover"
        :class="isOffline ? 'grayscale opacity-60' : ''"
        @error="loadFailed = true"
      />
      <div v-else class="absolute inset-0 flex flex-col items-center justify-center gap-1.5 text-slate-700">
        <ServerIcon class="w-10 h-10" />
        <span class="font-mono text-[9px] text-slate-600">belum ada foto perangkat</span>
      </div>

      <!-- Scrim gradient overlay for clear contrast on numbers -->
      <div class="absolute inset-x-0 bottom-0 h-[65%] bg-gradient-to-t from-dark-950 via-dark-950/85 to-transparent pointer-events-none"></div>

      <!-- Location Badge -->
      <div v-if="server.location" class="absolute top-2 left-2 flex items-center gap-1 font-mono text-[9px] text-slate-300 bg-dark-950/80 rounded px-1.5 py-0.5 backdrop-blur-sm border border-slate-800/60 shadow-sm">
        <MapPin class="w-2.5 h-2.5 text-cyan-400 shrink-0" />
        <span class="truncate max-w-[180px]">{{ server.location }}</span>
      </div>

      <!-- Telemetry overlay -->
      <div class="absolute inset-x-0 bottom-0 p-2.5 font-mono">
        <template v-if="isOffline">
          <div class="flex items-center gap-1.5 text-rose-400 text-xs font-bold mb-1">
            <PlugZap class="w-4 h-4 shrink-0" />
            <span>Agent tidak merespons</span>
          </div>
          <div class="text-[10px] text-slate-400 truncate" :title="entry?.error || ''">
            {{ entry?.error || 'Koneksi gagal' }}
          </div>
        </template>

        <template v-else-if="isPending">
          <div class="text-[11px] text-slate-400 mb-1.5">Memeriksa node...</div>
          <div class="space-y-1.5">
            <div v-for="n in 3" :key="n" class="h-1.5 rounded-full bg-slate-800/80 animate-pulse"></div>
          </div>
        </template>

        <template v-else>
          <div class="space-y-1.5">
            <div v-for="bar in bars" :key="bar.key" class="flex items-center gap-2">
              <span class="text-[10px] text-slate-400 w-7 shrink-0">{{ bar.label }}</span>
              <div class="flex-1 h-2 bg-slate-900/80 rounded-full overflow-hidden min-w-0">
                <div
                  class="h-full rounded-full transition-all duration-500"
                  :class="bar.color"
                  :style="{ width: `${Math.min(100, Math.max(0, bar.value))}%` }"
                ></div>
              </div>
              <span class="text-[11px] font-bold text-slate-100 w-9 text-right tabular-nums shrink-0">
                {{ bar.value.toFixed(0) }}%
              </span>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- Context strip -->
    <div
      v-if="!isOffline && !isPending"
      class="shrink-0 px-2.5 py-1.5 border-t border-slate-800/70 flex items-center justify-between font-mono text-[9px] text-slate-400 bg-dark-900/60 gap-2"
    >
      <span class="truncate">up {{ formatUptimeShort(metrics?.uptime ?? 0) }}</span>
      <span class="truncate">load {{ load1 !== undefined ? load1.toFixed(2) : '--' }}/{{ cores }}c</span>
      <span class="truncate">
        {{ metrics?.memory ? toGB(metrics.memory.used, 1) + '/' + toGB(metrics.memory.total, 0) + 'G' : '--' }}
      </span>
      <span v-if="disk" class="truncate text-slate-500">{{ disk.mount }}</span>
    </div>
    <div v-else class="shrink-0 px-2.5 py-1.5 border-t border-slate-800/70 font-mono text-[9px] text-slate-500 bg-dark-900/60 truncate">
      {{ server.location || server.description }}
    </div>
  </button>
</template>
