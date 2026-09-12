<template>
  <div class="home">
    <!-- ======== Hero Banner ======== -->
    <section class="hero">
      <!-- 装饰光晕 -->
      <div class="hero-glow hero-glow-1"></div>
      <div class="hero-glow hero-glow-2"></div>
      <div class="hero-glow hero-glow-3"></div>
      <!-- 网格 -->
      <div class="hero-grid"></div>

      <div class="hero-content">
        <div class="hero-tag">GOFORGE SUMMER SALE</div>
        <h1 class="hero-title">618 年中大促</h1>
        <p class="hero-subtitle">全场低至五折 · 限时特惠 · 品质保障</p>
        <div class="hero-actions">
          <button class="cta-btn" @click="$router.push('/seckill')">
            立即抢购
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="cta-arrow">
              <path d="M5 12h14M12 5l7 7-7 7"/>
            </svg>
          </button>
          <button class="cta-btn-secondary" @click="$router.push('/products')">浏览全部商品</button>
        </div>
        <div class="hero-stats">
          <div class="stat-item"><span class="stat-num">10万+</span><span class="stat-label">品质商品</span></div>
          <div class="stat-divider"></div>
          <div class="stat-item"><span class="stat-num">50万+</span><span class="stat-label">信赖用户</span></div>
          <div class="stat-divider"></div>
          <div class="stat-item"><span class="stat-num">24h</span><span class="stat-label">极速发货</span></div>
        </div>
      </div>
    </section>

    <!-- ======== 秒杀专区 ======== -->
    <section class="section">
      <div class="section-head">
        <div class="section-head-left">
          <h2 class="section-title">限时秒杀</h2>
          <p class="section-desc">每日精选 · 手慢无</p>
        </div>
        <div class="section-head-right">
          <div class="countdown" v-if="countdown.h > 0 || countdown.m > 0 || countdown.s > 0">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="countdown-icon"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
            <span>{{ pad(countdown.h) }}:{{ pad(countdown.m) }}:{{ pad(countdown.s) }}</span>
          </div>
          <a class="section-link" @click="$router.push('/seckill')">查看全部 →</a>
        </div>
      </div>

      <div v-if="seckillActivities.length === 0" class="empty-block">
        <p>暂无秒杀活动，敬请期待</p>
      </div>
      <div v-else class="seckill-grid">
        <div v-for="act in seckillActivities.slice(0,4)" :key="act.id" class="seckill-card" @click="$router.push(`/seckill/${act.id}`)">
          <div class="seckill-img">
            <img :src="getSeckillImage(act)" :alt="act.name" @error="e => (e.target as HTMLImageElement).src=placeholderUri" />
            <span class="seckill-badge">SEKILL</span>
          </div>
          <div class="seckill-body">
            <h4 class="seckill-name">{{ act.name || '秒杀商品' }}</h4>
            <div class="seckill-price"><strong>¥{{ fmt(act.seckill_price) }}</strong><del v-if="act.original_price">¥{{ fmt(act.original_price) }}</del></div>
            <div class="seckill-progress">
              <div class="progress-bar"><div class="progress-fill" :style="{width:seckillPercent(act)+'%'}"></div></div>
              <span class="progress-text">已抢{{ seckillPercent(act) }}%</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ======== 分类商品区 ======== -->
    <section v-for="cat in displayCategories" :key="cat.id" class="section">
      <div class="section-head">
        <div class="section-head-left">
          <h2 class="section-title">{{ cat.name }}</h2>
          <p class="section-desc">热门精选</p>
        </div>
        <a class="section-link" @click="$router.push(`/products?category_id=${cat.id}`)">查看更多 →</a>
      </div>

      <div v-if="!hotByCat[cat.id] || hotByCat[cat.id].length === 0" class="empty-block">
        <p>暂无商品</p>
      </div>
      <div v-else class="product-grid">
        <div v-for="p in (hotByCat[cat.id] || [])" :key="p.id" class="product-card" @click="$router.push(`/products/${p.id}`)">
          <div class="product-img">
            <img :src="getProductImage(p)" :alt="p.name" @error="e => (e.target as HTMLImageElement).src=placeholderUri" />
            <span v-if="p.original_price > p.price" class="discount-tag">{{ Math.round((1-p.price/p.original_price)*100) }}%</span>
            <!-- hover 加购按钮 -->
            <div class="product-overlay">
              <button class="add-cart-btn" @click.stop="addToCart(p)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="add-cart-icon"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/></svg>
                加入购物车
              </button>
            </div>
          </div>
          <div class="product-body">
            <h4 class="product-name">{{ p.name }}</h4>
            <p class="product-subtitle" v-if="p.subtitle">{{ p.subtitle }}</p>
            <div class="product-price-row">
              <span class="price">¥{{ fmt(p.price) }}</span>
              <span v-if="p.original_price > p.price" class="price-old">¥{{ fmt(p.original_price) }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 页面底部间距 -->
    <div style="height:80px"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useCartStore } from '@/stores/cart'
import { ElMessage } from 'element-plus'
import { getProductList } from '@/api/product'
import { getCategoryTree } from '@/api/category'
import { listSeckillActivities, type SeckillActivity } from '@/api/seckill'
import { addItem } from '@/api/cart'
import { getDefaultSkuId } from '@/api/sku'
import type { Product } from '@/api/product'
import type { Category } from '@/api/category'

const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()

const mainCategories = ref<Category[]>([])
const hotByCat = ref<Record<number, Product[]>>({})
const seckillActivities = ref<SeckillActivity[]>([])

// 显示前5个主分类
const displayCategories = computed(() => mainCategories.value.slice(0, 5))

// 秒杀倒计时
const countdown = reactive({ h: 2, m: 15, s: 46 })
let timer: ReturnType<typeof setInterval> | null = null
const pad = (n: number) => String(n).padStart(2, '0')
const tick = () => {
  if (countdown.s > 0) { countdown.s-- }
  else if (countdown.m > 0) { countdown.m--; countdown.s = 59 }
  else if (countdown.h > 0) { countdown.h--; countdown.m = 59; countdown.s = 59 }
  else { if (timer) { clearInterval(timer); timer = null } }
}
const seckillPercent = (act: SeckillActivity) => {
  const stock = Number(act.stock || 1), sold = Number(act.sold || 0)
  return Math.min(Math.round((sold / Math.max(stock + sold, 1)) * 100), 100)
}

// 加购
const addToCart = async (p: Product) => {
  if (!userStore.token) { ElMessage.warning('请先登录'); router.push('/login'); return }
  // 首页商品是 SPU，必须先解析出真实 SKU ID 再加购（不能把商品 ID 当 SKU ID 传）
  const skuId = await getDefaultSkuId(p.id)
  if (!skuId) { ElMessage.warning('该商品暂无可用规格'); return }
  try {
    await addItem({ skuId, quantity: 1 })
    ElMessage.success('已添加到购物车')
    cartStore.fetchCart()
  } catch { ElMessage.error('添加失败') }
}

const placeholderUri = 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200" viewBox="0 0 200 200"><rect fill="#1a1f2e" width="200" height="200"/><text fill="#4a5068" font-size="14" x="50%" y="50%" text-anchor="middle" dominant-baseline="central">无图</text></svg>')

const fmt = (v: any) => { const n = Number(v||0); return n.toFixed(2) }

const resolveUrl = (url: string) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  if (url.startsWith('/')) return 'http://localhost:8080' + url
  return url
}

const getProductImage = (p: Product) => {
  const img = p.local_main_image || p.main_image || (p.images?.[0])
  return img ? resolveUrl(img) : placeholderUri
}
const getSeckillImage = (act: SeckillActivity) => act.sku_image ? resolveUrl(act.sku_image) : placeholderUri

const fetchCategories = async () => {
  try { const r = await getCategoryTree({status:-1}) as any; if (r.code===0&&r.data) mainCategories.value = r.data } catch {}
}
const fetchSeckill = async () => {
  try { const r = await listSeckillActivities({page:1,page_size:6,status:1}) as any; seckillActivities.value = r?.data?.list || [] } catch {}
}
const fetchHot = async (categoryId: number) => {
  try {
    const r = await getProductList({category_id:categoryId,page:1,page_size:8,status:1,is_hot:1}) as any
    if (r.code===0 && r.data?.list) hotByCat.value[categoryId] = r.data.list
  } catch {}
}

onMounted(async () => {
  await fetchCategories()
  fetchSeckill()
  timer = setInterval(tick, 1000)
  const cats = mainCategories.value.slice(0, 5)
  for (const c of cats) await fetchHot(c.id)
})

onUnmounted(() => {
  if (timer) { clearInterval(timer); timer = null }
})
</script>

<style scoped>
.home { --accent:#00F5FF; --accent-dim:rgba(0,245,255,0.1); --text:#EDF0F5; --text-dim:#8890A5; --border:rgba(255,255,255,0.06); --card:rgba(255,255,255,0.03); --radius:14px; --radius-sm:10px; font-family:'Inter','PingFang SC','SF Pro Display',-apple-system,sans-serif; }

/* ======== Hero ======== */
.hero { position:relative; overflow:hidden; min-height:460px; display:flex; align-items:center; justify-content:center; padding:100px 24px 80px; }
.hero-glow { position:absolute; border-radius:50%; filter:blur(120px); opacity:.3; pointer-events:none; }
.hero-glow-1 { width:600px; height:600px; background:radial-gradient(circle,var(--accent),transparent); top:-200px; left:-100px; }
.hero-glow-2 { width:400px; height:400px; background:radial-gradient(circle,rgba(99,102,241,.5),transparent); bottom:-100px; right:-50px; }
.hero-glow-3 { width:300px; height:300px; background:radial-gradient(circle,rgba(6,182,212,.4),transparent); top:50%; left:50%; transform:translate(-50%,-50%); }
.hero-grid { position:absolute; inset:0; background-image:linear-gradient(rgba(255,255,255,.015)1px,transparent 1px),linear-gradient(90deg,rgba(255,255,255,.015)1px,transparent 1px); background-size:60px 60px; pointer-events:none; }
.hero-content { position:relative; z-index:1; text-align:center; max-width:720px; }
.hero-tag { display:inline-block; font-size:11px; font-weight:600; letter-spacing:.15em; color:var(--accent); background:var(--accent-dim); padding:6px 16px; border-radius:100px; margin-bottom:24px; }
.hero-title { font-size:clamp(40px,7vw,72px); font-weight:800; line-height:1.1; letter-spacing:-.02em; margin:0 0 16px; background:linear-gradient(135deg,#fff,#00F5FF,#fff); background-size:200% auto; -webkit-background-clip:text; -webkit-text-fill-color:transparent; background-clip:text; }
.hero-subtitle { font-size:18px; color:var(--text-dim); margin:0 0 40px; }
.hero-actions { display:flex; justify-content:center; gap:16px; margin-bottom:56px; flex-wrap:wrap; }
.cta-btn { display:inline-flex; align-items:center; gap:8px; padding:14px 32px; border-radius:100px; border:none; background:var(--accent); color:#0A0F1C; font-size:15px; font-weight:700; cursor:pointer; transition:all .3s; }
.cta-btn:hover { transform:translateY(-2px); box-shadow:0 8px 32px rgba(0,245,255,.35); }
.cta-arrow { width:18px; height:18px; }
.cta-btn-secondary { padding:14px 28px; border-radius:100px; border:1px solid rgba(255,255,255,.15); background:transparent; color:var(--text); font-size:14px; font-weight:500; cursor:pointer; transition:all .2s; }
.cta-btn-secondary:hover { border-color:rgba(255,255,255,.3); background:rgba(255,255,255,.04); }
.hero-stats { display:flex; align-items:center; justify-content:center; }
.stat-item { display:flex; flex-direction:column; align-items:center; gap:4px; padding:0 32px; }
.stat-num { font-size:26px; font-weight:700; }
.stat-label { font-size:12px; color:var(--text-dim); letter-spacing:.04em; }
.stat-divider { width:1px; height:36px; background:var(--border); }

/* ======== Sections ======== */
.section { max-width:1280px; margin:0 auto; padding:72px 24px 0; }
.section-head { display:flex; justify-content:space-between; align-items:flex-end; margin-bottom:32px; gap:16px; }
.section-head-left { display:flex; flex-direction:column; gap:4px; }
.section-head-right { display:flex; align-items:center; gap:20px; }
.section-title { font-size:24px; font-weight:700; margin:0; color:var(--text); letter-spacing:-.01em; }
.section-desc { font-size:13px; color:var(--text-dim); margin:0; }
.section-link { color:var(--accent); cursor:pointer; font-size:14px; font-weight:500; transition:opacity .2s; white-space:nowrap; }
.section-link:hover { opacity:.75; }

/* Countdown */
.countdown { display:inline-flex; align-items:center; gap:6px; background:rgba(255,59,48,.1); border:1px solid rgba(255,59,48,.2); border-radius:100px; padding:6px 16px; font-size:15px; font-weight:700; color:#FF5A50; letter-spacing:.04em; font-variant-numeric:tabular-nums; }
.countdown-icon { width:16px; height:16px; }

/* Empty */
.empty-block { text-align:center; padding:48px 0; color:var(--text-dim); font-size:14px; }

/* Seckill Grid */
.seckill-grid { display:grid; grid-template-columns:repeat(4,1fr); gap:16px; }
.seckill-card { background:var(--card); border:1px solid var(--border); border-radius:var(--radius); overflow:hidden; cursor:pointer; transition:all .3s; }
.seckill-card:hover { border-color:var(--accent); transform:translateY(-4px); box-shadow:0 12px 40px rgba(0,245,255,.08); }
.seckill-img { position:relative; height:180px; background:#141a2b; display:flex; align-items:center; justify-content:center; overflow:hidden; }
.seckill-img img { width:100%; height:100%; object-fit:cover; transition:transform .4s; }
.seckill-card:hover .seckill-img img { transform:scale(1.06); }
.seckill-badge { position:absolute; top:10px; left:10px; background:var(--accent); color:#0A0F1C; font-size:10px; font-weight:700; padding:4px 10px; border-radius:6px; letter-spacing:.06em; }
.seckill-body { padding:14px 16px; }
.seckill-name { font-size:14px; font-weight:500; margin:0 0 8px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.seckill-price { display:flex; align-items:baseline; gap:8px; margin-bottom:10px; }
.seckill-price strong { font-size:18px; color:var(--accent); }
.seckill-price del { font-size:12px; color:var(--text-dim); }
.seckill-progress { display:flex; align-items:center; gap:8px; }
.progress-bar { flex:1; height:4px; background:rgba(255,255,255,.06); border-radius:2px; overflow:hidden; }
.progress-fill { height:100%; background:linear-gradient(90deg,var(--accent),#FF5A50); border-radius:2px; transition:width .6s; }
.progress-text { font-size:11px; color:var(--text-dim); white-space:nowrap; }

/* Product Grid */
.product-grid { display:grid; grid-template-columns:repeat(4,1fr); gap:16px; }
.product-card { background:var(--card); border:1px solid var(--border); border-radius:var(--radius); overflow:hidden; cursor:pointer; transition:all .3s; position:relative; }
.product-card:hover { border-color:rgba(255,255,255,.12); transform:translateY(-6px); box-shadow:0 20px 60px rgba(0,0,0,.4),0 0 0 1px rgba(0,245,255,.06); }
.product-img { position:relative; height:220px; background:#111827; display:flex; align-items:center; justify-content:center; overflow:hidden; }
.product-img img { width:100%; height:100%; object-fit:cover; transition:transform .5s; }
.product-card:hover .product-img img { transform:scale(1.08); }
.discount-tag { position:absolute; top:10px; left:10px; background:var(--accent); color:#0A0F1C; font-size:11px; font-weight:700; padding:4px 10px; border-radius:6px; z-index:2; }
/* Hover overlay + 加购按钮 */
.product-overlay { position:absolute; inset:0; background:linear-gradient(to top,rgba(10,15,28,.85) 0%,transparent 60%); display:flex; align-items:flex-end; justify-content:center; padding:16px; opacity:0; transition:opacity .3s; z-index:3; }
.product-card:hover .product-overlay { opacity:1; }
.add-cart-btn { display:inline-flex; align-items:center; gap:6px; padding:10px 22px; border-radius:100px; border:1px solid var(--accent); background:var(--accent); color:#0A0F1C; font-size:13px; font-weight:600; cursor:pointer; transition:all .25s; transform:translateY(8px); }
.product-card:hover .add-cart-btn { transform:translateY(0); }
.add-cart-btn:hover { box-shadow:0 0 24px rgba(0,245,255,.4); }
.add-cart-icon { width:16px; height:16px; }
.product-body { padding:16px; }
.product-name { font-size:14px; font-weight:500; margin:0 0 4px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; color:var(--text); }
.product-subtitle { font-size:12px; color:var(--text-dim); margin:0 0 8px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.product-price-row { display:flex; align-items:baseline; gap:8px; }
.price { font-size:18px; color:var(--accent); font-weight:700; }
.price-old { font-size:12px; color:var(--text-dim); text-decoration:line-through; }

@media (max-width:900px) { .seckill-grid,.product-grid { grid-template-columns:repeat(2,1fr); } .hero-stats { flex-wrap:wrap; gap:20px; } .stat-divider { display:none; } .stat-item { padding:0; } }
@media (max-width:600px) { .seckill-grid,.product-grid { grid-template-columns:1fr; } }
</style>
