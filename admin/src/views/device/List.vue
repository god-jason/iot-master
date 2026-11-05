<template>
  <div class="device-list">
    <div class="content-wrapper">
      <div v-if="showGroup" class="sidebar" :class="{ 'is-collapsed': groupCollapsed }">
        <div v-show="!groupCollapsed" class="sidebar-body">
          <GroupTree ref="groupTreeRef" @check-change="handleGroupCheck" />
        </div>
        <div v-if="groupCollapsed" class="sidebar-rail" title="展开分组" @click="toggleGroupSidebar">
          <el-icon><Expand /></el-icon>
          <span class="rail-label">设备分组</span>
        </div>
        <el-button
          v-if="!groupCollapsed"
          class="sidebar-toggle"
          text
          title="收起分组"
          @click="toggleGroupSidebar"
        >
          <el-icon><Fold /></el-icon>
        </el-button>
      </div>

      <div class="main-content">
        <SmartTable
          ref="tableRef"
          :card-title="showGroup ? '我的设备' : '设备列表'"
          :loader="searchDevices"
          :columns="columns"
          :operators="operators"
          :filter="filters"
          :show-selection="true"
          :show-batch-delete="true"
          :delete-fn="batchDeleteDevices"
          :operator-width="180"
          storage-key="device"
          keyword-placeholder="搜索设备名称/SN"
          row-clickable
          @row-click="handleRowClick"
        >
          <template #toolbar-left>
            <el-select
              v-if="showGroup"
              v-model="filterModelId"
              placeholder="物模型"
              clearable
              style="width: 150px"
            >
              <el-option v-for="item in myModels" :key="item.id" :label="item.name" :value="item.id" />
            </el-select>
          </template>

          <template #toolbar-right>
            <el-button type="primary" :icon="Plus" @click="openCreateDrawer">新建设备</el-button>
            <el-button :icon="Upload" @click="$router.push('/devices/import')">批量导入</el-button>
          </template>

          <template #icon="{ row }">
            <div class="icon-cell">
              <img v-if="row.model_icon" :src="getImageUrl(row.model_icon)" />
              <el-icon v-else :size="22"><Monitor /></el-icon>
            </div>
          </template>

          <template #sn="{ row }">
            <span class="mono">{{ row.sn || '-' }}</span>
          </template>

          <template #secret="{ row }">
            <div class="secret-cell">
              <span class="mono">{{ row.secret ? row.secret.slice(0, 6) + '****' : '-' }}</span>
              <el-button link type="primary" size="small" :disabled="!row.secret" @click.stop="copySecret(row)">复制</el-button>
            </div>
          </template>

          <template #location="{ row }">
            <span v-if="row.lat && row.lng" class="location mono">{{ row.lng.toFixed(4) }}, {{ row.lat.toFixed(4) }}</span>
            <span v-else class="muted">-</span>
          </template>
        </SmartTable>
      </div>
    </div>

    <DeviceForm
      v-model:visible="drawerVisible"
      :model-value="currentEditData"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Monitor, Plus, Upload, Fold, Expand } from '@element-plus/icons-vue'
import SmartTable from '@/components/SmartTable.vue'
import GroupTree from '@/components/GroupTree.vue'
import DeviceForm from './DeviceForm.vue'
import { searchModels } from '@/api/model'
import { searchDevices, createDevice, updateDevice, deleteDevice } from '@/api/device'

const props = defineProps({
  modelId: {
    type: String,
    default: ''
  },
  showGroup: {
    type: Boolean,
    default: true
  }
})

const router = useRouter()
const tableRef = ref(null)
const groupTreeRef = ref(null)

const baseUrl = import.meta.env.VITE_API_BASE_URL || ''

const getImageUrl = (url) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return baseUrl + url
}

const filterModelId = ref('')
const selectedGroupIds = ref([])
const myModels = ref([])

// 设备分组侧栏折叠状态（记录到 localStorage）
const GROUP_COLLAPSE_KEY = 'device_group_sidebar_collapsed'
const groupCollapsed = ref(false)
try {
  groupCollapsed.value = localStorage.getItem(GROUP_COLLAPSE_KEY) === '1'
} catch (e) { /* localStorage 不可用时忽略 */ }

const toggleGroupSidebar = () => {
  groupCollapsed.value = !groupCollapsed.value
  try {
    localStorage.setItem(GROUP_COLLAPSE_KEY, groupCollapsed.value ? '1' : '0')
  } catch (e) { /* localStorage 不可用时忽略 */ }
}

const filters = computed(() => {
  const list = []
  const modelFilter = filterModelId.value || props.modelId
  if (modelFilter) {
    list.push({ field: 'model_id', operator: 'eq', value: modelFilter })
  }
  if (selectedGroupIds.value.length > 0) {
    list.push({ field: 'group_id', operator: 'in', value: selectedGroupIds.value })
  }
  return list
})

const columns = [
  { prop: 'model_icon', label: '图标', width: 60, align: 'center', slot: 'icon', showOverflowTooltip: false },
  { prop: 'name', label: '设备名称', minWidth: 140 },
  { prop: 'sn', label: '设备 SN', minWidth: 160, slot: 'sn', showOverflowTooltip: false },
  { label: '密钥', width: 180, slot: 'secret', showOverflowTooltip: false },
  { prop: 'model_name', label: '物模型', minWidth: 120 },
  { prop: 'group_name', label: '分组', minWidth: 100 },
  { label: '位置', width: 160, slot: 'location', showOverflowTooltip: false },
  {
    prop: 'online',
    label: '在线',
    width: 80,
    align: 'center',
    type: 'tag',
    formatter: (row) => (row.online ? '在线' : '离线'),
    tagType: (row) => (row.online ? 'success' : 'info')
  },
  { prop: 'created', label: '创建时间', width: 170, type: 'datetime' }
]

const operators = [
  { key: 'detail', label: '详情', type: 'primary', handler: (row) => handleViewDetail(row) },
  { key: 'edit', label: '编辑', type: 'primary', handler: (row) => openEditDrawer(row) },
  { key: 'delete', label: '删除', type: 'danger', handler: (row) => handleDelete(row) }
]

const drawerVisible = ref(false)
const currentEditId = ref(null)
const currentEditData = ref({})

const handleFormSubmit = async (formData) => {
  if (currentEditId.value) {
    await updateDevice(currentEditId.value, formData)
  } else {
    await createDevice(formData)
  }
  drawerVisible.value = false
  tableRef.value?.refresh()
}

const loadMyModels = async () => {
  try {
    const res = await searchModels({ page: 1, page_size: 100 })
    myModels.value = res.data || []
  } catch (e) { console.error(e) }
}

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

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除设备"${row.name}"吗？`, '提示', { type: 'warning' })
    await deleteDevice(row.id)
    ElMessage.success('删除成功')
    tableRef.value?.refresh()
  } catch (e) { /* cancel */ }
}

const batchDeleteDevices = async (ids) => {
  await Promise.all(ids.map(id => deleteDevice(id).catch(() => null)))
}

const copySecret = (row) => {
  if (!row.secret) return
  navigator.clipboard.writeText(row.secret).then(() => ElMessage.success('密钥已复制'))
}

const handleGroupCheck = (checkedKeys) => {
  selectedGroupIds.value = checkedKeys
}

const handleRowClick = (row) => {
  router.push(`/devices/${row.id}`)
}

const handleViewDetail = (row) => {
  router.push(`/devices/${row.id}`)
}

onMounted(() => {
  loadMyModels()
})
</script>

<style lang="scss" scoped>
.device-list {
  .content-wrapper {
    display: flex;
    gap: 20px;
  }

  .sidebar {
    position: relative;
    width: 240px;
    flex-shrink: 0;
    background: #fff;
    border-radius: 8px;
    padding: 16px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
    transition: width 0.2s ease, padding 0.2s ease;

    &.is-collapsed {
      width: 44px;
      padding: 8px 6px;
      display: flex;
      align-items: center;
      justify-content: center;
      overflow: hidden;
    }

    .sidebar-body {
      height: 100%;
    }

    .sidebar-toggle {
      position: absolute;
      top: 50%;
      right: -10px;
      transform: translateY(-50%);
      width: 22px;
      height: 48px;
      padding: 0;
      background: #fff;
      border: 1px solid #ebeef5;
      border-left: none;
      border-radius: 0 6px 6px 0;
      box-shadow: 2px 0 6px rgba(0, 0, 0, 0.06);
      color: #909399;
      z-index: 3;

      &:hover {
        color: #409eff;
      }
    }

    .sidebar-rail {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 8px;
      cursor: pointer;
      color: #606266;
      user-select: none;

      .rail-label {
        writing-mode: vertical-rl;
        letter-spacing: 4px;
        font-size: 12px;
      }

      &:hover {
        color: #409eff;
      }
    }
  }

  .main-content {
    flex: 1;
    min-width: 0;
  }

  .mono { font-family: 'Consolas', 'Menlo', monospace; font-size: 12px; }
  .secret-cell { display: flex; gap: 6px; align-items: center;
    .el-button { padding: 0 4px; }
  }
  .location { color: #606266; }
  .muted { color: #c0c4cc; }

  .icon-cell {
    width: 40px;
    height: 40px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #f5f5f5;
    overflow: hidden;

    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }
}
</style>
