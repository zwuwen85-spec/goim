<template>
  <div class="ai-chat-container">
    <!-- Bot list -->
    <div class="bot-selector" v-if="!aiStore.currentBot">
      <h3>选择 AI 助手</h3>
      <div class="bot-grid">
        <div
          v-for="bot in aiStore.bots"
          :key="bot.id"
          class="bot-card"
          @click="selectBot(bot)"
        >
          <el-avatar :size="60" :style="{ backgroundColor: getBotColor(bot.id) }">
            <el-icon size="30"><ChatDotRound /></el-icon>
          </el-avatar>
          <div class="bot-info">
            <div class="bot-name">{{ bot.name }}</div>
            <div class="bot-desc">{{ getBotDesc(bot.personality) }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Chat area -->
    <div v-else class="chat-area">
      <!-- Header -->
      <div class="chat-header">
        <el-button text @click="backToList">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <div class="bot-title">
          <el-avatar :size="40" :style="{ backgroundColor: getBotColor(aiStore.currentBot.id) }">
            <el-icon size="20"><ChatDotRound /></el-icon>
          </el-avatar>
          <div>
            <div class="bot-name">{{ aiStore.currentBot.name }}</div>
            <div class="bot-status">{{ getBotDesc(aiStore.currentBot.personality) }}</div>
          </div>
        </div>
        <el-button text @click="clearChat">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>

      <!-- Messages -->
      <div class="messages-container" ref="messagesRef">
        <div
          v-for="(msg, index) in botMessages"
          :key="index"
          class="message"
          :class="{ 'is-user': msg.role === 'user', 'streaming': msg.streaming, 'error': msg.error }"
        >
          <el-avatar v-if="msg.role === 'assistant'" :size="32" :style="{ backgroundColor: getBotColor(aiStore.currentBot!.id) }">
            <el-icon><ChatDotRound /></el-icon>
          </el-avatar>
          <div class="message-content">
            <div class="message-sender">
              {{ msg.role === 'user' ? '我' : aiStore.currentBot?.name }}
            </div>
            <div class="message-body">
              <template v-if="msg.streaming && !msg.content">
                <div class="typing">
                  <span></span>
                  <span></span>
                  <span></span>
                </div>
              </template>
              <template v-else-if="msg.role === 'assistant'">
                <MarkdownRenderer :content="msg.content" />
                <span v-if="msg.streaming" class="cursor"></span>
              </template>
              <template v-else>
                {{ msg.content }}
              </template>
            </div>
            <div v-if="msg.timestamp" class="message-time">
              {{ formatTime(msg.timestamp) }}
            </div>
          </div>
          <el-avatar v-if="msg.role === 'user'" :size="32">
            {{ userStore.currentUser?.nickname?.[0] }}
          </el-avatar>
        </div>

        <!-- Loading indicator -->
        <div v-if="aiStore.sending" class="message is-user">
          <el-avatar :size="32">
            {{ userStore.currentUser?.nickname?.[0] }}
          </el-avatar>
          <div class="message-content">
            <div class="message-body typing">
              <span></span>
              <span></span>
              <span></span>
            </div>
          </div>
        </div>
      </div>

      <!-- Input area -->
      <div class="input-area">
        <el-input
          v-model="messageInput"
          type="textarea"
          :rows="3"
          placeholder="和 AI 对话..."
          @keydown.enter.exact="handleSend"
          :disabled="aiStore.sending"
        />
        <div class="input-actions">
          <span class="hint">按 Enter 发送，Shift + Enter 换行</span>
          <el-button
            type="primary"
            @click="handleSend"
            :disabled="!messageInput.trim() || aiStore.sending"
            :loading="aiStore.sending"
          >
            发送
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAIStore } from '../store/ai'
import { useUserStore } from '../store/user'
import type { AIBot } from '../api/chat'
import MarkdownRenderer from './MarkdownRenderer.vue'

const aiStore = useAIStore()
const userStore = useUserStore()

const messageInput = ref('')
const messagesRef = ref<HTMLElement>()

const botMessages = computed(() => {
  return aiStore.currentBot ? aiStore.getBotMessages(aiStore.currentBot.id) : []
})

const botColors: Record<number, string> = {
  9001: '#409EFF', // assistant - blue
  9002: '#67C23A', // companion - green
  9003: '#E6A23C', // tutor - orange
  9004: '#F56C6C'  // creative - red
}

const getBotColor = (botId: number) => {
  return botColors[botId] || '#909399'
}

const getBotDesc = (personality: string) => {
  const descs: Record<string, string> = {
    assistant: '智能助手 - 有帮助、知识渊博',
    companion: '聊天伙伴 - 友好、有趣',
    tutor: '学习导师 - 耐心、专业',
    creative: '创意助手 - 富有想象力'
  }
  return descs[personality] || 'AI 助手'
}

const formatTime = (timestamp: number) => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

const selectBot = (bot: AIBot) => {
  aiStore.setCurrentBot(bot)
  nextTick(() => {
    scrollToBottom()
  })
}

const backToList = () => {
  aiStore.currentBot = null
}

const clearChat = async () => {
  try {
    await ElMessageBox.confirm('确定要清空对话记录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    if (aiStore.currentBot) {
      aiStore.clearMessages(aiStore.currentBot.id)
    }
    ElMessage.success('对话记录已清空')
  } catch {
    // Cancelled
  }
}

const handleSend = async () => {
  if (!messageInput.value.trim() || !aiStore.currentBot || aiStore.sending) return

  const userMessage = messageInput.value
  messageInput.value = ''

  try {
    // Use streaming send for better UX
    await aiStore.streamMessage(aiStore.currentBot.id, userMessage)
    await nextTick()
    scrollToBottom()
  } catch (error: any) {
    ElMessage.error(error.message || '发送失败')
    messageInput.value = userMessage // Restore message on error
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
}

// Watch for new messages
watch(() => botMessages.value.length, () => {
  nextTick(() => {
    scrollToBottom()
  })
})

onMounted(async () => {
  try {
    await aiStore.loadBots()
  } catch (error) {
    ElMessage.error('加载 AI 助手失败')
  }
})
</script>

<style scoped>
.ai-chat-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.bot-selector {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.bot-selector h3 {
  margin: 0 0 20px;
  font-size: 16px;
  color: #303133;
}

.bot-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
}

.bot-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
}

.bot-card:hover {
  border-color: #409EFF;
  box-shadow: 0 2px 12px rgba(64, 158, 255, 0.2);
  transform: translateY(-2px);
}

.bot-info {
  margin-top: 12px;
  text-align: center;
}

.bot-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.bot-desc {
  font-size: 12px;
  color: #909399;
}

.chat-area {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.chat-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
}

.bot-title {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
}

.bot-title .bot-name {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.bot-status {
  font-size: 12px;
  color: #909399;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: #f5f7fa;
}

.message {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}

.message.is-user {
  flex-direction: row-reverse;
}

.message-content {
  max-width: 70%;
}

.message-sender {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.message.is-user .message-sender {
  text-align: right;
}

.message-body {
  padding: 10px 14px;
  background: #fff;
  border-radius: 8px;
  word-break: break-word;
  line-height: 1.5;
  white-space: pre-wrap;
}

.message.is-user .message-body {
  background: #95ec69;
}

.message-time {
  font-size: 11px;
  color: #c0c4cc;
  margin-top: 4px;
}

.message.is-user .message-time {
  text-align: right;
}

.message.streaming .message-body {
  position: relative;
}

.cursor {
  display: inline-block;
  width: 2px;
  height: 1em;
  background: #409EFF;
  margin-left: 2px;
  animation: blink 1s infinite;
  vertical-align: text-bottom;
}

@keyframes blink {
  0%, 50% {
    opacity: 1;
  }
  51%, 100% {
    opacity: 0;
  }
}

.message.error .message-body {
  background: #fef0f0;
  color: #f56c6c;
}

.typing {
  display: flex;
  gap: 4px;
  padding: 12px 16px;
}

.typing span {
  width: 8px;
  height: 8px;
  background: #909399;
  border-radius: 50%;
  animation: typing 1.4s infinite;
}

.typing span:nth-child(2) {
  animation-delay: 0.2s;
}

.typing span:nth-child(3) {
  animation-delay: 0.4s;
}

@keyframes typing {
  0%, 60%, 100% {
    transform: translateY(0);
  }
  30% {
    transform: translateY(-10px);
  }
}

.input-area {
  padding: 16px 20px;
  background: #fff;
  border-top: 1px solid #e4e7ed;
}

.input-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
}

.hint {
  font-size: 12px;
  color: #909399;
}
</style>
