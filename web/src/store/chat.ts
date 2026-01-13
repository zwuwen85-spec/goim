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
  isHistoryLoaded?: boolean
  isHistoryAllLoaded?: boolean
}

export const useChatStore = defineStore('chat', () => {
  const conversations = ref<Conversation[]>([])
  const friends = ref<Friend[]>([])
  const friendRequests = ref<FriendRequest[]>([])
  const currentSession = ref<ChatSession | null>(null)
  const sessions = ref<Map<string, ChatSession>>(new Map())

  // Persistence
  const saveSessionsToStorage = () => {
    try {
      const sessionsArray = Array.from(sessions.value.entries())
      localStorage.setItem('chat_sessions', JSON.stringify(sessionsArray))
    } catch (e) {
      console.error('Failed to save sessions to storage', e)
    }
  }

  const loadSessionsFromStorage = () => {
    try {
      const stored = localStorage.getItem('chat_sessions')
      if (stored) {
        const sessionsArray = JSON.parse(stored)
        sessions.value = new Map(sessionsArray)
      }
    } catch (e) {
      console.error('Failed to load sessions from storage', e)
    }
  }

  // Initialize
  loadSessionsFromStorage()

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
    // For AI chat (type 3), conversationId is usually just the botId (targetId)
    // No special calculation needed as it is like a group chat (id is unique)

    const sessionKey = `${conversationId}:${conversationType}`
    let session = sessions.value.get(sessionKey)

    if (!session || !session.isHistoryLoaded || session.messages.length === 0) {
      // Load history messages
      try {
        const historyResponse = await messageApi.getHistory({
          conversation_id: conversationId,
          conversation_type: conversationType,
          limit: 50
        })

        const historyMessages = (historyResponse as any)?.data?.messages || []

        if (!session) {
          session = {
            id: sessionKey,
            targetId: targetId,
            targetType: conversationType === 1 ? 'user' : (conversationType === 2 ? 'group' : 'ai'),
            name: name || (conversationType === 3 ? 'AI Assistant' : `Chat ${targetId}`),
            avatar,
            unreadCount: 0,
            messages: historyMessages,
            isHistoryLoaded: true,
            isHistoryAllLoaded: historyMessages.length < 50
          }
          sessions.value.set(sessionKey, session)
        } else {
          // Merge history with existing messages
          const existingIds = new Set(session.messages.map(m => m.msg_id || m.id))
          const newMsgs = historyMessages.filter((m: Message) => !existingIds.has(m.msg_id || m.id))
          
          session.messages = [...newMsgs, ...session.messages].sort((a, b) => a.seq - b.seq)
          session.isHistoryLoaded = true
          if (historyMessages.length < 50) {
             session.isHistoryAllLoaded = true
          }
        }
        saveSessionsToStorage()
      } catch (error) {
        console.error('Failed to load history:', error)
        if (!session) {
          // Create empty session on error
          session = {
            id: sessionKey,
            targetId: targetId,
            targetType: conversationType === 1 ? 'user' : (conversationType === 2 ? 'group' : 'ai'),
            name: name || (conversationType === 3 ? 'AI Assistant' : `Chat ${targetId}`),
            avatar,
            unreadCount: 0,
            messages: [],
            isHistoryLoaded: true // Avoid retrying immediately on error
          }
          sessions.value.set(sessionKey, session)
          saveSessionsToStorage()
        }
      }
    }

    currentSession.value = sessions.value.get(sessionKey)!

    // Ensure conversation exists in the list (for sidebar)
    const existingConv = conversations.value.find(
      c => c.target_id === targetId && c.conversation_type === conversationType
    )
    if (!existingConv) {
       const userStore = useUserStore()
       const newConv: Conversation = {
         id: conversationId,
         user_id: userStore.currentUser?.id || 0,
         target_id: targetId,
         conversation_type: conversationType,
         unread_count: 0,
         last_msg_content: '',
         last_msg_time: new Date().toISOString(),
         is_pinned: 0,
         is_muted: 0,
         target_user: {
            id: targetId,
            username: name || 'User',
            nickname: name || 'User',
            avatar: avatar || '',
            status: 1
         }
       }
       conversations.value.unshift(newConv)
    }

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
      saveSessionsToStorage()
      
      // Update conversation list
      updateConversation(
        currentSession.value.targetId,
        currentSession.value.targetType === 'user' ? 1 : (currentSession.value.targetType === 'group' ? 2 : 3),
        content,
        (response as any).data?.created_at || new Date().toISOString()
      )

      return true
    }
    return false
  }

  const addMessage = (msg: Message) => {
    const sessionKey = `${msg.conversation_id}:${msg.conversation_type}`
    let session = sessions.value.get(sessionKey)
    
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
            targetId = id1 === myId ? id2 : id1 
        }
    }

    if (!session) {
      session = {
        id: sessionKey,
        targetId: targetId,
        targetType: msg.conversation_type === 1 ? 'user' : (msg.conversation_type === 2 ? 'group' : 'ai'),
        name: `Chat ${targetId}`, // Name might be updated later
        unreadCount: 0,
        messages: []
      }
      sessions.value.set(sessionKey, session)
    }

    session.messages.push(msg)
    saveSessionsToStorage()

    // Update conversation list
    // If it's a new message from someone else, we should update the list
    updateConversation(
        targetId,
        msg.conversation_type,
        msg.content,
        msg.created_at
    )
    
    // If we really need to sync from server:
    // loadConversations()
  }

  const loadMoreMessages = async () => {
    if (!currentSession.value) return 0
    if (currentSession.value.isHistoryAllLoaded) return 0
    
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
        
        if (newMessages.length < 20) {
           currentSession.value.isHistoryAllLoaded = true
        }

        if (newMessages.length > 0) {
          // Merge and sort
          // Filter out duplicates just in case
          const existingIds = new Set(currentSession.value.messages.map(m => m.msg_id || m.id))
          const uniqueNewMessages = newMessages.filter((m: Message) => !existingIds.has(m.msg_id || m.id))
          
          if (uniqueNewMessages.length > 0) {
            const combined = [...uniqueNewMessages, ...currentSession.value.messages]
            combined.sort((a, b) => a.seq - b.seq)
            currentSession.value.messages = combined
            saveSessionsToStorage()
            return uniqueNewMessages.length
          }
        } else {
           // No new messages found
           currentSession.value.isHistoryAllLoaded = true
        }
      }
    } catch (error) {
      console.error('Failed to load more messages:', error)
    }
    return 0
  }

  const updateConversation = (targetId: number, type: number, content: string, time: string) => {
    // For AI chat, we might need to handle targetId carefully
    // In chat store, conversation_id for AI is just the bot_id (targetId)
    // But in conversations list, it might be stored with a specific ID
    
    let conv = conversations.value.find(c => c.target_id === targetId && c.conversation_type === type)
    
    if (!conv && type === 3) {
        // If not found and it's AI, try to find it again or create it
        // Sometimes type 3 might be missing if it wasn't loaded from server list yet
    }

    if (conv) {
      conv.last_msg_content = content
      conv.last_msg_time = time
      conv.unread_count = 0 // Reset unread count if we are updating it (usually means we saw it or sent it)
      
      // Move to top
      const index = conversations.value.indexOf(conv)
      if (index > 0) {
        conversations.value.splice(index, 1)
        conversations.value.unshift(conv)
      }
    } else {
        // If conversation doesn't exist in list, create it!
        // This is crucial for new chats (especially AI) to show up with content
        const userStore = useUserStore()
        
        // Try to get name/avatar
        let name = `Chat ${targetId}`
        let avatar = ''
        
        // If AI, try to fetch bot info from AI store (if available in context, but we are in chat store)
        // We can use a default name for now, usually openChat handles this better, but this is for background updates
        if (type === 3) {
             name = `AI ${targetId}` 
             // We can try to import useAIStore but circular dependency might occur. 
             // Ideally openChat should have created this.
        }

        const newConv: Conversation = {
             id: targetId, // Simplified ID for AI/Group
             user_id: userStore.currentUser?.id || 0,
             target_id: targetId,
             conversation_type: type,
             unread_count: 0,
             last_msg_content: content,
             last_msg_time: time,
             is_pinned: 0,
             is_muted: 0,
             target_user: {
                id: targetId,
                username: name,
                nickname: name,
                avatar: avatar,
                status: 1
             }
        }
        conversations.value.unshift(newConv)
    }
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
    updateConversation,
    getSessionKey,
  }
})
