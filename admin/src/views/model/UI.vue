<template>
  <div class="model-ui-page">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>App 界面</span>
          <el-button size="small" type="primary" @click="goDesign">界面设计</el-button>
        </div>
      </template>

        <AppPreview
          :title="modelName"
          :controls="controls"
          :properties="properties"
        />

    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import AppPreview from '@/components/AppPreview.vue'
import { getModel } from '@/api/model'

const router = useRouter()

const props = defineProps({
  modelId: {
    type: String,
    required: true
  }
})

const loading = ref(false)
const controls = ref([])
const properties = ref([])
const services = ref([])
const modelName = ref('')

const loadModel = async () => {
  loading.value = true
  try {
    const res = await getModel(props.modelId)
    const model = res.data || {}
    controls.value = model.controls || []
    properties.value = model.properties || []
    services.value = model.services || []
    modelName.value = model.name || '设备控制'
  } catch (error) {
    console.error(error)
    ElMessage.error('加载物模型信息失败')
  } finally {
    loading.value = false
  }
}

const goDesign = () => {
  router.push(`/models/${props.modelId}/designer`)
}

onMounted(() => {
  loadModel()
})
</script>

<style lang="scss" scoped>
.model-ui-page {
  height: calc(100vh - 180px);

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .preview-mode, .design-mode {
    height: 100%;
  }
}
</style>
