<template>
  <el-dialog
    v-model="visible"
    :title="title"
    width="720px"
    :close-on-click-modal="false"
    append-to-body
    destroy-on-close
    @closed="handleClosed"
  >
    <div class="map-picker">
      <div class="search-bar">
        <el-input
          v-model="keyword"
          placeholder="输入地址搜索位置"
          clearable
          @input="handleKeywordInput"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
          <template #append v-if="searchLoading">
            <el-icon class="is-loading"><Loading /></el-icon>
          </template>
        </el-input>
        <div v-if="suggestions.length && keyword" class="suggestions-dropdown">
          <div
            v-for="(item, index) in suggestions"
            :key="index"
            class="suggestion-item"
            :class="{ disabled: item.disabled }"
            @click="!item.disabled && selectSuggestion(item)"
          >
            <div class="suggestion-name">{{ item.name }}</div>
            <div v-if="item.district || item.address" class="suggestion-address">
              {{ item.district }}{{ item.address && item.district ? ' - ' : '' }}{{ item.address }}
            </div>
          </div>
        </div>
      </div>

      <div ref="mapRef" class="map-container"></div>

      <div class="selected-info">
        <div v-if="selectedPosition" class="info-item">
          <span class="label">经度：</span>
          <span class="value">{{ selectedPosition.longitude.toFixed(6) }}</span>
        </div>
        <div v-if="selectedPosition" class="info-item">
          <span class="label">纬度：</span>
          <span class="value">{{ selectedPosition.latitude.toFixed(6) }}</span>
        </div>
        <div v-if="selectedPosition?.geoHash" class="info-item">
          <span class="label">GeoHash：</span>
          <span class="value mono">{{ selectedPosition.geoHash }}</span>
        </div>
        <div v-if="selectedPosition?.location" class="info-item full-width">
          <span class="label">地址：</span>
          <span class="value">{{ selectedPosition.location }}</span>
        </div>
        <div v-if="!selectedPosition" class="info-item placeholder">
          请在地图上点击选择位置，或使用搜索框搜索地址
        </div>
      </div>
    </div>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :disabled="!selectedPosition" @click="handleConfirm">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'
import AMapLoader from '@amap/amap-jsapi-loader'
import { Search, Loading } from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: Boolean,
  title: { type: String, default: '位置选择' },
  initialLongitude: { type: Number, default: 0 },
  initialLatitude: { type: Number, default: 0 },
})

const emit = defineEmits(['update:modelValue', 'confirm'])

const visible = ref(false)
const keyword = ref('')
const selectedPosition = ref(null)
const searchLoading = ref(false)
const suggestions = ref([])

const mapRef = ref()
let map = null
let AMap = null
let marker = null
let autocomplete = null
let geocoder = null
let searchTimer = null

const JSAPI_KEY = 'eb6a831c04b6dfedda190d6254febb58'
const SECURITY_CODE = '55de9923dc16159e4750b7c743117e0d'

watch(() => props.modelValue, (val) => {
  visible.value = val || false
  if (val) {
    nextTick(() => {
      initMap()
      if (props.initialLongitude && props.initialLatitude) {
        selectedPosition.value = {
          longitude: props.initialLongitude,
          latitude: props.initialLatitude,
        }
      }
    })
  }
})

watch(visible, (val) => { emit('update:modelValue', val) })

async function initMap() {
  if (map) { map.destroy(); map = null }

  try {
    window._AMapSecurityConfig = { securityJsCode: SECURITY_CODE }
    if (!AMap) {
      AMap = await AMapLoader.load({
        key: JSAPI_KEY,
        version: '2.0',
        plugins: ['AMap.AutoComplete', 'AMap.Geocoder'],
      })
    }

    const hasInit = props.initialLongitude && props.initialLatitude
    const center = hasInit ? [props.initialLongitude, props.initialLatitude] : [116.397428, 39.90923]

    map = new AMap.Map(mapRef.value, {
      center,
      zoom: hasInit ? 15 : 11,
      viewMode: '2D',
    })

    if (!autocomplete && AMap.AutoComplete) {
      autocomplete = new AMap.AutoComplete({ city: '全国' })
    }
    if (!geocoder && AMap.Geocoder) {
      geocoder = new AMap.Geocoder({})
    }

    map.on('click', async (e) => {
      await updatePosition(e.lnglat.getLng(), e.lnglat.getLat())
    })

    if (hasInit) {
      setMarker(props.initialLongitude, props.initialLatitude)
    }
  } catch (e) {
    console.error('地图加载失败:', e)
  }
}

function setMarker(lng, lat) {
  if (!map || !AMap) return
  if (marker) marker.setMap(null)
  marker = new AMap.Marker({ position: [lng, lat], anchor: 'bottom-center' })
  map.add(marker)
  map.setCenter([lng, lat])
}

// 简易 GeoHash 编码（12 位精度，纯 JS 实现，无需 geohashing 库依赖）
function encodeGeoHash(lat, lng, precision = 12) {
  const BASE32 = '0123456789bcdefghjkmnpqrstuvwxyz'
  const latRange = [-90, 90]
  const lngRange = [-180, 180]
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

async function updatePosition(lng, lat) {
  setMarker(lng, lat)
  const geoHash = encodeGeoHash(lat, lng)
  selectedPosition.value = { longitude: lng, latitude: lat, location: '', geoHash }
  try {
    const addr = await reverseGeocode(lng, lat)
    if (addr) selectedPosition.value = { ...selectedPosition.value, location: addr }
  } catch {}
}

function reverseGeocode(lng, lat) {
  return new Promise((resolve) => {
    if (!geocoder) { resolve(''); return }
    geocoder.getAddress([lng, lat], (status, result) => {
      if (status === 'complete' && result && result.regeocode) {
        resolve(result.regeocode.formattedAddress || '')
      } else { resolve('') }
    })
  })
}

function handleKeywordInput() {
  if (searchTimer) clearTimeout(searchTimer)
  if (!keyword.value) { suggestions.value = []; return }
  searchTimer = setTimeout(() => fetchSuggestions(keyword.value), 300)
}

function fetchSuggestions(q) {
  if (!q) { suggestions.value = []; return }
  if (!autocomplete) {
    suggestions.value = [{ name: '搜索服务未就绪', disabled: true }]
    return
  }
  searchLoading.value = true
  try {
    autocomplete.search(q, (status, result) => {
      searchLoading.value = false
      try {
        if (status === 'complete' && result?.tips?.length) {
          const tips = result.tips.filter(t => t.location).map(t => {
            let lng, lat
            const loc = t.location
            if (typeof loc === 'string') {
              const p = loc.split(',')
              lng = parseFloat(p[0]); lat = parseFloat(p[1])
            } else if (loc) {
              lng = loc.lng != null ? Number(loc.lng) : (typeof loc.getLng === 'function' ? loc.getLng() : 0)
              lat = loc.lat != null ? Number(loc.lat) : (typeof loc.getLat === 'function' ? loc.getLat() : 0)
            }
            if (!lng || !lat) return null
            return { name: t.name, address: t.address || '', district: t.district || '', location: { lng, lat } }
          }).filter(Boolean)
          suggestions.value = tips.length ? tips : [{ name: '未找到相关地点', disabled: true }]
        } else {
          suggestions.value = [{ name: '未找到相关地点', disabled: true }]
        }
      } catch { suggestions.value = [{ name: '搜索解析失败', disabled: true }] }
    })
  } catch (e) {
    searchLoading.value = false
    suggestions.value = [{ name: '搜索服务异常', disabled: true }]
  }
}

function selectSuggestion(item) {
  if (!item.location) return
  keyword.value = item.name
  suggestions.value = []
  updatePosition(item.location.lng, item.location.lat)
}

function handleConfirm() {
  if (selectedPosition.value) {
    emit('confirm', { ...selectedPosition.value })
    visible.value = false
  }
}

function handleClosed() {
  emit('update:modelValue', false)
  keyword.value = ''
  suggestions.value = []
  selectedPosition.value = null
  searchTimer = null
  if (map) { map.destroy(); map = null }
  marker = null; autocomplete = null; geocoder = null
}
</script>

<style scoped lang="scss">
.map-picker {
  position: relative;

  .search-bar {
    display: flex; gap: 8px; margin-bottom: 12px; position: relative;
    .el-input { flex: 1; }
  }

  .suggestions-dropdown {
    position: absolute; top: 100%; left: 0; right: 0; margin-top: 4px;
    background: #fff; border: 1px solid #e4e7ed; border-radius: 4px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1); max-height: 260px; overflow-y: auto; z-index: 2000;
    .suggestion-item {
      padding: 10px 12px; cursor: pointer; border-bottom: 1px solid #f0f0f0; transition: background 0.2s;
      &:last-child { border-bottom: none }
      &:hover:not(.disabled) { background: #f5f7fa }
      &.disabled { cursor: not-allowed; color: #c0c4cc }
      .suggestion-name { font-size: 14px; color: #303133; font-weight: 500 }
      .suggestion-address { font-size: 12px; color: #909399; margin-top: 2px }
    }
  }

  .map-container { width: 100%; height: 380px; border-radius: 4px; border: 1px solid #e4e7ed }

  .selected-info {
    margin-top: 12px; padding: 10px 12px; background: #f5f7fa; border-radius: 4px;
    display: flex; flex-wrap: wrap; gap: 8px 16px;
    .info-item { font-size: 13px; color: #606266;
      .label { color: #909399 }
      .value { color: #303133; font-weight: 500 }
      .value.mono { font-family: monospace; font-size: 12px }
      &.full-width { width: 100% }
      &.placeholder { color: #c0c4cc; font-style: italic }
    }
  }
}
</style>
