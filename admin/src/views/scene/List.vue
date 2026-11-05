<template>
  <div class="scene-list">
    <SmartTable
      ref="tableRef"
      card-title="智能场景"
      :loader="searchScenes"
      :columns="columns"
      :operators="operators"
      :filter="filters"
      :show-selection="true"
      :show-batch-delete="true"
      :delete-fn="batchDeleteScenes"
      :operator-width="210"
      keyword-placeholder="搜索场景名称"
      storage-key="scene"
    >
      <template #toolbar-left>
        <el-radio-group v-model="modeFilter">
          <el-radio-button value="">全部</el-radio-button>
          <el-radio-button value="online">在线模式</el-radio-button>
          <el-radio-button value="offline">离线模式</el-radio-button>
        </el-radio-group>
      </template>

      <template #toolbar-right>
        <el-button type="primary" :icon="Plus" @click="openCreate">创建场景</el-button>
      </template>

      <template #actionCount="{ row }">{{ (row.actions || []).length }}</template>

      <template #status="{ row }">
        <el-switch :model-value="!row.disabled" @change="(v) => handleSwitch(row, !v)" />
      </template>
    </SmartTable>

    <SceneForm v-model:visible="formVisible" :model-value="current" @submit="handleSubmit" />

    <!-- 执行结果 -->
    <el-dialog v-model="resultVisible" :title="`执行结果 - ${current?.name || ''}`" width="520px">
      <el-table :data="results" size="small" stripe>
        <el-table-column prop="device_id" label="设备ID" min-width="160" show-overflow-tooltip />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">{{ row.type === 'property' ? '下发属性' : '调用服务' }}</template>
        </el-table-column>
        <el-table-column label="结果" min-width="140">
          <template #default="{ row }">
            <el-tag size="small" :type="row.success ? 'success' : 'danger'">
              {{ row.success ? '成功' : '失败' }}
            </el-tag>
            <span class="result-msg">{{ row.message }}</span>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button type="primary" @click="resultVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SmartTable from '@/components/SmartTable.vue'
import SceneForm from './SceneForm.vue'
import {
  searchScenes, createScene, updateScene, deleteScene,
  switchScene, executeScene
} from '@/api/scene'

const tableRef = ref(null)

const modeFilter = ref('')

const filters = computed(() => {
  const list = []
  if (modeFilter.value) {
    list.push({ field: 'mode', operator: 'eq', value: modeFilter.value })
  }
  return list
})

const columns = [
  { prop: 'name', label: '场景名称', minWidth: 160 },
  {
    prop: 'mode',
    label: '运行模式',
    width: 120,
    type: 'tag',
    formatter: (row) => (row.mode === 'offline' ? '离线模式' : '在线模式'),
    tagType: (row) => (row.mode === 'offline' ? 'warning' : 'success')
  },
  { prop: 'desc', label: '描述', minWidth: 180 },
  { label: '动作数', width: 90, slot: 'actionCount', showOverflowTooltip: false },
  { label: '状态', width: 90, slot: 'status', showOverflowTooltip: false }
]

const operators = [
  { key: 'execute', label: '执行', type: 'success', handler: (row) => handleExecute(row) },
  { key: 'edit', label: '编辑', type: 'primary', handler: (row) => openEdit(row) },
  { key: 'delete', label: '删除', type: 'danger', handler: (row) => handleDelete(row) }
]

const formVisible = ref(false)
const current = ref(null)
const resultVisible = ref(false)
const results = ref([])

const openCreate = () => {
  current.value = {}
  formVisible.value = true
}
const openEdit = (row) => {
  current.value = row
  formVisible.value = true
}

const handleSubmit = async (payload) => {
  if (current.value?.id) {
    await updateScene(current.value.id, payload)
    ElMessage.success('更新成功')
  } else {
    await createScene(payload)
    ElMessage.success('创建成功')
  }
  formVisible.value = false
  tableRef.value?.refresh()
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除场景"${row.name}"吗？关联的定时任务将无法触发该场景。`, '提示', { type: 'warning' })
    await deleteScene(row.id)
    ElMessage.success('删除成功')
    tableRef.value?.refresh()
  } catch (e) {
    // 取消
  }
}

const batchDeleteScenes = async (ids) => {
  await Promise.all(ids.map((id) => deleteScene(id).catch(() => null)))
}

const handleSwitch = async (row, disabled) => {
  try {
    await switchScene(row.id, disabled)
    row.disabled = disabled
  } catch (e) {
    // 拦截器已提示
  }
}

const handleExecute = async (row) => {
  try {
    const res = await executeScene(row.id)
    current.value = row
    results.value = res.data?.results || []
    resultVisible.value = true
  } catch (e) {
    // 拦截器已提示
  }
}
</script>

<style lang="scss" scoped>
.scene-list {
  .result-msg {
    margin-left: 6px;
    color: #909399;
    font-size: 12px;
  }
}
</style>
