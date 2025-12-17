import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { storage } from '@/utils/storage'

export type ThemeType = 'light' | 'dark'

const THEME_KEY = 'app-theme'

export const useThemeStore = defineStore('theme', () => {
  // 状态
  const theme = ref<ThemeType>(storage.get<ThemeType>(THEME_KEY) || 'light')

  // 计算属性
  const isDark = computed(() => theme.value === 'dark')
  const isLight = computed(() => theme.value === 'light')

  // 应用主题到 DOM
  function applyTheme(newTheme: ThemeType) {
    document.documentElement.setAttribute('data-theme', newTheme)
    // 同时设置 Element Plus 的暗色模式 class
    if (newTheme === 'dark') {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }

  // 初始化主题（在应用启动时调用）
  function initTheme() {
    applyTheme(theme.value)
  }

  // 切换主题
  function toggleTheme() {
    const newTheme: ThemeType = theme.value === 'light' ? 'dark' : 'light'
    setTheme(newTheme)
  }

  // 设置主题
  function setTheme(newTheme: ThemeType) {
    theme.value = newTheme
    storage.set(THEME_KEY, newTheme)
    applyTheme(newTheme)
  }

  return {
    // 状态
    theme,
    // 计算属性
    isDark,
    isLight,
    // 方法
    initTheme,
    toggleTheme,
    setTheme
  }
})
