import { ref } from 'vue'
import { encodeMessage, decodeMessage, OP_HEARTBEAT, OP_AUTH, OP_CHANGE_ROOM, OP_CHANGE_ROOM_REPLY, getOpName, PROTOCOL_VERSION } from './protocol'

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

export interface GroupUpdateMessage {
  type: 'group_update'
  group_id: number
  group_no?: string
  name?: string
  avatar?: string
  timestamp: number
}

export interface UserUpdateMessage {
  type: 'user_update'
  user_id: number
  nickname?: string
  avatar?: string
  signature?: string
  timestamp: number
}

export interface GroupMemberUpdateMessage {
  type: 'group_member_update'
  group_id: number
  user_id: number
  nickname?: string
  avatar?: string
  timestamp: number
}

export function useWebSocket(wsUrl: string) {
  const status = ref<WebSocketStatus>('disconnected')
  const ws = ref<WebSocket | null>(null)
  const messages = ref<Message[]>([])
  const groupUpdates = ref<GroupUpdateMessage[]>([])
  const userUpdates = ref<UserUpdateMessage[]>([])
  const groupMemberUpdates = ref<GroupMemberUpdateMessage[]>([])

  let seq = 1
  let heartbeatInterval: number | null = null
  let reconnectTimeout: number | null = null

  // Send auth message
  const sendAuth = (_token: string, userId: number) => {
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
        } else if (msg.op === 3) {
          // Heartbeat reply - no body to parse
          console.log('[WS] Heartbeat reply')
        } else if (msg.op === 9) {
          // Message push - handle special case for extra whitespace/binary data
          if (msg.body) {
            try {
              // Remove all non-printable characters and null bytes
              // This handles the Room batch processing prefix bytes
              let cleanBody = msg.body
              // Remove null bytes and other control characters
              cleanBody = cleanBody.replace(/[\x00-\x08\x0B-\x0C\x0E-\x1F\x7F]/g, '')
              // Also trim whitespace
              cleanBody = cleanBody.trim()
              // Find the first { character in case there's prefix data
              const jsonStart = cleanBody.indexOf('{')
              if (jsonStart > 0) {
                cleanBody = cleanBody.substring(jsonStart)
              }
              const data = JSON.parse(cleanBody)
              console.log('[WS] Parsed message data:', data)

              // Check if this is a group update notification
              if (data.type === 'group_update') {
                console.log('[WS] Group update notification:', data)
                groupUpdates.value.push(data)
              } else if (data.type === 'user_update') {
                console.log('[WS] User update notification:', data)
                userUpdates.value.push(data)
              } else if (data.type === 'group_member_update') {
                console.log('[WS] Group member update notification:', data)
                groupMemberUpdates.value.push(data)
              } else {
                messages.value.push(data)
                console.log('[WS] Added to messages array, new length:', messages.value.length)
              }
            } catch (e) {
              console.log('[WS] Failed to parse body:', e)
              console.log('[WS] Raw body length:', msg.body.length, 'First 100 chars:', msg.body.substring(0, 100))
            }
          }
        } else if (msg.op === OP_CHANGE_ROOM_REPLY) {
          // Change room reply - body is the room ID string, not JSON
          console.log('[WS] Change room successful:', msg.body)
        } else if (msg.body && msg.body.trim().length > 0) {
          try {
            const data = JSON.parse(msg.body)
            messages.value.push(data)
          } catch (e) {
            console.log('[WS] Failed to parse body:', e)
          }
        } else {
          console.log('[WS] Message has no body, op:', msg.op)
        }
      }

      ws.value.onclose = (event) => {
        console.log('[WS] Disconnected', { code: event.code, reason: event.reason, wasClean: event.wasClean })
        status.value = 'disconnected'
        stopHeartbeat()
        // Auto reconnect in both production and development
        reconnectTimeout = window.setTimeout(() => connect(token, userId), 5000)
      }

      ws.value.onerror = (error) => {
        console.error('[WS] Error:', error)
        status.value = 'error'
        // Auto-retry on error
        reconnectTimeout = window.setTimeout(() => connect(token, userId), 5000)
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
    console.log('[WS] Sending: op=' + op + ' (' + getOpName(op) + '), seq=' + (seq-1) + ', body=' + (body || '""') + ', len=' + buffer.byteLength)
    console.log('[WS] WebSocket state: ' + ws.value.readyState + ', bufferedAmount: ' + ws.value.bufferedAmount)
    try {
      ws.value.send(buffer)
      console.log('[WS] Send completed successfully')
    } catch (e) {
      console.error('[WS] Send failed:', e)
      return false
    }
    return true
  }

  // Clear messages
  const clearMessages = () => {
    messages.value = []
  }

  // Clear group updates
  const clearGroupUpdates = () => {
    groupUpdates.value = []
  }

  // Clear user updates
  const clearUserUpdates = () => {
    userUpdates.value = []
  }

  // Clear group member updates
  const clearGroupMemberUpdates = () => {
    groupMemberUpdates.value = []
  }

  // Change room (join/leave room for group chat)
  const changeRoom = (roomId: string) => {
    console.log('[WS] Changing to room:', roomId)
    return send(OP_CHANGE_ROOM, roomId)
  }

  return {
    status,
    messages,
    groupUpdates,
    userUpdates,
    groupMemberUpdates,
    connect,
    disconnect,
    send,
    sendHeartbeat,
    clearMessages,
    clearGroupUpdates,
    clearUserUpdates,
    clearGroupMemberUpdates,
    changeRoom
  }
}
