<template>
  <SmartTable
    card-title="设备日志"
    :loader="loadDeviceLogs"
    :columns="columns"
    :filter="filters"
    :show-keyword-search="false"
    pagination-layout="total, prev, pager, next"
  />
</template>

<script setup>
import { computed } from 'vue'
import SmartTable from '@/components/SmartTable.vue'
import { searchDeviceLogs } from '@/api/device'

const props = defineProps({
  deviceId: {
    type: String,
    required: true
  }
})

const logLevelLabels = {
  debug: '调试',
  info: '信息',
  warn: '警告',
  error: '错误'
}

const logLevelTagTypes = {
  debug: 'info',
  info: 'success',
  warn: 'warning',
  error: 'danger'
}

const filters = computed(() => [
  { field: 'device_id', operator: 'eq', value: props.deviceId }
])

const columns = [
  { prop: 'user_name', label: '用户名', width: 100 },
  { prop: 'device_name', label: '设备名称', width: 120 },
  { prop: 'device_sn', label: '设备SN', width: 150 },
  {
    prop: 'level',
    label: '级别',
    width: 80,
    type: 'tag',
    formatter: (row) => logLevelLabels[row.level] || row.level,
    tagType: (row) => logLevelTagTypes[row.level] || 'info'
  },
  { prop: 'message', label: '日志消息', minWidth: 300 },
  { prop: 'created', label: '时间', width: 180, type: 'datetime' }
]

const loadDeviceLogs = (params) => searchDeviceLogs(params)
</script>
