<template>
  <div class="gateway-connections-page">
    <SmartTable
      ref="tableRef"
      :card-title="isServerScope ? '连接列表' : '连接'"
      :loader="loader"
      :columns="columns"
      :operators="operators"
      row-key="conn_id"
      :show-pagination="false"
      :show-keyword-search="false"
      :operator-width="100"
    >
      <template #toolbar-left>
        <el-button v-if="isServerScope" link type="primary" @click="router.push('/gateway/servers')">← 返回服务器</el-button>
        <el-tag v-if="isServerScope && serverRunning" type="success" size="small">Server#{{ serverId }} 运行中</el-tag>
        <el-tag v-else-if="isServerScope && !serverRunning" type="warning" size="small">Server#{{ serverId }} 未运行</el-tag>
      </template>

      <template #link="{ row }"><span class="mono">{{ row.link_id || '-' }}</span></template>
      <template #imei="{ row }"><span class="mono">{{ row.imei || '-' }}</span></template>
      <template #devices="{ row }">
        <el-tag v-if="row.devices?.length" size="small" type="info">{{ row.devices.length }} 个</el-tag>
        <span v-else class="muted">无</span>
      </template>
      <template #remote="{ row }"><span class="mono">{{ row.remote }}</span></template>
      <template #poll="{ row }">{{ row.poll?.interval || '未配置' }}</template>
    </SmartTable>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import SmartTable from '@/components/SmartTable.vue'
import { listGatewayConnections } from '@/api/gateway'

const route = useRoute()
const router = useRouter()
// 支持两种模式：
//   1. /gateway/servers/:id/connections  → 带 serverId 过滤
//   2. /gateway/connections              → 不带，返回所有连接
const serverId = computed(() => {
  const v = route.params.id
  return v ? Number(v) : null
})
const isServerScope = computed(() => serverId.value != null && !isNaN(serverId.value))

const tableRef = ref(null)
const serverRunning = ref(true)
let timer = null

const loader = async () => {
  try {
    const res = await listGatewayConnections(isServerScope.value ? serverId.value : undefined)
    serverRunning.value = true
    return res
  } catch (e) {
    serverRunning.value = isServerScope.value ? false : true
    return { data: { items: [] } }
  }
}

const columns = [
  { prop: 'conn_id', label: 'ConnID', width: 80 },
  { label: '链路 ID', minWidth: 160, slot: 'link', showOverflowTooltip: false },
  { label: 'IMEI', minWidth: 140, slot: 'imei', showOverflowTooltip: false },
  { label: '设备', width: 90, slot: 'devices', showOverflowTooltip: false },
  { prop: 'protocol', label: '协议', width: 90 },
  { label: '远端', minWidth: 140, slot: 'remote', showOverflowTooltip: false },
  {
    prop: 'registered',
    label: '状态',
    width: 90,
    type: 'tag',
    formatter: (row) => (row.registered ? '已注册' : '未注册'),
    tagType: (row) => (row.registered ? 'success' : 'warning')
  },
  { label: '兜底轮询', width: 100, slot: 'poll', showOverflowTooltip: false }
]

const operators = [
  { key: 'detail', label: '详情', type: 'primary', handler: (row) => goDetail(row) }
]

function goDetail(row) {
  router.push(`/gateway/servers/${row.server_id || serverId.value}/connections/${row.conn_id}`)
}

onMounted(() => {
  timer = setInterval(() => tableRef.value?.refresh(), 10000)
})
onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style lang="scss" scoped>
.gateway-connections-page {
  .mono { font-family: 'Consolas', 'Menlo', monospace; }
  .muted { color: #999; font-size: 12px; }
}
</style>
