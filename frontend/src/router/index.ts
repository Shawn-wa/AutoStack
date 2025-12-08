import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('@/pages/Dashboard.vue'),
        meta: { title: '控制台' }
      },
      {
        path: 'projects',
        name: 'Projects',
        component: () => import('@/pages/Projects.vue'),
        meta: { title: '项目管理' }
      },
      {
        path: 'deployments',
        name: 'Deployments',
        component: () => import('@/pages/Deployments.vue'),
        meta: { title: '部署管理' }
      },
      {
        path: 'templates',
        name: 'Templates',
        component: () => import('@/pages/Templates.vue'),
        meta: { title: '模板市场' }
      },
    ]
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/pages/Login.vue'),
    meta: { title: '登录' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, _from, next) => {
  document.title = `${to.meta.title || 'AutoStack'} - AutoStack`
  next()
})

export default router

