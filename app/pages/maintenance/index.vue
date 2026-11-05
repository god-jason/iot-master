<template>
  <view class="page-maintenance-index">
    <DsNavBar title="维保工单" :showBack="false" />

    <DsSearchBar v-model="searchText" placeholder="搜索单号或设备名" @confirm="onSearch" />

    <view class="maintenance-content">
      <view class="status-tabs">
        <DsTabs v-model="activeStatus">
          <DsTabItem
            v-for="(tab, index) in statusTabs"
            :key="tab.value"
            :label="tab.label"
            :index="index"
            :active="activeStatus === index"
            @click="selectStatus(index)"
          />
        </DsTabs>
      </view>

      <view class="order-list">
        <DsEmpty v-if="orderList.length === 0 && !isLoading" text="暂无维保工单" subText="下拉刷新获取最新工单" />

        <DsCard
          v-for="order in orderList"
          :key="order.id"
          customClass="order-card"
          @click="goDetail(order.id)"
        >
          <view class="order-head">
            <text class="order-no">{{ order.order_no || '未命名工单' }}</text>
            <DsTag :text="statusText(order.status)" :type="statusTagType(order.status)" size="small" />
          </view>

          <text class="order-title">{{ order.title || '维保任务' }}</text>

          <view class="order-device">
            <DsIcon name="computer" :size="18" color="#5BA3E8" />
            <text class="device-name">{{ order.device_name || '未知设备' }}</text>
            <text v-if="order.device_sn" class="device-sn">{{ order.device_sn }}</text>
          </view>

          <view class="order-meta">
            <view class="meta-item">
              <text class="meta-label">负责人</text>
              <text class="meta-value">{{ order.assignee_name || '未分配' }}</text>
            </view>
            <view class="meta-item">
              <text class="meta-label">应完成</text>
              <text class="meta-value" :class="{ 'meta-overdue': isOverdue(order) }">{{ formatTime(order.due_at) }}</text>
              <text v-if="isOverdue(order)" class="overdue-badge">逾期</text>
            </view>
            <view class="meta-priority">
              <DsTag :text="priorityText(order.priority)" :type="priorityTagType(order.priority)" size="mini" />
            </view>
          </view>
        </DsCard>

        <DsLoading v-if="isLoading" text="加载中..." />
        <DsNoMore v-if="!hasMore && orderList.length > 0" text="— 已加载全部 —" />
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad, onShow, onReachBottom, onPullDownRefresh } from '@dcloudio/uni-app'
import { searchMaintenanceOrders } from '@/api/maintenance'
import { navigateToLogin } from '@/utils/navigation.js'
import dayjs from 'dayjs'
import DsNavBar from '@/components/DsNavBar.vue'
import DsSearchBar from '@/components/DsSearchBar.vue'
import DsTabs from '@/components/DsTabs.vue'
import DsTabItem from '@/components/DsTabItem.vue'
import DsEmpty from '@/components/DsEmpty.vue'
import DsCard from '@/components/DsCard.vue'
import DsTag from '@/components/DsTag.vue'
import DsIcon from '@/components/DsIcon.vue'
import DsLoading from '@/components/DsLoading.vue'
import DsNoMore from '@/components/DsNoMore.vue'

const statusTabs = [
  { label: '全部', value: '' },
  { label: '待处理', value: 'pending' },
  { label: '处理中', value: 'processing' },
  { label: '已完成', value: 'completed' }
]

const statusMap = {
  pending: { text: '待处理', type: 'warning' },
  processing: { text: '处理中', type: 'primary' },
  reviewing: { text: '待复核', type: 'warning' },
  completed: { text: '已完成', type: 'success' },
  cancelled: { text: '已取消', type: 'default' }
}

const priorityMap = {
  1: { text: '优先级 高', type: 'danger' },
  2: { text: '优先级 中', type: 'warning' },
  3: { text: '优先级 低', type: 'default' }
}

const searchText = ref('')
const activeStatus = ref(0)
const orderList = ref([])
const page = ref(1)
const pageSize = 20
const hasMore = ref(true)
const isLoading = ref(false)
const loaded = ref(false)

const ensureLoggedIn = () => {
  const token = uni.getStorageSync('token')
  if (token) return true
  navigateToLogin()
  return false
}

const statusText = (status) => {
  return statusMap[status]?.text || '未知状态'
}

const statusTagType = (status) => {
  return statusMap[status]?.type || 'default'
}

const priorityText = (priority) => {
  return priorityMap[String(priority)]?.text || '优先级 低'
}

const priorityTagType = (priority) => {
  return priorityMap[String(priority)]?.type || 'default'
}

const formatTime = (value) => {
  if (!value) return '-'
  const time = dayjs(value)
  return time.isValid() ? time.format('YYYY-MM-DD HH:mm') : '-'
}

const isOverdue = (order) => {
  if (!order) return false
  if (order.overdue === true || order.overdue === 1 || order.overdue === 'true') return true
  if (order.status === 'completed' || order.status === 'cancelled') return false
  if (!order.due_at) return false
  const due = dayjs(order.due_at)
  return due.isValid() && due.isBefore(dayjs())
}

const loadOrders = async (isRefresh = false) => {
  if (isLoading.value || !ensureLoggedIn()) return
  isLoading.value = true
  try {
    const res = await searchMaintenanceOrders({
      page: isRefresh ? 1 : page.value,
      page_size: pageSize,
      keyword: searchText.value.trim(),
      status: statusTabs[activeStatus.value].value
    })
    const list = Array.isArray(res.data) ? res.data : []
    const total = typeof res.total === 'number' ? res.total : null

    if (isRefresh) {
      orderList.value = list
      page.value = 1
    } else {
      orderList.value = [...orderList.value, ...list]
    }

    hasMore.value = total !== null ? orderList.value.length < total : list.length >= pageSize
  } catch (error) {
    console.error('加载维保工单失败:', error)
  } finally {
    isLoading.value = false
    loaded.value = true
    uni.stopPullDownRefresh()
  }
}

const selectStatus = (index) => {
  if (activeStatus.value === index) return
  activeStatus.value = index
  hasMore.value = true
  page.value = 1
  loadOrders(true)
}

const onSearch = () => {
  hasMore.value = true
  page.value = 1
  loadOrders(true)
}

const goDetail = (id) => {
  if (!id) return
  uni.navigateTo({ url: `/pages/maintenance/detail?id=${id}` })
}

onLoad(() => {
  loadOrders(true)
})

onShow(() => {
  if (loaded.value) {
    loadOrders(true)
  }
})

onPullDownRefresh(() => {
  hasMore.value = true
  page.value = 1
  loadOrders(true)
})

onReachBottom(() => {
  if (hasMore.value && !isLoading.value) {
    page.value++
    loadOrders()
  }
})
</script>

<style lang="scss">
@import '@/components/variables.scss';

page {
  background: $bg-secondary;
}

.page-maintenance-index {
  min-height: 100vh;
}

.maintenance-content {
  padding-bottom: $spacing-xl;
}

.status-tabs {
  background: $bg-primary;
  margin-bottom: $spacing-xs;
}

.order-list {
  padding: 0;
}

.order-card {
  margin-bottom: $spacing-sm;
}

.order-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: $spacing-sm;
  margin-bottom: $spacing-xs;

  .order-no {
    flex: 1;
    min-width: 0;
    font-size: $font-size-sm;
    color: $text-secondary;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.order-title {
  display: block;
  font-size: $font-size-lg;
  font-weight: 600;
  color: $text-primary;
  line-height: 1.5;
  margin-bottom: $spacing-sm;
}

.order-device {
  display: flex;
  align-items: center;
  gap: $spacing-xs;
  padding: $spacing-sm 0;
  border-top: 1rpx solid $border-light;

  .device-name {
    font-size: $font-size-base;
    color: $primary-dark;
    max-width: 60%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .device-sn {
    font-size: $font-size-xs;
    color: $text-placeholder;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.order-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: $spacing-sm $spacing-lg;
  padding-top: $spacing-sm;
  border-top: 1rpx solid $border-light;

  .meta-item {
    display: flex;
    align-items: center;
    gap: $spacing-xs;
  }

  .meta-label {
    font-size: $font-size-xs;
    color: $text-placeholder;
  }

  .meta-value {
    font-size: $font-size-sm;
    color: $text-secondary;
  }

  .meta-overdue {
    color: $danger-color;
    font-weight: 600;
  }

  .meta-priority {
    margin-left: auto;
  }
}

.overdue-badge {
  padding: 2rpx 8rpx;
  font-size: $font-size-xs;
  color: #fff;
  background: $danger-color;
  border-radius: $radius-sm;
}
</style>
