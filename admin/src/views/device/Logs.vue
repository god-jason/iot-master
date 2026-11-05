<template>
  <SmartTable
    card-title="设备日志"
    :loader="loadData"
    :columns="columns"
    :page-sizes="[20, 50, 100]"
    keyword-placeholder="搜索设备ID或日志消息"
    storage-key="device_logs"
  >
    <template #deviceName="{ row }">{{ row.device_name || '-' }}</template>
    <template #userName="{ row }">{{ row.user_name || '-' }}</template>
  </SmartTable>
</template>

<script setup>
import SmartTable from '@/components/SmartTable.vue'
import { searchDeviceLogs } from '@/api/device'

const columns = [
  { prop: 'device_id', label: '设备ID', minWidth: 180 },
  { label: '设备名称', minWidth: 140, slot: 'deviceName' },
  {
    prop: 'type',
    label: '类型',
    width: 100,
    type: 'tag',
    formatter: (row) => row.type || 'info'
  },
  { prop: 'message', label: '消息', minWidth: 240 },
  { label: '操作人', width: 120, slot: 'userName' },
  { prop: 'created', label: '时间', width: 180, type: 'datetime' }
]

const loadData = (params) => searchDeviceLogs(params)
</script>
