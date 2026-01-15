import { ref } from 'vue'

export type Theme = 'light' | 'dark'

const THEME_KEY = 'app_theme'

export function useTheme() {
  const theme = ref<Theme>((localStorage.getItem(THEME_KEY) as Theme) || 'light')

  const setTheme = (newTheme: Theme) => {
    theme.value = newTheme
    document.documentElement.setAttribute('data-theme', newTheme)
    localStorage.setItem(THEME_KEY, newTheme)
  }

  const toggleTheme = () => {
    setTheme(theme.value === 'light' ? 'dark' : 'light')
  }

  const initTheme = () => {
    document.documentElement.setAttribute('data-theme', theme.value)
  }

  return {
    theme,
    setTheme,
    toggleTheme,
    initTheme
  }
}
