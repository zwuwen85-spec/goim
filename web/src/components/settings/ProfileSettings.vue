<template>
  <div class="profile-settings">
    <div class="settings-header">
      <h3>个人资料</h3>
      <p>管理您的个人信息</p>
    </div>

    <div class="settings-body">
      <!-- Avatar Upload -->
      <div class="setting-section">
        <div class="section-title">头像</div>
        <div class="avatar-upload">
          <el-avatar :size="80" :src="avatarUrl" :key="avatarUrl">
            {{ userStore.currentUser?.nickname?.[0] }}
          </el-avatar>
          <el-upload
            :show-file-list="false"
            :auto-upload="false"
            :on-change="handleAvatarSelect"
            accept="image/*"
          >
            <el-button size="small">更换头像</el-button>
          </el-upload>
        </div>
      </div>

      <!-- Nickname -->
      <div class="setting-section">
        <div class="section-title">昵称</div>
        <el-input
          v-model="form.nickname"
          placeholder="请输入昵称"
          maxlength="20"
          show-word-limit
        />
      </div>

      <!-- Signature -->
      <div class="setting-section">
        <div class="section-title">个性签名</div>
        <el-input
          v-model="form.signature"
          type="textarea"
          placeholder="介绍一下自己..."
          maxlength="200"
          show-word-limit
          :rows="3"
        />
      </div>

      <!-- Save Button -->
      <div class="setting-actions">
        <el-button type="primary" @click="handleSave" :loading="saving">
          保存修改
        </el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import type { UploadFile } from 'element-plus'
import { useUserStore } from '../../store/user'

const userStore = useUserStore()
const saving = ref(false)

const form = reactive({
  nickname: '',
  signature: ''
})

const avatarUrl = computed(() => {
  const avatar = userStore.currentUser?.avatar
  if (!avatar) return undefined

  // 如果是完整URL直接返回
  if (avatar.startsWith('http://') || avatar.startsWith('https://')) {
    return avatar
  }

  // 如果是相对路径，通过 API 代理访问
  // 开发环境：前端代理 /api -> 后端 3112 端口
  // 生产环境：需要配置 nginx 或后端直接提供静态文件
  if (avatar.startsWith('/uploads/')) {
    // 开发环境使用代理路径
    return window.location.origin + avatar
  }

  return avatar
})

onMounted(() => {
  form.nickname = userStore.currentUser?.nickname || ''
  form.signature = (userStore.currentUser as any)?.signature || ''
})

const handleAvatarSelect = async (uploadFile: UploadFile) => {
  const file = uploadFile.raw
  if (!file) return

  const isImage = file.type.startsWith('image/')
  const isLt2M = file.size / 1024 / 1024 < 2

  if (!isImage) {
    ElMessage.error('只能上传图片文件!')
    return
  }
  if (!isLt2M) {
    ElMessage.error('图片大小不能超过 2MB!')
    return
  }

  console.log('开始上传头像...', file.name)
  try {
    const response = await userStore.uploadAvatar(file)
    console.log('上传响应:', response)
    // 刷新用户信息
    await userStore.refreshProfile()
    ElMessage.success('头像上传成功')
  } catch (error: any) {
    console.error('上传失败:', error)
    ElMessage.error(error.message || '头像上传失败')
  }
}

const handleSave = async () => {
  if (!form.nickname.trim()) {
    ElMessage.warning('昵称不能为空')
    return
  }

  saving.value = true
  try {
    await userStore.updateProfile({
      nickname: form.nickname,
      signature: form.signature
    })
    ElMessage.success('保存成功')
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  } finally {
    saving.value = false
  }
}

const handleReset = () => {
  form.nickname = userStore.currentUser?.nickname || ''
  form.signature = (userStore.currentUser as any)?.signature || ''
}
</script>

<style scoped>
.profile-settings {
  max-width: 600px;
  width: 100%;
}

.settings-header {
  margin-bottom: 32px;
}

.settings-header h3 {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 8px 0;
}

.settings-header p {
  color: var(--text-secondary);
  margin: 0;
}

.settings-body {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.setting-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

.avatar-upload {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}

.setting-actions {
  display: flex;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
  flex-wrap: wrap;
}

/* Responsive */
@media (max-width: 768px) {
  .settings-header h3 {
    font-size: 20px;
  }

  .settings-body {
    gap: 24px;
  }

  .avatar-upload {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
}
</style>
