<template>
  <div class="settings-container">
    <div class="settings-layout">
      <!-- Left Sidebar -->
      <div class="settings-sidebar">
        <div class="sidebar-header">
          <div class="header-with-back">
            <el-button class="back-button" :icon="ArrowLeft" @click="$router.push('/')" circle />
            <h2>设置</h2>
          </div>
        </div>
        <div class="sidebar-menu">
          <div
            v-for="item in menuItems"
            :key="item.key"
            class="menu-item"
            :class="{ active: activeTab === item.key }"
            @click="activeTab = item.key"
          >
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </div>
        </div>
      </div>

      <!-- Right Content -->
      <div class="settings-content">
        <ProfileSettings v-if="activeTab === 'profile'" />
        <SecuritySettings v-else-if="activeTab === 'security'" />
        <ThemeSettings v-else-if="activeTab === 'theme'" />
        <NotificationSettings v-else-if="activeTab === 'notification'" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { User, Lock, Moon, Bell, ArrowLeft } from '@element-plus/icons-vue'
import ProfileSettings from '../components/settings/ProfileSettings.vue'
import SecuritySettings from '../components/settings/SecuritySettings.vue'
import ThemeSettings from '../components/settings/ThemeSettings.vue'
import NotificationSettings from '../components/settings/NotificationSettings.vue'

const activeTab = ref('profile')

const menuItems = [
  { key: 'profile', label: '个人资料', icon: User },
  { key: 'security', label: '账号安全', icon: Lock },
  { key: 'theme', label: '主题设置', icon: Moon },
  { key: 'notification', label: '通知设置', icon: Bell }
]
</script>

<style scoped>
.settings-container {
  width: 100%;
  height: 100%;
  background-color: var(--bg-body);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  overflow: hidden;
}

.settings-layout {
  width: 100%;
  max-width: 1000px;
  height: 100%;
  max-height: 700px;
  background-color: var(--bg-surface);
  border-radius: var(--border-radius-xl);
  box-shadow: var(--shadow-lg);
  display: flex;
  overflow: hidden;
}

.settings-sidebar {
  width: 240px;
  min-width: 240px;
  background-color: var(--bg-sidebar);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-header {
  padding: 24px 20px;
  border-bottom: 1px solid var(--border-color);
}

.sidebar-header h2 {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.header-with-back {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-button {
  flex-shrink: 0;
}

.sidebar-menu {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
  overflow-y: auto;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: var(--border-radius-md);
  cursor: pointer;
  color: var(--text-secondary);
  transition: all 0.2s;
  white-space: nowrap;
}

.menu-item:hover {
  background-color: var(--primary-light);
  color: var(--primary-color);
}

.menu-item.active {
  background-color: var(--primary-color);
  color: white;
}

.settings-content {
  flex: 1;
  padding: 32px;
  overflow-y: auto;
  overflow-x: hidden;
}

/* Scrollbar styling */
.settings-content::-webkit-scrollbar,
.sidebar-menu::-webkit-scrollbar {
  width: 6px;
}

.settings-content::-webkit-scrollbar-track,
.sidebar-menu::-webkit-scrollbar-track {
  background: transparent;
}

.settings-content::-webkit-scrollbar-thumb,
.sidebar-menu::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 3px;
}

.settings-content::-webkit-scrollbar-thumb:hover,
.sidebar-menu::-webkit-scrollbar-thumb:hover {
  background: var(--text-light);
}

/* Responsive design */
@media (max-width: 768px) {
  .settings-container {
    padding: 0;
    align-items: flex-start;
  }

  .settings-layout {
    max-width: 100%;
    max-height: 100%;
    border-radius: 0;
    flex-direction: column;
  }

  .settings-sidebar {
    width: 100%;
    min-width: 100%;
    border-right: none;
    border-bottom: 1px solid var(--border-color);
  }

  .sidebar-header {
    padding: 16px;
  }

  .header-with-back {
    gap: 8px;
  }

  .sidebar-header h2 {
    font-size: 18px;
  }

  .sidebar-menu {
    padding: 8px;
  }

  .menu-item {
    padding: 10px 12px;
    font-size: 14px;
  }

  .settings-content {
    padding: 20px;
  }

  .back-button {
    width: 32px;
    height: 32px;
  }
}
</style>
