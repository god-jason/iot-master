<template>
  <div class="device-group-page">
    <div class="page-header">
      <h2 class="page-title">分组管理</h2>
      <el-button type="primary" @click="openCreateModal">
        <el-icon><Plus /></el-icon>新建分组
      </el-button>
    </div>

    <el-card class="group-card">
      <el-tree
        :data="groupTree"
        :props="treeProps"
        :expand-on-click-node="false"
        default-expand-all
        node-key="id"
        class="group-tree"
      >
        <template #default="{ node, data }">
          <div class="tree-node">
            <span>{{ node.label }}</span>
            <span class="node-actions">
              <el-icon class="action-icon" @click="editGroup(data)"><Edit /></el-icon>
              <el-icon class="action-icon delete" @click="deleteGroup(data)"><Delete /></el-icon>
            </span>
          </div>
        </template>
      </el-tree>
    </el-card>

    <el-dialog v-model="showModal" :title="modalTitle" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="分组名称" required>
          <el-input v-model="form.name" placeholder="请输入分组名称" />
        </el-form-item>
        <el-form-item label="父级分组">
          <el-select v-model="form.parent_id" placeholder="请选择父级分组" style="width: 100%">
            <el-option label="无（顶级分组）" value="" />
            <el-option
              v-for="group in parentOptions"
              :key="group.id"
              :label="group.name"
              :value="group.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="closeModal">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="loading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Plus, Edit, Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { searchGroups, createGroup, updateGroup, deleteGroup as delGroup } from '@/api/group'

const loading = ref(false)
const showModal = ref(false)
const modalTitle = ref('新建分组')
const groups = ref([])
const form = ref({
  id: '',
  name: '',
  parent_id: ''
})

const treeProps = {
  children: 'children',
  label: 'name'
}

const groupTree = computed(() => {
  const map = new Map()
  const roots = []
  const ZERO_ID = '000000000000000000000000'

  groups.value.forEach(group => {
    map.set(group.id, { ...group, children: [] })
  })

  map.forEach(group => {
    const pid = group.parent_id
    if (pid && pid !== ZERO_ID) {
      const parent = map.get(pid)
      if (parent) {
        parent.children.push(group)
      } else {
        // parent 已不存在时也放到根
        roots.push(group)
      }
    } else {
      roots.push(group)
    }
  })

  return roots
})

const parentOptions = computed(() => {
  return groups.value.filter(g => g.id !== form.value.id)
})

const loadData = async () => {
  try {
    const res = await searchGroups({ page: 1, page_size: 1000 })
    groups.value = res.data || []
  } catch (e) {
    console.error(e)
    ElMessage.error('加载失败')
  }
}

const openCreateModal = () => {
  modalTitle.value = '新建分组'
  form.value = {
    id: '',
    name: '',
    parent_id: ''
  }
  showModal.value = true
}

const editGroup = (data) => {
  modalTitle.value = '编辑分组'
  form.value = {
    id: data.id,
    name: data.name,
    parent_id: data.parent_id || ''
  }
  showModal.value = true
}

const deleteGroup = async (data) => {
  try {
    await ElMessageBox.confirm('确定要删除该分组吗？', '提示', {
      type: 'warning'
    })
    await delGroup(data.id)
    ElMessage.success('删除成功')
    loadData()
  } catch (e) {
    console.error(e)
  }
}

const closeModal = () => {
  showModal.value = false
}

const submitForm = async () => {
  if (!form.value.name.trim()) {
    ElMessage.warning('请输入分组名称')
    return
  }

  loading.value = true
  try {
    if (form.value.id) {
      await updateGroup(form.value.id, {
        name: form.value.name,
        parent_id: form.value.parent_id || null
      })
      ElMessage.success('更新成功')
    } else {
      await createGroup({
        name: form.value.name,
        parent_id: form.value.parent_id || null
      })
      ElMessage.success('创建成功')
    }
    showModal.value = false
    loadData()
  } catch (e) {
    console.error(e)
    ElMessage.error('操作失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style lang="scss" scoped>
.device-group-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;

    .page-title {
      font-size: 20px;
      font-weight: 600;
      margin: 0;
    }
  }

  .group-card {
    border-radius: 12px;

    .group-tree {
      :deep(.el-tree-node) {
        .tree-node {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 8px 0;

          .node-actions {
            visibility: hidden;
            display: flex;
            gap: 8px;

            .action-icon {
              font-size: 18px;
              color: #909399;
              cursor: pointer;
              padding: 4px;
              border-radius: 4px;
              transition: all 0.2s;

              &:hover {
                color: #409eff;
                background: #ecf5ff;
              }

              &.delete:hover {
                color: #f56c6c;
                background: #fef0f0;
              }
            }
          }
        }

        &:hover .node-actions {
          visibility: visible;
        }
      }
    }
  }
}
</style>
