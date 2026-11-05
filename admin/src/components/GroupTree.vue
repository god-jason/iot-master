<template>
  <div class="group-tree">
    <div class="tree-header">
      <span class="title">设备分组</span>
      <div class="actions">
        <el-button text @click="handleRefresh" title="刷新">
          <el-icon><Refresh /></el-icon>
        </el-button>
        <el-button text @click="handleCreateRoot" title="新建分组">
          <el-icon><Plus /></el-icon>
        </el-button>
      </div>
    </div>

    <el-tree
      ref="treeRef"
      :data="treeData"
      :props="{ children: 'children', label: 'name' }"
      node-key="id"
      default-expand-all
      :expand-on-click-node="false"
      :show-checkbox="showCheckbox"
      :check-strictly="checkStrictly"
      :default-checked-keys="checkedKeys"
      @check="handleCheck"
      @node-click="handleNodeClick"
    >
      <template #default="{ node, data }">
        <span class="tree-node">
          <span class="node-label">{{ data.name }}</span>
          <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, data)">
            <el-button link size="small" class="node-action">
              <el-icon><MoreFilled /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="rename">重命名</el-dropdown-item>
                <el-dropdown-item command="createChild">新建子分组</el-dropdown-item>
                <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </span>
      </template>
    </el-tree>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="400px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="分组名称">
          <el-input v-model="form.name" placeholder="请输入分组名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, MoreFilled } from '@element-plus/icons-vue'
import { searchGroups, createGroup, updateGroup, deleteGroup } from '@/api/group'

const props = defineProps({
  showCheckbox: {
    type: Boolean,
    default: true
  },
  checkStrictly: {
    type: Boolean,
    default: true
  },
  checkedKeys: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['check-change', 'node-click', 'refresh'])

const treeRef = ref(null)
const treeData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('新建分组')
const form = ref({
  id: null,
  name: '',
  parent_id: null
})
const actionMode = ref('create')

const loadData = async () => {
  try {
    const res = await searchGroups({ page: 1, page_size: 1000 })
    if (res.code === 0) {
      treeData.value = buildTree(res.data || [])
    }
  } catch (e) {
    console.error('加载分组失败:', e)
  }
}

const buildTree = (list) => {
  const map = {}
  const roots = []
  list.forEach(item => {
    map[item.id] = { ...item, children: [] }
  })
  list.forEach(item => {
    const node = map[item.id]
    if (item.parent_id && map[item.parent_id]) {
      map[item.parent_id].children.push(node)
    } else {
      roots.push(node)
    }
  })
  return roots
}

const handleRefresh = () => {
  loadData()
  emit('refresh')
}

const handleCreateRoot = () => {
  actionMode.value = 'create'
  dialogTitle.value = '新建分组'
  form.value = { id: null, name: '', parent_id: null }
  dialogVisible.value = true
}

const handleCommand = (cmd, data) => {
  switch (cmd) {
    case 'rename':
      actionMode.value = 'rename'
      dialogTitle.value = '重命名'
      form.value = { id: data.id, name: data.name, parent_id: data.parent_id }
      dialogVisible.value = true
      break
    case 'createChild':
      actionMode.value = 'create'
      dialogTitle.value = '新建子分组'
      form.value = { id: null, name: '', parent_id: data.id }
      dialogVisible.value = true
      break
    case 'delete':
      handleDelete(data)
      break
  }
}

const handleDelete = (data) => {
  ElMessageBox.confirm(`确定要删除分组"${data.name}"吗？`, '提示', {
    type: 'warning'
  }).then(async () => {
    const res = await deleteGroup(data.id)
    if (res.code === 0) {
      ElMessage.success('删除成功')
      loadData()
    }
  }).catch(() => {})
}

const handleSubmit = async () => {
  if (!form.value.name) {
    ElMessage.warning('请输入分组名称')
    return
  }
  try {
    let res
    if (actionMode.value === 'create') {
      res = await createGroup(form.value)
    } else {
      res = await updateGroup(form.value.id, { name: form.value.name })
    }
    if (res.code === 0) {
      ElMessage.success(actionMode.value === 'create' ? '创建成功' : '更新成功')
      dialogVisible.value = false
      loadData()
    }
  } catch (e) {
    console.error(e)
  }
}

const handleCheck = (data, info) => {
  emit('check-change', info.checkedKeys, data)
}

const handleNodeClick = (data) => {
  emit('node-click', data)
}

const getCheckedNodes = () => {
  return treeRef.value?.getCheckedNodes() || []
}

const setCheckedKeys = (keys) => {
  treeRef.value?.setCheckedKeys(keys)
}

defineExpose({
  loadData,
  getCheckedNodes,
  setCheckedKeys
})

onMounted(() => {
  loadData()
})
</script>

<style lang="scss" scoped>
.group-tree {
  height: 100%;
  display: flex;
  flex-direction: column;

  .tree-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid #ebeef5;

    .title {
      font-weight: 600;
      font-size: 14px;
      color: #303133;
    }

    .actions {
      display: flex;
      gap: 4px;
    }
  }

  :deep(.el-tree) {
    flex: 1;
    overflow-y: auto;
    border: none;
  }

  .tree-node {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-right: 8px;

    .node-label {
      font-size: 14px;
    }

    .node-action {
      opacity: 0;
      transition: opacity 0.2s;
    }
  }

  :deep(.el-tree-node__content:hover .node-action) {
    opacity: 1;
  }
}
</style>
