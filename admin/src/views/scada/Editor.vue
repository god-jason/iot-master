<template>
  <div class="scada-editor">
    <!-- 顶部工具栏 -->
    <div class="editor-toolbar">
      <div class="toolbar-left">
        <el-button text @click="handleBack">
          <el-icon><ArrowLeft /></el-icon>返回
        </el-button>
        <el-divider direction="vertical" />
        <el-input v-model="scadaName" placeholder="画面名称" class="name-input" size="small" />
        <el-tag :type="scadaType === 'scada' ? 'primary' : 'success'" size="small">
          {{ scadaType === 'scada' ? '组态' : '大屏' }}
        </el-tag>
      </div>
      <div class="toolbar-center">
        <el-tooltip content="撤销 (Ctrl+Z)">
          <el-button text :disabled="!canUndo" @click="undo"><el-icon><RefreshLeft /></el-icon></el-button>
        </el-tooltip>
        <el-tooltip content="重做 (Ctrl+Y)">
          <el-button text :disabled="!canRedo" @click="redo"><el-icon><RefreshRight /></el-icon></el-button>
        </el-tooltip>
        <el-divider direction="vertical" />
        <el-tooltip content="缩小">
          <el-button text @click="zoomCanvas('out')"><el-icon><ZoomOut /></el-icon></el-button>
        </el-tooltip>
        <el-tooltip content="放大">
          <el-button text @click="zoomCanvas('in')"><el-icon><ZoomIn /></el-icon></el-button>
        </el-tooltip>
        <el-tooltip content="适应画布">
          <el-button text @click="zoomCanvas('fit')"><el-icon><FullScreen /></el-icon></el-button>
        </el-tooltip>
        <el-divider direction="vertical" />
        <el-tooltip content="删除选中 (Delete)">
          <el-button text :disabled="!hasSelection" @click="deleteSelected"><el-icon><Delete /></el-icon></el-button>
        </el-tooltip>
        <el-tooltip content="置顶">
          <el-button text :disabled="!hasSelection" @click="layerToTop"><el-icon><DArrowRight class="rotate90" /></el-icon></el-button>
        </el-tooltip>
        <el-tooltip content="置底">
          <el-button text :disabled="!hasSelection" @click="layerToBottom"><el-icon><DArrowRight class="rotate270" /></el-icon></el-button>
        </el-tooltip>
        <el-divider direction="vertical" />
        <el-tooltip content="分组 (Ctrl+G)">
          <el-button text :disabled="!canGroup" @click="groupSelected"><el-icon><Files /></el-icon></el-button>
        </el-tooltip>
        <el-tooltip content="取消分组 (Ctrl+Shift+G)">
          <el-button text :disabled="!canUngroup" @click="ungroupSelected"><el-icon><FolderOpened /></el-icon></el-button>
        </el-tooltip>
      </div>
      <div class="toolbar-right">
        <!-- 画布设置：网格 / 坐标轴 开关 -->
        <el-divider direction="vertical" class="divider-left" />
        <div class="canvas-toggles">
          <el-tooltip content="网格 (G)">
            <el-button text :class="{ active: showGrid }" @click="toggleGrid"><el-icon><Grid /></el-icon></el-button>
          </el-tooltip>
          <el-tooltip content="坐标轴 (R)">
            <el-button text :class="{ active: showRuler }" @click="toggleRuler"><el-icon><DataLine /></el-icon></el-button>
          </el-tooltip>
        </div>
        <el-divider direction="vertical" />
        <el-tooltip content="画布尺寸">
          <el-dropdown trigger="click" @command="applyPresetCanvas">
            <el-button text size="small"><el-icon><Setting /></el-icon>画布 {{ canvasConfig.width }}×{{ canvasConfig.height }}</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item :command="{w: 1920, h: 1080}">1920 × 1080（16:9）</el-dropdown-item>
                <el-dropdown-item :command="{w: 1280, h: 720}">1280 × 720（16:9）</el-dropdown-item>
                <el-dropdown-item :command="{w: 2560, h: 1440}">2560 × 1440（2K）</el-dropdown-item>
                <el-dropdown-item :command="{w: 1024, h: 768}">1024 × 768（4:3）</el-dropdown-item>
                <el-dropdown-item :command="{w: 3840, h: 2160}">3840 × 2160（4K）</el-dropdown-item>
                <el-dropdown-item :command="{w: 800, h: 600}">800 × 600</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </el-tooltip>
        <el-button size="small" @click="handlePreview">预览</el-button>
        <el-button size="small" type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </div>
    </div>

    <div class="editor-body">
      <!-- 左侧：分组折叠面板 -->
      <div class="panel-left">
        <el-collapse v-model="leftActiveGroups" class="left-collapse">
          <!-- 基础图元 -->
          <el-collapse-item title="基础图元" name="basic">
            <div class="element-grid">
              <div
                v-for="item in basicElements"
                :key="item.type"
                class="element-item"
                @click="addElement(item.type)"
              >
                <span class="element-icon" v-html="item.icon"></span>
                <span class="element-name">{{ item.name }}</span>
              </div>
            </div>
          </el-collapse-item>
          <!-- 页面控件 -->
          <el-collapse-item title="页面控件" name="control">
            <div class="element-grid">
              <div
                v-for="item in controlElements"
                :key="item.type"
                class="element-item"
                @click="addElement(item.type)"
              >
                <span class="element-icon" v-html="item.icon"></span>
                <span class="element-name">{{ item.name }}</span>
              </div>
            </div>
          </el-collapse-item>
          <!-- 组件库 -->
          <el-collapse-item name="components">
            <template #title>
              <span class="collapse-title">
                组件库
                <el-button size="small" text type="primary" :disabled="!hasSelection" @click="saveAsComponent">存为组件</el-button>
              </span>
            </template>
            <div v-for="(items, key) in groupedComponentList" :key="key" class="comp-group">
              <div class="comp-group-title">{{ key }}</div>
              <div class="component-grid" v-loading="componentLoading">
                <div
                  v-for="comp in items"
                  :key="comp.id"
                  class="component-item"
                  @click="addComponent(comp)"
                >
                  <!-- 组件缩略图：有 icon 显示图片，否则显示占位 -->
                  <div class="comp-thumb">
                    <img v-if="comp.icon" :src="comp.icon.startsWith('data:') || comp.icon.startsWith('http') ? comp.icon : baseUrl + comp.icon" alt="" />
                    <el-icon v-else :size="26"><FolderOpened /></el-icon>
                  </div>
                  <span class="component-name">{{ comp.name }}</span>
                </div>
              </div>
            </div>
            <div v-if="!componentLoading && componentList.length === 0" class="component-empty">
              暂无组件，选中图元后点击"存为组件"
            </div>
          </el-collapse-item>
        </el-collapse>
      </div>

      <!-- 中间：画布 -->
      <div class="canvas-wrapper" ref="canvasWrapper">
        <div class="canvas-view" ref="canvasView"></div>
      </div>

      <!-- 右侧：折叠面板组 -->
      <div class="panel-right">
        <el-collapse v-model="rightActiveGroups" class="right-collapse" accordion>
          <!-- 属性 -->
          <el-collapse-item title="属性" name="props">
            <div v-if="!selectedElement" class="empty-tip">未选中图元</div>
            <div v-else class="props-form">
              <div class="setting-row">
                <span>名称</span>
                <el-input v-model="propsForm.name" size="small" @change="applyProp('name')" />
              </div>
              <div class="setting-row">
                <span>X</span>
                <el-input-number v-model="propsForm.x" size="small" controls-position="right" @change="applyProp('x')" />
              </div>
              <div class="setting-row">
                <span>Y</span>
                <el-input-number v-model="propsForm.y" size="small" controls-position="right" @change="applyProp('y')" />
              </div>
              <template v-if="!isLine && !isGroup">
                <div class="setting-row">
                  <span>宽度</span>
                  <el-input-number v-model="propsForm.width" :min="1" size="small" controls-position="right" @change="applyProp('width')" />
                </div>
                <div class="setting-row">
                  <span>高度</span>
                  <el-input-number v-model="propsForm.height" :min="1" size="small" controls-position="right" @change="applyProp('height')" />
                </div>
              </template>
              <div class="setting-row">
                <span>旋转</span>
                <el-input-number v-model="propsForm.rotation" :min="-360" :max="360" size="small" controls-position="right" @change="applyProp('rotation')" />
              </div>
              <div class="setting-row">
                <span>透明度</span>
                <el-slider v-model="propsForm.opacity" :min="0" :max="1" :step="0.05" @change="applyProp('opacity')" />
              </div>
              <div class="setting-row" v-if="hasFill">
                <span>填充</span>
                <el-color-picker v-model="propsForm.fill" size="small" @change="applyProp('fill')" />
              </div>
              <div class="setting-row" v-if="hasStroke">
                <span>描边</span>
                <el-color-picker v-model="propsForm.stroke" size="small" show-alpha @change="applyProp('stroke')" />
              </div>
              <div class="setting-row" v-if="hasStroke">
                <span>线宽</span>
                <el-input-number v-model="propsForm.strokeWidth" :min="0" :max="50" size="small" controls-position="right" @change="applyProp('strokeWidth')" />
              </div>
              <div class="setting-row" v-if="isRect">
                <span>圆角</span>
                <el-input-number v-model="propsForm.cornerRadius" :min="0" size="small" controls-position="right" @change="applyProp('cornerRadius')" />
              </div>
              <template v-if="isText">
                <div class="setting-row">
                  <span>内容</span>
                  <el-input v-model="propsForm.text" type="textarea" :rows="2" size="small" @change="applyProp('text')" />
                </div>
                <div class="setting-row">
                  <span>字号</span>
                  <el-input-number v-model="propsForm.fontSize" :min="8" :max="200" size="small" controls-position="right" @change="applyProp('fontSize')" />
                </div>
                <div class="setting-row">
                  <span>字体</span>
                  <el-select v-model="propsForm.fontFamily" size="small" @change="applyProp('fontFamily')">
                    <el-option v-for="f in FONT_FAMILIES" :key="f" :label="f" :value="f" />
                  </el-select>
                </div>
                <div class="setting-row">
                  <span>加粗</span>
                  <el-switch v-model="propsForm.fontWeight" size="small" @change="applyProp('fontWeight')" />
                </div>
              </template>
              <div class="setting-row" v-if="isImage">
                <span>图片URL</span>
                <el-input v-model="propsForm.url" size="small" placeholder="图片地址" @change="applyProp('url')" />
              </div>
            </div>
          </el-collapse-item>
          <!-- 图层 -->
          <el-collapse-item title="图层" name="layers">
            <div v-if="layers.length === 0" class="empty-tip">暂无图元</div>
            <div v-else class="layer-list">
              <div
                v-for="layer in layers"
                :key="layer.id"
                class="layer-item"
                :class="{ active: layer.selected }"
                @click="selectLayer(layer)"
              >
                <el-icon><Component :is="layerIcon(layer)" /></el-icon>
                <span class="layer-name">{{ layer.name }}</span>
                <el-icon class="layer-del" @click.stop="deleteElement(layer.el)"><Close /></el-icon>
              </div>
            </div>
          </el-collapse-item>
          <!-- 绑定 -->
          <el-collapse-item v-if="selectedElement" title="设备绑定" name="binding">
            <div class="binding-form">
              <div class="setting-row">
                <span>设备</span>
                <el-select
                  v-model="bindingForm.device_id"
                  size="small" filterable remote clearable
                  :remote-method="searchDeviceOptions" :loading="deviceLoading"
                  placeholder="搜索设备" @change="onDeviceChange">
                  <el-option v-for="d in deviceOptions" :key="d.id" :label="d.name || d.sn" :value="d.id" />
                </el-select>
              </div>
              <div class="setting-row">
                <span>属性</span>
                <el-select v-model="bindingForm.property" size="small" clearable
                  :disabled="!bindingForm.device_id" placeholder="选择物模型属性" @change="onPropertyChange">
                  <el-option v-for="p in propertyOptions" :key="p.key" :label="`${p.name}(${p.key})`" :value="p.key" />
                </el-select>
              </div>
              <div class="setting-row">
                <span>最小值</span>
                <el-input-number v-model="bindingForm.min" size="small" controls-position="right" />
              </div>
              <div class="setting-row">
                <span>最大值</span>
                <el-input-number v-model="bindingForm.max" size="small" controls-position="right" />
              </div>
              <div class="setting-row">
                <span>精度</span>
                <el-input-number v-model="bindingForm.precision" :min="0" :max="6" size="small" controls-position="right" />
              </div>
              <div class="setting-row">
                <span>单位</span>
                <el-input v-model="bindingForm.unit" size="small" placeholder="如 ℃" />
              </div>
              <div class="threshold-section">
                <div class="threshold-header">
                  <span>阈值变色规则</span>
                  <el-button size="small" text type="primary" @click="addThreshold">添加</el-button>
                </div>
                <div v-for="(t, idx) in bindingForm.thresholds" :key="idx" class="threshold-row">
                  <span class="threshold-op">≥</span>
                  <el-input-number v-model="t.value" size="small" controls-position="right" />
                  <el-color-picker v-model="t.color" size="small" />
                  <el-icon class="threshold-del" @click="removeThreshold(idx)"><Close /></el-icon>
                </div>
                <div v-if="bindingForm.thresholds.length === 0" class="threshold-empty">无阈值规则（数值≥阈值时图元变色）</div>
              </div>
              <div class="binding-actions">
                <el-button type="primary" size="small" :disabled="!bindingForm.device_id || !bindingForm.property" @click="applyBinding">应用绑定</el-button>
                <el-button size="small" :disabled="!currentBinding" @click="clearBinding">清除绑定</el-button>
              </div>
              <div v-if="currentBinding" class="binding-status">
                已绑定：{{ currentBinding.property }} @ {{ currentBinding.device_id }}
              </div>
            </div>
          </el-collapse-item>
          <!-- 图表 -->
          <el-collapse-item v-if="selectedElement && isChart" title="图表配置" name="chart">
            <div class="chart-form">
              <div class="setting-row">
                <span>标题</span>
                <el-input v-model="chartForm.title" size="small" placeholder="图表标题" @change="applyChartConfig" />
              </div>
              <div class="setting-row">
                <span>数据源</span>
                <el-select v-model="chartForm.dsType" size="small" @change="onDsTypeChange">
                  <el-option label="静态数据" value="static" />
                  <el-option label="历史聚合" value="history" />
                </el-select>
              </div>
              <template v-if="chartForm.dsType === 'static'">
                <div class="static-section">
                  <div class="static-header">
                    <span>数据项</span>
                    <el-button size="small" text type="primary" @click="addStaticItem">添加</el-button>
                  </div>
                  <div v-for="(item, idx) in chartForm.staticData" :key="idx" class="static-row">
                    <el-input v-model="item.name" size="small" placeholder="名称" />
                    <el-input-number v-model="item.value" size="small" controls-position="right" />
                    <el-icon class="static-del" @click="removeStaticItem(idx)"><Close /></el-icon>
                  </div>
                  <div v-if="chartForm.staticData.length === 0" class="static-empty">暂无数据项</div>
                </div>
              </template>
              <template v-if="chartForm.dsType === 'history'">
                <div class="setting-row">
                  <span>设备</span>
                  <el-select v-model="chartForm.device_id" size="small" filterable remote clearable
                    :remote-method="searchDeviceOptions" :loading="deviceLoading"
                    placeholder="搜索设备" @change="onChartDeviceChange">
                    <el-option v-for="d in deviceOptions" :key="d.id" :label="d.name || d.sn" :value="d.id" />
                  </el-select>
                </div>
                <div class="setting-row">
                  <span>属性</span>
                  <el-select v-model="chartForm.property" size="small" clearable
                    :disabled="!chartForm.device_id" placeholder="选择属性">
                    <el-option v-for="p in propertyOptions" :key="p.key" :label="`${p.name}(${p.key})`" :value="p.key" />
                  </el-select>
                </div>
                <template v-if="chartForm.dsType === 'history'">
                  <div class="setting-row">
                    <span>窗口</span>
                    <el-select v-model="chartForm.window" size="small">
                      <el-option label="5分钟" value="5m" /><el-option label="15分钟" value="15m" />
                      <el-option label="1小时" value="1h" /><el-option label="6小时" value="6h" /><el-option label="1天" value="1d" />
                    </el-select>
                  </div>
                  <div class="setting-row">
                    <span>聚合</span>
                    <el-select v-model="chartForm.method" size="small">
                      <el-option label="平均值" value="mean" /><el-option label="最大值" value="max" />
                      <el-option label="最小值" value="min" /><el-option label="求和" value="sum" /><el-option label="最后值" value="last" />
                    </el-select>
                  </div>
                  <div class="setting-row">
                    <span>点数</span>
                    <el-input-number v-model="chartForm.limit" :min="2" :max="200" size="small" controls-position="right" />
                  </div>
                </template>
              </template>
              <div class="chart-actions">
                <el-button type="primary" size="small" @click="applyChartConfig">应用配置</el-button>
              </div>
            </div>
          </el-collapse-item>
          <!-- 管道 -->
          <el-collapse-item v-if="selectedElement && isPipe" title="管道配置" name="pipe">
            <div class="pipe-form">
              <div class="setting-row">
                <span>流动颜色</span>
                <el-color-picker v-model="pipeForm.flowColor" size="small" @change="applyPipeConfig" />
              </div>
              <div class="setting-row">
                <span>流速</span>
                <el-slider v-model="pipeForm.speed" :min="1" :max="10" :step="1" show-stops @change="applyPipeConfig" />
              </div>
              <div class="setting-row">
                <span>方向</span>
                <el-radio-group v-model="pipeForm.direction" size="small" @change="applyPipeConfig">
                  <el-radio-button :value="1">正向</el-radio-button>
                  <el-radio-button :value="-1">反向</el-radio-button>
                </el-radio-group>
              </div>
              <div class="setting-row">
                <span>流动动画</span>
                <el-switch v-model="pipeForm.animated" @change="applyPipeConfig" />
              </div>
              <div class="pipe-tip">预览时管道将按配置播放流动动画；绑定数值属性后可按属性值动态调节流速</div>
            </div>
          </el-collapse-item>
        </el-collapse>
      </div>
    </div>

    <!-- 预览弹窗：直接把当前画布图元渲染到独立 Leafer 画布，无需先保存 -->
    <el-dialog v-model="previewVisible" title="实时预览" width="900px" top="4vh" :close-on-click-modal="false" destroy-on-close>
      <div class="preview-dialog-body">
        <div class="preview-canvas-holder" ref="previewCanvasView"></div>
      </div>
      <template #footer>
        <el-button @click="closePreview">关闭</el-button>
        <el-button type="primary" plain @click="handleSave">保存并全屏预览</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowLeft, RefreshLeft, RefreshRight, ZoomIn, ZoomOut, FullScreen,
  Delete, DArrowRight, Close, CopyDocument, Aim, Minus, Document,
  Picture, Odometer, Notebook, Files, FolderOpened, Grid, DataLine, Setting
} from '@element-plus/icons-vue'
import {
  App, Rect, Ellipse, Line, Text, Group, UI, Leafer,
  EditorEvent, EditorMoveEvent, EditorScaleEvent, EditorRotateEvent
} from 'leafer-editor'
import './PointsEditTool'  // 注册自定义顶点编辑工具
import '@leafer-in/text-editor'  // 注册文本双击编辑 internal editor
import { ELEMENT_LIBRARY, getElementsByCategory, CATEGORY_NAMES, CATEGORY_ORDER, ELEMENT_CATEGORY, createElement, getElementTypeName, CANVAS_FRAME_ID, DEFAULT_CANVAS, isChartElement, isPipeElement, patchEditable, FONT_FAMILIES } from './elements'
import { getScada, createScada, updateScada } from '@/api/scada'
import { getDeviceList, getDevice } from '@/api/device'
import { getModel } from '@/api/model'
import { searchComponents, createComponent, getComponent } from '@/api/component'

const route = useRoute()
const router = useRouter()

// ========== 基础状态 ==========
const canvasView = ref(null)
const canvasWrapper = ref(null)
const scadaName = ref('未命名画面')
const scadaType = ref(route.query.type === 'screen' ? 'screen' : 'scada')
const scadaId = ref(route.params.id === 'new' ? '' : route.params.id)
const saving = ref(false)
const canvasConfig = reactive({ ...DEFAULT_CANVAS })

// 左侧面板：折叠状态（默认全部展开）
const leftActiveGroups = ref(['basic', 'control', 'components'])
// 画布网格开关
const showGrid = ref(true)
// 坐标轴（标尺）开关
const showRuler = ref(true)
const baseUrl = import.meta.env.VITE_API_BASE_URL || ''

// 右侧折叠面板默认展开
const rightActiveGroups = ref(['props', 'layers'])

// ===== 预览弹窗相关 =====
const previewVisible = ref(false)
const previewCanvasView = ref(null)
let previewLeafer = null  // 预览画布只读渲染实例

// 左侧图元分组 computed
const basicElements = computed(() => getElementsByCategory('basic'))
const controlElements = computed(() => getElementsByCategory('control'))
// 组件库按 category 分组显示（组件缩略图在 canvas 的 app.view 容器里渲染后导出 dataURL）
const groupedComponentList = computed(() => {
  const groups = {}
  componentList.value.forEach(comp => {
    const key = CATEGORY_NAMES[comp.category] || '自定义分组'
    if (!groups[key]) groups[key] = []
    groups[key].push(comp)
  })
  return groups
})

let app = null           // Leafer App 实例
let canvasFrame = null   // 画布背景框图元
let undoStack = []       // 撤销栈
let redoStack = []       // 重做栈
let pendingSnapshot = null // 拖拽/缩放/旋转开始前的快照
const HISTORY_LIMIT = 50

// ground 层上的装饰元素（网格点、刻度线、数字标签），用于开关控制可见性
let groundDeco = { gridDots: [], rulerLines: [], rulerLabels: [] }

// ========== 网格 & 坐标轴 ==========

// canvasFrame Group 在 tree 层的偏移量（坐标轴和刻度用这个偏移定位）
const CANVAS_OFFSET = 30

// 清除所有装饰元素（grid + ruler）
function clearDecoration() {
  const remove = (el) => { try { el.remove() } catch (_) {} }
  groundDeco.gridDots.forEach(remove)
  groundDeco.rulerLines.forEach(remove)
  groundDeco.rulerLabels.forEach(remove)
  groundDeco.gridDots = []
  groundDeco.rulerLines = []
  groundDeco.rulerLabels = []
}

// 重画所有装饰（画布尺寸变化时调用）
function redrawDecoration() {
  clearDecoration()
  drawGrid()
  drawRuler()
}

// 在 canvasFrame Group 内部画网格点（画面可见区域，Group 本地坐标）
function drawGrid() {
  if (!canvasFrame) return
  const W = canvasConfig.width, H = canvasConfig.height
  const GAP = 20
  for (let x = GAP; x < W; x += GAP) {
    for (let y = GAP; y < H; y += GAP) {
      const dot = new Ellipse({
        x: x - 1, y: y - 1, width: 2, height: 2,
        fill: '#b4bccc', hittable: false, visible: showGrid.value,
        userData: { __decoration: true }
      })
      canvasFrame.add(dot)
      groundDeco.gridDots.push(dot)
    }
  }
}

// 在 tree 层画坐标轴（紧贴 canvasFrame Group 外边缘，tree 层坐标 = 屏幕坐标，全在视口内可见）
// canvasFrame 在 tree 层坐标 (CANVAS_OFFSET, CANVAS_OFFSET)
function drawRuler() {
  if (!app || !app.tree) return
  const W = canvasConfig.width, H = canvasConfig.height
  const TICK = 100
  const O = CANVAS_OFFSET              // canvasFrame 在 tree 层的偏移
  const DEC = { __decoration: true }
  const stroke = '#5a5a7a'
  const labelColor = '#909399'
  const base = { hittable: false, visible: showRuler.value, userData: DEC }

  // ===== X 轴水平线：画面上边缘 (y=O) =====
  // 从画面左侧深灰区 (0, O) 画到画面右侧深灰区 (O+W+O, O)
  const axisX = new Line({
    x: 0, y: O, toPoint: { x: W + O * 2, y: 0 },
    stroke, strokeWidth: 1, ...base
  })
  // ===== Y 轴垂直线：画面左边缘 (x=O) =====
  // 从画面上方深灰区 (O, 0) 画到画面下方深灰区 (O, O+H+O)
  const axisY = new Line({
    x: O, y: 0, toPoint: { x: 0, y: H + O * 2 },
    stroke, strokeWidth: 1, ...base
  })
  app.tree.addMany(axisX, axisY)
  groundDeco.rulerLines.push(axisX, axisY)

  // ===== 原点 "0" 标签（在 X/Y 轴交汇处，画面左上外侧）=====
  const origin = new Text({
    x: O - 20, y: O + 2, width: 24, height: 14,
    text: '0', fontSize: 10, fill: labelColor, textAlign: 'left',
    ...base
  })
  app.tree.add(origin)
  groundDeco.rulerLabels.push(origin)

  // ===== X 轴刻度 + 数字 =====
  // 沿画面顶部（tree 层 y=O），每 TICK 一个，tick 向上延伸（画面外）
  for (let i = 1; i * TICK < W; i++) {
    const x = O + i * TICK              // tree 层绝对坐标
    const tick = new Line({
      x, y: O, toPoint: { x: 0, y: -8 },
      stroke, strokeWidth: 1, ...base
    })
    const lbl = new Text({
      x: x - 14, y: O - 22, width: 40, height: 14,
      text: String(i * TICK), fontSize: 10, fill: labelColor, textAlign: 'left',
      ...base
    })
    app.tree.addMany(tick, lbl)
    groundDeco.rulerLines.push(tick)
    groundDeco.rulerLabels.push(lbl)
  }

  // ===== Y 轴刻度 + 数字 =====
  // 沿画面左侧（tree 层 x=O），每 TICK 一个，tick 向左延伸（画面外）
  for (let i = 1; i * TICK < H; i++) {
    const y = O + i * TICK
    const tick = new Line({
      x: O, y, toPoint: { x: -8, y: 0 },
      stroke, strokeWidth: 1, ...base
    })
    const lbl = new Text({
      x: O - 50, y: y - 8, width: 46, height: 14,
      text: String(i * TICK), fontSize: 10, fill: labelColor, textAlign: 'right',
      ...base
    })
    app.tree.addMany(tick, lbl)
    groundDeco.rulerLines.push(tick)
    groundDeco.rulerLabels.push(lbl)
  }
}

// 网格开关
function toggleGrid() {
  showGrid.value = !showGrid.value
  groundDeco.gridDots.forEach(d => d.visible = showGrid.value)
}

// 坐标轴开关
function toggleRuler() {
  showRuler.value = !showRuler.value
  groundDeco.rulerLines.forEach(l => l.visible = showRuler.value)
  groundDeco.rulerLabels.forEach(l => l.visible = showRuler.value)
}

// 画布尺寸下拉菜单项
function applyPresetCanvas(cmd) {
  canvasConfig.width = cmd.w
  canvasConfig.height = cmd.h
  applyCanvasConfig()
  redrawDecoration()
}

// ========== 选中状态 ==========
const selectedElement = ref(null)
const hasSelection = computed(() => !!selectedElement.value)
const isLine = computed(() => selectedElement.value?.tag === 'Line')
const isGroup = computed(() => selectedElement.value?.tag === 'Group')
const isRect = computed(() => selectedElement.value?.tag === 'Rect')
const isText = computed(() => selectedElement.value?.tag === 'Text')
const isImage = computed(() => selectedElement.value?.tag === 'Image')
const isChart = computed(() => isChartElement(selectedElement.value))
const isPipe = computed(() => isPipeElement(selectedElement.value))
// 颜色编辑项：Line 无填充，Text 无描边；其余图元（含 Group 容器）均可编辑填充/描边
const hasFill = computed(() => !isLine.value)
const hasStroke = computed(() => !isText.value)

// 属性表单（与选中图元双向同步）
const propsForm = reactive({
  name: '', x: 0, y: 0, width: 0, height: 0, rotation: 0, opacity: 1,
  fill: '', stroke: '', strokeWidth: 0, cornerRadius: 0,
  text: '', fontSize: 16, fontFamily: 'Arial', fontWeight: false, url: ''
})

// ========== 图层列表 ==========
const layers = ref([])
const layerIcon = (layer) => {
  const map = { Rect: CopyDocument, Ellipse: Aim, Line: Minus, Text: Document, Image: Picture, Group: Odometer }
  return map[layer.tag] || Notebook
}

function refreshLayers() {
  if (!canvasFrame) return
  const list = []
  // 倒序遍历 canvasFrame.children（非装饰），最上层图元显示在最前面
  const children = canvasFrame.children
  for (let i = children.length - 1; i >= 0; i--) {
    const el = children[i]
    if (isDecoration(el)) continue
    list.push({
      id: el.innerId || el.id,
      el,
      tag: el.tag,
      name: getElementTypeName(el),
      selected: selectedElement.value === el
    })
  }
  layers.value = list
}

function selectLayer(layer) {
  app.editor.select(layer.el)
}

// ========== 序列化 ==========
function isDecoration(el) { return el?.userData?.__decoration === true }

// 将画布图元序列化为 JSON 数组（只序列化 canvasFrame Group 内的非装饰 children）
function serializeElements() {
  if (!canvasFrame) return []
  return canvasFrame.children
    .filter(el => !isDecoration(el))
    .map(el => el.toJSON())
}

// 快照当前状态（用于撤销/重做）
function snapshot() {
  return JSON.stringify(serializeElements())
}

// 压入撤销栈
function pushHistory() {
  undoStack.push(snapshot())
  if (undoStack.length > HISTORY_LIMIT) undoStack.shift()
  redoStack = []
}

const canUndo = ref(false)
const canRedo = ref(false)
function updateHistoryState() {
  canUndo.value = undoStack.length > 0
  canRedo.value = redoStack.length > 0
}

// 从快照恢复画布
function restoreSnapshot(json) {
  const list = JSON.parse(json)
  clearElements()
  list.forEach(data => {
    const el = UI.one(data)
    if (el) { patchEditable(el); canvasFrame.add(el) }
  })
  app.editor.cancel()
  refreshLayers()
}

// 清空所有用户图元（保留 bgRect 和网格点等装饰）
function clearElements() {
  const children = [...canvasFrame.children]
  children.forEach(el => {
    if (!isDecoration(el)) el.remove()
  })
}

function undo() {
  if (!undoStack.length) return
  redoStack.push(snapshot())
  restoreSnapshot(undoStack.pop())
  updateHistoryState()
}

function redo() {
  if (!redoStack.length) return
  undoStack.push(snapshot())
  restoreSnapshot(redoStack.pop())
  updateHistoryState()
}

// ========== 图元操作 ==========
function addElement(type) {
  const el = createElement(type)
  if (!el) return
  pushHistory()
  canvasFrame.add(el)
  app.editor.select(el)
  refreshLayers()
  updateHistoryState()
}

function deleteElement(el) {
  pushHistory()
  el.remove()
  if (selectedElement.value === el) app.editor.cancel()
  refreshLayers()
  updateHistoryState()
}

function deleteSelected() {
  const list = app.editor.list
  if (!list.length) return
  pushHistory()
  list.forEach(el => el.remove())
  app.editor.cancel()
  refreshLayers()
  updateHistoryState()
}

function layerToTop() {
  const el = selectedElement.value
  if (!el) return
  pushHistory()
  canvasFrame.add(el)
  refreshLayers()
  updateHistoryState()
}

function layerToBottom() {
  const el = selectedElement.value
  if (!el) return
  pushHistory()
  // 插到装饰元素之后（装饰元素有 __decoration 标记，先找到最后一个装饰元素的位置）
  const lastDecoIdx = canvasFrame.children.reduce((last, c, i) => isDecoration(c) ? i : last, -1)
  canvasFrame.addAt(el, lastDecoIdx + 1)
  refreshLayers()
  updateHistoryState()
}

// ========== 图层分组 ==========
// 可分组：多选至少 2 个图元
const canGroup = computed(() => app && app.editor && app.editor.list.length >= 2)
// 可取消分组：选中单个 Group 图元
const canUngroup = computed(() => {
  const el = selectedElement.value
  return !!el && el.tag === 'Group'
})

// 分组：将选中的多个图元包装为一个 Group
function groupSelected() {
  if (!app) return
  const list = app.editor.list
  if (list.length < 2) return
  pushHistory()
  // 记录第一个图元在 canvasFrame 中的位置，分组后放回原位
  const first = list[0]
  const index = canvasFrame.children.indexOf(first)
  const group = new Group({ name: '分组' })
  // 先加入 canvasFrame 再移动子图元
  canvasFrame.addAt(group, index >= 0 ? index : canvasFrame.children.length)
  list.forEach(el => group.add(el))
  app.editor.select(group)
  refreshLayers()
  updateHistoryState()
}

// 取消分组：将 Group 内的子图元释放回树
// 分组创建时位于原点，子图元坐标与树坐标系一致；若分组被拖动过，需叠加其位移
// 注意：仅处理平移，分组被旋转/缩放后取消分组可能存在偏差
function ungroupSelected() {
  const el = selectedElement.value
  if (!el || el.tag !== 'Group') return
  pushHistory()
  const index = app.tree.children.indexOf(el)
  const children = [...el.children]
  const offsetX = el.x || 0
  const offsetY = el.y || 0
  // 将子图元逐个移回树的分组位置，并叠加分组位移
  children.forEach((child, i) => {
    child.x = (child.x || 0) + offsetX
    child.y = (child.y || 0) + offsetY
    app.tree.addAt(child, index + i)
  })
  el.remove()
  app.editor.cancel()
  refreshLayers()
  updateHistoryState()
}

// ========== 属性面板 ==========
// 控件类 Group 的"颜色承接子元素"：将填充/描边映射到第一个子元素（背景），
// 避免填充被内部不透明背景遮挡而不可见。普通分组（无 data.component）返回自身，
// 颜色直接作用于 Group 容器。
function getColorTarget(el) {
  if (el && el.tag === 'Group' && el.data && el.data.component && el.children && el.children.length) {
    return el.children[0]
  }
  return el
}

// 选中图元变化时同步属性表单
function syncPropsForm() {
  const el = selectedElement.value
  if (!el) return
  propsForm.name = el.name || ''
  propsForm.x = Math.round(el.x || 0)
  propsForm.y = Math.round(el.y || 0)
  propsForm.width = Math.round(el.width || 0)
  propsForm.height = Math.round(el.height || 0)
  propsForm.rotation = Math.round(el.rotation || 0)
  propsForm.opacity = el.opacity ?? 1
  const colorEl = getColorTarget(el)
  propsForm.fill = typeof colorEl.fill === 'string' ? colorEl.fill : ''
  propsForm.stroke = typeof colorEl.stroke === 'string' ? colorEl.stroke : ''
  propsForm.strokeWidth = colorEl.strokeWidth || 0
  propsForm.cornerRadius = el.cornerRadius || 0
  propsForm.text = el.text || ''
  propsForm.fontSize = el.fontSize || 16
  propsForm.fontFamily = el.fontFamily || 'Arial'
  propsForm.fontWeight = (el.fontWeight || '').includes('bold')
  propsForm.url = el.url || ''
}

// 应用属性修改到选中图元
function applyProp(key) {
  const el = selectedElement.value
  if (!el) return
  pushHistory()
  // fontWeight 特殊处理：propsForm 里是 bool，leafer el.fontWeight 用字符串 'bold' / 'normal'
  if (key === 'fontWeight') {
    el.fontWeight = propsForm.fontWeight ? 'bold' : 'normal'
  } else if (key === 'fill' || key === 'stroke' || key === 'strokeWidth') {
    // 控件 Group 的颜色作用于内部背景子元素，其余图元作用于自身
    getColorTarget(el)[key] = propsForm[key]
  } else {
    el[key] = propsForm[key]
  }
  if (key === 'name') refreshLayers()
  updateHistoryState()
}

// ========== 数据绑定面板 ==========
// 绑定表单：设备、属性、量程、精度、单位、阈值规则
const bindingForm = reactive({
  device_id: '',
  property: '',
  min: 0,
  max: 100,
  precision: 1,
  unit: '',
  thresholds: [] // [{value, color}]
})
const deviceOptions = ref([])   // 设备下拉选项
const deviceLoading = ref(false)
const propertyOptions = ref([]) // 物模型属性下拉选项

// 当前选中图元的绑定配置（只读展示）
const currentBinding = computed(() => {
  const el = selectedElement.value
  const binding = el && el.data && el.data.binding
  return (binding && binding.device_id && binding.property) ? binding : null
})

// 搜索设备下拉选项
async function searchDeviceOptions(keyword) {
  deviceLoading.value = true
  try {
    const res = await getDeviceList({ page: 1, page_size: 20, keyword: keyword || '' })
    deviceOptions.value = res.data || []
  } catch (e) {
    console.error('search devices failed', e)
  } finally {
    deviceLoading.value = false
  }
}

// 设备选择变化：加载该设备物模型的属性列表
async function onDeviceChange(deviceID) {
  bindingForm.property = ''
  propertyOptions.value = []
  if (!deviceID) return
  try {
    // 优先从下拉选项中查找，找不到则查询设备详情（回显场景）
    let device = deviceOptions.value.find(d => d.id === deviceID)
    if (!device || !device.model_id) {
      const res = await getDevice(deviceID)
      device = res.data
      if (device && !deviceOptions.value.find(d => d.id === deviceID)) {
        deviceOptions.value.push(device)
      }
    }
    if (!device || !device.model_id) return
    const res = await getModel(device.model_id)
    const model = res.data
    if (model && Array.isArray(model.properties)) {
      propertyOptions.value = model.properties
    }
  } catch (e) {
    console.error('load model failed', e)
  }
}

// 属性选择变化：自动填充量程/单位/精度/枚举
function onPropertyChange(key) {
  if (!key) return
  const prop = propertyOptions.value.find(p => p.key === key)
  if (!prop) return
  if (prop.unit) bindingForm.unit = prop.unit
  if (Number.isFinite(prop.precision)) bindingForm.precision = prop.precision
  if (Number.isFinite(prop.threshold)) bindingForm.min = 0
}

// 添加阈值规则
function addThreshold() {
  bindingForm.thresholds.push({ value: 0, color: '#f56c6c' })
}

// 删除阈值规则
function removeThreshold(idx) {
  bindingForm.thresholds.splice(idx, 1)
}

// 应用绑定到选中图元
function applyBinding() {
  const el = selectedElement.value
  if (!el) return
  if (!bindingForm.device_id || !bindingForm.property) {
    ElMessage.warning('请选择设备和属性')
    return
  }
  pushHistory()
  // 写入图元 data.binding（序列化时会一并保存）
  el.data = {
    ...(el.data || {}),
    binding: {
      device_id: bindingForm.device_id,
      property: bindingForm.property,
      min: bindingForm.min,
      max: bindingForm.max,
      precision: bindingForm.precision,
      unit: bindingForm.unit,
      thresholds: bindingForm.thresholds.filter(t => t.color),
      enums: propertyOptions.value.find(p => p.key === bindingForm.property)?.enums || []
    }
  }
  updateHistoryState()
  ElMessage.success('绑定已应用')
}

// 清除选中图元的绑定
function clearBinding() {
  const el = selectedElement.value
  if (!el || !el.data) return
  pushHistory()
  delete el.data.binding
  updateHistoryState()
  ElMessage.success('绑定已清除')
}

// 选中图元变化时同步绑定表单
function syncBindingForm() {
  const el = selectedElement.value
  const binding = el && el.data && el.data.binding
  if (binding && binding.device_id) {
    bindingForm.device_id = binding.device_id
    bindingForm.property = binding.property || ''
    bindingForm.min = Number(binding.min ?? 0)
    bindingForm.max = Number(binding.max ?? 100)
    bindingForm.precision = Number(binding.precision ?? 1)
    bindingForm.unit = binding.unit || ''
    bindingForm.thresholds = Array.isArray(binding.thresholds)
      ? binding.thresholds.map(t => ({ ...t }))
      : []
    // 回显设备选项并加载属性列表
    if (!deviceOptions.value.find(d => d.id === binding.device_id)) {
      deviceOptions.value = [{ id: binding.device_id, name: binding.device_id }]
    }
    onDeviceChange(binding.device_id).then(() => {
      bindingForm.property = binding.property || ''
    })
  } else {
    bindingForm.device_id = ''
    bindingForm.property = ''
    bindingForm.min = 0
    bindingForm.max = 100
    bindingForm.precision = 1
    bindingForm.unit = ''
    bindingForm.thresholds = []
    propertyOptions.value = []
  }
}

// ========== 图表配置面板 ==========
// 图表表单：标题、数据源类型、静态数据、设备/属性、历史聚合参数
const chartForm = reactive({
  title: '',
  dsType: 'static',
  staticData: [],
  device_id: '',
  property: '',
  window: '1h',
  method: 'mean',
  limit: 24
})

// 数据源类型切换
function onDsTypeChange() {
  // 切换数据源时不清空已有配置，保留用户输入
}

// 图表设备选择变化：复用绑定面板的属性加载逻辑
async function onChartDeviceChange(deviceID) {
  chartForm.property = ''
  propertyOptions.value = []
  if (!deviceID) return
  try {
    let device = deviceOptions.value.find(d => d.id === deviceID)
    if (!device || !device.model_id) {
      const res = await getDevice(deviceID)
      device = res.data
      if (device && !deviceOptions.value.find(d => d.id === deviceID)) {
        deviceOptions.value.push(device)
      }
    }
    if (!device || !device.model_id) return
    const res = await getModel(device.model_id)
    const model = res.data
    if (model && Array.isArray(model.properties)) {
      propertyOptions.value = model.properties
    }
  } catch (e) {
    console.error('load model failed', e)
  }
}

// 添加静态数据项
function addStaticItem() {
  chartForm.staticData.push({ name: '项目' + (chartForm.staticData.length + 1), value: 0 })
}

// 删除静态数据项
function removeStaticItem(idx) {
  chartForm.staticData.splice(idx, 1)
}

// 应用图表配置到选中图元
function applyChartConfig() {
  const el = selectedElement.value
  if (!el || !isChartElement(el)) return
  pushHistory()
  const chart = el.data.chart
  chart.title = chartForm.title
  chart.dataSource = {
    type: chartForm.dsType,
    staticData: chartForm.staticData.filter(i => i.name),
    device_id: chartForm.device_id,
    property: chartForm.property,
    window: chartForm.window,
    method: chartForm.method,
    limit: chartForm.limit
  }
  updateHistoryState()
  ElMessage.success('图表配置已应用')
}

// 选中图元变化时同步图表表单
function syncChartForm() {
  const el = selectedElement.value
  if (!isChartElement(el)) return
  const chart = el.data.chart
  const ds = chart.dataSource || {}
  chartForm.title = chart.title || ''
  chartForm.dsType = ds.type === 'history' ? 'history' : 'static'
  chartForm.staticData = Array.isArray(ds.staticData)
    ? ds.staticData.map(i => ({ ...i }))
    : []
  chartForm.device_id = ds.device_id || ''
  chartForm.property = ds.property || ''
  chartForm.window = ds.window || '1h'
  chartForm.method = ds.method || 'mean'
  chartForm.limit = Number(ds.limit) || 24
  // 回显设备并加载属性列表
  if (chartForm.device_id) {
    if (!deviceOptions.value.find(d => d.id === chartForm.device_id)) {
      deviceOptions.value = [{ id: chartForm.device_id, name: chartForm.device_id }]
    }
    onChartDeviceChange(chartForm.device_id).then(() => {
      chartForm.property = ds.property || ''
    })
  }
}

// ========== 管道配置面板 ==========
// 管道表单：流动颜色、流速、方向、是否启用流动动画
const pipeForm = reactive({
  flowColor: '#00d9ff',
  speed: 3,
  direction: 1,
  animated: true
})

// 应用管道配置到选中图元
function applyPipeConfig() {
  const el = selectedElement.value
  if (!el || !isPipeElement(el)) return
  pushHistory()
  Object.assign(el.data.pipe, {
    flowColor: pipeForm.flowColor,
    speed: pipeForm.speed,
    direction: pipeForm.direction,
    animated: pipeForm.animated
  })
  // 同步更新流动层颜色，便于编辑器内直观查看
  const flow = el.children && el.children.find(c => c.name === 'pipe_flow')
  if (flow) flow.stroke = pipeForm.flowColor
  updateHistoryState()
}

// 选中图元变化时同步管道表单
function syncPipeForm() {
  const el = selectedElement.value
  if (!isPipeElement(el)) return
  const pipe = el.data.pipe
  pipeForm.flowColor = pipe.flowColor || '#00d9ff'
  pipeForm.speed = Number(pipe.speed) || 3
  pipeForm.direction = pipe.direction === -1 ? -1 : 1
  pipeForm.animated = pipe.animated !== false
}

// ========== 组件库 ==========
// 组件列表与加载状态
const componentList = ref([])
const componentLoading = ref(false)

// 加载组件库列表
async function loadComponents() {
  componentLoading.value = true
  try {
    const res = await searchComponents({ page: 1, page_size: 50 })
    componentList.value = res.data || []
  } catch (e) {
    console.error('load components failed', e)
  } finally {
    componentLoading.value = false
  }
}

// 将组件添加到画布（实例化为 Group）
async function addComponent(comp) {
  try {
    // 拉取组件完整定义（含图元）
    const res = await getComponent(comp.id)
    const component = res.data
    if (!component || !Array.isArray(component.elements) || component.elements.length === 0) {
      ElMessage.warning('该组件暂无图元')
      return
    }
    pushHistory()
    // 将组件图元反序列化为 Group
    const group = new Group({ name: component.name || '组件', editable: true })
    component.elements.forEach(data => {
      const el = UI.one(data)
      if (el) { patchEditable(el); group.add(el) }
    })
    // 记录组件来源与变量/事件/脚本配置，供运行时使用
    group.data = {
      ...(group.data || {}),
      component: {
        id: component.id,
        name: component.name,
        variables: component.variables || [],
        events: component.events || [],
        scripts: component.scripts || []
      }
    }
    group.x = 120
    group.y = 120
    canvasFrame.add(group)
    app.editor.select(group)
    refreshLayers()
    updateHistoryState()
  } catch (e) {
    console.error('add component failed', e)
  }
}

// 将当前选中图元存为组件
async function saveAsComponent() {
  const list = app.editor.list
  if (!list.length) {
    ElMessage.warning('请先选中要保存为组件的图元')
    return
  }
  try {
    const { value: name } = await ElMessageBox.prompt('请输入组件名称', '存为组件', {
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputPattern: /.+/,
      inputErrorMessage: '组件名称不能为空'
    })
    // 序列化选中图元
    const elements = list.map(el => el.toJSON())
    await createComponent({
      name: name.trim(),
      elements,
      variables: [],
      events: [],
      scripts: []
    })
    ElMessage.success('已保存为组件')
    loadComponents()
  } catch (e) {
    // 用户取消或保存失败
    if (e !== 'cancel') console.error('save component failed', e)
  }
}

// ========== 画布配置 ==========
function applyCanvasConfig() {
  if (!canvasFrame) return
  // canvasFrame 是 Group，尺寸/背景色改内部的 bgRect
  const bgRect = canvasFrame.children.find(c => c.id === CANVAS_FRAME_ID + '_bg')
  if (bgRect) {
    bgRect.width = canvasConfig.width
    bgRect.height = canvasConfig.height
    bgRect.fill = canvasConfig.fill
  }
  // 画布尺寸改变时重画网格/坐标轴
  redrawDecoration()
}

// ========== 缩放 ==========
function zoomCanvas(type) {
  app.tree.zoom(type)
}

// ========== 保存 / 预览 / 返回 ==========
async function handleSave() {
  if (!scadaName.value.trim()) {
    ElMessage.warning('请输入画面名称')
    return
  }
  saving.value = true
  try {
    const payload = {
      name: scadaName.value.trim(),
      type: scadaType.value,
      canvas: { ...canvasConfig },
      elements: serializeElements(),
      updated: new Date().toISOString()
    }
    if (scadaId.value) {
      await updateScada(scadaId.value, payload)
      ElMessage.success('保存成功')
    } else {
      const res = await createScada(payload)
      scadaId.value = res.data.id
      // 替换路由，避免重复创建
      router.replace({ path: `/scada/${scadaId.value}/editor`, query: { type: scadaType.value } })
      ElMessage.success('创建成功')
    }
  } catch (e) {
    console.error('save scada failed', e)
  } finally {
    saving.value = false
  }
}

// 打开预览弹窗：直接把当前画布图元序列化后在弹窗内渲染，无需先保存
async function handlePreview() {
  previewVisible.value = true
  await nextTick()
  // 等弹窗 DOM 出来后再初始化 Leafer
  if (!previewCanvasView.value) return
  if (previewLeafer) { previewLeafer.destroy(); previewLeafer = null }

  const elements = serializeElements()
  // 如果当前画布为空，显示提示
  if (elements.length === 0) {
    ElMessage.info('画布为空，没有可预览的图元')
  }

  // 创建只读渲染的 Leafer（design 类型）
  previewLeafer = new Leafer({
    view: previewCanvasView.value,
    type: 'design',
    width: canvasConfig.width,
    height: canvasConfig.height,
    fill: canvasConfig.fill
  })

  // 反序列化并渲染所有图元
  elements.forEach(data => {
    const el = UI.one(data)
    if (el) { patchEditable(el); previewLeafer.add(el) }
  })
}

// 关闭预览弹窗：销毁 Leafer 实例释放资源
function closePreview() {
  if (previewLeafer) {
    previewLeafer.destroy()
    previewLeafer = null
  }
  previewVisible.value = false
}

function handleBack() {
  router.push('/scada')
}

// ========== 加载画面 ==========
async function loadScada() {
  if (!scadaId.value) return
  try {
    const res = await getScada(scadaId.value)
    const scada = res.data
    if (!scada) return
    scadaName.value = scada.name || '未命名画面'
    scadaType.value = scada.type || scadaType.value
    if (scada.canvas) {
      Object.assign(canvasConfig, scada.canvas)
      applyCanvasConfig()
    }
    if (Array.isArray(scada.elements)) {
      scada.elements.forEach(data => {
        const el = UI.one(data)
        if (el) { patchEditable(el); canvasFrame.add(el) }
      })
      refreshLayers()
    }
  } catch (e) {
    console.error('load scada failed', e)
  }
}

// ========== 键盘快捷键 ==========
function onKeydown(e) {
  // 输入框中不处理
  const tag = e.target?.tagName
  if (tag === 'INPUT' || tag === 'TEXTAREA' || e.target?.isContentEditable) return
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'z') {
    e.preventDefault()
    e.shiftKey ? redo() : undo()
  } else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'y') {
    e.preventDefault()
    redo()
  } else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'g') {
    // Ctrl+G 分组 / Ctrl+Shift+G 取消分组
    e.preventDefault()
    e.shiftKey ? ungroupSelected() : groupSelected()
  } else if (e.key === 'Delete' || e.key === 'Backspace') {
    e.preventDefault()
    deleteSelected()
  }
}

// ========== 初始化 ==========
onMounted(async () => {
  await nextTick()
  // 创建 Leafer 应用：ground 装饰层 / tree 设计层（图元） / sky 编辑器层
  // realCanvas: true 让 App 直接用真实 canvas 接收交互事件，避免 unrealCanvas() 把 canvas 替换成 div 导致 pointer-events: none
  app = new App({
    view: canvasView.value,
    realCanvas: true,
    // ground 装饰层：只有背景色，网格和坐标轴由 drawGrid()/drawRuler() 添加
    ground: {
      type: 'draw',
      fill: '#2b2b3c',
      hittable: false
    },
    tree: { type: 'design' },
    sky: { type: 'draw' },
    // move 配置：右键拖动画布
    move: {
      type: 'position',
      rightKey: true,
      pointGap: 0
    },
    zoom: {
      type: 'wheel',
      pointGap: 0
    },
    editor: {
      keyEvent: true,
      stroke: '#00d9ff',
      pointFill: '#ffffff',
      hover: true,
      // 单击（按下）即选中，按下即可拖动
      select: 'press',
      moveable: true,
      resizeable: true,
      rotateable: true
    }
  })

  // ========== 画布容器架构 ==========
  // canvasFrame 是一个 Group 容器，整体偏移 CANVAS_OFFSET 像素
  // 这样画面内部坐标从 (0,0) 开始，而坐标轴画在 Group 外侧 (0,0) 深灰区（可见）
  // 用户图元/网格点都是 Group 的 children（自动偏移），坐标轴画在 tree 层 Group 之外
  const CANVAS_OFFSET = 30

  // canvasFrame Group（容器，editable=false 让它不被选中，但 children 能被选）
  canvasFrame = new Group({
    id: CANVAS_FRAME_ID,
    x: CANVAS_OFFSET, y: CANVAS_OFFSET,
    editable: false
  })

  // 画面背景 Rect（Group 的 children，fill 用户设定的颜色，尺寸 = canvasConfig）
  const bgRect = new Rect({
    id: CANVAS_FRAME_ID + '_bg',
    x: 0, y: 0,
    width: canvasConfig.width,
    height: canvasConfig.height,
    fill: canvasConfig.fill,
    hittable: false,
    editable: false,
    userData: { __decoration: true }
  })
  canvasFrame.add(bgRect)
  app.tree.add(canvasFrame)

  // 网格点加到 canvasFrame Group 内部（画面可见区域）
  drawGrid()
  // 坐标轴画在 tree 层 Group 外侧（画面外深灰色区域，(0,0) 附近可见）
  drawRuler()

  // 选中事件：同步属性面板与图层列表（选中和取消选中都走这个事件，value=null 就是取消选中）
  app.editor.on(EditorEvent.SELECT, (e) => {
    const val = e.value
    if (!val) {
      selectedElement.value = null
    } else if (Array.isArray(val)) {
      selectedElement.value = val.length === 1 ? val[0] : null
    } else {
      selectedElement.value = val
    }
    console.log('[SELECT] value=', val?.tag || val, 'selected=', selectedElement.value?.tag)
    if (selectedElement.value) {
      syncPropsForm()
      syncBindingForm()
      syncChartForm()
      syncPipeForm()
    }
    refreshLayers()
  })

  // 拖拽/缩放/旋转：开始时记录快照，首次变化时压入撤销栈
  const onTransformStart = () => {
    pendingSnapshot = snapshot()
  }
  const onTransform = () => {
    if (pendingSnapshot) {
      undoStack.push(pendingSnapshot)
      if (undoStack.length > HISTORY_LIMIT) undoStack.shift()
      redoStack = []
      pendingSnapshot = null
      updateHistoryState()
    }
    // 实时同步属性面板数值
    if (selectedElement.value) syncPropsForm()
  }
  app.editor.on(EditorMoveEvent.BEFORE_MOVE, onTransformStart)
  app.editor.on(EditorScaleEvent.BEFORE_SCALE, onTransformStart)
  app.editor.on(EditorRotateEvent.BEFORE_ROTATE, onTransformStart)
  app.editor.on(EditorMoveEvent.MOVE, onTransform)
  app.editor.on(EditorScaleEvent.SCALE, onTransform)
  app.editor.on(EditorRotateEvent.ROTATE, onTransform)

  window.addEventListener('keydown', onKeydown)

  // 加载已有画面
  await loadScada()
  // 加载组件库列表
  loadComponents()
  // 初始视图适应画布
  setTimeout(() => app.tree.zoom('fit'), 100)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
  if (previewLeafer) { previewLeafer.destroy(); previewLeafer = null }
  if (app) { app.destroy(); app = null }
})
</script>

<style lang="scss" scoped>
.scada-editor {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 130px);
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
}

.editor-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 48px;
  padding: 0 12px;
  border-bottom: 1px solid #e6e8eb;
  flex-shrink: 0;

  .toolbar-left, .toolbar-center, .toolbar-right {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .name-input {
    width: 180px;
  }

  .rotate90 { transform: rotate(-90deg); }
  .rotate270 { transform: rotate(90deg); }

  .canvas-toggles {
    display: flex;
    align-items: center;
    gap: 2px;
    :deep(.el-button.active) {
      color: #409EFF;
      background: #ecf5ff;
    }
  }
}

.editor-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.panel-left {
  width: 240px;
  border-right: 1px solid #e6e8eb;
  padding: 4px;
  overflow-y: auto;
  flex-shrink: 0;

  // el-collapse 全局样式
  .left-collapse {
    border: none;
    --el-collapse-header-height: 34px;
    --el-collapse-content-box-padding: 8px 0 12px;

    :deep(.el-collapse-item__header) {
      font-size: 13px;
      font-weight: 600;
      color: #303133;
      border-bottom: 1px solid #ebeef5;
    }
    :deep(.el-collapse-item__content) {
      border-bottom: 1px solid #ebeef5;
    }
    :deep(.el-collapse-item__wrap) {
      border-bottom: none;
    }
    :deep(.el-collapse) {
      border: none;
    }
    :deep(.el-collapse-item__header .el-collapse-item__arrow) {
      font-size: 12px;
    }
  }

  .collapse-title {
    display: flex; align-items: center; gap: 8px;
  }

  .element-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 8px;
  }

  .element-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 8px 4px;
    border: 1px solid #e6e8eb;
    border-radius: 6px;
    cursor: pointer;
    color: #606266;
    transition: all 0.2s;

    &:hover {
      border-color: #00d9ff;
      color: #00a8c6;
      background: #f0fbff;
    }

    .element-name { font-size: 12px; }
  }

  .canvas-setting {
    .setting-row {
      display: flex; align-items: center; justify-content: space-between;
      margin-bottom: 8px; gap: 8px;
      span { font-size: 12px; color: #606266; flex-shrink: 0; }
    }
  }

  .comp-group {
    margin-bottom: 10px;
    &-title {
      font-size: 11px;
      font-weight: 600;
      color: #909399;
      padding: 2px 0 4px;
      border-bottom: 1px dashed #ebeef5;
      margin-bottom: 4px;
    }
  }

  .component-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;

    .component-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 4px;
      padding: 6px 4px;
      border: 1px solid #e6e8eb;
      border-radius: 6px;
      cursor: pointer;
      color: #606266;
      transition: all 0.2s;

      &:hover {
        border-color: #00d9ff;
        color: #00a8c6;
        background: #f0fbff;
      }

      .comp-thumb {
        width: 100%;
        height: 42px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: #f5f7fa;
        border-radius: 4px;
        overflow: hidden;

        img { width: 100%; height: 100%; object-fit: cover; }
        svg { width: 22px; height: 22px; color: #c0c4cc; }
      }

      .component-name {
        font-size: 11px;
        overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
        max-width: 100%;
      }
    }

    .component-empty {
      grid-column: 1 / -1;
      font-size: 12px; color: #c0c4cc;
      text-align: center; padding: 12px 0;
    }
  }
}

.panel-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.component-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  margin-bottom: 16px;

  .component-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 8px 4px;
    border: 1px solid #e6e8eb;
    border-radius: 6px;
    cursor: pointer;
    color: #606266;
    transition: all 0.2s;

    &:hover {
      border-color: #00d9ff;
      color: #00a8c6;
      background: #f0fbff;
    }

    .component-name {
      font-size: 12px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      max-width: 100%;
    }
  }

  .component-empty {
    grid-column: 1 / -1;
    font-size: 12px;
    color: #c0c4cc;
    text-align: center;
    padding: 12px 0;
  }
}

.canvas-setting, .props-form {
  .setting-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
    gap: 8px;

    > span {
      font-size: 12px;
      color: #909399;
      white-space: nowrap;
      width: 48px;
    }

    .el-input-number {
      width: 130px;
    }

    .el-input, .el-slider {
      flex: 1;
    }
  }
}

.binding-form {
  .setting-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
    gap: 8px;

    > span {
      font-size: 12px;
      color: #909399;
      white-space: nowrap;
      width: 48px;
    }

    .el-select {
      flex: 1;
    }

    .el-input-number {
      width: 130px;
    }

    .el-input {
      flex: 1;
    }
  }

  .threshold-section {
    margin-top: 12px;
    padding-top: 10px;
    border-top: 1px dashed #e6e8eb;

    .threshold-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      font-size: 12px;
      color: #606266;
      margin-bottom: 8px;
    }

    .threshold-row {
      display: flex;
      align-items: center;
      gap: 6px;
      margin-bottom: 6px;

      .threshold-op {
        font-size: 12px;
        color: #909399;
      }

      .el-input-number {
        width: 90px;
      }

      .threshold-del {
        color: #f56c6c;
        cursor: pointer;
      }
    }

    .threshold-empty {
      font-size: 12px;
      color: #c0c4cc;
    }
  }

  .binding-actions {
    margin-top: 14px;
    display: flex;
    gap: 8px;
  }

  .binding-status {
    margin-top: 10px;
    font-size: 12px;
    color: #00a8c6;
    background: #e8f9ff;
    padding: 6px 8px;
    border-radius: 4px;
    word-break: break-all;
  }
}

.chart-form {
  .setting-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
    gap: 8px;

    > span {
      font-size: 12px;
      color: #909399;
      white-space: nowrap;
      width: 48px;
    }

    .el-select {
      flex: 1;
    }

    .el-input-number {
      width: 130px;
    }

    .el-input {
      flex: 1;
    }
  }

  .static-section {
    margin-top: 8px;

    .static-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      font-size: 12px;
      color: #606266;
      margin-bottom: 8px;
    }

    .static-row {
      display: flex;
      align-items: center;
      gap: 6px;
      margin-bottom: 6px;

      .el-input {
        width: 90px;
      }

      .el-input-number {
        width: 90px;
      }

      .static-del {
        color: #f56c6c;
        cursor: pointer;
      }
    }

    .static-empty {
      font-size: 12px;
      color: #c0c4cc;
    }
  }

  .chart-actions {
    margin-top: 14px;
  }
}

.pipe-form {
  .setting-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
    gap: 8px;

    > span {
      font-size: 12px;
      color: #909399;
      white-space: nowrap;
      width: 60px;
    }

    .el-slider {
      flex: 1;
      padding: 0 6px;
    }
  }

  .pipe-tip {
    margin-top: 10px;
    font-size: 12px;
    color: #c0c4cc;
    line-height: 1.6;
  }
}

.canvas-wrapper {
  flex: 1;
  overflow: hidden;
  background: #2b2b3c;
  position: relative;

  .canvas-view {
    width: 100%;
    height: 100%;
  }
}

.panel-right {
  width: 260px;
  border-left: 1px solid #e6e8eb;
  padding: 0 12px;
  overflow-y: auto;
  flex-shrink: 0;

  .right-collapse {
    border: none;

    :deep(.el-collapse-item__header) {
      font-size: 13px;
      font-weight: 600;
      color: #303133;
      height: 36px;
      line-height: 36px;
      border-bottom: 1px solid #ebeef5;
    }
    :deep(.el-collapse-item__content) {
      border-bottom: 1px solid #ebeef5;
    }
    :deep(.el-collapse-item__wrap) { border-bottom: none; }
    :deep(.el-collapse) { border: none; }
  }

  .empty-tip {
    text-align: center;
    color: #909399;
    font-size: 13px;
    padding: 30px 0;
  }
}

.layer-list {
  .layer-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 10px;
    border-radius: 6px;
    cursor: pointer;
    font-size: 13px;
    color: #606266;

    &:hover {
      background: #f5f7fa;
    }

    &.active {
      background: #e8f9ff;
      color: #00a8c6;
    }

    .layer-name {
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .layer-del {
      display: none;
      color: #f56c6c;
    }

    &:hover .layer-del {
      display: block;
    }
  }
}

// 预览弹窗（el-dialog 是全局组件，这里用 :deep 或非 scoped）
.preview-dialog-body {
  height: 68vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #1a1a2e;
  border-radius: 4px;
  overflow: hidden;
}

.preview-canvas-holder {
  // Leafer 画布会自动填满父容器
  width: 100%;
  height: 100%;
}
</style>
