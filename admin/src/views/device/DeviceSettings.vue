<template>
  <div>
    <div class="setting-list">
      <el-card v-for="setting in settings" :key="setting.key" class="setting-card" shadow="hover">
        <div class="setting-header">
          <span class="setting-name">{{ setting.name }}</span>
          <el-button type="primary" size="small" @click="openSettingDialog(setting)">配置</el-button>
        </div>
        <div class="setting-desc">{{ setting.key }} - {{ setting.params?.length || 0 }} 个参数</div>
      </el-card>
      <el-empty v-if="!settings || settings.length === 0" description="暂无远程配置" />
    </div>
  </div>

  <el-dialog v-model="settingDialogVisible" :title="`配置 ${currentSetting?.name}`" width="500px">
    <el-form ref="settingFormRef" :model="settingForm" label-width="120px">
      <el-form-item
        v-for="param in currentSetting?.params"
        :key="param.key"
        :label="param.name"
        :prop="param.key"
      >
        <el-input-number controls-position="right"
          v-if="param.type === 'int' || param.type === 'float'"
          v-model="settingForm[param.key]"
          :min="param.min"
          :max="param.max"
          :step="param.step || 1"
        />
        <el-switch v-else-if="param.type === 'bool'" v-model="settingForm[param.key]" />
        <el-input v-else v-model="settingForm[param.key]" placeholder="请输入参数值" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="settingDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="applySetting" :loading="settingLoading">应用配置</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { setDeviceConfig } from '@/api/device'

const props = defineProps({
  deviceId: {
    type: String,
    required: true
  },
  settings: {
    type: Array,
    default: () => []
  }
})

const settingDialogVisible = ref(false)
const settingFormRef = ref(null)
const settingForm = ref({})
const currentSetting = ref(null)
const settingLoading = ref(false)

const openSettingDialog = (setting) => {
  currentSetting.value = setting
  settingForm.value = {}
  setting.params?.forEach(param => {
    settingForm.value[param.key] = param.default
  })
  settingDialogVisible.value = true
}

const applySetting = async () => {
  settingLoading.value = true
  try {
    const config = {}
    config[currentSetting.value.key] = settingForm.value
    await setDeviceConfig(props.deviceId, config)
    ElMessage.success('配置应用成功')
    settingDialogVisible.value = false
  } catch (error) {
    console.error(error)
    ElMessage.error('配置应用失败')
  } finally {
    settingLoading.value = false
  }
}
</script>

<style lang="scss" scoped>
.setting-list {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;

  .setting-card {
    .setting-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 8px;
      gap: 8px;

      .setting-name {
        font-weight: 600;
        flex: 1;
      }
    }

    .setting-desc {
      color: #909399;
      font-size: 13px;
    }
  }
}
</style>