import { ref } from 'vue'

export interface NotificationSettings {
  desktop: boolean
  sound: boolean
  preview: boolean
}

const NOTIFICATION_KEY = 'notification_settings'

export function useNotification() {
  const loadSettings = (): NotificationSettings => {
    try {
      const saved = localStorage.getItem(NOTIFICATION_KEY)
      if (saved) {
        return JSON.parse(saved)
      }
    } catch (e) {
      console.error('Failed to load notification settings', e)
    }
    return { desktop: true, sound: true, preview: true }
  }

  const settings = ref<NotificationSettings>(loadSettings())

  const updateSettings = (newSettings: Partial<NotificationSettings>) => {
    settings.value = { ...settings.value, ...newSettings }
    localStorage.setItem(NOTIFICATION_KEY, JSON.stringify(settings.value))
  }

  const requestPermission = async (): Promise<boolean> => {
    if ('Notification' in window) {
      const permission = await Notification.requestPermission()
      return permission === 'granted'
    }
    return false
  }

  const showNotification = (title: string, body: string) => {
    if (settings.value.desktop && 'Notification' in window && Notification.permission === 'granted') {
      new Notification(title, { body, icon: '/logo.png' })
    }
    if (settings.value.sound) {
      // Play notification sound
      try {
        const audio = new Audio('/notification.mp3')
        audio.play().catch(console.error)
      } catch (e) {
        console.error('Failed to play notification sound', e)
      }
    }
  }

  return {
    settings,
    updateSettings,
    requestPermission,
    showNotification
  }
}
