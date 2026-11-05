<template>
  <div class="window-manager">
    <Window
      v-for="window in windows"
      :key="window.id"
      :window-data="window"
      :max-z-index="maxZIndex"
      @activate="activateWindow"
      @minimize="minimizeWindow"
      @maximize="toggleMaximize"
      @close="closeWindow"
      @resize="resizeWindow"
    />
    <TaskBar
      :windows="windows"
      @restore="restoreWindow"
      @activate="activateWindow"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import Window from './Window.vue'
import TaskBar from './TaskBar.vue'

const windows = ref([])
let zIndexCounter = 100
let windowIdCounter = 0

const maxZIndex = computed(() => {
  if (windows.value.length === 0) return 100
  return Math.max(...windows.value.map(w => w.zIndex))
})

const openWindow = (app) => {
  const existingWindow = windows.value.find(w => w.appId === app.id)
  if (existingWindow) {
    activateWindow(existingWindow.id)
    return
  }

  const windowId = ++windowIdCounter
  const offset = (windows.value.length % 4) * 30
  
  const defaultX = 100 + offset
  const defaultY = 100 + offset
  const defaultWidth = 800
  const defaultHeight = 600
  
  windows.value.push({
    id: windowId,
    appId: app.id,
    title: app.name,
    icon: app.icon,
    url: app.url,
    x: defaultX,
    y: defaultY,
    width: defaultWidth,
    height: defaultHeight,
    prevX: defaultX,
    prevY: defaultY,
    prevWidth: defaultWidth,
    prevHeight: defaultHeight,
    zIndex: ++zIndexCounter,
    minimized: false,
    maximized: true
  })
}

const closeWindow = (id) => {
  const index = windows.value.findIndex(w => w.id === id)
  if (index !== -1) {
    windows.value.splice(index, 1)
  }
}

const minimizeWindow = (id) => {
  const window = windows.value.find(w => w.id === id)
  if (window) {
    window.minimized = true
  }
}

const restoreWindow = (id) => {
  const window = windows.value.find(w => w.id === id)
  if (window) {
    window.restoring = true
    setTimeout(() => {
      window.minimized = false
      window.restoring = false
      activateWindow(id)
    }, 300)
  }
}

const toggleMaximize = (id) => {
  const window = windows.value.find(w => w.id === id)
  if (window) {
    if (!window.maximized) {
      window.prevX = window.x
      window.prevY = window.y
      window.prevWidth = window.width
      window.prevHeight = window.height
    } else {
      window.x = window.prevX
      window.y = window.prevY
      window.width = window.prevWidth
      window.height = window.prevHeight
    }
    window.maximized = !window.maximized
    activateWindow(id)
  }
}

const activateWindow = (id) => {
  const window = windows.value.find(w => w.id === id)
  if (window) {
    window.zIndex = ++zIndexCounter
    window.minimized = false
  }
}

const resizeWindow = ({ id, x, y, width, height }) => {
  const window = windows.value.find(w => w.id === id)
  if (window) {
    if (x !== undefined) window.x = x
    if (y !== undefined) window.y = y
    if (width !== undefined) window.width = width
    if (height !== undefined) window.height = height
  }
}

const minimizeAll = () => {
  windows.value.forEach(w => {
    w.minimized = true
  })
}

defineExpose({
  openWindow,
  closeWindow,
  minimizeAll
})
</script>

<style lang="scss" scoped>
.window-manager {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  z-index: 1000;

  :deep(.desktop-window) {
    pointer-events: auto;
  }

  :deep(.task-bar) {
    pointer-events: auto;
  }
}
</style>