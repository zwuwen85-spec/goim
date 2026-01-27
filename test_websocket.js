// WebSocket 测试脚本 - 使用正确的 GoIM 二进制协议
const WebSocket = require('ws');

// 用户ID
const USER_ID = 9002;
const WS_URL = 'ws://localhost:3102/sub';

// GoIM 协议常量
const PROTOCOL_VERSION = 102;
const HEADER_SIZE = 16;
const OP_AUTH = 7;
const OP_AUTH_REPLY = 8;
const OP_HEARTBEAT = 2;
const OP_HEARTBEAT_REPLY = 3;
const OP_MESSAGE = 5;

// 编码消息 (GoIM 二进制协议)
function encodeMessage(ver, op, seq, body = '') {
  const bodyBytes = Buffer.from(body, 'utf8');
  const bodyLen = bodyBytes.length;
  const packLen = HEADER_SIZE + bodyLen;

  const buffer = Buffer.alloc(packLen);

  // packLen (int32, big endian)
  buffer.writeUInt32BE(packLen, 0);
  // headerLen (int16, big endian)
  buffer.writeUInt16BE(HEADER_SIZE, 4);
  // ver (int16, big endian)
  buffer.writeUInt16BE(ver, 6);
  // op (int32, big endian)
  buffer.writeUInt32BE(op, 8);
  // seq (int32, big endian)
  buffer.writeUInt32BE(seq, 12);

  // body
  if (bodyLen > 0) {
    bodyBytes.copy(buffer, HEADER_SIZE);
  }

  return buffer;
}

// 解码消息
function decodeMessage(buffer) {
  const packLen = buffer.readUInt32BE(0);
  const headerLen = buffer.readUInt16BE(4);
  const ver = buffer.readUInt16BE(6);
  const op = buffer.readUInt32BE(8);
  const seq = buffer.readUInt32BE(12);
  const bodyLen = packLen - headerLen;

  let body = '';
  if (bodyLen > 0) {
    body = buffer.toString('utf8', headerLen);
  }

  return { ver, op, seq, body };
}

function getOpName(op) {
  const names = {
    2: 'HEARTBEAT',
    3: 'HEARTBEAT_REPLY',
    5: 'MESSAGE',
    7: 'AUTH',
    8: 'AUTH_REPLY'
  };
  return names[op] || `UNKNOWN(${op})`;
}

// 连接到 Comet WebSocket 服务
const ws = new WebSocket(WS_URL);
let seq = 1;

ws.on('open', function open() {
  console.log('✅ WebSocket 连接成功!');

  // 发送认证消息 (使用二进制协议)
  const authData = JSON.stringify({
    mid: USER_ID,
    key: `user_${USER_ID}_${Date.now()}`,
    room_id: '',
    platform: 'web',
    accepts: [1001, 1002, 1003]
  });

  console.log('📤 发送认证消息 (OP_AUTH=7):', authData);
  const authBuffer = encodeMessage(PROTOCOL_VERSION, OP_AUTH, seq++, authData);
  ws.send(authBuffer);
});

ws.on('message', function message(data) {
  const msg = decodeMessage(data);
  console.log(`📥 收到消息: OP=${getOpName(msg.op)} (${msg.op}), SEQ=${msg.seq}`);

  if (msg.op === OP_AUTH_REPLY) {
    console.log('✅ 认证成功！');
    console.log('   Body:', msg.body);
  } else if (msg.op === OP_HEARTBEAT_REPLY) {
    console.log('💓 心跳响应');
  } else if (msg.op === OP_MESSAGE) {
    console.log('📨 收到消息:', msg.body);
  }
});

ws.on('error', function error(err) {
  console.error('❌ WebSocket 错误:', err.message);
});

ws.on('close', function close() {
  console.log('🔌 WebSocket 连接关闭');
  process.exit(0);
});

// 30秒后自动关闭
setTimeout(() => {
  console.log('⏱️  测试结束，关闭连接');
  ws.close();
}, 30000);
