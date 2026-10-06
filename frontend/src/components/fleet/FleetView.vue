<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import { useMetricsStore } from '@/stores/metricsStore'
import { useFleetStore } from '@/stores/fleetStore'
import { useFleetPoll } from '@/composables/useFleetPoll'
import FleetCard from './FleetCard.vue'
import FleetHealthPie from '@/components/charts/FleetHealthPie.vue'
import FleetTrendChart from '@/components/charts/FleetTrendChart.vue'
import { worstFilesystem } from '@/config/thresholds'
import { formatTime24 } from '@/utils/format'
import {
  RefreshCw, ShieldCheck, TriangleAlert, Loader2,
  Pause, Play, ArrowUpDown, Clock, Activity,
} from 'lucide-vue-next'
import { prometheusServerApi, resolveServerIp } from '@/services/prometheus'
import { TIME_RANGES, type TimeRangeKey } from '@/stores/metricsStore'

const store = useMetricsStore()
const fleet = useFleetStore()
const { pollAll } = useFleetPoll()

const counts = computed(() => fleet.counts)
const scrollContainerRef = ref<HTMLElement | null>(null)
const mobileTab = ref<'servers' | 'health'>('servers')

// ---------------------------------------------------------------------------
// Prometheus Fleet SLA State & Loader
// ---------------------------------------------------------------------------
const fleetSla = ref<number | null>(null)
const fleetSlas = ref<Record<string, number>>({})
const fleetRange = ref<TimeRangeKey>('24h')
const slaLoading = ref(false)

const loadFleetSla = async () => {
  slaLoading.value = true
  try {
    const [avg, all] = await Promise.all([
      prometheusServerApi.getFleetSla(fleetRange.value),
      prometheusServerApi.getAllFleetSla(fleetRange.value),
    ])
    fleetSla.value = avg
    fleetSlas.value = all
  } catch {
    fleetSla.value = null
  } finally {
    slaLoading.value = false
  }
}

// ---------------------------------------------------------------------------
// Auto-Scroll, Rotation Duration & URL Parameter Configuration
// ---------------------------------------------------------------------------
const DEFAULT_ROTATE_MS = 10000
const DURATION_OPTIONS = [3, 5, 10, 15, 20, 30] // options in seconds (including 3s)

const readUrlParams = () => {
  try {
    const params = new URLSearchParams(window.location.search)
    const rotParam = params.get('rotate')?.toLowerCase() || params.get('scroll')?.toLowerCase() || params.get('interval')?.toLowerCase()

    let rotateMs = DEFAULT_ROTATE_MS
    let isPaused = false

    if (rotParam) {
      if (rotParam === 'false' || rotParam === '0' || rotParam === 'off' || rotParam === 'pause' || rotParam === 'paused') {
        isPaused = true
      } else {
        const num = parseFloat(rotParam)
        if (!isNaN(num) && num > 0) {
          // If <= 120, interpret as seconds; otherwise as milliseconds
          const calculatedMs = num <= 120 ? Math.round(num * 1000) : Math.round(num)
          rotateMs = Math.max(1000, calculatedMs)
        }
      }
    }

    return { rotateMs, isPaused }
  } catch {
    return { rotateMs: DEFAULT_ROTATE_MS, isPaused: false }
  }
}

const initialConfig = readUrlParams()
const rotateMs = ref(initialConfig.rotateMs)

const hoverPaused = ref(false)
const manualPaused = ref(initialConfig.isPaused)
const isPaused = computed(() => hoverPaused.value || manualPaused.value)

let scrollTimer: ReturnType<typeof setInterval> | null = null
let autoScrollDirection = 'down' // 'down' | 'up'

const syncFleetUrl = () => {
  try {
    const url = new URL(window.location.href)
    if (store.view === 'fleet') {
      // Clean up legacy pagination params
      url.searchParams.delete('paginate')
      url.searchParams.delete('page')

      // Sync rotate parameter cleanly
      if (manualPaused.value) {
        url.searchParams.set('rotate', '0')
      } else if (rotateMs.value !== DEFAULT_ROTATE_MS) {
        url.searchParams.set('rotate', String(Math.round(rotateMs.value / 1000)))
      } else {
        url.searchParams.delete('rotate')
      }

      window.history.replaceState({}, '', url.toString())
    }
  } catch {}
}

const stopTimer = () => {
  if (scrollTimer) {
    clearInterval(scrollTimer)
    scrollTimer = null
  }
}

// Perform smooth auto-scroll ping-pong
const performAutoScroll = () => {
  const el = scrollContainerRef.value
  if (!el || isPaused.value) return

  const maxScroll = el.scrollHeight - el.clientHeight
  if (maxScroll <= 10) return // No scrollable overflow needed

  const currentScroll = el.scrollTop

  if (autoScrollDirection === 'down') {
    if (currentScroll >= maxScroll - 20) {
      // Reached bottom: switch direction to scroll back to top next cycle
      autoScrollDirection = 'up'
      el.scrollTo({ top: 0, behavior: 'smooth' })
    } else {
      // Scroll down by 1 viewport
      el.scrollTo({ top: Math.min(maxScroll, currentScroll + (el.clientHeight * 0.75)), behavior: 'smooth' })
      if (currentScroll + (el.clientHeight * 0.75) >= maxScroll - 20) {
        autoScrollDirection = 'up'
      }
    }
  } else {
    // Scroll smoothly back to top
    el.scrollTo({ top: 0, behavior: 'smooth' })
    autoScrollDirection = 'down'
  }
}

const startTimer = () => {
  stopTimer()
  scrollTimer = setInterval(() => {
    if (!isPaused.value) {
      performAutoScroll()
    }
  }, rotateMs.value)
}

const toggleManualPause = () => {
  manualPaused.value = !manualPaused.value
  syncFleetUrl()
  if (!manualPaused.value) {
    startTimer()
  }
}

const setDuration = (sec: number) => {
  rotateMs.value = sec * 1000
  manualPaused.value = false
  syncFleetUrl()
  startTimer()
}

let slaTimer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  startTimer()
  syncFleetUrl()
  loadFleetSla()
  slaTimer = setInterval(loadFleetSla, 60000)
})

onUnmounted(() => {
  stopTimer()
  if (slaTimer) clearInterval(slaTimer)
})

watch(fleetRange, () => {
  loadFleetSla()
})

const lastUpdate = computed(() =>
  fleet.lastPollAt ? formatTime24(fleet.lastPollAt / 1000) : '--:--:--'
)

const SEVERITY_ORDER: Record<string, number> = {
  offline: 0,
  critical: 1,
  warning: 2,
}

/** Anything not healthy, worst first — the short list worth acting on. */
const attention = computed(() => {
  const rows = store.servers
    .map(server => ({ server, status: fleet.statusOf(server.id), entry: fleet.getEntry(server.id) }))
    .filter(row => row.status in SEVERITY_ORDER)
    .sort((a, b) => SEVERITY_ORDER[a.status] - SEVERITY_ORDER[b.status])

  return rows.map(row => {
    const metrics = row.entry?.metrics
    const worstFs = worstFilesystem(metrics?.filesystems)

    let detail = row.entry?.error || 'Tidak dapat dihubungi'
    if (row.status !== 'offline' && metrics) {
      const candidates = [
        { label: 'CPU', value: metrics.cpu?.usage ?? 0 },
        { label: 'RAM', value: metrics.memory?.usage ?? 0 },
        { label: worstFs ? `Disk ${worstFs.mountpoint}` : 'Disk', value: worstFs?.usage ?? 0 },
      ]
      const driver = candidates.reduce((a, b) => (b.value > a.value ? b : a))
      detail = `${driver.label} ${driver.value.toFixed(0)}%`
    }

    return { ...row, detail }
  })
})

const statusStyle = (status: string) => {
  if (status === 'offline') return 'text-rose-400 border-rose-500/40 bg-rose-950/30'
  if (status === 'critical') return 'text-rose-400 border-rose-500/40 bg-rose-950/30'
  return 'text-amber-400 border-amber-500/40 bg-amber-950/30'
}

const statusLabel = (status: string) => {
  if (status === 'offline') return 'OFFLINE'
  if (status === 'critical') return 'KRITIS'
  return 'WARN'
}
</script>

<template>
  <div class="flex-1 min-h-0 p-2 sm:p-2.5 flex flex-col xl:grid xl:grid-cols-12 gap-2">
    <!-- Mobile Alert & Segmented Tabs (Screens < xl) -->
    <div class="xl:hidden flex flex-col gap-2 shrink-0">
      <!-- Attention Banner if any server has warnings/offline -->
      <div
        v-if="attention.length > 0"
        class="bg-amber-950/40 border border-amber-500/40 rounded-lg px-2.5 py-1.5 flex items-center justify-between gap-2 font-mono text-[11px]"
      >
        <div class="flex items-center gap-2 min-w-0">
          <TriangleAlert class="w-3.5 h-3.5 text-amber-400 shrink-0 animate-pulse" />
          <span class="text-amber-200 font-semibold truncate">
            {{ attention.length }} server perlu perhatian:
            <span class="text-amber-400 font-bold">{{ attention[0].server.name }} ({{ attention[0].detail }})</span>
          </span>
        </div>
        <button
          type="button"
          class="px-2 py-0.5 rounded bg-amber-900/60 border border-amber-700/60 text-amber-300 font-bold text-[10px] shrink-0 hover:bg-amber-800/80"
          @click="mobileTab = 'health'"
        >
          Lihat
        </button>
      </div>

      <!-- Mobile Tab Switcher -->
      <div class="flex items-center bg-dark-900/90 border border-slate-800/80 rounded-lg p-1 font-mono text-xs shadow-md">
        <button
          type="button"
          class="flex-1 py-1.5 px-3 rounded-md font-semibold flex items-center justify-center gap-2 transition-all"
          :class="mobileTab === 'servers'
            ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40 shadow-sm'
            : 'text-slate-400 hover:text-slate-200 border border-transparent'"
          @click="mobileTab = 'servers'"
        >
          <Activity class="w-3.5 h-3.5" />
          <span>Server ({{ store.servers.length }})</span>
        </button>
        <button
          type="button"
          class="flex-1 py-1.5 px-3 rounded-md font-semibold flex items-center justify-center gap-2 transition-all relative"
          :class="mobileTab === 'health'
            ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40 shadow-sm'
            : 'text-slate-400 hover:text-slate-200 border border-transparent'"
          @click="mobileTab = 'health'"
        >
          <ShieldCheck v-if="attention.length === 0" class="w-3.5 h-3.5 text-emerald-400" />
          <TriangleAlert v-else class="w-3.5 h-3.5 text-amber-400" />
          <span>Status &amp; SLA</span>
          <span
            v-if="attention.length > 0"
            class="px-1.5 py-0.2 rounded-full text-[9px] font-bold bg-rose-500/90 text-white"
          >
            {{ attention.length }}
          </span>
        </button>
      </div>
    </div>

    <!-- Server grid container -->
    <div
      class="col-span-12 xl:col-span-9 min-h-0 flex-col gap-2"
      :class="mobileTab === 'servers' ? 'flex' : 'hidden xl:flex'"
      @mouseenter="hoverPaused = true"
      @mouseleave="hoverPaused = false"
      @touchstart="hoverPaused = true"
      @touchend="hoverPaused = false"
    >
      <!-- All Servers in Proportional Grid with Smooth Auto-Scroll -->
      <div
        ref="scrollContainerRef"
        class="flex-1 min-h-0 grid gap-2.5 grid-cols-1 sm:grid-cols-2 2xl:grid-cols-3 auto-rows-auto xl:auto-rows-[calc((100%-0.625rem)/2)] overflow-y-auto pr-1 sm:pr-1.5 scroll-smooth"
      >
        <FleetCard
          v-for="server in store.servers"
          :key="server.id"
          :server="server"
          :entry="fleet.getEntry(server.id)"
          :sla="fleetSlas[server.id] ?? fleetSlas[resolveServerIp(server.id, server.description)]"
          @open="store.openServer"
        />

        <div
          v-if="store.servers.length === 0"
          class="col-span-full row-span-full flex items-center justify-center text-sm font-mono text-slate-500 py-12"
        >
          Registry server kosong. Periksa /servers.json
        </div>
      </div>

      <!-- Auto-Scroll Control Bar -->
      <nav
        class="shrink-0 flex flex-wrap items-center justify-between gap-2 px-2.5 sm:px-3 py-1.5 bg-dark-900/80 border border-slate-800/80 rounded-lg font-mono text-[11px] text-slate-400 shadow-md backdrop-blur-sm"
        aria-label="Kontrol Auto-Scroll dan Fleet"
      >
        <!-- Left: Active Node Count & Status -->
        <div class="flex items-center gap-2 text-[10px] sm:text-[11px]">
          <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse shrink-0"></span>
          <span class="font-bold text-slate-200">FLEET</span>
          <span class="text-slate-500 hidden sm:inline">|</span>
          <span class="text-cyan-400 font-semibold truncate">{{ store.servers.length }} Server Aktif</span>
        </div>

        <!-- Center: Auto-Scroll Play/Pause & Duration Selector -->
        <div class="flex items-center gap-2 sm:gap-3 flex-wrap">
          <!-- Play / Pause Toggle Button -->
          <button
            type="button"
            class="flex items-center gap-1.5 px-2.5 sm:px-3 py-1 rounded-md border font-semibold transition-all shadow-sm text-[10px] sm:text-[11px]"
            :class="manualPaused
              ? 'text-amber-300 bg-amber-950/50 border-amber-500/50 hover:bg-amber-900/50'
              : 'text-cyan-300 bg-cyan-950/40 border-cyan-500/40 hover:bg-cyan-900/40'"
            :title="manualPaused ? 'Klik untuk melanjutkan auto-scroll' : 'Klik untuk menjeda auto-scroll'"
            @click="toggleManualPause"
          >
            <Play v-if="manualPaused" class="w-3.5 h-3.5 text-amber-400 fill-amber-400/20" />
            <Pause v-else class="w-3.5 h-3.5 text-cyan-400 fill-cyan-400/20" />
            <ArrowUpDown class="w-3 h-3 opacity-70 hidden sm:inline" />
            <span>{{ manualPaused ? 'PAUSED' : 'AUTO-SCROLL' }}</span>
          </button>

          <!-- Duration Quick Picker Pills (with 3s option) -->
          <div class="flex items-center gap-0.5 sm:gap-1 bg-dark-950 p-0.5 rounded-md border border-slate-800 overflow-x-auto max-w-[200px] sm:max-w-none">
            <div class="hidden sm:flex items-center gap-1 px-1.5 text-slate-500 text-[10px]">
              <Clock class="w-3 h-3 text-slate-500" />
              <span>Jeda:</span>
            </div>
            <button
              v-for="sec in DURATION_OPTIONS"
              :key="sec"
              type="button"
              class="px-1.5 sm:px-2 py-0.5 rounded text-[10px] font-semibold transition-all tabular-nums shrink-0"
              :class="Math.round(rotateMs / 1000) === sec && !manualPaused
                ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40 shadow-sm'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'"
              :title="`Set jeda auto-scroll ${sec} detik`"
              @click="setDuration(sec)"
            >
              {{ sec }}s
            </button>
          </div>
        </div>

        <!-- Right: Status / Hint -->
        <div class="text-[10px] text-slate-500 hidden md:flex items-center gap-1.5">
          <span v-if="hoverPaused" class="text-amber-400 font-medium">
            (Dijeda karena kursor aktif di atas kartu)
          </span>
          <span v-else-if="manualPaused" class="text-slate-400">
            Auto-scroll berhenti
          </span>
          <span v-else class="text-slate-400">
            Loop scroll atas-bawah tiap {{ Math.round(rotateMs / 1000) }}s
          </span>
        </div>
      </nav>
    </div>

    <!-- Right rail: aggregate health + what needs attention -->
    <aside
      class="col-span-12 xl:col-span-3 flex-col gap-2 min-h-0"
      :class="mobileTab === 'health' ? 'flex' : 'hidden xl:flex'"
    >
      <!-- Prometheus Fleet SLA Card -->
      <section
        class="bg-dark-900/90 border border-slate-800/80 rounded-lg p-2.5 shadow-lg flex flex-col min-h-0 shrink-0 font-mono"
      >
        <div class="flex items-center justify-between gap-1 mb-1.5 shrink-0">
          <div class="flex items-center gap-1.5 text-slate-300">
            <Activity class="w-3.5 h-3.5 text-cyan-400" />
            <span class="text-[11px] font-semibold uppercase tracking-wider">
              SLA FLEET
            </span>
          </div>

          <!-- Range Pills: 24h, 7d, 30d, 90d, 1th -->
          <div class="flex items-center gap-0.5 bg-dark-950 p-0.5 rounded border border-slate-800 shrink-0">
            <button
              v-for="r in TIME_RANGES"
              :key="r.key"
              type="button"
              class="px-1.5 py-0.5 text-[9px] font-bold rounded transition-colors"
              :class="fleetRange === r.key
                ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40'
                : 'text-slate-500 hover:text-slate-300'"
              :title="`Pilih rentang SLA ${r.label}`"
              @click="fleetRange = r.key"
            >
              {{ r.key }}
            </button>
          </div>
        </div>

        <div class="flex items-baseline justify-between my-1">
          <span
            class="text-2xl font-bold tracking-tight tabular-nums"
            :class="slaLoading ? 'text-slate-600' : fleetSla !== null && fleetSla >= 99.0 ? 'text-emerald-400' : fleetSla !== null && fleetSla >= 95.0 ? 'text-cyan-400' : 'text-amber-400'"
          >
            {{ slaLoading ? '--' : fleetSla !== null ? `${fleetSla.toFixed(2)}%` : '--' }}
          </span>
          <span class="text-[10px] text-slate-500">
            Target 99.9% &middot; TSDB 3th
          </span>
        </div>

        <div class="w-full bg-dark-950 rounded-full h-1.5 overflow-hidden">
          <div
            v-if="fleetSla !== null && !slaLoading"
            class="h-full rounded-full transition-all duration-300"
            :class="fleetSla >= 99.0 ? 'bg-emerald-500' : fleetSla >= 95.0 ? 'bg-cyan-500' : 'bg-amber-500'"
            :style="{ width: `${Math.min(100, Math.max(0, fleetSla))}%` }"
          ></div>
        </div>
      </section>

      <section
        class="bg-dark-900/90 border border-slate-800/80 rounded-lg p-3 shadow-lg flex flex-col min-h-0 shrink-0"
      >
        <div class="flex items-center justify-between mb-1 shrink-0">
          <span class="text-[11px] font-mono font-semibold uppercase tracking-wider text-slate-400">
            Status Fleet
          </span>
          <button
            type="button"
            class="flex items-center gap-1 text-[10px] font-mono text-slate-500 hover:text-cyan-300 transition-colors"
            title="Perbarui sekarang"
            @click="pollAll"
          >
            <Loader2 v-if="fleet.isPolling" class="w-3 h-3 animate-spin text-cyan-400" />
            <RefreshCw v-else class="w-3 h-3" />
            <span class="tabular-nums">{{ lastUpdate }}</span>
          </button>
        </div>

        <FleetHealthPie :counts="counts" class="shrink-0" />
      </section>

      <FleetTrendChart :samples="fleet.trend" class="shrink-0 h-[170px]" />

      <section
        class="bg-dark-900/90 border border-slate-800/80 rounded-lg p-3 shadow-lg flex flex-col flex-1 min-h-[110px]"
      >
        <div class="flex items-center gap-1.5 mb-2 shrink-0">
          <TriangleAlert v-if="attention.length" class="w-3.5 h-3.5 text-amber-400" />
          <ShieldCheck v-else class="w-3.5 h-3.5 text-emerald-400" />
          <span class="text-[11px] font-mono font-semibold uppercase tracking-wider text-slate-300">
            Perlu Perhatian ({{ attention.length }})
          </span>
        </div>

        <div class="flex-1 min-h-0 overflow-y-auto space-y-1.5 pr-1">
          <button
            v-for="row in attention"
            :key="row.server.id"
            type="button"
            class="w-full text-left rounded border px-2 py-1.5 font-mono transition-colors hover:brightness-125"
            :class="statusStyle(row.status)"
            @click="store.openServer(row.server.id)"
          >
            <div class="flex items-center justify-between gap-2">
              <span class="text-[11px] font-bold text-slate-100 truncate">{{ row.server.name }}</span>
              <span class="text-[9px] font-bold tracking-wider shrink-0">{{ statusLabel(row.status) }}</span>
            </div>
            <div class="text-[10px] opacity-90 truncate">{{ row.detail }}</div>
          </button>

          <div
            v-if="attention.length === 0"
            class="h-full flex flex-col items-center justify-center gap-2 text-center px-2"
          >
            <ShieldCheck class="w-7 h-7 text-emerald-500/70" />
            <p class="text-[11px] font-mono text-slate-400 leading-relaxed">
              Semua server dalam batas normal
            </p>
          </div>
        </div>
      </section>
    </aside>
  </div>
</template>
