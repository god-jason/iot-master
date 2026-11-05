<template>
  <view class="page-device-sub-devices">
    <DsNavBar title="子设备" />
    <DsCard v-if="subDevices.length > 0">
      <DsList>
        <DsListItem 
          v-for="device in subDevices" 
          :key="device.id"
          :title="device.name"
          :desc="`SN: ${device.sn}`"
          :icon="getDeviceIcon(device.type)"
          :icon-color="device.online ? '#fff' : '#999'"
          :icon-bg="device.online ? 'linear-gradient(135deg, #5BA3E8 0%, #1E3A5F 100%)' : '#F0F7FF'"
          @click="goToDetail(device.id)"
        >
          <template #extra>
            <DsBadge :type="device.online ? 'success' : 'default'" :text="device.online ? '在线' : '离线'" />
          </template>
        </DsListItem>
      </DsList>
    </DsCard>

    <DsEmpty v-else-if="!isLoading" text="暂无子设备" />

    <DsLoading v-if="isLoading" text="加载中..." />
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { getSubDevices } from '@/api/device'
import DsNavBar from '@/components/DsNavBar.vue'
import DsCard from '@/components/DsCard.vue'
import DsList from '@/components/DsList.vue'
import DsListItem from '@/components/DsListItem.vue'
import DsBadge from '@/components/DsBadge.vue'
import DsEmpty from '@/components/DsEmpty.vue'
import DsLoading from '@/components/DsLoading.vue'
const subDevices = ref([])
const isLoading = ref(false)
const deviceId = ref('')

const goToDetail = (id) => {
  uni.navigateTo({
    url: `/pages/device/detail?id=${id}`
  })
}

const getDeviceIcon = (type) => {
  const icons = {
    'sensor': 'bar-chart',
    'switch': 'electricity',
    'outlet': 'electricity',
    'light': 'bulb',
    'thermostat': 'thermometer'
  }
  return icons[type] || 'phone'
}

const loadSubDevices = async () => {
  isLoading.value = true
  
  try {
    const res = await getSubDevices(deviceId.value)
    const data = res.data || []
    subDevices.value = data.map(item => ({
      ...item,
      // 直接从设备数据中获取在线状态
      online: item.status === '1' || item.status === 1 || (item.status_data || {}).status === '1' || (item.status_data || {}).status === 1
    }))
  } catch (error) {
    console.error('加载子设备失败:', error)
    subDevices.value = []
  } finally {
    isLoading.value = false
    uni.stopPullDownRefresh()
  }
}

onLoad((options) => {
  deviceId.value = options?.id || ''
  loadSubDevices()
})

onPullDownRefresh(() => {
  loadSubDevices()
})
</script>

<style lang="scss">
@import '@/components/variables.scss';

page {
  background: $bg-secondary;
}

.sub-devices-container {
  min-height: 100vh;
}
</style>
