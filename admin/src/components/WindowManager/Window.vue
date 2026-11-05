<template>
  <div
    class="desktop-window"
    :class="{
      'is-minimized': windowData.minimized,
      'is-maximized': windowData.maximized,
      'is-active': windowData.zIndex === maxZIndex,
      'is-closing': isClosing,
      'is-minimizing': isMinimizing,
      'is-restoring': isRestoring,
      'is-opening': isOpening && !isRestoring
    }"
    :style="windowStyle"
    @mousedown="handleMouseDown"
  >
    <div class="window-header" @mousedown="startDrag">
      <div class="window-icon">
        <img v-if="windowData.icon" :src="getIconUrl(windowData.icon)" alt="" />
        <el-icon v-else><Monitor /></el-icon>
      </div>
      <span class="window-title">{{ windowData.title }}</span>
      <div class="window-controls">
        <button class="control-btn minimize" @click.stop="minimize">
          <el-icon><Minus /></el-icon>
        </button>
        <button class="control-btn maximize" @click.stop="toggleMaximize">
          <el-icon><ZoomIn v-if="!windowData.maximized" /><ZoomOut v-else /></el-icon>
        </button>
        <button class="control-btn close" @click.stop="close">
          <el-icon><Close /></el-icon>
        </button>
      </div>
    </div>
    <div class="window-content">
      <iframe :src="windowData.url" frameborder="0" class="window-iframe"></iframe>
    </div>
    <div class="resize-handle" @mousedown="startResize"></div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { Monitor, Minus, ZoomIn, ZoomOut, Close } from '@element-plus/icons-vue'

const props = defineProps({
  windowData: {
    type: Object,
    required: true
  },
  maxZIndex: {
    type: Number,
    default: 100
  }
})

const emit = defineEmits(['activate', 'minimize', 'maximize', 'close', 'resize'])

const isDragging = ref(false)
const isResizing = ref(false)
const isClosing = ref(false)
const isMinimizing = ref(false)
const isRestoring = computed(() => props.windowData.restoring || false)
const isOpening = ref(true)
const hasTransition = ref(true)
const startX = ref(0)
const startY = ref(0)
const startWidth = ref(0)
const startHeight = ref(0)
const startLeft = ref(0)
const startTop = ref(0)

const baseUrl = import.meta.env.VITE_API_BASE_URL || ''

const getIconUrl = (url) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return baseUrl + url
}

const windowStyle = computed(() => {
  if (props.windowData.minimized && !isRestoring.value) {
    return {
      display: 'none'
    }
  }
  const baseStyle = {
    zIndex: props.windowData.zIndex,
    transition: hasTransition.value ? 'left 0.3s ease-out, top 0.3s ease-out, width 0.3s ease-out, height 0.3s ease-out' : 'none'
  }
  if (props.windowData.maximized) {
    return {
      ...baseStyle,
      left: '0',
      top: '0',
      width: '100%',
      height: '100%',
      borderRadius: '0'
    }
  }
  return {
    ...baseStyle,
    left: `${props.windowData.x}px`,
    top: `${props.windowData.y}px`,
    width: `${props.windowData.width}px`,
    height: `${props.windowData.height}px`,
    borderRadius: '12px'
  }
})

const handleMouseDown = () => {
  emit('activate', props.windowData.id)
}

const startDrag = (e) => {
  if (props.windowData.maximized) return
  isDragging.value = true
  hasTransition.value = false
  startX.value = e.clientX
  startY.value = e.clientY
  startLeft.value = props.windowData.x
  startTop.value = props.windowData.y
  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', stopDrag)
}

const onDrag = (e) => {
  if (!isDragging.value) return
  const deltaX = e.clientX - startX.value
  const deltaY = e.clientY - startY.value
  emit('resize', {
    id: props.windowData.id,
    x: Math.max(0, startLeft.value + deltaX),
    y: Math.max(0, startTop.value + deltaY)
  })
}

const stopDrag = () => {
  isDragging.value = false
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
  setTimeout(() => {
    hasTransition.value = true
  }, 50)
}

const startResize = (e) => {
  if (props.windowData.maximized) return
  isResizing.value = true
  hasTransition.value = false
  startX.value = e.clientX
  startY.value = e.clientY
  startWidth.value = props.windowData.width
  startHeight.value = props.windowData.height
  document.addEventListener('mousemove', onResize)
  document.addEventListener('mouseup', stopResize)
}

const onResize = (e) => {
  if (!isResizing.value) return
  const deltaX = e.clientX - startX.value
  const deltaY = e.clientY - startY.value
  emit('resize', {
    id: props.windowData.id,
    width: Math.max(400, startWidth.value + deltaX),
    height: Math.max(300, startHeight.value + deltaY)
  })
}

const stopResize = () => {
  isResizing.value = false
  document.removeEventListener('mousemove', onResize)
  document.removeEventListener('mouseup', stopResize)
  setTimeout(() => {
    hasTransition.value = true
  }, 50)
}

const minimize = () => {
  isMinimizing.value = true
  setTimeout(() => {
    isMinimizing.value = false
    emit('minimize', props.windowData.id)
  }, 300)
}

const toggleMaximize = () => {
  emit('maximize', props.windowData.id)
}

const close = () => {
  isClosing.value = true
  setTimeout(() => {
    emit('close', props.windowData.id)
  }, 300)
}

watch(isOpening, (val) => {
  if (val) {
    setTimeout(() => {
      isOpening.value = false
    }, 300)
  }
})

watch(() => props.windowData.restoring, (val) => {
  if (val) {
    isOpening.value = false
  }
})

onUnmounted(() => {
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
  document.removeEventListener('mousemove', onResize)
  document.removeEventListener('mouseup', stopResize)
})
</script>

<style lang="scss" scoped>
.desktop-window {
  position: absolute;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
  overflow: hidden;
  transition: box-shadow 0.3s, opacity 0.3s, transform 0.3s;
  border: 1px solid rgba(0, 0, 0, 0.1);

  &.is-opening {
    animation: windowOpen 0.3s ease-out;
  }

  &.is-active {
    box-shadow: 0 10px 50px rgba(0, 217, 255, 0.3);
    border: 1px solid rgba(0, 217, 255, 0.4);
  }

  &.is-closing {
    animation: windowClose 0.3s ease-in forwards;
  }

  &.is-minimizing {
    animation: windowMinimize 0.3s ease-in forwards;
  }

  &.is-restoring {
    animation: windowRestore 0.3s ease-out forwards;
  }

  .window-header {
    height: 52px;
    background: linear-gradient(90deg, #1a1a2e 0%, #16213e 100%);
    display: flex;
    align-items: center;
    padding: 0 16px;
    cursor: move;

    .window-icon {
      width: 28px;
      height: 28px;
      margin-right: 10px;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #00d9ff;

      img {
        width: 100%;
        height: 100%;
        border-radius: 6px;
      }
    }

    .window-title {
      flex: 1;
      color: #fff;
      font-size: 16px;
      font-weight: 600;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .window-controls {
      display: flex;
      gap: 8px;

      .control-btn {
        width: 36px;
        height: 30px;
        display: flex;
        align-items: center;
        justify-content: center;
        border: none;
        border-radius: 6px;
        cursor: pointer;
        transition: all 0.2s;

        &.minimize {
          background: rgba(255, 255, 255, 0.1);
          color: rgba(255, 255, 255, 0.8);

          &:hover {
            background: rgba(255, 255, 255, 0.2);
            color: #fff;
          }
        }

        &.maximize {
          background: rgba(255, 255, 255, 0.1);
          color: rgba(255, 255, 255, 0.8);

          &:hover {
            background: rgba(255, 255, 255, 0.2);
            color: #fff;
          }
        }

        &.close {
          background: rgba(231, 76, 60, 0.8);
          color: #fff;

          &:hover {
            background: #e74c3c;
          }
        }

        .el-icon {
          font-size: 16px;
        }
      }
    }
  }

  .window-content {
    height: calc(100% - 52px);
    overflow: hidden;

    .window-iframe {
      width: 100%;
      height: 100%;
      border: none;
    }
  }

  .resize-handle {
    position: absolute;
    right: 0;
    bottom: 0;
    width: 12px;
    height: 12px;
    cursor: nwse-resize;
    background: linear-gradient(135deg, transparent 50%, rgba(0, 217, 255, 0.3) 50%);

    &:hover {
      background: linear-gradient(135deg, transparent 50%, rgba(0, 217, 255, 0.5) 50%);
    }
  }
}

@keyframes windowOpen {
  from {
    opacity: 0;
    transform: scale(0.6);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

@keyframes windowClose {
  from {
    opacity: 1;
    transform: scale(1);
  }
  to {
    opacity: 0;
    transform: scale(0.6);
  }
}

@keyframes windowMinimize {
  from {
    opacity: 1;
    transform: scale(1);
  }
  to {
    opacity: 0;
    transform: scale(0.3) translateY(100%);
  }
}

@keyframes windowRestore {
  from {
    opacity: 0;
    transform: scale(0.4) translateY(200%);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}
</style>