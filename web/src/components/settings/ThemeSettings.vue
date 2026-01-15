<template>
  <div class="theme-settings">
    <div class="settings-header">
      <h3>主题设置</h3>
      <p>选择您喜欢的主题风格</p>
    </div>

    <div class="theme-options">
      <div
        v-for="theme in themes"
        :key="theme.value"
        class="theme-card"
        :class="{ active: currentTheme === theme.value }"
        @click="handleThemeChange(theme.value)"
      >
        <div class="theme-preview" :style="{ background: theme.preview }">
          <el-icon v-if="currentTheme === theme.value" class="check-icon">
            <Check />
          </el-icon>
        </div>
        <div class="theme-info">
          <div class="theme-name">{{ theme.label }}</div>
          <div class="theme-desc">{{ theme.description }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { Check } from '@element-plus/icons-vue'
import { useTheme } from '../../composables/useTheme'

const { theme: currentTheme, setTheme, initTheme } = useTheme()

const themes = [
  {
    value: 'light',
    label: '浅色模式',
    description: '明亮的主题，适合白天使用',
    preview: 'linear-gradient(135deg, #f3f4f6 0%, #ffffff 100%)'
  },
  {
    value: 'dark',
    label: '深色模式',
    description: '暗黑主题，保护眼睛',
    preview: 'linear-gradient(135deg, #111827 0%, #1f2937 100%)'
  }
]

const handleThemeChange = (newTheme: 'light' | 'dark') => {
  setTheme(newTheme)
}

onMounted(() => {
  initTheme()
})
</script>

<style scoped>
.theme-settings {
  max-width: 600px;
  width: 100%;
}

.settings-header {
  margin-bottom: 32px;
}

.settings-header h3 {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 8px 0;
}

.settings-header p {
  color: var(--text-secondary);
  margin: 0;
}

.theme-options {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.theme-card {
  border: 2px solid var(--border-color);
  border-radius: var(--border-radius-lg);
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s;
}

.theme-card:hover {
  border-color: var(--primary-color);
  box-shadow: var(--shadow-md);
}

.theme-card.active {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px var(--primary-light);
}

.theme-preview {
  height: 120px;
  border-radius: var(--border-radius-md);
  margin-bottom: 16px;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.check-icon {
  font-size: 32px;
  color: var(--primary-color);
}

.theme-info {
  text-align: center;
}

.theme-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.theme-desc {
  font-size: 13px;
  color: var(--text-secondary);
}

/* Responsive */
@media (max-width: 768px) {
  .settings-header h3 {
    font-size: 20px;
  }

  .theme-options {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .theme-preview {
    height: 100px;
  }

  .theme-name {
    font-size: 14px;
  }

  .theme-desc {
    font-size: 12px;
  }
}
</style>
