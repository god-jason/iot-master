<template>
  <div class="control-hue">
    <div class="hue-header">
      <span class="hue-label">{{ control.name }}</span>
      <div class="hue-value-box">
        <span class="hue-value">{{ currentValue }}</span>
        <span class="hue-unit">°</span>
        <span class="hue-color-block" :style="{ background: currentHueColor }"></span>
      </div>
    </div>
    <div class="hue-slider-container">
      <div class="hue-track">
        <div class="hue-gradient"></div>
        <div 
          class="hue-pointer" 
          :style="{ left: fillWidth + '%', background: currentHueColor }"
        ></div>
      </div>
      <input 
        type="range" 
        :min="control.min || 0" 
        :max="control.max || 360" 
        :step="control.step || 1"
        :value="currentValue"
        @input="handleInput"
        class="hue-range"
      />
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
    default: 0
  }
})

const emit = defineEmits(['change', 'action'])

const currentValue = ref(toNum(props.value, 0))

const min = computed(() => props.control.min || 0)
const max = computed(() => props.control.max || 360)

const fillWidth = computed(() => {
  return ((currentValue.value - min.value) / (max.value - min.value)) * 100
})

const currentHueColor = computed(() => {
  return `hsl(${currentValue.value}, 100%, 50%)`
})

watch(() => props.value, (val) => {
  currentValue.value = toNum(val, 0)
})

const handleInput = (event) => {
  currentValue.value = parseInt(event.target.value)
  emit('change', { type: 'hue', value: currentValue.value, control: props.control, bindKey: props.control.binds?.value })
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
.control-hue {
  .hue-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 8px;

    .hue-label {
      font-size: 14px;
      color: #333;
    }

    .hue-value-box {
      display: flex;
      align-items: center;
      gap: 4px;

      .hue-value {
        font-size: 13px;
        font-weight: 600;
        color: #5BA3E8;
      }

      .hue-unit {
        font-size: 12px;
        color: #999;
      }

      .hue-color-block {
        display: inline-block;
        width: 14px;
        height: 14px;
        border-radius: 3px;
        border: 1px solid #e0e0e0;
        margin-left: 4px;
      }
    }
  }

  .hue-slider-container {
    position: relative;
    height: 40px;
    display: flex;
    align-items: center;
    
    .hue-track {
      position: relative;
      width: 100%;
      height: 8px;
      border-radius: 4px;
      overflow: visible;
      
      .hue-gradient {
        position: absolute;
        left: 0;
        top: 0;
        width: 100%;
        height: 100%;
        background: linear-gradient(to right, 
          hsl(0, 100%, 50%), 
          hsl(60, 100%, 50%), 
          hsl(120, 100%, 50%), 
          hsl(180, 100%, 50%), 
          hsl(240, 100%, 50%), 
          hsl(300, 100%, 50%), 
          hsl(360, 100%, 50%)
        );
        border-radius: 4px;
      }
      
      .hue-pointer {
        position: absolute;
        top: 50%;
        transform: translate(-50%, -50%);
        width: 20px;
        height: 20px;
        border-radius: 50%;
        border: 3px solid #fff;
        box-shadow: 0 2px 6px rgba(0,0,0,0.2);
        cursor: pointer;
        transition: left 0.1s;
      }
    }
    
    .hue-range {
      position: absolute;
      width: 100%;
      height: 100%;
      opacity: 0;
      cursor: pointer;
    }
  }
}
</style>
