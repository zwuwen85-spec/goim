<template>
  <div class="notification-settings">
    <div class="settings-header">
      <h3>通知设置</h3>
      <p>管理消息通知偏好</p>
    </div>

    <div class="settings-body">
      <!-- Desktop Notification -->
      <div class="setting-item">
        <div class="item-header">
          <div class="item-info">
            <div class="item-title">桌面通知</div>
            <div class="item-desc">在桌面显示新消息通知</div>
          </div>
          <el-switch
            v-model="settings.desktop"
            @change="handleDesktopChange"
          />
        </div>
        <div v-if="!notificationPermission" class="permission-tip">
          <el-alert
            type="warning"
            :closable="false"
            show-icon
          >
            <template #title>
              <span>浏览器未授权通知权限，</span>
              <el-button type="primary" link @click="requestPermission">
                点击授权
              </el-button>
            </template>
          </el-alert>
        </div>
      </div>

      <!-- Sound Notification -->
      <div class="setting-item">
        <div class="item-header">
          <div class="item-info">
            <div class="item-title">声音提醒</div>
            <div class="item-desc">接收消息时播放提示音</div>
          </div>
          <el-switch
            v-model="settings.sound"
            @change="handleSettingChange"
          />
        </div>
      </div>

      <!-- Message Preview -->
      <div class="setting-item">
        <div class="item-header">
          <div class="item-info">
            <div class="item-title">消息预览</div>
            <div class="item-desc">在通知中显示消息内容</div>
          </div>
          <el-switch
            v-model="settings.preview"
            @change="handleSettingChange"
          />
        </div>
      </div>

      <!-- Test Notification -->
      <div class="setting-actions">
        <el-button @click="handleTestNotification">
          <el-icon><Bell /></el-icon>
          测试通知
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Bell } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useNotification } from '../../composables/useNotification'

const { settings, updateSettings, requestPermission, showNotification } = useNotification()

const notificationPermission = computed(() => {
  if ('Notification' in window) {
    return Notification.permission === 'granted'
  }
  return false
})

const handleDesktopChange = async (value: boolean) => {
  if (value && !notificationPermission.value) {
    const granted = await requestPermission()
    if (!granted) {
      settings.desktop = false
      ElMessage.warning('请在浏览器设置中允许通知权限')
      return
    }
  }
  handleSettingChange()
}

const handleSettingChange = () => {
  updateSettings({
    desktop: settings.desktop,
    sound: settings.sound,
    preview: settings.preview
  })
}

const handleTestNotification = () => {
  showNotification(
    '测试通知',
    settings.preview ? '这是一条测试消息内容' : '您收到了一条新消息'
  )
}
</script>

<style scoped>
.notification-settings {
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

.settings-body {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.setting-item {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.item-info {
  flex: 1;
  min-width: 0;
}

.item-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.item-desc {
  font-size: 14px;
  color: var(--text-secondary);
}

.permission-tip {
  margin-left: 0;
}

.setting-actions {
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
}

/* Responsive */
@media (max-width: 768px) {
  .settings-header h3 {
    font-size: 20px;
  }

  .item-header {
    flex-wrap: wrap;
  }

  .item-title {
    font-size: 14px;
  }

  .item-desc {
    font-size: 13px;
  }

  :deep(.el-alert) {
    font-size: 13px;
  }

  :deep(.el-alert__title) {
    font-size: 13px;
  }
}
</style>
