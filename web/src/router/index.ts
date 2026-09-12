import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/Register.vue'),
    meta: { title: '注册' },
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    children: [
      {
        path: '',
        name: 'Home',
        component: () => import('@/views/Home.vue'),
        meta: { title: '首页' },
      },
      {
        path: 'products',
        name: 'Products',
        component: () => import('@/views/products/ProductList.vue'),
        meta: { title: '商品列表' },
      },
      {
        path: 'products/:id',
        name: 'ProductDetail',
        component: () => import('@/views/products/ProductDetail.vue'),
        meta: { title: '商品详情' },
      },
      {
        path: 'seckill',
        name: 'Seckill',
        component: () => import('@/views/seckill/SeckillList.vue'),
        meta: { title: '秒杀' },
      },
      {
        path: 'seckill/:id',
        name: 'SeckillDetail',
        component: () => import('@/views/seckill/SeckillDetail.vue'),
        meta: { title: '秒杀详情' },
      },
      {
        path: 'cart',
        name: 'Cart',
        component: () => import('@/views/cart/Cart.vue'),
        meta: { title: '购物车', requiresAuth: true },
      },
      {
        path: 'orders',
        name: 'Orders',
        component: () => import('@/views/orders/OrderList.vue'),
        meta: { title: '我的订单', requiresAuth: true },
      },
      {
        path: 'orders/create',
        name: 'OrderCreate',
        component: () => import('@/views/orders/OrderCreate.vue'),
        meta: { title: '创建订单', requiresAuth: true },
      },
      {
        path: 'orders/:id',
        name: 'OrderDetail',
        component: () => import('@/views/orders/OrderDetail.vue'),
        meta: { title: '订单详情', requiresAuth: true },
      },
      {
        path: 'pay/:id',
        name: 'Pay',
        component: () => import('@/views/orders/Pay.vue'),
        meta: { title: '收银台', requiresAuth: true },
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/profile/Profile.vue'),
        meta: { title: '个人中心', requiresAuth: true },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  // 如果访问登录/注册页且已登录，跳转到首页
  if ((to.path === '/login' || to.path === '/register') && userStore.token) {
    next('/')
    return
  }
  
  // 需要登录的页面
  if (to.meta.requiresAuth && !userStore.token) {
    next('/login')
    return
  }
  
  next()
})

export default router

