<template>
  <div class="group-settings">
    <div class="settings-content">
      <!-- Group Avatar (Owner/Admin only) -->
      <div v-if="canManage" class="setting-section">
        <div class="section-title">群头像</div>
        <div class="avatar-section">
          <el-avatar :size="80" :src="displayAvatarUrl" class="group-avatar">
            <el-icon size="40"><ChatDotRound /></el-icon>
          </el-avatar>
          <el-upload
            :action="uploadUrl"
            :headers="uploadHeaders"
            :show-file-list="false"
            :before-upload="beforeUpload"
            :on-success="onAvatarSuccess"
            accept="image/*"
            name="avatar"
            class="avatar-upload"
          >
            <el-button size="small" :loading="uploading">
              {{ uploading ? '上传中...' : '更换头像' }}
            </el-button>
          </el-upload>
        </div>
      </div>

      <!-- Group Name (Owner/Admin only) -->
      <div v-if="canManage" class="setting-section">
        <div class="section-title">群名称</div>
        <el-input
          v-model="groupName"
          placeholder="请输入群名称"
          maxlength="50"
          show-word-limit
          :disabled="updatingName"
        >
          <template #append>
            <el-button
              :disabled="!groupName.trim() || groupName === originalGroupName"
              :loading="updatingName"
              @click="updateGroupName"
            >
              保存
            </el-button>
          </template>
        </el-input>
      </div>

      <!-- Member Nickname (All members) -->
      <div class="setting-section">
        <div class="section-title">我的群昵称</div>
        <el-input
          v-model="myNicknameValue"
          placeholder="请输入群昵称"
          maxlength="20"
          show-word-limit
          :disabled="updatingNickname"
        >
          <template #append>
            <el-button
              :disabled="!myNicknameValue.trim() || myNicknameValue === originalNickname"
              :loading="updatingNickname"
              @click="updateMyNickname"
            >
              保存
            </el-button>
          </template>
        </el-input>
        <div class="section-hint">设置在群内显示的名称，不影响全局昵称</div>
      </div>

      <!-- Group Info -->
      <div class="setting-section">
        <div class="section-title">群信息</div>
        <div class="info-item">
          <span class="info-label">群号:</span>
          <span class="info-value">{{ group?.group_no || '-' }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">群主:</span>
          <span class="info-value">{{ ownerName }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">成员数:</span>
          <span class="info-value">{{ group?.member_count || 0 }} 人</span>
        </div>
        <div class="info-item">
          <span class="info-label">我的角色:</span>
          <span class="info-value">{{ roleText }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ChatDotRound } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { groupApi, type Group } from '../api/chat'
import { useUserStore } from '../store/user'

const props = defineProps<{
  group: Group
  myRole?: number // 1=member, 2=admin, 3=owner
  myNickname?: string
}>()

const emit = defineEmits(['close', 'updated'])

const userStore = useUserStore()

const uploading = ref(false)
const updatingName = ref(false)
const updatingNickname = ref(false)

const groupName = ref(props.group.name || '')
const originalGroupName = ref(props.group.name || '')
const groupAvatar = ref(props.group.avatar || '')

// Computed property for displaying avatar with proper URL
const displayAvatarUrl = computed(() => {
  const avatar = groupAvatar.value
  if (!avatar) return undefined

  // If it's already a full URL, return as is
  if (avatar.startsWith('http://') || avatar.startsWith('https://')) {
    return avatar
  }

  // If it's a relative path starting with /uploads/, use origin + path
  // This will be proxied by vite to the backend
  if (avatar.startsWith('/uploads/')) {
    return window.location.origin + avatar
  }

  return avatar
})

// Handle sql.NullString format for myNickname prop
const initialNickname = props.myNickname && typeof props.myNickname === 'object' && 'String' in props.myNickname
  ? props.myNickname.String
  : (props.myNickname || '')
const myNicknameValue = ref(initialNickname || userStore.currentUser?.nickname || '')
const originalNickname = ref(initialNickname || userStore.currentUser?.nickname || '')

// Upload configuration
const uploadUrl = computed(() => {
  const baseUrl = import.meta.env.VITE_API_BASE_URL || ''
  return `${baseUrl}/api/group/avatar/${props.group.id}`
})
const uploadHeaders = computed(() => ({
  'Authorization': `Bearer ${userStore.token || sessionStorage.getItem('token') || ''}`
}))

// Permission check
const canManage = computed(() => {
  const role = props.myRole || 1
  return role === 3 || role === 2 // Owner or Admin
})

const ownerName = computed(() => {
  if (!props.group) return '-'
  return '群主'
})

const roleText = computed(() => {
  const role = props.myRole || 1
  switch (role) {
    case 3: return '群主'
    case 2: return '管理员'
    default: return '成员'
  }
})

// Watch for prop changes
watch(() => props.group, (newGroup) => {
  if (newGroup) {
    groupName.value = newGroup.name || ''
    originalGroupName.value = newGroup.name || ''
    groupAvatar.value = newGroup.avatar || ''
  }
}, { deep: true })

watch(() => props.myNickname, (newNickname) => {
  if (newNickname) {
    // Handle sql.NullString format
    const nicknameStr = newNickname && typeof newNickname === 'object' && 'String' in newNickname
      ? newNickname.String
      : newNickname
    myNicknameValue.value = nicknameStr || ''
    originalNickname.value = nicknameStr || ''
  }
})

// Avatar upload validation
const beforeUpload = (file: File) => {
  const isImage = file.type.startsWith('image/')
  if (!isImage) {
    ElMessage.error('只能上传图片文件')
    return false
  }
  const isLt10M = file.size / 1024 / 1024 < 10
  if (!isLt10M) {
    ElMessage.error('图片大小不能超过 10MB')
    return false
  }
  uploading.value = true
  return true
}

// Avatar upload success
const onAvatarSuccess = (response: any) => {
  uploading.value = false
  // Handle wrapped response format: {code: 0, data: {avatar_url: "..."}}
  const avatarUrl = response.data?.avatar_url || response.avatar_url
  if (avatarUrl) {
    // Add timestamp to prevent browser caching
    const urlWithTimestamp = avatarUrl + (avatarUrl.includes('?') ? '&' : '?') + 't=' + Date.now()
    groupAvatar.value = urlWithTimestamp
    ElMessage.success('头像上传成功')
    emit('updated', { avatar: urlWithTimestamp })
  } else {
    ElMessage.error('头像上传失败')
  }
}

// Update group name
const updateGroupName = async () => {
  if (!groupName.value.trim()) return
  if (groupName.value === originalGroupName.value) return

  updatingName.value = true
  try {
    const response = await groupApi.update(props.group.id, { name: groupName.value })
    if ((response as any).code === 0) {
      originalGroupName.value = groupName.value
      ElMessage.success('群名称更新成功')
      emit('updated', { name: groupName.value })
    } else {
      ElMessage.error((response as any).message || '更新失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '更新失败')
  } finally {
    updatingName.value = false
  }
}

// Update my nickname
const updateMyNickname = async () => {
  if (!myNicknameValue.value.trim()) return
  if (myNicknameValue.value === originalNickname.value) return

  updatingNickname.value = true
  try {
    const response = await groupApi.setMemberNickname(
      props.group.id,
      userStore.currentUser?.id || 0,
      myNicknameValue.value
    )
    if ((response as any).code === 0) {
      originalNickname.value = myNicknameValue.value
      ElMessage.success('群昵称更新成功')
      emit('updated', { myNickname: myNicknameValue.value })
    } else {
      ElMessage.error((response as any).message || '更新失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '更新失败')
  } finally {
    updatingNickname.value = false
  }
}
</script>

<style scoped>
.group-settings {
  display: flex;
  flex-direction: column;
}

.settings-content {
  padding: 16px;
}

.setting-section {
  margin-bottom: 20px;
}

.setting-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 10px;
}

.section-hint {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 6px;
}

.avatar-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.group-avatar {
  flex-shrink: 0;
}

.avatar-upload {
  flex-shrink: 0;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  border-bottom: 1px solid var(--border-color);
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  font-size: 14px;
  color: var(--text-secondary);
}

.info-value {
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 500;
}
</style>
