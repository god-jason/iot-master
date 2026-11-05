<template>
  <div class="designer-page">
    <AppDesigner
      v-model="controls"
      :properties="properties"
      :services="services"
      @save="handleSave"
    />

  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import AppDesigner from '@/components/AppDesigner.vue'
import { getModel, updateModel } from '@/api/model'

const router = useRouter()
const route = useRoute()

const controls = ref([])
const properties = ref([])
const services = ref([])

const loadModel = async () => {
  const modelId = route.params.id
  try {
    const res = await getModel(modelId)
    const model = res.data || {}
    controls.value = JSON.parse(JSON.stringify(model.controls || []))
    properties.value = JSON.parse(JSON.stringify(model.properties || []))
    services.value = JSON.parse(JSON.stringify(model.services || []))
  } catch (error) {
    console.error(error)
    ElMessage.error('加载物模型信息失败')
  }
}

const handleSave = async (newControls) => {
  const modelId = route.params.id
  try {
    await updateModel(modelId, { controls: newControls })
    ElMessage.success('保存成功')
    router.push(`/models/${modelId}`)
  } catch (error) {
    console.error(error)
    ElMessage.error('保存失败')
  }
}

onMounted(() => {
  loadModel()
})

watch(() => route.params.id, () => {
  loadModel()
})
</script>

<style lang="scss" scoped>
.designer-page {
  height: calc(100vh - 60px);
}
</style>
