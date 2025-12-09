<script setup lang="ts">
import { ref } from 'vue'

const stats = ref([
  { label: '运行中项目', value: 12, icon: '▶', color: 'success' },
  { label: '部署总数', value: 48, icon: '◉', color: 'primary' },
  { label: '可用模板', value: 24, icon: '❖', color: 'accent' },
  { label: '今日部署', value: 6, icon: '↑', color: 'warning' },
])

const recentDeployments = ref([
  { id: 1, name: 'web-frontend', env: 'prod', status: 'running', time: '2分钟前' },
  { id: 2, name: 'api-gateway', env: 'staging', status: 'pending', time: '15分钟前' },
  { id: 3, name: 'user-service', env: 'dev', status: 'running', time: '1小时前' },
  { id: 4, name: 'database-backup', env: 'prod', status: 'stopped', time: '3小时前' },
])
</script>

<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <section class="stats-grid">
      <div 
        v-for="stat in stats" 
        :key="stat.label" 
        class="stat-card"
        :class="`stat-${stat.color}`"
      >
        <div class="stat-icon">{{ stat.icon }}</div>
        <div class="stat-content">
          <div class="stat-value">{{ stat.value }}</div>
          <div class="stat-label">{{ stat.label }}</div>
        </div>
      </div>
    </section>

    <!-- 快速操作 -->
    <section class="quick-actions">
      <h2 class="section-title">快速开始</h2>
      <div class="actions-grid">
        <button class="action-card">
          <div class="action-icon">🚀</div>
          <div class="action-text">
            <div class="action-title">一键部署</div>
            <div class="action-desc">从模板快速创建</div>
          </div>
        </button>
        <button class="action-card">
          <div class="action-icon">📦</div>
          <div class="action-text">
            <div class="action-title">导入项目</div>
            <div class="action-desc">从 Git 仓库导入</div>
          </div>
        </button>
        <button class="action-card">
          <div class="action-icon">⚙️</div>
          <div class="action-text">
            <div class="action-title">自定义配置</div>
            <div class="action-desc">手动配置部署</div>
          </div>
        </button>
      </div>
    </section>

    <!-- 最近部署 -->
    <section class="recent-section">
      <h2 class="section-title">最近部署</h2>
      <div class="deployments-list">
        <div 
          v-for="deploy in recentDeployments" 
          :key="deploy.id" 
          class="deployment-item"
        >
          <div class="deploy-info">
            <div class="deploy-name">{{ deploy.name }}</div>
            <div class="deploy-env">{{ deploy.env }}</div>
          </div>
          <div class="deploy-status">
            <span class="status" :class="`status-${deploy.status}`">
              {{ deploy.status === 'running' ? '运行中' : deploy.status === 'pending' ? '等待中' : '已停止' }}
            </span>
          </div>
          <div class="deploy-time">{{ deploy.time }}</div>
          <div class="deploy-actions">
            <button class="btn btn-ghost">查看</button>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped lang="scss">
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.stat-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 20px;
  transition: all var(--transition-normal);
  
  &:hover {
    transform: translateY(-4px);
    box-shadow: var(--shadow-card);
  }
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.stat-success .stat-icon {
  background: rgba(0, 255, 136, 0.1);
  color: var(--color-success);
}

.stat-primary .stat-icon {
  background: rgba(0, 212, 255, 0.1);
  color: var(--color-primary);
}

.stat-accent .stat-icon {
  background: rgba(0, 255, 136, 0.1);
  color: var(--color-accent);
}

.stat-warning .stat-icon {
  background: rgba(255, 170, 0, 0.1);
  color: var(--color-warning);
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  font-family: var(--font-mono);
}

.stat-label {
  font-size: 14px;
  color: var(--text-secondary);
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 16px;
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.action-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  text-align: left;
  transition: all var(--transition-normal);
  
  &:hover {
    border-color: var(--color-primary);
    background: var(--bg-hover);
    
    .action-icon {
      transform: scale(1.1);
    }
  }
}

.action-icon {
  font-size: 32px;
  transition: transform var(--transition-fast);
}

.action-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 4px;
}

.action-desc {
  font-size: 13px;
  color: var(--text-secondary);
}

.deployments-list {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.deployment-item {
  display: grid;
  grid-template-columns: 1fr auto auto auto;
  align-items: center;
  gap: 24px;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color);
  transition: background var(--transition-fast);
  
  &:last-child {
    border-bottom: none;
  }
  
  &:hover {
    background: var(--bg-hover);
  }
}

.deploy-name {
  font-weight: 500;
  font-family: var(--font-mono);
}

.deploy-env {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
}

.deploy-time {
  font-size: 13px;
  color: var(--text-secondary);
}

@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .actions-grid {
    grid-template-columns: 1fr;
  }
}
</style>
