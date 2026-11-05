<template>
  <div class="account-list">
    <SmartTable
      ref="tableRef"
      card-title="账号管理"
      :loader="listAccounts"
      :columns="columns"
      :operators="operators"
      :show-create="false"
      :show-selection="true"
      :operator-width="220"
      keyword-placeholder="搜索用户名/邮箱/手机号"
      storage-key="account"
      row-clickable
      @row-click="handleViewDetail"
      @selection-change="rows => (selectedRows = rows)"
    >
      <template #toolbar-right>
        <el-button type="primary" :icon="Plus" @click="openCreateDrawer">创建账号</el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="batchToggleDisable">
          批量禁用 <span v-if="selectedRows.length">({{ selectedRows.length }})</span>
        </el-button>
      </template>

      <template #avatar="{ row }">
        <div class="avatar-cell">
          <el-avatar :size="40" :src="getAvatarUrl(row.avatar)">
            {{ row.username?.[0]?.toUpperCase() || 'U' }}
          </el-avatar>
        </div>
      </template>

      <template #org="{ row }">
        <el-tag v-if="row.org_ids?.length" size="small" type="info">{{ orgName(row.org_ids[0]) }}</el-tag>
        <span v-else class="org-missing">未归属组织</span>
      </template>

      <template #accountType="{ row }">
        <el-tag size="small" :type="row.admin || row.platform ? 'success' : 'info'">
          {{ row.admin || row.platform ? '平台账号' : '用户账号' }}
        </el-tag>
      </template>
    </SmartTable>

    <AccountForm
      v-model:visible="drawerVisible"
      :model-value="currentEditData"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SmartTable from '@/components/SmartTable.vue'
import { listAccounts, createAccount, updateAccount } from '@/api/account'
import { getOrgTree } from '@/api/org'
import AccountForm from './AccountForm.vue'

const router = useRouter()
const tableRef = ref(null)

const baseUrl = import.meta.env.VITE_API_BASE_URL || ''

const getAvatarUrl = (url) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return baseUrl + url
}

const columns = [
  { prop: 'avatar', label: '头像', width: 80, align: 'center', slot: 'avatar', showOverflowTooltip: false },
  { prop: 'username', label: '用户名', minWidth: 120 },
  { label: '所属组织', minWidth: 140, slot: 'org' },
  { label: '账号类型', width: 110, slot: 'accountType' },
  { prop: 'email', label: '邮箱', minWidth: 180 },
  { prop: 'cellphone', label: '手机号', width: 150 },
  {
    prop: 'disabled',
    label: '状态',
    width: 100,
    type: 'tag',
    formatter: (row) => (row.disabled ? '禁用' : '正常'),
    tagType: (row) => (row.disabled ? 'danger' : 'success')
  },
  { prop: 'created', label: '创建时间', width: 180, type: 'datetime' }
]

const operators = [
  { key: 'detail', label: '详情', type: 'primary', handler: (row) => handleViewDetail(row) },
  { key: 'edit', label: '编辑', type: 'primary', handler: (row) => openEditDrawer(row) },
  {
    key: 'toggle',
    label: (row) => (row.disabled ? '启用' : '禁用'),
    type: (row) => (row.disabled ? 'success' : 'warning'),
    handler: (row) => handleToggleStatus(row)
  }
]

const drawerVisible = ref(false)
const currentEditId = ref(null)
const currentEditData = ref({})
const selectedRows = ref([])
const orgMap = ref({})

const orgName = (orgId) => orgMap.value[orgId] || `组织#${orgId}`

const loadOrgs = async () => {
  try {
    const res = await getOrgTree()
    const map = {}
    const walk = (nodes) => {
      (nodes || []).forEach((n) => {
        map[n.id] = n.name
        walk(n.children)
      })
    }
    walk(res.data)
    orgMap.value = map
  } catch (error) {
    console.error('加载组织失败', error)
  }
}

const openCreateDrawer = () => {
  currentEditId.value = null
  currentEditData.value = {}
  drawerVisible.value = true
}

const openEditDrawer = (row) => {
  currentEditId.value = row.id
  currentEditData.value = row
  drawerVisible.value = true
}

const handleFormSubmit = async (formData) => {
  try {
    if (currentEditId.value) {
      const updateData = {
        nickname: formData.nickname,
        email: formData.email,
        cellphone: formData.cellphone,
        disabled: formData.disabled,
        org_id: formData.org_id,
        platform: formData.platform
      }
      if (formData.password) {
        updateData.password = formData.password
      }
      await updateAccount(currentEditId.value, updateData)
    } else {
      await createAccount({
        username: formData.username,
        password: formData.password,
        nickname: formData.nickname,
        avatar: formData.avatar,
        email: formData.email,
        cellphone: formData.cellphone,
        org_id: formData.org_id,
        platform: formData.platform
      })
    }
    drawerVisible.value = false
    tableRef.value?.refresh()
  } catch (error) {
    console.error(error)
    ElMessage.error(error.response?.data?.message || (currentEditId.value ? '更新失败' : '创建失败'))
  }
}

const handleToggleStatus = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定${row.disabled ? '启用' : '禁用'}账号"${row.username}"吗？`,
      '提示',
      { type: 'warning' }
    )
    await updateAccount(row.id, { disabled: !row.disabled })
    ElMessage.success(row.disabled ? '已启用' : '已禁用')
    tableRef.value?.refresh()
  } catch (error) {
    if (error !== 'cancel') console.error(error)
  }
}

const handleViewDetail = (row) => {
  router.push(`/accounts/${row.id}`)
}

const batchToggleDisable = async () => {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(`确定禁用选中的 ${selectedRows.value.length} 个账号吗？`, '提示', { type: 'warning' })
    const ids = selectedRows.value.map(r => r.id)
    await Promise.all(ids.map(id => updateAccount(id, { disabled: true }).catch(() => null)))
    ElMessage.success('批量禁用完成')
    tableRef.value?.refresh()
  } catch (e) { /* cancel */ }
}

onMounted(() => {
  loadOrgs()
})
</script>

<style lang="scss" scoped>
.account-list {
  .avatar-cell {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .org-missing {
    color: #909399;
    font-size: 13px;
  }
}
</style>
