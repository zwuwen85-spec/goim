// 完整的消息流程测试
const WebSocket = require('ws');

const USER_ID = 1001;
const WS_URL = 'ws://localhost:3102/sub';

// GoIM 协议常量
const PROTOCOL_VERSION = 102;
const HEADER_SIZE = 16;
const OP_AUTH = 7;
const OP_AUTH_REPLY = 8;
const OP_MESSAGE = 5;

// 编码消息
function encodeMessage(ver, op, seq, body = '') {
  const bodyBytes = Buffer.from(body, 'utf8');
  const bodyLen = bodyBytes.length;
  const packLen = HEADER_SIZE + bodyLen;

  const buffer = Buffer.alloc(packLen);
  buffer.writeUInt32BE(packLen, 0);
  buffer.writeUInt16BE(HEADER_SIZE, 4);
  buffer.writeUInt16BE(ver, 6);
  buffer.writeUInt32BE(op, 8);
  buffer.writeUInt32BE(seq, 12);

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
    2: 'HEARTBEAT', 3: 'HEARTBEAT_REPLY',
    5: 'MESSAGE', 7: 'AUTH', 8: 'AUTH_REPLY',
    9: 'MESSAGE_PUSH', 10: 'HEARTBEAT'
  };
  return names[op] || `UNKNOWN(${op})`;
}

// 创建连接
const ws = new WebSocket(WS_URL);
let seq = 1;
let messageCount = 0;

ws.on('open', () => {
  console.log('✅ WebSocket 连接成功!');

  const authData = JSON.stringify({
    mid: USER_ID,
    key: `user_${USER_ID}`,
    room_id: '',
    platform: 'web',
    accepts: [1001, 1002, 1003]
  });

  console.log('📤 发送认证...');
  ws.send(encodeMessage(PROTOCOL_VERSION, OP_AUTH, seq++, authData));
});

ws.on('message', (data) => {
  const msg = decodeMessage(data);

  // 详细打印所有消息
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log(`📥 [消息 #${++messageCount}] OP=${getOpName(msg.op)} (${msg.op}), SEQ=${msg.seq}`);
  console.log(`   Body Length: ${msg.body.length} bytes`);

  if (msg.body) {
    console.log(`   Body: ${msg.body.substring(0, 200)}`);
    try {
      const jsonData = JSON.parse(msg.body);
      console.log(`   Parsed JSON:`, JSON.stringify(jsonData, null, 2));
    } catch (e) {
      console.log(`   (不是 JSON 格式)`);
    }
  }

  if (msg.op === OP_AUTH_REPLY) {
    console.log(`   ✅ 认证成功！`);
    console.log(`   📝 现在可以在另一个终端通过 API 发送消息给用户 ${USER_ID}`);
    console.log(`   curl -X POST http://localhost:3112/api/message/send \\`);
    console.log(`     -H "Authorization: Bearer <TOKEN>" \\`);
    console.log(`     -H "Content-Type: application/json" \\`);
    console.log(`     -d '{"to_user_id":${USER_ID},"conversation_type":1,"msg_type":1,"content":"Hello 1001!"}'`);
  }
});

ws.on('error', (err) => {
  console.error('❌ WebSocket 错误:', err.message);
});

ws.on('close', () => {
  console.log('🔌 WebSocket 连接关闭');
  process.exit(0);
});

// 保持运行
console.log(`开始监听用户 ${USER_ID} 的消息...`);
console.log(`按 Ctrl+C 退出\n`);

setTimeout(() => {
  console.log('\n⏱️  60秒后自动退出...');
  ws.close();
}, 60000);
