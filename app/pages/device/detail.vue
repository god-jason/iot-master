<template>
  <view class="page-device-detail">
    <DsNavBar title="设备详情">
      <template #right>
        <DsIcon name="more" :size="24" color="#5BA3E8" @click="goToEditSimple"/>
      </template>
    </DsNavBar>

    <view class="detail-header header-panel">
      <view class="device-cover">
        <image v-if="device.model_icon" :src="device.model_icon" mode="aspectFill" class="device-cover-image" />
        <view v-else class="device-cover-fallback">
          <DsIcon v-if="device.model_type === 'gateway'" name="server" :size="40" color="#fff" />
          <DsIcon v-else-if="device.model_type === 'sub'" name="electricity" :size="40" color="#fff" />
          <DsIcon v-else name="phone" :size="40" color="#fff" />
        </view>
        <view class="device-state-chip" :class="device.online ? 'online' : 'offline'">
          <text>{{ device.online ? '在线' : '离线' }}</text>
        </view>
      </view>

      <view class="device-info">
        <text class="device-name">{{ device.name || '未命名设备' }}</text>

        <view class="device-sub-info">
          <DsTag v-if="device.model_name" :text="device.model_name" type="primary" size="small" />
          <DsTag v-if="device.room_name" :text="device.room_name" type="warning" size="small" />
          <DsTag :text="deviceTypeText" type="default" size="small" />
        </view>

        <text class="device-desc">{{ device.model_name ? `所属物模型：${device.model_name}` : '该设备暂未绑定物模型说明。' }}</text>

        <view class="device-meta-list">
          <view class="device-meta-item">
            <text class="meta-label">设备编号</text>
            <text class="meta-value">{{ device.sn || '-' }}</text>
          </view>
          <view class="device-meta-item">
            <text class="meta-label">房间位置</text>
            <text class="meta-value">{{ device.room_name || '-' }}</text>
          </view>
          <view class="device-meta-item">
            <text class="meta-label">最近上报</text>
            <text class="meta-value">{{ formatLastTime(device.last_time) }}</text>
          </view>
        </view>

        <view class="device-status-tags">
          <view v-if="device.csq" class="status-tag csq-tag">
            <text class="tag-label">信号</text>
            <view class="csq-bar-wrapper">
              <view class="csq-bar">
                <view class="csq-fill" :style="{ width: getCsqPercent(device.csq) + '%' }" :class="getCsqLevel(device.csq)"></view>
              </view>
              <text class="tag-value">{{ device.csq }}</text>
            </view>
          </view>
          <view v-if="device.error" class="status-tag error-tag">
            <text class="tag-label">错误</text>
            <text class="tag-value">{{ device.error }}</text>
          </view>
          <view v-if="device.longitude && device.latitude" class="status-tag location-tag" @click="openLocation">
            <DsIcon name="map" :size="18" color="#5BA3E8" />
            <text class="tag-value">查看位置</text>
          </view>
        </view>
      </view>
    </view>

    <view v-if="modelControls.length > 0" class="controls">
      <view
        v-for="control in modelControls"
        :key="control.id"
        class="control"
        :style="{ gridColumn: `span ${control.span || 1}` }"
      >
        <UserRender :control="control" :value="getControlValue(control)"
          @change="handleControlChange" @action="handleControlAction" />
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { getDeviceDetail } from '@/api/device'
import { getModel } from '@/utils/models'
import UserRender from '@/controls/UserRender.vue'
import DsNavBar from '@/components/DsNavBar.vue'
import DsIcon from '@/components/DsIcon.vue'
import DsTag from '@/components/DsTag.vue'
import dayjs from 'dayjs'

const device = reactive({
  id: '',
  name: '',
  sn: '',
  type: '',
  online: false,
  gateway_id: '',
  model_id: '',
  model_icon: '',
  model_type: '',
  model_name: '',
  room_name: '',
  last_time: '',
  csq: '',
  error: '',
  longitude: '',
  latitude: ''
})

const modelControls = ref([])
const controlValues = reactive({})
const realtimeProps = ref({})

const deviceTypeText = computed(() => {
  const map = {
    gateway: '网关设备',
    sub: '子设备',
    device: '直连设备'
  }
  return map[device.model_type] || '物联设备'
})

const formatLastTime = (value) => {
  if (!value) return '-'
  return dayjs(value).format('YYYY-MM-DD HH:mm')
}

const goToEdit = () => {
  uni.navigateTo({
    url: `/pages/device/edit?id=${device.id}`
  })
}

const goToEditSimple = () => {
  uni.navigateTo({
    url: `/pages/device/edit-simple?id=${device.id}`
  })
}

const openLocation = () => {
  if (device.longitude && device.latitude) {
    uni.openLocation({
      longitude: parseFloat(device.longitude),
      latitude: parseFloat(device.latitude),
      name: device.name || '设备位置',
      address: '设备所在位置',
      scale: 15
    })
  }
}

const goToLogs = () => {
  uni.navigateTo({
    url: `/pages/device/logs?id=${device.id}`
  })
}

const goToSubDevices = () => {
  uni.navigateTo({
    url: `/pages/device/sub-devices?id=${device.id}`
  })
}

const loadDeviceDetail = async (id) => {
  const deviceId = id || '1'

  try {
    const res = await getDeviceDetail(deviceId)
    const data = res.data || res
    device.id = data.id || deviceId
    device.name = data.name || data.deviceName || '未知设备'
    device.sn = data.sn || data.deviceSn || ''
    device.type = data.type || data.deviceType || 'direct'
    device.gateway_id = data.gateway_id || ''
    device.model_id = data.model_id || ''
    device.model_icon = data.model_icon || ''
    device.model_type = data.model_type || ''
    device.model_name = data.model_name || ''
    device.room_name = data.room_name || ''

    // 从设备数据中获取状态信息
    const statusData = data.status_data || {}
    device.online = data.status === '1' || data.status === 1 || statusData.status === '1' || statusData.status === 1
    device.last_time = statusData.last_time || ''
    device.csq = statusData.csq || ''
    device.error = statusData.error || ''
    device.longitude = statusData.longitude || ''
    device.latitude = statusData.latitude || ''
    realtimeProps.value = statusData

    if (device.model_id) {
      await loadModel(device.model_id)
    }
  } catch (error) {
    device.id = deviceId
    device.name = '未知设备'
    device.sn = ''
    device.type = 'direct'
    device.model_id = ''
    device.model_icon = ''
    device.model_type = ''
    device.model_name = ''
    device.online = false
    device.last_time = ''
    device.csq = ''
    device.error = ''
    device.longitude = ''
    device.latitude = ''
  }
}

const loadModel = async (modelId) => {
  try {
    const model = await getModel(modelId)
    modelControls.value = model.controls || []
  } catch (error) {
    modelControls.value = []
  }
}

const getControlValue = (control) => {
  if (control.binds) {
    for (const paramName of Object.keys(control.binds)) {
      const bindKey = control.binds[paramName]
      if (bindKey && typeof bindKey === 'string') {
        const value = realtimeProps.value[bindKey]
        if (value !== undefined && value !== null && value !== '') {
          return typeof value === 'object' ? value.value : value
        }
      }
    }
  }
  return controlValues[control.id] ?? null
}

const getControlById = (controlId) => {
  return modelControls.value.find(c => c.id === controlId)
}

const handleControlChange = (event) => {
  const { value, controlId } = event || {}
  if (value !== undefined && controlId) {
    controlValues[controlId] = value
  }
}

const handleControlAction = (event) => {
  const { controlId, eventKey } = event || {}
  if (!getControlById(controlId)) return
  // 后端当前未提供设备属性/服务控制接口，仅保留控件交互
  console.warn('设备控制接口未开放:', controlId, eventKey)
  uni.showToast({ title: '设备控制接口暂未开放', icon: 'none' })
}

const formatDateTime = (timeStr) => {
  if (!timeStr) return ''
  return dayjs(Number(timeStr)).format('YYYY/MM/DD HH:mm:ss')
}

const getCsqPercent = (csq) => {
  const value = parseInt(csq) || 0
  return Math.min((value / 30) * 100, 100)
}

const getCsqLevel = (csq) => {
  const value = parseInt(csq) || 0
  if (value >= 25) return 'excellent'
  if (value >= 20) return 'good'
  if (value >= 15) return 'normal'
  if (value >= 10) return 'poor'
  return 'weak'
}

onLoad(async (options) => {
  await loadDeviceDetail(options?.id)
})

onPullDownRefresh(async () => {
  await loadDeviceDetail(device.id)
  uni.stopPullDownRefresh()
})
</script>

<style lang="scss">
@import '@/components/variables.scss';

page {
  background: $bg-secondary;
}

.detail-container {
}

.detail-header {
  display: flex;
  gap: $spacing-lg;
  flex-wrap: nowrap;
  align-items: stretch;
  padding: $spacing-lg;
  margin-bottom: $spacing-base;
}

.header-panel {
  margin: $spacing-base;
  background: $bg-primary;
  border-radius: $radius-lg;
  box-shadow: $shadow-sm;
}

.device-cover {
  position: relative;
  width: 240rpx;
  min-width: 240rpx;
  height: 240rpx;
  border-radius: $radius-xl;
  overflow: hidden;
  background: linear-gradient(135deg, $primary-color 0%, $primary-dark 100%);
  box-shadow: 0 10rpx 24rpx rgba(91, 163, 232, 0.24);
}

.device-cover-image,
.device-cover-fallback {
  width: 100%;
  height: 100%;
}

.device-cover-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
}

.device-state-chip {
  position: absolute;
  left: 16rpx;
  bottom: 16rpx;
  padding: 8rpx 16rpx;
  border-radius: 999rpx;
  background: rgba(15, 23, 42, 0.56);

  text {
    font-size: $font-size-xs;
    color: #fff;
    font-weight: 600;
  }

  &.online {
    background: rgba(34, 197, 94, 0.88);
  }

  &.offline {
    background: rgba(100, 116, 139, 0.88);
  }
}

.device-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;

  .device-name {
    font-size: 40rpx;
    font-weight: 700;
    color: $primary-dark;
  }

  .device-sub-info {
    display: flex;
    flex-wrap: wrap;
    gap: $spacing-sm;
  }

  .device-desc {
    font-size: $font-size-sm;
    color: $text-secondary;
    line-height: 1.6;
  }

  .device-meta-list {
    display: flex;
    flex-direction: column;
    gap: 12rpx;
  }

  .device-meta-item {
    display: flex;
    align-items: flex-start;
    gap: $spacing-sm;

    .meta-label {
      width: 112rpx;
      flex-shrink: 0;
      font-size: $font-size-xs;
      color: $text-placeholder;
    }

    .meta-value {
      flex: 1;
      min-width: 0;
      font-size: $font-size-sm;
      color: $text-primary;
      word-break: break-all;
    }
  }

  .device-status-tags {
    display: flex;
    flex-wrap: wrap;
    gap: $spacing-sm;
    margin-top: 8rpx;

    .status-tag {
      display: flex;
      align-items: center;
      gap: 8rpx;
      padding: 8rpx 16rpx;
      background: $bg-secondary;
      border-radius: $radius-sm;

      .tag-label {
        font-size: $font-size-xs;
        color: $text-secondary;
        flex-shrink: 0;
      }

      .tag-value {
        font-size: $font-size-xs;
        color: $primary-dark;
      }

      &.error-tag {
        background: #fef2f2;
        .tag-label, .tag-value { color: $danger-color; }
      }

      &.location-tag {
        background: #eff6ff;
        .tag-value { color: $primary-color; }
      }

      &.csq-tag {
        flex: 1;
        min-width: 200rpx;

        .csq-bar-wrapper {
          flex: 1;
          display: flex;
          align-items: center;
          gap: 8rpx;
        }

        .csq-bar {
          flex: 1;
          height: 12rpx;
          background: #e2e8f0;
          border-radius: 6rpx;
          overflow: hidden;
        }

        .csq-fill {
          height: 100%;
          border-radius: 6rpx;
          transition: width 0.3s ease;

          &.excellent { background: linear-gradient(90deg, #84cc16 0%, #65a30d 100%); }
          &.good { background: linear-gradient(90deg, $primary-color 0%, $primary-dark 100%); }
          &.normal { background: linear-gradient(90deg, #fbbf24 0%, #f59e0b 100%); }
          &.poor { background: linear-gradient(90deg, #f97316 0%, #ea580c 100%); }
          &.weak { background: linear-gradient(90deg, $danger-color 0%, #dc2626 100%); }
        }
      }
    }
  }
}

.controls {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: $spacing-base;
  padding: $spacing-sm;
  background: $bg-secondary;
  border-radius: $radius-base;

  .control {
    background: $bg-primary;
    border-radius: $radius-base;
    padding: $spacing-base;
    box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.06);
  }
}

.detail-actions {
  margin-bottom: $spacing-base;
}

.actions-grid {
  display: grid;
  gap: 16rpx;
}

.action-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8rpx;

  .action-text {
    font-size: $font-size-xs;
    color: $text-secondary;
  }
}

.data-list {
  padding-top: $spacing-sm;
}

.data-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-sm 0;
  border-bottom: 1rpx solid $border-color;

  &:last-child {
    border-bottom: none;
  }

  .data-left {
    display: flex;
    align-items: center;
    gap: 8rpx;
  }

  .data-label {
    font-size: $font-size-base;
    color: $text-secondary;
  }

  .data-unit {
    font-size: $font-size-xs;
    color: $text-placeholder;
  }

  .data-value {
    font-size: $font-size-lg;
    font-weight: 600;
    color: $primary-color;
  }
}

.control-list {
  padding-top: $spacing-sm;
}

.control-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-base 0;
  border-bottom: 1rpx solid $border-color;

  &:last-child {
    border-bottom: none;
  }

  .control-info {
    flex: 1;

    .control-label {
      font-size: $font-size-base;
      color: $text-primary;
      display: block;
      margin-bottom: 4rpx;
    }

    .control-desc {
      font-size: $font-size-xs;
      color: $text-placeholder;
    }
  }

  .control-value {
    flex-shrink: 0;
    margin-left: $spacing-sm;
  }

  .slider-wrapper {
    width: 400rpx;
    margin-left: $spacing-sm;
  }
}
</style>
