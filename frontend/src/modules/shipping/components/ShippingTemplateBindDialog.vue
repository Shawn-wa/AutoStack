<template>
  <el-dialog
    v-model="visible"
    :title="dialogTitle"
    width="700px"
    destroy-on-close
    @close="handleClose"
  >
    <!-- 已绑定的运费模版列表 -->
    <div class="mb-4">
      <div class="flex justify-between items-center mb-2">
        <span class="font-medium">已绑定运费模版</span>
        <el-button type="primary" size="small" @click="showAddDialog = true">
          添加绑定
        </el-button>
      </div>
      
      <el-table :data="boundTemplates" v-loading="loading" size="small">
        <el-table-column prop="template_name" label="模版名称" />
        <el-table-column prop="carrier" label="物流商" width="120" />
        <el-table-column label="默认" width="80" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_default" type="success" size="small">默认</el-tag>
            <el-button v-else link size="small" @click="handleSetDefault(row)">设为默认</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="sort_order" label="排序" width="80" align="center" />
        <el-table-column label="操作" width="80" align="center">
          <template #default="{ row }">
            <el-popconfirm title="确定解绑此运费模版吗？" @confirm="handleUnbind(row)">
              <template #reference>
                <el-button type="danger" link size="small">解绑</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      
      <el-empty v-if="!loading && boundTemplates.length === 0" description="暂无绑定的运费模版" />
    </div>

    <!-- 添加绑定对话框 -->
    <el-dialog
      v-model="showAddDialog"
      title="添加运费模版绑定"
      width="400px"
      append-to-body
    >
      <el-form :model="bindForm" label-width="100px">
        <el-form-item label="运费模版" required>
          <el-select v-model="bindForm.shipping_template_id" placeholder="请选择运费模版" style="width: 100%">
            <el-option
              v-for="item in availableTemplates"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="设为默认">
          <el-switch v-model="bindForm.is_default" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="bindForm.sort_order" :min="0" :max="999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="handleBind" :loading="binding">确定</el-button>
      </template>
    </el-dialog>

    <template #footer>
      <el-button @click="handleClose">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import shippingApi, { 
  ProductShippingTemplate, 
  PlatformProductShippingTemplate,
  TemplateOption 
} from '../api'

// Props
const props = defineProps<{
  modelValue: boolean
  type: 'product' | 'platform_product'  // 绑定类型
  targetId: number  // 产品ID或平台产品ID
  targetName?: string  // 产品名称（用于显示）
}>()

// Emits
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'updated'): void
}>()

// 响应式数据
const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const dialogTitle = computed(() => {
  const typeName = props.type === 'product' ? '本地产品' : '平台产品'
  return `${typeName}运费模版绑定${props.targetName ? ` - ${props.targetName}` : ''}`
})

const loading = ref(false)
const binding = ref(false)
const showAddDialog = ref(false)
const boundTemplates = ref<(ProductShippingTemplate | PlatformProductShippingTemplate)[]>([])
const allTemplates = ref<TemplateOption[]>([])

const bindForm = ref({
  shipping_template_id: null as number | null,
  is_default: false,
  sort_order: 0
})

// 计算可用的运费模版（排除已绑定的）
const availableTemplates = computed(() => {
  const boundIds = boundTemplates.value.map(t => t.shipping_template_id)
  return allTemplates.value.filter(t => !boundIds.includes(t.id))
})

// 监听对话框打开
watch(visible, async (val) => {
  if (val && props.targetId) {
    await loadData()
  }
})

// 加载数据
async function loadData() {
  loading.value = true
  try {
    // 并行加载
    const [templatesRes, allRes] = await Promise.all([
      props.type === 'product'
        ? shippingApi.getProductShippingTemplates(props.targetId)
        : shippingApi.getPlatformProductShippingTemplates(props.targetId),
      shippingApi.listAllTemplates()
    ])
    
    boundTemplates.value = (templatesRes as any).data || []
    allTemplates.value = (allRes as any).data || []
  } catch (error) {
    console.error('加载运费模版失败', error)
  } finally {
    loading.value = false
  }
}

// 绑定运费模版
async function handleBind() {
  if (!bindForm.value.shipping_template_id) {
    ElMessage.warning('请选择运费模版')
    return
  }
  
  binding.value = true
  try {
    if (props.type === 'product') {
      await shippingApi.bindProductShippingTemplate({
        product_id: props.targetId,
        shipping_template_id: bindForm.value.shipping_template_id,
        is_default: bindForm.value.is_default,
        sort_order: bindForm.value.sort_order
      })
    } else {
      await shippingApi.bindPlatformProductShippingTemplate({
        platform_product_id: props.targetId,
        shipping_template_id: bindForm.value.shipping_template_id,
        is_default: bindForm.value.is_default,
        sort_order: bindForm.value.sort_order
      })
    }
    
    ElMessage.success('绑定成功')
    showAddDialog.value = false
    resetBindForm()
    await loadData()
    emit('updated')
  } catch (error) {
    ElMessage.error('绑定失败')
  } finally {
    binding.value = false
  }
}

// 解绑运费模版
async function handleUnbind(row: ProductShippingTemplate | PlatformProductShippingTemplate) {
  try {
    if (props.type === 'product') {
      await shippingApi.unbindProductShippingTemplate(row.id)
    } else {
      await shippingApi.unbindPlatformProductShippingTemplate(row.id)
    }
    
    ElMessage.success('解绑成功')
    await loadData()
    emit('updated')
  } catch (error) {
    ElMessage.error('解绑失败')
  }
}

// 设置默认运费模版
async function handleSetDefault(row: ProductShippingTemplate | PlatformProductShippingTemplate) {
  try {
    if (props.type === 'product') {
      await shippingApi.setProductDefaultShippingTemplate(props.targetId, row.shipping_template_id)
    } else {
      await shippingApi.setPlatformProductDefaultShippingTemplate(props.targetId, row.shipping_template_id)
    }
    
    ElMessage.success('设置成功')
    await loadData()
    emit('updated')
  } catch (error) {
    ElMessage.error('设置失败')
  }
}

// 重置绑定表单
function resetBindForm() {
  bindForm.value = {
    shipping_template_id: null,
    is_default: false,
    sort_order: 0
  }
}

// 关闭对话框
function handleClose() {
  visible.value = false
  resetBindForm()
}
</script>
