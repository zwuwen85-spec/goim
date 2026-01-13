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
          <el-badge :value="chatStore.friendRequests.length" :hidden="chatStore.friendRequests.length === 0" class="sidebar-badge">
            <el-icon :size="24"><User /></el-icon>
          </el-badge>
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
        <div class="chat-container">
          <ChatWindow
            ref="chatWindowRef"
            :key="activeSessionType + '-' + (chatStore.currentSession?.id || aiStore.currentBot?.id || '0')"
            :messages="currentMessages"
            :current-user-id="userStore.currentUser?.id || 0"
            :title="sessionTitle"
            :subtitle="sessionSubtitle"
            :avatar="sessionAvatar"
            :avatar-style="sessionAvatarStyle"
            :sending="isSending"
            :loading="isLoadingMessages"
            :has-more="hasMoreMessages"
            :show-sender-name="activeSessionType === 'group'"
            :get-sender-name="getSenderName"
            :get-sender-avatar="getSenderAvatar"
            :get-sender-style="getSenderStyle"
            @send="handleSendMessage"
            @load-more="handleLoadMore"
          >
            <template #actions>
              <div class="header-actions-wrapper">
                <el-button 
                  v-show="activeSessionType === 'group'" 
                  key="group-btn"
                  text 
                  circle 
                  @click.stop="toggleGroupSidebar"
                >
                  <el-icon><MoreFilled /></el-icon>
                </el-button>
                <el-button 
                  v-show="activeSessionType === 'ai'" 
                  key="ai-btn"
                  text 
                  circle 
                  @click="clearAiChat"
                >
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </template>
          </ChatWindow>

          <!-- Group Sidebar -->
          <div class="sidebar-wrapper" v-if="activeSessionType === 'group' && groupStore.currentGroup">
            <GroupSidebar
              v-show="showGroupSidebar"
              :key="groupStore.currentGroup.id"
              :group="groupStore.currentGroup"
              @close-sidebar="showGroupSidebar = false"
              @close="handleGroupClose"
              @scroll-to-message="handleScrollToMessage"
            />
          </div>
        </div>
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
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
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
import GroupSidebar from '../components/GroupSidebar.vue'

const router = useRouter()
const userStore = useUserStore()
const chatStore = useChatStore()
const groupStore = useGroupStore()
const aiStore = useAIStore()

// State
const activeTab = ref('chats')
const searchText = ref('')
const activeSessionType = ref<'private' | 'group' | 'ai' | null>(null)
const showGroupSidebar = ref(false)
const chatWindowRef = ref<InstanceType<typeof ChatWindow>>()

// WebSocket
const { messages: wsMessages, connect, disconnect } = useWebSocket('ws://localhost:3102/sub')

// Computed for ChatWindow
const currentMessages = computed(() => {
  if (activeSessionType.value === 'ai') {
    return aiStore.currentBot ? aiStore.getBotMessages(aiStore.currentBot.id) : []
  }
  return chatStore.currentSession?.messages || []
})

// Polling for friend requests
let pollInterval: ReturnType<typeof setInterval>

const startPolling = () => {
  pollInterval = setInterval(() => {
    chatStore.loadFriendRequests()
  }, 10000) // 10 seconds
}

const stopPolling = () => {
  if (pollInterval) {
    clearInterval(pollInterval)
  }
}

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

const isLoadingMessages = ref(false)

const hasMoreMessages = computed(() => {
  if (activeSessionType.value === 'ai') return false // AI history not paginated yet or managed by AI store
  return !(chatStore.currentSession as any)?.isHistoryAllLoaded
})

// Handlers
const handleConversationSelect = async (conv: any) => {
  console.log('handleConversationSelect:', conv.conversation_type, 'current activeSessionType:', activeSessionType.value)

  // Close sidebar first
  showGroupSidebar.value = false
  console.log('set showGroupSidebar to false')

  await nextTick()

  if (conv.conversation_type === 3) {
    activeSessionType.value = 'ai'
    const bot = aiStore.getBotById(conv.target_id)
    if (bot) {
      aiStore.setCurrentBot(bot)
    }
  } else {
    activeSessionType.value = conv.conversation_type === 2 ? 'group' : 'private'
  }

  const name = parseSqlNullString(conv.target_user?.nickname)
  const avatar = parseSqlNullString(conv.target_user?.avatar)
  
  // For AI, we might need to get name/avatar from bot info if target_user is empty
  let finalName = name
  let finalAvatar = avatar
  if (conv.conversation_type === 3) {
     const bot = aiStore.getBotById(conv.target_id)
     if (bot) {
       finalName = bot.name
       // finalAvatar = bot.avatar // bots don't have avatar url yet, use style
     }
  }

  await chatStore.openChat(conv.target_id, conv.conversation_type, finalName, finalAvatar)

  if (activeSessionType.value === 'group') {
    // Sync group store if needed
    // groupStore.setCurrentGroup(...) // logic needs to match
    await loadGroupMembers(conv.target_id)
  }

  // Ensure sidebar switches to 'chats' if we are in search or other tabs, 
  // though typically ConversationList is in 'chats' tab already.
  if (activeTab.value !== 'chats') {
    activeTab.value = 'chats'
  }

  console.log('after handleConversationSelect, activeSessionType:', activeSessionType.value, 'showGroupSidebar:', showGroupSidebar.value)
}

const handleFriendChat = (friend: any) => {
  console.log('handleFriendChat')

  // Close sidebar first
  showGroupSidebar.value = false
  console.log('set showGroupSidebar to false')

  const userId = friend.friend_user?.id || friend.friend_id
  
  // Clear AI state if switching from AI
  if (activeSessionType.value === 'ai') {
     aiStore.setCurrentBot(null as any)
  }

  activeSessionType.value = 'private'
  chatStore.openChat(userId, 1, friend.remark || friend.friend_user?.nickname)
  activeTab.value = 'chats' // Switch to chat tab

  console.log('after handleFriendChat, activeSessionType:', activeSessionType.value, 'showGroupSidebar:', showGroupSidebar.value)
}

const handleGroupSelect = async (group: any) => {
  console.log('handleGroupSelect')

  // Close sidebar first
  showGroupSidebar.value = false
  console.log('set showGroupSidebar to false')

  await nextTick()

  activeSessionType.value = 'group'
  await groupStore.setCurrentGroup(group)
  chatStore.openChat(group.id, 2, group.name)
  await loadGroupMembers(group.id)
  activeTab.value = 'chats' // Switch to chat tab to show conversation list

  console.log('after handleGroupSelect, activeSessionType:', activeSessionType.value, 'showGroupSidebar:', showGroupSidebar.value)
}

const handleBotSelect = (bot: any) => {
  console.log('handleBotSelect')

  activeSessionType.value = 'ai'
  showGroupSidebar.value = false
  aiStore.setCurrentBot(bot)
  
  // Add to conversation list
  chatStore.openChat(bot.id, 3, bot.name)
  
  activeTab.value = 'chats' // Switch to chat tab

  console.log('after handleBotSelect, activeSessionType:', activeSessionType.value, 'showGroupSidebar:', showGroupSidebar.value)
}

const toggleGroupSidebar = () => {
  showGroupSidebar.value = !showGroupSidebar.value
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

const handleLoadMore = async () => {
  if (isLoadingMessages.value) return
  
  isLoadingMessages.value = true
  try {
    if (activeSessionType.value === 'ai') {
      if (aiStore.currentBot) {
        await aiStore.loadMoreMessages(aiStore.currentBot.id)
      }
    } else {
      const count = await chatStore.loadMoreMessages()
      if (count === 0) {
        // No more messages
      }
    }
  } finally {
    isLoadingMessages.value = false
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

const handleGroupClose = async () => {
  // Handle group close (e.g., after leaving or deleting)
  showGroupSidebar.value = false
  groupStore.clearCurrentGroup()
  activeSessionType.value = null
  await chatStore.loadConversations()
}

const handleScrollToMessage = (msg: any) => {
  const messageId = msg.msg_id || msg.id
  if (chatWindowRef.value && messageId) {
    chatWindowRef.value.scrollToMessage(messageId)
  }
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
    chatStore.loadFriends(),
    chatStore.loadFriendRequests()
  ])
  
  startPolling()

  try {
    connect(userStore.token, userStore.currentUser!.id)
  } catch (e) {
    console.warn('WS failed', e)
  }

  // Restore UI state
  try {
    const stored = localStorage.getItem('app_state')
    if (stored) {
      const state = JSON.parse(stored)
      if (state.activeTab) activeTab.value = state.activeTab
      
      if (state.activeSessionType && state.currentSessionId) {
        if (state.activeSessionType === 'ai') {
          await aiStore.loadBots()
          const bot = aiStore.getBotById(state.currentSessionId)
          if (bot) {
             aiStore.setCurrentBot(bot)
             activeSessionType.value = 'ai'
          }
        } else if (state.activeSessionType === 'group') {
           await groupStore.loadGroups()
           const group = groupStore.getGroupById(state.currentSessionId)
           if (group) {
             await groupStore.setCurrentGroup(group)
             chatStore.openChat(group.id, 2, group.name)
             activeSessionType.value = 'group'
             await loadGroupMembers(group.id)
           }
        } else if (state.activeSessionType === 'private') {
           chatStore.openChat(state.currentSessionId, 1, state.currentSessionName, state.currentSessionAvatar)
           activeSessionType.value = 'private'
        }
      }
    }
  } catch (e) {
    console.error('Failed to restore state', e)
  }
})

// Persistence state
const saveState = () => {
  try {
    const state = {
      activeTab: activeTab.value,
      activeSessionType: activeSessionType.value,
      currentSessionId: activeSessionType.value === 'ai' ? aiStore.currentBot?.id : 
                        activeSessionType.value === 'group' ? groupStore.currentGroup?.id :
                        chatStore.currentSession?.targetId,
      currentSessionName: activeSessionType.value === 'ai' ? aiStore.currentBot?.name :
                          activeSessionType.value === 'group' ? groupStore.currentGroup?.name :
                          chatStore.currentSession?.name,
      currentSessionAvatar: activeSessionType.value === 'ai' ? undefined :
                            activeSessionType.value === 'group' ? undefined :
                            chatStore.currentSession?.avatar
    }
    localStorage.setItem('app_state', JSON.stringify(state))
  } catch (e) {
    console.error('Failed to save state', e)
  }
}

// Watchers for persistence
watch([activeTab, activeSessionType, () => chatStore.currentSession?.id, () => aiStore.currentBot?.id], () => {
  saveState()
}, { deep: true })

onUnmounted(() => {
  stopPolling()
})

// Watch activeSessionType to close sidebar when switching away from group
watch(activeSessionType, (newType, oldType) => {
  console.log('activeSessionType changed from', oldType, 'to', newType)
  // Close sidebar when switching to non-group chat
  if (newType !== 'group') {
    showGroupSidebar.value = false
    console.log('auto-close sidebar because newType is not group')
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

// Watchers for sidebar control
watch(activeTab, () => {
  showGroupSidebar.value = false
})

watch(() => chatStore.currentSession?.id, () => {
  showGroupSidebar.value = false
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

.sidebar-badge :deep(.el-badge__content) {
  top: 0;
  right: 0;
  transform: translateY(-50%) translateX(50%);
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

.header-actions-wrapper {
  display: flex;
  gap: 8px;
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
  overflow: hidden;
}

.chat-container {
  display: flex;
  width: 100%;
  height: 100%;
  position: relative; /* Ensure positioning context */
}

.sidebar-wrapper {
  height: 100%;
  flex-shrink: 0;
}

.chat-container :deep(.chat-window) {
  flex: 1;
  min-width: 0;
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
