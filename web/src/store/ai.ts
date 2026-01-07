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
    
    // Trigger reactivity by reassigning array or using reactive object
    // For simple ref object, pushing to array inside might not trigger deep watch if not deep
    // But store refs are usually reactive. 
    // Let's create a new array to be sure
    const newMsgs = [...(messages.value[botId] || []), userMsg]
    messages.value = { ...messages.value, [botId]: newMsgs }

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
  }

  // Clear all messages
  const clearAllMessages = () => {
    messages.value = {}
  }

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
