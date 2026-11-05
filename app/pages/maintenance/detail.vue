<template>
  <view class="page-maintenance-detail">
    <DsNavBar title="工单详情" />

    <DsLoading v-if="loading" text="加载中..." />

    <view v-else-if="loadError" class="state-wrap">
      <DsEmpty text="工单加载失败" subText="请返回列表后重试" type="error" />
    </view>

    <view v-else class="detail-body">
      <DsCard customClass="order-card">
        <view class="order-head">
          <text class="order-no">{{ order.order_no || '-' }}</text>
          <DsTag :text="statusText" :type="statusType" size="small" />
        </view>

        <text class="order-title">{{ order.title || '维保任务' }}</text>

        <view class="info-list">
          <view class="info-row">
            <text class="label">设备</text>
            <text class="value">
              {{ order.device_name || '未知设备' }}
              <text v-if="order.device_sn" class="value-sub">（{{ order.device_sn }}）</text>
            </text>
          </view>
          <view class="info-row">
            <text class="label">类型</text>
            <text class="value">{{ typeText }}</text>
          </view>
          <view class="info-row">
            <text class="label">优先级</text>
            <text class="value">{{ priorityText }}</text>
          </view>
          <view class="info-row">
            <text class="label">负责人</text>
            <text class="value">{{ order.assignee_name || '未分配' }}</text>
          </view>
          <view class="info-row">
            <text class="label">应完成</text>
            <text class="value" :class="{ 'overdue-text': isOverdue }">
              {{ formatTime(order.due_at) }}
              <text v-if="isOverdue" class="overdue-badge">逾期</text>
            </text>
          </view>
          <view v-if="order.plan_name" class="info-row">
            <text class="label">维保计划</text>
            <text class="value">{{ order.plan_name }}</text>
          </view>
          <view v-if="order.finished_at" class="info-row">
            <text class="label">完成时间</text>
            <text class="value">{{ formatTime(order.finished_at) }}</text>
          </view>
        </view>
      </DsCard>

      <DsCard title="检查项">
        <view v-if="itemsForm.length === 0" class="empty-text">该工单暂无检查项</view>

        <view v-for="(item, index) in itemsForm" :key="index" class="check-item">
          <text class="check-name">{{ item.name || `检查项 ${index + 1}` }}</text>

          <template v-if="canEdit">
            <view class="result-options">
              <view
                v-for="option in itemResultOptions"
                :key="option"
                class="result-option"
                :class="{ active: item.result === option }"
                @click="item.result = option"
              >
                {{ option }}
              </view>
            </view>
            <input
              v-model="item.remark"
              class="remark-input"
              type="text"
              placeholder="填写检查备注（选填）"
              placeholder-class="input-placeholder"
            />
          </template>

          <view v-else class="check-readonly">
            <text class="readonly-result">{{ item.result || '未填写' }}</text>
            <text v-if="item.remark" class="readonly-remark">{{ item.remark }}</text>
          </view>
        </view>
      </DsCard>

      <DsCard title="现场照片">
        <view v-if="attachments.length === 0 && !canEdit" class="empty-text">暂无现场照片</view>

        <view class="photo-grid">
          <view v-for="(url, index) in attachments" :key="`${url}-${index}`" class="photo-item">
            <image :src="url" mode="aspectFill" @click="previewImage(url)" />
            <view v-if="canEdit" class="photo-remove" @click.stop="removeAttachment(index)">
              <DsIcon name="close" :size="16" color="#fff" />
            </view>
          </view>
          <view v-if="canEdit && attachments.length < 9" class="photo-add" @click="chooseImages">
            <DsIcon name="camera" :size="26" color="#999" />
            <text class="photo-add-text">{{ uploading ? '上传中...' : '拍照 / 选图' }}</text>
          </view>
        </view>
      </DsCard>

      <DsCard title="备件登记">
        <template v-if="canEdit">
          <view v-if="partsForm.length === 0" class="empty-text">暂未登记备件</view>

          <view v-for="(part, index) in partsForm" :key="index" class="part-item">
            <view class="part-main">
              <input
                v-model="part.name"
                class="part-name"
                type="text"
                placeholder="备件名称"
                placeholder-class="input-placeholder"
              />
              <DsIcon name="dustbin" :size="18" color="#f56c6c" @click="removePart(index)" />
            </view>
            <view class="part-numbers">
              <view class="part-field">
                <text class="part-label">数量</text>
                <input v-model="part.qty" class="part-input" type="number" />
              </view>
              <view class="part-field">
                <text class="part-label">单价(元)</text>
                <input v-model="part.price" class="part-input" type="digit" />
              </view>
              <text class="part-subtotal">小计 ¥{{ partSubtotal(part) }}</text>
            </view>
          </view>

          <view class="add-part" @click="addPart">
            <DsIcon name="add" :size="18" color="#5BA3E8" />
            <text>添加备件</text>
          </view>
        </template>

        <template v-else>
          <view v-if="partsForm.length === 0" class="empty-text">暂无备件记录</view>
          <view v-for="(part, index) in partsForm" :key="index" class="part-readonly">
            <text class="part-readonly-name">{{ part.name }}</text>
            <text class="part-readonly-qty">x{{ part.qty }}</text>
            <text class="part-readonly-price">¥{{ part.price }}</text>
          </view>
        </template>

        <view class="cost-row">
          <text class="cost-label">备件成本合计</text>
          <text class="cost-value">¥{{ partsCost }}</text>
        </view>
      </DsCard>

      <DsCard title="执行结果">
        <template v-if="canEdit">
          <view class="section-label">
            整体结果
            <text class="required">*</text>
          </view>
          <view class="result-options overall">
            <view
              v-for="option in overallResultOptions"
              :key="option.value"
              class="result-option"
              :class="{ active: completeForm.result === option.value }"
              @click="completeForm.result = option.value"
            >
              {{ option.label }}
            </view>
          </view>
          <textarea
            v-model="completeForm.remark"
            class="remark-textarea"
            placeholder="填写处理说明（选填）"
            placeholder-class="input-placeholder"
            maxlength="500"
          />
        </template>

        <template v-else>
          <view class="info-row">
            <text class="label">整体结果</text>
            <text class="value">{{ overallResultText }}</text>
          </view>
          <view v-if="order.remark" class="info-row">
            <text class="label">处理说明</text>
            <text class="value">{{ order.remark }}</text>
          </view>
        </template>
      </DsCard>

      <view class="bottom-space"></view>
    </view>

    <view v-if="canEdit && !loading" class="footer-bar">
      <DsButton
        v-if="order.status === 'pending'"
        text="开始处理"
        type="primary"
        size="large"
        :loading="busy"
        @click="handleStart"
      />
      <DsButton
        v-if="order.status === 'processing'"
        text="提交完成"
        type="success"
        size="large"
        :loading="busy"
        @click="handleComplete"
      />
      <DsButton text="取消工单" type="ghost" size="large" :disabled="busy" @click="handleCancel" />
    </view>
  </view>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import {
  getMaintenanceOrderDetail,
  startMaintenanceOrder,
  completeMaintenanceOrder,
  cancelMaintenanceOrder
} from '@/api/maintenance'
import dayjs from 'dayjs'
import DsNavBar from '@/components/DsNavBar.vue'
import DsCard from '@/components/DsCard.vue'
import DsTag from '@/components/DsTag.vue'
import DsIcon from '@/components/DsIcon.vue'
import DsButton from '@/components/DsButton.vue'
import DsEmpty from '@/components/DsEmpty.vue'
import DsLoading from '@/components/DsLoading.vue'

const apiBaseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'

const statusMap = {
  pending: { text: '待处理', type: 'warning' },
  processing: { text: '处理中', type: 'primary' },
  reviewing: { text: '待复核', type: 'warning' },
  completed: { text: '已完成', type: 'success' },
  cancelled: { text: '已取消', type: 'default' }
}

const priorityMap = {
  1: { text: '高', type: 'danger' },
  2: { text: '中', type: 'warning' },
  3: { text: '低', type: 'default' }
}

const typeMap = {
  plan: '计划维保',
  fault: '故障维修',
  patrol: '巡检任务',
  install: '安装任务'
}

const resultMap = {
  normal: '正常',
  fixed: '已修复',
  replaced: '已更换',
  unresolved: '未解决'
}

const itemResultOptions = ['正常', '异常', '已修复']
const overallResultOptions = [
  { value: 'normal', label: '正常' },
  { value: 'fixed', label: '已修复' },
  { value: 'replaced', label: '已更换' },
  { value: 'unresolved', label: '未解决' }
]

const orderId = ref('')
const loading = ref(true)
const loadError = ref(false)
const busy = ref(false)
const uploading = ref(false)

const order = reactive({
  id: '',
  order_no: '',
  device_id: '',
  device_name: '',
  device_sn: '',
  type: '',
  title: '',
  priority: '',
  status: '',
  assignee_id: '',
  assignee_name: '',
  due_at: '',
  finished_at: '',
  result: '',
  cost: 0,
  remark: '',
  overdue: false,
  plan_name: ''
})

const itemsForm = ref([])
const partsForm = ref([])
const attachments = ref([])
const completeForm = reactive({
  result: '',
  remark: ''
})

const canEdit = computed(() => {
  return order.status === 'pending' || order.status === 'processing'
})

const statusText = computed(() => statusMap[order.status]?.text || '未知状态')
const statusType = computed(() => statusMap[order.status]?.type || 'default')
const priorityText = computed(() => priorityMap[String(order.priority)]?.text || '低')
const typeText = computed(() => typeMap[order.type] || '维保任务')
const overallResultText = computed(() => resultMap[order.result] || (order.result ? order.result : '未填写'))

const isOverdue = computed(() => {
  if (order.overdue === true || order.overdue === 1 || order.overdue === 'true') return true
  if (order.status === 'completed' || order.status === 'cancelled') return false
  if (!order.due_at) return false
  const due = dayjs(order.due_at)
  return due.isValid() && due.isBefore(dayjs())
})

const partsCost = computed(() => {
  const total = partsForm.value.reduce((sum, part) => {
    if (!part.name || !part.name.trim()) return sum
    return sum + (Number(part.qty) || 0) * (Number(part.price) || 0)
  }, 0)
  return Number(total.toFixed(2))
})

const parseList = (value) => {
  if (Array.isArray(value)) return value
  if (typeof value === 'string' && value) {
    try {
      const parsed = JSON.parse(value)
      return Array.isArray(parsed) ? parsed : []
    } catch (error) {
      return []
    }
  }
  return []
}

const formatTime = (value) => {
  if (!value) return '-'
  const time = dayjs(value)
  return time.isValid() ? time.format('YYYY-MM-DD HH:mm') : '-'
}

const partSubtotal = (part) => {
  const total = (Number(part.qty) || 0) * (Number(part.price) || 0)
  return Number(total.toFixed(2))
}

const loadDetail = async () => {
  if (!orderId.value) return
  loading.value = true
  loadError.value = false
  try {
    const res = await getMaintenanceOrderDetail(orderId.value)
    const data = res.data || res || {}

    order.id = data.id || orderId.value
    order.order_no = data.order_no || ''
    order.device_id = data.device_id || ''
    order.device_name = data.device_name || ''
    order.device_sn = data.device_sn || ''
    order.type = data.type || ''
    order.title = data.title || ''
    order.priority = data.priority ?? ''
    order.status = data.status || ''
    order.assignee_id = data.assignee_id || ''
    order.assignee_name = data.assignee_name || ''
    order.due_at = data.due_at || ''
    order.finished_at = data.finished_at || ''
    order.result = data.result || ''
    order.cost = data.cost || 0
    order.remark = data.remark || ''
    order.overdue = data.overdue
    order.plan_name = data.plan_name || ''

    itemsForm.value = parseList(data.items).map((item, index) => ({
      name: item?.name || `检查项 ${index + 1}`,
      result: item?.result || '',
      remark: item?.remark || ''
    }))

    partsForm.value = parseList(data.parts).map((part) => ({
      name: part?.name || '',
      qty: part?.qty ?? 1,
      price: part?.price ?? 0
    }))

    attachments.value = parseList(data.attachments)
      .map((item) => (typeof item === 'string' ? item : item?.url || ''))
      .filter(Boolean)

    completeForm.result = data.result || ''
    completeForm.remark = data.remark || ''
  } catch (error) {
    console.error('加载工单详情失败:', error)
    loadError.value = true
  } finally {
    loading.value = false
    uni.stopPullDownRefresh()
  }
}

const previewImage = (url) => {
  uni.previewImage({
    urls: attachments.value,
    current: url
  })
}

const uploadAttachment = (filePath) => {
  return new Promise((resolve, reject) => {
    const token = uni.getStorageSync('token')
    uni.uploadFile({
      url: `${apiBaseURL}/upload/attachment`,
      filePath,
      name: 'file',
      header: {
        Authorization: token ? `Bearer ${token}` : ''
      },
      success: (res) => {
        try {
          const data = JSON.parse(res.data)
          if (data.code === 0) {
            resolve(data.data?.url || data.url || '')
            return
          }
          reject(new Error(data.message || '上传失败'))
        } catch (error) {
          reject(new Error('上传返回格式异常'))
        }
      },
      fail: () => reject(new Error('上传失败'))
    })
  })
}

const chooseImages = () => {
  if (uploading.value) return
  const remain = Math.max(1, 9 - attachments.value.length)
  uni.chooseImage({
    count: remain,
    sizeType: ['compressed'],
    sourceType: ['camera', 'album'],
    success: async (res) => {
      const paths = res.tempFilePaths || []
      if (paths.length === 0) return
      uploading.value = true
      try {
        for (const filePath of paths) {
          const url = await uploadAttachment(filePath)
          if (url) attachments.value.push(url)
        }
      } catch (error) {
        console.error('上传现场照片失败:', error)
        uni.showToast({ title: error.message || '照片上传失败', icon: 'none' })
      } finally {
        uploading.value = false
      }
    }
  })
}

const removeAttachment = (index) => {
  uni.showModal({
    title: '删除照片',
    content: '确定要删除这张现场照片吗？',
    success: (res) => {
      if (res.confirm) {
        attachments.value.splice(index, 1)
      }
    }
  })
}

const addPart = () => {
  partsForm.value.push({ name: '', qty: 1, price: 0 })
}

const removePart = (index) => {
  partsForm.value.splice(index, 1)
}

const handleStart = async () => {
  if (busy.value || !order.id) return
  busy.value = true
  try {
    await startMaintenanceOrder(order.id)
    order.status = 'processing'
    uni.showToast({ title: '已开始处理', icon: 'success' })
  } catch (error) {
    console.error('开始处理工单失败:', error)
  } finally {
    busy.value = false
  }
}

const handleComplete = async () => {
  if (busy.value || !order.id) return
  if (!completeForm.result) {
    uni.showToast({ title: '请选择整体结果', icon: 'none' })
    return
  }

  busy.value = true
  try {
    const items = itemsForm.value.map((item) => ({
      name: item.name || '',
      result: item.result || '',
      remark: item.remark || ''
    }))
    const parts = partsForm.value
      .filter((part) => part.name && part.name.trim())
      .map((part) => ({
        name: part.name.trim(),
        qty: Number(part.qty) || 0,
        price: Number(part.price) || 0
      }))

    await completeMaintenanceOrder(order.id, {
      result: completeForm.result,
      items,
      attachments: attachments.value.slice(),
      parts,
      cost: partsCost.value,
      remark: completeForm.remark || ''
    })

    uni.showToast({ title: '提交成功', icon: 'success' })
    setTimeout(() => {
      uni.navigateBack()
    }, 600)
  } catch (error) {
    console.error('提交工单失败:', error)
  } finally {
    busy.value = false
  }
}

const handleCancel = () => {
  uni.showModal({
    title: '取消工单',
    content: '确认取消该维保工单吗？',
    editable: true,
    placeholderText: '请输入取消原因（选填）',
    success: async (res) => {
      if (!res.confirm) return
      busy.value = true
      try {
        const remark = (res.content || '').trim() || '用户取消工单'
        await cancelMaintenanceOrder(order.id, remark)
        uni.showToast({ title: '工单已取消', icon: 'success' })
        setTimeout(() => {
          uni.navigateBack()
        }, 600)
      } catch (error) {
        console.error('取消工单失败:', error)
      } finally {
        busy.value = false
      }
    }
  })
}

onLoad((options) => {
  orderId.value = options?.id || ''
  if (!orderId.value) {
    loading.value = false
    loadError.value = true
    return
  }
  loadDetail()
})

onPullDownRefresh(() => {
  loadDetail()
})
</script>

<style lang="scss">
@import '@/components/variables.scss';

page {
  background: $bg-secondary;
}

.page-maintenance-detail {
  min-height: 100vh;
}

.state-wrap {
  padding-top: 120rpx;
}

.detail-body {
  padding-top: $spacing-xs;
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
  font-size: 34rpx;
  font-weight: 700;
  color: $text-primary;
  line-height: 1.5;
  margin-bottom: $spacing-sm;
}

.info-list {
  padding-top: $spacing-sm;
  border-top: 1rpx solid $border-light;
}

.info-row {
  display: flex;
  align-items: flex-start;
  padding: $spacing-sm 0;

  .label {
    width: 140rpx;
    flex-shrink: 0;
    font-size: $font-size-sm;
    color: $text-placeholder;
  }

  .value {
    flex: 1;
    min-width: 0;
    font-size: $font-size-sm;
    color: $text-primary;
    line-height: 1.6;
    word-break: break-all;
  }

  .value-sub {
    font-size: $font-size-xs;
    color: $text-placeholder;
  }

  .overdue-text {
    color: $danger-color;
    font-weight: 600;
  }
}

.overdue-badge {
  display: inline-block;
  margin-left: $spacing-xs;
  padding: 2rpx 8rpx;
  font-size: $font-size-xs;
  color: #fff;
  background: $danger-color;
  border-radius: $radius-sm;
}

.empty-text {
  padding: $spacing-base 0;
  font-size: $font-size-sm;
  color: $text-placeholder;
  text-align: center;
}

.check-item {
  padding: $spacing-base 0;
  border-bottom: 1rpx solid $border-light;

  &:last-child {
    border-bottom: none;
  }

  .check-name {
    display: block;
    font-size: $font-size-base;
    font-weight: 500;
    color: $text-primary;
    margin-bottom: $spacing-sm;
  }
}

.result-options {
  display: flex;
  flex-wrap: wrap;
  gap: $spacing-sm;

  .result-option {
    padding: 10rpx 28rpx;
    font-size: $font-size-sm;
    color: $text-secondary;
    background: $bg-secondary;
    border: 2rpx solid transparent;
    border-radius: $radius-base;
    transition: $transition-base;

    &.active {
      color: $primary-color;
      background: $primary-light;
      border-color: $primary-color;
      font-weight: 600;
    }
  }

  &.overall {
    margin-bottom: $spacing-base;
  }
}

.remark-input {
  width: 100%;
  height: 68rpx;
  margin-top: $spacing-sm;
  padding: 0 $spacing-sm;
  font-size: $font-size-sm;
  color: $text-primary;
  background: $bg-secondary;
  border-radius: $radius-sm;
}

.check-readonly {
  display: flex;
  align-items: center;
  gap: $spacing-sm;

  .readonly-result {
    font-size: $font-size-sm;
    color: $primary-color;
  }

  .readonly-remark {
    flex: 1;
    font-size: $font-size-xs;
    color: $text-placeholder;
  }
}

.photo-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: $spacing-sm;
  padding-top: $spacing-xs;
}

.photo-item,
.photo-add {
  position: relative;
  width: 100%;
  height: 200rpx;
  border-radius: $radius-base;
  overflow: hidden;
  background: $bg-secondary;
}

.photo-item image {
  width: 100%;
  height: 100%;
}

.photo-remove {
  position: absolute;
  top: 8rpx;
  right: 8rpx;
  width: 40rpx;
  height: 40rpx;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
}

.photo-add {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: $spacing-xs;
  border: 2rpx dashed #d9e2ec;

  .photo-add-text {
    font-size: $font-size-xs;
    color: $text-placeholder;
  }
}

.part-item {
  padding: $spacing-sm 0;
  border-bottom: 1rpx solid $border-light;

  &:last-child {
    border-bottom: none;
  }

  .part-main {
    display: flex;
    align-items: center;
    gap: $spacing-sm;
  }

  .part-name {
    flex: 1;
    height: 68rpx;
    padding: 0 $spacing-sm;
    font-size: $font-size-sm;
    color: $text-primary;
    background: $bg-secondary;
    border-radius: $radius-sm;
  }

  .part-numbers {
    display: flex;
    align-items: center;
    gap: $spacing-base;
    margin-top: $spacing-sm;
  }

  .part-field {
    display: flex;
    align-items: center;
    gap: $spacing-xs;
  }

  .part-label {
    font-size: $font-size-xs;
    color: $text-placeholder;
  }

  .part-input {
    width: 120rpx;
    height: 60rpx;
    padding: 0 $spacing-xs;
    font-size: $font-size-sm;
    color: $text-primary;
    background: $bg-secondary;
    border-radius: $radius-sm;
    text-align: center;
  }

  .part-subtotal {
    margin-left: auto;
    font-size: $font-size-sm;
    color: $primary-color;
  }
}

.part-readonly {
  display: flex;
  align-items: center;
  gap: $spacing-base;
  padding: $spacing-sm 0;
  border-bottom: 1rpx solid $border-light;

  &:last-child {
    border-bottom: none;
  }

  .part-readonly-name {
    flex: 1;
    font-size: $font-size-sm;
    color: $text-primary;
  }

  .part-readonly-qty {
    font-size: $font-size-sm;
    color: $text-secondary;
  }

  .part-readonly-price {
    font-size: $font-size-sm;
    color: $primary-color;
  }
}

.add-part {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: $spacing-xs;
  padding: $spacing-sm;
  margin-top: $spacing-sm;
  font-size: $font-size-sm;
  color: $primary-color;
  background: $primary-light;
  border-radius: $radius-sm;
}

.cost-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: $spacing-base;
  margin-top: $spacing-xs;
  border-top: 1rpx solid $border-light;

  .cost-label {
    font-size: $font-size-sm;
    color: $text-secondary;
  }

  .cost-value {
    font-size: $font-size-lg;
    font-weight: 700;
    color: $danger-color;
  }
}

.section-label {
  font-size: $font-size-base;
  color: $text-primary;
  margin-bottom: $spacing-sm;

  .required {
    color: $danger-color;
    margin-left: 4rpx;
  }
}

.remark-textarea {
  width: 100%;
  height: 160rpx;
  padding: $spacing-sm;
  font-size: $font-size-sm;
  color: $text-primary;
  background: $bg-secondary;
  border-radius: $radius-sm;
  line-height: 1.6;
}

.input-placeholder {
  color: $text-placeholder;
}

.bottom-space {
  height: 180rpx;
}

.footer-bar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  gap: $spacing-sm;
  padding: $spacing-base;
  padding-bottom: calc(#{$spacing-base} + env(safe-area-inset-bottom));
  background: $bg-primary;
  box-shadow: 0 -2rpx 12rpx rgba(0, 0, 0, 0.06);

  .ds-button {
    flex: 1;
  }
}
</style>
