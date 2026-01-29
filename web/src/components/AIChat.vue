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
      <!-- Header with bot switcher -->
      <div class="chat-header">
        <el-button text @click="backToList">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>

        <div class="bot-title">
          <el-dropdown @command="switchBot" trigger="click" v-if="aiStore.bots.length > 1">
            <span class="bot-switcher-trigger">
              <el-avatar :size="40" :style="{ backgroundColor: getBotColor(aiStore.currentBot.id) }">
                <el-icon size="20"><ChatDotRound /></el-icon>
              </el-avatar>
              <div class="bot-info-dropdown">
                <div class="bot-name">{{ aiStore.currentBot.name }}</div>
                <div class="bot-status">{{ getBotDesc(aiStore.currentBot.personality) }}</div>
              </div>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item
                  v-for="bot in aiStore.bots"
                  :key="bot.id"
                  :command="bot"
                  :disabled="bot.id === aiStore.currentBot.id"
                >
                  <el-avatar :size="24" :style="{ backgroundColor: getBotColor(bot.id) }" style="margin-right: 8px">
                    <el-icon><ChatDotRound /></el-icon>
                  </el-avatar>
                  {{ bot.name }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <div v-else class="bot-title-static">
            <el-avatar :size="40" :style="{ backgroundColor: getBotColor(aiStore.currentBot.id) }">
              <el-icon size="20"><ChatDotRound /></el-icon>
            </el-avatar>
            <div>
              <div class="bot-name">{{ aiStore.currentBot.name }}</div>
              <div class="bot-status">{{ getBotDesc(aiStore.currentBot.personality) }}</div>
            </div>
          </div>
        </div>

        <el-button text @click="clearChat">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>

      <!-- Messages -->
      <div class="messages-container" ref="messagesRef" @paste="handlePaste">
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
              <!-- Show attached images -->
              <div v-if="msg.images && msg.images.length > 0" class="message-images">
                <el-image
                  v-for="(img, imgIdx) in msg.images"
                  :key="imgIdx"
                  :src="img"
                  :preview-src-list="msg.images"
                  fit="cover"
                  class="message-image"
                />
              </div>

              <!-- Show attached files -->
              <div v-if="msg.files && msg.files.length > 0" class="message-files">
                <div v-for="(file, fileIdx) in msg.files" :key="fileIdx" class="message-file">
                  <el-icon><Document /></el-icon>
                  <span>{{ file.name }}</span>
                </div>
              </div>

              <!-- Text content -->
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

      <!-- Input area with file/attachment support -->
      <div class="input-area">
        <!-- Attachment previews -->
        <div v-if="attachedImages.length > 0 || attachedFiles.length > 0" class="attachments-preview">
          <div v-for="(img, idx) in attachedImages" :key="'img-' + idx" class="attachment-item">
            <el-image :src="img" fit="cover" class="attachment-image" />
            <el-button
              circle
              size="small"
              class="attachment-remove"
              @click="removeImage(idx)"
            >
              <el-icon><Close /></el-icon>
            </el-button>
          </div>
          <div v-for="(file, idx) in attachedFiles" :key="'file-' + idx" class="attachment-item attachment-file">
            <el-icon><Document /></el-icon>
            <span class="file-name">{{ file.name }}</span>
            <el-button
              circle
              size="small"
              class="attachment-remove"
              @click="removeFile(idx)"
            >
              <el-icon><Close /></el-icon>
            </el-button>
          </div>
        </div>

        <el-input
          v-model="messageInput"
          type="textarea"
          :rows="3"
          placeholder="和 AI 对话... (可粘贴图片或拖拽文件)"
          @keydown.enter.exact="handleSend"
          @paste="handlePaste"
          :disabled="aiStore.sending"
        />
        <div class="input-actions">
          <div class="input-left">
            <el-upload
              ref="uploadRef"
              :auto-upload="false"
              :show-file-list="false"
              :on-change="handleFileSelect"
              accept=".txt,.md,.json,.xml,.yaml,.yml,.log,.jpg,.jpeg,.png,.gif,.webp"
            >
              <template #trigger>
                <el-button text>
                  <el-icon><Folder /></el-icon>
                  文件
                </el-button>
              </template>
            </el-upload>
            <span class="hint">支持粘贴图片、拖拽文件</span>
          </div>
          <el-button
            type="primary"
            @click="handleSend"
            :disabled="!canSend"
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
import {
  ChatDotRound,
  ArrowLeft,
  ArrowDown,
  Delete,
  Document,
  Folder,
  Close
} from '@element-plus/icons-vue'
import { useAIStore } from '../store/ai'
import { useUserStore } from '../store/user'
import { aiApi, type AIBot, type UploadedFile } from '../api/chat'
import MarkdownRenderer from './MarkdownRenderer.vue'

const aiStore = useAIStore()
const userStore = useUserStore()

const messageInput = ref('')
const messagesRef = ref<HTMLElement>()
const uploadRef = ref()
const attachedImages = ref<string[]>([])
const attachedFiles = ref<UploadedFile[]>([])
const uploadingFiles = ref(false)

const botMessages = computed(() => {
  return aiStore.currentBot ? aiStore.getBotMessages(aiStore.currentBot.id) : []
})

const canSend = computed(() => {
  return (messageInput.value.trim() || attachedImages.value.length > 0 || attachedFiles.value.length > 0) && !aiStore.sending
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

const switchBot = (bot: AIBot) => {
  if (bot.id === aiStore.currentBot?.id) return

  // Clear attachments when switching
  attachedImages.value = []
  attachedFiles.value = []

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

// Handle paste event for images
const handlePaste = async (event: ClipboardEvent) => {
  const items = event.clipboardData?.items
  console.log('[AIChat] Paste event, items:', items?.length)
  if (!items) return

  for (const item of Array.from(items)) {
    console.log('[AIChat] Paste item type:', item.type)
    if (item.type.indexOf('image') !== -1) {
      event.preventDefault()
      const file = item.getAsFile()
      console.log('[AIChat] Pasted image file:', file?.name, file?.size, file?.type)
      if (file) {
        await uploadFile(file)
      }
    }
  }
}

// Handle file selection
const handleFileSelect = async (file: any) => {
  console.log('[AIChat] File selected:', file)
  if (file.raw) {
    await uploadFile(file.raw)
  }
}

// Upload file and get URL
const uploadFile = async (file: File) => {
  try {
    console.log('[AIChat] Uploading file:', file.name, file.size, file.type)
    uploadingFiles.value = true
    const response = await aiApi.uploadFile(file) as any
    console.log('[AIChat] Upload response:', response)

    if (response.code === 0) {
      const data = response.data
      console.log('[AIChat] Upload data:', data)
      if (data.is_image) {
        attachedImages.value.push(data.file_url)
        ElMessage.success('图片上传成功')
      } else {
        attachedFiles.value.push({
          name: data.file_name,
          url: data.file_url,
          type: data.file_type,
          is_text: data.is_text,
          is_image: data.is_image,
          content: data.content
        })
        ElMessage.success('文件上传成功')
      }
    } else {
      ElMessage.error(response.message || '上传失败')
    }
  } catch (error: any) {
    console.error('[AIChat] Upload error:', error)
    ElMessage.error(error.message || '文件上传失败')
  } finally {
    uploadingFiles.value = false
  }
}

const removeImage = (index: number) => {
  attachedImages.value.splice(index, 1)
}

const removeFile = (index: number) => {
  attachedFiles.value.splice(index, 1)
}

const handleSend = async () => {
  if (!aiStore.currentBot || aiStore.sending || uploadingFiles.value) return

  const userMessage = messageInput.value
  const hasAttachments = attachedImages.value.length > 0 || attachedFiles.value.length > 0

  if (!userMessage.trim() && !hasAttachments) return

  // Prepare message with attachments
  const imageUrls = [...attachedImages.value]
  const fileUrls = attachedFiles.value.map(f => f.url)

  // Clear input
  messageInput.value = ''

  // Add user message with attachments to store
  const userMsg: any = {
    role: 'user',
    content: userMessage,
    timestamp: Date.now()
  }

  if (imageUrls.length > 0) {
    userMsg.images = imageUrls
  }

  if (attachedFiles.value.length > 0) {
    userMsg.files = [...attachedFiles.value]
  }

  const newMsgs = [...botMessages.value, userMsg]
  aiStore.messages.value = { ...aiStore.messages.value, [aiStore.currentBot.id]: newMsgs }

  try {
    // Use multimodal streaming if there are attachments
    if (hasAttachments) {
      await aiStore.streamMultimodalMessage(
        aiStore.currentBot.id,
        userMessage,
        imageUrls,
        fileUrls
      )
    } else {
      await aiStore.streamMessage(aiStore.currentBot.id, userMessage)
    }

    await nextTick()
    scrollToBottom()
  } catch (error: any) {
    ElMessage.error(error.message || '发送失败')
    messageInput.value = userMessage // Restore message on error
  } finally {
    // Clear attachments
    attachedImages.value = []
    attachedFiles.value = []
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
}

.bot-switcher-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
}

.bot-switcher-trigger:hover {
  background: #f5f7fa;
}

.bot-info-dropdown {
  text-align: left;
}

.bot-title-static {
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

.message-images {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 8px;
}

.message-image {
  width: 120px;
  height: 120px;
  border-radius: 8px;
  cursor: pointer;
}

.message-files {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 8px;
}

.message-file {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 6px;
  font-size: 13px;
  color: #606266;
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

.attachments-preview {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

.attachment-item {
  position: relative;
}

.attachment-image {
  width: 80px;
  height: 80px;
  border-radius: 8px;
  object-fit: cover;
}

.attachment-file {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 6px;
}

.file-name {
  font-size: 13px;
  color: #606266;
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.attachment-remove {
  position: absolute;
  top: -8px;
  right: -8px;
  width: 20px;
  height: 20px;
  min-width: 20px;
  padding: 0;
}

.input-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
}

.input-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.hint {
  font-size: 12px;
  color: #909399;
}
</style>
