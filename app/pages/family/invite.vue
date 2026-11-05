<template>
  <view class="page-family-invite">
    <DsNavBar title="邀请成员" />

    <DsCard customClass="invite-card">
      <view class="card-header">
        <text class="card-title">邀请二维码</text>
        <text class="card-subtitle">邀请码 5 分钟内有效，家人扫码即可加入家庭</text>
      </view>

      <view class="qrcode-section">
        <view class="qrcode-box">
          <DsLoading v-if="isLoading && !inviteContent" text="生成中..." />
          <DsQrCode v-else-if="inviteContent" :value="inviteContent" :size="260" />
          <DsEmpty v-else text="暂无邀请码" subText="请尝试刷新邀请码" />
        </view>
        <text class="qrcode-tip">有效期 5 分钟，过期后请刷新</text>
      </view>

      <view class="info-section">
        <DsDescription>
          <DsDescriptionItem label="邀请码">{{ inviteKey || '-' }}</DsDescriptionItem>
          <DsDescriptionItem label="家庭">{{ familyName || '-' }}</DsDescriptionItem>
        </DsDescription>
      </view>

      <view class="action-section">
        <DsButton text="刷新二维码" icon="refresh" type="ghost" size="medium" @click="refreshInvite" />
        <DsButton text="复制邀请码" icon="copy" type="default" size="medium" @click="copyInviteKey" />
      </view>

      <DsButton text="复制邀请内容并分享" icon="share" type="primary" size="large" block @click="shareInvite" />
    </DsCard>

    <DsCard customClass="tips-card">
      <view class="card-header compact">
        <text class="card-title">使用说明</text>
        <text class="card-subtitle">将二维码或邀请码发给家人，对方可在设备页扫码加入</text>
      </view>

      <view class="tips-list">
        <view v-for="(tip, index) in tips" :key="index" class="tips-item">
          <DsTag :text="String(index + 1)" type="primary" size="small" />
          <text class="tips-text">{{ tip }}</text>
        </view>
      </view>
    </DsCard>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { createFamilyInvite, getFamilyDetail } from '@/api/family'
import DsButton from '@/components/DsButton.vue'
import DsCard from '@/components/DsCard.vue'
import DsDescription from '@/components/DsDescription.vue'
import DsDescriptionItem from '@/components/DsDescriptionItem.vue'
import DsEmpty from '@/components/DsEmpty.vue'
import DsLoading from '@/components/DsLoading.vue'
import DsNavBar from '@/components/DsNavBar.vue'
import DsQrCode from '@/components/DsQrCode.vue'
import DsTag from '@/components/DsTag.vue'

const familyId = ref('')
const familyName = ref('')
const inviteKey = ref('')
const inviteUrl = ref('')
const isLoading = ref(false)

const tips = [
  '请让家人在“设备管理”页面点击左上角扫码按钮。',
  '扫描当前二维码，或输入你分享的邀请码。',
  '加入成功后，家人即可查看并管理该家庭下的设备。'
]

const inviteContent = computed(() => {
  if (inviteKey.value) return `noob-invite:${inviteKey.value}`
  if (inviteUrl.value) return inviteUrl.value
  return ''
})

const shareText = computed(() => {
  if (!inviteKey.value) return ''
  const lines = [
    `邀请你加入家庭：${familyName.value || '我的家庭'}`,
    `邀请码：${inviteKey.value}`,
    `扫码内容：${inviteContent.value}`
  ]

  if (inviteUrl.value) {
    lines.push(`邀请链接：${inviteUrl.value}`)
  }

  return lines.join('\n')
})

const loadFamilyInfo = async () => {
  const res = await getFamilyDetail(familyId.value)
  const family = res.data || res
  familyName.value = family.name || ''
}

const createInvite = async () => {
  const res = await createFamilyInvite(familyId.value)
  const data = res.data || res
  inviteKey.value = data.invite_key || ''
  inviteUrl.value = data.url || ''
}

const loadPageData = async () => {
  if (!familyId.value) return

  isLoading.value = true
  try {
    await Promise.all([loadFamilyInfo(), createInvite()])
  } catch (error) {
    console.error('加载邀请信息失败:', error)
    uni.showToast({ title: error.message || '加载邀请信息失败', icon: 'none' })
  } finally {
    isLoading.value = false
  }
}

const refreshInvite = async () => {
  isLoading.value = true
  inviteKey.value = ''
  inviteUrl.value = ''

  try {
    await createInvite()
    uni.showToast({ title: '已刷新邀请码', icon: 'success' })
  } catch (error) {
    console.error('刷新邀请码失败:', error)
    uni.showToast({ title: error.message || '刷新失败', icon: 'none' })
  } finally {
    isLoading.value = false
  }
}

const copyInviteKey = () => {
  if (!inviteKey.value) {
    uni.showToast({ title: '暂无邀请码', icon: 'none' })
    return
  }

  uni.setClipboardData({
    data: inviteKey.value,
    success: () => {
      uni.showToast({ title: '邀请码已复制', icon: 'success' })
    }
  })
}

const shareInvite = () => {
  if (!shareText.value) {
    uni.showToast({ title: '暂无可分享内容', icon: 'none' })
    return
  }

  uni.setClipboardData({
    data: shareText.value,
    success: () => {
      uni.showModal({
        title: '分享邀请',
        content: '邀请内容已复制，你可以直接粘贴发送给家人。',
        showCancel: false
      })
    }
  })
}

onLoad((options) => {
  familyId.value = options?.id || ''
  if (familyId.value) {
    loadPageData()
  }
})
</script>

<style lang="scss">
@import '@/components/variables.scss';

page {
  background: $bg-secondary;
}

.invite-container {
  min-height: 100vh;
  padding-bottom: $spacing-xl;
}

.invite-card,
.tips-card {
  margin-bottom: $spacing-base;
}

.card-header {
  margin-bottom: $spacing-lg;

  &.compact {
    margin-bottom: $spacing-base;
  }
}

.card-title {
  display: block;
  font-size: $font-size-lg;
  font-weight: 600;
  color: $text-primary;
  margin-bottom: 8rpx;
}

.card-subtitle {
  font-size: $font-size-sm;
  color: $text-placeholder;
  line-height: 1.6;
}

.qrcode-section {
  text-align: center;
  margin-bottom: $spacing-lg;
}

.qrcode-box {
  min-height: 320rpx;
  padding: $spacing-base;
  background: $bg-secondary;
  border-radius: $radius-lg;
  display: flex;
  align-items: center;
  justify-content: center;
}

.qrcode-tip {
  display: block;
  margin-top: $spacing-sm;
  font-size: $font-size-sm;
  color: $text-placeholder;
}

.info-section {
  padding: $spacing-base 0;
  border-top: 1rpx solid $border-color;
  border-bottom: 1rpx solid $border-color;
  margin-bottom: $spacing-lg;
}

.action-section {
  display: flex;
  gap: $spacing-base;
  margin-bottom: $spacing-lg;
}

.tips-list {
  display: flex;
  flex-direction: column;
  gap: $spacing-base;
}

.tips-item {
  display: flex;
  align-items: flex-start;
  gap: $spacing-sm;
}

.tips-text {
  flex: 1;
  font-size: $font-size-base;
  color: $text-secondary;
  line-height: 1.6;
}
</style>
