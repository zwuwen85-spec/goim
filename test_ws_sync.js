// WebSocket Real-time Message Sync Test
// This script tests that when User A sends a message to User B,
// User B receives it via WebSocket in real-time

const http = require('http');
const WebSocket = require('ws');

const API_BASE = 'http://localhost:3112/api';
const WS_URL = 'ws://localhost:3102/sub';

// Test configuration
const USER1 = { username: 'testuser', password: '123456', userId: 9002, targetId: 9006 };
const USER2 = { username: 'testuser2', password: 'password123', userId: 9006, targetId: 9002 };

let user1Token, user2Token;
let user1WS, user2WS;
let messagesReceived = {
    user1: [],
    user2: []
};
let testResults = {
    user1Login: false,
    user2Login: false,
    user1WSConnected: false,
    user2WSConnected: false,
    user1WSAuthed: false,
    user2WSAuthed: false,
    messageSent: false,
    user2Received: false,
    user1ReceivedReply: false
};

// HTTP request helper
function httpRequest(method, path, data, token = null) {
    return new Promise((resolve, reject) => {
        const options = {
            hostname: 'localhost',
            port: 3112,
            path: `/api${path}`,
            method,
            headers: {
                'Content-Type': 'application/json'
            }
        };

        if (token) {
            options.headers['Authorization'] = `Bearer ${token}`;
        }

        const req = http.request(options, (res) => {
            let body = '';
            res.on('data', chunk => body += chunk);
            res.on('end', () => {
                try {
                    resolve(JSON.parse(body));
                } catch (e) {
                    reject(new Error(`Parse error: ${body}`));
                }
            });
        });

        req.on('error', reject);
        if (data) req.write(JSON.stringify(data));
        req.end();
    });
}

// Login function
async function login(username, password) {
    const data = await httpRequest('POST', '/user/login', { username, password });
    if (data.code !== 0) throw new Error(`Login failed: ${data.message}`);
    return { token: data.data.token, userId: data.data.user_id };
}

// Send message function
async function sendMessage(token, toUserId, content) {
    const data = await httpRequest('POST', '/message/send', {
        to_user_id: toUserId,
        conversation_type: 1,
        msg_type: 1,
        content
    }, token);
    if (data.code !== 0) throw new Error(`Send failed: ${data.message}`);
    return data;
}

// Encode WebSocket message (goim uses BIG endian)
function encodeMessage(ver, op, seq, body) {
    const payload = body ? Buffer.from(body) : Buffer.alloc(0);
    const packLen = 16 + payload.length;
    const buffer = Buffer.alloc(packLen);

    // goim protocol uses BIG endian
    buffer.writeUInt32BE(packLen, 0);
    buffer.writeUInt16BE(16, 4);
    buffer.writeUInt16BE(ver, 6);
    buffer.writeUInt32BE(op, 8);
    buffer.writeUInt32BE(seq, 12);

    if (payload.length > 0) {
        payload.copy(buffer, 16);
    }

    return buffer;
}

// Decode WebSocket message (goim uses BIG endian)
function decodeMessage(buffer) {
    if (buffer.length < 16) return null;

    // goim protocol uses BIG endian
    const packLen = buffer.readUInt32BE(0);
    const headerLen = buffer.readUInt16BE(4);
    const ver = buffer.readUInt16BE(6);
    const op = buffer.readUInt32BE(8);
    const seq = buffer.readUInt32BE(12);
    const body = buffer.slice(16).toString('utf8');

    return { packLen, headerLen, ver, op, seq, body };
}

// Connect WebSocket
function connectWebSocket(userId, token, userLabel) {
    return new Promise((resolve, reject) => {
        const ws = new WebSocket(WS_URL, { handshakeTimeout: 5000 });

        const timeout = setTimeout(() => {
            ws.close();
            reject(new Error(`${userLabel}: WebSocket connection timeout`));
        }, 10000);

        ws.on('open', () => {
            console.log(`✅ ${userLabel}: WebSocket connected`);
            clearTimeout(timeout);

            // Send auth
            const authData = JSON.stringify({
                mid: userId,
                key: `user_${userId}`,
                room_id: '',
                platform: 'web'
            });
            ws.send(encodeMessage(1, 7, 1, authData));
            console.log(`✅ ${userLabel}: Auth sent`);
        });

        ws.on('message', (data) => {
            const msg = decodeMessage(data);
            if (!msg) return;

            if (msg.op === 8) {
                // Auth success
                console.log(`✅ ${userLabel}: WebSocket authenticated`);
                clearTimeout(timeout);
                resolve(ws);
            } else if (msg.op === 3) {
                // Heartbeat reply
            } else if (msg.op === 5 && msg.body) {
                // Data message
                try {
                    const messageData = JSON.parse(msg.body);
                    console.log(`📨 ${userLabel}: Received message: "${messageData.content}" from User ${messageData.from_user_id}`);

                    if (userLabel === 'User 2') {
                        messagesReceived.user2.push(messageData);
                        testResults.user2Received = true;
                    } else {
                        messagesReceived.user1.push(messageData);
                        testResults.user1ReceivedReply = true;
                    }
                } catch (e) {
                    console.log(`📨 ${userLabel}: Received raw message: ${msg.body}`);
                }
            }
        });

        ws.on('error', (error) => {
            clearTimeout(timeout);
            console.error(`❌ ${userLabel}: WebSocket error: ${error.message}`);
            reject(error);
        });

        ws.on('close', () => {
            console.log(`🔌 ${userLabel}: WebSocket closed`);
        });
    });
}

// Run test
async function runTest() {
    console.log('\n========================================');
    console.log('🧪 WebSocket Real-time Message Sync Test');
    console.log('========================================\n');

    try {
        // Step 1: Login both users
        console.log('📝 Step 1: Logging in users...');
        const user1 = await login(USER1.username, USER1.password);
        user1Token = user1.token;
        console.log(`✅ User 1 (testuser) logged in, ID: ${user1.userId}`);

        const user2 = await login(USER2.username, USER2.password);
        user2Token = user2.token;
        console.log(`✅ User 2 (testuser2) logged in, ID: ${user2.userId}`);
        testResults.user1Login = true;
        testResults.user2Login = true;
        console.log('');

        // Step 2: Connect WebSocket for both users
        console.log('🔌 Step 2: Connecting WebSocket...');
        user1WS = await connectWebSocket(user1.userId, user1Token, 'User 1');
        testResults.user1WSConnected = true;
        testResults.user1WSAuthed = true;

        // Small delay
        await new Promise(r => setTimeout(r, 500));

        user2WS = await connectWebSocket(user2.userId, user2Token, 'User 2');
        testResults.user2WSConnected = true;
        testResults.user2WSAuthed = true;
        console.log('');

        // Step 3: Wait for WebSocket to be ready
        console.log('⏳ Step 3: Waiting for connections to stabilize...');
        await new Promise(r => setTimeout(r, 2000));
        console.log('');

        // Step 4: User 1 sends message to User 2
        console.log('📤 Step 4: User 1 sending message to User 2...');
        const testMessage = `WebSocket sync test at ${new Date().toISOString()}`;
        await sendMessage(user1Token, USER2.userId, testMessage);
        console.log(`✅ Message sent: "${testMessage}"`);
        testResults.messageSent = true;
        console.log('');

        // Step 5: Wait for User 2 to receive via WebSocket
        console.log('⏳ Step 5: Waiting for User 2 to receive message via WebSocket...');
        let attempts = 0;
        while (attempts < 10 && messagesReceived.user2.length === 0) {
            await new Promise(r => setTimeout(r, 500));
            attempts++;
        }

        if (messagesReceived.user2.length > 0) {
            console.log(`✅ User 2 received message via WebSocket: "${messagesReceived.user2[0].content}"`);
        } else {
            console.log('❌ User 2 did NOT receive message via WebSocket');
        }
        console.log('');

        // Step 6: User 2 replies
        console.log('📤 Step 6: User 2 sending reply to User 1...');
        const replyMessage = `Reply from User 2 at ${new Date().toISOString()}`;
        await sendMessage(user2Token, USER1.userId, replyMessage);
        console.log(`✅ Reply sent: "${replyMessage}"`);
        console.log('');

        // Step 7: Wait for User 1 to receive reply
        console.log('⏳ Step 7: Waiting for User 1 to receive reply via WebSocket...');
        attempts = 0;
        while (attempts < 10 && messagesReceived.user1.length === 0) {
            await new Promise(r => setTimeout(r, 500));
            attempts++;
        }

        if (messagesReceived.user1.length > 0) {
            console.log(`✅ User 1 received reply via WebSocket: "${messagesReceived.user1[0].content}"`);
        } else {
            console.log('❌ User 1 did NOT receive reply via WebSocket');
        }
        console.log('');

        // Final results
        console.log('========================================');
        console.log('📊 Test Results');
        console.log('========================================');
        console.log(`User 1 Login:         ${testResults.user1Login ? '✅ PASS' : '❌ FAIL'}`);
        console.log(`User 2 Login:         ${testResults.user2Login ? '✅ PASS' : '❌ FAIL'}`);
        console.log(`User 1 WebSocket:     ${testResults.user1WSConnected ? '✅ PASS' : '❌ FAIL'}`);
        console.log(`User 2 WebSocket:     ${testResults.user2WSConnected ? '✅ PASS' : '❌ FAIL'}`);
        console.log(`Message Sent:         ${testResults.messageSent ? '✅ PASS' : '❌ FAIL'}`);
        console.log(`User 2 Received:      ${testResults.user2Received ? '✅ PASS' : '❌ FAIL'}`);
        console.log(`User 1 Received Reply: ${testResults.user1ReceivedReply ? '✅ PASS' : '❌ FAIL'}`);
        console.log('========================================\n');

        const allPassed = testResults.user1Login && testResults.user2Login &&
                         testResults.user1WSConnected && testResults.user2WSConnected &&
                         testResults.messageSent && testResults.user2Received && testResults.user1ReceivedReply;

        if (allPassed) {
            console.log('🎉 ALL TESTS PASSED! WebSocket real-time message sync is working!\n');
        } else {
            console.log('⚠️  Some tests failed. Check the results above.\n');
        }

        // Cleanup
        setTimeout(() => {
            console.log('🧹 Closing connections...');
            if (user1WS) user1WS.close();
            if (user2WS) user2WS.close();
            console.log('✅ Test complete.\n');
            process.exit(allPassed ? 0 : 1);
        }, 1000);

    } catch (error) {
        console.error('\n❌ Test failed with error:', error.message);
        console.error(error.stack);
        process.exit(1);
    }
}

// Start test
runTest();
