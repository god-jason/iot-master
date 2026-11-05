<template>
  <div class="control-slider">
    <span class="slider-label">{{ control.name }}</span>
    <div class="slider-value">{{ currentValue }}{{ control.unit || '' }}</div>
    <div class="slider-container">
      <div class="slider-track">
        <div 
          class="slider-fill" 
          :style="{ width: fillWidth + '%', backgroundColor: trackStyle.background_color || '#5BA3E8' }"
        ></div>
        <div 
          class="slider-pointer" 
          :style="{ 
            left: fillWidth + '%', 
            backgroundColor: pointerStyle.background_color || '#5BA3E8',
            width: pointerStyle.width || '20px',
            height: pointerStyle.height || '20px'
          }"
        ></div>
      </div>
      <input 
        type="range" 
        :min="control.min || 0" 
        :max="control.max || 100" 
        :step="control.step || 1"
        :value="currentValue"
        @input="handleInput"
        class="slider-input"
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
    default: 50
  }
})

const emit = defineEmits(['change', 'action'])

const currentValue = ref(toNum(props.value, 50))

const min = computed(() => props.control.min || 0)
const max = computed(() => props.control.max || 100)

const fillWidth = computed(() => {
  return ((currentValue.value - min.value) / (max.value - min.value)) * 100
})

const trackStyle = computed(() => props.control.styles?.default || {})
const pointerStyle = computed(() => props.control.styles?.pointer || {})

watch(() => props.value, (val) => {
  currentValue.value = toNum(val, 50)
})

const handleInput = (event) => {
  currentValue.value = parseFloat(event.target.value)
  emit('change', { type: 'input', value: currentValue.value, control: props.control })
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
.control-slider {
  .slider-label {
    font-size: 14px;
    color: #333;
    display: block;
    margin-bottom: 8px;
  }

  .slider-value {
    font-size: 18px;
    font-weight: 600;
    color: #5BA3E8;
    text-align: right;
    margin-bottom: 8px;
  }

  .slider-container {
    position: relative;
    height: 40px;
    display: flex;
    align-items: center;

    .slider-track {
      position: relative;
      width: 100%;
      height: 8px;
      background: #e0e0e0;
      border-radius: 4px;
      overflow: visible;

      .slider-fill {
        position: absolute;
        left: 0;
        top: 0;
        height: 100%;
        border-radius: 4px;
        transition: width 0.1s;
      }

      .slider-pointer {
        position: absolute;
        top: 50%;
        transform: translate(-50%, -50%);
        border-radius: 50%;
        box-shadow: 0 2px 6px rgba(0,0,0,0.15);
        cursor: pointer;
        transition: left 0.1s;
      }
    }

    .slider-input {
      position: absolute;
      width: 100%;
      height: 100%;
      opacity: 0;
      cursor: pointer;
    }
  }
}
</style>
