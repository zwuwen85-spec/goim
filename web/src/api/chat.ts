import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json'
  }
})

export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = sessionStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor to handle errors
api.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    if (error.response?.status === 401) {
      sessionStorage.removeItem('token')
      sessionStorage.removeItem('user')
      
      // Force reload if not on login page
      if (!window.location.pathname.includes('/login')) {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error.response?.data || error.message)
  }
)

// Types
export interface User {
  id: number
  username: string
  nickname: string
  avatar?: string
  status: number
}

export interface Message {
  id: number
  msg_id: string
  from_user_id: number
  conversation_id: number
  conversation_type: number
  msg_type: number
  content: string
  seq: number
  created_at: string
  from_user?: User
}

export interface Conversation {
  id: number
  user_id: number
  target_id: number
  conversation_type: number
  unread_count: number
  last_msg_content: string
  last_msg_time: string
  is_pinned: number
  is_muted: number
  target_user?: User
}

export interface Friend {
  id: number
  user_id: number
  friend_id: number
  remark?: string
  group_name: string
  friend_user?: User
}

export interface FriendRequest {
  id: number
  from_user_id: number
  to_user_id: number
  message?: string
  status: number
  created_at: string
  from_user?: User
}

// Group Types
export interface Group {
  id: number
  group_no?: string
  name: string
  owner_id: number
  max_members: number
  member_count: number
  avatar?: string
  description?: string
  join_type?: number
  mute_all?: number
  created_at: string
}

export interface GroupMember {
  id: number
  group_id: number
  user_id: number
  role: number // 1: member, 2: admin, 3: owner
  nickname?: string
  joined_at: string
  mute_until?: string
  user?: User
}

export interface GroupJoinRequest {
  id: number
  group_id: number
  user_id: number
  message?: string
  status: number
  created_at: string
  user?: User
}

// AI Types
export interface AIBot {
  id: number
  name: string
  personality: string
  role?: string
  tone?: string
  traits?: string[]
  is_default: boolean
  model?: string
}

export interface AIMessage {
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp?: number
  seq?: number
  streaming?: boolean
  error?: boolean
  msgId?: string
}

// Auth API
export const authApi = {
  register: (data: { username: string; password: string; nickname: string }) =>
    api.post('/user/register', data),

  login: (data: { username: string; password: string }) =>
    api.post('/user/login', data),

  getProfile: () => api.get('/user/profile'),

  updateProfile: (data: { nickname?: string; signature?: string }) =>
    api.put('/user/profile', data),

  uploadAvatar: (file: File) => {
    const formData = new FormData()
    formData.append('avatar', file)
    return api.post('/user/avatar', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  changePassword: (data: { old_password: string; new_password: string }) =>
    api.put('/user/password', data),

  searchUsers: (keyword: string) => api.get('/user/search', { params: { keyword } })
}

// Friend API
export const friendApi = {
  sendRequest: (data: { to_user_id: number; message?: string }) =>
    api.post('/friend/request', data),

  acceptRequest: (id: number) => api.post(`/friend/accept/${id}`),

  rejectRequest: (id: number) => api.post(`/friend/reject/${id}`),

  updateRemark: (data: { friend_id: number; remark: string; group_name: string }) =>
    api.put('/friend/remark', data),

  deleteFriend: (id: number) => api.delete(`/friend/delete/${id}`),

  getFriends: () => api.get('/friend/list'),

  getRequests: () => api.get('/friend/requests')
}

// Message API
export const messageApi = {
  send: (data: {
    to_user_id?: number
    to_group_id?: number
    conversation_type: number
    msg_type: number
    content: string
  }) => api.post('/message/send', data),

  getHistory: (params: {
    conversation_id: number
    conversation_type: number
    last_seq?: number
    limit?: number
  }) => api.get('/message/history', { params }),

  markRead: (data: {
    conversation_id: number
    conversation_type: number
    msg_id: number
  }) => api.post('/message/read', data)
}

// Conversation API
export const conversationApi = {
  getList: () => api.get('/conversation/list')
}

// Group API
export const groupApi = {
  getList: () => api.get('/group/list'),

  create: (data: { name: string; max_members?: number }) =>
    api.post('/group/create', data),

  getInfo: (id: number) => api.get(`/group/info/${id}`),

  getMembers: (id: number) => api.get(`/group/members/${id}`),

  join: (id: number, message?: string) =>
    api.post(`/group/join/${id}`, { message }),

  leave: (id: number) => api.delete(`/group/leave/${id}`),

  update: (id: number, data: { name?: string }) =>
    api.put(`/group/info/${id}`, data),

  uploadAvatar: (id: number, file: File) => {
    const formData = new FormData()
    formData.append('avatar', file)
    return api.post(`/group/avatar/${id}`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  // New: invite, kick, manage members
  invite: (id: number, data: { user_id: number; role?: number }) =>
    api.post(`/group/invite/${id}`, data),

  kick: (id: number, userId: number) =>
    api.delete(`/group/kick/${id}/${userId}`),

  setMemberRole: (id: number, userId: number, role: number) =>
    api.put(`/group/role/${id}/${userId}`, { role }),

  setMemberNickname: (id: number, userId: number, nickname: string) =>
    api.put(`/group/nickname/${id}/${userId}`, { nickname }),

  muteMember: (id: number, userId: number, duration: number) =>
    api.put(`/group/mute/${id}/${userId}`, { duration }),

  transferOwnership: (id: number, userId: number) =>
    api.post(`/group/transfer/${id}/${userId}`),

  deleteGroup: (id: number) =>
    api.delete(`/group/${id}`)
}

// AI API
export const aiApi = {
  getBots: () => api.get('/ai/bots'),

  getBot: (id: number) => api.get(`/ai/bot/${id}`),

  createBot: (data: {
    name: string
    personality: string
    model?: string
    temperature?: number
  }) => api.post('/ai/bot/create', data),

  updateBot: (id: number, data: {
    name?: string
    personality?: string
    temperature?: number
  }) => api.put(`/ai/bot/${id}`, data),

  deleteBot: (id: number) => api.delete(`/ai/bot/${id}`),

  sendMessage: (data: {
    bot_id: number
    message: string
  }) => api.post('/ai/chat/send', data),

  // Stream message API
  streamMessage: async (data: {
    bot_id: number
    message: string
  }, onChunk: (chunk: string) => void, onDone: (fullContent: string, msgId: string) => void, onError: (error: string) => void): Promise<void> => {
    const token = sessionStorage.getItem('token')

    try {
      const response = await fetch('/api/ai/chat/stream', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify(data)
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const reader = response.body?.getReader()
      if (!reader) {
        throw new Error('No response body')
      }

      const decoder = new TextDecoder()
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()

        if (done) break

        buffer += decoder.decode(value, { stream: true })

        // Process SSE messages
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        for (const line of lines) {
          if (line.startsWith('data:')) {
            try {
              const data = JSON.parse(line.slice(5).trim())

              if (data.type === 'chunk') {
                onChunk(data.delta)
              } else if (data.type === 'done') {
                onDone(data.content, data.msg_id)
              } else if (data.type === 'error') {
                onError(data.message || 'Unknown error')
              }
            } catch (e) {
              console.error('Failed to parse SSE data:', line, e)
            }
          }
        }
      }
    } catch (error: any) {
      onError(error.message || 'Stream failed')
      throw error
    }
  },

  getConversations: () => api.get('/ai/chat')
}

export default api
