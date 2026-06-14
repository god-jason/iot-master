// 上传页面配置
return {
  title: '上传',
  icon: '/icons/upload.svg',
  template: 'edit',
  fields: [{ key: 'file', label: '上传文件', type: 'file', upload: '/api/upload' }]
}
