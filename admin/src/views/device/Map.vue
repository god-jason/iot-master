<template>
  <div class="device-map">
    <div class="page-header">
      <h2 class="page-title">设备地图分布</h2>
    </div>
    <el-card class="map-card">
      <div class="map-placeholder">
        <el-icon :size="80" color="#00d2ff"><Location /></el-icon>
        <p>地图展示区域</p>
        <span>集成高德/百度地图显示设备分布位置</span>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { Location } from '@element-plus/icons-vue'
</script>

<style lang="scss" scoped>
.device-map {
  .page-header {
    margin-bottom: 20px;

    .page-title {
      font-size: 24px;
      color: #e0e6ed;
    }
  }

  .map-card {
    background: rgba(26, 31, 58, 0.8);
    border: 1px solid rgba(0, 210, 255, 0.1);

    :deep(.el-card__body) {
      height: calc(100vh - 200px);
    }
  }

  .map-placeholder {
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: rgba(224, 230, 237, 0.6);

    p {
      font-size: 18px;
      margin-top: 20px;
    }

    span {
      font-size: 14px;
      margin-top: 10px;
    }
  }
}
</style>
