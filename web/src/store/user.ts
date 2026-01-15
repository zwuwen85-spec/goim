import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, type User } from '../api/chat'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(sessionStorage.getItem('token') || '')
  const currentUser = ref<User | null>(null)

  const isLoggedIn = computed(() => !!token.value)

  const setToken = (newToken: string) => {
    token.value = newToken
    sessionStorage.setItem('token', newToken)
  }

  const setUser = (user: User) => {
    currentUser.value = user
    sessionStorage.setItem('user', JSON.stringify(user))
  }

  const loadUser = () => {
    const saved = sessionStorage.getItem('user')
    if (saved) {
      try {
        currentUser.value = JSON.parse(saved)
      } catch {
        sessionStorage.removeItem('user')
      }
    }
  }

  const login = async (username: string, password: string) => {
    // Clear all old account data before login
    const { useChatStore } = await import('./chat')
    const { useGroupStore } = await import('./group')
    const { useAIStore } = await import('./ai')

    const chatStore = useChatStore()
    const groupStore = useGroupStore()
    const aiStore = useAIStore()

    // Clear all stores
    chatStore.clearAll()
    groupStore.clearAll()
    aiStore.clearAll()

    // Clear theme and notification settings
    localStorage.removeItem('app_theme')
    localStorage.removeItem('notification_settings')

    const response = await authApi.login({ username, password }) as any
    if (response.code === 0) {
      setToken(response.data.token)
      const user: User = {
        id: response.data.user_id,
        username: response.data.username,
        nickname: response.data.nickname,
        avatar: response.data.avatar,
        status: response.data.status
      }
      setUser(user)
      return user
    }
    throw new Error(response.message)
  }

  const register = async (username: string, password: string, nickname: string) => {
    // Clear all old account data before register
    const { useChatStore } = await import('./chat')
    const { useGroupStore } = await import('./group')
    const { useAIStore } = await import('./ai')

    const chatStore = useChatStore()
    const groupStore = useGroupStore()
    const aiStore = useAIStore()

    // Clear all stores
    chatStore.clearAll()
    groupStore.clearAll()
    aiStore.clearAll()

    // Clear theme and notification settings
    localStorage.removeItem('app_theme')
    localStorage.removeItem('notification_settings')

    const response = await authApi.register({ username, password, nickname }) as any
    if (response.code === 0) {
      setToken(response.data.token)
      const user: User = {
        id: response.data.user_id,
        username: response.data.username,
        nickname: response.data.nickname,
        status: 1
      }
      setUser(user)
      return user
    }
    throw new Error(response.message)
  }

  const logout = () => {
    token.value = ''
    currentUser.value = null
    sessionStorage.removeItem('token')
    sessionStorage.removeItem('user')
  }

  const refreshProfile = async () => {
    const response = await authApi.getProfile() as any
    if (response.code === 0) {
      const user: User = {
        id: response.data.user_id,
        username: response.data.username,
        nickname: response.data.nickname,
        avatar: response.data.avatar,
        status: response.data.status
      }
      setUser(user)
      return user
    }
    return null
  }

  const updateProfile = async (data: { nickname?: string; signature?: string }) => {
    const response = await authApi.updateProfile(data) as any
    if (response.code === 0 && currentUser.value) {
      if (data.nickname) currentUser.value.nickname = data.nickname
      if (data.signature !== undefined) {
        (currentUser.value as any).signature = data.signature
      }
      sessionStorage.setItem('user', JSON.stringify(currentUser.value))
    }
    return response
  }

  const uploadAvatar = async (file: File) => {
    const response = await authApi.uploadAvatar(file) as any
    if (response.code === 0 && currentUser.value) {
      // 更新本地存储的用户信息
      currentUser.value.avatar = response.data.avatar_url
      sessionStorage.setItem('user', JSON.stringify(currentUser.value))
    }
    return response
  }

  const changePassword = async (oldPassword: string, newPassword: string) => {
    const response = await authApi.changePassword({
      old_password: oldPassword,
      new_password: newPassword
    }) as any
    return response
  }

  // Load user on store creation
  loadUser()

  return {
    token,
    currentUser,
    isLoggedIn,
    setToken,
    setUser,
    login,
    register,
    logout,
    refreshProfile,
    updateProfile,
    uploadAvatar,
    changePassword
  }
})
