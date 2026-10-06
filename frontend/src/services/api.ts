import type { HostInfo, ServerMetrics, HistoryResponse, DiskPartition, ProcessInfo, NetworkInterface } from '@/types/metrics'

/** A dead host must fail fast, otherwise one node stalls the whole fleet poll. */
const DEFAULT_TIMEOUT_MS = 4000

export class ApiClient {
  private baseUrl: string
  private token?: string
  private timeoutMs: number

  constructor(baseUrl: string, token?: string, timeoutMs: number = DEFAULT_TIMEOUT_MS) {
    this.baseUrl = baseUrl.replace(/\/$/, '')
    this.token = token
    this.timeoutMs = timeoutMs
  }

  private getHeaders(): HeadersInit {
    const headers: HeadersInit = {
      'Accept': 'application/json'
    }
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`
    }
    return headers
  }

  private async request<T>(endpoint: string): Promise<T> {
    const url = `${this.baseUrl}/api/v1${endpoint}`
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), this.timeoutMs)

    try {
      const res = await fetch(url, {
        headers: this.getHeaders(),
        signal: controller.signal
      })

      const contentType = res.headers.get('content-type') || ''

      if (!res.ok) {
        if (contentType.includes('application/json')) {
          const errorBody = await res.json().catch(() => ({}))
          const message = errorBody?.error?.message || `HTTP ${res.status}: ${res.statusText}`
          throw new Error(message)
        }
        throw new Error(`HTTP ${res.status}: ${res.statusText}`)
      }

      // Check if response is HTML instead of JSON (common when Nginx proxy is missing)
      if (!contentType.includes('application/json')) {
        const text = await res.text()
        if (text.trim().startsWith('<') || text.includes('<!DOCTYPE')) {
          throw new Error(`Nginx belum memiliki konfigurasi proxy untuk ${this.baseUrl}. Pastikan nginx.conf sudah di-update dan di-reload.`)
        }
        try {
          return JSON.parse(text)
        } catch {
          throw new Error(`Invalid non-JSON response from ${url}`)
        }
      }

      return res.json()
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') {
        throw new Error(`Timeout setelah ${this.timeoutMs}ms`)
      }
      throw err
    } finally {
      clearTimeout(timer)
    }
  }

  async getHealth(): Promise<{ status: string; server_id: string; uptime: number }> {
    return this.request('/health')
  }

  async getServerInfo(): Promise<HostInfo> {
    return this.request('/server')
  }

  async getStatus(): Promise<{ status: string; timestamp: number; uptime: number }> {
    return this.request('/status')
  }

  async getMetrics(): Promise<ServerMetrics> {
    return this.request('/metrics')
  }

  async getHistory(range: '1m' | '5m' | '15m' = '1m'): Promise<HistoryResponse> {
    return this.request(`/history?range=${range}`)
  }

  async getFilesystems(): Promise<DiskPartition[]> {
    return this.request('/filesystems')
  }

  async getProcesses(): Promise<ProcessInfo[]> {
    return this.request('/processes')
  }

  async getInterfaces(): Promise<NetworkInterface[]> {
    return this.request('/network/interfaces')
  }
}
