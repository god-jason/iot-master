<template>
  <el-drawer v-model="localVisible" :title="isEdit ? '编辑物模型' : '创建物模型'" size="500px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="物模型图标">
        <ImageUpload v-model="form.icon" :is-icon="true" tip="建议使用128x128正方形图片" />
      </el-form-item>
      <el-form-item label="物模型名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入物模型名称" />
      </el-form-item>
      <el-form-item label="物模型ID" prop="id">
        <el-input
          v-model="form.id"
          placeholder="留空则自动生成，也可自定义"
          maxlength="32"
          show-word-limit
        />
      </el-form-item>
      <el-form-item label="物模型秘钥" prop="secret">
        <el-input v-model="form.secret" placeholder="请输入物模型秘钥" />
      </el-form-item>
      <el-form-item label="设备类型" prop="type">
        <el-select v-model="form.type" placeholder="请选择设备类型" style="width: 100%">
          <el-option label="网关" value="gateway" />
          <el-option label="直连设备" value="direct" />
          <el-option label="子设备" value="sub" />
        </el-select>
      </el-form-item>
      <el-form-item label="协议" prop="protocol">
        <el-select v-model="form.protocol" placeholder="请选择协议" style="width: 100%" clearable>
          <el-option v-for="p in protocolOptions" :key="p.id" :label="p.name" :value="p.name" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <div style="display: flex; justify-content: flex-end; gap: 10px;">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </div>
    </template>
  </el-drawer>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import ImageUpload from '@/components/ImageUpload.vue'
import { searchProtocols } from '@/api/protocol'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  modelValue: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['update:visible', 'submit'])

const formRef = ref(null)
const submitting = ref(false)
const isEdit = ref(false)
const localVisible = ref(false)
const protocols = ref([])

// 内置协议：直连设备用 mqtt / custom，排在协议列表最前面
const builtinProtocols = [
  { id: 'mqtt', name: 'mqtt' },
  { id: 'custom', name: 'custom' }
]

const protocolOptions = computed(() => [...builtinProtocols, ...protocols.value])

const loadProtocols = async () => {
  try {
    const res = await searchProtocols({ page: 1, page_size: 100 })
    protocols.value = res.data || []
  } catch (e) { console.error(e) }
}

const form = reactive({
  id: '',
  icon: '',
  name: '',
  secret: '',
  type: '',
  protocol: ''
})

const rules = {
  name: [{ required: true, message: '请输入物模型名称', trigger: 'blur' }],
  // 物模型ID 允许留空（创建时由后端自动生成），填写时只允许字母、数字、下划线、中划线，长度不超过 32
  id: [{
    validator: (rule, value, callback) => {
      if (value && !/^[A-Za-z0-9_-]{1,32}$/.test(value)) {
        callback(new Error('仅支持字母、数字、下划线、中划线，长度不超过32'))
      } else {
        callback()
      }
    },
    trigger: ['blur', 'change']
  }],
  secret: [{ required: true, message: '请输入物模型秘钥', trigger: 'blur' }],
  type: [{ required: true, message: '请选择设备类型', trigger: 'change' }]
}

watch(() => props.visible, (val) => {
  localVisible.value = val
  if (val) {
    isEdit.value = !!props.modelValue.id
    Object.assign(form, {
      id: props.modelValue.id || '',
      icon: props.modelValue.icon || '',
      name: props.modelValue.name || '',
      secret: props.modelValue.secret || '',
      type: props.modelValue.type || '',
      protocol: props.modelValue.protocol || ''
    })
    loadProtocols()
  }
})

watch(localVisible, (val) => {
  if (!val) {
    emit('update:visible', false)
  }
})

const handleClose = () => {
  localVisible.value = false
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    await emit('submit', { ...form, id: form.id.trim() })
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    emit('update:visible', false)
  } catch (e) {
    console.error(e)
    ElMessage.error('操作失败')
  } finally {
    submitting.value = false
  }
}
</script>
