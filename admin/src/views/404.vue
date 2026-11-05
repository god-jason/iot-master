<template>
  <div class="not-found-container">
    <div class="content">
      <div class="error-code">404</div>
      <div class="error-message">页面不存在</div>
      <div class="error-description">抱歉，您访问的页面不存在或已被删除</div>
      <el-button type="primary" @click="goHome">返回首页</el-button>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const goHome = () => {
  router.push('/')
}
</script>

<style lang="scss" scoped>
.not-found-container {
  width: 100%;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;

  .content {
    text-align: center;
  }

  .error-code {
    font-size: 120px;
    font-weight: 700;
    color: #409eff;
    line-height: 1;
    margin-bottom: 20px;
  }

  .error-message {
    font-size: 32px;
    margin-bottom: 10px;
  }

  .error-description {
    font-size: 16px;
    color: #999;
    margin-bottom: 30px;
  }
}
</style>
