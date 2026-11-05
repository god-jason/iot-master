<template>
  <div class="install-page">
    <div class="install-card">
      <div class="install-header">
        <h1>IoT-DMP 安装向导</h1>
        <p>首次安装 — 完成以下步骤即可启动您的智能物联管理平台</p>
      </div>

      <el-steps v-if="!finished" :active="step - 1" finish-status="success" align-center class="install-steps">
        <el-step v-for="title in stepTitles" :key="title" :title="title" />
      </el-steps>

      <!-- ========== 安装表单 ========== -->
      <div v-if="!finished" class="install-body">
        <!-- 步骤 1：数据库 -->
        <div v-show="step === 1" class="step-page">
          <div class="section-title">关系数据库（MySQL / PostgreSQL）</div>
          <el-form label-position="top">
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="数据库驱动" required>
                  <el-select v-model="form.db.driver" style="width: 100%" @change="onDriverChange">
                    <el-option label="MySQL" value="mysql" />
                    <el-option label="PostgreSQL" value="postgres" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="数据库名" required>
                  <el-input v-model="form.db.database" placeholder="iot_dmp" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="主机" required>
                  <el-input v-model="form.db.host" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="端口" required>
                  <el-input-number
                    v-model="form.db.port"
                    :min="1"
                    :max="65535"
                    controls-position="right"
                    style="width: 100%"
                  />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="用户名" required>
                  <el-input v-model="form.db.username" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="密码">
                  <el-input v-model="form.db.password" type="password" show-password />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="SSL Mode（PostgreSQL）">
                  <el-select v-model="form.db.sslmode" style="width: 100%" clearable placeholder="默认">
                    <el-option label="disable" value="disable" />
                    <el-option label="require" value="require" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="额外参数（可选）">
                  <el-input v-model="form.db.params" placeholder="charset=utf8mb4&parseTime=True&loc=Local" />
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>
          <div class="test-bar">
            <el-button size="small" :loading="dbTest.loading" @click="onTestDB">测试连接</el-button>
            <span v-if="dbTest.message" :class="['test-result', dbTest.ok ? 'pass' : 'fail']">
              {{ dbTest.message }}
            </span>
          </div>

          <div class="section-title">
            Redis 内存数据库 <span class="optional">（可选，不填则降级为内存存储）</span>
          </div>
          <el-form label-position="top">
            <el-row :gutter="16">
              <el-col :span="6">
                <el-form-item label="主机">
                  <el-input v-model="form.redis.host" />
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <el-form-item label="端口">
                  <el-input-number v-model="form.redis.port" :min="1" :max="65535" controls-position="right" style="width: 100%" />
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <el-form-item label="密码">
                  <el-input v-model="form.redis.password" type="password" show-password />
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <el-form-item label="DB 编号">
                  <el-input-number v-model="form.redis.db" :min="0" controls-position="right" style="width: 100%" />
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>
          <div class="test-bar">
            <el-button size="small" :loading="redisTest.loading" @click="onTestRedis">测试连接</el-button>
            <span v-if="redisTest.message" :class="['test-result', redisTest.ok ? 'pass' : 'fail']">
              {{ redisTest.message }}
            </span>
          </div>

          <div class="section-title">
            InfluxDB 时序数据库 <span class="optional">（可选，不填则不启用历史存储）</span>
          </div>
          <el-form label-position="top">
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="URL">
                  <el-input v-model="form.influx.url" placeholder="http://127.0.0.1:8086" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="Token">
                  <el-input v-model="form.influx.token" type="password" show-password />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="Organization">
                  <el-input v-model="form.influx.org" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="Bucket">
                  <el-input v-model="form.influx.bucket" />
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>
          <div class="test-bar">
            <el-button size="small" :loading="influxTest.loading" @click="onTestInflux">测试连接</el-button>
            <span v-if="influxTest.message" :class="['test-result', influxTest.ok ? 'pass' : 'fail']">
              {{ influxTest.message }}
            </span>
          </div>
        </div>

        <!-- 步骤 2：HTTP -->
        <div v-show="step === 2" class="step-page">
          <div class="section-title">HTTP 服务器</div>
          <el-form label-position="top">
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="监听地址" required>
                  <el-input v-model="form.web.host" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="端口" required>
                  <el-input-number v-model="form.web.port" :min="1" :max="65535" controls-position="right" style="width: 100%" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item label="启用 HTTPS（使用自定义证书）">
              <el-switch v-model="form.web.https" />
            </el-form-item>
            <el-row v-if="form.web.https" :gutter="16">
              <el-col :span="12">
                <el-form-item label="证书文件路径">
                  <el-input v-model="form.web.cert_file" placeholder="/path/to/cert.pem" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="私钥文件路径">
                  <el-input v-model="form.web.key_file" placeholder="/path/to/key.pem" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item label="启用 Let's Encrypt 自动证书">
              <el-switch v-model="form.web.lets_encrypt" />
            </el-form-item>
          </el-form>
        </div>

        <!-- 步骤 3：MQTT -->
        <div v-show="step === 3" class="step-page">
          <div class="section-title">MQTT Broker</div>
          <p class="hint">IoT-DMP 支持两种设备接入方式。</p>
          <div class="broker-modes">
            <div
              class="mode-card"
              :class="{ active: form.broker.mode === 'builtin' }"
              @click="form.broker.mode = 'builtin'"
            >
              <div class="mode-icon">📡</div>
              <div class="mode-label">内置代理</div>
              <div class="mode-desc">Master 进程内嵌 MQTT Broker，单机部署最简单</div>
            </div>
            <div
              class="mode-card"
              :class="{ active: form.broker.mode === 'external' }"
              @click="form.broker.mode = 'external'"
            >
              <div class="mode-icon">🔗</div>
              <div class="mode-label">外置代理</div>
              <div class="mode-desc">设备连接独立的 EMQX/Mosquitto 等 Broker</div>
            </div>
          </div>
          <el-form v-if="form.broker.mode === 'builtin'" label-position="top">
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="监听地址">
                  <el-input v-model="form.broker.host" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="监听端口">
                  <el-input-number v-model="form.broker.port" :min="1" :max="65535" controls-position="right" style="width: 100%" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item label="启用物模型校验（设备必须上报已注册的属性）">
              <el-switch v-model="form.broker.validator" />
            </el-form-item>
          </el-form>
          <el-alert
            v-else
            title="外置代理已独立运行，Master 进程不再启动内嵌 Broker；设备需连接到外置 Broker 的地址与端口。"
            type="info"
            :closable="false"
          />
        </div>

        <!-- 步骤 4：管理员 -->
        <div v-show="step === 4" class="step-page">
          <div class="section-title">管理员账号</div>
          <p class="hint">系统首次启动后，使用此账号登录 Admin 控制台。密码请妥善保管。</p>
          <el-form label-position="top">
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="用户名" required>
                  <el-input v-model="form.admin.username" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="昵称">
                  <el-input v-model="form.admin.nickname" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="密码" required>
                  <el-input v-model="form.admin.password" type="password" show-password placeholder="至少 6 位" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="确认密码" required>
                  <el-input v-model="form.admin.password2" type="password" show-password placeholder="再次输入密码" />
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>

          <div class="section-title">其他配置</div>
          <el-form label-position="top">
            <el-form-item label="上传文件目录（留空使用默认 ./uploads）">
              <el-input v-model="form.uploadDir" placeholder="./uploads" />
            </el-form-item>
            <el-form-item label="启用调试模式">
              <el-switch v-model="form.web.debug" />
            </el-form-item>
          </el-form>
        </div>

        <!-- 步骤 5：品牌 -->
        <div v-show="step === 5" class="step-page">
          <div class="section-title">品牌定制（OEM）</div>
          <p class="hint">设置系统展示的品牌信息，后续可在后台随时修改。</p>
          <el-form label-position="top">
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="系统名称">
                  <el-input v-model="form.oem.system_name" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="公司名称">
                  <el-input v-model="form.oem.company" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="Logo 路径（URL 或相对路径）">
                  <el-input v-model="form.oem.logo" placeholder="/logo.jpg 或 https://..." />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="备案号 ICP">
                  <el-input v-model="form.oem.icp" placeholder="沪ICP备00000000号" />
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>
        </div>

        <!-- 步骤 6：完成 -->
        <div v-show="step === 6" class="step-page">
          <div class="finish-box">
            <el-icon class="finish-icon"><CircleCheckFilled /></el-icon>
            <h2>准备就绪</h2>
            <p>点击「完成安装」，系统将写入配置文件、创建数据库表结构并初始化管理员账号。</p>
          </div>
        </div>
      </div>

      <!-- ========== 安装结果 ========== -->
      <div v-else class="install-body">
        <div class="finish-box">
          <el-icon v-if="result.ok" class="finish-icon success"><CircleCheckFilled /></el-icon>
          <el-icon v-else class="finish-icon error"><CircleCloseFilled /></el-icon>
          <h2>{{ result.ok ? '安装完成' : '安装失败' }}</h2>
          <p>{{ result.message }}</p>
          <p v-if="result.ok" class="hint">请在终端重启本程序，即可进入正常模式。</p>
          <el-alert
            v-if="result.warnings?.length"
            class="warnings"
            type="warning"
            :closable="false"
            title="以下警告需要关注"
          >
            <ul>
              <li v-for="(w, i) in result.warnings" :key="i">{{ w }}</li>
            </ul>
          </el-alert>
        </div>
      </div>

      <!-- ========== 底部操作 ========== -->
      <div v-if="!finished" class="install-footer">
        <el-button :disabled="step === 1" @click="prevStep">上一步</el-button>
        <el-button v-if="step < maxStep" type="primary" @click="nextStep">下一步</el-button>
        <el-button v-else type="success" :loading="saving" @click="finishInstall">完成安装</el-button>
      </div>
      <div v-else-if="!result.ok" class="install-footer">
        <el-button @click="backToForm">返回修改</el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { testDB, testRedis, testInflux, saveSetup } from '@/api/setup'

const stepTitles = ['数据库', 'HTTP', 'MQTT', '管理员', '品牌', '完成']
const maxStep = 6

const step = ref(1)
const saving = ref(false)
const finished = ref(false)
const result = reactive({ ok: false, message: '', warnings: [] })

const form = reactive({
  db: {
    driver: 'mysql',
    host: '127.0.0.1',
    port: 3306,
    database: 'iot_dmp',
    sslmode: '',
    username: 'root',
    password: '',
    params: ''
  },
  redis: { host: '127.0.0.1', port: 6379, password: '', db: 0 },
  influx: { url: '', token: '', org: '', bucket: '' },
  web: {
    host: '0.0.0.0',
    port: 8080,
    https: false,
    cert_file: '',
    key_file: '',
    lets_encrypt: false,
    debug: true
  },
  broker: { mode: 'builtin', host: '0.0.0.0', port: 1883, validator: true },
  admin: { username: 'admin', password: '', password2: '', nickname: '管理员' },
  uploadDir: '',
  oem: { system_name: '小白白', company: '本易物联', logo: '/logo.jpg', icp: '' }
})

// 连接测试状态（loading / ok / message）
const dbTest = reactive({ loading: false, ok: false, message: '' })
const redisTest = reactive({ loading: false, ok: false, message: '' })
const influxTest = reactive({ loading: false, ok: false, message: '' })

// 切换数据库驱动时同步默认端口
function onDriverChange(driver) {
  form.db.port = driver === 'postgres' ? 5432 : 3306
}

async function onTestDB() {
  dbTest.loading = true
  dbTest.message = ''
  try {
    const r = await testDB({ ...form.db })
    dbTest.ok = !!r.ok
    dbTest.message = r.message || (r.ok ? '连接成功' : '连接失败')
  } catch (e) {
    dbTest.ok = false
    dbTest.message = e.message || '连接失败'
  } finally {
    dbTest.loading = false
  }
}

async function onTestRedis() {
  redisTest.loading = true
  redisTest.message = ''
  try {
    const r = await testRedis({ host: form.redis.host, port: form.redis.port })
    redisTest.ok = !!r.ok
    redisTest.message = r.message || (r.ok ? '连接成功' : '连接失败')
  } catch (e) {
    redisTest.ok = false
    redisTest.message = e.message || '连接失败'
  } finally {
    redisTest.loading = false
  }
}

async function onTestInflux() {
  if (!form.influx.url) {
    influxTest.ok = false
    influxTest.message = ''
    ElMessage.warning('请先填写 InfluxDB URL')
    return
  }
  influxTest.loading = true
  influxTest.message = ''
  try {
    const r = await testInflux({ url: form.influx.url })
    influxTest.ok = !!r.ok
    influxTest.message = r.message || (r.ok ? '连接成功' : '连接失败')
  } catch (e) {
    influxTest.ok = false
    influxTest.message = e.message || '连接失败'
  } finally {
    influxTest.loading = false
  }
}

// 下一步前做必要校验（与后端一致：库用户名/库名必填，管理员密码至少 6 位）
function nextStep() {
  if (step.value === 1) {
    if (!form.db.username || !form.db.database) {
      ElMessage.warning('请填写数据库用户名和数据库名')
      return
    }
    if (!dbTest.ok) {
      ElMessage.warning('尚未测试数据库连接，建议先测试')
    }
  }
  if (step.value === 4) {
    if (!form.admin.password) {
      ElMessage.warning('请设置管理员密码')
      return
    }
    if (form.admin.password !== form.admin.password2) {
      ElMessage.warning('两次输入的密码不一致')
      return
    }
    if (form.admin.password.length < 6) {
      ElMessage.warning('密码至少 6 位')
      return
    }
  }
  if (step.value < maxStep) step.value++
}

function prevStep() {
  if (step.value > 1) step.value--
}

async function finishInstall() {
  saving.value = true
  try {
    const payload = {
      db: { ...form.db },
      redis: { ...form.redis },
      influx: { ...form.influx },
      web: { ...form.web, domains: [] },
      broker: { ...form.broker },
      admin: {
        username: form.admin.username,
        password: form.admin.password,
        nickname: form.admin.nickname
      },
      upload_dir: form.uploadDir,
      oem: { ...form.oem }
    }
    const r = await saveSetup(payload)
    result.ok = !!r.ok
    result.message = r.message || ''
    result.warnings = r.warnings || []
    finished.value = true
  } catch (e) {
    result.ok = false
    result.message = e.message || '安装失败'
    result.warnings = []
    finished.value = true
  } finally {
    saving.value = false
  }
}

// 安装失败后返回表单继续修改
function backToForm() {
  finished.value = false
  step.value = maxStep
}
</script>

<style scoped>
.install-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.install-card {
  width: 100%;
  max-width: 760px;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
  overflow: hidden;
}

.install-header {
  padding: 28px 36px;
  color: #fff;
  background: linear-gradient(135deg, #409eff, #667eea);
}

.install-header h1 {
  margin: 0 0 6px;
  font-size: 24px;
  font-weight: 600;
}

.install-header p {
  margin: 0;
  font-size: 14px;
  opacity: 0.85;
}

.install-steps {
  padding: 20px 24px;
  border-bottom: 1px solid #f0f0f0;
  background: #fafbfc;
}

.install-body {
  padding: 28px 36px;
  min-height: 320px;
  max-height: 62vh;
  overflow-y: auto;
}

.section-title {
  margin: 20px 0 14px;
  padding-bottom: 8px;
  border-bottom: 1px solid #f0f0f0;
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}

.section-title:first-child {
  margin-top: 0;
}

.optional {
  margin-left: 4px;
  font-size: 12px;
  font-weight: 400;
  color: #909399;
}

.hint {
  margin: 0 0 16px;
  font-size: 13px;
  color: #909399;
}

.test-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: -4px 0 4px;
}

.test-result {
  font-size: 12px;
}

.test-result.pass {
  color: #67c23a;
}

.test-result.fail {
  color: #f56c6c;
}

.broker-modes {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.mode-card {
  flex: 1;
  padding: 14px;
  border: 2px solid #e8eaed;
  border-radius: 8px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
}

.mode-card:hover {
  border-color: #c0c4cc;
}

.mode-card.active {
  border-color: #409eff;
  background: #ecf5ff;
}

.mode-icon {
  margin-bottom: 6px;
  font-size: 24px;
}

.mode-label {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.mode-desc {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}

.finish-box {
  padding: 40px 20px;
  text-align: center;
}

.finish-icon {
  margin-bottom: 16px;
  font-size: 64px;
}

.finish-icon.success {
  color: #67c23a;
}

.finish-icon.error {
  color: #f56c6c;
}

.finish-box h2 {
  margin: 0 0 8px;
  font-size: 22px;
  color: #303133;
}

.finish-box p {
  margin: 0 0 24px;
  font-size: 14px;
  line-height: 1.6;
  color: #909399;
}

.warnings {
  text-align: left;
}

.warnings ul {
  margin: 8px 0 0;
  padding-left: 18px;
}

.warnings li {
  margin-bottom: 4px;
  font-size: 13px;
}

.install-footer {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 20px 36px;
  border-top: 1px solid #f0f0f0;
  background: #fafbfc;
}
</style>
