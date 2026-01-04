import { defineStore } from 'pinia'
import { ref } from 'vue'
import { aiApi, type AIBot, type AIMessage } from '../api/chat'

export const useAIStore = defineStore('ai', () => {
  const bots = ref<AIBot[]>([])
  const currentBot = ref<AIBot | null>(null)
  const messages = ref<Map<number, AIMessage[]>>(new Map())
  const loading = ref(false)
  const sending = ref(false)

  // Load available AI bots
  const loadBots = async () => {
    try {
      const response = await aiApi.getBots()
      if (response.code === 0) {
        bots.value = response.data.bots || []
      }
    } catch (error: any) {
      console.error('Failed to load AI bots:', error)
      throw error
    }
  }

  // Get messages for a bot
  const getBotMessages = (botId: number): AIMessage[] => {
    return messages.value.get(botId) || []
  }

  // Set current bot
  const setCurrentBot = (bot: AIBot) => {
    currentBot.value = bot
    if (!messages.value.has(bot.id)) {
      messages.value.set(bot.id, [])
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

    const botMessages = messages.value.get(botId) || []
    botMessages.push(userMsg)
    messages.value.set(botId, botMessages)

    try {
      const response = await aiApi.sendMessage({
        bot_id: botId,
        message: userMessage
      })

      if (response.code === 0) {
        // Add AI response
        const aiMsg: AIMessage = {
          role: 'assistant',
          content: response.data.reply,
          timestamp: Date.now()
        }
        botMessages.push(aiMsg)
        return response.data.reply
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
    messages.value.delete(botId)
  }

  // Clear all messages
  const clearAllMessages = () => {
    messages.value.clear()
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
