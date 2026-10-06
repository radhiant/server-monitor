<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useMetricsStore } from '@/stores/metricsStore'
import { Server, ChevronDown, Check } from 'lucide-vue-next'

const store = useMetricsStore()
const isOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

const selectedServerName = computed(() => {
  return store.activeServer?.name || 'Select Server'
})

const selectServer = (id: string) => {
  store.setActiveServerId(id)
  isOpen.value = false
}

const handleClickOutside = (e: MouseEvent) => {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target as Node)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div ref="dropdownRef" class="relative inline-block text-left font-mono">
    <button
      type="button"
      @click="isOpen = !isOpen"
      class="inline-flex items-center gap-1.5 sm:gap-2 px-2.5 py-1 sm:py-1.5 rounded-md bg-dark-900 border border-slate-700/80 hover:border-cyan-500/50 hover:bg-dark-850 text-slate-200 text-xs font-semibold shadow-sm transition-all"
    >
      <Server class="w-3.5 h-3.5 text-cyan-400 shrink-0" />
      <span class="truncate max-w-[100px] min-[400px]:max-w-[140px] sm:max-w-[200px]">{{ selectedServerName }}</span>
      <ChevronDown class="w-3.5 h-3.5 text-slate-400 transition-transform duration-200 shrink-0" :class="{ 'rotate-180': isOpen }" />
    </button>

    <transition
      enter-active-class="transition duration-100 ease-out"
      enter-from-class="transform scale-95 opacity-0"
      enter-to-class="transform scale-100 opacity-100"
      leave-active-class="transition duration-75 ease-in"
      leave-from-class="transform scale-100 opacity-100"
      leave-to-class="transform scale-95 opacity-0"
    >
      <div
        v-if="isOpen"
        class="absolute left-0 mt-1.5 w-60 sm:w-64 max-w-[calc(100vw-1.5rem)] rounded-md bg-dark-900 border border-slate-700 shadow-xl z-50 py-1 divide-y divide-slate-800"
      >
        <div class="px-3 py-1.5 text-[10px] uppercase font-bold text-slate-500 tracking-wider">
          Monitored Servers ({{ store.servers.length }})
        </div>
        <div class="max-h-60 overflow-y-auto py-1">
          <button
            v-for="s in store.servers"
            :key="s.id"
            @click="selectServer(s.id)"
            class="w-full text-left px-3 py-2 flex items-center justify-between hover:bg-slate-800/80 transition-colors group"
            :class="{ 'bg-slate-800/40 text-cyan-400': s.id === store.activeServerId }"
          >
            <div>
              <div class="text-xs font-semibold text-slate-200 group-hover:text-cyan-300">
                {{ s.name }}
              </div>
              <div class="text-[10px] text-slate-400 truncate max-w-[180px]">
                {{ s.description || s.baseUrl }}
              </div>
            </div>
            <Check v-if="s.id === store.activeServerId" class="w-4 h-4 text-cyan-400" />
          </button>
        </div>
      </div>
    </transition>
  </div>
</template>
