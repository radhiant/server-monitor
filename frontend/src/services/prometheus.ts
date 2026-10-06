/**
 * Prometheus service for querying SLA and historical metrics for the server fleet.
 * Accesses Prometheus via the /prometheus/api/ reverse proxy.
 */

export interface PrometheusMetricResult {
  metric: Record<string, string>
  value?: [number, string]
  values?: [number, string][]
}

export interface PrometheusQueryResponse {
  status: string
  data: {
    resultType: string
    result: PrometheusMetricResult[]
  }
}

export interface HistoryTrendPoint {
  time: string
  timestamp: number
  cpu: number
  memory: number
}

export const SERVER_IP_MAP: Record<string, string> = {
  web01: '192.0.2.11',
  app01: '192.0.2.12',
  db01: '192.0.2.13',
  monitor01: '192.0.2.14',
  proxy01: '192.0.2.15',
  edge01: '198.51.100.20',
}

/** Reverse lookup: IP -> serverId */
export const IP_TO_SERVER_MAP: Record<string, string> = Object.entries(SERVER_IP_MAP).reduce(
  (acc, [id, ip]) => {
    acc[ip] = id
    return acc
  },
  {} as Record<string, string>
)

export function resolveServerIp(serverId: string, description?: string): string {
  if (SERVER_IP_MAP[serverId]) return SERVER_IP_MAP[serverId]
  if (description) {
    const match = description.match(/(\d+\.\d+\.\d+\.\d+)/)
    if (match) return match[1]
  }
  const directMatch = serverId.match(/(\d+\.\d+\.\d+\.\d+)/)
  if (directMatch) return directMatch[1]
  return serverId
}

export const prometheusServerApi = {
  /**
   * Get SLA uptime percentage for a given server over a window.
   * e.g. '24h', '7d', '30d', '90d', '1y' / '1th'
   */
  async getServerSla(serverId: string, rangeKey: string, description?: string): Promise<number | null> {
    const ip = resolveServerIp(serverId, description)
    const window = mapRangeToPromWindow(rangeKey)
    const query = `avg_over_time(up{job="server_fleet",instance=~".*${ip}.*"}[${window}]) * 100`

    try {
      const res = await fetch(`/prometheus/api/v1/query?query=${encodeURIComponent(query)}`)
      if (!res.ok) return null
      const json: PrometheusQueryResponse = await res.json()
      if (json.status === 'success' && json.data.result.length > 0) {
        const val = parseFloat(json.data.result[0].value?.[1] ?? '')
        return isNaN(val) ? null : Number(val.toFixed(2))
      }
      return null
    } catch {
      return null
    }
  },

  /**
   * Get SLA uptime percentage for all servers in a single query.
   * Returns dictionary keyed by both serverId and IP.
   */
  async getAllFleetSla(rangeKey: string): Promise<Record<string, number>> {
    const window = mapRangeToPromWindow(rangeKey)
    const query = `avg_over_time(up{job="server_fleet"}[${window}]) * 100`

    const result: Record<string, number> = {}
    try {
      const res = await fetch(`/prometheus/api/v1/query?query=${encodeURIComponent(query)}`)
      if (!res.ok) return result
      const json: PrometheusQueryResponse = await res.json()
      if (json.status === 'success' && Array.isArray(json.data.result)) {
        for (const item of json.data.result) {
          const instance = item.metric.instance || ''
          const ip = instance.split(':')[0]
          const val = parseFloat(item.value?.[1] ?? '')
          if (!isNaN(val)) {
            const rounded = Number(val.toFixed(2))
            if (ip) {
              result[ip] = rounded
              const mappedId = IP_TO_SERVER_MAP[ip]
              if (mappedId) result[mappedId] = rounded
            }
          }
        }
      }
      return result
    } catch {
      return result
    }
  },

  /**
   * Get overall Fleet average SLA percentage.
   */
  async getFleetSla(rangeKey: string): Promise<number | null> {
    const window = mapRangeToPromWindow(rangeKey)
    const query = `avg(avg_over_time(up{job="server_fleet"}[${window}])) * 100`

    try {
      const res = await fetch(`/prometheus/api/v1/query?query=${encodeURIComponent(query)}`)
      if (!res.ok) return null
      const json: PrometheusQueryResponse = await res.json()
      if (json.status === 'success' && json.data.result.length > 0) {
        const val = parseFloat(json.data.result[0].value?.[1] ?? '')
        return isNaN(val) ? null : Number(val.toFixed(2))
      }
      return null
    } catch {
      return null
    }
  },

  /**
   * Get CPU and Memory historical trend series for a server.
   */
  async getServerTrend(serverId: string, rangeKey: string, description?: string): Promise<HistoryTrendPoint[] | null> {
    const ip = resolveServerIp(serverId, description)
    const { startSec, stepSec } = resolveRangeParams(rangeKey)
    const nowSec = Math.floor(Date.now() / 1000)

    const cpuQuery = `server_cpu_usage_percent{instance=~".*${ip}.*"}`
    const memQuery = `server_memory_usage_percent{instance=~".*${ip}.*"}`

    try {
      const [cpuRes, memRes] = await Promise.all([
        fetch(`/prometheus/api/v1/query_range?query=${encodeURIComponent(cpuQuery)}&start=${startSec}&end=${nowSec}&step=${stepSec}`),
        fetch(`/prometheus/api/v1/query_range?query=${encodeURIComponent(memQuery)}&start=${startSec}&end=${nowSec}&step=${stepSec}`),
      ])

      if (!cpuRes.ok || !memRes.ok) return null

      const cpuJson: PrometheusQueryResponse = await cpuRes.json()
      const memJson: PrometheusQueryResponse = await memRes.json()

      const cpuVals = cpuJson.data.result[0]?.values ?? []
      const memMap = new Map<number, number>()

      for (const [ts, valStr] of memJson.data.result[0]?.values ?? []) {
        memMap.set(ts, parseFloat(valStr) || 0)
      }

      if (cpuVals.length === 0) return null

      return cpuVals.map(([ts, valStr]) => {
        const d = new Date(ts * 1000)
        const timeStr = formatPointDate(d, rangeKey)
        return {
          time: timeStr,
          timestamp: ts,
          cpu: Number(parseFloat(valStr).toFixed(1)),
          memory: Number((memMap.get(ts) ?? 0).toFixed(1)),
        }
      })
    } catch {
      return null
    }
  }
}

function formatPointDate(d: Date, rangeKey: string): string {
  const hh = d.getHours().toString().padStart(2, '0')
  const mm = d.getMinutes().toString().padStart(2, '0')
  if (rangeKey === '24h') {
    return `${hh}:${mm}`
  }
  const dd = d.getDate().toString().padStart(2, '0')
  const mo = (d.getMonth() + 1).toString().padStart(2, '0')
  return `${dd}/${mo} ${hh}:${mm}`
}

function mapRangeToPromWindow(rangeKey: string): string {
  switch (rangeKey) {
    case '7d': return '7d'
    case '30d': return '30d'
    case '90d': return '90d'
    case '1y':
    case '1th':
    case '365d': return '365d'
    case '3y':
    case '3th': return '1095d'
    default: return '24h'
  }
}

function resolveRangeParams(rangeKey: string): { startSec: number; stepSec: number } {
  const now = Math.floor(Date.now() / 1000)
  switch (rangeKey) {
    case '7d': return { startSec: now - 7 * 86400, stepSec: 1800 }
    case '30d': return { startSec: now - 30 * 86400, stepSec: 7200 }
    case '90d': return { startSec: now - 90 * 86400, stepSec: 21600 }
    case '1y':
    case '1th':
    case '365d': return { startSec: now - 365 * 86400, stepSec: 86400 }
    default: return { startSec: now - 86400, stepSec: 300 }
  }
}
