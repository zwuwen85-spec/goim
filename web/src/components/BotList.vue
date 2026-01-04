<template>
  <div class="bot-list-container">
    <div class="list-header">
      <div class="list-title">AI 助手</div>
    </div>

    <div class="bot-list">
      <div
        v-for="bot in aiStore.bots"
        :key="bot.id"
        class="list-item"
        :class="{ active: currentBotId === bot.id }"
        @click="handleSelect(bot)"
      >
        <el-avatar :size="48" :style="{ backgroundColor: getBotColor(bot.id) }" shape="square" class="item-avatar">
          <el-icon size="24" color="white"><ChatDotRound /></el-icon>
        </el-avatar>
        <div class="item-content">
          <div class="item-name">{{ bot.name }}</div>
          <div class="item-desc">{{ getBotDesc(bot.personality) }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { ChatDotRound } from '@element-plus/icons-vue'
import { useAIStore } from '../store/ai'
import type { AIBot } from '../api/chat'

const emit = defineEmits(['select'])
const aiStore = useAIStore()

const currentBotId = computed(() => aiStore.currentBot?.id)

const botColors: Record<number, string> = {
  9001: '#409EFF', // assistant - blue
  9002: '#67C23A', // companion - green
  9003: '#E6A23C', // tutor - orange
  9004: '#F56C6C'  // creative - red
}

const getBotColor = (botId: number) => {
  return botColors[botId] || '#909399'
}

const getBotDesc = (personality: string) => {
  const descs: Record<string, string> = {
    assistant: '智能助手',
    companion: '聊天伙伴',
    tutor: '学习导师',
    creative: '创意助手'
  }
  return descs[personality] || 'AI 助手'
}

const handleSelect = (bot: AIBot) => {
  emit('select', bot)
}

onMounted(() => {
  if (!aiStore.bots.length) {
    aiStore.loadBots()
  }
})
</script>

<style scoped>
.bot-list-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.list-header {
  padding: 16px;
  display: flex;
  align-items: center;
}

.list-title {
  font-weight: 600;
  color: var(--text-primary);
}

.bot-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.list-item {
  display: flex;
  padding: 12px;
  gap: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: 4px;
}

.list-item:hover {
  background-color: var(--bg-body);
}

.list-item.active {
  background-color: var(--primary-light);
}

.item-avatar {
  flex-shrink: 0;
  border-radius: 12px;
}

.item-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 4px;
}

.item-name {
  font-weight: 500;
  color: var(--text-primary);
  font-size: 14px;
}

.item-desc {
  font-size: 12px;
  color: var(--text-secondary);
}
</style>
