<template>
  <div class="user-page">
    <PageHeader title="用户管理" desc="管理平台注册用户">
      <template #actions>
        <button class="btn-add" @click="handleAdd">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          新增用户
        </button>
      </template>
    </PageHeader>

    <!-- 搜索栏 -->
    <div class="search-bar">
      <div class="search-box">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="search-icon"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
        <input v-model="searchKeyword" type="text" placeholder="搜索用户名 / 手机号 / 邮箱..." class="search-input" @keyup.enter="handleSearch" />
      </div>
      <select v-model="searchStatus" class="filter-select" @change="handleSearch">
        <option :value="null">全部状态</option>
        <option :value="1">正常</option>
        <option :value="0">禁用</option>
      </select>
    </div>

    <!-- 表格 -->
    <DarkCard v-loading="loading" no-pad>
      <el-table :data="userList" class="dark-table" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="用户" min-width="180">
          <template #default="{ row }">
            <div class="user-cell">
              <div class="user-avatar" :class="levelClass(row.member_level)">
                <img v-if="resolveAvatar(row.avatar)" :src="resolveAvatar(row.avatar)" class="avatar-img" @error="e=>(e.target as HTMLImageElement).style.display='none'" />
                <span v-else>{{ (row.nickname || row.username || 'U')[0] }}</span>
              </div>
              <div class="user-info">
                <span class="user-name">{{ row.username }}</span>
                <span class="user-nick">{{ row.nickname || '-' }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column prop="email" label="邮箱" min-width="160" show-overflow-tooltip />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <span class="status-dot" :class="row.status===1?'on':'off'">{{ row.status===1?'正常':'禁用' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="会员" width="80">
          <template #default="{ row }">
            <span class="level-tag" :class="levelClass(row.member_level)">{{ levelText(row.member_level) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="points" label="积分" width="70" />
        <el-table-column label="注册时间" width="160">
          <template #default="{ row }">{{ row.created_at?.slice(0,10) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <button class="tbl-btn edit" @click="handleEdit(row)">编辑</button>
            <button class="tbl-btn del" @click="handleDelete(row)">删除</button>
          </template>
        </el-table-column>
      </el-table>
    </DarkCard>

    <!-- 分页 -->
    <div class="pagination" v-if="total > 0">
      <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" :total="total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="fetchList" @current-change="fetchList" />
    </div>

    <!-- ======== 对话框 ======== -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="520px" @close="handleDialogClose">
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="formData.username" :disabled="isEdit" placeholder="用户名" maxlength="50" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!isEdit">
          <el-input v-model="formData.password" type="password" show-password placeholder="6位以上" />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="formData.nickname" placeholder="昵称（选填）" maxlength="20" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="手机号"><el-input v-model="formData.phone" placeholder="手机号" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="邮箱"><el-input v-model="formData.email" placeholder="邮箱" /></el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="性别">
              <el-radio-group v-model="formData.gender">
                <el-radio :value="0">未知</el-radio>
                <el-radio :value="1">男</el-radio>
                <el-radio :value="2">女</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态" v-if="isEdit">
              <el-radio-group v-model="formData.status">
                <el-radio :value="1">正常</el-radio>
                <el-radio :value="0">禁用</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible=false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { getUserList, deleteUser, register, updateUserInfo, type UserInfo, type ListUsersParams } from '@/api/user'
import PageHeader from '@/components/PageHeader.vue'
import DarkCard from '@/components/DarkCard.vue'

const loading = ref(false); const userList = ref<UserInfo[]>([])
const currentPage = ref(1); const pageSize = ref(10); const total = ref(0)
const searchKeyword = ref(''); const searchStatus = ref<number|null>(null)
const dialogVisible = ref(false); const dialogTitle = ref('新增用户'); const isEdit = ref(false)
const submitting = ref(false); const formRef = ref<FormInstance>()
const formData = ref({ id:0, username:'', password:'', nickname:'', phone:'', email:'', gender:0, status:1 })

const formRules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }, { min: 6, message: '至少6位', trigger: 'blur' }],
  email: [{ type: 'email', message: '邮箱格式不正确', trigger: 'blur' }],
}

const resolveAvatar = (url: string) => { if (!url) return ''; if (url.startsWith('http')) return url; if (url.startsWith('/')) return 'http://localhost:8080' + url; return '' }
const levelText = (l: number) => ({ 0:'普通',1:'VIP1',2:'VIP2',3:'VIP3' }[l]||'普通')
const levelClass = (l: number) => l >= 3 ? 'vip3' : l >= 2 ? 'vip2' : l >= 1 ? 'vip1' : ''

const fetchList = async () => {
  loading.value=true
  try {
    const params: ListUsersParams = { page: currentPage.value, page_size: pageSize.value }
    if (searchKeyword.value) params.keyword = searchKeyword.value
    if (searchStatus.value !== null) params.status = searchStatus.value
    const r = await getUserList(params) as any
    if (r.code===0) { userList.value = r.data.users||r.data.list||[]; total.value = r.data.total||0 }
  } catch { ElMessage.error('获取失败') } finally { loading.value=false }
}

const handleSearch = () => { currentPage.value=1; fetchList() }
const handleAdd = () => { isEdit.value=false; dialogTitle.value='新增用户'; formData.value={ id:0, username:'', password:'', nickname:'', phone:'', email:'', gender:0, status:1 }; dialogVisible.value=true }
const handleEdit = (row: UserInfo) => { isEdit.value=true; dialogTitle.value='编辑用户'; formData.value={ id:row.id, username:row.username, password:'', nickname:row.nickname||'', phone:row.phone||'', email:row.email||'', gender:row.gender, status:row.status }; dialogVisible.value=true }
const handleDelete = async (row: UserInfo) => {
  try { await ElMessageBox.confirm('确定删除？', '提示', { type:'warning' }); await deleteUser(row.id); ElMessage.success('已删除'); fetchList() }
  catch(e:any) { if (e!=='cancel') ElMessage.error(e.message||'删除失败') }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return; submitting.value=true
    try {
      if (isEdit.value) {
        await updateUserInfo({ nickname: formData.value.nickname, gender: formData.value.gender })
        ElMessage.success('已更新')
      } else {
        await register({ username: formData.value.username, password: formData.value.password, phone: formData.value.phone, email: formData.value.email })
        ElMessage.success('已创建')
      }
      dialogVisible.value=false; fetchList()
    } catch (e:any) { ElMessage.error(e.message||'操作失败') }
    finally { submitting.value=false }
  })
}
const handleDialogClose = () => formRef.value?.resetFields()
watch(searchStatus, () => handleSearch())
onMounted(() => fetchList())
</script>

<style scoped>
.user-page { padding: 0; }

.btn-add { display: inline-flex; align-items: center; gap: 6px; padding: 9px 20px; border-radius: 10px; border: none; background: #00F5FF; color: #0A0F1C; font-size: 13px; font-weight: 600; cursor: pointer; transition: all .2s; }
.btn-add:hover { box-shadow: 0 0 20px rgba(0,245,255,.3); transform: translateY(-1px); }
.btn-add svg { width: 16px; height: 16px; }

/* Search Bar */
.search-bar { display: flex; gap: 12px; align-items: center; margin-bottom: 20px; }
.search-box { position: relative; display: flex; align-items: center; flex: 1; }
.search-icon { position: absolute; left: 12px; width: 16px; height: 16px; color: #8890A5; pointer-events: none; }
.search-input { width: 100%; padding: 10px 16px 10px 36px; border-radius: 12px; background: rgba(255,255,255,0.04); border: 1px solid rgba(255,255,255,0.06); color: #EDF0F5; font-size: 13px; outline: none; transition: all .25s; font-family: inherit; }
.search-input::placeholder { color: #8890A5; }
.search-input:focus { border-color: #00F5FF; box-shadow: 0 0 0 3px rgba(0,245,255,.08); }
.filter-select { appearance: none; -webkit-appearance: none; padding: 10px 32px 10px 14px; border-radius: 12px; background: rgba(255,255,255,0.04); border: 1px solid rgba(255,255,255,0.06); color: #EDF0F5; font-size: 13px; cursor: pointer; outline: none; font-family: inherit; background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 24 24' fill='none' stroke='%238890A5' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E"); background-repeat: no-repeat; background-position: right 8px center; background-size: 16px; }
.filter-select:focus { border-color: #00F5FF; }
.filter-select option { background: #111827; color: #EDF0F5; }

/* User cell */
.user-cell { display: flex; align-items: center; gap: 10px; }
.user-avatar { width: 34px; height: 34px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 14px; background: rgba(0,245,255,.1); color: #00F5FF; flex-shrink: 0; overflow: hidden; }
.user-avatar .avatar-img { width: 100%; height: 100%; object-fit: cover; }
.user-avatar.vip1 { background: rgba(16,185,129,.1); color: #10B981; }
.user-avatar.vip2 { background: rgba(139,92,246,.1); color: #8B5CF6; }
.user-avatar.vip3 { background: rgba(245,158,11,.1); color: #F59E0B; }
.user-info { display: flex; flex-direction: column; gap: 2px; }
.user-name { font-size: 14px; font-weight: 500; color: #EDF0F5; }
.user-nick { font-size: 12px; color: #8890A5; }

/* Tags */
.status-dot { font-size: 11px; padding: 2px 10px; border-radius: 100px; font-weight: 600; }
.status-dot.on { background: rgba(16,185,129,.12); color: #10B981; }
.status-dot.off { background: rgba(255,255,255,.06); color: #8890A5; }
.level-tag { font-size: 11px; padding: 2px 8px; border-radius: 100px; font-weight: 600; background: rgba(255,255,255,.06); color: #8890A5; }
.level-tag.vip1 { background: rgba(16,185,129,.1); color: #10B981; }
.level-tag.vip2 { background: rgba(139,92,246,.1); color: #8B5CF6; }
.level-tag.vip3 { background: rgba(245,158,11,.1); color: #F59E0B; }

/* Action buttons */
.tbl-btn { padding: 4px 12px; border-radius: 6px; border: 1px solid transparent; background: transparent; cursor: pointer; font-size: 12px; transition: all .15s; }
.tbl-btn.edit { color: #00F5FF; }
.tbl-btn.edit:hover { background: rgba(0,245,255,.1); border-color: rgba(0,245,255,.2); }
.tbl-btn.del { color: #F87171; }
.tbl-btn.del:hover { background: rgba(248,113,113,.1); border-color: rgba(248,113,113,.2); }

.pagination { margin-top: 20px; display: flex; justify-content: flex-end; }
</style>
