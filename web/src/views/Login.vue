<template>
  <div class="login-wrapper">
    <div class="login-split">
      <!-- Left Side: Visual -->
      <div class="login-visual">
        <div class="visual-content">
          <h1>GoIM Chat</h1>
          <p>连接你我，沟通无限</p>
        </div>
        <img 
          src="https://trae-api-sg.mchost.guru/api/ide/v1/text_to_image?prompt=abstract%20modern%20technology%20blue%20gradient%203d%20shapes%20minimal&image_size=portrait_16_9" 
          alt="Login Visual" 
          class="visual-bg"
        />
        <div class="visual-overlay"></div>
      </div>

      <!-- Right Side: Form -->
      <div class="login-form-container">
        <div class="form-content">
          <div class="form-header">
            <h2>欢迎回来</h2>
            <p class="subtitle">请登录您的账号以继续</p>
          </div>

          <el-form :model="form" :rules="rules" ref="formRef" label-position="top" size="large">
            <el-form-item label="用户名" prop="username">
              <el-input
                v-model="form.username"
                placeholder="请输入用户名"
                :prefix-icon="User"
                @keyup.enter="handleLogin"
              />
            </el-form-item>

            <el-form-item label="密码" prop="password">
              <el-input
                v-model="form.password"
                type="password"
                placeholder="请输入密码"
                show-password
                :prefix-icon="Lock"
                @keyup.enter="handleLogin"
              />
            </el-form-item>

            <el-form-item>
              <el-button type="primary" @click="handleLogin" :loading="loading" class="submit-btn">
                登录
              </el-button>
            </el-form-item>

            <div class="form-footer">
              <span class="text-gray">还没有账号？</span>
              <router-link to="/register" class="link-primary">立即注册</router-link>
            </div>
          </el-form>

          <div class="demo-users">
            <p>快速体验（点击登录）：</p>
            <div class="demo-tags">
              <el-tag 
                v-for="user in ['alice', 'bob', 'charlie']" 
                :key="user"
                @click="fillAndLogin(user)" 
                class="cursor-pointer hover-effect"
                effect="plain"
              >
                {{ user }}
              </el-tag>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { useUserStore } from '../store/user'
import { useChatStore } from '../store/chat'

const router = useRouter()
const userStore = useUserStore()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      // Clear any existing chat state first
      const chatStore = useChatStore()
      chatStore.clearAll()

      await userStore.login(form.username, form.password)
      ElMessage.success('登录成功')

      // Wait a tick to ensure store is reset before loading new user data
      // Although userStore.login is async, extra safety
      
      // Explicitly load new user's sessions using the NEW user ID
      await chatStore.loadSessionsFromStorage()

      router.push('/')
    } catch (error: any) {
      ElMessage.error(error.message || '登录失败')
    } finally {
      loading.value = false
    }
  })
}

const fillAndLogin = (username: string) => {
  form.username = username
  form.password = '123456'
  handleLogin()
}
</script>

<style scoped>
.login-wrapper {
  min-height: 100vh;
  background-color: var(--bg-surface);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.login-split {
  display: flex;
  width: 100%;
  max-width: 1200px;
  height: 800px;
  max-height: 90vh;
  background: var(--bg-surface);
  border-radius: var(--border-radius-xl);
  box-shadow: var(--shadow-lg);
  overflow: hidden;
}

.login-visual {
  flex: 1;
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  color: white;
  padding: 40px;
  overflow: hidden;
}

.visual-bg {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  z-index: 0;
}

.visual-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.8) 0%, rgba(236, 72, 153, 0.8) 100%);
  z-index: 1;
}

.visual-content {
  position: relative;
  z-index: 2;
  text-align: center;
}

.visual-content h1 {
  font-size: 3.5rem;
  font-weight: 800;
  margin-bottom: 1rem;
  text-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.visual-content p {
  font-size: 1.5rem;
  opacity: 0.9;
}

.login-form-container {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  background: white;
}

.form-content {
  width: 100%;
  max-width: 400px;
}

.form-header {
  margin-bottom: 40px;
  text-align: center;
}

.form-header h2 {
  font-size: 2rem;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.subtitle {
  color: var(--text-secondary);
  font-size: 1rem;
}

.submit-btn {
  width: 100%;
  font-weight: 600;
  padding: 12px;
  height: auto;
  font-size: 1rem;
  border-radius: var(--border-radius-md);
  background: linear-gradient(to right, var(--primary-color), var(--primary-hover));
  border: none;
  transition: transform 0.2s;
}

.submit-btn:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

.form-footer {
  text-align: center;
  margin-top: 16px;
  font-size: 0.95rem;
}

.text-gray {
  color: var(--text-secondary);
}

.link-primary {
  color: var(--primary-color);
  text-decoration: none;
  font-weight: 600;
  margin-left: 4px;
}

.link-primary:hover {
  text-decoration: underline;
}

.demo-users {
  margin-top: 40px;
  padding-top: 24px;
  border-top: 1px solid var(--border-color);
  text-align: center;
}

.demo-users p {
  color: var(--text-light);
  font-size: 0.875rem;
  margin-bottom: 12px;
}

.demo-tags {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.hover-effect {
  transition: all 0.2s;
}

.hover-effect:hover {
  transform: scale(1.05);
  background-color: var(--primary-light);
  color: var(--primary-color);
  border-color: var(--primary-color);
}

@media (max-width: 768px) {
  .login-split {
    flex-direction: column;
    height: auto;
    max-height: none;
  }
  
  .login-visual {
    padding: 60px 20px;
    min-height: 200px;
  }

  .visual-content h1 {
    font-size: 2.5rem;
  }

  .visual-content p {
    font-size: 1.2rem;
  }
}
</style>
