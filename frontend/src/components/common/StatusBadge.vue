<script setup lang="ts">
import { computed } from 'vue'
import type { ConnectionState } from '@/types/metrics'

const props = defineProps<{
  state: ConnectionState
  isMock?: boolean
}>()

const badgeConfig = computed(() => {
  if (props.isMock) {
    return {
      label: 'MOCK SIMULATION',
      dotClass: 'bg-purple-400 animate-pulse',
      containerClass: 'bg-purple-950/60 border-purple-500/30 text-purple-300'
    }
  }

  switch (props.state) {
    case 'CONNECTED':
      return {
        label: 'LIVE',
        dotClass: 'bg-emerald-400 animate-pulse',
        containerClass: 'bg-emerald-950/60 border-emerald-500/30 text-emerald-300 shadow-sm shadow-emerald-500/10'
      }
    case 'CONNECTING':
      return {
        label: 'CONNECTING...',
        dotClass: 'bg-cyan-400 animate-ping',
        containerClass: 'bg-cyan-950/60 border-cyan-500/30 text-cyan-300'
      }
    case 'RECONNECTING':
      return {
        label: 'RECONNECTING...',
        dotClass: 'bg-amber-400 animate-bounce',
        containerClass: 'bg-amber-950/60 border-amber-500/30 text-amber-300'
      }
    case 'DISCONNECTED':
    case 'ERROR':
    default:
      return {
        label: 'OFFLINE',
        dotClass: 'bg-rose-500',
        containerClass: 'bg-rose-950/60 border-rose-500/30 text-rose-300 shadow-sm shadow-rose-500/10'
      }
  }
})
</script>

<template>
  <div
    class="inline-flex items-center gap-2 px-2.5 py-1 rounded-full border text-xs font-mono font-medium tracking-wide transition-colors"
    :class="badgeConfig.containerClass"
  >
    <span class="relative flex h-2 w-2">
      <span
        class="absolute inline-flex h-full w-full rounded-full opacity-75"
        :class="badgeConfig.dotClass"
      ></span>
      <span
        class="relative inline-flex rounded-full h-2 w-2"
        :class="badgeConfig.dotClass"
      ></span>
    </span>
    <span>{{ badgeConfig.label }}</span>
  </div>
</template>
