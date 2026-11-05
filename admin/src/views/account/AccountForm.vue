<template>
  <el-drawer v-model="localVisible" :title="isEdit ? '编辑账号' : '添加账号'" size="500px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="头像" v-if="isEdit">
        <ImageUpload v-model="form.avatar" tip="建议使用正方形头像" />
      </el-form-item>
      <el-form-item label="账号名称" prop="username">
        <el-input v-model="form.username" :disabled="isEdit" />
      </el-form-item>
      <el-form-item v-if="!isEdit" label="密码" prop="password">
        <el-input v-model="form.password" type="password" placeholder="请输入密码" />
      </el-form-item>
      <el-form-item label="昵称" prop="nickname">
        <el-input v-model="form.nickname" placeholder="请输入昵称" />
      </el-form-item>
      <el-form-item label="所属组织" prop="org_id">
        <el-tree-select
          v-model="form.org_id"
          :data="orgOptions"
          check-strictly
          :render-after-expand="false"
          placeholder="请选择所属组织"
          style="width: 100%"
        />
      </el-form-item>
      <el-form-item label="账号类型" prop="platform">
        <el-radio-group v-model="form.platform">
          <el-radio :label="false">用户账号</el-radio>
          <el-radio :label="true">平台账号</el-radio>
        </el-radio-group>
        <div class="account-type-tip">
          平台账号可登录管理后台，进行组织管理、权限分配与设备维护；用户账号仅可通过移动端使用。
        </div>
      </el-form-item>
      <el-form-item label="邮箱">
        <el-input v-model="form.email" placeholder="请输入邮箱" />
      </el-form-item>
      <el-form-item label="手机号">
        <el-input v-model="form.cellphone" placeholder="请输入手机号" />
      </el-form-item>
      <el-form-item label="状态" v-if="isEdit">
        <el-switch v-model="form.disabled" :active-text="form.disabled ? '禁用' : '正常'" />
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
import { ref, reactive, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import ImageUpload from '@/components/ImageUpload.vue'
import { getOrgTree } from '@/api/org'

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

const form = reactive({
  username: '',
  password: '',
  nickname: '',
  email: '',
  cellphone: '',
  avatar: '',
  disabled: false,
  org_id: 0,
  platform: false
})

const orgOptions = ref([])

const loadOrgs = async () => {
  try {
    const res = await getOrgTree()
    const convert = (nodes) =>
      (nodes || []).map((n) => ({ value: n.id, label: n.name, children: convert(n.children) }))
    orgOptions.value = convert(res.data)
  } catch (error) {
    console.error('加载组织失败', error)
  }
}

onMounted(loadOrgs)

const rules = {
  username: [{ required: true, message: '请输入账号名称', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  org_id: [{ required: true, message: '请选择所属组织', trigger: 'change' }]
}

watch(() => props.visible, (val) => {
  localVisible.value = val
  if (val) {
    isEdit.value = !!props.modelValue.id
    Object.assign(form, {
      username: props.modelValue.username || '',
      password: '',
      nickname: props.modelValue.nickname || '',
      email: props.modelValue.email || '',
      cellphone: props.modelValue.cellphone || '',
      avatar: props.modelValue.avatar || '',
      disabled: props.modelValue.disabled || false,
      org_id: (props.modelValue.org_ids || [])[0] || 0,
      platform: props.modelValue.admin ? true : (props.modelValue.platform || false)
    })
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

<style lang="scss" scoped>
.account-type-tip {
  width: 100%;
  margin-top: 4px;
  color: #909399;
  font-size: 12px;
  line-height: 1.5;
}
</style>
