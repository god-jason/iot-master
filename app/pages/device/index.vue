<template>
  <view class="page-device-index">
    <DsNavBar title="设备管理" :show-back="false">
      <template #left>
        <DsIcon name="scan" :size="24" color="#5BA3E8" @click="scanBind" />
      </template>
      <template #right>
        <DsIcon name="conversation" :size="24" color="#5BA3E8" @click="goToMessage" />
      </template>
    </DsNavBar>

    <DsSearchBar v-model="searchText" placeholder="搜索设备" @confirm="onSearch" />

    <view class="device-content">
      <view class="room-tabs">
        <DsTabs v-model="activeRoom" scrollable>
          <DsTabItem label="全部" :index="0" :active="activeRoom === 0" @click="activeRoom = 0" />
          <DsTabItem v-for="room in roomList" :key="room.id" :label="room.name" :index="room.id" :active="activeRoom === room.id" @click="selectRoom(room.id)" />
        </DsTabs>
      </view>

      <view class="device-list">
        <DsEmpty v-if="showNoFamilyTip" text="还没有家庭" subText="创建家庭后开始管理设备" actionText="创建家庭" @action="goToCreateFamily" />
        
        <DsEmpty v-else-if="deviceList.length === 0 && !isLoading" text="暂无设备" subText="点击扫码绑定设备" @click="scanBind" />
        
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
              <text class="device-sn">{{ device.room_name || '-' }}</text>
              <text v-if="device.error" class="device-error">{{ device.error }}</text>
            </view>
            <DsIcon name="next" :size="10" color="#ccc" />
          </view>
        </DsCard>

        <DsLoading v-if="isLoading" text="加载中..." />
        
        <DsNoMore v-if="!hasMore && deviceList.length > 0" text="— 已加载全部 —" />
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh, onShow } from '@dcloudio/uni-app'
import { getDeviceList } from '@/api/device'
import { getRoomList } from '@/api/family'
import { familyList, currentFamily, loadFamilyData } from '@/utils/user'
import { scanCode } from '@/utils/scan'
import { navigateToLogin } from '@/utils/navigation.js'
import DsNavBar from '@/components/DsNavBar.vue'
import DsIcon from '@/components/DsIcon.vue'
import DsSearchBar from '@/components/DsSearchBar.vue'
import DsTabs from '@/components/DsTabs.vue'
import DsTabItem from '@/components/DsTabItem.vue'
import DsEmpty from '@/components/DsEmpty.vue'
import DsCard from '@/components/DsCard.vue'
import DsTag from '@/components/DsTag.vue'
import DsLoading from '@/components/DsLoading.vue'
import DsNoMore from '@/components/DsNoMore.vue'
import dayjs from 'dayjs'

const searchText = ref('')
const deviceList = ref([])
const page = ref(1)
const pageSize = 20
const hasMore = ref(true)
const isLoading = ref(false)
const roomList = ref([])
const activeRoom = ref(0)

const showNoFamilyTip = computed(() => {
  return familyList.value.length === 0 && !isLoading.value
})

const ensureLoggedIn = () => {
  const token = uni.getStorageSync('token')
  if (token) return true
  navigateToLogin()
  return false
}

const scanBind = () => {
  scanCode(() => {
    loadData()
  })
}

const goToMessage = () => {
  uni.navigateTo({ url: '/pages/message/index' })
}

const goToCreateFamily = () => {
  uni.navigateTo({ url: '/pages/family/index' })
}

const goToDetail = (id) => {
  uni.navigateTo({ url: `/pages/device/detail?id=${id}` })
}

const loadRooms = async () => {
  const familyId = currentFamily.value?.id
  if (!familyId) {
    roomList.value = []
    return
  }
  try {
    const res = await getRoomList(familyId)
    roomList.value = res.data || []
  } catch (error) {
    console.error('加载房间列表失败:', error)
  }
}

const selectRoom = (roomId) => {
  activeRoom.value = roomId
  hasMore.value = true
  page.value = 1
  loadDevices(true)
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
    if (currentFamily.value?.id) {
      params.family_id = currentFamily.value.id
    }
    if (activeRoom.value) {
      params.room_id = activeRoom.value
    }
    const res = await getDeviceList(params)
    const data = res.data || []
    const totalCount = res.total || 0
    // 直接从设备数据中获取状态信息
    const formattedData = data.map((item) => {
      const statusData = item.status_data || {}
      return {
        ...item,
        online: item.status === '1' || item.status === 1 || statusData.status === '1' || statusData.status === 1,
        error: statusData.error || '',
        last_time: statusData.last_time || ''
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
    if (error?.message === '登录已过期' || error?.message === '请求失败' || error?.message === '网络错误' || error?.errMsg?.startsWith('request:fail')) {
      return
    }
    console.error('加载设备列表失败:', error)
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

const initPage = async () => {
  activeRoom.value = 0
  // 家庭数据可能由 App.vue 异步加载，首次进入时若尚未就绪则在此补加载
  if (familyList.value.length === 0 && uni.getStorageSync('token')) {
    await loadFamilyData()
  }
  familyIdSnapshot = currentFamily.value?.id || ''
  if (currentFamily.value?.id) {
    await loadRooms()
  } else {
    roomList.value = []
  }
  if (!showNoFamilyTip.value) {
    hasMore.value = true
    page.value = 1
    loadDevices(true)
  }
}

let familyIdSnapshot = ''

onLoad(() => {
  initPage()
})

onShow(() => {
  if ((currentFamily.value?.id || '') !== familyIdSnapshot) {
    initPage()
  }
})

onPullDownRefresh(() => {
  initPage()
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

.page-device-index {
  page {
    background: #F0F7FF;
  }

  & {
    min-height: 100vh;
  }

  .device-content {
  }

  .device-list {
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

    image {
      width: 80rpx;
      height: 80rpx;
      border-radius: 20rpx;
    }
  }

  .device-detail {
    flex: 1;

    .device-name-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 8rpx;
    }

    .device-name {
      font-size: 32rpx;
      font-weight: 600;
      color: #333;
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .device-sn {
      font-size: 24rpx;
      color: #999;
    }

    .device-error {
      font-size: 22rpx;
      color: #ef4444;
      margin-top: 8rpx;
    }
  }
}
</style>
