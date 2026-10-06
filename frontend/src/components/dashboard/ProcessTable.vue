<script setup lang="ts">
import { ref, computed } from 'vue'
import { useMetricsStore } from '@/stores/metricsStore'
import { formatBytes } from '@/utils/format'
import { Terminal, Search } from 'lucide-vue-next'

const store = useMetricsStore()
const searchQuery = ref('')
const sortBy = ref<'cpu' | 'memory'>('cpu')
const sortOrder = ref<'desc' | 'asc'>('desc')

const toggleSort = (type: 'cpu' | 'memory') => {
  if (sortBy.value === type) {
    sortOrder.value = sortOrder.value === 'desc' ? 'asc' : 'desc'
  } else {
    sortBy.value = type
    sortOrder.value = 'desc'
  }
}

const sortIndicator = (type: 'cpu' | 'memory') => {
  if (sortBy.value !== type) return ''
  return sortOrder.value === 'desc' ? '↓' : '↑'
}

const processes = computed(() => {
  const list = store.latestMetrics?.processes || []
  let filtered = list

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    filtered = filtered.filter(p =>
      p.name.toLowerCase().includes(q) ||
      p.pid.toString().includes(q) ||
      p.username.toLowerCase().includes(q)
    )
  }

  return [...filtered].sort((a, b) => {
    const valA = sortBy.value === 'cpu' ? a.cpu_percent : a.memory_percent
    const valB = sortBy.value === 'cpu' ? b.cpu_percent : b.memory_percent
    return sortOrder.value === 'desc' ? valB - valA : valA - valB
  })
})

const cpuColor = (value: number) => {
  if (value >= 50) return 'text-rose-400'
  if (value >= 20) return 'text-amber-400'
  return 'text-cyan-400'
}
</script>

<template>
  <div class="bg-dark-900/90 border border-slate-800/80 rounded-lg p-2 shadow-lg flex flex-col font-mono min-h-0">
    <!-- Toolbar -->
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2 mb-1.5 pb-1.5 border-b border-slate-800 shrink-0">
      <div class="flex items-center gap-2">
        <Terminal class="w-3.5 h-3.5 text-cyan-400" />
        <span class="text-[11px] font-bold uppercase tracking-wider text-slate-200">
          Top Processes ({{ processes.length }})
        </span>
      </div>

      <div class="flex items-center gap-2 w-full sm:w-auto">
        <div class="relative flex-1 sm:w-48">
          <Search class="w-3.5 h-3.5 text-slate-500 absolute left-2.5 top-1/2 -translate-y-1/2" />
          <input
            v-model="searchQuery"
            type="text"
            aria-label="Cari proses berdasarkan PID, nama, atau user"
            placeholder="Cari PID, nama..."
            class="w-full bg-dark-950 border border-slate-700/80 rounded px-2.5 py-1 pl-8 text-[11px] text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500/50"
          />
        </div>

        <div class="flex items-center gap-1 bg-dark-950 border border-slate-800 rounded p-0.5 text-[10px]">
          <button
            type="button"
            class="px-2 py-0.5 rounded transition-colors"
            :class="sortBy === 'cpu' ? 'bg-cyan-950 text-cyan-300 font-bold' : 'text-slate-400 hover:text-slate-200'"
            @click="toggleSort('cpu')"
          >
            CPU {{ sortIndicator('cpu') }}
          </button>
          <button
            type="button"
            class="px-2 py-0.5 rounded transition-colors"
            :class="sortBy === 'memory' ? 'bg-purple-950 text-purple-300 font-bold' : 'text-slate-400 hover:text-slate-200'"
            @click="toggleSort('memory')"
          >
            MEM {{ sortIndicator('memory') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Rows scroll inside the panel so the page itself never grows. -->
    <div class="flex-1 min-h-0 overflow-auto">
      <table class="w-full text-left text-[11px] border-collapse">
        <thead class="sticky top-0 bg-dark-900 text-slate-400 text-[10px] uppercase border-b border-slate-800 z-10">
          <tr>
            <th scope="col" class="py-1 px-1.5 sm:px-2 w-12 sm:w-14 font-semibold">PID</th>
            <th scope="col" class="py-1 px-1.5 sm:px-2 font-semibold">Process</th>
            <th scope="col" class="py-1 px-1.5 sm:px-2 w-16 sm:w-20 text-right font-semibold">
              <!-- Sorting is reachable by keyboard here, not only from the toolbar. -->
              <button type="button" class="hover:text-cyan-300 transition-colors" @click="toggleSort('cpu')">
                CPU % {{ sortIndicator('cpu') }}
              </button>
            </th>
            <th scope="col" class="py-1 px-1.5 sm:px-2 w-16 sm:w-20 text-right font-semibold">
              <button type="button" class="hover:text-purple-300 transition-colors" @click="toggleSort('memory')">
                MEM % {{ sortIndicator('memory') }}
              </button>
            </th>
            <th scope="col" class="hidden md:table-cell py-1 px-2 w-20 text-right font-semibold">RSS</th>
            <th scope="col" class="hidden lg:table-cell py-1 px-2 w-12 text-center font-semibold">THR</th>
            <th scope="col" class="hidden sm:table-cell py-1 px-2 w-20 font-semibold">USER</th>
            <th scope="col" class="hidden sm:table-cell py-1 px-2 w-16 text-center font-semibold">STAT</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-800/50">
          <tr
            v-for="p in processes"
            :key="p.pid"
            class="hover:bg-slate-800/40 transition-colors"
            :class="{ 'bg-rose-950/20': p.cpu_percent >= 50 }"
          >
            <td class="py-1 px-1.5 sm:px-2 text-slate-500 tabular-nums">{{ p.pid }}</td>
            <td class="py-1 px-1.5 sm:px-2 font-semibold text-slate-200 truncate max-w-[120px] min-[400px]:max-w-[160px] sm:max-w-[220px]" :title="p.name">
              {{ p.name }}
            </td>
            <td class="py-1 px-1.5 sm:px-2 text-right font-bold tabular-nums" :class="cpuColor(p.cpu_percent)">
              {{ p.cpu_percent.toFixed(1) }}
            </td>
            <td class="py-1 px-1.5 sm:px-2 text-right font-bold text-purple-400 tabular-nums">
              {{ p.memory_percent.toFixed(1) }}
            </td>
            <td class="hidden md:table-cell py-1 px-2 text-right text-slate-300 tabular-nums">
              {{ formatBytes(p.memory_rss) }}
            </td>
            <td class="hidden lg:table-cell py-1 px-2 text-center text-slate-500 tabular-nums">
              {{ p.num_threads }}
            </td>
            <td class="hidden sm:table-cell py-1 px-2 text-slate-400 truncate max-w-[80px]">
              {{ p.username || 'root' }}
            </td>
            <td class="hidden sm:table-cell py-1 px-2 text-center">
              <span
                class="px-1.5 py-0.5 rounded text-[9px] font-bold uppercase"
                :class="p.status === 'running' || p.status === 'R'
                  ? 'bg-emerald-950 text-emerald-400 border border-emerald-800/50'
                  : 'bg-slate-800 text-slate-400'"
              >
                {{ p.status || 'R' }}
              </span>
            </td>
          </tr>
          <tr v-if="processes.length === 0">
            <td colspan="8" class="text-center py-6 text-slate-500">
              {{ store.hasData ? 'Tidak ada proses yang cocok' : 'Menunggu telemetri proses...' }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
