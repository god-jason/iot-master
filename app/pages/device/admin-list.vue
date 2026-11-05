<template>
  <view class="page-device-admin-list">
    <DsNavBar title="管理所有设备" :show-back="true">
      <template #right>
        <DsIcon name="scan" :size="24" color="#5BA3E8" @click="handleScan" />
      </template>
    </DsNavBar>

    <DsSearchBar v-model="searchText" placeholder="按设备名 / SN 搜索" @confirm="onSearch" />

    <view class="device-list">
      <DsEmpty v-if="deviceList.length === 0 && !isLoading" text="暂无设备" subText="当前系统下还没有任何设备" />

      <DsCard v-for="device in deviceList" :key="device.id" customClass="device-card" @click="goToDetail(device.id)">
        <view class="device-info">
          <view class="device-icon">
            <image v-if="device.model_icon" :src="device.model_icon" mode="aspectFill" />
            <DsIcon v-else-if="device.model_type === 'gateway'" name="server" :size="22" color="#fff" />
            <DsIcon v-else-if="device.model_type === 'sub'" name="electricity" :size="22" color="#fff" />
            <DsIcon v-else name="phone" :size="22" color="#fff" />
          </view>
          <view class="device-detail">
            <view class="device-name-row">
              <text class="device-name">{{ device.name || '未知设备' }}</text>
              <DsTag :type="device.online ? 'success' : 'default'" :text="device.online ? '在线' : '离线'" size="small" />
            </view>
            <text class="device-sn">SN: {{ device.sn || '-' }}</text>
            <view class="device-meta">
              <text class="meta-item" v-if="device.model_name">模型: {{ device.model_name }}</text>
              <text class="meta-item" v-if="device.family_name">家庭: {{ device.family_name }}</text>
              <text class="meta-item" v-if="device.user_name">归属: {{ device.user_name }}</text>
            </view>
          </view>
          <DsIcon name="next" :size="10" color="#ccc" />
        </view>
      </DsCard>

      <DsLoading v-if="isLoading" text="加载中..." />
      <DsNoMore v-if="!hasMore && deviceList.length > 0" text="— 已加载全部 —" />
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { getDeviceList } from '@/api/device'
import { navigateToLogin } from '@/utils/navigation.js'
import DsNavBar from '@/components/DsNavBar.vue'
import DsIcon from '@/components/DsIcon.vue'
import DsSearchBar from '@/components/DsSearchBar.vue'
import DsEmpty from '@/components/DsEmpty.vue'
import DsCard from '@/components/DsCard.vue'
import DsTag from '@/components/DsTag.vue'
import DsLoading from '@/components/DsLoading.vue'
import DsNoMore from '@/components/DsNoMore.vue'

const searchText = ref('')
const deviceList = ref([])
const page = ref(1)
const pageSize = 20
const hasMore = ref(true)
const isLoading = ref(false)

const ensureLoggedIn = () => {
  const token = uni.getStorageSync('token')
  if (token) return true
  navigateToLogin()
  return false
}

const goToDetail = (id) => {
  uni.navigateTo({ url: `/pages/device/detail?id=${id}` })
}

const handleScan = () => {
  uni.scanCode({
    onlyFromCamera: false,
    success: (res) => {
      uni.setClipboardData({ data: res.result, showToast: false })
      uni.showToast({ title: '已复制 SN', icon: 'none' })
    },
    fail: () => {}
  })
}

const loadDevices = async (isRefresh = false) => {
  if (isLoading.value || !ensureLoggedIn()) return
  isLoading.value = true
  try {
    const params = {
      page: isRefresh ? 1 : page.value,
      size: pageSize,
      keyword: searchText.value
    }
    // 后端管理员默认返回全量（不按 family_id / room_id 过滤）
    const res = await getDeviceList(params)
    const data = res.data || []
    const totalCount = res.total || 0
    const formattedData = data.map((item) => {
      const statusData = item.status_data || {}
      return {
        ...item,
        online: item.status === '1' || item.status === 1 || statusData.status === '1' || statusData.status === 1
      }
    })
    if (isRefresh) {
      deviceList.value = formattedData
      page.value = 1
    } else {
      deviceList.value = [...deviceList.value, ...formattedData]
    }
    hasMore.value = deviceList.value.length < totalCount
  } catch (error) {
    console.error('加载管理员设备列表失败:', error)
    uni.showToast({ title: '加载失败', icon: 'none' })
  } finally {
    isLoading.value = false
    uni.stopPullDownRefresh()
  }
}

const onSearch = () => {
  hasMore.value = true
  page.value = 1
  loadDevices(true)
}

onLoad(() => {
  loadDevices(true)
})

onPullDownRefresh(() => {
  loadDevices(true)
})

onReachBottom(() => {
  if (hasMore.value && !isLoading.value) {
    page.value++
    loadDevices()
  }
})
</script>

<style lang="scss">
@import '@/components/variables.scss';

.page-device-admin-list {
  min-height: 100vh;
  background: $bg-secondary;
  padding-bottom: $spacing-xl;

  .device-list {
    padding: $spacing-sm $spacing-base;
  }

  .device-card {
    margin-bottom: 20rpx;
  }

  .device-info {
    display: flex;
    align-items: center;
    gap: 24rpx;
  }

  .device-icon {
    width: 88rpx;
    height: 88rpx;
    border-radius: 20rpx;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #F0F7FF;
    flex-shrink: 0;

    image {
      width: 80rpx;
      height: 80rpx;
      border-radius: 20rpx;
    }
  }

  .device-detail {
    flex: 1;
    min-width: 0;

    .device-name-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 6rpx;
    }

    .device-name {
      font-size: 32rpx;
      font-weight: 600;
      color: #333;
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      margin-right: 16rpx;
    }

    .device-sn {
      display: block;
      font-size: 24rpx;
      color: #999;
      margin-bottom: 4rpx;
    }

    .device-meta {
      display: flex;
      flex-wrap: wrap;
      gap: 8rpx 16rpx;

      .meta-item {
        font-size: 22rpx;
        color: #6ba3d6;
      }
    }
  }
}
</style>
