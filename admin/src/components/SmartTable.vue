<template>
  <el-card class="smart-table-card" shadow="never">
    <template v-if="cardTitle || $slots.header" #header>
      <div class="smart-table-header">
        <span class="smart-table-title">
          <slot name="header">{{ cardTitle }}</slot>
        </span>
        <slot name="header-extra" />
      </div>
    </template>

    <div class="smart-table-container">
      <!-- 工具栏 -->
      <div v-if="showToolbar" class="smart-table-toolbar">
        <div class="toolbar-left">
          <el-button v-if="showRefresh" :icon="Refresh" circle @click="handleRefresh" />
          <div v-if="showKeywordSearch" class="toolbar-search">
            <el-input
              v-model="keyword"
              :placeholder="keywordPlaceholder"
              clearable
              style="width: 220px"
              @keyup.enter="handleKeywordSearch"
              @clear="handleKeywordSearch"
            >
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
            <el-button :icon="Search" @click="handleKeywordSearch" />
          </div>
          <slot name="toolbar-left" :refresh="handleRefresh" :params="queryParams" />
        </div>
        <div class="toolbar-right">
          <slot name="toolbar-right" :refresh="handleRefresh" :selected="selectedRows" />
          <el-button
            v-if="showBatchDelete"
            type="danger"
            :icon="Delete"
            :disabled="!selectedRows.length"
            @click="handleBatchDelete"
          >
            批量删除<span v-if="selectedRows.length"> ({{ selectedRows.length }})</span>
          </el-button>
          <el-button v-if="showCreate" type="primary" :icon="Plus" @click="handleCreate">
            {{ createText }}
          </el-button>
        </div>
      </div>

      <!-- 选中提示条 -->
      <div v-if="showSelection && selectedRows.length" class="table-toolbar">
        <span class="selection-info">已选择 <b>{{ selectedRows.length }}</b> 项</span>
        <el-button link type="primary" @click="clearSelection">清空选择</el-button>
      </div>

      <el-table
        ref="tableRef"
        v-loading="loading"
        :data="tableData"
        :row-key="rowKey"
        :stripe="stripe"
        :max-height="maxHeight"
        class="smart-table"
        @selection-change="handleSelectionChange"
        @row-click="handleRowClick"
        @sort-change="handleSortChange"
      >
        <el-table-column v-if="showSelection" type="selection" width="44" />
        <el-table-column
          v-if="showIndex"
          type="index"
          label="序号"
          width="60"
          :index="indexMethod"
        />

        <template v-for="(col, idx) in columns" :key="col.prop || col.label + '_' + idx">
          <el-table-column
            v-if="!col.hidden"
            :prop="col.prop"
            :label="col.label"
            :width="col.width"
            :min-width="col.minWidth"
            :align="col.align || 'left'"
            :fixed="col.fixed"
            :sortable="col.sortable ? 'custom' : false"
            :show-overflow-tooltip="col.showOverflowTooltip !== false"
          >
            <template #default="{ row, column, index }">
              <slot
                v-if="col.slot"
                :name="col.slot"
                :row="row"
                :column="column"
                :index="index"
              />
              <el-tag v-else-if="col.type === 'tag'" :type="tagTypeOf(row, col)" size="small">
                {{ tagTextOf(row, col) }}
              </el-tag>
              <el-tag
                v-else-if="col.type === 'boolean'"
                :type="cellValue(row, col) ? 'success' : 'info'"
                size="small"
              >
                {{ cellValue(row, col) ? '是' : '否' }}
              </el-tag>
              <span v-else-if="col.type === 'datetime'">{{ formatDateTime(cellValue(row, col)) }}</span>
              <span v-else-if="col.type === 'date'">{{ (cellValue(row, col) || '').substring(0, 10) || '-' }}</span>
              <span v-else-if="col.type === 'money'">{{ formatMoney(cellValue(row, col)) }}</span>
              <template v-else-if="col.type === 'image'">
                <el-image
                  v-if="cellValue(row, col)"
                  :src="imageUrl(cellValue(row, col), col)"
                  :preview-src-list="[imageUrl(cellValue(row, col), col)]"
                  :preview-teleported="true"
                  fit="cover"
                  :style="`width:${col.imageSize || 40}px;height:${col.imageSize || 40}px;border-radius:4px`"
                  @click.stop
                />
                <span v-else class="muted">-</span>
              </template>
              <span v-else class="cell-text">{{ cellValue(row, col) ?? '-' }}</span>
            </template>
          </el-table-column>
        </template>

        <el-table-column
          v-if="hasActions"
          label="操作"
          :width="operatorWidth"
          :fixed="operatorFixed"
          class-name="operator-column"
        >
          <template #default="{ row, index }">
            <slot v-if="$slots.actions" name="actions" :row="row" :index="index" :refresh="handleRefresh" />
            <template v-else>
              <el-button
                v-for="op in visibleOperators(row)"
                :key="op.key"
                link
                :type="typeof op.type === 'function' ? op.type(row) : (op.type || 'primary')"
                @click.stop="handleOperator(op, row, index)"
              >
                {{ typeof op.label === 'function' ? op.label(row) : op.label }}
              </el-button>
            </template>
          </template>
        </el-table-column>

        <template #empty>
          <slot name="empty">
            <el-empty description="暂无数据" :image-size="80" />
          </slot>
        </template>
      </el-table>

      <div v-if="showPagination" class="smart-table-pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSizeInner"
          :total="total"
          :page-sizes="pageSizes"
          :layout="paginationLayout"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </div>
  </el-card>
</template>

<script setup>
import { ref, computed, watch, onMounted, useSlots } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search, Plus, Delete } from '@element-plus/icons-vue'

const props = defineProps({
  // 列表加载函数，入参 { page, page_size, keyword, filter }，返回 { data, total }
  loader: { type: Function, required: true },
  // 列配置
  columns: { type: Array, default: () => [] },
  // 行操作按钮 [{ key, label, type, handler(row, index), show(row) }]
  operators: { type: Array, default: () => [] },
  rowKey: { type: String, default: 'id' },
  // 额外查询参数，变化后自动刷新
  params: { type: Object, default: () => ({}) },
  // 额外过滤条件数组 [{ field, operator, value }]
  filter: { type: Array, default: () => [] },
  cardTitle: { type: String, default: '' },
  showToolbar: { type: Boolean, default: true },
  showRefresh: { type: Boolean, default: true },
  showKeywordSearch: { type: Boolean, default: true },
  keywordPlaceholder: { type: String, default: '请输入关键字搜索' },
  showCreate: { type: Boolean, default: false },
  createText: { type: String, default: '新建' },
  showSelection: { type: Boolean, default: false },
  showBatchDelete: { type: Boolean, default: false },
  showIndex: { type: Boolean, default: false },
  showPagination: { type: Boolean, default: true },
  pageSize: { type: Number, default: 20 },
  pageSizes: { type: Array, default: () => [10, 20, 50, 100] },
  paginationLayout: { type: String, default: 'total, sizes, prev, pager, next' },
  operatorWidth: { type: [Number, String], default: 180 },
  operatorFixed: { type: String, default: 'right' },
  stripe: { type: Boolean, default: true },
  maxHeight: { type: [Number, String], default: undefined },
  // 批量删除处理器
  deleteFn: { type: Function, default: null },
  // 是否点击行触发 row-click
  rowClickable: { type: Boolean, default: false },
  autoLoad: { type: Boolean, default: true },
  // 分页大小本地缓存的 key，为空则不缓存
  storageKey: { type: String, default: '' }
})

const emit = defineEmits([
  'create',
  'row-click',
  'selection-change',
  'batch-delete',
  'operator',
  'load-success',
  'load-error'
])

const slots = useSlots()
const tableRef = ref(null)
const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const keyword = ref('')
const selectedRows = ref([])
const sortState = ref(null)

/* ---------------- 分页大小缓存 ---------------- */
function storageName() {
  return props.storageKey ? `iot_table_page_size_${props.storageKey}` : ''
}

function savedPageSize() {
  const key = storageName()
  if (key) {
    const saved = parseInt(localStorage.getItem(key))
    if (props.pageSizes.includes(saved)) return saved
  }
  return props.pageSize
}

const pageSizeInner = ref(savedPageSize())

function savePageSize(size) {
  const key = storageName()
  if (key) localStorage.setItem(key, String(size))
}

/* ---------------- 计算属性 ---------------- */
const hasActions = computed(() => props.operators.length > 0 || !!slots.actions)

const queryParams = computed(() => {
  const extra = {}
  Object.entries(props.params || {}).forEach(([k, v]) => {
    if (v !== '' && v !== null && v !== undefined) extra[k] = v
  })
  return {
    page: page.value,
    page_size: pageSizeInner.value,
    keyword: keyword.value.trim(),
    filter: props.filter,
    ...extra
  }
})

function visibleOperators(row) {
  return props.operators.filter((op) => (op.show ? op.show(row) : true))
}

/* ---------------- 单元格渲染 ---------------- */
function cellValue(row, col) {
  return row[col.displayProp || col.prop]
}

function tagTextOf(row, col) {
  if (typeof col.formatter === 'function') return col.formatter(row, col)
  const val = cellValue(row, col)
  const map = col.tagMap || {}
  return map[val] ?? (val === null || val === undefined || val === '' ? '-' : val)
}

function tagTypeOf(row, col) {
  if (typeof col.tagType === 'function') return col.tagType(row, col)
  const val = cellValue(row, col)
  if (['online', 'success', 'active', 'normal', 'running', '1', 1, true, '正常', '在线', '运行中', '启用'].includes(val)) return 'success'
  if (['error', 'fail', 'danger', 'disabled', '异常', '禁用', '已停止'].includes(val)) return 'danger'
  if (['offline', 'warning', 'pending', '离线', '警告', '待处理'].includes(val)) return 'warning'
  return 'info'
}

function formatDateTime(val) {
  if (!val) return '-'
  if (typeof val === 'number') {
    const d = new Date(val > 1e12 ? val : val * 1000)
    return d.toLocaleString('zh-CN', { hour12: false })
  }
  return String(val).replace('T', ' ').replace(/Z$/, '').substring(0, 19)
}

function formatMoney(val) {
  if (val === null || val === undefined || val === '') return '-'
  const n = Number(val)
  return Number.isNaN(n) ? val : `¥${n.toFixed(2)}`
}

function imageUrl(url, col) {
  if (!url) return ''
  if (/^http(s)?:\/\//.test(url) || url.startsWith('data:')) return url
  const base = col.baseUrl !== undefined ? col.baseUrl : (import.meta.env.VITE_API_BASE_URL || '')
  return base + url
}

function indexMethod(index) {
  return (page.value - 1) * pageSizeInner.value + index + 1
}

/* ---------------- 数据加载 ---------------- */
function normalize(res) {
  if (!res) return { list: [], total: 0 }
  let list = res.data
  if (!Array.isArray(list)) {
    list = res.list || res.items || res.records || res.data?.list || res.data?.items || res.data?.records || []
  }
  const totalNum = res.total ?? res.data?.total ?? list.length
  return { list, total: Number(totalNum) || 0 }
}

async function loadData() {
  loading.value = true
  try {
    const res = await props.loader(queryParams.value)
    const { list, total: t } = normalize(res)
    tableData.value = list
    total.value = t
    emit('load-success', res)
  } catch (e) {
    console.error('加载表格数据失败', e)
    emit('load-error', e)
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadData()
}

function handleKeywordSearch() {
  page.value = 1
  loadData()
}

function handleSizeChange(size) {
  pageSizeInner.value = size
  savePageSize(size)
  page.value = 1
  loadData()
}

function handleCurrentChange(current) {
  page.value = current
  loadData()
}

function handleSortChange({ prop, order }) {
  sortState.value = prop && order ? { prop, order: order === 'ascending' ? 'asc' : 'desc' } : null
  // 本项目后端暂不支持排序参数，仅保留状态供业务页面扩展
  page.value = 1
  loadData()
}

function handleRowClick(row, column) {
  if (!props.rowClickable) return
  if (column?.label === '操作' || column?.type === 'selection' || column?.type === 'index') return
  emit('row-click', row)
}

function handleSelectionChange(rows) {
  selectedRows.value = rows
  emit('selection-change', rows)
}

function clearSelection() {
  selectedRows.value = []
  tableRef.value?.clearSelection()
  emit('selection-change', [])
}

function handleCreate() {
  emit('create')
}

function handleOperator(op, row, index) {
  if (typeof op.handler === 'function') {
    op.handler(row, index)
  } else {
    emit('operator', op, row, index)
  }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(
      `确定删除选中的 ${selectedRows.value.length} 条记录吗？`,
      '提示',
      { type: 'warning' }
    )
    if (props.deleteFn) {
      await props.deleteFn(selectedRows.value.map((r) => r[props.rowKey]))
    } else {
      ElMessage.warning('未配置删除方法')
      return
    }
    ElMessage.success('删除成功')
    emit('batch-delete', selectedRows.value.map((r) => r[props.rowKey]))
    clearSelection()
    loadData()
  } catch (e) {
    if (e !== 'cancel') console.error(e)
  }
}

/* ---------------- 外部参数变化自动刷新 ---------------- */
watch(
  () => props.params,
  () => {
    page.value = 1
    loadData()
  },
  { deep: true }
)

watch(
  () => props.filter,
  () => {
    page.value = 1
    loadData()
  },
  { deep: true }
)

watch(
  () => props.loader,
  () => {
    page.value = 1
    loadData()
  }
)

onMounted(() => {
  if (props.autoLoad) loadData()
})

defineExpose({
  refresh: loadData,
  loadData,
  tableRef,
  clearSelection,
  selectedRows,
  page,
  pageSize: pageSizeInner,
  keyword
})
</script>

<style lang="scss" scoped>
.smart-table-card {
  :deep(.el-card__header) {
    padding: 14px 16px;
  }
  :deep(.el-card__body) {
    padding: 16px;
  }
}

.smart-table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;

  .smart-table-title {
    font-size: 16px;
    font-weight: 600;
    color: #303133;
  }
}

.smart-table-container {
  width: 100%;

  .smart-table-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 12px;
    flex-wrap: wrap;
  }

  .toolbar-left,
  .toolbar-right {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .toolbar-search {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .table-toolbar {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 8px 12px;
    background: #ecf5ff;
    border-radius: 4px;
    margin-bottom: 8px;

    .selection-info {
      color: #409eff;
      font-size: 14px;

      b {
        font-weight: 600;
      }
    }
  }

  .smart-table {
    width: 100%;

    :deep(.el-table__row) {
      cursor: default;
    }
  }

  .muted {
    color: #c0c4cc;
  }

  .smart-table-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }

  @media (max-width: 768px) {
    .smart-table-toolbar {
      flex-direction: column;
      align-items: stretch;
    }
    .toolbar-search :deep(.el-input) {
      width: 100% !important;
    }
  }
}
</style>
