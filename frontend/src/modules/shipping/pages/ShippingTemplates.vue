<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit, Delete } from '@element-plus/icons-vue'
import api, { type ShippingTemplate, type ShippingTemplateRule, type CreateTemplateRequest, type CreateRuleRequest } from '../api'
import { formatDateTime } from '@/utils/format'

defineOptions({ name: 'ShippingTemplates' })

// 列表数据
const loading = ref(false)
const tableData = ref<ShippingTemplate[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

// 筛选条件
const keyword = ref('')
const filterStatus = ref('')

// 模板对话框
const dialogVisible = ref(false)
const dialogTitle = ref('新增运费模板')
const isEdit = ref(false)
const formLoading = ref(false)
const formData = ref<CreateTemplateRequest & { id?: number }>({
  name: '',
  carrier: '',
  from_region: '',
  description: ''
})

// 规则对话框
const ruleDialogVisible = ref(false)
const ruleLoading = ref(false)
const currentTemplate = ref<ShippingTemplate | null>(null)
const ruleList = ref<ShippingTemplateRule[]>([])

// 规则表单对话框
const ruleFormVisible = ref(false)
const ruleFormLoading = ref(false)
const isEditRule = ref(false)
const ruleFormData = ref<CreateRuleRequest & { id?: number }>({
  to_region: '',
  min_weight: 0,
  max_weight: 0,
  first_weight: 0,
  first_price: 0,
  additional_unit: 100,
  additional_price: 0,
  currency: 'CNY',
  estimated_days: 0
})

// 获取模板列表
const fetchTemplates = async () => {
  loading.value = true
  try {
    const params: any = {
      page: currentPage.value,
      page_size: pageSize.value
    }
    if (keyword.value) params.keyword = keyword.value
    if (filterStatus.value) params.status = filterStatus.value

    const res = await api.listTemplates(params)
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (error) {
    console.error('获取模板列表失败', error)
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  currentPage.value = 1
  fetchTemplates()
}

// 重置
const handleReset = () => {
  keyword.value = ''
  filterStatus.value = ''
  currentPage.value = 1
  fetchTemplates()
}

// 分页
const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchTemplates()
}

// 新增模板
const handleCreate = () => {
  isEdit.value = false
  dialogTitle.value = '新增运费模板'
  formData.value = {
    name: '',
    carrier: '',
    from_region: '',
    description: ''
  }
  dialogVisible.value = true
}

// 编辑模板
const handleEdit = (row: ShippingTemplate) => {
  isEdit.value = true
  dialogTitle.value = '编辑运费模板'
  formData.value = {
    id: row.id,
    name: row.name,
    carrier: row.carrier,
    from_region: row.from_region,
    description: row.description
  }
  dialogVisible.value = true
}

// 提交模板表单
const handleSubmit = async () => {
  if (!formData.value.name) {
    ElMessage.warning('请输入模板名称')
    return
  }

  formLoading.value = true
  try {
    if (isEdit.value && formData.value.id) {
      await api.updateTemplate(formData.value.id, formData.value)
      ElMessage.success('更新成功')
    } else {
      await api.createTemplate(formData.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchTemplates()
  } catch (error) {
    console.error('保存失败', error)
  } finally {
    formLoading.value = false
  }
}

// 切换状态
const handleToggleStatus = async (row: ShippingTemplate) => {
  const newStatus = row.status === 'active' ? 'inactive' : 'active'
  const actionText = newStatus === 'active' ? '启用' : '停用'
  
  try {
    await ElMessageBox.confirm(`确定要${actionText}该模板吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await api.updateTemplate(row.id, { status: newStatus })
    ElMessage.success(`${actionText}成功`)
    fetchTemplates()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('操作失败', error)
    }
  }
}

// 删除模板
const handleDelete = async (row: ShippingTemplate) => {
  try {
    await ElMessageBox.confirm('删除后将无法恢复，确定要删除该模板吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await api.deleteTemplate(row.id)
    ElMessage.success('删除成功')
    fetchTemplates()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除失败', error)
    }
  }
}

// ========== 规则管理 ==========

// 打开规则管理
const handleManageRules = async (row: ShippingTemplate) => {
  currentTemplate.value = row
  ruleDialogVisible.value = true
  await fetchRules()
}

// 获取规则列表
const fetchRules = async () => {
  if (!currentTemplate.value) return
  
  ruleLoading.value = true
  try {
    const res = await api.getTemplateRules(currentTemplate.value.id)
    ruleList.value = res.data || []
  } catch (error) {
    console.error('获取规则列表失败', error)
  } finally {
    ruleLoading.value = false
  }
}

// 新增规则
const handleAddRule = () => {
  isEditRule.value = false
  ruleFormData.value = {
    to_region: '',
    min_weight: 0,
    max_weight: 0,
    first_weight: 0,
    first_price: 0,
    additional_unit: 100,
    additional_price: 0,
    currency: 'CNY',
    estimated_days: 0
  }
  ruleFormVisible.value = true
}

// 编辑规则
const handleEditRule = (rule: ShippingTemplateRule) => {
  isEditRule.value = true
  ruleFormData.value = {
    id: rule.id,
    to_region: rule.to_region,
    min_weight: rule.min_weight,
    max_weight: rule.max_weight,
    first_weight: rule.first_weight,
    first_price: rule.first_price,
    additional_unit: rule.additional_unit,
    additional_price: rule.additional_price,
    currency: rule.currency,
    estimated_days: rule.estimated_days
  }
  ruleFormVisible.value = true
}

// 提交规则表单
const handleSubmitRule = async () => {
  if (!ruleFormData.value.to_region) {
    ElMessage.warning('请输入收货区域')
    return
  }
  if (!currentTemplate.value) return

  ruleFormLoading.value = true
  try {
    if (isEditRule.value && ruleFormData.value.id) {
      await api.updateRule(currentTemplate.value.id, ruleFormData.value.id, ruleFormData.value)
      ElMessage.success('更新成功')
    } else {
      await api.createRule(currentTemplate.value.id, ruleFormData.value)
      ElMessage.success('创建成功')
    }
    ruleFormVisible.value = false
    fetchRules()
    fetchTemplates() // 刷新规则数量
  } catch (error) {
    console.error('保存失败', error)
  } finally {
    ruleFormLoading.value = false
  }
}

// 删除规则
const handleDeleteRule = async (rule: ShippingTemplateRule) => {
  if (!currentTemplate.value) return
  
  try {
    await ElMessageBox.confirm('确定要删除该规则吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await api.deleteRule(currentTemplate.value.id, rule.id)
    ElMessage.success('删除成功')
    fetchRules()
    fetchTemplates()
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除失败', error)
    }
  }
}

onMounted(() => {
  fetchTemplates()
})
</script>

<template>
  <div class="shipping-templates">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">运费模板</h1>
        <p class="page-desc">管理物流运费模板，支持根据重量和区域计算运费</p>
      </div>
      <div class="header-right">
        <el-button type="primary" :icon="Plus" @click="handleCreate">
          新增模板
        </el-button>
      </div>
    </div>

    <div class="filter-card">
      <el-form inline>
        <el-form-item label="状态">
          <el-select
            v-model="filterStatus"
            placeholder="全部状态"
            clearable
            style="width: 120px"
            @change="handleSearch"
          >
            <el-option label="启用" value="active" />
            <el-option label="停用" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input
            v-model="keyword"
            placeholder="模板名称/物流商"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="table-card">
      <el-table
        v-loading="loading"
        :data="tableData"
        stripe
      >
        <el-table-column label="模板名称" prop="name" min-width="150" />
        <el-table-column label="物流商" prop="carrier" width="150" />
        <el-table-column label="发货区域" prop="from_region" width="120" />
        <el-table-column label="规则数" prop="rule_count" width="90" align="center">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.rule_count || 0 }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
              {{ row.status === 'active' ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="描述" prop="description" min-width="200" show-overflow-tooltip />
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleManageRules(row)">规则</el-button>
            <el-button type="primary" link @click="handleEdit(row)">编辑</el-button>
            <el-button 
              :type="row.status === 'active' ? 'warning' : 'success'" 
              link 
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 'active' ? '停用' : '启用' }}
            </el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="handlePageChange"
        />
      </div>
    </div>

    <!-- 模板表单对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
      destroy-on-close
    >
      <el-form :model="formData" label-width="80px">
        <el-form-item label="模板名称" required>
          <el-input v-model="formData.name" placeholder="请输入模板名称" />
        </el-form-item>
        <el-form-item label="物流商">
          <el-input v-model="formData.carrier" placeholder="如: 顺丰、中通、极兔等" />
        </el-form-item>
        <el-form-item label="发货区域">
          <el-input v-model="formData.from_region" placeholder="如: 中国、广东等" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="模板描述信息"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 规则管理对话框 -->
    <el-dialog
      v-model="ruleDialogVisible"
      :title="`运费规则 - ${currentTemplate?.name || ''}`"
      width="1100px"
      destroy-on-close
    >
      <div class="rule-header">
        <el-button type="primary" size="small" :icon="Plus" @click="handleAddRule">
          添加规则
        </el-button>
      </div>

      <el-table
        v-loading="ruleLoading"
        :data="ruleList"
        stripe
      >
        <el-table-column label="收货区域" prop="to_region" width="120">
          <template #default="{ row }">
            {{ row.to_region === '*' ? '全部区域' : row.to_region }}
          </template>
        </el-table-column>
        <el-table-column label="重量范围(g)" width="140" align="center">
          <template #default="{ row }">
            {{ row.min_weight }} - {{ row.max_weight || '不限' }}
          </template>
        </el-table-column>
        <el-table-column label="首重(g)" prop="first_weight" width="90" align="center" />
        <el-table-column label="首重费用" width="100" align="right">
          <template #default="{ row }">
            {{ row.first_price.toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column label="续重单位(g)" prop="additional_unit" width="110" align="center" />
        <el-table-column label="续重单价" width="100" align="right">
          <template #default="{ row }">
            {{ row.additional_price.toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column label="货币" prop="currency" width="70" align="center" />
        <el-table-column label="时效(天)" prop="estimated_days" width="90" align="center">
          <template #default="{ row }">
            {{ row.estimated_days || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEditRule(row)">编辑</el-button>
            <el-button type="danger" link @click="handleDeleteRule(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <template v-if="ruleList.length === 0 && !ruleLoading">
        <el-empty description="暂无运费规则" :image-size="80" />
      </template>
    </el-dialog>

    <!-- 规则表单对话框 -->
    <el-dialog
      v-model="ruleFormVisible"
      :title="isEditRule ? '编辑规则' : '添加规则'"
      width="550px"
      destroy-on-close
    >
      <el-form :model="ruleFormData" label-width="100px">
        <el-form-item label="收货区域" required>
          <el-input v-model="ruleFormData.to_region" placeholder="如: RU、*（全部）" />
          <div class="form-tip">* 表示匹配所有区域</div>
        </el-form-item>
        <el-form-item label="重量范围">
          <div class="weight-range">
            <el-input-number
              v-model="ruleFormData.min_weight"
              :min="0"
              :precision="0"
              controls-position="right"
              placeholder="最小"
            />
            <span class="range-separator">-</span>
            <el-input-number
              v-model="ruleFormData.max_weight"
              :min="0"
              :precision="0"
              controls-position="right"
              placeholder="最大(0=不限)"
            />
            <span class="unit">g</span>
          </div>
        </el-form-item>
        <el-form-item label="首重">
          <el-input-number
            v-model="ruleFormData.first_weight"
            :min="0"
            :precision="0"
            controls-position="right"
          />
          <span class="unit">g</span>
        </el-form-item>
        <el-form-item label="首重费用">
          <el-input-number
            v-model="ruleFormData.first_price"
            :min="0"
            :precision="2"
            controls-position="right"
          />
          <span class="unit">{{ ruleFormData.currency }}</span>
        </el-form-item>
        <el-form-item label="续重单位">
          <el-input-number
            v-model="ruleFormData.additional_unit"
            :min="1"
            :precision="0"
            controls-position="right"
          />
          <span class="unit">g</span>
        </el-form-item>
        <el-form-item label="续重单价">
          <el-input-number
            v-model="ruleFormData.additional_price"
            :min="0"
            :precision="2"
            controls-position="right"
          />
          <span class="unit">{{ ruleFormData.currency }}/单位</span>
        </el-form-item>
        <el-form-item label="货币">
          <el-select v-model="ruleFormData.currency" style="width: 120px">
            <el-option label="CNY" value="CNY" />
            <el-option label="USD" value="USD" />
            <el-option label="EUR" value="EUR" />
            <el-option label="RUB" value="RUB" />
          </el-select>
        </el-form-item>
        <el-form-item label="预估时效">
          <el-input-number
            v-model="ruleFormData.estimated_days"
            :min="0"
            :precision="0"
            controls-position="right"
          />
          <span class="unit">天</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleFormVisible = false">取消</el-button>
        <el-button type="primary" :loading="ruleFormLoading" @click="handleSubmitRule">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style lang="scss" scoped>
.shipping-templates {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.header-left {
  .page-title {
    font-size: 20px;
    font-weight: 600;
    margin: 0 0 4px 0;
  }
  .page-desc {
    color: var(--text-secondary);
    font-size: 14px;
    margin: 0;
  }
}

.filter-card {
  background: var(--bg-card);
  border-radius: 8px;
  padding: 16px 20px 0;
  margin-bottom: 16px;
}

.table-card {
  background: var(--bg-card);
  border-radius: 8px;
  padding: 16px 20px;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.rule-header {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 16px;
}

.weight-range {
  display: flex;
  align-items: center;
  gap: 8px;
  
  .range-separator {
    color: var(--text-secondary);
  }
}

.unit {
  margin-left: 8px;
  color: var(--text-secondary);
  font-size: 14px;
}

.form-tip {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 4px;
}
</style>
