<template>
  <div class="connection-detail">
    <el-card class="info-card">
      <div class="info-header">
        <div class="info-left">
          <el-button link type="primary" @click="router.push(`/gateway/servers/${serverId}/connections`)">← 返回连接列表</el-button>
          <h2 class="mt-8">{{ detail.sn || detail.link_id || '连接 #' + connId }}</h2>
          <div class="info-meta">
            <el-tag :type="detail.registered ? 'success' : 'warning'">{{ detail.registered ? '已注册' : '未注册' }}</el-tag>
            <span class="mono">ConnID: {{ detail.conn_id }}</span>
            <span class="mono">链路: {{ detail.link_id || '-' }}</span>
            <span class="mono">{{ detail.remote }}</span>
            <el-tag>{{ detail.protocol?.toUpperCase() || 'TCP' }}</el-tag>
          </div>
        </div>
        <div class="info-right">
          <el-button :icon="Refresh" @click="load">刷新</el-button>
        </div>
      </div>
    </el-card>

    <el-row :gutter="20" class="main-row">
      <!-- 左侧：兜底轮询配置 -->
      <el-col :span="8">
        <el-card>
          <template #header>
            <div class="section-title"><el-icon><Timer /></el-icon><span>兜底轮询配置</span></div>
          </template>
          <el-form :model="pollForm" label-width="100px">
            <el-form-item label="启用兜底">
              <el-switch v-model="pollForm.enabled" />
            </el-form-item>
            <el-form-item v-if="pollForm.enabled" label="间隔">
              <el-input v-model="pollForm.interval" placeholder="如 5s" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="savePoll">保存</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <!-- 右侧：链路设备管理 -->
      <el-col :span="16">
        <el-card>
          <template #header>
            <div class="section-title">
              <el-icon><Cpu /></el-icon><span>链路设备 ({{ devices.length }})</span>
              <el-button type="primary" link class="add-btn" :icon="Plus" @click="openDeviceDialog()">添加设备</el-button>
            </div>
          </template>
          <el-table :data="devices" v-loading="deviceLoading" stripe size="small" max-height="420">
            <el-table-column prop="name" label="名称" min-width="120" />
            <el-table-column prop="sn" label="SN" min-width="140">
              <template #default="{ row }"><span class="mono">{{ row.sn || '-' }}</span></template>
            </el-table-column>
            <el-table-column prop="model_name" label="物模型" min-width="110">
              <template #default="{ row }">{{ row.model_name || '-' }}</template>
            </el-table-column>
            <el-table-column prop="address" label="地址" width="90" />
            <el-table-column label="状态" width="130">
              <template #default="{ row }">
                <el-tag v-if="row.disabled" type="info" size="small">已禁用</el-tag>
                <el-tag v-else :type="row.online ? 'success' : 'warning'" size="small">
                  {{ row.online ? '在线' : (row.mounted ? '已挂载' : '未挂载') }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button link size="small" @click="openDeviceDialog(row)">编辑</el-button>
                <el-popconfirm title="删除设备？" @confirm="delDevice(row)">
                  <template #reference><el-button link type="danger" size="small">删除</el-button></template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- 设备 Dialog -->
    <el-dialog v-model="deviceDialogVisible" :title="editingDevice ? '编辑设备' : '添加设备'" width="480px">
      <el-form :model="deviceForm" label-width="90px">
        <el-form-item label="名称"><el-input v-model="deviceForm.name" placeholder="设备名称" /></el-form-item>
        <el-form-item label="SN"><el-input v-model="deviceForm.sn" placeholder="留空自动生成" /></el-form-item>
        <el-form-item label="物模型" required>
          <el-select v-model="deviceForm.model_id" placeholder="请选择物模型" filterable style="width: 100%">
            <el-option v-for="m in models" :key="m.id" :label="m.name" :value="m.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="协议地址" required>
          <el-input v-model="deviceForm.address" placeholder="如 1" />
        </el-form-item>
        <el-form-item label="禁用">
          <el-switch v-model="deviceForm.disabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="deviceDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveDevice">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Timer, Cpu, Plus } from '@element-plus/icons-vue'
import {
  getGatewayConnection, setGatewayPoll,
  listLinkDevices, createLinkDevice, updateLinkDevice, deleteLinkDevice
} from '@/api/gateway'
import { searchModels } from '@/api/model'

const route = useRoute()
const router = useRouter()
const serverId = Number(route.params.serverId || route.params.server_id)
const connId = Number(route.params.connId || route.params.conn_id)

const detail = ref({})
const devices = ref([])
const deviceLoading = ref(false)
const models = ref([])
let timer = null

async function load() {
  try {
    const res = await getGatewayConnection(serverId, connId)
    detail.value = res.data || res
    if (detail.value.poll?.interval) {
      pollForm.enabled = true
      pollForm.interval = detail.value.poll.interval
    }
  } catch (e) { /* ignore */ }
  loadDevices()
}

async function loadModels() {
  try {
    const res = await searchModels({ page: 1, page_size: 100 })
    models.value = res.data || []
  } catch (e) { console.error(e) }
}

async function loadDevices() {
  if (!detail.value.link_id) {
    devices.value = []
    return
  }
  deviceLoading.value = true
  try {
    const res = await listLinkDevices(detail.value.link_id, { page: 1, page_size: 100 })
    devices.value = res.data?.items || res.items || []
  } catch (e) { /* ignore */ }
  finally { deviceLoading.value = false }
}

// ===== poll =====
const pollForm = reactive({ enabled: false, interval: '5s' })
async function savePoll() {
  try {
    await setGatewayPoll(serverId, connId, pollForm.enabled ? pollForm : { interval: '' })
    ElMessage.success('保存成功')
    load()
  } catch (e) { ElMessage.error('保存失败') }
}

// ===== device =====
const deviceDialogVisible = ref(false)
const editingDevice = ref(null)
const deviceForm = reactive({
  name: '', sn: '', model_id: '', address: '', disabled: false
})

function openDeviceDialog(row) {
  editingDevice.value = row || null
  if (row) {
    Object.assign(deviceForm, {
      name: row.name || '',
      sn: row.sn || '',
      model_id: row.model_id || '',
      address: row.address || '',
      disabled: !!row.disabled
    })
  } else {
    Object.assign(deviceForm, { name: '', sn: '', model_id: '', address: '', disabled: false })
  }
  deviceDialogVisible.value = true
}

async function saveDevice() {
  if (!deviceForm.model_id) { ElMessage.warning('请选择物模型'); return }
  if (!deviceForm.address) { ElMessage.warning('请填写协议地址'); return }
  try {
    const linkId = detail.value.link_id
    const body = { ...deviceForm, address: String(deviceForm.address) }
    if (editingDevice.value) {
      await updateLinkDevice(linkId, editingDevice.value.id, body)
    } else {
      await createLinkDevice(linkId, body)
    }
    deviceDialogVisible.value = false
    ElMessage.success('保存成功')
    loadDevices()
  } catch (e) { ElMessage.error('保存失败') }
}

function delDevice(row) {
  deleteLinkDevice(detail.value.link_id, row.id).then(() => {
    ElMessage.success('已删除')
    loadDevices()
  })
}

onMounted(() => { loadModels(); load(); timer = setInterval(load, 10000) })
onUnmounted(() => { if (timer) clearInterval(timer) })
</script>

<style lang="scss" scoped>
.connection-detail {
  .info-card { margin-bottom: 20px; }
  .info-header { display: flex; justify-content: space-between; }
  .info-left h2 { margin: 8px 0; font-size: 18px; }
  .info-meta { display: flex; gap: 10px; color: #666; font-size: 13px; align-items: center; }
  .info-right { display: flex; gap: 8px; align-self: flex-start; }
  .mt-8 { margin-top: 8px; }
  .section-title { display: flex; align-items: center; gap: 6px; font-weight: 600; font-size: 14px; }
  .add-btn { margin-left: auto; }
  .main-row { margin-top: 20px; }
  .mono { font-family: 'Consolas', 'Menlo', monospace; }
}
</style>
