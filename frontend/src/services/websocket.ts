import type { WebSocketMessage, ConnectionState } from '@/types/metrics'

export type MessageHandler = (msg: WebSocketMessage) => void
export type StateChangeHandler = (state: ConnectionState) => void

export class WebSocketClient {
  private url: string
  private token?: string
  private ws: WebSocket | null = null
  private isExplicitlyClosed = false
  private reconnectAttempts = 0
  private maxReconnectDelay = 4000
  private reconnectTimer: any = null
  private pingInterval: any = null
  private onMessageCallback: MessageHandler | null = null
  private onStateChangeCallback: StateChangeHandler | null = null

  constructor(baseUrl: string, token?: string) {
    this.token = token
    const loc = window.location
    const protocol = loc.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = loc.host
    const cleanBase = baseUrl.replace(/\/$/, '')
    
    if (cleanBase.startsWith('http://') || cleanBase.startsWith('https://')) {
      const parsed = new URL(cleanBase)
      const wsProto = parsed.protocol === 'https:' ? 'wss:' : 'ws:'
      this.url = `${wsProto}//${parsed.host}${parsed.pathname}/ws/v1`
    } else if (cleanBase.startsWith('ws://') || cleanBase.startsWith('wss://')) {
      this.url = `${cleanBase}/ws/v1`
    } else {
      this.url = `${protocol}//${host}${cleanBase}/ws/v1`
    }

    if (this.token) {
      this.url += `?token=${encodeURIComponent(this.token)}`
    }
  }

  onMessage(cb: MessageHandler) {
    this.onMessageCallback = cb
  }

  onStateChange(cb: StateChangeHandler) {
    this.onStateChangeCallback = cb
  }

  private setState(state: ConnectionState) {
    if (this.onStateChangeCallback) {
      this.onStateChangeCallback(state)
    }
  }

  connect() {
    this.isExplicitlyClosed = false
    clearTimeout(this.reconnectTimer)

    // Cleanup any existing websocket instance completely
    this.cleanupSocket()

    this.setState(this.reconnectAttempts > 0 ? 'RECONNECTING' : 'CONNECTING')

    try {
      const socket = new WebSocket(this.url)
      this.ws = socket

      socket.onopen = () => {
        if (this.ws !== socket) return
        this.reconnectAttempts = 0
        this.setState('CONNECTED')
        this.startHeartbeat()
      }

      socket.onmessage = (event) => {
        if (this.ws !== socket) return
        try {
          const data: WebSocketMessage = JSON.parse(event.data)
          if (this.onMessageCallback) {
            this.onMessageCallback(data)
          }
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e)
        }
      }

      socket.onerror = () => {
        if (this.ws !== socket) return
        // Don't set error permanently if we are in reconnect loop
        if (this.reconnectAttempts === 0) {
          this.setState('ERROR')
        }
      }

      socket.onclose = () => {
        if (this.ws !== socket) return
        this.stopHeartbeat()
        if (!this.isExplicitlyClosed) {
          this.setState(this.reconnectAttempts > 0 ? 'RECONNECTING' : 'DISCONNECTED')
          this.scheduleReconnect()
        }
      }
    } catch (e) {
      this.setState('ERROR')
      this.scheduleReconnect()
    }
  }

  private cleanupSocket() {
    if (this.ws) {
      this.ws.onopen = null
      this.ws.onclose = null
      this.ws.onerror = null
      this.ws.onmessage = null
      try {
        if (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING) {
          this.ws.close()
        }
      } catch {}
      this.ws = null
    }
  }

  private scheduleReconnect() {
    if (this.isExplicitlyClosed) return
    clearTimeout(this.reconnectTimer)

    // Fast retry cadence: 1s, 2s, 3s, max 4s
    const delay = Math.min(1000 + (this.reconnectAttempts * 1000), this.maxReconnectDelay)
    this.reconnectAttempts++

    this.reconnectTimer = setTimeout(() => {
      if (!this.isExplicitlyClosed) {
        this.connect()
      }
    }, delay)
  }

  private startHeartbeat() {
    this.stopHeartbeat()
    this.pingInterval = setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        try {
          this.ws.send(JSON.stringify({ type: 'ping' }))
        } catch {}
      }
    }, 15000)
  }

  private stopHeartbeat() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval)
      this.pingInterval = null
    }
  }

  disconnect() {
    this.isExplicitlyClosed = true
    clearTimeout(this.reconnectTimer)
    this.stopHeartbeat()
    this.cleanupSocket()
    this.setState('DISCONNECTED')
  }
}
