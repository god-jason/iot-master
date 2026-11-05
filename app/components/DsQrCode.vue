<template>
  <view class="ds-qrcode" @click="handleClick">
    <canvas 
      :canvas-id="canvasId" 
      :id="canvasId"
      class="ds-qrcode-canvas"
      :style="{ width: size + 'rpx', height: size + 'rpx' }"
    ></canvas>
    <view v-if="loading" class="ds-qrcode-loading">
      <DsLoading type="dots" text="" />
    </view>
    <view v-if="text" class="ds-qrcode-text">{{ text }}</view>
  </view>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import QRCode from 'qrcode'
import DsLoading from './DsLoading.vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  value: {
    type: String,
    default: ''
  },
  size: {
    type: Number,
    default: 200
  },
  text: {
    type: String,
    default: ''
  },
  level: {
    type: String,
    default: 'H',
    validator: (val) => ['L', 'M', 'Q', 'H'].includes(val)
  }
})

const canvasId = ref('qrcode-' + Date.now())
const loading = ref(true)

const emit = defineEmits(['click'])

const handleClick = (event) => {
  emit('click', event)
}

const drawQRCode = () => {
  if (!props.value) {
    loading.value = false
    return
  }

  loading.value = true
  
  const query = uni.createSelectorQuery()
  query.select(`#${canvasId.value}`)
    .fields({ node: true, size: true })
    .exec((res) => {
      if (!res[0]) {
        loading.value = false
        return
      }

      const canvas = res[0].node
      const ctx = canvas.getContext('2d')
      const dpr = uni.getSystemInfoSync().pixelRatio
      const size = props.size / 2

      canvas.width = size * dpr
      canvas.height = size * dpr
      ctx.scale(dpr, dpr)

      QRCode.toCanvas(canvas, props.value, {
        width: size,
        margin: 2,
        color: {
          dark: '#000000',
          light: '#ffffff'
        },
        errorCorrectionLevel: props.level
      }).then(() => {
        loading.value = false
      }).catch(() => {
        loading.value = false
      })
    })
}

onMounted(() => {
  setTimeout(drawQRCode, 100)
})

watch(() => props.value, () => {
  drawQRCode()
})
</script>

<style lang="scss">
@import './variables.scss';

.ds-qrcode {
  display: flex;
  flex-direction: column;
  align-items: center;

  .ds-qrcode-canvas {
    background: #fff;
    border-radius: $radius-base;
    padding: $spacing-sm;
  }

  .ds-qrcode-loading {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
  }

  .ds-qrcode-text {
    margin-top: $spacing-sm;
    font-size: $font-size-xs;
    color: $text-placeholder;
    text-align: center;
    max-width: 200rpx;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>