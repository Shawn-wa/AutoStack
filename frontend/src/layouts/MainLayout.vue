<script setup lang="ts">
import { computed, ref, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/modules/auth/stores'
import { useThemeStore } from '@/stores/theme'
import { useTabsStore, type TabItem } from '@/stores/tabs'
import { ElMessageBox } from 'element-plus'
import { Close } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const themeStore = useThemeStore()
const tabsStore = useTabsStore()
const sidebarCollapsed = ref(false)

// 所有菜单项（带 permission 字段）
const allMenuItems = [
  { path: '/', name: 'Dashboard', icon: '🏠', label: '首页', permission: 'dashboard:view' },
  { 
    icon: '📦', 
    label: '产品管理',
    permission: 'product:view',
    children: [
      { path: '/product/products', name: 'LocalProducts', label: '系统产品' },
      { path: '/product/platform-products', name: 'PlatformProducts', label: '平台产品' },
      { path: '/product/summary', name: 'OrderSummary', label: '订单汇总' },
    ]
  },
  { path: '/order/auths', name: 'PlatformAuths', icon: '🔑', label: '平台授权', permission: 'platform_auth:view' },
  { path: '/order/orders', name: 'Orders', icon: '📋', label: '订单列表', permission: 'order:view' },
  { 
    icon: '📊', 
    label: '报表',
    permission: 'report:view',
    children: [
      { path: '/order/cashflow', name: 'CashFlow', label: '财务报告' },
      { path: '/order/settlement', name: 'Settlement', label: '结算报告' },
    ]
  },
  { 
    icon: '🏭', 
    label: '仓库',
    permission: 'warehouse:view',
    children: [
      { path: '/warehouse/list', name: 'WarehouseList', label: '仓库列表' },
      { path: '/warehouse/inventory', name: 'InventoryList', label: '库存明细' },
      { path: '/warehouse/stock-in', name: 'StockInOrders', label: '入库单' },
    ]
  },
  { 
    icon: '🚚', 
    label: '物流',
    permission: 'shipping:view',
    children: [
      { path: '/shipping/templates', name: 'ShippingTemplates', label: '运费模板' },
    ]
  },
  { path: '/users', name: 'UserManagement', icon: '👤', label: '用户管理', permission: 'user:view' },
]

// 展开的菜单
const expandedMenus = ref<string[]>(['报表'])

// 切换菜单展开
const toggleMenu = (label: string) => {
  const index = expandedMenus.value.indexOf(label)
  if (index > -1) {
    expandedMenus.value.splice(index, 1)
  } else {
    expandedMenus.value.push(label)
  }
}

// 检查菜单是否展开
const isMenuExpanded = (label: string) => expandedMenus.value.includes(label)

// 检查子菜单是否激活
const isChildActive = (item: any) => {
  if (!item.children) return false
  return item.children.some((child: any) => route.path === child.path)
}

// 计算显示的菜单项（根据用户权限过滤）
const menuItems = computed(() => {
  return allMenuItems.filter(item => {
    if (!item.permission) return true
    return userStore.hasPermission(item.permission)
  })
})

// 路由名称到标题的映射
const routeTitleMap: Record<string, string> = {
  'Dashboard': '首页',
  'PlatformAuths': '平台授权',
  'Orders': '订单列表',
  'OrderDetail': '订单详情',
  'CashFlow': '财务报告',
  'Settlement': '结算报告',
  'LocalProducts': '系统产品',
  'PlatformProducts': '平台产品',
  'OrderSummary': '订单汇总',
  'StockInOrders': '入库单',
  'StockInCreate': '新建入库单',
  'InventoryList': '库存明细',
  'WarehouseList': '仓库列表',
  'ShippingTemplates': '运费模板',
  'UserManagement': '用户管理',
}

// 不可关闭的标签
const unclosableTabs = ['Dashboard']

// 用户名首字母
const userInitial = computed(() => {
  return userStore.username?.charAt(0)?.toUpperCase() || 'U'
})

// 用户角色显示
const userRoleDisplay = computed(() => {
  const role = userStore.user?.role
  if (role === 'super_admin') return '超级管理员'
  if (role === 'admin') return '管理员'
  return '用户'
})

const isActive = (path: string) => route.path === path

// 导航并添加标签
const navigateTo = (path: string) => {
  router.push(path)
}

// 监听路由变化，自动添加标签
watch(() => route.fullPath, () => {
  if (route.name && typeof route.name === 'string') {
    const title = routeTitleMap[route.name] || (route.meta?.title as string) || route.name
    const tab: TabItem = {
      name: route.name,
      title: title,
      path: route.fullPath,
      closable: !unclosableTabs.includes(route.name)
    }
    tabsStore.addTab(tab)
  }
}, { immediate: true })

// 点击标签
const handleTabClick = (tab: TabItem) => {
  if (route.fullPath !== tab.path) {
    router.push(tab.path)
  }
}

// 关闭标签
const handleTabClose = (tab: TabItem, event: Event) => {
  event.stopPropagation()
  const redirectPath = tabsStore.closeTab(tab.name)
  if (redirectPath) {
    router.push(redirectPath)
  }
}

// 拖拽排序
const dragIndex = ref<number | null>(null)
const dragOverIndex = ref<number | null>(null)

const handleDragStart = (index: number, event: DragEvent) => {
  dragIndex.value = index
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', String(index))
  }
}

const handleDragOver = (index: number, event: DragEvent) => {
  event.preventDefault()
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'move'
  }
  dragOverIndex.value = index
}

const handleDragLeave = () => {
  dragOverIndex.value = null
}

const handleDrop = (index: number, event: DragEvent) => {
  event.preventDefault()
  if (dragIndex.value !== null && dragIndex.value !== index) {
    tabsStore.moveTab(dragIndex.value, index)
  }
  dragIndex.value = null
  dragOverIndex.value = null
}

const handleDragEnd = () => {
  dragIndex.value = null
  dragOverIndex.value = null
}

// 右键菜单
const contextMenuVisible = ref(false)
const contextMenuPosition = ref({ x: 0, y: 0 })
const contextMenuTab = ref<TabItem | null>(null)

// 显示右键菜单
const handleContextMenu = (tab: TabItem, event: MouseEvent) => {
  event.preventDefault()
  contextMenuTab.value = tab
  contextMenuPosition.value = { x: event.clientX, y: event.clientY }
  contextMenuVisible.value = true
  
  // 点击其他地方关闭菜单
  document.addEventListener('click', closeContextMenu)
}

// 关闭右键菜单
const closeContextMenu = () => {
  contextMenuVisible.value = false
  document.removeEventListener('click', closeContextMenu)
}

// 刷新当前页面
const handleRefreshTab = () => {
  if (contextMenuTab.value) {
    const tabName = contextMenuTab.value.name
    const path = contextMenuTab.value.path
    
    // 从缓存中移除，触发组件重新创建
    tabsStore.refreshTab(tabName)
    
    // 使用 nextTick 确保缓存已更新，然后导航
    nextTick(() => {
      router.replace('/redirect' + path).then(() => {
        // 导航完成后恢复缓存
        setTimeout(() => {
          tabsStore.restoreCache(tabName)
        }, 100)
      }).catch(() => {
        // 如果 redirect 路由失败，直接刷新页面
        router.go(0)
      })
    })
  }
  closeContextMenu()
}

// 固定/取消固定标签
const handlePinTab = () => {
  if (contextMenuTab.value) {
    const tabName = contextMenuTab.value.name
    if (unclosableTabs.includes(tabName)) {
      // 取消固定
      const index = unclosableTabs.indexOf(tabName)
      if (index > -1 && tabName !== 'Dashboard') {
        unclosableTabs.splice(index, 1)
      }
    } else {
      // 固定
      unclosableTabs.push(tabName)
    }
    // 更新标签的 closable 属性
    const tab = tabsStore.tabs.find(t => t.name === tabName)
    if (tab) {
      tab.closable = !unclosableTabs.includes(tabName)
    }
  }
  closeContextMenu()
}

// 关闭当前标签
const handleCloseCurrentTab = () => {
  if (contextMenuTab.value && contextMenuTab.value.closable) {
    const redirectPath = tabsStore.closeTab(contextMenuTab.value.name)
    if (redirectPath) {
      router.push(redirectPath)
    }
  }
  closeContextMenu()
}

// 关闭其他标签
const handleCloseOtherTabs = () => {
  if (contextMenuTab.value) {
    tabsStore.closeOtherTabs(contextMenuTab.value.name)
    router.push(contextMenuTab.value.path)
  }
  closeContextMenu()
}

// 判断标签是否已固定
const isTabPinned = computed(() => {
  return contextMenuTab.value ? unclosableTabs.includes(contextMenuTab.value.name) : false
})

const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

// 登出
const handleLogout = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要退出登录吗？',
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    userStore.logout()
  } catch {
    // 用户取消
  }
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
        <template v-for="item in menuItems" :key="item.path || item.label">
          <!-- 有子菜单的项 -->
          <template v-if="item.children">
            <button
              class="nav-item nav-parent"
              :class="{ active: isChildActive(item), expanded: isMenuExpanded(item.label) }"
              @click="toggleMenu(item.label)"
            >
              <span class="nav-icon">{{ item.icon }}</span>
              <span class="nav-label" v-show="!sidebarCollapsed">{{ item.label }}</span>
              <span class="nav-arrow" v-show="!sidebarCollapsed">{{ isMenuExpanded(item.label) ? '▾' : '▸' }}</span>
            </button>
            <div class="nav-children" v-show="isMenuExpanded(item.label) && !sidebarCollapsed">
              <button
                v-for="child in item.children"
                :key="child.path"
                class="nav-item nav-child"
                :class="{ active: isActive(child.path) }"
                @click="navigateTo(child.path)"
              >
                <span class="nav-label">{{ child.label }}</span>
              </button>
            </div>
          </template>
          <!-- 无子菜单的项 -->
          <button
            v-else
            class="nav-item"
            :class="{ active: isActive(item.path) }"
            @click="navigateTo(item.path)"
          >
            <span class="nav-icon">{{ item.icon }}</span>
            <span class="nav-label" v-show="!sidebarCollapsed">{{ item.label }}</span>
          </button>
        </template>
      </nav>
      
      <div class="sidebar-footer">
        <div class="user-info" v-show="!sidebarCollapsed">
          <div class="user-avatar">{{ userInitial }}</div>
          <div class="user-details">
            <div class="user-name">{{ userStore.username }}</div>
            <div class="user-role">{{ userRoleDisplay }}</div>
          </div>
          <button class="logout-btn" @click="handleLogout" title="退出登录">
            ⏻
          </button>
        </div>
        <button 
          v-show="sidebarCollapsed" 
          class="logout-btn-collapsed" 
          @click="handleLogout" 
          title="退出登录"
        >
          ⏻
        </button>
      </div>
    </aside>
    
    <!-- 主内容区 -->
    <main class="main-content">
      <header class="top-header">
        <!-- 标签栏 -->
        <div class="tabs-bar">
          <div 
            v-for="(tab, index) in tabsStore.tabs" 
            :key="tab.path"
            class="tab-item"
            :class="{ 
              active: tabsStore.activeTab === tab.name, 
              pinned: !tab.closable,
              dragging: dragIndex === index,
              'drag-over': dragOverIndex === index && dragIndex !== index
            }"
            draggable="true"
            @click="handleTabClick(tab)"
            @contextmenu="handleContextMenu(tab, $event)"
            @dragstart="handleDragStart(index, $event)"
            @dragover="handleDragOver(index, $event)"
            @dragleave="handleDragLeave"
            @drop="handleDrop(index, $event)"
            @dragend="handleDragEnd"
          >
            <span v-if="!tab.closable" class="tab-pin">📌</span>
            <span class="tab-title">{{ tab.title }}</span>
            <span 
              v-if="tab.closable" 
              class="tab-close"
              @click="handleTabClose(tab, $event)"
            >
              <el-icon :size="12"><Close /></el-icon>
            </span>
          </div>
        </div>
        
        <!-- 右键菜单 -->
        <Teleport to="body">
          <div 
            v-if="contextMenuVisible" 
            class="tab-context-menu"
            :style="{ left: contextMenuPosition.x + 'px', top: contextMenuPosition.y + 'px' }"
          >
            <div class="context-menu-item" @click="handleRefreshTab">
              刷新当前页面
            </div>
            <div class="context-menu-item" @click="handlePinTab">
              {{ isTabPinned ? '取消固定' : '固定当前页面' }}
            </div>
            <div 
              class="context-menu-item" 
              :class="{ disabled: !contextMenuTab?.closable }"
              @click="handleCloseCurrentTab"
            >
              关闭当前页面
            </div>
            <div class="context-menu-item" @click="handleCloseOtherTabs">
              关闭其他页面
            </div>
          </div>
        </Teleport>
        
        <div class="header-actions">
          <button class="theme-toggle" @click="themeStore.toggleTheme" :title="themeStore.isDark ? '切换到浅色模式' : '切换到深色模式'">
            {{ themeStore.isDark ? '☀️' : '🌙' }}
          </button>
          <span class="user-greeting">{{ userStore.companyName ? userStore.companyName + ' - ' : '' }}{{ userStore.username }}</span>
        </div>
      </header>
      
      <div class="content-wrapper">
        <RouterView v-slot="{ Component }">
          <keep-alive :include="tabsStore.cachedViews">
            <component :is="Component" />
          </keep-alive>
        </RouterView>
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

.nav-parent {
  justify-content: flex-start;
  
  .nav-arrow {
    margin-left: auto;
    font-size: 10px;
    color: var(--text-muted);
    transition: transform var(--transition-fast);
  }
  
  &.expanded .nav-arrow {
    color: var(--color-primary);
  }
}

.nav-children {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-left: 20px;
  padding-left: 16px;
  border-left: 1px solid var(--border-color);
}

.nav-child {
  padding: 10px 16px;
  font-size: 13px;
  
  &.active {
    background: rgba(0, 212, 255, 0.1);
    color: var(--color-primary);
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
  flex-shrink: 0;
}

.user-details {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-role {
  font-size: 12px;
  color: var(--text-muted);
}

.logout-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 16px;
  transition: all var(--transition-fast);
  flex-shrink: 0;
  
  &:hover {
    background: rgba(255, 77, 79, 0.1);
    color: #ff4d4f;
  }
}

.logout-btn-collapsed {
  width: 100%;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 18px;
  transition: all var(--transition-fast);
  
  &:hover {
    background: rgba(255, 77, 79, 0.1);
    color: #ff4d4f;
  }
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

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.theme-toggle {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  border-radius: 50%;
  font-size: 18px;
  transition: all var(--transition-fast);
  
  &:hover {
    background: var(--bg-hover);
    transform: rotate(15deg);
  }
}

.user-greeting {
  font-size: 14px;
  color: var(--text-secondary);
}

.content-wrapper {
  padding: 32px;
}

.tabs-bar {
  display: flex;
  align-items: center;
  gap: 4px;
  flex: 1;
  overflow-x: auto;
  padding-right: 16px;
  
  &::-webkit-scrollbar {
    height: 4px;
  }
  
  &::-webkit-scrollbar-thumb {
    background: var(--border-color);
    border-radius: 2px;
  }
}

.tab-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
  font-size: 13px;
  color: var(--text-secondary);
  border: 1px solid transparent;
  
  &:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
    
    .tab-close {
      opacity: 1;
    }
  }
  
  &.active {
    background: var(--bg-card);
    color: var(--color-primary);
    border-color: var(--color-primary);
    
    .tab-close {
      opacity: 1;
    }
  }
}

.tab-title {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tab-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  opacity: 0;
  transition: all var(--transition-fast);
  
  &:hover {
    background: rgba(255, 77, 79, 0.2);
    color: #ff4d4f;
  }
}

.tab-pin {
  font-size: 10px;
  margin-right: -4px;
}

.tab-item.pinned {
  .tab-title {
    font-weight: 500;
  }
}

.tab-item.dragging {
  opacity: 0.5;
  cursor: grabbing;
}

.tab-item.drag-over {
  border-left: 2px solid var(--color-primary);
  margin-left: -2px;
}

.tab-item {
  cursor: grab;
  
  &:active {
    cursor: grabbing;
  }
}

.tab-context-menu {
  position: fixed;
  z-index: 9999;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  padding: 4px 0;
  min-width: 140px;
}

.context-menu-item {
  padding: 8px 16px;
  font-size: 13px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all var(--transition-fast);
  
  &:hover {
    background: var(--bg-hover);
    color: var(--color-primary);
  }
  
  &.disabled {
    color: var(--text-muted);
    cursor: not-allowed;
    
    &:hover {
      background: transparent;
      color: var(--text-muted);
    }
  }
}
</style>
