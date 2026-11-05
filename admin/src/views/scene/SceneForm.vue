<template>
  <el-dialog
    :model-value="visible"
    :title="form.id ? '编辑场景' : '创建场景'"
    width="780px"
    top="6vh"
    @update:model-value="$emit('update:visible', $event)"
  >
    <el-form ref="formRef" :model="form" label-width="100px" class="scene-form">
      <!-- 基本信息 -->
      <div class="section-title">基本信息</div>
      <el-form-item label="场景名称" required>
        <el-input v-model="form.name" placeholder="如：回家模式、睡眠模式" />
      </el-form-item>
      <el-form-item label="场景描述">
        <el-input v-model="form.desc" type="textarea" :rows="2" placeholder="场景用途说明" />
      </el-form-item>
      <el-row :gutter="16">
        <el-col :span="14">
          <el-form-item label="运行模式" required>
            <el-radio-group v-model="form.mode">
              <el-radio-button value="online">在线模式</el-radio-button>
              <el-radio-button value="offline">离线模式</el-radio-button>
            </el-radio-group>
          </el-form-item>
        </el-col>
        <el-col :span="10">
          <el-form-item label="禁用">
            <el-switch v-model="form.disabled" />
          </el-form-item>
        </el-col>
      </el-row>
      <div class="mode-tip">
        {{ form.mode === 'online'
          ? '在线模式：由 master 服务器执行，服务器运行时生效。'
          : '离线模式：由本地 MQTT 网关执行，断网/服务器不可用时仍可在局域网内执行。' }}
      </div>

      <!-- 动作 -->
      <div class="section-title">
        设备动作
        <span class="hint">按顺序执行</span>
      </div>
      <div v-for="(a, i) in form.actions" :key="i" class="action-block">
        <div class="action-head">
          <span class="action-index">动作 {{ i + 1 }}</span>
          <el-button link type="danger" @click="form.actions.splice(i, 1)">删除动作</el-button>
        </div>

        <el-form-item label="目标设备" class="inner-item">
          <el-select
            v-model="a.device_id"
            filterable
            remote
            reserve-keyword
            placeholder="输入设备名称搜索"
            :remote-method="(kw) => searchDevice(a, kw)"
            :loading="a.loading"
            style="width: 100%"
            @change="(id) => onDeviceChange(a, id)"
          >
            <el-option v-for="d in a.deviceOptions" :key="d.id" :label="d.name || d.sn || d.id" :value="d.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="动作类型" class="inner-item">
          <el-radio-group v-model="a.type">
            <el-radio-button value="property">下发属性</el-radio-button>
            <el-radio-button value="service">调用服务</el-radio-button>
          </el-radio-group>
        </el-form-item>

        <!-- 下发属性 -->
        <template v-if="a.type === 'property'">
          <div v-for="(p, pi) in a._props" :key="pi" class="prop-row">
            <el-select v-model="p.key" placeholder="选择属性" style="width: 200px">
              <el-option
                v-for="prop in a.properties"
                :key="prop.key"
                :label="`${prop.name || prop.key} (${prop.key})`"
                :value="prop.key"
              />
            </el-select>
            <el-input v-model="p.value" placeholder="属性值" style="width: 200px" />
            <el-button link type="danger" @click="a._props.splice(pi, 1)">删除</el-button>
          </div>
          <el-button text type="primary" @click="a._props.push({ key: '', value: '' })">
            <el-icon><Plus /></el-icon> 添加属性
          </el-button>
        </template>

        <!-- 调用服务 -->
        <template v-if="a.type === 'service'">
          <el-form-item label="服务" class="inner-item">
            <el-select v-model="a.service" placeholder="选择设备服务" style="width: 100%">
              <el-option
                v-for="srv in a.services"
                :key="srv.key"
                :label="`${srv.name || srv.key} (${srv.key})`"
                :value="srv.key"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="入参JSON" class="inner-item">
            <el-input v-model="a._paramsText" type="textarea" :rows="2" placeholder='可选，如 {"mode":1}' />
          </el-form-item>
        </template>
      </div>

      <el-button text type="primary" @click="addAction">
        <el-icon><Plus /></el-icon> 添加动作
      </el-button>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:visible', false)">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { searchDevices } from '@/api/device'
import { getModel } from '@/api/model'

const props = defineProps({
  visible: { type: Boolean, default: false },
  modelValue: { type: Object, default: () => ({}) },
  familyId: { type: String, default: '' }
})
const emit = defineEmits(['update:visible', 'submit'])

let savingFlag = false

const defaultForm = () => ({
  id: '',
  name: '',
  desc: '',
  family_id: '',
  mode: 'online',
  disabled: false,
  actions: []
})

const form = reactive(defaultForm())

// 新增动作行
const addAction = () => {
  form.actions.push({
    device_id: '',
    type: 'property',
    properties: [],
    services: [],
    deviceOptions: [],
    loading: false,
    _props: [{ key: '', value: '' }],
    _paramsText: ''
  })
}

// 远程搜索设备
const searchDevice = async (action, keyword) => {
  action.loading = true
  try {
    const res = await searchDevices({ page: 1, page_size: 30, keyword })
    action.deviceOptions = res.data || []
  } finally {
    action.loading = false
  }
}

// 设备选中后加载物模型的属性与服务定义
const onDeviceChange = async (action, deviceID) => {
  action.properties = []
  action.services = []
  if (!deviceID) return
  const opt = action.deviceOptions.find(d => d.id === deviceID)
  const modelID = opt?.model_id
  if (!modelID) return
  try {
    const res = await getModel(modelID)
    const model = res.data || {}
    action.properties = model.properties || []
    action.services = model.services || []
  } catch (e) {
    // 物模型加载失败时仍允许手工填写
  }
}

// 值类型转换：数字/布尔按字面量，其余保持字符串
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
    Object.assign(form, JSON.parse(JSON.stringify(defaultForm())))
    // 从家庭管理页进入时自动归属当前家庭
    if (props.familyId) form.family_id = props.familyId
    if (!src.id) {
      addAction()
      return
    }
    Object.assign(form, JSON.parse(JSON.stringify(src)))
    // 还原 UI 辅助字段
    form.actions = (src.actions || []).map(a => {
      const item = {
        device_id: a.device_id || '',
        type: a.type || 'property',
        properties: [],
        services: [],
        deviceOptions: a.device_id ? [{ id: a.device_id, name: a.device_id }] : [],
        loading: false,
        _props: [],
        _paramsText: ''
      }
      if (a.type === 'property') {
        item._props = Object.entries(a.properties || {}).map(([key, value]) => ({ key, value: String(value) }))
        if (item._props.length === 0) item._props = [{ key: '', value: '' }]
      }
      if (a.type === 'service') {
        item.service = a.service || ''
        if (a.params) item._paramsText = JSON.stringify(a.params)
      }
      // 异步补全属性/服务选项，不阻塞回显
      if (item.device_id) onDeviceChange(item, item.device_id)
      return item
    })
  }
)

const handleSubmit = async () => {
  if (savingFlag) return
  if (!form.name) {
    ElMessage.warning('请输入场景名称')
    return
  }
  if (form.actions.length === 0) {
    ElMessage.warning('请至少添加一个动作')
    return
  }

  const actions = []
  for (const a of form.actions) {
    if (!a.device_id) {
      ElMessage.warning('请为每个动作选择目标设备')
      return
    }
    if (a.type === 'property') {
      const properties = {}
      a._props.filter(p => p.key).forEach(p => { properties[p.key] = parseValue(p.value) })
      if (Object.keys(properties).length === 0) {
        ElMessage.warning('请至少配置一个下发属性')
        return
      }
      actions.push({ device_id: a.device_id, type: 'property', properties })
    } else {
      if (!a.service) {
        ElMessage.warning('请选择要调用的服务')
        return
      }
      const act = { device_id: a.device_id, type: 'service', service: a.service, params: {} }
      if (a._paramsText.trim()) {
        try {
          act.params = JSON.parse(a._paramsText)
        } catch (e) {
          ElMessage.warning('服务入参 JSON 格式错误')
          return
        }
      }
      actions.push(act)
    }
  }

  savingFlag = true
  try {
    emit('submit', {
      id: form.id || undefined,
      name: form.name,
      desc: form.desc,
      mode: form.mode,
      disabled: form.disabled,
      family_id: form.family_id || undefined,
      actions
    })
  } finally {
    savingFlag = false
  }
}
</script>

<style lang="scss" scoped>
.scene-form {
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
    font-weight: 400;
  }

  .mode-tip {
    margin: -6px 0 10px 100px;
    color: #909399;
    font-size: 12px;
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

      .action-index {
        font-weight: 600;
        color: #0099ff;
      }
    }

    .inner-item {
      margin-bottom: 12px;
    }
  }

  .prop-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 10px;
  }
}
</style>
