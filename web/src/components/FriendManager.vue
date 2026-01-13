<template>
  <div class="friend-manager-container">
    <!-- Sub tabs -->
    <el-tabs v-model="subTab" class="friend-tabs">
      <!-- Friend List -->
      <el-tab-pane label="好友列表" name="list">
        <div class="friend-list-content">
          <div class="list-header">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索好友"
              :prefix-icon="Search"
              clearable
              @input="handleSearch"
            />
          </div>

          <div class="friend-items">
            <div
              v-for="friend in filteredFriends"
              :key="friend.id"
              class="friend-item"
            >
              <div class="friend-main" @click="startChat(friend)">
                <el-avatar :size="48" :src="friend.friend_user?.avatar">
                  {{ friend.friend_user?.nickname?.[0] || '?' }}
                </el-avatar>
                <div class="friend-info">
                  <div class="friend-name">{{ friend.remark || friend.friend_user?.nickname }}</div>
                  <div class="friend-meta">
                    <span class="friend-nickname" v-if="friend.remark !== friend.friend_user?.nickname">
                      昵称: {{ friend.friend_user?.nickname }}
                    </span>
                    <span class="friend-group">{{ friend.group_name || '默认分组' }}</span>
                  </div>
                </div>
              </div>
              <el-dropdown trigger="click" @command="(cmd: string | number | object) => handleFriendCommand(cmd as string, friend)">
                <el-button text circle>
                  <el-icon><MoreFilled /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="chat">发消息</el-dropdown-item>
                    <el-dropdown-item command="remark">设置备注</el-dropdown-item>
                    <el-dropdown-item command="delete" divided>删除好友</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>

            <el-empty v-if="filteredFriends.length === 0" description="暂无好友" />
          </div>
        </div>
      </el-tab-pane>

      <!-- Friend Requests -->
      <el-tab-pane name="requests">
        <template #label>
          <span>好友请求</span>
          <el-badge v-if="chatStore.friendRequests.length > 0" :value="chatStore.friendRequests.length" class="request-badge" />
        </template>
        <div class="requests-content">
          <div class="request-items">
            <div
              v-for="request in chatStore.friendRequests"
              :key="request.id"
              class="request-item"
            >
              <el-avatar :size="48" :src="request.from_user?.avatar">
                {{ request.from_user?.nickname?.[0] || '?' }}
              </el-avatar>
              <div class="request-info">
                <div class="request-name">{{ request.from_user?.nickname }}</div>
                <div class="request-message" v-if="request.message">
                  {{ request.message }}
                </div>
                <div class="request-time">{{ formatTime(request.created_at) }}</div>
              </div>
              <div class="request-actions">
                <el-button type="primary" size="small" @click="handleRequest(request.id, 'accept')">
                  接受
                </el-button>
                <el-button size="small" @click="handleRequest(request.id, 'reject')">
                  拒绝
                </el-button>
              </div>
            </div>

            <el-empty v-if="chatStore.friendRequests.length === 0" description="暂无好友请求" />
          </div>
        </div>
      </el-tab-pane>

      <!-- Add Friend -->
      <el-tab-pane label="添加好友" name="add">
        <div class="add-friend-content">
          <el-form :model="searchForm" label-position="top">
            <el-form-item label="搜索用户">
              <el-input
                v-model="searchForm.keyword"
                placeholder="输入用户名或昵称"
                @input="handleSearchInput"
                :prefix-icon="Search"
                clearable
              >
              </el-input>
            </el-form-item>
          </el-form>

          <div v-if="searchResults.length > 0" class="search-results">
            <div class="results-title">搜索结果</div>
            <div
              v-for="user in searchResults"
              :key="user.id"
              class="user-item"
            >
              <el-avatar :size="40" :src="user.avatar">
                {{ user.nickname?.[0] || '?' }}
              </el-avatar>
              <div class="user-info">
                <div class="user-name">{{ user.nickname }}</div>
                <div class="user-username">@{{ user.username }}</div>
              </div>
              <el-button
                type="primary"
                size="small"
                @click="showSendRequestDialog(user)"
              >
                添加
              </el-button>
            </div>
          </div>

          <el-empty v-else-if="hasSearched && searchResults.length === 0" description="未找到用户" />
          <el-empty v-else description="搜索用户来添加好友" />
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- Set Remark Dialog -->
    <el-dialog v-model="showRemarkDialog" title="设置备注" width="400px">
      <el-form :model="remarkForm" label-width="80px">
        <el-form-item label="备注名">
          <el-input v-model="remarkForm.remark" placeholder="请输入备注名" maxlength="50" />
        </el-form-item>
        <el-form-item label="分组">
          <el-input v-model="remarkForm.groupName" placeholder="请输入分组名" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRemarkDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveRemark">保存</el-button>
      </template>
    </el-dialog>

    <!-- Send Friend Request Dialog -->
    <el-dialog v-model="showRequestDialog" title="添加好友" width="400px">
      <el-form :model="requestForm" label-width="80px">
        <el-form-item label="对方">
          <div class="target-user">
            <el-avatar :size="32" :src="targetUser?.avatar">
              {{ targetUser?.nickname?.[0] }}
            </el-avatar>
            <span>{{ targetUser?.nickname }}</span>
          </div>
        </el-form-item>
        <el-form-item label="验证消息">
          <el-input
            v-model="requestForm.message"
            type="textarea"
            :rows="3"
            placeholder="请输入验证消息（可选）"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRequestDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSendRequest">发送请求</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { friendApi, authApi } from '../api/chat'
import { useChatStore } from '../store/chat'
import { useUserStore } from '../store/user'
import type { Friend, FriendRequest, User } from '../api/chat'
import { parseSqlNullString } from '../utils/format'

const chatStore = useChatStore()
const userStore = useUserStore()

const subTab = ref('list')
const searchKeyword = ref('')
const searchForm = ref({ keyword: '' })
const searchResults = ref<User[]>([])
const hasSearched = ref(false)

const showRemarkDialog = ref(false)
const remarkForm = ref({ remark: '', groupName: '' })
const currentFriend = ref<Friend | null>(null)

const showRequestDialog = ref(false)
const requestForm = ref({ message: '' })
const targetUser = ref<User | null>(null)

const filteredFriends = computed(() => {
  if (!searchKeyword.value) return chatStore.friends

  const keyword = searchKeyword.value.toLowerCase()
  return chatStore.friends.filter(friend => {
    const nickname = parseSqlNullString(friend.friend_user?.nickname).toLowerCase()
    const remark = friend.remark?.toLowerCase() || ''
    return nickname.includes(keyword) || remark.includes(keyword)
  })
})

const getNickname = (user: any) => {
  return parseSqlNullString(user?.nickname)
}

const formatTime = (timestamp: string) => {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  if (diff < 604800000) return `${Math.floor(diff / 86400000)} 天前`

  return date.toLocaleDateString('zh-CN')
}

const handleSearch = () => {
  // Filter happens in computed
}

const startChat = (friend: Friend) => {
  const userId = friend.friend_user?.id || friend.friend_id
  // Emit 'chat' event to parent (Main.vue) to handle view switching
  emit('chat', friend)
}

const emit = defineEmits(['chat'])

const handleFriendCommand = async (command: string, friend: Friend) => {
  switch (command) {
    case 'chat':
      startChat(friend)
      break
    case 'remark':
      currentFriend.value = friend
      remarkForm.value = {
        remark: friend.remark || friend.friend_user?.nickname || '',
        groupName: friend.group_name || ''
      }
      showRemarkDialog.value = true
      break
    case 'delete':
      await handleDeleteFriend(friend)
      break
  }
}

const handleSaveRemark = async () => {
  if (!currentFriend.value) return

  try {
    const response = await friendApi.updateRemark({
      friend_id: currentFriend.value.friend_id,
      remark: remarkForm.value.remark,
      group_name: remarkForm.value.groupName
    })

    if ((response as any).code === 0) {
      ElMessage.success('备注已更新')
      showRemarkDialog.value = false
      await chatStore.loadFriends()
    }
  } catch (error: any) {
    ElMessage.error(error.message || '更新失败')
  }
}

const handleDeleteFriend = async (friend: Friend) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除好友 "${friend.remark || friend.friend_user?.nickname}" 吗？`,
      '删除好友',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const response = await friendApi.deleteFriend(friend.friend_id)
    if ((response as any).code === 0) {
      ElMessage.success('已删除好友')
      await chatStore.loadFriends()
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

const handleSearchInput = async () => {
  if (!searchForm.value.keyword.trim()) {
    searchResults.value = []
    hasSearched.value = false
    return
  }
  
  // Debounce search
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    searchUsers()
  }, 300)
}

let searchTimeout: ReturnType<typeof setTimeout>

const searchUsers = async () => {
  if (!searchForm.value.keyword.trim()) return

  hasSearched.value = true
  try {
    const response = await authApi.searchUsers(searchForm.value.keyword)
    if ((response as any).code === 0) {
      searchResults.value = (response as any).data.users || []
    }
  } catch (error: any) {
    ElMessage.error(error.message || '搜索失败')
  }
}

const showSendRequestDialog = (user: User) => {
  targetUser.value = user
  requestForm.value.message = ''
  showRequestDialog.value = true
}

const handleSendRequest = async () => {
  if (!targetUser.value) return

  // Prevent adding self
  if (targetUser.value.id === userStore.currentUser?.id) {
    ElMessage.warning('不能添加自己为好友')
    return
  }

  try {
    const response = await friendApi.sendRequest({
      to_user_id: targetUser.value.id,
      message: requestForm.value.message
    })

    if ((response as any).code === 0) {
      ElMessage.success('好友请求已发送')
      showRequestDialog.value = false
      searchResults.value = []
      searchForm.value.keyword = ''
      subTab.value = 'list' // Switch back to list tab
    } else {
      // Handle business error (e.g. already friends, request pending)
      ElMessage.error((response as any).message || '发送失败')
      // Stay on the add friend tab but close dialog to allow retry/search others
      showRequestDialog.value = false
    }
  } catch (error: any) {
    // Handle network or other errors
    ElMessage.error(error.message || '发送失败')
    showRequestDialog.value = false
  }
}

const handleRequest = async (requestId: number, action: 'accept' | 'reject') => {
  try {
    const api = action === 'accept' ? friendApi.acceptRequest : friendApi.rejectRequest
    const response = await api(requestId)

    if ((response as any).code === 0) {
      ElMessage.success(action === 'accept' ? '已接受好友请求' : '已拒绝好友请求')
      await chatStore.loadFriendRequests()
      if (action === 'accept') {
        await chatStore.loadFriends()
      }
    }
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  }
}

onMounted(async () => {
  await chatStore.loadFriends()
  await chatStore.loadFriendRequests()
})
</script>

<style scoped>
.friend-manager-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.friend-tabs {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.friend-tabs :deep(.el-tabs__content) {
  flex: 1;
  overflow: hidden;
}

.friend-tabs :deep(.el-tabs__nav-scroll) {
  display: flex;
  justify-content: center;
}

.friend-tabs :deep(.el-tab-pane) {
  height: 100%;
}

.friend-list-content,
.requests-content,
.add-friend-content {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 12px;
}

.list-header {
  margin-bottom: 12px;
}

.friend-items,
.request-items,
.search-results {
  flex: 1;
  overflow-y: auto;
}

.friend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px;
  border-radius: 8px;
  transition: background 0.2s;
}

.friend-item:hover {
  background: #f5f7fa;
}

.friend-main {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.friend-info {
  flex: 1;
  min-width: 0;
}

.friend-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 2px;
}

.friend-meta {
  font-size: 12px;
  color: #909399;
  display: flex;
  gap: 8px;
}

.friend-nickname {
  color: #67C23A;
}

.request-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  margin-bottom: 8px;
}

.request-info {
  flex: 1;
  min-width: 0;
}

.request-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.request-message {
  font-size: 12px;
  color: #606266;
  margin-bottom: 4px;
}

.request-time {
  font-size: 11px;
  color: #909399;
}

.request-actions {
  display: flex;
  gap: 8px;
}

.user-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  margin-bottom: 8px;
}

.user-info {
  flex: 1;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.user-username {
  font-size: 12px;
  color: #909399;
}

.target-user {
  display: flex;
  align-items: center;
  gap: 8px;
}

.results-title {
  font-size: 13px;
  color: #909399;
  margin-bottom: 8px;
  padding-top: 8px;
}

.request-badge {
  margin-left: 4px;
}
</style>
