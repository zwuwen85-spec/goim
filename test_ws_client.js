#!/usr/bin/env node

const WebSocket = require('ws');

const WS_URL = 'ws://localhost:3102/sub';

// Protocol constants
const OP_HEARTBEAT = 2;
const OP_HEARTBEAT_REPLY = 3;
const OP_MESSAGE = 5;
const OP_AUTH = 7;
const OP_AUTH_REPLY = 8;

let seq = 1;
let heartbeatInterval = null;

// Protocol header format (16 bytes):
// packLen (int32) + headerLen (int16) + ver (int16) + op (int32) + seq (int32)
const HEADER_SIZE = 16;

function encodeMessage(ver, op, seq, body = null) {
    const bodyBytes = body ? Buffer.from(body, 'utf8') : Buffer.alloc(0);
    const bodyLen = bodyBytes.length;
    const packLen = HEADER_SIZE + bodyLen;

    const buffer = Buffer.alloc(HEADER_SIZE + bodyLen);

    buffer.writeInt32BE(packLen, 0);        // total packet length
    buffer.writeInt16BE(HEADER_SIZE, 4);    // header length (always 16)
    buffer.writeInt16BE(ver, 6);            // protocol version
    buffer.writeInt32BE(op, 8);             // operation
    buffer.writeInt32BE(seq, 12);           // sequence

    if (bodyLen > 0) {
        bodyBytes.copy(buffer, HEADER_SIZE);
    }

    return buffer;
}

function decodeMessage(buffer) {
    const packLen = buffer.readInt32BE(0);
    const headerLen = buffer.readInt16BE(4);
    const ver = buffer.readInt16BE(6);
    const op = buffer.readInt32BE(8);
    const seq = buffer.readInt32BE(12);
    const bodyLen = packLen - headerLen;

    let body = null;
    if (bodyLen > 0) {
        body = buffer.toString('utf8', headerLen, headerLen + bodyLen);
    }

    return { ver, op, seq, body };
}

function getOpName(op) {
    switch (op) {
        case OP_HEARTBEAT: return 'Heartbeat';
        case OP_HEARTBEAT_REPLY: return 'Heartbeat Reply';
        case OP_MESSAGE: return 'Message';
        case OP_AUTH: return 'Auth';
        case OP_AUTH_REPLY: return 'Auth Reply';
        default: return `Unknown(${op})`;
    }
}

function log(msg) {
    console.log(`[${new Date().toLocaleTimeString()}] ${msg}`);
}

function connect() {
    log(`正在连接到 ${WS_URL}...`);

    const ws = new WebSocket(WS_URL, {
        headers: {
            'User-Agent': 'goim-test-client'
        },
        binaryType: 'nodebuffer'  // Use Node Buffer instead of ArrayBuffer
    });

    ws.on('open', () => {
        log('WebSocket 连接成功！');
        log('发送认证消息...');

        // Send auth message
        const token = JSON.stringify({
            mid: 123,
            key: 'test-key-123',
            room_id: 'test://room-001',
            platform: 'web',
            accepts: [1000, 1001, 1002]
        });

        const authMsg = encodeMessage(102, OP_AUTH, seq++, token);
        ws.send(authMsg);
        log(`发送认证: ${token}`);

        // Start heartbeat
        startHeartbeat(ws);
    });

    ws.on('message', (data) => {
        const msg = decodeMessage(data);
        log(`收到消息: ver=${msg.ver}, op=${msg.op}(${getOpName(msg.op)}), seq=${msg.seq}`);
        if (msg.body) {
            log(`  Body: ${msg.body}`);
        }

        // Handle auth reply
        if (msg.op === OP_AUTH_REPLY) {
            log('认证成功！');
        }
    });

    ws.on('close', () => {
        log('WebSocket 连接关闭');
        stopHeartbeat();
    });

    ws.on('error', (error) => {
        log(`WebSocket 错误: ${error.message}`);
    });

    return ws;
}

function startHeartbeat(ws) {
    heartbeatInterval = setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
            const msg = encodeMessage(102, OP_HEARTBEAT, seq++);
            ws.send(msg);
            log('发送心跳');
        }
    }, 30000); // 30 seconds
    log('自动心跳已开启 (每30秒)');
}

function stopHeartbeat() {
    if (heartbeatInterval) {
        clearInterval(heartbeatInterval);
        heartbeatInterval = null;
        log('自动心跳已关闭');
    }
}

// Handle Ctrl+C
process.on('SIGINT', () => {
    log('\n正在退出...');
    stopHeartbeat();
    process.exit(0);
});

// Start connection
connect();
