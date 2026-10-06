import type { ServerMetrics, HostInfo, ProcessInfo } from '@/types/metrics'

export function useMockMetrics(serverID = 'web-01') {
  let tickCount = 0
  let cpuTrend = 28.5
  let memUsed = 14200000000 // ~14.2 GB
  const totalMem = 34359738368 // 32 GB
  const swapTotal = 8589934592 // 8 GB

  const mockHostInfo: HostInfo = {
    server_id: serverID,
    hostname: `${serverID}.prod.cloud`,
    os: 'linux',
    platform: 'ubuntu',
    platform_family: 'debian',
    platform_version: '24.04 LTS (Noble Numbat)',
    kernel_version: '6.8.0-40-generic',
    kernel_arch: 'x86_64',
    uptime: 1428590,
    boot_time: Date.now() / 1000 - 1428590,
    cpu_model: 'AMD EPYC 7763 64-Core Processor @ 2.45GHz',
    cpu_cores: 8,
    total_memory: totalMem
  }

  const baseProcesses: ProcessInfo[] = [
    { pid: 1420, name: 'postgres: pooler', cpu_percent: 14.2, memory_percent: 6.4, memory_rss: 2199023255, memory_vms: 4294967296, num_threads: 18, username: 'postgres', status: 'running' },
    { pid: 2891, name: 'nginx: worker process', cpu_percent: 8.5, memory_percent: 1.2, memory_rss: 412450000, memory_vms: 1073741824, num_threads: 4, username: 'www-data', status: 'running' },
    { pid: 3104, name: 'redis-server *:6379', cpu_percent: 5.1, memory_percent: 3.8, memory_rss: 1305000000, memory_vms: 2147483648, num_threads: 6, username: 'redis', status: 'running' },
    { pid: 5612, name: 'server-monitor', cpu_percent: 0.8, memory_percent: 0.1, memory_rss: 34500000, memory_vms: 120000000, num_threads: 8, username: 'root', status: 'running' },
    { pid: 1102, name: 'docker-containerd', cpu_percent: 3.2, memory_percent: 2.1, memory_rss: 720000000, memory_vms: 1500000000, num_threads: 24, username: 'root', status: 'running' },
    { pid: 7801, name: 'node /app/server.js', cpu_percent: 11.4, memory_percent: 5.2, memory_rss: 1780000000, memory_vms: 3200000000, num_threads: 12, username: 'node', status: 'running' },
    { pid: 8990, name: 'systemd-journald', cpu_percent: 0.2, memory_percent: 0.4, memory_rss: 140000000, memory_vms: 300000000, num_threads: 2, username: 'root', status: 'running' },
    { pid: 9412, name: 'prometheus', cpu_percent: 4.6, memory_percent: 4.1, memory_rss: 1409000000, memory_vms: 2800000000, num_threads: 16, username: 'prometheus', status: 'running' },
    { pid: 4001, name: 'sshd: user [priv]', cpu_percent: 0.0, memory_percent: 0.1, memory_rss: 35000000, memory_vms: 80000000, num_threads: 1, username: 'root', status: 'sleeping' },
    { pid: 1024, name: 'rsyslogd', cpu_percent: 0.1, memory_percent: 0.1, memory_rss: 25000000, memory_vms: 60000000, num_threads: 4, username: 'syslog', status: 'sleeping' },
  ]

  const getNextSample = (): ServerMetrics => {
    tickCount++
    const now = Math.floor(Date.now() / 1000)

    // Realistic trended CPU oscillation
    const noise = (Math.random() - 0.48) * 4.0
    cpuTrend = Math.max(8.0, Math.min(92.0, cpuTrend + noise + Math.sin(tickCount / 10) * 1.5))
    const cores = 8
    const perCore = Array.from({ length: cores }, (_, i) => {
      const coreNoise = (Math.random() - 0.5) * 8.0
      return Math.max(2.0, Math.min(100.0, +(cpuTrend + coreNoise + Math.sin(tickCount / 5 + i) * 10).toFixed(1)))
    })

    // Memory variation
    memUsed = Math.max(10000000000, Math.min(28000000000, memUsed + (Math.random() - 0.49) * 20000000))
    const memUsagePct = +((memUsed / totalMem) * 100).toFixed(1)
    const swapUsed = 1073741824 + Math.floor(Math.sin(tickCount / 20) * 200000000)

    // Network delta rate variation (KB/s to MB/s)
    const burst = Math.random() > 0.85 ? 12000000 : 0
    const rxRate = 1500000 + Math.random() * 800000 + burst + Math.sin(tickCount / 7) * 600000
    const txRate = 800000 + Math.random() * 400000 + (burst * 0.4) + Math.cos(tickCount / 7) * 300000

    // Disk I/O rates
    const diskIoBurst = Math.random() > 0.9 ? 25000000 : 0
    const diskRead = 1200000 + Math.random() * 500000 + diskIoBurst
    const diskWrite = 3400000 + Math.random() * 1200000 + (diskIoBurst * 0.8)

    // Load Average
    const load1 = +(cores * (cpuTrend / 100) * 0.95).toFixed(2)
    const load5 = +(cores * (cpuTrend / 100) * 0.88).toFixed(2)
    const load15 = +(cores * (cpuTrend / 100) * 0.80).toFixed(2)

    let serverStatus: 'healthy' | 'warning' | 'critical' = 'healthy'
    if (cpuTrend >= 90 || memUsagePct >= 90) serverStatus = 'critical'
    else if (cpuTrend >= 70 || memUsagePct >= 70) serverStatus = 'warning'

    // Update process CPU jitter
    const liveProcesses = baseProcesses.map((p) => {
      const pNoise = (Math.random() - 0.5) * 1.5
      return {
        ...p,
        cpu_percent: Math.max(0.1, +(p.cpu_percent + pNoise).toFixed(1))
      }
    })

    return {
      timestamp: now,
      server_id: serverID,
      status: serverStatus,
      uptime: mockHostInfo.uptime + tickCount,
      cpu: {
        usage: +cpuTrend.toFixed(1),
        cores: cores,
        per_core: perCore,
        frequency_mhz: 2450,
        model_name: mockHostInfo.cpu_model
      },
      memory: {
        total: totalMem,
        used: Math.floor(memUsed),
        available: totalMem - Math.floor(memUsed),
        free: totalMem - Math.floor(memUsed) - 2147483648,
        usage: memUsagePct,
        cached: 4294967296,
        buffers: 1073741824,
        swap_total: swapTotal,
        swap_used: swapUsed,
        swap_free: swapTotal - swapUsed,
        swap_usage: +((swapUsed / swapTotal) * 100).toFixed(1)
      },
      filesystems: [
        { device: '/dev/nvme0n1p2', mountpoint: '/', fstype: 'ext4', total: 250000000000, used: 115000000000, free: 135000000000, usage: 46.0 },
        { device: '/dev/nvme1n1', mountpoint: '/var/lib/docker', fstype: 'ext4', total: 500000000000, used: 310000000000, free: 190000000000, usage: 62.0 },
        { device: '/dev/sda1', mountpoint: '/mnt/storage', fstype: 'xfs', total: 2000000000000, used: 840000000000, free: 1160000000000, usage: 42.0 }
      ],
      disk_io: {
        read_bytes_sec: diskRead,
        write_bytes_sec: diskWrite,
        read_count_sec: 140,
        write_count_sec: 320
      },
      network: [
        {
          name: 'eth0',
          rx_bytes_sec: rxRate,
          tx_bytes_sec: txRate,
          rx_packets_sec: rxRate / 1200,
          tx_packets_sec: txRate / 1200,
          total_rx_bytes: 428900120390,
          total_tx_bytes: 184719203910
        },
        {
          name: 'docker0',
          rx_bytes_sec: rxRate * 0.4,
          tx_bytes_sec: txRate * 0.4,
          rx_packets_sec: (rxRate * 0.4) / 1000,
          tx_packets_sec: (txRate * 0.4) / 1000,
          total_rx_bytes: 54900120390,
          total_tx_bytes: 32719203910
        }
      ],
      load: {
        '1m': load1,
        '5m': load5,
        '15m': load15,
        status: load1 >= cores * 1.5 ? 'critical' : (load1 >= cores * 0.9 ? 'warning' : 'healthy')
      },
      processes: liveProcesses
    }
  }

  // Prepopulate 60 historical samples
  const generateHistory = (count = 60): ServerMetrics[] => {
    const samples: ServerMetrics[] = []
    const now = Math.floor(Date.now() / 1000)
    for (let i = count - 1; i >= 0; i--) {
      const s = getNextSample()
      s.timestamp = now - i
      samples.push(s)
    }
    return samples
  }

  return {
    mockHostInfo,
    getNextSample,
    generateHistory
  }
}
