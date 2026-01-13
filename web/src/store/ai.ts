import { defineStore } from 'pinia'
import { ref } from 'vue'
import { aiApi, messageApi, type AIBot, type AIMessage } from '../api/chat'
import { useChatStore } from './chat'

export const useAIStore = defineStore('ai', () => {
  const bots = ref<AIBot[]>([])
  const currentBot = ref<AIBot | null>(null)
  const messages = ref<Record<number, AIMessage[]>>({})
  const loading = ref(false)
  const sending = ref(false)

  // Load available AI bots
  const loadBots = async () => {
    try {
      const response = await aiApi.getBots()
      if ((response as any).code === 0) {
        bots.value = (response as any).data.bots || []
      }
    } catch (error: any) {
      console.error('Failed to load AI bots:', error)
      throw error
    }
  }

  // Get messages for a bot
  const getBotMessages = (botId: number): AIMessage[] => {
    return messages.value[botId] || []
  }

  // Load messages from localStorage
  const loadMessagesFromStorage = () => {
    try {
      const stored = localStorage.getItem('ai_messages')
      if (stored) {
        messages.value = JSON.parse(stored)
      }
    } catch (e) {
      console.error('Failed to load AI messages from storage', e)
    }
  }

  // Save messages to localStorage
  const saveMessagesToStorage = () => {
    try {
      localStorage.setItem('ai_messages', JSON.stringify(messages.value))
    } catch (e) {
      console.error('Failed to save AI messages to storage', e)
    }
  }

  // Load history from server
  const loadHistory = async (botId: number) => {
    try {
      // AI chat uses conversation_type = 3
      // conversation_id = botId
      const response = await messageApi.getHistory({
        conversation_id: botId,
        conversation_type: 3,
        limit: 50
      })
      
      if ((response as any).code === 0) {
        const backendMessages = (response as any).data?.messages || []
        // Map to AIMessage
        const mappedMessages: AIMessage[] = backendMessages.map((m: any) => ({
          role: m.from_user_id === botId ? 'assistant' : 'user',
          content: m.content,
          timestamp: new Date(m.created_at).getTime(),
          seq: m.seq // Store sequence number for pagination
        }))
        
        // Sort by timestamp
        mappedMessages.sort((a, b) => (a.timestamp || 0) - (b.timestamp || 0))

        // Update store
        messages.value = { ...messages.value, [botId]: mappedMessages }
        // Update localStorage as cache
        saveMessagesToStorage()
      }
    } catch (e) {
      console.error('Failed to load AI history', e)
    }
  }

  // Load more history from server (pagination)
  const loadMoreMessages = async (botId: number): Promise<number> => {
    try {
      const currentMessages = messages.value[botId] || []
      if (currentMessages.length === 0) return 0

      // Find oldest message with a sequence number
      const oldestMsg = currentMessages.find(m => m.seq !== undefined)
      if (!oldestMsg || oldestMsg.seq === undefined) return 0

      const response = await messageApi.getHistory({
        conversation_id: botId,
        conversation_type: 3,
        limit: 20,
        last_seq: oldestMsg.seq
      })

      if ((response as any).code === 0) {
        const newBackendMessages = (response as any).data?.messages || []
        if (newBackendMessages.length === 0) return 0

        // Map to AIMessage
        const mappedNewMessages: AIMessage[] = newBackendMessages.map((m: any) => ({
          role: m.from_user_id === botId ? 'assistant' : 'user',
          content: m.content,
          timestamp: new Date(m.created_at).getTime(),
          seq: m.seq
        }))

        // Filter duplicates
        const existingTimestamps = new Set(currentMessages.map(m => m.timestamp))
        const uniqueNewMessages = mappedNewMessages.filter(m => !existingTimestamps.has(m.timestamp))

        if (uniqueNewMessages.length > 0) {
          const combined = [...uniqueNewMessages, ...currentMessages]
          combined.sort((a, b) => (a.timestamp || 0) - (b.timestamp || 0))
          
          messages.value = { ...messages.value, [botId]: combined }
          saveMessagesToStorage()
          return uniqueNewMessages.length
        }
      }
    } catch (e) {
      console.error('Failed to load more AI history', e)
    }
    return 0
  }

  // Set current bot
  const setCurrentBot = (bot: AIBot) => {
    currentBot.value = bot
    if (bot) {
      if (!messages.value[bot.id]) {
        messages.value[bot.id] = []
      }
      // Load history from server
      loadHistory(bot.id)
    }
  }

  // Send message to AI
  const sendMessage = async (botId: number, userMessage: string): Promise<string | null> => {
    if (!userMessage.trim()) return null

    sending.value = true

    // Add user message (optimistic update)
    const userMsg: AIMessage = {
      role: 'user',
      content: userMessage,
      timestamp: Date.now()
    }

    if (!messages.value[botId]) {
      messages.value[botId] = []
    }
    
    const newMsgs = [...(messages.value[botId] || []), userMsg]
    messages.value = { ...messages.value, [botId]: newMsgs }
    saveMessagesToStorage()

    try {
      const response = await aiApi.sendMessage({
        bot_id: botId,
        message: userMessage
      })

      if ((response as any).code === 0) {
        // Add AI response
        const aiMsg: AIMessage = {
          role: 'assistant',
          content: (response as any).data.reply,
          timestamp: Date.now()
        }
        
        const updatedMsgs = [...(messages.value[botId] || []), aiMsg]
        messages.value = { ...messages.value, [botId]: updatedMsgs }
        saveMessagesToStorage()

        // Update conversation list preview with AI reply
        const chatStore = useChatStore()
        chatStore.updateConversation(
          botId, 
          3, 
          aiMsg.content, 
          new Date().toISOString()
        )

        return (response as any).data.reply
      }
      return null
    } catch (error: any) {
      console.error('Failed to send AI message:', error)
      throw error
    } finally {
      sending.value = false
    }
  }

  // Clear messages for a bot
  const clearMessages = (botId: number) => {
    messages.value = { ...messages.value, [botId]: [] }
    saveMessagesToStorage()
    
    // Also clear conversation list preview
    const chatStore = useChatStore()
    chatStore.updateConversation(
      botId, 
      3, 
      '', 
      new Date().toISOString()
    )
  }

  // Clear all messages
  const clearAllMessages = () => {
    messages.value = {}
    saveMessagesToStorage()
  }

  // Initialize
  loadBots()
  loadMessagesFromStorage()

  // Get default bots (system bots)
  const defaultBots = ref<AIBot[]>([
    { id: 9001, name: '智能助手', personality: 'assistant', role: 'assistant', tone: 'friendly', is_default: true },
    { id: 9002, name: '聊天伙伴', personality: 'companion', role: 'companion', tone: 'casual', is_default: true },
    { id: 9003, name: '学习导师', personality: 'tutor', role: 'tutor', tone: 'professional', is_default: true },
    { id: 9004, name: '创意助手', personality: 'creative', role: 'creative', tone: 'imaginative', is_default: true }
  ])

  // Get bot by ID
  const getBotById = (botId: number): AIBot | undefined => {
    return bots.value.find(b => b.id === botId)
  }

  return {
    bots,
    currentBot,
    messages,
    loading,
    sending,
    defaultBots,
    loadBots,
    getBotMessages,
    setCurrentBot,
    sendMessage,
    clearMessages,
    clearAllMessages,
    getBotById,
    loadMoreMessages
  }
})
