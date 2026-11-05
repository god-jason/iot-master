<template>
  <el-dialog
    :model-value="visible"
    :title="form.id ? '编辑规则' : '创建规则'"
    width="820px"
    top="6vh"
    @update:model-value="$emit('update:visible', $event)"
    @closed="handleClosed"
  >
    <el-form ref="formRef" :model="form" label-width="100px" class="rule-form">
      <!-- 基本信息 -->
      <div class="section-title">基本信息</div>
      <el-form-item label="规则名称" required>
        <el-input v-model="form.name" placeholder="请输入规则名称" />
      </el-form-item>
      <el-form-item label="规则描述">
        <el-input v-model="form.desc" type="textarea" :rows="2" placeholder="规则用途说明" />
      </el-form-item>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="优先级">
            <el-input-number v-model="form.priority" :min="0" :max="9999" controls-position="right" />
            <span class="hint">数字越小越先执行</span>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="禁用">
            <el-switch v-model="form.disabled" />
          </el-form-item>
        </el-col>
      </el-row>

      <!-- 作用范围 -->
      <div class="section-title">作用范围</div>
      <el-form-item label="范围类型" required>
        <el-radio-group v-model="form.scope.type">
          <el-radio-button value="single">单个设备</el-radio-button>
          <el-radio-button value="model">物模型</el-radio-button>
          <el-radio-button value="group">设备分组</el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item v-if="form.scope.type === 'single'" label="选择设备" required>
        <el-select
          v-model="form.scope.device_id"
          filterable
          remote
          reserve-keyword
          placeholder="输入设备名称搜索"
          :remote-method="searchDeviceOptions"
          :loading="optionLoading"
          style="width: 100%"
        >
          <el-option v-for="d in deviceOptions" :key="d.id" :label="d.name || d.id" :value="d.id" />
        </el-select>
      </el-form-item>
      <el-form-item v-if="form.scope.type === 'model'" label="选择物模型" required>
        <el-select
          v-model="form.scope.model_id"
          filterable
          remote
          reserve-keyword
          placeholder="输入物模型名称搜索"
          :remote-method="searchModelOptions"
          :loading="optionLoading"
          style="width: 100%"
        >
          <el-option v-for="m in modelOptions" :key="m.id" :label="m.name" :value="m.id" />
        </el-select>
      </el-form-item>
      <el-form-item v-if="form.scope.type === 'group'" label="选择分组" required>
        <el-select
          v-model="form.scope.group_id"
          filterable
          remote
          reserve-keyword
          placeholder="输入分组名称搜索"
          :remote-method="searchGroupOptions"
          :loading="optionLoading"
          style="width: 100%"
        >
          <el-option v-for="g in groupOptions" :key="g.id" :label="g.name" :value="g.id" />
        </el-select>
      </el-form-item>

      <!-- 触发方式 -->
      <div class="section-title">触发方式</div>
      <el-form-item label="触发类型" required>
        <el-select v-model="form.trigger.type" style="width: 220px">
          <el-option label="属性上报" value="property" />
          <el-option label="设备事件" value="event" />
          <el-option label="上下线" value="status" />
          <el-option label="定时触发" value="schedule" />
        </el-select>
      </el-form-item>
      <el-form-item v-if="form.trigger.type === 'property'" label="属性KEY">
        <el-input v-model="form.trigger.property" placeholder="留空表示任一属性上报即触发" style="width: 360px" />
      </el-form-item>
      <el-form-item v-if="form.trigger.type === 'event'" label="事件名称">
        <el-input v-model="form.trigger.event_name" placeholder="留空表示所有事件都触发" style="width: 360px" />
      </el-form-item>
      <el-form-item v-if="form.trigger.type === 'status'" label="状态">
        <el-radio-group v-model="form.trigger.status">
          <el-radio value="online">上线</el-radio>
          <el-radio value="offline">下线</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item v-if="form.trigger.type === 'schedule'" label="Cron表达式" required>
        <el-input v-model="form.trigger.cron" placeholder="分 时 日 月 周，如 0 9 * * *（每天9点）" style="width: 360px" />
        <span class="hint">5 段标准 cron，按分钟精度调度</span>
      </el-form-item>

      <!-- 条件 -->
      <div class="section-title">
        触发条件
        <el-radio-group v-model="condMode" size="small" style="margin-left: 12px">
          <el-radio-button value="structured">结构化条件</el-radio-button>
          <el-radio-button value="expr">表达式</el-radio-button>
        </el-radio-group>
      </div>

      <template v-if="condMode === 'structured'">
        <el-form-item label="组合逻辑">
          <el-radio-group v-model="form.logic">
            <el-radio-button value="and">全部满足(AND)</el-radio-button>
            <el-radio-button value="or">任一满足(OR)</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <div v-for="(c, i) in form.conditions" :key="i" class="cond-row">
          <el-input v-model="c.field" placeholder="属性KEY" style="width: 150px" />
          <el-select v-model="c.op" style="width: 120px">
            <el-option label="&gt;" value=">" />
            <el-option label="&gt;=" value=">=" />
            <el-option label="&lt;" value="<" />
            <el-option label="&lt;=" value="<=" />
            <el-option label="==" value="==" />
            <el-option label="!=" value="!=" />
            <el-option label="区间内" value="between" />
            <el-option label="区间外" value="out_of" />
          </el-select>
          <template v-if="c.op === 'between' || c.op === 'out_of'">
            <el-input-number controls-position="right" v-model="c.min" :controls="false" placeholder="最小值" style="width: 110px" />
            <span style="margin: 0 4px">~</span>
            <el-input-number controls-position="right" v-model="c.max" :controls="false" placeholder="最大值" style="width: 110px" />
          </template>
          <el-input v-else v-model="c.value" placeholder="比较值" style="width: 150px" />
          <el-tooltip content="边沿触发：仅在由不满足跨越到满足时命中" placement="top">
            <el-checkbox v-model="c.edge">边沿</el-checkbox>
          </el-tooltip>
          <el-button link type="danger" @click="form.conditions.splice(i, 1)">删除</el-button>
        </div>
        <el-button text type="primary" @click="addCondition">
          <el-icon><Plus /></el-icon> 添加条件
        </el-button>
      </template>

      <template v-else>
        <el-form-item label="表达式">
          <el-input
            v-model="form.expr"
            type="textarea"
            :rows="4"
            placeholder="如：value.temperature > 30 && value.humidity < 60"
          />
          <span class="hint">
            可用变量：value（当前数据）、prev（上次值）、device_id、ts；非空时优先于结构化条件
          </span>
        </el-form-item>
      </template>

      <!-- 动作 -->
      <div class="section-title">执行动作</div>
      <div v-for="(a, i) in form.actions" :key="i" class="action-block">
        <div class="action-head">
          <el-select v-model="a.type" style="width: 200px" @change="onActionTypeChange(a)">
            <el-option label="异常报警" value="alarm" />
            <el-option label="下发属性" value="control_property" />
            <el-option label="调用服务" value="control_service" />
            <el-option label="通知消息" value="notify" />
            <el-option label="数据加工" value="data" />
          </el-select>
          <el-button link type="danger" @click="form.actions.splice(i, 1)">删除动作</el-button>
        </div>

        <!-- 报警 / 通知 -->
        <template v-if="a.type === 'alarm'">
          <el-form-item label="告警级别" class="inner-item">
            <el-select v-model="a.params.level" style="width: 180px">
              <el-option label="提醒" value="alert" />
              <el-option label="警告" value="warning" />
              <el-option label="严重" value="critical" />
            </el-select>
          </el-form-item>
          <el-form-item label="消息内容" class="inner-item">
            <el-input v-model="a.params.message" placeholder="留空则使用规则名称" />
          </el-form-item>
        </template>
        <template v-if="a.type === 'notify'">
          <el-form-item label="消息内容" class="inner-item">
            <el-input v-model="a.params.message" placeholder="通知内容" />
          </el-form-item>
        </template>

        <!-- 下发属性 -->
        <template v-if="a.type === 'control_property'">
          <div v-for="(p, pi) in a._props" :key="pi" class="cond-row">
            <el-input v-model="p.key" placeholder="属性KEY" style="width: 180px" />
            <el-input v-model="p.value" placeholder="属性值" style="width: 200px" />
            <el-button link type="danger" @click="a._props.splice(pi, 1)">删除</el-button>
          </div>
          <el-button text type="primary" @click="a._props.push({ key: '', value: '' })">
            <el-icon><Plus /></el-icon> 添加属性
          </el-button>
        </template>

        <!-- 调用服务 -->
        <template v-if="a.type === 'control_service'">
          <el-form-item label="服务名" class="inner-item">
            <el-input v-model="a.params.service" placeholder="物模型服务标识" />
          </el-form-item>
          <el-form-item label="入参JSON" class="inner-item">
            <el-input v-model="a._paramsText" type="textarea" :rows="2" placeholder='可选，如 {"mode":1}' />
          </el-form-item>
        </template>

        <!-- 数据加工 -->
        <template v-if="a.type === 'data'">
          <el-form-item label="衍生字段" class="inner-item">
            <el-input v-model="a.params.field" placeholder="写入时序库的字段名" />
          </el-form-item>
          <el-form-item label="取值来源" class="inner-item">
            <el-input v-model="a.params.source" placeholder="从触发数据哪个字段取值（留空同衍生字段）" />
          </el-form-item>
          <el-form-item label="固定值" class="inner-item">
            <el-input v-model="a.params.value" placeholder="留空则取“取值来源”的实时值" />
          </el-form-item>
        </template>
      </div>
      <el-button text type="primary" @click="addAction">
        <el-icon><Plus /></el-icon> 添加动作
      </el-button>

      <!-- 防抖限流 -->
      <div class="section-title">防抖与限流</div>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="去抖(秒)">
            <el-input-number v-model="form.debounce" :min="0" controls-position="right" />
            <span class="hint">持续满足该时长才告警</span>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="重复间隔">
            <el-input-number v-model="form.throttle" :min="0" controls-position="right" />
            <span class="hint">0 表示只告警一次</span>
          </el-form-item>
        </el-col>
      </el-row>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:visible', false)">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { searchDevices } from '@/api/device'
import { searchModels } from '@/api/model'
import { searchGroups } from '@/api/group'

const props = defineProps({
  visible: { type: Boolean, default: false },
  modelValue: { type: Object, default: () => ({}) }
})
const emit = defineEmits(['update:visible', 'submit'])

const formRef = ref(null)
const saving = ref(false)
const condMode = ref('structured')
const optionLoading = ref(false)
const deviceOptions = ref([])
const modelOptions = ref([])
const groupOptions = ref([])

// 默认空表单
const defaultForm = () => ({
  id: '',
  name: '',
  desc: '',
  disabled: false,
  priority: 100,
  scope: { type: 'single', device_id: '', model_id: '', group_id: '' },
  trigger: { type: 'property', property: '', event_name: '', status: 'online', cron: '' },
  expr: '',
  conditions: [],
  logic: 'and',
  actions: [],
  debounce: 0,
  throttle: 0
})

const form = reactive(defaultForm())

// 远程搜索：设备/物模型/分组
const searchDeviceOptions = async (keyword) => {
  optionLoading.value = true
  try {
    const res = await searchDevices({ page: 1, page_size: 30, keyword })
    deviceOptions.value = res.data || []
  } finally {
    optionLoading.value = false
  }
}
const searchModelOptions = async (keyword) => {
  optionLoading.value = true
  try {
    const res = await searchModels({ page: 1, page_size: 30, keyword })
    modelOptions.value = res.data || []
  } finally {
    optionLoading.value = false
  }
}
const searchGroupOptions = async (keyword) => {
  optionLoading.value = true
  try {
    const res = await searchGroups({ page: 1, page_size: 30, keyword })
    groupOptions.value = res.data || []
  } finally {
    optionLoading.value = false
  }
}

// 条件/动作操作
const addCondition = () => {
  form.conditions.push({ field: '', op: '>', value: '', min: null, max: null, edge: false })
}
const addAction = () => {
  const a = { type: 'alarm', params: { level: 'alert', message: '' }, _props: [], _paramsText: '' }
  form.actions.push(a)
}
const onActionTypeChange = (a) => {
  a.params = {}
  a._props = []
  a._paramsText = ''
  if (a.type === 'alarm') a.params = { level: 'alert', message: '' }
}

// 尝试把字符串值解析为数字/布尔
const parseValue = (v) => {
  if (v === '' || v === null || v === undefined) return v
  if (v === 'true') return true
  if (v === 'false') return false
  const n = Number(v)
  return Number.isNaN(n) ? v : n
}

watch(
  () => props.visible,
  (open) => {
    if (!open) return
    const src = props.modelValue || {}
    const base = defaultForm()
    Object.assign(form, JSON.parse(JSON.stringify(base)))
    if (src.id) {
      Object.assign(form, JSON.parse(JSON.stringify(src)))
      form.scope = Object.assign({ type: 'single', device_id: '', model_id: '', group_id: '' }, src.scope || {})
      form.trigger = Object.assign({ type: 'property', property: '', event_name: '', status: 'online', cron: '' }, src.trigger || {})
      form.conditions = (src.conditions || []).map((c) => {
        const row = { ...c, min: null, max: null, value: c.value ?? '' }
        if ((c.op === 'between' || c.op === 'out_of') && Array.isArray(c.value)) {
          row.min = c.value[0]
          row.max = c.value[1]
        }
        return row
      })
      form.actions = (src.actions || []).map((a) => {
        const item = { type: a.type, params: { ...(a.params || {}) }, _props: [], _paramsText: '' }
        if (a.type === 'control_property' && a.params.properties) {
          item._props = Object.entries(a.params.properties).map(([key, value]) => ({ key, value: String(value) }))
        }
        if (a.type === 'control_service' && a.params.params) {
          item._paramsText = JSON.stringify(a.params.params)
        }
        return item
      })
      condMode.value = src.expr ? 'expr' : 'structured'
      // 编辑时预载当前选中项，保证回显名称
      if (form.scope.type === 'single' && form.scope.device_id) {
        deviceOptions.value = [{ id: form.scope.device_id, name: form.scope.device_id }]
      }
      if (form.scope.type === 'model' && form.scope.model_id) {
        modelOptions.value = [{ id: form.scope.model_id, name: form.scope.model_id }]
      }
      if (form.scope.type === 'group' && form.scope.group_id) {
        groupOptions.value = [{ id: form.scope.group_id, name: form.scope.group_id }]
      }
    } else {
      condMode.value = 'structured'
    }
  }
)

const handleSubmit = async () => {
  if (!form.name) {
    ElMessage.warning('请输入规则名称')
    return
  }
  if (form.scope.type === 'single' && !form.scope.device_id) return ElMessage.warning('请选择设备')
  if (form.scope.type === 'model' && !form.scope.model_id) return ElMessage.warning('请选择物模型')
  if (form.scope.type === 'group' && !form.scope.group_id) return ElMessage.warning('请选择分组')
  if (form.trigger.type === 'schedule' && !form.trigger.cron) return ElMessage.warning('请输入 Cron 表达式')

  // 组装提交数据（剥离 UI 辅助字段）
  const payload = JSON.parse(JSON.stringify(form))
  payload.conditions = form.conditions
    .filter((c) => c.field)
    .map((c) => {
      if (c.op === 'between' || c.op === 'out_of') {
        return { field: c.field, op: c.op, edge: !!c.edge, value: [c.min, c.max] }
      }
      return { field: c.field, op: c.op, edge: !!c.edge, value: parseValue(c.value) }
    })
  payload.expr = condMode.value === 'expr' ? form.expr.trim() : ''
  payload.actions = form.actions.map((a) => {
    const act = { type: a.type, params: { ...a.params } }
    if (a.type === 'control_property') {
      const props = {}
      a._props.filter((p) => p.key).forEach((p) => { props[p.key] = parseValue(p.value) })
      act.params = { properties: props }
    }
    if (a.type === 'control_service') {
      act.params = { service: a.params.service || '' }
      if (a._paramsText.trim()) {
        try {
          act.params.params = JSON.parse(a._paramsText)
        } catch (e) {
          throw new Error('服务入参 JSON 格式错误')
        }
      } else {
        act.params.params = {}
      }
    }
    if (a.type === 'data' && act.params.value !== undefined && act.params.value !== '') {
      act.params.value = parseValue(act.params.value)
    }
    return act
  })

  if (payload.actions.length === 0) {
    ElMessage.warning('请至少添加一个动作')
    return
  }

  saving.value = true
  try {
    emit('submit', payload)
  } finally {
    saving.value = false
  }
}

const handleClosed = () => {
  formRef.value?.clearValidate?.()
}
</script>

<style lang="scss" scoped>
.rule-form {
  max-height: 68vh;
  overflow-y: auto;
  padding-right: 8px;

  .section-title {
    font-size: 15px;
    font-weight: 600;
    color: #1a1a2e;
    margin: 18px 0 14px;
    padding-left: 10px;
    border-left: 3px solid #0099ff;
  }

  .hint {
    margin-left: 10px;
    color: #999;
    font-size: 12px;
  }

  .cond-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 10px;
  }

  .action-block {
    border: 1px solid #ebeef5;
    border-radius: 8px;
    padding: 14px 14px 0;
    margin-bottom: 12px;
    background: #fafbfc;

    .action-head {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 12px;
    }

    .inner-item {
      margin-bottom: 12px;
    }
  }
}
</style>
