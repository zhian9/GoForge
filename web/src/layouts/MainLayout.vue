<template>
  <div class="main-layout" :class="{ 'is-home': isHome }">
    <!-- 液态玻璃背景层：所有半透明材质都要有内容可折射，否则玻璃会退化成灰色方块 -->
    <div class="gf-aurora" aria-hidden="true">
      <span class="gf-aurora__blob gf-aurora__blob--1"></span>
      <span class="gf-aurora__blob gf-aurora__blob--2"></span>
      <span class="gf-aurora__blob gf-aurora__blob--3"></span>
    </div>
    <div class="gf-grain" aria-hidden="true"></div>

    <!-- ======== Header ======== -->
    <header class="header" :class="{ 'is-scrolled': scrolled }">
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
          <router-link to="/coupons" class="nav-item" active-class="active">领券中心</router-link>
        </nav>

        <!-- 右侧操作区 -->
        <div class="header-actions">
          <div class="search-box">
            <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
            <input v-model="searchKeyword" type="text" placeholder="搜索商品..." class="search-input" @keyup.enter="handleSearch" />
          </div>

          <!-- 购物车入口对游客也可见：点进去再要求登录，而不是先藏起来 -->
          <button class="cart-btn" type="button" title="购物车" @click="goCart">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="cart-icon">
              <circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/>
              <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>
            </svg>
            <span v-if="cartCount > 0" class="cart-badge">{{ cartCount }}</span>
          </button>

          <template v-if="userStore.token">
            <!-- 消息中心入口：未读数来自 /messages/unread-count -->
            <button class="cart-btn" type="button" title="消息中心" @click="goMessages">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="cart-icon">
                <path d="M18 8a6 6 0 1 0-12 0c0 7-3 9-3 9h18s-3-2-3-9"/>
                <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
              </svg>
              <span v-if="unreadCount > 0" class="cart-badge">{{ unreadCount > 99 ? '99+' : unreadCount }}</span>
            </button>
            <div class="avatar-dropdown" @click.stop="showMenu = !showMenu">
              <div class="avatar-sm">
                <img v-if="avatarUrl" :src="avatarUrl" />
                <span v-else>{{ initial }}</span>
              </div>
              <div v-if="showMenu" class="dropdown-menu" @click.stop>
                <router-link to="/profile" class="dropdown-item" @click="showMenu=false">个人中心</router-link>
                <router-link to="/orders" class="dropdown-item" @click="showMenu=false">我的订单</router-link>
                <router-link to="/coupons" class="dropdown-item" @click="showMenu=false">我的优惠券</router-link>
                <router-link to="/messages" class="dropdown-item" @click="showMenu=false">消息中心</router-link>
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
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { resolveAssetUrl as resolveUrl } from '@/utils/api'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { useCartStore } from '@/stores/cart'
import { getCategoryTree } from '@/api/category'
import type { Category } from '@/api/category'
import { loginLocation } from '@/utils/auth'
import { getUnreadCount } from '@/api/message'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()
const searchKeyword = ref('')
const mainCategories = ref<Category[]>([])
const showMenu = ref(false)

// Safari 不支持 ?. 顶层调用，用 computed 延迟访问
const isHome = computed(() => route.path === '/')

const avatarUrl = computed(() => resolveUrl(userStore.userInfo?.avatar || ''))
const initial = computed(() => (userStore.userInfo?.nickname || userStore.userInfo?.username || 'U')[0])
const cartCount = computed(() => cartStore.totalCount)

// 未读消息数：登录后拉一次，路由切换时刷新（够用且简单，不引入 WebSocket）
const unreadCount = ref(0)
const refreshUnread = async () => {
  if (!userStore.token) {
    unreadCount.value = 0
    return
  }
  try {
    const res: any = await getUnreadCount()
    unreadCount.value = Number(res?.count ?? 0)
  } catch {
    // 未读数不是关键路径，失败保持原值
  }
}
const goMessages = () => router.push('/messages')

const showBackTop = ref(false)
const scrolled = ref(false)
const scrollToTop = () => window.scrollTo({top:0,behavior:'smooth'})
const onScroll = () => {
  showBackTop.value = window.scrollY > 400
  scrolled.value = window.scrollY > 8
}

const handleSearch = () => {
  if (searchKeyword.value.trim()) router.push({ path: '/products', query: { keyword: searchKeyword.value } })
}
const goCart = () => {
  if (!userStore.token) {
    ElMessage.warning('登录后即可查看购物车')
    router.push(loginLocation('/cart'))
    return
  }
  router.push('/cart')
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
    refreshUnread()
  }
  document.addEventListener('click', () => { showMenu.value = false })
  window.addEventListener('scroll', onScroll)
})

// 从消息中心返回时未读数需要立即变化
watch(() => route.path, refreshUnread)

onUnmounted(() => {
  window.removeEventListener('scroll', onScroll)
})
</script>

<style>
/* 全局暗色基座由 src/styles/base.css 统一提供（背景色为 --gf-void #04060C）。 */
</style>

<style scoped>
/* 设计 token 已统一到 src/styles/tokens.css，这里只保留布局职责 */
.main-layout { min-height: 100vh; display: flex; flex-direction: column; color: var(--text); font-family: var(--gf-font); }

/* ======== Header ======== */
/* 导航做成悬浮胶囊：玻璃厚于卡片，且随滚动加深，形成“贴在内容之上”的层次 */
.header { position: sticky; top: 0; z-index: var(--gf-z-header); padding: 12px 24px 0; }
.header::before {
  content: ''; position: absolute; inset: 0; z-index: 0; pointer-events: none;
  background: linear-gradient(180deg, var(--gf-bg) 0%, rgba(6,9,17,.72) 52%, transparent 100%);
  opacity: 0; transition: opacity .35s ease;
}
.header.is-scrolled::before { opacity: 1; }
.header-inner {
  position: relative; z-index: 1;
  max-width: 1280px; margin: 0 auto; padding: 0 20px; height: 64px;
  display: flex; align-items: center; gap: 28px;
  background: var(--gf-glass-nav);
  -webkit-backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke); border-radius: var(--radius-pill);
  box-shadow: var(--gf-shadow-1), var(--gf-inner-shadow);
  transition: background .35s ease, border-color .35s ease, box-shadow .35s ease;
}
.header.is-scrolled .header-inner { background: rgba(8,12,20,.74); border-color: var(--gf-stroke-strong); box-shadow: var(--gf-shadow-2), var(--gf-inner-shadow); }
.logo { cursor: pointer; flex-shrink: 0; }
.logo-img { height: 32px; width: auto; }
.nav-links { display: flex; align-items: center; gap: 4px; flex: 1; }
.nav-item { color: var(--text-dim); text-decoration: none; font-size: 14px; font-weight: 500; padding: 8px 16px; border-radius: var(--radius-pill); transition: color .2s, background .2s, box-shadow .2s; white-space: nowrap; }
.nav-item:hover { color: var(--text); background: var(--gf-glass-2); }
.nav-item.active { color: var(--accent); background: var(--accent-dim); box-shadow: inset 0 0 0 1px rgba(79,216,255,.16); }
.header-actions { display: flex; align-items: center; gap: 16px; flex-shrink: 0; }

/* Search */
.search-box { position: relative; display: flex; align-items: center; }
.search-icon { position: absolute; left: 12px; width: 16px; height: 16px; color: var(--text-dim); pointer-events: none; }
.search-input {
  background: var(--gf-glass-1); border: 1px solid var(--gf-stroke); border-radius: var(--radius-pill);
  padding: 8px 16px 8px 36px; color: var(--text); font-size: 13px; width: 200px; outline: none;
  box-shadow: var(--gf-inner-shadow-soft);
  -webkit-backdrop-filter: blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));
  transition: width .3s cubic-bezier(.22,1,.36,1), background .25s ease, border-color .25s ease, box-shadow .25s ease;
  font-family: var(--gf-font);
}
.search-input::placeholder { color: var(--text-dim); }
.search-input:focus { background: var(--gf-glass-2); border-color: rgba(79,216,255,.5); box-shadow: var(--gf-inner-shadow-soft), 0 0 0 3px var(--accent-dim); width: 260px; }

/* Cart */
.cart-btn { position: relative; display: flex; align-items: center; padding: 8px; border: 0; background: transparent; border-radius: var(--radius-sm); color: var(--text-dim); cursor: pointer; font-family: var(--gf-font); transition: all .2s; }
.cart-btn:hover { color: var(--accent); background: var(--gf-glass-2); }
.cart-icon { width: 22px; height: 22px; }
.cart-badge { position: absolute; top: 2px; right: 0; background: var(--gf-gradient); color: #04121a; font-size: 10px; font-weight: 700; min-width: 18px; height: 18px; border-radius: 9px; display: flex; align-items: center; justify-content: center; padding: 0 4px; box-shadow: 0 4px 12px -4px var(--accent-glow), var(--gf-inner-shadow-soft); }

/* Avatar */
.avatar-dropdown { position: relative; cursor: pointer; }
.avatar-sm { width: 34px; height: 34px; border-radius: 50%; overflow: hidden; background: var(--accent-dim); display: flex; align-items: center; justify-content: center; font-weight: 600; font-size: 14px; color: var(--accent); border: 1px solid var(--gf-stroke-strong); box-shadow: var(--gf-inner-shadow-soft), 0 0 0 3px rgba(255,255,255,.02); transition: border-color .25s, box-shadow .25s; }
.avatar-sm img { width: 100%; height: 100%; object-fit: cover; }
.avatar-dropdown:hover .avatar-sm { border-color: rgba(79,216,255,.6); box-shadow: var(--gf-inner-shadow-soft), 0 0 0 4px var(--accent-dim); }
.dropdown-menu {
  position: absolute; top: calc(100% + 10px); right: 0; min-width: 148px; padding: 6px;
  background: var(--gf-glass-deep);
  -webkit-backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke-strong); border-radius: var(--radius-sm);
  box-shadow: var(--gf-shadow-3), var(--gf-inner-shadow);
}
.dropdown-item { display: block; padding: 10px 14px; border-radius: var(--radius-xs); color: var(--text); text-decoration: none; font-size: 13px; transition: all .15s; cursor: pointer; }
.dropdown-item:hover { background: var(--accent-dim); color: var(--accent); }
.dropdown-item.logout { color: var(--danger); }
.dropdown-item.logout:hover { background: rgba(255,107,129,.12); color: var(--danger); }

/* Login */
.btn-login { display: inline-flex; align-items: center; padding: 9px 22px; border-radius: var(--radius-pill); background: var(--gf-gradient); color: #04121a; font-size: 13px; font-weight: 700; text-decoration: none; box-shadow: var(--gf-inner-shadow-soft), 0 8px 22px -10px var(--accent-glow); transition: transform .25s cubic-bezier(.22,1,.36,1), box-shadow .25s, filter .25s; }
.btn-login:hover { transform: translateY(-1px); filter: brightness(1.06); box-shadow: var(--gf-inner-shadow-soft), 0 14px 32px -12px var(--accent-glow); }

/* Main Content */
/* 必须建立层叠上下文，否则静态内容会绘制在固定定位的极光层之下 */
.main-content { position: relative; z-index: var(--gf-z-content); flex: 1; }

/* ======== Footer ======== */
.footer {
  position: relative; z-index: var(--gf-z-content); margin-top: 100px;
  background: linear-gradient(180deg, rgba(6,9,17,.35) 0%, rgba(4,6,12,.82) 100%);
  -webkit-backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  border-top: 1px solid var(--gf-stroke);
}
.footer-inner { max-width: 1280px; margin: 0 auto; padding: 64px 24px 32px; }

/* Grid */
.footer-grid { display: grid; grid-template-columns: 2fr 1fr 1fr 1fr 1fr; gap: 40px; margin-bottom: 48px; }

/* Brand */
.footer-brand { display: flex; flex-direction: column; gap: 16px; }
.footer-logo { cursor: pointer; }
.footer-logo img { height: 30px; width: auto; }
.footer-slogan { font-size: 13px; color: var(--text-dim); margin: 0; line-height: 1.6; max-width: 240px; }
.footer-social { display: flex; gap: 10px; margin-top: 4px; }
.social-icon { width: 36px; height: 36px; border-radius: 50%; border: 1px solid var(--gf-stroke); background: var(--gf-glass-1); display: flex; align-items: center; justify-content: center; color: var(--text-dim); transition: all .2s; }
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
.back-to-top {
  position: absolute; right: 24px; bottom: 24px; width: 44px; height: 44px; border-radius: 50%;
  border: 1px solid var(--gf-stroke); background: var(--gf-glass-2); color: var(--text-dim);
  -webkit-backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  box-shadow: var(--gf-shadow-2), var(--gf-inner-shadow);
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  opacity: 0; visibility: hidden;
  transition: opacity .3s, visibility .3s, color .2s, border-color .2s, transform .2s;
}
.back-to-top svg { width: 20px; height: 20px; }
.back-to-top.visible { opacity: 1; visibility: visible; }
.back-to-top:hover { color: var(--accent); border-color: rgba(79,216,255,.5); background: var(--gf-glass-3); transform: translateY(-2px); }

@media (max-width:900px) { .footer-grid { grid-template-columns: repeat(2,1fr); } .footer-brand { grid-column: 1/-1; } }
@media (max-width:768px) { .header { padding: 8px 12px 0; } .header-inner { padding: 0 14px; gap: 16px; } .nav-links { display: none; } .search-input { width: 140px; } .search-input:focus { width: 160px; } .footer-bottom { flex-direction: column; text-align: center; } }
</style>
