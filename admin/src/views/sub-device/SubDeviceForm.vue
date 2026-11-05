<template>
  <el-drawer v-model="localVisible" :title="isEdit ? '编辑设备' : '添加设备'" size="500px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="设备名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入设备名称" />
      </el-form-item>
      <el-form-item label="设备SN" prop="sn">
        <el-input v-model="form.sn" placeholder="留空自动生成" />
      </el-form-item>
      <el-form-item label="所属物模型" prop="model_id">
        <el-select
          v-model="form.model_id"
          placeholder="请选择物模型"
          filterable
          style="width: 100%"
          clearable
        >
          <el-option
            v-for="item in models"
            :key="item.id"
            :label="item.name"
            :value="item.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="协议地址" prop="address">
        <el-input v-model="form.address" placeholder="如 1" />
      </el-form-item>
      <el-form-item label="禁用">
        <el-switch v-model="form.disabled" />
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
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { searchModels } from '@/api/model'

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
const models = ref([])

const loadModels = async () => {
  try {
    const res = await searchModels({ page: 1, page_size: 100 })
    models.value = res.data || []
  } catch (e) { console.error(e) }
}

const form = reactive({
  name: '',
  sn: '',
  model_id: null,
  address: '',
  disabled: false
})

const rules = {
  name: [{ required: true, message: '请输入设备名称', trigger: 'blur' }],
  model_id: [{ required: true, message: '请选择物模型', trigger: 'change' }],
  address: [{ required: true, message: '请输入协议地址', trigger: 'blur' }]
}

watch(() => props.visible, (val) => {
  localVisible.value = val
  if (val) {
    isEdit.value = !!props.modelValue.id
    Object.assign(form, {
      name: props.modelValue.name || '',
      sn: props.modelValue.sn || '',
      model_id: props.modelValue.model_id || null,
      address: props.modelValue.address || '',
      disabled: !!props.modelValue.disabled
    })
    loadModels()
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
    await emit('submit', { ...form })
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
