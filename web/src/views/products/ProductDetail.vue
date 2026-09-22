<template>
  <div class="detail-page" v-loading="loading">
    <div v-if="product" class="container">
      <!-- ======== 面包屑 ======== -->
      <div class="breadcrumb">
        <a @click="$router.push('/')" class="crumb-link">首页</a>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="crumb-sep"><polyline points="9 18 15 12 9 6"/></svg>
        <a @click="$router.push({path:'/products',query:{category_id:product.category_id}})" class="crumb-link">{{ categoryName }}</a>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="crumb-sep"><polyline points="9 18 15 12 9 6"/></svg>
        <span class="crumb-current">{{ product.name }}</span>
      </div>

      <!-- ======== 商品主区域 ======== -->
      <div class="product-main">
        <!-- 左侧图片 -->
        <div class="product-gallery">
          <div class="main-image" @click="showPreview = true">
            <img :src="currentImage" :alt="product.name" @error="e=>(e.target as HTMLImageElement).src=placeholderUri" />
            <button class="zoom-btn" v-if="imageList.length>1">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/><path d="M11 8v6M8 11h6"/></svg>
            </button>
          </div>
          <div class="thumbnail-row" v-if="imageList.length > 1">
            <div v-for="(img,idx) in imageList" :key="idx" class="thumb" :class="{active:currentImage===img}" @click="currentImage=img">
              <img :src="img" :alt="`图片${idx+1}`" @error="e=>(e.target as HTMLImageElement).src=thumbPlaceholderUri" />
            </div>
          </div>
        </div>

        <!-- 右侧信息 -->
        <div class="product-info">
          <h1 class="product-name">{{ product.name }}</h1>
          <p class="product-subtitle" v-if="product.subtitle">{{ product.subtitle }}</p>

          <!-- 价格 -->
          <div class="price-box">
            <span class="price-current">¥{{ fmt(selectedSku?.price ?? product.price) }}</span>
            <span class="price-original" v-if="(selectedSku?.original_price ?? product.original_price) && (selectedSku?.original_price || product.original_price) > (selectedSku?.price || product.price)">
              原价 ¥{{ fmt(selectedSku?.original_price ?? product.original_price) }}
            </span>
            <span class="price-discount" v-if="product.original_price && product.original_price > product.price">
              {{ Math.round((1-product.price/product.original_price)*100) }}% OFF
            </span>
          </div>

          <!-- SKU 规格 -->
          <div v-if="skus.length > 0" class="sku-section">
            <div v-if="Object.keys(availableSpecs).length === 0" class="sku-group">
              <span class="sku-label">规格</span>
              <span class="sku-option active">默认</span>
            </div>
            <div v-for="(spec, specName) in availableSpecs" :key="specName" class="sku-group">
              <span class="sku-label">{{ specName }}</span>
              <div class="sku-options">
                <button
                  v-for="val in spec.values" :key="val"
                  class="sku-option"
                  :class="{active:selectedSpecs[specName]===val,disabled:!isSkuAvailable(specName,val)}"
                  @click="selectSpec(specName,val)"
                >{{ val }}</button>
              </div>
            </div>
          </div>

          <!-- 库存 + 数量 -->
          <div class="meta-row">
            <div class="meta-item">
              <span class="meta-label">库存</span>
              <span class="meta-value" :class="{'out':!hasStock}">{{ hasStock ? (selectedSku?.stock ?? product.stock)+' 件' : '暂时缺货' }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">数量</span>
              <div class="qty-control">
                <button class="qty-btn" @click="quantity=Math.max(1,quantity-1)" :disabled="!hasStock">−</button>
                <span class="qty-val">{{ quantity }}</span>
                <button class="qty-btn" @click="quantity=Math.min(maxBuy,quantity+1)" :disabled="!hasStock">+</button>
              </div>
            </div>
          </div>

          <!-- 按钮 -->
          <div class="action-buttons">
            <button class="btn-cart" :disabled="!hasStock" :class="{loading:adding}" @click="addToCart">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/></svg>
              {{ adding ? '添加中...' : '加入购物车' }}
            </button>
            <button class="btn-buy" :disabled="!hasStock" @click="buyNow">立即购买</button>
          </div>

          <!-- 服务保障 -->
          <div class="service-box">
            <div v-for="s in services" :key="s" class="service-item">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="service-icon"><polyline points="20 6 9 17 4 12"/></svg>
              {{ s }}
            </div>
          </div>
        </div>
      </div>

      <!-- ======== 详情 Tab ======== -->
      <div class="detail-tabs">
        <div class="tabs-nav">
          <button v-for="t in tabItems" :key="t.key" class="tab-btn" :class="{active:activeTab===t.key}" @click="switchTab(t.key)">{{ t.label }}</button>
        </div>

        <!-- 商品详情 -->
        <div v-show="activeTab==='detail'" class="tab-panel">
          <div class="detail-content" v-html="product.detail || '<p class=&quot;detail-empty-hint&quot;>暂无详情</p>'"></div>
        </div>

        <!-- 规格参数 -->
        <div v-show="activeTab==='specs'" class="tab-panel">
          <div class="specs-grid">
            <div class="spec-row"><span class="spec-label">商品名称</span><span>{{ product.name }}</span></div>
            <div class="spec-row"><span class="spec-label">商品分类</span><span>{{ categoryName }}</span></div>
            <div class="spec-row"><span class="spec-label">商品价格</span><span class="accent">¥{{ fmt(product.price) }}</span></div>
            <div class="spec-row"><span class="spec-label">库存数量</span><span>{{ product.stock }} 件</span></div>
            <div class="spec-row"><span class="spec-label">销量</span><span>{{ product.sales }}</span></div>
            <div class="spec-row"><span class="spec-label">状态</span><span class="accent">{{ product.status===1?'在售':'已下架' }}</span></div>
          </div>
        </div>

        <!-- 评价 -->
        <div v-show="activeTab==='reviews'" class="tab-panel">
          <!-- 评分概览 -->
          <div class="review-summary" v-if="reviewStats">
            <div class="review-avg">
              <span class="avg-num">{{ reviewStats.average_rating.toFixed(1) }}</span>
              <div class="avg-stars">
                <span v-for="i in 5" :key="i" class="star" :class="{filled:i<=Math.round(reviewStats.average_rating)}">★</span>
              </div>
              <span class="avg-count">{{ reviewStats.total_count }} 条评价</span>
            </div>
            <div class="review-dist">
              <div v-for="i in 5" :key="i" class="dist-row">
                <span>{{ i }}星</span>
                <div class="dist-bar"><div class="dist-fill" :style="{width:((reviewStats as any)['rating_'+i+'_count']/Math.max(reviewStats.total_count,1)*100)+'%'}"></div></div>
                <span class="dist-count">{{ (reviewStats as any)['rating_'+i+'_count'] }}</span>
              </div>
            </div>
          </div>

          <!-- 评分筛选 -->
          <div class="review-filter">
            <select v-model="reviewRatingFilter" class="sort-select" @change="fetchReviews">
              <option :value="0">全部评价</option><option :value="5">5星</option><option :value="4">4星</option><option :value="3">3星</option><option :value="2">2星</option><option :value="1">1星</option>
            </select>
          </div>

          <!-- 写评价 -->
          <div class="write-review">
            <h4>我要评价</h4>
            <div class="write-row">
              <span class="write-label">评分</span>
              <div class="star-input"><span v-for="i in 5" :key="i" class="star" :class="{filled:i<=myReviewRating}" @click="myReviewRating=i">★</span></div>
            </div>
            <textarea v-model="myReviewContent" class="write-textarea" rows="3" maxlength="300" placeholder="写下你的真实体验吧～"></textarea>
            <button class="btn-submit" :disabled="myReviewSubmitting" @click="submitMyReview">{{ myReviewSubmitting?'提交中...':'提交评价' }}</button>
          </div>

          <!-- 评价列表 -->
          <div v-if="reviewList.length===0" class="empty-block">暂无评价</div>
          <div v-else class="review-list">
            <div v-for="r in reviewList" :key="r.id" class="review-item">
              <div class="review-head">
                <div class="star-row"><span v-for="i in 5" :key="i" class="star small" :class="{filled:i<=r.rating}">★</span></div>
                <span class="review-time">{{ r.created_at }}</span>
              </div>
              <p class="review-text">{{ r.content || '（无文字评价）' }}</p>
              <div v-if="r.reply_content" class="review-reply"><strong>商家回复：</strong>{{ r.reply_content }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- ======== 相关推荐 ======== -->
      <div v-if="relatedProducts.length > 0" class="related-section">
        <h3 class="section-title">相关推荐</h3>
        <div class="product-grid">
          <div v-for="p in relatedProducts" :key="p.id" class="product-card" @click="$router.push(`/products/${p.id}`)">
            <div class="card-img"><img :src="getProductImage(p)" :alt="p.name" @error="e=>(e.target as HTMLImageElement).src=placeholderUri" /></div>
            <div class="card-body">
              <h4 class="card-name">{{ p.name }}</h4>
              <span class="card-price">¥{{ fmt(p.price) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 图片预览 -->
    <div v-if="showPreview" class="preview-overlay" @click="showPreview=false">
      <img :src="currentImage" class="preview-img" @click.stop />
      <button class="preview-close" @click="showPreview=false">✕</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { resolveAssetUrl as resolveUrl } from '@/utils/api'
import { ElMessage } from 'element-plus'
import { getProductDetail, getProductList } from '@/api/product'
import { getSkusByProductId } from '@/api/sku'
import { getCategoryDetail } from '@/api/category'
import { addItem } from '@/api/cart'
import { createReview, getProductReviews, getReviewStats, type Review, type ReviewStats } from '@/api/review'
import { useCartStore } from '@/stores/cart'
import { useUserStore } from '@/stores/user'
import type { Product } from '@/api/product'
import type { Sku } from '@/api/sku'
import { placeholderImage } from '@/utils/placeholder'
import { ensureLogin } from '@/utils/auth'

const route = useRoute(); const router = useRouter()
const userStore = useUserStore(); const cartStore = useCartStore()

const loading = ref(false); const adding = ref(false)
const product = ref<Product|null>(null); const skus = ref<Sku[]>([])
const categoryName = ref(''); const quantity = ref(1)
const currentImage = ref(''); const activeTab = ref('detail'); const showPreview = ref(false)
const relatedProducts = ref<Product[]>([])
const selectedSpecs = ref<Record<string,string>>({})

const reviewLoading = ref(false); const reviewList = ref<Review[]>([])
const reviewTotal = ref(0); const reviewStats = ref<ReviewStats|null>(null)
const reviewRatingFilter = ref<number>(0)
const myReviewSubmitting = ref(false); const myReviewRating = ref(5); const myReviewContent = ref('')

const tabItems = [{key:'detail',label:'商品详情'},{key:'specs',label:'规格参数'},{key:'reviews',label:'商品评价'}]
const services = ['7天无理由退货','满69元包邮','15天免费换货','1100余家售后网点']
const placeholderUri = placeholderImage(400, '暂无图片')
// 缩略图只有 72px，用不带文字的小尺寸占位图
const thumbPlaceholderUri = placeholderImage(160, '')

const fmt = (v:any) => {const n=Number(v||0);return isNaN(n)?'0.00':n.toFixed(2)}
const getImageUrl = (url:string) => resolveUrl(url)
const getProductImage = (p:Product) => {const img=p.local_main_image||p.main_image||(p.images?.[0]);return img?resolveUrl(img):placeholderUri}

const availableSpecs = computed(() => {
  const out:Record<string,{values:string[]}> = {}
  if(!skus.value?.length) return out
  const all = skus.value.map(s=>(s?.specs||{}) as Record<string,string>)
  let keys:string[] = Object.keys(all[0]||{})
  for(let i=1;i<all.length;i++){const s=new Set(Object.keys(all[i]||{}));keys=keys.filter(k=>s.has(k))}
  skus.value.forEach(sku=>{if(!sku||sku.status!==1)return;const sp=(sku.specs||{}) as Record<string,string>;keys.forEach(k=>{const v=sp[k];if(!v)return;if(!out[k])out[k]={values:[]};if(!out[k].values.includes(v))out[k].values.push(v)})})
  return out
})

const selectedSku = computed(() => {
  if(!skus.value.length) return null
  const keys = Object.keys(availableSpecs.value)
  if(!keys.length || Object.keys(selectedSpecs.value).length < keys.length){
    const inStock = skus.value.find(s=>s&&s.status===1&&s.stock>0&&Object.entries(selectedSpecs.value).every(([k,v])=>(s.specs as any)?.[k]===v))
    return inStock||skus.value[0]
  }
  return skus.value.find(s=>s&&keys.every(k=>(s.specs as any)?.[k]===selectedSpecs.value[k]))||null
})

const imageList = computed(() => {
  const imgs:string[] = []; const p=product.value; if(!p) return [placeholderUri]
  if(p.local_main_image) imgs.push(getImageUrl(p.local_main_image)); else if(p.main_image) imgs.push(getImageUrl(p.main_image))
  const local = p.local_images?.length?p.local_images:p.images; if(local?.length) local.forEach((i:string)=>imgs.push(getImageUrl(i)))
  if(selectedSku.value?.image) imgs.push(getImageUrl(selectedSku.value.image))
  return imgs.length?imgs:[placeholderUri]
})

const hasStock = computed(() => {
  if(!skus.value?.length) return product.value?product.value.stock>0:false
  return selectedSku.value?selectedSku.value.stock>0:false
})

const maxBuy = computed(() => Math.max(1,selectedSku.value?.stock??product.value?.stock??999))

const isSkuAvailable = (specName:string,specValue:string) => {
  if(selectedSpecs.value[specName]===specValue) return true
  const temp = {...selectedSpecs.value,[specName]:specValue}
  return skus.value.some(s=>s&&s.status===1&&s.stock>0&&Object.entries(temp).every(([k,v])=>(s.specs as any)?.[k]===v))
}

const selectSpec = (name:string,val:string) => {
  if(selectedSpecs.value[name]===val) return
  if(!isSkuAvailable(name,val)){ElMessage.warning('该规格暂无库存');return}
  selectedSpecs.value[name]=val
  if(selectedSku.value?.image) currentImage.value=getImageUrl(selectedSku.value.image)
}

const switchTab = (key:string) => { activeTab.value=key; if(key==='reviews') fetchReviews() }

const addToCart = async () => {
  if(!ensureLogin(router, '请先登录后再加入购物车'))return
  if(!product.value) return
  // 必须以真实的 SKU ID 为准；若拿不到 SKU，不能把商品 ID 当成 SKU ID 传给后端，否则会串到其它商品
  const skuId = selectedSku.value?.id
  if(!skuId){ElMessage.warning('暂无可用规格');return}
  adding.value=true
  try{await addItem({skuId,quantity:quantity.value});ElMessage.success('已添加到购物车');cartStore.fetchCart()}
  catch(e:any){ElMessage.error(e.message||'添加失败')}
  finally{adding.value=false}
}

const buyNow = async () => {
  if(!ensureLogin(router, '请先登录后再下单'))return
  if(!product.value) return
  const id = selectedSku.value?.id
  if(!id){ElMessage.warning('暂无可用规格');return}
  try{await addItem({skuId:id,quantity:quantity.value});router.push({path:'/orders/create',query:{sku_id:id.toString(),quantity:quantity.value.toString()}})}
  catch(e:any){ElMessage.error(e.message||'操作失败')}
}

const fetchProduct = async () => {
  loading.value=true
  try{
    const pid=Number(route.params.id)
    const [pr,sk]=await Promise.all([getProductDetail(pid),getSkusByProductId(pid)])
    if(pr.code===0&&pr.data){
      product.value=pr.data
      currentImage.value = pr.data.local_main_image?getImageUrl(pr.data.local_main_image):pr.data.main_image?getImageUrl(pr.data.main_image):(pr.data.images?.[0]||placeholderUri)
      if(pr.data.category_id){try{const cr=await getCategoryDetail(pr.data.category_id);if(cr.code===0&&cr.data)categoryName.value=cr.data.name}catch{}}
      try{const rr=await getProductList({category_id:pr.data.category_id,page:1,page_size:4,status:1}) as any;if(rr.code===0&&rr.data)relatedProducts.value=(rr.data.list||[]).filter((p:Product)=>p.id!==pid).slice(0,4)}catch{}
    }
    if(sk.code===0&&sk.data){
      let list:any[]=[]; if(Array.isArray(sk.data))list=sk.data; else if((sk.data as any)?.list)list=(sk.data as any).list
      const normSpecs=(raw:any):Record<string,string>=>{if(!raw)return{};if(typeof raw==='string'){try{const o=JSON.parse(raw);const r:Record<string,string>={};Object.entries(o).forEach(([k,v])=>r[k]=String(v));return r}catch{return{}}}if(typeof raw==='object'){const r:Record<string,string>={};Object.entries(raw as any).forEach(([k,v])=>r[k]=String(v));return r}return{}}
      skus.value=list.map((s:any)=>({id:Number(s.id),product_id:Number(s.product_id??s.productId??0),sku_code:s.sku_code??s.skuCode??'',name:s.name??'',specs:normSpecs(s.specs),price:Number(s.price??0),original_price:s.original_price??s.originalPrice,stock:Number(s.stock??0),image:s.image??'',weight:s.weight,volume:s.volume,status:Number(s.status??0)}))
      const first = skus.value.find((s:any)=>s.status===1&&s.stock>0)||skus.value[0]
      if(first){selectedSpecs.value={...(first.specs||{})};if(first.image)currentImage.value=getImageUrl(first.image)}
    }
  }catch(e){console.error(e);ElMessage.error('加载失败')}
  finally{loading.value=false}
}

const fetchReviews = async () => {
  const pid=Number(route.params.id);if(!pid)return;reviewLoading.value=true
  try{const[st,lr]=await Promise.all([getReviewStats(pid),getProductReviews(pid,{page:1,page_size:20,rating:reviewRatingFilter.value})]);if(st.code===0)reviewStats.value=st.data;if(lr.code===0){reviewList.value=lr.data||[];reviewTotal.value=Number(lr.total??0)}}
  catch{}finally{reviewLoading.value=false}
}

const submitMyReview = async () => {
  if(!ensureLogin(router, '请先登录后再评价'))return
  const pid=Number(route.params.id);if(!pid){ElMessage.error('商品信息异常');return}
  const c=(myReviewContent.value||'').trim();if(!c){ElMessage.warning('请输入评价内容');return}
  myReviewSubmitting.value=true
  try{await createReview({user_id:Number(userStore.userId),order_id:0,order_item_id:0,product_id:pid,sku_id:Number(selectedSku.value?.id??0),rating:myReviewRating.value,content:c,images:[],videos:[]});ElMessage.success('评价已提交');myReviewContent.value='';myReviewRating.value=5;await fetchReviews()}
  catch(e:any){ElMessage.error(e?.message||'提交失败')}
  finally{myReviewSubmitting.value=false}
}

watch(()=>route.params.id,(n,o)=>{if(n&&n!==o){product.value=null;skus.value=[];categoryName.value='';relatedProducts.value=[];selectedSpecs.value={};quantity.value=1;currentImage.value='';activeTab.value='detail';reviewList.value=[];reviewStats.value=null;reviewRatingFilter.value=0;fetchProduct()}})
watch(selectedSku,(s)=>{if(s?.image)currentImage.value=getImageUrl(s.image)})
onMounted(()=>fetchProduct())
</script>

<style scoped>
.detail-page { min-height:calc(100vh-64px); padding:24px; font-family:var(--gf-font); }
.container { max-width:1200px; margin:0 auto; }

/* Breadcrumb */
.breadcrumb { display:flex;align-items:center;gap:8px;margin-bottom:28px;font-size:13px; }
.crumb-link { color:var(--text-dim);cursor:pointer;transition:color .2s; }
.crumb-link:hover { color:var(--accent); }
.crumb-sep { width:14px;height:14px;color:var(--text-dim); }
.crumb-current { color:var(--text); }

/* ======== Main Product ======== */
.product-main { display:flex;gap:48px;margin-bottom:40px; }

/* Gallery */
.product-gallery { width:480px;flex-shrink:0; }
.main-image { position:relative;width:100%;height:480px;background:linear-gradient(150deg,rgba(255,255,255,.055),rgba(255,255,255,.012));-webkit-backdrop-filter:blur(var(--gf-blur)) saturate(var(--gf-saturate));backdrop-filter:blur(var(--gf-blur)) saturate(var(--gf-saturate));border:1px solid var(--gf-stroke);border-radius:var(--radius);box-shadow:var(--gf-shadow-2),var(--gf-inner-shadow);overflow:hidden;display:flex;align-items:center;justify-content:center;cursor:zoom-in; }
.main-image img { width:100%;height:100%;object-fit:contain;transition:transform .3s; }
.main-image:hover img { transform:scale(1.03); }
.zoom-btn { position:absolute;bottom:12px;right:12px;width:36px;height:36px;border-radius:50%;border:1px solid var(--gf-stroke-strong);background:var(--gf-glass-deep);-webkit-backdrop-filter:blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));backdrop-filter:blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));color:var(--text);cursor:pointer;display:flex;align-items:center;justify-content:center;transition:all .2s; }
.zoom-btn svg { width:18px;height:18px; }
.zoom-btn:hover { border-color:rgba(79,216,255,.5);color:var(--accent); }
.thumbnail-row { display:flex;gap:10px;margin-top:12px; }
.thumb { width:72px;height:72px;border-radius:var(--radius-sm);overflow:hidden;border:1px solid var(--gf-stroke);cursor:pointer;transition:all .2s;background:var(--gf-glass-1);box-shadow:var(--gf-inner-shadow-soft); }
.thumb:hover { border-color:var(--gf-stroke-strong); }
.thumb.active { border-color:rgba(79,216,255,.65);box-shadow:var(--gf-inner-shadow-soft),0 0 0 3px var(--accent-dim); }
.thumb img { width:100%;height:100%;object-fit:cover; }

/* Product Info */
.product-info { flex:1;min-width:0; }
.product-name { font-size:26px;font-weight:700;color:var(--text);margin:0 0 8px;line-height:1.3;letter-spacing:-.01em; }
.product-subtitle { font-size:15px;color:var(--text-dim);margin:0 0 24px; }

/* Price */
.price-box { display:flex;align-items:baseline;gap:14px;padding:20px 24px;background:var(--gf-glass-2);-webkit-backdrop-filter:blur(var(--gf-blur)) saturate(var(--gf-saturate));backdrop-filter:blur(var(--gf-blur)) saturate(var(--gf-saturate));border:1px solid var(--gf-stroke);border-radius:var(--radius-sm);box-shadow:var(--gf-inner-shadow);margin-bottom:28px; }
.price-current { font-size:36px;font-weight:800;color:var(--price);letter-spacing:-.02em; }
.price-original { font-size:14px;color:var(--text-dim);text-decoration:line-through; }
.price-discount { font-size:12px;font-weight:700;background:var(--gf-gradient);color:#04121a;padding:3px 10px;border-radius:var(--radius-xs);margin-left:auto;box-shadow:var(--gf-inner-shadow-soft); }

/* SKU */
.sku-section { margin-bottom:24px; }
.sku-group { margin-bottom:14px; }
.sku-label { display:inline-block;font-size:13px;color:var(--text-dim);margin-right:12px;width:40px;vertical-align:top;padding-top:6px; }
.sku-options { display:inline-flex;flex-wrap:wrap;gap:8px; }
.sku-option { padding:8px 18px;border-radius:var(--radius-pill);border:1px solid var(--gf-stroke);background:var(--gf-glass-1);color:var(--text-dim);font-size:13px;cursor:pointer;box-shadow:var(--gf-inner-shadow-soft);transition:background .2s,border-color .2s,color .2s,transform .2s; }
.sku-option:hover:not(.disabled):not(.active) { border-color:var(--gf-stroke-strong);background:var(--gf-glass-2);color:var(--text);transform:translateY(-1px); }
.sku-option.active { background:var(--gf-gradient);border-color:transparent;color:#04121a;font-weight:700;box-shadow:var(--gf-inner-shadow-soft),0 8px 20px -10px var(--accent-glow); }
.sku-option.disabled { opacity:.3;cursor:not-allowed;text-decoration:line-through; }

/* Meta */
.meta-row { display:flex;gap:32px;align-items:center;margin-bottom:28px; }
.meta-item { display:flex;align-items:center;gap:10px; }
.meta-label { font-size:13px;color:var(--text-dim); }
.meta-value { font-size:14px;color:var(--text);font-weight:500; }
.meta-value.out { color:var(--danger); }
.qty-control { display:flex;align-items:center;gap:0;background:var(--gf-glass-1);border:1px solid var(--gf-stroke);border-radius:var(--radius-pill);box-shadow:var(--gf-inner-shadow-soft);overflow:hidden; }
.qty-btn { width:36px;height:36px;border:none;background:transparent;color:var(--text);font-size:18px;cursor:pointer;transition:all .15s;display:flex;align-items:center;justify-content:center; }
.qty-btn:hover:not(:disabled) { background:var(--accent-dim);color:var(--accent); }
.qty-btn:disabled { opacity:.3;cursor:not-allowed; }
.qty-val { width:40px;text-align:center;font-size:15px;font-weight:600;color:var(--text); }

/* Buttons */
.action-buttons { display:flex;gap:14px;margin-bottom:28px; }
.btn-cart,.btn-buy { flex:1;padding:16px;border-radius:100px;font-size:16px;font-weight:600;cursor:pointer;transition:all .2s;display:flex;align-items:center;justify-content:center;gap:8px; }
.btn-cart { border:1px solid var(--gf-stroke-strong);background:var(--gf-glass-1);color:var(--text);-webkit-backdrop-filter:blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));backdrop-filter:blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));box-shadow:var(--gf-inner-shadow-soft); }
.btn-cart:hover:not(:disabled) { border-color:rgba(79,216,255,.5);color:var(--accent);background:var(--gf-glass-2); }
.btn-cart svg { width:20px;height:20px; }
.btn-buy { border:none;background:var(--gf-gradient);color:#04121a;box-shadow:var(--gf-inner-shadow-soft),0 12px 32px -14px var(--accent-glow); }
.btn-buy:hover:not(:disabled) { filter:brightness(1.06);box-shadow:var(--gf-inner-shadow-soft),0 20px 44px -16px var(--accent-glow);transform:translateY(-1px); }
.btn-cart:disabled,.btn-buy:disabled { opacity:.35;cursor:not-allowed; }
.btn-cart.loading { opacity:.7; }

/* Service */
.service-box { display:flex;flex-wrap:wrap;gap:20px;padding:18px 20px;background:var(--gf-glass-1);border:1px solid var(--gf-stroke);border-radius:var(--radius-sm);box-shadow:var(--gf-inner-shadow-soft); }
.service-item { display:flex;align-items:center;gap:6px;font-size:13px;color:var(--text-dim); }
.service-icon { width:16px;height:16px;color:var(--success);flex-shrink:0; }

/* ======== Tabs ======== */
.detail-tabs { margin-bottom:40px; }
.tabs-nav { display:flex;gap:0;border-bottom:1px solid var(--gf-stroke);margin-bottom:0; }
.tab-btn { padding:14px 28px;border:none;background:none;color:var(--text-dim);font-size:14px;font-weight:500;cursor:pointer;border-bottom:2px solid transparent;transition:all .15s;margin-bottom:-1px; }
.tab-btn:hover { color:var(--text); }
.tab-btn.active { color:var(--accent);border-bottom-color:transparent;background:linear-gradient(90deg,transparent,var(--gf-gradient),transparent) bottom/100% 2px no-repeat; }
.tab-panel { padding:24px 0; }
.detail-content { color:var(--text-dim);line-height:1.8;font-size:14px; }
.detail-content :deep(img) { max-width:100%;border-radius:var(--radius-sm); }
.detail-content :deep(.detail-empty-hint) { color:var(--text-dim);margin:0; }

/* Specs */
.specs-grid { max-width:600px; }
.spec-row { display:flex;justify-content:space-between;padding:14px 0;border-bottom:1px solid var(--border);font-size:14px;color:var(--text); }
.spec-label { color:var(--text-dim); }
.accent { color:var(--accent);font-weight:600; }

/* Reviews */
.review-summary { display:flex;gap:40px;padding:20px;background:var(--gf-glass-2);border:1px solid var(--gf-stroke);border-radius:var(--radius-sm);box-shadow:var(--gf-inner-shadow);margin-bottom:20px; }
.review-avg { display:flex;flex-direction:column;align-items:center;gap:6px;min-width:100px; }
.avg-num { font-size:42px;font-weight:800;color:var(--price);line-height:1; }
.star { color:rgba(255,255,255,.1);font-size:16px; }
.star.filled { color:var(--gf-warning); }
.star.small { font-size:13px; }
.avg-count { font-size:12px;color:var(--text-dim); }
.review-dist { flex:1;display:flex;flex-direction:column;gap:6px; }
.dist-row { display:flex;align-items:center;gap:8px;font-size:12px;color:var(--text-dim); }
.dist-bar { flex:1;height:6px;background:rgba(255,255,255,.06);border-radius:3px;overflow:hidden; }
.dist-fill { height:100%;background:var(--gf-gradient);border-radius:3px;box-shadow:0 0 8px -2px var(--accent-glow);transition:width .4s; }
.dist-count { width:24px;text-align:right; }

.review-filter { margin-bottom:16px; }
.sort-select { appearance:none;-webkit-appearance:none;background:var(--gf-glass-1);border:1px solid var(--gf-stroke);border-radius:var(--radius-pill);padding:8px 32px 8px 14px;color:var(--text);font-size:13px;cursor:pointer;outline:none;box-shadow:var(--gf-inner-shadow-soft);background-image:url("data:image/svg+xml,%3Csvg viewBox='0 0 24 24' fill='none' stroke='%239BA7C0' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E");background-repeat:no-repeat;background-position:right 10px center;background-size:16px;font-family:var(--gf-font); }
.sort-select option { background:#111827;color:var(--text); }

.write-review { padding:20px;background:var(--gf-glass-2);border:1px solid var(--gf-stroke);border-radius:var(--radius-sm);box-shadow:var(--gf-inner-shadow);margin-bottom:20px; }
.write-review h4 { margin:0 0 14px;font-size:14px;font-weight:600;color:var(--text); }
.write-row { display:flex;align-items:center;gap:12px;margin-bottom:12px; }
.write-label { font-size:13px;color:var(--text-dim); }
.star-input .star { cursor:pointer;font-size:22px;transition:all .1s; }
.star-input .star:hover { transform:scale(1.15); }
.write-textarea { width:100%;padding:12px;border-radius:var(--radius-sm);background:var(--gf-glass-1);border:1px solid var(--gf-stroke);color:var(--text);font-size:13px;resize:vertical;outline:none;font-family:inherit;box-sizing:border-box;box-shadow:var(--gf-inner-shadow-soft);transition:border-color .2s,box-shadow .2s; }
.write-textarea:focus { border-color:rgba(79,216,255,.5);box-shadow:var(--gf-inner-shadow-soft),0 0 0 3px var(--accent-dim); }
.write-textarea::placeholder { color:var(--text-dim); }
.btn-submit { margin-top:12px;padding:10px 24px;border-radius:var(--radius-pill);border:none;background:var(--gf-gradient);color:#04121a;font-size:13px;font-weight:700;cursor:pointer;box-shadow:var(--gf-inner-shadow-soft),0 10px 26px -12px var(--accent-glow);transition:filter .2s,box-shadow .2s; }
.btn-submit:hover:not(:disabled) { filter:brightness(1.06);box-shadow:var(--gf-inner-shadow-soft),0 16px 34px -14px var(--accent-glow); }
.btn-submit:disabled { opacity:.5;cursor:not-allowed; }

.review-list { display:flex;flex-direction:column;gap:12px; }
.review-item { padding:16px;background:var(--gf-glass-1);border:1px solid var(--gf-stroke);border-radius:var(--radius-sm);box-shadow:var(--gf-inner-shadow-soft); }
.review-head { display:flex;justify-content:space-between;align-items:center;margin-bottom:8px; }
.review-time { font-size:12px;color:var(--text-dim); }
.review-text { font-size:14px;color:var(--text);margin:0;line-height:1.6; }
.review-reply { margin-top:10px;padding:10px 14px;background:rgba(79,216,255,.05);border:1px solid rgba(79,216,255,.12);border-radius:var(--radius-sm);font-size:13px;color:var(--text-dim); }
.review-reply strong { color:var(--accent); }

.empty-block { text-align:center;padding:48px;color:var(--text-dim);font-size:14px; }

/* Related */
.related-section { margin-bottom:60px; }
.section-title { font-size:22px;font-weight:700;color:var(--text);margin:0 0 20px;display:flex;align-items:center;gap:10px; }
.section-title::before { content:'';width:3px;height:18px;border-radius:2px;background:var(--gf-gradient);box-shadow:0 0 12px var(--accent-glow); }
.product-grid { display:grid;grid-template-columns:repeat(4,1fr);gap:16px; }
.product-card { background:var(--gf-glass-2);-webkit-backdrop-filter:blur(var(--gf-blur)) saturate(var(--gf-saturate));backdrop-filter:blur(var(--gf-blur)) saturate(var(--gf-saturate));border:1px solid var(--gf-stroke);border-radius:var(--radius);box-shadow:var(--gf-inner-shadow);overflow:hidden;cursor:pointer;transition:transform .35s cubic-bezier(.22,1,.36,1),border-color .3s,box-shadow .35s; }
.product-card:hover { border-color:var(--gf-stroke-strong);transform:translateY(-5px);box-shadow:var(--gf-shadow-3),var(--gf-inner-shadow); }
.card-img { height:180px;background:linear-gradient(150deg,rgba(255,255,255,.05),rgba(255,255,255,.01));display:flex;align-items:center;justify-content:center;overflow:hidden; }
.card-img img { width:100%;height:100%;object-fit:cover;transition:transform .4s; }
.product-card:hover .card-img img { transform:scale(1.06); }
.card-body { padding:14px; }
.card-name { font-size:14px;font-weight:500;color:var(--text);margin:0 0 8px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap; }
.card-price { font-size:16px;color:var(--price);font-weight:700; }

/* Preview overlay */
.preview-overlay { position:fixed;inset:0;z-index:200;background:rgba(4,6,12,.82);-webkit-backdrop-filter:blur(var(--gf-blur)) saturate(var(--gf-saturate));backdrop-filter:blur(var(--gf-blur)) saturate(var(--gf-saturate));display:flex;align-items:center;justify-content:center;cursor:zoom-out; }
.preview-img { max-width:90vw;max-height:90vh;object-fit:contain;cursor:default; }
.preview-close { position:absolute;top:24px;right:24px;width:44px;height:44px;border-radius:50%;border:1px solid var(--gf-stroke-strong);background:var(--gf-glass-deep);-webkit-backdrop-filter:blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));backdrop-filter:blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));color:var(--text);font-size:20px;cursor:pointer;box-shadow:var(--gf-inner-shadow-soft);transition:all .2s; }
.preview-close:hover { background:var(--gf-glass-3);border-color:rgba(79,216,255,.5);color:var(--accent); }

@media(max-width:900px) { .product-main { flex-direction:column; } .product-gallery { width:100%; } .main-image { height:340px; } .product-grid { grid-template-columns:repeat(2,1fr); } .review-summary { flex-direction:column;gap:16px; } }
@media(max-width:500px) { .product-grid { grid-template-columns:1fr; } .action-buttons { flex-direction:column; } }
</style>
