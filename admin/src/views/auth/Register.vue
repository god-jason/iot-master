<template>
  <div class="register-container">
    <div class="bg-effects">
      <div class="grid-lines"></div>
      <div class="glow-orb orb1"></div>
      <div class="glow-orb orb2"></div>
    </div>

    <div class="register-box animate-slide-in">
      <div class="register-header">
        <h2>用户注册</h2>
        <p>暂未开放注册功能，敬请期待</p>
      </div>

      <div class="coming-soon">
        <div class="icon-box">
          <el-icon :size="60" color="#00d2ff"><Clock /></el-icon>
        </div>
        <h3>注册功能正在开发中</h3>
        <p>我们正在努力开发注册功能，届时您将可以：</p>
        <ul>
          <li>创建您的个人账户</li>
          <li>管理您的IoT设备</li>
          <li>使用物模型库和应用市场</li>
        </ul>
        <el-button type="primary" size="large" class="back-btn" @click="$router.push('/login')">
          返回登录
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Clock } from '@element-plus/icons-vue'
</script>

<style lang="scss" scoped>
.register-container {
  width: 100%;
  height: 100vh;
  background: linear-gradient(135deg, #0a0e27 0%, #1a1f3a 50%, #0d1225 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}

.bg-effects {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.grid-lines {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(0, 210, 255, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 210, 255, 0.03) 1px, transparent 1px);
  background-size: 50px 50px;
}

.glow-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
  opacity: 0.3;
}

.orb1 {
  width: 300px;
  height: 300px;
  background: #00d2ff;
  top: 10%;
  left: 20%;
}

.orb2 {
  width: 200px;
  height: 200px;
  background: #0066ff;
  bottom: 20%;
  right: 25%;
}

.register-box {
  width: 480px;
  background: rgba(26, 31, 58, 0.8);
  backdrop-filter: blur(20px);
  border-radius: 20px;
  border: 1px solid rgba(0, 210, 255, 0.2);
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.5);
  padding: 50px 40px;
  position: relative;
  z-index: 1;
}

.register-header {
  text-align: center;
  margin-bottom: 40px;

  h2 {
    font-size: 28px;
    color: #e0e6ed;
    margin-bottom: 8px;
  }

  p {
    font-size: 14px;
    color: rgba(224, 230, 237, 0.6);
  }
}

.coming-soon {
  text-align: center;
  color: #e0e6ed;

  .icon-box {
    width: 120px;
    height: 120px;
    margin: 0 auto 30px;
    background: rgba(0, 210, 255, 0.1);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid rgba(0, 210, 255, 0.3);
    animation: glow 2s ease-in-out infinite;
  }

  h3 {
    font-size: 20px;
    margin-bottom: 15px;
    color: #00d2ff;
  }

  p {
    color: rgba(224, 230, 237, 0.7);
    margin-bottom: 20px;
  }

  ul {
    list-style: none;
    text-align: left;
    padding: 0 40px;
    margin-bottom: 30px;

    li {
      padding: 8px 0;
      color: rgba(224, 230, 237, 0.8);
      position: relative;
      padding-left: 20px;

      &::before {
        content: '';
        position: absolute;
        left: 0;
        top: 50%;
        transform: translateY(-50%);
        width: 6px;
        height: 6px;
        background: #00d2ff;
        border-radius: 50%;
      }
    }
  }
}

.back-btn {
  width: 200px;
  background: linear-gradient(135deg, #00d2ff, #0066ff);
  border: none;
  border-radius: 8px;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 10px 30px rgba(0, 210, 255, 0.4);
  }
}
</style>
