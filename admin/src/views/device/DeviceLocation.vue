<template>
  <div class="location-content">
    <div class="location-actions">
      <el-button type="primary" @click="getLatestLocation">获取最新定位</el-button>
      <el-button @click="showHistoryTrack">查看历史轨迹</el-button>
    </div>
    <div class="map-container" ref="mapRef"></div>
    <div class="location-info" v-if="deviceLocation">
      <span>当前位置：{{ deviceLocation.lat }}°, {{ deviceLocation.lng }}°</span>
      <span>更新时间：{{ $formatDateTime(deviceLocation.time) }}</span>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'

defineProps({
  deviceId: {
    type: String,
    required: true
  }
})

const mapRef = ref(null)
const mapInstance = ref(null)
const deviceLocation = ref(null)
const showTrack = ref(false)

const initMap = async () => {
  await nextTick()
}

const getLatestLocation = async () => {
  try {
    ElMessage.info('获取定位功能待实现')
  } catch (error) {
    console.error(error)
  }
}

const showHistoryTrack = () => {
  showTrack.value = !showTrack.value
  ElMessage.info(showTrack.value ? '显示历史轨迹' : '隐藏历史轨迹')
}

onMounted(() => {
  initMap()
})
</script>

<style lang="scss" scoped>
.location-content {
  .location-actions {
    margin-bottom: 16px;
  }

  .map-container {
    height: 400px;
    background: #f5f7fa;
    border-radius: 8px;
  }

  .location-info {
    margin-top: 12px;
    display: flex;
    gap: 24px;
    color: #606266;
  }
}
</style>