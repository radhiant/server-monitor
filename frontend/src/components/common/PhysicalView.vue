<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  resolveServerImage,
  getImageOverride,
  setImageOverride,
  clearImageOverride,
  fileToScaledDataUrl,
  useServerImages,
} from '@/services/serverImages'
import { ImagePlus, RotateCcw, Server as ServerIcon, MapPin, Loader2 } from 'lucide-vue-next'
import type { ServerRegistryItem } from '@/types/server'

/**
 * Photo of the physical machine behind the telemetry.
 *
 * Operators can drop in their own picture (rack shot, front panel, label) so
 * the person reading the dashboard knows which box to walk to. The default
 * comes from servers.json; replacements live in localStorage on the display.
 */
const props = defineProps<{
  server: ServerRegistryItem | null
}>()

// Touch the reactive override map so the img re-renders after a change.
const overrides = useServerImages()

const fileInput = ref<HTMLInputElement | null>(null)
const isDragging = ref(false)
const isProcessing = ref(false)
const errorMessage = ref('')

const imageSrc = computed(() => {
  if (!props.server) return null
  void overrides.value
  return resolveServerImage(props.server.id, props.server.image)
})

const hasOverride = computed(() => {
  if (!props.server) return false
  void overrides.value
  return getImageOverride(props.server.id) !== null
})

const applyFile = async (file: File | undefined) => {
  if (!file || !props.server) return

  errorMessage.value = ''
  isProcessing.value = true
  try {
    const dataUrl = await fileToScaledDataUrl(file)
    setImageOverride(props.server.id, dataUrl)
  } catch (err) {
    errorMessage.value = err instanceof Error ? err.message : 'Gagal memproses gambar.'
  } finally {
    isProcessing.value = false
  }
}

const onPick = (event: Event) => {
  const input = event.target as HTMLInputElement
  void applyFile(input.files?.[0])
  // Reset so picking the same file twice still fires a change event.
  input.value = ''
}

const onDrop = (event: DragEvent) => {
  isDragging.value = false
  void applyFile(event.dataTransfer?.files?.[0])
}

const onReset = () => {
  if (!props.server) return
  errorMessage.value = ''
  clearImageOverride(props.server.id)
}

// A broken path in servers.json should fall back to the placeholder, not a
// browser's torn-image icon on a wall display.
const loadFailed = ref(false)
const onImgError = () => { loadFailed.value = true }

/**
 * The frame follows the photo, not the other way round.
 *
 * A fixed 16:9 frame left a square photo floating between two wide bands of
 * empty panel. Reading the real aspect on load and shaping the frame to match
 * means the image fills it edge to edge whatever proportions it was shot in.
 *
 * Clamped so an extreme panorama or a tall portrait cannot squeeze the memory
 * donut below a usable height; inside the clamp the fit is exact.
 */
const MIN_ASPECT = 0.95
const MAX_ASPECT = 2.4
const DEFAULT_ASPECT = 4 / 3

const photoAspect = ref(DEFAULT_ASPECT)

const onImgLoad = (event: Event) => {
  const img = event.target as HTMLImageElement
  if (!img.naturalWidth || !img.naturalHeight) return
  const ratio = img.naturalWidth / img.naturalHeight
  photoAspect.value = Math.min(MAX_ASPECT, Math.max(MIN_ASPECT, ratio))
}

// Clear the failure flag whenever the source changes, otherwise one 404 would
// suppress the photo for every server visited afterwards. The frame goes back
// to its neutral shape until the next image reports its own.
watch(imageSrc, () => {
  loadFailed.value = false
  photoAspect.value = DEFAULT_ASPECT
})
</script>

<template>
  <div
    class="relative bg-dark-900/90 border rounded-lg overflow-hidden shadow-lg flex flex-col min-h-0 group transition-colors"
    :class="isDragging ? 'border-cyan-500/70' : 'border-slate-800/80'"
    @dragover.prevent="isDragging = true"
    @dragleave.prevent="isDragging = false"
    @drop.prevent="onDrop"
  >
    <div class="flex items-center justify-between px-2.5 pt-2 pb-1 shrink-0">
      <span class="text-[11px] font-mono font-semibold uppercase tracking-wider text-slate-400">
        Perangkat Fisik
      </span>
      <div class="flex items-center gap-1">
        <button
          v-if="hasOverride"
          type="button"
          aria-label="Kembalikan ke gambar bawaan"
          title="Kembalikan ke gambar bawaan"
          class="p-1 rounded text-slate-500 hover:text-amber-300 hover:bg-slate-800/80 transition-colors"
          @click="onReset"
        >
          <RotateCcw class="w-3.5 h-3.5" />
        </button>
        <button
          type="button"
          aria-label="Ganti gambar perangkat"
          title="Ganti gambar perangkat"
          class="p-1 rounded text-slate-500 hover:text-cyan-300 hover:bg-slate-800/80 transition-colors"
          @click="fileInput?.click()"
        >
          <ImagePlus class="w-3.5 h-3.5" />
        </button>
      </div>
    </div>

    <!-- Frame takes its shape from the image, so a square photo fills the width
         instead of sitting between two empty bands. -->
    <div
      class="relative mx-2 mb-2 rounded overflow-hidden bg-dark-950 border border-slate-800/60 max-h-full"
      :style="{ aspectRatio: String(photoAspect) }"
    >
      <img
        v-if="imageSrc && !loadFailed"
        :src="imageSrc"
        :alt="`Foto fisik ${server?.name || 'server'}`"
        class="absolute inset-0 w-full h-full object-contain"
        @load="onImgLoad"
        @error="onImgError"
      />

      <!-- Placeholder doubles as the drop target hint -->
      <button
        v-else
        type="button"
        class="absolute inset-0 w-full h-full flex flex-col items-center justify-center gap-1.5 text-slate-600 hover:text-cyan-400 transition-colors"
        @click="fileInput?.click()"
      >
        <ServerIcon class="w-7 h-7" />
        <span class="text-[10px] font-mono text-center px-2 leading-tight">
          Klik atau jatuhkan gambar<br />untuk mengatur foto perangkat
        </span>
      </button>

      <div
        v-if="isProcessing"
        class="absolute inset-0 bg-dark-950/80 flex items-center justify-center"
      >
        <Loader2 class="w-5 h-5 text-cyan-400 animate-spin" />
      </div>

      <!-- Location strip, legible over any photo -->
      <div
        v-if="server?.location"
        class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-dark-950/95 to-transparent px-2 pt-5 pb-1.5 pointer-events-none"
      >
        <div class="flex items-center gap-1 text-[10px] font-mono text-slate-300 truncate">
          <MapPin class="w-3 h-3 text-cyan-400 shrink-0" />
          <span class="truncate">{{ server.location }}</span>
        </div>
      </div>

      <div
        v-if="isDragging"
        class="absolute inset-0 bg-cyan-500/10 border-2 border-dashed border-cyan-400/70 rounded flex items-center justify-center"
      >
        <span class="text-[11px] font-mono text-cyan-300">Lepaskan untuk mengunggah</span>
      </div>
    </div>

    <p v-if="errorMessage" class="px-2.5 pb-1.5 text-[10px] font-mono text-rose-400 leading-tight">
      {{ errorMessage }}
    </p>

    <input
      ref="fileInput"
      type="file"
      accept="image/*"
      class="hidden"
      @change="onPick"
    />
  </div>
</template>
