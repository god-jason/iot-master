<template>
  <div class="image-upload">
    <div class="upload-trigger" :class="{ 'has-image': modelValue }" @click="handleSelect">
      <img v-if="modelValue" :src="fullUrl" class="preview-image" :style="previewStyle" />
      <div v-else class="upload-placeholder">
        <el-icon class="upload-icon"><Plus /></el-icon>
        <span class="upload-text">{{ placeholder }}</span>
      </div>
      <div v-if="modelValue && !disabled" class="upload-overlay">
        <el-icon class="overlay-icon" @click.stop="handleSelect"><Edit /></el-icon>
        <el-icon class="overlay-icon" @click.stop="handleRemove"><Delete /></el-icon>
      </div>
    </div>
    <input
      ref="fileInputRef"
      type="file"
      accept="image/*"
      style="display: none"
      @change="handleChange"
    />
    <div class="upload-tip" v-if="tip">{{ tip }}</div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Plus, Edit, Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { uploadImage, uploadIcon } from '@/api/upload'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  placeholder: {
    type: String,
    default: '点击上传'
  },
  tip: {
    type: String,
    default: '建议使用正方形图片'
  },
  width: {
    type: Number,
    default: 0
  },
  height: {
    type: Number,
    default: 0
  },
  format: {
    type: String,
    default: 'jpg'
  },
  quality: {
    type: Number,
    default: 85
  },
  disabled: {
    type: Boolean,
    default: false
  },
  isIcon: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'success'])

const fileInputRef = ref(null)
const uploading = ref(false)

const baseUrl = import.meta.env.VITE_API_BASE_URL || ''

const fullUrl = computed(() => {
  if (!props.modelValue) return ''
  if (props.modelValue.startsWith('http')) return props.modelValue
  return baseUrl + props.modelValue
})

const previewStyle = computed(() => {
  if (props.width && props.height) {
    return {
      width: `${props.width}px`,
      height: `${props.height}px`
    }
  }
  return {
    width: '100%',
    height: '100%'
  }
})

const handleSelect = () => {
  if (props.disabled) return
  fileInputRef.value?.click()
}

const handleChange = async (e) => {
  const file = e.target.files[0]
  if (!file) return

  // 验证文件类型
  if (!file.type.startsWith('image/')) {
    ElMessage.error('请选择图片文件')
    return
  }

  // 验证文件大小 (5MB)
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.error('图片大小不能超过5MB')
    return
  }

  uploading.value = true
  try {
    let res
    if (props.isIcon) {
      res = await uploadIcon(file)
    } else {
      res = await uploadImage(file, {
        width: props.width,
        height: props.height,
        format: props.format,
        quality: props.quality
      })
    }

    emit('update:modelValue', res.data.url)
    emit('success', res.data)
    ElMessage.success('上传成功')
  } catch (e) {
    console.error(e)
    ElMessage.error('上传失败')
  } finally {
    uploading.value = false
    e.target.value = ''
  }
}

const handleRemove = () => {
  emit('update:modelValue', '')
}
</script>

<style lang="scss" scoped>
.image-upload {
  display: inline-block;

  .upload-trigger {
    position: relative;
    width: 120px;
    height: 120px;
    border: 1px dashed #d9d9d9;
    border-radius: 8px;
    cursor: pointer;
    overflow: hidden;
    transition: all 0.3s;

    &:hover {
      border-color: var(--primary-color);

      .upload-overlay {
        opacity: 1;
      }
    }

    &.has-image {
      border-style: solid;
    }

    .preview-image {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    .upload-placeholder {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      height: 100%;
      color: #999;

      .upload-icon {
        font-size: 28px;
        margin-bottom: 8px;
      }

      .upload-text {
        font-size: 12px;
      }
    }

    .upload-overlay {
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: rgba(0, 0, 0, 0.5);
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 12px;
      opacity: 0;
      transition: opacity 0.3s;

      .overlay-icon {
        font-size: 20px;
        color: #fff;
        cursor: pointer;

        &:hover {
          color: var(--primary-color);
        }
      }
    }
  }

  .upload-tip {
    margin-top: 8px;
    font-size: 12px;
    color: #999;
  }
}
</style>
