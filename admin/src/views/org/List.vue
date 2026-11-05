<template>
  <div class="org-page">
    <SmartTable
      ref="tableRef"
      card-title="组织管理"
      :loader="loadOrgTree"
      :columns="columns"
      :show-pagination="false"
      :show-keyword-search="false"
      :show-create="isAdmin"
      create-text="新建根组织"
      :operator-width="300"
      @create="openCreateRoot"
      @load-success="handleLoadSuccess"
      @load-error="handleLoadError"
    >
      <template #toolbar-left>
        <div class="org-tip">
          <el-icon><InfoFilled /></el-icon>
          <span v-if="isAdmin">平台管理员可管理所有组织；负责人和成员都可以管理该组织下的设备与子级组织。</span>
          <span v-else>你只能查看自己负责 / 加入的组织及其子级组织；关联组织本身不可编辑或删除，其成员由负责人管理。</span>
        </div>
      </template>

      <template #name="{ row }">
        <div class="name-cell">
          <el-icon><OfficeBuilding /></el-icon>
          <span class="org-name">{{ row.name }}</span>
          <el-tag size="small" :type="typeTag(row.type)">{{ typeLabel(row.type) }}</el-tag>
          <el-tag v-if="row.is_owner" size="small" type="success">我负责</el-tag>
          <el-tag v-if="row.is_member" size="small" type="info">成员</el-tag>
        </div>
      </template>

      <template #owner="{ row }">{{ row.owner_name || row.owner_user_id || '-' }}</template>

      <template #region="{ row }">{{ row.region || '-' }}</template>

      <template #actions="{ row }">
        <el-button link type="primary" @click="openCreateChild(row)">添加子级</el-button>
        <el-button link type="primary" :disabled="!row.can_manage" @click="openEdit(row)">编辑</el-button>
        <el-button link type="danger" :disabled="!row.can_manage" @click="handleDelete(row)">删除</el-button>
        <el-button link type="primary" :disabled="!row.can_members" @click="openMembers(row)">成员</el-button>
      </template>
    </SmartTable>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="560px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="96px">
        <el-form-item label="上级组织">
          <el-tree-select
            v-model="form.parent_id"
            :data="parentOptions"
            check-strictly
            :render-after-expand="false"
            :disabled="isEdit"
            placeholder="请选择上级组织"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="组织名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入组织名称" maxlength="64" />
        </el-form-item>
        <el-form-item label="组织类型">
          <el-select v-model="form.type" placeholder="请选择类型" style="width: 100%">
            <el-option label="平台" value="platform" />
            <el-option label="代理商" value="agent" />
            <el-option label="经销商" value="dealer" />
            <el-option label="门店" value="store" />
          </el-select>
        </el-form-item>
        <el-form-item label="负责人">
          <el-input :model-value="form.owner_display" readonly placeholder="点击右侧按钮选择负责人">
            <template #append>
              <el-button @click="openUserPicker('owner')">选择</el-button>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="所在区域">
          <el-input v-model="form.region" placeholder="如：广东省深圳市" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item v-if="isEdit" label="状态">
          <el-switch v-model="form.disabled" :active-text="form.disabled ? '停用' : '正常'" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="memberDialogVisible" :title="`成员管理 - ${memberOrg.name || ''}`" width="640px">
      <div class="member-owner">
        <span class="member-label">负责人：</span>
        <span>{{ memberOwnerName || memberOrg.owner_name || memberOrg.owner_user_id || '未设置' }}</span>
        <el-button class="member-add" size="small" type="primary" :icon="Plus" @click="openUserPicker('member')">
          添加成员
        </el-button>
      </div>
      <el-table :data="members" v-loading="membersLoading" size="small" max-height="360">
        <el-table-column label="昵称" min-width="120">
          <template #default="{ row }">{{ row.nickname || '-' }}</template>
        </el-table-column>
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column label="手机号" width="140">
          <template #default="{ row }">{{ row.cellphone || '-' }}</template>
        </el-table-column>
        <el-table-column label="加入时间" width="170">
          <template #default="{ row }">{{ $formatDateTime(row.created) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="90">
          <template #default="{ row }">
            <el-button link type="danger" @click="handleRemoveMember(row)">移除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="member-hint">负责人和成员都可以管理该组织下的设备及子级组织。</div>
      <template #footer>
        <el-button @click="memberDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="pickerVisible"
      :title="pickerMode === 'owner' ? '选择负责人' : '添加成员'"
      width="600px"
    >
      <div class="picker-search">
        <el-input
          v-model="pickerKeyword"
          placeholder="搜索用户名 / 昵称 / 手机号 / 邮箱"
          clearable
          @clear="loadCandidates"
          @keyup.enter="loadCandidates"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button type="primary" @click="loadCandidates">搜索</el-button>
      </div>
      <el-table
        :data="candidates"
        v-loading="pickerLoading"
        height="340"
        highlight-current-row
        @row-click="selectCandidate"
      >
        <el-table-column label="昵称" min-width="140">
          <template #default="{ row }">{{ row.nickname || '-' }}</template>
        </el-table-column>
        <el-table-column prop="username" label="用户名" min-width="140" />
        <el-table-column label="手机号" width="140">
          <template #default="{ row }">{{ row.cellphone || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-button link type="primary" @click.stop="selectCandidate(row)">选择</el-button>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="pickerVisible = false">取消</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, reactive, ref, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, InfoFilled, OfficeBuilding, Search } from '@element-plus/icons-vue'
import SmartTable from '@/components/SmartTable.vue'
import {
  getOrgTree,
  createOrg,
  updateOrg,
  deleteOrg,
  searchUsers,
  getOrgMembers,
  addOrgMember,
  removeOrgMember
} from '@/api/org'
import { useUserStore } from '@/store'

const userStore = useUserStore()
const isAdmin = computed(() => userStore.isAdmin)

const tableRef = ref(null)
const tree = ref([])
const dialogVisible = ref(false)
const submitting = ref(false)
const isEdit = ref(false)
const editingId = ref(0)
const formRef = ref(null)

const form = reactive({
  parent_id: 0,
  name: '',
  type: 'agent',
  owner_user_id: '',
  owner_display: '',
  region: '',
  sort: 0,
  disabled: false
})

const rules = {
  name: [{ required: true, message: '请输入组织名称', trigger: 'blur' }]
}

const dialogTitle = computed(() => (isEdit.value ? '编辑组织' : '新建组织'))

const columns = [
  { prop: 'name', label: '组织名称', minWidth: 240, slot: 'name', showOverflowTooltip: false },
  { prop: 'level', label: '层级', width: 80, align: 'center' },
  { prop: 'path', label: '路径', width: 140 },
  { label: '负责人', width: 160, slot: 'owner' },
  { label: '区域', width: 140, slot: 'region' },
  { prop: 'sort', label: '排序', width: 80, align: 'center' },
  {
    prop: 'disabled',
    label: '状态',
    width: 90,
    type: 'tag',
    formatter: (row) => (row.disabled ? '停用' : '正常'),
    tagType: (row) => (row.disabled ? 'danger' : 'success')
  }
]

const typeLabel = (type) => {
  const map = { platform: '平台', agent: '代理商', dealer: '经销商', store: '门店' }
  return map[type] || type || '未分类'
}

const typeTag = (type) => {
  const map = { platform: 'danger', agent: 'warning', dealer: 'success', store: 'info' }
  return map[type] || 'info'
}

// 上级组织选择：管理员可建根组织（0），其他用户只能选可管理组织树内的节点
const parentOptions = computed(() => {
  const convert = (nodes) =>
    (nodes || []).map((n) => ({
      value: n.id,
      label: n.name,
      children: convert(n.children)
    }))
  const list = convert(tree.value)
  if (isAdmin.value) {
    return [{ value: 0, label: '顶级组织（无上级）', children: list }]
  }
  return list
})

// 组织树为非分页 GET 接口
const loadOrgTree = () => getOrgTree()

const handleLoadSuccess = async (res) => {
  tree.value = res?.data || []
  await nextTick()
  // 默认展开全部层级，保持原树形表格的展示效果
  const expand = (nodes) => {
    (nodes || []).forEach((node) => {
      if (node.children?.length) {
        tableRef.value?.tableRef?.toggleRowExpansion(node, true)
        expand(node.children)
      }
    })
  }
  expand(tree.value)
}

const handleLoadError = () => {
  ElMessage.error('加载组织树失败')
}

const resetForm = () => {
  Object.assign(form, {
    parent_id: 0,
    name: '',
    type: 'agent',
    owner_user_id: '',
    owner_display: '',
    region: '',
    sort: 0,
    disabled: false
  })
  editingId.value = 0
  isEdit.value = false
}

const openCreateRoot = () => {
  resetForm()
  form.parent_id = 0
  dialogVisible.value = true
}

const openCreateChild = (row) => {
  resetForm()
  form.parent_id = row.id
  form.type = row.type === 'agent' ? 'dealer' : 'store'
  dialogVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  editingId.value = row.id
  Object.assign(form, {
    parent_id: row.parent_id || 0,
    name: row.name || '',
    type: row.type || '',
    owner_user_id: row.owner_user_id || '',
    owner_display: row.owner_name || row.owner_user_id || '',
    region: row.region || '',
    sort: row.sort || 0,
    disabled: !!row.disabled
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    const data = {
      name: form.name,
      type: form.type,
      owner_user_id: form.owner_user_id || '',
      region: form.region,
      sort: form.sort,
      disabled: form.disabled
    }
    if (isEdit.value) {
      await updateOrg(editingId.value, data)
      ElMessage.success('组织已更新')
    } else {
      await createOrg({ parent_id: form.parent_id, ...data })
      ElMessage.success('组织已创建')
    }
    dialogVisible.value = false
    tableRef.value?.refresh()
  } catch (error) {
    console.error(error)
    ElMessage.error(error.response?.data?.message || '保存失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(
    `确定删除组织「${row.name}」吗？存在子组织、成员、用户或绑定设备时无法删除。`,
    '删除确认',
    { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
  )
    .then(async () => {
      await deleteOrg(row.id)
      ElMessage.success('组织已删除')
      tableRef.value?.refresh()
    })
    .catch(() => {})
}

// ========== 成员管理 ==========

const memberDialogVisible = ref(false)
const memberOrg = ref({})
const members = ref([])
const memberOwnerName = ref('')
const membersLoading = ref(false)

const openMembers = async (row) => {
  memberOrg.value = row
  memberDialogVisible.value = true
  await loadMembers()
}

const loadMembers = async () => {
  if (!memberOrg.value.id) return
  membersLoading.value = true
  try {
    const res = await getOrgMembers(memberOrg.value.id)
    const data = res.data || {}
    members.value = data.members || []
    memberOwnerName.value = data.owner_name || ''
  } catch (error) {
    console.error(error)
    ElMessage.error('加载成员失败')
  } finally {
    membersLoading.value = false
  }
}

const handleRemoveMember = (row) => {
  ElMessageBox.confirm(`确定将「${row.nickname || row.username || row.user_id}」移出组织吗？`, '移除确认', {
    type: 'warning'
  })
    .then(async () => {
      await removeOrgMember(memberOrg.value.id, row.user_id)
      ElMessage.success('成员已移除')
      await loadMembers()
    })
    .catch(() => {})
}

// ========== 用户选择器（负责人 / 成员） ==========

const pickerVisible = ref(false)
const pickerMode = ref('owner')
const pickerKeyword = ref('')
const pickerLoading = ref(false)
const candidates = ref([])

const openUserPicker = (mode) => {
  pickerMode.value = mode
  pickerKeyword.value = ''
  candidates.value = []
  pickerVisible.value = true
  loadCandidates()
}

const loadCandidates = async () => {
  pickerLoading.value = true
  try {
    const res = await searchUsers(pickerKeyword.value)
    candidates.value = res.data || []
  } catch (error) {
    console.error(error)
    ElMessage.error('搜索用户失败')
  } finally {
    pickerLoading.value = false
  }
}

const userLabel = (u) => (u.nickname ? `${u.nickname}（${u.username}）` : u.username)

const selectCandidate = async (row) => {
  if (pickerMode.value === 'owner') {
    form.owner_user_id = row.id
    form.owner_display = userLabel(row)
    pickerVisible.value = false
    return
  }
  try {
    await addOrgMember(memberOrg.value.id, row.id)
    ElMessage.success('成员已添加')
    pickerVisible.value = false
    await loadMembers()
  } catch (error) {
    console.error(error)
    ElMessage.error(error.response?.data?.message || '添加成员失败')
  }
}
</script>

<style lang="scss" scoped>
.org-page {
  padding: 16px;
}

.org-tip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  color: #606266;
  font-size: 13px;
  background: #f4f8ff;
  border-radius: 6px;
}

.name-cell {
  display: flex;
  align-items: center;
  gap: 8px;

  .org-name {
    font-weight: 600;
  }
}

.member-owner {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 12px;
  font-size: 14px;

  .member-label {
    color: #909399;
  }

  .member-add {
    margin-left: auto;
  }
}

.member-hint {
  margin-top: 10px;
  color: #909399;
  font-size: 12px;
}

.picker-search {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}
</style>
