import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, type User } from '../api/chat'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const currentUser = ref<User | null>(null)

  const isLoggedIn = computed(() => !!token.value)

  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  const setUser = (user: User) => {
    currentUser.value = user
    localStorage.setItem('user', JSON.stringify(user))
  }

  const loadUser = () => {
    const saved = localStorage.getItem('user')
    if (saved) {
      try {
        currentUser.value = JSON.parse(saved)
      } catch {
        localStorage.removeItem('user')
      }
    }
  }

  const login = async (username: string, password: string) => {
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
    localStorage.removeItem('token')
    localStorage.removeItem('user')
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
    refreshProfile
  }
})
