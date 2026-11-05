<template>
  <div class="firmware-list">
    <SmartTable
      ref="tableRef"
      card-title="固件管理"
      :loader="searchFirmware"
      :columns="columns"
      :operators="operators"
      :filter="filters"
      :operator-width="300"
      keyword-placeholder="搜索固件名称 / 版本号"
      storage-key="ota-firmware"
    >
      <template #toolbar-left>
        <el-select v-model="statusFilter" placeholder="全部状态" clearable style="width: 130px">
          <el-option label="草稿" :value="0" />
          <el-option label="已发布" :value="1" />
          <el-option label="已停用" :value="2" />
        </el-select>
        <el-select v-model="channelFilter" placeholder="全部通道" clearable style="width: 120px">
          <el-option label="稳定版" value="stable" />
          <el-option label="测试版" value="beta" />
        </el-select>
      </template>

      <template #toolbar-right>
        <el-button type="primary" :icon="Upload" @click="openUpload">上传固件</el-button>
      </template>

      <template #model="{ row }">
        <span v-if="row.model_name">{{ row.model_name }}</span>
        <el-tag v-else size="small" type="info">通用</el-tag>
      </template>

      <template #version="{ row }">
        <span class="mono">{{ row.version }}</span>
      </template>

      <template #size="{ row }">{{ formatSize(row.file_size) }}</template>

      <template #sha="{ row }">
        <span class="mono sha" :title="row.sha256">{{ row.sha256 || '-' }}</span>
      </template>

      <template #created="{ row }">{{ formatTime(row.created) }}</template>
    </SmartTable>

    <!-- 上传固件 -->
    <el-dialog v-model="uploadVisible" title="上传固件" width="660px" destroy-on-close>
      <el-form :model="uploadForm" label-width="100px">
        <el-form-item label="固件文件" required>
          <el-upload
            ref="uploadRef"
            drag
            :auto-upload="false"
            :limit="1"
            :on-change="onFileChange"
            :on-remove="onFileRemove"
            accept=".bin,.hex,.img,.fw,.zip,.tar,.gz,.tgz,.rar,.7z"
            style="width: 100%"
          >
            <el-icon class="upload-icon"><UploadFilled /></el-icon>
            <div class="el-upload__text">拖拽固件到此处，或 <em>点击选择</em></div>
            <template #tip>
              <div class="el-upload__tip">支持 .bin/.hex/.img/.zip 等，最大 {{ maxSizeMB }}MB</div>
            </template>
          </el-upload>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="固件名称" required>
              <el-input v-model="uploadForm.name" placeholder="如：温控器固件" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="版本号" required>
              <el-input v-model="uploadForm.version" placeholder="如：1.2.3" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="适用型号">
              <el-select v-model="uploadForm.model_id" placeholder="不选表示通用固件" clearable style="width: 100%">
                <el-option v-for="m in models" :key="m.id" :label="m.name" :value="m.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="发布通道">
              <el-select v-model="uploadForm.channel" style="width: 100%">
                <el-option label="稳定版 stable" value="stable" />
                <el-option label="测试版 beta" value="beta" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="最低版本">
              <el-input v-model="uploadForm.min_version" placeholder="低于该版本的设备不可升级（可选）" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="固件签名">
              <el-input v-model="uploadForm.signature" placeholder="可选" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="备注">
          <el-input v-model="uploadForm.notes" type="textarea" :rows="2" placeholder="升级说明 / 变更内容" />
        </el-form-item>
        <el-form-item v-if="uploading">
          <el-progress :percentage="uploadPercent" :stroke-width="10" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="submitUpload">上传</el-button>
      </template>
    </el-dialog>

    <!-- 编辑固件 -->
    <el-dialog v-model="editVisible" title="编辑固件" width="560px">
      <el-form :model="editForm" label-width="90px">
        <el-form-item label="固件名称" required>
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="版本号" required>
          <el-input v-model="editForm.version" />
        </el-form-item>
        <el-form-item label="适用型号">
          <el-select v-model="editForm.model_id" placeholder="不选表示通用固件" clearable style="width: 100%">
            <el-option v-for="m in models" :key="m.id" :label="m.name" :value="m.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="发布通道">
          <el-select v-model="editForm.channel" style="width: 100%">
            <el-option label="稳定版 stable" value="stable" />
            <el-option label="测试版 beta" value="beta" />
          </el-select>
        </el-form-item>
        <el-form-item label="最低版本">
          <el-input v-model="editForm.min_version" />
        </el-form-item>
        <el-form-item label="固件签名">
          <el-input v-model="editForm.signature" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="editForm.notes" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" @click="submitEdit">保存</el-button>
      </template>
    </el-dialog>

    <!-- 黑白名单 -->
    <el-drawer v-model="rulesVisible" title="固件黑白名单" size="720px" destroy-on-close>
      <div class="rules-body" v-loading="rulesLoading">
        <el-alert type="info" :closable="false" show-icon class="rules-tip">
          <div>白名单非空时，仅名单内的设备可以升级该固件；黑名单内的设备始终禁止升级，冲突时以黑名单为准。</div>
          <div>名单在预览和创建升级任务时生效，已创建的任务不受影响。</div>
        </el-alert>

        <div class="rules-head">
          <div class="rules-fw">
            {{ rulesFirmware.name }}
            <span class="mono">{{ rulesFirmware.version }}</span>
          </div>
          <div class="rules-actions">
            <el-button size="small" type="success" @click="openRuleAdd('white')">添加白名单</el-button>
            <el-button size="small" type="danger" @click="openRuleAdd('black')">添加黑名单</el-button>
          </div>
        </div>

        <div class="section-title">白名单（{{ whiteRules.length }}）</div>
        <el-table :data="whiteRules" size="small" border empty-text="白名单为空，表示不限制设备">
          <el-table-column label="类型" width="80">
            <template #default="{ row }">
              <el-tag size="small" :type="ruleTargetTag(row.target_type)">{{ ruleTargetText(row.target_type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="目标" min-width="150">
            <template #default="{ row }">
              <div>{{ row.target_name || row.target_id }}</div>
              <div v-if="row.target_sn" class="sub mono">{{ row.target_sn }}</div>
            </template>
          </el-table-column>
          <el-table-column prop="device_count" label="设备数" width="80" />
          <el-table-column label="添加时间" width="150">
            <template #default="{ row }">{{ formatTime(row.created) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="70">
            <template #default="{ row }">
              <el-button link type="danger" @click="removeRule(row)">移除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <div class="section-title">黑名单（{{ blackRules.length }}）</div>
        <el-table :data="blackRules" size="small" border empty-text="黑名单为空">
          <el-table-column label="类型" width="80">
            <template #default="{ row }">
              <el-tag size="small" :type="ruleTargetTag(row.target_type)">{{ ruleTargetText(row.target_type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="目标" min-width="150">
            <template #default="{ row }">
              <div>{{ row.target_name || row.target_id }}</div>
              <div v-if="row.target_sn" class="sub mono">{{ row.target_sn }}</div>
            </template>
          </el-table-column>
          <el-table-column prop="device_count" label="设备数" width="80" />
          <el-table-column label="添加时间" width="150">
            <template #default="{ row }">{{ formatTime(row.created) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="70">
            <template #default="{ row }">
              <el-button link type="danger" @click="removeRule(row)">移除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-drawer>

    <!-- 添加名单规则 -->
    <el-dialog v-model="ruleAddVisible" title="添加名单规则" width="560px" destroy-on-close>
      <el-form :model="ruleForm" label-width="90px">
        <el-form-item label="名单类型">
          <el-radio-group v-model="ruleForm.list_type">
            <el-radio-button value="white">白名单</el-radio-button>
            <el-radio-button value="black">黑名单</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="目标类型">
          <el-radio-group v-model="ruleForm.target_type" @change="onRuleTargetTypeChange">
            <el-radio-button value="device">设备</el-radio-button>
            <el-radio-button value="group">分组</el-radio-button>
            <el-radio-button value="org">组织</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="选择目标" required>
          <el-select
            v-if="ruleForm.target_type === 'device'"
            v-model="ruleForm.target_ids"
            multiple
            filterable
            remote
            :remote-method="searchRuleDevices"
            :loading="ruleDeviceLoading"
            placeholder="输入设备名称/SN 搜索"
            style="width: 100%"
            @visible-change="(v) => v && searchRuleDevices('')"
          >
            <el-option
              v-for="d in ruleDeviceOptions"
              :key="d.id"
              :label="`${d.name}（${d.sn || d.id}）`"
              :value="d.id"
            />
          </el-select>
          <el-select
            v-else-if="ruleForm.target_type === 'group'"
            v-model="ruleForm.target_ids"
            multiple
            filterable
            placeholder="选择设备分组"
            style="width: 100%"
          >
            <el-option v-for="g in groups" :key="g.id" :label="g.name" :value="String(g.id)" />
          </el-select>
          <el-tree-select
            v-else
            v-model="ruleForm.target_ids"
            :data="orgTree"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            node-key="id"
            multiple
            show-checkbox
            check-strictly
            placeholder="选择组织（自动包含子组织设备）"
            style="width: 100%"
          />
        </el-form-item>
        <el-alert v-if="ruleForm.target_type === 'org'" type="warning" :closable="false" show-icon>
          组织规则会自动覆盖其全部子组织绑定的设备。
        </el-alert>
      </el-form>
      <template #footer>
        <el-button @click="ruleAddVisible = false">取消</el-button>
        <el-button type="primary" :loading="ruleSaving" @click="submitRuleAdd">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Upload, UploadFilled } from '@element-plus/icons-vue'
import SmartTable from '@/components/SmartTable.vue'
import { searchModels } from '@/api/model'
import { searchGroups } from '@/api/group'
import { getOrgTree } from '@/api/org'
import { searchDevices } from '@/api/device'
import {
  searchFirmware, updateFirmware, deleteFirmware,
  releaseFirmware, disableFirmware, uploadFirmware,
  listFirmwareRules, addFirmwareRules, deleteFirmwareRule
} from '@/api/ota'

const tableRef = ref(null)
const statusFilter = ref(undefined)
const channelFilter = ref('')
const models = ref([])
const maxSizeMB = 200

const uploadVisible = ref(false)
const uploading = ref(false)
const uploadPercent = ref(0)
const uploadRef = ref(null)
const uploadFile = ref(null)
const uploadForm = ref(defaultUploadForm())

const editVisible = ref(false)
const editForm = ref({})

// ========== 黑白名单 ==========
const rulesVisible = ref(false)
const rulesLoading = ref(false)
const rulesFirmware = ref({})
const rules = ref([])
const whiteRules = computed(() => rules.value.filter(r => r.list_type === 'white'))
const blackRules = computed(() => rules.value.filter(r => r.list_type === 'black'))

const ruleAddVisible = ref(false)
const ruleSaving = ref(false)
const ruleForm = ref(defaultRuleForm())
const ruleDeviceOptions = ref([])
const ruleDeviceLoading = ref(false)
const groups = ref([])
const orgTree = ref([])

function defaultRuleForm() {
  return { list_type: 'white', target_type: 'device', target_ids: [] }
}

function defaultUploadForm() {
  return { name: '', version: '', model_id: '', channel: 'stable', min_version: '', signature: '', notes: '' }
}

const columns = [
  { prop: 'name', label: '固件名称', minWidth: 150 },
  { label: '适用型号', minWidth: 130, slot: 'model', showOverflowTooltip: false },
  { prop: 'version', label: '版本', width: 120, slot: 'version' },
  {
    prop: 'channel',
    label: '通道',
    width: 90,
    type: 'tag',
    formatter: (row) => (row.channel === 'beta' ? '测试版' : '稳定版'),
    tagType: (row) => (row.channel === 'beta' ? 'warning' : 'success')
  },
  { label: '文件大小', width: 105, slot: 'size' },
  { label: 'SHA256', minWidth: 170, slot: 'sha', showOverflowTooltip: false },
  {
    prop: 'status',
    label: '状态',
    width: 90,
    type: 'tag',
    formatter: (row) => statusText(row.status),
    tagType: (row) => statusTagType(row.status)
  },
  { label: '创建时间', width: 170, slot: 'created' }
]

const operators = [
  { key: 'edit', label: '编辑', type: 'primary', handler: (row) => openEdit(row) },
  { key: 'release', label: '发布', type: 'success', handler: (row) => handleRelease(row), show: (row) => row.status !== 1 },
  { key: 'disable', label: '停用', type: 'warning', handler: (row) => handleDisable(row), show: (row) => row.status === 1 },
  { key: 'rules', label: '黑白名单', type: 'primary', handler: (row) => openRules(row) },
  { key: 'download', label: '下载', type: 'primary', handler: (row) => download(row) },
  { key: 'delete', label: '删除', type: 'danger', handler: (row) => handleDelete(row) }
]

const filters = computed(() => {
  const list = []
  if (statusFilter.value !== undefined && statusFilter.value !== '') {
    list.push({ field: 'status', operator: 'eq', value: statusFilter.value })
  }
  if (channelFilter.value) {
    list.push({ field: 'channel', operator: 'eq', value: channelFilter.value })
  }
  return list
})

const loadModels = async () => {
  try {
    const res = await searchModels({ page: 1, page_size: 100 })
    models.value = res.data || []
  } catch (e) { /* ignore */ }
}

onMounted(() => {
  loadModels()
})

const openUpload = () => {
  uploadForm.value = defaultUploadForm()
  uploadFile.value = null
  uploadPercent.value = 0
  uploadVisible.value = true
}

const onFileChange = (file) => {
  uploadFile.value = file
}
const onFileRemove = () => {
  uploadFile.value = null
}

const submitUpload = async () => {
  if (!uploadFile.value) {
    ElMessage.warning('请选择固件文件')
    return
  }
  if (!uploadForm.value.name || !uploadForm.value.version) {
    ElMessage.warning('请填写固件名称和版本号')
    return
  }
  const fd = new FormData()
  fd.append('file', uploadFile.value.raw)
  Object.entries(uploadForm.value).forEach(([k, v]) => {
    if (v !== undefined && v !== null && v !== '') fd.append(k, v)
  })
  uploading.value = true
  uploadPercent.value = 0
  try {
    await uploadFirmware(fd, (e) => {
      if (e.total) uploadPercent.value = Math.round((e.loaded / e.total) * 100)
    })
    ElMessage.success('固件上传成功，当前为草稿状态')
    uploadVisible.value = false
    tableRef.value?.refresh()
  } finally {
    uploading.value = false
  }
}

const openEdit = (row) => {
  editForm.value = {
    id: row.id,
    name: row.name,
    version: row.version,
    model_id: row.model_id || '',
    channel: row.channel || 'stable',
    min_version: row.min_version || '',
    signature: row.signature || '',
    notes: row.notes || ''
  }
  editVisible.value = true
}

const submitEdit = async () => {
  if (!editForm.value.name || !editForm.value.version) {
    ElMessage.warning('固件名称和版本号不能为空')
    return
  }
  const { id, ...data } = editForm.value
  await updateFirmware(id, data)
  ElMessage.success('保存成功')
  editVisible.value = false
  tableRef.value?.refresh()
}

const handleRelease = async (row) => {
  await ElMessageBox.confirm(`确认发布固件「${row.name} ${row.version}」？发布后可用于升级任务。`, '发布固件', { type: 'warning' })
  await releaseFirmware(row.id)
  ElMessage.success('已发布')
  tableRef.value?.refresh()
}

const handleDisable = async (row) => {
  await ElMessageBox.confirm(`确认停用固件「${row.name} ${row.version}」？进行中的任务将暂停。`, '停用固件', { type: 'warning' })
  await disableFirmware(row.id)
  ElMessage.success('已停用')
  tableRef.value?.refresh()
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm(`确认删除固件「${row.name} ${row.version}」？`, '删除固件', { type: 'warning' })
  await deleteFirmware(row.id)
  ElMessage.success('已删除')
  tableRef.value?.refresh()
}

const download = (row) => {
  if (!row.download_url) {
    ElMessage.warning('固件下载地址不可用')
    return
  }
  window.open(row.download_url, '_blank')
}

// ========== 黑白名单操作 ==========
const openRules = async (row) => {
  rulesFirmware.value = row
  rulesVisible.value = true
  await loadRules(row.id)
  loadRuleRefs()
}

const loadRules = async (id) => {
  rulesLoading.value = true
  try {
    const res = await listFirmwareRules(id || rulesFirmware.value.id)
    rules.value = res.data?.rules || []
  } finally {
    rulesLoading.value = false
  }
}

const loadRuleRefs = async () => {
  if (!groups.value.length) {
    try {
      const res = await searchGroups({ page: 1, page_size: 100 })
      groups.value = res.data || []
    } catch (e) { /* ignore */ }
  }
  if (!orgTree.value.length) {
    try {
      const res = await getOrgTree()
      orgTree.value = res.data || []
    } catch (e) { /* ignore */ }
  }
}

const openRuleAdd = (listType) => {
  ruleForm.value = { ...defaultRuleForm(), list_type: listType }
  ruleAddVisible.value = true
  if (ruleForm.value.target_type === 'device') searchRuleDevices('')
}

const onRuleTargetTypeChange = () => {
  ruleForm.value.target_ids = []
  if (ruleForm.value.target_type === 'device') searchRuleDevices('')
}

const searchRuleDevices = async (kw) => {
  ruleDeviceLoading.value = true
  try {
    const res = await searchDevices({ page: 1, page_size: 20, keyword: kw || '' })
    ruleDeviceOptions.value = res.data || []
  } finally {
    ruleDeviceLoading.value = false
  }
}

const submitRuleAdd = async () => {
  if (!ruleForm.value.target_ids?.length) {
    ElMessage.warning('请选择要加入名单的目标')
    return
  }
  ruleSaving.value = true
  try {
    const res = await addFirmwareRules(rulesFirmware.value.id, {
      list_type: ruleForm.value.list_type,
      target_type: ruleForm.value.target_type,
      target_ids: ruleForm.value.target_ids.map(String)
    })
    const added = res.data?.added || 0
    const skipped = res.data?.skipped || 0
    ElMessage.success(skipped ? `已添加 ${added} 项，${skipped} 项已存在` : `已添加 ${added} 项`)
    ruleAddVisible.value = false
    await loadRules()
  } finally {
    ruleSaving.value = false
  }
}

const removeRule = async (row) => {
  await ElMessageBox.confirm(`确认将「${row.target_name || row.target_id}」移出名单？`, '移除名单', { type: 'warning' })
  await deleteFirmwareRule(rulesFirmware.value.id, row.id)
  ElMessage.success('已移除')
  await loadRules()
}

const ruleTargetText = (t) => ({ device: '设备', group: '分组', org: '组织' }[t] || t)
const ruleTargetTag = (t) => ({ device: 'primary', group: 'warning', org: 'success' }[t] || 'info')

const statusText = (status) => ({ 0: '草稿', 1: '已发布', 2: '已停用' }[status] || '未知')
const statusTagType = (status) => ({ 0: 'info', 1: 'success', 2: 'warning' }[status] || 'info')

const formatSize = (size) => {
  if (!size) return '-'
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(2)} MB`
}

const formatTime = (val) => {
  if (!val) return '-'
  const d = new Date(val)
  if (Number.isNaN(d.getTime())) return '-'
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}
</script>

<style lang="scss" scoped>
.firmware-list {
  .mono {
    font-family: ui-monospace, SFMono-Regular, Consolas, monospace;
  }

  .sha {
    display: inline-block;
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    vertical-align: middle;
    font-size: 12px;
    color: #909399;
  }

  .upload-icon {
    font-size: 42px;
    color: #c0c4cc;
    margin-top: 10px;
  }

  .rules-body {
    .rules-tip {
      margin-bottom: 16px;
    }

    .rules-head {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 8px;
    }

    .rules-fw {
      font-weight: 600;

      .mono {
        margin-left: 8px;
        font-weight: 400;
        font-size: 12px;
        color: #909399;
      }
    }

    .section-title {
      margin: 20px 0 10px;
      font-weight: 600;
      font-size: 15px;
    }

    .sub {
      font-size: 12px;
      color: #909399;
    }
  }
}
</style>
