<template>
  <div class="profile-page">
    <!-- ==================== 身份卡 ==================== -->
    <section class="identity">
      <div class="identity-wash" aria-hidden="true"></div>

      <div class="identity-body">
        <!-- 头像（点击更换） -->
        <div
          class="avatar"
          role="button"
          tabindex="0"
          title="点击更换头像"
          @click="triggerUpload"
          @keyup.enter="triggerUpload"
        >
          <div class="avatar-ring">
            <img v-if="avatarPreview" :src="avatarPreview" :alt="displayName" class="avatar-img" />
            <span v-else class="avatar-letter">{{ initial }}</span>
          </div>
          <span class="avatar-cam">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z" />
              <circle cx="12" cy="13" r="4" />
            </svg>
          </span>
          <input ref="fileInput" type="file" accept="image/*" hidden @change="handleFileSelect" />
        </div>

        <!-- 身份信息 -->
        <div class="identity-text">
          <div class="identity-name-row">
            <h1 class="identity-name">{{ displayName }}</h1>
            <span class="level-chip" :class="`level-${level}`">{{ levelText }}</span>
          </div>
          <p class="identity-meta">
            <span>@{{ userStore.userInfo?.username || '—' }}</span>
            <template v-if="joinedDate">
              <i class="meta-dot"></i>
              <span>注册于 {{ joinedDate }}</span>
            </template>
            <template v-if="daysSinceJoin">
              <i class="meta-dot"></i>
              <span>已陪伴 {{ daysSinceJoin }} 天</span>
            </template>
          </p>
        </div>

        <!-- 快捷操作 -->
        <div class="identity-actions">
          <button class="btn-ghost" @click="activeTab = 'info'">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.12 2.12 0 0 1 3 3L12 15l-4 1 1-4z"/></svg>
            编辑资料
          </button>
          <button class="btn-grad" :disabled="signInLoading" @click="handleSignIn">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2"/><path d="M16 2v4M8 2v4M3 10h18"/><path d="m9 16 2 2 4-4"/></svg>
            {{ signInLoading ? '签到中…' : '每日签到' }}
          </button>
        </div>
      </div>

      <!-- 等级成长进度 -->
      <div class="level-bar">
        <div class="level-head">
          <span class="level-title">成长进度</span>
          <span class="level-count">{{ levelProgressText }}</span>
        </div>
        <div class="level-track">
          <div class="level-fill" :style="{ width: levelPercent + '%' }"></div>
        </div>
        <p class="level-hint">{{ levelHint }}</p>
      </div>
    </section>

    <!-- ==================== 数据概览 ==================== -->
    <section class="stat-row">
      <button
        v-for="s in statCards"
        :key="s.key"
        class="stat-card"
        type="button"
        @click="activeTab = s.tab"
      >
        <span class="stat-icon" :class="`tint-${s.tint}`">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" v-html="s.icon"></svg>
        </span>
        <span class="stat-text">
          <span class="stat-label">{{ s.label }}</span>
          <span class="stat-value">{{ s.value }}</span>
        </span>
      </button>
    </section>

    <!-- ==================== 分段导航 ==================== -->
    <nav class="segmented" role="tablist">
      <span class="seg-indicator" :style="{ transform: `translateX(${activeIndex * 100}%)` }" aria-hidden="true"></span>
      <button
        v-for="tab in tabs"
        :key="tab.key"
        class="seg-item"
        type="button"
        role="tab"
        :aria-selected="activeTab === tab.key"
        :class="{ active: activeTab === tab.key }"
        @click="activeTab = tab.key"
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" v-html="tab.icon"></svg>
        {{ tab.label }}
      </button>
    </nav>

    <!-- ==================== 面板 ==================== -->
    <transition name="panel" mode="out-in">
      <!-- ---------- 资料设置 ---------- -->
      <section v-if="activeTab === 'info'" key="info" class="panel">
        <header class="panel-head">
          <div>
            <h2 class="panel-title">资料设置</h2>
            <p class="panel-desc">完善资料有助于提升账号安全与购物体验</p>
          </div>
        </header>

        <div class="form-grid">
          <div class="field">
            <label class="field-label">用户名</label>
            <div class="field-readonly">
              <span>{{ userStore.userInfo?.username || '-' }}</span>
              <span class="lock-chip">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
                不可修改
              </span>
            </div>
          </div>

          <div class="field">
            <label class="field-label">昵称</label>
            <input v-model="formData.nickname" class="field-input" maxlength="20" placeholder="输入昵称" />
          </div>

          <div class="field">
            <label class="field-label">性别</label>
            <div class="radio-group">
              <label
                v-for="g in genders"
                :key="g.value"
                class="radio-item"
                :class="{ active: formData.gender === g.value }"
              >
                <input type="radio" v-model="formData.gender" :value="g.value" hidden />
                {{ g.label }}
              </label>
            </div>
          </div>

          <div class="field">
            <label class="field-label">生日</label>
            <input type="date" v-model="formData.birthday" class="field-input" />
          </div>

          <div class="field">
            <label class="field-label">手机号</label>
            <input v-model="formData.phone" class="field-input" maxlength="20" placeholder="用于接收订单通知" />
          </div>

          <div class="field">
            <label class="field-label">邮箱</label>
            <input v-model="formData.email" class="field-input" maxlength="100" placeholder="用于接收账单与通知" />
          </div>

          <div class="field field-wide">
            <label class="field-label">头像链接</label>
            <input v-model="formData.avatar" class="field-input" placeholder="也可以直接粘贴一张图片 URL" />
          </div>
        </div>

        <footer class="panel-foot">
          <button class="btn-ghost" :disabled="saving" @click="handleResetForm">重置</button>
          <button class="btn-grad" :disabled="saving" @click="handleSave">
            {{ saving ? '保存中…' : '保存修改' }}
          </button>
        </footer>
      </section>

      <!-- ---------- 收货地址 ---------- -->
      <section v-else-if="activeTab === 'address'" key="address" class="panel">
        <header class="panel-head">
          <div>
            <h2 class="panel-title">收货地址</h2>
            <p class="panel-desc">已保存 {{ addresses.length }} 个地址，默认地址将在下单时自动选中</p>
          </div>
          <button class="btn-grad btn-sm" @click="openAddressDialog()">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M12 5v14M5 12h14"/></svg>
            新增地址
          </button>
        </header>

        <div v-if="addresses.length === 0" class="empty">
          <div class="empty-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>
          </div>
          <p class="empty-title">还没有收货地址</p>
          <p class="empty-desc">添加地址后，下单时可以直接选择</p>
          <button class="btn-grad" @click="openAddressDialog()">添加第一个地址</button>
        </div>

        <div v-else class="address-grid">
          <article
            v-for="addr in addresses"
            :key="addr.id"
            class="address-card"
            :class="{ 'is-default': addr.is_default === 1 }"
          >
            <header class="addr-head">
              <span class="addr-name">{{ addr.receiver_name }}</span>
              <span class="addr-phone">{{ addr.receiver_phone }}</span>
              <span v-if="addr.is_default === 1" class="tag-default">默认</span>
            </header>
            <p class="addr-detail">{{ addr.province }}{{ addr.city }}{{ addr.district }} {{ addr.detail }}</p>
            <p v-if="addr.postal_code" class="addr-postal">邮编 {{ addr.postal_code }}</p>
            <footer class="addr-actions">
              <button v-if="addr.is_default !== 1" class="btn-text" @click="handleSetDefault(addr)">设为默认</button>
              <button class="btn-text" @click="openAddressDialog(addr)">编辑</button>
              <button class="btn-text danger" @click="handleDeleteAddress(addr.id)">删除</button>
            </footer>
          </article>

          <!-- 新增地址占位块 -->
          <button class="address-add" type="button" @click="openAddressDialog()">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M12 5v14M5 12h14"/></svg>
            添加新地址
          </button>
        </div>
      </section>

      <!-- ---------- 账户信息 ---------- -->
      <section v-else key="account" class="panel">
        <header class="panel-head">
          <div>
            <h2 class="panel-title">账户信息</h2>
            <p class="panel-desc">积分、余额与账号绑定状态</p>
          </div>
        </header>

        <div class="account-grid">
          <div class="account-card">
            <span class="account-icon tint-gold">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 8h18v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><path d="M3 8 5.5 3h13L21 8"/><path d="M12 12v5"/></svg>
            </span>
            <div class="account-body">
              <span class="account-label">账户余额</span>
              <span class="account-value gold">¥{{ balance.toFixed(2) }}</span>
            </div>
            <button class="btn-ghost btn-sm" @click="handleRecharge">充值</button>
          </div>

          <div class="account-card">
            <span class="account-icon tint-cyan">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/></svg>
            </span>
            <div class="account-body">
              <span class="account-label">可用积分</span>
              <span class="account-value">{{ formatNumber(points) }}</span>
            </div>
            <button class="btn-ghost btn-sm" :disabled="signInLoading" @click="handleSignIn">
              {{ signInLoading ? '签到中…' : '签到' }}
            </button>
          </div>

          <div class="account-card">
            <span class="account-icon tint-violet">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2 15 9l7 .6-5.3 4.6L18.2 21 12 17.3 5.8 21l1.5-6.8L2 9.6 9 9z"/></svg>
            </span>
            <div class="account-body">
              <span class="account-label">会员等级</span>
              <span class="account-value">{{ levelText }}</span>
              <span class="account-hint">{{ levelHint }}</span>
            </div>
          </div>

          <div class="account-card">
            <span class="account-icon tint-mint">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2"/><path d="M16 2v4M8 2v4M3 10h18"/></svg>
            </span>
            <div class="account-body">
              <span class="account-label">注册时间</span>
              <span class="account-value sm">{{ createdAtText }}</span>
            </div>
          </div>
        </div>

        <h3 class="section-subtitle">账号绑定</h3>
        <div class="bind-list">
          <div class="bind-row">
            <span class="bind-label">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="2" width="14" height="20" rx="3"/><path d="M12 18h.01"/></svg>
              手机号
            </span>
            <span class="bind-value">{{ userStore.userInfo?.phone || '未填写' }}</span>
            <span class="bind-state" :class="userStore.userInfo?.phone ? 'on' : 'off'">
              {{ userStore.userInfo?.phone ? '已绑定' : '未绑定' }}
            </span>
          </div>
          <div class="bind-row">
            <span class="bind-label">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="4" width="20" height="16" rx="3"/><path d="m3 7 9 6 9-6"/></svg>
              邮箱
            </span>
            <span class="bind-value">{{ userStore.userInfo?.email || '未填写' }}</span>
            <span class="bind-state" :class="userStore.userInfo?.email ? 'on' : 'off'">
              {{ userStore.userInfo?.email ? '已绑定' : '未绑定' }}
            </span>
          </div>
          <div class="bind-row">
            <span class="bind-label">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/></svg>
              用户 ID
            </span>
            <span class="bind-value mono">{{ userStore.userInfo?.id || userStore.userId || '-' }}</span>
            <span class="bind-state neutral">系统生成</span>
          </div>
        </div>
      </section>
    </transition>

    <!-- ==================== 地址弹窗 ==================== -->
    <transition name="fade">
      <div v-if="addressDialogVisible" class="modal-overlay" @click.self="addressDialogVisible = false">
        <div class="modal-card">
          <header class="modal-head">
            <h3 class="modal-title">{{ editingAddress ? '编辑地址' : '新增地址' }}</h3>
            <button class="modal-close" type="button" aria-label="关闭" @click="addressDialogVisible = false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M18 6 6 18M6 6l12 12"/></svg>
            </button>
          </header>

          <div class="modal-body">
            <div class="field-row">
              <div class="field">
                <label class="field-label">收件人 <i>*</i></label>
                <input v-model="addressForm.receiver_name" class="field-input" placeholder="姓名" />
              </div>
              <div class="field">
                <label class="field-label">手机号 <i>*</i></label>
                <input v-model="addressForm.receiver_phone" class="field-input" maxlength="11" placeholder="11 位手机号" />
              </div>
            </div>
            <div class="field">
              <label class="field-label">地区 <i>*</i></label>
              <input v-model="addressForm.region" class="field-input" placeholder="省 市 区，用空格分隔" />
            </div>
            <div class="field">
              <label class="field-label">详细地址 <i>*</i></label>
              <input v-model="addressForm.detail" class="field-input" placeholder="街道、门牌号" />
            </div>
            <div class="field-row">
              <div class="field">
                <label class="field-label">邮编</label>
                <input v-model="addressForm.postal_code" class="field-input" placeholder="选填" />
              </div>
              <div class="field">
                <label class="field-label">默认地址</label>
                <label class="switch">
                  <input
                    type="checkbox"
                    :checked="addressForm.is_default === 1"
                    @change="addressForm.is_default = addressForm.is_default === 1 ? 0 : 1"
                  />
                  <span class="switch-slider"></span>
                  <span class="switch-text">{{ addressForm.is_default === 1 ? '下单时默认使用' : '不设为默认' }}</span>
                </label>
              </div>
            </div>
          </div>

          <footer class="modal-foot">
            <button class="btn-ghost" @click="addressDialogVisible = false">取消</button>
            <button class="btn-grad" :disabled="savingAddress" @click="handleSaveAddress">
              {{ savingAddress ? '保存中…' : '保存地址' }}
            </button>
          </footer>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { resolveAssetUrl as resolveUrl, uploadUrl } from '@/utils/api'
import { updateUserInfo, getAddressList, addAddress, updateAddress, deleteAddress, signIn, recharge } from '@/api/user'
import type { Address } from '@/api/user'

const userStore = useUserStore()

/**
 * 会员等级配置。
 *
 * 后端只返回 member_level 与 points，没有暴露成长值/升级规则，所以这里用积分为口径
 * 做展示层的映射。若后端后续给出真实规则，改这一处即可，页面其余部分无需调整。
 */
const LEVEL_CONFIG = [
  { level: 1, name: '普通会员', min: 0, next: 1000 },
  { level: 2, name: '银卡会员', min: 1000, next: 5000 },
  { level: 3, name: '金卡会员', min: 5000, next: 20000 },
  { level: 4, name: '钻石会员', min: 20000, next: null as number | null },
]

const tabs = [
  { key: 'info', label: '资料设置', icon: '<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>' },
  { key: 'address', label: '收货地址', icon: '<path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/>' },
  { key: 'account', label: '账户信息', icon: '<rect x="2" y="3" width="20" height="18" rx="3"/><line x1="6" y1="9" x2="18" y2="9"/><line x1="6" y1="13" x2="14" y2="13"/>' },
]
const genders = [{ value: 0, label: '未知' }, { value: 1, label: '男' }, { value: 2, label: '女' }]

const activeTab = ref('info')
const activeIndex = computed(() => Math.max(0, tabs.findIndex((t) => t.key === activeTab.value)))

// ====== 派生展示数据 ======
const displayName = computed(() => userStore.userInfo?.nickname || userStore.userInfo?.username || '未登录')
const initial = computed(() => (displayName.value || 'U')[0])
const points = computed(() => Number(userStore.userInfo?.points ?? 0))
const balance = computed(() => Number(userStore.userInfo?.balance ?? 0))
const level = computed(() => {
  const l = Number(userStore.userInfo?.member_level || 1)
  return LEVEL_CONFIG.some((c) => c.level === l) ? l : 1
})
const levelText = computed(() => LEVEL_CONFIG.find((c) => c.level === level.value)?.name || '普通会员')

const currentLevelConfig = computed(() => LEVEL_CONFIG.find((c) => c.level === level.value)!)
const levelPercent = computed(() => {
  const cfg = currentLevelConfig.value
  if (cfg.next === null) return 100
  const span = cfg.next - cfg.min
  if (span <= 0) return 100
  return Math.min(100, Math.max(0, Math.round(((points.value - cfg.min) / span) * 100)))
})
const levelProgressText = computed(() => {
  const cfg = currentLevelConfig.value
  return cfg.next === null ? '已达最高等级' : `${formatNumber(points.value)} / ${formatNumber(cfg.next)} 积分`
})
const levelHint = computed(() => {
  const cfg = currentLevelConfig.value
  if (cfg.next === null) return '已是最高等级，尊享全部会员权益'
  const nextName = LEVEL_CONFIG.find((c) => c.level === cfg.level + 1)?.name || '下一等级'
  const gap = Math.max(0, cfg.next - points.value)
  return `再获得 ${formatNumber(gap)} 积分即可升级为${nextName}`
})

const joinedDate = computed(() => (userStore.userInfo?.created_at || '').replace(' ', 'T').slice(0, 10))
// 账户卡里直接展示 ISO 串太长，截到分钟即可
const createdAtText = computed(() => {
  const raw = userStore.userInfo?.created_at
  return raw ? raw.replace('T', ' ').slice(0, 16) : '-'
})
const daysSinceJoin = computed(() => {
  const raw = userStore.userInfo?.created_at
  if (!raw) return 0
  const t = new Date(raw.replace(' ', 'T')).getTime()
  if (Number.isNaN(t)) return 0
  return Math.max(1, Math.ceil((Date.now() - t) / 86400000))
})

const statCards = computed(() => [
  { key: 'points', label: '可用积分', value: formatNumber(points.value), tab: 'account', tint: 'cyan',
    icon: '<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>' },
  { key: 'balance', label: '账户余额', value: `¥${balance.value.toFixed(2)}`, tab: 'account', tint: 'gold',
    icon: '<path d="M3 8h18v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><path d="M3 8 5.5 3h13L21 8"/><path d="M12 12v5"/>' },
  { key: 'address', label: '收货地址', value: `${addresses.value.length} 个`, tab: 'address', tint: 'violet',
    icon: '<path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/>' },
  { key: 'level', label: '会员等级', value: levelText.value, tab: 'account', tint: 'mint',
    icon: '<path d="M12 2 15 9l7 .6-5.3 4.6L18.2 21 12 17.3 5.8 21l1.5-6.8L2 9.6 9 9z"/>' },
])

const formatNumber = (n: number) => new Intl.NumberFormat('zh-CN').format(Number.isFinite(n) ? n : 0)

// ====== 签到 / 充值 ======
const signInLoading = ref(false)
const handleSignIn = async () => {
  const uid = userStore.userInfo?.id || userStore.userId
  if (!uid) { ElMessage.warning('请先登录'); return }
  if (signInLoading.value) return
  signInLoading.value = true
  try {
    const r = await signIn(Number(uid))
    ElMessage.success(`签到成功 +${r.addedPoints ?? 10} 积分`)
    await userStore.fetchUserInfo()
  } catch {
    // 已签到等错误由请求拦截器统一提示
  } finally {
    signInLoading.value = false
  }
}

const handleRecharge = async () => {
  const uid = userStore.userInfo?.id || userStore.userId
  if (!uid) { ElMessage.warning('请先登录'); return }
  try {
    const { value } = await ElMessageBox.prompt('输入充值金额（元）', '余额充值', {
      confirmButtonText: '充值', cancelButtonText: '取消',
      inputPattern: /^\d+(\.\d{1,2})?$/, inputErrorMessage: '请输入正确的金额',
    })
    const amount = parseFloat(value)
    if (!(amount > 0)) { ElMessage.warning('金额需大于 0'); return }
    await recharge(Number(uid), amount)
    ElMessage.success('充值成功')
    await userStore.fetchUserInfo()
  } catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message || '充值失败') }
}

// ====== 资料表单 ======
const saving = ref(false)
const fileInput = ref<HTMLInputElement>()
const avatarPreview = ref('')
const formData = reactive({ nickname: '', avatar: '', birthday: '', gender: 0, phone: '', email: '' })

const initForm = () => {
  if (!userStore.userInfo) return
  formData.nickname = userStore.userInfo.nickname || ''
  formData.avatar = userStore.userInfo.avatar || ''
  avatarPreview.value = resolveUrl(userStore.userInfo.avatar || '')
  formData.birthday = userStore.userInfo.birthday || ''
  formData.gender = userStore.userInfo.gender || 0
  formData.phone = userStore.userInfo.phone || ''
  formData.email = userStore.userInfo.email || ''
}

const handleResetForm = () => {
  initForm()
  ElMessage.success('已还原为当前保存的资料')
}

const triggerUpload = () => fileInput.value?.click()
const handleFileSelect = async (e: Event) => {
  const f = (e.target as HTMLInputElement).files?.[0]; if (!f) return
  if (f.size > 10 * 1024 * 1024) { ElMessage.warning('图片不能超过10MB'); return }
  try {
    const fd = new FormData(); fd.append('file', f)
    const res = await fetch(uploadUrl, { method: 'POST', headers: { Authorization: `Bearer ${userStore.token}` }, body: fd })
    const d = await res.json()
    const url = d.data?.file_url || d.data?.fileUrl
    if (d.code === 0 && url) { formData.avatar = url; avatarPreview.value = resolveUrl(url); ElMessage.success('头像上传成功') }
    else ElMessage.error(d.message || '上传失败')
  } catch { ElMessage.error('上传失败') }
  (e.target as HTMLInputElement).value = ''
}

const handleSave = async () => {
  saving.value = true
  try {
    await updateUserInfo({ nickname: formData.nickname, avatar: formData.avatar, gender: formData.gender, birthday: formData.birthday || undefined, phone: formData.phone || undefined, email: formData.email || undefined })
    ElMessage.success('保存成功'); await userStore.fetchUserInfo(); initForm()
  } catch (e: any) { ElMessage.error(e.message || '保存失败') }
  finally { saving.value = false }
}

// ====== 收货地址 ======
const addresses = ref<Address[]>([])
const addressDialogVisible = ref(false)
const savingAddress = ref(false)
const editingAddress = ref<Address | null>(null)
const addressForm = reactive({ receiver_name: '', receiver_phone: '', region: '', detail: '', postal_code: '', is_default: 0 })
const getUserId = () => userStore.userId || userStore.userInfo?.id || 0

const parseRegion = (p: string, c: string, d: string) => [p, c, d].filter(Boolean).join(' ')
const splitRegion = (r: string): { province: string; city: string; district: string } => {
  const parts = r.split(/\s+/).filter(Boolean)
  if (parts.length >= 3) return { province: parts[0], city: parts[1], district: parts.slice(2).join(' ') }
  return { province: parts[0] || '', city: parts[1] || '', district: parts.slice(2).join(' ') || '' }
}

const fetchAddresses = async () => {
  try { const r = await getAddressList(getUserId()); if (r.code === 0) addresses.value = r.data || [] } catch {}
}

const openAddressDialog = (addr?: Address) => {
  if (addr) {
    editingAddress.value = addr
    addressForm.receiver_name = addr.receiver_name; addressForm.receiver_phone = addr.receiver_phone
    addressForm.region = parseRegion(addr.province, addr.city, addr.district)
    addressForm.detail = addr.detail; addressForm.postal_code = addr.postal_code || ''
    addressForm.is_default = addr.is_default
  } else {
    editingAddress.value = null
    addressForm.receiver_name = addressForm.receiver_phone = addressForm.region = addressForm.detail = addressForm.postal_code = ''
    addressForm.is_default = 0
  }
  addressDialogVisible.value = true
}

const handleSaveAddress = async () => {
  if (!addressForm.receiver_name.trim()) { ElMessage.warning('请输入收件人姓名'); return }
  if (!addressForm.receiver_phone.trim() || !/^1[3-9]\d{9}$/.test(addressForm.receiver_phone)) { ElMessage.warning('请输入正确的手机号'); return }
  if (!addressForm.region.trim()) { ElMessage.warning('请输入省市区'); return }
  if (!addressForm.detail.trim()) { ElMessage.warning('请输入详细地址'); return }
  savingAddress.value = true
  try {
    const { province, city, district } = splitRegion(addressForm.region)
    if (!province || !city) { ElMessage.warning('地区格式错误，请用空格分隔，如：广东省 深圳市 南山区'); savingAddress.value = false; return }
    const base = { user_id: getUserId(), receiver_name: addressForm.receiver_name.trim(), receiver_phone: addressForm.receiver_phone.trim(), province, city, district, detail: addressForm.detail.trim(), postal_code: addressForm.postal_code, is_default: addressForm.is_default }
    if (editingAddress.value) { await updateAddress(editingAddress.value.id, { id: editingAddress.value.id, ...base }); ElMessage.success('已更新') }
    else { await addAddress(base); ElMessage.success('已添加') }
    addressDialogVisible.value = false; await fetchAddresses()
  } catch (e: any) { ElMessage.error(e.message || '操作失败') }
  finally { savingAddress.value = false }
}

const handleSetDefault = async (addr: Address) => {
  try {
    await updateAddress(addr.id, {
      id: addr.id, user_id: addr.user_id ?? getUserId(),
      receiver_name: addr.receiver_name, receiver_phone: addr.receiver_phone,
      province: addr.province, city: addr.city, district: addr.district,
      detail: addr.detail, postal_code: addr.postal_code || '', is_default: 1,
    })
    ElMessage.success('已设为默认地址')
    await fetchAddresses()
  } catch (e: any) { ElMessage.error(e.message || '设置失败') }
}

const handleDeleteAddress = async (id: number) => {
  try { await ElMessageBox.confirm('确定删除？', '提示', { type: 'warning' }); await deleteAddress(id, getUserId()); ElMessage.success('已删除'); await fetchAddresses() }
  catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message || '删除失败') }
}

onMounted(async () => {
  if (!userStore.userInfo) await userStore.fetchUserInfo()
  initForm(); fetchAddresses()
})
</script>

<style scoped>
.profile-page {
  max-width: 1100px;
  margin: 0 auto;
  padding: 24px 24px 80px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  font-family: var(--gf-font);
}

/* ==================== 身份卡 ==================== */
.identity {
  position: relative;
  overflow: hidden;
  border-radius: var(--radius-lg);
  background: var(--gf-glass-2);
  -webkit-backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke-strong);
  box-shadow: var(--gf-shadow-3), var(--gf-inner-shadow);
  padding: 28px 28px 22px;
}

/* 身份卡内部的一层虹彩底光，让卡片自身有方向感 */
.identity-wash {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    radial-gradient(120% 140% at 0% 0%, rgba(79, 216, 255, 0.16) 0%, transparent 55%),
    radial-gradient(90% 120% at 100% 10%, rgba(139, 124, 255, 0.16) 0%, transparent 60%);
}

.identity-body {
  position: relative;
  display: flex;
  align-items: center;
  gap: 22px;
}

/* ---- 头像 ---- */
.avatar {
  position: relative;
  flex-shrink: 0;
  width: 96px;
  height: 96px;
  cursor: pointer;
  border-radius: 50%;
}

.avatar-ring {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  padding: 2px;
  background: var(--gf-gradient);
  box-shadow: 0 12px 32px -14px var(--accent-glow);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.avatar-img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
  border: 3px solid var(--gf-bg);
}

.avatar-letter {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  border: 3px solid var(--gf-bg);
  background: var(--gf-bg-elev);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36px;
  font-weight: 700;
  color: var(--gf-accent);
}

.avatar-cam {
  position: absolute;
  right: -2px;
  bottom: -2px;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--gf-glass-deep);
  -webkit-backdrop-filter: blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke-strong);
  color: var(--text);
  box-shadow: var(--gf-shadow-2);
  opacity: 0;
  transform: scale(0.85);
  transition: opacity 0.25s ease, transform 0.25s cubic-bezier(0.22, 1, 0.36, 1);
}

.avatar-cam svg { width: 15px; height: 15px; }
.avatar:hover .avatar-cam,
.avatar:focus-visible .avatar-cam { opacity: 1; transform: scale(1); }
.avatar:hover .avatar-ring { box-shadow: 0 16px 40px -14px var(--accent-glow); }

/* ---- 身份文字 ---- */
.identity-text { flex: 1; min-width: 0; }

.identity-name-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.identity-name {
  margin: 0;
  font-size: 26px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--text);
}

.level-chip {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  border-radius: var(--radius-pill);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.02em;
  background: var(--accent-dim);
  color: var(--gf-accent);
  border: 1px solid rgba(79, 216, 255, 0.28);
  box-shadow: var(--gf-inner-shadow-soft);
}

.level-chip.level-3 {
  background: linear-gradient(120deg, rgba(232, 200, 138, 0.18), rgba(139, 124, 255, 0.18));
  color: var(--gf-gold);
  border-color: rgba(232, 200, 138, 0.35);
}

.level-chip.level-4 {
  background: var(--gf-gradient);
  color: #04121a;
  border-color: transparent;
}

.identity-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  margin: 8px 0 0;
  font-size: 13px;
  color: var(--text-dim);
}

.meta-dot {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: var(--gf-text-mute);
}

/* ---- 快捷操作 ---- */
.identity-actions {
  display: flex;
  gap: 10px;
  flex-shrink: 0;
}

.btn-grad,
.btn-ghost {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  padding: 11px 22px;
  border-radius: var(--radius-pill);
  font-size: 13px;
  font-weight: 600;
  font-family: var(--gf-font);
  cursor: pointer;
  white-space: nowrap;
  transition: transform 0.2s cubic-bezier(0.22, 1, 0.36, 1), box-shadow 0.25s ease, background 0.2s ease, filter 0.2s ease;
}

.btn-grad svg,
.btn-ghost svg { width: 15px; height: 15px; }

.btn-grad {
  border: none;
  background: var(--gf-gradient);
  color: #04121a;
  font-weight: 700;
  box-shadow: var(--gf-inner-shadow-soft), 0 10px 28px -12px var(--accent-glow);
}

.btn-grad:hover:not(:disabled) { transform: translateY(-1px); filter: brightness(1.06); box-shadow: var(--gf-inner-shadow-soft), 0 16px 36px -14px var(--accent-glow); }
.btn-grad:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-ghost {
  border: 1px solid var(--gf-stroke-strong);
  background: var(--gf-glass-1);
  color: var(--text);
  box-shadow: var(--gf-inner-shadow-soft);
  -webkit-backdrop-filter: blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));
}

.btn-ghost:hover:not(:disabled) { border-color: rgba(79, 216, 255, 0.45); color: var(--gf-accent); background: var(--gf-glass-2); transform: translateY(-1px); }
.btn-ghost:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-sm { padding: 9px 18px; font-size: 13px; }

/* ---- 成长进度 ---- */
.level-bar {
  position: relative;
  margin-top: 24px;
  padding-top: 18px;
  border-top: 1px solid var(--gf-stroke);
}

.level-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 10px;
}

.level-title { font-size: 12px; color: var(--text-dim); letter-spacing: 0.04em; }
.level-count { font-size: 12px; color: var(--text); font-weight: 600; font-variant-numeric: tabular-nums; }

.level-track {
  height: 6px;
  border-radius: 3px;
  background: rgba(255, 255, 255, 0.07);
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.45);
  overflow: hidden;
}

.level-fill {
  height: 100%;
  border-radius: 3px;
  background: var(--gf-gradient);
  box-shadow: 0 0 12px -2px var(--accent-glow);
  transition: width 0.8s cubic-bezier(0.22, 1, 0.36, 1);
}

.level-hint {
  margin: 10px 0 0;
  font-size: 12px;
  color: var(--gf-text-mute);
}

/* ==================== 数据概览 ==================== */
.stat-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 18px;
  text-align: left;
  cursor: pointer;
  font-family: var(--gf-font);
  border-radius: var(--radius);
  background: var(--gf-glass-1);
  -webkit-backdrop-filter: blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-sm)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke);
  box-shadow: var(--gf-inner-shadow-soft);
  transition: transform 0.3s cubic-bezier(0.22, 1, 0.36, 1), background 0.25s ease, border-color 0.25s ease, box-shadow 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-3px);
  background: var(--gf-glass-2);
  border-color: var(--gf-stroke-strong);
  box-shadow: var(--gf-shadow-2), var(--gf-inner-shadow);
}

.stat-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: var(--gf-inner-shadow-soft);
}

.stat-icon svg { width: 19px; height: 19px; }

.tint-cyan { background: rgba(79, 216, 255, 0.14); color: var(--gf-accent); }
.tint-gold { background: rgba(232, 200, 138, 0.15); color: var(--gf-gold); }
.tint-violet { background: rgba(139, 124, 255, 0.15); color: var(--gf-accent-2); }
.tint-mint { background: rgba(110, 231, 200, 0.14); color: var(--gf-accent-3); }

.stat-text { display: flex; flex-direction: column; gap: 3px; min-width: 0; }
.stat-label { font-size: 12px; color: var(--text-dim); }
.stat-value { font-size: 17px; font-weight: 700; color: var(--text); font-variant-numeric: tabular-nums; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

/* ==================== 分段导航 ==================== */
.segmented {
  position: relative;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0;
  padding: 4px;
  border-radius: var(--radius-pill);
  background: var(--gf-glass-1);
  -webkit-backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke);
  box-shadow: var(--gf-inner-shadow-soft);
}

.seg-indicator {
  position: absolute;
  top: 4px;
  bottom: 4px;
  left: 4px;
  width: calc((100% - 8px) / 3);
  border-radius: var(--radius-pill);
  background: var(--gf-gradient);
  box-shadow: var(--gf-inner-shadow-soft), 0 10px 26px -12px var(--accent-glow);
  transition: transform 0.35s cubic-bezier(0.22, 1, 0.36, 1);
}

.seg-item {
  position: relative;
  z-index: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 11px 10px;
  border: none;
  background: none;
  border-radius: var(--radius-pill);
  font-family: var(--gf-font);
  font-size: 14px;
  font-weight: 600;
  color: var(--text-dim);
  cursor: pointer;
  transition: color 0.25s ease;
}

.seg-item svg { width: 16px; height: 16px; }
.seg-item:hover { color: var(--text); }
.seg-item.active { color: #04121a; }

/* ==================== 面板 ==================== */
.panel {
  position: relative;
  border-radius: var(--radius-lg);
  background: var(--gf-glass-1);
  -webkit-backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke);
  box-shadow: var(--gf-inner-shadow);
  padding: 26px 28px;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 24px;
}

.panel-title {
  margin: 0;
  font-size: 17px;
  font-weight: 700;
  color: var(--text);
  letter-spacing: -0.01em;
}

.panel-desc { margin: 6px 0 0; font-size: 13px; color: var(--text-dim); }

.panel-foot {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 26px;
  padding-top: 20px;
  border-top: 1px solid var(--gf-stroke);
}

/* 面板切换动画 */
.panel-enter-active,
.panel-leave-active { transition: opacity 0.25s ease, transform 0.25s cubic-bezier(0.22, 1, 0.36, 1); }
.panel-enter-from { opacity: 0; transform: translateY(10px); }
.panel-leave-to { opacity: 0; transform: translateY(-8px); }

.fade-enter-active,
.fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from,
.fade-leave-to { opacity: 0; }

/* ==================== 表单 ==================== */
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px 20px;
}

.field { display: flex; flex-direction: column; gap: 8px; min-width: 0; }
.field-wide { grid-column: 1 / -1; }

.field-label {
  font-size: 13px;
  color: var(--text-dim);
  font-weight: 500;
}

.field-label i { color: var(--gf-danger); font-style: normal; }

.field-input {
  width: 100%;
  padding: 11px 14px;
  border-radius: var(--radius-sm);
  /* 输入框比卡片底色更沉，形成明确的"凹陷"关系，深色界面里最容易被忽略的一层 */
  background: rgba(6, 9, 17, 0.55);
  border: 1px solid var(--gf-stroke);
  box-shadow: var(--gf-inner-shadow-soft);
  color: var(--text);
  font-size: 14px;
  font-family: var(--gf-font);
  outline: none;
  box-sizing: border-box;
  transition: border-color 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
}

.field-input::placeholder { color: var(--gf-text-mute); }
.field-input:hover { border-color: var(--gf-stroke-strong); }
.field-input:focus { border-color: rgba(79, 216, 255, 0.5); background: rgba(6, 9, 17, 0.72); box-shadow: var(--gf-inner-shadow-soft), 0 0 0 3px var(--accent-dim); }
.field-input[type='date'] { color-scheme: dark; }
.field-input::-webkit-datetime-edit { color: var(--text); }
.field-input::-webkit-datetime-edit-fields-wrapper { color: var(--text); }
.field-input::-webkit-calendar-picker-indicator { opacity: 0.55; cursor: pointer; }
.field-input::-webkit-calendar-picker-indicator:hover { opacity: 0.9; }

.field-readonly {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 11px 14px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.02);
  border: 1px dashed var(--gf-stroke);
  color: var(--text-dim);
  font-size: 14px;
}

.lock-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  color: var(--gf-text-mute);
  white-space: nowrap;
}

.lock-chip svg { width: 12px; height: 12px; }

.radio-group {
  display: inline-flex;
  gap: 2px;
  padding: 3px;
  border-radius: var(--radius-pill);
  background: var(--gf-glass-1);
  border: 1px solid var(--gf-stroke);
  box-shadow: var(--gf-inner-shadow-soft);
  width: fit-content;
}

.radio-item {
  padding: 8px 20px;
  border-radius: var(--radius-pill);
  cursor: pointer;
  font-size: 13px;
  color: var(--text-dim);
  transition: color 0.2s ease, background 0.2s ease, box-shadow 0.2s ease;
}

.radio-item:hover { color: var(--text); }
.radio-item.active { background: var(--gf-gradient); color: #04121a; font-weight: 700; box-shadow: 0 6px 18px -8px var(--accent-glow); }

/* ==================== 空状态 ==================== */
.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 56px 20px;
}

.empty-icon {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 18px;
  background: var(--gf-glass-1);
  border: 1px solid var(--gf-stroke);
  box-shadow: var(--gf-inner-shadow-soft);
  color: var(--text-dim);
}

.empty-icon svg { width: 30px; height: 30px; }
.empty-title { margin: 0 0 6px; font-size: 16px; font-weight: 600; color: var(--text); }
.empty-desc { margin: 0 0 22px; font-size: 13px; color: var(--text-dim); }

/* ==================== 收货地址 ==================== */
.address-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.address-card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 18px;
  border-radius: var(--radius);
  background: var(--gf-glass-1);
  border: 1px solid var(--gf-stroke);
  box-shadow: var(--gf-inner-shadow-soft);
  transition: transform 0.3s cubic-bezier(0.22, 1, 0.36, 1), background 0.25s ease, border-color 0.25s ease, box-shadow 0.3s ease;
}

.address-card:hover {
  transform: translateY(-3px);
  background: var(--gf-glass-2);
  border-color: var(--gf-stroke-strong);
  box-shadow: var(--gf-shadow-2), var(--gf-inner-shadow);
}

/* 默认地址用左上角渐变细条标识，比整块高亮更克制 */
.address-card.is-default::before {
  content: '';
  position: absolute;
  left: 0;
  top: 18px;
  bottom: 18px;
  width: 3px;
  border-radius: 0 3px 3px 0;
  background: var(--gf-gradient);
  box-shadow: 0 0 12px var(--accent-glow);
}

.address-card.is-default { border-color: rgba(79, 216, 255, 0.28); }

.addr-head { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.addr-name { font-size: 15px; font-weight: 600; color: var(--text); }
.addr-phone { font-size: 13px; color: var(--text-dim); font-variant-numeric: tabular-nums; }

.tag-default {
  font-size: 11px;
  font-weight: 700;
  padding: 2px 9px;
  border-radius: var(--radius-xs);
  background: var(--gf-gradient);
  color: #04121a;
  box-shadow: var(--gf-inner-shadow-soft);
}

.addr-detail { margin: 0; font-size: 13px; line-height: 1.6; color: var(--text-dim); }
.addr-postal { margin: 0; font-size: 12px; color: var(--gf-text-mute); }

.addr-actions {
  display: flex;
  gap: 16px;
  /* 卡片在网格里等高拉伸，操作行统一贴底，避免两张卡片的按钮不在同一水平线 */
  margin-top: auto;
  padding-top: 12px;
  border-top: 1px solid var(--gf-stroke);
}

.btn-text {
  padding: 0;
  border: none;
  background: none;
  font-family: var(--gf-font);
  font-size: 13px;
  color: var(--text-dim);
  cursor: pointer;
  transition: color 0.2s ease;
}

.btn-text:hover { color: var(--gf-accent); }
.btn-text.danger:hover { color: var(--gf-danger); }

.address-add {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  min-height: 132px;
  padding: 18px;
  border-radius: var(--radius);
  border: 1px dashed var(--gf-stroke-strong);
  background: transparent;
  color: var(--text-dim);
  font-family: var(--gf-font);
  font-size: 13px;
  cursor: pointer;
  transition: border-color 0.25s ease, color 0.25s ease, background 0.25s ease;
}

.address-add svg { width: 22px; height: 22px; }
.address-add:hover { border-color: rgba(79, 216, 255, 0.5); color: var(--gf-accent); background: var(--gf-glass-1); }

/* ==================== 账户信息 ==================== */
.account-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.account-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 18px;
  border-radius: var(--radius);
  background: var(--gf-glass-1);
  border: 1px solid var(--gf-stroke);
  box-shadow: var(--gf-inner-shadow-soft);
}

.account-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: var(--gf-inner-shadow-soft);
}

.account-icon svg { width: 19px; height: 19px; }

.account-body { display: flex; flex-direction: column; gap: 4px; min-width: 0; flex: 1; }
.account-label { font-size: 12px; color: var(--text-dim); }
.account-value { font-size: 18px; font-weight: 700; color: var(--text); font-variant-numeric: tabular-nums; }
.account-value.gold { color: var(--gf-price); }
.account-value.sm { font-size: 13px; font-weight: 500; color: var(--text-dim); }
.account-hint { font-size: 11px; color: var(--gf-text-mute); line-height: 1.5; }

.section-subtitle {
  margin: 28px 0 14px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text);
}

.bind-list { display: flex; flex-direction: column; }

.bind-row {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 2px;
  border-bottom: 1px solid var(--gf-stroke);
}

.bind-row:last-child { border-bottom: none; }

.bind-label {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  width: 120px;
  flex-shrink: 0;
  font-size: 13px;
  color: var(--text-dim);
}

.bind-label svg { width: 15px; height: 15px; }
.bind-value { flex: 1; min-width: 0; font-size: 14px; color: var(--text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.bind-value.mono { font-family: var(--gf-font-num); font-size: 13px; color: var(--text-dim); }

.bind-state {
  flex-shrink: 0;
  font-size: 12px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: var(--radius-pill);
}

.bind-state.on { color: var(--gf-success); background: rgba(74, 222, 155, 0.12); }
.bind-state.off { color: var(--gf-warning); background: rgba(251, 191, 107, 0.12); }
.bind-state.neutral { color: var(--text-dim); background: rgba(255, 255, 255, 0.06); }

/* ==================== 弹窗 ==================== */
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 200;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: rgba(4, 6, 12, 0.62);
  -webkit-backdrop-filter: blur(6px);
  backdrop-filter: blur(6px);
}

.modal-card {
  width: 520px;
  max-width: 100%;
  max-height: 88vh;
  overflow-y: auto;
  padding: 24px 26px;
  border-radius: var(--radius-lg);
  background: var(--gf-glass-deep);
  -webkit-backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  backdrop-filter: blur(var(--gf-blur-lg)) saturate(var(--gf-saturate));
  border: 1px solid var(--gf-stroke-strong);
  box-shadow: var(--gf-shadow-3), var(--gf-inner-shadow);
}

.modal-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 22px;
}

.modal-title { margin: 0; font-size: 18px; font-weight: 700; color: var(--text); }

.modal-close {
  width: 34px;
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  border: 1px solid var(--gf-stroke);
  background: var(--gf-glass-1);
  color: var(--text-dim);
  cursor: pointer;
  transition: all 0.2s ease;
}

.modal-close svg { width: 15px; height: 15px; }
.modal-close:hover { color: var(--text); border-color: var(--gf-stroke-strong); background: var(--gf-glass-2); }

.modal-body { display: flex; flex-direction: column; gap: 16px; }
.field-row { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 16px; }

.modal-foot {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid var(--gf-stroke);
}

/* 开关 */
.switch { display: inline-flex; align-items: center; gap: 10px; cursor: pointer; user-select: none; }
.switch input { opacity: 0; width: 0; height: 0; position: absolute; }

.switch-slider {
  position: relative;
  width: 44px;
  height: 26px;
  flex-shrink: 0;
  border-radius: 26px;
  background: rgba(255, 255, 255, 0.09);
  border: 1px solid var(--gf-stroke);
  transition: background 0.25s ease, border-color 0.25s ease, box-shadow 0.25s ease;
}

.switch-slider::before {
  content: '';
  position: absolute;
  left: 3px;
  bottom: 3px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--text);
  transition: transform 0.25s cubic-bezier(0.22, 1, 0.36, 1);
}

.switch input:checked + .switch-slider {
  background: var(--gf-gradient);
  border-color: transparent;
  box-shadow: 0 6px 18px -8px var(--accent-glow);
}

.switch input:checked + .switch-slider::before { transform: translateX(18px); background: #04121a; }

.switch-text { font-size: 13px; color: var(--text-dim); }

/* ==================== 响应式 ==================== */
@media (max-width: 900px) {
  .profile-page { padding: 20px 16px 64px; }
  .identity { padding: 22px 20px 18px; }
  .identity-body { flex-wrap: wrap; gap: 18px; }
  .identity-actions { width: 100%; }
  .identity-actions .btn-grad,
  .identity-actions .btn-ghost { flex: 1; }
  .stat-row { grid-template-columns: repeat(2, 1fr); }
  .form-grid,
  .account-grid { grid-template-columns: 1fr; }
  .address-grid { grid-template-columns: 1fr; }
  .seg-item { font-size: 13px; padding: 10px 6px; }
  .seg-item svg { display: none; }
}

@media (max-width: 520px) {
  .identity-name { font-size: 22px; }
  .panel { padding: 20px 18px; }
  .panel-head { flex-direction: column; align-items: stretch; }
  .field-row { grid-template-columns: 1fr; }
  .bind-label { width: 92px; }
  .modal-card { padding: 20px 18px; }
}
</style>
