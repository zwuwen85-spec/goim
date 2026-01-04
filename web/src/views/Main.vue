<template>
  <div class="main-container">
    <!-- Sidebar -->
    <div class="sidebar">
      <!-- User info -->
      <div class="user-info">
        <el-avatar :size="40" :src="userStore.currentUser?.avatar">
          {{ userStore.currentUser?.nickname?.[0] }}
        </el-avatar>
        <div class="user-details">
          <div class="user-name">{{ userStore.currentUser?.nickname }}</div>
          <div class="user-status">在线</div>
        </div>
        <el-button text @click="handleLogout">
          <el-icon><SwitchButton /></el-icon>
        </el-button>
      </div>

      <!-- Tabs -->
      <el-tabs v-model="activeTab" class="sidebar-tabs">
        <!-- Conversations -->
        <el-tab-pane label="聊天" name="conversations">
          <div class="tab-content">
            <div
              v-for="conv in chatStore.conversations"
              :key="`${conv.target_id}:${conv.conversation_type}`"
              class="conversation-item"
              :class="{ active: currentConvId === `${conv.target_id}:${conv.conversation_type}` }"
              @click="openConversation(conv)"
            >
              <el-avatar :size="46">{{ conv.target_user?.nickname?.[0] || '?' }}</el-avatar>
              <div class="conv-info">
                <div class="conv-name">{{ conv.target_user?.nickname || `User ${conv.target_id}` }}</div>
                <div class="conv-preview">{{ getLastMessage(conv) }}</div>
              </div>
              <el-badge v-if="conv.unread_count > 0" :value="conv.unread_count > 99 ? '99+' : conv.unread_count" class="unread-badge" />
            </div>
          </div>
        </el-tab-pane>

        <!-- Friends -->
        <el-tab-pane label="好友" name="friends">
          <div class="tab-content friends-tab">
            <FriendManager />
          </div>
        </el-tab-pane>

        <!-- Groups -->
        <el-tab-pane label="群聊" name="groups">
          <div class="tab-content groups-tab">
            <GroupChat />
          </div>
        </el-tab-pane>

        <!-- AI Chat -->
        <el-tab-pane label="AI" name="ai">
          <div class="tab-content ai-tab">
            <AIChat />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- Chat area -->
    <div class="chat-area">
      <div v-if="chatStore.currentSession" class="chat-container">
        <!-- Chat header -->
        <div class="chat-header">
          <h3>{{ chatStore.currentSession.name }}</h3>
          <el-button text @click="showUserInfo = true">
            <el-icon><InfoFilled /></el-icon>
          </el-button>
        </div>

        <!-- Messages -->
        <div class="messages-container" ref="messagesRef">
          <div
            v-for="msg in chatStore.currentSession.messages"
            :key="msg.msg_id"
            class="message"
            :class="{ 'is-me': isFromMe(msg) }"
          >
            <!-- 头像 - 每条消息都显示 -->
            <el-avatar
              :size="36"
              :src="getSenderInfo(msg).avatar"
            >
              {{ getSenderInfo(msg).nickname?.[0] || getSenderInfo(msg).name?.[0] || '?' }}
            </el-avatar>
            <div class="message-content">
              <div class="message-header">
                <span class="message-sender">{{ getSenderInfo(msg).name }}</span>
                <span class="message-time">{{ formatTime(msg.created_at) }}</span>
              </div>
              <div class="message-body">{{ parseContent(msg.content) }}</div>
            </div>
          </div>
        </div>

        <!-- Input area -->
        <div class="input-area">
          <el-input
            v-model="messageInput"
            type="textarea"
            :rows="3"
            placeholder="输入消息..."
            @keydown.enter.exact="handleSend"
          />
          <div class="input-actions">
            <el-button type="primary" @click="handleSend" :disabled="!messageInput.trim()">
              发送
            </el-button>
          </div>
        </div>
      </div>

      <!-- Empty state -->
      <div v-else class="empty-state">
        <el-empty description="选择一个聊天或好友开始对话" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick, watch } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { useUserStore } from '../store/user'
import { useChatStore } from '../store/chat'
import { useGroupStore } from '../store/group'
import { useWebSocket } from '../utils/websocket'
import { authApi, groupApi, type User, type GroupMember } from '../api/chat'
import type { Message as ChatMessage } from '../utils/websocket'
import AIChat from '../components/AIChat.vue'
import GroupChat from '../components/GroupChat.vue'
import FriendManager from '../components/FriendManager.vue'

const router = useRouter()
const userStore = useUserStore()
const chatStore = useChatStore()
const groupStore = useGroupStore()

const activeTab = ref('conversations')
const messageInput = ref('')
const messagesRef = ref<HTMLElement>()
const currentConvId = ref('')

// 群成员缓存：groupId -> Map<userId, GroupMember>
const groupMembersCache = ref<Map<number, Map<number, GroupMember>>>(new Map())

// WebSocket
const { status: wsStatus, messages: wsMessages, connect, disconnect, clearMessages } = useWebSocket('ws://localhost:3102/sub')

const isFromMe = (msg: any) => {
  return msg.from_user_id === userStore.currentUser?.id
}

// 加载群成员信息
const loadGroupMembers = async (groupId: number) => {
  if (groupMembersCache.value.has(groupId)) {
    return
  }

  try {
    const response = await groupApi.getMembers(groupId)
    if (response.code === 0) {
      const membersMap = new Map<number, GroupMember>()
      response.data.members?.forEach((member: GroupMember) => {
        membersMap.set(member.user_id, member)
      })
      groupMembersCache.value.set(groupId, membersMap)
    }
  } catch (error) {
    console.error('Failed to load group members:', error)
  }
}

// 获取发送者信息
const getSenderInfo = (msg: any) => {
  const myId = userStore.currentUser?.id

  // 如果是自己的消息
  if (msg.from_user_id === myId) {
    return {
      name: '我',
      avatar: userStore.currentUser?.avatar || '',
      nickname: userStore.currentUser?.nickname || '我'
    }
  }

  // 如果是群聊消息
  if (chatStore.currentSession?.targetType === 'group') {
    const membersMap = groupMembersCache.value.get(chatStore.currentSession.targetId)
    const member = membersMap?.get(msg.from_user_id)
    if (member?.user) {
      // 处理 SQL.NullString 类型的 avatar_url
      const avatar = member.user.avatar_url?.Valid ? member.user.avatar_url.String : member.user.avatar_url || ''
      return {
        name: member.user.nickname,
        avatar: avatar,
        nickname: member.user.nickname
      }
    }
  }

  // 私聊消息：使用会话名称
  return {
    name: chatStore.currentSession?.name || '用户',
    avatar: chatStore.currentSession?.avatar || '',
    nickname: chatStore.currentSession?.name || '用户'
  }
}

const formatTime = (timestamp: string | number) => {
  let date: Date
  if (typeof timestamp === 'string') {
    // ISO format string like "2026-01-04T15:12:15+08:00"
    date = new Date(timestamp)
  } else if (timestamp < 10000000000) {
    // Unix timestamp in seconds
    date = new Date(timestamp * 1000)
  } else {
    // Unix timestamp in milliseconds
    date = new Date(timestamp)
  }
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

const parseContent = (content: string) => {
  try {
    const parsed = JSON.parse(content)
    return parsed.text || content
  } catch {
    return content
  }
}

const getLastMessage = (conv: any) => {
  if (!conv.last_msg_content) return '暂无消息'
  try {
    const parsed = JSON.parse(conv.last_msg_content)
    return parsed.text || conv.last_msg_content
  } catch {
    return conv.last_msg_content
  }
}

const openConversation = async (conv: any) => {
  currentConvId.value = `${conv.target_id}:${conv.conversation_type}`
  chatStore.openChat(conv.target_id, conv.conversation_type, conv.target_user?.nickname)

  // 如果是群聊，加载群成员信息
  if (conv.conversation_type === 2) {
    await loadGroupMembers(conv.target_id)
  }
}

const handleSend = async () => {
  if (!messageInput.value.trim() || !chatStore.currentSession) return

  const success = await chatStore.sendMessage(JSON.stringify({ text: messageInput.value }), 1)
  if (success) {
    messageInput.value = ''
    await nextTick()
    scrollToBottom()
  } else {
    ElMessage.error('发送失败')
  }
}

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    disconnect()
    userStore.logout()
    router.push('/login')
  } catch {
    // Cancelled
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
}

// Watch for new messages from WebSocket
watch(wsMessages, (newMessages) => {
  if (newMessages.length > 0) {
    const latestMsg = newMessages[newMessages.length - 1]
    if (latestMsg.conversation_type === 1 || latestMsg.conversation_type === 2) {
      chatStore.addMessage(latestMsg)
      if (chatStore.currentSession?.id === `${latestMsg.conversation_id}:${latestMsg.conversation_type}`) {
        scrollToBottom()
      }
    }
  }
})

// Watch for session changes to scroll to bottom
watch(() => chatStore.currentSession, async (newSession) => {
  if (newSession && newSession.targetType === 'group') {
    await loadGroupMembers(newSession.targetId)
  }
  nextTick(() => {
    scrollToBottom()
  })
})

onMounted(async () => {
  if (!userStore.isLoggedIn) return

  // Load initial data
  await Promise.all([
    chatStore.loadConversations(),
    chatStore.loadFriends()
  ])

  // Connect WebSocket (don't block if it fails)
  try {
    connect(userStore.token, userStore.currentUser!.id)
  } catch (error) {
    console.warn('WebSocket connection failed, continuing without it:', error)
    // WebSocket is optional for basic functionality
    ElMessage.warning({
      message: '实时通信未连接，刷新页面获取最新消息',
      duration: 3000,
      showClose: true
    })
  }
})

onUnmounted(() => {
  try {
    disconnect()
  } catch (error) {
    console.warn('Error disconnecting WebSocket:', error)
  }
})
</script>

<style scoped>
.main-container {
  display: flex;
  height: 100vh;
}

.sidebar {
  width: 280px;
  background: #fff;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
}

.user-info {
  display: flex;
  align-items: center;
  padding: 16px;
  gap: 12px;
  border-bottom: 1px solid #e4e7ed;
}

.user-details {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.user-status {
  font-size: 12px;
  color: #67c23a;
}

.sidebar-tabs {
  flex: 1;
  overflow: hidden;
}

.sidebar-tabs :deep(.el-tabs__content) {
  height: calc(100% - 40px);
}

.sidebar-tabs :deep(.el-tab-pane) {
  height: 100%;
}

.tab-content {
  height: 100%;
  overflow-y: auto;
  padding: 12px;
}

.conversation-item,
.friend-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  margin-bottom: 4px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.25s ease;
  position: relative;
}

.conversation-item :deep(.el-avatar),
.friend-item :deep(.el-avatar) {
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border: 2px solid #fff;
  transition: all 0.25s ease;
}

.conversation-item:hover :deep(.el-avatar),
.friend-item:hover :deep(.el-avatar) {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.conversation-item:hover,
.friend-item:hover {
  background: #f5f7fa;
  transform: translateX(2px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.conversation-item.active {
  background: linear-gradient(135deg, #ecf5ff 0%, #e1f0ff 100%);
  box-shadow: 0 2px 12px rgba(64, 158, 255, 0.15);
}

.conversation-item.active .conv-name {
  color: #409eff;
  font-weight: 600;
}

.conv-info,
.friend-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.conv-name,
.friend-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  line-height: 1.4;
}

.conv-preview {
  font-size: 12px;
  color: #909399;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.3;
}

.friend-remark {
  font-size: 12px;
  color: #909399;
}

.unread-badge {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
}

.unread-badge :deep(.el-badge__content) {
  background: linear-gradient(135deg, #ff6b6b 0%, #ff5252 100%);
  border: 2px solid #fff;
  box-shadow: 0 2px 8px rgba(255, 82, 82, 0.3);
  font-weight: 600;
}

.search-box {
  padding: 8px;
}

.search-results {
  margin-top: 10px;
  border-top: 1px solid #e4e7ed;
  padding-top: 10px;
}

.search-title {
  font-size: 12px;
  color: #909399;
  padding: 0 8px 8px;
}

.search-result-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  cursor: pointer;
}

.search-result-item:hover {
  background: #f5f7fa;
}

.chat-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #f5f7fa;
}

.chat-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
}

.chat-header h3 {
  margin: 0;
  font-size: 16px;
  color: #303133;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.message {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}

.message.is-me {
  flex-direction: row-reverse;
}

.message-content {
  max-width: 60%;
}

.message-header {
  display: flex;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 12px;
}

.message-sender {
  font-weight: 500;
  color: #606266;
}

.message-time {
  color: #909399;
}

.message-body {
  padding: 10px 14px;
  background: #fff;
  border-radius: 8px;
  word-break: break-word;
}

.message.is-me .message-body {
  background: #95ec69;
}

.input-area {
  padding: 16px 20px;
  background: #fff;
  border-top: 1px solid #e4e7ed;
}

.input-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
}

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* AI Tab Styles */
.ai-tab {
  height: 100%;
  padding: 0;
}

.ai-tab :deep(.ai-chat-container) {
  height: 100%;
}

/* Groups Tab Styles */
.groups-tab {
  height: 100%;
  padding: 0;
}

.groups-tab :deep(.group-chat-container) {
  height: 100%;
}

/* Friends Tab Styles */
.friends-tab {
  height: 100%;
  padding: 0;
}

.friends-tab :deep(.friend-manager-container) {
  height: 100%;
}
</style>
