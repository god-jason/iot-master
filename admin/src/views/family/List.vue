<template>
  <div class="family-list">
    <SmartTable
      ref="tableRef"
      card-title="家庭管理"
      :loader="loadFamilies"
      :columns="columns"
      :operators="operators"
      :show-selection="true"
      :show-batch-delete="true"
      :delete-fn="batchDeleteFamilies"
      :operator-width="320"
      show-create
      create-text="新建家庭"
      keyword-placeholder="搜索家庭名称"
      storage-key="family"
      @create="openCreate"
    >
      <template #family="{ row }">
        <div class="family-cell">
          <div class="family-icon" v-if="row.icon">{{ row.icon }}</div>
          <el-icon v-else class="default-icon"><HomeFilled /></el-icon>
          <span class="family-name">{{ row.name }}</span>
        </div>
      </template>

      <template #role="{ row }">
        <el-tag size="small" :type="isOwner(row) ? 'warning' : 'info'">
          {{ isOwner(row) ? '拥有者' : '成员' }}
        </el-tag>
      </template>

      <template #created="{ row }">{{ formatDate(row.created) }}</template>

      <template #empty>
        <el-empty description="还没有创建任何家庭" :image-size="80" />
      </template>
    </SmartTable>

    <!-- 创建/编辑 弹窗 -->
    <el-dialog v-model="formVisible" :title="current?.id ? '编辑家庭' : '创建家庭'" width="440px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="家庭名称" required>
          <el-input v-model="form.name" placeholder="请输入家庭名称" maxlength="32" show-word-limit />
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="form.icon" placeholder="emoji 或图标 URL" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" :title="current?.name + ' - 详情'" width="720px" top="8vh">
      <el-tabs v-model="activeTab">
        <!-- 成员 -->
        <el-tab-pane label="成员" name="members">
          <div class="tab-toolbar">
            <el-button type="primary" size="small" @click="openInviteUser">
              <el-icon><Plus /></el-icon> 邀请用户
            </el-button>
          </div>
          <el-table :data="members" v-loading="membersLoading" stripe size="small">
            <el-table-column label="头像" width="60">
              <template #default="{ row }">
                <el-avatar :size="32" :src="row.nickname ? '' : ''">
                  {{ (row.nickname || 'U')[0]?.toUpperCase() }}
                </el-avatar>
              </template>
            </el-table-column>
            <el-table-column prop="nickname" label="昵称" min-width="120" />
            <el-table-column label="角色" width="120">
              <template #default="{ row }">
                <el-tag size="small" :type="row.admin ? 'warning' : 'info'">
                  {{ row.admin ? '管理员' : '普通成员' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="160">
              <template #default="{ row }">
                <el-button link size="small" @click="toggleMemberAdmin(row)">
                  {{ row.admin ? '取消管理员' : '设为管理员' }}
                </el-button>
                <el-button link size="small" type="danger" @click="handleRemoveMember(row)">移除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 房间 -->
        <el-tab-pane label="房间" name="rooms">
          <div class="tab-toolbar">
            <el-button type="primary" size="small" @click="openCreateRoom">
              <el-icon><Plus /></el-icon> 新建房间
            </el-button>
          </div>
          <el-table :data="rooms" v-loading="roomsLoading" stripe size="small">
            <el-table-column prop="name" label="房间名称" min-width="160" />
            <el-table-column prop="sort" label="排序" width="80" />
            <el-table-column label="操作" width="160">
              <template #default="{ row }">
                <el-button link size="small" @click="openEditRoom(row)">编辑</el-button>
                <el-button link size="small" type="danger" @click="handleDeleteRoom(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>

    <!-- 房间创建/编辑弹窗 -->
    <el-dialog v-model="roomFormVisible" :title="currentRoom?.id ? '编辑房间' : '新建房间'" width="400px">
      <el-form :model="roomForm" label-width="80px">
        <el-form-item label="房间名称" required>
          <el-input v-model="roomForm.name" placeholder="如：客厅、主卧" maxlength="24" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number controls-position="right" v-model="roomForm.sort" :min="0" :max="999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roomFormVisible = false">取消</el-button>
        <el-button type="primary" @click="handleRoomSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 邀请链接弹窗 -->
    <el-dialog v-model="inviteVisible" title="邀请成员" width="460px">
      <div v-if="inviteInfo">
        <el-alert type="info" :closable="false" class="invite-tip">
          邀请码有效期 5 分钟，请尽快发送给对方。
        </el-alert>
        <el-form label-width="80px">
          <el-form-item label="邀请码">
            <div class="copy-row">
              <el-input :model-value="inviteInfo.invite_key" readonly />
              <el-button @click="copyText(inviteInfo.invite_key)">复制</el-button>
            </div>
          </el-form-item>
          <el-form-item label="邀请链接">
            <div class="copy-row">
              <el-input :model-value="inviteInfo.url" readonly />
              <el-button @click="copyText(inviteInfo.url)">复制</el-button>
            </div>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button type="primary" @click="inviteVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 邀请用户（按用户名）弹窗 -->
    <el-dialog v-model="inviteUserVisible" title="邀请用户加入" width="400px">
      <el-form label-width="80px">
        <el-form-item label="用户名">
          <el-input v-model="inviteForm.username" placeholder="输入目标用户名" />
        </el-form-item>
        <el-form-item label="管理员">
          <el-switch v-model="inviteForm.admin" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="inviteUserVisible = false">取消</el-button>
        <el-button type="primary" @click="handleInviteUser">邀请</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, HomeFilled } from '@element-plus/icons-vue'
import SmartTable from '@/components/SmartTable.vue'
import {
  getFamilies, createFamily, updateFamily, deleteFamily,
  createInvite, searchMembers, inviteMember, updateMember, removeMember
} from '@/api/family'
import { searchRooms, createRoom, updateRoom, deleteRoom } from '@/api/room'

const tableRef = ref(null)

const formVisible = ref(false)
const current = ref(null)
const form = reactive({ name: '', icon: '' })

// 详情
const detailVisible = ref(false)
const activeTab = ref('members')
const members = ref([])
const membersLoading = ref(false)

// 房间
const rooms = ref([])
const roomsLoading = ref(false)
const roomFormVisible = ref(false)
const currentRoom = ref(null)
const roomForm = reactive({ name: '', sort: 0 })

// 邀请链接
const inviteVisible = ref(false)
const inviteInfo = ref(null)

// 按用户名邀请
const inviteUserVisible = ref(false)
const inviteForm = reactive({ username: '', admin: false })

const loadFamilies = ({ page, page_size, keyword }) =>
  getFamilies({
    page,
    page_size,
    ...(keyword ? { keyword } : {})
  })

const columns = [
  { prop: 'name', label: '家庭', minWidth: 180, slot: 'family', showOverflowTooltip: false },
  { label: '角色', width: 100, slot: 'role' },
  { prop: 'created', label: '创建时间', width: 180, slot: 'created' }
]

const operators = [
  { key: 'detail', label: '详情', type: 'primary', handler: (row) => openDetail(row) },
  { key: 'invite', label: '邀请', type: 'success', handler: (row) => openInvite(row) },
  { key: 'edit', label: '编辑', type: 'primary', handler: (row) => openEdit(row) },
  { key: 'delete', label: '删除', type: 'danger', handler: (row) => handleDelete(row) }
]

const batchDeleteFamilies = (ids) =>
  Promise.all(ids.map((id) => deleteFamily(id).catch(() => null)))

const isOwner = (fam) => fam.user_id === currentUserId.value

const currentUserId = computed(() => {
  // 从 store 读取当前用户 ID
  return ''
})

const formatDate = (s) => {
  if (!s) return ''
  const d = new Date(s)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const resetForm = () => {
  form.name = ''
  form.icon = ''
}

const openCreate = () => {
  current.value = null
  resetForm()
  formVisible.value = true
}

const openEdit = (row) => {
  current.value = row
  form.name = row.name
  form.icon = row.icon || ''
  formVisible.value = true
}

const handleSubmit = async () => {
  if (!form.name.trim()) {
    ElMessage.warning('请输入家庭名称')
    return
  }
  try {
    if (current.value?.id) {
      await updateFamily(current.value.id, form)
      ElMessage.success('更新成功')
    } else {
      await createFamily({ name: form.name, icon: form.icon })
      ElMessage.success('创建成功')
    }
    formVisible.value = false
    tableRef.value?.refresh()
  } catch (e) {
    // 拦截器已提示
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除家庭"${row.name}"吗？将同时删除关联的成员、房间和设备绑定。`, '提示', { type: 'warning' })
    await deleteFamily(row.id)
    ElMessage.success('删除成功')
    tableRef.value?.refresh()
  } catch (e) {
    // 取消
  }
}

// ========== 详情 ==========
const openDetail = async (row) => {
  current.value = row
  activeTab.value = 'members'
  detailVisible.value = true
  await loadMembers()
}

const loadMembers = async () => {
  if (!current.value?.id) return
  membersLoading.value = true
  try {
    const res = await searchMembers({
      page: 1,
      page_size: 100,
      filter: [{ field: 'family_id', operator: 'eq', value: current.value.id }]
    })
    members.value = res.data || []
  } finally {
    membersLoading.value = false
  }
}

// tab 切换时懒加载房间
watch(activeTab, async (tab) => {
  if (tab === 'rooms' && current.value?.id && rooms.value.length === 0) {
    await loadRooms()
  }
})

const loadRooms = async () => {
  if (!current.value?.id) return
  roomsLoading.value = true
  try {
    const res = await searchRooms({
      page: 1,
      page_size: 100,
      filter: [{ field: 'family_id', operator: 'eq', value: current.value.id }]
    })
    rooms.value = res.data || []
  } finally {
    roomsLoading.value = false
  }
}

const resetRoomForm = () => {
  roomForm.name = ''
  roomForm.sort = 0
}

const openCreateRoom = () => {
  currentRoom.value = null
  resetRoomForm()
  roomFormVisible.value = true
}

const openEditRoom = (row) => {
  currentRoom.value = row
  roomForm.name = row.name
  roomForm.sort = row.sort || 0
  roomFormVisible.value = true
}

const handleRoomSubmit = async () => {
  if (!roomForm.name.trim()) {
    ElMessage.warning('请输入房间名称')
    return
  }
  try {
    if (currentRoom.value?.id) {
      await updateRoom(currentRoom.value.id, { name: roomForm.name, sort: roomForm.sort, family_id: current.value.id })
      ElMessage.success('更新成功')
    } else {
      await createRoom({ name: roomForm.name, sort: roomForm.sort, family_id: current.value.id })
      ElMessage.success('创建成功')
    }
    roomFormVisible.value = false
    await loadRooms()
  } catch (e) {
    // 拦截器已提示
  }
}

const handleDeleteRoom = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除房间"${row.name}"吗？`, '提示', { type: 'warning' })
    await deleteRoom(row.id)
    ElMessage.success('已删除')
    await loadRooms()
  } catch (e) {
    // 取消
  }
}

const toggleMemberAdmin = async (row) => {
  try {
    await updateMember(row.id, !row.admin)
    row.admin = !row.admin
    ElMessage.success('更新成功')
  } catch (e) {
    // 拦截器已提示
  }
}

const handleRemoveMember = async (row) => {
  try {
    await ElMessageBox.confirm(`确定移除成员"${row.nickname || row.user_id}"吗？`, '提示', { type: 'warning' })
    await removeMember(row.id)
    ElMessage.success('已移除')
    await loadMembers()
  } catch (e) {
    // 取消
  }
}

// ========== 邀请 ==========
const openInvite = async (row) => {
  try {
    const res = await createInvite(row.id)
    inviteInfo.value = res.data
    inviteVisible.value = true
  } catch (e) {
    // 拦截器已提示
  }
}

const copyText = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制到剪贴板')
  } catch (e) {
    ElMessage.warning('复制失败，请手动复制')
  }
}

const openInviteUser = () => {
  inviteForm.username = ''
  inviteForm.admin = false
  inviteUserVisible.value = true
}

const handleInviteUser = async () => {
  if (!inviteForm.username.trim()) {
    ElMessage.warning('请输入用户名')
    return
  }
  try {
    await inviteMember({
      family_id: current.value.id,
      username: inviteForm.username,
      admin: inviteForm.admin
    })
    ElMessage.success('邀请已发送')
    inviteUserVisible.value = false
    await loadMembers()
  } catch (e) {
    // 拦截器已提示
  }
}
</script>

<style lang="scss" scoped>
.family-list {
  .family-cell {
    display: flex;
    align-items: center;
    gap: 10px;

    .family-icon {
      width: 36px;
      height: 36px;
      border-radius: 8px;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: #fff;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 18px;
      flex-shrink: 0;
    }

    .default-icon {
      font-size: 22px;
      color: #667eea;
      flex-shrink: 0;
    }

    .family-name {
      font-size: 14px;
      font-weight: 500;
      color: #1a1a2e;
    }
  }

  .tab-toolbar {
    margin-bottom: 12px;
  }

  .invite-tip {
    margin-bottom: 16px;
  }

  .copy-row {
    display: flex;
    gap: 8px;
    width: 100%;
  }
}
</style>
