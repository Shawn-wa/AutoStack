<script setup lang="ts">
import { RouterView } from 'vue-router'
import { useThemeStore } from '@/stores/theme'

// 初始化主题
const themeStore = useThemeStore()
themeStore.initTheme()
</script>

<template>
  <RouterView />
</template>

<style scoped>
</style>

