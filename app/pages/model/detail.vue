<template>
  <view class="page-model-detail">
    <DsNavBar :title="model?.name || '物模型详情'" />

    <view v-if="model">
    <view class="detail-header header-panel">
      <view class="header-cover">
        <image v-if="model.icon" :src="model.icon" mode="aspectFill" class="cover-image" />
        <view v-else class="cover-fallback">
          <DsIcon name="cpu" :size="36" color="#fff" />
        </view>
      </view>

      <view class="header-info">
        <text class="header-name">{{ model.name || '未命名物模型' }}</text>
        <view class="header-tags">
          <DsTag type="primary" size="small">{{ modelTypeText }}</DsTag>
          <DsTag v-if="model.protocol" type="success" size="small">{{ model.protocol.toUpperCase() }}</DsTag>
        </view>

        <view class="meta-list">
          <view class="meta-item">
            <text class="meta-label">物模型标识</text>
            <text class="meta-value">{{ model.id || '-' }}</text>
          </view>
          <view class="meta-item">
            <text class="meta-label">协议类型</text>
            <text class="meta-value">{{ model.protocol || '-' }}</text>
          </view>
          <view class="meta-item">
            <text class="meta-label">创建时间</text>
            <text class="meta-value">{{ formatDate(model.created) }}</text>
          </view>
        </view>
      </view>
    </view>

    <DsSection v-if="model" title="基础信息" type="card">
      <DsDescription>
        <DsDescriptionItem label="物模型名称" :value="model.name || '-'" />
        <DsDescriptionItem label="物模型类型" :value="modelTypeText" />
        <DsDescriptionItem label="协议类型" :value="model.protocol || '-'" />
        <DsDescriptionItem label="物模型标识" :value="model.id || '-'" />
        <DsDescriptionItem label="创建时间" :value="formatDate(model.created)" />
      </DsDescription>
    </DsSection>

    <DsSection v-if="model?.properties?.length > 0" title="属性" type="card">
      <DsTable :columns="propColumns" :data="model.properties" />
    </DsSection>

      <DsSection v-if="model.pollers?.length > 0 && hasProtocol" title="轮询器" type="card">
        <DsTable :columns="pollerColumns" :data="model.pollers" />
      </DsSection>

    <DsSection v-if="model?.events?.length > 0" title="事件" type="card">
      <DsTable :columns="eventColumns" :data="model.events" />
    </DsSection>

    <DsSection v-if="model?.services?.length > 0" title="服务" type="card">
      <DsTable :columns="serviceColumns" :data="model.services" />
    </DsSection>

    <DsSection v-if="model?.settings?.length > 0" title="配置" type="card">
      <DsTable :columns="configColumns" :data="model.settings" />
    </DsSection>

    <view v-if="model?.controls?.length > 0" class="controls-preview">
      <view
        v-for="control in model.controls"
        :key="control.id"
        class="control-item"
        :style="{ gridColumn: `span ${control.span || 1}` }"
      >
        <UserRender :control="control" :value="getControlValue(control)" @change="handleControlChange" @action="handleControlAction" />
      </view>
    </view>

    </view>

    <DsEmpty v-if="!model && !loading" text="加载失败" />
    <DsLoading v-if="loading" text="加载中..." />
  </view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { onPullDownRefresh } from '@dcloudio/uni-app'
import { getModel } from '@/api/model'
import UserRender from '@/controls/UserRender.vue'
import DsNavBar from '@/components/DsNavBar.vue'
import DsIcon from '@/components/DsIcon.vue'
import DsTag from '@/components/DsTag.vue'
import DsSection from '@/components/DsSection.vue'
import DsDescription from '@/components/DsDescription.vue'
import DsDescriptionItem from '@/components/DsDescriptionItem.vue'
import DsTable from '@/components/DsTable.vue'
import DsEmpty from '@/components/DsEmpty.vue'
import DsLoading from '@/components/DsLoading.vue'

const model = ref(null)
const loading = ref(true)
const controlValues = ref({})

const modelTypeText = computed(() => {
  const type = model.value?.type
  const map = {
    gateway: '网关',
    device: '设备',
    sensor: '传感器'
  }
  return map[type] || type || '-'
})

const hasProtocol = computed(() => !!model.value?.protocol)

const propColumns = computed(() => {
  const base = [
    { title: '名称', key: 'name' },
    { title: '标识', key: 'key' },
    { title: '类型', key: 'type' },
    { title: '单位', key: 'unit' },
    { title: '访问', key: 'mode' }
  ]
  if (hasProtocol.value) {
    base.push({ title: '寄存器', key: 'register' })
    base.push({ title: '格式', key: 'format' })
    base.push({ title: '倍率', key: 'rate' })
    base.push({ title: '地址', key: 'address' })
  }
  return base
})

const eventColumns = [
  { title: '名称', key: 'name' },
  { title: '标识', key: 'key' },
  { title: '说明', key: 'description' }
]

const serviceColumns = [
  { title: '名称', key: 'name' },
  { title: '标识', key: 'key' },
  { title: '说明', key: 'description' }
]

const configColumns = [
  { title: '名称', key: 'name' },
  { title: '标识', key: 'key' },
  { title: '类型', key: 'type' },
  { title: '默认值', key: 'value' }
]

const pollerColumns = [
  { title: '寄存器', key: 'register' },
  { title: '地址', key: 'address' },
  { title: '长度', key: 'length' }
]


const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  try {
    const date = new Date(dateStr)
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
  } catch {
    return dateStr
  }
}

const getControlValue = (control) => {
  if (control.binds?.value && Object.prototype.hasOwnProperty.call(controlValues.value, control.binds.value)) {
    return controlValues.value[control.binds.value]
  }
  if (control.type === 'switch') return false
  if (control.type === 'slider' || control.type === 'brightness' || control.type === 'color-temperature') {
    return control.min || 0
  }
  if (control.type === 'number') return 0
  return ''
}

const handleControlChange = (event) => {
  if (event.bindKey) {
    controlValues.value[event.bindKey] = event.value
  }
}

const handleControlAction = (event) => {
  console.log('Control action:', event)
}

const loadModel = async () => {
  loading.value = true
  try {
    const pages = getCurrentPages()
    const currentPage = pages[pages.length - 1]
    const id = currentPage.options.id
    const res = await getModel(id)
    model.value = res.data || res
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadModel()
})

onPullDownRefresh(async () => {
  await loadModel()
  uni.stopPullDownRefresh()
})
</script>

<style lang="scss">
@import '@/components/variables.scss';

page {
  background: $bg-secondary;
  padding-bottom: 40rpx;
}

.detail-header {
  display: flex;
  gap: $spacing-lg;
  flex-wrap: nowrap;
  align-items: stretch;
  margin: $spacing-base;
  margin-bottom: $spacing-base;
  padding: $spacing-lg;
}

.header-panel {
  background: $bg-primary;
  border-radius: $radius-lg;
  box-shadow: $shadow-sm;
}

.header-cover {
  width: 240rpx;
  min-width: 240rpx;
  height: 240rpx;
  border-radius: $radius-xl;
  overflow: hidden;
  background: linear-gradient(135deg, $primary-color 0%, $primary-dark 100%);
  box-shadow: 0 10rpx 24rpx rgba(91, 163, 232, 0.24);
}

.cover-image,
.cover-fallback {
  width: 100%;
  height: 100%;
}

.cover-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
}

.header-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
}

.header-name {
  font-size: 40rpx;
  font-weight: 700;
  color: $text-primary;
}

.header-tags {
  display: flex;
  flex-wrap: wrap;
  gap: $spacing-sm;
}

.header-desc {
  font-size: $font-size-sm;
  color: $text-secondary;
  line-height: 1.6;
}

.meta-list {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  margin-top: 4rpx;
}

.meta-item {
  display: flex;
  align-items: flex-start;
  gap: $spacing-sm;
}

.meta-label {
  width: 140rpx;
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

.controls-preview {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: $spacing-base;
  padding: $spacing-sm;
  background: $bg-secondary;
  border-radius: $radius-base;

  .control-item {
    background: $bg-primary;
    border-radius: $radius-base;
    padding: $spacing-base;
    box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.06);
  }
}

.preview-container {
  background: $bg-secondary;
  border-radius: $radius-base;
  overflow: hidden;

  .preview-img {
    width: 100%;
  }
}
</style>
