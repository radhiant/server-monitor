export interface HostInfo {
  server_id: string
  hostname: string
  os: string
  platform: string
  platform_family?: string
  platform_version?: string
  kernel_version: string
  kernel_arch: string
  uptime: number
  boot_time: number
  cpu_model?: string
  cpu_cores: number
  total_memory: number
}

export interface CpuMetrics {
  usage: number
  cores: number
  per_core: number[]
  frequency_mhz?: number
  model_name?: string
}

export interface MemoryMetrics {
  total: number
  used: number
  available: number
  free: number
  usage: number
  cached: number
  buffers: number
  swap_total: number
  swap_used: number
  swap_free: number
  swap_usage: number
}

export interface DiskPartition {
  device: string
  mountpoint: string
  fstype: string
  total: number
  used: number
  free: number
  usage: number
}

export interface DiskIoMetrics {
  read_bytes_sec: number
  write_bytes_sec: number
  read_count_sec: number
  write_count_sec: number
}

export interface NetworkInterface {
  name: string
  rx_bytes_sec: number
  tx_bytes_sec: number
  rx_packets_sec: number
  tx_packets_sec: number
  total_rx_bytes: number
  total_tx_bytes: number
}

export interface LoadMetrics {
  '1m': number
  '5m': number
  '15m': number
  status: 'healthy' | 'warning' | 'critical'
}

export interface ProcessInfo {
  pid: number
  name: string
  cpu_percent: number
  memory_percent: number
  memory_rss: number
  memory_vms: number
  num_threads: number
  username: string
  status: string
}

export interface ServerMetrics {
  timestamp: number
  server_id: string
  status: 'healthy' | 'warning' | 'critical'
  uptime: number
  host?: HostInfo
  cpu: CpuMetrics
  memory: MemoryMetrics
  filesystems: DiskPartition[]
  disk_io: DiskIoMetrics
  network: NetworkInterface[]
  load: LoadMetrics
  processes?: ProcessInfo[]
}

export interface HistoryResponse {
  count: number
  samples: ServerMetrics[]
}

export interface WebSocketMessage {
  type: 'metrics' | 'processes' | 'server_status' | 'error' | 'pong'
  timestamp: number
  data: any
}

export type ConnectionState = 'CONNECTING' | 'CONNECTED' | 'RECONNECTING' | 'DISCONNECTED' | 'ERROR'
