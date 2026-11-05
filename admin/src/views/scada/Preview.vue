<template>
  <div class="scada-preview" :style="{ background: canvasConfig.fill }">
    <div class="preview-header" v-if="!fullscreen">
      <span class="preview-title">{{ scadaName }}</span>
      <div class="preview-actions">
        <el-button size="small" @click="toggleFullscreen">
          <el-icon><FullScreen /></el-icon>全屏投放
        </el-button>
        <el-button size="small" @click="handleBack">返回</el-button>
      </div>
    </div>
    <div class="preview-canvas" ref="previewView">
      <!-- echarts 图表覆盖层：随画布缩放定位 -->
      <div class="chart-overlay" ref="chartOverlay"></div>
    </div>
    <div v-if="loading" class="preview-loading" v-loading="true"></div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { FullScreen } from '@element-plus/icons-vue'
import { Leafer, UI } from 'leafer-editor'
import { getScada } from '@/api/scada'
import { DEFAULT_CANVAS, isChartElement, isPipeElement } from './elements'
import { ChartOverlayManager, fetchHistoryPoints } from './charts'

const route = useRoute()
const router = useRouter()

const previewView = ref(null)
const chartOverlay = ref(null)
const scadaName = ref('')
const loading = ref(true)
const fullscreen = ref(false)
const canvasConfig = reactive({ ...DEFAULT_CANVAS })

let leafer = null // 只读渲染画布实例
let chartManager = null // echarts 图表覆盖层管理器

// ========== 管道流动动画 ==========
let pipeAnimFrame = null // requestAnimationFrame 句柄
let pipeLastTime = 0     // 上一帧时间戳（毫秒）

// 启动管道流动动画：循环递减流动层 dashOffset 产生流动效果
// 流速优先取绑定属性映射的动态流速（_dynamicSpeed），否则取配置流速
function startPipeAnimation(pipes) {
  const animatedPipes = pipes.filter(el => el.data.pipe.animated !== false)
  if (animatedPipes.length === 0) return

  const step = (now) => {
    const dt = pipeLastTime ? Math.min(now - pipeLastTime, 100) : 16
    pipeLastTime = now
    animatedPipes.forEach(el => {
      const pipe = el.data.pipe
      const flow = el.children && el.children.find(c => c.name === 'pipe_flow')
      if (!flow) return
      // 动态流速优先（绑定属性实时调节），否则使用配置流速
      const speed = pipe._dynamicSpeed != null ? pipe._dynamicSpeed : (Number(pipe.speed) || 3)
      if (speed <= 0) return
      const direction = pipe.direction === -1 ? -1 : 1
      // dashOffset 递减使虚线沿正向流动（速度系数 0.06 将 1~10 映射为合适的像素速度）
      flow.dashOffset = (flow.dashOffset || 0) - direction * speed * 0.06 * dt
    })
    pipeAnimFrame = requestAnimationFrame(step)
  }
  pipeAnimFrame = requestAnimationFrame(step)
}

// 停止管道流动动画
function stopPipeAnimation() {
  if (pipeAnimFrame) {
    cancelAnimationFrame(pipeAnimFrame)
    pipeAnimFrame = null
  }
  pipeLastTime = 0
}

// 等比缩放画布以适应窗口，并同步图表覆盖层布局
function fitScale() {
  if (!leafer) return
  const wrapper = previewView.value
  if (!wrapper) return
  const scaleX = wrapper.clientWidth / canvasConfig.width
  const scaleY = wrapper.clientHeight / canvasConfig.height
  const scale = Math.min(scaleX, scaleY)
  leafer.scale = scale
  // 居中显示
  leafer.x = (wrapper.clientWidth - canvasConfig.width * scale) / 2
  leafer.y = (wrapper.clientHeight - canvasConfig.height * scale) / 2
  // 同步图表覆盖层位置与尺寸
  if (chartManager) {
    chartManager.updateLayout(scale, leafer.x, leafer.y)
  }
}

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen()
    fullscreen.value = true
  } else {
    document.exitFullscreen()
    fullscreen.value = false
  }
}

function onFullscreenChange() {
  fullscreen.value = !!document.fullscreenElement
  setTimeout(fitScale, 100)
}

function onResize() {
  fitScale()
}

function handleBack() {
  router.push('/scada')
}

onMounted(async () => {
  // 创建只读画布（不启用编辑器，仅渲染）
  leafer = new Leafer({
    view: previewView.value,
    fill: 'transparent',
    hittable: false
  })

  const loadedElements = []
  try {
    const res = await getScada(route.params.id)
    const scada = res.data
    if (scada) {
      scadaName.value = scada.name || ''
      if (scada.canvas) Object.assign(canvasConfig, scada.canvas)
      if (Array.isArray(scada.elements)) {
        scada.elements.forEach(data => {
          const el = UI.one(data)
          if (el) {
            el.hittable = false
            leafer.add(el)
            loadedElements.push(el)
          }
        })
      }
    }
  } catch (e) {
    console.error('load scada failed', e)
  } finally {
    loading.value = false
  }

  // 启动管道流动动画
  startPipeAnimation(loadedElements.filter(isPipeElement))

  // 初始化图表覆盖层：为所有图表图元创建 echarts 实例
  chartManager = new ChartOverlayManager(chartOverlay.value)
  const chartElements = loadedElements.filter(isChartElement)
  chartElements.forEach(el => chartManager.addChart(el))

  fitScale()

  // 加载历史聚合数据源（异步拉取后重绘）
  chartElements.forEach(async el => {
    const ds = el.data.chart.dataSource
    if (ds && ds.type === 'history') {
      const points = await fetchHistoryPoints(ds)
      if (chartManager) chartManager.updateData(el, points)
    }
  })

  window.addEventListener('resize', onResize)
  document.addEventListener('fullscreenchange', onFullscreenChange)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  document.removeEventListener('fullscreenchange', onFullscreenChange)
  // 停止管道流动动画
  stopPipeAnimation()
  // 销毁图表覆盖层
  if (chartManager) {
    chartManager.destroy()
    chartManager = null
  }
  if (leafer) {
    leafer.destroy()
    leafer = null
  }
})
</script>

<style lang="scss" scoped>
.scada-preview {
  position: fixed;
  inset: 0;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  z-index: 2000;

  .preview-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 48px;
    padding: 0 16px;
    background: rgba(26, 26, 46, 0.9);
    color: #fff;
    flex-shrink: 0;

    .preview-title {
      font-size: 15px;
      font-weight: 600;
    }

    .preview-actions {
      display: flex;
      gap: 8px;
    }
  }

  .preview-canvas {
    flex: 1;
    overflow: hidden;
    position: relative;

    .chart-overlay {
      position: absolute;
      inset: 0;
      pointer-events: none;
      z-index: 10;
    }
  }

  .preview-loading {
    position: absolute;
    inset: 0;
  }
}
</style>
