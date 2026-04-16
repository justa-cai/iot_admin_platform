import { ref, onUnmounted } from 'vue'
import { getWebSocketUrl } from '@/api/client'
import type { WSEvent } from '@/types'

export function useWebSocket() {
  const ws = ref<WebSocket | null>(null)
  const connected = ref(false)
  const events = ref<WSEvent[]>([])
  const latestEvent = ref<WSEvent | null>(null)
  const reconnectTimer = ref<ReturnType<typeof setTimeout> | null>(null)
  const maxReconnectDelay = 30000
  let reconnectAttempts = 0

  function connect() {
    if (ws.value && ws.value.readyState === WebSocket.OPEN) return

    const url = getWebSocketUrl()
    ws.value = new WebSocket(url)

    ws.value.onopen = () => {
      connected.value = true
      reconnectAttempts = 0
    }

    ws.value.onmessage = (event) => {
      try {
        const data: WSEvent = JSON.parse(event.data)
        latestEvent.value = data
        events.value.push(data)
        if (events.value.length > 200) {
          events.value = events.value.slice(-200)
        }
      } catch {
        // ignore non-JSON messages
      }
    }

    ws.value.onclose = () => {
      connected.value = false
      scheduleReconnect()
    }

    ws.value.onerror = () => {
      connected.value = false
    }
  }

  function scheduleReconnect() {
    if (reconnectTimer.value) clearTimeout(reconnectTimer.value)
    const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), maxReconnectDelay)
    reconnectAttempts++
    reconnectTimer.value = setTimeout(() => {
      connect()
    }, delay)
  }

  function disconnect() {
    if (reconnectTimer.value) {
      clearTimeout(reconnectTimer.value)
      reconnectTimer.value = null
    }
    if (ws.value) {
      ws.value.onclose = null
      ws.value.close()
      ws.value = null
    }
    connected.value = false
  }

  function send(data: Record<string, unknown>) {
    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify(data))
    }
  }

  function clearEvents() {
    events.value = []
    latestEvent.value = null
  }

  onUnmounted(() => {
    disconnect()
  })

  return {
    connected,
    events,
    latestEvent,
    connect,
    disconnect,
    send,
    clearEvents,
  }
}
