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
  const getStorageKey = (key: string) => {
    const userStore = useUserStore()
    const userId = userStore.currentUser?.id
    if (!userId) return null
    return `${key}_${userId}`
  }

  const saveSessionsToStorage = () => {
    try {
      const keySessions = getStorageKey('chat_sessions')
      const keyConversations = getStorageKey('chat_conversations')
      
      if (keySessions && keyConversations) {
        const sessionsArray = Array.from(sessions.value.entries())
        localStorage.setItem(keySessions, JSON.stringify(sessionsArray))
        localStorage.setItem(keyConversations, JSON.stringify(conversations.value))
      }
    } catch (e) {
      console.error('Failed to save sessions to storage', e)
    }
  }

  const loadSessionsFromStorage = () => {
    try {
      const keySessions = getStorageKey('chat_sessions')
      const keyConversations = getStorageKey('chat_conversations')
      
      if (keySessions) {
        const storedSessions = localStorage.getItem(keySessions)
        if (storedSessions) {
          const sessionsArray = JSON.parse(storedSessions)
          sessions.value = new Map(sessionsArray)
        } else {
            sessions.value = new Map()
        }
      }

      if (keyConversations) {
        const storedConversations = localStorage.getItem(keyConversations)
        if (storedConversations) {
          conversations.value = JSON.parse(storedConversations)
        } else {
            conversations.value = []
        }
      }
    } catch (e) {
      console.error('Failed to load sessions from storage', e)
    }
  }

  // Clear all data (for account switching)
  // Modified to NOT clear storage, just memory state
  const clearAll = () => {
    conversations.value = []
    friends.value = []
    friendRequests.value = []
    currentSession.value = null
    sessions.value = new Map() // Properly reset Map
    // We do NOT remove from localStorage here anymore, 
    // because we want to persist data per user.
    // Each user has their own key now.
  }

  // Initialize
  // Do NOT load automatically on init, wait for explicit call after login
  // loadSessionsFromStorage()

  const loadConversations = async () => {
    try {
      const response = await conversationApi.getList()
      if ((response as any).code === 0) {
        const rawList = (response as any).data?.conversations || []
        // Normalize types
        const newList = rawList.map((c: any) => ({
          ...c,
          target_id: Number(c.target_id),
          conversation_type: Number(c.conversation_type),
          unread_count: Number(c.unread_count)
        }))
        
        // Merge with local conversations to preserve last message info if server doesn't provide it fully
        // Or simply overwrite if server is source of truth. Usually server list has last_msg info.
        conversations.value = newList
        saveSessionsToStorage()
      }
    } catch (e) {
        console.error('Failed to load conversations from server', e)
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
    
    const userStore = useUserStore()
    const myId = userStore.currentUser?.id || 0

    // For single chat, calculate the conversation pair ID
    if (conversationType === 1) {
      // If targetId is myself, conversationId is just the user pair ID where both are myId
      // Or we can treat it as a special case.
      // Current logic: id1 * 1e9 + id2.
      // For self chat: myId * 1e9 + myId.
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
          // Update name and avatar if provided
          if (name) session.name = name
          if (avatar !== undefined) session.avatar = avatar
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

    // For self-chat, targetId is currentUserId
    const toUserId = currentSession.value.targetType === 'user' ? currentSession.value.targetId : undefined
    
    console.log('Sending message to', toUserId, 'type', currentSession.value.targetType)

    const response = await messageApi.send({
      to_user_id: toUserId,
      to_group_id: currentSession.value.targetType === 'group' ? currentSession.value.targetId : undefined,
      conversation_type: currentSession.value.targetType === 'user' ? 1 : 2,
      msg_type: msgType,
      content
    })

    console.log('Send response', response)

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

      // Update conversation list (isFromSelf=true because we sent this message)
      updateConversation(
        currentSession.value.targetId,
        currentSession.value.targetType === 'user' ? 1 : (currentSession.value.targetType === 'group' ? 2 : 3),
        content,
        (response as any).data?.created_at || new Date().toISOString(),
        true // isFromSelf: true because we are the sender
      )

      return true
    }
    return false
  }

  const addMessage = (msg: Message, isFromSelf: boolean = false) => {
    console.log('addMessage called', msg, isFromSelf)
    const sessionKey = `${msg.conversation_id}:${msg.conversation_type}`
    let session = sessions.value.get(sessionKey)

    let targetId = Number(msg.conversation_id)
    const convType = Number(msg.conversation_type)

    if (convType === 1) {
        const userStore = useUserStore()
        const myId = userStore.currentUser?.id || 0
        const convId = Number(msg.conversation_id)
        
        // Reverse Pair ID: PairID = id1 * 1e9 + id2
        const id1 = Math.floor(convId / 1000000000)
        const id2 = convId % 1000000000

        if (id1 === myId) targetId = id2
        else if (id2 === myId) targetId = id1
        else {
            // Fallback if myId is not in the pair (should not happen)
            targetId = id1 === myId ? id2 : id1
        }
        
        // Special check for self-chat: if myId * 1e9 + myId == convId, then targetId is myId
        if (id1 === myId && id2 === myId) {
            targetId = myId
        }
    }

    if (!session) {
      session = {
        id: sessionKey,
        targetId: targetId,
        targetType: convType === 1 ? 'user' : (convType === 2 ? 'group' : 'ai'),
        name: `Chat ${targetId}`, // Name might be updated later
        unreadCount: 0,
        messages: []
      }
      sessions.value.set(sessionKey, session)
    }

    // Check if message already exists to avoid duplication (especially for self-chat where we push manually first)
    // Use a more robust check involving msg_id or temporary ID
    const exists = session.messages.some(m => 
        (m.msg_id && m.msg_id === msg.msg_id) || 
        (m.id && m.id === msg.id) ||
        // Check for self-sent message that might have been pushed optimistically without msg_id yet (unlikely here as msg comes from WS/API response)
        (m.seq && m.seq === msg.seq && m.from_user_id === msg.from_user_id)
    )

    if (!exists) {
        session.messages.push(msg)
        saveSessionsToStorage()

        // Update conversation list
        // If it's a new message from someone else, we should update the list
        console.log('Calling updateConversation with', targetId, convType)
        updateConversation(
            targetId,
            convType,
            msg.content,
            msg.created_at,
            isFromSelf // Pass whether this is from self to handle unread count correctly
        )
    } else {
        console.log('Message already exists, skipping push', msg)
        // Ensure conversation is updated even if message exists (e.g. optimistic update didn't update list?)
        // Actually optimistic update in sendMessage DOES update list.
        // But if WS comes later, we might want to refresh state?
        // Let's safe-guard:
        if (session.messages.length > 0) {
             const lastMsg = session.messages[session.messages.length - 1]
             updateConversation(
                targetId,
                convType,
                lastMsg.content,
                lastMsg.created_at || new Date().toISOString(),
                isFromSelf
             )
        }
    }

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

  const updateConversation = (targetId: number, type: number, content: string, time: string, isFromSelf: boolean = false) => {
    // For AI chat, we might need to handle targetId carefully
    // In chat store, conversation_id for AI is just the bot_id (targetId)
    // But in conversations list, it might be stored with a specific ID

    // Ensure targetId is a number for comparison
    const targetIdNum = Number(targetId)
    const typeNum = Number(type)

    const userStore = useUserStore()
    const currentUserId = userStore.currentUser?.id || 0

    // Allow self-chat, but ensure we don't create duplicates.
    // In self-chat, targetId is currentUserId.
    
    let conv = conversations.value.find(c => c.target_id === targetIdNum && c.conversation_type === typeNum)

    if (!conv && typeNum === 3) {
        // If not found and it's AI, try to find it again or create it
        // Sometimes type 3 might be missing if it wasn't loaded from server list yet
    }

    if (conv) {
      console.log('Updating existing conversation', targetIdNum, content)
      conv.last_msg_content = content
      conv.last_msg_time = time

      // Handle unread count based on who sent the message
      if (isFromSelf) {
        // If we sent the message, reset unread count
        conv.unread_count = 0
      } else {
        // If someone else sent the message and we're not currently viewing this chat, increment unread
        const conversationId = typeNum === 1 ? (() => {
            const id1 = Math.min(currentUserId, targetIdNum)
            const id2 = Math.max(currentUserId, targetIdNum)
            return id1 * 1000000000 + id2
        })() : targetIdNum

        const sessionKey = `${conversationId}:${typeNum}`
        // Only increment if this is not the currently active session
        if (currentSession.value?.id !== sessionKey) {
          conv.unread_count = (conv.unread_count || 0) + 1
        } else {
          conv.unread_count = 0
        }
      }

      // Move to top
      const index = conversations.value.indexOf(conv)
      if (index > 0) {
        conversations.value.splice(index, 1)
        conversations.value.unshift(conv)
      }
    } else {
        console.log('Creating new conversation', targetIdNum, content)
        // If conversation doesn't exist in list, create it!
        // This is crucial for new chats (especially AI) to show up with content

        // Try to get name/avatar
        let name = `Chat ${targetIdNum}`
        let avatar = ''

        // If private chat, try to find in friends list to get better name/avatar
        if (typeNum === 1) {
            if (targetIdNum === currentUserId) {
                // Self chat
                name = userStore.currentUser?.nickname || userStore.currentUser?.username || 'Me'
                avatar = userStore.currentUser?.avatar || ''
            } else {
                const friend = friends.value.find(f => 
                    (f.friend_user?.id === targetIdNum) || (f.friend_id === targetIdNum)
                )
                if (friend) {
                    name = friend.remark || friend.friend_user?.nickname || name
                    avatar = friend.friend_user?.avatar || ''
                }
            }
        }

        // If AI, try to fetch bot info from AI store (if available in context, but we are in chat store)
        // We can use a default name for now, usually openChat handles this better, but this is for background updates
        if (typeNum === 3) {
             name = `AI ${targetIdNum}`
             // We can try to import useAIStore but circular dependency might occur.
             // Ideally openChat should have created this.
        }

        const newConv: Conversation = {
             id: typeNum === 1 ? (() => {
                 const id1 = Math.min(currentUserId, targetIdNum)
                 const id2 = Math.max(currentUserId, targetIdNum)
                 return id1 * 1000000000 + id2
             })() : targetIdNum,
             user_id: currentUserId,
             target_id: targetIdNum,
             conversation_type: typeNum,
             unread_count: isFromSelf ? 0 : 1, // Set unread count based on who sent the message
             last_msg_content: content,
             last_msg_time: time,
             is_pinned: 0,
             is_muted: 0,
             target_user: {
                id: targetIdNum,
                username: name,
                nickname: name,
                avatar: avatar,
                status: 1
             }
        }
        conversations.value.unshift(newConv)
    }
    
    // Save to storage immediately after update
    saveSessionsToStorage()
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
    clearAll,
    loadConversations,
    loadFriends,
    loadFriendRequests,
    openChat,
    sendMessage,
    addMessage,
    loadMoreMessages,
    updateConversation,
    getSessionKey,
    loadSessionsFromStorage // Export this to call it after login
  }
})
