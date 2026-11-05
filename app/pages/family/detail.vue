<template>
  <view class="page-family-detail">
    <DsNavBar :title="family.name || '家庭详情'" />

    <view v-if="isLoading && !family.id" class="page-loading">
      <DsLoading text="加载家庭详情中..." />
    </view>

    <view v-else class="family-detail-content">
      <DsCard v-if="family.id" customClass="family-header">
        <view class="family-icon">
          <DsIcon v-if="family.icon" :name="family.icon" :size="24" color="#fff" />
          <DsIcon v-else name="home" :size="24" color="#fff" />
        </view>
        <view class="family-info">
          <text class="family-name">{{ family.name || '未命名家庭' }}</text>
          <text class="family-desc">共 {{ memberList.length }} 位成员，{{ roomList.length }} 个房间</text>
        </view>
        <DsButton
          v-if="canManageFamily"
          text="编辑家庭"
          icon="edit"
          type="ghost"
          size="small"
          @click="openEditFamilyDialog"
        />
      </DsCard>

      <DsCard>
        <view class="panel-header">
          <view>
            <text class="panel-title">房间管理</text>
            <text class="panel-subtitle">管理当前家庭的房间顺序与名称</text>
          </view>
          <DsButton
            v-if="canManageFamily"
            text="添加房间"
            icon="add-circle"
            type="ghost"
            size="small"
            @click="openCreateRoomDialog"
          />
        </view>

        <DsEmpty v-if="roomList.length === 0" text="暂无房间" subText="添加房间后可按房间管理设备" />

        <view v-else class="room-list">
          <view v-for="(room, index) in roomList" :key="room.id" class="room-row">
            <view v-if="canManageFamily" class="room-order">
              <DsButton :disabled="index === 0" text="上移" type="default" size="mini" @click.stop="moveRoom(index, -1)" />
              <DsButton :disabled="index === roomList.length - 1" text="下移" type="default" size="mini" @click.stop="moveRoom(index, 1)" />
            </view>

            <view class="room-row-icon">
              <text>{{ getRoomInitial(room.name) }}</text>
            </view>

            <view class="room-main">
              <text class="room-row-name">{{ room.name || '未命名房间' }}</text>
              <text class="room-row-desc">排序 {{ index + 1 }}</text>
            </view>

            <view v-if="canManageFamily" class="room-actions">
              <DsButton icon="edit" type="default" size="mini" @click.stop="editRoom(room)" />
              <DsButton icon="trash-bin" type="danger" size="mini" @click.stop="removeRoom(room)" />
            </view>
          </view>
        </view>
      </DsCard>

      <DsCard>
        <view class="panel-header">
          <view>
            <text class="panel-title">家庭成员</text>
            <text class="panel-subtitle">邀请成员加入，按角色管理访问权限</text>
          </view>
          <DsButton
            v-if="canManageFamily"
            text="邀请成员"
            icon="add"
            type="ghost"
            size="small"
            @click="goToInvite"
          />
        </view>

        <DsEmpty v-if="memberList.length === 0" text="暂无成员" subText="邀请成员后可以一起管理家庭设备" />

        <view v-else class="member-list">
          <view v-for="member in memberList" :key="member.id" class="member-item">
            <DsAvatar v-if="member.avatar" :src="member.avatar" size="small" />
            <view v-else class="member-avatar">
              <DsIcon name="my" :size="24" color="#fff" />
            </view>

            <view class="member-info">
              <text class="member-name">{{ member.nickname || member.username || member.user_id || '未知成员' }}</text>
              <text class="member-role">{{ getRoleText(member) }}</text>
            </view>

            <view
              v-if="canManageFamily && member.user_id !== currentUserId && !(family.user_id && family.user_id === member.user_id)"
              class="member-actions"
            >
              <DsButton
                :text="member.admin ? '取消管理员' : '设为管理员'"
                type="default"
                size="mini"
                @click="toggleAdmin(member)"
              />
              <DsButton text="移除" type="danger" size="mini" @click="removeMember(member)" />
            </view>
          </view>
        </view>
      </DsCard>

      <DsCard v-if="currentUserRole === 'owner'">
        <view class="danger-zone">
          <text class="danger-title">危险操作</text>
          <text class="danger-desc">删除家庭后，成员关系和房间数据将一并清除，请谨慎操作。</text>
          <DsButton text="删除家庭" type="danger" size="large" block @click="handleDeleteFamily" />
        </view>
      </DsCard>
    </view>

    <DsDialog v-model:visible="showEditDialog" title="编辑家庭" @confirm="handleEditFamily" @cancel="closeEditDialog">
      <DsFormItem label="家庭名称">
        <DsInput v-model="editForm.name" placeholder="请输入家庭名称" />
      </DsFormItem>
    </DsDialog>

    <DsDialog
      v-model:visible="showRoomDialog"
      :title="editingRoom ? '编辑房间' : '添加房间'"
      @confirm="handleSaveRoom"
      @cancel="closeRoomDialog"
    >
      <DsFormItem label="房间名称">
        <DsInput v-model="roomForm.name" placeholder="请输入房间名称" />
      </DsFormItem>
    </DsDialog>
  </view>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import {
  createRoom,
  deleteFamily,
  deleteRoom,
  getFamilyDetail,
  getMemberList,
  getRoomList,
  removeMember as apiRemoveMember,
  updateFamily,
  updateMemberRole,
  updateRoom
} from '@/api/family'
import { currentFamily, getUserInfo, loadFamilyData, setCurrentFamily } from '@/utils/user'
import DsAvatar from '@/components/DsAvatar.vue'
import DsButton from '@/components/DsButton.vue'
import DsCard from '@/components/DsCard.vue'
import DsDialog from '@/components/DsDialog.vue'
import DsEmpty from '@/components/DsEmpty.vue'
import DsFormItem from '@/components/DsFormItem.vue'
import DsIcon from '@/components/DsIcon.vue'
import DsInput from '@/components/DsInput.vue'
import DsLoading from '@/components/DsLoading.vue'
import DsNavBar from '@/components/DsNavBar.vue'

const familyId = ref('')
const isLoading = ref(false)

const family = reactive({
  id: '',
  name: '',
  icon: '',
  user_id: ''
})

const roomList = ref([])
const memberList = ref([])

const showEditDialog = ref(false)
const editForm = reactive({ name: '' })

const showRoomDialog = ref(false)
const editingRoom = ref(null)
const roomForm = reactive({ name: '' })

const currentUserId = computed(() => getUserInfo()?.id || '')

const currentUserMember = computed(() => {
  if (!currentUserId.value) return null
  return memberList.value.find((item) => item.user_id === currentUserId.value) || null
})

const canManageFamily = computed(() => {
  if (!currentUserId.value) return false
  if (family.user_id && family.user_id === currentUserId.value) return true
  return !!currentUserMember.value?.admin
})

const currentUserRole = computed(() => {
  if (!currentUserId.value) return ''
  if (family.user_id && family.user_id === currentUserId.value) return 'owner'
  return currentUserMember.value?.admin ? 'admin' : 'member'
})

const normalizeRooms = (rooms) => {
  return [...(rooms || [])].sort((left, right) => {
    const leftSort = Number(left?.sort || 0)
    const rightSort = Number(right?.sort || 0)
    if (leftSort !== rightSort) return leftSort - rightSort
    return String(left?.name || '').localeCompare(String(right?.name || ''))
  })
}

const syncCurrentFamily = () => {
  if (currentFamily.value?.id === family.id) {
    setCurrentFamily({
      ...currentFamily.value,
      ...family
    })
  }
}

const getRoleText = (member) => {
  if (family.user_id && family.user_id === member.user_id) return '所有者'
  if (member.admin) return '管理员'
  return '成员'
}

const getRoomInitial = (roomName) => {
  return String(roomName || '房').slice(0, 1)
}

const closeEditDialog = () => {
  showEditDialog.value = false
  editForm.name = family.name || ''
}

const openEditFamilyDialog = () => {
  editForm.name = family.name || ''
  showEditDialog.value = true
}

const resetRoomForm = () => {
  editingRoom.value = null
  roomForm.name = ''
}

const closeRoomDialog = () => {
  showRoomDialog.value = false
  resetRoomForm()
}

const openCreateRoomDialog = () => {
  resetRoomForm()
  showRoomDialog.value = true
}

const loadData = async () => {
  if (!familyId.value) return

  isLoading.value = true
  try {
    const [familyRes, roomRes, memberRes] = await Promise.all([
      getFamilyDetail(familyId.value),
      getRoomList(familyId.value),
      getMemberList(familyId.value)
    ])

    Object.assign(family, familyRes.data || {})
    roomList.value = normalizeRooms(roomRes.data || [])
    memberList.value = memberRes.data || []
    syncCurrentFamily()
  } catch (error) {
    console.error('加载家庭详情失败:', error)
    uni.showToast({ title: error.message || '加载家庭详情失败', icon: 'none' })
  } finally {
    isLoading.value = false
    uni.stopPullDownRefresh()
  }
}

const goToInvite = () => {
  if (!familyId.value) return
  uni.navigateTo({
    url: `/pages/family/invite?id=${familyId.value}`
  })
}

const handleEditFamily = async () => {
  const name = editForm.name.trim()
  if (!name) {
    uni.showToast({ title: '请输入家庭名称', icon: 'none' })
    return
  }

  try {
    await updateFamily(familyId.value, { name })
    family.name = name
    syncCurrentFamily()
    showEditDialog.value = false
    uni.showToast({ title: '修改成功', icon: 'success' })
  } catch (error) {
    console.error('修改家庭失败:', error)
    uni.showToast({ title: error.message || '修改失败', icon: 'none' })
  }
}

const moveRoom = async (index, direction) => {
  const newIndex = index + direction
  if (newIndex < 0 || newIndex >= roomList.value.length) return

  const rooms = [...roomList.value]
  const current = rooms[index]
  rooms[index] = rooms[newIndex]
  rooms[newIndex] = current
  roomList.value = rooms

  try {
    await Promise.all(
      rooms.map((room, sortIndex) => updateRoom(room.id, { sort: sortIndex + 1 }))
    )
    roomList.value = rooms.map((room, sortIndex) => ({
      ...room,
      sort: sortIndex + 1
    }))
  } catch (error) {
    console.error('房间排序失败:', error)
    uni.showToast({ title: error.message || '排序失败', icon: 'none' })
    await loadData()
  }
}

const editRoom = (room) => {
  editingRoom.value = room
  roomForm.name = room.name || ''
  showRoomDialog.value = true
}

const handleSaveRoom = async () => {
  const name = roomForm.name.trim()
  if (!name) {
    uni.showToast({ title: '请输入房间名称', icon: 'none' })
    return
  }

  try {
    if (editingRoom.value) {
      await updateRoom(editingRoom.value.id, { name })
    } else {
      await createRoom({
        name,
        family_id: familyId.value,
        sort: roomList.value.length + 1
      })
    }

    closeRoomDialog()
    await loadData()
    uni.showToast({ title: '保存成功', icon: 'success' })
  } catch (error) {
    console.error('保存房间失败:', error)
    uni.showToast({ title: error.message || '保存失败', icon: 'none' })
  }
}

const removeRoom = (room) => {
  uni.showModal({
    title: '确认删除',
    content: `确定要删除房间“${room.name || '未命名房间'}”吗？`,
    success: async (result) => {
      if (!result.confirm) return

      try {
        await deleteRoom(room.id)
        await loadData()
        uni.showToast({ title: '删除成功', icon: 'success' })
      } catch (error) {
        console.error('删除房间失败:', error)
        uni.showToast({ title: error.message || '删除失败', icon: 'none' })
      }
    }
  })
}

const toggleAdmin = (member) => {
  const targetName = member.nickname || member.username || member.user_id || '该成员'
  const action = member.admin ? '取消管理员' : '设为管理员'

  uni.showModal({
    title: '修改权限',
    content: `确定将${targetName}${action}吗？`,
    success: async (result) => {
      if (!result.confirm) return

      try {
        await updateMemberRole(member.id, !member.admin)
        member.admin = !member.admin
        uni.showToast({ title: '修改成功', icon: 'success' })
      } catch (error) {
        console.error('修改权限失败:', error)
        uni.showToast({ title: error.message || '修改失败', icon: 'none' })
      }
    }
  })
}

const removeMember = (member) => {
  const targetName = member.nickname || member.username || member.user_id || '该成员'

  uni.showModal({
    title: '确认移除',
    content: `确定要将${targetName}移出家庭吗？`,
    success: async (result) => {
      if (!result.confirm) return

      try {
        await apiRemoveMember(member.id)
        memberList.value = memberList.value.filter((item) => item.id !== member.id)
        uni.showToast({ title: '移除成功', icon: 'success' })
      } catch (error) {
        console.error('移除成员失败:', error)
        uni.showToast({ title: error.message || '移除失败', icon: 'none' })
      }
    }
  })
}

const handleDeleteFamily = () => {
  uni.showModal({
    title: '确认删除',
    content: '删除家庭后，所有成员和房间数据将被清除，确定要删除吗？',
    success: async (result) => {
      if (!result.confirm) return

      try {
        await deleteFamily(familyId.value)
        await loadFamilyData()
        uni.showToast({ title: '删除成功', icon: 'success' })
        setTimeout(() => {
          uni.redirectTo({ url: '/pages/family/index' })
        }, 300)
      } catch (error) {
        console.error('删除家庭失败:', error)
        uni.showToast({ title: error.message || '删除失败', icon: 'none' })
      }
    }
  })
}

onLoad((options) => {
  familyId.value = options?.id || ''
  if (familyId.value) {
    loadData()
  }
})

onPullDownRefresh(() => {
  loadData()
})
</script>

<style lang="scss">
@import '@/components/variables.scss';

page {
  background: $bg-secondary;
}

.family-detail-container {
  min-height: 100vh;
}

.page-loading {
  padding: $spacing-xl;
}

.family-detail-content {
  padding-bottom: $spacing-xl;
}

.family-header {
  display: flex;
  align-items: center;
  gap: $spacing-lg;
  background: linear-gradient(135deg, $primary-color 0%, $primary-dark 100%);
  padding: $spacing-lg;
  margin: $spacing-base;
  border-radius: 24rpx;

  .family-icon {
    width: 120rpx;
    height: 120rpx;
    border-radius: $radius-lg;
    background: rgba(255, 255, 255, 0.2);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .family-info {
    flex: 1;

    .family-name {
      font-size: 40rpx;
      font-weight: 600;
      color: #fff;
      display: block;
      margin-bottom: 8rpx;
    }

    .family-desc {
      font-size: $font-size-sm;
      color: rgba(255, 255, 255, 0.85);
    }
  }
}

.panel-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: $spacing-base;
  margin-bottom: $spacing-base;

  .panel-title {
    display: block;
    font-size: $font-size-lg;
    font-weight: 600;
    color: $text-primary;
    margin-bottom: 6rpx;
  }

  .panel-subtitle {
    font-size: $font-size-sm;
    color: $text-placeholder;
  }
}

.room-list,
.member-list {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
}

.room-row,
.member-item {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  padding: $spacing-base;
  background: #f8fafc;
  border-radius: $radius-base;
}

.room-order {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.room-row-icon {
  width: 64rpx;
  height: 64rpx;
  border-radius: $radius-base;
  background: linear-gradient(135deg, $primary-color 0%, $primary-dark 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;

  text {
    font-size: $font-size-base;
    font-weight: 600;
    color: #fff;
  }
}

.room-main {
  flex: 1;
  min-width: 0;
}

.room-row-name {
  display: block;
  font-size: $font-size-base;
  color: $primary-dark;
  font-weight: 500;
  margin-bottom: 4rpx;
}

.room-row-desc {
  font-size: $font-size-sm;
  color: $text-secondary;
}

.room-actions,
.member-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
  justify-content: flex-end;
}

.member-avatar {
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  background: $primary-color;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.member-info {
  flex: 1;
  min-width: 0;

  .member-name {
    display: block;
    font-size: $font-size-base;
    font-weight: 500;
    color: $primary-dark;
    margin-bottom: 6rpx;
  }

  .member-role {
    font-size: $font-size-sm;
    color: $text-secondary;
  }
}

.danger-zone {
  .danger-title {
    display: block;
    font-size: $font-size-lg;
    font-weight: 600;
    color: $danger-color;
    margin-bottom: $spacing-xs;
  }

  .danger-desc {
    display: block;
    font-size: $font-size-sm;
    color: $text-secondary;
    line-height: 1.6;
    margin-bottom: $spacing-base;
  }
}
</style>
