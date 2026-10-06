<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  cores: number[]
}>()

function getCoreColor(usage: number): string {
  if (usage >= 90) return 'bg-rose-500 text-rose-400'
  if (usage >= 70) return 'bg-amber-500 text-amber-400'
  return 'bg-emerald-500 text-emerald-400'
}
</script>

<template>
  <div class="grid grid-cols-4 sm:grid-cols-8 gap-1.5 font-mono">
    <div
      v-for="(usage, index) in cores"
      :key="index"
      class="bg-dark-950/80 border border-slate-800 rounded p-1.5 flex flex-col justify-between"
    >
      <div class="flex items-center justify-between text-[10px] text-slate-400 mb-1">
        <span>C{{ index }}</span>
        <span :class="getCoreColor(usage).split(' ')[1]">{{ usage.toFixed(0) }}%</span>
      </div>
      <div class="w-full bg-slate-900 rounded-sm h-1 overflow-hidden">
        <div
          class="h-full rounded-sm transition-all duration-300"
          :class="getCoreColor(usage).split(' ')[0]"
          :style="{ width: `${Math.min(100, Math.max(0, usage))}%` }"
        ></div>
      </div>
    </div>
  </div>
</template>
