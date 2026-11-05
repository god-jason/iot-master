<template>
  <div class="login-container">
    <div class="login-box">
      <!-- 左侧品牌区 -->
      <div class="brand-section">
        <div class="brand-content">
          <h1 class="brand-title">{{ oemStore.getSystemName }}</h1>
          <p class="brand-subtitle">智能物联 · 万物互联</p>
          <div class="feature-list">
            <div class="feature-item">
              <el-icon><Connection /></el-icon>
              <span>海量设备接入</span>
            </div>
            <div class="feature-item">
              <el-icon><DataLine /></el-icon>
              <span>实时数据采集</span>
            </div>
            <div class="feature-item">
              <el-icon><Cpu /></el-icon>
              <span>智能数据分析</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧登录区 -->
      <div class="form-section">
        <div class="form-header">
          <h2>欢迎登录</h2>
          <p>登录您的账户开始管理您的IoT设备</p>
        </div>

        <!-- 登录方式切换 -->
        <div class="login-tabs">
          <el-radio-group v-model="loginType">
            <el-radio-button label="password">账号密码</el-radio-button>
            <el-radio-button label="app">App扫码</el-radio-button>
          </el-radio-group>
        </div>

        <!-- 账号密码登录 -->
        <el-form v-if="loginType === 'password'" ref="formRef" :model="form" :rules="rules" class="login-form">
          <el-form-item prop="username">
            <el-input
              v-model="form.username"
              placeholder="请输入用户名/邮箱"
              size="large"
              :prefix-icon="User"
            />
          </el-form-item>
          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              size="large"
              :prefix-icon="Lock"
              show-password
            />
          </el-form-item>
          <div class="form-options">
            <el-checkbox v-model="rememberMe">记住我</el-checkbox>
            <a href="#" class="forgot-link">忘记密码？</a>
          </div>
          <el-form-item>
            <el-button type="primary" size="large" class="submit-btn" @click="handleLogin">
              登 录
            </el-button>
          </el-form-item>
        </el-form>

        <!-- App扫码登录 -->
        <div v-else class="app-login">
          <div class="qrcode-box">
            <img v-if="qrcodeUrl" :src="qrcodeUrl" alt="登录二维码" class="qrcode-img" />
            <div v-else class="qrcode-placeholder">
              <el-icon :size="80" color="#409eff"><Scissor /></el-icon>
              <p>加载中...</p>
            </div>
          </div>
          <p class="app-tip">打开{{ oemStore.getSystemName }}扫描二维码登录</p>
          <div class="refresh-btn" @click="refreshQRCode">
            <el-icon><Refresh /></el-icon>
            <span>刷新二维码</span>
          </div>
        </div>

        <div class="register-link">
          还没有账户？ <router-link to="/register">立即注册</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, Connection, DataLine, Cpu, Scissor, Refresh } from '@element-plus/icons-vue'
import { login } from '@/api/auth'
import { useUserStore, useOEMStore } from '@/store'

const router = useRouter()
const userStore = useUserStore()
const oemStore = useOEMStore()

const loginType = ref('password')
const rememberMe = ref(false)
const formRef = ref(null)
const loading = ref(false)
const qrcodeUrl = ref('')
const loginSessionId = ref('')
let ws = null
let pollTimer = null

const form = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const res = await login({ ...form, client: 'admin' })
    userStore.setToken(res.data?.token || res.token)
    if (res.data?.user) {
      userStore.setUser(res.data.user)
    }
    ElMessage.success('登录成功')
    router.push('/')
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const generateQRCode = async () => {
  try {
    const res = await fetch('/api/login/wait', { method: 'GET' })
    const data = await res.json()
    loginSessionId.value = data.id
    
    qrcodeUrl.value = `/api/qrcode/generate?content=${encodeURIComponent(data.url)}&size=200`
    
    connectWebSocket(data.id)
  } catch (error) {
    console.error('生成二维码失败:', error)
  }
}

const connectWebSocket = (id) => {
  if (ws) {
    ws.close()
  }
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  ws = new WebSocket(`${protocol}//${window.location.host}/api/login/ws?id=${id}`)
  
  let wsTimeout = setTimeout(() => {
    console.log('WebSocket连接超时，切换到轮询')
    if (ws) {
      ws.close()
      ws = null
    }
    startPolling(id)
  }, 3000)
  
  ws.onopen = () => {
    console.log('WebSocket连接成功')
  }
  
  ws.onmessage = (event) => {
    clearTimeout(wsTimeout)
    const data = JSON.parse(event.data)
    if (data.code === 0 && data.token) {
      userStore.setToken(data.token)
      ElMessage.success('登录成功')
      router.push('/')
      closeWebSocket()
    } else if (data.code === 1) {
      ElMessage.warning('二维码已过期，请刷新')
    }
  }
  
  ws.onclose = () => {
    clearTimeout(wsTimeout)
    console.log('WebSocket连接关闭')
    if (!loginSessionId.value) return
    startPolling(id)
  }
  
  ws.onerror = (error) => {
    clearTimeout(wsTimeout)
    console.error('WebSocket错误:', error)
    startPolling(id)
  }
}

const startPolling = (id) => {
  if (pollTimer) {
    clearInterval(pollTimer)
  }
  pollTimer = setInterval(async () => {
    try {
      const res = await fetch(`/api/login/wait?id=${id}`, { method: 'GET' })
      const data = await res.json()
      if (data.token) {
        userStore.setToken(data.token)
        ElMessage.success('登录成功')
        router.push('/')
        clearInterval(pollTimer)
      }
    } catch (error) {
      console.error('轮询失败:', error)
    }
  }, 2000)
}

const closeWebSocket = () => {
  if (ws) {
    ws.close()
    ws = null
  }
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

const refreshQRCode = () => {
  closeWebSocket()
  qrcodeUrl.value = ''
  generateQRCode()
}

watch(loginType, (newVal) => {
  if (newVal === 'app') {
    generateQRCode()
  } else {
    closeWebSocket()
    qrcodeUrl.value = ''
  }
})

onUnmounted(() => {
  closeWebSocket()
})
</script>

<style lang="scss" scoped>
.login-container {
  width: 100%;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
}

.login-box {
  width: 900px;
  height: 560px;
  border-radius: 20px;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.1);
  display: flex;
  overflow: hidden;
}

.brand-section {
  width: 380px;
  background: linear-gradient(135deg, #409eff 0%, #667eea 100%);
  padding: 60px 40px;
  display: flex;
  align-items: center;
}

.brand-content {
  color: #fff;
}

.brand-title {
  font-size: 36px;
  font-weight: 700;
  margin-bottom: 10px;
}

.brand-subtitle {
  font-size: 16px;
  opacity: 0.9;
  margin-bottom: 40px;
}

.feature-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 15px;
  font-size: 14px;
  opacity: 0.9;

  .el-icon {
    font-size: 24px;
  }
}

.form-section {
  flex: 1;
  padding: 60px 50px;
}

.form-header {
  margin-bottom: 30px;

  h2 {
    font-size: 28px;
    margin-bottom: 8px;
  }

  p {
    font-size: 14px;
    color: #999;
  }
}

.login-tabs {
  margin-bottom: 30px;
}

.app-login {
  text-align: center;
  padding: 20px 0;
}

.qrcode-box {
  width: 220px;
  height: 220px;
  margin: 0 auto 20px;
  background: #f5f7fa;
  border: 2px dashed #d9d9d9;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.qrcode-img {
  width: 200px;
  height: 200px;
}

.qrcode-placeholder {
  text-align: center;
  color: #999;

  p {
    margin-top: 10px;
    font-size: 14px;
  }
}

.app-tip {
  color: #999;
  font-size: 14px;
  margin-bottom: 15px;
}

.refresh-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 14px;
  color: #409eff;
  cursor: pointer;

  &:hover {
    color: #66b1ff;
  }
}

.login-form {
  :deep(.el-form-item) {
    margin-bottom: 24px;
  }
}

.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;

  .forgot-link {
    font-size: 14px;
  }
}

.submit-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
}

.register-link {
  text-align: center;
  margin-top: 30px;
  font-size: 14px;
  color: #999;

  a {
    margin-left: 5px;
  }
}
</style>
