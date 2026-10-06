import { defineStore } from 'pinia'
import { shallowRef, computed } from 'vue'
import type { ServerMetrics } from '@/types/metrics'
import { serverHealth, type HealthStatus } from '@/config/thresholds'

export type FleetNodeState = 'ONLINE' | 'OFFLINE' | 'PENDING'

export interface FleetEntry {
  state: FleetNodeState
  metrics: ServerMetrics | null
  error: string | null
  updatedAt: number
}

/**
 * Snapshot of every server in the registry, refreshed by a REST poll.
 *
 * Kept separate from `metricsStore` (which owns the deep 1Hz stream of the one
 * server being inspected) because the fleet only needs a coarse heartbeat.
 */
export interface FleetTrendSample {
  timestamp: number
  cpu: number
  memory: number
  online: number
}

/** 60 polls at 5s ≈ the last five minutes. */
const TREND_LIMIT = 60

export const useFleetStore = defineStore('fleet', () => {
  // shallowRef + whole-map replacement: entries hold full metric snapshots and
  // deep-proxying every one of them on each poll is pure overhead.
  const entries = shallowRef<Record<string, FleetEntry>>({})
  const lastPollAt = shallowRef<number>(0)
  const isPolling = shallowRef<boolean>(false)

  // The cards are pure snapshots, so without this the overview has no memory at
  // all — it cannot distinguish a steady 60% from one climbing through 60%.
  const trend = shallowRef<FleetTrendSample[]>([])

  const setEntry = (serverId: string, entry: FleetEntry) => {
    entries.value = { ...entries.value, [serverId]: entry }
  }

  const recordTrend = (next: Record<string, FleetEntry>) => {
    let cpuSum = 0
    let memSum = 0
    let online = 0

    for (const id of Object.keys(next)) {
      const entry = next[id]
      if (entry.state !== 'ONLINE' || !entry.metrics) continue
      cpuSum += entry.metrics.cpu?.usage ?? 0
      memSum += entry.metrics.memory?.usage ?? 0
      online++
    }

    // The seed poll carries no metrics yet; skip it rather than plot a zero.
    if (online === 0) return

    const sample: FleetTrendSample = {
      timestamp: Date.now() / 1000,
      cpu: cpuSum / online,
      memory: memSum / online,
      online,
    }

    const nextTrend = trend.value.length >= TREND_LIMIT
      ? trend.value.slice(trend.value.length - TREND_LIMIT + 1)
      : trend.value.slice()
    nextTrend.push(sample)
    trend.value = nextTrend
  }

  const setEntries = (next: Record<string, FleetEntry>) => {
    entries.value = next
    lastPollAt.value = Date.now()
    recordTrend(next)
  }

  const reset = () => {
    entries.value = {}
    lastPollAt.value = 0
    trend.value = []
  }

  const getEntry = (serverId: string): FleetEntry | null => entries.value[serverId] || null

  /** OFFLINE outranks every in-band health status. */
  const statusOf = (serverId: string): HealthStatus | 'offline' | 'pending' => {
    const entry = entries.value[serverId]
    if (!entry || entry.state === 'PENDING') return 'pending'
    if (entry.state === 'OFFLINE') return 'offline'
    return serverHealth(entry.metrics)
  }

  const counts = computed(() => {
    const result = { healthy: 0, warning: 0, critical: 0, offline: 0, pending: 0, total: 0 }
    for (const id of Object.keys(entries.value)) {
      const status = statusOf(id)
      result[status]++
      result.total++
    }
    return result
  })

  return {
    entries,
    lastPollAt,
    isPolling,
    trend,
    counts,
    setEntry,
    setEntries,
    getEntry,
    statusOf,
    reset,
  }
})
