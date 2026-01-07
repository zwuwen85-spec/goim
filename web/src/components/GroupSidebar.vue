<template>
  <div class="group-sidebar">
    <!-- Close button -->
    <div class="sidebar-close">
      <el-button text circle @click="emit('close-sidebar')">
        <el-icon><Close /></el-icon>
      </el-button>
    </div>

    <el-tabs v-model="activeTab" class="sidebar-tabs">
      <!-- 群信息 & 成员 -->
      <el-tab-pane label="群信息" name="info">
        <!-- 群信息卡片 -->
        <div class="group-info-section">
          <div class="group-header">
            <el-avatar
              :size="60"
              :src="group?.avatar"
              :style="{ backgroundColor: groupColor }"
              shape="square"
            >
              <el-icon size="30" color="white"><ChatDotRound /></el-icon>
            </el-avatar>
            <div class="group-meta">
              <h3 class="group-name">{{ group?.name }}</h3>
              <p class="group-no">群号: {{ group?.id || '未设置' }}</p>
            </div>
          </div>

          <div class="group-stats">
            <div class="stat-item">
              <span class="stat-label">群主</span>
              <span class="stat-value">{{ ownerName }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">成员</span>
              <span class="stat-value">{{ members.length }}/{{ group?.max_members || 500 }}</span>
            </div>
          </div>

          <div class="group-actions">
            <el-button
              v-if="isOwnerOrAdmin"
              type="primary"
              size="small"
              @click="showInviteDialog = true"
            >
              <el-icon><Plus /></el-icon> 邀请成员
            </el-button>
            <el-button
              v-if="!isOwner"
              type="danger"
              size="small"
              @click="handleLeaveGroup"
            >
              退出群聊
            </el-button>
            <el-dropdown v-if="isOwner" trigger="click">
              <el-button size="small">
                更多 <el-icon><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="handleTransferOwnership">转让群主</el-dropdown-item>
                  <el-dropdown-item @click="handleDeleteGroup" style="color: #f56c6c">解散群聊</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>

        <!-- 群成员列表 -->
        <div class="members-section">
          <div class="section-header">
            <span>群成员 ({{ members.length }})</span>
          </div>
          <div class="members-search">
            <el-input
              v-model="memberSearch"
              placeholder="搜索成员"
              :prefix-icon="Search"
              size="small"
              clearable
            />
          </div>
          <div class="members-list">
            <div
              v-for="member in filteredMembers"
              :key="member.user_id"
              class="member-item"
            >
              <el-avatar :size="36" :src="member.user?.avatar">
                {{ member.user?.nickname?.[0] || '?' }}
              </el-avatar>
              <div class="member-info">
                <div class="member-name">
                  {{ getMemberDisplayName(member) }}
                  <el-tag v-if="member.role === 3" size="small" type="danger">群主</el-tag>
                  <el-tag v-if="member.role === 2" size="small" type="warning">管理员</el-tag>
                </div>
                <div class="member-id">ID: {{ member.user_id }}</div>
              </div>
              <el-dropdown v-if="isOwnerOrAdmin && member.user_id !== currentUserId" trigger="click">
                <el-button text circle size="small">
                  <el-icon><MoreFilled /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="handleSetNickname(member)">设置昵称</el-dropdown-item>
                    <el-dropdown-item v-if="isOwner" @click="handleSetRole(member)">
                      设为{{ member.role === 2 ? '普通成员' : '管理员' }}
                    </el-dropdown-item>
                    <el-dropdown-item @click="handleMuteMember(member)">
                      {{ member.mute_until ? '解除禁言' : '禁言' }}
                    </el-dropdown-item>
                    <el-dropdown-item @click="handleKickMember(member)" style="color: #f56c6c">
                      移除成员
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <!-- 历史消息 -->
      <el-tab-pane label="历史消息" name="history">
        <div class="history-section">
          <div class="history-search">
            <el-input
              v-model="messageSearch"
              placeholder="搜索消息内容"
              :prefix-icon="Search"
              size="small"
              clearable
              @input="handleSearchMessages"
            />
          </div>
          <div class="history-list" v-loading="loadingHistory" ref="historyListRef" @scroll="handleHistoryScroll">
            <div
              v-for="msg in filteredMessages"
              :key="msg.id"
              class="history-message"
              @click="handleScrollToMessage(msg)"
            >
              <div class="msg-header">
                <span class="msg-sender">{{ getSenderName(msg) }}</span>
                <span class="msg-time">{{ formatTime(msg.created_at) }}</span>
              </div>
              <div class="msg-content">{{ getMessageContent(msg.content) }}</div>
            </div>
            <div v-if="filteredMessages.length === 0 && !loadingHistory" class="empty-history">
              暂无历史消息
            </div>
            <div v-if="loadingHistory" class="loading-more">
              <el-icon class="is-loading"><Loading /></el-icon> 加载中...
            </div>
            <div v-if="!hasMore && filteredMessages.length > 0" class="no-more">
              没有更多消息了
            </div>
            <div v-if="hasMore && !loadingHistory" class="load-more-manual" @click="loadMoreHistory">
              <span class="text">点击加载更多历史消息</span>
              <el-icon><ArrowDown /></el-icon>
            </div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- 邀请成员对话框 -->
    <el-dialog v-model="showInviteDialog" title="邀请成员" width="400px">
      <el-select
        v-model="inviteUserId"
        filterable
        remote
        reserve-keyword
        placeholder="搜索用户ID或昵称（边输入边搜索）"
        :remote-method="handleSearch"
        :loading="searchLoading"
        loading-text="正在搜索..."
        no-data-text="未找到匹配用户"
        style="width: 100%"
        default-first-option
      >
        <el-option
          v-for="item in searchResults"
          :key="item.id"
          :label="item.nickname + ' (ID:' + item.id + ')'"
          :value="item.id"
        >
          <div style="display: flex; align-items: center; justify-content: space-between">
            <div style="display: flex; align-items: center; gap: 8px">
              <el-avatar :size="24" :src="item.avatar">{{ item.nickname?.[0] }}</el-avatar>
              <span>{{ item.nickname }}</span>
            </div>
            <span style="color: var(--el-text-color-secondary); font-size: 12px">
              ID: {{ item.id }}
            </span>
          </div>
        </el-option>
      </el-select>
      <template #footer>
        <el-button @click="showInviteDialog = false">取消</el-button>
        <el-button type="primary" @click="handleInviteMember">邀请</el-button>
      </template>
    </el-dialog>

    <!-- 转让群主对话框 -->
    <el-dialog v-model="showTransferDialog" title="转让群主" width="400px">
      <el-select v-model="transferUserId" placeholder="选择新群主" style="width: 100%">
        <el-option
          v-for="member in adminMembers"
          :key="member.user_id"
          :label="member.nickname || member.user?.nickname"
          :value="member.user_id"
        />
      </el-select>
      <template #footer>
        <el-button @click="showTransferDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmTransfer">确认转让</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import {
  ChatDotRound, Plus, Search, ArrowDown, MoreFilled, Close, Loading
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, ElTag } from 'element-plus'
import { groupApi, messageApi, authApi, type Group, type GroupMember, type Message, type User } from '../api/chat'
import { useUserStore } from '../store/user'
import { useGroupStore } from '../store/group'
import { useChatStore } from '../store/chat'
import { parseSqlNullString } from '../utils/format'

const props = defineProps<{
  group: Group | null
}>()

const emit = defineEmits(['close', 'close-sidebar', 'scroll-to-message'])

const userStore = useUserStore()
const groupStore = useGroupStore()
const chatStore = useChatStore()

const activeTab = ref('info')
const memberSearch = ref('')
const messageSearch = ref('')
const showInviteDialog = ref(false)
const showTransferDialog = ref(false)
const inviteUserId = ref('') // This will now store the selected user ID (number or string)
const searchLoading = ref(false)
const searchResults = ref<User[]>([])
let searchTimeout: ReturnType<typeof setTimeout> | null = null

const handleSearch = (query: string) => {
  if (query.trim().length === 0) {
    searchResults.value = []
    return
  }
  
  searchLoading.value = true
  
  // Debounce logic
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }
  
  searchTimeout = setTimeout(async () => {
    try {
      const res = await authApi.searchUsers(query)
      if ((res as any).code === 0) {
        searchResults.value = (res as any).data?.users || []
      }
    } catch (error) {
      console.error(error)
      searchResults.value = []
    } finally {
      searchLoading.value = false
    }
  }, 300) // 300ms delay for better UX
}

const transferUserId = ref('')
const loadingHistory = ref(false)
const historyMessages = ref<Message[]>([])
const hasMore = ref(true)
const currentUserId = userStore.currentUser?.id || 0

// Group data
const members = computed(() => groupStore.members)
const groupInfo = ref<any>({})

// Computed
const groupColor = computed(() => {
  const colors = ['#F56C6C', '#E6A23C', '#67C23A', '#409EFF', '#909399']
  const id = props.group?.id || 0
  return colors[id % colors.length]
})

const getMemberDisplayName = (member: GroupMember | undefined) => {
  if (!member) return '未知'
  try {
    // Priority: Group Nickname -> User Nickname -> '未知'
    let name = parseSqlNullString(member.nickname) || parseSqlNullString(member.user?.nickname) || ''
    
    // Handle potential JSON string format
    if (name && name.startsWith('{')) {
      try {
        const parsed = JSON.parse(name)
        name = parsed.String || ''
      } catch {
        // keep original name if parse fails
      }
    }
    
    // Fallback if empty
    if (!name) return '未知'
    
    // Clean up: Remove redundant role suffixes if they exist in the name itself
    // User complaint: "Alice群主" should just be "Alice" if there is a "群主" tag
    name = name.replace(/群主$/, '').replace(/管理员$/, '')
    
    return name
  } catch (e) {
    console.warn('Error parsing member name:', e)
    return '未知'
  }
}

const ownerName = computed(() => {
  const owner = members.value.find(m => m.role === 3)
  return getMemberDisplayName(owner)
})

const isOwner = computed(() => {
  return props.group?.owner_id === currentUserId
})

const isOwnerOrAdmin = computed(() => {
  const member = members.value.find(m => m.user_id === currentUserId)
  return member?.role === 3 || member?.role === 2
})

const filteredMembers = computed(() => {
  if (!memberSearch.value) return members.value
  const keyword = memberSearch.value.toLowerCase()
  return members.value.filter(m => {
    const name = getMemberDisplayName(m).toLowerCase()
    return name.includes(keyword) || m.user_id.toString().includes(keyword)
  })
})

const adminMembers = computed(() => {
  return members.value.filter(m => m.role === 2 || m.role === 3)
})

const filteredMessages = computed(() => {
  if (!messageSearch.value) return historyMessages.value
  const keyword = messageSearch.value.toLowerCase()
  return historyMessages.value.filter(m => {
    const content = getMessageContent(m.content).toLowerCase()
    return content.includes(keyword)
  })
})

// Methods
const getSenderName = (msg: Message) => {
  if (msg.from_user_id === currentUserId) return '我'
  const member = members.value.find(m => m.user_id === msg.from_user_id)
  if (member) {
    return getMemberDisplayName(member)
  }
  // Fallback for non-members or if member not found
  if (msg.from_user) {
    let name = parseSqlNullString(msg.from_user.nickname)
    if (name && name.startsWith('{')) {
      try {
        const parsed = JSON.parse(name)
        return parsed.String || name
      } catch {
        return name
      }
    }
    return name || `用户${msg.from_user_id}`
  }
  return `用户${msg.from_user_id}`
}

const formatTime = (time: string) => {
  const date = new Date(time)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))

  if (days === 0) {
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  } else if (days === 1) {
    return '昨天 ' + date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  } else if (days < 7) {
    const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
    return weekdays[date.getDay()] + ' ' + date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  } else {
    return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' })
  }
}

const getMessageContent = (content: string) => {
  try {
    const parsed = JSON.parse(content)
    return parsed.text || content
  } catch {
    return content
  }
}

const handleSearchMessages = async () => {
  if (!props.group || messageSearch.value.length < 2) return
  loadingHistory.value = true
  try {
    const response = await messageApi.getHistory({
      conversation_id: props.group.id,
      conversation_type: 2,
      limit: 100
    })
    if ((response as any).code === 0) {
      const msgs = (response as any).data?.messages || []
      // 按照时间降序排序（最新的在最上面）
      historyMessages.value = [...msgs].sort((a: Message, b: Message) => {
        const t1 = new Date(b.created_at).getTime()
        const t2 = new Date(a.created_at).getTime()
        return t1 - t2
      })
      hasMore.value = false // Search results are not paginated currently
    }
  } catch (error) {
    console.error('Failed to search messages:', error)
  } finally {
    loadingHistory.value = false
  }
}

const loadMoreHistory = async () => {
  if (!props.group || historyMessages.value.length === 0) return
  // Find the oldest message (which is at the bottom of the list in Newest->Oldest view)
  const lastSeq = historyMessages.value[historyMessages.value.length - 1].seq
  
  loadingHistory.value = true
  try {
    const response = await messageApi.getHistory({
      conversation_id: props.group.id,
      conversation_type: 2,
      last_seq: lastSeq,
      limit: 15
    })
    if ((response as any).code === 0) {
      const newMessages = (response as any).data?.messages || []
      
      if (newMessages.length === 0) {
        hasMore.value = false
        return
      }

      // Sort new messages descending too
      const sortedNewMessages = [...newMessages].sort((a: Message, b: Message) => {
        return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      })
      
      // Append to the end (bottom)
      historyMessages.value.push(...sortedNewMessages)
      hasMore.value = newMessages.length >= 15
    }
  } catch (error) {
    console.error('Failed to load more history:', error)
  } finally {
    loadingHistory.value = false
  }
}

const historyListRef = ref<HTMLElement>()

const handleHistoryScroll = () => {
  const el = historyListRef.value
  if (!el) return

  if (loadingHistory.value || !hasMore.value) return
  
  // Debug scroll
  // console.log('Scroll:', Math.ceil(el.scrollTop), el.scrollHeight, el.clientHeight)

  // Use a larger threshold and Math.ceil for robustness
  const distanceToBottom = el.scrollHeight - Math.ceil(el.scrollTop) - el.clientHeight
  const isBottom = distanceToBottom < 50
  
  if (isBottom) {
    console.log('Scroll hit bottom, loading more...')
    loadMoreHistory()
  }
}

const handleScrollToMessage = (msg: Message) => {
  emit('scroll-to-message', msg)
}

const handleInviteMember = async () => {
  if (!props.group) return
  if (!inviteUserId.value) {
    ElMessage.warning('请选择要邀请的用户')
    return
  }
  try {
    // inviteUserId is now bound to the selected user's ID directly from el-select
    const response = await groupApi.invite(props.group.id, { user_id: Number(inviteUserId.value) })
    if ((response as any).code === 0) {
      ElMessage.success('邀请成功')
      showInviteDialog.value = false
      inviteUserId.value = ''
      searchResults.value = [] // Clear results
      await groupStore.loadMembers(props.group.id)
    } else {
      ElMessage.error((response as any).message || '邀请失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '邀请失败')
  }
}

const handleLeaveGroup = async () => {
  if (!props.group) return
  try {
    await ElMessageBox.confirm('确定要退出群聊吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const response = await groupApi.leave(props.group.id)
    if ((response as any).code === 0) {
      ElMessage.success('已退出群聊')
      emit('close')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '操作失败')
    }
  }
}

const handleDeleteGroup = async () => {
  if (!props.group) return
  try {
    await ElMessageBox.confirm('确定要解散群聊吗？此操作不可恢复！', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const response = await groupApi.deleteGroup(props.group.id)
    if ((response as any).code === 0) {
      ElMessage.success('群聊已解散')
      emit('close')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '操作失败')
    }
  }
}

const handleTransferOwnership = () => {
  transferUserId.value = ''
  showTransferDialog.value = true
}

const confirmTransfer = async () => {
  if (!props.group) return
  if (!transferUserId.value) {
    ElMessage.warning('请选择新群主')
    return
  }
  try {
    await ElMessageBox.confirm(`确定要将群主转让给用户 ${transferUserId.value} 吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const response = await groupApi.transferOwnership(props.group.id, Number(transferUserId.value))
    if ((response as any).code === 0) {
      ElMessage.success('群主已转让')
      showTransferDialog.value = false
      emit('close')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '操作失败')
    }
  }
}

const handleSetNickname = async (member: GroupMember) => {
  if (!props.group) return
  try {
    const { value } = await ElMessageBox.prompt('请输入群昵称', '设置昵称', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputValue: member.nickname || '',
      inputPattern: /^.{1,20}$/,
      inputErrorMessage: '昵称长度为1-20个字符'
    })
    const response = await groupApi.setMemberNickname(
      props.group.id,
      member.user_id,
      value
    )
    if ((response as any).code === 0) {
      ElMessage.success('昵称已设置')
      await groupStore.loadMembers(props.group.id)
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '操作失败')
    }
  }
}

const handleSetRole = async (member: GroupMember) => {
  if (!props.group) return
  try {
    const newRole = member.role === 2 ? 1 : 2
    const response = await groupApi.setMemberRole(
      props.group.id,
      member.user_id,
      newRole
    )
    if ((response as any).code === 0) {
      ElMessage.success('角色已更新')
      await groupStore.loadMembers(props.group.id)
    }
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  }
}

const handleMuteMember = async (member: GroupMember) => {
  if (!props.group) return
  try {
    const duration = member.mute_until ? -1 : 60 // 默认禁言60分钟，-1表示解除
    const response = await groupApi.muteMember(
      props.group.id,
      member.user_id,
      duration
    )
    if ((response as any).code === 0) {
      ElMessage.success(duration > 0 ? '已禁言60分钟' : '已解除禁言')
      await groupStore.loadMembers(props.group.id)
    }
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  }
}

const handleKickMember = async (member: GroupMember) => {
  if (!props.group) return
  try {
    await ElMessageBox.confirm(`确定要将 ${member.nickname || member.user?.nickname} 移出群聊吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const response = await groupApi.kick(props.group.id, member.user_id)
    if ((response as any).code === 0) {
      ElMessage.success('成员已移除')
      await groupStore.loadMembers(props.group.id)
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '操作失败')
    }
  }
}

// Watch for group changes
watch(() => props.group, async (newGroup) => {
  if (newGroup) {
    await groupStore.loadMembers(newGroup.id)
    // Load initial history
    const response = await messageApi.getHistory({
      conversation_id: newGroup.id,
      conversation_type: 2,
      limit: 15
    })
    if ((response as any).code === 0) {
      const msgs = (response as any).data?.messages || []
      // Fix: Sort initial messages Newest -> Oldest (Descending)
      historyMessages.value = [...msgs].sort((a: Message, b: Message) => {
        return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      })
      hasMore.value = msgs.length >= 15
    }
  }
}, { immediate: true })
</script>

<style scoped>
.group-sidebar {
  width: 320px;
  height: 100%;
  background-color: var(--bg-surface);
  border-left: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  position: relative;
}

.sidebar-close {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 10;
}

.sidebar-close .el-button {
  background-color: var(--bg-body);
  transition: background-color 0.2s;
}

.sidebar-close .el-button:hover {
  background-color: var(--border-color);
}

.sidebar-tabs {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden; /* Ensure tabs don't spill out */
  min-height: 0; /* Critical for nested flex scrolling */
}

.sidebar-tabs :deep(.el-tabs__nav-scroll) {
  display: flex;
  justify-content: center;
}

.sidebar-tabs :deep(.el-tabs__content) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0; /* Critical for nested flex scrolling */
}

.sidebar-tabs :deep(.el-tab-pane) {
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0; /* Critical for nested flex scrolling */
}

/* Group Info Section */
.group-info-section {
  padding: 20px 20px; /* Reduced padding */
  border-bottom: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  background-color: var(--bg-surface);
  gap: 16px; /* Reduced gap */
  flex-shrink: 0; /* Prevent shrinking */
}

.group-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px; /* Reduced gap */
  margin-bottom: 0;
  width: 100%;
}

.group-meta {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.group-name {
  font-size: 18px; /* Slightly smaller */
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.4;
}

.group-no {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 0;
  opacity: 0.8;
}

.group-stats {
  display: flex;
  gap: 40px; /* Reduced gap */
  justify-content: center;
  width: 100%;
  padding: 12px 0; /* Reduced padding */
  border-top: 1px dashed var(--border-color);
  border-bottom: 1px dashed var(--border-color);
  margin-bottom: 0;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: center;
}

.stat-label {
  font-size: 12px;
  color: var(--text-light);
}

.stat-value {
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 600;
}

.group-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  justify-content: center;
  width: 100%;
}

/* Members Section */
.members-section {
  padding: 16px;
  display: flex;
  flex-direction: column;
  flex: 1; /* Take remaining space */
  overflow-y: auto;
  min-height: 0;
}

.section-header {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 16px;
  text-align: left;
  width: 100%;
  padding-left: 4px;
}

.members-search {
  margin-bottom: 16px;
  width: 100%;
}

.members-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
}

.member-item {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 16px;
  padding: 12px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.member-item:hover {
  background-color: var(--bg-body);
  transform: translateX(2px);
}

.member-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.member-name {
  font-size: 15px;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.member-id {
  font-size: 12px;
  color: var(--text-light);
}

/* History Section */
.history-section {
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

/* ... */

.history-list {
  flex: 1;
  overflow-y: auto !important;
  padding: 8px;
  min-height: 0;
  /* height: 0;  Removed, relying on min-height: 0 and flex: 1 */
}

.history-message {
  padding: 12px;
  border-radius: 8px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: background-color 0.2s;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.history-message:hover {
  background-color: var(--bg-body);
}

.msg-header {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.msg-sender {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

.msg-time {
  font-size: 12px;
  color: var(--text-light);
  font-weight: normal;
}

.msg-content {
  font-size: 13px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  margin-top: 4px;
}

.empty-history {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-light);
  font-size: 13px;
}

.loading-more,
.no-more {
  text-align: center;
  padding: 16px;
  color: var(--text-light);
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.load-more-manual {
  margin: 12px 24px 24px;
  padding: 8px 16px;
  background-color: var(--bg-body);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  cursor: pointer;
  color: var(--text-secondary);
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  transition: all 0.2s ease;
}

.load-more-manual:hover {
  color: var(--primary-color);
  border-color: var(--primary-color);
  background-color: var(--bg-surface);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  transform: translateY(-1px);
}

.load-more-manual:active {
  transform: translateY(0);
}
</style>
