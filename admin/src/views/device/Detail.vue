<template>
  <div class="device-detail-page">
    <PageHeader title="设备详情" :showBack="true">
      <template #extra>
        <el-button type="primary" @click="openEditDrawer">
          <el-icon>
            <Edit />
          </el-icon> 编辑
        </el-button>
      </template>
    </PageHeader>

    <el-card class="info-card" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>基本信息</span>
        </div>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="设备名称">{{ device.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="设备SN">{{ device.sn || '-' }}</el-descriptions-item>
        <el-descriptions-item label="所属链路">{{ device.link_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="所属物模型">
          <router-link v-if="device.model_id" :to="`/models/${device.model_id}`" class="model-link">
            {{ device.model_name || '-' }}
          </router-link>
          <span v-else>{{ '-' }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="位置">{{ device.location || '-' }}</el-descriptions-item>
        <el-descriptions-item label="固件版本">{{ device.firmware || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ $formatDateTime(device.created) }}</el-descriptions-item>
        <el-descriptions-item v-for="item in statusInfo" :key="item.key" :label="item.label">
          {{ item.value }}
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span>所属组织</span>
          <el-button size="small" type="primary" :icon="Plus" @click="openOrgDialog">绑定组织</el-button>
        </div>
      </template>
      <div class="org-binding">
        <el-tag
          v-for="item in boundOrgs"
          :key="item.id"
          closable
          type="info"
          class="org-tag"
          @close="handleUnbindOrg(item)"
        >
          {{ item.name }}
        </el-tag>
        <span v-if="boundOrgs.length === 0" class="org-empty">暂未绑定任何组织</span>
      </div>
      <div class="org-hint">一台设备可绑定多个组织；管理员可分配到任意组织，组织成员只能分配到本组织及子级组织。</div>
    </el-card>

    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="实时数据" name="realtime" lazy>
        <DeviceRealtime :device-id="deviceId" :properties="modelProperties" />
      </el-tab-pane>

      <el-tab-pane label="历史曲线" name="history" lazy>
        <DeviceHistory :device-id="deviceId" :properties="modelProperties" />
      </el-tab-pane>

      <el-tab-pane v-if="device.model_type !== 'sub'" label="远程操作" name="services" lazy>
        <DeviceServices :device-id="deviceId" :services="modelServices" />
      </el-tab-pane>

      <el-tab-pane v-if="device.model_type !== 'sub'" label="远程配置" name="settings" lazy>
        <DeviceSettings :device-id="deviceId" :settings="modelSettings" />
      </el-tab-pane>

      <el-tab-pane label="事件记录" name="events" lazy>
        <DeviceEvents :device-id="deviceId" />
      </el-tab-pane>

      <el-tab-pane label="设备日志" name="logs" lazy>
        <DeviceLogs :device-id="deviceId" />
      </el-tab-pane>

      <el-tab-pane v-if="device.model_type === 'gateway'" label="子设备管理" name="subdevices" lazy>
        <SubDevice v-if="subLinkId" :link-id="subLinkId" />
        <el-empty v-else description="该网关设备未配置SN或所属链路，暂无子设备可管理" />
      </el-tab-pane>

      <el-tab-pane v-else-if="device.link_id" label="链路设备" name="linkdevices" lazy>
        <SubDevice :link-id="device.link_id" />
      </el-tab-pane>

      <el-tab-pane v-if="device.model_type !== 'sub'" label="地图定位" name="location" lazy>
        <DeviceLocation :device-id="deviceId" />
      </el-tab-pane>
    </el-tabs>

    <DeviceForm
      v-model:visible="editDrawerVisible"
      :model-value="device"
      @submit="handleEditSubmit"
    />

    <el-dialog v-model="orgDialogVisible" title="绑定设备到组织" width="480px">
      <el-form label-width="90px">
        <el-form-item label="目标组织">
          <el-tree-select
            v-model="selectedOrgId"
            :data="orgOptions"
            check-strictly
            :render-after-expand="false"
            placeholder="请选择组织"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="orgDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="orgSubmitting" @click="handleBindOrg">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Edit, Refresh, Plus } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import { getDevice, updateDevice } from '@/api/device'
import { getOrgTree, bindDeviceOrg, unbindDeviceOrg } from '@/api/org'
import { getModel } from '@/api/model'
import DeviceRealtime from './DeviceRealtime.vue'
import DeviceHistory from './DeviceHistory.vue'
import DeviceServices from './DeviceServices.vue'
import DeviceEvents from './DeviceEvents.vue'
import DeviceLogs from './DeviceLogs.vue'
import DeviceLocation from './DeviceLocation.vue'
import DeviceList from './List.vue'
import SubDevice from '@/views/sub-device/SubDevice.vue'
import DeviceForm from './DeviceForm.vue'
import DeviceSettings from './DeviceSettings.vue'
import { formatDateTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const deviceId = ref(route.params.id)

//console.log(id, deviceId)

const loading = ref(false)
const device = ref({})
const activeTab = ref('realtime')

const modelProperties = ref([])
const modelServices = ref([])
const modelEvents = ref([])
const modelSettings = ref([])
const statusInfo = ref([])

const loadDevice = async () => {
  loading.value = true
  try {
    const res = await getDevice(deviceId.value)
    device.value = res.data || {}

    if (device.value.model_id) {
      const modelRes = await getModel(device.value.model_id)
      const model = modelRes.data || {}
      modelProperties.value = model.properties || []
      modelServices.value = model.services || []
      modelEvents.value = model.events || []
      modelSettings.value = model.settings || []
    }

    // 从设备信息中获取状态数据
    if (device.value.status_data) {
      buildStatusInfo(device.value.status_data)
    }
  } catch (error) {
    console.error(error)
    ElMessage.error('加载设备信息失败')
  } finally {
    loading.value = false
  }
}

const buildStatusInfo = (data) => {
  const statusMap = {
    'status': { label: '状态', formatter: (v) => v === '1' || v === 1 ? '在线' : '离线' },
    'online_time': { label: '上线时间', formatter: (v) => v ? formatDateTime(Number(v)) : '-' },
    'offline_time': { label: '离线时间', formatter: (v) => v ? formatDateTime(Number(v)) : '-' },
    'last_time': { label: '最后更新', formatter: (v) => v ? formatDateTime(Number(v)) : '-' },
    'error': { label: '错误信息', formatter: (v) => v || '-' },
    'longitude': { label: '经度', formatter: (v) => v },
    'latitude': { label: '纬度', formatter: (v) => v },
    'csq': { label: '信号强度', formatter: (v) => v },
    'firmware': { label: '固件版本', formatter: (v) => v || '-' },
    'version': { label: '软件版本', formatter: (v) => v || '-' },
  }

  const info = []
  for (const key in statusMap) {
    if (data[key] !== undefined && data[key] !== null && data[key] !== '') {
      info.push({
        key,
        label: statusMap[key].label,
        value: statusMap[key].formatter(data[key])
      })
    }
  }
  statusInfo.value = info
}

const editDrawerVisible = ref(false)

const orgDialogVisible = ref(false)
const orgSubmitting = ref(false)
const selectedOrgId = ref(null)
const orgTree = ref([])

const boundOrgs = computed(() => {
  const ids = device.value.org_ids || []
  const names = device.value.org_names || []
  return ids.map((id, index) => ({ id, name: names[index] || `组织#${id}` }))
})

// 网关子设备挂载在链路下（Link ID = 注册包 SN）：优先取设备自身的 link_id，否则以网关 SN 作为链路 ID
const subLinkId = computed(() => device.value.link_id || device.value.sn || '')

const orgOptions = computed(() => {
  const convert = (nodes) =>
    (nodes || []).map((n) => ({ value: n.id, label: n.name, children: convert(n.children) }))
  return convert(orgTree.value)
})

const loadOrgTree = async () => {
  try {
    const res = await getOrgTree()
    orgTree.value = res.data || []
  } catch (error) {
    console.error('加载组织树失败', error)
  }
}

const openOrgDialog = () => {
  selectedOrgId.value = null
  orgDialogVisible.value = true
}

const handleBindOrg = async () => {
  if (!selectedOrgId.value) {
    ElMessage.warning('请选择组织')
    return
  }
  orgSubmitting.value = true
  try {
    await bindDeviceOrg(deviceId.value, selectedOrgId.value)
    ElMessage.success('已绑定组织')
    orgDialogVisible.value = false
    await loadDevice()
  } catch (error) {
    console.error(error)
    ElMessage.error(error.response?.data?.message || '绑定失败')
  } finally {
    orgSubmitting.value = false
  }
}

const handleUnbindOrg = async (item) => {
  try {
    await ElMessageBox.confirm(`确定将设备从组织「${item.name}」解绑吗？`, '解绑确认', {
      type: 'warning'
    })
    await unbindDeviceOrg(deviceId.value, item.id)
    ElMessage.success('已解绑')
    await loadDevice()
  } catch (error) {
    if (error !== 'cancel') console.error(error)
  }
}

const openEditDrawer = () => {
  editDrawerVisible.value = true
}

const handleEditSubmit = async (formData) => {
  await updateDevice(deviceId.value, formData)
  ElMessage.success('更新成功')
  editDrawerVisible.value = false
  loadDevice()
}

onMounted(() => {
  loadDevice()
  loadOrgTree()
})
</script>

<style lang="scss" scoped>
.device-detail-page {
  .info-card {
    margin-bottom: 20px;
  }

  .status-card {
    margin-bottom: 20px;
  }

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .org-binding {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
  }

  .org-empty {
    color: #909399;
    font-size: 13px;
  }

  .org-hint {
    margin-top: 10px;
    color: #909399;
    font-size: 12px;
  }

  .detail-tabs {
    background: #fff;
    border-radius: 8px;
    padding: 20px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
  }

  .model-link {
    color: #409eff;
    text-decoration: none;

    &:hover {
      text-decoration: underline;
    }
  }
}
</style>