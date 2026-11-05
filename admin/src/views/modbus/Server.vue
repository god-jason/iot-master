<template>
  <div class="gateway-server-page">
    <SmartTable
      ref="tableRef"
      card-title="网关服务器"
      :loader="loader"
      :columns="columns"
      :operators="operators"
      :show-create="true"
      create-text="创建服务器"
      :show-pagination="false"
      :show-keyword-search="false"
      :operator-width="220"
      @create="openCreate"
    >
      <template #address="{ row }">
        <span class="mono">{{ row.port || '-' }}</span>
        <el-tag size="small" type="info" style="margin-left: 6px">全接口</el-tag>
      </template>
      <template #protocol="{ row }">{{ protocolLabel(row.protocol) }}</template>
      <template #defaultDevices="{ row }">{{ row.default_devices?.length || 0 }} 个</template>
      <template #status="{ row }">
        <el-tag :type="row.running ? 'success' : 'info'" size="small">
          {{ row.running ? '运行中' : '已停止' }}
        </el-tag>
        <el-tag v-if="row.disabled" type="warning" size="small" style="margin-left: 4px">已禁用</el-tag>
      </template>
    </SmartTable>

    <!-- 创建/编辑 Dialog -->
    <el-dialog v-model="showDialog" :title="editing ? '编辑服务器' : '创建服务器'" width="720px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如 车间A-TCP" maxlength="64" show-word-limit />
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input-number controls-position="right" v-model="form.port" :min="1" :max="65535" />
          <span class="field-hint">监听所有网卡地址（0.0.0.0），无需指定 IP</span>
        </el-form-item>
        <el-form-item label="协议" prop="protocol">
          <el-select v-model="form.protocol">
            <el-option v-for="p in protocolList" :key="p.value" :label="p.label" :value="p.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="轮询间隔 (秒)">
          <el-input-number controls-position="right" v-model="form.interval" :min="1" :max="3600" />
        </el-form-item>
        <el-form-item label="并发数">
          <el-input-number controls-position="right" v-model="form.concurrency" :min="1" :max="64" />
        </el-form-item>
        <el-form-item label="禁用">
          <el-switch v-model="form.disabled" />
        </el-form-item>

        <el-divider content-position="left">默认设备</el-divider>
        <div class="default-devices">
          <div class="default-devices-hint">
            新链路注册且名下没有设备时，按此列表自动创建设备并绑定到该链路。
          </div>
          <el-table :data="form.default_devices" size="small" stripe>
            <el-table-column label="名称" min-width="110">
              <template #default="{ row }">
                <el-input v-model="row.name" placeholder="设备名称" />
              </template>
            </el-table-column>
            <el-table-column label="物模型" min-width="150">
              <template #default="{ row }">
                <el-select v-model="row.model_id" placeholder="选择物模型" filterable style="width: 100%">
                  <el-option v-for="m in models" :key="m.id" :label="m.name" :value="m.id" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="协议地址" width="130">
              <template #default="{ row }">
                <el-input v-model="row.address" placeholder="如 1" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="60">
              <template #default="{ $index }">
                <el-button link type="danger" @click="removeDefaultDevice($index)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-button class="add-default-btn" link type="primary" :icon="Plus" @click="addDefaultDevice">添加默认设备</el-button>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="saveForm">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SmartTable from '@/components/SmartTable.vue'
import {
  listGatewayServers, createGatewayServer, updateGatewayServer,
  deleteGatewayServer, startGatewayServer, stopGatewayServer
} from '@/api/gateway'
import { searchModels } from '@/api/model'

const protocolList = [
  { value: 'tcp',    label: 'Modbus TCP' },
  { value: 'rtu',    label: 'Modbus RTU' },
  { value: 'dlt645', label: 'DLT645' },
  { value: 'cjt188', label: 'CJT188' },
  { value: 'iec104', label: 'IEC104' },
  { value: 'jt808',  label: 'JT808' },
]

function protocolLabel(v) {
  const p = protocolList.find(x => x.value === v)
  return p ? p.label : v || '-'
}

// XSS 危险模式（不用 g 标志避免 lastIndex 污染）
function containsXSS(value) {
  if (!value) return false
  const p1 = /<\s*(script|iframe|object|embed|form|link|meta|style|base|body|img)\b/i
  const p2 = /on\w+\s*=/i
  const p3 = /javascript:/i
  return p1.test(value) || p2.test(value) || p3.test(value)
}

const rules = {
  name: [
    { required: true, message: '请输入服务器名称', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (!value || !value.trim()) {
          callback(new Error('服务器名称不能为空'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    },
    { validator: (rule, value, callback) => { if (containsXSS(value)) callback(new Error('包含不允许的特殊字符')); else callback() }, trigger: 'blur' }
  ],
  port: [
    { required: true, message: '请输入端口号', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (!value || value < 1 || value > 65535) {
          callback(new Error('端口号需在 1-65535 之间'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  protocol: [{ required: true, message: '请选择协议', trigger: 'change' }]
}

const router = useRouter()
const tableRef = ref(null)
const models = ref([])
let timer = null

const loader = async () => {
  try {
    return await listGatewayServers()
  } catch (e) {
    return { data: { items: [] } }
  }
}

const columns = [
  { prop: 'id', label: 'ID', width: 60 },
  { prop: 'name', label: '名称', minWidth: 120 },
  { label: '监听端口', width: 130, slot: 'address', showOverflowTooltip: false },
  { label: '协议', width: 100, slot: 'protocol' },
  { label: '默认设备', width: 100, slot: 'defaultDevices' },
  { label: '状态', width: 150, slot: 'status', showOverflowTooltip: false },
  { prop: 'conn_count', label: '连接数', width: 80 },
  { prop: 'poll_entries', label: '轮询项', width: 80 }
]

const operators = [
  {
    key: 'toggle',
    label: (row) => (row.running ? '停止' : '启动'),
    type: (row) => (row.running ? 'danger' : 'primary'),
    handler: (row) => (row.running ? stopServer(row) : startServer(row))
  },
  { key: 'connections', label: '连接', type: 'primary', handler: (row) => goConnections(row) },
  { key: 'edit', label: '编辑', type: 'primary', handler: (row) => openEdit(row) },
  { key: 'delete', label: '删除', type: 'danger', handler: (row) => delServer(row) }
]

async function loadModels() {
  try {
    const res = await searchModels({ page: 1, page_size: 100 })
    models.value = res.data || []
  } catch (e) { console.error(e) }
}

function startServer(row) {
  startGatewayServer(row.id).then(() => { ElMessage.success('启动请求已发送'); tableRef.value?.refresh() })
}
function stopServer(row) {
  stopGatewayServer(row.id).then(() => { ElMessage.success('已停止'); tableRef.value?.refresh() })
}
function goConnections(row) {
  router.push(`/gateway/servers/${row.id}/connections`)
}
async function delServer(row) {
  try {
    await ElMessageBox.confirm('确认删除？会同时停止运行实例', '提示', { type: 'warning' })
    await deleteGatewayServer(row.id)
    ElMessage.success('已删除')
    tableRef.value?.refresh()
  } catch (e) { /* cancel */ }
}

const showDialog = ref(false)
const editing = ref(null)
const formRef = ref(null)
const form = reactive({
  name: '', host: '0.0.0.0', port: 502, protocol: 'tcp',
  interval: 5, concurrency: 4, disabled: false,
  default_devices: []
})

function resetForm() {
  Object.assign(form, {
    name: '', host: '0.0.0.0', port: 502, protocol: 'tcp',
    interval: 5, concurrency: 4, disabled: false,
    default_devices: []
  })
}

function openCreate() {
  editing.value = null
  resetForm()
  showDialog.value = true
}
function openEdit(row) {
  editing.value = row
  resetForm()
  Object.assign(form, { ...row })
  // 固定全接口监听，忽略历史记录中的 host
  form.host = '0.0.0.0'
  // 后端 interval 为字符串（如 "5s"），表单使用数字（秒）
  if (typeof form.interval === 'string') {
    const m = form.interval.match(/(\d+)/)
    form.interval = m ? parseInt(m[1]) : 5
  }
  form.default_devices = (row.default_devices || []).map(d => ({ ...d }))
  showDialog.value = true
}

function addDefaultDevice() {
  form.default_devices.push({ name: '', address: '', model_id: '' })
}
function removeDefaultDevice(index) {
  form.default_devices.splice(index, 1)
}

function saveForm() {
  formRef.value.validate((valid) => {
    if (!valid) return
    const payload = {
      name: form.name.trim(),
      host: '0.0.0.0',
      port: form.port,
      protocol: form.protocol,
      interval: `${form.interval || 5}s`,
      concurrency: form.concurrency,
      disabled: form.disabled,
      default_devices: form.default_devices
        .filter(d => d.model_id && d.address)
        .map(d => ({ name: d.name || '', address: String(d.address), model_id: d.model_id }))
    }
    const req = editing.value ? updateGatewayServer(editing.value.id, payload) : createGatewayServer(payload)
    req.then(() => { showDialog.value = false; ElMessage.success('保存成功'); tableRef.value?.refresh() })
  })
}

onMounted(() => {
  loadModels()
  timer = setInterval(() => tableRef.value?.refresh(), 10000)
})
onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style lang="scss" scoped>
.gateway-server-page {
  .mono { font-family: 'Consolas', 'Menlo', monospace; }
  .default-devices-hint { color: #909399; font-size: 12px; margin-bottom: 8px; }
  .field-hint { margin-left: 8px; color: #909399; font-size: 12px; }
  .add-default-btn { margin-top: 8px; }
}
</style>
