<template>
  <view class="page-mine-avatar">
    <DsNavBar title="更换头像" />

    <view class="avatar-panel">
      <view class="avatar-stage" @click="chooseImage">
        <DsAvatar v-if="previewUrl" :src="previewUrl" size="xlarge" />
        <view v-else class="avatar-placeholder">
          <DsIcon name="my" :size="44" color="#8ea6c1" />
        </view>

        <view class="avatar-camera-badge">
          <DsIcon name="camera" :size="18" color="#fff" />
        </view>
      </view>

      <text class="avatar-title">点击头像重新选择图片</text>
      <text class="avatar-subtitle">建议使用清晰正方形头像，支持 JPG、PNG，小于 5MB。</text>

      <view class="tips-card">
        <view class="tip-item">
          <view class="tip-dot"></view>
          <text class="tip-text">头像会同步展示在个人主页、家庭成员和消息相关页面。</text>
        </view>
        <view class="tip-item">
          <view class="tip-dot"></view>
          <text class="tip-text">上传前会保留当前预览，确认保存后才会提交到服务器。</text>
        </view>
      </view>
    </view>

    <view class="avatar-footer">
      <DsButton text="重新选择" type="ghost" size="large" @click="chooseImage" />
      <DsButton text="保存头像" type="primary" size="large" block :loading="isUploading" :disabled="!canSave" @click="handleSave" />
    </view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getCurrentUser } from '@/api/auth'
import { updateUserInfo } from '@/api/user'
import { currentUser } from '@/utils/user'
import DsNavBar from '@/components/DsNavBar.vue'
import DsAvatar from '@/components/DsAvatar.vue'
import DsIcon from '@/components/DsIcon.vue'
import DsButton from '@/components/DsButton.vue'

const apiBaseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'

const previewUrl = ref('')
const originalAvatarUrl = ref('')
const isUploading = ref(false)
const userId = ref('')

const isTempFile = (value) => {
  if (!value) return false
  return value.startsWith('wxfile://') || value.startsWith('http://tmp') || value.startsWith('file://')
}

const canSave = computed(() => {
  if (!previewUrl.value || isUploading.value) return false
  return previewUrl.value !== originalAvatarUrl.value
})

const chooseImage = () => {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: (res) => {
      const filePath = res.tempFilePaths?.[0]
      if (filePath) {
        previewUrl.value = filePath
      }
    }
  })
}

const syncUserCache = (avatar) => {
  const nextUser = {
    ...(currentUser.value || {}),
    id: userId.value,
    avatar
  }
  currentUser.value = nextUser
  uni.setStorageSync('user', JSON.stringify(nextUser))
}

const handleSave = async () => {
  if (!canSave.value) return

  isUploading.value = true
  try {
    let avatarUrl = previewUrl.value
    if (isTempFile(previewUrl.value)) {
      const token = uni.getStorageSync('token')
      const uploadRes = await new Promise((resolve, reject) => {
        uni.uploadFile({
          url: `${apiBaseURL}/upload/avatar`,
          filePath: previewUrl.value,
          name: 'file',
          header: {
            Authorization: token ? `Bearer ${token}` : ''
          },
          success: (res) => {
            try {
              const data = JSON.parse(res.data)
              if (data.code === 0) {
                resolve(data)
                return
              }
              reject(new Error(data.message || '上传失败'))
            } catch (error) {
              reject(error)
            }
          },
          fail: (error) => reject(error)
        })
      })
      avatarUrl = uploadRes.data?.url || uploadRes.url || ''
    }

    if (!avatarUrl) {
      throw new Error('头像地址无效')
    }

    await updateUserInfo({ avatar: avatarUrl })
    previewUrl.value = avatarUrl
    originalAvatarUrl.value = avatarUrl
    syncUserCache(avatarUrl)

    uni.showToast({ title: '头像已更新', icon: 'success' })
    setTimeout(() => {
      uni.navigateBack()
    }, 600)
  } catch (error) {
    console.error('上传头像失败:', error)
    uni.showToast({ title: error.message || '保存失败', icon: 'none' })
  } finally {
    isUploading.value = false
  }
}

const loadUserInfo = async () => {
  try {
    const res = await getCurrentUser()
    const data = res.data || res || {}
    userId.value = data.id || ''
    previewUrl.value = data.avatar || ''
    originalAvatarUrl.value = data.avatar || ''
  } catch (error) {
    console.error('获取用户信息失败:', error)
  }
}

onShow(() => {
  loadUserInfo()
})
</script>

<style lang="scss">
@import '@/components/variables.scss';

page {
  background: linear-gradient(180deg, #f4f8fc 0%, #eef3f8 100%);
}

.avatar-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #f4f8fc 0%, #eef3f8 100%);
  padding-bottom: 220rpx;
}

.avatar-panel {
  margin: $spacing-base;
  padding: 72rpx 48rpx;
  border-radius: $radius-xl;
  background: $bg-primary;
  box-shadow: 0 12rpx 32rpx rgba(15, 23, 42, 0.06);
  display: flex;
  flex-direction: column;
  align-items: center;
}

.avatar-stage {
  position: relative;
  padding: 20rpx;
  border-radius: 50%;
  background: linear-gradient(135deg, rgba(91, 163, 232, 0.16) 0%, rgba(91, 163, 232, 0.04) 100%);
  margin-bottom: 32rpx;
}

.avatar-placeholder {
  width: 160rpx;
  height: 160rpx;
  border-radius: 50%;
  background: #e9f1f8;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-camera-badge {
  position: absolute;
  right: 14rpx;
  bottom: 14rpx;
  width: 56rpx;
  height: 56rpx;
  border-radius: 50%;
  background: $primary-color;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 6rpx 16rpx rgba(91, 163, 232, 0.3);
}

.avatar-title {
  font-size: 36rpx;
  font-weight: 700;
  color: $text-primary;
  margin-bottom: 12rpx;
}

.avatar-subtitle {
  font-size: $font-size-sm;
  color: $text-secondary;
  line-height: 1.7;
  text-align: center;
  margin-bottom: 40rpx;
}

.tips-card {
  width: 100%;
  padding: 28rpx 32rpx;
  border-radius: $radius-lg;
  background: #f7fafc;
}

.tip-item {
  display: flex;
  align-items: flex-start;
  gap: 16rpx;

  & + .tip-item {
    margin-top: 20rpx;
  }
}

.tip-dot {
  width: 12rpx;
  height: 12rpx;
  border-radius: 50%;
  background: $primary-color;
  margin-top: 12rpx;
  flex-shrink: 0;
}

.tip-text {
  flex: 1;
  font-size: $font-size-sm;
  color: $text-secondary;
  line-height: 1.7;
}

.avatar-footer {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: $spacing-base;
  padding-bottom: calc(#{$spacing-base} + env(safe-area-inset-bottom));
  background: linear-gradient(180deg, rgba(244, 248, 252, 0) 0%, rgba(244, 248, 252, 0.96) 24%, #f4f8fc 100%);
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
}
</style>
