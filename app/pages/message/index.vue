<template>
  <view class="page-message-index">
    <DsNavBar title="消息中心" />
    <DsEmpty v-if="messageList.length === 0 && !isLoading" text="暂无消息">
      <DsIcon name="conversation" :size="24" color="#ccc" />
    </DsEmpty>

    <DsCard 
      v-for="msg in messageList" 
      :key="msg.id" 
      class="message-card"
      @click="handleMessageClick(msg)"
    >
      <view class="message-icon" :class="msg.type">
        <DsIcon v-if="msg.type === 'device'" name="phone" :size="24" color="#666" />
        <DsIcon v-else-if="msg.type === 'event'" name="warning" :size="24" color="#f59e0b" />
        <DsIcon v-else-if="msg.type === 'system'" name="conversation" :size="24" color="#3b82f6" />
        <DsIcon v-else-if="msg.type === 'alert'" name="warning" :size="24" color="#ef4444" />
        <DsIcon v-else name="text" :size="24" color="#666" />
      </view>
      <view class="message-content">
        <view class="message-header">
          <text class="message-title">{{ msg.title }}</text>
          <text class="message-time">{{ formatTime(msg.created) }}</text>
        </view>
        <text class="message-desc">{{ msg.content }}</text>
      </view>
      <DsBadge v-if="!msg.read" type="dot" />
    </DsCard>

    <DsLoading v-if="isLoading" text="加载中..." />

    <DsNoMore v-if="!hasMore && messageList.length > 0" text="— 已加载全部 —" />
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { getDeviceEvents } from '@/api/device'
import DsNavBar from '@/components/DsNavBar.vue'
import DsEmpty from '@/components/DsEmpty.vue'
import DsIcon from '@/components/DsIcon.vue'
import DsCard from '@/components/DsCard.vue'
import DsBadge from '@/components/DsBadge.vue'
import DsLoading from '@/components/DsLoading.vue'
import DsNoMore from '@/components/DsNoMore.vue'
import dayjs from 'dayjs'

const messageList = ref([])
const page = ref(1)
const pageSize = 20
const total = ref(0)
const hasMore = ref(true)
const isLoading = ref(false)
const isRefreshing = ref(false)

const formatTime = (timestamp) => {
  if (!timestamp) return ''
  
  const now = dayjs()
  const date = dayjs(timestamp)
  const diffMinutes = now.diff(date, 'minute')
  const diffHours = now.diff(date, 'hour')
  
  if (diffMinutes < 1) {
    return '刚刚'
  } else if (diffMinutes < 60) {
    return `${diffMinutes}分钟前`
  } else if (diffHours < 24) {
    return `${diffHours}小时前`
  } else {
    return date.format('MM-DD')
  }
}

const handleMessageClick = (msg) => {
  uni.showToast({
    title: '消息详情开发中',
    icon: 'none'
  })
}

const loadMessages = async (isRefresh = false) => {
  if (isLoading.value) return
  
  isLoading.value = true
  
  try {
    const res = await getDeviceEvents({ 
      page: isRefresh ? 1 : page.value, 
      size: pageSize 
    })
    
    const data = res.data || []
    const totalCount = res.total || 0
    
    const formattedData = data.map(item => ({
      id: item.id,
      title: item.type || '设备事件',
      content: item.message || item.code || '暂无描述',
      type: item.type || 'event',
      read: item.status === 1,
      created: item.created || Date.now()
    }))
    
    if (isRefresh) {
      messageList.value = formattedData
      page.value = 1
    } else {
      messageList.value = [...messageList.value, ...formattedData]
    }
    
    total.value = totalCount
    hasMore.value = messageList.value.length < totalCount
  } catch (error) {
    if (error?.message === '登录已过期' || error?.message === '请求失败' || error?.message === '网络错误') {
      return
    }

    if (error?.errMsg?.startsWith('request:fail')) {
      return
    }

    console.error('加载消息列表失败:', error)
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
  loadMessages(true)
}

const onLoadMore = () => {
  if (hasMore.value && !isLoading.value) {
    page.value++
    loadMessages()
  }
}

onLoad(() => {
  loadMessages()
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

.message-container {
  min-height: 100vh;
}

.message-card {
  display: flex;
  gap: $spacing-base;
  padding: $spacing-base;
  margin-bottom: $spacing-sm;

  &:active {
    background: $bg-secondary;
  }
}

.message-icon {
  width: 72rpx;
  height: 72rpx;
  border-radius: $radius-base;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;

  &.device {
    background: #dbeafe;
  }

  &.event {
    background: #fef3c7;
  }

  &.system {
    background: #f3e8ff;
  }

  &.alert {
    background: #fee2e2;
  }

  &.info {
    background: #e0f2fe;
  }
}

.message-content {
  flex: 1;
  overflow: hidden;
}

.message-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10rpx;

  .message-title {
    font-size: $font-size-base;
    font-weight: 600;
    color: $text-primary;
  }

  .message-time {
    font-size: $font-size-xs;
    color: $text-placeholder;
    flex-shrink: 0;
    margin-left: $spacing-base;
  }
}

.message-desc {
  font-size: $font-size-sm;
  color: $text-secondary;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
