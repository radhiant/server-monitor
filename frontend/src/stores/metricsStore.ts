import { defineStore } from 'pinia'
import { ref, shallowRef, computed } from 'vue'
import type { ServerMetrics, HostInfo, ConnectionState } from '@/types/metrics'
import type { ServerRegistryItem } from '@/types/server'
import { pickPrimaryInterface } from '@/utils/network'

export type DashboardView = 'fleet' | 'server'
/** 'tv' locks the dashboard to one screen; 'scroll' is the classic long page. */
export type FitMode = 'tv' | 'scroll'

export type TimeRangeKey = '24h' | '7d' | '30d' | '90d' | '1th'

export const TIME_RANGES: { key: TimeRangeKey; label: string }[] = [
  { key: '24h', label: '24 JAM' },
  { key: '7d', label: '7 HARI' },
  { key: '30d', label: '30 HARI' },
  { key: '90d', label: '90 HARI' },
  { key: '1th', label: '1 TAHUN' },
]

const FIT_MODE_KEY = 'srvmon:fitMode'
const HISTORY_LIMIT = 60

export const useMetricsStore = defineStore('metrics', () => {
  const servers = ref<ServerRegistryItem[]>([])

  const readParams = (): URLSearchParams => {
    try {
      return new URLSearchParams(window.location.search)
    } catch {
      return new URLSearchParams()
    }
  }

  const getUrlServerId = (): string => readParams().get('server') || ''

  const getInitialView = (): DashboardView => {
    const params = readParams()
    const view = params.get('view')
    if (view === 'fleet' || view === 'server') return view
    // Older copy-links only carried ?server=, and they meant the detail page.
    return params.get('server') ? 'server' : 'fleet'
  }

  /** Dividing a short viewport this many ways leaves charts too small to read,
   *  so anything under a 1080p-class display starts scrollable. An explicit
   *  choice always wins over the heuristic. */
  const TV_MIN_HEIGHT = 900
  const TV_MIN_WIDTH = 1200

  const getInitialFitMode = (): FitMode => {
    try {
      if (typeof window !== 'undefined' && window.innerWidth < 1024) {
        return 'scroll'
      }
      const saved = localStorage.getItem(FIT_MODE_KEY)
      if (saved === 'scroll' || saved === 'tv') return saved
    } catch {
      // Storage unavailable; fall through to the size heuristic.
    }
    return (typeof window !== 'undefined' && window.innerHeight >= TV_MIN_HEIGHT && window.innerWidth >= TV_MIN_WIDTH) ? 'tv' : 'scroll'
  }

  const view = ref<DashboardView>(getInitialView())
  const fitMode = ref<FitMode>(getInitialFitMode())
  const activeServerId = ref<string>(getUrlServerId())

  // shallowRef throughout: every snapshot carries ~20 process objects plus
  // filesystem and interface arrays. Deep-proxying that at 1Hz — and then
  // deep-watching it from four charts — was the dashboard's biggest cost.
  // Snapshots are replaced wholesale, never mutated, so shallow is sufficient.
  const hostInfo = shallowRef<HostInfo | null>(null)
  const latestMetrics = shallowRef<ServerMetrics | null>(null)
  const history = shallowRef<ServerMetrics[]>([])

  // Start in CONNECTING state on initial page load to avoid flashing disconnected banner
  const connectionState = ref<ConnectionState>('CONNECTING')
  const isMockMode = ref<boolean>(import.meta.env.VITE_DATA_SOURCE === 'mock')
  const selectedIface = ref<string>('')
  const timeRange = ref<TimeRangeKey>('24h')

  const setTimeRange = (range: TimeRangeKey) => {
    timeRange.value = range
  }

  const activeServer = computed(() => {
    if (servers.value.length === 0) return null
    return servers.value.find(s => s.id === activeServerId.value) || servers.value[0] || null
  })

  const isLive = computed(() => connectionState.value === 'CONNECTED')

  /** True until the first snapshot lands, so panels can show placeholders
   *  instead of a confident-looking 0.0%. */
  const hasData = computed(() => latestMetrics.value !== null)

  const syncUrl = () => {
    try {
      const url = new URL(window.location.href)
      url.searchParams.set('view', view.value)
      if (view.value === 'server' && activeServerId.value) {
        url.searchParams.set('server', activeServerId.value)
      } else if (view.value === 'fleet') {
        url.searchParams.delete('server')
      }
      window.history.replaceState({}, '', url.toString())
    } catch {
      // Ignore in non-browser environments
    }
  }

  const clearTelemetry = () => {
    latestMetrics.value = null
    history.value = []
    hostInfo.value = null
    selectedIface.value = ''
  }

  const setServers = (list: ServerRegistryItem[]) => {
    servers.value = list
    if (list.length === 0) return

    const urlId = getUrlServerId()
    const foundInList = list.find(s => s.id === urlId)

    if (foundInList) {
      activeServerId.value = foundInList.id
    } else if (!activeServerId.value || !list.find(s => s.id === activeServerId.value)) {
      activeServerId.value = list[0].id
    }

    syncUrl()
  }

  const setView = (next: DashboardView) => {
    if (view.value === next) return
    view.value = next
    if (next === 'fleet') {
      // Detail telemetry is stale the moment we leave; drop it so returning to
      // a server never paints the previous server's numbers for a frame.
      clearTelemetry()
    }
    syncUrl()
  }

  const setActiveServerId = (id: string) => {
    activeServerId.value = id
    clearTelemetry()
    syncUrl()
  }

  /** Fleet card click: select the server and open its detail page at once. */
  const openServer = (id: string) => {
    activeServerId.value = id
    view.value = 'server'
    clearTelemetry()
    syncUrl()
  }

  const setFitMode = (mode: FitMode) => {
    fitMode.value = mode
    try {
      localStorage.setItem(FIT_MODE_KEY, mode)
    } catch {
      // Preference simply won't persist.
    }
  }

  const toggleFitMode = () => {
    setFitMode(fitMode.value === 'tv' ? 'scroll' : 'tv')
  }

  const setHostInfo = (info: HostInfo) => {
    hostInfo.value = info
  }

  const setLatestMetrics = (m: ServerMetrics) => {
    latestMetrics.value = m

    if (m.host) {
      hostInfo.value = m.host
    }

    if (!selectedIface.value) {
      // network[0] is `lo` on every one of these hosts, which pinned the
      // network card at a permanent 0.00 MB/s.
      const primary = pickPrimaryInterface(m.network)
      if (primary) selectedIface.value = primary.name
    }

    // New array reference so shallowRef watchers fire without deep tracking.
    const next = history.value.length >= HISTORY_LIMIT
      ? history.value.slice(history.value.length - HISTORY_LIMIT + 1)
      : history.value.slice()
    next.push(m)
    history.value = next
  }

  const setHistory = (samples: ServerMetrics[]) => {
    history.value = samples.slice(-HISTORY_LIMIT)
    const last = samples[samples.length - 1]
    if (last?.host) {
      hostInfo.value = last.host
    }
  }

  const setConnectionState = (state: ConnectionState) => {
    connectionState.value = state
  }

  const setSelectedInterface = (iface: string) => {
    selectedIface.value = iface
  }

  // Handle browser back/forward buttons
  if (typeof window !== 'undefined') {
    window.addEventListener('popstate', () => {
      const nextView = getInitialView()
      const urlId = getUrlServerId()

      if (urlId && urlId !== activeServerId.value && servers.value.find(s => s.id === urlId)) {
        activeServerId.value = urlId
        clearTelemetry()
      }
      if (nextView !== view.value) {
        view.value = nextView
      }
    })
  }

  return {
    servers,
    view,
    fitMode,
    activeServerId,
    activeServer,
    hostInfo,
    latestMetrics,
    history,
    connectionState,
    isMockMode,
    selectedIface,
    timeRange,
    setTimeRange,
    TIME_RANGES,
    isLive,
    hasData,
    setServers,
    setView,
    openServer,
    setFitMode,
    toggleFitMode,
    setActiveServerId,
    setHostInfo,
    setLatestMetrics,
    setHistory,
    setConnectionState,
    setSelectedInterface,
  }
})
