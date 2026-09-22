<template>
  <div class="seckill-page">
    <!-- 背景装饰 -->
    <div class="bg-grid"></div>

    <!-- ======== Header ======== -->
    <div class="page-head">
      <div class="head-left">
        <h1 class="page-title">限时秒杀</h1>
        <p class="page-subtitle">每日精选 · 手慢无</p>
      </div>
    </div>

    <!-- ======== 倒计时 ======== -->
    <div class="countdown-bar">
      <div class="countdown-label">
        <span class="cd-dot"></span>
        {{ statusTab === '1' ? '距离结束仅剩' : statusTab === '0' ? '距离开始还有' : '活动' }}
      </div>
      <div class="countdown-digits">
        <div class="cd-block"><span class="cd-num">{{ pad(cd.h) }}</span><span class="cd-unit">时</span></div>
        <span class="cd-colon">:</span>
        <div class="cd-block"><span class="cd-num">{{ pad(cd.m) }}</span><span class="cd-unit">分</span></div>
        <span class="cd-colon">:</span>
        <div class="cd-block"><span class="cd-num">{{ pad(cd.s) }}</span><span class="cd-unit">秒</span></div>
      </div>
    </div>

    <!-- ======== Filter Tabs ======== -->
    <div class="filter-tabs">
      <button v-for="t in tabs" :key="t.value" class="tab-btn" :class="{active:statusTab===t.value}" @click="statusTab=t.value;handleTabChange()">
        <span v-if="t.value==='1'" class="tab-dot live"></span>
        {{ t.label }}
      </button>
    </div>

    <!-- ======== Content ======== -->
    <div class="content" v-loading="loading">
      <div v-if="!loading && activities.length===0" class="empty-block">
        <svg viewBox="0 0 120 120" fill="none" class="empty-icon"><circle cx="60" cy="60" r="48" stroke="currentColor" stroke-width="2" stroke-dasharray="10 5"/><path d="M48 42h24M52 54h16M46 66h28" stroke="currentColor" stroke-width="3" stroke-linecap="round"/></svg>
        <p class="empty-text">暂无秒杀活动</p>
        <p class="empty-desc">精彩活动即将上线，敬请期待</p>
      </div>

      <div v-else class="grid">
        <div v-for="item in activities" :key="String(item.id)" class="card" :class="statusClass(item)" @click="goDetail(item)">
          <!-- 角标 -->
          <div class="corner-ribbon">限时秒杀</div>

          <!-- 图片 -->
          <div class="card-img">
            <img :src="getImage(item)" :alt="item.name" @error="e=>(e.target as HTMLImageElement).src=placeholderUri" />
            <div class="img-shine"></div>
          </div>

          <!-- 信息 -->
          <div class="card-body">
            <h3 class="card-name">{{ item.name }}</h3>
            <p class="card-sku">{{ item.sku_name || '精选商品' }}</p>

            <div class="price-row">
              <span class="price-seckill">¥{{ fmt(item.seckill_price) }}</span>
              <span class="price-original" v-if="item.original_price">¥{{ fmt(item.original_price) }}</span>
            </div>

            <div class="progress-wrap">
              <div class="progress-bar">
                <div class="progress-fill" :class="statusClass(item)" :style="{width:soldPercent(item)+'%'}"></div>
              </div>
              <span class="progress-num">{{ soldPercent(item) }}%</span>
            </div>

            <div class="stock-row">
              <span class="stock-label">剩余 <strong>{{ remain(item) }}</strong> 件</span>
              <span class="status-tag" :class="statusClass(item)">{{ statusText(item) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Pagination -->
      <div v-if="total>pageSize" class="pagination">
        <button class="page-btn" :disabled="page<=1" @click="handlePageChange(page-1)">←</button>
        <span class="page-info">{{ page }} / {{ Math.ceil(total/pageSize) }}</span>
        <button class="page-btn" :disabled="page>=Math.ceil(total/pageSize)" @click="handlePageChange(page+1)">→</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { resolveAssetUrl as resolveUrl } from '@/utils/api'
import { listSeckillActivities, type SeckillActivity } from '@/api/seckill'
import { placeholderImage } from '@/utils/placeholder'

const router = useRouter()
const loading = ref(false)
const activities = ref<SeckillActivity[]>([])
const statusTab = ref<'1'|'0'|'2'|'-1'>('1')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const tabs = [
  {value:'1' as const, label:'进行中'},
  {value:'0' as const, label:'未开始'},
  {value:'2' as const, label:'已结束'},
  {value:'-1' as const, label:'全部'},
]

// 倒计时
const cd = reactive({ h: 21, m: 45, s: 36 })
let timer: ReturnType<typeof setInterval>|null = null
const pad = (n:number) => String(n).padStart(2,'0')
const tick = () => {
  if(cd.s>0) cd.s--
  else if(cd.m>0){ cd.m--; cd.s=59 }
  else if(cd.h>0){ cd.h--; cd.m=59; cd.s=59 }
  else { if(timer){clearInterval(timer);timer=null} }
}

const placeholderUri = placeholderImage(400, '暂无图片')
const fmt = (v:any) => {const n=Number(v||0);return isNaN(n)?'0.00':n.toFixed(2)}
const getImage = (item:SeckillActivity) => item.sku_image?resolveUrl(item.sku_image):placeholderUri

const nowSec = () => Math.floor(Date.now()/1000)
const statusText = (item:SeckillActivity) => {
  const s=Number(item.status??-1)
  if(s===0)return'未开始';if(s===1)return'进行中';if(s===2)return'已结束'
  const st=Number(item.start_time||0),et=Number(item.end_time||0),n=nowSec()
  if(st&&n<st)return'未开始';if(et&&n>et)return'已结束';return'进行中'
}
const statusClass = (item:SeckillActivity) => {const t=statusText(item);if(t==='进行中')return'live';if(t==='未开始')return'pending';return'ended'}
const remain = (item:SeckillActivity) => Math.max(0,Number(item.stock||0)-Number(item.sold||0))
const soldPercent = (item:SeckillActivity) => {
  const stock=Number(item.stock||1),sold=Number(item.sold||0)
  return Math.min(100,Math.round(sold/Math.max(stock+sold,1)*100))
}

const fetchList = async () => {
  loading.value=true
  try{
    const resp = await listSeckillActivities({page:page.value,page_size:pageSize.value,status:Number(statusTab.value)}) as any
    activities.value=resp.data?.list||[];total.value=Number(resp.data?.total||0)
  }finally{loading.value=false}
}
const handleTabChange = () => {page.value=1;fetchList()}
const handlePageChange = (p:number) => {page.value=p;fetchList();window.scrollTo({top:0,behavior:'smooth'})}
const goDetail = (item:SeckillActivity) => router.push(`/seckill/${Number(item.id)}`)

onMounted(()=>{fetchList();timer=setInterval(tick,1000)})
onUnmounted(()=>{if(timer){clearInterval(timer);timer=null}})
</script>

<style scoped>
.seckill-page { position:relative; max-width:1100px; margin:0 auto; padding:32px 24px 60px; min-height:calc(100vh-64px); font-family:var(--gf-font); }

/* Background grid */
.bg-grid { position:fixed; inset:0; pointer-events:none; z-index:0; opacity:.3;
  background-image:linear-gradient(rgba(79,216,255,.025) 1px,transparent 1px),linear-gradient(90deg,rgba(79,216,255,.025) 1px,transparent 1px);
  background-size:80px 80px; mask-image:radial-gradient(ellipse 60% 60% at 50% 0%,black 30%,transparent 70%); -webkit-mask-image:radial-gradient(ellipse 60% 60% at 50% 0%,black 30%,transparent 70%);
}
.page-head,.filter-tabs,.content { position:relative; z-index:1; }

/* Head */
.page-head { display:flex; justify-content:space-between; align-items:flex-end; margin-bottom:24px; }
.page-title { font-size:32px; font-weight:800; margin:0; color:var(--text); letter-spacing:-.02em; }
.page-subtitle { font-size:14px; color:var(--text-dim); margin:6px 0 0; }

/* Countdown */
.countdown-bar { display:flex; align-items:center; gap:20px; padding:18px 24px; background:var(--gf-glass-2); -webkit-backdrop-filter:blur(var(--gf-blur-lg)) saturate(var(--gf-saturate)); backdrop-filter:blur(var(--gf-blur-lg)) saturate(var(--gf-saturate)); border:1px solid var(--gf-stroke-strong); border-radius:var(--radius); box-shadow:var(--gf-shadow-2),var(--gf-inner-shadow); margin-bottom:24px; flex-wrap:wrap; position:relative; z-index:1; }
.countdown-label { display:flex; align-items:center; gap:8px; font-size:15px; font-weight:600; color:var(--danger); }
.cd-dot { width:10px;height:10px;border-radius:50%;background:var(--danger);box-shadow:0 0 10px var(--danger);animation:pulse-dot 1.2s infinite; }
@keyframes pulse-dot { 0%,100%{opacity:1;transform:scale(1)} 50%{opacity:.4;transform:scale(1.4)} }
.countdown-digits { display:flex; align-items:center; gap:4px; }
.cd-block { display:flex; flex-direction:column; align-items:center; gap:2px; }
.cd-num { font-size:32px; font-weight:800; color:#fff; background:linear-gradient(180deg,rgba(255,107,129,.26),rgba(255,107,129,.08)); border:1px solid rgba(255,107,129,.22); border-radius:var(--radius-sm); box-shadow:var(--gf-inner-shadow-soft); padding:4px 12px; min-width:48px; text-align:center; font-variant-numeric:tabular-nums; line-height:1.2; }
.cd-unit { font-size:11px; color:var(--text-dim); text-transform:uppercase; letter-spacing:.06em; }
.cd-colon { font-size:28px; font-weight:700; color:var(--danger); margin:0 2px 14px; }

/* Tabs */
.filter-tabs { display:flex; gap:6px; margin-bottom:28px; }
.tab-btn { padding:10px 24px; border-radius:var(--radius-pill); border:1px solid var(--gf-stroke); background:var(--gf-glass-1); color:var(--text-dim); font-size:14px; font-weight:500; cursor:pointer; box-shadow:var(--gf-inner-shadow-soft); transition:background .2s,border-color .2s,color .2s,transform .2s; display:flex; align-items:center; gap:6px; }
.tab-btn:hover { color:var(--text); border-color:var(--gf-stroke-strong); background:var(--gf-glass-2); transform:translateY(-1px); }
.tab-btn.active { background:var(--gf-gradient); border-color:transparent; color:#04121a; font-weight:700; box-shadow:var(--gf-inner-shadow-soft),0 10px 24px -12px var(--accent-glow); }
.tab-dot { width:8px;height:8px;border-radius:50%; }
.tab-dot.live { background:#04121a; animation:pulse 1.5s infinite; }
@keyframes pulse { 0%,100%{opacity:1} 50%{opacity:.3} }

/* Empty */
.empty-block { text-align:center; padding:100px 0; }
.empty-icon { width:120px;height:120px;color:var(--text-dim);opacity:.2;margin-bottom:24px; }
.empty-text { font-size:18px;color:var(--text-dim);margin:0 0 8px; }
.empty-desc { font-size:13px;color:var(--text-dim);opacity:.6;margin:0; }

/* Grid */
.grid { display:grid; grid-template-columns:repeat(2,1fr); gap:20px; }

/* Card */
.card { background:var(--gf-glass-2); border:1px solid var(--gf-stroke); border-radius:var(--radius); box-shadow:var(--gf-inner-shadow); overflow:hidden; cursor:pointer; transition:all .4s cubic-bezier(.4,0,.2,1); display:flex; position:relative; -webkit-backdrop-filter:blur(var(--gf-blur)) saturate(var(--gf-saturate)); backdrop-filter:blur(var(--gf-blur)) saturate(var(--gf-saturate)); }
.card:hover { border-color:rgba(79,216,255,.35); transform:translateY(-6px); box-shadow:var(--gf-shadow-3),var(--gf-inner-shadow); background:var(--gf-glass-3); }
.card.live { border-left:3px solid var(--danger); }
.card.pending { border-left:3px solid var(--warning); }
.card.ended { border-left:3px solid var(--text-dim); opacity:.72; }

/* Corner Ribbon */
.corner-ribbon { position:absolute; top:14px; right:-32px; background:linear-gradient(135deg,#FF5A6E,#FF8A9B); color:#fff; font-size:10px; font-weight:700; padding:4px 36px; transform:rotate(45deg); z-index:4; letter-spacing:.06em; box-shadow:0 2px 10px rgba(255,107,129,.35); }
.card.ended .corner-ribbon { background:rgba(255,255,255,.1); color:var(--text-dim); }
.card.pending .corner-ribbon { background:linear-gradient(135deg,#E8B45C,#FBD38D); }

/* Image */
.card-img { width:200px;height:200px;flex-shrink:0;background:linear-gradient(150deg,rgba(255,255,255,.05),rgba(255,255,255,.01));position:relative;overflow:hidden;display:flex;align-items:center;justify-content:center; }
.card-img img { width:100%;height:100%;object-fit:cover;transition:transform .5s; }
.card:hover .card-img img { transform:scale(1.08); filter:brightness(1.1); }
.img-shine { position:absolute;top:0;left:-100%;width:60%;height:100%;background:linear-gradient(90deg,transparent,rgba(255,255,255,.05),transparent);transform:skewX(-20deg);transition:left .6s; }
.card:hover .img-shine { left:120%; }

/* Body */
.card-body { flex:1; padding:18px 20px; display:flex; flex-direction:column; gap:8px; min-width:0; }
.card-name { font-size:16px; font-weight:600; color:var(--text); margin:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.card-sku { font-size:12px; color:var(--text-dim); margin:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }

/* Price */
.price-row { display:flex; align-items:baseline; gap:10px; }
.price-seckill { font-size:26px; font-weight:800; background:linear-gradient(135deg,#F3DCAF,#E8C88A 55%,#C9A468); -webkit-background-clip:text; -webkit-text-fill-color:transparent; background-clip:text; letter-spacing:-.01em; filter:drop-shadow(0 0 8px rgba(232,200,138,.25)); }
.price-original { font-size:13px; color:var(--text-dim); text-decoration:line-through; }

/* Progress */
.progress-wrap { display:flex; align-items:center; gap:10px; margin-top:4px; }
.progress-bar { flex:1; height:6px; background:rgba(255,255,255,.07); border-radius:3px; box-shadow:inset 0 1px 2px rgba(0,0,0,.4); overflow:hidden; }
.progress-fill { height:100%; border-radius:3px; transition:width .6s cubic-bezier(.4,0,.2,1); }
.progress-fill.live { background:linear-gradient(90deg,#FF8A9B,#FF5A6E); box-shadow:0 0 8px rgba(255,107,129,.35); }
.progress-fill.pending { background:linear-gradient(90deg,#FBD38D,#E8B45C); }
.progress-fill.ended { background:var(--text-dim); }
.progress-num { font-size:12px; color:var(--text-dim); font-weight:600; min-width:36px; }

.stock-row { display:flex; justify-content:space-between; align-items:center; }
.stock-label { font-size:12px; color:var(--text-dim); }
.stock-label strong { color:var(--text); font-size:13px; }
.status-tag { font-size:11px; font-weight:600; padding:3px 10px; border-radius:100px; }
.status-tag.live { background:rgba(255,107,129,.14); color:var(--danger); }
.status-tag.pending { background:rgba(251,191,107,.14); color:var(--warning); }
.status-tag.ended { background:rgba(255,255,255,.06); color:var(--text-dim); }

/* Pagination */
.pagination { display:flex; justify-content:center; align-items:center; gap:16px; margin-top:48px; position:relative; z-index:1; }
.page-btn { width:44px;height:42px;border-radius:var(--radius-sm);border:1px solid var(--gf-stroke);background:var(--gf-glass-1);color:var(--text);font-size:16px;cursor:pointer;box-shadow:var(--gf-inner-shadow-soft);transition:all .2s;display:flex;align-items:center;justify-content:center; }
.page-btn:hover:not(:disabled) { border-color:rgba(79,216,255,.5);color:var(--accent);background:var(--gf-glass-2); }
.page-btn:disabled { opacity:.3;cursor:not-allowed; }
.page-info { font-size:14px;color:var(--text-dim);min-width:80px;text-align:center; }

@media(max-width:800px) {
  .grid { grid-template-columns:1fr; }
  .card { flex-direction:column; }
  .card-img { width:100%;height:200px; }
  .countdown-bar { flex-direction:column; align-items:flex-start; gap:12px; }
  .cd-num { font-size:24px; min-width:38px; padding:4px 10px; }
  .cd-colon { font-size:22px; margin-bottom:10px; }
}
</style>
