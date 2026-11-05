<template>
  <view class="page-mine-index">
    <DsNavBar title="我的" :showBack="false">
      <template #left>
        <DsIcon name="scan" :size="24" color="#5BA3E8" @click="handleScanCode" />
      </template>
    </DsNavBar>

    <DsCard
      customClass="profile-card"
      customStyle="margin-top: -16rpx;"
    >
      <view class="profile-edit" @click="goToEditInfo">
        <DsIcon name="edit" :size="18" color="#fff" />
      </view>

      <view class="profile-main">
        <view class="profile-avatar" @click="goToAvatar">
          <DsAvatar v-if="userInfo.avatar" :src="userInfo.avatar" size="large" />
          <view v-else class="profile-avatar-fallback">
            <DsIcon name="my" :size="28" color="#fff" />
          </view>
        </view>

        <view class="profile-info">
          <view class="profile-title-row">
            <text class="profile-name">{{ userInfo.nickname || userInfo.username || '用户' }}</text>
          </view>
          <text class="profile-subtitle">{{ userInfo.username || '未设置账号' }}</text>
          <text class="profile-desc">当前家庭：{{ familyName }}</text>
        </view>
      </view>
    </DsCard>

	<DsCard>
    <DsList>
      <DsListItem v-if="isAdmin" title="管理所有设备" icon="server" arrow @click="goToAdminDevices" />
      <DsListItem title="维保工单" icon="order" arrow @click="goToMaintenance" />
      <DsListItem title="我的家庭" :extra="familyName" icon="home" arrow @click="goToFamily" />
      <DsListItem title="修改密码" icon="key" arrow @click="goToChangePassword" />
      <DsListItem title="消息通知" icon="conversation" arrow @click="goToMessage" />
      <DsListItem title="关于我们" icon="text" arrow @click="goToAbout" />
      <DsListItem title="退出登录" icon="quit" arrow @click="handleLogout" />
    </DsList>
	</DsCard>
	
  </view>
</template>

<script setup>
import { reactive, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { currentFamily, currentUser } from '@/utils/user'
import { scanCode } from '@/utils/scan'
import { getSystemName } from '@/utils/oem'
import DsAvatar from '@/components/DsAvatar.vue'
import DsNavBar from '@/components/DsNavBar.vue'
import DsIcon from '@/components/DsIcon.vue'
import DsCard from '@/components/DsCard.vue'
import DsList from '@/components/DsList.vue'
import DsListItem from '@/components/DsListItem.vue'
const userInfo = reactive({
  id: '',
  username: '',
  nickname: '',
  avatar: ''
})
const familyName = computed(() => currentFamily.value?.name || '未选择')

const isAdmin = computed(() => {
  const u = currentUser.value
  if (!u) return false
  return u.admin === true || u.admin === 1
})

const goToAdminDevices = () => {
  uni.navigateTo({ url: '/pages/device/admin-list' })
}

const goToMaintenance = () => {
  uni.navigateTo({ url: '/pages/maintenance/index' })
}

const goToFamily = () => {
  uni.navigateTo({ url: '/pages/family/index' })
}

const handleScanCode = () => {
  scanCode()
}

const goToEditInfo = () => {
  uni.navigateTo({ url: '/pages/mine/edit-info' })
}

const goToAvatar = () => {
  uni.navigateTo({ url: '/pages/mine/avatar' })
}

const goToChangePassword = () => {
  uni.navigateTo({ url: '/pages/mine/change-password' })
}

const goToMessage = () => {
  uni.navigateTo({ url: '/pages/message/index' })
}

const goToAbout = () => {
	uni.showModal({ title: `关于${getSystemName()}`, content: '版本 1.0.0\n智能物联 · 万物互联', showCancel: false })
}

const handleLogout = () => {
  uni.showModal({
    title: '确认退出',
    content: '确定要退出登录吗？',
    success: (res) => {
      if (res.confirm) {
        uni.removeStorageSync('token')
        uni.removeStorageSync('user')
        uni.removeStorageSync('current_family_id')
        uni.reLaunch({ url: '/pages/login/index' })
      }
    }
  })
}

onShow(() => {
  //TODO 调用 api/me，获取最新信息
  if (currentUser.value) {
    userInfo.id = currentUser.value.id || ''
    userInfo.username = currentUser.value.username || ''
    userInfo.nickname = currentUser.value.nickname || ''
    userInfo.avatar = currentUser.value.avatar || ''
  }
})
</script>

<style lang="scss">
@import '@/components/variables.scss';

page {
  background: $bg-secondary;
}

.page-mine-index {
  padding-bottom: $spacing-xl;
}

.profile-card {
  position: relative;
  background: linear-gradient(135deg, $primary-color 0%, $primary-dark 100%);
  color: #fff;
  margin-top: -$spacing-sm;
  padding: $spacing-lg;
  border-radius: $radius-lg;
}

.profile-edit {
  position: absolute;
  top: $spacing-base;
  right: $spacing-base;
  width: 56rpx;
  height: 56rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.16);
  display: flex;
  align-items: center;
  justify-content: center;
}

.profile-main {
  display: flex;
  align-items: center;
  gap: $spacing-base;
}

.profile-avatar {
  flex-shrink: 0;
}

.profile-avatar-fallback {
  width: 96rpx;
  height: 96rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.profile-info {
  flex: 1;
  min-width: 0;
}

.profile-title-row {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  margin-bottom: 8rpx;
}

.profile-name {
  font-size: 38rpx;
  font-weight: 700;
  color: #fff;
}

.profile-subtitle {
  display: block;
  font-size: $font-size-sm;
  color: rgba(255, 255, 255, 0.8);
  margin-bottom: 8rpx;
}

.profile-desc {
  display: block;
  font-size: $font-size-sm;
  color: rgba(255, 255, 255, 0.9);
  line-height: 1.6;
  word-break: break-all;
}

.profile-card .tag {
  background: rgba(255, 255, 255, 0.18);
  color: #fff;
}

.profile-card .button-ghost {
  border-color: rgba(255, 255, 255, 0.4);
  color: #fff;
}

.profile-card .button-ghost .button-text {
  color: #fff;
}

.profile-card .button-ghost .iconfont {
  color: #fff !important;
}
</style>
