<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useMetricsStore } from '@/stores/metricsStore'
import ServerDropdown from '@/components/server-selector/ServerDropdown.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { formatUptime, nowTime24 } from '@/utils/format'
import {
  Clock, Activity, Link as LinkIcon, Check, LayoutGrid,
  Monitor, ScrollText, Maximize2, Minimize2,
} from 'lucide-vue-next'

const store = useMetricsStore()
const copied = ref(false)
const currentTime = ref(nowTime24())
const isFullscreen = ref(false)

let timer: ReturnType<typeof setInterval> | null = null

const syncFullscreen = () => {
  isFullscreen.value = document.fullscreenElement !== null
}

onMounted(() => {
  timer = setInterval(() => {
    currentTime.value = nowTime24()
  }, 1000)
  document.addEventListener('fullscreenchange', syncFullscreen)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
  document.removeEventListener('fullscreenchange', syncFullscreen)
})

const formattedUptime = computed(() =>
  formatUptime(store.latestMetrics?.uptime || store.hostInfo?.uptime || 0)
)

const hostDisplay = computed(() => {
  const h = store.hostInfo
  const s = store.latestMetrics
  const fallbackName = store.activeServer?.name || store.activeServerId || 'Server'

  return {
    hostname: h?.hostname || s?.server_id || fallbackName,
    os: h?.platform ? `${h.platform} ${h.platform_version || ''}`.trim() : (h?.os || 'Linux'),
    kernel: h?.kernel_version ? `${h.kernel_version} (${h.kernel_arch || 'x86_64'})` : ''
  }
})

// Mock data is indistinguishable from live telemetry at a glance. On a wall
// display one stray click would show a fully green fake fleet, so the toggle
// only exists in development builds.
const showMockToggle = import.meta.env.DEV

const toggleMockMode = () => {
  store.isMockMode = !store.isMockMode
}

const copyServerLink = async () => {
  try {
    await navigator.clipboard.writeText(window.location.href)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch {
    // Clipboard unavailable over plain HTTP on some browsers; nothing to show.
  }
}

const toggleFullscreen = async () => {
  try {
    if (document.fullscreenElement) {
      await document.exitFullscreen()
    } else {
      await document.documentElement.requestFullscreen()
    }
  } catch {
    // User denied or the browser blocked it.
  }
}
</script>

<template>
  <header class="bg-dark-900/80 border-b border-slate-800/80 px-3 py-1.5 z-40 backdrop-blur-md shrink-0">
    <div class="max-w-[2560px] mx-auto flex flex-wrap items-center justify-between gap-x-3 gap-y-1.5">
      <!-- Left: brand, view switch, server context -->
      <div class="flex flex-wrap items-center gap-2">
        <button
          type="button"
          class="flex items-center gap-2 pr-2 border-r border-slate-800 hover:opacity-80 transition-opacity"
          title="Kembali ke ikhtisar fleet"
          @click="store.setView('fleet')"
        >
          <Activity class="w-5 h-5 text-cyan-400" />
          <span class="font-mono font-bold tracking-tight text-slate-100 text-sm">
            SERVER<span class="text-cyan-400">MONITOR</span>
          </span>
        </button>

        <!-- Fleet is the home view; the detail page always offers a way back. -->
        <div class="flex items-center gap-1.5">
          <button
            type="button"
            class="flex items-center gap-1.5 px-2.5 py-1 rounded-md border font-mono text-[11px] font-semibold transition-colors"
            :class="store.view === 'fleet'
              ? 'bg-cyan-950/60 border-cyan-500/50 text-cyan-300'
              : 'bg-dark-900 border-slate-700/80 text-slate-400 hover:text-cyan-300 hover:border-cyan-500/50'"
            @click="store.setView('fleet')"
          >
            <LayoutGrid class="w-3.5 h-3.5" />
            <span>FLEET</span>
          </button>

          <!-- Copy link button on Fleet view -->
          <button
            v-if="store.view === 'fleet'"
            type="button"
            class="p-1.5 rounded bg-dark-900 border border-slate-700/80 hover:border-cyan-500/50 text-slate-400 hover:text-cyan-300 transition-colors"
            :aria-label="copied ? 'Link tersalin' : 'Salin link fleet ini'"
            :title="copied ? 'Link tersalin!' : 'Salin link fleet ini untuk monitor NOC'"
            @click="copyServerLink"
          >
            <Check v-if="copied" class="w-3.5 h-3.5 text-emerald-400" />
            <LinkIcon v-else class="w-3.5 h-3.5" />
          </button>
        </div>

        <template v-if="store.view === 'server'">
          <div class="flex items-center gap-1.5">
            <ServerDropdown />

            <button
              type="button"
              class="p-1.5 rounded bg-dark-900 border border-slate-700/80 hover:border-cyan-500/50 text-slate-400 hover:text-cyan-300 transition-colors"
              :aria-label="copied ? 'Link tersalin' : 'Salin link langsung ke server ini'"
              :title="copied ? 'Link tersalin!' : 'Salin link langsung untuk display'"
              @click="copyServerLink"
            >
              <Check v-if="copied" class="w-3.5 h-3.5 text-emerald-400" />
              <LinkIcon v-else class="w-3.5 h-3.5" />
            </button>
          </div>

          <div class="hidden lg:flex items-center gap-1.5 text-[11px] font-mono text-slate-400">
            <span class="px-2 py-0.5 rounded bg-dark-950 border border-slate-800 text-slate-300">
              {{ hostDisplay.hostname }}
            </span>
            <span v-if="hostDisplay.os" class="px-2 py-0.5 rounded bg-dark-950 border border-slate-800">
              {{ hostDisplay.os }}
            </span>
            <span v-if="hostDisplay.kernel" class="px-2 py-0.5 rounded bg-dark-950 border border-slate-800 hidden 2xl:inline">
              {{ hostDisplay.kernel }}
            </span>
          </div>
        </template>

        <div v-else class="font-mono text-[11px] text-slate-500 hidden sm:inline">
          Memantau {{ store.servers.length }} server
        </div>
      </div>

      <!-- Right: uptime, clock, display controls, connection -->
      <div class="flex flex-wrap items-center gap-2 font-mono text-[11px]">
        <div
          v-if="store.view === 'server'"
          class="hidden sm:flex items-center gap-1.5 text-slate-400 bg-dark-950/60 px-2 py-0.5 rounded border border-slate-800/60"
        >
          <Clock class="w-3.5 h-3.5 text-slate-500" />
          <span class="text-slate-500">UPTIME</span>
          <span class="font-semibold text-slate-300 tabular-nums">{{ formattedUptime }}</span>
        </div>

        <div class="hidden sm:flex items-center gap-1.5 text-slate-400 bg-dark-950/60 px-2 py-0.5 rounded border border-slate-800/60">
          <span class="text-slate-500">TIME</span>
          <span class="font-semibold text-slate-300 tracking-wider tabular-nums">{{ currentTime }}</span>
        </div>

        <!-- TV mode locks everything to one screen; scroll mode is for laptops/NOC. Hide on mobile phones. -->
        <button
          type="button"
          class="hidden sm:flex items-center gap-1.5 px-2 py-1 rounded border font-semibold transition-colors"
          :class="store.fitMode === 'tv'
            ? 'bg-cyan-950/50 border-cyan-500/40 text-cyan-300'
            : 'bg-dark-950 border-slate-800 text-slate-400 hover:text-slate-200'"
          :title="store.fitMode === 'tv'
            ? 'Mode TV: seluruh dashboard dikunci dalam satu layar'
            : 'Mode gulir: halaman panjang, cocok untuk layar kecil'"
          @click="store.toggleFitMode()"
        >
          <Monitor v-if="store.fitMode === 'tv'" class="w-3.5 h-3.5" />
          <ScrollText v-else class="w-3.5 h-3.5" />
          <span>{{ store.fitMode === 'tv' ? 'TV' : 'SCROLL' }}</span>
        </button>

        <button
          type="button"
          class="hidden sm:flex p-1.5 rounded bg-dark-950 border border-slate-800 text-slate-400 hover:text-cyan-300 hover:border-cyan-500/50 transition-colors"
          :aria-label="isFullscreen ? 'Keluar dari layar penuh' : 'Layar penuh'"
          :title="isFullscreen ? 'Keluar dari layar penuh' : 'Layar penuh'"
          @click="toggleFullscreen"
        >
          <Minimize2 v-if="isFullscreen" class="w-3.5 h-3.5" />
          <Maximize2 v-else class="w-3.5 h-3.5" />
        </button>

        <button
          v-if="showMockToggle"
          type="button"
          class="px-2 py-1 rounded border text-[10px] font-semibold transition-all"
          :class="store.isMockMode
            ? 'bg-purple-950/60 border-purple-500/50 text-purple-300 hover:bg-purple-900/60'
            : 'bg-dark-950 border-slate-800 text-slate-400 hover:text-slate-200 hover:border-slate-700'"
          title="Hanya tersedia pada build development"
          @click="toggleMockMode"
        >
          {{ store.isMockMode ? 'MOCK ON' : 'MOCK OFF' }}
        </button>

        <StatusBadge
          v-if="store.view === 'server'"
          :state="store.connectionState"
          :is-mock="store.isMockMode"
        />
        <div
          v-else-if="store.isMockMode"
          class="px-2 py-1 rounded-full border border-purple-500/30 bg-purple-950/60 text-purple-300 text-[10px] font-semibold"
        >
          MOCK SIMULATION
        </div>
      </div>
    </div>
  </header>
</template>
