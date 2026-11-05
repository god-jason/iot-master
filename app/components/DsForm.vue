<template>
  <view class="ds-form">
    <slot></slot>
  </view>
</template>

<script setup>
import { reactive } from 'vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  model: {
    type: Object,
    required: true
  },
  rules: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['validate', 'submit'])

const errors = reactive({})

const validateField = (field) => {
  const rule = props.rules[field]
  if (!rule) return true

  const value = props.model[field]
  const { required, min, max, pattern, message } = rule

  if (required && !value) {
    errors[field] = message || '该字段必填'
    emit('validate', { field, valid: false, error: errors[field] })
    return false
  }

  if (min != null && value.length < min) {
    errors[field] = message || `最少${min}个字符`
    emit('validate', { field, valid: false, error: errors[field] })
    return false
  }

  if (max != null && value.length > max) {
    errors[field] = message || `最多${max}个字符`
    emit('validate', { field, valid: false, error: errors[field] })
    return false
  }

  if (pattern && !pattern.test(value)) {
    errors[field] = message || '格式不正确'
    emit('validate', { field, valid: false, error: errors[field] })
    return false
  }

  delete errors[field]
  emit('validate', { field, valid: true })
  return true
}

const validate = () => {
  const fields = Object.keys(props.rules)
  let allValid = true

  fields.forEach(field => {
    if (!validateField(field)) {
      allValid = false
    }
  })

  return allValid
}

const submit = () => {
  const isValid = validate()
  emit('submit', { valid: isValid, errors: { ...errors } })
  return isValid
}

defineExpose({
  validate,
  validateField,
  submit,
  errors
})
</script>

<style lang="scss">
@import './variables.scss';

.ds-form {
  padding: $spacing-base;
}
</style>