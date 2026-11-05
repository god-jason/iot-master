<template>
  <div class="device-events-page">
    <div class="page-header">
      <h2 class="page-title">设备事件</h2>
    </div>

    <el-card class="events-card">
      <DeviceEvents />
    </el-card>
  </div>
</template>

<script setup>
import DeviceEvents from './DeviceEvents.vue'
</script>

<style lang="scss" scoped>
.device-events-page {
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

  .events-card {
    border-radius: 12px;

    :deep(.el-table__header-wrapper) {
      .el-table__header {
        background: #f5f7fa;
      }
    }
  }
}
</style>