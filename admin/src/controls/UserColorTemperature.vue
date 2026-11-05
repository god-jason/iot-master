<template>
  <div class="control-color-temperature">
    <div class="temp-header">
      <span class="temp-label">{{ control.name }}</span>
      <div class="temp-value-box">
        <span class="temp-value">{{ currentValue }}</span>
        <span class="temp-unit">{{ control.unit || 'K' }}</span>
      </div>
    </div>
    <div class="temp-slider-container">
      <div class="temp-track">
        <div 
          class="temp-fill" 
          :style="{ width: fillWidth + '%', background: gradientBackground }"
        ></div>
        <div 
          class="temp-pointer" 
          :style="{ left: fillWidth + '%' }"
        ></div>
      </div>
      <input 
        type="range" 
        :min="control.min || 2700" 
        :max="control.max || 6500" 
        :step="control.step || 100"
        :value="currentValue"
        @input="handleInput"
        class="temp-range"
      />
    </div>
    <div class="temp-labels">
      <span class="temp-cool">冷色</span>
      <span class="temp-warm">暖色</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { toNum } from './Control.js'

const props = defineProps({
  control: {
    type: Object,
    required: true
  },
  value: {
    type: [String, Number, Boolean],
    default: 4000
  }
})

const emit = defineEmits(['change', 'action'])

const currentValue = ref(toNum(props.value, 4000))

const min = computed(() => props.control.min || 2700)
const max = computed(() => props.control.max || 6500)

const fillWidth = computed(() => {
  return ((currentValue.value - min.value) / (max.value - min.value)) * 100
})

const gradientBackground = computed(() => {
  return 'linear-gradient(to right, #1E90FF, #FF8C00)'
})

watch(() => props.value, (val) => {
  currentValue.value = toNum(val, 4000)
})

const handleInput = (event) => {
  currentValue.value = parseInt(event.target.value)
  emit('change', { type: 'color-temperature', value: currentValue.value, control: props.control, bindKey: props.control.binds?.value })
  if (props.control.events?.change?.key) {
    emit('action', {
      type: 'change',
      control: props.control,
      eventKey: 'change',
    })
  }
}
</script>

<style lang="scss" scoped>
.control-color-temperature {
  .temp-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;

    .temp-label {
      font-size: 14px;
      color: #333;
    }

    .temp-value-box {
      display: flex;
      align-items: baseline;
      gap: 2px;

      .temp-value {
        font-size: 13px;
        font-weight: 600;
        color: #5BA3E8;
      }

      .temp-unit {
        font-size: 12px;
        color: #999;
        margin-left: 2px;
      }
    }
  }

  .temp-slider-container {
    position: relative;
    height: 40px;
    display: flex;
    align-items: center;
    
    .temp-track {
      position: relative;
      width: 100%;
      height: 8px;
      background: #e0e0e0;
      border-radius: 4px;
      overflow: visible;
      
      .temp-fill {
        position: absolute;
        left: 0;
        top: 0;
        height: 100%;
        border-radius: 4px;
        transition: width 0.1s;
      }
      
      .temp-pointer {
        position: absolute;
        top: 50%;
        transform: translate(-50%, -50%);
        width: 20px;
        height: 20px;
        background: #fff;
        border-radius: 50%;
        border: 3px solid #5BA3E8;
        box-shadow: 0 2px 6px rgba(0,0,0,0.15);
        cursor: pointer;
        transition: left 0.1s;
      }
    }
    
    .temp-range {
      position: absolute;
      width: 100%;
      height: 100%;
      opacity: 0;
      cursor: pointer;
    }
  }

  .temp-labels {
    display: flex;
    justify-content: space-between;
    margin-top: 8px;
    
    .temp-cool {
      font-size: 12px;
      color: #1E90FF;
    }
    
    .temp-warm {
      font-size: 12px;
      color: #FF8C00;
    }
  }
}
</style>
