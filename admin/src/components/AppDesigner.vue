<template>
  <div class="designer-container">
    <div class="designer-header">
      <span class="header-title">App 界面设计器</span>
      <div class="header-actions">
        <el-button size="small" @click="handleReset">重置</el-button>
        <el-button size="small" type="primary" @click="handleSave">保存</el-button>
      </div>
    </div>
    
    <div class="designer-body">
      <div class="panel panel-left">
        <div class="panel-header">控件库</div>
        <div class="panel-content">
          <div 
            v-for="item in controlTypes" 
            :key="item.type"
            class="control-item"
            @click="addControl(item.type)"
          >
            <span class="control-name">{{ item.label }}</span>
            <div class="control-preview">
              <component 
                :is="getControlComponent(item.type)" 
                :control="getPreviewControl(item.type)"
                :value="item.type === 'switch' ? false : 50"
              />
            </div>
          </div>
        </div>
      </div>
      
      <div class="panel panel-center">
        <div class="panel-header">预览区</div>
        <div class="preview-area" ref="previewRef">
          <div 
            v-if="controls.length === 0" 
            class="empty-state"
          >
            <span class="empty-icon">👆</span>
            <span>请从左侧选择控件添加</span>
          </div>
          
          <div class="controls-grid" @dragover.prevent @drop="handleDrop">
            <div 
              v-for="(control, index) in controls" 
              :key="control.id"
              :data-id="control.id"
              class="control-wrapper"
              :class="{ 
                selected: selectedControl?.id === control.id,
                dragging: draggingIndex === index
              }"
              :style="getControlStyle(control)"
              draggable="true"
              @dragstart="handleDragStart(index)"
              @dragend="handleDragEnd"
              @click="selectControl(control)"
            >
              <component 
                :is="getControlComponent(control.type)" 
                :control="control"
                :value="getControlValue(control)"
                @change="handleControlChange"
                @action="handleControlAction"
              />
              <div class="control-delete" @click.stop="deleteControl(control.id)">✕</div>
              <div class="drag-handle">⋮⋮</div>
            </div>
          </div>
        </div>
      </div>
      
      <div class="panel panel-right">
        <div class="panel-header">属性设置</div>
        <div class="panel-content">
          <template v-if="selectedControl">
            <el-form :model="selectedControl" label-width="80px" size="small">
              <template v-for="prop in currentConfig.properties" :key="prop">
                <el-form-item :label="getPropLabel(prop)">
                  <el-input v-model="selectedControl[prop]" v-if="prop === 'name' || prop === 'unit' || prop === 'url'" :placeholder="getPropPlaceholder(prop)" />
                  <el-input v-model="selectedControl[prop]" type="textarea" :rows="3" v-else-if="prop === 'text'" placeholder="请输入文本内容" />
                  <el-select v-model="selectedControl[prop]" placeholder="选择占列" v-else-if="prop === 'span'">
                    <el-option label="1列" :value="1" />
                    <el-option label="2列" :value="2" />
                  </el-select>
                  <el-input-number controls-position="right" v-model="selectedControl[prop]" v-else-if="prop === 'min'" />
                  <el-input-number controls-position="right" v-model="selectedControl[prop]" v-else-if="prop === 'max'" />
                  <el-input-number controls-position="right" v-model="selectedControl[prop]" v-else-if="prop === 'step'" />
                  <el-input-number controls-position="right" v-model="selectedControl[prop]" :min="0" :max="6" v-else-if="prop === 'precision'" />
                  <div v-else-if="prop === 'images'" class="images-editor">
                    <div v-for="(img, idx) in selectedControl.images" :key="idx" class="image-item">
                      <el-input v-model="selectedControl.images[idx]" placeholder="图片地址" size="small" />
                      <el-button type="danger" size="small" icon="Delete" circle @click="removeImage(idx)" />
                    </div>
                    <el-button size="small" type="primary" plain @click="addImage">+ 添加图片</el-button>
                  </div>
                </el-form-item>
              </template>
              
              <template v-if="currentConfig.inputs && currentConfig.inputs.length > 0">
                <el-divider content-position="left">数值绑定</el-divider>
                <div v-for="input in currentConfig.inputs" :key="input.name" class="bind-item">
                  <span class="bind-label">{{ input.label }}</span>
                  <el-select v-model="selectedControl.binds[input.name]" placeholder="选择属性" size="small" class="bind-key">
                    <el-option 
                      v-for="item in properties" 
                      :key="item.key" 
                      :label="item.name" 
                      :value="item.key"
                    />
                  </el-select>
                </div>
              </template>
              
              <template v-if="currentConfig.events && currentConfig.events.length > 0">
                <el-divider content-position="left">事件绑定</el-divider>
                <div v-for="evt in currentConfig.events" :key="evt.name" class="bind-item">
                  <span class="bind-label">{{ evt.label }}</span>
                  <el-select v-model="selectedControl.events[evt.name].type" placeholder="类型" size="small" style="width: 80px">
                    <el-option label="修改属性" value="property" />
                    <el-option label="调用服务" value="service" />
                  </el-select>
                  <el-select 
                    v-if="selectedControl.events[evt.name].type === 'property'" 
                    v-model="selectedControl.events[evt.name].key" 
                    placeholder="选择属性" size="small" class="bind-key"
                  >
                    <el-option 
                      v-for="item in properties" 
                      :key="item.key" 
                      :label="item.name" 
                      :value="item.key"
                    />
                  </el-select>
                  <template v-if="selectedControl.events[evt.name].type === 'property' && selectedControl.events[evt.name].key">
                    <div class="event-params">
                      <div class="event-param-row">
                        <span class="param-label">设置值</span>
                        <el-select v-model="getEventParam(evt.name, 'value').source" placeholder="选择值来源" size="small" class="param-select">
                          <el-option label="固定值" value="fixed" />
                          <el-option 
                            v-for="output in (currentConfig.outputs || [])" 
                            :key="output.name" 
                            :label="output.label" 
                            :value="output.name"
                          />
                        </el-select>
                        <el-input 
                          v-if="getEventParam(evt.name, 'value').source === 'fixed'" 
                          v-model="getEventParam(evt.name, 'value').value" 
                          placeholder="输入固定值" 
                          size="small" 
                          class="param-fixed-input"
                        />
                      </div>
                    </div>
                  </template>
                  <el-select 
                    v-if="selectedControl.events[evt.name].type === 'service'" 
                    v-model="selectedControl.events[evt.name].key" 
                    placeholder="选择服务" size="small" class="bind-key"
                  >
                    <el-option 
                      v-for="item in services" 
                      :key="item.key" 
                      :label="item.name" 
                      :value="item.key"
                    />
                  </el-select>
                  <template v-if="selectedControl.events[evt.name].type === 'service' && selectedControl.events[evt.name].key && getServiceParams(selectedControl.events[evt.name].key).length > 0">
                    <div class="event-params">
                      <div v-for="sparam in getServiceParams(selectedControl.events[evt.name].key)" :key="sparam.key" class="event-param-row">
                        <span class="param-label">{{ sparam.name || sparam.key }}</span>
                        <el-select v-model="getEventParam(evt.name, sparam.key).source" placeholder="选择参数来源" size="small" class="param-select">
                          <el-option label="固定值" value="fixed" />
                          <el-option 
                            v-for="output in (currentConfig.outputs || [])" 
                            :key="output.name" 
                            :label="output.label" 
                            :value="output.name"
                          />
                        </el-select>
                        <el-input 
                          v-if="getEventParam(evt.name, sparam.key).source === 'fixed'" 
                          v-model="getEventParam(evt.name, sparam.key).value" 
                          placeholder="输入固定值" 
                          size="small" 
                          class="param-fixed-input"
                        />
                      </div>
                    </div>
                  </template>
                </div>
              </template>
              
              <el-divider content-position="left">样式设置</el-divider>
              
              <div class="style-sections">
                <div 
                  v-for="part in currentConfig.parts" 
                  :key="part.name"
                  class="style-section"
                >
                  <div 
                    class="style-section-header" 
                    @click="toggleStyleSection(part.name)"
                  >
                    <span class="style-section-title">{{ part.label }}</span>
                    <span class="style-section-arrow" :class="{ expanded: expandedStyleSections.includes(part.name) }">▼</span>
                  </div>
                  <div 
                    class="style-section-content" 
                    v-show="expandedStyleSections.includes(part.name)"
                  >
                    <template v-for="styleKey in part.styles" :key="styleKey">
                      <el-form-item :label="getStyleLabel(styleKey)">
                        <template v-if="styleKey === 'icon'">
                          <div class="icon-picker-input">
                            <el-input 
                              :model-value="selectedControl?.styles?.[part.name]?.[styleKey]" 
                              placeholder="选择图标"
                            />
                            <el-button 
                              type="primary" 
                              size="small"
                              @click="handleIconPick(part.name, styleKey)"
                            >
                              选择
                            </el-button>
                          </div>
                        </template>
                        <el-color-picker v-model="selectedControl.styles[part.name][styleKey]" v-else-if="isColorStyle(styleKey)" />
                        <el-input v-model="selectedControl.styles[part.name][styleKey]" v-else />
                      </el-form-item>
                    </template>
                  </div>
                </div>
              </div>
              
              <template v-if="currentConfig.options && currentConfig.options.length > 0">
                <el-divider content-position="left">选项设置</el-divider>
                
                <div v-for="(option, index) in selectedControl.options" :key="index" class="option-item">
                  <el-input 
                    v-model="option.label" 
                    placeholder="选项标签" 
                    size="small"
                    class="option-input"
                  />
                  <el-input 
                    v-model="option.value" 
                    placeholder="选项值" 
                    size="small"
                    class="option-input"
                  />
                  <div class="option-icon-picker option-input">
                    <i v-if="option.icon" class="iotfont option-icon-preview" :class="`iot-${option.icon}`"></i>
                    <el-input 
                      v-model="option.icon" 
                      placeholder="图标" 
                      size="small"
                      readonly
                      @click="openIconPicker({ setValue: (v) => { option.icon = v } })"
                    />
                    <el-button 
                      size="small" 
                      type="primary"
                      @click="openIconPicker({ setValue: (v) => { option.icon = v } })"
                    >
                      选
                    </el-button>
                  </div>
                  <el-button size="small" type="danger" @click="removeOption(index)" v-if="selectedControl.options.length > 1">删除</el-button>
                </div>
                
                <el-button size="small" @click="addOption">添加选项</el-button>
              </template>
            </el-form>
          </template>
          
          <template v-else>
            <div class="empty-tip">
              <span class="empty-icon">👆</span>
              <span>请选择一个控件</span>
            </div>
          </template>
        </div>
      </div>
    </div>
    
    <el-drawer
      v-model="iconDrawerVisible"
      title="选择图标"
      size="500px"
    >
      <div class="icon-picker-drawer">
        <el-input
          v-model="iconSearchKeyword"
          placeholder="搜索图标..."
          clearable
          class="icon-search-input"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        
        <div class="icon-grid">
          <div 
            v-for="icon in filteredIcons" 
            :key="icon.font_class"
            class="icon-grid-item"
            :title="icon.name"
            @click="selectIcon(icon.font_class)"
          >
            <i class="iotfont" :class="`iot-${icon.font_class}`"></i>
            <span class="icon-name">{{ icon.name }}</span>
          </div>
        </div>
        
        <div v-if="filteredIcons.length === 0" class="icon-empty">
          未找到匹配的图标
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, reactive, watch, computed } from 'vue'
import { controlTypes, getControlComponent, getControlConfig, createControl } from '../controls'
import { ElMessage } from 'element-plus'
import iotfontData from '../iotfont/iotfont.json'
import { Search } from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: {
    type: Array,
    default: () => []
  },
  properties: {
    type: Array,
    default: () => []
  },
  services: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:modelValue', 'save'])

const controls = ref(JSON.parse(JSON.stringify(props.modelValue)))

watch(() => props.modelValue, (newVal) => {
  controls.value = JSON.parse(JSON.stringify(newVal))
  selectedControl.value = null
  expandedStyleSections.value = []
}, { deep: true })

const selectedControl = ref(null)
const controlValues = reactive({})
const draggingIndex = ref(-1)
const dropTargetIndex = ref(-1)
const expandedStyleSections = ref([])

const iconDrawerVisible = ref(false)
const iconPickerTarget = ref(null)
const iconSearchKeyword = ref('')

const filteredIcons = computed(() => {
  const icons = iotfontData.glyphs || []
  if (!iconSearchKeyword.value) return icons
  const kw = iconSearchKeyword.value.toLowerCase()
  return icons.filter(item => 
    item.name.toLowerCase().includes(kw) || 
    item.font_class.toLowerCase().includes(kw)
  )
})

const handleIconPick = (partName, styleKey) => {
  if (!selectedControl.value) return
  if (!selectedControl.value.styles) {
    selectedControl.value.styles = {}
  }
  if (!selectedControl.value.styles[partName]) {
    selectedControl.value.styles[partName] = {}
  }
  openIconPicker({ 
    setValue: (v) => { 
      selectedControl.value.styles[partName][styleKey] = v 
    } 
  })
}

const openIconPicker = (target) => {
  iconPickerTarget.value = target
  iconSearchKeyword.value = ''
  iconDrawerVisible.value = true
}

const selectIcon = (iconName) => {
  if (iconPickerTarget.value && iconPickerTarget.value.setValue) {
    iconPickerTarget.value.setValue(iconName)
  }
  iconDrawerVisible.value = false
}

const currentConfig = computed(() => {
  if (!selectedControl.value) return getControlConfig('number')
  return getControlConfig(selectedControl.value.type)
})

const getPropLabel = (prop) => {
  const labels = {
    name: '控件名称',
    span: '占列数',
    min: '最小值',
    max: '最大值',
    step: '步长',
    unit: '单位',
    text: '文本内容',
    precision: '小数精度',
    url: '视频地址',
    images: '图片列表'
  }
  return labels[prop] || prop
}

const getPropPlaceholder = (prop) => {
  const placeholders = {
    name: '请输入名称',
    unit: '请输入单位',
    url: '请输入视频URL地址'
  }
  return placeholders[prop] || ''
}

const addImage = () => {
  if (!selectedControl.value.images) {
    selectedControl.value.images = []
  }
  selectedControl.value.images.push('')
}

const removeImage = (index) => {
  if (selectedControl.value.images) {
    selectedControl.value.images.splice(index, 1)
  }
}

const getStyleLabel = (styleKey) => {
  const labels = {
    icon: '图标',
    color: '颜色',
    background_color: '背景色',
    font_size: '字体大小',
    border: '边框',
    border_radius: '圆角',
    width: '宽度',
    height: '高度',
    padding: '内边距',
    line_width: '线条宽度',
    border_color: '边框颜色',
    high_color: '高值颜色',
    medium_color: '中值颜色',
    low_color: '低值颜色'
  }
  return labels[styleKey] || styleKey
}

const isColorStyle = (styleKey) => {
  return styleKey.includes('color') && styleKey !== 'font_size' && styleKey !== 'border'
}

const toggleStyleSection = (partName) => {
  const index = expandedStyleSections.value.indexOf(partName)
  if (index > -1) {
    expandedStyleSections.value.splice(index, 1)
  } else {
    expandedStyleSections.value.push(partName)
  }
}

const getPreviewControl = (type) => {
  return createControl(type)
}

const addControl = (type) => {
  const newControl = createControl(type)
  controls.value.push(newControl)
  selectedControl.value = newControl
  expandedStyleSections.value = currentConfig.value.parts.map(p => p.name)
}

const selectControl = (control) => {
  selectedControl.value = control
  
  const config = getControlConfig(control.type)
  if (!control.binds) control.binds = {}
  if (!control.events) control.events = {}
  if (!control.styles) control.styles = {}
  
  const inputs = Array.isArray(config.inputs) ? config.inputs : []
  inputs.forEach(p => {
    if (control.binds[p.name] === undefined || control.binds[p.name] === null) {
      control.binds[p.name] = ''
    } else if (typeof control.binds[p.name] === 'object') {
      control.binds[p.name] = control.binds[p.name].key || ''
    }
  })
  
  const eventList = Array.isArray(config.events) ? config.events : []
  eventList.forEach(e => {
    if (!control.events[e.name]) {
      control.events[e.name] = { type: '', key: '', params: {} }
    }
    const evt = control.events[e.name]
    if (!evt.type) evt.type = ''
    if (!evt.key) evt.key = ''
    if (!evt.params) evt.params = {}
    if (typeof evt.params === 'object' && !Array.isArray(evt.params)) {
      for (const paramKey of Object.keys(evt.params)) {
        const val = evt.params[paramKey]
        if (typeof val === 'string') {
          if (val.startsWith('$output:')) {
            evt.params[paramKey] = { source: val.substring(8), value: '' }
          } else {
            evt.params[paramKey] = { source: 'fixed', value: val }
          }
        }
      }
    }
  })
  
  const parts = Array.isArray(config.parts) ? config.parts : []
  parts.forEach(part => {
    if (!control.styles[part.name]) {
      control.styles[part.name] = config.defaultStyles && config.defaultStyles[part.name]
        ? { ...config.defaultStyles[part.name] }
        : {}
    }
  })
  
  expandedStyleSections.value = parts.map(p => p.name)
}

const deleteControl = (id) => {
  const index = controls.value.findIndex(c => c.id === id)
  if (index > -1) {
    controls.value.splice(index, 1)
    if (selectedControl.value?.id === id) {
      selectedControl.value = null
    }
  }
}

const addOption = () => {
  selectedControl.value.options.push({ label: '', value: '', icon: '' })
}

const removeOption = (index) => {
  selectedControl.value.options.splice(index, 1)
}

const getControlValue = (control) => {
  return controlValues[control.id] || (control.type === 'switch' ? false : 50)
}

const handleControlChange = (event) => {
  if (event.value !== undefined) {
    controlValues[event.control.id] = event.value
  }
}

const handleControlAction = (event) => {
  emit('action', event)
}

const getServiceParams = (serviceKey) => {
  const service = props.services.find(s => s.key === serviceKey)
  return (service && service.params) || []
}

const getEventParam = (evtName, paramKey) => {
  if (!selectedControl.value.events[evtName]) {
    selectedControl.value.events[evtName] = { type: '', key: '', params: {} }
  }
  if (!selectedControl.value.events[evtName].params) {
    selectedControl.value.events[evtName].params = {}
  }
  if (!selectedControl.value.events[evtName].params[paramKey]) {
    selectedControl.value.events[evtName].params[paramKey] = { source: '', value: '' }
  }
  return selectedControl.value.events[evtName].params[paramKey]
}

const getControlStyle = (control) => {
  const span = control.span || 1
  if (span === 2) {
    return {
      flex: '1 1 100%',
      minWidth: '100%'
    }
  }
  return {
    flex: '1 1 calc(50% - 6px)',
    minWidth: 'calc(50% - 6px)'
  }
}

const handleDragStart = (index) => {
  draggingIndex.value = index
}

const handleDragEnd = () => {
  draggingIndex.value = -1
  dropTargetIndex.value = -1
}

const handleDrop = (event) => {
  if (draggingIndex.value === -1) return
  
  const target = event.target.closest('.control-wrapper')
  if (!target) {
    draggingIndex.value = -1
    return
  }
  
  const targetId = target.getAttribute('data-id')
  const targetIndex = controls.value.findIndex(c => c.id === targetId)
  
  if (targetIndex === -1 || targetIndex === draggingIndex.value) {
    draggingIndex.value = -1
    return
  }
  
  const [draggedItem] = controls.value.splice(draggingIndex.value, 1)
  
  if (draggingIndex.value < targetIndex) {
    controls.value.splice(targetIndex, 0, draggedItem)
  } else {
    controls.value.splice(targetIndex, 0, draggedItem)
  }
  
  draggingIndex.value = -1
}

const handleReset = () => {
  controls.value = []
  selectedControl.value = null
  emit('update:modelValue', [])
  ElMessage.success('已重置')
}

const handleSave = () => {
  const outputControls = controls.value.map(control => {
    const controlCopy = JSON.parse(JSON.stringify(control))
    if (controlCopy.events) {
      for (const evtName of Object.keys(controlCopy.events)) {
        const evt = controlCopy.events[evtName]
        if (evt.params && typeof evt.params === 'object' && !Array.isArray(evt.params)) {
          const paramMap = {}
          for (const [paramKey, paramConfig] of Object.entries(evt.params)) {
            if (paramConfig && paramConfig.source) {
              if (paramConfig.source === 'fixed') {
                paramMap[paramKey] = paramConfig.value || ''
              } else {
                paramMap[paramKey] = '$output:' + paramConfig.source
              }
            }
          }
          evt.params = paramMap
        }
      }
    }
    return controlCopy
  })
  emit('update:modelValue', outputControls)
  emit('save', outputControls)
  ElMessage.success('保存成功')
}
</script>

<style lang="scss" scoped>
.designer-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #f5f5f5;

  .designer-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 20px;
    background: #fff;
    border-bottom: 1px solid #eee;

    .header-title {
      font-size: 16px;
      font-weight: 600;
    }

    .header-actions {
      display: flex;
      gap: 8px;
    }
  }

  .designer-body {
    flex: 1;
    display: flex;
    gap: 12px;
    padding: 12px;
    overflow: hidden;

    .panel {
      background: #fff;
      border-radius: 8px;
      display: flex;
      flex-direction: column;
      overflow: hidden;

      .panel-header {
        padding: 12px 16px;
        font-size: 14px;
        font-weight: 600;
        border-bottom: 1px solid #eee;
        background: #fafafa;
      }

      .panel-content {
        flex: 1;
        overflow-y: auto;
        padding: 12px;
      }
    }

    .panel-left {
      width: 280px;

      .control-item {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 6px;
        padding: 12px;
        cursor: pointer;
        border-radius: 8px;
        transition: all 0.2s;

        &:hover {
          background: #f0f7ff;
        }

        .control-preview {
          width: 100%;
          padding: 8px;
          background: #f8f9fa;
          border-radius: 6px;
          display: flex;
          justify-content: center;
          align-items: center;

          & > * {
            width: 100%;
          }
        }

        .control-name {
          font-size: 12px;
          color: #666;
          text-align: center;
        }
      }
    }

    .panel-center {
      flex: 1;

      .preview-area {
        padding: 20px;
        min-height: 400px;
        background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
        border-radius: 8px;
        overflow-y: auto;

        .empty-state {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          height: 300px;
          color: #999;

          .empty-icon {
            font-size: 48px;
            margin-bottom: 12px;
          }
        }

        .controls-grid {
          display: flex;
          flex-wrap: wrap;
          gap: 12px;
        }

        .control-wrapper {
          position: relative;
          padding: 12px;
          background: #fff;
          border-radius: 8px;
          transition: all 0.2s;
          cursor: move;
          border: 2px solid transparent;

          &.selected {
            border-color: #5BA3E8;
            background: rgba(91, 163, 232, 0.05);
          }

          &.dragging {
            opacity: 0.5;
            transform: scale(0.98);
          }

          .control-delete {
            position: absolute;
            top: -8px;
            right: -8px;
            width: 20px;
            height: 20px;
            background: #e74c3c;
            color: #fff;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 12px;
            cursor: pointer;
            opacity: 0;
            transition: opacity 0.2s;
          }

          .drag-handle {
            position: absolute;
            top: 4px;
            right: 4px;
            font-size: 14px;
            color: #ccc;
            cursor: grab;
            opacity: 0;
            transition: opacity 0.2s;
          }

          &:hover {
            .control-delete,
            .drag-handle {
              opacity: 1;
            }
          }

          .drag-handle:active {
            cursor: grabbing;
          }
        }
      }
    }

    .panel-right {
      width: 320px;

      .empty-tip {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 200px;
        color: #999;

        .empty-icon {
          font-size: 48px;
          margin-bottom: 12px;
        }
      }

      .option-item {
        display: flex;
        gap: 6px;
        margin-bottom: 8px;
        align-items: center;

        .option-input {
          flex: 1;
        }

        .option-icon-picker {
          display: flex;
          align-items: center;
          gap: 4px;

          .option-icon-preview {
            font-size: 16px;
            flex-shrink: 0;
          }

          .el-input {
            flex: 1;
            cursor: pointer;

            :deep(.el-input__wrapper) {
              cursor: pointer;
            }
          }
        }
      }

      .bind-item {
        display: flex;
        gap: 6px;
        margin-bottom: 8px;
        align-items: center;
        flex-wrap: wrap;

        .bind-label {
          font-size: 12px;
          color: #666;
          width: 60px;
        }

        .bind-key {
          flex: 1;
        }

        .event-params {
          width: 100%;
          padding-left: 60px;

          .event-param-row {
            display: flex;
            align-items: center;
            gap: 6px;
            margin-top: 4px;

            .param-label {
              font-size: 12px;
              color: #999;
              width: 70px;
            }

            .param-select {
              flex: 1;
            }

            .param-fixed-input {
              width: 100px;
            }
          }
        }
      }

      .images-editor {
        width: 100%;

        .image-item {
          display: flex;
          gap: 6px;
          margin-bottom: 6px;
          align-items: center;

          .el-input {
            flex: 1;
          }
        }
      }

      .icon-picker-input {
        display: flex;
        align-items: center;
        gap: 6px;
        width: 100%;

        .icon-preview {
          width: 28px;
          height: 28px;
          display: flex;
          align-items: center;
          justify-content: center;
          background: #f5f5f5;
          border-radius: 4px;
          flex-shrink: 0;

          .iotfont {
            font-size: 18px;
          }
        }

        .el-input {
          flex: 1;
          cursor: pointer;

          :deep(.el-input__wrapper) {
            cursor: pointer;
          }
        }
      }

      .style-sections {
        .style-section {
          margin-bottom: 8px;
          border: 1px solid #eee;
          border-radius: 6px;
          overflow: hidden;

          .style-section-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 8px 12px;
            background: #fafafa;
            cursor: pointer;

            .style-section-title {
              font-size: 13px;
              font-weight: 500;
              color: #333;
            }

            .style-section-arrow {
              font-size: 10px;
              color: #999;
              transition: transform 0.2s;

              &.expanded {
                transform: rotate(180deg);
              }
            }
          }

          .style-section-content {
            padding: 8px 12px;
            background: #fff;
          }
        }
      }
    }
  }
}

.icon-picker-drawer {
  padding: 12px;

  .icon-search-input {
    margin-bottom: 16px;
  }

  .icon-grid {
    display: grid;
    grid-template-columns: repeat(5, 1fr);
    gap: 12px;

    .icon-grid-item {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: 12px 8px;
        border: 1px solid #eee;
        border-radius: 8px;
        cursor: pointer;
        transition: all 0.2s;

        .iotfont {
          font-size: 24px;
          color: #333;
          margin-bottom: 6px;
        }

        .icon-name {
          font-size: 11px;
          color: #999;
          text-align: center;
          line-height: 1.2;
          word-break: break-all;
        }

        &:hover {
          border-color: #5BA3E8;
          background: #f0f7ff;

          .iotfont {
            color: #5BA3E8;
          }
        }
      }
  }

  .icon-empty {
    text-align: center;
    color: #999;
    padding: 40px 0;
    font-size: 14px;
  }
}
</style>
