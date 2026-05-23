<template>
  <div class="main-layout" :class="{ 'is-home': isHome }">
    <!-- ======== Header ======== -->
    <header class="header">
      <div class="header-inner">
        <!-- Logo -->
        <div class="logo" @click="$router.push('/')">
          <img src="/logo-forge.png" alt="GoForge" class="logo-img" />
        </div>

        <!-- 导航链接 -->
        <nav class="nav-links">
          <router-link to="/" class="nav-item" exact-active-class="active">首页</router-link>
          <router-link to="/seckill" class="nav-item" active-class="active">秒杀</router-link>
          <router-link
            v-for="cat in mainCategories.slice(0, 2)"
            :key="cat.id"
            :to="`/products?category_id=${cat.id}`"
            class="nav-item"
          >{{ cat.name }}</router-link>
          <router-link to="/products" class="nav-item" active-class="active">全部商品</router-link>
        </nav>

        <!-- 右侧操作区 -->
        <div class="header-actions">
          <div class="search-box">
            <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
            <input v-model="searchKeyword" type="text" placeholder="搜索商品..." class="search-input" @keyup.enter="handleSearch" />
          </div>

          <template v-if="userStore.token">
            <router-link to="/cart" class="cart-btn">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="cart-icon">
                <circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/>
                <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>
              </svg>
              <span v-if="cartCount > 0" class="cart-badge">{{ cartCount }}</span>
            </router-link>

            <div class="avatar-dropdown" @click.stop="showMenu = !showMenu">
              <div class="avatar-sm">
                <img v-if="avatarUrl" :src="avatarUrl" />
                <span v-else>{{ initial }}</span>
              </div>
              <div v-if="showMenu" class="dropdown-menu" @click.stop>
                <router-link to="/profile" class="dropdown-item" @click="showMenu=false">个人中心</router-link>
                <router-link to="/orders" class="dropdown-item" @click="showMenu=false">我的订单</router-link>
                <div class="dropdown-item logout" @click="handleLogout">退出登录</div>
              </div>
            </div>
          </template>
          <template v-else>
            <router-link to="/login" class="btn-login">登录</router-link>
          </template>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="main-content">
      <router-view />
    </main>

    <!-- Footer -->
    <footer class="footer">
      <div class="footer-inner">
        <!-- Top: Logo + Links -->
        <div class="footer-grid">
          <!-- Brand Column -->
          <div class="footer-brand">
            <div class="footer-logo" @click="$router.push('/')">
              <img src="/logo-forge.png" alt="GoForge" />
            </div>
            <p class="footer-slogan">用 Go 锻造极致电商体验</p>
            <div class="footer-social">
              <a href="#" title="GitHub" class="social-icon"><svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61-.546-1.385-1.335-1.755-1.335-1.755-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.605-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 21.795 24 17.295 24 12 24 5.37 18.63 0 12 0z"/></svg></a>
              <a href="#" title="Twitter" class="social-icon"><svg viewBox="0 0 24 24" fill="currentColor"><path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/></svg></a>
            </div>
          </div>
          <!-- Link Columns -->
          <div class="footer-col">
            <h4>购物指南</h4>
            <a href="#">购物流程</a><a href="#">会员介绍</a><a href="#">生活旅行</a><a href="#">常见问题</a>
          </div>
          <div class="footer-col">
            <h4>配送方式</h4>
            <a href="#">免费配送</a><a href="#">海外配送</a><a href="#">211限时达</a><a href="#">EMS</a>
          </div>
          <div class="footer-col">
            <h4>支付方式</h4>
            <a href="#">在线支付</a><a href="#">分期付款</a><a href="#">货到付款</a><a href="#">邮局汇款</a>
          </div>
          <div class="footer-col">
            <h4>售后服务</h4>
            <a href="#">售后政策</a><a href="#">价格保护</a><a href="#">退款说明</a><a href="#">退换货</a>
          </div>
        </div>

        <!-- Bottom -->
        <div class="footer-bottom">
          <div class="footer-bottom-left">
            <span>&copy; 2026 GoForge. All rights reserved.</span>
            <span class="sep">·</span>
            <a href="#">隐私政策</a>
            <span class="sep">·</span>
            <a href="#">服务条款</a>
          </div>
          <div class="footer-bottom-right">
            <span>ICP 备 2026XXXXXXXX 号</span>
          </div>
        </div>
      </div>

      <!-- Back to Top -->
      <button class="back-to-top" :class="{visible:showBackTop}" @click="scrollToTop" title="返回顶部">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="18 15 12 9 6 15"/></svg>
      </button>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useCartStore } from '@/stores/cart'
import { getCategoryTree } from '@/api/category'
import type { Category } from '@/api/category'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()
const searchKeyword = ref('')
const mainCategories = ref<Category[]>([])
const showMenu = ref(false)

// Safari 不支持 ?. 顶层调用，用 computed 延迟访问
const isHome = computed(() => route.path === '/')

const resolveUrl = (url: string) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  if (url.startsWith('/')) return 'http://localhost:8080' + url
  return url
}
const avatarUrl = computed(() => resolveUrl(userStore.userInfo?.avatar || ''))
const initial = computed(() => (userStore.userInfo?.nickname || userStore.userInfo?.username || 'U')[0])
const cartCount = computed(() => cartStore.totalCount)

const showBackTop = ref(false)
const scrollToTop = () => window.scrollTo({top:0,behavior:'smooth'})
const onScroll = () => { showBackTop.value = window.scrollY > 400 }

const handleSearch = () => {
  if (searchKeyword.value.trim()) router.push({ path: '/products', query: { keyword: searchKeyword.value } })
}
const handleLogout = () => {
  userStore.logout(); cartStore.clearCart(); showMenu.value = false; router.push('/')
}

const fetchCategories = async () => {
  try {
    const res = await getCategoryTree({ status: -1 }) as any
    if (res.code === 0 && res.data) mainCategories.value = res.data
  } catch { /* ignore */ }
}

onMounted(async () => {
  fetchCategories()
  if (userStore.token) {
    if (!userStore.userInfo) await userStore.fetchUserInfo()
    cartStore.fetchCart()
  }
  document.addEventListener('click', () => { showMenu.value = false })
  window.addEventListener('scroll', onScroll)
})

onUnmounted(() => {
  window.removeEventListener('scroll', onScroll)
})
</script>

<style>
/* 全局暗色主题 */
html, body { margin: 0; padding: 0; background: #0A0F1C; }
</style>

<style scoped>
.main-layout { --bg: #0A0F1C; --accent: #00F5FF; --accent-dim: rgba(0,245,255,0.12); --text: #EDF0F5; --text-dim: #8890A5; --border: rgba(255,255,255,0.06); --radius-sm: 8px; min-height: 100vh; display: flex; flex-direction: column; background: var(--bg); color: var(--text); font-family: 'Inter','PingFang SC','SF Pro Display',-apple-system,sans-serif; }

/* ======== Header ======== */
.header { position: sticky; top: 0; z-index: 100; backdrop-filter: blur(20px) saturate(180%); -webkit-backdrop-filter: blur(20px) saturate(180%); background: rgba(10,15,28,0.78); border-bottom: 1px solid var(--border); }
.header-inner { max-width: 1280px; margin: 0 auto; padding: 0 24px; height: 64px; display: flex; align-items: center; gap: 32px; }
.logo { cursor: pointer; flex-shrink: 0; }
.logo-img { height: 32px; width: auto; }
.nav-links { display: flex; align-items: center; gap: 4px; flex: 1; }
.nav-item { color: var(--text-dim); text-decoration: none; font-size: 14px; font-weight: 500; padding: 8px 16px; border-radius: var(--radius-sm); transition: all .2s; white-space: nowrap; }
.nav-item:hover { color: var(--text); background: rgba(255,255,255,0.04); }
.nav-item.active { color: var(--accent); }
.header-actions { display: flex; align-items: center; gap: 16px; flex-shrink: 0; }

/* Search */
.search-box { position: relative; display: flex; align-items: center; }
.search-icon { position: absolute; left: 12px; width: 16px; height: 16px; color: var(--text-dim); pointer-events: none; }
.search-input { background: rgba(255,255,255,0.05); border: 1px solid transparent; border-radius: 100px; padding: 8px 16px 8px 36px; color: var(--text); font-size: 13px; width: 200px; outline: none; transition: all .25s; }
.search-input::placeholder { color: var(--text-dim); }
.search-input:focus { background: rgba(255,255,255,0.08); border-color: var(--accent); box-shadow: 0 0 0 3px var(--accent-dim); width: 260px; }

/* Cart */
.cart-btn { position: relative; display: flex; align-items: center; padding: 8px; border-radius: var(--radius-sm); color: var(--text-dim); transition: all .2s; }
.cart-btn:hover { color: var(--accent); background: rgba(255,255,255,0.04); }
.cart-icon { width: 22px; height: 22px; }
.cart-badge { position: absolute; top: 2px; right: 0; background: var(--accent); color: var(--bg); font-size: 10px; font-weight: 700; min-width: 18px; height: 18px; border-radius: 9px; display: flex; align-items: center; justify-content: center; padding: 0 4px; }

/* Avatar */
.avatar-dropdown { position: relative; cursor: pointer; }
.avatar-sm { width: 34px; height: 34px; border-radius: 50%; overflow: hidden; background: var(--accent-dim); display: flex; align-items: center; justify-content: center; font-weight: 600; font-size: 14px; color: var(--accent); border: 2px solid transparent; transition: border-color .2s; }
.avatar-sm img { width: 100%; height: 100%; object-fit: cover; }
.avatar-dropdown:hover .avatar-sm { border-color: var(--accent); }
.dropdown-menu { position: absolute; top: calc(100% + 8px); right: 0; min-width: 140px; background: rgba(20,25,42,0.97); backdrop-filter: blur(16px); border: 1px solid var(--border); border-radius: 12px; padding: 6px; box-shadow: 0 12px 40px rgba(0,0,0,0.5); }
.dropdown-item { display: block; padding: 10px 14px; border-radius: var(--radius-sm); color: var(--text); text-decoration: none; font-size: 13px; transition: all .15s; cursor: pointer; }
.dropdown-item:hover { background: var(--accent-dim); color: var(--accent); }
.dropdown-item.logout { color: #F87171; }
.dropdown-item.logout:hover { background: rgba(248,113,113,0.1); }

/* Login */
.btn-login { display: inline-flex; align-items: center; padding: 8px 20px; border-radius: 100px; background: var(--accent); color: var(--bg); font-size: 13px; font-weight: 600; text-decoration: none; transition: all .2s; }
.btn-login:hover { box-shadow: 0 0 24px rgba(0,245,255,0.3); transform: translateY(-1px); }

/* Main Content */
.main-content { flex: 1; }
.is-home .main-content { background: var(--bg); }

/* ======== Footer ======== */
.footer { background: linear-gradient(180deg, #0A0F1C 0%, #060912 100%); border-top: 1px solid var(--border); margin-top: 100px; position: relative; }
.footer-inner { max-width: 1280px; margin: 0 auto; padding: 64px 24px 32px; }

/* Grid */
.footer-grid { display: grid; grid-template-columns: 2fr 1fr 1fr 1fr 1fr; gap: 40px; margin-bottom: 48px; }

/* Brand */
.footer-brand { display: flex; flex-direction: column; gap: 16px; }
.footer-logo { cursor: pointer; }
.footer-logo img { height: 30px; width: auto; }
.footer-slogan { font-size: 13px; color: var(--text-dim); margin: 0; line-height: 1.6; max-width: 240px; }
.footer-social { display: flex; gap: 10px; margin-top: 4px; }
.social-icon { width: 36px; height: 36px; border-radius: 50%; border: 1px solid var(--border); display: flex; align-items: center; justify-content: center; color: var(--text-dim); transition: all .2s; }
.social-icon svg { width: 16px; height: 16px; }
.social-icon:hover { color: var(--accent); border-color: var(--accent); background: var(--accent-dim); }

/* Link Columns */
.footer-col h4 { color: var(--text); font-size: 13px; font-weight: 600; margin: 0 0 18px; letter-spacing: .02em; text-transform: uppercase; }
.footer-col a { display: block; color: var(--text-dim); text-decoration: none; font-size: 13px; margin-bottom: 12px; transition: all .15s; width: fit-content; }
.footer-col a:hover { color: var(--accent); padding-left: 4px; }

/* Bottom */
.footer-bottom { border-top: 1px solid var(--border); padding-top: 24px; display: flex; justify-content: space-between; align-items: center; color: var(--text-dim); font-size: 12px; flex-wrap: wrap; gap: 8px; }
.footer-bottom a { color: var(--text-dim); text-decoration: none; transition: color .15s; }
.footer-bottom a:hover { color: var(--accent); }
.sep { margin: 0 6px; opacity: .4; }

/* Back to Top */
.back-to-top { position: absolute; right: 24px; bottom: 24px; width: 44px; height: 44px; border-radius: 50%; border: 1px solid var(--border); background: var(--card-bg); color: var(--text-dim); cursor: pointer; display: flex; align-items: center; justify-content: center; opacity: 0; visibility: hidden; transition: all .3s; }
.back-to-top svg { width: 20px; height: 20px; }
.back-to-top.visible { opacity: 1; visibility: visible; }
.back-to-top:hover { color: var(--accent); border-color: var(--accent); background: var(--accent-dim); }

@media (max-width:900px) { .footer-grid { grid-template-columns: repeat(2,1fr); } .footer-brand { grid-column: 1/-1; } }
@media (max-width:768px) { .nav-links { display: none; } .search-input { width: 140px; } .search-input:focus { width: 160px; } .footer-bottom { flex-direction: column; text-align: center; } }
</style>
