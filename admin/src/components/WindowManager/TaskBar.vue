<template>
  <div 
    v-if="windows.length > 0"
    class="task-bar" 
    :class="{ 'is-visible': isVisible }"
    @mouseenter="show"
    @mouseleave="hide"
  >
    <div class="task-bar-inner">
      <div class="task-items">
        <div
          v-for="window in windows"
          :key="window.id"
          class="task-item"
          :class="{ 'is-active': window.zIndex === maxZIndex && !window.minimized }"
          @click="handleClick(window.id)"
          :title="window.title"
        >
          <div class="task-icon">
            <img v-if="window.icon" :src="getIconUrl(window.icon)" alt="" />
            <el-icon v-else><Monitor /></el-icon>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Monitor } from '@element-plus/icons-vue'

const props = defineProps({
  windows: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['restore', 'activate'])

const isVisible = ref(false)
let hideTimeout = null

const baseUrl = import.meta.env.VITE_API_BASE_URL || ''

const getIconUrl = (url) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return baseUrl + url
}

const maxZIndex = computed(() => {
  if (props.windows.length === 0) return 100
  return Math.max(...props.windows.map(w => w.zIndex))
})

const show = () => {
  isVisible.value = true
  if (hideTimeout) {
    clearTimeout(hideTimeout)
  }
}

const hide = () => {
  hideTimeout = setTimeout(() => {
    isVisible.value = false
  }, 500)
}

const handleClick = (id) => {
  const window = props.windows.find(w => w.id === id)
  if (window) {
    if (window.minimized) {
      emit('restore', id)
    } else {
      emit('activate', id)
    }
  }
}

const handleMouseMove = (e) => {
  if (window.innerHeight - e.clientY < 50) {
    show()
  }
}

onMounted(() => {
  document.addEventListener('mousemove', handleMouseMove)
})

onUnmounted(() => {
  document.removeEventListener('mousemove', handleMouseMove)
  if (hideTimeout) {
    clearTimeout(hideTimeout)
  }
})
</script>

<style lang="scss" scoped>
.task-bar {
  position: fixed;
  bottom: -70px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(26, 26, 46, 0.85);
  backdrop-filter: blur(10px);
  border-radius: 16px 16px 0 0;
  padding: 10px 20px;
  transition: bottom 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 10000;
  min-width: 100px;

  &.is-visible {
    bottom: 0;
  }

  .task-bar-inner {
    display: flex;
    justify-content: center;
  }

  .task-items {
    display: flex;
    gap: 12px;
  }

  .task-item {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 52px;
    height: 52px;
    border-radius: 10px;
    cursor: pointer;
    transition: all 0.3s;
    background: rgba(255, 255, 255, 0.05);

    &:hover {
      background: rgba(255, 255, 255, 0.15);
      transform: translateY(-3px);
    }

    &.is-active {
      background: linear-gradient(135deg, rgba(0, 217, 255, 0.3) 0%, rgba(0, 153, 255, 0.2) 100%);
      border: 1px solid rgba(0, 217, 255, 0.3);
    }

    .task-icon {
      width: 36px;
      height: 36px;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #00d9ff;

      img {
        width: 100%;
        height: 100%;
        border-radius: 8px;
      }
    }
  }
}
</style>