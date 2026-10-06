/**
 * Shared display formatters.
 *
 * These used to be copy-pasted into every chart and panel, which let the
 * rounding drift apart between views (0 vs 1 decimal for the same value).
 */

const GB = 1024 * 1024 * 1024
const MB = 1024 * 1024

/** Unix seconds -> "HH:MM:SS" in the viewer's local timezone. */
export function formatTime24(ts: number): string {
  const d = new Date(ts * 1000)
  const h = String(d.getHours()).padStart(2, '0')
  const m = String(d.getMinutes()).padStart(2, '0')
  const s = String(d.getSeconds()).padStart(2, '0')
  return `${h}:${m}:${s}`
}

/** Current wall clock as "HH:MM:SS". */
export function nowTime24(): string {
  return formatTime24(Date.now() / 1000)
}

export function toGB(bytes: number, digits = 1): string {
  return (bytes / GB).toFixed(digits)
}

export function toMB(bytes: number, digits = 2): string {
  return (bytes / MB).toFixed(digits)
}

/** Human-readable byte size that picks its own unit. */
export function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B'
  if (bytes >= GB) return `${(bytes / GB).toFixed(1)} GB`
  if (bytes >= MB) return `${(bytes / MB).toFixed(0)} MB`
  if (bytes >= 1024) return `${(bytes / 1024).toFixed(0)} KB`
  return `${bytes} B`
}

/** Seconds -> "12d 4h 30m" / "4h 30m 12s" / "30m 12s". */
export function formatUptime(seconds: number): string {
  const total = Math.max(0, Math.floor(seconds))
  const days = Math.floor(total / 86400)
  const hours = Math.floor((total % 86400) / 3600)
  const minutes = Math.floor((total % 3600) / 60)
  const secs = total % 60

  if (days > 0) return `${days}d ${hours}h ${minutes}m`
  if (hours > 0) return `${hours}h ${minutes}m ${secs}s`
  return `${minutes}m ${secs}s`
}

/** Compact uptime for dense fleet cards: "12d 4h" / "4h 30m" / "30m". */
export function formatUptimeShort(seconds: number): string {
  const total = Math.max(0, Math.floor(seconds))
  const days = Math.floor(total / 86400)
  const hours = Math.floor((total % 86400) / 3600)
  const minutes = Math.floor((total % 3600) / 60)

  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
}
