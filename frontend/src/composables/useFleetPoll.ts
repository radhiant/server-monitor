import { watch, onUnmounted } from 'vue'
import { useMetricsStore } from '@/stores/metricsStore'
import { useFleetStore, type FleetEntry } from '@/stores/fleetStore'
import { ApiClient } from '@/services/api'
import { useMockMetrics } from './useMockMetrics'
import type { ServerRegistryItem } from '@/types/server'

/** The fleet grid is an at-a-glance heartbeat, not a 1Hz stream. */
const DEFAULT_INTERVAL_MS = 5000
const REQUEST_TIMEOUT_MS = 4000

type MockGenerator = ReturnType<typeof useMockMetrics>

/**
 * Polls every server in the registry over REST and pushes the result into
 * `fleetStore`.
 *
 * REST rather than six parallel WebSockets: the overview only needs a coarse
 * refresh, and a plain request with a timeout gives an unambiguous
 * online/offline verdict per node. Requests are settled independently so one
 * unreachable host never blocks the rest of the grid.
 */
export function useFleetPoll(intervalMs: number = DEFAULT_INTERVAL_MS) {
  const store = useMetricsStore()
  const fleet = useFleetStore()

  let timer: ReturnType<typeof setInterval> | null = null
  let stopped = false
  const mockGenerators = new Map<string, MockGenerator>()

  const mockEntry = (server: ServerRegistryItem): FleetEntry => {
    let generator = mockGenerators.get(server.id)
    if (!generator) {
      generator = useMockMetrics(server.id)
      mockGenerators.set(server.id, generator)
    }
    return {
      state: 'ONLINE',
      metrics: generator.getNextSample(),
      error: null,
      updatedAt: Date.now(),
    }
  }

  const pollOne = async (server: ServerRegistryItem): Promise<[string, FleetEntry]> => {
    const api = new ApiClient(server.baseUrl, server.apiToken, REQUEST_TIMEOUT_MS)
    try {
      const metrics = await api.getMetrics()
      return [server.id, { state: 'ONLINE', metrics, error: null, updatedAt: Date.now() }]
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Tidak dapat dihubungi'
      return [server.id, { state: 'OFFLINE', metrics: null, error: message, updatedAt: Date.now() }]
    }
  }

  const pollAll = async () => {
    const list = store.servers
    if (list.length === 0) return

    fleet.isPolling = true

    try {
      if (store.isMockMode) {
        const next: Record<string, FleetEntry> = {}
        for (const server of list) {
          next[server.id] = mockEntry(server)
        }
        if (!stopped) fleet.setEntries(next)
        return
      }

      const settled = await Promise.allSettled(list.map(pollOne))
      if (stopped) return

      const next: Record<string, FleetEntry> = {}
      settled.forEach((result, index) => {
        if (result.status === 'fulfilled') {
          const [id, entry] = result.value
          next[id] = entry
        } else {
          // pollOne already traps its own errors, so this is a defensive branch.
          const server = list[index]
          next[server.id] = {
            state: 'OFFLINE',
            metrics: null,
            error: 'Polling gagal',
            updatedAt: Date.now(),
          }
        }
      })

      fleet.setEntries(next)
    } finally {
      fleet.isPolling = false
    }
  }

  const start = () => {
    stop()
    stopped = false

    // Seed every card as PENDING so the grid renders its full shape instantly
    // rather than popping in one card at a time.
    const pending: Record<string, FleetEntry> = {}
    for (const server of store.servers) {
      pending[server.id] = fleet.getEntry(server.id) || {
        state: 'PENDING',
        metrics: null,
        error: null,
        updatedAt: 0,
      }
    }
    fleet.setEntries(pending)

    void pollAll()
    timer = setInterval(() => void pollAll(), intervalMs)
  }

  const stop = () => {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
  }

  watch(
    () => [store.servers.length, store.isMockMode] as const,
    ([count]) => {
      if (count > 0) start()
    },
    { immediate: true },
  )

  onUnmounted(() => {
    stopped = true
    stop()
  })

  return { pollAll, start, stop }
}
