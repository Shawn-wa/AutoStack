<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'

const visible = ref(false)
const previewSrc = ref('')
const positionStyle = ref<{ top: string; left: string }>({ top: '0', left: '0' })

const PREVIEW_WIDTH = 400
const PREVIEW_HEIGHT = 400
const GAP = 8 // 预览框与原图的间距

const show = (src: string, event?: MouseEvent) => {
  previewSrc.value = src
  
  if (event) {
    const target = event.currentTarget as HTMLElement
    const rect = target.getBoundingClientRect()
    
    // 计算预览框位置
    let left = rect.right + GAP
    let top = rect.top
    
    // 如果右侧空间不足，显示在左侧
    if (left + PREVIEW_WIDTH > window.innerWidth) {
      left = rect.left - PREVIEW_WIDTH - GAP
    }
    
    // 如果左侧也不够，则居中显示
    if (left < 0) {
      left = (window.innerWidth - PREVIEW_WIDTH) / 2
    }
    
    // 确保不超出顶部
    if (top < 0) {
      top = GAP
    }
    
    // 确保不超出底部
    if (top + PREVIEW_HEIGHT > window.innerHeight) {
      top = window.innerHeight - PREVIEW_HEIGHT - GAP
    }
    
    positionStyle.value = {
      top: `${top}px`,
      left: `${left}px`
    }
  } else {
    // 没有事件时居中显示
    positionStyle.value = {
      top: `${(window.innerHeight - PREVIEW_HEIGHT) / 2}px`,
      left: `${(window.innerWidth - PREVIEW_WIDTH) / 2}px`
    }
  }
  
  visible.value = true
}

const hide = () => {
  visible.value = false
}

// ESC 键退出预览
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && visible.value) {
    hide()
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})

defineExpose({ show, hide })
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div 
        v-if="visible" 
        class="image-preview-container" 
        :style="positionStyle"
      >
        <img :src="previewSrc" class="preview-image" />
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped lang="scss">
.image-preview-container {
  position: fixed;
  cursor: default;
  width: 400px;
  height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  z-index: 9999;
  pointer-events: none;
}

.preview-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  border-radius: 4px;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
