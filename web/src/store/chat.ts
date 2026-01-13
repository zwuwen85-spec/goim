import { defineStore } from 'pinia'
import { ref } from 'vue'
import { conversationApi, friendApi, messageApi, type Conversation, type Friend, type Message, type FriendRequest } from '../api/chat'
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
  const friendRequests = ref<FriendRequest[]>([])
  const currentSession = ref<ChatSession | null>(null)
  const sessions = ref<Map<string, ChatSession>>(new Map())

  const loadConversations = async () => {
    const response = await conversationApi.getList()
    if ((response as any).code === 0) {
      conversations.value = (response as any).data?.conversations || []
    }
  }

  const loadFriends = async () => {
    const response = await friendApi.getFriends()
    if ((response as any).code === 0) {
      friends.value = (response as any).data?.friends || []
    }
  }

  const loadFriendRequests = async () => {
    try {
      const response = await friendApi.getRequests()
      if ((response as any).code === 0) {
        // API returns pending requests (status === 1)
        friendRequests.value = (response as any).data.requests || []
      }
    } catch (error) {
      console.error('Failed to load friend requests:', error)
    }
  }

  const openChat = async (targetId: number, conversationType: number, name?: string, avatar?: string) => {
    let conversationId = targetId
    
    // For single chat, calculate the conversation pair ID
    if (conversationType === 1) {
      const userStore = useUserStore()
      const myId = userStore.currentUser?.id || 0
      const id1 = Math.min(myId, targetId)
      const id2 = Math.max(myId, targetId)
      conversationId = id1 * 1000000000 + id2
    }

    const sessionKey = `${conversationId}:${conversationType}`

    if (!sessions.value.has(sessionKey)) {
      // Load history messages
      try {
        const historyResponse = await messageApi.getHistory({
          conversation_id: conversationId,
          conversation_type: conversationType,
          limit: 50
        })

        const session: ChatSession = {
          id: sessionKey,
          targetId: targetId,
          targetType: conversationType === 1 ? 'user' : 'group',
          name: name || `Chat ${targetId}`,
          avatar,
          unreadCount: 0,
          messages: (historyResponse as any)?.data?.messages || []
        }
        sessions.value.set(sessionKey, session)
      } catch (error) {
        console.error('Failed to load history:', error)
        // Create empty session on error
        const session: ChatSession = {
          id: sessionKey,
          targetId: targetId,
          targetType: conversationType === 1 ? 'user' : 'group',
          name: name || `Chat ${targetId}`,
          avatar,
          unreadCount: 0,
          messages: []
        }
        sessions.value.set(sessionKey, session)
      }
    }

    currentSession.value = sessions.value.get(sessionKey)!

    // Clear unread count
    const conv = conversations.value.find(
      c => c.target_id === targetId && c.conversation_type === conversationType
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

    if ((response as any).code === 0) {
      // Add message to current session
      const msg: Message = {
        id: Date.now(),
        msg_id: (response as any).data?.msg_id,
        from_user_id: currentUserId,
        conversation_id: (response as any).data?.conversation_id,
        conversation_type: (response as any).data?.conversation_type,
        msg_type: msgType,
        content,
        seq: (response as any).data?.seq,
        created_at: (response as any).data?.created_at
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
      let targetId = msg.conversation_id
      if (msg.conversation_type === 1) {
        const userStore = useUserStore()
        const myId = userStore.currentUser?.id || 0
        // Reverse Pair ID: PairID = id1 * 1e9 + id2
        const id1 = Math.floor(msg.conversation_id / 1000000000)
        const id2 = msg.conversation_id % 1000000000
        
        if (id1 === myId) targetId = id2
        else if (id2 === myId) targetId = id1
        else {
            // Fallback, shouldn't happen if I am part of the conversation
            targetId = id1 === myId ? id2 : id1 
        }
      }

      session = {
        id: sessionKey,
        targetId: targetId,
        targetType: msg.conversation_type === 1 ? 'user' : 'group',
        name: `Chat ${targetId}`,
        unreadCount: 0,
        messages: []
      }
      sessions.value.set(sessionKey, session)
    }

    session.messages.push(msg)

    // Update conversation list
    loadConversations()
  }

  const loadMoreMessages = async () => {
    if (!currentSession.value) return 0
    
    const messages = currentSession.value.messages
    if (messages.length === 0) return 0
    
    // Find the oldest seq (smallest seq)
    const oldestSeq = messages.reduce((min, msg) => (msg.seq < min ? msg.seq : min), messages[0].seq)
    
    const [convIdStr, typeStr] = currentSession.value.id.split(':')
    const conversationId = parseInt(convIdStr)
    const conversationType = parseInt(typeStr)

    try {
      const response = await messageApi.getHistory({
        conversation_id: conversationId,
        conversation_type: conversationType,
        limit: 20, // Load 20 at a time
        last_seq: oldestSeq
      })

      if ((response as any).code === 0) {
        const newMessages = (response as any).data?.messages || []
        if (newMessages.length > 0) {
          // Merge and sort
          // Filter out duplicates just in case
          const existingIds = new Set(currentSession.value.messages.map(m => m.msg_id || m.id))
          const uniqueNewMessages = newMessages.filter((m: Message) => !existingIds.has(m.msg_id || m.id))
          
          if (uniqueNewMessages.length > 0) {
            const combined = [...uniqueNewMessages, ...currentSession.value.messages]
            combined.sort((a, b) => a.seq - b.seq)
            currentSession.value.messages = combined
            return uniqueNewMessages.length
          }
        }
      }
    } catch (error) {
      console.error('Failed to load more messages:', error)
    }
    return 0
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
    friendRequests,
    currentSession,
    sessions,
    loadConversations,
    loadFriends,
    loadFriendRequests,
    openChat,
    sendMessage,
    addMessage,
    loadMoreMessages,
    getSessionKey,
  }
})
