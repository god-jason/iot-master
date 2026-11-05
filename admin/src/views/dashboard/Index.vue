<template>
  <div class="dashboard">
    <div class="page-header">
      <h2 class="page-title">控制台</h2>
      <span class="page-time">{{ currentTime }}</span>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon device">
              <el-icon :size="28"><Cpu /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.deviceCount }}</div>
              <div class="stat-label">设备总数</div>
            </div>
            <div class="stat-trend up">
              <el-icon><Top /></el-icon> 12%
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon online">
              <el-icon :size="28"><Connection /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.onlineCount }}</div>
              <div class="stat-label">在线设备</div>
            </div>
            <div class="stat-trend up">
              <el-icon><Top /></el-icon> 8%
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon model">
              <el-icon :size="28"><Box /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.modelCount }}</div>
              <div class="stat-label">物模型数量</div>
            </div>
            <div class="stat-trend down">
              <el-icon><Bottom /></el-icon> 3%
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Cpu, Connection, Box, Top, Bottom } from '@element-plus/icons-vue'

const currentTime = ref('')
const stats = ref({
  deviceCount: 128,
  onlineCount: 86,
  modelCount: 24
})

let timer = null

import dayjs from 'dayjs'

const updateTime = () => {
  currentTime.value = dayjs().format('YYYY/MM/DD HH:mm:ss')
}

onMounted(() => {
  updateTime()
  timer = setInterval(updateTime, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style lang="scss" scoped>
.dashboard {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
    
    .page-title {
      font-size: 24px;
      font-weight: 600;
      color: #1a1a2e;
      margin: 0;
    }
    
    .page-time {
      color: #999;
      font-size: 14px;
    }
  }
  
  .stats-row {
    margin-bottom: 20px;
  }
  
  .stat-card {
    border-radius: 12px;
    transition: all 0.3s;
    
    &:hover {
      transform: translateY(-4px);
    }
    
    .stat-content {
      display: flex;
      align-items: center;
      gap: 16px;
      position: relative;
    }
    
    .stat-icon {
      width: 56px;
      height: 56px;
      border-radius: 14px;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #fff;
      
      &.device { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
      &.online { background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%); }
      &.model { background: linear-gradient(135deg, #fc4a1a 0%, #f7b733 100%); }
    }
    
    .stat-info {
      .stat-value {
        font-size: 28px;
        font-weight: 700;
        color: #1a1a2e;
        line-height: 1.2;
      }
      
      .stat-label {
        font-size: 14px;
        color: #999;
        margin-top: 4px;
      }
    }
    
    .stat-trend {
      position: absolute;
      right: 0;
      top: 0;
      font-size: 12px;
      display: flex;
      align-items: center;
      gap: 2px;
      padding: 4px 8px;
      border-radius: 6px;
      
      &.up {
        color: #67c23a;
        background: rgba(103, 194, 58, 0.1);
      }
      
      &.down {
        color: #f56c6c;
        background: rgba(245, 108, 108, 0.1);
      }
    }
  }
}
</style>