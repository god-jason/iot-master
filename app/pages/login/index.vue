<template>
  <view class="page-login-index">
    <view class="login-header">
      <image class="logo" src="/static/logo.png" mode="aspectFill" />
      <text class="title">{{ systemName }}</text>
      <text class="subtitle">智能物联 · 万物互联</text>
    </view>

    <view class="login-form">
      <DsTabs v-model="activeTab" type="card">
        <DsTabItem label="账号密码" :index="0" :active="activeTab === 0" @click="activeTab = 0" />
        <DsTabItem label="微信登录" :index="1" :active="activeTab === 1" @click="activeTab = 1" />
      </DsTabs>

      <view v-if="activeTab === 0" class="form-content">
        <DsFormItem label="">
          <DsInput v-model="form.username" placeholder="请输入用户名/邮箱" prefixIcon="my" />
        </DsFormItem>
        <DsFormItem label="">
          <DsInput v-model="form.password" type="password" placeholder="请输入密码" prefixIcon="key" />
        </DsFormItem>
        <DsButton text="登 录" type="primary" size="large" :loading="loading" :disabled="loading" @click="handleLogin" />
      </view>

      <view v-else class="wechat-content">
        <DsButton text="微信授权登录" bgColor="#07C160" color="#fff" size="large" @click="handleWechatLogin" />
      </view>

      <view class="form-options">
        <view class="checkbox" @click="acceptLicense = !acceptLicense">
          <view :class="['checkbox-icon', { checked: acceptLicense }]">
            <DsIcon v-if="acceptLicense" name="affirm" :size="14" color="#fff" />
          </view>
          <text class="checkbox-text" @click.stop="goToLicense">我已经阅读并接受用户使用协议</text>
        </view>
        <text class="forgot-link" @click="forgotPassword">忘记密码？</text>
      </view>
    </view>

    <DsDialog v-model:visible="showProfileModal" title="完善资料" :showCancel="false" @confirm="submitProfile">
      <view class="profile-body">
        <view class="profile-item">
          <text class="profile-label">头像</text>
          <button 
            class="avatar-btn" 
            open-type="chooseAvatar" 
            @chooseavatar="onChooseAvatar"
          >
            <DsAvatar v-if="profileForm.avatar" :src="profileForm.avatar" size="large" />
            <DsAvatar v-else icon="my" size="large" bgColor="#f5f5f5" />
          </button>
        </view>
        <DsFormItem label="昵称">
          <DsInput v-model="profileForm.nickname" placeholder="请输入昵称" />
        </DsFormItem>
        <DsFormItem label="手机号">
          <DsInput v-model="profileForm.phone" placeholder="请输入手机号" maxlength="11" />
        </DsFormItem>
        <DsButton
          text="授权微信绑定的手机号"
          bgColor="#07c160"
          color="#fff"
          size="medium"
          open-type="getPhoneNumber"
          @getphonenumber="onGetPhoneNumber"
        />
      </view>
    </DsDialog>
  </view>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { login, wechatCode2Session, wechatAuth, getPhoneNumber } from '@/api/auth'
import { initUserData } from '@/utils/user'
import { getSystemName } from '@/utils/oem'
import DsNavBar from '@/components/DsNavBar.vue'
import DsTabs from '@/components/DsTabs.vue'
import DsTabItem from '@/components/DsTabItem.vue'
import DsFormItem from '@/components/DsFormItem.vue'
import DsInput from '@/components/DsInput.vue'
import DsButton from '@/components/DsButton.vue'
import DsIcon from '@/components/DsIcon.vue'
import DsDialog from '@/components/DsDialog.vue'
import DsAvatar from '@/components/DsAvatar.vue'
const systemName = getSystemName()
const activeTab = ref(0)
const acceptLicense = ref(false)
const loading = ref(false)
const showProfileModal = ref(false)
const isNavigating = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const profileForm = reactive({
  openid: '',
  nickname: '',
  avatar: '',
  phone: ''
})

const handleLogin = async () => {
  if (!form.username || !form.password) {
    uni.showToast({ title: '请输入用户名和密码', icon: 'none' })
    return
  }
  if (!acceptLicense.value) {
    uni.showToast({ title: '请先接受使用协议', icon: 'none' })
    return
  }
  loading.value = true
  try {
    const res = await login(form)
    const token = res.data?.token || res.token
    if (token) {
      uni.setStorageSync('token', token)
      if (res.data?.user) {
        uni.setStorageSync('user', JSON.stringify(res.data.user))
      }
      await initUserData()
      if (isNavigating.value) return
      isNavigating.value = true
      uni.showToast({ title: '登录成功', icon: 'success' })
      setTimeout(() => {
        uni.switchTab({
          url: '/pages/device/index',
          complete: () => {
            isNavigating.value = false
          }
        })
      }, 300)
    }
  } catch (error) {
    console.error(error)
    uni.showToast({ title: error.message || '登录失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

const goToLicense = () => {
  uni.navigateTo({ url: '/pages/login/license' })
}

const forgotPassword = () => {
  uni.showModal({ title: '忘记密码', content: '请联系杰神 15161515197', showCancel: false, confirmText: '知道了' })
}

const handleWechatLogin = async () => {
  if (!acceptLicense.value) {
    uni.showToast({ title: '请先接受使用协议', icon: 'none' })
    return
  }
  if (process.env.UNI_PLATFORM === 'mp-weixin') {
    loading.value = true
    try {
      const res = await uni.login({ provider: 'weixin' })
      const wxRes = await wechatCode2Session(res.code)
      const data = wxRes.data || wxRes
      if (data.token) {
        uni.setStorageSync('token', data.token)
        if (data.user) {
          uni.setStorageSync('user', JSON.stringify(data.user))
        }
        await initUserData()
        if (isNavigating.value) return
        isNavigating.value = true
        uni.switchTab({
          url: '/pages/device/index',
          complete: () => {
            isNavigating.value = false
          }
        })
      } else if (data.need_auth) {
        profileForm.openid = data.openid
        showProfileModal.value = true
      }
    } catch (error) {
      console.error(error)
      uni.showToast({ title: error.message || '微信授权失败', icon: 'none' })
    } finally {
      loading.value = false
    }
  } else {
    uni.showToast({ title: '仅微信小程序支持微信登录', icon: 'none' })
  }
}

const onChooseAvatar = (e) => {
  profileForm.avatar = e.detail.avatarUrl
}

const onGetPhoneNumber = async (e) => {
  if (e.detail.errMsg && e.detail.errMsg !== 'getPhoneNumber:ok') {
    uni.showToast({ title: '已取消授权', icon: 'none' })
    return
  }
  if (!e.detail.code) {
    uni.showToast({ title: '获取手机号失败', icon: 'none' })
    return
  }
  try {
    const res = await getPhoneNumber(e.detail.code)
    const phone = res.data?.phone || res.phone
    if (phone) {
      profileForm.phone = phone
      uni.showToast({ title: '手机号已获取', icon: 'success' })
    } else {
      uni.showToast({ title: '未获取到手机号', icon: 'none' })
    }
  } catch (error) {
    console.error(error)
    uni.showToast({ title: error.message || '获取手机号失败', icon: 'none' })
  }
}

const submitProfile = async () => {
  if (!profileForm.nickname.trim()) {
    uni.showToast({ title: '请输入昵称', icon: 'none' })
    return
  }
  loading.value = true
  try {
    let avatarUrl = profileForm.avatar
    if (profileForm.avatar && (profileForm.avatar.startsWith('wxfile://') || profileForm.avatar.startsWith('http://tmp'))) {
      const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'
      const uploadRes = await new Promise((resolve, reject) => {
        uni.uploadFile({
          url: `${baseURL}/weixin/avatar`,
          filePath: profileForm.avatar,
          name: 'file',
          success: (res) => {
            const data = JSON.parse(res.data)
            if (data.code === 0) {
              resolve(data)
            } else {
              reject(new Error(data.message || '头像上传失败'))
            }
          },
          fail: (err) => reject(err)
        })
      })
      avatarUrl = uploadRes.data.url
    }
    const wxRes = await wechatAuth({
      openid: profileForm.openid,
      nickname: profileForm.nickname,
      avatar: avatarUrl,
      phone: profileForm.phone
    })
    const token = wxRes.data?.token || wxRes.token
    const user = wxRes.data?.user || wxRes.user
    if (token) {
      uni.setStorageSync('token', token)
      if (user) {
        uni.setStorageSync('user', JSON.stringify(user))
      }
      showProfileModal.value = false
      await initUserData()
      if (isNavigating.value) return
      isNavigating.value = true
      uni.switchTab({
        url: '/pages/device/index',
        complete: () => {
          isNavigating.value = false
        }
      })
    }
  } catch (error) {
    console.error(error)
    uni.showToast({ title: error.message || '提交失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss">
@import '@/components/variables.scss';

.page-login-index {
  page {
    background: linear-gradient(135deg, #5BA3E8 0%, #1E3A5F 100%);
    min-height: 100vh;
  }

  & {
    display: flex;
    flex-direction: column;
    align-items: center;
    min-height: 100vh;
    background: linear-gradient(135deg, #5BA3E8 0%, #1E3A5F 100%);
    box-sizing: border-box;
  }

  .login-header {
    display: flex;
    flex-direction: column;
    align-items: center;
    margin-bottom: 80rpx;
    margin-top: 80rpx;

    .logo {
      width: 200rpx;
      height: 200rpx;
      margin-bottom: 30rpx;
      border-radius: 100%;
      background: rgba(255, 255, 255, 0.2);
      overflow: hidden;
    }

    .title {
      font-size: 48rpx;
      font-weight: 700;
      color: #fff;
      margin-bottom: 10rpx;
    }

    .subtitle {
      font-size: 26rpx;
      color: rgba(255, 255, 255, 0.8);
    }
  }

  .login-form {
    width: 100%;
    max-width: 600rpx;
    background: #fff;
    border-radius: 24rpx;
    padding: 40rpx;
    box-shadow: 0 20rpx 60rpx rgba(0, 0, 0, 0.15);
  }

  .form-content {
    display: flex;
    flex-direction: column;
    gap: 30rpx;
    margin-top: 30rpx;
  }

  .form-options {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 30rpx;
  }

  .checkbox {
    display: flex;
    align-items: center;
    gap: 12rpx;
  }

  .checkbox-icon {
    width: 36rpx;
    height: 36rpx;
    border: 2rpx solid #ddd;
    border-radius: 6rpx;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 24rpx;
    color: #fff;
    transition: all 0.2s;

    &.checked {
      background: #5BA3E8;
      border-color: #5BA3E8;
    }
  }

  .checkbox-text {
    font-size: 26rpx;
    color: #6B7280;
  }

  .forgot-link {
    font-size: 26rpx;
    color: #5BA3E8;
  }

  .wechat-content {
    padding: 40rpx 0;
  }

  .profile-body {
    padding: 20rpx 0;

    .profile-item {
      display: flex;
      align-items: center;
      gap: 30rpx;
      margin-bottom: 30rpx;
    }

    .profile-label {
      font-size: 30rpx;
      color: #666;
      width: 120rpx;
    }

    .avatar-btn {
      width: 120rpx;
      height: 120rpx;
      border-radius: 50%;
      border: none;
      background: transparent;
      padding: 0;
      margin: 0;
    }
  }
}
</style>
