import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { storage } from '@/utils/storage'

const routes: RouteRecordRaw[] = [
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
