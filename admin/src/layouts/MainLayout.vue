<template>
  <div class="admin-layout">
    <!-- 低强度极光：给玻璃材质提供可折射的内容 -->
    <div class="gf-aurora" aria-hidden="true">
      <span class="gf-aurora__blob gf-aurora__blob--1"></span>
      <span class="gf-aurora__blob gf-aurora__blob--2"></span>
    </div>

    <!-- ======== 侧边栏 ======== -->
    <aside class="sidebar" :class="{ collapsed }">
      <div class="sidebar-logo" @click="$router.push('/dashboard')">
        <img src="/logo-forge.png" alt="GoForge" class="logo-img" />
        <div v-show="!collapsed" class="logo-text">
          <strong>GoForge</strong>
          <span>管理后台</span>
        </div>
      </div>

      <nav class="sidebar-nav">
        <template v-for="group in menuGroups" :key="group.label">
          <div v-if="group.label" v-show="!collapsed" class="nav-group-label">{{ group.label }}</div>
          <div v-else v-show="!collapsed" class="nav-group-label nav-group-label--first">概览</div>
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

      <button class="collapse-btn" type="button" :title="collapsed ? '展开菜单' : '收起菜单'" @click="collapsed = !collapsed">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" :class="{ rotated: collapsed }">
          <polyline points="15 18 9 12 15 6" />
        </svg>
      </button>
    </aside>

    <!-- ======== 主区域 ======== -->
    <div class="main-area">
      <header class="topbar">
        <div class="topbar-left">
          <button class="icon-btn" type="button" title="切换菜单" @click="collapsed = !collapsed">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <line x1="3" y1="6" x2="21" y2="6" /><line x1="3" y1="12" x2="21" y2="12" /><line x1="3" y1="18" x2="21" y2="18" />
            </svg>
          </button>

          <nav class="breadcrumb" aria-label="面包屑">
            <span class="bc-root">GoForge</span>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="bc-sep"><polyline points="9 18 15 12 9 6" /></svg>
            <span v-if="currentGroup" class="bc-group">{{ currentGroup }}</span>
            <svg v-if="currentGroup" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="bc-sep"><polyline points="9 18 15 12 9 6" /></svg>
            <span class="bc-current">{{ currentTitle }}</span>
          </nav>
        </div>

        <div class="topbar-right">
          <div class="search-box">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="search-icon"><circle cx="11" cy="11" r="8" /><path d="m21 21-4.35-4.35" /></svg>
            <input v-model="keyword" type="text" placeholder="搜索菜单…" class="search-input" @keyup.enter="handleSearch" />
          </div>

          <button class="icon-btn" type="button" title="通知">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
              <path d="M13.73 21a2 2 0 0 1-3.46 0" />
            </svg>
            <span class="notif-dot"></span>
          </button>

          <div class="user-dropdown" @click.stop="showMenu = !showMenu">
            <div class="user-avatar">{{ initial }}</div>
            <div class="user-info">
              <span class="user-name">{{ userStore.userInfo?.username || '管理员' }}</span>
              <span class="user-role">超级管理员</span>
            </div>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="arrow"><polyline points="6 9 12 15 18 9" /></svg>

            <div v-if="showMenu" class="dropdown-menu" @click.stop>
              <div class="dm-header">
                <span class="dm-avatar">{{ initial }}</span>
                <div class="dm-text">
                  <div class="dm-name">{{ userStore.userInfo?.username || '管理员' }}</div>
                  <div class="dm-email">{{ userStore.userInfo?.email || 'admin@goforge.dev' }}</div>
                </div>
              </div>
              <div class="dm-divider"></div>
              <div class="dropdown-item" @click="handleLogout">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" /><polyline points="16 17 21 12 16 7" /><line x1="21" y1="12" x2="9" y2="12" />
                </svg>
                退出登录
              </div>
            </div>
          </div>
        </div>
      </header>

      <main class="content">
        <router-view v-slot="{ Component }">
          <transition name="gf-page" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage, ElMessageBox } from 'element-plus'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const collapsed = ref(false)
const showMenu = ref(false)
const keyword = ref('')
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

const currentItem = computed(() => {
  for (const g of menuGroups) {
    const m = g.items.find((i) => isActive(i.path))
    if (m) return m
  }
  return null
})
const currentTitle = computed(() => currentItem.value?.label || '管理后台')
/** 面包屑带上分组名，和主流后台一致：GoForge / 基础数据 / 商品管理 */
const currentGroup = computed(() => {
  for (const g of menuGroups) {
    if (g.items.some((i) => isActive(i.path))) return g.label
  }
  return ''
})

/** 顶栏搜索在菜单里做模糊匹配，回车直接跳转 */
const handleSearch = () => {
  const kw = keyword.value.trim()
  if (!kw) return
  for (const g of menuGroups) {
    const hit = g.items.find((i) => i.label.includes(kw))
    if (hit) { router.push(hit.path); keyword.value = ''; return }
  }
  ElMessage.warning(`没有找到「${kw}」相关的菜单`)
}

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定退出登录？', '提示', { type: 'warning' })
    userStore.logout()
    router.push('/login')
  } catch { /* 取消 */ }
}

if (typeof document !== 'undefined') {
  document.addEventListener('click', () => { showMenu.value = false })
}
</script>

<style scoped>
.admin-layout {
  position: relative;
  display: flex;
  height: 100vh;
  overflow: hidden;
  color: var(--gf-text);
  font-family: var(--gf-font);
}

/* ==================== 侧边栏 ==================== */
.sidebar {
  position: relative;
  z-index: var(--gf-z-sidebar);
  width: 226px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, rgba(9, 13, 22, 0.9) 0%, rgba(5, 8, 14, 0.94) 100%);
  -webkit-backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  border-right: 1px solid var(--gf-stroke);
  transition: width 0.26s cubic-bezier(0.22, 1, 0.36, 1);
}

.sidebar.collapsed { width: 68px; }
.sidebar.collapsed .nav-label,
.sidebar.collapsed .logo-text { display: none; }
.sidebar.collapsed .nav-item { justify-content: center; padding: 11px 0; }

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 11px;
  padding: 18px 18px 16px;
  cursor: pointer;
  border-bottom: 1px solid var(--gf-stroke);
}

.logo-img { height: 27px; width: auto; flex-shrink: 0; }
.logo-text { display: flex; flex-direction: column; gap: 1px; min-width: 0; }
.logo-text strong { font-size: 14px; font-weight: 700; color: var(--gf-text); letter-spacing: -0.01em; }
.logo-text span { font-size: 10px; color: var(--gf-text-mute); letter-spacing: 0.12em; }

.sidebar-nav { flex: 1; overflow-y: auto; padding: 10px 12px 64px; }

.nav-group-label {
  padding: 16px 10px 7px;
  font-size: 10px;
  font-weight: 600;
  color: var(--gf-text-mute);
  letter-spacing: 0.1em;
}

.nav-group-label--first { padding-top: 6px; }

.nav-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 11px;
  margin-bottom: 2px;
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  color: var(--gf-text-dim);
  font-size: 13px;
  font-weight: 500;
  text-decoration: none;
  white-space: nowrap;
  transition: color 0.2s ease, background 0.2s ease, box-shadow 0.2s ease;
}

.nav-item:hover { color: var(--gf-text); background: var(--gf-glass-1); }

.nav-item.active {
  color: var(--gf-accent);
  background: var(--accent-dim);
  font-weight: 600;
  box-shadow: inset 0 0 0 1px rgba(79, 216, 255, 0.16);
}

/* 选中态左侧的渐变指示条 */
.nav-item.active::before {
  content: '';
  position: absolute;
  left: -12px;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 20px;
  border-radius: 0 3px 3px 0;
  background: var(--gf-gradient);
  box-shadow: 0 0 12px var(--accent-glow);
}

.nav-icon { width: 19px; height: 19px; flex-shrink: 0; display: flex; align-items: center; justify-content: center; }
.nav-icon :deep(svg) { width: 19px; height: 19px; fill: none; stroke: currentColor; stroke-width: 1.9; stroke-linecap: round; stroke-linejoin: round; }

.collapse-btn {
  position: absolute;
  bottom: 14px;
  right: 14px;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  border: 1px solid var(--gf-stroke);
  background: var(--gf-glass-1);
  box-shadow: var(--gf-inner-shadow-soft);
  color: var(--gf-text-dim);
  cursor: pointer;
  transition: color 0.2s ease, border-color 0.2s ease, background 0.2s ease;
}

.collapse-btn:hover { color: var(--gf-accent); border-color: rgba(79, 216, 255, 0.4); background: var(--gf-glass-2); }
.collapse-btn svg { width: 15px; height: 15px; transition: transform 0.26s cubic-bezier(0.22, 1, 0.36, 1); }
.collapse-btn svg.rotated { transform: rotate(180deg); }

/* ==================== 主区域 ==================== */
.main-area { flex: 1; display: flex; flex-direction: column; min-width: 0; overflow: hidden; }

/* ==================== 顶栏 ==================== */
.topbar {
  position: relative;
  z-index: var(--gf-z-topbar);
  height: 60px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 0 20px;
  background: var(--gf-topbar);
  -webkit-backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  border-bottom: 1px solid var(--gf-stroke);
}

.topbar-left { display: flex; align-items: center; gap: 14px; min-width: 0; }
.topbar-right { display: flex; align-items: center; gap: 10px; }

.icon-btn {
  position: relative;
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  border: 1px solid var(--gf-stroke);
  background: var(--gf-glass-1);
  box-shadow: var(--gf-inner-shadow-soft);
  color: var(--gf-text-dim);
  cursor: pointer;
  transition: color 0.2s ease, border-color 0.2s ease, background 0.2s ease;
}

.icon-btn:hover { color: var(--gf-accent); border-color: rgba(79, 216, 255, 0.4); background: var(--gf-glass-2); }
.icon-btn svg { width: 17px; height: 17px; }

.notif-dot {
  position: absolute;
  top: 8px;
  right: 9px;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--gf-danger);
  box-shadow: 0 0 8px var(--gf-danger);
}

/* 面包屑 */
.breadcrumb { display: flex; align-items: center; gap: 7px; font-size: 13px; min-width: 0; }
.bc-root { color: var(--gf-text-mute); font-weight: 500; }
.bc-sep { width: 13px; height: 13px; color: var(--gf-text-mute); flex-shrink: 0; }
.bc-group { color: var(--gf-text-dim); white-space: nowrap; }
.bc-current { color: var(--gf-text); font-weight: 600; white-space: nowrap; }

/* 搜索 */
.search-box { position: relative; display: flex; align-items: center; }
.search-icon { position: absolute; left: 11px; width: 15px; height: 15px; color: var(--gf-text-mute); pointer-events: none; }

.search-input {
  width: 190px;
  padding: 9px 14px 9px 33px;
  border-radius: var(--radius-sm);
  background: var(--gf-glass-1);
  border: 1px solid var(--gf-stroke);
  box-shadow: var(--gf-inner-shadow-soft);
  color: var(--gf-text);
  font-size: 13px;
  font-family: var(--gf-font);
  outline: none;
  transition: width 0.3s cubic-bezier(0.22, 1, 0.36, 1), border-color 0.2s ease, background 0.2s ease, box-shadow 0.2s ease;
}

.search-input::placeholder { color: var(--gf-text-mute); }
.search-input:focus {
  width: 240px;
  border-color: rgba(79, 216, 255, 0.5);
  background: var(--gf-glass-2);
  box-shadow: var(--gf-inner-shadow-soft), 0 0 0 3px var(--accent-dim);
}

/* 用户菜单 */
.user-dropdown {
  position: relative;
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 4px 10px 4px 4px;
  border-radius: var(--radius-pill);
  border: 1px solid transparent;
  cursor: pointer;
  transition: background 0.2s ease, border-color 0.2s ease;
}

.user-dropdown:hover { background: var(--gf-glass-1); border-color: var(--gf-stroke); }

.user-avatar {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: var(--gf-gradient);
  color: #04121a;
  font-size: 14px;
  font-weight: 700;
  box-shadow: var(--gf-inner-shadow-soft), 0 6px 16px -8px var(--accent-glow);
}

.user-info { display: flex; flex-direction: column; gap: 0; }
.user-name { font-size: 13px; font-weight: 600; color: var(--gf-text); line-height: 1.35; }
.user-role { font-size: 10px; color: var(--gf-text-mute); line-height: 1.35; }
.arrow { width: 14px; height: 14px; color: var(--gf-text-mute); }

.dropdown-menu {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  min-width: 216px;
  padding: 6px;
  border-radius: var(--radius);
  background: var(--gf-glass-deep);
  -webkit-backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke-strong);
  box-shadow: var(--gf-shadow-3), var(--gf-inner-shadow);
  z-index: 50;
}

.dm-header { display: flex; align-items: center; gap: 10px; padding: 10px 12px; }
.dm-avatar {
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: var(--gf-gradient);
  color: #04121a;
  font-size: 15px;
  font-weight: 700;
}
.dm-text { min-width: 0; }
.dm-name { font-size: 13px; font-weight: 600; color: var(--gf-text); }
.dm-email { font-size: 11px; color: var(--gf-text-mute); overflow: hidden; text-overflow: ellipsis; }
.dm-divider { height: 1px; margin: 6px 0; background: var(--gf-stroke); }

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 10px 12px;
  border-radius: var(--radius-xs);
  font-size: 13px;
  color: var(--gf-danger);
  cursor: pointer;
  transition: background 0.2s ease;
}

.dropdown-item svg { width: 16px; height: 16px; }
.dropdown-item:hover { background: rgba(255, 107, 129, 0.12); }

/* ==================== 内容区 ==================== */
.content {
  position: relative;
  z-index: var(--gf-z-content);
  flex: 1;
  overflow-y: auto;
  padding: 22px 24px 32px;
}

@media (max-width: 900px) {
  .sidebar { position: absolute; height: 100%; transform: translateX(-100%); transition: transform 0.26s cubic-bezier(0.22, 1, 0.36, 1); }
  .sidebar.collapsed { width: 226px; transform: translateX(0); }
  .sidebar.collapsed .nav-label,
  .sidebar.collapsed .logo-text { display: flex; }
  .sidebar.collapsed .nav-item { justify-content: flex-start; padding: 10px 12px; }
  .content { padding: 16px 14px 28px; }
  .search-input { width: 140px; }
  .search-input:focus { width: 170px; }
}
</style>
