<template>
  <div class="security-settings">
    <div class="settings-header">
      <h3>账号安全</h3>
      <p>修改密码以保护账号安全</p>
    </div>

    <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
      <el-form-item label="当前密码" prop="oldPassword">
        <el-input
          v-model="form.oldPassword"
          type="password"
          placeholder="请输入当前密码"
          show-password
        />
      </el-form-item>

      <el-form-item label="新密码" prop="newPassword">
        <el-input
          v-model="form.newPassword"
          type="password"
          placeholder="请输入新密码（至少6位）"
          show-password
        />
      </el-form-item>

      <el-form-item label="确认密码" prop="confirmPassword">
        <el-input
          v-model="form.confirmPassword"
          type="password"
          placeholder="请再次输入新密码"
          show-password
        />
      </el-form-item>

      <el-form-item>
        <el-button type="primary" @click="handleSubmit" :loading="loading">
          修改密码
        </el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useUserStore } from '../../store/user'
import { useChatStore } from '../../store/chat'
import { useWebSocket } from '../../utils/websocket'

const router = useRouter()
const userStore = useUserStore()
const chatStore = useChatStore()
const { disconnect } = useWebSocket('ws://localhost:3102/sub')
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const validateConfirmPassword = (rule: any, value: any, callback: any) => {
  if (value === '') {
    callback(new Error('请再次输入密码'))
  } else if (value !== form.newPassword) {
    callback(new Error('两次输入密码不一致'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  oldPassword: [
    { required: true, message: '请输入当前密码', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        await userStore.changePassword(form.oldPassword, form.newPassword)
        ElMessage.success('密码修改成功，请重新登录')
        formRef.value?.resetFields()
        
        // Force logout sequence
        disconnect()
        chatStore.clearAll()
        userStore.logout()
        
        // Use timeout to allow message to be seen
        setTimeout(() => {
            window.location.href = '/login'
        }, 1000)
        
      } catch (error: any) {
        ElMessage.error(error.message || '密码修改失败')
      } finally {
        loading.value = false
      }
    }
  })
}

const handleReset = () => {
  formRef.value?.resetFields()
}
</script>

<style scoped>
.security-settings {
  max-width: 500px;
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

/* Responsive */
@media (max-width: 768px) {
  .security-settings {
    max-width: 100%;
  }

  .settings-header h3 {
    font-size: 20px;
  }

  :deep(.el-form-item__label) {
    width: 100px !important;
  }
}
</style>
