<template>
  <div class="group-chat-container">
    <!-- Group list -->
    <div class="group-list" v-if="!groupStore.currentGroup">
      <div class="list-header">
        <h3>群聊</h3>
        <el-button type="primary" size="small" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          创建群组
        </el-button>
      </div>

      <div class="group-items">
        <template v-if="groupStore.loading">
          <el-skeleton :rows="3" animated />
        </template>
        <template v-else-if="groupStore.groups && groupStore.groups.length > 0">
          <div
            v-for="group in groupStore.groups.filter(g => g != null)"
            :key="group.id"
            class="group-item"
            @click="openGroup(group)"
          >
            <el-avatar :size="48" :style="{ backgroundColor: getGroupColor(group.id) }">
              <el-icon size="24"><ChatDotRound /></el-icon>
            </el-avatar>
            <div class="group-info">
              <div class="group-name">{{ group.name }}</div>
              <div class="group-meta">
                <span>{{ group.member_count }} 人</span>
                <span v-if="group.description" class="group-desc">{{ group.description }}</span>
              </div>
            </div>
          </div>
        </template>
        <el-empty v-else
                  description="暂无群组，点击上方按钮创建一个吧" />
      </div>
    </div>

    <!-- Chat area -->
    <div v-else class="chat-area">
      <!-- Header -->
      <div class="chat-header">
        <el-button text @click="backToList">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <div class="group-title">
          <el-avatar :size="40" :style="{ backgroundColor: getGroupColor(groupStore.currentGroup.id) }">
            <el-icon size="20"><ChatDotRound /></el-icon>
          </el-avatar>
          <div>
            <div class="group-name">{{ groupStore.currentGroup.name }}</div>
            <div class="group-status">{{ groupStore.members.length }} 位成员</div>
          </div>
        </div>
        <el-dropdown trigger="click">
          <el-button text>
            <el-icon><MoreFilled /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="showMembers = true">查看成员</el-dropdown-item>
              <el-dropdown-item @click="leaveGroup" divided>退出群组</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>

      <!-- Messages -->
      <div class="messages-container" ref="messagesRef">
        <div
          v-for="(msg, index) in groupMessages"
          :key="index"
          class="message"
          :class="{ 'is-me': isFromMe(msg) }"
        >
          <el-avatar v-if="!isFromMe(msg)" :size="32" :style="getMemberStyle(msg.from_user_id)">
            {{ getMemberInitial(msg) }}
          </el-avatar>
          <div class="message-content">
            <div class="message-header">
              <span class="message-sender">
                {{ isFromMe(msg) ? '我' : getMemberName(msg) }}
              </span>
              <span class="message-time">{{ formatTime(msg.created_at) }}</span>
            </div>
            <div class="message-body">{{ parseContent(msg.content) }}</div>
          </div>
          <el-avatar v-if="isFromMe(msg)" :size="32">
            {{ userStore.currentUser?.nickname?.[0] }}
          </el-avatar>
        </div>

        <!-- Empty state -->
        <el-empty v-if="groupMessages.length === 0"
                  description="暂无消息，开始聊天吧"
                  :image-size="100" />
      </div>

      <!-- Input area -->
      <div class="input-area">
        <el-input
          v-model="messageInput"
          type="textarea"
          :rows="3"
          placeholder="发送群消息..."
          @keydown.enter.exact="handleSend"
        />
        <div class="input-actions">
          <span class="hint">按 Enter 发送，Shift + Enter 换行</span>
          <el-button
            type="primary"
            @click="handleSend"
            :disabled="!messageInput.trim()"
          >
            发送
          </el-button>
        </div>
      </div>
    </div>

    <!-- Create Group Dialog -->
    <el-dialog v-model="showCreateDialog" title="创建群组" width="400px">
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="群名称">
          <el-input v-model="createForm.name" placeholder="请输入群名称" maxlength="50" />
        </el-form-item>
        <el-form-item label="群人数">
          <el-input-number v-model="createForm.max_members" :min="2" :max="500" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreateGroup" :disabled="!createForm.name.trim()">
          创建
        </el-button>
      </template>
    </el-dialog>

    <!-- Members Dialog -->
    <el-dialog v-model="showMembers" title="群成员" width="500px">
      <div class="members-list">
        <template v-if="groupStore.members && groupStore.members.length > 0">
          <div
            v-for="member in groupStore.members.filter(m => m != null)"
            :key="member.id"
            class="member-item"
          >
            <el-avatar :size="40">
              {{ member.user?.nickname?.[0] || '?' }}
            </el-avatar>
            <div class="member-info">
              <div class="member-name">{{ member.user?.nickname || member.nickname }}</div>
              <div class="member-role">{{ getRoleText(member.role) }}</div>
            </div>
          </div>
        </template>
        <el-empty v-else description="暂无成员" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useGroupStore } from '../store/group'
import { useUserStore } from '../store/user'
import { useChatStore } from '../store/chat'
import type { Group } from '../api/chat'

const groupStore = useGroupStore()
const userStore = useUserStore()
const chatStore = useChatStore()

const messageInput = ref('')
const messagesRef = ref<HTMLElement>()
const showCreateDialog = ref(false)
const showMembers = ref(false)

const createForm = ref({
  name: '',
  max_members: 100
})

const groupMessages = computed(() => {
  if (!groupStore.currentGroup) return []
  return chatStore.currentSession?.messages || []
})

const groupColors: Record<number, string> = {}

const getGroupColor = (groupId: number) => {
  if (!groupColors[groupId]) {
    const colors = ['#F56C6C', '#E6A23C', '#67C23A', '#409EFF', '#909399']
    groupColors[groupId] = colors[groupId % colors.length]
  }
  return groupColors[groupId]
}

const getMemberStyle = (userId: number) => {
  const member = groupStore.members.find(m => m.user_id === userId)
  if (member?.role === 'owner') return { backgroundColor: '#F56C6C' }
  if (member?.role === 'admin') return { backgroundColor: '#E6A23C' }
  return {}
}

const getMemberInitial = (msg: any) => {
  const member = groupStore.members.find(m => m.user_id === msg.from_user_id)
  return member?.user?.nickname?.[0] || member?.nickname?.[0] || '?'
}

const getMemberName = (msg: any) => {
  const member = groupStore.members.find(m => m.user_id === msg.from_user_id)
  return member?.user?.nickname || member?.nickname || `User ${msg.from_user_id}`
}

const getRoleText = (role: string) => {
  const roles: Record<string, string> = {
    owner: '群主',
    admin: '管理员',
    member: '成员'
  }
  return roles[role] || '成员'
}

const formatTime = (timestamp: string | number) => {
  const date = new Date(typeof timestamp === 'string' ? parseInt(timestamp) : timestamp * 1000)
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

const isFromMe = (msg: any) => {
  return msg.from_user_id === userStore.currentUser?.id
}

const openGroup = async (group: Group) => {
  await groupStore.setCurrentGroup(group)
  // Open chat session
  chatStore.openChat(group.id, 2, group.name)
  await nextTick()
  scrollToBottom()
}

const backToList = () => {
  groupStore.clearCurrentGroup()
  chatStore.currentSession = null
}

const handleCreateGroup = async () => {
  if (!createForm.value.name.trim()) return

  try {
    const newGroup = await groupStore.createGroup({
      name: createForm.value.name,
      max_members: createForm.value.max_members
    })

    if (newGroup) {
      ElMessage.success('群组创建成功')
      showCreateDialog.value = false
      createForm.value.name = ''
      createForm.value.max_members = 100
    }
  } catch (error: any) {
    ElMessage.error(error.message || '创建失败')
  }
}

const leaveGroup = async () => {
  if (!groupStore.currentGroup) return

  try {
    await ElMessageBox.confirm('确定要退出该群组吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const success = await groupStore.leaveGroup(groupStore.currentGroup.id)
    if (success) {
      ElMessage.success('已退出群组')
      backToList()
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '退出失败')
    }
  }
}

const handleSend = async () => {
  if (!messageInput.value.trim() || !groupStore.currentGroup) return

  const success = await groupStore.sendMessage(groupStore.currentGroup.id, messageInput.value)
  if (success) {
    messageInput.value = ''
    await nextTick()
    scrollToBottom()
  } else {
    ElMessage.error('发送失败')
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
}

onMounted(async () => {
  try {
    await groupStore.loadGroups()
  } catch (error) {
    ElMessage.error('加载群组失败')
  }
})
</script>

<style scoped>
.group-chat-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.group-list {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid #e4e7ed;
}

.list-header h3 {
  margin: 0;
  font-size: 16px;
  color: #303133;
}

.group-items {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.group-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
}

.group-item:hover {
  background: #f5f7fa;
}

.group-info {
  flex: 1;
  min-width: 0;
}

.group-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.group-meta {
  font-size: 12px;
  color: #909399;
  display: flex;
  gap: 8px;
  align-items: center;
}

.group-desc {
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

.group-title {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
}

.group-title .group-name {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.group-status {
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

.message.is-me {
  flex-direction: row-reverse;
}

.message-content {
  max-width: 70%;
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
  line-height: 1.5;
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
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
}

.hint {
  font-size: 12px;
  color: #909399;
}

.members-list {
  max-height: 400px;
  overflow-y: auto;
}

.member-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.member-item:last-child {
  border-bottom: none;
}

.member-info {
  flex: 1;
}

.member-name {
  font-size: 14px;
  color: #303133;
}

.member-role {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}
</style>
