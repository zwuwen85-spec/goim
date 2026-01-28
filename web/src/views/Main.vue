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
          <el-avatar :size="40" :src="userStore.currentUser?.avatar" :key="userStore.currentUser?.avatar" class="user-avatar">
            {{ userStore.currentUser?.nickname?.[0] }}
          </el-avatar>
          <template #dropdown>
            <el-dropdown-menu>
              <div class="user-dropdown-header">
                <div class="dropdown-name">{{ userStore.currentUser?.nickname }}</div>
                <div class="dropdown-status">在线</div>
              </div>
              <el-dropdown-item divided @click="router.push('/settings')">
                <el-icon><Setting /></el-icon>设置
              </el-dropdown-item>
              <el-dropdown-item @click="handleLogout">
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
            @avatar-click="handleAvatarClick"
          >
            <template #actions>
              <div class="header-actions-wrapper">
                <el-button
                  v-show="activeSessionType === 'group'"
                  key="group-settings-btn"
                  text
                  circle
                  title="群设置"
                  @click.stop="handleOpenGroupSettings"
                >
                  <el-icon><Setting /></el-icon>
                </el-button>
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

  <!-- Group Settings Dialog -->
  <el-dialog
    v-model="showGroupSettings"
    title="群设置"
    width="90%"
    :style="{ maxWidth: '600px' }"
    class="group-settings-dialog"
    append-to-body
    destroy-on-close
  >
    <GroupSettings
      v-if="groupStore.currentGroup"
      :group="groupStore.currentGroup"
      :my-role="myGroupRole"
      :my-nickname="myGroupNickname"
      @close="showGroupSettings = false"
      @updated="handleGroupSettingsUpdated"
    />
  </el-dialog>

  <!-- User Profile Dialog -->
  <el-dialog
    v-model="showUserProfile"
    title="用户资料"
    width="90%"
    :style="{ maxWidth: '500px' }"
    append-to-body
    destroy-on-close
  >
    <UserProfile
      v-if="selectedUserId"
      :user-id="selectedUserId"
      @close="showUserProfile = false"
      @start-chat="handleStartChatFromProfile"
    />
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import {
  ChatLineRound, User, Connection, Cpu,
  SwitchButton, ChatDotRound, MoreFilled, Delete, Setting
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
import GroupSettings from '../components/GroupSettings.vue'
import UserProfile from '../components/UserProfile.vue'

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
const showGroupSettings = ref(false)
const showUserProfile = ref(false)
const selectedUserId = ref<number | null>(null)
const chatWindowRef = ref<InstanceType<typeof ChatWindow>>()
const processedGroupUpdateCount = ref(0)
const processedUserUpdateCount = ref(0)
const processedGroupMemberUpdateCount = ref(0)
const membersRefreshKey = ref(0) // Used to trigger message re-renders

// WebSocket
const { messages: wsMessages, groupUpdates: wsGroupUpdates, userUpdates: wsUserUpdates, groupMemberUpdates: wsGroupMemberUpdates, connect, disconnect, changeRoom } = useWebSocket('ws://localhost:3102/sub')

// Computed for ChatWindow - returns array but depends on membersRefreshKey
const currentMessages = computed(() => {
  const msgs = activeSessionType.value === 'ai'
    ? (aiStore.currentBot ? aiStore.getBotMessages(aiStore.currentBot.id) : [])
    : (chatStore.currentSession?.messages || [])

  // Access membersRefreshKey to create dependency (triggers re-evaluation when updated)
  const key = membersRefreshKey.value

  // Add a non-enumerable property to track version without affecting array appearance
  if (Array.isArray(msgs) && !Object.prototype.hasOwnProperty.call(msgs, '_version')) {
    Object.defineProperty(msgs, '_version', {
      value: key,
      enumerable: false,
      configurable: true,
      writable: true
    })
  } else if (Array.isArray(msgs)) {
    msgs._version = key
  }

  return msgs
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

// Current user's role in the group
const myGroupRole = computed(() => {
  if (activeSessionType.value !== 'group' || !groupStore.currentGroup) return 1
  const myId = userStore.currentUser?.id || 0
  const member = groupStore.members.find(m => m.user_id === myId)
  return member?.role || 1
})

// Current user's nickname in the group
const myGroupNickname = computed(() => {
  if (activeSessionType.value !== 'group' || !groupStore.currentGroup) return ''
  const myId = userStore.currentUser?.id || 0
  const member = groupStore.members.find(m => m.user_id === myId)
  // Handle sql.NullString format: {String: 'xxx', Valid: true}
  const nickname = member?.nickname
  if (nickname && typeof nickname === 'object' && 'String' in nickname) {
    return nickname.String || ''
  }
  return nickname || ''
})

const sessionAvatar = computed(() => {
  if (activeSessionType.value === 'ai') return '' // Use style
  const avatar = chatStore.currentSession?.avatar
  console.log('[sessionAvatar] computed:', {
    activeSessionType: activeSessionType.value,
    avatar: avatar,
    result: avatar ? (avatar.startsWith('/uploads/') ? window.location.origin + avatar : avatar) : ''
  })
  if (!avatar) return ''
  // Handle relative path for uploaded files
  if (avatar.startsWith('/uploads/')) {
    return window.location.origin + avatar
  }
  return avatar
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
    // Find and set the current group
    await groupStore.loadGroups()
    const group = groupStore.getGroupById(conv.target_id)
    if (group) {
      await groupStore.setCurrentGroup(group)
    }
    await loadGroupMembers(conv.target_id)
    // Join the group room to receive messages
    const roomId = `group://${conv.target_id}`
    console.log('[Main] Opening group chat, joining room:', roomId)
    changeRoom(roomId)
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
  chatStore.openChat(userId, 1, friend.remark || friend.friend_user?.nickname, parseSqlNullString(friend.friend_user?.avatar))
  activeTab.value = 'chats' // Switch to chat tab

  console.log('after handleFriendChat, activeSessionType:', activeSessionType.value, 'showGroupSidebar:', showGroupSidebar.value)
}

const handleStartChatFromProfile = (userId: number) => {
  console.log('handleStartChatFromProfile:', userId)

  // Close sidebar
  showGroupSidebar.value = false

  // Find user in friends list to get name/avatar
  const friend = chatStore.friends.find(f =>
    (f.friend_user?.id === userId) || (f.friend_id === userId)
  )

  // Clear AI state if switching from AI
  if (activeSessionType.value === 'ai') {
     aiStore.setCurrentBot(null as any)
  }

  activeSessionType.value = 'private'
  chatStore.openChat(
    userId,
    1,
    friend?.remark || friend?.friend_user?.nickname || `User${userId}`,
    friend ? parseSqlNullString(friend.friend_user?.avatar) : undefined
  )
  activeTab.value = 'chats'
}

const handleGroupSelect = async (group: any) => {
  console.log('handleGroupSelect')

  // Close sidebar first
  showGroupSidebar.value = false
  console.log('set showGroupSidebar to false')

  await nextTick()

  activeSessionType.value = 'group'
  await groupStore.setCurrentGroup(group)
  chatStore.openChat(group.id, 2, group.name, parseSqlNullString(group.avatar))
  await loadGroupMembers(group.id)

  // Room already joined on mount, no need to join again
  // But if this is a newly joined group, we may need to ensure room is joined
  const roomId = `group://${group.id}`
  console.log('[Main] Ensuring group room:', roomId)
  changeRoom(roomId) // Safe to call again, will check if already in room

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

const handleOpenGroupSettings = () => {
  showGroupSettings.value = true
}

const handleAvatarClick = (userId: number) => {
  selectedUserId.value = userId
  showUserProfile.value = true
}

const handleSendMessage = async (content: string) => {
  if (activeSessionType.value === 'ai') {
    if (aiStore.currentBot) {
      // Use streaming for AI messages
      await aiStore.streamMessage(aiStore.currentBot.id, content)
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
    
    // Clear chat store explicitly
    chatStore.clearAll()
    
    userStore.logout()
    
    // Force reload to clear all memory state absolutely
    window.location.href = '/login'
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

const handleGroupSettingsUpdated = async (data: any) => {
  console.log('[handleGroupSettingsUpdated] START')
  console.log('[handleGroupSettingsUpdated] Received data:', data)

  // Extract updated fields
  const newAvatarUrl = data?.avatar
  const newName = data?.name
  const newMyNickname = data?.myNickname
  const currentGroupId = groupStore.currentGroup?.id

  console.log('[handleGroupSettingsUpdated] newAvatarUrl:', newAvatarUrl)
  console.log('[handleGroupSettingsUpdated] newName:', newName)
  console.log('[handleGroupSettingsUpdated] newMyNickname:', newMyNickname)
  console.log('[handleGroupSettingsUpdated] currentGroupId:', currentGroupId)

  // Reload group info and members
  if (currentGroupId) {
    console.log('[handleGroupSettingsUpdated] Loading groups...')
    await groupStore.loadGroups()
    const updatedGroup = groupStore.getGroupById(currentGroupId)
    console.log('[handleGroupSettingsUpdated] updatedGroup:', updatedGroup)
    if (updatedGroup) {
      // If avatar was updated, update the group object with the new URL
      if (newAvatarUrl) {
        updatedGroup.avatar = newAvatarUrl
      }
      // If name was updated, use the new name from data (not from server to avoid stale data)
      if (newName) {
        updatedGroup.name = newName
      }
      groupStore.setCurrentGroup(updatedGroup)
    }
  }

  // If nickname was updated, reload members to get the latest data
  if (newMyNickname && currentGroupId) {
    const myId = userStore.currentUser?.id || 0

    // First, update the member directly in groupStore.members for immediate UI update
    const member = groupStore.members.find(m => m.user_id === myId)
    if (member) {
      // Update the nickname in the store
      if (typeof member.nickname === 'object' && 'String' in member.nickname) {
        member.nickname.String = newMyNickname
      } else {
        member.nickname = newMyNickname
      }
    }

    // Then clear the cache and reload to sync with server
    groupMembersCache.value.delete(currentGroupId)
    await groupStore.loadMembers(currentGroupId)

    // Also update the members cache directly
    const membersMap = new Map<number, any>()
    groupStore.members.forEach(m => {
      membersMap.set(m.user_id, m)
    })
    groupMembersCache.value.set(currentGroupId, membersMap)

    // Force ChatWindow to re-render to show the updated name
    membersRefreshKey.value++
  }

  // Reload conversations to update the list
  console.log('[handleGroupSettingsUpdated] Loading conversations...')
  await chatStore.loadConversations()

  // After reload, update current session and conversation with new data
  if (activeSessionType.value === 'group' && chatStore.currentSession && currentGroupId) {
    console.log('[handleGroupSettingsUpdated] Updating currentSession and conversations...')

    // Update current session using $patch for reactivity
    chatStore.$patch((state) => {
      if (state.currentSession) {
        if (newAvatarUrl) {
          state.currentSession.avatar = newAvatarUrl
        }
        if (newName) {
          state.currentSession.name = newName
        }
      }
    })

    // Update the conversation in the list using $patch for reactivity
    const convIndex = chatStore.conversations.findIndex(
      c => c.target_id === currentGroupId && c.conversation_type === 2
    )
    console.log('[handleGroupSettingsUpdated] convIndex:', convIndex)
    if (convIndex !== -1) {
      chatStore.$patch((state) => {
        if (newAvatarUrl) {
          state.conversations[convIndex].avatar = newAvatarUrl
        }
        if (newName) {
          state.conversations[convIndex].name = newName
        }
      })
    }
  }

  console.log('[handleGroupSettingsUpdated] END')
}

// Helpers for ChatWindow
const groupMembersCache = ref<Map<number, Map<number, any>>>(new Map())

const loadGroupMembers = async (groupId: number) => {
  // Logic from original Main.vue to load members
  // This is needed for getSenderName/Avatar in groups
  if (!groupMembersCache.value.has(groupId)) {
    // Call API to load members
    await groupStore.loadMembers(groupId)
    // Store in cache
    const membersMap = new Map<number, any>()
    groupStore.members.forEach(m => {
      membersMap.set(m.user_id, m)
    })
    groupMembersCache.value.set(groupId, membersMap)
  }
}

// Helper functions for ChatWindow - these will be reactive through watchEffect
const getSenderName = (msg: any) => {
  if (msg.role) return msg.role === 'user' ? '我' : (aiStore.currentBot?.name || 'AI')

  if (msg.from_user_id === userStore.currentUser?.id) return '我'

  if (activeSessionType.value === 'group') {
    const member = groupStore.members.find(m => m.user_id === msg.from_user_id)
    if (member) {
      // Handle sql.NullString format for nickname
      const groupNickname = member.nickname && typeof member.nickname === 'object' && 'String' in member.nickname
        ? member.nickname.String
        : member.nickname
      const userNickname = member.user?.nickname

      return groupNickname || userNickname || `User ${msg.from_user_id}`
    }
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

  // Ensure we load sessions for the current user if not already loaded correctly
  // This is a safety check in case page was refreshed or store was reset
  // But be careful not to load if we already have data for *this* user
  // The store doesn't track *which* user the data belongs to easily,
  // but if it's empty, we definitely need to load.
  if (chatStore.conversations.length === 0) {
      chatStore.loadSessionsFromStorage()
  }

  await Promise.all([
    chatStore.loadConversations(),
    chatStore.loadFriends(),
    chatStore.loadFriendRequests(),
    groupStore.loadGroups() // Load all groups
  ])

  startPolling()

  try {
    connect(userStore.token, userStore.currentUser!.id)

    // After connecting, join all group rooms to receive messages
    // Wait a bit for connection to be established
    setTimeout(() => {
      const groups = groupStore.groups
      console.log('[Main] Joining all group rooms, groups count:', groups.length)
      groups.forEach(group => {
        const roomId = `group://${group.id}`
        console.log('[Main] Auto-joining group room:', roomId)
        changeRoom(roomId)
      })
    }, 1000)
  } catch (e) {
    console.warn('WS failed', e)
  }

  // Restore UI state
  try {
    const userId = userStore.currentUser?.id
    if (userId) {
      const stored = localStorage.getItem(`app_state_${userId}`)
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
               // Don't need to join room here anymore, already joined on mount
               // const roomId = `group://${group.id}`
               // console.log('[Main] Restoring group, joining room:', roomId)
               // changeRoom(roomId)
             }
          } else if (state.activeSessionType === 'private') {
             chatStore.openChat(state.currentSessionId, 1, state.currentSessionName, state.currentSessionAvatar)
             activeSessionType.value = 'private'
          }
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
    const userId = userStore.currentUser?.id
    if (!userId) return

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
    localStorage.setItem(`app_state_${userId}`, JSON.stringify(state))
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
const processedMsgCount = ref(0)

// Debug: log wsMessages reference
console.log('[Main Debug] wsMessages ref:', wsMessages)
console.log('[Main Debug] wsMessages.value:', wsMessages.value)

watch(wsMessages, (newMessages) => {
  console.log('[Main] wsMessages watch triggered, length:', newMessages.length, 'processed:', processedMsgCount.value)
  console.log('[Main] newMessages value:', newMessages)

  if (newMessages.length > processedMsgCount.value) {
    const start = processedMsgCount.value
    const end = newMessages.length

    console.log('[Main] Processing messages from', start, 'to', end)

    for (let i = start; i < end; i++) {
      const latestMsg = newMessages[i]
      console.log('[Main] Processing message:', latestMsg)
      if (latestMsg.conversation_type === 1 || latestMsg.conversation_type === 2) {
        // Convert WS message to Store Message
        const msg = {
          id: Date.now() + i, // Generate a temporary ID (add i to ensure uniqueness in same tick)
          msg_id: latestMsg.msg_id,
          from_user_id: latestMsg.from_user_id,
          conversation_id: Number(latestMsg.conversation_id),
          conversation_type: Number(latestMsg.conversation_type),
          msg_type: Number(latestMsg.msg_type),
          content: latestMsg.content,
          seq: latestMsg.seq,
          created_at: new Date(latestMsg.created_at * 1000).toISOString()  // Unix timestamp in seconds, convert to milliseconds
        }
        // Determine if this message is from self
        const isFromSelf = latestMsg.from_user_id === userStore.currentUser?.id
        console.log('[Main] Calling addMessage with:', msg, 'isFromSelf:', isFromSelf)
        chatStore.addMessage(msg, isFromSelf)
      }
    }
    processedMsgCount.value = end
  }
}, { deep: true })

// Watch for group update notifications via WebSocket
watch(wsGroupUpdates, async (newUpdates) => {
  console.log('[Main] wsGroupUpdates watch triggered, length:', newUpdates.length, 'processed:', processedGroupUpdateCount.value)

  if (newUpdates.length > processedGroupUpdateCount.value) {
    const start = processedGroupUpdateCount.value
    const end = newUpdates.length

    console.log('[Main] Processing group updates from', start, 'to', end)

    for (let i = start; i < end; i++) {
      const update = newUpdates[i]
      console.log('[Main] Processing group update:', update)

      if (update.type === 'group_update') {
        const groupId = update.group_id
        console.log('[Main] Group update for group:', groupId)

        // Reload groups to get the latest data
        await groupStore.loadGroups()

        // If this is the current group, update currentGroup
        if (groupStore.currentGroup?.id === groupId) {
          const updatedGroup = groupStore.getGroupById(groupId)
          if (updatedGroup) {
            // Apply the update data (avatar/name from notification)
            if (update.avatar) {
              updatedGroup.avatar = update.avatar
            }
            if (update.name) {
              updatedGroup.name = update.name
            }
            groupStore.setCurrentGroup(updatedGroup)
          }
        }

        // Reload conversations to update the list
        await chatStore.loadConversations()

        // Update current session if it's this group
        if (activeSessionType.value === 'group' && chatStore.currentSession?.targetId === groupId) {
          chatStore.$patch((state) => {
            if (state.currentSession) {
              if (update.avatar) {
                state.currentSession.avatar = update.avatar
              }
              if (update.name) {
                state.currentSession.name = update.name
              }
            }
          })

          // Update conversation in list
          const convIndex = chatStore.conversations.findIndex(
            c => c.target_id === groupId && c.conversation_type === 2
          )
          if (convIndex !== -1) {
            chatStore.$patch((state) => {
              if (update.avatar) {
                state.conversations[convIndex].avatar = update.avatar
              }
              if (update.name) {
                state.conversations[convIndex].name = update.name
              }
            })
          }
        }

        console.log('[Main] Group update processed successfully')
      }
    }

    processedGroupUpdateCount.value = end
  }
}, { deep: true })

// Watch for user update notifications via WebSocket
watch(wsUserUpdates, async (newUpdates) => {
  console.log('[Main] wsUserUpdates watch triggered, length:', newUpdates.length, 'processed:', processedUserUpdateCount.value)

  if (newUpdates.length > processedUserUpdateCount.value) {
    const start = processedUserUpdateCount.value
    const end = newUpdates.length

    console.log('[Main] Processing user updates from', start, 'to', end)

    for (let i = start; i < end; i++) {
      const update = newUpdates[i]
      console.log('[Main] Processing user update:', update)

      if (update.type === 'user_update') {
        const userId = update.user_id
        console.log('[Main] User update for user:', userId)

        // Update member info in group store
        groupStore.updateMemberInfo(userId, {
          nickname: update.nickname,
          avatar: update.avatar,
          signature: update.signature
        })

        // Also update in conversations if this user is in a private chat
        const convIndex = chatStore.conversations.findIndex(
          c => c.target_id === userId && c.conversation_type === 1
        )
        if (convIndex !== -1) {
          chatStore.$patch((state) => {
            if (update.avatar) {
              state.conversations[convIndex].avatar = update.avatar
            }
            if (update.nickname) {
              state.conversations[convIndex].name = update.nickname
            }
            // Also update target_user info if available
            if (state.conversations[convIndex].target_user) {
              if (update.avatar) {
                state.conversations[convIndex].target_user!.avatar = update.avatar
              }
              if (update.nickname) {
                state.conversations[convIndex].target_user!.nickname = update.nickname
              }
            }
          })
        }

        // Update current session if it's a chat with this user
        if (activeSessionType.value === 'single' && chatStore.currentSession?.targetId === userId) {
          chatStore.$patch((state) => {
            if (state.currentSession) {
              if (update.avatar) {
                state.currentSession.avatar = update.avatar
              }
              if (update.nickname) {
                state.currentSession.name = update.nickname
              }
            }
          })
        }

        // Notify ChatWindow to update member display
        membersRefreshKey.value++
      }

      console.log('[Main] User update processed successfully')
    }

    processedUserUpdateCount.value = end
  }
}, { deep: true })

// Watch for group member update notifications via WebSocket
watch(wsGroupMemberUpdates, async (newUpdates) => {
  console.log('[Main] wsGroupMemberUpdates watch triggered, length:', newUpdates.length, 'processed:', processedGroupMemberUpdateCount.value)

  if (newUpdates.length > processedGroupMemberUpdateCount.value) {
    const start = processedGroupMemberUpdateCount.value
    const end = newUpdates.length

    console.log('[Main] Processing group member updates from', start, 'to', end)

    for (let i = start; i < end; i++) {
      const update = newUpdates[i]
      console.log('[Main] Processing group member update:', update)

      if (update.type === 'group_member_update') {
        const userId = update.user_id
        const groupId = update.group_id
        console.log('[Main] Group member update for user:', userId, 'in group:', groupId)

        // If this is the current group, update the member info
        if (activeSessionType.value === 'group' && groupStore.currentGroup?.id === groupId) {
          const member = groupStore.members.find(m => m.user_id === userId)
          if (member && update.nickname !== undefined) {
            member.nickname = update.nickname
            console.log('[Main] Updated member nickname:', member.nickname)
          }
        }

        // Notify ChatWindow to update member display
        membersRefreshKey.value++
      }

      console.log('[Main] Group member update processed successfully')
    }

    processedGroupMemberUpdateCount.value = end
  }
}, { deep: true })

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

<style>
/* Group Settings Dialog - Make it fit within viewport */
.group-settings-dialog {
  height: auto !important;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
}

.group-settings-dialog .el-dialog__body {
  max-height: calc(85vh - 120px);
  overflow-y: auto;
  overflow-x: hidden;
}
</style>
