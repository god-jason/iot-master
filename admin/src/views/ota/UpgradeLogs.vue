<template>
  <div class="upgrade-logs">
    <SmartTable
      ref="tableRef"
      card-title="升级日志"
      :loader="searchOTATasks"
      :columns="columns"
      :operators="operators"
      :filter="filters"
      :operator-width="220"
      keyword-placeholder="搜索任务名称"
      storage-key="ota-upgrade-task"
    >
      <template #toolbar-left>
        <el-select v-model="statusFilter" placeholder="全部状态" clearable style="width: 130px">
          <el-option v-for="s in statusOptions" :key="s.value" :label="s.label" :value="s.value" />
        </el-select>
      </template>

      <template #toolbar-right>
        <el-button type="primary" :icon="Plus" @click="openCreate">创建升级任务</el-button>
      </template>

      <template #firmware="{ row }">
        <div class="firmware-cell">
          <span>{{ row.firmware_name || '-' }}</span>
          <span class="mono version">{{ row.firmware_version || '' }}</span>
        </div>
      </template>

      <template #scope="{ row }">{{ scopeText(row.scope_type) }}</template>

      <template #progress="{ row }">
        <div class="progress-cell">
          <el-progress :percentage="taskPercent(row)" :stroke-width="8" :status="row.failed ? 'exception' : undefined" />
          <span class="progress-text">成功 {{ row.success || 0 }} / 失败 {{ row.failed || 0 }} / 共 {{ row.total || 0 }}</span>
        </div>
      </template>

      <template #created="{ row }">{{ formatTime(row.created) }}</template>
    </SmartTable>

    <!-- 创建任务 -->
    <el-dialog v-model="createVisible" title="创建升级任务" width="720px" destroy-on-close>
      <el-form :model="form" label-width="100px">
        <el-form-item label="任务名称">
          <el-input v-model="form.name" placeholder="留空自动生成" />
        </el-form-item>
        <el-form-item label="升级固件" required>
          <el-select v-model="form.firmware_id" placeholder="仅可选择已发布固件" style="width: 100%">
            <el-option
              v-for="fw in firmwares"
              :key="fw.id"
              :label="`${fw.name}  ${fw.version}`"
              :value="fw.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="升级范围" required>
          <el-radio-group v-model="form.scope_type" @change="onScopeTypeChange">
            <el-radio-button value="model">按型号</el-radio-button>
            <el-radio-button value="group">按分组</el-radio-button>
            <el-radio-button value="org">按组织</el-radio-button>
            <el-radio-button value="device">指定设备</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="scopeSelectLabel" required>
          <el-select
            v-if="form.scope_type === 'model'"
            v-model="form.scope_id"
            placeholder="选择物模型"
            style="width: 100%"
          >
            <el-option v-for="m in models" :key="m.id" :label="m.name" :value="m.id" />
          </el-select>
          <el-select
            v-else-if="form.scope_type === 'group'"
            v-model="form.scope_id"
            placeholder="选择设备分组"
            style="width: 100%"
          >
            <el-option v-for="g in groups" :key="g.id" :label="g.name" :value="String(g.id)" />
          </el-select>
          <el-tree-select
            v-else-if="form.scope_type === 'org'"
            v-model="form.scope_id"
            :data="orgTree"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            node-key="id"
            check-strictly
            placeholder="选择组织（含子组织）"
            style="width: 100%"
          />
          <el-select
            v-else
            v-model="form.scope_id"
            filterable
            remote
            :remote-method="searchDeviceOptions"
            :loading="deviceLoading"
            placeholder="输入设备名称/SN 搜索"
            style="width: 100%"
            @visible-change="(v) => v && searchDeviceOptions('')"
          >
            <el-option
              v-for="d in deviceOptions"
              :key="d.id"
              :label="`${d.name}（${d.sn || d.id}）`"
              :value="d.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="灰度方式">
          <el-radio-group v-model="form.strategy_type">
            <el-radio-button value="all">一次性全量</el-radio-button>
            <el-radio-button value="canary">灰度分批</el-radio-button>
          </el-radio-group>
        </el-form-item>

        <template v-if="form.strategy_type === 'canary'">
          <el-form-item label="批次设置">
            <div class="batch-editor">
              <div class="batch-head">
                <span>批次</span>
                <span>百分比(%)</span>
                <span>等待(分钟)</span>
                <span></span>
              </div>
              <div v-for="(b, i) in form.batches" :key="i" class="batch-row">
                <span>第 {{ i + 1 }} 批</span>
                <el-input-number v-model="b.percent" :min="1" :max="100" controls-position="right" />
                <el-input-number v-model="b.wait_minutes" :min="0" :max="10080" controls-position="right" />
                <el-button link type="danger" :disabled="form.batches.length <= 1" @click="removeBatch(i)">删除</el-button>
              </div>
              <el-button link type="primary" :icon="Plus" @click="addBatch">添加批次</el-button>
            </div>
          </el-form-item>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="成功率门禁">
                <el-input-number v-model="form.success_threshold" :min="0.5" :max="1" :step="0.01" :precision="2" style="width: 100%" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="失败率上限">
                <el-input-number v-model="form.failure_threshold" :min="0.01" :max="0.5" :step="0.01" :precision="2" style="width: 100%" />
              </el-form-item>
            </el-col>
          </el-row>
        </template>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="计划执行">
              <el-date-picker
                v-model="form.scheduled_at"
                type="datetime"
                placeholder="留空立即执行"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="失败回滚">
              <el-switch v-model="form.auto_rollback" />
              <span class="hint">设备支持时失败自动回滚旧固件</span>
            </el-form-item>
          </el-col>
        </el-row>

        <el-alert v-if="preview" type="info" :closable="false" class="preview-box">
          <div class="preview-title">
            目标设备 {{ preview.total }} 台（在线 {{ preview.online }}，离线 {{ preview.offline }}）
            <span v-if="preview.excluded">，已排除 {{ preview.excluded }} 台有进行中任务的设备</span>
            <span v-if="preview.rule_filtered">
              ，已按固件黑白名单排除 {{ preview.rule_filtered }} 台
              <template v-if="preview.whitelist_active">（白名单生效）</template>
            </span>
          </div>
          <div class="preview-batches">
            <span v-for="p in preview.batches" :key="p.seq" class="preview-chip">
              第{{ p.seq }}批 {{ p.count }}台
            </span>
          </div>
        </el-alert>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button :loading="previewing" @click="handlePreview">预览设备</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">创建任务</el-button>
      </template>
    </el-dialog>

    <!-- 任务详情 -->
    <el-drawer v-model="detailVisible" title="升级任务详情" size="72%" @close="closeDetail">
      <div v-if="detail.task" class="detail-body">
        <div class="detail-head">
          <div>
            <h3>{{ detail.task.name }}</h3>
            <div class="detail-sub">
              <el-tag size="small" :type="taskTagType(detail.task.status)">{{ taskStatusText(detail.task.status) }}</el-tag>
              <span class="mono">{{ detail.firmware.version }}</span>
              <span>{{ scopeText(detail.task.scope_type) }}升级</span>
            </div>
          </div>
          <div class="detail-actions">
            <el-button v-if="canPause(detail.task)" size="small" type="warning" @click="handlePause(detail.task)">暂停</el-button>
            <el-button v-if="detail.task.status === 'paused'" size="small" type="success" @click="handleResume(detail.task)">恢复</el-button>
            <el-button v-if="canCancel(detail.task)" size="small" type="danger" @click="handleCancel(detail.task)">取消</el-button>
            <el-button v-if="canRetry(detail.task)" size="small" type="primary" @click="handleRetry(detail.task)">重试失败设备</el-button>
          </div>
        </div>

        <el-descriptions :column="3" border class="detail-desc">
          <el-descriptions-item label="总设备数">{{ detail.task.total || 0 }}</el-descriptions-item>
          <el-descriptions-item label="成功">{{ detail.task.success || 0 }}</el-descriptions-item>
          <el-descriptions-item label="失败">{{ detail.task.failed || 0 }}</el-descriptions-item>
          <el-descriptions-item label="开始时间">{{ formatTime(detail.task.started_at) }}</el-descriptions-item>
          <el-descriptions-item label="结束时间">{{ formatTime(detail.task.finished_at) }}</el-descriptions-item>
          <el-descriptions-item label="失败回滚">{{ detail.task.auto_rollback ? '开启' : '关闭' }}</el-descriptions-item>
          <el-descriptions-item label="计划时间">{{ formatTime(detail.task.scheduled_at) }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTime(detail.task.created) }}</el-descriptions-item>
          <el-descriptions-item label="暂停原因">{{ detail.task.pause_reason === 'auto' ? '门禁自动暂停' : (detail.task.pause_reason === 'manual' ? '手动暂停' : '-') }}</el-descriptions-item>
        </el-descriptions>

        <div class="section-title">灰度批次</div>
        <el-table :data="detail.batches" size="small" border>
          <el-table-column prop="seq" label="批次" width="70">
            <template #default="{ row }">第{{ row.seq }}批</template>
          </el-table-column>
          <el-table-column label="目标比例" width="100">
            <template #default="{ row }">{{ row.percent ? row.percent + '%' : '重试批' }}</template>
          </el-table-column>
          <el-table-column prop="wait_minutes" label="等待(分钟)" width="100" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag size="small" :type="batchTagType(row.status)">{{ batchStatusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="开始" width="165">
            <template #default="{ row }">{{ formatTime(row.started_at) }}</template>
          </el-table-column>
          <el-table-column label="完成" width="165">
            <template #default="{ row }">{{ formatTime(row.finished_at) }}</template>
          </el-table-column>
        </el-table>

        <div class="section-title">
          设备升级明细
          <div class="device-filter">
            <el-select v-model="deviceStatus" placeholder="全部状态" clearable size="small" style="width: 130px" @change="loadDevices">
              <el-option v-for="s in deviceStatusOptions" :key="s.value" :label="s.label" :value="s.value" />
            </el-select>
            <el-input
              v-model="deviceKeyword"
              size="small"
              placeholder="设备名称 / SN"
              clearable
              style="width: 190px"
              @clear="loadDevices"
              @keyup.enter="loadDevices"
            />
          </div>
        </div>
        <el-table :data="detail.devices" size="small" border v-loading="deviceTableLoading">
          <el-table-column label="设备" min-width="170">
            <template #default="{ row }">
              <div class="device-name">{{ row.device_name || row.device_id }}</div>
              <div class="device-sn">{{ row.device_sn || '' }}</div>
            </template>
          </el-table-column>
          <el-table-column label="在线" width="70">
            <template #default="{ row }">
              <el-tag size="small" :type="row.online ? 'success' : 'info'">{{ row.online ? '在线' : '离线' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="版本变化" width="150">
            <template #default="{ row }">
              <span class="mono">{{ row.from_version || '-' }} → {{ row.to_version || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag size="small" :type="deviceTagType(row.status)">{{ deviceStatusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="进度" width="140">
            <template #default="{ row }">
              <el-progress :percentage="row.progress || 0" :stroke-width="6" :status="row.status === 'failed' || row.status === 'timeout' ? 'exception' : (row.status === 'success' ? 'success' : undefined)" />
            </template>
          </el-table-column>
          <el-table-column prop="retry_count" label="重试" width="65" />
          <el-table-column prop="error" label="失败原因" min-width="150" show-overflow-tooltip />
          <el-table-column label="最近上报" width="165">
            <template #default="{ row }">{{ formatTime(row.updated) }}</template>
          </el-table-column>
        </el-table>

        <el-pagination
          v-model:current-page="devicePage"
          v-model:page-size="devicePageSize"
          :total="deviceTotal"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="loadDevices"
          @current-change="loadDevices"
          class="pagination"
        />
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SmartTable from '@/components/SmartTable.vue'
import { searchModels } from '@/api/model'
import { searchGroups } from '@/api/group'
import { getOrgTree } from '@/api/org'
import { searchDevices } from '@/api/device'
import {
  searchFirmware, searchOTATasks, getOTATask, getOTATaskDevices,
  previewOTATask, createOTATask, pauseOTATask, resumeOTATask,
  cancelOTATask, retryFailedOTATask
} from '@/api/ota'

const tableRef = ref(null)
const statusFilter = ref('')

const statusOptions = [
  { value: 'scheduled', label: '待执行' },
  { value: 'running', label: '执行中' },
  { value: 'paused', label: '已暂停' },
  { value: 'completed', label: '已完成' },
  { value: 'partial', label: '部分成功' },
  { value: 'failed', label: '失败' },
  { value: 'cancelled', label: '已取消' }
]

const deviceStatusOptions = [
  { value: 'pending', label: '等待下发' },
  { value: 'dispatched', label: '已下发' },
  { value: 'downloading', label: '下载中' },
  { value: 'verifying', label: '校验中' },
  { value: 'installing', label: '写入中' },
  { value: 'rebooting', label: '重启中' },
  { value: 'success', label: '成功' },
  { value: 'failed', label: '失败' },
  { value: 'timeout', label: '超时' },
  { value: 'cancelled', label: '已取消' },
  { value: 'rolled_back', label: '已回滚' }
]

const firmwares = ref([])
const models = ref([])
const groups = ref([])
const orgTree = ref([])
const deviceOptions = ref([])
const deviceLoading = ref(false)

const createVisible = ref(false)
const previewing = ref(false)
const creating = ref(false)
const preview = ref(null)
const form = ref(defaultForm())

const detailVisible = ref(false)
const detail = ref({ task: null, firmware: {}, batches: [], devices: [] })
const devicePage = ref(1)
const devicePageSize = ref(50)
const deviceTotal = ref(0)
const deviceStatus = ref('')
const deviceKeyword = ref('')
const deviceTableLoading = ref(false)
let pollTimer = null

function defaultForm() {
  return {
    name: '',
    firmware_id: '',
    scope_type: 'model',
    scope_id: '',
    strategy_type: 'all',
    batches: [
      { percent: 5, wait_minutes: 30 },
      { percent: 20, wait_minutes: 60 },
      { percent: 100, wait_minutes: 0 }
    ],
    success_threshold: 0.95,
    failure_threshold: 0.10,
    scheduled_at: null,
    auto_rollback: false
  }
}

const scopeSelectLabel = computed(() => {
  return { model: '选择型号', group: '选择分组', org: '选择组织', device: '选择设备' }[form.value.scope_type] || '升级目标'
})

const columns = [
  { prop: 'name', label: '任务名称', minWidth: 170 },
  { label: '固件版本', width: 150, slot: 'firmware', showOverflowTooltip: false },
  { label: '升级范围', width: 100, slot: 'scope' },
  {
    prop: 'status',
    label: '状态',
    width: 100,
    type: 'tag',
    formatter: (row) => taskStatusText(row.status),
    tagType: (row) => taskTagType(row.status)
  },
  { label: '进度', minWidth: 190, slot: 'progress', showOverflowTooltip: false },
  { label: '创建时间', width: 170, slot: 'created' }
]

const operators = [
  { key: 'detail', label: '详情', type: 'primary', handler: (row) => openDetail(row) },
  { key: 'pause', label: '暂停', type: 'warning', handler: (row) => handlePause(row), show: (row) => canPause(row) },
  { key: 'resume', label: '恢复', type: 'success', handler: (row) => handleResume(row), show: (row) => row.status === 'paused' },
  { key: 'cancel', label: '取消', type: 'danger', handler: (row) => handleCancel(row), show: (row) => canCancel(row) },
  { key: 'retry', label: '重试失败', type: 'primary', handler: (row) => handleRetry(row), show: (row) => canRetry(row) }
]

const filters = computed(() => {
  const list = []
  if (statusFilter.value) list.push({ field: 'status', operator: 'eq', value: statusFilter.value })
  return list
})

const loadRefs = async () => {
  try {
    const fwRes = await searchFirmware({ page: 1, page_size: 100, filter: [{ field: 'status', operator: 'eq', value: 1 }] })
    firmwares.value = fwRes.data || []
  } catch (e) { /* ignore */ }
  try {
    const mRes = await searchModels({ page: 1, page_size: 100 })
    models.value = mRes.data || []
  } catch (e) { /* ignore */ }
  try {
    const gRes = await searchGroups({ page: 1, page_size: 100 })
    groups.value = gRes.data || []
  } catch (e) { /* ignore */ }
  try {
    const oRes = await getOrgTree()
    orgTree.value = oRes.data || []
  } catch (e) { /* ignore */ }
}

onMounted(() => {
  loadRefs()
})

onUnmounted(() => {
  stopPolling()
})

const openCreate = async () => {
  form.value = defaultForm()
  preview.value = null
  if (!firmwares.value.length) await loadRefs()
  createVisible.value = true
}

const onScopeTypeChange = () => {
  form.value.scope_id = ''
  preview.value = null
}

const searchDeviceOptions = async (kw) => {
  deviceLoading.value = true
  try {
    const res = await searchDevices({ page: 1, page_size: 20, keyword: kw || '' })
    deviceOptions.value = res.data || []
  } finally {
    deviceLoading.value = false
  }
}

const addBatch = () => {
  form.value.batches.push({ percent: 10, wait_minutes: 10 })
}
const removeBatch = (index) => {
  form.value.batches.splice(index, 1)
}

const buildPayload = () => {
  const strategy = {
    type: form.value.strategy_type,
    success_threshold: form.value.success_threshold,
    failure_threshold: form.value.failure_threshold,
    batches: form.value.strategy_type === 'all'
      ? [{ percent: 100, wait_minutes: 0 }]
      : form.value.batches.map(b => ({ percent: Number(b.percent), wait_minutes: Number(b.wait_minutes) }))
  }
  return {
    name: form.value.name,
    firmware_id: form.value.firmware_id,
    scope_type: form.value.scope_type,
    scope_id: String(form.value.scope_id || ''),
    strategy,
    scheduled_at: form.value.scheduled_at ? form.value.scheduled_at.getTime() : 0,
    auto_rollback: form.value.auto_rollback
  }
}

const handlePreview = async () => {
  if (!form.value.firmware_id) {
    ElMessage.warning('请选择升级固件')
    return
  }
  if (!form.value.scope_id) {
    ElMessage.warning('请选择升级目标')
    return
  }
  previewing.value = true
  try {
    const res = await previewOTATask(buildPayload())
    preview.value = res.data
    if (!res.data.total) ElMessage.warning('目标范围内没有可升级的设备')
  } finally {
    previewing.value = false
  }
}

const submitCreate = async () => {
  if (!form.value.firmware_id) {
    ElMessage.warning('请选择升级固件')
    return
  }
  if (!form.value.scope_id) {
    ElMessage.warning('请选择升级目标')
    return
  }
  if (form.value.strategy_type === 'canary') {
    const sum = form.value.batches.reduce((s, b) => s + Number(b.percent || 0), 0)
    if (sum < 100) {
      ElMessage.warning('灰度批次百分比合计建议达到 100%（不足部分会归入最后一批）')
    }
  }
  creating.value = true
  try {
    const res = await createOTATask(buildPayload())
    const excluded = res.data?.excluded || 0
    const ruleFiltered = res.data?.rule_filtered || 0
    const notes = []
    if (ruleFiltered) notes.push(`黑白名单排除 ${ruleFiltered} 台`)
    if (excluded) notes.push(`冲突任务排除 ${excluded} 台`)
    ElMessage.success(notes.length ? `任务已创建（${notes.join('，')}）` : '任务已创建')
    createVisible.value = false
    tableRef.value?.refresh()
  } finally {
    creating.value = false
  }
}

// ========== 详情 ==========
const openDetail = async (row) => {
  detailVisible.value = true
  devicePage.value = 1
  deviceStatus.value = ''
  deviceKeyword.value = ''
  await loadDetail(row.id)
  await loadDevices()
  startPolling(row.id)
}

const loadDetail = async (id) => {
  const res = await getOTATask(id)
  detail.value = {
    task: res.data.task,
    firmware: res.data.firmware || {},
    batches: res.data.batches || [],
    devices: detail.value.devices || []
  }
}

const loadDevices = async () => {
  const id = detail.value.task?.id
  if (!id) return
  deviceTableLoading.value = true
  try {
    const res = await getOTATaskDevices(id, {
      page: devicePage.value,
      page_size: devicePageSize.value,
      status: deviceStatus.value,
      keyword: deviceKeyword.value
    })
    detail.value.devices = res.data || []
    deviceTotal.value = res.total || 0
  } finally {
    deviceTableLoading.value = false
  }
}

const startPolling = (id) => {
  stopPolling()
  pollTimer = setInterval(async () => {
    if (!detailVisible.value) return stopPolling()
    try {
      await loadDetail(id)
      await loadDevices()
    } catch (e) { /* ignore */ }
  }, 5000)
}

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

const closeDetail = () => {
  stopPolling()
  detail.value = { task: null, firmware: {}, batches: [], devices: [] }
  deviceTotal.value = 0
}

const handlePause = async (row) => {
  await pauseOTATask(row.id)
  ElMessage.success('已暂停')
  refreshCurrent(row.id)
}

const handleResume = async (row) => {
  await resumeOTATask(row.id)
  ElMessage.success('已恢复')
  refreshCurrent(row.id)
}

const handleCancel = async (row) => {
  await ElMessageBox.confirm('取消后未完成的设备将停止升级，确认取消任务？', '取消任务', { type: 'warning' })
  await cancelOTATask(row.id)
  ElMessage.success('已取消')
  refreshCurrent(row.id)
}

const handleRetry = async (row) => {
  const res = await retryFailedOTATask(row.id)
  ElMessage.success(`已安排 ${res.data?.retry_count || 0} 台设备重试`)
  refreshCurrent(row.id)
}

const refreshCurrent = async (id) => {
  tableRef.value?.refresh()
  if (detailVisible.value && detail.value.task?.id === id) {
    await loadDetail(id)
    await loadDevices()
  }
}

// ========== 状态与格式化 ==========
const canPause = (row) => row.status === 'running' || row.status === 'scheduled'
const canCancel = (row) => ['scheduled', 'running', 'paused'].includes(row.status)
const canRetry = (row) => ['paused', 'partial', 'failed', 'completed'].includes(row.status) && row.failed > 0

const taskPercent = (row) => {
  if (!row.total) return 0
  return Math.min(100, Math.round(((row.success || 0) + (row.failed || 0)) / row.total * 100))
}

const taskStatusText = (s) => ({
  scheduled: '待执行', running: '执行中', paused: '已暂停',
  completed: '已完成', partial: '部分成功', failed: '失败', cancelled: '已取消'
}[s] || s)

const taskTagType = (s) => ({
  scheduled: 'info', running: 'primary', paused: 'warning',
  completed: 'success', partial: 'warning', failed: 'danger', cancelled: 'info'
}[s] || 'info')

const batchStatusText = (s) => ({ pending: '等待中', running: '下发中', completed: '已完成' }[s] || s)
const batchTagType = (s) => ({ pending: 'info', running: 'primary', completed: 'success' }[s] || 'info')

const deviceStatusText = (s) => ({
  pending: '等待下发', dispatched: '已下发', downloading: '下载中', verifying: '校验中',
  installing: '写入中', rebooting: '重启中', success: '成功', failed: '失败',
  timeout: '超时', cancelled: '已取消', rolled_back: '已回滚'
}[s] || s)

const deviceTagType = (s) => ({
  success: 'success', failed: 'danger', timeout: 'danger', rolled_back: 'warning',
  cancelled: 'info', pending: 'info', dispatched: 'primary',
  downloading: 'primary', verifying: 'primary', installing: 'primary', rebooting: 'primary'
}[s] || 'info')

const scopeText = (s) => ({ model: '型号', group: '分组', org: '组织', device: '设备' }[s] || s)

const formatTime = (val) => {
  if (!val) return '-'
  const d = new Date(val)
  if (Number.isNaN(d.getTime()) || d.getFullYear() < 1970) return '-'
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}
</script>

<style lang="scss" scoped>
.upgrade-logs {
  .mono {
    font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
  }

  .firmware-cell {
    display: flex;
    flex-direction: column;
    line-height: 1.3;

    .version {
      color: #909399;
      font-size: 12px;
    }
  }

  .progress-cell {
    display: flex;
    flex-direction: column;
    gap: 4px;

    .progress-text {
      font-size: 12px;
      color: #909399;
    }
  }

  .hint {
    margin-left: 8px;
    font-size: 12px;
    color: #909399;
  }

  .pagination {
    margin-top: 16px;
    justify-content: flex-end;
  }

  .batch-editor {
    width: 100%;
    border: 1px solid #ebeef5;
    border-radius: 6px;
    padding: 10px;

    .batch-head,
    .batch-row {
      display: grid;
      grid-template-columns: 80px 1fr 1fr 60px;
      gap: 10px;
      align-items: center;
    }

    .batch-head {
      font-size: 12px;
      color: #909399;
      margin-bottom: 8px;
    }

    .batch-row {
      margin-bottom: 8px;
    }
  }

  .preview-box {
    margin-top: 4px;

    .preview-title {
      font-weight: 500;
      margin-bottom: 6px;
    }

    .preview-chip {
      display: inline-block;
      margin-right: 8px;
      padding: 2px 8px;
      border-radius: 10px;
      background: rgba(64, 158, 255, 0.1);
      color: #409eff;
      font-size: 12px;
    }
  }

  .detail-body {
    padding: 0 4px;

    .detail-head {
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      margin-bottom: 16px;

      h3 {
        margin: 0 0 8px;
        font-size: 18px;
      }

      .detail-sub {
        display: flex;
        align-items: center;
        gap: 12px;
        font-size: 13px;
        color: #909399;
      }
    }

    .detail-actions {
      display: flex;
      gap: 8px;
    }

    .detail-desc {
      margin-bottom: 18px;
    }

    .section-title {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin: 20px 0 10px;
      font-weight: 600;
      font-size: 15px;

      .device-filter {
        display: flex;
        gap: 8px;
      }
    }

    .device-name {
      font-weight: 500;
    }

    .device-sn {
      font-size: 12px;
      color: #909399;
      font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
    }
  }
}
</style>
