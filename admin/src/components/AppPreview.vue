<template>
  <div class="app-preview-container">
    
    <div class="phone-frame">
      <div class="phone-header">
        <div class="status-bar">
          <span class="signal">📶</span>
          <span class="time">9:41</span>
          <span class="battery">🔋</span>
        </div>
        <div class="nav-bar">
          <span class="nav-back">←</span>
          <span class="nav-title">{{ title }}</span>
          <span class="nav-more">⋮</span>
        </div>
      </div>
      
      <div class="phone-body">
        <div v-if="controls.length === 0" class="empty-state">
          <span class="empty-icon">📱</span>
          <span class="empty-text">暂无控件，请点击设计添加</span>
        </div>
        
        <div v-else class="controls-grid">
          <div 
            v-for="control in controls" 
            :key="control.id"
            class="control-item"
            :style="getControlStyle(control)"
          >
            <component 
              :is="getControlComponent(control.type)" 
              :control="control"
              :value="getControlValue(control)"
              @change="handleControlChange"
              @action="handleControlAction"
            />
          </div>
        </div>
      </div>
      
      <div class="phone-bottom">
        <div class="home-indicator"></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { getControlComponent, getControlConfig } from '../controls'

const props = defineProps({
  title: {
    type: String,
    default: '设备控制'
  },
  controls: {
    type: Array,
    default: () => []
  },
  properties: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['change', 'action'])

const controlValues = reactive({})

const getControlValue = (control) => {
  if (control.binds) {
    for (const key of Object.keys(control.binds)) {
      const bindKey = control.binds[key]
      if (bindKey && props.properties.length > 0) {
        const prop = props.properties.find(p => p.key === bindKey)
        if (prop) {
          return prop.default || (control.type === 'switch' ? false : 50)
        }
      }
    }
  }
  return controlValues[control.id] || (control.type === 'switch' ? false : 50)
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

const handleControlChange = (event) => {
  controlValues[event.control.id] = event.value
  emit('change', event)
}

const handleControlAction = (event) => {
  emit('action', event)
}
</script>

<style lang="scss" scoped>
.app-preview-container {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #f5f5f5;

  .preview-header {
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
  }

  .phone-frame {
    flex: 1;
    display: flex;
    flex-direction: column;
    max-width: 375px;
    max-height: calc(100vh - 100px);
    margin: 20px auto;
    background: #000;
    border-radius: 40px;
    padding: 12px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
    overflow: hidden;

    .phone-header {
      background: linear-gradient(135deg, #5BA3E8 0%, #1E3A5F 100%);
      border-radius: 30px 30px 0 0;
      padding-top: 30px;

      .status-bar {
        display: flex;
        justify-content: space-between;
        padding: 0 24px;
        color: #fff;
        font-size: 12px;
        margin-bottom: 16px;

        .time {
          font-weight: 600;
        }
      }

      .nav-bar {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0 20px 16px;
        color: #fff;

        .nav-back {
          font-size: 20px;
          cursor: pointer;
        }

        .nav-title {
          font-size: 18px;
          font-weight: 600;
        }

        .nav-more {
          font-size: 24px;
        }
      }
    }

    .phone-body {
      flex: 1;
      background: #F0F7FF;
      padding: 16px;
      overflow-y: auto;

      .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 400px;
        color: #999;

        .empty-icon {
          font-size: 64px;
          margin-bottom: 16px;
        }

        .empty-text {
          font-size: 14px;
        }
      }

      .controls-grid {
        display: flex;
        flex-wrap: wrap;
        gap: 12px;

        .control-item {
          flex: 1;
        }
      }
    }

    .phone-bottom {
      background: #000;
      padding: 12px;
      display: flex;
      justify-content: center;

      .home-indicator {
        width: 120px;
        height: 4px;
        background: #333;
        border-radius: 2px;
      }
    }
  }
}
</style>
