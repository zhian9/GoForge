<template>
  <div class="product-page">
    <!-- ======== 面包屑 ======== -->
    <div class="breadcrumb">
      <a @click="$router.push('/')" class="crumb-link">首页</a>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="crumb-sep"><polyline points="9 18 15 12 9 6"/></svg>
      <span v-if="currentCategory" class="crumb-current">{{ currentCategory.name }}</span>
      <span v-else class="crumb-current">全部商品</span>
      <span v-if="keyword" class="crumb-keyword">— "{{ keyword }}"</span>
    </div>

    <!-- ======== 主体：侧边栏 + 内容 ======== -->
    <div class="layout">
      <!-- 左侧分类 -->
      <aside class="sidebar">
        <div class="sidebar-head">
          <h3>商品分类</h3>
        </div>
        <div class="sidebar-body">
          <div
            v-for="cat in categoryTree"
            :key="cat.id"
            class="cat-group"
          >
            <div
              class="cat-parent"
              :class="{ active: selectedCategoryId === cat.id }"
              @click="selectCategory(cat.id)"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="cat-icon"><rect x="3" y="3" width="18" height="18" rx="3"/><line x1="9" y1="9" x2="15" y2="9"/><line x1="9" y1="13" x2="15" y2="13"/></svg>
              {{ cat.name }}
            </div>
            <div
              v-for="child in cat.children"
              :key="child.id"
              class="cat-child"
              :class="{ active: selectedCategoryId === child.id }"
              @click="selectCategory(child.id)"
            >{{ child.name }}</div>
          </div>
        </div>
      </aside>

      <!-- 右侧内容 -->
      <main class="content">
        <!-- 工具栏 -->
        <div class="toolbar">
          <span class="result-count" v-if="!loading">共找到 <strong>{{ total }}</strong> 件商品</span>
          <span class="result-count" v-else>加载中…</span>
          <div class="toolbar-right">
            <select v-model="sortBy" class="sort-select" @change="handleSortChange">
              <option value="">综合排序</option>
              <option value="price_asc">价格从低到高</option>
              <option value="price_desc">价格从高到低</option>
              <option value="sales_desc">销量从高到低</option>
              <option value="created_desc">最新上架</option>
            </select>
          </div>
        </div>

        <!-- 商品网格 -->
        <div v-if="!loading && products.length > 0" class="product-grid">
          <div v-for="p in products" :key="p.id" class="product-card" @click="$router.push(`/products/${p.id}`)">
            <div class="product-img">
              <img :src="getProductImage(p)" :alt="p.name" @error="onImgError" />
              <span v-if="p.original_price > p.price" class="discount-tag">{{ Math.round((1-p.price/p.original_price)*100) }}%</span>
              <div class="product-overlay">
                <button class="add-cart-btn" @click.stop="addToCart(p)">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="add-cart-icon"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/></svg>
                  加入购物车
                </button>
              </div>
            </div>
            <div class="product-body">
              <h4 class="product-name" :title="p.name">{{ p.name }}</h4>
              <p class="product-subtitle" v-if="p.subtitle" :title="p.subtitle">{{ p.subtitle }}</p>
              <div class="product-price-row">
                <span class="price">¥{{ fmt(p.price) }}</span>
                <span v-if="p.original_price > p.price" class="price-old">¥{{ fmt(p.original_price) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 空状态 -->
        <div v-if="!loading && products.length === 0" class="empty-block">
          <svg viewBox="0 0 80 80" fill="none" class="empty-icon"><rect x="10" y="20" width="60" height="48" rx="8" stroke="currentColor" stroke-width="2"/><path d="M28 44h24M28 52h16" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
          <p class="empty-text">暂无商品</p>
          <p class="empty-desc">换个分类或关键词试试</p>
        </div>

        <!-- 分页 -->
        <div v-if="total > pageSize" class="pagination">
          <button class="page-btn" :disabled="currentPage<=1" @click="goPage(currentPage-1)">← 上一页</button>
          <template v-for="p in pages" :key="p">
            <button v-if="p==='...'" class="page-btn disabled">...</button>
            <button v-else class="page-btn" :class="{active:p===currentPage}" @click="goPage(p as number)">{{ p }}</button>
          </template>
          <button class="page-btn" :disabled="currentPage>=totalPages" @click="goPage(currentPage+1)">下一页 →</button>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useCartStore } from '@/stores/cart'
import { ElMessage } from 'element-plus'
import { getProductList } from '@/api/product'
import { searchProducts } from '@/api/search'
import { getCategoryTree } from '@/api/category'
import { addItem } from '@/api/cart'
import { getDefaultSkuId } from '@/api/sku'
import type { Product } from '@/api/product'
import type { Category } from '@/api/category'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()

const loading = ref(false)
const products = ref<Product[]>([])
const categoryTree = ref<Category[]>([])
const selectedCategoryId = ref<number | null>(null)
const keyword = ref('')
const currentPage = ref(1)
const pageSize = ref(12)
const total = ref(0)
const sortBy = ref('')

const placeholderUri = 'data:image/svg+xml,' + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200" viewBox="0 0 200 200"><rect fill="#1a1f2e" width="200" height="200"/><text fill="#4a5068" font-size="14" x="50%" y="50%" text-anchor="middle" dominant-baseline="central">无图</text></svg>')

const fmt = (v: any) => { const n = Number(v||0); return n.toFixed(2) }
const resolveUrl = (url: string) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  if (url.startsWith('/')) return 'http://localhost:8080' + url
  return url
}

const currentCategory = computed(() => {
  if (!selectedCategoryId.value) return null
  const find = (cats: Category[]): Category|null => {
    for (const c of cats) { if (c.id===selectedCategoryId.value) return c; if (c.children) { const f=find(c.children); if (f) return f } }
    return null
  }
  return find(categoryTree.value)
})

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

const pages = computed(() => {
  const t = totalPages.value; const c = currentPage.value
  if (t <= 7) return Array.from({length:t},(_,i)=>i+1)
  if (c <= 4) return [1,2,3,4,5,'...',t]
  if (c >= t-3) return [1,'...',t-4,t-3,t-2,t-1,t]
  return [1,'...',c-1,c,c+1,'...',t]
})

const getProductImage = (p: Product) => {
  const img = p.local_main_image || p.main_image || (p.images?.[0])
  return img ? resolveUrl(img) : placeholderUri
}
const onImgError = (e: Event) => { (e.target as HTMLImageElement).src = placeholderUri }

const selectCategory = (id: number) => {
  selectedCategoryId.value = id; currentPage.value = 1
  router.push({ path:'/products', query:{ category_id:id } })
  fetchProducts()
}
const goPage = (n: number) => { currentPage.value = n; fetchProducts(); window.scrollTo({top:0,behavior:'smooth'}) }
const handleSortChange = () => { currentPage.value = 1; fetchProducts() }

const addToCart = async (p: Product) => {
  if (!userStore.token) { ElMessage.warning('请先登录'); router.push('/login'); return }
  // 列表商品是 SPU，必须先解析出真实 SKU ID 再加购（不能把商品 ID 当 SKU ID 传）
  const skuId = await getDefaultSkuId(p.id)
  if (!skuId) { ElMessage.warning('该商品暂无可用规格'); return }
  try { await addItem({ skuId, quantity:1 }); ElMessage.success('已添加到购物车'); cartStore.fetchCart() } catch { ElMessage.error('添加失败') }
}

const fetchProducts = async () => {
  loading.value = true
  try {
    // 有关键词时走 Elasticsearch 全文检索；没有关键词时仍走商品列表接口。
    // 这样搜索能用到分词/相关性排序，而普通浏览不受影响。
    if (keyword.value) {
      const r = await searchProducts({
        keyword: keyword.value,
        page: currentPage.value,
        pageSize: pageSize.value,
        categoryId: selectedCategoryId.value ?? undefined,
        sortBy: sortBy.value,
      }) as any
      if (r.code === 0) {
        // ES 返回的是精简文档（只含列表需要的字段），这里映射成列表页使用的 Product 结构
        products.value = (r.data || []).map((it: any) => ({
          id: Number(it.productId),
          name: it.name,
          price: Number(it.price || 0),
          main_image: it.mainImage,
          sales: Number(it.sales || 0),
        })) as unknown as Product[]
        total.value = Number(r.total || (r.data ? r.data.length : 0))
      } else {
        products.value = []; total.value = 0
      }
      return
    }

    const params: any = { page:currentPage.value, page_size:pageSize.value, status:1 }
    if (selectedCategoryId.value) params.category_id = selectedCategoryId.value
    if (keyword.value) params.keyword = keyword.value
    if (sortBy.value) params.sort = sortBy.value
    const r = await getProductList(params) as any
    if (r.code===0 && r.data) { products.value = r.data.list||[]; total.value = Number(r.data.total||0) }
  } catch { products.value = []; total.value = 0 } finally { loading.value = false }
}

const loadCategories = async () => {
  try { const r = await getCategoryTree({status:-1}) as any; if (r.code===0&&r.data) categoryTree.value = r.data } catch {}
}

watch(()=>route.query.category_id, (v)=>{ selectedCategoryId.value=v?Number(v):null; currentPage.value=1; fetchProducts() },{immediate:true})
watch(()=>route.query.keyword, (v)=>{ keyword.value=(v as string)||''; currentPage.value=1; fetchProducts() },{immediate:true})

onMounted(()=>{ loadCategories(); fetchProducts() })
</script>

<style scoped>
.product-page { --accent:#00F5FF; --accent-dim:rgba(0,245,255,0.1); --text:#EDF0F5; --text-dim:#8890A5; --border:rgba(255,255,255,0.06); --card:rgba(255,255,255,0.025); --radius:14px; --radius-sm:10px; min-height:calc(100vh - 64px); padding:24px; max-width:1280px; margin:0 auto; font-family:'Inter','PingFang SC','SF Pro Display',-apple-system,sans-serif; }

/* Breadcrumb */
.breadcrumb { display:flex; align-items:center; gap:8px; padding:8px 0 24px; font-size:13px; }
.crumb-link { color:var(--text-dim); cursor:pointer; transition:color .2s; }
.crumb-link:hover { color:var(--accent); }
.crumb-sep { width:14px; height:14px; color:var(--text-dim); }
.crumb-current { color:var(--text); }
.crumb-keyword { color:var(--accent); }

/* Layout */
.layout { display:flex; gap:24px; align-items:flex-start; }

/* Sidebar */
.sidebar { width:220px; flex-shrink:0; background:var(--card); border:1px solid var(--border); border-radius:var(--radius); position:sticky; top:88px; overflow:hidden; }
.sidebar-head { padding:18px 20px 12px; border-bottom:1px solid var(--border); }
.sidebar-head h3 { font-size:14px; font-weight:600; margin:0; color:var(--text); }
.sidebar-body { padding:8px; }
.cat-parent { display:flex; align-items:center; gap:8px; padding:10px 12px; border-radius:var(--radius-sm); cursor:pointer; font-size:14px; font-weight:500; color:var(--text); transition:all .15s; }
.cat-parent:hover { background:var(--accent-dim); color:var(--accent); }
.cat-parent.active { background:var(--accent-dim); color:var(--accent); font-weight:600; }
.cat-icon { width:16px; height:16px; flex-shrink:0; }
.cat-child { padding:7px 12px 7px 36px; border-radius:var(--radius-sm); cursor:pointer; font-size:13px; color:var(--text-dim); transition:all .15s; }
.cat-child:hover { color:var(--accent); background:var(--accent-dim); }
.cat-child.active { color:var(--accent); font-weight:500; }

/* Content */
.content { flex:1; min-width:0; }

/* Toolbar */
.toolbar { display:flex; justify-content:space-between; align-items:center; margin-bottom:24px; }
.result-count { font-size:13px; color:var(--text-dim); }
.result-count strong { color:var(--text); font-weight:600; }
.sort-select { appearance:none; -webkit-appearance:none; background:var(--card); border:1px solid var(--border); border-radius:100px; padding:8px 32px 8px 14px; color:var(--text); font-size:13px; cursor:pointer; outline:none; background-image:url("data:image/svg+xml,%3Csvg viewBox='0 0 24 24' fill='none' stroke='%238890A5' stroke-width='2' xmlns='http://www.w3.org/2000/svg'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E"); background-repeat:no-repeat; background-position:right 10px center; background-size:16px; transition:border-color .2s; }
.sort-select:focus { border-color:var(--accent); }
.sort-select option { background:#1a1f2e; color:var(--text); }

/* Product Grid */
.product-grid { display:grid; grid-template-columns:repeat(4,1fr); gap:16px; }
.product-card { background:var(--card); border:1px solid var(--border); border-radius:var(--radius); overflow:hidden; cursor:pointer; transition:all .3s; }
.product-card:hover { border-color:rgba(255,255,255,.12); transform:translateY(-6px); box-shadow:0 20px 60px rgba(0,0,0,.4); }
.product-img { position:relative; height:200px; background:#111827; display:flex; align-items:center; justify-content:center; overflow:hidden; }
.product-img img { width:100%; height:100%; object-fit:cover; transition:transform .5s; }
.product-card:hover .product-img img { transform:scale(1.08); }
.discount-tag { position:absolute; top:10px; left:10px; background:var(--accent); color:#0A0F1C; font-size:11px; font-weight:700; padding:4px 10px; border-radius:6px; z-index:2; }
.product-overlay { position:absolute; inset:0; background:linear-gradient(to top,rgba(10,15,28,.85) 0%,transparent 60%); display:flex; align-items:flex-end; justify-content:center; padding:16px; opacity:0; transition:opacity .3s; z-index:3; }
.product-card:hover .product-overlay { opacity:1; }
.add-cart-btn { display:inline-flex; align-items:center; gap:6px; padding:10px 22px; border-radius:100px; border:none; background:var(--accent); color:#0A0F1C; font-size:13px; font-weight:600; cursor:pointer; transition:all .25s; transform:translateY(8px); }
.product-card:hover .add-cart-btn { transform:translateY(0); }
.add-cart-btn:hover { box-shadow:0 0 24px rgba(0,245,255,.4); }
.add-cart-icon { width:16px; height:16px; }
.product-body { padding:14px 16px; }
.product-name { font-size:14px; font-weight:500; margin:0 0 4px; color:var(--text); overflow:hidden; text-overflow:ellipsis; display:-webkit-box; -webkit-line-clamp:2; -webkit-box-orient:vertical; line-height:1.4; }
.product-subtitle { font-size:12px; color:var(--text-dim); margin:0 0 8px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.product-price-row { display:flex; align-items:baseline; gap:8px; }
.price { font-size:18px; color:var(--accent); font-weight:700; }
.price-old { font-size:12px; color:var(--text-dim); text-decoration:line-through; }

/* Empty */
.empty-block { text-align:center; padding:80px 0; }
.empty-icon { width:80px; height:80px; color:var(--text-dim); margin-bottom:16px; opacity:.5; }
.empty-text { font-size:16px; color:var(--text-dim); margin:0 0 8px; }
.empty-desc { font-size:13px; color:var(--text-dim); opacity:.6; margin:0; }

/* Pagination */
.pagination { display:flex; justify-content:center; align-items:center; gap:6px; margin-top:40px; padding-top:24px; border-top:1px solid var(--border); }
.page-btn { min-width:40px; height:38px; border-radius:var(--radius-sm); border:1px solid var(--border); background:var(--card); color:var(--text-dim); font-size:13px; cursor:pointer; transition:all .15s; display:inline-flex; align-items:center; justify-content:center; padding:0 12px; }
.page-btn:hover:not(:disabled):not(.disabled):not(.active) { border-color:rgba(255,255,255,.15); color:var(--text); }
.page-btn.active { background:var(--accent); border-color:var(--accent); color:#0A0F1C; font-weight:700; }
.page-btn:disabled,.page-btn.disabled { opacity:.3; cursor:not-allowed; }

@media (max-width:1100px) { .product-grid { grid-template-columns:repeat(3,1fr); } }
@media (max-width:768px) { .layout { flex-direction:column; } .sidebar { width:100%; position:static; } .product-grid { grid-template-columns:repeat(2,1fr); } }
@media (max-width:500px) { .product-grid { grid-template-columns:1fr; } }
</style>
