<template>
  <el-drawer v-model="localVisible" :title="isEdit ? '编辑设备' : '添加设备'" size="640px">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="设备名称" prop="name">
            <el-input v-model="form.name" placeholder="请输入设备名称" maxlength="64" show-word-limit />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="设备 SN" prop="sn">
            <el-input v-model="form.sn" placeholder="设备序列号 / MAC 地址" maxlength="128" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="物模型" prop="model_id">
            <el-select v-model="form.model_id" placeholder="请选择物模型" filterable clearable style="width: 100%">
              <el-option v-for="m in modelList" :key="m.id" :label="m.name" :value="m.id" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="分组" prop="group_id">
            <el-tree-select
              v-model="form.group_id"
              :data="groupTree"
              :props="{ label: 'name', value: 'id', children: 'children' }"
              placeholder="请选择分组" filterable clearable check-strictly style="width: 100%"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <!-- Secret（一机一密）— 创建时可改，编辑时只读 + 重置 -->
      <el-form-item label="设备密钥">
        <div class="secret-row">
          <el-input v-model="form.secret" placeholder="自动生成，可手动指定" :readonly="isEdit" show-password />
          <el-button v-if="!isEdit" @click="form.secret = newSecret()">随机生成</el-button>
          <el-button v-else type="warning" @click="resetSecret">重置密钥</el-button>
          <el-button @click="copySecret">复制</el-button>
        </div>
        <div class="form-hint">MQTT 连接密码 / API 密钥（一机一密）</div>
      </el-form-item>

      <el-form-item label="接入方式">
        <el-radio-group v-model="form.link">
          <el-radio-button label="mqtt">MQTT 直连</el-radio-button>
          <el-radio-button label="gateway">网关接入</el-radio-button>
          <el-radio-button label="http">HTTP API</el-radio-button>
          <el-radio-button label="coap">CoAP</el-radio-button>
          <el-radio-button label="modbus">Modbus</el-radio-button>
          <el-radio-button label="other">其他</el-radio-button>
        </el-radio-group>
      </el-form-item>

      <!-- GPS 位置 -->
      <el-divider content-position="left">地理位置</el-divider>

      <el-form-item label="地址">
        <div class="location-row">
          <el-input v-model="form.address" placeholder="详细地址" />
          <el-button type="primary" :icon="MapLocation" @click="mapVisible = true">地图选点</el-button>
        </div>
      </el-form-item>

      <el-row :gutter="16">
        <el-col :span="8">
          <el-form-item label="经度">
            <el-input-number v-model="form.lng" :precision="6" :step="0.001" :min="-180" :max="180" controls-position="right" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="纬度">
            <el-input-number v-model="form.lat" :precision="6" :step="0.001" :min="-90" :max="90" controls-position="right" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="GeoHash">
            <el-input v-model="form.geo_hash" placeholder="自动计算" readonly />
          </el-form-item>
        </el-col>
      </el-row>
    </el-form>

    <template #footer>
      <div style="display: flex; justify-content: flex-end; gap: 10px;">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ isEdit ? '保存' : '创建设备' }}</el-button>
      </div>
    </template>

    <!-- 地图选点弹窗 -->
    <MapPicker
      v-model="mapVisible"
      title="选择设备位置"
      :initial-longitude="form.lng"
      :initial-latitude="form.lat"
      @confirm="onMapConfirm"
    />
  </el-drawer>
</template>

<script setup>
import { ref, reactive, watch, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MapLocation } from '@element-plus/icons-vue'
import MapPicker from '@/components/MapPicker.vue'
import { searchGroups } from '@/api/group'
import { searchModels } from '@/api/model'

const props = defineProps({
  visible: Boolean,
  modelValue: { type: Object, default: () => ({}) },
})
const emit = defineEmits(['update:visible', 'submit'])

const formRef = ref(null)
const submitting = ref(false)
const localVisible = ref(false)
const mapVisible = ref(false)
const groupTree = ref([])
const modelList = ref([])

const isEdit = computed(() => !!props.modelValue.id)

const form = reactive({
  name: '', sn: '', secret: '',
  model_id: '', group_id: null, link: 'mqtt',
  address: '', lat: 0, lng: 0, geo_hash: '',
})

const rules = {
  name: [{ required: true, message: '请输入设备名称', trigger: 'blur' }],
}

// GeoHash 编码（和 MapPicker 里保持一致）
function encodeGeoHash(lat, lng, precision = 12) {
  if (!lat || !lng) return ''
  const BASE32 = '0123456789bcdefghjkmnpqrstuvwxyz'
  const latRange = [-90, 90], lngRange = [-180, 180]
  let even = true, bit = 0, ch = 0, hash = ''
  while (hash.length < precision) {
    if (even) {
      const mid = (lngRange[0] + lngRange[1]) / 2
      if (lng >= mid) { ch |= (1 << (4 - bit)); lngRange[0] = mid } else { lngRange[1] = mid }
    } else {
      const mid = (latRange[0] + latRange[1]) / 2
      if (lat >= mid) { ch |= (1 << (4 - bit)); latRange[0] = mid } else { latRange[1] = mid }
    }
    even = !even
    if (bit === 4) { hash += BASE32[ch]; bit = 0; ch = 0 } else { bit++ }
  }
  return hash
}

function newSecret() {
  // 32 字符随机 hex
  const bytes = new Uint8Array(16)
  crypto.getRandomValues(bytes)
  return Array.from(bytes, b => b.toString(16).padStart(2, '0')).join('')
}

function resetSecret() {
  ElMessageBox.confirm('确定重置设备密钥？重置后需要更新设备端的连接密码。', '提示', { type: 'warning' })
    .then(() => { form.secret = newSecret(); ElMessage.success('密钥已重置，记得保存') })
    .catch(() => {})
}

function copySecret() {
  if (!form.secret) return
  navigator.clipboard.writeText(form.secret).then(() => ElMessage.success('密钥已复制'))
}

function onMapConfirm(pos) {
  form.lng = Number(pos.longitude.toFixed(6))
  form.lat = Number(pos.latitude.toFixed(6))
  form.geo_hash = pos.geoHash || encodeGeoHash(form.lat, form.lng)
  if (pos.location) form.address = pos.location
}

// 经纬度变化时自动更新 GeoHash
watch(() => [form.lat, form.lng], ([lat, lng]) => {
  if (lat && lng) form.geo_hash = encodeGeoHash(lat, lng)
})

// ============ 初始化 ============
function buildGroupTree(list) {
  const map = {}, roots = []
  list.forEach(item => { map[item.id] = { ...item, children: [] } })
  list.forEach(item => {
    const node = map[item.id]
    if (item.parent_id && map[item.parent_id]) map[item.parent_id].children.push(node)
    else roots.push(node)
  })
  return roots
}

async function loadOptions() {
  try {
    const [gRes, mRes] = await Promise.all([
      searchGroups({ page: 1, page_size: 1000 }),
      searchModels({ page: 1, page_size: 1000 }),
    ])
    groupTree.value = buildGroupTree(gRes.data || [])
    modelList.value = mRes.data || []
  } catch (e) { console.error(e) }
}

watch(() => props.visible, (val) => {
  localVisible.value = val
  if (val) {
    // 回填或初始化
    const v = props.modelValue
    Object.assign(form, {
      name: v.name || '',
      sn: v.sn || '',
      secret: v.secret || (isEdit.value ? '' : newSecret()),
      model_id: v.model_id || '',
      group_id: v.group_id || null,
      link: v.link || 'mqtt',
      address: v.address || '',
      lat: v.lat || 0,
      lng: v.lng || 0,
      geo_hash: v.geo_hash || '',
    })
    loadOptions()
  }
})

watch(localVisible, (val) => { if (!val) emit('update:visible', false) })

const handleClose = () => { localVisible.value = false }

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const payload = { ...form }
    if (!isEdit.value) {
      // 创建：前端不发 id，后端自动生成
      delete payload.id
    }
    await emit('submit', payload)
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    emit('update:visible', false)
  } catch (e) {
    // 拦截器已提示
  } finally {
    submitting.value = false
  }
}
</script>

<style lang="scss" scoped>
.secret-row {
  display: flex;
  gap: 8px;
  width: 100%;
  :deep(.el-input) { flex: 1 }
}
.form-hint {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}
.location-row {
  display: flex;
  gap: 8px;
  width: 100%;
  :deep(.el-input) { flex: 1 }
}
</style>
