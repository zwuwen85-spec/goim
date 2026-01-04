<template>
  <div class="main-layout">
    <!-- 1. Icon Sidebar -->
    <div class="icon-sidebar">
      <div class="app-logo">
        <el-icon :size="32" color="#6366f1"><ChatDotRound /></el-icon>
      </div>
      
      <div class="nav-items">
        <div 
          class="nav-item" 
          :class="{ active: activeTab === 'chats' }" 
          @click="activeTab = 'chats'"
          title="聊天"
        >
          <el-icon :size="24"><ChatLineRound /></el-icon>
        </div>
        <div 
          class="nav-item" 
          :class="{ active: activeTab === 'friends' }" 
          @click="activeTab = 'friends'"
          title="好友"
        >
          <el-icon :size="24"><User /></el-icon>
        </div>
        <div 
          class="nav-item" 
          :class="{ active: activeTab === 'groups' }" 
          @click="activeTab = 'groups'"
          title="群组"
        >
          <el-icon :size="24"><Connection /></el-icon>
        </div>
        <div 
          class="nav-item" 
          :class="{ active: activeTab === 'ai' }" 
          @click="activeTab = 'ai'"
          title="AI 助手"
        >
          <el-icon :size="24"><Cpu /></el-icon>
        </div>
      </div>

      <div class="bottom-actions">
        <el-dropdown trigger="click" placement="right-end">
          <el-avatar :size="40" :src="userStore.currentUser?.avatar" class="user-avatar">
            {{ userStore.currentUser?.nickname?.[0] }}
          </el-avatar>
          <template #dropdown>
            <el-dropdown-menu>
              <div class="user-dropdown-header">
                <div class="dropdown-name">{{ userStore.currentUser?.nickname }}</div>
                <div class="dropdown-status">在线</div>
              </div>
              <el-dropdown-item divided @click="handleLogout">
                <el-icon><SwitchButton /></el-icon>退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <!-- 2. List Sidebar -->
    <div class="list-sidebar">
      <!-- Search Bar -->
      <div class="sidebar-search">
        <el-input 
          v-model="searchText" 
          placeholder="搜索..." 
          prefix-icon="Search"
          class="custom-input"
        />
      </div>
      
      <!-- Content Lists -->
      <div class="sidebar-content">
        <ConversationList 
          v-if="activeTab === 'chats'" 
          @select="handleConversationSelect" 
        />
        <FriendManager 
          v-else-if="activeTab === 'friends'" 
          @chat="handleFriendChat" 
        />
        <GroupList 
          v-else-if="activeTab === 'groups'" 
          @select="handleGroupSelect" 
        />
        <BotList 
          v-else-if="activeTab === 'ai'" 
          @select="handleBotSelect" 
        />
      </div>
    </div>

    <!-- 3. Main Content -->
    <div class="main-content">
      <template v-if="activeSessionType">
        <ChatWindow 
          :messages="currentMessages"
          :current-user-id="userStore.currentUser?.id || 0"
          :title="sessionTitle"
          :subtitle="sessionSubtitle"
          :avatar="sessionAvatar"
          :avatar-style="sessionAvatarStyle"
          :sending="isSending"
          :loading="isLoadingMessages"
          :show-sender-name="activeSessionType === 'group'"
          :get-sender-name="getSenderName"
          :get-sender-avatar="getSenderAvatar"
          :get-sender-style="getSenderStyle"
          @send="handleSendMessage"
        >
          <template #actions>
            <el-button text circle v-if="activeSessionType === 'group'" @click="showGroupMembers">
              <el-icon><MoreFilled /></el-icon>
            </el-button>
            <el-button text circle v-if="activeSessionType === 'ai'" @click="clearAiChat">
              <el-icon><Delete /></el-icon>
            </el-button>
          </template>
        </ChatWindow>
      </template>
      
      <div v-else class="empty-state">
        <div class="empty-content">
          <img src="https://trae-api-sg.mchost.guru/api/ide/v1/text_to_image?prompt=minimalist%20chat%20illustration%20vector%20flat%20design%20blue&image_size=square_hd" alt="Empty State" class="empty-img" />
          <h3>开始聊天</h3>
          <p>选择一个联系人或群组开始对话</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  ChatLineRound, User, Connection, Cpu, 
  SwitchButton, ChatDotRound, MoreFilled, Delete 
} from '@element-plus/icons-vue'

import { useUserStore } from '../store/user'
import { useChatStore } from '../store/chat'
import { useGroupStore } from '../store/group'
import { useAIStore } from '../store/ai'
import { useWebSocket } from '../utils/websocket'
import { parseSqlNullString } from '../utils/format'

import ConversationList from '../components/ConversationList.vue'
import FriendManager from '../components/FriendManager.vue'
import GroupList from '../components/GroupList.vue'
import BotList from '../components/BotList.vue'
import ChatWindow from '../components/ChatWindow.vue'

const router = useRouter()
const userStore = useUserStore()
const chatStore = useChatStore()
const groupStore = useGroupStore()
const aiStore = useAIStore()

// State
const activeTab = ref('chats')
const searchText = ref('')
const activeSessionType = ref<'private' | 'group' | 'ai' | null>(null)

// WebSocket
const { messages: wsMessages, connect, disconnect } = useWebSocket('ws://localhost:3102/sub')

// Computed for ChatWindow
const currentMessages = computed(() => {
  if (activeSessionType.value === 'ai') {
    return aiStore.currentBot ? aiStore.getBotMessages(aiStore.currentBot.id) : []
  }
  return chatStore.currentSession?.messages || []
})

const sessionTitle = computed(() => {
  if (activeSessionType.value === 'ai') return aiStore.currentBot?.name || 'AI'
  return chatStore.currentSession?.name || 'Chat'
})

const sessionSubtitle = computed(() => {
  if (activeSessionType.value === 'ai') return aiStore.currentBot?.personality || 'Assistant'
  if (activeSessionType.value === 'group') return `${groupStore.members.length} members`
  return 'Online'
})

const sessionAvatar = computed(() => {
  if (activeSessionType.value === 'ai') return '' // Use style
  return chatStore.currentSession?.avatar
})

const sessionAvatarStyle = computed(() => {
  if (activeSessionType.value === 'ai' && aiStore.currentBot) {
    const colors: Record<number, string> = {
      9001: '#409EFF', 9002: '#67C23A', 9003: '#E6A23C', 9004: '#F56C6C'
    }
    return { backgroundColor: colors[aiStore.currentBot.id] || '#909399' }
  }
  if (activeSessionType.value === 'group') {
     // Generate color based on ID
     const colors = ['#F56C6C', '#E6A23C', '#67C23A', '#409EFF', '#909399']
     const id = chatStore.currentSession?.targetId || 0
     return { backgroundColor: colors[id % colors.length] }
  }
  return {}
})

const isSending = computed(() => {
  return activeSessionType.value === 'ai' ? aiStore.sending : false
})

const isLoadingMessages = computed(() => {
  // Can add loading state logic here
  return false
})

// Handlers
const handleConversationSelect = async (conv: any) => {
  activeSessionType.value = conv.conversation_type === 2 ? 'group' : 'private'
  const name = parseSqlNullString(conv.target_user?.nickname)
  const avatar = parseSqlNullString(conv.target_user?.avatar)
  await chatStore.openChat(conv.target_id, conv.conversation_type, name, avatar)
  
  if (activeSessionType.value === 'group') {
    // Sync group store if needed
    // groupStore.setCurrentGroup(...) // logic needs to match
    await loadGroupMembers(conv.target_id)
  }
}

const handleFriendChat = (friend: any) => {
  const userId = friend.friend_user?.id || friend.friend_id
  activeSessionType.value = 'private'
  chatStore.openChat(userId, 1, friend.remark || friend.friend_user?.nickname)
  activeTab.value = 'chats' // Switch to chat tab
}

const handleGroupSelect = async (group: any) => {
  activeSessionType.value = 'group'
  await groupStore.setCurrentGroup(group)
  chatStore.openChat(group.id, 2, group.name)
  await loadGroupMembers(group.id)
  // activeTab.value = 'chats' // Optional: stay on groups or switch? Switching is better for "Chat" context
}

const handleBotSelect = (bot: any) => {
  activeSessionType.value = 'ai'
  aiStore.setCurrentBot(bot)
}

const handleSendMessage = async (content: string) => {
  if (activeSessionType.value === 'ai') {
    if (aiStore.currentBot) {
      await aiStore.sendMessage(aiStore.currentBot.id, content)
    }
  } else {
    // Private or Group
    const type = activeSessionType.value === 'group' ? 2 : 1
    await chatStore.sendMessage(JSON.stringify({ text: content }), type)
  }
}

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
    })
    disconnect()
    userStore.logout()
    router.push('/login')
  } catch {}
}

const showGroupMembers = () => {
  ElMessage.info('查看群成员功能待实现 (Refactored)')
  // logic to show dialog
}

const clearAiChat = async () => {
  if (aiStore.currentBot) {
     aiStore.clearMessages(aiStore.currentBot.id)
  }
}

// Helpers for ChatWindow
const groupMembersCache = ref<Map<number, Map<number, any>>>(new Map())

const loadGroupMembers = async (groupId: number) => {
  // Logic from original Main.vue to load members
  // This is needed for getSenderName/Avatar in groups
  if (!groupMembersCache.value.has(groupId)) {
    // Call API... reused logic
    await groupStore.loadGroups() // Simplification
    // Actually we need to call groupApi.getMembers
    // For now, rely on groupStore.members if it's the current group
  }
}

const getSenderName = (msg: any) => {
  if (msg.role) return msg.role === 'user' ? '我' : (aiStore.currentBot?.name || 'AI')
  
  if (msg.from_user_id === userStore.currentUser?.id) return '我'
  
  if (activeSessionType.value === 'group') {
    const member = groupStore.members.find(m => m.user_id === msg.from_user_id)
    return member?.user?.nickname || member?.nickname || `User ${msg.from_user_id}`
  }
  
  return chatStore.currentSession?.name || 'User'
}

const getSenderAvatar = (msg: any): string => {
  if (msg.role === 'assistant') return '' // Use style
  if (msg.from_user_id === userStore.currentUser?.id) return userStore.currentUser?.avatar || ''
  
  if (activeSessionType.value === 'group') {
    const member = groupStore.members.find(m => m.user_id === msg.from_user_id)
    return member?.user?.avatar || ''
  }
  
  return chatStore.currentSession?.avatar || ''
}

const getSenderStyle = (msg: any) => {
  if (msg.role === 'assistant' && aiStore.currentBot) {
     // Reusing color logic
     return sessionAvatarStyle.value
  }
  return {}
}

// Lifecycle
onMounted(async () => {
  if (!userStore.isLoggedIn) return
  
  await Promise.all([
    chatStore.loadConversations(),
    chatStore.loadFriends()
  ])
  
  try {
    connect(userStore.token, userStore.currentUser!.id)
  } catch (e) {
    console.warn('WS failed', e)
  }
})

// Watch WS messages
watch(wsMessages, (newMessages) => {
  if (newMessages.length > 0) {
    const latestMsg = newMessages[newMessages.length - 1]
    if (latestMsg.conversation_type === 1 || latestMsg.conversation_type === 2) {
      // Convert WS message to Store Message
      chatStore.addMessage({
        id: Date.now(), // Generate a temporary ID
        msg_id: latestMsg.msg_id,
        from_user_id: latestMsg.from_user_id,
        conversation_id: latestMsg.conversation_id,
        conversation_type: latestMsg.conversation_type,
        msg_type: latestMsg.msg_type,
        content: latestMsg.content,
        seq: latestMsg.seq,
        created_at: new Date(latestMsg.created_at).toISOString()
      })
    }
  }
})

</script>

<style scoped>
.main-layout {
  display: flex;
  width: 100vw;
  height: 100vh;
  background-color: var(--bg-body);
  overflow: hidden;
}

/* 1. Icon Sidebar */
.icon-sidebar {
  width: 72px;
  background-color: var(--bg-surface);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 0;
  z-index: 10;
}

.app-logo {
  margin-bottom: 40px;
}

.nav-items {
  display: flex;
  flex-direction: column;
  gap: 24px;
  flex: 1;
}

.nav-item {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  cursor: pointer;
  color: var(--text-light);
  transition: all 0.2s;
}

.nav-item:hover {
  background-color: var(--primary-light);
  color: var(--primary-color);
}

.nav-item.active {
  background-color: var(--primary-color);
  color: white;
  box-shadow: var(--shadow-md);
}

.bottom-actions {
  margin-top: auto;
}

.user-avatar {
  cursor: pointer;
  border: 2px solid transparent;
  transition: all 0.2s;
}

.user-avatar:hover {
  border-color: var(--primary-color);
}

.user-dropdown-header {
  padding: 8px 16px;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 8px;
}

.dropdown-name {
  font-weight: 600;
  color: var(--text-primary);
}

.dropdown-status {
  font-size: 12px;
  color: #67c23a;
}

/* 2. List Sidebar */
.list-sidebar {
  width: 320px;
  background-color: var(--bg-surface);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
}

.sidebar-search {
  padding: 20px;
}

.custom-input :deep(.el-input__wrapper) {
  background-color: var(--bg-body);
  box-shadow: none;
  border-radius: 8px;
}

.sidebar-content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* 3. Main Content */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  background-color: var(--bg-chat);
  position: relative;
}

.empty-state {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-light);
}

.empty-content {
  text-align: center;
}

.empty-img {
  width: 200px;
  height: 200px;
  object-fit: cover;
  margin-bottom: 24px;
  opacity: 0.8;
}

.empty-content h3 {
  font-size: 1.5rem;
  color: var(--text-primary);
  margin-bottom: 8px;
}
</style>
