<template>
  <div class="maintenance-stats">
    <PageHeader title="维保统计">
      <template #extra>
        <el-button :icon="Refresh" :loading="loading" @click="loadData">刷新</el-button>
      </template>
    </PageHeader>

    <div v-loading="loading" class="stats-body">
      <div class="stat-grid">
        <div class="stat-card">
          <div class="stat-label">计划数</div>
          <div class="stat-value">{{ stats.plans || 0 }}</div>
        </div>
        <div class="stat-card warning">
          <div class="stat-label">待处理</div>
          <div class="stat-value">{{ stats.pending || 0 }}</div>
        </div>
        <div class="stat-card primary">
          <div class="stat-label">处理中</div>
          <div class="stat-value">{{ stats.processing || 0 }}</div>
        </div>
        <div class="stat-card danger">
          <div class="stat-label">逾期</div>
          <div class="stat-value">{{ stats.overdue || 0 }}</div>
        </div>
        <div class="stat-card success">
          <div class="stat-label">本月完成</div>
          <div class="stat-value">{{ stats.month_completed || 0 }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">本月成本</div>
          <div class="stat-value">¥ {{ money(stats.month_cost) }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">平均耗时</div>
          <div class="stat-value">{{ hours(stats.avg_hours) }} <span class="unit">小时</span></div>
        </div>
      </div>

      <div class="panel-grid">
        <div class="panel">
          <div class="panel-title">状态分布</div>
          <div v-if="statusRows.length" class="status-list">
            <div v-for="row in statusRows" :key="row.label" class="status-row">
              <div class="status-head">
                <span>{{ row.label }}</span>
                <span class="count">{{ row.value }} 单</span>
              </div>
              <el-progress :percentage="statusPercent(row.value)" :stroke-width="10" :color="row.color" />
            </div>
          </div>
          <el-empty v-else description="暂无工单数据" :image-size="80" />
        </div>

        <div class="panel">
          <div class="panel-title">按类型分布</div>
          <el-table :data="typeRows" size="small" border empty-text="暂无数据">
            <el-table-column label="工单类型" min-width="140">
              <template #default="{ row }">{{ typeText(row.type) }}</template>
            </el-table-column>
            <el-table-column prop="count" label="工单数" width="90" />
            <el-table-column label="占比" width="140">
              <template #default="{ row }">
                <el-progress :percentage="typePercent(row.count)" :stroke-width="8" :show-text="false" />
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>

      <div class="panel">
        <div class="panel-title">负责人排行</div>
        <el-table :data="assigneeRows" size="small" border empty-text="暂无数据">
          <el-table-column label="排名" width="70">
            <template #default="{ $index }">{{ $index + 1 }}</template>
          </el-table-column>
          <el-table-column label="负责人" min-width="160">
            <template #default="{ row }">{{ row.assignee_name || row.assignee_id || '-' }}</template>
          </el-table-column>
          <el-table-column prop="count" label="工单数" width="110" />
          <el-table-column prop="completed" label="已完成" width="110" />
          <el-table-column label="完成率" width="200">
            <template #default="{ row }">
              <el-progress :percentage="completionRate(row)" :stroke-width="8" />
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="bottom-actions">
        <el-button type="primary" @click="$router.push('/maintenance/plans')">前往维保计划</el-button>
        <el-button @click="$router.push('/maintenance/orders')">前往维保工单</el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import { getMaintenanceStats } from '@/api/maintenance'

const loading = ref(false)
const stats = ref({})

const statusRows = computed(() => {
  const s = stats.value
  const rows = [
    { label: '待处理', value: Number(s.pending || 0), color: '#e6a23c' },
    { label: '处理中', value: Number(s.processing || 0), color: '#409eff' },
    { label: '待验收', value: Number(s.reviewing || 0), color: '#f2c94c' },
    { label: '已完成', value: Number(s.completed || 0), color: '#67c23a' },
    { label: '已取消', value: Number(s.cancelled || 0), color: '#909399' }
  ]
  return rows.filter(r => r.value > 0)
})

const typeRows = computed(() => stats.value.by_type || [])
const assigneeRows = computed(() => stats.value.by_assignee || [])

const statusTotal = computed(() => {
  const s = stats.value
  return Number(s.pending || 0) + Number(s.processing || 0) + Number(s.reviewing || 0) +
    Number(s.completed || 0) + Number(s.cancelled || 0)
})

const typeTotal = computed(() => typeRows.value.reduce((sum, row) => sum + Number(row.count || 0), 0))

const statusPercent = (value) => {
  const base = statusTotal.value
  if (!base) return 0
  return Math.min(100, Math.round((Number(value || 0) / base) * 100))
}

const typePercent = (value) => {
  const base = typeTotal.value
  if (!base) return 0
  return Math.min(100, Math.round((Number(value || 0) / base) * 100))
}

const completionRate = (row) => {
  const total = Number(row.count || 0)
  if (!total) return 0
  return Math.min(100, Math.round((Number(row.completed || 0) / total) * 100))
}

const money = (value) => Number(value || 0).toFixed(2)

const hours = (value) => Number(value || 0).toFixed(1)

const typeText = (t) => ({ plan: '计划保养', fault: '故障维修', patrol: '巡检', install: '安装' }[t] || t || '-')

const loadData = async () => {
  loading.value = true
  try {
    const res = await getMaintenanceStats()
    stats.value = res.data || {}
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style lang="scss" scoped>
.maintenance-stats {
  .stats-body {
    min-height: 200px;
  }

  .stat-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    gap: 14px;
    margin-bottom: 20px;
  }

  .stat-card {
    background: #fff;
    border-radius: 8px;
    padding: 16px 18px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
    border-left: 4px solid #c0c4cc;

    &.primary {
      border-left-color: #409eff;
    }

    &.success {
      border-left-color: #67c23a;
    }

    &.warning {
      border-left-color: #e6a23c;
    }

    &.danger {
      border-left-color: #f56c6c;
    }

    .stat-label {
      font-size: 13px;
      color: #909399;
      margin-bottom: 8px;
    }

    .stat-value {
      font-size: 24px;
      font-weight: 600;
      color: #303133;

      .unit {
        font-size: 12px;
        font-weight: 400;
        color: #909399;
      }
    }
  }

  .panel-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 14px;
    margin-bottom: 14px;
  }

  .panel {
    background: #fff;
    border-radius: 8px;
    padding: 16px 18px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
    margin-bottom: 14px;

    .panel-title {
      font-weight: 600;
      font-size: 15px;
      margin-bottom: 14px;
    }
  }

  .status-list {
    .status-row {
      margin-bottom: 14px;

      .status-head {
        display: flex;
        justify-content: space-between;
        font-size: 13px;
        color: #606266;
        margin-bottom: 4px;

        .count {
          color: #909399;
        }
      }
    }
  }

  .bottom-actions {
    display: flex;
    justify-content: center;
    gap: 12px;
    margin-top: 10px;
  }
}
</style>
