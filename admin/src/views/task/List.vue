<template>
  <div class="task-list">
    <SmartTable
      ref="tableRef"
      card-title="定时任务"
      :loader="searchTasks"
      :columns="columns"
      :operators="operators"
      :filter="filters"
      :show-selection="true"
      :show-batch-delete="true"
      :delete-fn="batchDeleteTasks"
      :operator-width="150"
      keyword-placeholder="搜索任务名称"
      storage-key="task"
    >
      <template #toolbar-left>
        <el-radio-group v-model="modeFilter">
          <el-radio-button value="">全部</el-radio-button>
          <el-radio-button value="online">在线模式</el-radio-button>
          <el-radio-button value="offline">离线模式</el-radio-button>
        </el-radio-group>
      </template>

      <template #toolbar-right>
        <el-button type="primary" :icon="Plus" @click="openCreate">创建任务</el-button>
      </template>

      <template #cron="{ row }"><span class="mono">{{ row.cron }}</span></template>

      <template #lastRun="{ row }">
        {{ row.last_run ? formatDateTime(row.last_run) : '未执行' }}
      </template>

      <template #status="{ row }">
        <el-switch :model-value="!row.disabled" @change="(v) => handleSwitch(row, !v)" />
      </template>
    </SmartTable>

    <TaskForm v-model:visible="formVisible" :model-value="current" @submit="handleSubmit" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SmartTable from '@/components/SmartTable.vue'
import { formatDateTime } from '@/utils/format'
import TaskForm from './TaskForm.vue'
import { searchTasks, createTask, updateTask, deleteTask, switchTask } from '@/api/task'

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
  { prop: 'name', label: '任务名称', minWidth: 160 },
  {
    prop: 'mode',
    label: '运行模式',
    width: 110,
    type: 'tag',
    formatter: (row) => (row.mode === 'offline' ? '离线' : '在线'),
    tagType: (row) => (row.mode === 'offline' ? 'warning' : 'success')
  },
  { prop: 'cron', label: 'Cron', width: 150, slot: 'cron', showOverflowTooltip: false },
  { prop: 'desc', label: '描述', minWidth: 160 },
  { label: '最近执行', width: 180, slot: 'lastRun', showOverflowTooltip: false },
  { label: '状态', width: 90, slot: 'status', showOverflowTooltip: false }
]

const operators = [
  { key: 'edit', label: '编辑', type: 'primary', handler: (row) => openEdit(row) },
  { key: 'delete', label: '删除', type: 'danger', handler: (row) => handleDelete(row) }
]

const formVisible = ref(false)
const current = ref(null)

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
    await updateTask(current.value.id, payload)
    ElMessage.success('更新成功')
  } else {
    await createTask(payload)
    ElMessage.success('创建成功')
  }
  formVisible.value = false
  tableRef.value?.refresh()
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除定时任务"${row.name}"吗？`, '提示', { type: 'warning' })
    await deleteTask(row.id)
    ElMessage.success('删除成功')
    tableRef.value?.refresh()
  } catch (e) {
    // 取消
  }
}

const batchDeleteTasks = async (ids) => {
  await Promise.all(ids.map((id) => deleteTask(id).catch(() => null)))
}

const handleSwitch = async (row, disabled) => {
  try {
    await switchTask(row.id, disabled)
    row.disabled = disabled
  } catch (e) {
    // 拦截器已提示
  }
}
</script>

<style lang="scss" scoped>
.task-list {
  .mono {
    font-family: Consolas, Monaco, monospace;
    font-size: 12px;
    color: #606266;
  }
}
</style>
