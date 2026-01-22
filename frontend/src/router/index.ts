import { createRouter, createWebHistory, useRoute, useRouter } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { storage } from '@/utils/storage'

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
        meta: { title: '控制台' }
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
        meta: { title: '用户管理', requiresAdmin: true }
      },
      {
        path: 'product/products',
        name: 'LocalProducts',
        component: () => import('@/modules/product/pages/LocalProducts.vue'),
        meta: { title: '本地产品' }
      },
      {
        path: 'product/platform-products',
        name: 'PlatformProducts',
        component: () => import('@/modules/product/pages/PlatformProducts.vue'),
        meta: { title: '平台产品' }
      },
      {
        path: 'product/summary',
        name: 'OrderSummary',
        component: () => import('@/modules/product/pages/OrderSummary.vue'),
        meta: { title: '订单汇总' }
      },
      {
        path: 'order/auths',
        name: 'PlatformAuths',
        component: () => import('@/modules/order/pages/PlatformAuths.vue'),
        meta: { title: '平台授权' }
      },
      {
        path: 'order/orders',
        name: 'Orders',
        component: () => import('@/modules/order/pages/Orders.vue'),
        meta: { title: '订单列表' }
      },
      {
        path: 'order/orders/:id',
        name: 'OrderDetail',
        component: () => import('@/modules/order/pages/OrderDetail.vue'),
        meta: { title: '订单详情' }
      },
      {
        path: 'order/cashflow',
        name: 'CashFlow',
        component: () => import('@/modules/order/pages/CashFlow.vue'),
        meta: { title: '财务报告' }
      },
      {
        path: 'order/settlement',
        name: 'Settlement',
        component: () => import('@/modules/order/pages/Settlement.vue'),
        meta: { title: '结算报告' }
      },
      {
        path: 'warehouse/stock-in',
        name: 'StockInOrders',
        component: () => import('@/modules/product/pages/StockInOrders.vue'),
        meta: { title: '入库单' }
      },
      {
        path: 'warehouse/inventory',
        name: 'InventoryList',
        component: () => import('@/modules/product/pages/InventoryList.vue'),
        meta: { title: '库存明细' }
      },
      {
        path: 'warehouse/list',
        name: 'WarehouseList',
        component: () => import('@/modules/product/pages/WarehouseList.vue'),
        meta: { title: '仓库列表' }
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

// 路由守卫
router.beforeEach((to, _from, next) => {
  // 设置页面标题
  document.title = `${to.meta.title || 'AutoStack'} - AutoStack`

  const token = storage.get('token')
  const user = storage.get<{ role: string }>('user')

  // 需要登录的页面
  if (to.matched.some(record => record.meta.requiresAuth)) {
    if (!token) {
      next({ path: '/login', query: { redirect: to.fullPath } })
      return
    }
  }

  // 需要管理员权限的页面
  if (to.meta.requiresAdmin) {
    if (!user || (user.role !== 'admin' && user.role !== 'super_admin')) {
      next({ path: '/' })
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
