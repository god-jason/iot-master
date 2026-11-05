<template>
  <div class="page-header">
    <div class="header-left">
      <el-button type="text" @click="$router.back()" v-if="showBack">
        <el-icon><ArrowLeft /></el-icon>
      </el-button>
      <h2 class="page-title">{{ title }}</h2>
    </div>
    <div class="header-right">
      <slot name="extra"></slot>
    </div>
  </div>
</template>

<script setup>
import { ArrowLeft } from '@element-plus/icons-vue'

defineProps({
  title: {
    type: String,
    required: true
  },
  showBack: {
    type: Boolean,
    default: false
  }
})
</script>

<style lang="scss" scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  margin-bottom: 20px;
  border-bottom: 1px solid #ebeef5;

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;

    .page-title {
      font-size: 20px;
      font-weight: 600;
      color: #303133;
      margin: 0;
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }
}
</style>