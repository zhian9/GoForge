<template>
  <div class="profile-page">
    <h1 class="page-title">个人中心</h1>

    <div class="layout">
      <!-- ======== 左侧导航 ======== -->
      <aside class="side-nav">
        <div
          v-for="tab in tabs"
          :key="tab.key"
          class="nav-item"
          :class="{ active: activeTab === tab.key }"
          @click="activeTab = tab.key"
        >
          <svg v-html="tab.icon" class="nav-icon"></svg>
          {{ tab.label }}
        </div>
      </aside>

      <!-- ======== 右侧内容 ======== -->
      <main class="main-content">
        <!-- 基本信息 -->
        <div v-show="activeTab === 'info'" class="card">
          <h3 class="card-title">基本信息</h3>
          <div class="info-layout">
            <!-- 头像 -->
            <div class="avatar-col">
              <div class="avatar-upload" @click="triggerUpload" title="点击更换头像">
                <div class="avatar-preview">
                  <img v-if="avatarPreview" :src="avatarPreview" />
                  <span v-else class="avatar-letter">{{ initial }}</span>
                </div>
                <div class="avatar-overlay">更换</div>
              </div>
              <input ref="fileInput" type="file" accept="image/*" hidden @change="handleFileSelect" />
              <p class="avatar-hint">点击更换头像</p>
            </div>

            <!-- 表单 -->
            <div class="form-col">
              <div class="form-item">
                <label class="form-label">用户名</label>
                <div class="form-value disabled">{{ userStore.userInfo?.username || '-' }}</div>
              </div>
              <div class="form-row">
                <div class="form-item">
                  <label class="form-label">昵称</label>
                  <input v-model="formData.nickname" class="form-input" maxlength="20" placeholder="输入昵称" />
                </div>
                <div class="form-item">
                  <label class="form-label">性别</label>
                  <div class="radio-group">
                    <label v-for="g in genders" :key="g.value" class="radio-item" :class="{ active: formData.gender === g.value }">
                      <input type="radio" v-model="formData.gender" :value="g.value" hidden />
                      {{ g.label }}
                    </label>
                  </div>
                </div>
              </div>
              <div class="form-row">
                <div class="form-item">
                  <label class="form-label">生日</label>
                  <input type="date" v-model="formData.birthday" class="form-input" />
                </div>
                <div class="form-item">
                  <label class="form-label">手机号</label>
                  <input v-model="formData.phone" class="form-input" maxlength="20" placeholder="手机号" />
                </div>
              </div>
              <div class="form-item">
                <label class="form-label">邮箱</label>
                <input v-model="formData.email" class="form-input" maxlength="100" placeholder="输入邮箱" />
              </div>
              <div class="form-item">
                <label class="form-label">头像 URL</label>
                <input v-model="formData.avatar" class="form-input" placeholder="或输入头像图片 URL" />
              </div>
              <button class="btn-save" :disabled="saving" @click="handleSave">
                {{ saving ? '保存中...' : '保存修改' }}
              </button>
            </div>
          </div>
        </div>

        <!-- 收货地址 -->
        <div v-show="activeTab === 'address'" class="card">
          <div class="card-head">
            <h3 class="card-title">收货地址</h3>
            <button class="btn-add" @click="openAddressDialog()">+ 添加新地址</button>
          </div>
          <div v-if="addresses.length === 0" class="empty-block">暂无收货地址</div>
          <div v-else class="address-list">
            <div v-for="addr in addresses" :key="addr.id" class="address-card" :class="{ 'is-default': addr.is_default===1 }">
              <div class="addr-body">
                <div class="addr-line1">
                  <strong>{{ addr.receiver_name }}</strong>
                  <span class="phone">{{ addr.receiver_phone }}</span>
                  <span v-if="addr.is_default===1" class="default-tag">默认</span>
                </div>
                <div class="addr-line2">{{ addr.province }}{{ addr.city }}{{ addr.district }} {{ addr.detail }}</div>
              </div>
              <div class="addr-actions">
                <button class="btn-link" @click="openAddressDialog(addr)">编辑</button>
                <button class="btn-link danger" @click="handleDeleteAddress(addr.id)">删除</button>
              </div>
            </div>
          </div>
        </div>

        <!-- 账户信息 -->
        <div v-show="activeTab === 'account'" class="card">
          <h3 class="card-title">账户信息</h3>
          <div class="account-grid">
            <div class="account-item"><span class="a-label">会员等级</span><span class="a-value">{{ levelText }}</span></div>
            <div class="account-item"><span class="a-label">积分</span><span class="a-value accent">{{ userStore.userInfo?.points || 0 }}</span></div>
            <div class="account-item"><span class="a-label">注册时间</span><span class="a-value">{{ userStore.userInfo?.created_at || '-' }}</span></div>
            <div class="account-item"><span class="a-label">最后更新</span><span class="a-value">{{ userStore.userInfo?.updated_at || '-' }}</span></div>
          </div>
        </div>
      </main>
    </div>

    <!-- ======== 地址弹窗 ======== -->
    <div v-if="addressDialogVisible" class="modal-overlay" @click.self="addressDialogVisible=false">
      <div class="modal-card">
        <h3 class="modal-title">{{ editingAddress ? '编辑地址' : '添加新地址' }}</h3>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-item"><label class="form-label">收件人 *</label><input v-model="addressForm.receiver_name" class="form-input" placeholder="姓名" /></div>
            <div class="form-item"><label class="form-label">手机号 *</label><input v-model="addressForm.receiver_phone" class="form-input" placeholder="手机号" /></div>
          </div>
          <div class="form-item"><label class="form-label">地区 *</label><input v-model="addressForm.region" class="form-input" placeholder="省 市 区" /></div>
          <div class="form-item"><label class="form-label">详细地址 *</label><input v-model="addressForm.detail" class="form-input" placeholder="街道、门牌" /></div>
          <div class="form-row">
            <div class="form-item"><label class="form-label">邮编</label><input v-model="addressForm.postal_code" class="form-input" placeholder="选填" /></div>
            <div class="form-item">
              <label class="form-label">默认地址</label>
              <label class="switch"><input type="checkbox" :checked="addressForm.is_default===1" @change="addressForm.is_default=addressForm.is_default===1?0:1" /><span class="switch-slider"></span></label>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="addressDialogVisible=false">取消</button>
          <button class="btn-save" :disabled="savingAddress" @click="handleSaveAddress">{{ savingAddress?'保存中...':'保存' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { updateUserInfo, getAddressList, addAddress, updateAddress, deleteAddress } from '@/api/user'
import type { Address, UpdateAddressRequest } from '@/api/user'

const userStore = useUserStore()

const tabs = [
  { key: 'info', label: '基本信息', icon: '<circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/><circle cx="12" cy="12" r="3"/>' },
  { key: 'address', label: '收货地址', icon: '<path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/>' },
  { key: 'account', label: '账户信息', icon: '<rect x="2" y="3" width="20" height="18" rx="3"/><line x1="6" y1="9" x2="18" y2="9"/><line x1="6" y1="13" x2="14" y2="13"/>' },
]
const genders = [{ value:0, label:'未知' },{ value:1, label:'男' },{ value:2, label:'女' }]
const levelText = computed(() => ({ 1:'普通会员',2:'银卡会员',3:'金卡会员',4:'钻石会员' }[userStore.userInfo?.member_level||1]||'普通会员'))
const initial = computed(() => (userStore.userInfo?.nickname || userStore.userInfo?.username || 'U')[0])

const resolveUrl = (url: string) => { if(!url) return ''; if(url.startsWith('http')) return url; if(url.startsWith('/')) return 'http://localhost:8080'+url; return url }

const activeTab = ref('info')
const saving = ref(false)
const fileInput = ref<HTMLInputElement>()
const avatarPreview = ref('')
const formData = reactive({ nickname:'',avatar:'',birthday:'',gender:0,phone:'',email:'' })

const initForm = () => {
  if(!userStore.userInfo) return
  formData.nickname = userStore.userInfo.nickname || ''
  formData.avatar = userStore.userInfo.avatar || ''
  avatarPreview.value = resolveUrl(userStore.userInfo.avatar || '')
  formData.birthday = userStore.userInfo.birthday || ''
  formData.gender = userStore.userInfo.gender || 0
  formData.phone = userStore.userInfo.phone || ''
  formData.email = userStore.userInfo.email || ''
}

const triggerUpload = () => fileInput.value?.click()
const handleFileSelect = async (e: Event) => {
  const f = (e.target as HTMLInputElement).files?.[0]; if(!f) return
  if(f.size > 10*1024*1024) { ElMessage.warning('图片不能超过10MB'); return }
  try {
    const fd = new FormData(); fd.append('file', f)
    const res = await fetch('http://localhost:8080/api/v1/files/upload', { method:'POST', headers:{Authorization:`Bearer ${userStore.token}`}, body:fd })
    const d = await res.json()
    const url = d.data?.file_url || d.data?.fileUrl
    if(d.code===0 && url) { formData.avatar = url; avatarPreview.value = resolveUrl(url); ElMessage.success('头像上传成功') }
    else ElMessage.error(d.message||'上传失败')
  } catch { ElMessage.error('上传失败') }
  (e.target as HTMLInputElement).value = ''
}

const handleSave = async () => {
  saving.value = true
  try {
    await updateUserInfo({ nickname:formData.nickname, avatar:formData.avatar, gender:formData.gender, birthday:formData.birthday||undefined, phone:formData.phone||undefined, email:formData.email||undefined })
    ElMessage.success('保存成功'); await userStore.fetchUserInfo(); initForm()
  } catch(e:any) { ElMessage.error(e.message||'保存失败') }
  finally { saving.value = false }
}

// ====== 地址 ======
const addresses = ref<Address[]>([])
const addressDialogVisible = ref(false)
const savingAddress = ref(false)
const editingAddress = ref<Address|null>(null)
const addressForm = reactive({ receiver_name:'',receiver_phone:'',region:'',detail:'',postal_code:'',is_default:0 })
const getUserId = () => userStore.userId || userStore.userInfo?.id || 0

const parseRegion = (p:string,c:string,d:string) => [p,c,d].filter(Boolean).join(' ')
const splitRegion = (r:string):{province:string,city:string,district:string} => {
  const parts = r.split(/\s+/).filter(Boolean)
  if(parts.length>=3) return { province:parts[0],city:parts[1],district:parts.slice(2).join(' ') }
  return { province:parts[0]||'',city:parts[1]||'',district:parts.slice(2).join(' ')||'' }
}

const fetchAddresses = async () => {
  try { const r = await getAddressList(getUserId()); if(r.code===0) addresses.value = r.data||[] } catch {}
}
const openAddressDialog = (addr?: Address) => {
  if(addr) {
    editingAddress.value = addr
    addressForm.receiver_name = addr.receiver_name; addressForm.receiver_phone = addr.receiver_phone
    addressForm.region = parseRegion(addr.province,addr.city,addr.district)
    addressForm.detail = addr.detail; addressForm.postal_code = addr.postal_code||''
    addressForm.is_default = addr.is_default
  } else {
    editingAddress.value = null
    addressForm.receiver_name=addressForm.receiver_phone=addressForm.region=addressForm.detail=addressForm.postal_code=''
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
    const {province,city,district} = splitRegion(addressForm.region)
    if (!province || !city) { ElMessage.warning('地区格式错误，请用空格分隔，如：广东省 深圳市 南山区'); savingAddress.value = false; return }
    const base = { user_id:getUserId(),receiver_name:addressForm.receiver_name.trim(),receiver_phone:addressForm.receiver_phone.trim(),province,city,district,detail:addressForm.detail.trim(),postal_code:addressForm.postal_code,is_default:addressForm.is_default }
    if(editingAddress.value) { await updateAddress(editingAddress.value.id,{id:editingAddress.value.id,...base}); ElMessage.success('已更新') }
    else { await addAddress(base); ElMessage.success('已添加') }
    addressDialogVisible.value = false; await fetchAddresses()
  } catch(e:any) { ElMessage.error(e.message||'操作失败') }
  finally { savingAddress.value = false }
}
const handleDeleteAddress = async (id:number) => {
  try { await ElMessageBox.confirm('确定删除？','提示',{type:'warning'}); await deleteAddress(id,getUserId()); ElMessage.success('已删除'); await fetchAddresses() }
  catch(e:any) { if(e!=='cancel') ElMessage.error(e.message||'删除失败') }
}

onMounted(async () => {
  if(!userStore.userInfo) await userStore.fetchUserInfo()
  initForm(); fetchAddresses()
})
</script>

<style scoped>
.profile-page { --accent:#00F5FF; --accent-dim:rgba(0,245,255,0.08); --bg:#0A0F1C; --card-bg:rgba(255,255,255,0.02); --text:#EDF0F5; --text-dim:#8890A5; --border:rgba(255,255,255,0.06); --radius:16px; --radius-sm:10px; max-width:1100px; margin:0 auto; padding:32px 24px; min-height:calc(100vh - 64px); font-family:'Inter','PingFang SC','SF Pro Display',-apple-system,sans-serif; }
.page-title { font-size:28px; font-weight:700; margin:0 0 32px; color:var(--text); letter-spacing:-.01em; }
.layout { display:flex; gap:24px; align-items:flex-start; }

/* Side Nav */
.side-nav { width:180px; flex-shrink:0; background:var(--card-bg); border:1px solid var(--border); border-radius:var(--radius); padding:8px; position:sticky; top:88px; }
.nav-item { display:flex; align-items:center; gap:10px; padding:12px 14px; border-radius:var(--radius-sm); cursor:pointer; font-size:14px; font-weight:500; color:var(--text-dim); transition:all .15s; margin-bottom:2px; }
.nav-item:hover { color:var(--text); background:var(--accent-dim); }
.nav-item.active { color:var(--accent); background:var(--accent-dim); font-weight:600; }
.nav-icon { width:18px; height:18px; flex-shrink:0; fill:none; stroke:currentColor; stroke-width:2; stroke-linecap:round; stroke-linejoin:round; }

/* Main Content */
.main-content { flex:1; min-width:0; }

/* Card */
.card { background:var(--card-bg); border:1px solid var(--border); border-radius:var(--radius); padding:28px; margin-bottom:20px; }
.card-head { display:flex; justify-content:space-between; align-items:center; margin-bottom:20px; }
.card-title { font-size:18px; font-weight:600; margin:0; color:var(--text); }

/* Avatar */
.info-layout { display:flex; gap:40px; align-items:flex-start; }
.avatar-col { text-align:center; flex-shrink:0; }
.avatar-upload { position:relative; width:100px; height:100px; border-radius:50%; cursor:pointer; margin:0 auto 8px; }
.avatar-preview { width:100%;height:100%;border-radius:50%;overflow:hidden;background:var(--accent-dim);display:flex;align-items:center;justify-content:center;border:3px solid transparent;transition:border-color .2s; }
.avatar-preview img { width:100%;height:100%;object-fit:cover; }
.avatar-letter { font-size:36px;font-weight:700;color:var(--accent); }
.avatar-upload:hover .avatar-preview { border-color:var(--accent); }
.avatar-overlay { position:absolute;inset:0;border-radius:50%;background:rgba(0,0,0,.5);color:#fff;font-size:12px;font-weight:500;display:flex;align-items:center;justify-content:center;opacity:0;transition:opacity .2s; }
.avatar-upload:hover .avatar-overlay { opacity:1; }
.avatar-hint { font-size:12px;color:var(--text-dim);margin:0; }

/* Form */
.form-col { flex:1; display:flex; flex-direction:column; gap:16px; }
.form-item { flex:1; }
.form-label { display:block; font-size:13px; color:var(--text-dim); margin-bottom:6px; font-weight:500; }
.form-input { width:100%; padding:10px 14px; border-radius:var(--radius-sm); background:rgba(255,255,255,0.04); border:1px solid var(--border); color:var(--text); font-size:14px; outline:none; transition:all .2s; box-sizing:border-box; font-family:inherit; }
.form-input::placeholder { color:var(--text-dim); }
.form-input:focus { border-color:var(--accent); box-shadow:0 0 0 3px rgba(0,245,255,.1); }
.form-input[type=date] { color-scheme:dark; }
.form-value.disabled { padding:10px 14px; font-size:14px; color:var(--text-dim); background:rgba(255,255,255,0.02); border-radius:var(--radius-sm); }
.form-row { display:flex; gap:16px; }
.radio-group { display:inline-flex; gap:2px; background:rgba(255,255,255,0.04); border-radius:100px; padding:3px; }
.radio-item { padding:7px 18px; border-radius:100px; cursor:pointer; font-size:13px; color:var(--text-dim); transition:all .15s; }
.radio-item:hover { color:var(--text); }
.radio-item.active { background:var(--accent); color:#0A0F1C; font-weight:600; }

/* Buttons */
.btn-save { padding:12px 32px; border-radius:100px; border:none; background:var(--accent); color:#0A0F1C; font-size:14px; font-weight:600; cursor:pointer; transition:all .2s; }
.btn-save:hover:not(:disabled) { box-shadow:0 0 24px rgba(0,245,255,.3); transform:translateY(-1px); }
.btn-save:disabled { opacity:.5;cursor:not-allowed; }
.btn-add { padding:8px 18px; border-radius:100px; border:1px solid var(--accent); background:transparent; color:var(--accent); font-size:13px; font-weight:500; cursor:pointer; transition:all .2s; }
.btn-add:hover { background:var(--accent-dim); }
.btn-link { background:none; border:none; cursor:pointer; font-size:13px; color:var(--accent); padding:0; }
.btn-link.danger { color:#F87171; }
.btn-link:hover { opacity:.8; }

/* Address */
.address-list { display:flex; flex-direction:column; gap:12px; }
.address-card { background:rgba(255,255,255,0.02); border:1px solid var(--border); border-radius:var(--radius-sm); padding:16px; display:flex; justify-content:space-between; align-items:center; transition:border-color .2s; }
.address-card.is-default { border-color:var(--accent); }
.addr-line1 { display:flex; align-items:center; gap:12px; margin-bottom:6px; font-size:14px; }
.phone { color:var(--text-dim); }
.default-tag { font-size:11px; background:var(--accent); color:#0A0F1C; font-weight:600; padding:2px 8px; border-radius:4px; }
.addr-line2 { font-size:13px; color:var(--text-dim); }
.addr-actions { display:flex; gap:12px; }
.empty-block { text-align:center; padding:48px 0; color:var(--text-dim); font-size:14px; }

/* Account */
.account-grid { display:grid; grid-template-columns:repeat(2,1fr); gap:16px; }
.account-item { padding:16px; background:rgba(255,255,255,0.02); border-radius:var(--radius-sm); display:flex; flex-direction:column; gap:6px; }
.a-label { font-size:12px; color:var(--text-dim); }
.a-value { font-size:16px; color:var(--text); font-weight:500; }
.a-value.accent { color:var(--accent); font-weight:700; }

/* Modal */
.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.6); backdrop-filter:blur(4px); z-index:200; display:flex; align-items:center; justify-content:center; }
.modal-card { background:#111827; border:1px solid var(--border); border-radius:var(--radius); padding:28px; width:520px; max-height:90vh; overflow-y:auto; }
.modal-title { font-size:18px; font-weight:600; margin:0 0 24px; color:var(--text); }
.modal-body { display:flex; flex-direction:column; gap:16px; }
.modal-footer { display:flex; justify-content:flex-end; gap:12px; margin-top:24px; }
.btn-cancel { padding:10px 24px; border-radius:100px; border:1px solid var(--border); background:transparent; color:var(--text); font-size:14px; cursor:pointer; transition:all .2s; }
.btn-cancel:hover { background:rgba(255,255,255,0.04); }
.switch { position:relative; display:inline-block; width:44px; height:26px; }
.switch input { opacity:0;width:0;height:0; }
.switch-slider { position:absolute;cursor:pointer;inset:0;background:rgba(255,255,255,0.1);border-radius:26px;transition:.2s; }
.switch-slider:before { position:absolute;content:'';height:20px;width:20px;left:3px;bottom:3px;background:var(--text);border-radius:50%;transition:.2s; }
.switch input:checked+.switch-slider { background:var(--accent); }
.switch input:checked+.switch-slider:before { transform:translateX(18px); }

@media(max-width:768px) { .layout { flex-direction:column; } .side-nav { width:100%;display:flex;flex-direction:row;position:static;gap:4px; } .nav-item { flex:1;justify-content:center; } .info-layout { flex-direction:column;align-items:center; } .form-row { flex-direction:column; } .account-grid { grid-template-columns:1fr; } }
</style>
