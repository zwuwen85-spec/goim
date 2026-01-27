// 双用户测试 - 两个用户同时在线互发消息
const WebSocket = require('ws');

const USER1_ID = 9001;
const USER2_ID = 9002;
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
    5: 'MESSAGE', 7: 'AUTH', 8: 'AUTH_REPLY'
  };
  return names[op] || `UNKNOWN(${op})`;
}

// 创建用户连接
function createUserConnection(userId, name, onMessage) {
  const ws = new WebSocket(WS_URL);
  let seq = 1;

  ws.on('open', () => {
    console.log(`✅ [${name}] WebSocket 连接成功!`);

    const authData = JSON.stringify({
      mid: userId,
      key: `user_${userId}`,
      room_id: '',
      platform: 'web',
      accepts: [1001, 1002, 1003]
    });

    console.log(`📤 [${name}] 发送认证... key=user_${userId}`);
    ws.send(encodeMessage(PROTOCOL_VERSION, OP_AUTH, seq++, authData));
  });

  ws.on('message', (data) => {
    const msg = decodeMessage(data);
    console.log(`📥 [${name}] OP=${getOpName(msg.op)} (${msg.op}), SEQ=${msg.seq}`);

    if (msg.op === OP_AUTH_REPLY) {
      console.log(`   ✅ [${name}] 认证成功!`);
    } else if (msg.op === OP_MESSAGE && msg.body) {
      console.log(`   📨 [${name}] 收到消息: ${msg.body}`);
      if (onMessage) onMessage(msg.body);
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
  console.log('=== 双用户在线消息测试 ===\n');

  let user1WS, user2WS;
  let user1Ready = false;
  let user2Ready = false;

  // 创建用户1连接
  user1WS = createUserConnection(USER1_ID, `用户${USER1_ID}`, (msg) => {
    console.log(`\n🎉 [用户${USER1_ID}] 收到来自 [用户${USER2_ID}] 的消息!`);
  });

  // 等待用户1连接
  await new Promise(resolve => setTimeout(resolve, 2000));
  user1Ready = true;

  // 创建用户2连接
  user2WS = createUserConnection(USER2_ID, `用户${USER2_ID}`, (msg) => {
    console.log(`\n🎉 [用户${USER2_ID}] 收到来自 [用户${USER1_ID}] 的消息!`);
  });

  // 等待用户2连接
  await new Promise(resolve => setTimeout(resolve, 2000));
  user2Ready = true;

  if (user1Ready && user2Ready) {
    console.log('\n=== 两个用户都已在线，通过 API 测试发送消息 ===\n');
    console.log('提示：现在可以在另一个终端发送测试消息：\n');

    console.log(`curl -X POST http://localhost:3112/api/message/send \\
  -H "Authorization: Bearer <TOKEN>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "to_user_id": ${USER1_ID},
    "conversation_type": 1,
    "msg_type": 1,
    "content": "Hello from API!"
  }'\n`);

    console.log('保持连接运行，等待接收消息...\n');
  }

  // 保持连接 60 秒
  setTimeout(() => {
    console.log('\n=== 测试结束 ===');
    if (user1WS) user1WS.close();
    if (user2WS) user2WS.close();
    process.exit(0);
  }, 60000);
}

runTest();
