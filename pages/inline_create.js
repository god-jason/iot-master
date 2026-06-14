// 创建内联设备页面配置
return {
  title: '创建内联设备',
  icon: '/icons/inline.svg',
  template: 'edit',
  fields: [
    { key: 'id', label: 'ID', type: 'text', placeholder: '默认随机ID' },
    { key: 'name', label: '名称', type: 'text' },
    { key: 'description', label: '说明', type: 'text' },
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
          this.content.fields[3].tips = result.name
        }
      }
    },
    { 
      key: 'gateway_id', 
      label: '网关ID', 
      type: 'text',
      link_text: '选择',
      link_action: {
        type: 'dialog',
        page: 'device_choose',
        after_close(result, data, index) {
          this.editor.patchValue({ gateway_id: result.id })
          this.content.fields[4].tips = result.name
        }
      }
    },
    { key: 'link_id', label: '连接ID', type: 'text' },
    { key: 'disabled', label: '禁用', type: 'switch' }
  ],
  submit_api: 'table/inline/create',
  // 页面挂载时执行
  mount() {
    this.data.product_id = this.params.product_id
    this.data.gateway_id = this.params.gateway_id
    this.get_extend_fields()
  },
  methods: {
    get_extend_fields() {
      this.request.get('device/extend/fields').subscribe(res => {
        if (res.error) return
        (res.data || []).map(f => this.content.fields.push(f))
      })
    }
  }
}