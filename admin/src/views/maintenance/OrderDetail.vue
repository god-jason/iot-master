<template>
  <div class="order-detail">
    <PageHeader title="工单详情">
      <template #extra>
        <div class="actions">
          <el-button @click="goBack">返回</el-button>
          <el-button v-if="order && order.status === 'pending'" @click="openAssign">指派</el-button>
          <el-button v-if="order && order.status === 'pending'" type="success" @click="handleStart">开始</el-button>
          <el-button v-if="canComplete" type="primary" @click="openComplete">完成</el-button>
          <el-button v-if="canCancel" type="warning" @click="handleCancel">取消</el-button>
        </div>
      </template>
    </PageHeader>

    <div v-loading="loading" class="detail-body">
      <template v-if="order">
        <div class="detail-head">
          <div>
            <h3>
              <span class="mono">{{ order.order_no || order.id }}</span>
              <span class="title">{{ order.title }}</span>
            </h3>
            <div class="detail-sub">
              <el-tag size="small" :type="statusTagType(order.status)">{{ statusText(order.status) }}</el-tag>
              <el-tag size="small" :type="typeTagType(order.type)">{{ typeText(order.type) }}</el-tag>
              <el-tag size="small" :type="priorityTagType(order.priority)">{{ priorityText(order.priority) }}优先级</el-tag>
              <el-tag v-if="isOverdue" size="small" type="danger">逾期</el-tag>
            </div>
          </div>
        </div>

        <el-descriptions :column="3" border class="detail-desc">
          <el-descriptions-item label="所属设备">{{ order.device_name || order.device_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="设备 SN">{{ order.device_sn || '-' }}</el-descriptions-item>
          <el-descriptions-item label="关联计划">{{ order.plan_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="负责人">{{ order.assignee_name || order.assignee_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="应完成时间">
            <span :class="{ overdue: isOverdue }">{{ fmt(order.due_at) }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="开始时间">{{ fmt(order.started_at) }}</el-descriptions-item>
          <el-descriptions-item label="完成时间">{{ fmt(order.finished_at) }}</el-descriptions-item>
          <el-descriptions-item label="处理结果">{{ resultText(order.result) }}</el-descriptions-item>
          <el-descriptions-item label="成本">{{ order.cost ? `¥ ${Number(order.cost).toFixed(2)}` : '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建人">{{ order.created_by || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ fmt(order.created) }}</el-descriptions-item>
          <el-descriptions-item label="备注">{{ order.remark || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div class="section-title">检查项</div>
        <el-table :data="orderItems" size="small" border empty-text="暂无检查项">
          <el-table-column prop="name" label="检查项" min-width="180" />
          <el-table-column prop="result" label="结果" width="140" />
          <el-table-column prop="remark" label="备注" min-width="200" />
        </el-table>

        <div class="section-title">备件</div>
        <el-table :data="orderParts" size="small" border empty-text="暂无备件">
          <el-table-column prop="name" label="备件名称" min-width="180" />
          <el-table-column prop="qty" label="数量" width="100" />
          <el-table-column label="单价(元)" width="120">
            <template #default="{ row }">{{ Number(row.price || 0).toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="小计(元)" width="120">
            <template #default="{ row }">{{ (Number(row.qty || 0) * Number(row.price || 0)).toFixed(2) }}</template>
          </el-table-column>
        </el-table>
        <div v-if="orderParts.length" class="parts-total">备件合计：¥ {{ orderPartsTotal }}</div>

        <div class="section-title">现场照片</div>
        <div v-if="attachmentUrls.length" class="photo-list">
          <el-image
            v-for="(url, i) in attachmentUrls"
            :key="i"
            :src="url"
            :preview-src-list="attachmentUrls"
            :initial-index="i"
            fit="cover"
            class="photo"
          />
        </div>
        <div v-else class="empty-tip">暂无现场照片</div>
      </template>

      <el-empty v-else-if="!loading" description="工单不存在或已被删除" />
    </div>

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
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import { searchUsers } from '@/api/org'
import { uploadImage } from '@/api/upload'
import { formatDateTime } from '@/utils/format'
import {
  getMaintenanceOrder,
  assignMaintenanceOrder,
  startMaintenanceOrder,
  completeMaintenanceOrder,
  cancelMaintenanceOrder
} from '@/api/maintenance'

const route = useRoute()
const router = useRouter()
const baseUrl = import.meta.env.VITE_API_BASE_URL || ''

const loading = ref(true)
const order = ref(null)

const assignVisible = ref(false)
const assigning = ref(false)
const assignUserId = ref('')
const assignUsers = ref([])
const assignUserLoading = ref(false)

const completeVisible = ref(false)
const completing = ref(false)
const completeForm = ref(defaultCompleteForm())
const photoList = ref([])
const attachments = ref([])
const previewVisible = ref(false)
const previewUrls = ref([])
const previewIndex = ref(0)

const resultOptions = [
  { value: 'normal', label: '正常' },
  { value: 'fixed', label: '已修复' },
  { value: 'replaced', label: '已更换' },
  { value: 'unresolved', label: '未解决' }
]

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

const loadData = async () => {
  loading.value = true
  try {
    const res = await getMaintenanceOrder(route.params.id)
    order.value = res.data?.order || res.data || null
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})

const goBack = () => {
  router.push('/maintenance/orders')
}

const orderItems = computed(() => parseArray(order.value?.items))
const orderParts = computed(() => parseArray(order.value?.parts))
const attachmentUrls = computed(() => parseArray(order.value?.attachments).filter(Boolean).map(getImageUrl))
const orderPartsTotal = computed(() => {
  const sum = orderParts.value.reduce((s, p) => s + Number(p.qty || 0) * Number(p.price || 0), 0)
  return sum.toFixed(2)
})

const isOverdue = computed(() => {
  if (!order.value) return false
  if (order.value.overdue) return true
  if (!order.value.due_at || ['completed', 'cancelled'].includes(order.value.status)) return false
  return new Date(order.value.due_at).getTime() < Date.now()
})

const canComplete = computed(() => order.value && ['processing', 'reviewing'].includes(order.value.status))
const canCancel = computed(() => order.value && ['pending', 'processing', 'reviewing'].includes(order.value.status))

// ========== 指派 ==========
const searchAssignUsers = async (kw) => {
  assignUserLoading.value = true
  try {
    const res = await searchUsers(kw || '')
    assignUsers.value = (res.data || []).map(u => ({ id: u.id, label: userLabel(u) }))
  } finally {
    assignUserLoading.value = false
  }
}

const openAssign = async () => {
  assignUserId.value = order.value?.assignee_id || ''
  assignVisible.value = true
  await searchAssignUsers('')
  if (order.value?.assignee_id && order.value?.assignee_name) {
    const exists = assignUsers.value.some(u => u.id === order.value.assignee_id)
    if (!exists) assignUsers.value.push({ id: order.value.assignee_id, label: order.value.assignee_name })
  }
}

const submitAssign = async () => {
  if (!assignUserId.value) {
    ElMessage.warning('请选择负责人')
    return
  }
  assigning.value = true
  try {
    await assignMaintenanceOrder(order.value.id, assignUserId.value)
    ElMessage.success('已指派')
    assignVisible.value = false
    loadData()
  } finally {
    assigning.value = false
  }
}

// ========== 开始 / 取消 ==========
const handleStart = async () => {
  await ElMessageBox.confirm('确认开始处理该工单？', '开始处理', { type: 'warning' })
  await startMaintenanceOrder(order.value.id)
  ElMessage.success('已开始处理')
  loadData()
}

const handleCancel = async () => {
  const { value } = await ElMessageBox.prompt('请填写取消原因', '取消工单', {
    type: 'warning',
    confirmButtonText: '确认取消',
    cancelButtonText: '返回',
    inputPlaceholder: '取消原因（可选）'
  })
  await cancelMaintenanceOrder(order.value.id, value || '')
  ElMessage.success('工单已取消')
  loadData()
}

// ========== 完成 ==========
const openComplete = () => {
  const items = parseArray(order.value?.items)
  const atts = parseArray(order.value?.attachments).filter(Boolean)
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

const partTotal = (part) => (Number(part.qty || 0) * Number(part.price || 0)).toFixed(2)

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
    await completeMaintenanceOrder(order.value.id, {
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
    loadData()
  } finally {
    completing.value = false
  }
}

// ========== 展示辅助 ==========
const statusText = (s) => ({ pending: '待处理', processing: '处理中', reviewing: '待验收', completed: '已完成', cancelled: '已取消' }[s] || s)
const statusTagType = (s) => ({ pending: 'warning', processing: 'primary', reviewing: 'warning', completed: 'success', cancelled: 'info' }[s] || 'info')
const typeText = (t) => ({ plan: '计划保养', fault: '故障维修', patrol: '巡检', install: '安装' }[t] || t)
const typeTagType = (t) => ({ plan: 'primary', fault: 'danger', patrol: 'success', install: 'info' }[t] || 'info')
const priorityText = (p) => ({ 1: '高', 2: '中', 3: '低' }[p] || `P${p}`)
const priorityTagType = (p) => ({ 1: 'danger', 2: 'warning', 3: 'info' }[p] || 'info')
const resultText = (r) => ({ normal: '正常', fixed: '已修复', replaced: '已更换', unresolved: '未解决' }[r] || (r || '-'))
</script>

<style lang="scss" scoped>
.order-detail {
  .actions {
    display: flex;
    gap: 8px;
  }

  .mono {
    font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
  }

  .detail-body {
    padding: 0 4px;
    min-height: 200px;
  }

  .detail-head {
    margin-bottom: 16px;

    h3 {
      display: flex;
      align-items: center;
      gap: 12px;
      margin: 0 0 8px;
      font-size: 18px;

      .title {
        font-weight: 600;
      }
    }

    .detail-sub {
      display: flex;
      align-items: center;
      gap: 10px;
    }
  }

  .detail-desc {
    margin-bottom: 18px;

    .overdue {
      color: #f56c6c;
      font-weight: 500;
    }
  }

  .section-title {
    margin: 20px 0 10px;
    font-weight: 600;
    font-size: 15px;
  }

  .parts-total {
    margin-top: 8px;
    text-align: right;
    font-weight: 600;
    color: #f56c6c;
  }

  .photo-list {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;

    .photo {
      width: 110px;
      height: 110px;
      border-radius: 6px;
      border: 1px solid #ebeef5;
      cursor: pointer;
    }
  }

  .empty-tip {
    color: #909399;
    font-size: 13px;
    padding: 8px 0;
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
