import type { DiskPartition, ServerMetrics } from '@/types/metrics'

export type HealthStatus = 'healthy' | 'warning' | 'critical'

export const THRESHOLDS = {
  cpu: {
    warning: 70,
    critical: 90,
  },
  memory: {
    warning: 70,
    critical: 90,
  },
  disk: {
    warning: 75,
    critical: 90,
  },
  loadRatio: {
    warning: 0.9,
    critical: 1.5,
  }
} as const

export function getStatusFromValue(val: number, warningThresh: number, criticalThresh: number): HealthStatus {
  if (val >= criticalThresh) return 'critical'
  if (val >= warningThresh) return 'warning'
  return 'healthy'
}

/**
 * The single fullest partition.
 *
 * Summing every filesystem into one ratio hides the case that actually takes a
 * server down: a small root partition at 98% next to a huge, near-empty data
 * volume averages out to "healthy". Disk pressure is per-mount, so the worst
 * mount is what the status must be driven by.
 */
export function worstFilesystem(filesystems: DiskPartition[] | undefined): DiskPartition | null {
  if (!filesystems || filesystems.length === 0) return null
  return filesystems.reduce((worst, fs) => (fs.usage > worst.usage ? fs : worst))
}

/** Total bytes across every mounted filesystem, for the capacity subtitle. */
export function filesystemTotals(filesystems: DiskPartition[] | undefined): { used: number; total: number } {
  let used = 0
  let total = 0
  for (const fs of filesystems || []) {
    used += fs.used
    total += fs.total
  }
  return { used, total }
}

const RANK: Record<HealthStatus, number> = { healthy: 0, warning: 1, critical: 2 }

export function worseOf(a: HealthStatus, b: HealthStatus): HealthStatus {
  return RANK[b] > RANK[a] ? b : a
}

/**
 * Overall server health = the worst of CPU, memory and the fullest partition.
 * Used by the fleet grid so one bad subsystem is never averaged away.
 */
export function serverHealth(metrics: ServerMetrics | null | undefined): HealthStatus {
  if (!metrics) return 'healthy'

  let status: HealthStatus = getStatusFromValue(
    metrics.cpu?.usage ?? 0,
    THRESHOLDS.cpu.warning,
    THRESHOLDS.cpu.critical,
  )

  status = worseOf(status, getStatusFromValue(
    metrics.memory?.usage ?? 0,
    THRESHOLDS.memory.warning,
    THRESHOLDS.memory.critical,
  ))

  const worstFs = worstFilesystem(metrics.filesystems)
  if (worstFs) {
    status = worseOf(status, getStatusFromValue(
      worstFs.usage,
      THRESHOLDS.disk.warning,
      THRESHOLDS.disk.critical,
    ))
  }

  return status
}
