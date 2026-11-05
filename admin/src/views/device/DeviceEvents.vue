<template>
  <SmartTable
    ref="tableRef"
    :card-title="deviceId ? '' : '设备事件'"
    :loader="loadEvents"
    :columns="columns"
    :operators="operators"
    :filter="filters"
    keyword-placeholder="搜索事件描述"
    :operator-width="120"
    pagination-layout="total, prev, pager, next"
  >
    <template #toolbar-left>
      <el-select v-model="searchType" placeholder="事件类型" clearable style="width: 150px">
        <el-option label="全部" value="" />
        <el-option label="上线" value="online" />
        <el-option label="离线" value="offline" />
        <el-option label="告警" value="alert" />
        <el-option label="数据" value="data" />
        <el-option label="配置" value="config" />
      </el-select>
      <el-select v-model="searchStatus" placeholder="处理状态" clearable style="width: 150px">
        <el-option label="全部" value="" />
        <el-option label="未处理" :value="0" />
        <el-option label="已处理" :value="1" />
      </el-select>
    </template>
  </SmartTable>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import SmartTable from '@/components/SmartTable.vue'
import { getDeviceEventList, deleteDeviceEvent, handleDeviceEvent } from '@/api/device'

const props = defineProps({
  deviceId: {
    type: String,
    default: ''
  }
})

const tableRef = ref(null)
const searchType = ref('')
const searchStatus = ref('')

const typeLabels = {
  online: '上线',
  offline: '离线',
  alert: '告警',
  data: '数据',
  config: '配置'
}

const typeTagTypes = {
  online: 'success',
  offline: 'danger',
  alert: 'warning',
  data: 'info',
  config: 'primary'
}

const filters = computed(() => {
  const list = []
  if (searchType.value !== '' && searchType.value !== undefined) {
    list.push({ field: 'type', operator: 'eq', value: searchType.value })
  }
  if (searchStatus.value !== '' && searchStatus.value !== undefined) {
    list.push({ field: 'status', operator: 'eq', value: searchStatus.value })
  }
  if (props.deviceId) {
    list.push({ field: 'device_id', operator: 'eq', value: props.deviceId })
  }
  return list
})

const columns = [
  { prop: 'device_name', label: '设备名称', width: 120 },
  { prop: 'device_sn', label: '设备SN', width: 150 },
  {
    prop: 'type',
    label: '类型',
    width: 80,
    type: 'tag',
    formatter: (row) => typeLabels[row.type] || row.type,
    tagType: (row) => typeTagTypes[row.type] || 'info'
  },
  { prop: 'code', label: '代码', width: 100 },
  { prop: 'message', label: '事件描述', minWidth: 200 },
  {
    prop: 'status',
    label: '状态',
    width: 80,
    type: 'tag',
    formatter: (row) => (row.status === 1 ? '已处理' : '未处理'),
    tagType: (row) => (row.status === 1 ? 'success' : 'warning')
  },
  { prop: 'created', label: '时间', width: 180, type: 'datetime' }
]

const operators = computed(() => {
  if (props.deviceId) return []
  return [
    {
      key: 'handle',
      label: '处理',
      type: 'primary',
      show: (row) => row.status === 0,
      handler: (row) => handleEvent(row)
    },
    { key: 'delete', label: '删除', type: 'danger', handler: (row) => deleteEvent(row) }
  ]
})

const loadEvents = (params) => getDeviceEventList(params)

const deleteEvent = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该事件吗？', '提示', {
      type: 'warning'
    })
    await deleteDeviceEvent(row.id)
    ElMessage.success('删除成功')
    tableRef.value?.refresh()
  } catch (e) {
    if (e !== 'cancel') console.error(e)
  }
}

const handleEvent = async (row) => {
  try {
    await handleDeviceEvent(row.id)
    ElMessage.success('处理成功')
    tableRef.value?.refresh()
  } catch (e) {
    console.error(e)
    ElMessage.error('处理失败')
  }
}
</script>
