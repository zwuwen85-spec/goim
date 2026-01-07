<template>
  <div class="chat-window">
    <!-- Header -->
    <div class="chat-header">
      <div class="header-left">
        <slot name="header-prefix"></slot>
        <el-avatar :size="40" :src="avatar" :style="avatarStyle">
          {{ title?.[0] || '?' }}
          <el-icon v-if="!title && !avatar"><UserFilled /></el-icon>
        </el-avatar>
      </div>
      <div class="header-center">
        <div class="header-title">{{ title }}</div>
        <div class="header-subtitle" v-if="subtitle">{{ subtitle }}</div>
      </div>
      <div class="header-actions">
        <slot name="actions"></slot>
      </div>
    </div>

    <!-- Messages -->
    <div class="messages-container" ref="messagesRef">
      <div v-if="loading" class="loading-state">
        <el-skeleton :rows="3" animated />
      </div>
      <template v-else>
        <div
          v-for="(msg, index) in messages"
          :key="msg.id || index"
          :id="'msg-' + (msg.msg_id || msg.id)"
          class="message-wrapper"
          :class="{ 'is-me': isMe(msg) }"
        >
          <div class="message-time-divider" v-if="showTimeDivider(msg, messages[index - 1])">
            <span>{{ formatTime(msg.timestamp || msg.created_at) }}</span>
          </div>
          
          <div class="message-row">
            <el-avatar 
              class="message-avatar" 
              :size="36" 
              :src="getSenderAvatar(msg)" 
              :style="getSenderStyle(msg)"
            >
              {{ getSenderName(msg)?.[0] || '?' }}
            </el-avatar>
            
            <div class="message-content-group">
              <div class="message-sender-name" v-if="!isMe(msg) && showSenderName">
                {{ getSenderName(msg) }}
              </div>
              
              <div class="message-bubble" :class="{ 'typing': msg.isTyping }">
                <template v-if="msg.isTyping">
                  <div class="typing-indicator">
                    <span></span><span></span><span></span>
                  </div>
                </template>
                <template v-else>
                  <div class="message-text">{{ parseContent(msg.content) }}</div>
                </template>
              </div>
            </div>
          </div>
        </div>
      </template>

      <div v-if="messages.length === 0 && !loading" class="empty-state">
        <el-empty description="暂无消息" />
      </div>
    </div>

    <!-- Input -->
    <div class="input-area">
      <div class="input-toolbar">
        <!-- Future: Emoji, Image, File buttons -->
        <el-button text circle size="small">
          <el-icon><Picture /></el-icon>
        </el-button>
        <el-button text circle size="small">
          <el-icon><files /></el-icon>
        </el-button>
      </div>
      <el-input
        v-model="inputValue"
        type="textarea"
        :rows="3"
        placeholder="输入消息..."
        resize="none"
        @keydown.enter.exact.prevent="handleSend"
        class="chat-input"
      />
      <div class="input-footer">
        <span class="input-hint">Enter 发送，Shift + Enter 换行</span>
        <el-button type="primary" @click="handleSend" :disabled="!inputValue.trim() || sending" :loading="sending">
          发送
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, watch, onMounted } from 'vue'
import { UserFilled, Picture, Files } from '@element-plus/icons-vue'
import { parseSqlNullString } from '../utils/format'

const props = defineProps<{
  messages: any[]
  currentUserId: number
  title: string
  subtitle?: string
  avatar?: string
  avatarStyle?: any
  sending?: boolean
  loading?: boolean
  showSenderName?: boolean
  getSenderName?: (msg: any) => string
  getSenderAvatar?: (msg: any) => string
  getSenderStyle?: (msg: any) => any
}>()

const emit = defineEmits(['send'])

const inputValue = ref('')
const messagesRef = ref<HTMLElement>()

const isMe = (msg: any) => {
  return msg.from_user_id === props.currentUserId || msg.role === 'user'
}

const parseContent = (content: any) => {
  const rawContent = parseSqlNullString(content)
  if (!rawContent) return ''
  try {
    const parsed = JSON.parse(rawContent)
    return parsed.text || rawContent
  } catch {
    return rawContent
  }
}

const showTimeDivider = (current: any, prev: any) => {
  if (!prev) return true
  const cTime = new Date(current.timestamp || current.created_at).getTime()
  const pTime = new Date(prev.timestamp || prev.created_at).getTime()
  return cTime - pTime > 5 * 60 * 1000 // 5 minutes
}

const formatTime = (timestamp: string | number) => {
  if (!timestamp) return ''
  const date = new Date(typeof timestamp === 'string' && timestamp.length < 13 ? parseInt(timestamp) * 1000 : timestamp)
  return date.toLocaleString('zh-CN', { hour: '2-digit', minute: '2-digit', month: 'short', day: 'numeric' })
}

const handleSend = () => {
  if (!inputValue.value.trim() || props.sending) return
  emit('send', inputValue.value)
  inputValue.value = ''
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
}

const scrollToMessage = (messageId: number | string) => {
  nextTick(() => {
    const elementId = `msg-${messageId}`
    const element = document.getElementById(elementId)
    if (element) {
      element.scrollIntoView({ behavior: 'smooth', block: 'center' })
      element.classList.add('highlight-message')
      setTimeout(() => {
        element.classList.remove('highlight-message')
      }, 2000)
    }
  })
}

defineExpose({
  scrollToBottom,
  scrollToMessage
})

watch(() => props.messages, () => {
  scrollToBottom()
}, { deep: true })

onMounted(() => {
  scrollToBottom()
})

// Default helpers if not provided
const defaultGetSenderName = (msg: any) => msg.senderName || 'User'
const defaultGetSenderAvatar = (msg: any) => msg.senderAvatar || ''
const defaultGetSenderStyle = (msg: any) => ({})

const getSenderName = props.getSenderName || defaultGetSenderName
const getSenderAvatar = props.getSenderAvatar || defaultGetSenderAvatar
const getSenderStyle = props.getSenderStyle || defaultGetSenderStyle

</script>

<style scoped>
.chat-window {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-surface);
}

.chat-header {
  height: 64px;
  padding: 0 24px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(10px);
  position: relative;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-center {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.header-title {
  font-weight: 600;
  font-size: 16px;
  color: var(--text-primary);
}

.header-subtitle {
  font-size: 12px;
  color: var(--text-secondary);
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  background-color: var(--bg-chat);
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message-time-divider {
  text-align: center;
  margin: 16px 0;
}

.message-time-divider span {
  background: rgba(0, 0, 0, 0.05);
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  color: var(--text-secondary);
}

.message-wrapper {
  display: flex;
  flex-direction: column;
}

.message-avatar {
  flex-shrink: 0;
}

.message-row {
  display: flex;
  gap: 12px;
  max-width: 70%;
}

.is-me .message-row {
  flex-direction: row-reverse;
  align-self: flex-end;
}

.message-content-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.message-sender-name {
  font-size: 12px;
  color: var(--text-secondary);
  margin-left: 4px;
}

.message-bubble {
  padding: 12px 16px;
  border-radius: 12px;
  border-top-left-radius: 2px;
  background: var(--msg-other-bg);
  color: var(--msg-other-text);
  box-shadow: var(--shadow-sm);
  position: relative;
  word-break: break-word;
  white-space: pre-wrap;
  line-height: 1.5;
}

.is-me .message-bubble {
  background: var(--msg-me-bg);
  color: var(--msg-me-text);
  border-top-left-radius: 12px;
  border-top-right-radius: 2px;
}

.input-area {
  padding: 16px 24px;
  background: var(--bg-surface);
  border-top: 1px solid var(--border-color);
}

.input-toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}

.chat-input :deep(.el-textarea__inner) {
  box-shadow: none;
  background: var(--bg-body);
  border: 1px solid transparent;
  padding: 12px;
  border-radius: 8px;
  transition: all 0.2s;
}

.chat-input :deep(.el-textarea__inner):focus {
  background: var(--bg-surface);
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px var(--primary-light);
}

.input-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
}

.input-hint {
  font-size: 12px;
  color: var(--text-light);
}

.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 4px 0;
}

.typing-indicator span {
  width: 6px;
  height: 6px;
  background: currentColor;
  border-radius: 50%;
  animation: bounce 1.4s infinite ease-in-out both;
  opacity: 0.6;
}

.typing-indicator span:nth-child(1) { animation-delay: -0.32s; }
.typing-indicator span:nth-child(2) { animation-delay: -0.16s; }

@keyframes bounce {
  0%, 80%, 100% { transform: scale(0); }
  40% { transform: scale(1); }
}

.empty-state {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.highlight-message .message-bubble {
  animation: highlight 2s ease-out;
}

@keyframes highlight {
  0%, 20% {
    background-color: var(--primary-light);
    transform: scale(1.02);
  }
  100% {
    transform: scale(1);
  }
}
</style>
