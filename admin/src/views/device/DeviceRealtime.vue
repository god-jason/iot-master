<template>
  <div class="realtime-grid">
    <div v-for="(item, key) in realtimeData" :key="key" class="realtime-item">
      <div class="item-header">
        <span class="item-label">{{ item.label || key }}</span>
        <el-tag size="small">{{ item.unit || '' }}</el-tag>
      </div>
      <div class="item-value">{{ item.value ?? '-' }}</div>
      <div class="item-time">{{ $formatDateTime(item.time) }}</div>
    </div>
    <el-empty v-if="!realtimeData || Object.keys(realtimeData).length === 0" description="暂无实时数据" />
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { getDeviceProperty } from '@/api/device'

const props = defineProps({
  deviceId: {
    type: String,
    required: true
  },
  properties: {
    type: Array,
    default: () => []
  }
})

const realtimeData = ref({})
let refreshTimer = null

const initRealtimeData = () => {
  realtimeData.value = {}
  props.properties.forEach(prop => {
    realtimeData.value[prop.key] = {
      label: prop.name,
      unit: prop.unit,
      value: null,
      time: null
    }
  })
}

const loadRealtimeData = async () => {
  try {
    const res = await getDeviceProperty(props.deviceId)
    if (res.data) {
      for (const key in res.data) {
        if (realtimeData.value[key]) {
          realtimeData.value[key].value = res.data[key].value
          realtimeData.value[key].time = res.data[key].time * 1000
        }
      }
    }
  } catch (error) {
    console.error('加载实时数据失败', error)
  }
}

watch(() => props.properties, () => {
  initRealtimeData()
  loadRealtimeData()
}, { immediate: true })

watch(() => props.deviceId, () => {
  loadRealtimeData()
})

onMounted(() => {
  loadRealtimeData()
  refreshTimer = setInterval(loadRealtimeData, 5000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style lang="scss" scoped>
.realtime-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;

  .realtime-item {
    background: #f5f7fa;
    border-radius: 8px;
    padding: 16px;
    text-align: center;

    .item-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 8px;

      .item-label {
        font-weight: 600;
        color: #303133;
      }
    }

    .item-value {
      font-size: 28px;
      font-weight: bold;
      color: #409EFF;
      margin-bottom: 4px;
    }

    .item-time {
      font-size: 12px;
      color: #909399;
    }
  }
}
</style>