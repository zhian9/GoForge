<template>
  <div class="msg-page">
    <div class="page-head">
      <h1>消息中心</h1>
      <p class="sub">订单、物流、优惠券到账等通知都会在这里</p>
    </div>

    <div class="toolbar">
      <div class="type-tabs">
        <button
          v-for="t in typeTabs"
          :key="t.value"
          class="type-tab"
          :class="{ active: type === t.value }"
          @click="changeType(t.value)"
        >{{ t.label }}</button>
      </div>
      <el-button v-if="messages.length > 0" text @click="handleReadAll">全部标记已读</el-button>
    </div>

    <el-skeleton v-if="loading" :rows="5" animated />
    <el-empty v-else-if="messages.length === 0" description="暂无消息" />
    <div v-else class="msg-list">
      <div
        v-for="m in messages"
        :key="m.id"
        class="msg-card"
        :class="{ unread: m.is_read === 0 }"
        @click="handleOpen(m)"
      >
        <div class="msg-top">
          <span class="msg-type">{{ messageTypeText(m.type) }}</span>
          <span class="msg-time">{{ formatTime(m.created_at) }}</span>
        </div>
        <h3 class="msg-title">
          <span v-if="m.is_read === 0" class="dot" />
          {{ m.title }}
        </h3>
        <p class="msg-content">{{ m.content }}</p>
      </div>
    </div>

    <div v-if="hasMore && !loading" class="more">
      <el-button text @click="loadMore">加载更多</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getMessages, markAsRead, batchMarkAsRead, messageTypeText, type Message } from '@/api/message'

const router = useRouter()

const loading = ref(false)
const messages = ref<Message[]>([])
const type = ref(0)
const page = ref(1)
const pageSize = 10
const total = ref(0)

const typeTabs = [
  { label: '全部', value: 0 },
  { label: '系统通知', value: 1 },
  { label: '订单消息', value: 2 },
  { label: '营销消息', value: 3 },
  { label: '物流消息', value: 4 },
]

const hasMore = ref(false)

const formatTime = (value?: string) =>
  value ? new Date(value).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }) : ''

const fetchList = async (reset = false) => {
  loading.value = true
  try {
    if (reset) page.value = 1
    const res: any = await getMessages({ type: type.value, page: page.value, page_size: pageSize })
    const list: Message[] = res?.data || []
    total.value = Number(res?.total ?? list.length)
    messages.value = reset ? list : [...messages.value, ...list]
    hasMore.value = messages.value.length < total.value
  } catch {
    // 拦截器已提示
  } finally {
    loading.value = false
  }
}

const changeType = (value: number) => {
  type.value = value
  fetchList(true)
}

const loadMore = () => {
  page.value += 1
  fetchList()
}

/** 点击消息：标记已读，并按 link 跳转（link 形如 /orders/123） */
const handleOpen = async (m: Message) => {
  if (m.is_read === 0) {
    try {
      await markAsRead(Number(m.id))
      m.is_read = 1
    } catch {
      // 标记失败不阻塞跳转
    }
  }
  if (m.link) router.push(m.link)
}

const handleReadAll = async () => {
  const unread = messages.value.filter((m) => m.is_read === 0).map((m) => Number(m.id))
  if (unread.length === 0) {
    ElMessage.info('没有未读消息')
    return
  }
  try {
    await batchMarkAsRead(unread)
    messages.value.forEach((m) => { m.is_read = 1 })
    ElMessage.success('已全部标记为已读')
  } catch {
    ElMessage.warning('操作失败，请稍后再试')
  }
}

onMounted(() => fetchList(true))
</script>

<style scoped>
.msg-page { max-width: 900px; margin: 0 auto; padding: 28px 20px 60px; }
.page-head h1 { margin: 0 0 6px; font-size: 26px; color: var(--text); }
.page-head .sub { margin: 0 0 18px; color: var(--text-dim); font-size: 13px; }
.toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 14px; gap: 12px; flex-wrap: wrap; }
.type-tabs { display: flex; gap: 6px; flex-wrap: wrap; }
.type-tab {
  padding: 6px 14px; border-radius: 999px; cursor: pointer; font-size: 13px;
  border: 1px solid var(--gf-border, rgba(255,255,255,.12));
  background: transparent; color: var(--text-dim);
}
.type-tab.active { color: #0b1220; background: var(--gf-accent, #4FD8FF); border-color: transparent; font-weight: 600; }
.msg-list { display: flex; flex-direction: column; gap: 12px; }
.msg-card {
  padding: 16px 18px; border-radius: 14px; cursor: pointer;
  border: 1px solid var(--gf-border, rgba(255,255,255,.1));
  background: var(--gf-glass, rgba(16,22,34,.72));
  transition: border-color .2s, transform .2s;
}
.msg-card:hover { border-color: var(--gf-accent, #4FD8FF); transform: translateY(-1px); }
.msg-card.unread { border-left: 3px solid var(--gf-accent, #4FD8FF); }
.msg-top { display: flex; justify-content: space-between; font-size: 12px; color: var(--text-dim); margin-bottom: 6px; }
.msg-type { color: var(--gf-accent, #4FD8FF); }
.msg-title { margin: 0 0 6px; font-size: 15px; color: var(--text); display: flex; align-items: center; gap: 6px; }
.dot { width: 6px; height: 6px; border-radius: 50%; background: #ff6b6b; display: inline-block; }
.msg-content { margin: 0; font-size: 13px; color: var(--text-dim); line-height: 1.6; white-space: pre-wrap; }
.more { text-align: center; margin-top: 16px; }
</style>
