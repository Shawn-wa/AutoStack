import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface TabItem {
  name: string       // 路由 name
  title: string      // 标签显示标题
  path: string       // 完整路径（包含 query）
  closable: boolean  // 是否可关闭
}

export const useTabsStore = defineStore('tabs', () => {
  // 已打开的标签列表
  const tabs = ref<TabItem[]>([
    { name: 'Dashboard', title: '控制台', path: '/', closable: false }
  ])
  
  // 当前激活的标签
  const activeTab = ref('Dashboard')
  
  // 需要排除缓存的组件（用于刷新）
  const excludeCache = ref<string[]>([])
  
  // 缓存的组件名称列表（排除需要刷新的）
  const cachedViews = computed(() => 
    tabs.value.map(tab => tab.name).filter(name => !excludeCache.value.includes(name))
  )
  
  // 刷新标签（临时从缓存中移除）
  const refreshTab = (name: string) => {
    excludeCache.value.push(name)
  }
  
  // 恢复缓存
  const restoreCache = (name: string) => {
    const index = excludeCache.value.indexOf(name)
    if (index > -1) {
      excludeCache.value.splice(index, 1)
    }
  }
  
  // 添加标签
  const addTab = (tab: TabItem) => {
    // 检查是否已存在（按 name 和 path 判断）
    const existingIndex = tabs.value.findIndex(t => t.name === tab.name && t.path === tab.path)
    if (existingIndex === -1) {
      // 检查是否有相同 name 但不同 path 的标签（如订单详情）
      const sameNameIndex = tabs.value.findIndex(t => t.name === tab.name)
      if (sameNameIndex !== -1 && tab.name !== 'Orders') {
        // 替换现有标签（除了订单列表，因为它的 query 可能不同）
        tabs.value[sameNameIndex] = tab
      } else {
        tabs.value.push(tab)
      }
    }
    activeTab.value = tab.name
  }
  
  // 关闭标签
  const closeTab = (name: string) => {
    const index = tabs.value.findIndex(tab => tab.name === name)
    if (index === -1) return
    
    const tab = tabs.value[index]
    if (!tab.closable) return
    
    tabs.value.splice(index, 1)
    
    // 如果关闭的是当前激活的标签，激活前一个或后一个
    if (activeTab.value === name) {
      const newIndex = Math.min(index, tabs.value.length - 1)
      activeTab.value = tabs.value[newIndex]?.name || 'Dashboard'
      return tabs.value[newIndex]?.path || '/dashboard'
    }
    return null
  }
  
  // 关闭其他标签
  const closeOtherTabs = (name: string) => {
    tabs.value = tabs.value.filter(tab => !tab.closable || tab.name === name)
    activeTab.value = name
  }
  
  // 关闭所有可关闭的标签
  const closeAllTabs = () => {
    tabs.value = tabs.value.filter(tab => !tab.closable)
    activeTab.value = tabs.value[0]?.name || 'Dashboard'
    return tabs.value[0]?.path || '/dashboard'
  }
  
  // 设置当前激活标签
  const setActiveTab = (name: string) => {
    activeTab.value = name
  }
  
  // 根据路由名称获取标签
  const getTabByName = (name: string) => {
    return tabs.value.find(tab => tab.name === name)
  }
  
  // 移动标签位置
  const moveTab = (fromIndex: number, toIndex: number) => {
    if (fromIndex === toIndex) return
    if (fromIndex < 0 || fromIndex >= tabs.value.length) return
    if (toIndex < 0 || toIndex >= tabs.value.length) return
    
    const tab = tabs.value.splice(fromIndex, 1)[0]
    tabs.value.splice(toIndex, 0, tab)
  }
  
  return {
    tabs,
    activeTab,
    cachedViews,
    addTab,
    closeTab,
    closeOtherTabs,
    closeAllTabs,
    setActiveTab,
    getTabByName,
    refreshTab,
    restoreCache,
    moveTab
  }
})

