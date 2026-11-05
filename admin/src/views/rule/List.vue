<template>
  <div class="rule-list">
    <SmartTable
      ref="tableRef"
      card-title="规则引擎"
      :loader="searchRules"
      :columns="columns"
      :operators="operators"
      :show-selection="true"
      :show-batch-delete="true"
      :delete-fn="batchDeleteRules"
      :show-create="true"
      create-text="新建规则"
      :operator-width="230"
      storage-key="rule"
      keyword-placeholder="搜索规则名称"
      @create="openCreate"
    >
      <template #scope="{ row }">
        <el-tag size="small">{{ scopeMap[row.scope?.type] || row.scope?.type }}</el-tag>
      </template>
      <template #trigger="{ row }">
        <el-tag size="small" type="info">{{ triggerMap[row.trigger?.type] || row.trigger?.type }}</el-tag>
      </template>
      <template #conditions="{ row }">
        <span v-if="row.expr" class="mono">{{ row.expr }}</span>
        <span v-else>{{ describeConditions(row) }}</span>
      </template>
      <template #actionText="{ row }">
        {{ (row.actions || []).map((a) => actionMap[a.type] || a.type).join('、') || '-' }}
      </template>
      <template #enabled="{ row }">
        <el-switch :model-value="!row.disabled" @change="(v) => handleSwitch(row, !v)" />
      </template>
    </SmartTable>

    <!-- 创建/编辑 -->
    <RuleForm v-model:visible="formVisible" :model-value="current" @submit="handleSubmit" />

    <!-- 试跑 -->
    <el-dialog v-model="testVisible" title="规则试跑" width="560px">
      <div class="test-tip">输入模拟上报数据（JSON），仅校验条件是否命中，不会执行动作。</div>
      <el-input v-model="testText" type="textarea" :rows="8" placeholder='{"temperature": 35, "humidity": 40}' />
      <div v-if="testResult !== null" class="test-result" :class="{ hit: testResult }">
        {{ testResult ? '条件命中 ✓' : '条件未命中' }}
      </div>
      <template #footer>
        <el-button @click="testVisible = false">关闭</el-button>
        <el-button type="primary" :loading="testing" @click="handleTest">执行试跑</el-button>
      </template>
    </el-dialog>

    <!-- 执行日志 -->
    <el-drawer v-model="logsVisible" :title="`执行日志 - ${current?.name || ''}`" size="60%">
      <div style="padding: 0 16px">
        <el-table :data="logs" v-loading="logsLoading" stripe size="small">
          <el-table-column prop="created" label="时间" width="170">
            <template #default="{ row }">{{ $formatDateTime(row.created) }}</template>
          </el-table-column>
          <el-table-column prop="device_id" label="设备ID" min-width="180" show-overflow-tooltip />
          <el-table-column prop="trigger" label="触发" width="90">
            <template #default="{ row }">{{ triggerMap[row.trigger] || row.trigger }}</template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag size="small" :type="row.state === 'firing' ? 'danger' : row.state === 'resolved' ? 'success' : 'info'">
                {{ stateMap[row.state] || row.state }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="动作结果" min-width="220">
            <template #default="{ row }">
              <div v-for="(r, i) in row.results || []" :key="i" class="log-result">
                <el-tag size="small" :type="r.success ? 'success' : 'danger'">{{ actionMap[r.type] || r.type }}</el-tag>
                <span>{{ r.message }}</span>
              </div>
            </template>
          </el-table-column>
        </el-table>
        <el-pagination
          style="margin-top: 12px"
          v-model:current-page="logPage"
          v-model:page-size="logPageSize"
          :total="logTotal"
          layout="total, prev, pager, next"
          @current-change="loadLogs"
        />
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import SmartTable from '@/components/SmartTable.vue'
import RuleForm from './RuleForm.vue'
import {
  searchRules, createRule, updateRule, deleteRule,
  switchRule, testRule, searchRuleLogs
} from '@/api/rule'

const scopeMap = { single: '单设备', model: '物模型', group: '分组' }
const triggerMap = { property: '属性上报', event: '设备事件', status: '上下线', schedule: '定时' }
const actionMap = {
  alarm: '异常报警',
  control_property: '下发属性',
  control_service: '调用服务',
  notify: '通知消息',
  data: '数据加工'
}
const stateMap = { firing: '告警中', resolved: '已恢复', pending: '去抖中', ok: '正常' }
const opMap = { '>': '>', '>=': '≥', '<': '<', '<=': '≤', '==': '=', '!=': '≠', between: '在区间', out_of: '在区间外' }

const tableRef = ref(null)
const formVisible = ref(false)
const current = ref(null)

const columns = [
  { prop: 'name', label: '规则名称', minWidth: 160 },
  { label: '作用范围', width: 140, slot: 'scope', showOverflowTooltip: false },
  { label: '触发方式', width: 120, slot: 'trigger', showOverflowTooltip: false },
  { label: '条件', minWidth: 200, slot: 'conditions', showOverflowTooltip: false },
  { label: '动作', width: 110, slot: 'actionText' },
  { prop: 'priority', label: '优先级', width: 80 },
  { label: '状态', width: 90, slot: 'enabled', showOverflowTooltip: false }
]

const operators = [
  { key: 'test', label: '试跑', type: 'primary', handler: (row) => openTest(row) },
  { key: 'logs', label: '日志', type: 'primary', handler: (row) => openLogs(row) },
  { key: 'edit', label: '编辑', type: 'primary', handler: (row) => openEdit(row) },
  { key: 'delete', label: '删除', type: 'danger', handler: (row) => handleDelete(row) }
]

// 结构化条件的简要描述
const describeConditions = (row) => {
  if (!row.conditions || row.conditions.length === 0) return '-'
  return row.conditions.map((c) => {
    if (c.op === 'between' || c.op === 'out_of') {
      const range = Array.isArray(c.value) ? c.value.join('~') : ''
      return `${c.field} ${opMap[c.op]} [${range}]`
    }
    return `${c.field} ${opMap[c.op] || c.op} ${c.value}`
  }).join(row.logic === 'or' ? ' 或 ' : ' 且 ')
}

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
    await updateRule(current.value.id, payload)
    ElMessage.success('更新成功')
  } else {
    await createRule(payload)
    ElMessage.success('创建成功')
  }
  formVisible.value = false
  tableRef.value?.refresh()
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除规则"${row.name}"吗？`, '提示', { type: 'warning' })
    await deleteRule(row.id)
    ElMessage.success('删除成功')
    tableRef.value?.refresh()
  } catch (e) {
    // 用户取消
  }
}

const batchDeleteRules = async (ids) => {
  await Promise.all(ids.map((id) => deleteRule(id).catch(() => null)))
}

const handleSwitch = async (row, disabled) => {
  try {
    await switchRule(row.id, disabled)
    row.disabled = disabled
    ElMessage.success(disabled ? '规则已停用' : '规则已启用')
  } catch (e) {
    // 失败时由请求拦截器提示
  }
}

// ---- 试跑 ----
const testVisible = ref(false)
const testing = ref(false)
const testText = ref('')
const testResult = ref(null)

const openTest = (row) => {
  current.value = row
  testText.value = '{\n  \n}'
  testResult.value = null
  testVisible.value = true
}
const handleTest = async () => {
  let values
  try {
    values = JSON.parse(testText.value || '{}')
  } catch (e) {
    ElMessage.warning('JSON 格式错误')
    return
  }
  testing.value = true
  try {
    const res = await testRule(current.value.id, values)
    testResult.value = !!res.data?.matched
  } finally {
    testing.value = false
  }
}

// ---- 日志 ----
const logsVisible = ref(false)
const logsLoading = ref(false)
const logs = ref([])
const logPage = ref(1)
const logPageSize = ref(20)
const logTotal = ref(0)

const openLogs = (row) => {
  current.value = row
  logsVisible.value = true
  logPage.value = 1
  loadLogs()
}
const loadLogs = async () => {
  logsLoading.value = true
  try {
    const res = await searchRuleLogs({
      page: logPage.value,
      page_size: logPageSize.value,
      order_by: 'created',
      filter: [{ field: 'rule_id', operator: 'eq', value: current.value.id }]
    })
    logs.value = res.data || []
    logTotal.value = res.total || 0
  } finally {
    logsLoading.value = false
  }
}
</script>

<style lang="scss" scoped>
.rule-list {
  .mono {
    font-family: Consolas, Monaco, monospace;
    font-size: 12px;
    color: #606266;
  }

  .test-tip {
    color: #909399;
    font-size: 13px;
    margin-bottom: 10px;
  }

  .test-result {
    margin-top: 12px;
    padding: 10px 14px;
    border-radius: 6px;
    background: #fef0f0;
    color: #f56c6c;
    font-weight: 600;

    &.hit {
      background: #f0f9eb;
      color: #67c23a;
    }
  }

  .log-result {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;
  }
}
</style>
