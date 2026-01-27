// GoIM protocol constants
export const OP_HEARTBEAT = 2
export const OP_HEARTBEAT_REPLY = 3
export const OP_MESSAGE = 5
export const OP_AUTH = 7
export const OP_AUTH_REPLY = 8
export const OP_MESSAGE_PUSH = 9  // 消息推送
export const OP_CHANGE_ROOM = 12  // 加入/切换房间
export const OP_CHANGE_ROOM_REPLY = 13  // 加入/切换房间回复

export const PROTOCOL_VERSION = 102
export const HEADER_SIZE = 16

export interface ProtoMessage {
  ver: number
  op: number
  seq: number
  body?: string
}

// Encode a message according to goim protocol
export function encodeMessage(ver: number, op: number, seq: number, body?: string): ArrayBuffer {
  const bodyBytes = body ? new TextEncoder().encode(body) : new Uint8Array(0)
  const bodyLen = bodyBytes.length
  const packLen = HEADER_SIZE + bodyLen

  const buffer = new ArrayBuffer(packLen)
  const view = new DataView(buffer)
  const bytes = new Uint8Array(buffer)

  // packLen (int32, big endian)
  view.setUint32(0, packLen, false)
  // headerLen (int16, big endian)
  view.setUint16(4, HEADER_SIZE, false)
  // ver (int16, big endian)
  view.setUint16(6, ver, false)
  // op (int32, big endian)
  view.setUint32(8, op, false)
  // seq (int32, big endian)
  view.setUint32(12, seq, false)

  // body
  if (bodyLen > 0) {
    bytes.set(bodyBytes, HEADER_SIZE)
  }

  return buffer
}

// Decode a message from goim protocol
export function decodeMessage(buffer: ArrayBuffer): ProtoMessage {
  const view = new DataView(buffer)
  const packLen = view.getUint32(0, false)
  const headerLen = view.getUint16(4, false)
  const ver = view.getUint16(6, false)
  const op = view.getUint32(8, false)
  const seq = view.getUint32(12, false)
  const bodyLen = packLen - headerLen

  let body: string | undefined
  if (bodyLen > 0) {
    const bodyBytes = new Uint8Array(buffer, headerLen, bodyLen)
    body = new TextDecoder().decode(bodyBytes)
  }

  return { ver, op, seq, body }
}

export function getOpName(op: number): string {
  switch (op) {
    case OP_HEARTBEAT: return 'Heartbeat'
    case OP_HEARTBEAT_REPLY: return 'Heartbeat Reply'
    case OP_MESSAGE: return 'Message'
    case OP_AUTH: return 'Auth'
    case OP_AUTH_REPLY: return 'Auth Reply'
    case OP_MESSAGE_PUSH: return 'Message Push'
    case OP_CHANGE_ROOM: return 'Change Room'
    case OP_CHANGE_ROOM_REPLY: return 'Change Room Reply'
    default: return `Unknown(${op})`
  }
}
