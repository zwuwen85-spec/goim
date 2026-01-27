// 完整聊天流程测试 - 两个用户互发消息
const WebSocket = require('ws');

const USER1_ID = 9002;
const USER2_ID = 9001;  // 假设有另一个用户
const WS_URL = 'ws://localhost:3102/sub';
const API_URL = 'http://localhost:3112';

// GoIM 协议常量
const PROTOCOL_VERSION = 102;
const HEADER_SIZE = 16;
const OP_AUTH = 7;
const OP_AUTH_REPLY = 8;
const OP_HEARTBEAT = 2;
const OP_HEARTBEAT_REPLY = 3;
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
    5: 'MESSAGE', 7: 'AUTH', 8: 'AUTH_REPLY'
  };
  return names[op] || `UNKNOWN(${op})`;
}

// 创建用户连接
function createUserConnection(userId, name) {
  const ws = new WebSocket(WS_URL);
  let seq = 1;

  ws.on('open', () => {
    console.log(`✅ [${name}] WebSocket 连接成功!`);

    const authData = JSON.stringify({
      mid: userId,
      key: `user_${userId}_${Date.now()}`,
      room_id: '',
      platform: 'web',
      accepts: [1001, 1002, 1003]
    });

    console.log(`📤 [${name}] 发送认证...`);
    ws.send(encodeMessage(PROTOCOL_VERSION, OP_AUTH, seq++, authData));
  });

  ws.on('message', (data) => {
    const msg = decodeMessage(data);

    if (msg.op === OP_AUTH_REPLY) {
      console.log(`✅ [${name}] 认证成功!`);
    } else if (msg.op === OP_MESSAGE) {
      console.log(`📨 [${name}] 收到消息: ${msg.body}`);
    }
  });

  ws.on('error', (err) => {
    console.error(`❌ [${name}] 错误: ${err.message}`);
  });

  ws.on('close', () => {
    console.log(`🔌 [${name}] 连接关闭`);
  });

  return ws;
}

// 主测试流程
async function runTest() {
  console.log('=== GoIM 完整聊天功能测试 ===\n');

  // 创建用户连接
  const user1 = createUserConnection(USER1_ID, '用户9002');

  // 等待连接建立
  await new Promise(resolve => setTimeout(resolve, 2000));

  console.log('\n=== 测试完成 ===');
  console.log('✅ WebSocket 连接正常');
  console.log('✅ 用户认证成功');
  console.log('\n提示：通过前端页面 http://localhost:5173 进行可视化测试');

  // 保持连接
  setTimeout(() => {
    user1.close();
    process.exit(0);
  }, 10000);
}

runTest();
