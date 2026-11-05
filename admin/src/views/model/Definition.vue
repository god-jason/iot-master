<template>
  <div class="model-definition">
    <div class="definition-header">
      <span>定义模型的属性、事件和服务</span>
      <div class="header-actions">
        <el-button v-if="!editing" type="primary" @click="startEdit">
          <el-icon><Edit /></el-icon>编辑物模型
        </el-button>
        <template v-else>
          <el-button @click="cancelEdit">取消</el-button>
          <el-button type="primary" :loading="saving" @click="saveModel">
            <el-icon><Check /></el-icon>保存
          </el-button>
        </template>
      </div>
    </div>
    <ModelPreview v-if="!editing" :model="model" :current-protocol="currentProtocol" />
    <ModelEditor v-else v-model="editData" :protocols="protocols" :current-protocol="currentProtocol" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { Edit, Check } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import ModelPreview from '@/components/ModelPreview.vue'
import ModelEditor from '@/components/ModelEditor.vue'
import { getModel, updateModel } from '@/api/model'
import { searchProtocols } from '@/api/protocol'

const props = defineProps({
  modelId: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['saved'])

const model = ref({})
const protocols = ref([])
const editData = reactive({
  properties: [],
  pollers: [],
  expressions: [],
  filters: [],
  events: [],
  services: [],
  settings: []
})
const editing = ref(false)
const saving = ref(false)

const currentProtocol = computed(() => {
  if (!model.value.protocol || !protocols.value.length) {
    return null
  }
  return protocols.value.find(p => p.name === model.value.protocol)
})

const loadModel = async () => {
  try {
    const res = await getModel(props.modelId)
    model.value = res.data || {}
    editData.properties = res.data?.properties || []
    editData.pollers = res.data?.pollers || []
    editData.expressions = res.data?.expressions || []
    editData.filters = res.data?.filters || []
    editData.events = res.data?.events || []
    editData.services = res.data?.services || []
    editData.settings = res.data?.settings || []
  } catch (e) {
    console.error(e)
    ElMessage.error('加载物模型信息失败')
  }
}

const loadProtocols = async () => {
  try {
    const res = await searchProtocols({ page: 1, page_size: 100 })
    protocols.value = res.data || []
  } catch (e) {
    console.error(e)
    ElMessage.error('加载协议列表失败')
  }
}

const startEdit = () => {
  editing.value = true
}

const cancelEdit = () => {
  loadModel()
  editing.value = false
}

const saveModel = async () => {
  saving.value = true
  try {
    await updateModel(props.modelId, {
      properties: editData.properties,
      pollers: editData.pollers,
      expressions: editData.expressions,
      filters: editData.filters,
      events: editData.events,
      services: editData.services,
      settings: editData.settings
    })
    ElMessage.success('保存成功')
    editing.value = false
    loadModel()
    emit('saved')
  } catch (e) {
    console.error(e)
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

watch(() => props.modelId, () => {
  loadModel()
})

onMounted(() => {
  loadModel()
  loadProtocols()
})
</script>

<style lang="scss" scoped>
.model-definition {
  .definition-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    color: #909399;
    font-size: 14px;

    .header-actions {
      display: flex;
      gap: 8px;
    }
  }
}
</style>
