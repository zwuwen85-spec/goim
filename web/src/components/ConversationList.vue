<template>
  <div class="conversation-list">
    <div
      v-for="conv in mergedConversations"
      :key="`${conv.target_id}:${conv.conversation_type}`"
      class="list-item"
      :class="{ active: currentConvId === `${conv.target_id}:${conv.conversation_type}` }"
      @click="handleSelect(conv)"
    >
      <el-avatar 
        :size="48" 
        :src="conv.is_ai ? '' : getAvatar(conv)" 
        :style="conv.is_ai ? { backgroundColor: getBotColor(conv.target_id) } : {}"
        shape="square" 
        class="item-avatar"
      >
        <el-icon v-if="conv.is_ai" size="24" color="white"><ChatDotRound /></el-icon>
        <span v-else>{{ getName(conv)?.[0] || '?' }}</span>
      </el-avatar>
      
      <div class="item-content">
        <div class="item-top">
          <span class="item-name">{{ getName(conv) }}</span>
          <span class="item-time">{{ formatTime(conv.last_msg_time) }}</span>
        </div>
        <div class="item-bottom">
          <span class="item-preview">{{ getLastMessage(conv) }}</span>
          <el-badge 
            v-if="!conv.is_ai && conv.unread_count > 0" 
            :value="conv.unread_count > 99 ? '99+' : conv.unread_count" 
            class="unread-badge" 
          />
        </div>
      </div>
    </div>
    
    <el-empty v-if="mergedConversations.length === 0" description="暂无消息" :image-size="60" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useChatStore } from '../store/chat'
import { useGroupStore } from '../store/group'
import { useAIStore } from '../store/ai'
import { parseSqlNullString } from '../utils/format'
import { ChatDotRound } from '@element-plus/icons-vue'

const emit = defineEmits(['select'])
const chatStore = useChatStore()
const groupStore = useGroupStore()
const aiStore = useAIStore()

const mergedConversations = computed(() => {
  const normal = chatStore.conversations.map(c => ({
    ...c,
    is_ai: false,
    timestamp: new Date(c.last_msg_time).getTime()
  }))

  const ai = []
  const addedBotIds = new Set<number>()
  
  // 1. Add bots with messages or active state
  for (const botIdStr of Object.keys(aiStore.messages)) {
    const botId = Number(botIdStr)
    const msgs = aiStore.messages[botId]
    // Allow bots with empty messages to show up (if they were initialized/clicked)
    // if (!msgs || msgs.length === 0) continue 
    
    const bot = aiStore.getBotById(botId) || aiStore.defaultBots.find(b => b.id === botId)
    if (!bot) continue
    
    const lastMsg = msgs && msgs.length > 0 ? msgs[msgs.length - 1] : null
    const timestamp = lastMsg ? (lastMsg.timestamp || Date.now()) : Date.now() // Use current time for empty chats so they stay at top when clicked
    
    ai.push({
      target_id: botId,
      conversation_type: 'ai',
      is_ai: true,
      name: bot.name,
      avatar: '',
      last_msg_content: lastMsg ? JSON.stringify({ text: lastMsg.content }) : '',
      last_msg_time: new Date(timestamp).toISOString(),
      timestamp: timestamp,
      unread_count: 0
    })
    addedBotIds.add(botId)
  }

  // 2. Add current bot if not present (so it shows up immediately when clicked)
  if (aiStore.currentBot && !addedBotIds.has(aiStore.currentBot.id)) {
      ai.push({
          target_id: aiStore.currentBot.id,
          conversation_type: 'ai',
          is_ai: true,
          name: aiStore.currentBot.name,
          avatar: '',
          last_msg_content: '',
          last_msg_time: new Date().toISOString(),
          timestamp: Date.now(),
          unread_count: 0
      })
  }
  
  return [...normal, ...ai].sort((a, b) => (b.timestamp || 0) - (a.timestamp || 0))
})

const getBotColor = (id: number) => {
  const colors = ['#409EFF', '#67C23A', '#E6A23C', '#F56C6C', '#909399']
  return colors[id % colors.length]
}

const currentConvId = computed(() => {
  if (aiStore.currentBot) {
     return `${aiStore.currentBot.id}:ai`
  }
  if (!chatStore.currentSession) return ''
  return `${chatStore.currentSession.targetId}:${chatStore.currentSession.targetType === 'group' ? 2 : 1}`
})

const getName = (conv: any) => {
  if (conv.conversation_type === 'ai') return conv.name
  // 1. Try to get from target_user (if populated by API)
  const directName = parseSqlNullString(conv.target_user?.nickname)
  if (directName) return directName

  // 2. If private chat (type 1), look up in friends list
  if (conv.conversation_type === 1) {
    const friend = chatStore.friends.find(f => 
      (f.friend_user?.id === conv.target_id) || (f.friend_id === conv.target_id)
    )
    if (friend) {
      return friend.remark || parseSqlNullString(friend.friend_user?.nickname) || `User ${conv.target_id}`
    }
  }

  // 3. If group chat (type 2), look up in group store
  if (conv.conversation_type === 2) {
    const group = groupStore.getGroupById(conv.target_id)
    if (group) return group.name
    return `Group ${conv.target_id}`
  }

  return `User ${conv.target_id}`
}

const getAvatar = (conv: any) => {
  // 1. Try to get from target_user
  const directAvatar = parseSqlNullString(conv.target_user?.avatar)
  if (directAvatar) return directAvatar

  // 2. If private chat, look up in friends list
  if (conv.conversation_type === 1) {
    const friend = chatStore.friends.find(f => 
      (f.friend_user?.id === conv.target_id) || (f.friend_id === conv.target_id)
    )
    if (friend) {
      return parseSqlNullString(friend.friend_user?.avatar)
    }
  }

  // 3. If group chat, look up in group store
  if (conv.conversation_type === 2) {
    const group = groupStore.getGroupById(conv.target_id)
    if (group) return parseSqlNullString(group.avatar)
  }

  return ''
}

const getLastMessage = (conv: any) => {
  const content = parseSqlNullString(conv.last_msg_content)
  if (!content) return '暂无消息'
  try {
    const parsed = JSON.parse(content)
    return parsed.text || content
  } catch {
    return content
  }
}

const formatTime = (timestamp: string | number) => {
  if (!timestamp) return ''
  // Handle ISO string or number
  const date = new Date(timestamp)
  if (isNaN(date.getTime())) {
    // Try parsing as number if it's a string representing a number
    if (typeof timestamp === 'string' && !isNaN(Number(timestamp))) {
       const num = Number(timestamp)
       const d = new Date(num < 10000000000 ? num * 1000 : num)
       if (!isNaN(d.getTime())) return format(d)
    }
    return ''
  }
  return format(date)
}

const format = (date: Date) => {
  const now = new Date()
  const isToday = date.getDate() === now.getDate() && date.getMonth() === now.getMonth() && date.getFullYear() === now.getFullYear()
  
  if (isToday) {
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }
  return date.toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric' })
}

const handleSelect = (conv: any) => {
  emit('select', conv)
}
</script>

<style scoped>
.conversation-list {
  height: 100%;
  overflow-y: auto;
  padding: 8px;
}

.list-item {
  display: flex;
  padding: 12px;
  gap: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: 4px;
}

.list-item:hover {
  background-color: var(--bg-body);
}

.list-item.active {
  background-color: var(--primary-light);
}

.item-avatar {
  flex-shrink: 0;
  border-radius: 12px;
}

.item-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 4px;
}

.item-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.item-name {
  font-weight: 500;
  color: var(--text-primary);
  font-size: 14px;
}

.item-time {
  font-size: 11px;
  color: var(--text-light);
}

.item-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.item-preview {
  font-size: 12px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 140px;
}

.unread-badge :deep(.el-badge__content) {
  border: none;
  height: 16px;
  line-height: 16px;
  padding: 0 4px;
}
</style>
