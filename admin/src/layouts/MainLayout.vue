<template>
  <div class="admin-layout">
    <!-- ======== 侧边栏 ======== -->
    <aside class="sidebar" :class="{ collapsed: collapsed }">
      <div class="sidebar-logo" @click="$router.push('/dashboard')">
        <img src="/logo-forge.png" alt="GoForge" class="logo-img" />
        <span class="logo-text" v-show="!collapsed">管理后台</span>
      </div>

      <nav class="sidebar-nav">
        <template v-for="group in menuGroups" :key="group.label">
          <div class="nav-group-label" v-show="!collapsed" v-if="group.label">{{ group.label }}</div>
          <router-link
            v-for="item in group.items"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: isActive(item.path) }"
            :title="collapsed ? item.label : ''"
          >
            <span class="nav-icon" v-html="item.icon"></span>
            <span class="nav-label">{{ item.label }}</span>
          </router-link>
        </template>
      </nav>

      <button class="collapse-btn" @click="collapsed=!collapsed">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" :class="{rotated:collapsed}"><polyline points="15 18 9 12 15 6"/></svg>
      </button>
    </aside>

    <!-- ======== 主区域 ======== -->
    <div class="main-area">
      <header class="topbar">
        <!-- 左侧 -->
        <div class="topbar-left">
          <button class="hamburger" @click="collapsed=!collapsed">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/></svg>
          </button>
          <div class="breadcrumb">
            <span class="bc-root">GoForge</span>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="bc-sep"><polyline points="9 18 15 12 9 6"/></svg>
            <span class="bc-current">{{ currentTitle }}</span>
          </div>
        </div>

        <!-- 右侧 -->
        <div class="topbar-right">
          <!-- 搜索 -->
          <div class="search-box">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="search-icon"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
            <input type="text" placeholder="搜索..." class="search-input" />
          </div>

          <!-- 通知铃铛 -->
          <button class="notif-btn">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 01-3.46 0"/></svg>
            <span class="notif-dot"></span>
          </button>

          <!-- 头像下拉 -->
          <div class="user-dropdown" @click.stop="showMenu=!showMenu">
            <div class="user-avatar">{{ initial }}</div>
            <div class="user-info" v-show="!collapsed">
              <span class="user-name">{{ userStore.userInfo?.username || '管理员' }}</span>
              <span class="user-role">超级管理员</span>
            </div>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="arrow"><polyline points="6 9 12 15 18 9"/></svg>
            <div v-if="showMenu" class="dropdown-menu" @click.stop>
              <div class="dm-header">
                <span class="dm-avatar">{{ initial }}</span>
                <div>
                  <div class="dm-name">{{ userStore.userInfo?.username || '管理员' }}</div>
                  <div class="dm-email">{{ userStore.userInfo?.email || 'admin@goForge.dev' }}</div>
                </div>
              </div>
              <div class="dm-divider"></div>
              <div class="dropdown-item" @click="handleLogout">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 21H5a2 2 0 01-2-2V5a2 2 0 012-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
                退出登录
              </div>
            </div>
          </div>
        </div>
      </header>
      <main class="content"><router-view /></main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessageBox } from 'element-plus'

const route = useRoute(); const router = useRouter(); const userStore = useUserStore()
const collapsed = ref(false); const showMenu = ref(false)
const initial = computed(() => (userStore.userInfo?.username || 'A')[0].toUpperCase())

const menuGroups = [
  { label: '', items: [
    { path: '/dashboard', label: '仪表盘', icon: '<rect x="3" y="3" width="18" height="18" rx="3"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="21" x2="9" y2="9"/>' },
  ]},
  { label: '基础数据', items: [
    { path: '/categories', label: '分类管理', icon: '<rect x="3" y="3" width="18" height="18" rx="3"/><line x1="9" y1="3" x2="9" y2="21"/>' },
    { path: '/banners', label: 'Banner管理', icon: '<rect x="2" y="4" width="20" height="16" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/>' },
    { path: '/users', label: '用户管理', icon: '<circle cx="12" cy="8" r="4"/><path d="M4 20c0-4 4-7 8-7s8 3 8 7"/>' },
    { path: '/products', label: '商品管理', icon: '<path d="M6 2L3 6v14a2 2 0 002 2h14a2 2 0 002-2V6l-3-4zM3 6h18"/><path d="M16 10a4 4 0 01-8 0"/>' },
    { path: '/skus', label: 'SKU管理', icon: '<rect x="2" y="2" width="20" height="20" rx="3"/><rect x="8" y="8" width="8" height="8" rx="1"/>' },
    { path: '/inventory', label: '库存管理', icon: '<rect x="3" y="3" width="18" height="18" rx="3"/><path d="M12 8v8M8 12h8"/>' },
  ]},
  { label: '交易流程', items: [
    { path: '/carts', label: '购物车管理', icon: '<circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 002 1.61h9.72a2 2 0 002-1.61L23 6H6"/>' },
    { path: '/orders', label: '订单管理', icon: '<rect x="2" y="3" width="20" height="18" rx="3"/><line x1="6" y1="9" x2="18" y2="9"/><line x1="6" y1="13" x2="14" y2="13"/>' },
    { path: '/payments', label: '支付管理', icon: '<rect x="2" y="4" width="20" height="16" rx="2"/><line x1="2" y1="10" x2="22" y2="10"/>' },
    { path: '/logistics', label: '物流管理', icon: '<rect x="1" y="3" width="15" height="13"/><polyline points="16 8 20 8 23 11 23 16 16 16 16 8"/><circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/>' },
  ]},
  { label: '营销评价', items: [
    { path: '/promotions', label: '营销管理', icon: '<circle cx="12" cy="12" r="9"/><path d="M12 3v18M3 12h18"/>' },
    { path: '/seckill-activities', label: '秒杀活动', icon: '<polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>' },
    { path: '/reviews', label: '评价管理', icon: '<path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/>' },
  ]},
  { label: '系统管理', items: [
    { path: '/messages', label: '消息管理', icon: '<path d="M21 11.5a8.38 8.38 0 01-.9 3.8 8.5 8.5 0 01-7.6 4.7 8.38 8.38 0 01-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 01-.9-3.8 8.5 8.5 0 014.7-7.6 8.38 8.38 0 013.8-.9h.5a8.48 8.48 0 018 8v.5z"/>' },
  ]},
]

const isActive = (path: string) => route.path === path
const currentTitle = computed(() => {
  for (const g of menuGroups) { const m = g.items.find(i => isActive(i.path)); if (m) return m.label }
  return '管理后台'
})

const handleLogout = async () => {
  try { await ElMessageBox.confirm('确定退出？', '提示', { type: 'warning' }); userStore.logout(); router.push('/login') } catch {}
}
if (typeof document !== 'undefined') document.addEventListener('click', () => { showMenu.value = false })
</script>

<style>
html, body, #app { margin:0; padding:0; background:#0A0F1C; }
</style>

<style scoped>
.admin-layout { --accent:#00F5FF; --accent-dim:rgba(0,245,255,0.1); --sidebar-bg:#0D1320; --topbar-bg:rgba(13,19,32,0.85); --text:#EDF0F5; --text-dim:#8890A5; --border:rgba(255,255,255,0.06); --radius-sm:8px; display:flex; height:100vh; color:var(--text); font-family:'Inter','PingFang SC','SF Pro Display',-apple-system,sans-serif; overflow:hidden; }

/* Sidebar */
.sidebar { width:220px; flex-shrink:0; background:var(--sidebar-bg); border-right:1px solid var(--border); display:flex; flex-direction:column; transition:width .25s; position:relative; z-index:10; }
.sidebar.collapsed { width:64px; }
.sidebar.collapsed .nav-label,.sidebar.collapsed .logo-text,.sidebar.collapsed .nav-group-label { display:none; }
.sidebar.collapsed .nav-item { justify-content:center; padding:12px; }
.sidebar.collapsed .nav-icon { margin:0; }

.sidebar-logo { display:flex; align-items:center; gap:10px; padding:16px 18px; cursor:pointer; border-bottom:1px solid var(--border); }
.logo-img { height:26px; width:auto; }
.logo-text { font-size:14px; font-weight:700; color:var(--text); white-space:nowrap; letter-spacing:-.01em; }

.sidebar-nav { flex:1; overflow-y:auto; padding:8px 10px; }
.sidebar-nav::-webkit-scrollbar { width:4px; }
.sidebar-nav::-webkit-scrollbar-thumb { background:var(--border); border-radius:2px; }

.nav-group-label { font-size:10px; font-weight:600; color:var(--text-dim); text-transform:uppercase; letter-spacing:.08em; padding:16px 8px 6px; }

.nav-item { display:flex; align-items:center; gap:10px; padding:10px 12px; border-radius:var(--radius-sm); text-decoration:none; color:var(--text-dim); font-size:13px; font-weight:500; transition:all .15s; margin-bottom:1px; white-space:nowrap; position:relative; }
.nav-item:hover { color:var(--text); background:rgba(255,255,255,.03); }
.nav-item.active { color:var(--accent); background:var(--accent-dim); }
.nav-item.active::before { content:''; position:absolute; left:0; top:50%; transform:translateY(-50%); width:3px; height:18px; background:var(--accent); border-radius:0 3px 3px 0; }
.nav-icon { width:20px; height:20px; flex-shrink:0; display:flex; align-items:center; justify-content:center; }
.nav-icon :deep(svg) { width:20px; height:20px; fill:none; stroke:currentColor; stroke-width:2; stroke-linecap:round; stroke-linejoin:round; }

.collapse-btn { position:absolute; bottom:10px; right:10px; width:30px; height:30px; border-radius:8px; border:1px solid var(--border); background:var(--sidebar-bg); color:var(--text-dim); cursor:pointer; display:flex; align-items:center; justify-content:center; transition:all .2s; }
.collapse-btn:hover { color:var(--accent); border-color:var(--accent); }
.collapse-btn svg { width:15px; height:15px; transition:transform .25s; }
.collapse-btn svg.rotated { transform:rotate(180deg); }

/* Main */
.main-area { flex:1; display:flex; flex-direction:column; min-width:0; overflow:hidden; }

/* Topbar */
.topbar { height:56px; flex-shrink:0; background:rgba(13,19,32,0.82); backdrop-filter:blur(16px) saturate(180%); -webkit-backdrop-filter:blur(16px) saturate(180%); border-bottom:1px solid var(--border); display:flex; align-items:center; justify-content:space-between; padding:0 20px; }
.topbar-left { display:flex; align-items:center; gap:14px; }
.hamburger { width:34px; height:34px; border-radius:8px; border:1px solid var(--border); background:transparent; color:var(--text-dim); cursor:pointer; display:flex; align-items:center; justify-content:center; transition:all .2s; flex-shrink:0; }
.hamburger:hover { color:var(--accent); border-color:var(--accent); }
.hamburger svg { width:18px; height:18px; }

/* Breadcrumb */
.breadcrumb { display:flex; align-items:center; gap:8px; font-size:13px; }
.bc-root { color:var(--accent); font-weight:600; }
.bc-sep { width:14px; height:14px; color:var(--text-dim); }
.bc-current { color:var(--text-dim); }

/* Topbar Right */
.topbar-right { display:flex; align-items:center; gap:12px; }

/* Search */
.search-box { position:relative; display:flex; align-items:center; }
.search-icon { position:absolute; left:10px; width:15px; height:15px; color:var(--text-dim); pointer-events:none; }
.search-input { width:180px; padding:8px 14px 8px 32px; border-radius:10px; background:rgba(255,255,255,0.05); border:1px solid transparent; color:var(--text); font-size:12px; outline:none; transition:all .25s; font-family:inherit; }
.search-input::placeholder { color:var(--text-dim); }
.search-input:focus { width:240px; border-color:var(--accent); background:rgba(255,255,255,0.08); box-shadow:0 0 0 3px rgba(0,245,255,.08); }

/* Notification */
.notif-btn { position:relative; width:34px; height:34px; border-radius:8px; border:1px solid var(--border); background:transparent; color:var(--text-dim); cursor:pointer; display:flex; align-items:center; justify-content:center; transition:all .2s; flex-shrink:0; }
.notif-btn:hover { color:var(--accent); border-color:var(--accent); }
.notif-btn svg { width:18px; height:18px; }
.notif-dot { position:absolute; top:7px; right:7px; width:8px; height:8px; border-radius:50%; background:#FF3B30; box-shadow:0 0 6px rgba(255,59,48,.5); }

/* User Dropdown */
.user-dropdown { position:relative; display:flex; align-items:center; gap:8px; cursor:pointer; padding:4px 12px 4px 4px; border-radius:100px; transition:background .15s; }
.user-dropdown:hover { background:rgba(255,255,255,.04); }
.user-avatar { width:32px; height:32px; border-radius:50%; background:var(--accent-dim); color:var(--accent); font-weight:700; font-size:14px; display:flex; align-items:center; justify-content:center; flex-shrink:0; }
.user-info { display:flex; flex-direction:column; gap:0; }
.user-name { font-size:13px; color:var(--text); font-weight:500; line-height:1.3; }
.user-role { font-size:10px; color:var(--text-dim); line-height:1.3; }
.arrow { width:14px; height:14px; color:var(--text-dim); }

/* Dropdown */
.dropdown-menu { position:absolute; top:calc(100%+8px); right:0; min-width:200px; background:#111827; border:1px solid var(--border); border-radius:14px; overflow:hidden; box-shadow:0 16px 48px rgba(0,0,0,.5); z-index:50; }
.dm-header { display:flex; align-items:center; gap:10px; padding:14px 16px; }
.dm-avatar { width:36px; height:36px; border-radius:50%; background:var(--accent-dim); color:var(--accent); font-weight:700; font-size:15px; display:flex; align-items:center; justify-content:center; }
.dm-name { font-size:13px; color:var(--text); font-weight:600; }
.dm-email { font-size:11px; color:var(--text-dim); }
.dm-divider { height:1px; background:var(--border); }
.dropdown-item { display:flex; align-items:center; gap:8px; padding:10px 16px; cursor:pointer; font-size:13px; color:#F87171; transition:background .15s; }
.dropdown-item svg { width:16px; height:16px; }
.dropdown-item:hover { background:rgba(248,113,113,.1); }

/* Content */
.content { flex:1; overflow-y:auto; padding:24px; background:#0A0F1C; }
.content::-webkit-scrollbar { width:5px; }
.content::-webkit-scrollbar-thumb { background:var(--border); border-radius:3px; }
</style>
