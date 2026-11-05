import request from '@/utils/request'

// 上传图片（支持压缩）
export function uploadImage(file, options = {}) {
  const formData = new FormData()
  formData.append('file', file)
  if (options.width) formData.append('width', options.width)
  if (options.height) formData.append('height', options.height)
  if (options.format) formData.append('format', options.format)
  if (options.quality) formData.append('quality', options.quality)

  return request({
    url: '/upload/image',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// 上传图标（固定128x128）
export function uploadIcon(file) {
  const formData = new FormData()
  formData.append('file', file)

  return request({
    url: '/upload/icon',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}
