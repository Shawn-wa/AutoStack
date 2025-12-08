<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Edit, Delete, Refresh } from '@element-plus/icons-vue'

// 项目数据接口
interface Project {
  id: number
  name: string
  description: string
  status: 'running' | 'stopped' | 'pending'
  deployments: number
  lastDeploy: string
}

// 项目列表数据
const projects = ref<Project[]>([
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
  {
    id: 4,
    name: 'order-service',
    description: '订单服务微服务',
    status: 'pending',
    deployments: 3,
    lastDeploy: '2024-01-13 09:20',
  },
])

// 搜索关键词
const searchKeyword = ref('')

// 弹窗控制
const dialogVisible = ref(false)
const dialogTitle = ref('新建项目')
const isEditing = ref(false)

// 表单数据
const formRef = ref()
const form = reactive({
  id: 0,
  name: '',
  description: '',
  status: 'stopped' as 'running' | 'stopped' | 'pending',
})

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '请输入项目名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' },
  ],
  description: [
    { required: true, message: '请输入项目描述', trigger: 'blur' },
  ],
}

// 过滤后的项目列表
const filteredProjects = computed(() => {
  if (!searchKeyword.value) return projects.value
  return projects.value.filter(
    (p) =>
      p.name.toLowerCase().includes(searchKeyword.value.toLowerCase()) ||
      p.description.toLowerCase().includes(searchKeyword.value.toLowerCase())
  )
})

// 状态标签类型映射
const statusType = (status: string) => {
  const map: Record<string, string> = {
    running: 'success',
    stopped: 'danger',
    pending: 'warning',
  }
  return map[status] || 'info'
}

// 状态文字映射
const statusText = (status: string) => {
  const map: Record<string, string> = {
    running: '运行中',
    stopped: '已停止',
    pending: '等待中',
  }
  return map[status] || status
}

// 打开新建弹窗
const openCreateDialog = () => {
  dialogTitle.value = '新建项目'
  isEditing.value = false
  form.id = 0
  form.name = ''
  form.description = ''
  form.status = 'stopped'
  dialogVisible.value = true
}

// 打开编辑弹窗
const openEditDialog = (row: Project) => {
  dialogTitle.value = '编辑项目'
  isEditing.value = true
  form.id = row.id
  form.name = row.name
  form.description = row.description
  form.status = row.status
  dialogVisible.value = true
}

// 提交表单
const submitForm = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate((valid: boolean) => {
    if (valid) {
      if (isEditing.value) {
        // 编辑项目
        const index = projects.value.findIndex((p) => p.id === form.id)
        if (index !== -1) {
          projects.value[index] = {
            ...projects.value[index],
            name: form.name,
            description: form.description,
            status: form.status,
          }
        }
        ElMessage.success('项目更新成功')
      } else {
        // 新建项目
        const newProject: Project = {
          id: Date.now(),
          name: form.name,
          description: form.description,
          status: form.status,
          deployments: 0,
          lastDeploy: '-',
        }
        projects.value.push(newProject)
        ElMessage.success('项目创建成功')
      }
      dialogVisible.value = false
    }
  })
}

// 删除项目
const handleDelete = (row: Project) => {
  ElMessageBox.confirm(
    `确定要删除项目「${row.name}」吗？此操作不可恢复。`,
    '删除确认',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(() => {
    const index = projects.value.findIndex((p) => p.id === row.id)
    if (index !== -1) {
      projects.value.splice(index, 1)
      ElMessage.success('项目已删除')
    }
  }).catch(() => {
    // 取消删除
  })
}

// 部署项目
const handleDeploy = (row: Project) => {
  ElMessage.info(`正在部署项目: ${row.name}`)
}

// 刷新列表
const handleRefresh = () => {
  ElMessage.success('列表已刷新')
}
</script>

<template>
  <div class="projects-page">
    <!-- 页面操作栏 -->
    <div class="page-header">
      <div class="header-left">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索项目名称或描述..."
          :prefix-icon="Search"
          clearable
          class="search-input"
        />
      </div>
      <div class="header-right">
        <el-button :icon="Refresh" @click="handleRefresh">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreateDialog">
          新建项目
        </el-button>
      </div>
    </div>

    <!-- 项目表格 -->
    <el-table
      :data="filteredProjects"
      stripe
      class="projects-table"
    >
      <el-table-column prop="name" label="项目名称" min-width="150">
        <template #default="{ row }">
          <div class="project-name">
            <span class="name-icon">▦</span>
            <span class="name-text">{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      
      <el-table-column prop="description" label="描述" min-width="200" />
      
      <el-table-column prop="status" label="状态" width="120" align="center">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" effect="dark">
            {{ statusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      
      <el-table-column prop="deployments" label="部署次数" width="120" align="center">
        <template #default="{ row }">
          <span class="deploy-count">{{ row.deployments }}</span>
        </template>
      </el-table-column>
      
      <el-table-column prop="lastDeploy" label="最近部署" width="180" />
      
      <el-table-column label="操作" width="240" align="center" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleDeploy(row)">
            部署
          </el-button>
          <el-button size="small" :icon="Edit" @click="openEditDialog(row)">
            编辑
          </el-button>
          <el-button
            size="small"
            type="danger"
            :icon="Delete"
            @click="handleDelete(row)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 新建/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="80px"
        class="project-form"
      >
        <el-form-item label="项目名称" prop="name">
          <el-input
            v-model="form.name"
            placeholder="请输入项目名称"
            maxlength="50"
            show-word-limit
          />
        </el-form-item>
        
        <el-form-item label="项目描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入项目描述"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
        
        <el-form-item label="状态" prop="status">
          <el-select v-model="form.status" placeholder="请选择状态" style="width: 100%">
            <el-option label="运行中" value="running" />
            <el-option label="已停止" value="stopped" />
            <el-option label="等待中" value="pending" />
          </el-select>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">
          {{ isEditing ? '保存修改' : '创建项目' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.projects-page {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.header-left {
  flex: 1;
  max-width: 400px;
}

.header-right {
  display: flex;
  gap: 12px;
}

.search-input {
  :deep(.el-input__wrapper) {
    background-color: var(--bg-card) !important;
  }
}

.projects-table {
  :deep(.el-table__header-wrapper) {
    th {
      font-weight: 600;
    }
  }
}

.project-name {
  display: flex;
  align-items: center;
  gap: 10px;
  
  .name-icon {
    width: 32px;
    height: 32px;
    background: rgba(0, 212, 255, 0.1);
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
    color: var(--color-primary);
  }
  
  .name-text {
    font-weight: 500;
    font-family: var(--font-mono);
  }
}

.deploy-count {
  font-family: var(--font-mono);
  font-weight: 600;
  color: var(--color-primary);
}

.project-form {
  padding: 10px 20px;
}
</style>