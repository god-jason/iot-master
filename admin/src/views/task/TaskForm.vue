<template>
  <el-dialog
    :model-value="visible"
    :title="form.id ? '编辑定时任务' : '创建定时任务'"
    width="600px"
    top="8vh"
    @update:model-value="$emit('update:visible', $event)"
  >
    <el-form :model="form" label-width="100px">
      <el-form-item label="任务名称" required>
        <el-input v-model="form.name" placeholder="如：每天早上开窗帘" />
      </el-form-item>
      <el-form-item label="任务描述">
        <el-input v-model="form.desc" type="textarea" :rows="2" />
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
          ? '在线模式：由 master 服务器定时触发。'
          : '离线模式：由本地 MQTT 网关定时触发，断网仍按时执行。' }}
      </div>

      <el-form-item label="执行场景" required>
        <el-select
          v-model="form.scene_id"
          filterable
          remote
          reserve-keyword
          placeholder="搜索并选择场景"
          :remote-method="searchSceneOptions"
          :loading="sceneLoading"
          style="width: 100%"
        >
          <el-option
            v-for="s in sceneOptions"
            :key="s.id"
            :label="`${s.name}（${s.mode === 'offline' ? '离线' : '在线'}）`"
            :value="s.id"
          />
        </el-select>
        <div class="hint">任务模式建议与场景模式一致，否则可能在执行端找不到场景</div>
      </el-form-item>

      <el-form-item label="执行周期" required>
        <el-radio-group v-model="freq">
          <el-radio value="daily">每天</el-radio>
          <el-radio value="weekly">每周</el-radio>
          <el-radio value="custom">自定义Cron</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item v-if="freq !== 'custom'" label="执行时间" required>
        <el-time-picker
          v-model="timeValue"
          format="HH:mm"
          value-format="HH:mm"
          placeholder="选择时间"
          @change="syncCron"
        />
      </el-form-item>

      <el-form-item v-if="freq === 'weekly'" label="星期">
        <el-checkbox-group v-model="weekdays" @change="syncCron">
          <el-checkbox v-for="w in weekdayOptions" :key="w.value" :value="w.value">{{ w.label }}</el-checkbox>
        </el-checkbox-group>
      </el-form-item>

      <el-form-item v-if="freq === 'custom'" label="Cron表达式" required>
        <el-input v-model="form.cron" placeholder="分 时 日 月 周，如 0 9 * * 1-5" />
      </el-form-item>

      <el-form-item v-else label=" ">
        <span class="cron-preview">Cron：{{ form.cron || '（请选择执行时间）' }}</span>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:visible', false)">取消</el-button>
      <el-button type="primary" @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { searchScenes } from '@/api/scene'

const props = defineProps({
  visible: { type: Boolean, default: false },
  modelValue: { type: Object, default: () => ({}) }
})
const emit = defineEmits(['update:visible', 'submit'])

// 周几选项：cron 周字段 0=周日
const weekdayOptions = [
  { label: '周一', value: 1 },
  { label: '周二', value: 2 },
  { label: '周三', value: 3 },
  { label: '周四', value: 4 },
  { label: '周五', value: 5 },
  { label: '周六', value: 6 },
  { label: '周日', value: 0 }
]

const defaultForm = () => ({
  id: '',
  name: '',
  desc: '',
  mode: 'online',
  disabled: false,
  scene_id: '',
  cron: ''
})

const form = reactive(defaultForm())
const freq = ref('daily')
const timeValue = ref('08:00')
const weekdays = ref([1, 2, 3, 4, 5])
const sceneOptions = ref([])
const sceneLoading = ref(false)

const searchSceneOptions = async (keyword) => {
  sceneLoading.value = true
  try {
    const res = await searchScenes({ page: 1, page_size: 30, keyword })
    sceneOptions.value = res.data || []
  } finally {
    sceneLoading.value = false
  }
}

// 根据时间与星期生成 cron
const syncCron = () => {
  if (!timeValue.value) {
    form.cron = ''
    return
  }
  const [h, m] = timeValue.value.split(':').map(Number)
  if (freq.value === 'daily') {
    form.cron = `${m} ${h} * * *`
  } else if (freq.value === 'weekly') {
    if (weekdays.value.length === 0) {
      form.cron = ''
      return
    }
    const dow = [...weekdays.value].sort((a, b) => a - b).join(',')
    form.cron = `${m} ${h} * * ${dow}`
  }
}

// 解析已有 cron 回填可视化控件
const parseCron = (cron) => {
  const parts = (cron || '').trim().split(/\s+/)
  if (parts.length !== 5) {
    freq.value = 'custom'
    return
  }
  const [min, hour, dom, mon, dow] = parts
  const time = `${String(hour).padStart(2, '0')}:${String(min).padStart(2, '0')}`
  if (dom === '*' && mon === '*' && dow === '*' && /^\d+$/.test(min) && /^\d+$/.test(hour)) {
    freq.value = 'daily'
    timeValue.value = time
    return
  }
  if (dom === '*' && mon === '*' && /^[\d,]+$/.test(dow) && /^\d+$/.test(min) && /^\d+$/.test(hour)) {
    freq.value = 'weekly'
    timeValue.value = time
    weekdays.value = dow.split(',').map(Number)
    return
  }
  freq.value = 'custom'
}

watch(
  () => props.visible,
  async (open) => {
    if (!open) return
    const src = props.modelValue || {}
    Object.assign(form, JSON.parse(JSON.stringify(defaultForm())))
    freq.value = 'daily'
    timeValue.value = '08:00'
    weekdays.value = [1, 2, 3, 4, 5]
    sceneOptions.value = []

    await searchSceneOptions('')
    if (src.id) {
      Object.assign(form, JSON.parse(JSON.stringify(src)))
      parseCron(form.cron)
      if (form.scene_id && !sceneOptions.value.find(s => s.id === form.scene_id)) {
        sceneOptions.value.unshift({ id: form.scene_id, name: form.scene_id, mode: form.mode })
      }
    } else {
      syncCron()
    }
  }
)

const handleSubmit = () => {
  if (!form.name) return ElMessage.warning('请输入任务名称')
  if (!form.scene_id) return ElMessage.warning('请选择执行场景')
  if (freq.value !== 'custom') syncCron()
  if (!form.cron) return ElMessage.warning('请设置执行时间或 Cron 表达式')
  if (freq.value === 'weekly' && weekdays.value.length === 0) return ElMessage.warning('请至少选择一个星期')

  emit('submit', {
    id: form.id || undefined,
    name: form.name,
    desc: form.desc,
    mode: form.mode,
    disabled: form.disabled,
    scene_id: form.scene_id,
    cron: form.cron
  })
}
</script>

<style lang="scss" scoped>
.mode-tip {
  margin: -6px 0 12px 100px;
  color: #909399;
  font-size: 12px;
}

.hint {
  color: #999;
  font-size: 12px;
  line-height: 1.4;
}

.cron-preview {
  font-family: Consolas, Monaco, monospace;
  color: #606266;
  font-size: 13px;
}
</style>
