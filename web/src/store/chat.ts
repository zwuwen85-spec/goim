import { defineStore } from 'pinia'
import { ref } from 'vue'
import { conversationApi, friendApi, messageApi, type Conversation, type Friend, type Message, type User } from '../api/chat'
import { useUserStore } from './user'

export interface ChatSession {
  id: string // conversation_id:conversation_type
  targetId: number
  targetType: 'user' | 'group' | 'ai'
  name: string
  avatar?: string
  unreadCount: number
  messages: Message[]
}

export const useChatStore = defineStore('chat', () => {
  const conversations = ref<Conversation[]>([])
  const friends = ref<Friend[]>([])
  const currentSession = ref<ChatSession | null>(null)
  const sessions = ref<Map<string, ChatSession>>(new Map())
  // 用户信息映射：userId -> User
  const usersMap = ref<Map<number, User>>(new Map())

  const loadConversations = async () => {
    const response = await conversationApi.getList()
    if (response.code === 0) {
      conversations.value = response.data.conversations || []
    }
  }

  const loadFriends = async () => {
    const response = await friendApi.getFriends()
    if (response.code === 0) {
      friends.value = response.data.friends || []
    }
  }

  const openChat = async (conversationId: number, conversationType: number, name?: string, avatar?: string) => {
    const sessionKey = `${conversationId}:${conversationType}`

    if (!sessions.value.has(sessionKey)) {
      // Load history messages
      const historyResponse = await messageApi.getHistory({
        conversation_id: conversationId,
        conversation_type: conversationType,
        limit: 50
      })

      const session: ChatSession = {
        id: sessionKey,
        targetId: conversationId,
        targetType: conversationType === 1 ? 'user' : 'group',
        name: name || `Chat ${conversationId}`,
        avatar,
        unreadCount: 0,
        messages: historyResponse.data?.messages || []
      }
      sessions.value.set(sessionKey, session)
    }

    currentSession.value = sessions.value.get(sessionKey)!

    // Clear unread count
    const conv = conversations.value.find(
      c => c.target_id === conversationId && c.conversation_type === conversationType
    )
    if (conv && conv.unread_count > 0) {
      // Find last message to mark as read
      const session = sessions.value.get(sessionKey)
      if (session && session.messages.length > 0) {
        const lastMsg = session.messages[session.messages.length - 1]
        await messageApi.markRead({
          conversation_id: conversationId,
          conversation_type: conversationType,
          msg_id: lastMsg.id
        })
        conv.unread_count = 0
      }
    }
  }

  const sendMessage = async (content: string, msgType: number = 1) => {
    if (!currentSession.value) return false

    const userStore = useUserStore()
    const currentUserId = userStore.currentUser?.id || 0

    const response = await messageApi.send({
      to_user_id: currentSession.value.targetType === 'user' ? currentSession.value.targetId : undefined,
      to_group_id: currentSession.value.targetType === 'group' ? currentSession.value.targetId : undefined,
      conversation_type: currentSession.value.targetType === 'user' ? 1 : 2,
      msg_type: msgType,
      content
    })

    if (response.code === 0) {
      // Add message to current session
      const msg: Message = {
        id: Date.now(),
        msg_id: response.data.msg_id,
        from_user_id: currentUserId,
        conversation_id: response.data.conversation_id,
        conversation_type: response.data.conversation_type,
        msg_type: msgType,
        content,
        seq: response.data.seq,
        created_at: response.data.created_at
      }
      currentSession.value.messages.push(msg)
      return true
    }
    return false
  }

  const addMessage = (msg: Message) => {
    const sessionKey = `${msg.conversation_id}:${msg.conversation_type}`
    let session = sessions.value.get(sessionKey)

    if (!session) {
      session = {
        id: sessionKey,
        targetId: msg.conversation_id,
        targetType: msg.conversation_type === 1 ? 'user' : 'group',
        name: `Chat ${msg.conversation_id}`,
        messages: []
      }
      sessions.value.set(sessionKey, session)
    }

    session.messages.push(msg)

    // Update conversation list
    loadConversations()
  }

  const getSessionKey = (userId: number) => {
    // Calculate conversation pair ID for single chat
    const userStore = useUserStore()
    const myId = userStore.currentUser?.id || 0
    const id1 = Math.min(myId, userId)
    const id2 = Math.max(myId, userId)
    return `${id1 * 1000000000 + id2}:1`
  }

  return {
    conversations,
    friends,
    currentSession,
    sessions,
    loadConversations,
    loadFriends,
    openChat,
    sendMessage,
    addMessage,
    getSessionKey
  }
})
