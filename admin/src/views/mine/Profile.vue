<template>
  <div class="profile-page">
    <div class="page-header">
      <h2 class="page-title">个人中心</h2>
    </div>
    
    <el-row :gutter="24">
      <el-col :span="16">
        <el-card class="profile-card">
          <template #header>
            <div class="card-header">
              <span>基本信息</span>
            </div>
          </template>
          <el-form :model="form" label-width="100px">
            <el-form-item label="用户名">
              <el-input v-model="form.username" disabled />
            </el-form-item>
            <el-form-item label="昵称">
              <el-input v-model="form.nickname" placeholder="请输入昵称" />
            </el-form-item>
            <el-form-item label="邮箱">
              <el-input v-model="form.email" />
            </el-form-item>
            <el-form-item label="手机号">
              <el-input v-model="form.cellphone" />
            </el-form-item>
            <el-form-item label="头像">
              <ImageUpload v-model="form.avatar" tip="建议使用正方形头像，自动裁剪为200x200" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleSave">保存修改</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
      
      <el-col :span="8">
        <el-card class="password-card">
          <template #header>
            <div class="card-header">
              <span>修改密码</span>
            </div>
          </template>
          <el-form :model="passwordForm" label-width="80px">
            <el-form-item label="原密码">
              <el-input v-model="passwordForm.oldPassword" type="password" show-password />
            </el-form-item>
            <el-form-item label="新密码">
              <el-input v-model="passwordForm.newPassword" type="password" show-password />
            </el-form-item>
            <el-form-item label="确认密码">
              <el-input v-model="passwordForm.confirmPassword" type="password" show-password />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleChangePassword">修改密码</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/store'
import ImageUpload from '@/components/ImageUpload.vue'
import { getMe, updateMe, changePassword } from '@/api/user'

const router = useRouter()

const userStore = useUserStore()

const loading = ref(false)

const form = ref({
  username: '',
  nickname: '',
  email: '',
  cellphone: '',
  avatar: ''
})

const passwordForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const loadUser = async () => {
  loading.value = true
  try {
    const res = await getMe()
    const user = res.data
    form.value = {
      username: user.username || '',
      nickname: user.nickname || '',
      email: user.email || '',
      cellphone: user.cellphone || '',
      avatar: user.avatar || ''
    }
    userStore.setUser(user)
  } catch (e) {
    console.error(e)
    ElMessage.error('加载用户信息失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadUser()
})

const handleSave = async () => {
  loading.value = true
  try {
    await updateMe({
      nickname: form.value.nickname,
      email: form.value.email,
      cellphone: form.value.cellphone,
      avatar: form.value.avatar
    })
    ElMessage.success('保存成功')
    // 更新store中的用户信息
    userStore.updateUserInfo(form.value)
  } catch (e) {
    console.error(e)
    ElMessage.error('保存失败')
  } finally {
    loading.value = false
  }
}

const handleChangePassword = async () => {
  if (!passwordForm.value.oldPassword) {
    ElMessage.error('请输入原密码')
    return
  }
  if (!passwordForm.value.newPassword) {
    ElMessage.error('请输入新密码')
    return
  }
  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    ElMessage.error('两次输入的密码不一致')
    return
  }
  
  loading.value = true
  try {
    await changePassword({
      old_password: passwordForm.value.oldPassword,
      new_password: passwordForm.value.newPassword
    })
    ElMessage.success('密码修改成功，请重新登录')
    userStore.logout()
    setTimeout(() => {
      router.replace('/login')
    }, 0)
  } catch (e) {
    console.error(e)
    ElMessage.error(e.response?.data?.message || '密码修改失败')
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.profile-page {
  .page-header {
    margin-bottom: 24px;
    
    .page-title {
      font-size: 24px;
      font-weight: 600;
      color: #1a1a2e;
    }
  }
  
  .profile-card,
  .password-card {
    border-radius: 12px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.05);
    
    .card-header {
      font-weight: 600;
      color: #1a1a2e;
    }
  }
}
</style>
