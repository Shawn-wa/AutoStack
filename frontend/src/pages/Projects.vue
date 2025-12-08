<script setup lang="ts">
import { ref } from 'vue'

const projects = ref([
  {
    id: 1,
    name: 'web-frontend',
    description: '前端应用程序',
    status: 'running',
    deployments: 5,
    lastDeploy: '2024-01-15 14:30',
  },
  {
    id: 2,
    name: 'api-gateway',
    description: 'API 网关服务',
    status: 'running',
    deployments: 12,
    lastDeploy: '2024-01-15 10:00',
  },
  {
    id: 3,
    name: 'user-service',
    description: '用户服务微服务',
    status: 'stopped',
    deployments: 8,
    lastDeploy: '2024-01-14 18:45',
  },
])

const showCreateModal = ref(false)
</script>

<template>
  <div class="projects-page">
    <div class="page-actions">
      <div class="search-box">
        <input type="text" class="input" placeholder="搜索项目..." />
      </div>
      <button class="btn btn-primary" @click="showCreateModal = true">
        + 新建项目
      </button>
    </div>

    <div class="projects-grid">
      <div v-for="project in projects" :key="project.id" class="project-card">
        <div class="project-header">
          <div class="project-icon">▦</div>
          <span class="status" :class="`status-${project.status}`">
            {{ project.status === 'running' ? '运行中' : '已停止' }}
          </span>
        </div>
        
        <h3 class="project-name">{{ project.name }}</h3>
        <p class="project-desc">{{ project.description }}</p>
        
        <div class="project-stats">
          <div class="stat">
            <span class="stat-value">{{ project.deployments }}</span>
            <span class="stat-label">部署次数</span>
          </div>
          <div class="stat">
            <span class="stat-label">最近部署</span>
            <span class="stat-value text-sm">{{ project.lastDeploy }}</span>
          </div>
        </div>
        
        <div class="project-actions">
          <button class="btn btn-secondary">配置</button>
          <button class="btn btn-primary">部署</button>
        </div>
      </div>

      <!-- 新建项目卡片 -->
      <button class="project-card new-project" @click="showCreateModal = true">
        <div class="new-project-icon">+</div>
        <span class="new-project-text">新建项目</span>
      </button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.projects-page {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.page-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.search-box {
  flex: 1;
  max-width: 400px;
}

.projects-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.project-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 24px;
  transition: all var(--transition-normal);
  
  &:hover {
    border-color: var(--border-glow);
    box-shadow: var(--shadow-glow);
  }
}

.project-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.project-icon {
  width: 40px;
  height: 40px;
  background: rgba(0, 212, 255, 0.1);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: var(--color-primary);
}

.project-name {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 8px;
  font-family: var(--font-mono);
}

.project-desc {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 20px;
}

.project-stats {
  display: flex;
  justify-content: space-between;
  padding: 16px 0;
  border-top: 1px solid var(--border-color);
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 20px;
}

.project-stats .stat {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.project-stats .stat-value {
  font-weight: 600;
  font-family: var(--font-mono);
  
  &.text-sm {
    font-size: 12px;
    font-weight: 400;
  }
}

.project-stats .stat-label {
  font-size: 12px;
  color: var(--text-muted);
}

.project-actions {
  display: flex;
  gap: 12px;
  
  .btn {
    flex: 1;
  }
}

.new-project {
  border: 2px dashed var(--border-color);
  background: transparent;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 280px;
  cursor: pointer;
  
  &:hover {
    border-color: var(--color-primary);
    background: rgba(0, 212, 255, 0.05);
    
    .new-project-icon {
      transform: scale(1.1);
      color: var(--color-primary);
    }
  }
}

.new-project-icon {
  font-size: 48px;
  color: var(--text-muted);
  margin-bottom: 12px;
  transition: all var(--transition-fast);
}

.new-project-text {
  font-size: 14px;
  color: var(--text-secondary);
}
</style>

