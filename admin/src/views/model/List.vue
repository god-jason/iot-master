<template>
  <div class="model-list">
    <SmartTable
      ref="tableRef"
      card-title="我的物模型"
      :loader="searchModels"
      :columns="columns"
      :operators="operators"
      :show-selection="true"
      :show-batch-delete="true"
      :delete-fn="batchDeleteModels"
      :operator-width="180"
      show-create
      create-text="创建物模型"
      keyword-placeholder="搜索物模型名称"
      storage-key="model"
      row-clickable
      @create="openCreateDrawer"
      @row-click="handleViewDetail"
    >
      <template #icon="{ row }">
        <div class="icon-cell">
          <img v-if="row.icon" :src="getImageUrl(row.icon)" alt="图标" />
          <el-icon v-else :size="24"><Monitor /></el-icon>
        </div>
      </template>

      <template #key="{ row }">
        <span class="key-text">{{ row.id }}</span>
      </template>

      <template #type="{ row }">{{ typeMap[row.type] || '-' }}</template>
    </SmartTable>

    <ModelForm
      v-model:visible="drawerVisible"
      :model-value="currentEditData"
      @submit="handleFormSubmit"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Monitor } from '@element-plus/icons-vue'
import SmartTable from '@/components/SmartTable.vue'
import ModelForm from './ModelForm.vue'
import { searchModels, createModel, updateModel, deleteModel } from '@/api/model'

const router = useRouter()
const tableRef = ref(null)

const baseUrl = import.meta.env.VITE_API_BASE_URL || ''

const getImageUrl = (url) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return baseUrl + url
}

const typeMap = {
  gateway: '网关',
  direct: '直连设备',
  sub: '子设备'
}

const columns = [
  { prop: 'icon', label: '图标', width: 80, slot: 'icon', showOverflowTooltip: false },
  { prop: 'id', label: '物模型ID', minWidth: 200, slot: 'key', showOverflowTooltip: false },
  { prop: 'name', label: '物模型名称', minWidth: 120 },
  { prop: 'type', label: '类型', width: 120, slot: 'type' },
  { prop: 'protocol', label: '协议', width: 120 },
  { prop: 'created', label: '创建时间', width: 180, type: 'datetime' }
]

const operators = [
  { key: 'detail', label: '详情', type: 'primary', handler: (row) => handleViewDetail(row) },
  { key: 'edit', label: '编辑', type: 'primary', handler: (row) => openEditDrawer(row) },
  { key: 'delete', label: '删除', type: 'danger', handler: (row) => handleDelete(row) }
]

const drawerVisible = ref(false)
const currentEditId = ref(null)
const currentEditData = ref({})

const batchDeleteModels = (ids) =>
  Promise.all(ids.map((id) => deleteModel(id).catch(() => null)))

const handleFormSubmit = async (formData) => {
  if (currentEditId.value) {
    await updateModel(currentEditId.value, formData)
  } else {
    await createModel(formData)
  }
  tableRef.value?.refresh()
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
    await ElMessageBox.confirm(`确定删除物模型"${row.name}"吗？`, '提示', { type: 'warning' })
    await deleteModel(row.id)
    ElMessage.success('删除成功')
    tableRef.value?.refresh()
  } catch (error) {
    if (error !== 'cancel') console.error(error)
  }
}

const handleViewDetail = (row) => {
  router.push(`/models/${row.id}`)
}
</script>

<style lang="scss" scoped>
.model-list {
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

  :deep(.el-table__row) {
    cursor: pointer;
  }
}
</style>
