<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/modules/auth/stores'

const router = useRouter()
const userStore = useUserStore()

const isLogin = ref(true)
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const registerForm = reactive({
  username: '',
  password: '',
  confirmPassword: '',
  email: '',
  company_name: ''
})

const handleLogin = async () => {
  if (!loginForm.username || !loginForm.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }

  loading.value = true
  try {
    await userStore.login({
      username: loginForm.username,
      password: loginForm.password
    })
    ElMessage.success('登录成功')
    router.push('/')
  } catch (error: any) {
    // 错误已在响应拦截器中处理
  } finally {
    loading.value = false
  }
}

const handleRegister = async () => {
  if (!registerForm.username || !registerForm.password || !registerForm.email || !registerForm.company_name) {
    ElMessage.warning('请填写完整信息')
    return
  }

  if (registerForm.password !== registerForm.confirmPassword) {
    ElMessage.warning('两次输入的密码不一致')
    return
  }

  if (registerForm.password.length < 6) {
    ElMessage.warning('密码长度至少6位')
    return
  }

  loading.value = true
  try {
    await userStore.register({
      username: registerForm.username,
      password: registerForm.password,
      email: registerForm.email,
      company_name: registerForm.company_name
    })
    ElMessage.success('注册成功，请登录')
    isLogin.value = true
    // 清空注册表单
    registerForm.username = ''
    registerForm.password = ''
    registerForm.confirmPassword = ''
    registerForm.email = ''
    registerForm.company_name = ''
  } catch (error: any) {
    // 错误已在响应拦截器中处理
  } finally {
    loading.value = false
  }
}

const handleSubmit = () => {
  if (isLogin.value) {
    handleLogin()
  } else {
    handleRegister()
  }
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
        
        <!-- 登录表单 -->
        <template v-if="isLogin">
          <div class="form-group">
            <label class="form-label">用户名</label>
            <input 
              v-model="loginForm.username"
              type="text" 
              class="input" 
              placeholder="请输入用户名"
              :disabled="loading"
            />
          </div>
          
          <div class="form-group">
            <label class="form-label">密码</label>
            <input 
              v-model="loginForm.password"
              type="password" 
              class="input" 
              placeholder="请输入密码"
              :disabled="loading"
            />
          </div>
        </template>

        <!-- 注册表单 -->
        <template v-else>
          <div class="form-group">
            <label class="form-label">企业名称</label>
            <input 
              v-model="registerForm.company_name"
              type="text" 
              class="input" 
              placeholder="请输入企业名称"
              :disabled="loading"
            />
          </div>

          <div class="form-group">
            <label class="form-label">用户名</label>
            <input 
              v-model="registerForm.username"
              type="text" 
              class="input" 
              placeholder="请输入用户名"
              :disabled="loading"
            />
          </div>

          <div class="form-group">
            <label class="form-label">邮箱</label>
            <input 
              v-model="registerForm.email"
              type="email" 
              class="input" 
              placeholder="请输入邮箱"
              :disabled="loading"
            />
          </div>
          
          <div class="form-group">
            <label class="form-label">密码</label>
            <input 
              v-model="registerForm.password"
              type="password" 
              class="input" 
              placeholder="请输入密码"
              :disabled="loading"
            />
          </div>

          <div class="form-group">
            <label class="form-label">确认密码</label>
            <input 
              v-model="registerForm.confirmPassword"
              type="password" 
              class="input" 
              placeholder="请再次输入密码"
              :disabled="loading"
            />
          </div>
        </template>
        
        <button 
          type="submit" 
          class="btn btn-primary submit-btn"
          :disabled="loading"
        >
          {{ loading ? '处理中...' : (isLogin ? '登录' : '注册') }}
        </button>
        
        <div class="form-footer">
          <span>{{ isLogin ? '没有账号？' : '已有账号？' }}</span>
          <button type="button" class="link-btn" @click="toggleMode" :disabled="loading">
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

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
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
