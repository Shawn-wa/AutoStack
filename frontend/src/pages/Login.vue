<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const isLogin = ref(true)
const loading = ref(false)

const form = ref({
  username: '',
  password: '',
  email: '',
})

const handleSubmit = async () => {
  loading.value = true
  // TODO: 实际登录逻辑
  setTimeout(() => {
    loading.value = false
    router.push('/')
  }, 1000)
}

const toggleMode = () => {
  isLogin.value = !isLogin.value
}
</script>

<template>
  <div class="login-page">
    <div class="login-container">
      <div class="login-header">
        <div class="logo">
          <span class="logo-icon">⚡</span>
          <span class="logo-text">AutoStack</span>
        </div>
        <p class="tagline">低代码快捷部署平台</p>
      </div>

      <form class="login-form" @submit.prevent="handleSubmit">
        <h2 class="form-title">{{ isLogin ? '登录' : '注册' }}</h2>
        
        <div class="form-group">
          <label class="form-label">用户名</label>
          <input 
            v-model="form.username"
            type="text" 
            class="input" 
            placeholder="请输入用户名"
            required
          />
        </div>
        
        <div v-if="!isLogin" class="form-group">
          <label class="form-label">邮箱</label>
          <input 
            v-model="form.email"
            type="email" 
            class="input" 
            placeholder="请输入邮箱"
            required
          />
        </div>
        
        <div class="form-group">
          <label class="form-label">密码</label>
          <input 
            v-model="form.password"
            type="password" 
            class="input" 
            placeholder="请输入密码"
            required
          />
        </div>
        
        <button 
          type="submit" 
          class="btn btn-primary submit-btn"
          :disabled="loading"
        >
          {{ loading ? '处理中...' : (isLogin ? '登录' : '注册') }}
        </button>
        
        <div class="form-footer">
          <span>{{ isLogin ? '没有账号？' : '已有账号？' }}</span>
          <button type="button" class="link-btn" @click="toggleMode">
            {{ isLogin ? '立即注册' : '立即登录' }}
          </button>
        </div>
      </form>

      <div class="features">
        <div class="feature">
          <span class="feature-icon">🚀</span>
          <span>一键部署</span>
        </div>
        <div class="feature">
          <span class="feature-icon">📦</span>
          <span>模板市场</span>
        </div>
        <div class="feature">
          <span class="feature-icon">⚡</span>
          <span>低代码配置</span>
        </div>
      </div>
    </div>

    <div class="login-bg">
      <div class="bg-grid"></div>
      <div class="bg-glow"></div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}

.login-bg {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.bg-grid {
  position: absolute;
  inset: 0;
  background-image: 
    linear-gradient(rgba(0, 212, 255, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 212, 255, 0.05) 1px, transparent 1px);
  background-size: 40px 40px;
}

.bg-glow {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 600px;
  height: 600px;
  background: radial-gradient(circle, rgba(0, 212, 255, 0.15) 0%, transparent 70%);
  pointer-events: none;
}

.login-container {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 420px;
  padding: 20px;
}

.login-header {
  text-align: center;
  margin-bottom: 40px;
}

.logo {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-bottom: 12px;
}

.logo-icon {
  font-size: 36px;
  text-shadow: 0 0 20px var(--color-primary);
}

.logo-text {
  font-size: 32px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--color-primary), var(--color-accent));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.tagline {
  color: var(--text-secondary);
  font-size: 14px;
}

.login-form {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 32px;
  box-shadow: var(--shadow-card);
}

.form-title {
  font-size: 24px;
  font-weight: 600;
  text-align: center;
  margin-bottom: 24px;
}

.form-group {
  margin-bottom: 20px;
}

.form-label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 8px;
  color: var(--text-secondary);
}

.submit-btn {
  width: 100%;
  padding: 14px;
  font-size: 16px;
  margin-top: 8px;
}

.form-footer {
  text-align: center;
  margin-top: 20px;
  font-size: 14px;
  color: var(--text-secondary);
}

.link-btn {
  background: none;
  color: var(--color-primary);
  font-weight: 500;
  margin-left: 4px;
  
  &:hover {
    text-decoration: underline;
  }
}

.features {
  display: flex;
  justify-content: center;
  gap: 32px;
  margin-top: 40px;
}

.feature {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary);
}

.feature-icon {
  font-size: 18px;
}
</style>

