<template>
  <view class="page-device-logs">
    <DsNavBar title="设备日志" />
    <view class="logs-list">
      <DsEmpty v-if="logsList.length === 0 && !isLoading" text="暂无日志" />

      <DsCard 
        v-for="log in logsList" 
        :key="log.id" 
        class="log-card"
      >
        <view class="log-header">
          <DsTag :type="getLogTagType(log.level)" :text="getLogLevelText(log.level)" />
          <text class="log-time">{{ formatTime(log.created) }}</text>
        </view>
        <text class="log-message">{{ log.message }}</text>
        <text class="log-device">{{ log.device_name || log.sn || '未知设备' }}</text>
      </DsCard>

      <DsLoading v-if="isLoading" text="加载中..." />

      <DsNoMore v-if="!hasMore && logsList.length > 0" text="— 已加载全部 —" />
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { getDeviceLogs } from '@/api/device'
import DsNavBar from '@/components/DsNavBar.vue'
import DsEmpty from '@/components/DsEmpty.vue'
import DsCard from '@/components/DsCard.vue'
import DsTag from '@/components/DsTag.vue'
import DsLoading from '@/components/DsLoading.vue'
import DsNoMore from '@/components/DsNoMore.vue'
import dayjs from 'dayjs'

const logsList = ref([])
const page = ref(1)
const pageSize = 20
const total = ref(0)
const hasMore = ref(true)
const isLoading = ref(false)
const isRefreshing = ref(false)
const deviceId = ref('')

const getLogLevelText = (level) => {
  const levels = {
    'error': '错误',
    'warn': '警告',
    'info': '信息',
    'debug': '调试'
  }
  return levels[level] || '信息'
}

const getLogTagType = (level) => {
  const types = {
    'error': 'danger',
    'warn': 'warning',
    'info': 'primary',
    'debug': 'default'
  }
  return types[level] || 'default'
}

const formatTime = (timestamp) => {
  if (!timestamp) return ''
  return dayjs(timestamp).format('YYYY-MM-DD HH:mm:ss')
}

const loadLogs = async (isRefresh = false) => {
  if (isLoading.value) return
  if (!deviceId.value) return
  
  isLoading.value = true
  
  try {
    const res = await getDeviceLogs(deviceId.value, { 
      page: isRefresh ? 1 : page.value, 
      size: pageSize 
    })
    
    const data = res.data || []
    const totalCount = res.total || 0
    
    const formattedData = data.map(item => ({
      id: item.id,
      message: item.message || item.content || '',
      level: item.level || 'info',
      created: item.created || item.timestamp || Date.now(),
      device_name: item.device_name || '',
      sn: item.sn || ''
    }))
    
    if (isRefresh) {
      logsList.value = formattedData
      page.value = 1
    } else {
      logsList.value = [...logsList.value, ...formattedData]
    }
    
    total.value = totalCount
    hasMore.value = logsList.value.length < totalCount
  } catch (error) {
    if (error?.message === '登录已过期' || error?.message === '请求失败' || error?.message === '网络错误') {
      return
    }

    if (error?.errMsg?.startsWith('request:fail')) {
      return
    }

    console.error('加载日志失败:', error)
    uni.showToast({
      title: '加载失败',
      icon: 'none'
    })
  } finally {
    isLoading.value = false
    isRefreshing.value = false
    uni.stopPullDownRefresh()
  }
}

const onRefresh = () => {
  hasMore.value = true
  loadLogs(true)
}

const onLoadMore = () => {
  if (hasMore.value && !isLoading.value) {
    page.value++
    loadLogs()
  }
}

onLoad((options) => {
  deviceId.value = options?.id || ''
  loadLogs()
})

onPullDownRefresh(() => {
  onRefresh()
})

onReachBottom(() => {
  onLoadMore()
})
</script>

<style lang="scss">
@import '@/components/variables.scss';

page {
  background: $bg-secondary;
}

.logs-container {
  min-height: 100vh;
}

.logs-list {
}

.log-card {
  padding: $spacing-base;
  margin-bottom: $spacing-sm;
}

.log-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-sm;
}

.log-time {
  font-size: $font-size-xs;
  color: $text-placeholder;
}

.log-message {
  font-size: $font-size-base;
  color: $text-primary;
  display: block;
  margin-bottom: 12rpx;
  line-height: 1.6;
}

.log-device {
  font-size: $font-size-sm;
  color: $text-placeholder;
}
</style>
