<template>
  <div class="control-switch">
    <Icon 
      v-if="iconStyle.icon" 
      :name="iconStyle.icon" 
      :color="iconStyle.color"
      :size="iconStyle.font_size"
      class="switch-icon" 
    />
    <span class="switch-label">{{ control.name }}</span>
    <div 
      class="switch-wrapper"
      :class="{ active: currentValue }"
      :style="switchStyle"
      @click="toggle"
    >
      <div class="switch-track">
        <div class="switch-thumb"></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessageBox } from 'element-plus'
import { toBool } from './Control.js'
import Icon from './Icon.vue'

const props = defineProps({
  control: {
    type: Object,
    required: true
  },
  value: {
    type: [String, Boolean, Number],
    default: false
  }
})

const emit = defineEmits(['change', 'action'])

const iconStyle = computed(() => {
  const style = props.control.styles?.icon || {}
  return {
    icon: style.icon || '',
    color: style.color || '#999',
    font_size: style.font_size || '20px'
  }
})

const currentValue = ref(toBool(props.value))

watch(() => props.value, (val) => {
  currentValue.value = toBool(val)
})

const switchStyle = computed(() => {
  const style = props.control.styles?.default || {}
  const thumbStyle = props.control.styles?.thumb || {}
  const activeColor = style.color || '#5BA3E8'
  return {
    '--active-color': activeColor,
    '--thumb-bg': thumbStyle.background_color || '#ffffff',
    '--thumb-radius': thumbStyle.border_radius || '50%'
  }
})

const doToggle = () => {
  currentValue.value = !currentValue.value
  emit('change', { type: 'toggle', value: currentValue.value, control: props.control })
  if (props.control.events?.toggle?.key) {
    emit('action', {
      type: 'toggle',
      control: props.control,
      eventKey: 'toggle',
    })
  }
}

const toggle = () => {
  if (props.control.confirm) {
    ElMessageBox.confirm('确定要执行此操作吗？', '确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(() => {
      doToggle()
    }).catch(() => {})
  } else {
    doToggle()
  }
}
</script>

<style lang="scss" scoped>
.control-switch {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;

  .switch-icon {
    font-size: 20px;
  }

  .switch-label {
    flex: 1;
    font-size: 16px;
    color: #333;
  }

  .switch-wrapper {
    .switch-track {
      width: 120px;
      height: 64px;
      background: #e0e0e0;
      border-radius: 32px;
      position: relative;
      transition: background 0.3s;
      cursor: pointer;

      .switch-thumb {
        width: 56px;
        height: 56px;
        background: var(--thumb-bg, #fff);
        border-radius: var(--thumb-radius, 50%);
        position: absolute;
        top: 4px;
        left: 4px;
        transition: left 0.3s;
        box-shadow: 0 2px 4px rgba(0,0,0,0.2);
      }
    }

    &.active {
      .switch-track {
        background: var(--active-color);
      }

      .switch-thumb {
        left: 60px;
      }
    }
  }
}
</style>
