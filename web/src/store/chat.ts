import { defineStore } from 'pinia'
import { ref } from 'vue'
import { conversationApi, friendApi, messageApi, type Conversation, type Friend, type Message, type User } from '../api/chat'
import { useUserStore } from './user'
import { useGroupStore } from './group'

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
    try {
      const response = await conversationApi.getList()
      if ((response as any).code === 0) {
        conversations.value = (response as any).data?.conversations || []
      }
    } catch (e) {
      console.error('Failed to load conversations', e)
    }
  }

  const loadFriends = async () => {
    try {
      const response = await friendApi.getFriends()
      if ((response as any).code === 0) {
        friends.value = (response as any).data?.friends || []
      }
    } catch (e) {
      console.error('Failed to load friends', e)
    }
  }

  const openChat = async (conversationId: number, conversationType: number, name?: string, avatar?: string) => {
    const sessionKey = `${conversationId}:${conversationType}`

    // 1. Optimistically create session if not exists (so UI can show something immediately)
    if (!sessions.value.has(sessionKey)) {
        const session: ChatSession = {
          id: sessionKey,
          targetId: conversationId,
          targetType: conversationType === 1 ? 'user' : 'group',
          name: name || `Chat ${conversationId}`,
          avatar,
          unreadCount: 0,
          messages: []
        }
        sessions.value.set(sessionKey, session)
        
        // Load history in background
        messageApi.getHistory({
          conversation_id: conversationId,
          conversation_type: conversationType,
          limit: 50
        }).then(response => {
            if ((response as any).code === 0) {
                const msgs = (response as any).data?.messages || []
                if (msgs.length > 0) {
                    const sess = sessions.value.get(sessionKey)
                    if (sess) {
                        sess.messages = msgs
                    }
                }
            }
        }).catch(e => {
            console.error('Failed to load history', e)
        })
    }

    currentSession.value = sessions.value.get(sessionKey)!

    // 2. Ensure it exists in conversation list (Optimistic update for list)
    const existingIdx = conversations.value.findIndex(c => c.target_id === conversationId && c.conversation_type === conversationType)
    if (existingIdx === -1) {
        // Add temporary conversation
        const newConv: Conversation = {
            id: Date.now(),
            user_id: 0, 
            target_id: conversationId,
            conversation_type: conversationType,
            unread_count: 0,
            last_msg_content: '',
            last_msg_time: new Date().toISOString(),
            is_pinned: 0,
            is_muted: 0,
            target_user: conversationType === 1 ? { 
                id: conversationId, 
                username: '', 
                nickname: name || `User ${conversationId}`, 
                avatar: avatar || '',
                status: 1, // Add status
                created_at: new Date().toISOString() // Add created_at
            } : undefined
        }
        conversations.value.unshift(newConv)
    } else {
        // Clear unread count if exists
        const conv = conversations.value[existingIdx]
        if (conv.unread_count > 0) {
            conv.unread_count = 0
            // Mark read logic (best effort)
            const session = sessions.value.get(sessionKey)
            if (session && session.messages.length > 0) {
                const lastMsg = session.messages[session.messages.length - 1]
                messageApi.markRead({
                  conversation_id: conversationId,
                  conversation_type: conversationType,
                  msg_id: lastMsg.id
                }).catch(() => {})
            }
        }
    }
  }

  const sendMessage = async (content: string, msgType: number = 1) => {
    if (!currentSession.value) return false

    const userStore = useUserStore()
    const currentUserId = userStore.currentUser?.id || 0

    try {
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
          conversation_id: (response as any).data?.conversation_id || currentSession.value.targetId, // Fallback to targetId
          conversation_type: (response as any).data?.conversation_type || (currentSession.value.targetType === 'user' ? 1 : 2),
          msg_type: msgType,
          content,
          seq: (response as any).data?.seq,
          created_at: (response as any).data?.created_at || new Date().toISOString()
        }
        currentSession.value.messages.push(msg)
        
        // Update conversation list
        const convId = currentSession.value.targetId
        const convType = currentSession.value.targetType === 'user' ? 1 : 2
        
        const existingIdx = conversations.value.findIndex(c => c.target_id === convId && c.conversation_type === convType)
        
        if (existingIdx > -1) {
          const conv = conversations.value[existingIdx]
          conv.last_msg_content = content
          conv.last_msg_time = msg.created_at
          // Move to top
          conversations.value.splice(existingIdx, 1)
          conversations.value.unshift(conv)
        } else {
           // Create a new conversation object locally
           const groupStore = useGroupStore()
           let targetUser = undefined
           if (convType === 1) {
              const friend = friends.value.find(f => f.friend_id === convId || f.user_id === convId)
              if (friend && friend.friend_user) {
                  targetUser = friend.friend_user
              }
           }

           const newConv: Conversation = {
             id: Date.now(), // Temporary ID
             user_id: currentUserId,
             target_id: convId,
             conversation_type: convType,
             unread_count: 0,
             last_msg_content: content,
             last_msg_time: msg.created_at,
             is_pinned: 0,
             is_muted: 0,
             target_user: targetUser
           }
           conversations.value.unshift(newConv)
           
           // Force reload from server to ensure data consistency (especially for ID and relations)
           // This helps if the local creation misses some data
           loadConversations()
        }

        return true
      }
    } catch (e) {
      console.error('Failed to send message', e)
    }
    return false
  }

  const addMessage = (msg: Message) => {
    // 1. Identify the conversation
    const userStore = useUserStore()
    const myId = userStore.currentUser?.id || 0
    
    // Check if it's an existing conversation
    let existingIdx = -1
    
    if (msg.conversation_type === 2) {
        // Group: target_id is conversation_id (group_id)
        existingIdx = conversations.value.findIndex(c => c.target_id === msg.conversation_id && c.conversation_type === 2)
    } else {
        // Private
        if (msg.from_user_id === myId) {
             // Outgoing echo
             // We generally handle this in sendMessage, but if it came from another device, we might need to handle it.
             // For now, skip to avoid duplication or complex logic, assuming single device usage primarily.
        } else {
            // Incoming. target_id is sender.
            existingIdx = conversations.value.findIndex(c => c.target_id === msg.from_user_id && c.conversation_type === 1)
        }
    }

    if (existingIdx > -1) {
      const conv = conversations.value[existingIdx]
      // Update last message
      conv.last_msg_content = msg.content
      conv.last_msg_time = msg.created_at
      
      // Update unread count if not current session
      const isCurrent = currentSession.value && 
                        currentSession.value.targetId === conv.target_id && 
                        ((currentSession.value.targetType === 'user' && conv.conversation_type === 1) || 
                         (currentSession.value.targetType === 'group' && conv.conversation_type === 2))
                         
      if (!isCurrent) {
        conv.unread_count++
      }

      // Move to top
      conversations.value.splice(existingIdx, 1)
      conversations.value.unshift(conv)
    } else {
      // New conversation
      let targetId = 0
      
      if (msg.conversation_type === 2) {
        targetId = msg.conversation_id
      } else if (msg.conversation_type === 1 && msg.from_user_id !== myId) {
        targetId = msg.from_user_id
      }
      
      if (targetId !== 0) {
        const newConv: Conversation = {
          id: Date.now(),
          user_id: myId,
          target_id: targetId,
          conversation_type: msg.conversation_type,
          unread_count: 1,
          last_msg_content: msg.content,
          last_msg_time: msg.created_at,
          is_pinned: 0,
          is_muted: 0
        }
        conversations.value.unshift(newConv)
        
        // Also reload from server to be safe
        loadConversations()
      }
    }

    // 2. Add to session messages if session is open
    let sessionKey = ''
    if (msg.conversation_type === 2) {
        sessionKey = `${msg.conversation_id}:2`
    } else {
        // Private
        if (msg.from_user_id !== myId) {
             sessionKey = `${msg.from_user_id}:1`
        } else {
             // Outgoing private msg
             // We can try to match if we know the recipient, but here we only have the msg.
             // If we are in a private chat, we can check if it matches.
             if (currentSession.value && currentSession.value.targetType === 'user') {
                 // Assume it matches current session for now if we want to support multi-device sync later
             }
        }
    }

    if (sessionKey) {
        const session = sessions.value.get(sessionKey)
        if (session) {
            // Check for duplicates
            const exists = session.messages.some(m => m.msg_id === msg.msg_id || (m.id === msg.id && m.id > 0))
            if (!exists) {
                session.messages.push(msg)
            }
        }
    }
  }

  const getSessionKey = (userId: number) => {
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
