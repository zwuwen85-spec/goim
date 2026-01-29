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
        <div class="header-info">
          <div class="header-title">{{ title }}</div>
          <div class="header-subtitle" v-if="subtitle">{{ subtitle }}</div>
        </div>
      </div>
      <div class="header-actions">
        <slot name="actions"></slot>
      </div>
    </div>

    <!-- Messages -->
    <div class="messages-container" ref="messagesRef" @scroll="handleScroll">
      <div v-if="loading" class="loading-state">
        <el-skeleton :rows="3" animated />
      </div>
      <template v-else>
        <div
          v-for="(msg, index) in messages"
          :key="(msg.msg_id || msg.id || index) + '-' + membersRefreshKey"
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
              @click="handleAvatarClick(msg)"
              :class="{ 'clickable': !isMe(msg) }"
            >
              {{ getSenderName(msg)?.[0] || '?' }}
            </el-avatar>
            
            <div class="message-content-group">
              <div class="message-sender-name" v-if="!isMe(msg) && showSenderName">
                {{ getSenderName(msg) }}
              </div>
              
              <div class="message-bubble" :class="{ 'typing': msg.isTyping, 'is-pure-file': isPureFile(msg.content) }">
                <template v-if="msg.isTyping">
                  <div class="typing-indicator">
                    <span></span><span></span><span></span>
                  </div>
                </template>
                <template v-else>
                  <div class="message-text" v-html="parseContent(msg.content)" @click="handleMessageClick"></div>
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

    <!-- Image Preview Overlay -->
    <div v-if="previewImageVisible" class="image-preview-overlay" @click="closePreview">
      <div class="image-preview-window" @click.stop>
        <div class="preview-header">
          <span class="preview-title">图片预览</span>
          <div class="preview-controls">
            <el-button link class="control-btn" @click="handleDownload">
              <el-icon><Download /></el-icon>
            </el-button>
            <el-button link class="control-btn close-btn" @click="closePreview">
              <el-icon><Close /></el-icon>
            </el-button>
          </div>
        </div>
        <div class="preview-body">
          <div class="nav-btn prev" v-if="chatImages.length > 1" @click.stop="prevImage" :class="{ disabled: currentImageIndex <= 0 }">
            <el-icon><ArrowLeft /></el-icon>
          </div>
          <img :src="previewImageUrl" class="preview-image" />
          <div class="nav-btn next" v-if="chatImages.length > 1" @click.stop="nextImage" :class="{ disabled: currentImageIndex >= chatImages.length - 1 }">
            <el-icon><ArrowRight /></el-icon>
          </div>
        </div>
      </div>
    </div>

    <!-- Input -->
    <div class="input-area">
      <!-- Attachments preview -->
      <div v-if="attachedImages.length > 0 || attachedFiles.length > 0" class="attachments-preview">
        <div v-for="(img, idx) in attachedImages" :key="'img-' + idx" class="attachment-item">
          <img :src="img" class="attachment-image" />
          <el-button circle size="small" class="attachment-remove" @click="removeImage(idx)">
            <el-icon><Close /></el-icon>
          </el-button>
        </div>
        <div v-for="(file, idx) in attachedFiles" :key="'file-' + idx" class="attachment-item">
          <el-icon><Document /></el-icon>
          <span>{{ file.name }}</span>
          <el-button circle size="small" class="attachment-remove" @click="removeFile(idx)">
            <el-icon><Close /></el-icon>
          </el-button>
        </div>
      </div>

      <div class="input-toolbar">
        <input
          ref="fileInputRef"
          type="file"
          style="display: none"
          accept="image/*,.txt,.md,.json,.xml,.yaml,.yml,.log"
          @change="handleFileInputChange"
        />
        <el-button text circle size="small" @click="() => fileInputRef?.click()" title="上传文件或图片">
          <el-icon><Folder /></el-icon>
        </el-button>
        <span class="input-hint-inline">支持粘贴图片、上传文件</span>
      </div>
      <el-input
        v-model="inputValue"
        type="textarea"
        :rows="3"
        :placeholder="isAIChat ? '和 AI 对话... (可粘贴图片)' : '输入消息...'"
        resize="none"
        @keydown.enter.exact.prevent="handleSend"
        @paste="handlePaste"
        class="chat-input"
      />
      <div class="input-footer">
        <span class="input-hint">Enter 发送，Shift + Enter 换行</span>
        <el-button type="primary" @click="handleSend" :disabled="!canSend" :loading="sending">
          发送
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, watch, onMounted, computed, onUnmounted } from 'vue'
import { UserFilled, Files, Folder, Document, Close, Download, ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { parseSqlNullString } from '../utils/format'
import { ElMessage } from 'element-plus'
import { aiApi } from '../api/chat'

const props = defineProps<{
  messages: any[]
  currentUserId: number
  title: string
  subtitle?: string
  avatar?: string
  avatarStyle?: any
  sending?: boolean
  loading?: boolean
  hasMore?: boolean
  showSenderName?: boolean
  getSenderName?: (msg: any) => string
  getSenderAvatar?: (msg: any) => string
  getSenderStyle?: (msg: any) => any
  isAIChat?: boolean
}>()

const emit = defineEmits(['send', 'send-multimodal', 'load-more', 'avatar-click'])

const inputValue = ref('')
const messagesRef = ref<HTMLElement>()
const fileInputRef = ref<HTMLInputElement>()
const firstMsgId = ref<string | number | null>(null)
const membersRefreshKey = computed(() => props.messages && Array.isArray(props.messages) ? (props.messages as any)._version || 0 : 0)

// File upload state for AI chat
const attachedImages = ref<string[]>([])
const attachedFiles = ref<any[]>([])
const uploadingFiles = ref(false)
const previewImageVisible = ref(false)
const previewImageUrl = ref('')
const currentImageIndex = ref(0)

// Extract all images from messages for gallery navigation
const chatImages = computed(() => {
  const images: string[] = []
  if (!props.messages) return images
  
  props.messages.forEach(msg => {
    const raw = parseSqlNullString(msg.content) || ''
    let text = raw
    try {
        const parsed = JSON.parse(raw)
        if (parsed && typeof parsed === 'object' && parsed.text) {
            text = parsed.text
        }
    } catch {}
    
    // Match [图片: url]
    const matches = text.matchAll(/\[图片:\s*([^\]]+)\]/g)
    for (const match of matches) {
        const url = match[1]
        const fullUrl = url.startsWith('/') ? `${window.location.origin}${url}` : url
        images.push(fullUrl)
    }
  })
  return images
})

const prevImage = () => {
  if (currentImageIndex.value > 0) {
    currentImageIndex.value--
    previewImageUrl.value = chatImages.value[currentImageIndex.value]
  }
}

const nextImage = () => {
  if (currentImageIndex.value < chatImages.value.length - 1) {
    currentImageIndex.value++
    previewImageUrl.value = chatImages.value[currentImageIndex.value]
  }
}

const handleKeydown = (e: KeyboardEvent) => {
  if (!previewImageVisible.value) return
  if (e.key === 'ArrowLeft') prevImage()
  if (e.key === 'ArrowRight') nextImage()
  if (e.key === 'Escape') closePreview()
}

// Handle message content clicks for image preview (must be defined before parseContent)
const handleMessageClick = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (target.tagName === 'IMG' && target.classList.contains('message-image')) {
    const src = target.getAttribute('src')
    if (src) {
      previewImageUrl.value = src
      previewImageVisible.value = true
      
      // Find index
      const index = chatImages.value.indexOf(src)
      if (index !== -1) {
        currentImageIndex.value = index
      }
    }
  }
}

const closePreview = () => {
  previewImageVisible.value = false
  previewImageUrl.value = ''
}

const handleScroll = () => {
  if (!messagesRef.value || props.loading) return
  const el = messagesRef.value

  // Load more when scrolling to top
  if (props.hasMore !== false && el.scrollTop < 20 && props.messages.length > 0) {
    emit('load-more')
  }
}

const isMe = (msg: any) => {
  // If role is 'user', it's usually AI chat from user
  if (msg.role === 'user') return true
  // Standard check
  return msg.from_user_id === props.currentUserId
}

const isPureFile = (content: any) => {
  let raw = parseSqlNullString(content) || ''
  try {
    const parsed = JSON.parse(raw)
    if (parsed && typeof parsed === 'object' && parsed.text) {
      raw = parsed.text
    }
  } catch (e) {
    // Not JSON, use raw string
  }
  // Check if string matches strictly [文件: ...] or [图片: ...] with no other text
  const trimmed = raw.trim()
  return /^\[文件:\s*[^\]]+\]$/.test(trimmed) || /^\[图片:\s*[^\]]+\]$/.test(trimmed)
}

const parseContent = (content: any) => {
  const rawContent = parseSqlNullString(content)
  if (!rawContent) return ''

  let text = rawContent
  try {
    const parsed = JSON.parse(rawContent)
    text = parsed.text || rawContent
  } catch {
    text = rawContent
  }

  // Convert [图片: url] to <img> tags with full URL
  text = text.replace(/\[图片:\s*([^\]]+)\]/g, (match, url) => {
    // Use full URL for preview to work correctly
    const src = url.startsWith('/') ? `${window.location.origin}${url}` : url
    return `<img src="${src}" class="message-image" alt="图片" data-url="${url}" />`
  })

  // Convert [文件: name] to clickable file links with download functionality
  text = text.replace(/\[文件:\s*([^\]]+)\]/g, (match, fileName) => {
    // Extract file URL if present in the format [文件: /uploads/.../filename]
    const urlMatch = fileName.match(/\/uploads\/ai\/ai_file_\d+\.(\w+)/)
    const displayName = urlMatch ? fileName.split('/').pop() : fileName
    const downloadUrl = fileName.startsWith('/') ? `${window.location.origin}${fileName}` : fileName
    
    // Determine file extension for icon style
    const ext = displayName?.split('.').pop()?.toLowerCase() || 'file'
    
    let iconColor = '#a6a6a6' // Default gray
    let iconType = 'file'
    
    if (['pdf'].includes(ext)) {
      iconColor = '#ff4d4f'
      iconType = 'pdf'
    } else if (['doc', 'docx'].includes(ext)) {
      iconColor = '#409eff'
      iconType = 'doc'
    } else if (['xls', 'xlsx', 'csv'].includes(ext)) {
      iconColor = '#67c23a'
      iconType = 'xls'
    } else if (['ppt', 'pptx'].includes(ext)) {
      iconColor = '#f56c6c'
      iconType = 'ppt'
    } else if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) {
      iconColor = '#e6a23c'
      iconType = 'zip'
    } else if (['txt', 'md', 'json', 'log', 'xml'].includes(ext)) {
      iconColor = '#909399'
      iconType = 'txt'
    } else if (['py', 'js', 'ts', 'java', 'c', 'cpp', 'go', 'html', 'css', 'sql'].includes(ext)) {
      iconColor = '#606266'
      iconType = 'code'
    }

    // IMPORTANT: Return a single line string to avoid issues with replace(/\n/g, '<br>') later
    return `<a href="${downloadUrl}" download="${displayName}" class="message-file" title="点击下载" onclick="event.stopPropagation()"><div class="message-file-info"><span class="message-file-name">${displayName}</span><span class="message-file-size">3.97 KB 已发送</span></div><div class="message-file-icon-wrapper"><svg viewBox="0 0 1024 1024" class="message-file-icon" xmlns="http://www.w3.org/2000/svg"><path fill="${iconColor === '#a6a6a6' ? '#dcdfe6' : iconColor}" opacity="${iconType === 'code' ? '0.1' : '1'}" d="M854.6 288.6L639.4 73.4c-6-6-14.1-9.4-22.6-9.4H192c-17.7 0-32 14.3-32 32v832c0 17.7 14.3 32 32 32h640c17.7 0 32-14.3 32-32V311.3c0-8.5-3.4-16.6-9.4-22.7zM790.2 326H602V137.8L790.2 326z m1.8 562H232V136h302v216c0 23.2 18.8 42 42 42h216v494z"/><path fill="${iconColor}" opacity="0.5" d="M304 464h216v48H304z m0 136h416v48H304z m0 136h416v48H304z"/>${iconType === 'pdf' ? `<path fill="${iconColor}" d="M304 736h160v48H304z"/>` : ''}${iconType === 'doc' ? `<path fill="${iconColor}" d="M304 736h100v48H304z"/>` : ''}${iconType === 'xls' ? `<path fill="${iconColor}" d="M304 736h100v48H304z m160 0h100v48H464z"/>` : ''}${iconType === 'code' ? `<path fill="${iconColor}" d="M365.4 395.7c-12.2-13.1-32.8-13.9-45.9-1.7l-128 120c-13.6 12.8-13.6 34.2 0 46.9l128 120c12.7 11.9 32.7 11.4 44.9-1.2 12.2-13.1 11.8-33.5-1.2-46.1L263.8 537.5l99.3-95.9c12.5-12 12.9-32.1 2.3-45.9z m304.6 1.7c-13.1-12.2-33.7-11.4-45.9 1.7-10.6 13.8-10.2 33.9 2.3 45.9l99.3 95.9-99.3 96.5c-13.1 12.7-13.4 33-1.2 46.1 12.2 12.7 32.2 13.1 44.9 1.2l128-120c13.6-12.8 13.6-34.2 0-46.9l-128-120z m-152.9-29.6c-17.1-4-34.2 6.5-38.2 23.6l-88 376c-4 17.1 6.5 34.2 23.6 38.2 17.1 4 34.2-6.5 38.2-23.6l88-376c4-17.1-6.5-34.2-23.6-38.2z"/>` : ''}${iconType === 'file' || iconType === 'txt' || iconType === 'zip' || iconType === 'ppt' ? `<path fill="${iconColor}" opacity="0.5" d="M304 464h216v48H304z m0 136h416v48H304z m0 136h416v48H304z"/>` : ''}</svg></div></a>`
  })

  // Convert newlines to <br>
  text = text.replace(/\n/g, '<br>')

  return text
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
  if (!inputValue.value.trim() && attachedImages.value.length === 0 && attachedFiles.value.length === 0) return
  if (props.sending || uploadingFiles.value) return

  // For AI chat with attachments, send multimodal message
  if (props.isAIChat && (attachedImages.value.length > 0 || attachedFiles.value.length > 0)) {
    emit('send-multimodal', {
      message: inputValue.value,
      imageUrls: [...attachedImages.value],
      fileUrls: attachedFiles.value.map(f => f.url)
    })
  } else if (attachedImages.value.length > 0 || attachedFiles.value.length > 0) {
    // For regular chat, include attachments in message
    const attachmentText = attachedImages.value.map(url => `[图片: ${url}]`).join('\n') +
                           attachedFiles.value.map(f => `[文件: ${f.name}]`).join('\n')
    emit('send', inputValue.value ? `${inputValue.value}\n${attachmentText}` : attachmentText)
  } else {
    emit('send', inputValue.value)
  }

  // Clear attachments and input
  attachedImages.value = []
  attachedFiles.value = []
  inputValue.value = ''
}

const handleAvatarClick = (msg: any) => {
  // Don't emit for own messages
  if (isMe(msg)) return

  // Emit the user id from the message
  emit('avatar-click', msg.from_user_id)
}

// Upload file and get URL (must be defined before functions that use it)
const uploadFile = async (file: File) => {
  try {
    uploadingFiles.value = true

    // Check file size (10MB limit)
    const maxSize = 10 * 1024 * 1024
    if (file.size > maxSize) {
      ElMessage.error('文件大小不能超过 10MB')
      return
    }

    const response = await aiApi.uploadFile(file) as any

    if (response.code === 0) {
      const data = response.data
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
        ElMessage.success(`文件 "${data.file_name}" 上传成功`)
      }
    } else {
      ElMessage.error(response.message || '上传失败')
    }
  } catch (error: any) {
    console.error('Upload error:', error)
    const errorMsg = error?.response?.data?.message || error?.message || '文件上传失败'
    ElMessage.error(errorMsg)
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

// Handle paste event for images
const handlePaste = async (event: ClipboardEvent) => {
  const items = event.clipboardData?.items
  if (!items) return

  for (const item of Array.from(items)) {
    if (item.type.indexOf('image') !== -1) {
      event.preventDefault()
      const file = item.getAsFile()
      if (file) {
        await uploadFile(file)
      }
    }
  }
}

// Handle native file input change
const handleFileInputChange = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    await uploadFile(file)
  }
  // Reset input so same file can be selected again
  target.value = ''
}

const handleDownload = () => {
  if (!previewImageUrl.value) return
  const link = document.createElement('a')
  link.href = previewImageUrl.value
  link.download = `image_${Date.now()}.png`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesRef.value) {
      const lastMessage = messagesRef.value.lastElementChild
      if (lastMessage) {
        lastMessage.scrollIntoView({ behavior: 'smooth', block: 'end' })
      } else {
        messagesRef.value.scrollTop = messagesRef.value.scrollHeight
      }
    }
  })
}

const canSend = computed(() => {
  return (inputValue.value.trim() || attachedImages.value.length > 0 || attachedFiles.value.length > 0) && !props.sending && !uploadingFiles.value
})

const scrollToMessage = (messageId: string | number) => {
  nextTick(() => {
    const el = document.getElementById('msg-' + messageId)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
  })
}

watch(() => props.messages, async (newMsgs) => {
  if (!messagesRef.value) return

  const el = messagesRef.value
  const oldHeight = el.scrollHeight
  const oldTop = el.scrollTop
  // Check if was at bottom (allow some slack)
  const wasAtBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 50

  const newFirstId = newMsgs.length > 0 ? (newMsgs[0].msg_id || newMsgs[0].id) : null
  const isPrepend = firstMsgId.value !== null && newFirstId !== null && newFirstId !== firstMsgId.value

  firstMsgId.value = newFirstId

  await nextTick()

  if (wasAtBottom) {
    el.scrollTop = el.scrollHeight
  }
  // else if (isPrepend && oldTop < 50) {
  //   // Browser scroll anchoring should handle this now
  //   // el.scrollTop = el.scrollHeight - oldHeight + oldTop
  // }
}, { deep: true })

onMounted(() => {
  scrollToBottom()
  if (props.messages.length > 0) {
    firstMsgId.value = props.messages[0].msg_id || props.messages[0].id
  }
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})

defineExpose({
  scrollToBottom,
  scrollToMessage
})

// Default helpers if not provided
const defaultGetSenderName = (msg: any) => msg.senderName || 'User'
const defaultGetSenderAvatar = (msg: any) => msg.senderAvatar || ''
const defaultGetSenderStyle = (_msg: any) => ({})

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
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-info {
  display: flex;
  flex-direction: column;
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

.no-more-history {
  text-align: center;
  padding: 12px 0;
  color: var(--text-secondary);
  font-size: 12px;
  opacity: 0.8;
  transform: scale(0.9);
}

.loading-bar {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 12px 0;
  color: var(--primary-color);
  height: 40px; /* Fixed height to minimize layout shift calculation errors */
  box-sizing: border-box;
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

.message-avatar.clickable {
  cursor: pointer;
  transition: all 0.2s;
}

.message-avatar.clickable:hover {
  transform: scale(1.05);
  opacity: 0.8;
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

.message-bubble.is-pure-file {
  background: transparent !important;
  padding: 0 !important;
  box-shadow: none !important;
  border-radius: 0;
}

.is-me .message-bubble.is-pure-file {
  background: transparent !important;
}

.is-pure-file .message-text :deep(.message-file) {
  /* Remove negative margins when bubble is transparent */
  margin: 0;
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
  align-items: center;
}

.input-hint-inline {
  font-size: 12px;
  color: #909399;
  font-weight: 500;
  padding: 0 8px;
  border-left: 1px solid #e4e7ed;
  margin-left: 8px;
}

.attachments-preview {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.attachment-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  border: 1px solid #dee2e6;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  color: #495057;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  transition: all 0.2s ease;
}

.attachment-item:hover {
  background: linear-gradient(135deg, #e9ecef 0%, #dee2e6 100%);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.12);
}

.attachment-image {
  max-height: 60px;
  max-width: 60px;
  border-radius: 4px;
  object-fit: cover;
}

.attachment-remove {
  position: absolute;
  top: -8px;
  right: -8px;
  padding: 4px;
  min-width: 20px;
  height: 20px;
  background: var(--bg-surface);
  border: 1px solid var(--border-color);
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

.message-text {
  line-height: 1.6;
  word-break: break-word;
  white-space: pre-wrap;
}

.message-text :deep(.message-image) {
  max-width: 100%;

  max-height: 200px;
  border-radius: 8px;
  margin: 4px 0;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.message-text :deep(.message-image):hover {
  transform: scale(1.02);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.message-text :deep(.message-file) {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 240px;
  max-width: 100%;
  padding: 12px 16px;
  background: #fff;
  border-radius: 8px;
  /* Use negative margin to counteract bubble padding and overlay the bubble background */
  margin: -12px -16px;
  text-decoration: none;
  cursor: pointer;
  border: 1px solid #ebeef5;
  transition: all 0.2s;
  box-shadow: 0 2px 12px 0 rgba(0,0,0,0.05);
}

.message-text :deep(.message-file:hover) {
  box-shadow: 0 4px 16px 0 rgba(0,0,0,0.1);
  transform: translateY(-1px);
}

.message-text :deep(.message-file-info) {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
  margin-right: 16px;
  height: 44px;
  justify-content: space-between;
}

.message-text :deep(.message-file-name) {
  font-size: 15px;
  color: #303133;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
  margin-bottom: 0;
}

.message-text :deep(.message-file-size) {
  font-size: 12px;
  color: #909399;
}

.message-text :deep(.message-file-icon-wrapper) {
  width: 42px;
  height: 42px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.message-text :deep(.message-file-icon) {
  width: 100%;
  height: 100%;
}

/* Image preview overlay */
.image-preview-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.65);
  z-index: 9999;
  display: flex;
  justify-content: center;
  align-items: center;
  backdrop-filter: blur(8px);
  animation: fadeIn 0.2s ease-out;
}

.image-preview-window {
  background: #2c2c2c;
  border-radius: 8px;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  max-width: 90vw;
  max-height: 90vh;
  min-width: 400px;
  min-height: 300px;
  animation: zoomIn 0.25s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.preview-header {
  height: 48px;
  background: #363636;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 16px;
  border-bottom: 1px solid #444;
}

.preview-title {
  color: #e0e0e0;
  font-size: 14px;
  font-weight: 500;
}

.preview-controls {
  display: flex;
  gap: 8px;
}

.control-btn {
  color: #a0a0a0;
  font-size: 18px;
  padding: 4px;
  height: auto;
}

.control-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
}

.control-btn.close-btn:hover {
  background: #ff4d4f;
  color: #fff;
}

.preview-body {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  background: #1e1e1e;
  padding: 20px;
  overflow: hidden;
  position: relative;
}

.preview-image {
  max-width: 100%;
  max-height: calc(90vh - 88px); /* 90vh window height - header height - padding */
  object-fit: contain;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  user-select: none;
}

.nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 44px;
  height: 44px;
  background: rgba(0, 0, 0, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  display: flex;
  justify-content: center;
  align-items: center;
  color: rgba(255, 255, 255, 0.8);
  cursor: pointer;
  transition: all 0.3s ease;
  z-index: 10;
  backdrop-filter: blur(4px);
  opacity: 0;
  visibility: hidden;
}

.preview-body:hover .nav-btn {
  opacity: 1;
  visibility: visible;
}

.nav-btn:hover {
  background: rgba(64, 158, 255, 0.9);
  color: #fff;
  border-color: transparent;
  transform: translateY(-50%) scale(1.1);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  opacity: 1; /* Ensure it stays visible on hover */
}

.nav-btn :deep(.el-icon) {
  font-size: 20px;
  font-weight: bold;
}

.nav-btn.prev {
  left: 12px;
}

.nav-btn.next {
  right: 12px;
}

.nav-btn.disabled {
  opacity: 0.3;
  cursor: not-allowed;
  pointer-events: none;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes zoomIn {
  from { transform: scale(0.95); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}
</style>
