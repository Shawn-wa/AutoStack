<script setup lang="ts">
import { ref } from 'vue'

const visible = ref(false)
const previewSrc = ref('')

const show = (src: string) => {
  previewSrc.value = src
  visible.value = true
}

const hide = () => {
  visible.value = false
}

defineExpose({ show })
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div v-if="visible" class="image-preview-overlay" @click="hide">
        <div class="image-preview-container" @click.stop>
          <img :src="previewSrc" class="preview-image" />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped lang="scss">
.image-preview-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  cursor: pointer;
}

.image-preview-container {
  cursor: default;
  max-width: 80vw;
  max-height: 80vh;
}

.preview-image {
  max-width: 100%;
  max-height: 80vh;
  border-radius: 8px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
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
