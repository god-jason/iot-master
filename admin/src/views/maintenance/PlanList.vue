<template>
  <div class="plan-list">
    <SmartTable
      ref="tableRef"
      card-title="维保计划"
      :loader="searchMaintenancePlans"
      :columns="columns"
      :operators="operators"
      :params="queryParams"
      :operator-width="300"
      keyword-placeholder="搜索计划名称"
      storage-key="maintenance-plan"
    >
      <template #toolbar-left>
        <el-select v-model="scopeFilter" placeholder="全部范围" clearable style="width: 130px">
          <el-option v-for="s in scopeOptions" :key="s.value" :label="s.label" :value="s.value" />
        </el-select>
        <el-select v-model="disabledFilter" placeholder="全部状态" clearable style="width: 120px">
          <el-option label="启用中" :value="false" />
          <el-option label="已停用" :value="true" />
        </el-select>
      </template>

      <template #toolbar-right>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建计划</el-button>
      </template>

      <template #scope="{ row }">
        <div class="scope-cell">
          <el-tag size="small" :type="scopeTagType(row.scope_type)">{{ scopeText(row.scope_type) }}</el-tag>
          <span class="scope-name">{{ row.scope_name || row.scope_id || '-' }}</span>
        </div>
      </template>

      <template #cycle="{ row }">{{ cycleText(row) }}</template>

      <template #remind="{ row }">{{ row.remind_days || 0 }} 天</template>

      <template #nextDue="{ row }">{{ fmt(row.next_due) }}</template>

      <template #owner="{ row }">{{ row.owner_name || row.owner_id || '-' }}</template>
    </SmartTable>

    <!-- 新建 / 编辑计划 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑维保计划' : '新建维保计划'" width="760px" destroy-on-close>
      <el-form v-loading="formLoading" :model="form" label-width="110px">
        <el-form-item label="计划名称" required>
          <el-input v-model="form.name" placeholder="如：空调季度保养" />
        </el-form-item>
        <el-form-item label="维保范围" required>
          <el-radio-group v-model="form.scope_type" @change="onScopeTypeChange">
            <el-radio-button v-for="s in scopeOptions" :key="s.value" :value="s.value">{{ s.label }}</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="scopeSelectLabel" required>
          <el-select
            v-if="form.scope_type === 'model'"
            v-model="form.scope_id"
            placeholder="选择物模型"
            filterable
            style="width: 100%"
          >
            <el-option v-for="m in models" :key="m.id" :label="m.name" :value="String(m.id)" />
          </el-select>
          <el-select
            v-else-if="form.scope_type === 'group'"
            v-model="form.scope_id"
            placeholder="选择设备分组"
            filterable
            style="width: 100%"
          >
            <el-option v-for="g in groups" :key="g.id" :label="g.name" :value="String(g.id)" />
          </el-select>
          <el-tree-select
            v-else-if="form.scope_type === 'org'"
            v-model="form.scope_id"
            :data="orgTree"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            node-key="id"
            check-strictly
            placeholder="选择组织（含子组织）"
            style="width: 100%"
          />
          <el-select
            v-else
            v-model="form.scope_id"
            filterable
            remote
            :remote-method="searchDeviceOptions"
            :loading="deviceLoading"
            placeholder="输入设备名称/SN 搜索"
            style="width: 100%"
            @visible-change="(v) => v && searchDeviceOptions('')"
          >
            <el-option
              v-for="d in deviceOptions"
              :key="d.id"
              :label="`${d.name}（${d.sn || d.id}）`"
              :value="String(d.id)"
            />
          </el-select>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="维保周期" required>
              <div class="cycle-editor">
                每
                <el-input-number v-model="form.cycle_value" :min="1" :max="3650" controls-position="right" />
                <el-select v-model="form.cycle_type" style="width: 100px">
                  <el-option label="天" value="day" />
                  <el-option label="个月" value="month" />
                  <el-option label="年" value="year" />
                </el-select>
              </div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="提前提醒">
              <el-input-number v-model="form.remind_days" :min="0" :max="365" controls-position="right" style="width: 140px" />
              <span class="hint">天</span>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="负责人">
              <el-select
                v-model="form.owner_id"
                filterable
                remote
                :remote-method="searchOwnerOptions"
                :loading="ownerLoading"
                placeholder="输入姓名/用户名搜索"
                clearable
                style="width: 100%"
              >
                <el-option v-for="u in ownerOptions" :key="u.id" :label="u.label" :value="u.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="下次维保时间">
              <el-date-picker v-model="form.next_due" type="datetime" placeholder="选择时间" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="检查项模板">
          <div class="items-editor">
            <div class="items-head">
              <span>检查项名称</span>
              <span>检查方法</span>
              <span></span>
            </div>
            <div v-for="(item, i) in form.items" :key="i" class="items-row">
              <el-input v-model="item.name" placeholder="如：外观检查" />
              <el-input v-model="item.method" placeholder="如：目视检查外壳" />
              <el-button link type="danger" @click="removeItem(i)">删除</el-button>
            </div>
            <el-button link type="primary" :icon="Plus" @click="addItem">添加检查项</el-button>
          </div>
        </el-form-item>

        <el-form-item label="启停状态">
          <el-switch v-model="form.disabled" active-text="停用" inactive-text="启用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitForm">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import SmartTable from '@/components/SmartTable.vue'
import { searchModels } from '@/api/model'
import { searchGroups } from '@/api/group'
import { getOrgTree } from '@/api/org'
import { searchUsers } from '@/api/org'
import { searchDevices } from '@/api/device'
import { formatDateTime } from '@/utils/format'
import {
  searchMaintenancePlans,
  getMaintenancePlan,
  createMaintenancePlan,
  updateMaintenancePlan,
  deleteMaintenancePlan,
  switchMaintenancePlan,
  generateMaintenancePlan
} from '@/api/maintenance'

const tableRef = ref(null)
const scopeFilter = ref('')
const disabledFilter = ref('')

const scopeOptions = [
  { value: 'model', label: '按物模型' },
  { value: 'group', label: '按分组' },
  { value: 'org', label: '按组织' },
  { value: 'device', label: '指定设备' }
]

const models = ref([])
const groups = ref([])
const orgTree = ref([])
const deviceOptions = ref([])
const deviceLoading = ref(false)
const ownerOptions = ref([])
const ownerLoading = ref(false)

const dialogVisible = ref(false)
const formLoading = ref(false)
const saving = ref(false)
const form = ref(defaultForm())

function defaultForm() {
  return {
    id: null,
    name: '',
    scope_type: 'model',
    scope_id: '',
    cycle_type: 'day',
    cycle_value: 30,
    remind_days: 3,
    owner_id: '',
    next_due: null,
    items: [{ name: '', method: '' }],
    disabled: false
  }
}

const scopeSelectLabel = computed(() => {
  return { model: '选择物模型', group: '选择分组', org: '选择组织', device: '选择设备' }[form.value.scope_type] || '维保范围'
})

const fmt = (val) => {
  const text = formatDateTime(val, 'YYYY-MM-DD HH:mm')
  return text || '-'
}

const toDate = (val) => {
  if (!val) return null
  if (val instanceof Date) return val
  if (typeof val === 'number') return new Date(val)
  if (typeof val === 'string' && /^\d+$/.test(val)) return new Date(Number(val))
  const d = new Date(val)
  return Number.isNaN(d.getTime()) ? null : d
}

const parseItems = (val) => {
  if (!val) return []
  if (Array.isArray(val)) return val
  if (typeof val === 'string') {
    try {
      const arr = JSON.parse(val)
      return Array.isArray(arr) ? arr : []
    } catch (e) {
      return []
    }
  }
  return []
}

const scopeText = (s) => ({ model: '物模型', group: '分组', org: '组织', device: '设备' }[s] || s || '-')
const scopeTagType = (s) => ({ model: 'primary', group: 'warning', org: 'success', device: 'info' }[s] || 'info')

const cycleText = (row) => {
  const unit = { day: '天', month: '个月', year: '年' }[row.cycle_type] || ''
  return `每 ${row.cycle_value || 1} ${unit}`
}

const userLabel = (u) => (u.nickname ? `${u.nickname}（${u.username}）` : (u.username || u.id))

const columns = [
  { prop: 'name', label: '计划名称', minWidth: 170 },
  { label: '维保范围', minWidth: 200, slot: 'scope', showOverflowTooltip: false },
  { label: '周期', width: 120, slot: 'cycle' },
  { label: '提前提醒', width: 100, slot: 'remind' },
  { label: '下次维保时间', width: 170, slot: 'nextDue' },
  { label: '负责人', width: 120, slot: 'owner' },
  {
    prop: 'disabled',
    label: '状态',
    width: 90,
    type: 'tag',
    formatter: (row) => (row.disabled ? '已停用' : '启用中'),
    tagType: (row) => (row.disabled ? 'info' : 'success')
  }
]

const operators = [
  { key: 'edit', label: '编辑', type: 'primary', handler: (row) => openEdit(row) },
  {
    key: 'switch',
    label: (row) => (row.disabled ? '启用' : '停用'),
    type: (row) => (row.disabled ? 'success' : 'warning'),
    handler: (row) => handleSwitch(row)
  },
  { key: 'generate', label: '立即生成', type: 'success', handler: (row) => handleGenerate(row) },
  { key: 'delete', label: '删除', type: 'danger', handler: (row) => handleDelete(row) }
]

const queryParams = computed(() => {
  const payload = { order_by: 'created', order: 'desc' }
  if (scopeFilter.value) payload.scope_type = scopeFilter.value
  if (disabledFilter.value !== '' && disabledFilter.value !== undefined) payload.disabled = disabledFilter.value
  return payload
})

const normalizeTree = (nodes) => (nodes || []).map(n => ({ ...n, id: String(n.id), children: normalizeTree(n.children) }))

const loadRefs = async () => {
  try {
    const res = await searchModels({ page: 1, page_size: 100 })
    models.value = res.data || []
  } catch (e) { /* ignore */ }
  try {
    const res = await searchGroups({ page: 1, page_size: 100 })
    groups.value = res.data || []
  } catch (e) { /* ignore */ }
  try {
    const res = await getOrgTree()
    // 组织 ID 统一转为字符串，与 plan.scope_id 保持一致
    orgTree.value = normalizeTree(res.data)
  } catch (e) { /* ignore */ }
}

onMounted(() => {
  loadRefs()
})

const searchDeviceOptions = async (kw) => {
  deviceLoading.value = true
  try {
    const res = await searchDevices({ page: 1, page_size: 20, keyword: kw || '' })
    deviceOptions.value = res.data || []
  } finally {
    deviceLoading.value = false
  }
}

const searchOwnerOptions = async (kw) => {
  ownerLoading.value = true
  try {
    const res = await searchUsers(kw || '')
    ownerOptions.value = (res.data || []).map(u => ({ id: u.id, label: userLabel(u) }))
  } finally {
    ownerLoading.value = false
  }
}

const openCreate = () => {
  form.value = defaultForm()
  dialogVisible.value = true
  if (!ownerOptions.value.length) searchOwnerOptions('')
}

const openEdit = async (row) => {
  form.value = defaultForm()
  dialogVisible.value = true
  formLoading.value = true
  try {
    await searchOwnerOptions('')
    const res = await getMaintenancePlan(row.id)
    const plan = res.data?.plan || res.data || row
    const items = parseItems(plan.items)
    form.value = {
      id: plan.id || row.id,
      name: plan.name || '',
      scope_type: plan.scope_type || 'model',
      scope_id: String(plan.scope_id || ''),
      cycle_type: plan.cycle_type || 'day',
      cycle_value: plan.cycle_value || 1,
      remind_days: plan.remind_days || 0,
      owner_id: plan.owner_id || '',
      next_due: toDate(plan.next_due),
      items: items.length ? items.map(i => ({ name: i.name || '', method: i.method || '' })) : [{ name: '', method: '' }],
      disabled: !!plan.disabled
    }
    if (plan.owner_id && plan.owner_name) {
      const exists = ownerOptions.value.some(u => u.id === plan.owner_id)
      if (!exists) ownerOptions.value.push({ id: plan.owner_id, label: plan.owner_name })
    }
  } finally {
    formLoading.value = false
  }
}

const onScopeTypeChange = () => {
  form.value.scope_id = ''
}

const addItem = () => {
  form.value.items.push({ name: '', method: '' })
}

const removeItem = (index) => {
  form.value.items.splice(index, 1)
}

const submitForm = async () => {
  if (!form.value.name) {
    ElMessage.warning('请填写计划名称')
    return
  }
  if (!form.value.scope_id) {
    ElMessage.warning('请选择维保范围')
    return
  }
  const items = form.value.items
    .filter(i => i.name && i.name.trim())
    .map(i => ({ name: i.name.trim(), method: (i.method || '').trim() }))
  const payload = {
    name: form.value.name.trim(),
    scope_type: form.value.scope_type,
    scope_id: String(form.value.scope_id || ''),
    cycle_type: form.value.cycle_type,
    cycle_value: Number(form.value.cycle_value || 1),
    remind_days: Number(form.value.remind_days || 0),
    items: JSON.stringify(items),
    owner_id: form.value.owner_id || '',
    disabled: !!form.value.disabled
  }
  // 后端时间为 RFC3339 字符串，未选择时不提交，交由后端按周期计算
  if (form.value.next_due) payload.next_due = new Date(form.value.next_due).toISOString()
  saving.value = true
  try {
    if (form.value.id) {
      await updateMaintenancePlan(form.value.id, payload)
      ElMessage.success('计划已更新')
    } else {
      await createMaintenancePlan(payload)
      ElMessage.success('计划已创建')
    }
    dialogVisible.value = false
    tableRef.value?.refresh()
  } finally {
    saving.value = false
  }
}

const handleSwitch = async (row) => {
  const next = !row.disabled
  await ElMessageBox.confirm(`确认${next ? '停用' : '启用'}计划「${row.name}」？`, '提示', { type: 'warning' })
  await switchMaintenancePlan(row.id, next)
  ElMessage.success(next ? '已停用' : '已启用')
  tableRef.value?.refresh()
}

const handleGenerate = async (row) => {
  await ElMessageBox.confirm(`确认按计划「${row.name}」立即生成维保工单？`, '立即生成', { type: 'warning' })
  const res = await generateMaintenancePlan(row.id)
  const generated = res.data?.generated || 0
  const count = res.data?.total || 0
  ElMessage.success(count ? `已生成 ${generated} 个工单，覆盖设备 ${count} 台` : `已生成 ${generated} 个工单`)
  tableRef.value?.refresh()
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm(`确认删除计划「${row.name}」？`, '删除计划', { type: 'warning' })
  await deleteMaintenancePlan(row.id)
  ElMessage.success('已删除')
  tableRef.value?.refresh()
}
</script>

<style lang="scss" scoped>
.plan-list {
  .scope-cell {
    display: flex;
    align-items: center;
    gap: 8px;

    .scope-name {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .cycle-editor {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .hint {
    margin-left: 8px;
    font-size: 12px;
    color: #909399;
  }

  .items-editor {
    width: 100%;
    border: 1px solid #ebeef5;
    border-radius: 6px;
    padding: 10px;

    .items-head,
    .items-row {
      display: grid;
      grid-template-columns: 1fr 1fr 60px;
      gap: 10px;
      align-items: center;
    }

    .items-head {
      font-size: 12px;
      color: #909399;
      margin-bottom: 8px;
    }

    .items-row {
      margin-bottom: 8px;
    }
  }
}
</style>
