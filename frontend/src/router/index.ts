import { createRouter, createWebHistory, useRoute, useRouter } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { storage } from '@/utils/storage'
import type { UserInfo } from '@/modules/auth/api'

const routes: RouteRecordRaw[] = [
  {
    // 用于刷新页面的中转路由
    path: '/redirect/:path(.*)',
    name: 'Redirect',
    component: {
      setup() {
        const route = useRoute()
        const router = useRouter()
        const path = '/' + (route.params.path as string)
        router.replace(path)
        return () => null
      }
    }
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('@/modules/dashboard/pages/Dashboard.vue'),
        meta: { title: '控制台', requiredPermission: 'dashboard:view' }
      },
      {
        path: 'projects',
        name: 'Projects',
        component: () => import('@/modules/project/pages/Projects.vue'),
        meta: { title: '项目管理' }
      },
      {
        path: 'deployments',
        name: 'Deployments',
        component: () => import('@/modules/deployment/pages/Deployments.vue'),
        meta: { title: '部署管理' }
      },
      {
        path: 'templates',
        name: 'Templates',
        component: () => import('@/modules/template/pages/Templates.vue'),
        meta: { title: '模板市场' }
      },
      {
        path: 'users',
        name: 'UserManagement',
        component: () => import('@/modules/user/pages/UserManagement.vue'),
        meta: { title: '用户管理', requiredPermission: 'user:view' }
      },
      {
        path: 'product/products',
        name: 'LocalProducts',
        component: () => import('@/modules/product/pages/LocalProducts.vue'),
        meta: { title: '系统产品', requiredPermission: 'product:view' }
      },
      {
        path: 'product/platform-products',
        name: 'PlatformProducts',
        component: () => import('@/modules/product/pages/PlatformProducts.vue'),
        meta: { title: '平台产品', requiredPermission: 'product:view' }
      },
      {
        path: 'product/summary',
        name: 'OrderSummary',
        component: () => import('@/modules/product/pages/OrderSummary.vue'),
        meta: { title: '订单汇总', requiredPermission: 'product:view' }
      },
      {
        path: 'order/auths',
        name: 'PlatformAuths',
        component: () => import('@/modules/order/pages/PlatformAuths.vue'),
        meta: { title: '平台授权', requiredPermission: 'platform_auth:view' }
      },
      {
        path: 'order/orders',
        name: 'Orders',
        component: () => import('@/modules/order/pages/Orders.vue'),
        meta: { title: '订单列表', requiredPermission: 'order:view' }
      },
      {
        path: 'order/orders/:id',
        name: 'OrderDetail',
        component: () => import('@/modules/order/pages/OrderDetail.vue'),
        meta: { title: '订单详情', requiredPermission: 'order:view' }
      },
      {
        path: 'order/cashflow',
        name: 'CashFlow',
        component: () => import('@/modules/order/pages/CashFlow.vue'),
        meta: { title: '财务报告', requiredPermission: 'report:view' }
      },
      {
        path: 'order/settlement',
        name: 'Settlement',
        component: () => import('@/modules/order/pages/Settlement.vue'),
        meta: { title: '结算报告', requiredPermission: 'report:view' }
      },
      {
        path: 'warehouse/stock-in',
        name: 'StockInOrders',
        component: () => import('@/modules/product/pages/StockInOrders.vue'),
        meta: { title: '入库单', requiredPermission: 'warehouse:view' }
      },
      {
        path: 'warehouse/stock-in/create',
        name: 'StockInCreate',
        component: () => import('@/modules/product/pages/StockInCreate.vue'),
        meta: { title: '新建入库单', requiredPermission: 'warehouse:view' }
      },
      {
        path: 'warehouse/inventory',
        name: 'InventoryList',
        component: () => import('@/modules/product/pages/InventoryList.vue'),
        meta: { title: '库存明细', requiredPermission: 'warehouse:view' }
      },
      {
        path: 'warehouse/list',
        name: 'WarehouseList',
        component: () => import('@/modules/product/pages/WarehouseList.vue'),
        meta: { title: '仓库列表', requiredPermission: 'warehouse:view' }
      },
      {
        path: 'shipping/templates',
        name: 'ShippingTemplates',
        component: () => import('@/modules/shipping/pages/ShippingTemplates.vue'),
        meta: { title: '运费模板', requiredPermission: 'shipping:view' }
      }
    ]
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/modules/auth/pages/Login.vue'),
    meta: { title: '登录', guest: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 检查用户是否拥有指定权限
function checkPermission(user: UserInfo | null, permission: string): boolean {
  if (!user) return false
  if (user.role === 'super_admin') return true
  return (user.permissions || []).includes(permission)
}

// 路由守卫
router.beforeEach((to, _from, next) => {
  // 设置页面标题
  document.title = `${to.meta.title || 'AutoStack'} - AutoStack`

  const token = storage.get('token')
  const user = storage.get<UserInfo>('user')

  // 需要登录的页面
  if (to.matched.some(record => record.meta.requiresAuth)) {
    if (!token) {
      next({ path: '/login', query: { redirect: to.fullPath } })
      return
    }
  }

  // 需要特定模块权限的页面
  const requiredPermission = to.meta.requiredPermission as string | undefined
  if (requiredPermission) {
    if (!checkPermission(user, requiredPermission)) {
      // 无权限时跳转到首页（若首页也无权限则停留在登录页）
      if (to.path !== '/') {
        next({ path: '/' })
      } else {
        next()
      }
      return
    }
  }

  // 已登录用户访问登录页面
  if (to.meta.guest && token) {
    next({ path: '/' })
    return
  }

  next()
})

export default router
