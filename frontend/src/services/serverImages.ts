import { ref } from 'vue'

/**
 * Per-server photo of the physical machine.
 *
 * The default comes from servers.json (`image`), but operators can replace it
 * from the UI without a rebuild. There is no upload endpoint — the backend is
 * a read-only telemetry agent — so the replacement is downscaled in the browser
 * and kept in localStorage on the display machine itself.
 */

const PREFIX = 'srvmon:img:'
const MAX_DIM = 900
const QUALITY = 0.82
/** localStorage tops out around 5MB per origin; stay well clear of it. */
const MAX_STORED_BYTES = 1_400_000

function loadAll(): Record<string, string> {
  const out: Record<string, string> = {}
  try {
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i)
      if (key && key.startsWith(PREFIX)) {
        const value = localStorage.getItem(key)
        if (value) out[key.slice(PREFIX.length)] = value
      }
    }
  } catch {
    // Private mode / storage disabled: fall back to registry defaults only.
  }
  return out
}

const overrides = ref<Record<string, string>>(loadAll())

/** Reactive map of serverId -> data URL. Empty when nothing was overridden. */
export function useServerImages() {
  return overrides
}

export function getImageOverride(serverId: string): string | null {
  return overrides.value[serverId] || null
}

/** Resolve the image to actually render: override first, then registry default. */
export function resolveServerImage(serverId: string, registryImage?: string): string | null {
  return overrides.value[serverId] || registryImage || null
}

export function clearImageOverride(serverId: string): void {
  const next = { ...overrides.value }
  delete next[serverId]
  overrides.value = next
  try {
    localStorage.removeItem(PREFIX + serverId)
  } catch {
    // Nothing to do; the in-memory map is already updated.
  }
}

export function setImageOverride(serverId: string, dataUrl: string): void {
  if (dataUrl.length > MAX_STORED_BYTES) {
    throw new Error('Gambar terlalu besar setelah dikompres. Coba gambar dengan resolusi lebih kecil.')
  }
  try {
    localStorage.setItem(PREFIX + serverId, dataUrl)
  } catch {
    throw new Error('Penyimpanan browser penuh. Hapus gambar server lain terlebih dahulu.')
  }
  overrides.value = { ...overrides.value, [serverId]: dataUrl }
}

/**
 * Read an image file and re-encode it to a bounded JPEG data URL.
 * A 4MB phone photo becomes roughly 100-200KB, which localStorage can hold.
 */
export function fileToScaledDataUrl(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    if (!file.type.startsWith('image/')) {
      reject(new Error('File harus berupa gambar (JPG, PNG, atau WEBP).'))
      return
    }

    const reader = new FileReader()

    reader.onerror = () => reject(new Error('Gagal membaca file gambar.'))
    reader.onload = () => {
      const img = new Image()

      img.onerror = () => reject(new Error('File gambar tidak dapat dibaca.'))
      img.onload = () => {
        const scale = Math.min(1, MAX_DIM / Math.max(img.width, img.height))
        const width = Math.max(1, Math.round(img.width * scale))
        const height = Math.max(1, Math.round(img.height * scale))

        const canvas = document.createElement('canvas')
        canvas.width = width
        canvas.height = height

        const ctx = canvas.getContext('2d')
        if (!ctx) {
          reject(new Error('Browser tidak mendukung pemrosesan gambar.'))
          return
        }

        ctx.drawImage(img, 0, 0, width, height)
        resolve(canvas.toDataURL('image/jpeg', QUALITY))
      }

      img.src = String(reader.result)
    }

    reader.readAsDataURL(file)
  })
}
