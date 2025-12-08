<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const sidebarCollapsed = ref(false)

const menuItems = [
  { path: '/', name: 'Dashboard', icon: '◇', label: '控制台' },
  { path: '/projects', name: 'Projects', icon: '▦', label: '项目管理' },
  { path: '/deployments', name: 'Deployments', icon: '▶', label: '部署管理' },
  { path: '/templates', name: 'Templates', icon: '❖', label: '模板市场' },
]

const isActive = (path: string) => route.path === path

const navigateTo = (path: string) => {
  router.push(path)
}

const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
}
</script>

<template>
  <div class="layout" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
    <!-- 侧边栏 -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <div class="logo">
          <span class="logo-icon">⚡</span>
          <span class="logo-text" v-show="!sidebarCollapsed">AutoStack</span>
        </div>
        <button class="toggle-btn" @click="toggleSidebar">
          {{ sidebarCollapsed ? '→' : '←' }}
        </button>
      </div>
      
      <nav class="sidebar-nav">
        <button
          v-for="item in menuItems"
          :key="item.path"
          class="nav-item"
          :class="{ active: isActive(item.path) }"
          @click="navigateTo(item.path)"
        >
          <span class="nav-icon">{{ item.icon }}</span>
          <span class="nav-label" v-show="!sidebarCollapsed">{{ item.label }}</span>
        </button>
      </nav>
      
      <div class="sidebar-footer">
        <div class="user-info" v-show="!sidebarCollapsed">
          <div class="user-avatar">A</div>
          <div class="user-details">
            <div class="user-name">Admin</div>
            <div class="user-role">管理员</div>
          </div>
        </div>
      </div>
    </aside>
    
    <!-- 主内容区 -->
    <main class="main-content">
      <header class="top-header">
        <div class="header-title">
          <h1>{{ route.meta.title || 'AutoStack' }}</h1>
        </div>
        <div class="header-actions">
          <button class="btn btn-primary">
            <span>+</span> 新建项目
          </button>
        </div>
      </header>
      
      <div class="content-wrapper">
        <RouterView />
      </div>
    </main>
  </div>
</template>

<style scoped lang="scss">
.layout {
  display: flex;
  min-height: 100vh;
}

.sidebar {
  width: 240px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  transition: width var(--transition-normal);
  position: fixed;
  height: 100vh;
  z-index: 100;
}

.sidebar-collapsed .sidebar {
  width: 72px;
}

.sidebar-header {
  padding: 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border-color);
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-icon {
  font-size: 24px;
  color: var(--color-primary);
  text-shadow: 0 0 10px var(--color-primary);
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  background: linear-gradient(135deg, var(--color-primary), var(--color-accent));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.toggle-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 12px;
  transition: all var(--transition-fast);
  
  &:hover {
    background: var(--bg-hover);
    color: var(--color-primary);
  }
}

.sidebar-nav {
  flex: 1;
  padding: 16px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  background: transparent;
  transition: all var(--transition-fast);
  width: 100%;
  text-align: left;
  
  &:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }
  
  &.active {
    background: rgba(0, 212, 255, 0.1);
    color: var(--color-primary);
    
    .nav-icon {
      text-shadow: 0 0 8px var(--color-primary);
    }
  }
}

.nav-icon {
  font-size: 18px;
  width: 24px;
  text-align: center;
}

.nav-label {
  font-size: 14px;
  font-weight: 500;
}

.sidebar-footer {
  padding: 16px;
  border-top: 1px solid var(--border-color);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--color-primary), var(--color-accent));
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  color: var(--bg-primary);
}

.user-details {
  flex: 1;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
}

.user-role {
  font-size: 12px;
  color: var(--text-muted);
}

.main-content {
  flex: 1;
  margin-left: 240px;
  transition: margin-left var(--transition-normal);
}

.sidebar-collapsed .main-content {
  margin-left: 72px;
}

.top-header {
  height: 64px;
  padding: 0 32px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  position: sticky;
  top: 0;
  z-index: 50;
}

.header-title h1 {
  font-size: 20px;
  font-weight: 600;
}

.content-wrapper {
  padding: 32px;
}
</style>

