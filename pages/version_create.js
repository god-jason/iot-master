// 创建版本页面配置
return {
  title: '创建版本',
  icon: '/icons/version.svg',
  template: 'edit',
  fields: [
    { key: 'id', label: 'ID', type: 'text', placeholder: '默认随机ID' },
    { key: 'name', label: '名称', type: 'text' },
    { key: 'description', label: '说明', type: 'text' },
    { key: 'url', label: '固件', type: 'file', upload: '/api/upload' },
    { 
      key: 'product_id', 
      label: '产品ID', 
      type: 'text',
      link_text: '选择',
      link_action: {
        type: 'dialog',
        page: 'product_choose',
        after_close(result, data, index) {
          this.editor.patchValue({ product_id: result.id })
        }
      }
    },
    { key: 'disabled', label: '禁用', type: 'switch' }
  ],
  submit_api: 'table/version/create',
  // 页面挂载时执行
  mount() {
    this.data.product_id = this.params.product_id
  }
}