import { defineStore } from 'pinia'
import { ref } from 'vue'
import { aiApi, type AIBot, type AIMessage } from '../api/chat'

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

  // Set current bot
  const setCurrentBot = (bot: AIBot) => {
    currentBot.value = bot
    if (!messages.value[bot.id]) {
      messages.value[bot.id] = []
    }
  }

  // Send message to AI
  const sendMessage = async (botId: number, userMessage: string): Promise<string | null> => {
    if (!userMessage.trim()) return null

    sending.value = true

    // Add user message
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
    const newMessages = { ...messages.value }
    delete newMessages[botId]
    messages.value = newMessages
    saveMessagesToStorage()
  }

  // Clear all messages
  const clearAllMessages = () => {
    messages.value = {}
    saveMessagesToStorage()
  }

  // Initialize
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
    getBotById
  }
})
