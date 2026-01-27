<template>
  <div class="user-profile">
    <div v-if="loading" class="loading-state">
      <el-skeleton :rows="3" animated />
    </div>
    <div v-else-if="userInfo" class="profile-content">
      <!-- Avatar Section -->
      <div class="avatar-section">
        <el-avatar :size="80" :src="displayAvatarUrl">
          <el-icon size="40"><User /></el-icon>
        </el-avatar>
      </div>

      <!-- User Info -->
      <div class="info-section">
        <div class="info-item">
          <span class="info-label">用户名:</span>
          <span class="info-value">{{ userInfo.username }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">昵称:</span>
          <span class="info-value">{{ userInfo.nickname }}</span>
        </div>
        <div class="info-item" v-if="userInfo.signature">
          <span class="info-label">个性签名:</span>
          <span class="info-value">{{ userInfo.signature }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">用户ID:</span>
          <span class="info-value">{{ userInfo.id }}</span>
        </div>
      </div>

      <!-- Actions -->
      <div class="action-section">
        <div class="action-buttons">
          <!-- Send message button - always show for other users -->
          <el-button
            v-if="!isCurrentUser"
            type="primary"
            @click="handleStartChat"
          >
            发送消息
          </el-button>

          <!-- Add friend button - only show if not already friends -->
          <el-button
            v-if="!isCurrentUser && !isFriend"
            type="success"
            @click="handleAddFriend"
          >
            申请加好友
          </el-button>

          <!-- Current user indicator -->
          <el-tag v-if="isCurrentUser" type="info">这是你自己</el-tag>
        </div>
      </div>
    </div>
    <div v-else class="empty-state">
      <el-empty description="用户不存在" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { User } from '@element-plus/icons-vue'
import { authApi, friendApi, type User as UserType } from '../api/chat'
import { useUserStore } from '../store/user'
import { useChatStore } from '../store/chat'
import { useGroupStore } from '../store/group'
import { parseSqlNullString } from '../utils/format'

const props = defineProps<{
  userId: number
}>()

const emit = defineEmits(['close', 'start-chat'])

const userStore = useUserStore()
const chatStore = useChatStore()
const groupStore = useGroupStore()

const userInfo = ref<any>(null)
const loading = ref(false)

const isCurrentUser = computed(() => {
  return props.userId === userStore.currentUser?.id
})

const isFriend = computed(() => {
  if (!userInfo.value) return false
  return chatStore.friends.some(f =>
    (f.friend_user?.id === props.userId) || (f.friend_id === props.userId)
  )
})

const displayAvatarUrl = computed(() => {
  const avatar = userInfo.value?.avatar
  if (!avatar) return undefined

  if (avatar.startsWith('http://') || avatar.startsWith('https://')) {
    return avatar
  }

  if (avatar.startsWith('/uploads/')) {
    return window.location.origin + avatar
  }

  return avatar
})

const loadUserProfile = async () => {
  loading.value = true

  // First, try to find user in friends list
  const friend = chatStore.friends.find(f =>
    (f.friend_user?.id === props.userId) || (f.friend_id === props.userId)
  )
  if (friend && friend.friend_user) {
    userInfo.value = {
      id: friend.friend_user.id,
      username: friend.friend_user.username,
      nickname: parseSqlNullString(friend.friend_user.nickname),
      avatar: friend.friend_user.avatar,
      status: friend.friend_user.status
    }
    loading.value = false
    return
  }

  // Try to find in group members
  for (const member of groupStore.members) {
    if (member.user_id === props.userId && member.user) {
      userInfo.value = {
        id: member.user.id,
        username: member.user.username,
        nickname: parseSqlNullString(member.user.nickname),
        avatar: member.user.avatar,
        status: member.user.status
      }
      loading.value = false
      return
    }
  }

  // If it's current user, use current user info
  if (props.userId === userStore.currentUser?.id) {
    userInfo.value = userStore.currentUser
    loading.value = false
    return
  }

  // Create minimal user object if not found
  userInfo.value = {
    id: props.userId,
    username: `User${props.userId}`,
    nickname: `用户${props.userId}`,
    status: 1
  }

  loading.value = false
}

const handleStartChat = () => {
  emit('close')
  emit('start-chat', props.userId)
}

const handleAddFriend = async () => {
  try {
    const response = await friendApi.sendRequest({
      to_user_id: props.userId,
      message: '你好，我想加你为好友'
    })
    if ((response as any).code === 0) {
      ElMessage.success('好友申请已发送')
      emit('close')
    } else {
      ElMessage.error((response as any).message || '发送申请失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '发送申请失败')
  }
}

onMounted(() => {
  loadUserProfile()
})
</script>

<style scoped>
.user-profile {
  display: flex;
  flex-direction: column;
}

.profile-content {
  padding: 16px;
}

.avatar-section {
  display: flex;
  justify-content: center;
  margin-bottom: 24px;
}

.info-section {
  margin-bottom: 24px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding: 10px 0;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  font-size: 14px;
  color: var(--text-secondary, #6b7280);
}

.info-value {
  font-size: 14px;
  color: var(--text-primary, #1f2937);
  font-weight: 500;
}

.action-section {
  padding-top: 16px;
  border-top: 1px solid var(--border-color, #e5e7eb);
}

.action-buttons {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.action-buttons .el-button {
  flex: 1;
  max-width: 200px;
}

.loading-state {
  padding: 20px;
}

.empty-state {
  padding: 40px 20px;
}
</style>
