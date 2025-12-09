<script setup lang="ts">
import { ref } from 'vue'

const templates = ref([
  {
    id: 1,
    name: 'Nginx 静态站点',
    description: '快速部署静态网站，支持 HTML/CSS/JS',
    icon: '🌐',
    category: 'Web',
    downloads: 2340,
  },
  {
    id: 2,
    name: 'Node.js 应用',
    description: '部署 Express/Koa/NestJS 等 Node.js 应用',
    icon: '💚',
    category: 'Backend',
    downloads: 1856,
  },
  {
    id: 3,
    name: 'Go 微服务',
    description: '高性能 Go 语言微服务部署模板',
    icon: '🐹',
    category: 'Backend',
    downloads: 1420,
  },
  {
    id: 4,
    name: 'Vue/React SPA',
    description: '单页应用前端部署，自动构建',
    icon: '⚡',
    category: 'Frontend',
    downloads: 3200,
  },
  {
    id: 5,
    name: 'MySQL 数据库',
    description: '一键部署 MySQL 数据库实例',
    icon: '🐬',
    category: 'Database',
    downloads: 980,
  },
  {
    id: 6,
    name: 'Redis 缓存',
    description: '部署 Redis 缓存服务',
    icon: '⚡',
    category: 'Database',
    downloads: 1120,
  },
  {
    id: 7,
    name: 'PostgreSQL',
    description: '高级关系型数据库部署',
    icon: '🐘',
    category: 'Database',
    downloads: 760,
  },
  {
    id: 8,
    name: '全栈应用',
    description: '前端 + 后端 + 数据库完整解决方案',
    icon: '🚀',
    category: 'Full Stack',
    downloads: 2100,
  },
])

const categories = ['全部', 'Web', 'Frontend', 'Backend', 'Database', 'Full Stack']
const activeCategory = ref('全部')

const filteredTemplates = ref(templates.value)

const filterByCategory = (category: string) => {
  activeCategory.value = category
  if (category === '全部') {
    filteredTemplates.value = templates.value
  } else {
    filteredTemplates.value = templates.value.filter(t => t.category === category)
  }
}
</script>

<template>
  <div class="templates-page">
    <div class="page-header">
      <h2 class="page-title">选择模板快速开始</h2>
      <p class="page-desc">预配置的部署模板，一键启动您的服务</p>
    </div>

    <div class="category-tabs">
      <button
        v-for="cat in categories"
        :key="cat"
        class="tab-btn"
        :class="{ active: activeCategory === cat }"
        @click="filterByCategory(cat)"
      >
        {{ cat }}
      </button>
    </div>

    <div class="templates-grid">
      <div 
        v-for="template in filteredTemplates" 
        :key="template.id" 
        class="template-card"
      >
        <div class="template-icon">{{ template.icon }}</div>
        <div class="template-content">
          <h3 class="template-name">{{ template.name }}</h3>
          <p class="template-desc">{{ template.description }}</p>
          <div class="template-meta">
            <span class="category-tag">{{ template.category }}</span>
            <span class="downloads">↓ {{ template.downloads }}</span>
          </div>
        </div>
        <div class="template-actions">
          <button class="btn btn-secondary">预览</button>
          <button class="btn btn-primary">使用</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.templates-page {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.page-header {
  text-align: center;
  padding: 20px 0;
}

.page-title {
  font-size: 28px;
  font-weight: 700;
  margin-bottom: 8px;
  background: linear-gradient(135deg, var(--color-primary), var(--color-accent));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.page-desc {
  color: var(--text-secondary);
}

.category-tabs {
  display: flex;
  gap: 8px;
  justify-content: center;
  flex-wrap: wrap;
}

.tab-btn {
  padding: 8px 20px;
  border-radius: 20px;
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-size: 14px;
  transition: all var(--transition-fast);
  
  &:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }
  
  &.active {
    background: var(--color-primary);
    color: var(--bg-primary);
  }
}

.templates-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 20px;
}

.template-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  transition: all var(--transition-normal);
  
  &:hover {
    border-color: var(--border-glow);
    box-shadow: var(--shadow-glow);
    transform: translateY(-4px);
  }
}

.template-icon {
  font-size: 40px;
  width: 64px;
  height: 64px;
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
}

.template-content {
  flex: 1;
}

.template-name {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 8px;
}

.template-desc {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.5;
  margin-bottom: 12px;
}

.template-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.category-tag {
  padding: 4px 10px;
  background: rgba(0, 212, 255, 0.1);
  color: var(--color-primary);
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.downloads {
  font-size: 12px;
  color: var(--text-muted);
}

.template-actions {
  display: flex;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
  
  .btn {
    flex: 1;
  }
}
</style>
