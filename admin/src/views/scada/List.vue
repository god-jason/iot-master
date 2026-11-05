<template>
  <div class="scada-list">
    <div class="page-header">
      <h2 class="page-title">组态大屏</h2>
      <div class="header-actions">
        <el-button @click="loadData">
          <el-icon><Refresh /></el-icon>
        </el-button>
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>新建画面
        </el-button>
      </div>
    </div>

    <div class="search-bar">
      <el-radio-group v-model="scadaType" @change="handleTypeChange">
        <el-radio-button value="scada">组态画面</el-radio-button>
        <el-radio-button value="screen">大屏画面</el-radio-button>
      </el-radio-group>
      <el-input v-model="searchKeyword" placeholder="搜索画面名称" style="width: 260px; margin-left: 16px" clearable @keyup.enter="loadData">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-button type="primary" @click="loadData">搜索</el-button>
    </div>

    <div class="scada-grid" v-loading="loading">
      <el-card v-for="scada in tableData" :key="scada.id" class="scada-card" shadow="hover">
        <div class="scada-thumb" @click="handleEdit(scada)">
          <img v-if="scada.icon" :src="getImageUrl(scada.icon)" alt="缩略图" />
          <div v-else class="thumb-placeholder">
            <el-icon :size="40"><Monitor /></el-icon>
          </div>
        </div>
        <div class="scada-info">
          <div class="scada-name">{{ scada.name }}</div>
          <div class="scada-desc">{{ scada.description || '暂无描述' }}</div>
          <div class="scada-meta">
            <span>{{ formatTime(scada.updated || scada.created) }}</span>
          </div>
        </div>
        <div class="scada-actions">
          <el-button link type="primary" @click="handleEdit(scada)">编辑</el-button>
          <el-button link type="primary" @click="handlePreview(scada)">预览</el-button>
          <el-button link type="danger" @click="handleDelete(scada)">删除</el-button>
        </div>
      </el-card>
      <el-empty v-if="!loading && tableData.length === 0" description="暂无画面" />
    </div>

    <el-pagination
      v-if="total > 0"
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :total="total"
      :page-sizes="[12, 24, 48]"
      layout="total, sizes, prev, pager, next"
      @size-change="loadData"
      @current-change="loadData"
    />

    <!-- 新建画面弹窗 -->
    <el-dialog v-model="createVisible" title="新建画面" width="420px">
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="画面名称">
          <el-input v-model="createForm.name" placeholder="请输入画面名称" />
        </el-form-item>
        <el-form-item label="画面类型">
          <el-radio-group v-model="createForm.type">
            <el-radio value="scada">组态</el-radio>
            <el-radio value="screen">大屏</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="createForm.description" type="textarea" :rows="2" placeholder="画面描述（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus, Monitor } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import { searchScadas, createScada, deleteScada } from '@/api/scada'

const router = useRouter()

const baseUrl = import.meta.env.VITE_API_BASE_URL || ''

const getImageUrl = (url) => {
  if (!url) return ''
  if (url.startsWith('http') || url.startsWith('data:')) return url
  return baseUrl + url
}

const formatTime = (t) => t ? dayjs(t).format('YYYY-MM-DD HH:mm') : '-'

const loading = ref(false)
const tableData = ref([])
const currentPage = ref(1)
const pageSize = ref(12)
const total = ref(0)
const searchKeyword = ref('')
const scadaType = ref('scada')

const loadData = async () => {
  loading.value = true
  try {
    const res = await searchScadas({
      page: currentPage.value,
      page_size: pageSize.value,
      keyword: searchKeyword.value,
      filter: [{ field: 'type', operator: 'eq', value: scadaType.value }]
    })
    tableData.value = res.data || []
    total.value = res.total || 0
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const handleTypeChange = () => {
  currentPage.value = 1
  loadData()
}

// ========== 新建 ==========
const createVisible = ref(false)
const creating = ref(false)
const createForm = reactive({ name: '', type: 'scada', description: '' })

const handleCreate = () => {
  createForm.name = ''
  createForm.description = ''
  createForm.type = scadaType.value
  createVisible.value = true
}

const submitCreate = async () => {
  if (!createForm.name.trim()) {
    ElMessage.warning('请输入画面名称')
    return
  }
  creating.value = true
  try {
    // 大屏画面默认深色背景，组态画面默认白色背景
    const fill = createForm.type === 'screen' ? '#1a1a2e' : '#ffffff'
    const res = await createScada({
      name: createForm.name.trim(),
      description: createForm.description,
      type: createForm.type,
      canvas: { width: 1280, height: 720, fill },
      elements: []
    })
    createVisible.value = false
    // 创建成功后直接进入编辑器
    router.push({ path: `/scada/${res.data.id}/editor`, query: { type: createForm.type } })
  } catch (e) {
    console.error(e)
  } finally {
    creating.value = false
  }
}

// ========== 操作 ==========
const handleEdit = (scada) => {
  router.push({ path: `/scada/${scada.id}/editor`, query: { type: scada.type } })
}

const handlePreview = (scada) => {
  const routeData = router.resolve({ path: `/scada/${scada.id}/preview` })
  window.open(routeData.href, '_blank')
}

const handleDelete = (scada) => {
  ElMessageBox.confirm(`确定删除画面「${scada.name}」吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await deleteScada(scada.id)
    ElMessage.success('删除成功')
    loadData()
  }).catch(() => {})
}

onMounted(() => loadData())
</script>

<style lang="scss" scoped>
.scada-list {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;

    .page-title {
      font-size: 24px;
    }

    .header-actions {
      display: flex;
      gap: 10px;
    }
  }

  .search-bar {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 20px;
  }

  .scada-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 20px;
  }

  .scada-card {
    transition: transform 0.2s, box-shadow 0.2s;

    &:hover {
      transform: translateY(-4px);
    }

    .scada-thumb {
      height: 150px;
      border-radius: 8px;
      overflow: hidden;
      background: #f5f7fa;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;

      img {
        width: 100%;
        height: 100%;
        object-fit: cover;
      }

      .thumb-placeholder {
        color: #c0c4cc;
      }
    }

    .scada-info {
      padding: 12px 4px 8px;

      .scada-name {
        font-size: 15px;
        font-weight: 600;
        margin-bottom: 4px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .scada-desc {
        font-size: 12px;
        color: #909399;
        margin-bottom: 6px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .scada-meta {
        font-size: 12px;
        color: #c0c4cc;
      }
    }

    .scada-actions {
      display: flex;
      gap: 12px;
      padding: 0 4px 4px;
      border-top: 1px solid #f0f2f5;
      padding-top: 8px;
    }
  }

  .el-pagination {
    margin-top: 20px;
    justify-content: flex-end;
  }
}
</style>
