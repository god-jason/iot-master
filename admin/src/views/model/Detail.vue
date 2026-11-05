<template>
  <div class="model-detail-page">
    <PageHeader title="物模型详情" :show-back="true">
      <template #extra>
        <el-button type="primary" @click="openEditDrawer">
          <el-icon><Edit /></el-icon> 编辑
        </el-button>
        <el-button type="danger" @click="handleDelete">
          <el-icon><Delete /></el-icon> 删除
        </el-button>
      </template>
    </PageHeader>

    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span>基本信息</span>
        </div>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="物模型ID">{{ model.id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="物模型秘钥">{{ model.secret || '-' }}</el-descriptions-item>
        <el-descriptions-item label="物模型名称">{{ model.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="设备类型">{{ typeMap[model.type] || '-' }}</el-descriptions-item>
        <el-descriptions-item label="协议">
          <el-select v-if="editing" v-model="model.protocol" placeholder="请选择协议" style="width: 100%">
            <el-option v-for="p in protocols" :key="p.id" :label="p.name" :value="p.name" />
          </el-select>
          <span v-else>{{ model.protocol || '-' }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间" :span="2">{{ $formatDateTime(model.created) }}</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="设备" name="devices" lazy>
        <DeviceList :model-id="id" :show-group="false" />
      </el-tab-pane>

      <el-tab-pane label="物模型" name="model" lazy>
        <ModelDefinition :model-id="id" />
      </el-tab-pane>

      <el-tab-pane label="界面设计" name="ui" lazy>
        <ModelUI :model-id="id" />
      </el-tab-pane>
    </el-tabs>

    <ModelForm
      v-model:visible="editDrawerVisible"
      :model-value="model"
      @submit="handleEditSubmit"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Edit, Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import PageHeader from '@/components/PageHeader.vue'
import DeviceList from '@/views/device/List.vue'
import ModelDefinition from './Definition.vue'
import ModelUI from './UI.vue'
import ModelForm from './ModelForm.vue'
import { getModel, updateModel, deleteModel } from '@/api/model'

const route = useRoute()
const router = useRouter()
const id = ref(route.params.id)

const typeMap = {
  gateway: '网关',
  direct: '直连设备',
  sub: '子设备'
}

const activeTab = ref('devices')
const model = ref({})

const loadModel = async () => {
  try {
    const res = await getModel(id.value)
    model.value = res.data || {}
  } catch (e) {
    console.error(e)
    ElMessage.error('加载物模型信息失败')
  }
}

const editDrawerVisible = ref(false)

const openEditDrawer = () => {
  editDrawerVisible.value = true
}

const handleEditSubmit = async (formData) => {
  await updateModel(id.value, formData)
  ElMessage.success('更新成功')
  editDrawerVisible.value = false
  loadModel()
}

const handleDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定删除物模型"${model.value.name}"吗？`, '提示', {
      type: 'warning',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消'
    })
    await deleteModel(id.value)
    ElMessage.success('删除成功')
    router.push('/models')
  } catch (e) {
    if (e !== 'cancel') {
      console.error(e)
      ElMessage.error('删除失败')
    }
  }
}

onMounted(() => {
  loadModel()
})
</script>

<style lang="scss" scoped>
.model-detail-page {
  .info-card {
    margin: 20px 0;
    border-radius: 8px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);

    .card-header {
      font-weight: 600;
      color: #303133;
    }
  }

  .detail-tabs {
    background: #fff;
    border-radius: 8px;
    padding: 20px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
  }
}
</style>
