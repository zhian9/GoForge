<template>
  <div class="home">
    <!-- ==================== 首屏三栏：分类导航 / 轮播 / 账户面板 ==================== -->
    <section class="hero">
      <div class="hero-glow hero-glow-1"></div>
      <div class="hero-glow hero-glow-2"></div>
      <div class="hero-glow hero-glow-3"></div>

      <div class="hero-zone">
        <!-- ---------- 左：全部分类 ---------- -->
        <aside class="cat-rail">
          <header class="rail-head">
            <h2 class="rail-title">全部分类</h2>
            <span class="rail-count">{{ mainCategories.length }} 个品类</span>
          </header>
          <ul class="rail-list">
            <li v-for="cat in mainCategories" :key="cat.id" class="rail-item">
              <a class="rail-link" @click="goCategory(cat.id)">
                <span class="rail-name">{{ cat.name }}</span>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
              </a>
              <div v-if="cat.children?.length" class="rail-children">
                <a v-for="child in cat.children.slice(0, 4)" :key="child.id" @click="goCategory(child.id)">{{ child.name }}</a>
              </div>
            </li>
            <li v-if="mainCategories.length === 0" class="rail-empty">分类加载中…</li>
          </ul>
          <a class="rail-all" @click="goPath('/products')">查看全部商品 →</a>
        </aside>

        <!-- ---------- 中：轮播 ---------- -->
        <div class="carousel" @mouseenter="pauseCarousel" @mouseleave="resumeCarousel">
          <article
            v-for="(s, i) in slides"
            :key="s.key"
            class="slide"
            :class="{ active: i === activeSlide }"
            @click="handleSlideClick(s)"
          >
            <img v-if="s.image" :src="s.image" class="slide-img" :alt="s.title" @error="onSlideImgError(i)" />
            <div class="slide-veil" aria-hidden="true"></div>
            <div class="slide-body">
              <span class="slide-tag">{{ s.tag }}</span>
              <h1 class="slide-title">{{ s.title }}</h1>
              <p v-if="s.desc" class="slide-desc">{{ s.desc }}</p>
              <span class="slide-cta">
                {{ s.cta }}
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
              </span>
            </div>
          </article>

          <button v-if="slides.length > 1" class="car-arrow prev" type="button" aria-label="上一张" @click.stop="prevSlide">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"/></svg>
          </button>
          <button v-if="slides.length > 1" class="car-arrow next" type="button" aria-label="下一张" @click.stop="nextSlide">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
          </button>

          <div v-if="slides.length > 1" class="car-dots">
            <button
              v-for="(s, i) in slides"
              :key="s.key"
              type="button"
              class="car-dot"
              :class="{ active: i === activeSlide }"
              :aria-label="`第 ${i + 1} 张`"
              @click.stop="activeSlide = i"
            ></button>
          </div>
        </div>

        <!-- ---------- 右：账户 + 快捷入口 ---------- -->
        <aside class="side-panel">
          <div class="user-card">
            <template v-if="userStore.token">
              <div class="uc-head">
                <div class="uc-avatar">
                  <img v-if="avatarUrl" :src="avatarUrl" alt="" />
                  <span v-else>{{ initial }}</span>
                </div>
                <div class="uc-text">
                  <span class="uc-hi">欢迎回来</span>
                  <strong class="uc-name">{{ displayName }}</strong>
                </div>
              </div>
              <div class="uc-links">
                <a @click="goPath('/orders')">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 2 3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4z"/><path d="M3 6h18"/><path d="M16 10a4 4 0 0 1-8 0"/></svg>
                  我的订单
                </a>
                <a @click="goPath('/cart')">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/></svg>
                  购物车<span v-if="cartStore.totalCount > 0" class="uc-badge">{{ cartStore.totalCount }}</span>
                </a>
                <a @click="goPath('/profile')">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/></svg>
                  个人中心
                </a>
              </div>
            </template>

            <template v-else>
              <div class="uc-head">
                <div class="uc-avatar guest">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/></svg>
                </div>
                <div class="uc-text">
                  <span class="uc-hi">你好，欢迎来到 GoForge</span>
                  <span class="uc-tip">登录后可查看订单与专属优惠</span>
                </div>
              </div>
              <div class="uc-actions">
                <button class="uc-btn" type="button" @click="goPath('/login')">登录</button>
                <button class="uc-btn ghost" type="button" @click="goPath('/register')">免费注册</button>
              </div>
            </template>
          </div>

          <div class="promise-list">
            <div v-for="item in promises" :key="item.label" class="promise-item">
              <span class="promise-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" v-html="item.icon"></svg>
              </span>
              <span class="promise-text">
                <strong>{{ item.label }}</strong>
                <small>{{ item.desc }}</small>
              </span>
            </div>
          </div>
        </aside>
      </div>
    </section>

    <!-- ==================== 金刚区 ==================== -->
    <section class="section">
      <nav class="quick-grid">
        <button
          v-for="q in quickEntries"
          :key="q.key"
          class="quick-item"
          type="button"
          @click="goPath(q.to)"
        >
          <span class="quick-icon" :class="`tint-${q.tint}`">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" v-html="q.icon"></svg>
          </span>
          <span class="quick-label">{{ q.label }}</span>
          <span class="quick-sub">{{ q.sub }}</span>
        </button>
      </nav>
    </section>

    <!-- ==================== 限时秒杀 ==================== -->
    <section class="section">
      <div class="section-head">
        <div class="section-head-left">
          <h2 class="section-title">限时秒杀</h2>
          <p class="section-desc">每日精选 · 手慢无</p>
        </div>
        <div class="section-head-right">
          <div v-if="hasSeckill" class="countdown">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="countdown-icon"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
            <span>{{ pad(countdown.h) }}:{{ pad(countdown.m) }}:{{ pad(countdown.s) }}</span>
          </div>
          <a class="section-link" @click="goPath('/seckill')">查看全部 →</a>
        </div>
      </div>

      <!-- 无活动时收成一条细带，不留整块空白 -->
      <div v-if="!hasSeckill" class="seckill-empty">
        <span class="seckill-empty-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M13 2 4.1 12.5a1 1 0 0 0 .7 1.7H10l-1 7.8L18.9 11.5a1 1 0 0 0-.7-1.7H13z"/></svg>
        </span>
        <span class="seckill-empty-text">今日场次筹备中，先去逛逛其它好物</span>
        <a class="seckill-empty-link" @click="goPath('/products')">去逛逛 →</a>
      </div>

      <div v-else class="seckill-grid">
        <div
          v-for="act in seckillActivities.slice(0, 4)"
          :key="act.id"
          class="seckill-card gf-iridescent"
          @click="goPath(`/seckill/${act.id}`)"
        >
          <div class="seckill-img">
            <img :src="getSeckillImage(act)" :alt="act.name" @error="e => (e.target as HTMLImageElement).src = placeholderUri" />
            <span class="seckill-badge">SEKILL</span>
          </div>
          <div class="seckill-body">
            <h4 class="seckill-name">{{ act.name || '秒杀商品' }}</h4>
            <div class="seckill-price">
              <strong>¥{{ fmt(act.seckill_price) }}</strong>
              <del v-if="act.original_price">¥{{ fmt(act.original_price) }}</del>
            </div>
            <div class="seckill-progress">
              <div class="progress-bar"><div class="progress-fill" :style="{ width: seckillPercent(act) + '%' }"></div></div>
              <span class="progress-text">已抢{{ seckillPercent(act) }}%</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ==================== 推荐流（分类筛选 + 换一批） ==================== -->
    <section class="section">
      <div class="section-head">
        <div class="section-head-left">
          <h2 class="section-title">热门推荐</h2>
          <p class="section-desc">共 {{ products.length }} 件在售商品</p>
        </div>
        <div class="section-head-right">
          <div class="feed-tabs">
            <button
              v-for="tab in feedTabs"
              :key="tab.key"
              type="button"
              class="feed-tab"
              :class="{ active: activeFeedTab === tab.key }"
              @click="switchFeedTab(tab.key)"
            >
              {{ tab.label }}<span class="feed-tab-count">{{ tab.count }}</span>
            </button>
          </div>
          <button class="refresh-btn" type="button" @click="shuffleFeed">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 1 1-3-6.7"/><polyline points="21 4 21 10 15 10"/></svg>
            换一批
          </button>
        </div>
      </div>

      <div v-if="feedProducts.length > 0" class="feed-grid">
        <article
          v-for="p in feedProducts"
          :key="p.id"
          class="product-card gf-iridescent"
          @click="goPath(`/products/${p.id}`)"
        >
          <div class="product-img">
            <img :src="getProductImage(p)" :alt="p.name" @error="e => (e.target as HTMLImageElement).src = placeholderUri" />
            <span v-if="p.original_price > p.price" class="discount-tag">
              {{ Math.round((1 - p.price / p.original_price) * 100) }}%
            </span>
            <span v-if="isLowStock(p)" class="stock-tag">仅剩 {{ p.stock }} 件</span>
            <div class="product-overlay">
              <button class="add-cart-btn" type="button" @click.stop="addToCart(p)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="add-cart-icon"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/></svg>
                加入购物车
              </button>
            </div>
          </div>
          <div class="product-body">
            <h4 class="product-name" :title="p.name">{{ p.name }}</h4>
            <p v-if="p.subtitle" class="product-subtitle" :title="p.subtitle">{{ p.subtitle }}</p>
            <div class="product-price-row">
              <span class="price">¥{{ fmt(p.price) }}</span>
              <span v-if="p.original_price > p.price" class="price-old">¥{{ fmt(p.original_price) }}</span>
            </div>
            <div class="product-meta">
              <span>{{ salesText(p) }}</span>
              <span v-if="categoryName(p)" class="product-cat">{{ categoryName(p) }}</span>
            </div>
          </div>
        </article>
      </div>

      <div v-else class="empty-block">
        <p>{{ loading ? '商品加载中…' : '该分类暂无在售商品' }}</p>
      </div>
    </section>

    <div class="page-bottom"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { resolveAssetUrl as resolveUrl } from '@/utils/api'
import { useCartStore } from '@/stores/cart'
import { ElMessage } from 'element-plus'
import { getProductList } from '@/api/product'
import { getCategoryTree } from '@/api/category'
import { getBannerList, type Banner } from '@/api/banner'
import { listSeckillActivities, type SeckillActivity } from '@/api/seckill'
import { addItem } from '@/api/cart'
import { getDefaultSkuId } from '@/api/sku'
import type { Product } from '@/api/product'
import type { Category } from '@/api/category'
import { placeholderImage } from '@/utils/placeholder'
import { ensureLogin } from '@/utils/auth'

const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()

const mainCategories = ref<Category[]>([])
const products = ref<Product[]>([])
const seckillActivities = ref<SeckillActivity[]>([])
const loading = ref(true)

/* ==================== 轮播 ==================== */
interface Slide {
  key: string
  tag: string
  title: string
  desc: string
  image: string
  cta: string
  link: string
  linkType: number
}

const banners = ref<Banner[]>([])
const activeSlide = ref(0)
const failedSlides = ref<number[]>([])

/** 后台没配 banner 时退回静态主视觉，保证首屏不空 */
const FALLBACK_SLIDES: Slide[] = [
  {
    key: 'fallback-1',
    tag: 'GOFORGE SUMMER SALE',
    title: '618 年中大促',
    desc: '全场低至五折 · 限时特惠 · 品质保障',
    image: '',
    cta: '立即抢购',
    link: '/seckill',
    linkType: 2,
  },
]

const slides = computed<Slide[]>(() => {
  if (banners.value.length === 0) return FALLBACK_SLIDES
  return banners.value.map((b, i) => ({
    key: `banner-${b.id}`,
    tag: 'GOFORGE 精选',
    title: b.title || '新品首发',
    desc: b.description || '',
    image: failedSlides.value.includes(i) ? '' : resolveUrl(b.image_local || b.image),
    cta: '立即查看',
    link: b.link || '',
    linkType: b.link_type ?? 4,
  }))
})

let slideTimer: ReturnType<typeof setInterval> | null = null
const nextSlide = () => { if (slides.value.length > 1) activeSlide.value = (activeSlide.value + 1) % slides.value.length }
const prevSlide = () => { if (slides.value.length > 1) activeSlide.value = (activeSlide.value - 1 + slides.value.length) % slides.value.length }
const startCarousel = () => { stopCarousel(); slideTimer = setInterval(nextSlide, 5000) }
const stopCarousel = () => { if (slideTimer) { clearInterval(slideTimer); slideTimer = null } }
const pauseCarousel = () => stopCarousel()
const resumeCarousel = () => startCarousel()
const onSlideImgError = (i: number) => { if (!failedSlides.value.includes(i)) failedSlides.value = [...failedSlides.value, i] }

/** 后台 banner 的 link_type：1-商品详情 2-分类页 3-外部链接 4-无链接 */
const handleSlideClick = (s: Slide) => {
  if (s.linkType === 3 && s.link) { window.open(s.link, '_blank', 'noopener'); return }
  if (s.linkType === 1 && s.link) { goPath(`/products/${s.link}`); return }
  if (s.linkType === 2 && s.link) { goPath(s.link.startsWith('/') ? s.link : '/products'); return }
  if (s.linkType === 4) return
  goPath(s.link || '/products')
}

/* ==================== 金刚区 ==================== */
const QUICK_GLYPHS = [
  '<rect x="6" y="2" width="12" height="20" rx="3"/><path d="M11 18h2"/>',
  '<rect x="3" y="4" width="18" height="12" rx="2"/><path d="M2 20h20"/>',
  '<rect x="4" y="2" width="16" height="20" rx="3"/><path d="M4 10h16"/><path d="M8 6h.01M8 14h.01"/>',
  '<path d="M8 3 4 5l-1.5 5L5 11v10h14V11l2.5-1L20 5l-4-2-2 2h-4z"/>',
  '<path d="M6 3v9a3 3 0 0 0 6 0V3"/><path d="M9 12v9"/><path d="M17 3c-1.5 2-2 3.5-2 5s.5 3 2 4v9"/>',
  '<path d="m3 10 9-7 9 7v10a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1z"/><path d="M9 21V12h6v9"/>',
]
const QUICK_TINTS = ['rose', 'cyan', 'violet', 'mint', 'gold', 'cyan', 'violet', 'mint']

const quickEntries = computed(() => {
  const entries = [
    {
      key: 'seckill', label: '限时秒杀', sub: hasSeckill.value ? '正在开抢' : '敬请期待', to: '/seckill', tint: 'rose',
      icon: '<path d="M13 2 4.1 12.5a1 1 0 0 0 .7 1.7H10l-1 7.8L18.9 11.5a1 1 0 0 0-.7-1.7H13z"/>',
    },
    {
      key: 'all', label: '全部商品', sub: `${products.value.length} 件在售`, to: '/products', tint: 'cyan',
      icon: '<rect x="3" y="3" width="7" height="7" rx="2"/><rect x="14" y="3" width="7" height="7" rx="2"/><rect x="3" y="14" width="7" height="7" rx="2"/><rect x="14" y="14" width="7" height="7" rx="2"/>',
    },
  ]

  mainCategories.value.forEach((cat, i) => {
    const count = productsOfCategory(cat.id).length
    entries.push({
      key: `cat-${cat.id}`,
      label: cat.name,
      sub: count > 0 ? `${count} 件` : '即将上新',
      to: `/products?category_id=${cat.id}`,
      tint: QUICK_TINTS[(i + 2) % QUICK_TINTS.length],
      icon: QUICK_GLYPHS[i % QUICK_GLYPHS.length],
    })
  })

  return entries
})

/* ==================== 服务承诺 ==================== */
const promises = [
  { label: '正品保障', desc: '官方渠道直供', icon: '<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/><path d="m9 12 2 2 4-4"/>' },
  { label: '极速发货', desc: '24 小时内出库', icon: '<path d="M1 3h15v13H1z"/><path d="M16 8h4l3 3v5h-7z"/><circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/>' },
  { label: '七天无理由', desc: '不满意可退换', icon: '<path d="M3 12a9 9 0 1 0 3-6.7"/><polyline points="3 4 3 10 9 10"/>' },
  { label: '专属客服', desc: '7×24 小时在线', icon: '<path d="M3 18v-6a9 9 0 0 1 18 0v6"/><path d="M21 19a2 2 0 0 1-2 2h-1v-6h3zM3 19a2 2 0 0 0 2 2h1v-6H3z"/>' },
]

/* ==================== 秒杀 ==================== */
const hasSeckill = computed(() => seckillActivities.value.length > 0)
const countdown = reactive({ h: 2, m: 15, s: 46 })
let timer: ReturnType<typeof setInterval> | null = null
const pad = (n: number) => String(n).padStart(2, '0')
const tick = () => {
  if (countdown.s > 0) countdown.s--
  else if (countdown.m > 0) { countdown.m--; countdown.s = 59 }
  else if (countdown.h > 0) { countdown.h--; countdown.m = 59; countdown.s = 59 }
  else if (timer) { clearInterval(timer); timer = null }
}
const seckillPercent = (act: SeckillActivity) => {
  const stock = Number(act.stock || 1), sold = Number(act.sold || 0)
  return Math.min(Math.round((sold / Math.max(stock + sold, 1)) * 100), 100)
}

/* ==================== 商品 ==================== */
const categoryIndex = computed(() => {
  const map = new Map<number, Category>()
  const walk = (list: Category[]) => list.forEach((c) => {
    map.set(c.id, c)
    if (c.children?.length) walk(c.children)
  })
  walk(mainCategories.value)
  return map
})

/** 分类 id → 含全部后代分类的 id 列表：商品挂在下级分类，按一级分类筛选必须带上下级 */
const descendantIds = (id: number) => {
  const out = [id]
  const walk = (node?: Category) => node?.children?.forEach((c) => { out.push(c.id); walk(c) })
  walk(categoryIndex.value.get(id))
  return out
}

const productsOfCategory = (id: number) => {
  const ids = descendantIds(id)
  return products.value.filter((p) => ids.includes(Number(p.category_id)))
}

const categoryName = (p: Product) => categoryIndex.value.get(Number(p.category_id))?.name || ''
const isLowStock = (p: Product) => Number(p.stock) > 0 && Number(p.stock) <= 20
const salesText = (p: Product) => (Number(p.sales) > 0 ? `已售 ${Number(p.sales)}` : '新品上架')

const activeFeedTab = ref('all')
const feedOffset = ref(0)

const feedTabs = computed(() => {
  const tabs = [{ key: 'all', label: '全部', count: products.value.length }]
  mainCategories.value.forEach((cat) => {
    const count = productsOfCategory(cat.id).length
    if (count > 0) tabs.push({ key: String(cat.id), label: cat.name, count })
  })
  return tabs
})

const feedProducts = computed(() => {
  const list = activeFeedTab.value === 'all'
    ? products.value
    : productsOfCategory(Number(activeFeedTab.value))
  if (list.length === 0 || feedOffset.value === 0) return list
  // 换一批：整体轮转，不引入随机数，保证同一批只出现一次且顺序稳定
  const n = feedOffset.value % list.length
  return list.slice(n).concat(list.slice(0, n))
})

const switchFeedTab = (key: string) => { activeFeedTab.value = key; feedOffset.value = 0 }
const shuffleFeed = () => { feedOffset.value += 3 }

/* ==================== 通用 ==================== */
const fmt = (v: any) => Number(v || 0).toFixed(2)

const placeholderUri = placeholderImage()

const getProductImage = (p: Product) => {
  const img = p.local_main_image || p.main_image || p.images?.[0]
  return img ? resolveUrl(img) : placeholderUri
}
const getSeckillImage = (act: SeckillActivity) => (act.sku_image ? resolveUrl(act.sku_image) : placeholderUri)

const displayName = computed(() => userStore.userInfo?.nickname || userStore.userInfo?.username || '会员')
const initial = computed(() => displayName.value[0] || 'G')
const avatarUrl = computed(() => resolveUrl(userStore.userInfo?.avatar || ''))

const goPath = (to: string) => router.push(to)
const goCategory = (id: number) => router.push({ path: '/products', query: { category_id: id } })

const addToCart = async (p: Product) => {
  if (!ensureLogin(router, '请先登录后再加入购物车')) return
  // 首页商品是 SPU，必须先解析出真实 SKU ID 再加购（不能把商品 ID 当 SKU ID 传）
  const skuId = await getDefaultSkuId(p.id)
  if (!skuId) { ElMessage.warning('该商品暂无可用规格'); return }
  try {
    await addItem({ skuId, quantity: 1 })
    ElMessage.success('已添加到购物车')
    cartStore.fetchCart()
  } catch { ElMessage.error('添加失败') }
}

/* ==================== 数据加载 ==================== */
const fetchCategories = async () => {
  try {
    const r = await getCategoryTree({ status: -1 }) as any
    if (r.code === 0 && r.data) mainCategories.value = r.data
  } catch { /* 分类失败不影响其余模块 */ }
}

const fetchBanners = async () => {
  try {
    const r = await getBannerList({ status: 1, limit: 6 }) as any
    banners.value = Array.isArray(r?.data) ? r.data : []
    activeSlide.value = 0
    if (banners.value.length > 1) startCarousel()
  } catch { /* 使用兜底主视觉 */ }
}

const fetchSeckill = async () => {
  try {
    const r = await listSeckillActivities({ page: 1, page_size: 6, status: 1 }) as any
    seckillActivities.value = r?.data?.list || []
  } catch { /* 秒杀区自动收成提示条 */ }
}

/**
 * 首页不再按分类逐个请求（原来 5 个分类 = 5 次请求，且每个分类只有 1-2 件商品，
 * 铺满一行会留下大片空洞）。改为一次性取全部在售商品，再在前端按分类树聚合，
 * 这样分类筛选、金刚区计数、推荐流都来自同一份数据。
 */
const fetchProducts = async () => {
  loading.value = true
  try {
    const r = await getProductList({ page: 1, page_size: 60, status: 1 }) as any
    if (r.code === 0 && r.data) products.value = r.data.list || []
  } catch { products.value = [] }
  finally { loading.value = false }
}

onMounted(async () => {
  timer = setInterval(tick, 1000)
  await Promise.all([fetchCategories(), fetchProducts()])
  fetchBanners()
  fetchSeckill()
})

onUnmounted(() => {
  if (timer) { clearInterval(timer); timer = null }
  stopCarousel()
})
</script>

<style scoped>
.home { font-family: var(--gf-font); }

/* ==================== 首屏 ==================== */
.hero { position: relative; overflow: hidden; padding: 24px 0 8px; }

.hero-glow { position: absolute; border-radius: 50%; filter: blur(120px); opacity: .24; pointer-events: none; }
.hero-glow-1 { width: 560px; height: 560px; background: radial-gradient(circle, var(--accent), transparent); top: -220px; left: -120px; }
.hero-glow-2 { width: 420px; height: 420px; background: radial-gradient(circle, rgba(139, 124, 255, .55), transparent); bottom: -160px; right: -80px; }
.hero-glow-3 { width: 320px; height: 320px; background: radial-gradient(circle, rgba(110, 231, 200, .4), transparent); top: 30%; left: 46%; }

.hero-zone {
  position: relative;
  z-index: 1;
  max-width: 1280px;
  margin: 0 auto;
  padding: 0 24px;
  display: grid;
  grid-template-columns: 216px minmax(0, 1fr) 264px;
  gap: 16px;
  align-items: stretch;
}

/* ---------- 左：分类导航 ---------- */
.cat-rail {
  display: flex;
  flex-direction: column;
  padding: 16px 14px;
  border-radius: var(--radius);
  background: var(--gf-glass-2);
  -webkit-backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke);
  box-shadow: var(--gf-inner-shadow);
}

.rail-head { display: flex; align-items: baseline; justify-content: space-between; gap: 8px; margin-bottom: 10px; padding: 0 4px; }
.rail-title { margin: 0; font-size: 14px; font-weight: 700; color: var(--text); }
.rail-count { font-size: 11px; color: var(--gf-text-mute); }

.rail-list { flex: 1; list-style: none; margin: 0; padding: 0; overflow-y: auto; }
.rail-item { padding: 6px 4px; border-radius: var(--radius-sm); transition: background .2s ease; }
.rail-item:hover { background: var(--gf-glass-2); }

.rail-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  color: var(--text);
  transition: color .2s ease;
}

.rail-link svg { width: 13px; height: 13px; color: var(--gf-text-mute); transition: transform .2s ease; }
.rail-item:hover .rail-link { color: var(--accent); }
.rail-item:hover .rail-link svg { transform: translateX(2px); color: var(--accent); }

.rail-children { display: flex; flex-wrap: wrap; gap: 4px 8px; margin-top: 5px; }
.rail-children a { font-size: 11px; color: var(--text-dim); cursor: pointer; transition: color .2s ease; }
.rail-children a:hover { color: var(--accent); }

.rail-empty { padding: 20px 4px; font-size: 12px; color: var(--text-dim); text-align: center; }

.rail-all {
  display: block;
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid var(--gf-stroke);
  font-size: 12px;
  color: var(--accent);
  cursor: pointer;
  text-align: center;
  transition: opacity .2s ease;
}

.rail-all:hover { opacity: .75; }

/* ---------- 中：轮播 ---------- */
.carousel {
  position: relative;
  min-height: 340px;
  border-radius: var(--radius-lg);
  overflow: hidden;
  border: 1px solid var(--gf-stroke-strong);
  box-shadow: var(--gf-shadow-3), var(--gf-inner-shadow);
  background: var(--gf-glass-2);
  cursor: pointer;
}

.slide {
  position: absolute;
  inset: 0;
  opacity: 0;
  visibility: hidden;
  transition: opacity .7s ease, visibility .7s ease;
  background:
    radial-gradient(120% 120% at 12% 8%, rgba(79, 216, 255, .3) 0%, transparent 58%),
    radial-gradient(100% 120% at 92% 92%, rgba(139, 124, 255, .32) 0%, transparent 60%),
    linear-gradient(140deg, #0d1424 0%, #070b14 100%);
}

.slide.active { opacity: 1; visibility: visible; }

.slide-img { width: 100%; height: 100%; object-fit: cover; }

/* 图片之上压一层由暗到透明的遮罩，保证白色标题在任何底图上都可读 */
.slide-veil {
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, rgba(4, 6, 12, .88) 0%, rgba(4, 6, 12, .55) 48%, rgba(4, 6, 12, .18) 100%);
}

.slide-body {
  position: absolute;
  left: 40px;
  bottom: 40px;
  right: 40px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 12px;
}

.slide-tag {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: .16em;
  color: var(--accent);
  background: rgba(79, 216, 255, .12);
  border: 1px solid rgba(79, 216, 255, .28);
  padding: 5px 14px;
  border-radius: var(--radius-pill);
  -webkit-backdrop-filter: blur(var(--gf-blur-sm));
  backdrop-filter: blur(var(--gf-blur-sm));
}

.slide-title {
  margin: 0;
  font-size: clamp(28px, 3.6vw, 44px);
  font-weight: 800;
  line-height: 1.12;
  letter-spacing: -.02em;
  background: linear-gradient(120deg, #fff 0%, #cfefff 45%, #b9b0ff 100%);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}

.slide-desc { margin: 0; font-size: 15px; color: var(--text-dim); }

.slide-cta {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
  padding: 12px 26px;
  border-radius: var(--radius-pill);
  background: var(--gf-gradient);
  color: #04121a;
  font-size: 14px;
  font-weight: 700;
  box-shadow: var(--gf-inner-shadow-soft), 0 12px 32px -14px var(--accent-glow);
  transition: transform .25s cubic-bezier(.22, 1, .36, 1), filter .25s ease;
}

.slide-cta svg { width: 16px; height: 16px; }
.carousel:hover .slide-cta { transform: translateY(-1px); filter: brightness(1.06); }

.car-arrow {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  border: 1px solid var(--gf-stroke-strong);
  background: var(--gf-glass-deep);
  -webkit-backdrop-filter: blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));
  color: var(--text);
  cursor: pointer;
  opacity: 0;
  transition: opacity .25s ease, background .2s ease, border-color .2s ease;
}

.car-arrow svg { width: 18px; height: 18px; }
.car-arrow.prev { left: 16px; }
.car-arrow.next { right: 16px; }
.carousel:hover .car-arrow { opacity: 1; }
.car-arrow:hover { background: var(--gf-glass-3); border-color: rgba(79, 216, 255, .5); color: var(--accent); }

.car-dots { position: absolute; right: 24px; bottom: 24px; display: flex; gap: 7px; }

.car-dot {
  width: 8px;
  height: 8px;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: rgba(255, 255, 255, .28);
  cursor: pointer;
  transition: width .3s cubic-bezier(.22, 1, .36, 1), background .3s ease;
}

.car-dot.active { width: 22px; border-radius: var(--radius-pill); background: var(--gf-gradient); }

/* ---------- 右：账户面板 ---------- */
.side-panel { display: flex; flex-direction: column; gap: 14px; }

.user-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 18px;
  border-radius: var(--radius);
  background: var(--gf-glass-2);
  -webkit-backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke);
  box-shadow: var(--gf-inner-shadow);
}

.uc-head { display: flex; align-items: center; gap: 12px; }

.uc-avatar {
  width: 46px;
  height: 46px;
  flex-shrink: 0;
  border-radius: 50%;
  padding: 1px;
  background: var(--gf-gradient);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  color: var(--gf-accent);
  font-size: 18px;
  font-weight: 700;
}

.uc-avatar img { width: 100%; height: 100%; border-radius: 50%; object-fit: cover; border: 2px solid var(--gf-bg); }
.uc-avatar span { display: flex; align-items: center; justify-content: center; width: 100%; height: 100%; border-radius: 50%; background: var(--gf-bg-elev); border: 2px solid var(--gf-bg); }
.uc-avatar.guest { color: var(--text-dim); }
.uc-avatar.guest svg { width: 22px; height: 22px; }

.uc-text { display: flex; flex-direction: column; gap: 3px; min-width: 0; }
.uc-hi { font-size: 12px; color: var(--text-dim); }
.uc-name { font-size: 15px; font-weight: 700; color: var(--text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.uc-tip { font-size: 13px; color: var(--text); line-height: 1.5; }

.uc-links { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; }

.uc-links a {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 10px 4px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: var(--text-dim);
  cursor: pointer;
  background: var(--gf-glass-1);
  border: 1px solid var(--gf-stroke);
  transition: color .2s ease, background .2s ease, transform .2s ease;
}

.uc-links a svg { width: 17px; height: 17px; }
.uc-links a:hover { color: var(--accent); background: var(--gf-glass-2); transform: translateY(-2px); }

.uc-badge {
  position: absolute;
  top: 4px;
  right: 8px;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  border-radius: 8px;
  background: var(--gf-gradient);
  color: #04121a;
  font-size: 10px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

.uc-actions { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }

.uc-btn {
  padding: 10px 12px;
  border: none;
  border-radius: var(--radius-pill);
  background: var(--gf-gradient);
  color: #04121a;
  font-family: var(--gf-font);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  box-shadow: var(--gf-inner-shadow-soft), 0 10px 26px -12px var(--accent-glow);
  transition: transform .2s ease, filter .2s ease;
}

.uc-btn.ghost {
  background: var(--gf-glass-1);
  border: 1px solid var(--gf-stroke-strong);
  color: var(--text);
  font-weight: 600;
  box-shadow: var(--gf-inner-shadow-soft);
}

.uc-btn:hover { transform: translateY(-1px); filter: brightness(1.05); }

/* ---------- 右：服务承诺 ---------- */
.promise-list {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 14px 16px;
  border-radius: var(--radius);
  background: var(--gf-glass-1);
  -webkit-backdrop-filter: blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke);
  box-shadow: var(--gf-inner-shadow-soft);
}

.promise-item { display: flex; align-items: center; gap: 10px; padding: 9px 0; }
.promise-item + .promise-item { border-top: 1px solid var(--gf-stroke); }

.promise-icon {
  width: 30px;
  height: 30px;
  flex-shrink: 0;
  border-radius: var(--radius-xs);
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(79, 216, 255, .12);
  color: var(--accent);
}

.promise-icon svg { width: 15px; height: 15px; }
.promise-text { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.promise-text strong { font-size: 13px; font-weight: 600; color: var(--text); }
.promise-text small { font-size: 11px; color: var(--gf-text-mute); }

/* ==================== 区块通用 ==================== */
.section { max-width: 1280px; margin: 0 auto; padding: 36px 24px 0; }

.section-head { display: flex; justify-content: space-between; align-items: flex-end; margin-bottom: 20px; gap: 16px; flex-wrap: wrap; }
.section-head-left { display: flex; flex-direction: column; gap: 4px; }
.section-head-right { display: flex; align-items: center; gap: 14px; flex-wrap: wrap; }

.section-title { font-size: 22px; font-weight: 700; margin: 0; color: var(--text); letter-spacing: -.01em; display: flex; align-items: center; gap: 10px; }
.section-title::before { content: ''; width: 3px; height: 20px; border-radius: 2px; background: var(--gf-gradient); box-shadow: 0 0 12px var(--accent-glow); }
.section-desc { font-size: 13px; color: var(--text-dim); margin: 0; }
.section-link { color: var(--accent); cursor: pointer; font-size: 13px; font-weight: 500; transition: opacity .2s; white-space: nowrap; }
.section-link:hover { opacity: .75; }

.page-bottom { height: 72px; }

/* ==================== 金刚区 ==================== */
.quick-grid { display: grid; grid-template-columns: repeat(8, minmax(0, 1fr)); gap: 12px; }

.quick-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px 8px 14px;
  border-radius: var(--radius);
  background: var(--gf-glass-1);
  -webkit-backdrop-filter: blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke);
  box-shadow: var(--gf-inner-shadow-soft);
  cursor: pointer;
  font-family: var(--gf-font);
  transition: transform .3s cubic-bezier(.22, 1, .36, 1), background .25s ease, border-color .25s ease, box-shadow .3s ease;
}

.quick-item:hover {
  transform: translateY(-4px);
  background: var(--gf-glass-2);
  border-color: var(--gf-stroke-strong);
  box-shadow: var(--gf-shadow-2), var(--gf-inner-shadow);
}

.quick-icon {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--gf-inner-shadow-soft);
}

.quick-icon svg { width: 21px; height: 21px; }

.tint-cyan { background: rgba(79, 216, 255, .14); color: var(--gf-accent); }
.tint-gold { background: rgba(232, 200, 138, .15); color: var(--gf-gold); }
.tint-violet { background: rgba(139, 124, 255, .15); color: var(--gf-accent-2); }
.tint-mint { background: rgba(110, 231, 200, .14); color: var(--gf-accent-3); }
.tint-rose { background: rgba(255, 107, 129, .15); color: var(--gf-danger); }

.quick-label { font-size: 13px; font-weight: 600; color: var(--text); }
.quick-sub { font-size: 11px; color: var(--gf-text-mute); }

/* ==================== 秒杀 ==================== */
.countdown {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 16px;
  border-radius: var(--radius-pill);
  background: rgba(255, 107, 129, .12);
  border: 1px solid rgba(255, 107, 129, .26);
  box-shadow: var(--gf-inner-shadow-soft);
  font-size: 15px;
  font-weight: 700;
  color: var(--danger);
  letter-spacing: .04em;
  font-variant-numeric: tabular-nums;
}

.countdown-icon { width: 16px; height: 16px; }

.seckill-empty {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  border-radius: var(--radius);
  background: var(--gf-glass-1);
  border: 1px dashed var(--gf-stroke-strong);
  color: var(--text-dim);
  font-size: 13px;
}

.seckill-empty-icon { display: flex; color: var(--gf-danger); }
.seckill-empty-icon svg { width: 18px; height: 18px; }
.seckill-empty-text { flex: 1; min-width: 0; }
.seckill-empty-link { color: var(--accent); cursor: pointer; white-space: nowrap; }
.seckill-empty-link:hover { opacity: .75; }

.seckill-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 16px; }

.seckill-card {
  position: relative;
  border-radius: var(--radius);
  background: var(--gf-glass-2);
  -webkit-backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke);
  box-shadow: var(--gf-inner-shadow);
  overflow: hidden;
  cursor: pointer;
  transition: transform .35s cubic-bezier(.22, 1, .36, 1), border-color .3s ease, box-shadow .35s ease;
}

.seckill-card:hover {
  transform: translateY(-6px);
  border-color: rgba(79, 216, 255, .4);
  box-shadow: var(--gf-shadow-2), var(--gf-inner-shadow), 0 18px 50px -24px var(--accent-glow);
}

.seckill-img {
  position: relative;
  height: 176px;
  background: linear-gradient(150deg, rgba(255, 255, 255, .05), rgba(255, 255, 255, .01));
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.seckill-img img { width: 100%; height: 100%; object-fit: cover; transition: transform .4s ease; }
.seckill-card:hover .seckill-img img { transform: scale(1.06); }

.seckill-badge {
  position: absolute;
  top: 10px;
  left: 10px;
  padding: 4px 10px;
  border-radius: var(--radius-xs);
  background: var(--gf-gradient);
  color: #04121a;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: .06em;
  box-shadow: 0 6px 16px -8px var(--accent-glow), var(--gf-inner-shadow-soft);
}

.seckill-body { padding: 14px 16px 16px; }
.seckill-name { font-size: 14px; font-weight: 500; margin: 0 0 8px; color: var(--text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.seckill-price { display: flex; align-items: baseline; gap: 8px; margin-bottom: 10px; }
.seckill-price strong { font-size: 18px; color: var(--price); font-variant-numeric: tabular-nums; }
.seckill-price del { font-size: 12px; color: var(--text-dim); }
.seckill-progress { display: flex; align-items: center; gap: 8px; }

.progress-bar { flex: 1; height: 4px; border-radius: 2px; background: rgba(255, 255, 255, .07); box-shadow: inset 0 1px 2px rgba(0, 0, 0, .4); overflow: hidden; }
.progress-fill { height: 100%; border-radius: 2px; background: linear-gradient(90deg, var(--accent), var(--accent-2)); box-shadow: 0 0 10px -2px var(--accent-glow); transition: width .6s cubic-bezier(.22, 1, .36, 1); }
.progress-text { font-size: 11px; color: var(--text-dim); white-space: nowrap; }

/* ==================== 推荐流 ==================== */
.feed-tabs { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }

.feed-tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border-radius: var(--radius-pill);
  border: 1px solid var(--gf-stroke);
  background: var(--gf-glass-1);
  box-shadow: var(--gf-inner-shadow-soft);
  color: var(--text-dim);
  font-family: var(--gf-font);
  font-size: 13px;
  cursor: pointer;
  transition: color .2s ease, background .2s ease, border-color .2s ease;
}

.feed-tab:hover { color: var(--text); background: var(--gf-glass-2); border-color: var(--gf-stroke-strong); }

.feed-tab.active {
  background: var(--gf-gradient);
  border-color: transparent;
  color: #04121a;
  font-weight: 700;
  box-shadow: var(--gf-inner-shadow-soft), 0 8px 22px -12px var(--accent-glow);
}

.feed-tab-count { font-size: 11px; opacity: .7; font-variant-numeric: tabular-nums; }

.refresh-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border-radius: var(--radius-pill);
  border: 1px solid var(--gf-stroke);
  background: var(--gf-glass-1);
  box-shadow: var(--gf-inner-shadow-soft);
  color: var(--text-dim);
  font-family: var(--gf-font);
  font-size: 13px;
  cursor: pointer;
  transition: color .2s ease, border-color .2s ease, transform .3s cubic-bezier(.22, 1, .36, 1);
}

.refresh-btn svg { width: 14px; height: 14px; }
.refresh-btn:hover { color: var(--accent); border-color: rgba(79, 216, 255, .4); transform: rotate(-12deg); }

/* 4 列而非 5 列：在 1280 容器内卡片更饱满，且商品数不整除时末行不会只剩一张孤卡 */
.feed-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 16px; }

.product-card {
  position: relative;
  border-radius: var(--radius);
  background: var(--gf-glass-2);
  -webkit-backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke);
  box-shadow: var(--gf-inner-shadow);
  overflow: hidden;
  cursor: pointer;
  transition: transform .35s cubic-bezier(.22, 1, .36, 1), border-color .3s ease, box-shadow .35s ease;
}

.product-card:hover {
  transform: translateY(-6px);
  border-color: var(--gf-stroke-strong);
  box-shadow: var(--gf-shadow-3), var(--gf-inner-shadow);
}

.product-img {
  position: relative;
  aspect-ratio: 1 / 1;
  background: linear-gradient(150deg, rgba(255, 255, 255, .05), rgba(255, 255, 255, .01));
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.product-img img { width: 100%; height: 100%; object-fit: cover; transition: transform .5s ease; }
.product-card:hover .product-img img { transform: scale(1.07); }

.discount-tag {
  position: absolute;
  top: 10px;
  left: 10px;
  z-index: 2;
  padding: 4px 10px;
  border-radius: var(--radius-xs);
  background: var(--gf-gradient);
  color: #04121a;
  font-size: 11px;
  font-weight: 700;
  box-shadow: 0 6px 18px -8px var(--accent-glow), var(--gf-inner-shadow-soft);
}

.stock-tag {
  position: absolute;
  top: 10px;
  right: 10px;
  z-index: 2;
  padding: 4px 9px;
  border-radius: var(--radius-xs);
  background: rgba(255, 107, 129, .16);
  border: 1px solid rgba(255, 107, 129, .3);
  color: #ffb3bf;
  font-size: 11px;
  font-weight: 600;
  -webkit-backdrop-filter: blur(var(--gf-blur-xs));
  backdrop-filter: blur(var(--gf-blur-xs));
}

.product-overlay {
  position: absolute;
  inset: 0;
  z-index: 3;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  padding: 14px;
  background: linear-gradient(to top, rgba(4, 6, 12, .86) 0%, transparent 62%);
  opacity: 0;
  transition: opacity .3s ease;
}

.product-card:hover .product-overlay { opacity: 1; }

.add-cart-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 9px 20px;
  border: none;
  border-radius: var(--radius-pill);
  background: var(--gf-gradient);
  color: #04121a;
  font-family: var(--gf-font);
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  box-shadow: var(--gf-inner-shadow-soft), 0 10px 26px -12px var(--accent-glow);
  transform: translateY(8px);
  transition: transform .25s cubic-bezier(.22, 1, .36, 1), filter .25s ease;
}

.product-card:hover .add-cart-btn { transform: translateY(0); }
.add-cart-btn:hover { filter: brightness(1.08); }
.add-cart-icon { width: 15px; height: 15px; }

.product-body { padding: 12px 13px 14px; display: flex; flex-direction: column; gap: 6px; }

.product-name {
  margin: 0;
  font-size: 13px;
  font-weight: 500;
  line-height: 1.4;
  color: var(--text);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 36px;
}

.product-subtitle { margin: 0; font-size: 11px; color: var(--text-dim); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.product-price-row { display: flex; align-items: baseline; gap: 8px; }
.price { font-size: 17px; font-weight: 700; color: var(--price); font-variant-numeric: tabular-nums; }
.price-old { font-size: 11px; color: var(--text-dim); text-decoration: line-through; }

.product-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding-top: 8px;
  border-top: 1px solid var(--gf-stroke);
  font-size: 11px;
  color: var(--gf-text-mute);
}

.product-cat { padding: 2px 7px; border-radius: var(--radius-xs); background: var(--gf-glass-2); color: var(--text-dim); }

/* 空状态 */
.empty-block { text-align: center; padding: 56px 0; color: var(--text-dim); font-size: 14px; }
.empty-block p { margin: 0; }

/* 卡片入场错峰 */
.product-card, .seckill-card, .quick-item { animation: gf-rise .5s cubic-bezier(.22, 1, .36, 1) both; }
.product-card:nth-child(5n+2) { animation-delay: .04s; }
.product-card:nth-child(5n+3) { animation-delay: .08s; }
.product-card:nth-child(5n+4) { animation-delay: .12s; }
.product-card:nth-child(5n) { animation-delay: .16s; }
.quick-item:nth-child(2) { animation-delay: .04s; }
.quick-item:nth-child(3) { animation-delay: .08s; }
.quick-item:nth-child(4) { animation-delay: .12s; }
.quick-item:nth-child(5) { animation-delay: .16s; }
.quick-item:nth-child(6) { animation-delay: .2s; }
.quick-item:nth-child(7) { animation-delay: .24s; }
.quick-item:nth-child(8) { animation-delay: .28s; }

/* ==================== 响应式 ==================== */
@media (max-width: 1180px) {
  .hero-zone { grid-template-columns: minmax(0, 1fr) 252px; }
  .cat-rail { display: none; }
  .quick-grid { grid-template-columns: repeat(5, minmax(0, 1fr)); }
}

@media (max-width: 920px) {
  .hero { padding-top: 16px; }
  .hero-zone { grid-template-columns: minmax(0, 1fr); padding: 0 16px; }
  .carousel { min-height: 280px; }
  .slide-body { left: 24px; right: 24px; bottom: 26px; }
  .side-panel { flex-direction: row; }
  .user-card, .promise-list { flex: 1; }
  .section { padding: 28px 16px 0; }
  .feed-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); }
}

@media (max-width: 680px) {
  .carousel { min-height: 220px; }
  .slide-title { font-size: 26px; }
  .slide-desc { font-size: 13px; }
  .slide-cta { padding: 10px 20px; font-size: 13px; }
  .car-dots { right: 16px; bottom: 16px; }
  .side-panel { flex-direction: column; }
  .quick-grid { grid-template-columns: repeat(4, minmax(0, 1fr)); }
  .quick-item { padding: 14px 6px 12px; }
  .feed-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
  .seckill-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .section-head-right { width: 100%; justify-content: space-between; }
}

@media (max-width: 420px) {
  .quick-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); }
}
</style>
