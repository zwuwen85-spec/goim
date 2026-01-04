import { ref, reactive } from 'vue'
import { encodeMessage, decodeMessage, OP_HEARTBEAT, OP_AUTH, getOpName, PROTOCOL_VERSION } from './protocol'

export type WebSocketStatus = 'connecting' | 'connected' | 'disconnected' | 'error'

export interface Message {
  msg_id: string
  from_user_id: number
  conversation_id: number
  conversation_type: number
  msg_type: number
  content: string
  seq: number
  created_at: number
}

export function useWebSocket(wsUrl: string) {
  const status = ref<WebSocketStatus>('disconnected')
  const ws = ref<WebSocket | null>(null)
  const messages = ref<Message[]>([])

  let seq = 1
  let heartbeatInterval: number | null = null
  let reconnectTimeout: number | null = null
  const messageHandlers: Map<number, (msg: any) => void> = new Map()

  // Send auth message
  const sendAuth = (token: string, userId: number) => {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) return

    const authData = JSON.stringify({
      mid: userId,
      key: `user_${userId}`,
      room_id: '',
      platform: 'web',
      accepts: [1001, 1002, 1003]
    })

    const buffer = encodeMessage(PROTOCOL_VERSION, OP_AUTH, seq++, authData)
    ws.value.send(buffer)
    console.log('[WS] Auth sent')
  }

  // Send heartbeat
  const sendHeartbeat = () => {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) return
    const buffer = encodeMessage(PROTOCOL_VERSION, OP_HEARTBEAT, seq++)
    ws.value.send(buffer)
  }

  // Start heartbeat
  const startHeartbeat = () => {
    stopHeartbeat()
    heartbeatInterval = window.setInterval(sendHeartbeat, 30000)
  }

  // Stop heartbeat
  const stopHeartbeat = () => {
    if (heartbeatInterval) {
      clearInterval(heartbeatInterval)
      heartbeatInterval = null
    }
  }

  // Connect
  const connect = (token: string, userId: number) => {
    if (ws.value?.readyState === WebSocket.OPEN) return

    status.value = 'connecting'

    try {
      ws.value = new WebSocket(wsUrl)
      ws.value.binaryType = 'arraybuffer'

      ws.value.onopen = () => {
        console.log('[WS] Connected')
        status.value = 'connected'
        sendAuth(token, userId)
        startHeartbeat()
      }

      ws.value.onmessage = (event) => {
        const msg = decodeMessage(event.data)
        console.log('[WS] Received:', getOpName(msg.op), msg.seq, msg.body)

        if (msg.op === 8) {
          // Auth reply
          console.log('[WS] Auth success')
        } else if (msg.body) {
          try {
            const data = JSON.parse(msg.body)
            messages.value.push(data)
          } catch {
            // Heartbeat or other message
          }
        }
      }

      ws.value.onclose = () => {
        console.log('[WS] Disconnected')
        status.value = 'disconnected'
        stopHeartbeat()
        // Only auto reconnect in production, not in development
        if (import.meta.env.PROD) {
          reconnectTimeout = window.setTimeout(() => connect(token, userId), 5000)
        }
      }

      ws.value.onerror = (error) => {
        console.error('[WS] Error:', error)
        status.value = 'error'
        // Don't auto-retry on error in development
      }
    } catch (error) {
      console.error('[WS] Connection error:', error)
      status.value = 'error'
    }
  }

  // Disconnect
  const disconnect = () => {
    if (reconnectTimeout) {
      clearTimeout(reconnectTimeout)
      reconnectTimeout = null
    }
    stopHeartbeat()
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
    status.value = 'disconnected'
  }

  // Send message
  const send = (op: number, body?: string) => {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) {
      console.error('[WS] Not connected')
      return false
    }

    const buffer = encodeMessage(PROTOCOL_VERSION, op, seq++, body)
    ws.value.send(buffer)
    return true
  }

  // Clear messages
  const clearMessages = () => {
    messages.value = []
  }

  return {
    status,
    messages,
    connect,
    disconnect,
    send,
    sendHeartbeat,
    clearMessages
  }
}
