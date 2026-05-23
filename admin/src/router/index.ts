import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘', icon: 'Odometer' },
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/users/UserList.vue'),
        meta: { title: '用户管理', icon: 'User' },
      },
      {
        path: 'categories',
        name: 'Categories',
        component: () => import('@/views/categories/CategoryList.vue'),
        meta: { title: '分类管理', icon: 'Menu' },
      },
      {
        path: 'banners',
        name: 'Banners',
        component: () => import('@/views/banners/BannerList.vue'),
        meta: { title: 'Banner管理', icon: 'Picture' },
      },
      {
        path: 'products',
        name: 'Products',
        component: () => import('@/views/products/ProductList.vue'),
        meta: { title: '商品管理', icon: 'Goods' },
      },
      {
        path: 'skus',
        name: 'Skus',
        component: () => import('@/views/skus/SkuList.vue'),
        meta: { title: 'SKU管理', icon: 'Box' },
      },
      {
        path: 'orders',
        name: 'Orders',
        component: () => import('@/views/orders/OrderList.vue'),
        meta: { title: '订单管理', icon: 'ShoppingBag' },
      },
      {
        path: 'carts',
        name: 'Carts',
        component: () => import('@/views/carts/CartList.vue'),
        meta: { title: '购物车管理', icon: 'ShoppingCart' },
      },
      {
        path: 'payments',
        name: 'Payments',
        component: () => import('@/views/payments/PaymentList.vue'),
        meta: { title: '支付管理', icon: 'CreditCard' },
      },
      {
        path: 'inventory',
        name: 'Inventory',
        component: () => import('@/views/inventory/InventoryList.vue'),
        meta: { title: '库存管理', icon: 'Box' },
      },
      {
        path: 'promotions',
        name: 'Promotions',
        component: () => import('@/views/promotions/PromotionList.vue'),
        meta: { title: '营销管理', icon: 'Tickets' },
      },
      {
        path: 'seckill-activities',
        name: 'SeckillActivities',
        component: () => import('@/views/seckill/SeckillActivityList.vue'),
        meta: { title: '秒杀活动', icon: 'Lightning' },
      },
      {
        path: 'reviews',
        name: 'Reviews',
        component: () => import('@/views/reviews/ReviewList.vue'),
        meta: { title: '评价管理', icon: 'ChatLineRound' },
      },
      {
        path: 'logistics',
        name: 'Logistics',
        component: () => import('@/views/logistics/LogisticsList.vue'),
        meta: { title: '物流管理', icon: 'Truck' },
      },
      {
        path: 'messages',
        name: 'Messages',
        component: () => import('@/views/messages/MessageList.vue'),
        meta: { title: '消息管理', icon: 'Message' },
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
  
  if (to.path === '/login') {
    // 如果已登录，跳转到首页
    if (userStore.token) {
      next('/')
    } else {
      next()
    }
  } else {
    // 需要登录且为管理员
    if (!userStore.token) {
      next('/login')
    } else if (userStore.userInfo && Number((userStore.userInfo as any).is_admin ?? (userStore.userInfo as any).isAdmin ?? 0) !== 1) {
      userStore.logout()
      ElMessage.error('您不是管理员，无法访问后台')
      next('/login')
    } else {
      next()
    }
  }
})

export default router

