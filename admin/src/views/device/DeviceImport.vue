<template>
  <div class="device-import">
    <PageHeader title="设备导入">
      <template #extra>
        <div class="actions">
          <el-button @click="downloadTemplate">
            <el-icon><Download /></el-icon> 下载模板
          </el-button>
          <el-button type="primary" @click="handleImport">
            <el-icon><Upload /></el-icon> 执行导入
          </el-button>
        </div>
      </template>
    </PageHeader>

    <div class="upload-section">
      <el-upload
        ref="uploadRef"
        :auto-upload="false"
        :limit="1"
        accept=".xlsx,.xls"
        drag
        :on-change="handleFileChange"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
        <template #tip>
          <div class="el-upload__tip">只能上传xlsx/xls文件，请先下载模板填写数据</div>
        </template>
      </el-upload>
    </div>

    <el-table v-if="importData.length > 0" :data="importData" v-loading="loading" stripe>
      <el-table-column type="index" label="序号" width="60" />
      <el-table-column prop="name" label="设备名称" min-width="120">
        <template #default="{ row }">
          <el-input v-model="row.name" placeholder="设备名称" size="small" />
        </template>
      </el-table-column>
      <el-table-column prop="sn" label="设备SN" min-width="150">
        <template #default="{ row }">
          <el-input v-model="row.sn" placeholder="设备SN" size="small" />
        </template>
      </el-table-column>
      <el-table-column prop="model_name" label="所属物模型" min-width="120">
        <template #default="{ row }">
          <el-select v-model="row.model_name" placeholder="选择物模型" size="small" style="width: 100%">
            <el-option v-for="p in models" :key="p.id" :label="p.name" :value="p.name" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column prop="group_name" label="所属分组" min-width="120">
        <template #default="{ row }">
          <el-input v-model="row.group_name" placeholder="所属分组" size="small" />
        </template>
      </el-table-column>
      <el-table-column prop="status" label="导入状态" width="120">
        <template #default="{ row }">
          <el-tag v-if="row.status === 'success'" type="success" size="small">成功</el-tag>
          <el-tag v-else-if="row.status === 'failed'" type="danger" size="small">失败</el-tag>
          <el-tag v-else type="info" size="small">待导入</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="message" label="提示信息" min-width="200">
        <template #default="{ row }">
          <span v-if="row.message" class="error-msg">{{ row.message }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="80">
        <template #default="{ row, $index }">
          <el-button link type="danger" size="small" @click="removeRow($index)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Download, Upload, UploadFilled } from '@element-plus/icons-vue'
import * as XLSX from 'xlsx'
import PageHeader from '@/components/PageHeader.vue'
import { searchModels } from '@/api/model'
import { createDevice } from '@/api/device'

const uploadRef = ref(null)
const loading = ref(false)
const importData = ref([])
const models = ref([])

const loadModels = async () => {
  try {
    const res = await searchModels({ page: 1, page_size: 200 })
    models.value = res.data || []
  } catch (e) {
    console.error(e)
  }
}

const downloadTemplate = () => {
  const templateData = [
    { '设备名称': '示例设备1', '设备SN': 'SN001', '所属物模型': '', '所属分组': '测试分组' },
    { '设备名称': '示例设备2', '设备SN': 'SN002', '所属物模型': '', '所属分组': '测试分组' }
  ]

  const ws = XLSX.utils.json_to_sheet(templateData)
  const wb = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(wb, ws, '设备导入模板')
  XLSX.writeFile(wb, '设备导入模板.xlsx')
}

const handleFileChange = (file) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const data = new Uint8Array(e.target.result)
      const workbook = XLSX.read(data, { type: 'array' })
      const sheetName = workbook.SheetNames[0]
      const worksheet = workbook.Sheets[sheetName]
      const jsonData = XLSX.utils.sheet_to_json(worksheet)

      importData.value = jsonData.map(item => ({
        name: item['设备名称'] || '',
        sn: item['设备SN'] || '',
        model_name: item['所属物模型'] || '',
        group_name: item['所属分组'] || '',
        status: 'pending',
        message: ''
      }))

      ElMessage.success(`成功读取 ${importData.value.length} 条数据`)
    } catch (error) {
      console.error(error)
      ElMessage.error('文件解析失败，请确保是有效的xlsx文件')
    }
  }
  reader.readAsArrayBuffer(file.raw)
}

const removeRow = (index) => {
  importData.value.splice(index, 1)
}

const getModelId = (modelName) => {
  const model = models.value.find(p => p.name === modelName)
  return model ? model.id : null
}

const handleImport = async () => {
  if (importData.value.length === 0) {
    ElMessage.warning('请先上传文件')
    return
  }

  try {
    await ElMessageBox.confirm(`确定要导入 ${importData.value.length} 条设备数据吗？`, '确认', {
      type: 'warning'
    })
  } catch (e) {
    return
  }

  loading.value = true

  for (let i = 0; i < importData.value.length; i++) {
    const row = importData.value[i]

    if (!row.name || !row.sn) {
      row.status = 'failed'
      row.message = '设备名称和SN不能为空'
      continue
    }

    const modelId = getModelId(row.model_name)

    try {
      await createDevice({
        name: row.name,
        sn: row.sn,
        model_id: modelId,
        group_id: null
      })
      row.status = 'success'
      row.message = ''
    } catch (error) {
      console.error(error)
      row.status = 'failed'
      row.message = error.response?.data?.msg || '导入失败'
    }
  }

  loading.value = false

  const successCount = importData.value.filter(r => r.status === 'success').length
  const failedCount = importData.value.filter(r => r.status === 'failed').length

  if (successCount === importData.value.length) {
    ElMessage.success('全部导入成功')
  } else if (failedCount === importData.value.length) {
    ElMessage.error('全部导入失败')
  } else {
    ElMessage.info(`导入完成：成功 ${successCount} 条，失败 ${failedCount} 条`)
  }
}

onMounted(() => {
  loadModels()
})
</script>

<style lang="scss" scoped>
.device-import {
  .upload-section {
    margin: 20px 0;
    background: #fff;
    border-radius: 8px;
    padding: 40px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
  }

  .error-msg {
    color: #f56c6c;
    font-size: 12px;
  }
}
</style>
