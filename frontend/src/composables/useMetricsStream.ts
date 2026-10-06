import { watch, onUnmounted } from 'vue'
import { useMetricsStore } from '@/stores/metricsStore'
import { ApiClient } from '@/services/api'
import { WebSocketClient } from '@/services/websocket'
import { useMockMetrics } from './useMockMetrics'

export function useMetricsStream() {
  const store = useMetricsStore()
  let wsClient: WebSocketClient | null = null
  let mockTimer: any = null

  const startMockStream = () => {
    stopStream()
    store.setConnectionState('CONNECTED')
    const mock = useMockMetrics(store.activeServerId || 'web01')
    store.setHostInfo(mock.mockHostInfo)
    
    // Initial 60 historical samples
    const history = mock.generateHistory(60)
    store.setHistory(history)
    store.setLatestMetrics(history[history.length - 1])

    mockTimer = setInterval(() => {
      const sample = mock.getNextSample()
      store.setLatestMetrics(sample)
    }, 1000)
  }

  const fetchRestTelemetry = async (api: ApiClient) => {
    try {
      // 1. Fetch initial host identity
      const hostInfo = await api.getServerInfo()
      if (hostInfo) {
        store.setHostInfo(hostInfo)
      }

      // 2. Fetch history for immediate chart render
      const historyResp = await api.getHistory('1m')
      if (historyResp && historyResp.samples && historyResp.samples.length > 0) {
        store.setHistory(historyResp.samples)
        store.setLatestMetrics(historyResp.samples[historyResp.samples.length - 1])
      } else {
        const latest = await api.getMetrics()
        if (latest) {
          store.setLatestMetrics(latest)
        }
      }
    } catch (err) {
      // REST might be warming up or proxying; WS will take over
    }
  }

  const startLiveStream = async () => {
    stopStream()
    const server = store.activeServer
    if (!server) {
      // Waiting for servers.json to load
      return
    }

    const api = new ApiClient(server.baseUrl, server.apiToken)
    store.setConnectionState('CONNECTING')

    // Initial REST load (non-blocking)
    fetchRestTelemetry(api)

    // Resilient WebSocket connection
    wsClient = new WebSocketClient(server.baseUrl, server.apiToken)
    
    wsClient.onStateChange((state) => {
      store.setConnectionState(state)
      if (state === 'CONNECTED') {
        // Refresh telemetry on reconnect
        fetchRestTelemetry(api)
      }
    })

    wsClient.onMessage((msg) => {
      if (msg.type === 'metrics' && msg.data) {
        store.setLatestMetrics(msg.data)
      }
    })

    wsClient.connect()
  }

  const initStream = () => {
    if (store.isMockMode) {
      startMockStream()
    } else {
      startLiveStream()
    }
  }

  const stopStream = () => {
    if (mockTimer) {
      clearInterval(mockTimer)
      mockTimer = null
    }
    if (wsClient) {
      wsClient.disconnect()
      wsClient = null
    }
  }

  // Reactively connect whenever activeServer becomes available or changes
  watch(
    () => [store.activeServer?.id, store.activeServer?.baseUrl, store.isMockMode],
    () => {
      if (store.isMockMode || store.activeServer) {
        initStream()
      }
    },
    { immediate: true }
  )

  onUnmounted(() => {
    stopStream()
  })

  return {
    initStream,
    stopStream
  }
}
