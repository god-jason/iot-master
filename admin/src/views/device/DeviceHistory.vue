<template>
  <div class="history-chart">
    <div class="chart-controls">
      <el-checkbox-group v-model="selectedProperties" @change="loadHistoryChart">
        <el-checkbox v-for="prop in properties" :key="prop.key" :label="prop.key">
          {{ prop.name }}
        </el-checkbox>
      </el-checkbox-group>
      <el-date-picker v-model="historyTimeRange" type="datetimerange" range-separator="至" start-placeholder="开始时间"
        end-placeholder="结束时间" :shortcuts="timeShortcuts" @change="loadHistoryChart" />
      <el-button type="primary" @click="loadHistoryChart">查询</el-button>
    </div>
    <div class="chart-container" ref="chartRef"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, onUnmounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import { getDeviceHistory } from '@/api/device'

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

const chartRef = ref(null)
const chartInstance = ref(null)
const selectedProperties = ref([])
const historyTimeRange = ref([
  new Date(Date.now() - 3600000 * 24),
  new Date()
])
const timeShortcuts = [
  { text: '最近1小时', value: () => [new Date(Date.now() - 3600000), new Date()] },
  { text: '最近6小时', value: () => [new Date(Date.now() - 21600000), new Date()] },
  { text: '最近24小时', value: () => [new Date(Date.now() - 86400000), new Date()] },
  { text: '最近7天', value: () => [new Date(Date.now() - 604800000), new Date()] }
]

const loadHistoryChart = async () => {
  if (!chartInstance.value) {
    await nextTick()
    chartInstance.value = echarts.init(chartRef.value)
  }

  if (selectedProperties.value.length === 0) {
    //默认选择所有属性
    chartInstance.value.setOption({ series: [] })
    return
  }


  try {
    const startTime = historyTimeRange.value[0]?.toISOString() || new Date(Date.now() - 3600000).toISOString()
    const endTime = historyTimeRange.value[1]?.toISOString() || new Date().toISOString()

    const res = await getDeviceHistory(props.deviceId, {
      //fields: selectedProperties.value.join(','),
      start: startTime,
      end: endTime,
      window: "1s" // 1秒窗口，尽可能获取更多数据
    })

    const series = selectedProperties.value.map(propKey => {
      const prop = props.properties.find(p => p.key === propKey)
      const points = (res.data?.[propKey] || []).map(item => [item.time * 1000, item.value])
      return {
        name: prop?.name || propKey,
        type: 'line',
        symbol: 'none',
        data: points,
        smooth: true
      }
    })

    chartInstance.value.setOption({
      tooltip: { trigger: 'axis' },
      legend: { data: series.map(s => s.name), top: 10 },
      toolbox: {
        feature: {
          dataZoom: {
            yAxisIndex: 'none'
          },
          restore: {},
          saveAsImage: {}
        }
      },
      xAxis: { type: 'time', boundaryGap: false },
      yAxis: { type: 'value', boundaryGap: [0, '100%'] },
      dataZoom: [
        { type: 'inside', start: 20, end: 100 },
        { start: 0, end: 20 }
      ],
      series
    })

    console.log(chartInstance.value.getOption())
  } catch (error) {
    console.error(error)
    ElMessage.error('加载历史数据失败')
  }
}

watch(() => props.properties, (newProps) => {
  selectedProperties.value = newProps.map(p => p.key)
}, { immediate: true })

onMounted(async () => {
  await nextTick()
  if (chartRef.value) {
    chartInstance.value = echarts.init(chartRef.value)
    if (selectedProperties.value.length > 0) {
      loadHistoryChart()
    }
  }
})

onUnmounted(() => {
  if (chartInstance.value) {
    chartInstance.value.dispose()
  }
})
</script>

<style lang="scss" scoped>
.history-chart {
  .chart-controls {
    display: flex;
    gap: 16px;
    align-items: center;
    margin-bottom: 16px;
  }

  .chart-container {
    height: 400px;
  }
}
</style>