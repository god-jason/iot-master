<template>
  <div>
    <div class="action-list">
      <el-card v-for="action in services" :key="action.key" class="action-card" shadow="hover">
        <div class="action-header">
          <span class="action-name">{{ action.name }}</span>
          <el-tag size="small" class="action-type">
            {{ getServiceTypeName(action.type) }}
          </el-tag>
        </div>
        <div class="action-desc">{{ action.key }} - {{ action.params?.length || 0 }} 个参数</div>
        
        <div v-if="action.type === 'switch' && action.params?.length" class="action-control">
          <el-switch 
            :checked="getParamValue(action.params[0])"
            @change="(val) => executeSwitchAction(action, action.params[0].key, val)"
            active-text="开启"
            inactive-text="关闭"
          />
        </div>
        
        <div v-else-if="action.type === 'slider' && action.params?.length" class="action-control">
          <el-slider 
            v-model="sliderValues[action.key]"
            :min="action.params[0].min || 0"
            :max="action.params[0].max || 100"
            :step="action.params[0].step || 1"
            :show-input="true"
            @change="(val) => executeSliderAction(action, action.params[0].key, val)"
          />
        </div>
        
        <div v-else class="action-footer">
          <el-button type="primary" size="small" @click="openActionDialog(action)">执行</el-button>
        </div>
      </el-card>
      <el-empty v-if="!services || services.length === 0" description="暂无远程操作" />
    </div>

    <div class="action-logs">
      <h4>操作日志</h4>
      <el-table :data="actionLogs" v-loading="actionLogsLoading">
        <el-table-column prop="time" label="时间" width="180">
          <template #default="{ row }">
            {{ $formatDateTime(row.time) }}
          </template>
        </el-table-column>
        <el-table-column prop="action" label="操作" />
        <el-table-column prop="params" label="参数">
          <template #default="{ row }">
            {{ row.params ? JSON.stringify(row.params) : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="result" label="结果">
          <template #default="{ row }">
            <el-tag :type="row.result === 'success' ? 'success' : 'danger'" size="small">
              {{ row.result === 'success' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="actionLogPage"
        v-model:page-size="actionLogPageSize"
        :total="actionLogTotal"
        layout="total, prev, pager, next"
        @current-change="loadActionLogs"
      />
    </div>
  </div>

  <el-dialog v-model="actionDialogVisible" :title="`执行 ${currentAction?.name}`" width="500px">
    <el-form ref="actionFormRef" :model="actionForm" label-width="120px">
      <el-form-item
        v-for="param in currentAction?.params"
        :key="param.key"
        :label="param.name"
        :prop="param.key"
      >
        <el-input-number controls-position="right"
          v-if="param.type === 'int' || param.type === 'float'"
          v-model="actionForm[param.key]"
          :min="param.min"
          :max="param.max"
          :step="param.step || 1"
        />
        <el-switch v-else-if="param.type === 'bool'" v-model="actionForm[param.key]" />
        <el-input v-else v-model="actionForm[param.key]" placeholder="请输入参数值" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="actionDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="executeAction" :loading="actionLoading">执行</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, defineProps } from 'vue'
import { ElMessage } from 'element-plus'
import { executeDeviceControl } from '@/api/device'

const props = defineProps({
  deviceId: {
    type: String,
    required: true
  },
  services: {
    type: Array,
    default: () => []
  },
  realtimeData: {
    type: Object,
    default: () => ({})
  }
})

const actionDialogVisible = ref(false)
const actionFormRef = ref(null)
const actionForm = ref({})
const currentAction = ref(null)
const actionLoading = ref(false)
const actionLogs = ref([])
const actionLogsLoading = ref(false)
const actionLogPage = ref(1)
const actionLogPageSize = ref(20)
const actionLogTotal = ref(0)
const sliderValues = ref({})

const getServiceTypeName = (type) => {
  const map = {
    switch: '开关',
    slider: '滑动条',
    button: '按钮',
    form: '表单'
  }
  return map[type] || type
}

const getParamValue = (param) => {
  if (param.register && props.realtimeData[param.key] !== undefined) {
    return props.realtimeData[param.key]
  }
  return param.default || false
}

const openActionDialog = (action) => {
  currentAction.value = action
  actionForm.value = {}
  action.params?.forEach(param => {
    actionForm.value[param.key] = getParamValue(param)
  })
  actionDialogVisible.value = true
}

const executeAction = async () => {
  actionLoading.value = true
  try {
    await executeDeviceControl(props.deviceId, currentAction.value.key, actionForm.value)
    ElMessage.success('操作执行成功')
    actionDialogVisible.value = false
    loadActionLogs()
  } catch (error) {
    console.error(error)
    ElMessage.error('操作执行失败')
  } finally {
    actionLoading.value = false
  }
}

const executeSwitchAction = async (action, key, value) => {
  try {
    await executeDeviceControl(props.deviceId, action.key, { [key]: value })
    ElMessage.success('操作执行成功')
  } catch (error) {
    console.error(error)
    ElMessage.error('操作执行失败')
  }
}

const executeSliderAction = async (action, key, value) => {
  try {
    await executeDeviceControl(props.deviceId, action.key, { [key]: value })
  } catch (error) {
    console.error(error)
    ElMessage.error('操作执行失败')
  }
}

const loadActionLogs = async () => {
  actionLogsLoading.value = true
  try {
    actionLogs.value = []
    actionLogTotal.value = 0
  } catch (error) {
    console.error(error)
  } finally {
    actionLogsLoading.value = false
  }
}
</script>

<style lang="scss" scoped>
.action-list {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 20px;

  .action-card {
    .action-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 8px;
      gap: 8px;

      .action-name {
        font-weight: 600;
        flex: 1;
      }

      .action-type {
        background: #ecf5ff;
        color: #409eff;
      }
    }

    .action-desc {
      color: #909399;
      font-size: 13px;
      margin-bottom: 12px;
    }

    .action-control {
      padding: 8px 0;
    }

    .action-footer {
      display: flex;
      justify-content: flex-end;
      margin-top: 8px;
    }
  }
}

.action-logs {
  h4 {
    margin-bottom: 12px;
  }
}

.el-pagination {
  margin-top: 16px;
  justify-content: flex-end;
}
</style>