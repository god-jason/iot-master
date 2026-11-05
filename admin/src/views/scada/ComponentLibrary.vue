<template>
  <div class="component-library">
    <div class="page-header">
      <h2 class="page-title">组件库</h2>
      <div class="header-actions">
        <el-button @click="loadData"><el-icon><Refresh /></el-icon></el-button>
        <el-button type="primary" @click="handleCreate"><el-icon><Plus /></el-icon>新建组件</el-button>
      </div>
    </div>

    <div class="search-bar">
      <el-select v-model="categoryFilter" placeholder="全部分组" clearable style="width: 140px" @change="loadData">
        <el-option v-for="(name, key) in categoryOptions" :key="key" :label="name" :value="key" />
      </el-select>
      <el-input v-model="searchKeyword" placeholder="搜索组件名称" style="width: 260px" clearable @keyup.enter="loadData">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-button type="primary" @click="loadData">搜索</el-button>
    </div>

    <!-- 按分组折叠展示 -->
    <el-collapse v-loading="loading" v-model="activeGroups">
      <el-collapse-item
        v-for="(items, key) in groupedData"
        :key="key"
        :title="`${categoryOptions[key] || '自定义分组'}（${items.length}）`"
        :name="key"
      >
        <div class="component-grid">
          <el-card v-for="comp in items" :key="comp.id" class="component-card" shadow="hover">
            <div class="component-thumb" @click="handleEdit(comp)">
              <img v-if="comp.icon" :src="getImageUrl(comp.icon)" alt="缩略图" />
              <div v-else class="thumb-placeholder">
                <el-icon :size="40"><FolderOpened /></el-icon>
              </div>
            </div>
            <div class="component-info">
              <div class="component-name">{{ comp.name }}</div>
              <div class="component-desc">{{ comp.description || '暂无描述' }}</div>
              <div class="component-meta">
                <el-tag size="small" type="info">{{ (comp.elements || []).length }} 图元</el-tag>
                <el-tag size="small" type="info">{{ (comp.variables || []).length }} 变量</el-tag>
                <el-tag size="small" type="info">{{ (comp.events || []).length }} 事件</el-tag>
                <el-tag size="small" type="info">{{ (comp.scripts || []).length }} 脚本</el-tag>
              </div>
            </div>
            <div class="component-actions">
              <el-button link type="primary" @click="handleEdit(comp)">编辑</el-button>
              <el-button link type="danger" @click="handleDelete(comp)">删除</el-button>
            </div>
          </el-card>
        </div>
      </el-collapse-item>
    </el-collapse>

    <el-empty v-if="!loading && tableData.length === 0" description="暂无组件，可在组态编辑器中选中图元保存为组件" />

    <!-- 新建组件弹窗 -->
    <el-dialog v-model="createVisible" title="新建组件" width="420px">
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="组件名称">
          <el-input v-model="createForm.name" placeholder="请输入组件名称" />
        </el-form-item>
        <el-form-item label="分组">
          <el-select v-model="createForm.category" placeholder="选择分组" style="width: 100%">
            <el-option v-for="(name, key) in categoryOptions" :key="key" :label="name" :value="key" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="createForm.description" type="textarea" :rows="2" placeholder="组件描述（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 编辑组件弹窗 -->
    <el-dialog v-model="editVisible" title="编辑组件" width="720px" top="6vh">
      <el-tabs v-model="editTab">
        <el-tab-pane label="基本信息" name="base">
          <el-form :model="editForm" label-width="80px">
            <el-form-item label="组件名称">
              <el-input v-model="editForm.name" placeholder="请输入组件名称" />
            </el-form-item>
            <el-form-item label="分组">
              <el-select v-model="editForm.category" placeholder="选择分组" style="width: 100%">
                <el-option v-for="(name, key) in categoryOptions" :key="key" :label="name" :value="key" />
              </el-select>
            </el-form-item>
            <el-form-item label="描述">
              <el-input v-model="editForm.description" type="textarea" :rows="2" placeholder="组件描述" />
            </el-form-item>
            <el-form-item label="图元数量">
              <el-tag>{{ (editForm.elements || []).length }} 个（在组态编辑器中维护图形）</el-tag>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="变量" name="variables">
          <div class="config-toolbar"><el-button size="small" type="primary" @click="addVariable">添加变量</el-button></div>
          <el-table :data="editForm.variables" size="small" empty-text="暂无变量">
            <el-table-column label="标识" width="140"><template #default="{ row }"><el-input v-model="row.key" size="small" placeholder="key" /></template></el-table-column>
            <el-table-column label="名称" width="140"><template #default="{ row }"><el-input v-model="row.name" size="small" placeholder="名称" /></template></el-table-column>
            <el-table-column label="类型" width="110">
              <template #default="{ row }">
                <el-select v-model="row.type" size="small">
                  <el-option label="整数" value="int" />
                  <el-option label="浮点" value="float" />
                  <el-option label="字符串" value="string" />
                  <el-option label="布尔" value="bool" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="默认值"><template #default="{ row }"><el-input v-model="row.default" size="small" placeholder="默认值" /></template></el-table-column>
            <el-table-column label="说明"><template #default="{ row }"><el-input v-model="row.description" size="small" placeholder="说明" /></template></el-table-column>
            <el-table-column width="50"><template #default="{ $index }"><el-icon class="row-del" @click="editForm.variables.splice($index, 1)"><Close /></el-icon></template></el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="事件" name="events">
          <div class="config-toolbar"><el-button size="small" type="primary" @click="addEvent">添加事件</el-button></div>
          <el-table :data="editForm.events" size="small" empty-text="暂无事件">
            <el-table-column label="标识" width="160"><template #default="{ row }"><el-input v-model="row.key" size="small" placeholder="key" /></template></el-table-column>
            <el-table-column label="名称" width="160"><template #default="{ row }"><el-input v-model="row.name" size="small" placeholder="名称" /></template></el-table-column>
            <el-table-column label="说明"><template #default="{ row }"><el-input v-model="row.description" size="small" placeholder="说明" /></template></el-table-column>
            <el-table-column width="50"><template #default="{ $index }"><el-icon class="row-del" @click="editForm.events.splice($index, 1)"><Close /></el-icon></template></el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="脚本" name="scripts">
          <div class="config-toolbar"><el-button size="small" type="primary" @click="addScript">添加脚本</el-button></div>
          <div v-for="(script, idx) in editForm.scripts" :key="idx" class="script-item">
            <div class="script-header">
              <el-input v-model="script.name" size="small" placeholder="脚本名称" style="width: 200px" />
              <el-icon class="row-del" @click="editForm.scripts.splice(idx, 1)"><Close /></el-icon>
            </div>
            <el-input v-model="script.code" type="textarea" :rows="6" placeholder="// JavaScript 脚本，预览运行时执行" class="script-code" />
          </div>
          <el-empty v-if="editForm.scripts.length === 0" description="暂无脚本" :image-size="60" />
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus, FolderOpened, Close } from '@element-plus/icons-vue'
import { searchComponents, createComponent, updateComponent, deleteComponent } from '@/api/component'

const baseUrl = import.meta.env.VITE_API_BASE_URL || ''
const getImageUrl = (url) => {
  if (!url) return ''
  if (url.startsWith('http') || url.startsWith('data:')) return url
  return baseUrl + url
}

// 分组选项：固定两个 + 动态自定义分组都显示
const categoryOptions = {
  basic: '基础图元',
  control: '页面控件'
}

const loading = ref(false)
const tableData = ref([])
const searchKeyword = ref('')
const categoryFilter = ref('')
const activeGroups = ref(['basic', 'control'])  // 默认展开前两个

const loadData = async () => {
  loading.value = true
  try {
    const params = { keyword: searchKeyword.value }
    if (categoryFilter.value) params.category = categoryFilter.value
    const res = await searchComponents(params)
    tableData.value = res.data || []
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

// 按 category 分组
const groupedData = computed(() => {
  const groups = {}
  tableData.value.forEach(comp => {
    const key = comp.category || 'custom'
    if (!groups[key]) groups[key] = []
    groups[key].push(comp)
  })
  // 把 custom 显示为「自定义分组」
  if (groups.custom) {
    groups['自定义分组'] = groups.custom
    delete groups.custom
  }
  // 固定顺序：basic → control → 自定义 → 其他
  const ordered = {}
  ;['basic', 'control', '自定义分组'].forEach(k => { if (groups[k]) ordered[k] = groups[k] })
  Object.keys(groups).filter(k => !['basic', 'control', '自定义分组'].includes(k)).sort().forEach(k => { ordered[k] = groups[k] })
  return ordered
})

// ========== 新建 ==========
const createVisible = ref(false)
const creating = ref(false)
const createForm = reactive({ name: '', description: '', category: 'custom' })

const handleCreate = () => {
  createForm.name = ''
  createForm.description = ''
  createForm.category = 'custom'
  createVisible.value = true
}

const submitCreate = async () => {
  if (!createForm.name.trim()) { ElMessage.warning('请输入组件名称'); return }
  creating.value = true
  try {
    await createComponent({
      name: createForm.name.trim(),
      description: createForm.description,
      category: createForm.category,
      elements: []
    })
    createVisible.value = false
    ElMessage.success('创建成功')
    loadData()
  } catch (e) { console.error(e) }
  finally { creating.value = false }
}

// ========== 编辑 ==========
const editVisible = ref(false)
const editTab = ref('base')
const saving = ref(false)
const editId = ref('')
const editForm = reactive({
  name: '', description: '', category: 'custom',
  elements: [], variables: [], events: [], scripts: []
})

const handleEdit = (comp) => {
  editId.value = comp.id
  editForm.name = comp.name || ''
  editForm.description = comp.description || ''
  editForm.category = comp.category || 'custom'
  editForm.elements = comp.elements || []
  editForm.variables = (comp.variables || []).map(v => ({ ...v }))
  editForm.events = (comp.events || []).map(e => ({ ...e }))
  editForm.scripts = (comp.scripts || []).map(s => ({ ...s }))
  editTab.value = 'base'
  editVisible.value = true
}

const addVariable = () => { editForm.variables.push({ key: '', name: '', type: 'float', default: '', description: '' }) }
const addEvent = () => { editForm.events.push({ key: '', name: '', description: '' }) }
const addScript = () => { editForm.scripts.push({ name: '', code: '' }) }

const submitEdit = async () => {
  if (!editForm.name.trim()) { ElMessage.warning('请输入组件名称'); return }
  saving.value = true
  try {
    await updateComponent(editId.value, {
      name: editForm.name.trim(),
      description: editForm.description,
      category: editForm.category,
      variables: editForm.variables.filter(v => v.key),
      events: editForm.events.filter(e => e.key),
      scripts: editForm.scripts.filter(s => s.name),
      updated: new Date().toISOString()
    })
    editVisible.value = false
    ElMessage.success('保存成功')
    loadData()
  } catch (e) { console.error(e) }
  finally { saving.value = false }
}

// ========== 删除 ==========
const handleDelete = (comp) => {
  ElMessageBox.confirm(`确定删除组件「${comp.name}」吗？`, '提示', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
    .then(async () => {
      await deleteComponent(comp.id)
      ElMessage.success('删除成功')
      loadData()
    })
    .catch(() => {})
}

onMounted(() => loadData())
</script>

<style lang="scss" scoped>
.component-library {
  .page-header {
    display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;
    .page-title { font-size: 24px; }
    .header-actions { display: flex; gap: 10px; }
  }
  .search-bar { display: flex; align-items: center; gap: 10px; margin-bottom: 20px; }
  .component-grid {
    display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 16px 20px;
  }
  .component-card {
    transition: transform 0.2s, box-shadow 0.2s;
    &:hover { transform: translateY(-4px); }
    .component-thumb {
      height: 130px; border-radius: 8px; overflow: hidden; background: #f5f7fa; cursor: pointer;
      display: flex; align-items: center; justify-content: center;
      img { width: 100%; height: 100%; object-fit: cover; }
      .thumb-placeholder { color: #c0c4cc; }
    }
    .component-info {
      padding: 12px 4px 8px;
      .component-name { font-size: 15px; font-weight: 600; margin-bottom: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
      .component-desc { font-size: 12px; color: #909399; margin-bottom: 6px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
      .component-meta { display: flex; gap: 6px; flex-wrap: wrap; }
    }
    .component-actions { display: flex; gap: 12px; padding: 8px 4px 4px; border-top: 1px solid #f0f2f5; }
  }
  .config-toolbar { margin-bottom: 10px; }
  .row-del { color: #f56c6c; cursor: pointer; }
  .script-item {
    margin-bottom: 14px; border: 1px solid #e6e8eb; border-radius: 6px; padding: 10px;
    .script-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
    .script-code :deep(textarea) { font-family: Consolas, Monaco, monospace; font-size: 12px; }
  }
}
</style>
