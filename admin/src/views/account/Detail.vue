<template>
  <div class="account-detail">
    <PageHeader :title="account.username || '账号详情'" show-back>
      <template #extra>
        <el-button type="primary" @click="openEditDialog">编辑</el-button>
      </template>
    </PageHeader>

    <el-card class="info-card" v-loading="loading">
      <div class="account-info">
        <div class="avatar-section">
          <el-avatar :size="120" :src="getAvatarUrl(account.avatar)">
            {{ account.username?.[0]?.toUpperCase() || 'U' }}
          </el-avatar>
        </div>
        
        <el-descriptions :column="2" border>
          <el-descriptions-item label="用户名">{{ account.username || '-' }}</el-descriptions-item>
          <el-descriptions-item label="昵称">{{ account.nickname || '-' }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ account.email || '-' }}</el-descriptions-item>
          <el-descriptions-item label="手机号">{{ account.cellphone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="账号类型">
            <el-tag :type="account.admin || account.platform ? 'success' : 'info'">
              {{ account.admin || account.platform ? '平台账号' : '用户账号' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="account.disabled ? 'danger' : 'success'">
              {{ account.disabled ? '禁用' : '正常' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ $formatDateTime(account.created) }}</el-descriptions-item>
        </el-descriptions>
      </div>
    </el-card>

    <AccountForm
      v-model:visible="dialogVisible"
      :model-value="account"
      @submit="handleSubmit"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import PageHeader from '@/components/PageHeader.vue'
import AccountForm from './AccountForm.vue'
import { getAccount, updateAccount } from '@/api/account'

const route = useRoute()
const accountId = route.params.id

const baseUrl = import.meta.env.VITE_API_BASE_URL || ''

const getAvatarUrl = (url) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return baseUrl + url
}

const loading = ref(false)
const account = ref({})
const dialogVisible = ref(false)

const loadAccount = async () => {
  loading.value = true
  try {
    const res = await getAccount(accountId)
    account.value = res.data || {}
  } catch (error) {
    console.error(error)
    ElMessage.error('加载账号信息失败')
  } finally {
    loading.value = false
  }
}

const openEditDialog = () => {
  dialogVisible.value = true
}

const handleSubmit = async (formData) => {
  await updateAccount(accountId, formData)
  ElMessage.success('更新成功')
  dialogVisible.value = false
  loadAccount()
}

onMounted(() => {
  loadAccount()
})
</script>

<style lang="scss" scoped>
.account-detail {
  .info-card {
    border-radius: 12px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.05);

    .account-info {
      .avatar-section {
        display: flex;
        justify-content: center;
        margin-bottom: 24px;
      }
    }
  }
}
</style>