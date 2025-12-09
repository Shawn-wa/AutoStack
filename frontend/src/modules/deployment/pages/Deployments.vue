<script setup lang="ts">
import { ref } from 'vue'

const deployments = ref([
  {
    id: 1,
    project: 'web-frontend',
    environment: 'production',
    status: 'running',
    version: 'v1.2.3',
    createdAt: '2024-01-15 14:30:00',
    cpu: '12%',
    memory: '256MB',
  },
  {
    id: 2,
    project: 'api-gateway',
    environment: 'staging',
    status: 'pending',
    version: 'v2.0.0-beta',
    createdAt: '2024-01-15 14:15:00',
    cpu: '-',
    memory: '-',
  },
  {
    id: 3,
    project: 'user-service',
    environment: 'development',
    status: 'running',
    version: 'v1.0.5',
    createdAt: '2024-01-15 10:00:00',
    cpu: '8%',
    memory: '128MB',
  },
  {
    id: 4,
    project: 'database-backup',
    environment: 'production',
    status: 'stopped',
    version: 'v1.0.0',
    createdAt: '2024-01-14 23:00:00',
    cpu: '-',
    memory: '-',
  },
])

const getEnvClass = (env: string) => {
  const classes: Record<string, string> = {
    production: 'env-prod',
    staging: 'env-staging',
    development: 'env-dev',
  }
  return classes[env] || ''
}
</script>

<template>
  <div class="deployments-page">
    <div class="page-actions">
      <div class="filters">
        <select class="input filter-select">
          <option value="">所有环境</option>
          <option value="production">Production</option>
          <option value="staging">Staging</option>
          <option value="development">Development</option>
        </select>
        <select class="input filter-select">
          <option value="">所有状态</option>
          <option value="running">运行中</option>
          <option value="pending">等待中</option>
          <option value="stopped">已停止</option>
        </select>
      </div>
      <button class="btn btn-primary">+ 新建部署</button>
    </div>

    <div class="deployments-table">
      <div class="table-header">
        <div class="col-project">项目</div>
        <div class="col-env">环境</div>
        <div class="col-status">状态</div>
        <div class="col-version">版本</div>
        <div class="col-resources">资源</div>
        <div class="col-time">部署时间</div>
        <div class="col-actions">操作</div>
      </div>
      
      <div 
        v-for="deploy in deployments" 
        :key="deploy.id" 
        class="table-row"
      >
        <div class="col-project">
          <span class="project-name">{{ deploy.project }}</span>
        </div>
        <div class="col-env">
          <span class="env-badge" :class="getEnvClass(deploy.environment)">
            {{ deploy.environment }}
          </span>
        </div>
        <div class="col-status">
          <span class="status" :class="`status-${deploy.status}`">
            {{ deploy.status === 'running' ? '运行中' : deploy.status === 'pending' ? '等待中' : '已停止' }}
          </span>
        </div>
        <div class="col-version">
          <code>{{ deploy.version }}</code>
        </div>
        <div class="col-resources">
          <span class="resource">CPU: {{ deploy.cpu }}</span>
          <span class="resource">MEM: {{ deploy.memory }}</span>
        </div>
        <div class="col-time">{{ deploy.createdAt }}</div>
        <div class="col-actions">
          <button class="btn btn-ghost" title="查看日志">📋</button>
          <button 
            v-if="deploy.status === 'running'" 
            class="btn btn-ghost" 
            title="停止"
          >⏹</button>
          <button 
            v-else 
            class="btn btn-ghost" 
            title="启动"
          >▶</button>
          <button class="btn btn-ghost" title="更多">⋯</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.deployments-page {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.page-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filters {
  display: flex;
  gap: 12px;
}

.filter-select {
  width: 160px;
}

.deployments-table {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.table-header {
  display: grid;
  grid-template-columns: 1.5fr 1fr 1fr 1fr 1.2fr 1.2fr auto;
  gap: 16px;
  padding: 16px 24px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.table-row {
  display: grid;
  grid-template-columns: 1.5fr 1fr 1fr 1fr 1.2fr 1.2fr auto;
  gap: 16px;
  padding: 16px 24px;
  align-items: center;
  border-bottom: 1px solid var(--border-color);
  transition: background var(--transition-fast);
  
  &:last-child {
    border-bottom: none;
  }
  
  &:hover {
    background: var(--bg-hover);
  }
}

.project-name {
  font-family: var(--font-mono);
  font-weight: 500;
}

.env-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.env-prod {
  background: rgba(255, 68, 102, 0.15);
  color: var(--color-danger);
}

.env-staging {
  background: rgba(255, 170, 0, 0.15);
  color: var(--color-warning);
}

.env-dev {
  background: rgba(0, 212, 255, 0.15);
  color: var(--color-primary);
}

.col-version code {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--text-secondary);
}

.col-resources {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.resource {
  font-size: 12px;
  font-family: var(--font-mono);
  color: var(--text-secondary);
}

.col-time {
  font-size: 13px;
  color: var(--text-secondary);
}

.col-actions {
  display: flex;
  gap: 4px;
}
</style>
