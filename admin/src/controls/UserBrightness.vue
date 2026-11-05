<template>
  <div class="control-brightness">
    <div class="brightness-header">
      <span class="brightness-label">{{ control.name }}</span>
    </div>
    <div class="brightness-value">{{ currentValue }}{{ control.unit || '%' }}</div>
    <div class="brightness-slider">
      <div class="brightness-track">
        <div 
          class="brightness-fill" 
          :style="{ width: fillWidth + '%' }"
        ></div>
        <div 
          class="brightness-pointer" 
          :style="{ left: fillWidth + '%' }"
        ></div>
      </div>
      <input 
        type="range" 
        :min="control.min || 0" 
        :max="control.max || 100" 
        :step="control.step || 1"
        :value="currentValue"
        @input="handleInput"
        class="brightness-input"
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

watch(() => props.value, (val) => {
  currentValue.value = toNum(val, 50)
})

const handleInput = (event) => {
  currentValue.value = parseFloat(event.target.value)
  emit('change', { type: 'brightness', value: currentValue.value, control: props.control })
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
.control-brightness {
  .brightness-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;

    .brightness-label {
      font-size: 14px;
      color: #333;
    }

    .brightness-icon {
      font-size: 24px;
    }
  }

  .brightness-value {
    font-size: 18px;
    font-weight: 600;
    color: #FFD700;
    text-align: right;
    margin-bottom: 8px;
  }

  .brightness-slider {
    position: relative;
    height: 40px;
    display: flex;
    align-items: center;

    .brightness-track {
      position: relative;
      width: 100%;
      height: 8px;
      background: linear-gradient(to right, #333, #FFD700);
      border-radius: 4px;
      overflow: visible;

      .brightness-fill {
        position: absolute;
        left: 0;
        top: 0;
        height: 100%;
        background: linear-gradient(to right, #FFD700, #FFD700);
        border-radius: 4px;
        transition: width 0.1s;
      }

      .brightness-pointer {
        position: absolute;
        top: 50%;
        transform: translate(-50%, -50%);
        width: 20px;
        height: 20px;
        background: #fff;
        border-radius: 50%;
        border: 3px solid #FFD700;
        box-shadow: 0 2px 6px rgba(0,0,0,0.15);
        cursor: pointer;
        transition: left 0.1s;
      }
    }

    .brightness-input {
      position: absolute;
      width: 100%;
      height: 100%;
      opacity: 0;
      cursor: pointer;
    }
  }
}
</style>
