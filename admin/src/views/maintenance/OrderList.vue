<template>
  <div class="order-list">
    <SmartTable
      ref="tableRef"
      card-title="维保工单"
      :loader="searchMaintenanceOrders"
      :columns="columns"
      :operators="operators"
      :params="queryParams"
      :operator-width="330"
      keyword-placeholder="搜索单号 / 标题 / 设备"
      storage-key="maintenance-order"
    >
      <template #toolbar-left>
        <el-select v-model="statusFilter" placeholder="全部状态" clearable style="width: 120px">
          <el-option v-for="s in statusOptions" :key="s.value" :label="s.label" :value="s.value" />
        </el-select>
        <el-select v-model="typeFilter" placeholder="全部类型" clearable style="width: 120px">
          <el-option v-for="t in typeOptions" :key="t.value" :label="t.label" :value="t.value" />
        </el-select>
        <el-select
          v-model="assigneeFilter"
          placeholder="全部负责人"
          filterable
          remote
          clearable
          :remote-method="searchFilterUsers"
          :loading="filterUserLoading"
          style="width: 180px"
          @visible-change="(v) => v && searchFilterUsers('')"
        >
          <el-option v-for="u in filterUsers" :key="u.id" :label="u.label" :value="u.id" />
        </el-select>
        <el-select v-model="overdueFilter" placeholder="是否逾期" clearable style="width: 120px">
          <el-option label="已逾期" :value="true" />
          <el-option label="未逾期" :value="false" />
        </el-select>
        <el-date-picker
          v-model="dueRange"
          type="daterange"
          range-separator="至"
          start-placeholder="应完成起"
          end-placeholder="应完成止"
          style="width: 240px"
        />
      </template>

      <template #toolbar-right>
        <el-button :icon="Download" :loading="exporting" @click="handleExport">导出</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建工单</el-button>
      </template>

      <template #orderNo="{ row }">
        <span class="mono">{{ row.order_no || row.id }}</span>
      </template>

      <template #device="{ row }">
        <div class="device-cell">
          <span>{{ row.device_name || row.device_id || '-' }}</span>
          <span class="device-sn mono">{{ row.device_sn || '' }}</span>
        </div>
      </template>

      <template #assignee="{ row }">{{ row.assignee_name || row.assignee_id || '-' }}</template>

      <template #due="{ row }">
        <div class="due-cell">
          <span :class="{ overdue: isOverdue(row) }">{{ fmt(row.due_at) }}</span>
          <el-tag v-if="isOverdue(row)" size="small" type="danger">逾期</el-tag>
        </div>
      </template>
    </SmartTable>

    <!-- 新建工单 -->
    <el-dialog v-model="createVisible" title="新建维保工单" width="640px" destroy-on-close>
      <el-form :model="createForm" label-width="100px">
        <el-form-item label="设备" required>
          <el-select
            v-model="createForm.device_id"
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
        <el-form-item label="标题" required>
          <el-input v-model="createForm.title" placeholder="如：空调滤芯更换" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="工单类型">
              <el-select v-model="createForm.type" style="width: 100%">
                <el-option v-for="t in typeOptions" :key="t.value" :label="t.label" :value="t.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="优先级">
              <el-select v-model="createForm.priority" style="width: 100%">
                <el-option v-for="p in priorityOptions" :key="p.value" :label="p.label" :value="p.value" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="负责人">
              <el-select
                v-model="createForm.assignee_id"
                filterable
                remote
                clearable
                :remote-method="searchFormUsers"
                :loading="formUserLoading"
                placeholder="输入姓名/用户名搜索"
                style="width: 100%"
                @visible-change="(v) => v && searchFormUsers('')"
              >
                <el-option v-for="u in formUsers" :key="u.id" :label="u.label" :value="u.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="应完成时间">
              <el-date-picker v-model="createForm.due_at" type="datetime" placeholder="选择时间" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="备注">
          <el-input v-model="createForm.remark" type="textarea" :rows="2" placeholder="故障描述 / 处理要求" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 指派负责人 -->
    <el-dialog v-model="assignVisible" title="指派负责人" width="480px" destroy-on-close>
      <el-form label-width="80px">
        <el-form-item label="负责人">
          <el-select
            v-model="assignUserId"
            filterable
            remote
            clearable
            :remote-method="searchAssignUsers"
            :loading="assignUserLoading"
            placeholder="输入姓名/用户名搜索"
            style="width: 100%"
            @visible-change="(v) => v && searchAssignUsers('')"
          >
            <el-option v-for="u in assignUsers" :key="u.id" :label="u.label" :value="u.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="assignVisible = false">取消</el-button>
        <el-button type="primary" :loading="assigning" @click="submitAssign">确定</el-button>
      </template>
    </el-dialog>

    <!-- 完成工单 -->
    <el-dialog v-model="completeVisible" title="完成维保工单" width="760px" destroy-on-close>
      <el-form :model="completeForm" label-width="90px">
        <el-form-item label="处理结果" required>
          <el-radio-group v-model="completeForm.result">
            <el-radio v-for="r in resultOptions" :key="r.value" :value="r.value">{{ r.label }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="检查项">
          <div class="items-editor">
            <div class="items-head">
              <span>检查项</span>
              <span>结果</span>
              <span>备注</span>
              <span></span>
            </div>
            <div v-for="(item, i) in completeForm.items" :key="i" class="items-row">
              <el-input v-model="item.name" placeholder="检查项名称" />
              <el-input v-model="item.result" placeholder="如：正常" />
              <el-input v-model="item.remark" placeholder="备注（可选）" />
              <el-button link type="danger" @click="removeCompleteItem(i)">删除</el-button>
            </div>
            <el-button link type="primary" :icon="Plus" @click="addCompleteItem">添加检查项</el-button>
          </div>
        </el-form-item>

        <el-form-item label="现场照片">
          <el-upload
            v-model:file-list="photoList"
            list-type="picture-card"
            accept="image/*"
            :http-request="handleUploadPhoto"
            :on-remove="handleRemovePhoto"
            :on-preview="handlePreviewPhoto"
          >
            <el-icon><Plus /></el-icon>
          </el-upload>
        </el-form-item>

        <el-form-item label="备件">
          <div class="parts-editor">
            <div class="parts-head">
              <span>备件名称</span>
              <span>数量</span>
              <span>单价(元)</span>
              <span>小计(元)</span>
              <span></span>
            </div>
            <div v-for="(part, i) in completeForm.parts" :key="i" class="parts-row">
              <el-input v-model="part.name" placeholder="如：滤芯" />
              <el-input-number v-model="part.qty" :min="1" :max="9999" controls-position="right" />
              <el-input-number v-model="part.price" :min="0" :max="999999" :precision="2" controls-position="right" />
              <span class="subtotal">{{ partTotal(part) }}</span>
              <el-button link type="danger" @click="removePart(i)">删除</el-button>
            </div>
            <el-button link type="primary" :icon="Plus" @click="addPart">添加备件</el-button>
          </div>
        </el-form-item>
        <el-form-item label="成本合计">
          <span class="cost-total">¥ {{ partsTotal }}</span>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="completeForm.remark" type="textarea" :rows="2" placeholder="处理说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="completeVisible = false">取消</el-button>
        <el-button type="primary" :loading="completing" @click="submitComplete">确认完成</el-button>
      </template>
    </el-dialog>

    <!-- 图片预览 -->
    <el-image-viewer
      v-if="previewVisible"
      :url-list="previewUrls"
      :initial-index="previewIndex"
      @close="previewVisible = false"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Download } from '@element-plus/icons-vue'
import SmartTable from '@/components/SmartTable.vue'
import { searchDevices } from '@/api/device'
import { searchUsers } from '@/api/org'
import { uploadImage } from '@/api/upload'
import { formatDateTime } from '@/utils/format'
import {
  searchMaintenanceOrders,
  createMaintenanceOrder,
  deleteMaintenanceOrder,
  assignMaintenanceOrder,
  startMaintenanceOrder,
  completeMaintenanceOrder,
  cancelMaintenanceOrder,
  exportMaintenanceOrders
} from '@/api/maintenance'

const router = useRouter()
const baseUrl = import.meta.env.VITE_API_BASE_URL || ''

const tableRef = ref(null)
const statusFilter = ref('')
const typeFilter = ref('')
const assigneeFilter = ref('')
const overdueFilter = ref('')
const dueRange = ref(null)
const exporting = ref(false)

const statusOptions = [
  { value: 'pending', label: '待处理' },
  { value: 'processing', label: '处理中' },
  { value: 'reviewing', label: '待验收' },
  { value: 'completed', label: '已完成' },
  { value: 'cancelled', label: '已取消' }
]
const typeOptions = [
  { value: 'plan', label: '计划保养' },
  { value: 'fault', label: '故障维修' },
  { value: 'patrol', label: '巡检' },
  { value: 'install', label: '安装' }
]
const priorityOptions = [
  { value: 1, label: '高' },
  { value: 2, label: '中' },
  { value: 3, label: '低' }
]
const resultOptions = [
  { value: 'normal', label: '正常' },
  { value: 'fixed', label: '已修复' },
  { value: 'replaced', label: '已更换' },
  { value: 'unresolved', label: '未解决' }
]

// ========== 新建 ==========
const createVisible = ref(false)
const creating = ref(false)
const createForm = ref(defaultCreateForm())
const formUsers = ref([])
const formUserLoading = ref(false)
const deviceOptions = ref([])
const deviceLoading = ref(false)

function defaultCreateForm() {
  return {
    device_id: '',
    title: '',
    type: 'plan',
    priority: 2,
    assignee_id: '',
    due_at: null,
    remark: ''
  }
}

// ========== 筛选负责人 ==========
const filterUsers = ref([])
const filterUserLoading = ref(false)

// ========== 指派 ==========
const assignVisible = ref(false)
const assigning = ref(false)
const assignUserId = ref('')
const assignOrder = ref(null)
const assignUsers = ref([])
const assignUserLoading = ref(false)

// ========== 完成 ==========
const completeVisible = ref(false)
const completing = ref(false)
const completeOrder = ref(null)
const completeForm = ref(defaultCompleteForm())
const photoList = ref([])
const attachments = ref([])
const previewVisible = ref(false)
const previewUrls = ref([])
const previewIndex = ref(0)

function defaultCompleteForm() {
  return { result: 'normal', items: [], parts: [], remark: '' }
}

const getImageUrl = (url) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return baseUrl + url
}

const fmt = (val) => {
  const text = formatDateTime(val, 'YYYY-MM-DD HH:mm')
  return text || '-'
}

const parseArray = (val) => {
  if (!val) return []
  if (Array.isArray(val)) return val
  if (typeof val === 'string') {
    try {
      const arr = JSON.parse(val)
      return Array.isArray(arr) ? arr : []
    } catch (e) {
      return []
    }
  }
  return []
}

const userLabel = (u) => (u.nickname ? `${u.nickname}（${u.username}）` : (u.username || u.id))

const columns = [
  { label: '单号', width: 170, slot: 'orderNo', showOverflowTooltip: false },
  { prop: 'device_name', label: '设备', minWidth: 180, slot: 'device', showOverflowTooltip: false },
  { prop: 'title', label: '标题', minWidth: 170 },
  {
    prop: 'type',
    label: '类型',
    width: 90,
    type: 'tag',
    formatter: (row) => typeText(row.type),
    tagType: (row) => typeTagType(row.type)
  },
  {
    prop: 'priority',
    label: '优先级',
    width: 90,
    type: 'tag',
    formatter: (row) => priorityText(row.priority),
    tagType: (row) => priorityTagType(row.priority)
  },
  { label: '负责人', width: 110, slot: 'assignee' },
  { label: '应完成时间', width: 190, slot: 'due', showOverflowTooltip: false },
  {
    prop: 'status',
    label: '状态',
    width: 90,
    type: 'tag',
    formatter: (row) => statusText(row.status),
    tagType: (row) => statusTagType(row.status)
  }
]

const operators = [
  { key: 'detail', label: '详情', type: 'primary', handler: (row) => openDetail(row) },
  { key: 'assign', label: '指派', type: 'primary', handler: (row) => openAssign(row), show: (row) => row.status === 'pending' },
  { key: 'start', label: '开始', type: 'success', handler: (row) => handleStart(row), show: (row) => row.status === 'pending' },
  { key: 'complete', label: '完成', type: 'success', handler: (row) => openComplete(row), show: (row) => canComplete(row) },
  { key: 'cancel', label: '取消', type: 'warning', handler: (row) => handleCancel(row), show: (row) => canCancel(row) },
  { key: 'delete', label: '删除', type: 'danger', handler: (row) => handleDelete(row) }
]

const queryParams = computed(() => {
  const payload = {}
  if (statusFilter.value) payload.status = statusFilter.value
  if (typeFilter.value) payload.type = typeFilter.value
  if (assigneeFilter.value !== '' && assigneeFilter.value !== undefined) payload.assignee_id = assigneeFilter.value
  if (overdueFilter.value !== '' && overdueFilter.value !== undefined) payload.overdue = overdueFilter.value
  if (dueRange.value?.length === 2) {
    const start = new Date(dueRange.value[0])
    start.setHours(0, 0, 0, 0)
    const end = new Date(dueRange.value[1])
    end.setHours(23, 59, 59, 999)
    payload.due_from = start.toISOString()
    payload.due_to = end.toISOString()
  }
  return payload
})

// ========== 用户/设备检索 ==========
const searchUserList = async (kw, target, loadingRef) => {
  loadingRef.value = true
  try {
    const res = await searchUsers(kw || '')
    target.value = (res.data || []).map(u => ({ id: u.id, label: userLabel(u) }))
  } finally {
    loadingRef.value = false
  }
}

const searchFilterUsers = (kw) => searchUserList(kw, filterUsers, filterUserLoading)
const searchFormUsers = (kw) => searchUserList(kw, formUsers, formUserLoading)
const searchAssignUsers = (kw) => searchUserList(kw, assignUsers, assignUserLoading)

const searchDeviceOptions = async (kw) => {
  deviceLoading.value = true
  try {
    const res = await searchDevices({ page: 1, page_size: 20, keyword: kw || '' })
    deviceOptions.value = res.data || []
  } finally {
    deviceLoading.value = false
  }
}

// ========== 新建工单 ==========
const openCreate = () => {
  createForm.value = defaultCreateForm()
  createVisible.value = true
  searchFormUsers('')
  if (!deviceOptions.value.length) searchDeviceOptions('')
}

const submitCreate = async () => {
  if (!createForm.value.device_id) {
    ElMessage.warning('请选择设备')
    return
  }
  if (!createForm.value.title) {
    ElMessage.warning('请填写工单标题')
    return
  }
  creating.value = true
  try {
    const payload = {
      device_id: createForm.value.device_id,
      type: createForm.value.type,
      title: createForm.value.title.trim(),
      priority: Number(createForm.value.priority || 2),
      assignee_id: createForm.value.assignee_id || '',
      remark: createForm.value.remark
    }
    // 后端时间为 RFC3339 字符串
    if (createForm.value.due_at) payload.due_at = new Date(createForm.value.due_at).toISOString()
    await createMaintenanceOrder(payload)
    ElMessage.success('工单已创建')
    createVisible.value = false
    tableRef.value?.refresh()
  } finally {
    creating.value = false
  }
}

const openDetail = (row) => {
  router.push(`/maintenance/orders/${row.id}`)
}

// ========== 指派 ==========
const openAssign = async (row) => {
  assignOrder.value = row
  assignUserId.value = row.assignee_id || ''
  assignVisible.value = true
  await searchAssignUsers('')
  if (row.assignee_id && row.assignee_name) {
    const exists = assignUsers.value.some(u => u.id === row.assignee_id)
    if (!exists) assignUsers.value.push({ id: row.assignee_id, label: row.assignee_name })
  }
}

const submitAssign = async () => {
  if (!assignUserId.value) {
    ElMessage.warning('请选择负责人')
    return
  }
  assigning.value = true
  try {
    await assignMaintenanceOrder(assignOrder.value.id, assignUserId.value)
    ElMessage.success('已指派')
    assignVisible.value = false
    tableRef.value?.refresh()
  } finally {
    assigning.value = false
  }
}

// ========== 开始/取消/删除 ==========
const handleStart = async (row) => {
  await ElMessageBox.confirm(`确认开始处理工单「${row.order_no || row.title}」？`, '开始处理', { type: 'warning' })
  await startMaintenanceOrder(row.id)
  ElMessage.success('已开始处理')
  tableRef.value?.refresh()
}

const handleCancel = async (row) => {
  const { value } = await ElMessageBox.prompt('请填写取消原因', `取消工单「${row.order_no || row.title}」`, {
    type: 'warning',
    confirmButtonText: '确认取消',
    cancelButtonText: '返回',
    inputPlaceholder: '取消原因（可选）'
  })
  await cancelMaintenanceOrder(row.id, value || '')
  ElMessage.success('工单已取消')
  tableRef.value?.refresh()
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm(`确认删除工单「${row.order_no || row.title}」？`, '删除工单', { type: 'warning' })
  await deleteMaintenanceOrder(row.id)
  ElMessage.success('已删除')
  tableRef.value?.refresh()
}

// ========== 完成 ==========
const canComplete = (row) => ['processing', 'reviewing'].includes(row.status)
const canCancel = (row) => ['pending', 'processing', 'reviewing'].includes(row.status)

const openComplete = (row) => {
  completeOrder.value = row
  const items = parseArray(row.items)
  const atts = parseArray(row.attachments).filter(Boolean)
  completeForm.value = {
    result: 'normal',
    items: items.map(i => ({ name: i.name || '', result: i.result || '', remark: i.remark || '' })),
    parts: [],
    remark: ''
  }
  attachments.value = [...atts]
  photoList.value = atts.map((url, i) => ({
    name: `现场照片${i + 1}`,
    url: getImageUrl(url),
    rawUrl: url,
    uid: Date.now() + i,
    status: 'success'
  }))
  completeVisible.value = true
}

const addCompleteItem = () => {
  completeForm.value.items.push({ name: '', result: '', remark: '' })
}
const removeCompleteItem = (i) => {
  completeForm.value.items.splice(i, 1)
}
const addPart = () => {
  completeForm.value.parts.push({ name: '', qty: 1, price: 0 })
}
const removePart = (i) => {
  completeForm.value.parts.splice(i, 1)
}

const partTotal = (part) => {
  return (Number(part.qty || 0) * Number(part.price || 0)).toFixed(2)
}

const partsTotal = computed(() => {
  const sum = completeForm.value.parts.reduce((s, p) => s + Number(p.qty || 0) * Number(p.price || 0), 0)
  return sum.toFixed(2)
})

const handleUploadPhoto = async (options) => {
  try {
    const res = await uploadImage(options.file)
    const url = res.data?.url || ''
    if (url) attachments.value.push(url)
    const item = photoList.value.find(f => f.uid === options.file.uid)
    if (item) {
      item.rawUrl = url
      item.url = getImageUrl(url)
    }
    options.onSuccess(res)
  } catch (e) {
    options.onError(e)
  }
}

const handleRemovePhoto = (file) => {
  const url = file.rawUrl || file.response?.data?.url
  if (!url) return
  const index = attachments.value.indexOf(url)
  if (index > -1) attachments.value.splice(index, 1)
}

const handlePreviewPhoto = (file) => {
  previewUrls.value = photoList.value.map(f => f.url).filter(Boolean)
  previewIndex.value = Math.max(0, photoList.value.findIndex(f => f.uid === file.uid))
  previewVisible.value = true
}

const submitComplete = async () => {
  if (!completeForm.value.result) {
    ElMessage.warning('请选择处理结果')
    return
  }
  completing.value = true
  try {
    await completeMaintenanceOrder(completeOrder.value.id, {
      result: completeForm.value.result,
      items: completeForm.value.items
        .filter(i => i.name && i.name.trim())
        .map(i => ({ name: i.name.trim(), result: (i.result || '').trim(), remark: (i.remark || '').trim() })),
      attachments: [...attachments.value],
      parts: completeForm.value.parts
        .filter(p => p.name && p.name.trim())
        .map(p => ({ name: p.name.trim(), qty: Number(p.qty || 0), price: Number(p.price || 0) })),
      cost: Number(partsTotal.value),
      remark: completeForm.value.remark
    })
    ElMessage.success('工单已完成')
    completeVisible.value = false
    tableRef.value?.refresh()
  } finally {
    completing.value = false
  }
}

// ========== 导出 ==========
const handleExport = async () => {
  exporting.value = true
  try {
    const blob = await exportMaintenanceOrders({
      ...queryParams.value,
      keyword: tableRef.value?.keyword || ''
    })
    const url = window.URL.createObjectURL(new Blob([blob], { type: 'text/csv;charset=utf-8' }))
    const a = document.createElement('a')
    a.href = url
    a.download = `maintenance-orders-${formatDateTime(new Date(), 'YYYYMMDD')}.csv`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    window.URL.revokeObjectURL(url)
  } finally {
    exporting.value = false
  }
}

// ========== 展示辅助 ==========
const isOverdue = (row) => {
  if (row.overdue) return true
  if (!row.due_at || ['completed', 'cancelled'].includes(row.status)) return false
  return new Date(row.due_at).getTime() < Date.now()
}

const statusText = (s) => ({ pending: '待处理', processing: '处理中', reviewing: '待验收', completed: '已完成', cancelled: '已取消' }[s] || s)
const statusTagType = (s) => ({ pending: 'warning', processing: 'primary', reviewing: 'warning', completed: 'success', cancelled: 'info' }[s] || 'info')
const typeText = (t) => ({ plan: '计划保养', fault: '故障维修', patrol: '巡检', install: '安装' }[t] || t)
const typeTagType = (t) => ({ plan: 'primary', fault: 'danger', patrol: 'success', install: 'info' }[t] || 'info')
const priorityText = (p) => ({ 1: '高', 2: '中', 3: '低' }[p] || `P${p}`)
const priorityTagType = (p) => ({ 1: 'danger', 2: 'warning', 3: 'info' }[p] || 'info')
</script>

<style lang="scss" scoped>
.order-list {
  .mono {
    font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
  }

  .device-cell,
  .due-cell {
    display: flex;
    flex-direction: column;
    gap: 2px;
    line-height: 1.3;
  }

  .device-sn {
    font-size: 12px;
    color: #909399;
  }

  .due-cell {
    flex-direction: row;
    align-items: center;
    gap: 6px;

    .overdue {
      color: #f56c6c;
      font-weight: 500;
    }
  }

  .items-editor,
  .parts-editor {
    width: 100%;
    border: 1px solid #ebeef5;
    border-radius: 6px;
    padding: 10px;
  }

  .items-head,
  .items-row {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr 60px;
    gap: 10px;
    align-items: center;
  }

  .parts-head,
  .parts-row {
    display: grid;
    grid-template-columns: 1fr 120px 140px 90px 60px;
    gap: 10px;
    align-items: center;
  }

  .items-head,
  .parts-head {
    font-size: 12px;
    color: #909399;
    margin-bottom: 8px;
  }

  .items-row,
  .parts-row {
    margin-bottom: 8px;
  }

  .subtotal {
    text-align: right;
    color: #606266;
  }

  .cost-total {
    font-size: 16px;
    font-weight: 600;
    color: #f56c6c;
  }
}
</style>
