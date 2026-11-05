<template>
  <div class="sub-device-list">
    <SmartTable
      ref="tableRef"
      card-title="子设备"
      :loader="loadDevices"
      :columns="columns"
      :operators="operators"
      :operator-width="150"
      keyword-placeholder="搜索子设备名称/SN"
      storage-key="sub-device"
    >
      <template #toolbar-right>
        <el-button type="success" :icon="Plus" @click="openCreateDrawer">添加子设备</el-button>
      </template>

      <template #icon="{ row }">
        <div class="icon-cell">
          <img v-if="row.model_icon" :src="getImageUrl(row.model_icon)" alt="图标" />
          <el-icon v-else :size="24"><Monitor /></el-icon>
        </div>
      </template>

      <template #sn="{ row }"><span class="mono">{{ row.sn || '-' }}</span></template>

      <template #state="{ row }">
        <el-tag v-if="row.disabled" type="info" size="small">已禁用</el-tag>
        <el-tag v-else :type="row.online ? 'success' : 'warning'" size="small">
          {{ row.online ? '在线' : (row.mounted ? '已挂载' : '未挂载') }}
        </el-tag>
      </template>

      <template #created="{ row }">{{ formatDateTime(row.created) || '-' }}</template>
    </SmartTable>

    <SubDeviceForm
      v-model:visible="drawerVisible"
      :model-value="currentEditData"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Monitor } from '@element-plus/icons-vue'
import SmartTable from '@/components/SmartTable.vue'
import { formatDateTime } from '@/utils/format'
import SubDeviceForm from './SubDeviceForm.vue'
import {
  listLinkDevices,
  createLinkDevice,
  updateLinkDevice,
  deleteLinkDevice
} from '@/api/gateway'

const props = defineProps({
  // 链路 ID（Device.LinkID，即注册包 SN）
  linkId: {
    type: String,
    required: true
  }
})

const tableRef = ref(null)

const baseUrl = import.meta.env.VITE_API_BASE_URL || ''

const getImageUrl = (url) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return baseUrl + url
}

const loadDevices = ({ page, page_size, keyword }) => {
  if (!props.linkId) return Promise.resolve({ data: [], total: 0 })
  return listLinkDevices(props.linkId, { page, page_size, keyword })
}

const columns = [
  { prop: 'model_icon', label: '图标', width: 70, align: 'center', slot: 'icon', showOverflowTooltip: false },
  { prop: 'name', label: '设备名称', minWidth: 120 },
  { prop: 'sn', label: '设备SN', minWidth: 150, slot: 'sn', showOverflowTooltip: false },
  { prop: 'model_name', label: '所属物模型', minWidth: 120 },
  { prop: 'address', label: '协议地址', width: 100 },
  { label: '状态', width: 100, slot: 'state', showOverflowTooltip: false },
  { prop: 'created', label: '创建时间', width: 180, slot: 'created', showOverflowTooltip: false }
]

const operators = [
  { key: 'edit', label: '编辑', type: 'primary', handler: (row) => openEditDrawer(row) },
  { key: 'delete', label: '删除', type: 'danger', handler: (row) => handleDelete(row) }
]

const drawerVisible = ref(false)
const currentEditId = ref(null)
const currentEditData = ref({})

const openCreateDrawer = () => {
  currentEditId.value = null
  currentEditData.value = {}
  drawerVisible.value = true
}

const openEditDrawer = (row) => {
  currentEditId.value = row.id
  currentEditData.value = row
  drawerVisible.value = true
}

const handleFormSubmit = async (formData) => {
  if (currentEditId.value) {
    await updateLinkDevice(props.linkId, currentEditId.value, formData)
  } else {
    await createLinkDevice(props.linkId, formData)
  }
  drawerVisible.value = false
  tableRef.value?.refresh()
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除设备"${row.name}"吗？`, '提示', {
      type: 'warning'
    })
    await deleteLinkDevice(props.linkId, row.id)
    ElMessage.success('删除成功')
    tableRef.value?.refresh()
  } catch (error) {
    if (error !== 'cancel') {
      console.error(error)
    }
  }
}
</script>

<style lang="scss" scoped>
.sub-device-list {
  .icon-cell {
    width: 40px;
    height: 40px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #f5f5f5;
    overflow: hidden;
    margin: 0 auto;

    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }

  .mono {
    font-family: 'Consolas', 'Menlo', monospace;
  }
}
</style>
